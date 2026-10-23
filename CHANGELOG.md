# Changelog

## 1.1.1 - 2026-10-23

### Changed

- Replace the `lib/pq` driver with `pgx`
- Build with Go 1.27.3
- Dependency updates

### Fixed

- Fix Go standard library vulnerabilities, including an HTTP/2 server crash
  (GO-2026-6617), HTTP/2 server memory exhaustion (GO-2026-6603) and an
  X.509 name constraint bypass (GO-2026-4866)

## 1.1.0 - 2026-04-04

- Dependency updates

## 1.0.0 - 2025-06-11

_Initial release._
