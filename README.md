# Umami Prometheus Exporter

A Prometheus metrics exporter for [Umami](https://umami.is/) website analytics
data. This exporter reads directly from Umami's PostgreSQL database to provide
website statistics as Prometheus metrics.

## Quick Start

Run the exporter using Docker:

```bash
docker run -p 8080:8080 \
  -e DATABASE_URL="postgres://user:password@host:5432/umami?sslmode=require" \
  -e WEBSITE_ID="your-website-uuid" \
  ghcr.io/csmith/umami-exporter:latest
```

Metrics will be available at `http://localhost:8080/metrics`.

## Configuration

The exporter is configured using environment variables or command-line flags:

| Environment Variable | Flag             | Description                              | Required |
|----------------------|------------------|------------------------------------------|----------|
| `DATABASE_URL`       | `--database-url` | PostgreSQL connection string             | Yes      |
| `WEBSITE_ID`         | `--website-id`   | Umami website UUID to export metrics for | Yes      |
| `PORT`               | `--port`         | HTTP server port (default: 8080)         | No       |
| `LOG_LEVEL`          | `--log.level`    | Log level: debug, info, warn, error      | No       |
| `LOG_FORMAT`         | `--log.format`   | Log format: text, json (default: text)   | No       |

### Database Connection

The `DATABASE_URL` should be a standard PostgreSQL connection string:
```
postgres://username:password@hostname:port/database?sslmode=require
```

### Website ID

Find your website ID in the Umami admin interface or by querying the `website`
table in your database.

## Exported Metrics

### `umami_page_views_total`
Counter tracking total page views with the following labels:
- `url_path` - The page path (e.g., `/`, `/about`)
- `referrer_domain` - Referring domain (empty string for direct visits)
- `browser` - Browser name (e.g., Chrome, Firefox)
- `os` - Operating system (e.g., Windows, macOS, Linux)
- `device` - Device type (e.g., desktop, mobile, tablet)
- `country` - Two-letter country code (e.g., US, GB, DE)

### `umami_pages_per_visit`
Histogram showing the distribution of pages viewed per visit, with buckets at
1, 2, 3, 5, 10, 20, 50, and 100 pages.

## ⚠️ Caveats

This exporter can generate metrics with very high cardinality due to the
`url_path` label. If your website has many unique URLs (dynamic paths, query
parameters, etc.), this could create thousands or millions of unique metric
series.

The exporter queries the entire dataset for the website in question, with no 
time limit. This may get slower and slower over time.

## Docker Compose Example

```yaml
services:
  umami:
    # ...

  umami-db:
    #...

  umami-exporter:
    image: ghcr.io/csmith/umami-exporter:latest
    environment:
      DATABASE_URL: "postgres://umami:umami@unami-db:5432/umami"
      WEBSITE_ID: "your-website-uuid-here"
      LOG_LEVEL: "info"
    depends_on:
      - unami-db
```

## Contributions/issues/etc

Pull requests and issues are more than welcome. There are probably better ways
to expose the data, more interesting things that could be exported, and so on.
I'm open to suggestions!