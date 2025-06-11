package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/csmith/envflag/v2"
	"github.com/csmith/slogflags"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	databaseURL = flag.String("database-url", "", "PostgreSQL connection URL (postgres://user:pass@host:port/dbname?sslmode=...)")
	websiteID   = flag.String("website-id", "", "Umami website ID (UUID)")
	port        = flag.Int("port", 8080, "HTTP server port")
)

func main() {
	envflag.Parse()
	_ = slogflags.Logger(slogflags.WithSetDefault(true))

	if *databaseURL == "" {
		slog.Error("database-url is required")
		os.Exit(1)
	}

	if *websiteID == "" {
		slog.Error("website-id is required")
		os.Exit(1)
	}

	collector, err := NewUmamiCollector(*databaseURL, *websiteID)
	if err != nil {
		slog.Error("failed to create collector", "error", err)
		os.Exit(1)
	}

	prometheus.MustRegister(collector)

	http.Handle("/metrics", promhttp.Handler())

	addr := fmt.Sprintf(":%d", *port)
	slog.Info("starting server", "addr", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		slog.Error("server failed", "error", err)
	}
}
