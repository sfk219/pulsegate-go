# PulseGate — Concurrent API Gateway (Go)

A runnable educational reverse proxy with per-client token-bucket rate limiting, circuit breaking, and Prometheus-style metrics. Uses only the Go standard library.

## Run
```bash
go test -race ./...
go run ./cmd/pulsegate
```
Start an upstream on `http://localhost:9000` or set `UPSTREAM_URL`. The gateway listens on port 8080. Metrics are available at `GET /_metrics`.

## Limitations
This is an independently authored **single-instance** portfolio demonstration, not a production API gateway. Rate-limit buckets are in-memory, clients are identified by remote IP (not trusted forwarding headers), and the circuit breaker protects only one upstream.
