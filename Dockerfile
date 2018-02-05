FROM golang:1.9.3-stretch as build

WORKDIR /go/src/upcloud-proxy
COPY upcloud-proxy.go .

RUN go get github.com/elazarl/goproxy \
    && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o upcloud-proxy upcloud-proxy.go

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /go/src/upcloud-proxy /
CMD ["/upcloud-proxy"]
