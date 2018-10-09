FROM golang:1.11.0-stretch as build

WORKDIR /src/upcloud-proxy

COPY go.* *.go ./

RUN CGO_ENABLED=0 go build -a -o upcloud-proxy

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /src/upcloud-proxy/upcloud-proxy /
CMD ["/upcloud-proxy"]
