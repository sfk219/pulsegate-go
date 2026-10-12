package gateway

import (
    "fmt"
    "net"
    "net/http"
    "net/http/httputil"
    "net/url"
    "sync/atomic"
    "time"
)

type Gateway struct {
    limiter *Limiter
    breaker *Breaker
    proxy *httputil.ReverseProxy
    received atomic.Uint64
    throttled atomic.Uint64
    rejected atomic.Uint64
    upstreamErrors atomic.Uint64
}

func New(upstream *url.URL, rate float64, burst, threshold int, cooldown time.Duration) *Gateway {
    g := &Gateway{limiter:NewLimiter(rate,burst), breaker:NewBreaker(threshold,cooldown)}
    proxy := httputil.NewSingleHostReverseProxy(upstream)
    transport := http.DefaultTransport.(*http.Transport).Clone()
    transport.ResponseHeaderTimeout = 5 * time.Second
    proxy.Transport = transport
    proxy.ModifyResponse = func(res *http.Response) error {
        g.breaker.Report(res.StatusCode < 500)
        if res.StatusCode >= 500 { g.upstreamErrors.Add(1) }
        return nil
    }
    proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
        g.breaker.Report(false)
        g.upstreamErrors.Add(1)
        http.Error(w,"upstream unavailable",http.StatusBadGateway)
    }
    g.proxy = proxy
    return g
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter,r *http.Request) {
    if r.URL.Path == "/_metrics" {
        w.Header().Set("Content-Type","text/plain; version=0.0.4")
        fmt.Fprintf(w,"# TYPE pulsegate_requests_total counter\npulsegate_requests_total %d\n",g.received.Load())
        fmt.Fprintf(w,"pulsegate_throttled_total %d\npulsegate_rejected_total %d\npulsegate_upstream_errors_total %d\n",g.throttled.Load(),g.rejected.Load(),g.upstreamErrors.Load())
        return
    }
    g.received.Add(1)
    ip, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil { ip = r.RemoteAddr }
    if !g.limiter.Allow(ip) {
        g.throttled.Add(1)
        http.Error(w,"rate limit exceeded",http.StatusTooManyRequests)
        return
    }
    if !g.breaker.Allow() {
        g.rejected.Add(1)
        http.Error(w,"upstream circuit open",http.StatusServiceUnavailable)
        return
    }
    g.proxy.ServeHTTP(w,r)
}
