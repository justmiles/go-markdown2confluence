.PHONY: build
VERSION=3.0.0

build:
	goreleaser release --snapshot --skip-publish --rm-dist

release-test:
	goreleaser release --skip-publish --rm-dist

release:
	goreleaser release --rm-dist