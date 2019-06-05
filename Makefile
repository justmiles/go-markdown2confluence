.PHONY: build
VERSION=2.0.0

build: COMMIT=$(shell git rev-list -1 HEAD | grep -o "^.\{10\}")
build: DATE=$(shell date +'%Y-%m-%d %H:%M')
build: 
	env GOOS=darwin  GOARCH=amd64 go build -ldflags '-X "main.Version=$(VERSION) ($(COMMIT) - $(DATE))"' -o build/$(VERSION)/markdown2confluence-$(VERSION)-darwin-amd64
	env GOOS=linux   GOARCH=amd64 go build -ldflags '-X "main.Version=$(VERSION) ($(COMMIT) - $(DATE))"' -o build/$(VERSION)/markdown2confluence-$(VERSION)-linux-amd64
	env GOOS=windows GOARCH=amd64 go build -ldflags '-X "main.Version=$(VERSION) ($(COMMIT) - $(DATE))"' -o build/$(VERSION)/markdown2confluence-$(VERSION)-windows-amd64.exe
