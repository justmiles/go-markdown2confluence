# markdown2confluence
Push markdown files to Confluence Cloud

## Installation

Download the [latest release](https://github.com/justmiles/go-markdown2confluence/releases) and add the binary in your local `PATH`

- Linux

        curl -LO https://github.com/justmiles/go-markdown2confluence/releases/download/v3.0.2/markdown2confluence_3.0.2_linux_x86_64.tar.gz
        sudo tar -xzvf markdown2confluence_3.0.2_linux_x86_64.tar.gz -C /usr/local/bin/ markdown2confluence

- OSX

        curl -LO https://github.com/justmiles/go-markdown2confluence/releases/download/v3.0.2/markdown2confluence_3.0.2_darwin_x86_64.tar.gz
        sudo tar -xzvf markdown2confluence_3.0.2_linux_x86_64.tar.gz -C /usr/local/bin/ markdown2confluence

- Windows
  - Download [the latest release](https://github.com/justmiles/go-markdown2confluence/releases/download/v3.0.2/markdown2confluence_3.0.2_windows_x86_64.zip) and add to your system `PATH`

## Environment Variables
For best practice we recommend you [authenticate using an API token](https://id.atlassian.com/manage/api-tokens).

- CONFLUENCE_USERNAME - username for Confluence Cloud. When using API tokens set this to your full email.
- CONFLUENCE_PASSWORD - API token or password for Confluence Cloud
- CONFLUENCE_ENDPOINT - endpoint for Confluence Cloud, eg `https://mycompanyname.atlassian.net/wiki`

## Usage

    Push markdown files to Confluence Cloud

    Usage:
    markdown2confluence [flags] (files or directories)

    Flags:
    -d, --debug                Enable debug logging
    -e, --endpoint string      Confluence endpoint. (Alternatively set CONFLUENCE_ENDPOINT environment variable) (default "https://mydomain.atlassian.net/wiki")
    -h, --help                 help for markdown2confluence
    -m, --modified-since int   Only upload files that have modifed in the past n minutes
    -p, --password string      Confluence password. (Alternatively set CONFLUENCE_PASSWORD environment variable)
    -s, --space string         Space in which page should be created
    -u, --username string      Confluence username. (Alternatively set CONFLUENCE_USERNAME environment variable)
        --version              version for markdown2confluence

## Examples

Upload a local directory of markdown files called `markdown-files` to Confluence.

    markdown2confluence --space 'MyTeamSpace' markdown-files

Upload the same directory, but only those modified in the last 30 minutes. This is particurlarly useful for cron jobs/recurring one-way syncs.

    markdown2confluence --space 'MyTeamSpace' --modified-since 30 markdown-files

Upload a single file

    markdown2confluence --space 'MyTeamSpace' markdown-files/test.md

Upload a directory of markdown files in space `MyTeamSpace` under the parent page `API Docs`

    markdown2confluence --space 'MyTeamSpace' --parent 'API Docs' markdown-files
