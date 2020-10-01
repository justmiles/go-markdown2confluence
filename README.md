# markdown2confluence

Push markdown files to Confluence Cloud

## Installation

Download the [latest
release](https://github.com/justmiles/go-markdown2confluence/releases)
and add the binary in your local `PATH`

- Linux

  ```shell
  curl -LO https://github.com/justmiles/go-markdown2confluence/releases/download/v3.1.2/go-markdown2confluence_3.1.2_linux_x86_64.tar.gz
  
  sudo tar -xzvf go-markdown2confluence_3.1.2_linux_x86_64.tar.gz -C /usr/local/bin/ markdown2confluence
  ```

- OSX

  ```shell
  curl -LO https://github.com/justmiles/go-markdown2confluence/releases/download/v3.1.2/go-markdown2confluence_3.1.2_darwin_x86_64.tar.gz
  
  sudo tar -xzvf go-markdown2confluence_3.1.2_darwin_x86_64.tar.gz -C /usr/local/bin/ markdown2confluence
  ```

- Windows
  
  Download [the latest release](https://github.com/justmiles/go-markdown2confluence/releases/download/v3.1.2/go-markdown2confluence_3.1.2_windows_x86_64.tar.gz) and add to your system `PATH`

## Build using docker

You can build locally using docker.

### Preparation

```shell
docker-compose build make
```

### Building the dist-directory

```shell
docker-compose run make
```

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
        --parent string        Optional parent page to nest content under
    -p, --password string      Confluence password. (Alternatively set CONFLUENCE_PASSWORD environment variable)
    -s, --space string         Space in which page should be created
    -t, --title string         Set the page title on upload (defaults to filename without extension)
    -u, --username string      Confluence username. (Alternatively set CONFLUENCE_USERNAME environment variable)
    -w, --hardwraps            Render newlines as <br />
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

## Enhancements

It is possible to insert Confluence macros using fenced code blocks.
The "language" for this is `CONFLUENCE-MACRO`, exactly like that in all-caps.
Here is an example for a ToC macro using all headlines starting at Level 2:

```markdown
    # Title

    ```CONFLUENCE-MACRO
    name:toc
      minLevel:2
    ```

    ## Section 1
```

In general almost all macros should be possible.
The general syntax is:

```markdown
    ```CONFLUENCE-MACRO
    name:Name of Macro
    attribute:Value of Attribute
      parameter-name:Value of Parameter
      next-parameter:Value of Parameter
    ```
```

So a fully fledged macro could look like:

```markdown
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
```

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
