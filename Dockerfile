FROM golang:1.14

WORKDIR /go/src/app

RUN curl -sfL https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh | sh

