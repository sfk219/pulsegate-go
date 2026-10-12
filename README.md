# PulseGate — Concurrent API Gateway

A runnable **Go reverse proxy** showing how backend infrastructure protects upstream services. Uses the Go standard library.

### Technical features
- Per-client token-bucket limiting with configurable rate and burst
- Circuit breaker rejecting traffic after repeated upstream 5xx or transport failures
- Reverse proxy with an upstream timeout
- Prometheus-format metrics at `/_metrics`
- Docker multi-stage build, concurrent HTTP tests, and a GitHub Actions workflow

### Run it
```bash
go test -race ./...
go run ./cmd/pulsegate
```
By default PulseGate proxies requests received at `localhost:8080` to `localhost:9000`. Set `UPSTREAM_URL=https://example.org` or another upstream. Example:
```bash
UPSTREAM_URL=https://httpbin.org REQUESTS_PER_SECOND=5 BURST=10 go run ./cmd/pulsegate
curl http://localhost:8080/get
curl http://localhost:8080/_metrics
```

**Architecture:** Client → rate limiter → circuit breaker → HTTP reverse proxy → upstream. Failed upstream responses affect the circuit breaker; clients see 429 when throttled and 503 when the circuit is open.

**Scope:** independent portfolio demonstration. In-memory limiter and breaker do not coordinate across replicas; buckets are not automatically purged; no authentication or gateway control plane. Never describe this as production-proven.
