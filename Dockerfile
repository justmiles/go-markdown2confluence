ARG VERSION=latest

FROM golang:1.14 as builder
ARG VERSION

WORKDIR /go/src

COPY . /go/src

RUN CGO_ENABLED=0 GOOS=linux go build -mod vendor -a -o markdown2confluence -ldflags "-s -w -X main.version=${VERSION}"
RUN md5sum markdown2confluence

# Create image from scratch
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/src/markdown2confluence /markdown2confluence
COPY --from=builder /tmp /tmp

ENTRYPOINT [ "/markdown2confluence" ]
