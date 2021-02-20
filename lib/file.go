package lib

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/justmiles/go-confluence"
)

// MarkdownFile contains information about the file to upload
type MarkdownFile struct {
	Path     string
	Title    string
	Parents  []string
	Ancestor string
	URL      string
}

func (f *MarkdownFile) String() (urlPath string) {
	return fmt.Sprintf("Path: %s, Title: %s, Parent: %s, Ancestor: %s", f.Path, f.Title, f.Parents, f.Ancestor)
}

// FormattedPath returns the Path with Parents
func (f *MarkdownFile) FormattedPath() (s string) {
	s = strings.Join(append(f.Parents, f.Title), "/")
	s = strings.TrimPrefix(s, "/")
	s = strings.TrimPrefix(s, "/")
	return s
}

// Upload a markdown file
func (f *MarkdownFile) Upload(m *Markdown2Confluence) (urlPath string, err error) {
	var ancestorID string
	// Content of Wiki
	dat, err := ioutil.ReadFile(f.Path)
	if err != nil {
		return urlPath, fmt.Errorf("Could not open file %s:\n\t%s", f.Path, err)
	}

	if m.Debug {
		fmt.Println(f.Path)
	}

	wikiContent := string(dat)
	var images []string
	wikiContent, images, err = renderContent(f.Path, wikiContent, m.WithHardWraps)

	if err != nil {
		return urlPath, fmt.Errorf("unable to render content from %s: %s", f.Path, err)
	}

	if m.Debug {
		fmt.Println("---- RENDERED CONTENT START ---------------------------------")
		fmt.Println(wikiContent)
		fmt.Println("---- RENDERED CONTENT END -----------------------------------")

		for _, image := range images {
			fmt.Printf("LOCAL IMAGE FOUND: %s\n", image)
		}
	}

	// search for existing page
	contentResults, err := m.client.GetContent(&confluence.GetContentQueryParameters{
		Title:    f.Title,
		Spacekey: m.Space,
		Limit:    1,
		Type:     "page",
		Expand:   []string{"version", "body.storage"},
	})
	if err != nil {
		return urlPath, fmt.Errorf("Error checking for existing page: %s", err)
	}

	if len(f.Parents) > 0 {
		ancestorID, err = f.FindOrCreateAncestors(m)
		if err != nil {
			return urlPath, err
		}
	}

	var content confluence.Content
	var currContentID string
	// if page exists, update it
	if len(contentResults) > 0 {
		content = contentResults[0]
		content.Version.Number++
		content.Version.Message = m.Comment
		content.Body.Storage.Representation = "storage"
		content.Body.Storage.Value = wikiContent
		content.Space.Key = m.Space
		if ancestorID != "" {
			content.Ancestors = append(content.Ancestors, Ancestor{
				ID: ancestorID,
			})
		}

		content, err = m.client.UpdateContent(&content, nil)
		if err != nil {
			return urlPath, fmt.Errorf("Error updating content: %s", err)
		}
		urlPath = m.client.Endpoint + content.Links.Tinyui
		currContentID = content.ID

		// if page does not exist, create it
	} else {

		bp := confluence.CreateContentBodyParameters{}
		bp.Title = f.Title
		bp.Type = "page"
		bp.Space.Key = m.Space
		bp.Body.Storage.Representation = "storage"
		bp.Body.Storage.Value = wikiContent

		if ancestorID != "" {
			bp.Ancestors = append(bp.Ancestors, Ancestor{
				ID: ancestorID,
			})
		}

		content, err := m.client.CreateContent(&bp, nil)
		if err != nil {
			return urlPath, fmt.Errorf("Error creating page: %s", err)
		}
		urlPath = m.client.Endpoint + content.Links.Tinyui
		currContentID = content.ID
	}

	_, errors := m.client.AddUpdateAttachments(currContentID, images)
	if len(errors) > 0 {
		fmt.Println(errors)
		err = errors[0]
	}

	return urlPath, err
}

// FindOrCreateAncestors creates an empty page to represent a local "folder" name
func (f *MarkdownFile) FindOrCreateAncestors(m *Markdown2Confluence) (ancestorID string, err error) {

	for _, parent := range f.Parents {
		ancestorID, err = f.FindOrCreateAncestor(m, m.client, ancestorID, parent)
		if err != nil {
			return "", err
		}
	}

	// Return the last ancestorID
	return ancestorID, nil
}

// ParentIndex caches parent page Ids for futures reference
var ParentIndex = make(map[string]string)

// FindOrCreateAncestor creates an empty page to represent a local "folder" name
func (f *MarkdownFile) FindOrCreateAncestor(m *Markdown2Confluence, client *confluence.Client, ancestorID, parent string) (string, error) {
	if parent == "" {
		return "", nil
	}

	if val, ok := ParentIndex[parent]; ok {
		return val, nil
	}

	if m.Debug {
		fmt.Printf("Searching for parent %s\n", parent)
	}

	contentResults, err := client.GetContent(&confluence.GetContentQueryParameters{
		Title:    parent,
		Spacekey: m.Space,
		Limit:    1,
		Type:     "page",
	})
	if err != nil {
		return "", fmt.Errorf("Error checking for parent page: %s", err)
	}

	if len(contentResults) > 0 {
		content := contentResults[0]
		ParentIndex[parent] = content.ID
		return content.ID, err
	}

	// if parent page does not exist, create it
	bp := confluence.CreateContentBodyParameters{}
	bp.Title = parent
	bp.Type = "page"
	bp.Space.Key = m.Space
	bp.Body.Storage.Representation = "storage"
	bp.Body.Storage.Value = defaultAncestorPage

	if m.Debug {
		fmt.Printf("Creating parent page '%s' with ancestor id %s\n", bp.Title, ancestorID)
	}

	if ancestorID != "" {
		bp.Ancestors = append(bp.Ancestors, Ancestor{
			ID: ancestorID,
		})
	}

	content, err := client.CreateContent(&bp, nil)
	if err != nil {
		return "", fmt.Errorf("Error creating parent page %s for %s: %s", f.Path, bp.Title, err)
	}
	ParentIndex[parent] = content.ID
	return content.ID, nil
}

// Ancestor TODO: move this to go-confluence api
type Ancestor struct {
	ID string `json:"id,omitempty"`
}

const defaultAncestorPage = `
<p>
   <ac:structured-macro ac:name="children" ac:schema-version="2" ac:macro-id="a93cdc19-61cd-4c21-8da7-0af3c6b76c07">
      <ac:parameter ac:name="all">true</ac:parameter>
      <ac:parameter ac:name="sort">title</ac:parameter>
   </ac:structured-macro>
</p>
`
