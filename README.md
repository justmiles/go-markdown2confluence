# markdown2confluence

Push markdown files to Confluence Cloud

[![Build Status](https://drone.justmiles.io/api/badges/justmiles/go-markdown2confluence/status.svg)](https://drone.justmiles.io/justmiles/go-markdown2confluence)

## Installation

Download the [latest
release](https://github.com/justmiles/go-markdown2confluence/releases)
and add the binary in your local `PATH`

- Linux

  ```shell
  curl -LO https://github.com/justmiles/go-markdown2confluence/releases/download/v4.0.0/go-markdown2confluence_4.0.0_linux_x86_64.tar.gz

  tar -xzvf go-markdown2confluence_4.0.0_linux_x86_64.tar.gz -C $HOME/.local/bin markdown2confluence
  ```

- OSX

  ```shell
  curl -LO https://github.com/justmiles/go-markdown2confluence/releases/download/v4.0.0/go-markdown2confluence_4.0.0_darwin_x86_64.tar.gz

  tar -xzvf go-markdown2confluence_4.0.0_darwin_x86_64.tar.gz -C $HOME/.local/bin markdown2confluence
  ```

- Windows

  Download [the latest release](https://github.com/justmiles/go-markdown2confluence/releases/download/v4.0.0/go-markdown2confluence_4.0.0_windows_x86_64.zip) and add to your system `PATH`

- or docker

```shell
docker run justmiles/markdown2confluence --version
```

## Environment Variables

For best practice we recommend you [authenticate using an API token](https://id.atlassian.com/manage/api-tokens).

- CONFLUENCE_ENDPOINT - endpoint for Confluence Cloud, eg `https://mycompanyname.atlassian.net/wiki`
- CONFLUENCE_USERNAME - Confluence username. When using API token this should be a valid email address.
- CONFLUENCE_PASSWORD - API token or password for Confluence Cloud
- CONFLUENCE_API_TOKEN - API token for Confluence Cloud
- CONFLUENCE_ACCESS_TOKEN - Bearer access token for Confluence Data Center

## Usage

```txt
Usage:
  markdown2confluence sync [flags]

Flags:
      --auto-approve         Automatically approve changes
  -c, --comment string       (Optional) Add comment to page
  -x, --exclude strings      Regex expressions to exclude matching files or file paths
  -f, --force                Force an upload regardless of whether or not it changed locally
  -w, --hardwraps            Render newlines as <br />
  -h, --help                 help for sync
      --parent string        Optional parent page to nest content under
  -t, --title string         Set the page title on upload (defaults to filename without extension)
      --use-document-title   Use Markdown document title (# Title) if available

Global Flags:
  -a, --access-token string   Access token for Confluence Data Center (CONFLUENCE_ACCESS_TOKEN environment variable can be used as an alternative)
      --api-token string      API token for Confluence Cloud (CONFLUENCE_API_TOKEN environment variable can be used as an alternative)
  -e, --endpoint string       Confluence endpoint (CONFLUENCE_ENDPOINT environment variable can be used as an alternative) (default "https://mydomain.atlassian.net/wiki")
  -i, --insecuretls           Skip certificate validation (e.g., for self-signed certificates)
  -l, --local-store string    Path to the local storage database (default "markdown2confluence.db")
      --log-level string      Verbosity log level (error, info, debug, or trace) (default "error")
  -p, --password string       Confluence password (CONFLUENCE_PASSWORD environment variable can be used as an alternative)
  -s, --space string          Space in which content should be created
  -u, --username string       Confluence username (CONFLUENCE_USERNAME environment variable can be used as an alternative)

---

Usage:
  markdown2confluence sync [flags]

Flags:
  -c, --comment string       (Optional) Add comment to page
  -x, --exclude strings      regex expression to exclude matching files or file paths
  -w, --hardwraps            Render newlines as <br />
  -h, --help                 help for sync
      --parent string        Optional parent page to nest content under
  -t, --title string         Set the page title on upload (defaults to filename without extension)
      --use-document-title   Will use the Markdown document title (# Title) if available
```

## Examples

Upload a local directory of markdown files called `markdown-files` to Confluence.

```shell
markdown2confluence sync \
  --space 'MyTeamSpace' \
  markdown-files
```

Upload a directory of markdown files in space `MyTeamSpace` under the parent page `API Docs`

```shell
markdown2confluence sync \
  --space 'MyTeamSpace' \
  --parent 'API Docs' \
  markdown-files
```

Use the markdown _document-title_ instead of the filname as document title (if available) in Confluence.

```shell
markdown2confluence sync \
  --space 'MyTeamSpace' \
  --use-document-title \
   markdown-files
```

## Confluence Specific Markup

It is possible to insert Confluence macros using fenced code blocks.
The "language" for this is `CONFLUENCE-MACRO`, exactly like that in all-caps.
Here is an example for a ToC macro using all headlines starting at Level 2:

````markdown
    # Title

    ```CONFLUENCE-MACRO
    name:toc
    schema-version:1
      minLevel:2
    ```

    ## Section 1
````

In general almost all macros should be possible.
The general syntax is:

````markdown
    ```CONFLUENCE-MACRO
    name:Name of Macro
    schema-version:Schema Version (use `1`)
      attribute:Value of Attribute
      parameter-name:Value of Parameter
      next-parameter:Value of Parameter
    ```
````
