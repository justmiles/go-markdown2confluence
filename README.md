# markdown2confluence
Push markdown files to Confluence Cloud

## Installation

    sudo curl -L https://github.com/justmiles/go-markdown2confluence/releases/download/v2.0.0/markdown2confluence-2.0.0-linux-amd64 -o /usr/local/bin/markdown2confluence
    sudo chmod +x /usr/local/bin/markdown2confluence

## Environment Variables

- CONFLUENCE_USERNAME - username for Confluence Cloud
- CONFLUENCE_PASSWORD - password for Confluence Cloud
- CONFLUENCE_ENDPOINT - endpoint for Confluence Cloud, eg `https://mycompanyname.atlassian.net/wiki`

## Usage of markdown2confluence:

    Push markdown files to Confluence Cloud

    Usage:
    markdown2confluence [flags] (files or directories)

    Flags:
    -d, --debug             Enable debug logging
    -e, --endpoint string   Confluence endpoint. (Alternatively set CONFLUENCE_ENDPOINT environment variable) (default "https://mydomain.atlassian.net/wiki")
    -h, --help              help for markdown2confluence
    -p, --password string   Confluence password. (Alternatively set CONFLUENCE_PASSWORD environment variable)
    -s, --space string      Space in which page should be created
    -u, --username string   Confluence username. (Alternatively set CONFLUENCE_USERNAME environment variable)
        --version           version for markdown2confluence
