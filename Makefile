
build:
	env GOOS=windows GOARCH=amd64 go build -o build/markdown2confluence.windows-amd64.exe
	env GOOS=linux GOARCH=amd64 go build -o build/markdown2confluence.linux-amd64
	env GOOS=darwin GOARCH=amd64 go build -o build/markdown2confluence.darwin-amd64
