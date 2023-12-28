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

- or with docker

```shell
docker run justmiles/markdown2confluence --version
```

## Environment Variables

For best practice we recommend you [authenticate using an API token](https://id.atlassian.com/manage/api-tokens).

However, you may also use [Personal Access Tokens](https://confluence.atlassian.com/enterprise/using-personal-access-tokens-1026032365.html),
which may help if your company uses SSO.

- CONFLUENCE_USERNAME - username for Confluence Cloud. When using API tokens set this to your full email.
- CONFLUENCE_PASSWORD - API token or password for Confluence Cloud
- CONFLUENCE_ENDPOINT - endpoint for Confluence Cloud, eg `https://mycompanyname.atlassian.net/wiki`
- CONFLUENCE_ACCESS_TOKEN - Bearer access token to use (instead of API token)

## Usage

```txt
A fast and flexible tool to syncronize or migrate your markdown documents to Confluence

Usage:
  markdown2confluence [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  purge-space Delete all pages from a Space - useful for a fresh sync
  sync        Sync markdown files to Confluence

Flags:
  -a, --access-token string   access-token for Confluence Data Center. (Alternatively set CONFLUENCE_ACCESS_TOKEN environment variable)
      --api-token string      api-token for Confluence Cloud. (Alternatively set CONFLUENCE_API_TOKEN environment variable)
  -d, --debug                 Enable debug logging
  -e, --endpoint string       Confluence endpoint. (Alternatively set CONFLUENCE_ENDPOINT environment variable) (default "https://mydomain.atlassian.net/wiki")
  -h, --help                  help for markdown2confluence
  -i, --insecuretls           Skip certificate validation. (e.g. for self-signed certificates)
  -p, --password string       Confluence password. (Alternatively set CONFLUENCE_PASSWORD environment variable)
  -s, --space string          Space in which content should be created
  -u, --username string       Confluence username. (Alternatively set CONFLUENCE_USERNAME environment variable)
  -v, --version               version for markdown2confluence

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

Upload a single file

```shell
markdown2confluence sync \
  --space 'MyTeamSpace' \
  markdown-files/test.md
```

Upload a directory of markdown files in space `MyTeamSpace` under the parent page `API Docs`

```shell
markdown2confluence \
  --space 'MyTeamSpace' \
  --parent 'API Docs' \
  markdown-files
```

Upload a directory of markdown files in space `MyTeamSpace` under a _nested_ parent page `Docs/API` and _exclude_ mardown files/directories that match `.*generated.*` or `.*temp.md`

```shell
markdown2confluence \
  --space 'MyTeamSpace' \
  --parent 'API/Docs' \
  --exclude '.*generated.*' \
  --exclude '.*temp.md' \
   markdown-files
```

Upload a directory of markdown files in space `MyTeamSpace` under the parent page `API Docs` and use the markdown _document-title_ instead of the filname as document title (if available) in Confluence.

```shell
markdown2confluence \
  --space 'MyTeamSpace' \
  --parent 'API Docs' \
  --use-document-title \
   markdown-files
```

## Enhancements

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

So a fully fledged macro could look like:

````markdown
    ```CONFLUENCE-MACRO
    name:toc
    schema-version:1
      maxLevel:5
      minLevel:2
      exclude:Beispiel.*
      style:none
      type:flat
      separator:pipe
    ```
````

Which will translate to:

```XML
<ac:structured-macro ac:name="toc" ac:schema-version="1" >
  <ac:parameter ac:name="maxLevel">5</ac:parameter>
  <ac:parameter ac:name="minLevel">2</ac:parameter>
  <ac:parameter ac:name="exclude">Beispiel.*</ac:parameter>
  <ac:parameter ac:name="style">none</ac:parameter>
  <ac:parameter ac:name="type">flat</ac:parameter>
  <ac:parameter ac:name="separator">pipe</ac:parameter>
</ac:structured-macro>
```

## Development

This project should use a little bit more test data, gnome sayin?

## V2 Release Notes

- Migrating to [Confluence Cloud REST API V2](https://blog.developer.atlassian.com/the-confluence-cloud-rest-api-v2-brings-major-performance-improvements/)
- Atlassian [removed public access to V1 endpoints January 1st 2024](https://community.developer.atlassian.com/t/deprecating-many-confluence-v1-apis-that-have-v2-equivalents/66883)

refactor

- [x] upload images
- [x] support --parent page
- [ ] imports remote page IDs to local database
- [ ] update readme (access token, download instructions, usage instructions)
- [ ] remove duplicate headers when --title option is exercised

features

- [x] set space homepage if root document provided
- [x] support deleting remote pages when source is deleted #14
- [ ] support linked documents #19
- [ ] support for mermaid #71

fixes

- [x] handle spaces in parent page names #37
- [x] old golang version in Dockerfile #68
- [x] handle error response and fail fast #7
- [ ] handle invalid URLs #59

stretch-features

- [ ] support of action item (task) #56
- [ ] write pages with wide mode #49
- [ ] support custom page headers/footers #13
- [ ] Add option to set labels #15 via https://github.com/yuin/goldmark-meta
