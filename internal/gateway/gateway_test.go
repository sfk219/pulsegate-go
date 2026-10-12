package gateway

import (
    "io"
    "net/http"
    "net/http/httptest"
    "net/url"
    "strings"
    "testing"
    "time"
)
func TestTokenBucket(t *testing.T) {
    l:=NewLimiter(0.0001,2)
    if !l.Allow("alice") || !l.Allow("alice") || l.Allow("alice") { t.Fatal("incorrect token limit") }
    if !l.Allow("bob") { t.Fatal("clients should have separate buckets") }
}
func TestBreaker(t *testing.T) {
    b:=NewBreaker(2,time.Hour)
    if !b.Allow(){t.Fatal("unexpected open breaker")}
    b.Report(false);b.Report(false)
    if b.Allow(){t.Fatal("breaker should reject")}
    b.Report(true)
    if !b.Allow(){t.Fatal("success should reset breaker")}
}
func TestGatewayProxyAndRateLimit(t *testing.T) {
    upstream:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){io.WriteString(w,"proxied:"+r.URL.Path)}))
    defer upstream.Close()
    target,_:=url.Parse(upstream.URL)
    gw:=httptest.NewServer(New(target,0.0001,1,3,time.Second))
    defer gw.Close()
    res,err:=http.Get(gw.URL+"/hello")
    if err!=nil{t.Fatal(err)}
    data,_:=io.ReadAll(res.Body);res.Body.Close()
    if res.StatusCode!=200 || string(data)!="proxied:/hello" {t.Fatalf("%d: %s",res.StatusCode,data)}
    res,err=http.Get(gw.URL+"/hello")
    if err!=nil{t.Fatal(err)}
    res.Body.Close()
    if res.StatusCode!=429 { t.Fatalf("expected 429, got %d",res.StatusCode) }
    res,err=http.Get(gw.URL+"/_metrics")
    if err!=nil{t.Fatal(err)}
    data,_=io.ReadAll(res.Body);res.Body.Close()
    if !strings.Contains(string(data),"pulsegate_throttled_total 1"){t.Fatal(string(data))}
}
func TestBreakerRejectsFailedUpstream(t *testing.T) {
    calls:=0
    upstream:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){calls++;w.WriteHeader(503)}))
    defer upstream.Close()
    target,_:=url.Parse(upstream.URL)
    gw:=httptest.NewServer(New(target,100,10,1,time.Hour))
    defer gw.Close()
    for i,code:=range []int{503,503} {
        resp,err:=http.Get(gw.URL+"/resource");if err!=nil{t.Fatal(err)}
        resp.Body.Close()
        if resp.StatusCode!=code{t.Fatalf("call %d returned %d",i,resp.StatusCode)}
    }
    if calls!=1{t.Fatalf("upstream called %d times; expected 1",calls)}
}
