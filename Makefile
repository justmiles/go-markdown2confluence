.PHONY: build
VERSION=`git describe --tags --abbrev=0`

build:
	goreleaser release --snapshot --skip-publish --rm-dist

release:
	goreleaser release --rm-dist

push-docker:
	docker build . -t justmiles/markdown2confluence --build-arg VERSION=$(VERSION)
	docker tag justmiles/markdown2confluence justmiles/markdown2confluence:$(VERSION)
	docker push justmiles/markdown2confluence
	docker push justmiles/markdown2confluence:$(VERSION)
