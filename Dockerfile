FROM golang:1.22-alpine AS build
WORKDIR /source
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -o /pulsegate ./cmd/pulsegate

FROM alpine:3.20
RUN apk --no-cache add ca-certificates && adduser -D -u 10001 app
COPY --from=build /pulsegate /usr/local/bin/pulsegate
USER app
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/pulsegate"]
