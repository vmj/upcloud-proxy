FROM golang:1.9.3-stretch as build

WORKDIR /go/src/
COPY vendor/github.com ./github.com
COPY upcloud-proxy.go ./upcloud-proxy.go

RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o upcloud-proxy

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /go/src/upcloud-proxy /
CMD ["/upcloud-proxy"]
