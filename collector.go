package main

import (
	"database/sql"
	"log/slog"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	pageViewsQuery = `
		SELECT 
			we.url_path,
			COALESCE(we.referrer_domain, '') as referrer_domain,
			COALESCE(s.browser, '') as browser,
			COALESCE(s.os, '') as os,
			COALESCE(s.device, '') as device,
			COALESCE(s.country, '') as country,
			COUNT(*) as count
		FROM website_event we
		JOIN session s ON we.session_id = s.session_id
		WHERE we.website_id = $1 AND we.event_type = 1
		GROUP BY we.url_path, we.referrer_domain, s.browser, s.os, s.device, s.country`

	pagesPerVisitQuery = `
		SELECT 
			visit_id,
			COUNT(*) as page_count
		FROM website_event
		WHERE website_id = $1 AND event_type = 1
		GROUP BY visit_id`
)

type DatabaseInterface interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type UmamiCollector struct {
	db        DatabaseInterface
	websiteID string

	pageViewsDesc     *prometheus.Desc
	pagesPerVisitDesc *prometheus.Desc
}

func NewUmamiCollector(databaseURL, websiteID string) (*UmamiCollector, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return NewUmamiCollectorWithDB(db, websiteID), nil
}

func NewUmamiCollectorWithDB(db DatabaseInterface, websiteID string) *UmamiCollector {
	return &UmamiCollector{
		db:        db,
		websiteID: websiteID,
		pageViewsDesc: prometheus.NewDesc(
			"umami_page_views_total",
			"Total number of page views",
			[]string{"url_path", "referrer_domain", "browser", "os", "device", "country"},
			nil,
		),
		pagesPerVisitDesc: prometheus.NewDesc(
			"umami_pages_per_visit",
			"Number of pages viewed per visit",
			nil,
			nil,
		),
	}
}

func (c *UmamiCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.pageViewsDesc
	ch <- c.pagesPerVisitDesc
}

func (c *UmamiCollector) Collect(ch chan<- prometheus.Metric) {
	c.collectPageViews(ch)
	c.collectPagesPerVisit(ch)
}

func (c *UmamiCollector) collectPageViews(ch chan<- prometheus.Metric) {
	rows, err := c.db.Query(pageViewsQuery, c.websiteID)
	if err != nil {
		slog.Error("failed to query page views", "error", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var urlPath, referrerDomain, browser, os, device, country string
		var count float64

		err := rows.Scan(&urlPath, &referrerDomain, &browser, &os, &device, &country, &count)
		if err != nil {
			slog.Error("failed to scan page view row", "error", err)
			continue
		}

		metric, err := prometheus.NewConstMetric(
			c.pageViewsDesc,
			prometheus.CounterValue,
			count,
			urlPath, referrerDomain, browser, os, device, country,
		)
		if err != nil {
			slog.Error("failed to create page view metric", "error", err)
			continue
		}

		ch <- metric
	}

	if err := rows.Err(); err != nil {
		slog.Error("error iterating page views", "error", err)
	}
}

func (c *UmamiCollector) collectPagesPerVisit(ch chan<- prometheus.Metric) {
	rows, err := c.db.Query(pagesPerVisitQuery, c.websiteID)
	if err != nil {
		slog.Error("failed to query pages per visit", "error", err)
		return
	}
	defer rows.Close()

	buckets := make(map[float64]uint64)
	var totalCount uint64
	var totalSum float64

	bucketBoundaries := []float64{1, 2, 3, 5, 10, 20, 50, 100}
	for _, boundary := range bucketBoundaries {
		buckets[boundary] = 0
	}

	for rows.Next() {
		var visitID string
		var pageCount float64

		err := rows.Scan(&visitID, &pageCount)
		if err != nil {
			slog.Error("failed to scan pages per visit row", "error", err)
			continue
		}

		totalCount++
		totalSum += pageCount

		for _, bucket := range bucketBoundaries {
			if pageCount <= bucket {
				buckets[bucket]++
			}
		}
	}

	if err := rows.Err(); err != nil {
		slog.Error("error iterating pages per visit", "error", err)
		return
	}

	if totalCount > 0 {
		metric, err := prometheus.NewConstHistogram(
			c.pagesPerVisitDesc,
			totalCount,
			totalSum,
			buckets,
		)
		if err != nil {
			slog.Error("failed to create pages per visit histogram", "error", err)
			return
		}

		ch <- metric
	}
}
