FROM vmj0/golang-dep:1.11.0-stretch-0.5.0 as build

WORKDIR /go/src/github.com/vmj/upcloud-proxy

COPY Gopkg.* *.go ./

RUN dep ensure && \
    CGO_ENABLED=0 go build -a -o upcloud-proxy

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /go/src/github.com/vmj/upcloud-proxy/upcloud-proxy /
CMD ["/upcloud-proxy"]
