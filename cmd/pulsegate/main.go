package main

import (
    "log"
    "net/http"
    "net/url"
    "os"
    "strconv"
    "time"
    "github.com/sfk219/pulsegate-go/internal/gateway"
)
func number(name string, fallback float64) float64 {
    raw:=os.Getenv(name)
    if raw=="" { return fallback }
    n,err:=strconv.ParseFloat(raw,64)
    if err!=nil{log.Fatalf("invalid %s: %v",name,err)}
    return n
}
func main(){
    upstream:=os.Getenv("UPSTREAM_URL")
    if upstream==""{upstream="http://localhost:9000"}
    target,err:=url.Parse(upstream)
    if err!=nil || (target.Scheme!="http" && target.Scheme!="https") || target.Host=="" {log.Fatal("UPSTREAM_URL must be an HTTP(S) URL")}
    port:=os.Getenv("PORT")
    if port==""{port="8080"}
    handler:=gateway.New(target,number("REQUESTS_PER_SECOND",5),int(number("BURST",10)),int(number("FAILURE_THRESHOLD",3)),time.Duration(number("COOLDOWN_SECONDS",10))*time.Second)
    log.Printf("PulseGate listening on :%s -> %s",port,target)
    log.Fatal(http.ListenAndServe(":"+port,handler))
}
