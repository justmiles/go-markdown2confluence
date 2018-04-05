package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/a8m/mark"
	"github.com/justmiles/go-confluence"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	spacePtr = flag.String("space", "", "space in which page should be created. Defaults to user's personal space")
	titlePtr = flag.String("title", "", "title for page. Defaults to file name without .md extension")
	filePtr  = flag.String("file", "", "markdown file to sync with Confluence")
	debugPtr = flag.Bool("debug", false, "enable debug logging")
	username = os.Getenv("CONFLUENCE_USERNAME")
	password = os.Getenv("CONFLUENCE_PASSWORD")
	endpoint = os.Getenv("CONFLUENCE_ENDPOINT")
)

func main() {
	flag.Parse()
	validateInput(username, "environment variable CONFLUENCE_USERNAME not set")
	validateInput(password, "environment variable CONFLUENCE_PASSWORD not set")
	validateInput(endpoint, "environment variable CONFLUENCE_ENDPOINT not set")
	validateInput(*filePtr, "no file provided")

	if *spacePtr == "" {
		*spacePtr = "~" + string(username)
	}

	if *titlePtr == "" {
		re := regexp.MustCompile(`.*[^.md]`)
		*titlePtr = re.FindString(filepath.Base(*filePtr))
		validateInput(*titlePtr, "title not provided")
	}

	// Content of Wiki
	dat, err := ioutil.ReadFile(*filePtr)
	check(err, fmt.Sprintf(`Could not open file "%s"`, *filePtr))
	wikiContent := string(dat)
	wikiContent = renderContent(wikiContent)

	// Create the Confluence client
	client := new(confluence.Client)
	client.Username = username
	client.Password = password
	client.Endpoint = endpoint
	client.Debug = *debugPtr

	// search for existing page
	contentResults, err := client.GetContent(&confluence.GetContentQueryParameters{
		Title:    *titlePtr,
		Spacekey: *spacePtr,
		Limit:    1,
		Type:     "page",
		Expand:   []string{"version", "body.storage"},
	})
	check(err, "")

	// if page exists, update it
	if len(contentResults) > 0 {
		content := contentResults[0]
		content.Version.Number++
		content.Body.Storage.Representation = "storage"
		content.Body.Storage.Value = wikiContent
		content, err = client.UpdateContent(&content, nil)
		check(err, "")
		fmt.Println(client.Endpoint + content.Links.Tinyui)

		// if page does not exist, create it
	} else {
		bp := confluence.CreateContentBodyParameters{}
		bp.Title = *titlePtr
		bp.Type = "page"
		bp.Space.Key = *spacePtr
		bp.Body.Storage.Representation = "storage"
		bp.Body.Storage.Value = wikiContent
		content, err := client.CreateContent(&bp, nil)
		check(err, "")
		fmt.Println(client.Endpoint + content.Links.Tinyui)
	}
}

func validateInput(s string, msg string) {
	if s == "" {
		fmt.Println(msg)
		os.Exit(1)
	}
}
func check(e error, s string) {
	if e != nil {
		if s != "" {
			fmt.Println(s)
		}
		fmt.Println(e)
		os.Exit(1)
	}
}

func renderContent(s string) string {
	m := mark.New(s, nil)
	m.AddRenderFn(mark.NodeCode, func(node mark.Node) (s string) {
		p, _ := node.(*mark.CodeNode)
		lineCount, _ := lineCounter(p.Text)
		s += `<ac:structured-macro ac:name="code" ac:schema-version="1" ac:macro-id="8f7842af-1a72-4902-89a1-2d4ed6d6c70d">`
		s += `<ac:parameter ac:name="theme">DJango</ac:parameter>`
		if lineCount > 5 {
			s += `<ac:parameter ac:name="linenumbers">true</ac:parameter>`
		}
		if lineCount > 30 {
			s += `<ac:parameter ac:name="collapse">true</ac:parameter>`
		}
		if p.Lang != "" {
			s += fmt.Sprintf(`<ac:parameter ac:name="language">%s</ac:parameter>`, p.Lang)
		}
		s += fmt.Sprintf(`<ac:plain-text-body><![CDATA[%s]]></ac:plain-text-body>`, p.Text)
		s += `</ac:structured-macro>`
		return s
	})
	return m.Render()
}

func lineCounter(s string) (int, error) {
	r := strings.NewReader(s)
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}
