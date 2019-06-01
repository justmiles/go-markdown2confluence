package markdown2confluence

import (
	"fmt"
	"io/ioutil"

	"github.com/justmiles/go-confluence"
)

// MarkdownFile contains information about the file to upload
type MarkdownFile struct {
	Path   string
	Title  string
	Parent string
}

// Upload a markdown file
func (f *MarkdownFile) Upload(m *Markdown2Confluence) (url string, err error) {

	// Content of Wiki
	dat, err := ioutil.ReadFile(f.Path)
	if err != nil {
		return url, fmt.Errorf("Could not open file %s:\n\t%s", f.Path, err)
	}

	wikiContent := string(dat)
	wikiContent = renderContent(wikiContent)

	if m.Debug {
		fmt.Println("---- RENDERED CONTENT START ---------------------------------")
		fmt.Println(wikiContent)
		fmt.Println("---- RENDERED CONTENT END -----------------------------------")
	}

	// Create the Confluence client
	client := new(confluence.Client)
	client.Username = m.Username
	client.Password = m.Password
	client.Endpoint = m.Endpoint
	client.Debug = m.Debug

	// search for existing page
	contentResults, err := client.GetContent(&confluence.GetContentQueryParameters{
		Title:    f.Title,
		Spacekey: m.Space,
		Limit:    1,
		Type:     "page",
		Expand:   []string{"version", "body.storage"},
	})
	if err != nil {
		return url, fmt.Errorf("Error checking for existing page: %s", err)
	}

	// if page exists, update it
	if len(contentResults) > 0 {
		content := contentResults[0]
		content.Version.Number++
		content.Body.Storage.Representation = "storage"
		content.Body.Storage.Value = wikiContent
		content, err = client.UpdateContent(&content, nil)
		if err != nil {
			return url, fmt.Errorf("Error updating content: %s", err)
		}
		url = client.Endpoint + content.Links.Tinyui

		// if page does not exist, create it
	} else {
		bp := confluence.CreateContentBodyParameters{}
		bp.Title = f.Title
		bp.Type = "page"
		bp.Space.Key = m.Space
		bp.Body.Storage.Representation = "storage"
		bp.Body.Storage.Value = wikiContent
		content, err := client.CreateContent(&bp, nil)
		if err != nil {
			return url, fmt.Errorf("Error creating page: %s", err)
		}
		url = client.Endpoint + content.Links.Tinyui
	}

	return url, nil
}
