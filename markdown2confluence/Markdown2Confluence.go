package markdown2confluence

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/justmiles/mark"
)

const (
	// DefaultEndpoint provides an example endpoint for users
	DefaultEndpoint = "https://mydomain.atlassian.net/wiki"
)

// Markdown2Confluence stores the settings for each run
type Markdown2Confluence struct {
	Space          string
	Title          string
	File           string
	Debug          bool
	Username       string
	Password       string
	Endpoint       string
	SourceMarkdown []string
}

// SourceEnvironmentVariables overrides Markdown2Confluence with any environment variables that are set
//  - CONFLUENCE_USERNAME
//  - CONFLUENCE_PASSWORD
//  - CONFLUENCE_ENDPOINT
func (m *Markdown2Confluence) SourceEnvironmentVariables() {
	var s string
	s = os.Getenv("CONFLUENCE_USERNAME")
	if s != "" {
		m.Username = s
	}

	s = os.Getenv("CONFLUENCE_PASSWORD")
	if s != "" {
		m.Password = s
	}

	s = os.Getenv("CONFLUENCE_ENDPOINT")
	if s != "" {
		m.Endpoint = s
	}
}

// Validate required configs are set
func (m Markdown2Confluence) Validate() error {
	if m.Space == "" {
		return fmt.Errorf("Space is not defined")
	}
	if m.Username == "" {
		return fmt.Errorf("Username is not defined")
	}
	if m.Password == "" {
		return fmt.Errorf("Password is not defined")
	}
	if m.Endpoint == "" {
		return fmt.Errorf("Endpoint is not defined")
	}
	if m.Endpoint == DefaultEndpoint {
		return fmt.Errorf("Endpoint is not defined")
	}
	if len(m.SourceMarkdown) == 0 {
		return fmt.Errorf("No markdown to upload")
	}
	return nil
}

// Run the sync
func (m *Markdown2Confluence) Run() error {
	var markdownFiles []MarkdownFile

	for _, f := range m.SourceMarkdown {
		file, err := os.Open(f)
		defer file.Close()
		if err != nil {
			return fmt.Errorf("Error opening file %s", err)
		}

		stat, err := file.Stat()
		if err != nil {
			return fmt.Errorf("Error reading file meta %s", err)
		}
		if stat.IsDir() {
			err := filepath.Walk(f,
				func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if strings.HasSuffix(path, ".md") {
						markdownFiles = append(markdownFiles, MarkdownFile{
							Path:   path,
							Parent: filepath.Base(filepath.Dir(path)),
							Title:  strings.TrimSuffix(filepath.Base(path), ".md"),
						})
					}
					return nil
				})
			if err != nil {
				fmt.Println(err)
			}

		} else {
			markdownFiles = append(markdownFiles, MarkdownFile{
				Path:  f,
				Title: strings.TrimSuffix(filepath.Base(f), ".md"),
			})
		}
	}

	for _, markdownFile := range markdownFiles {
		url, err := markdownFile.Upload(m)
		if err != nil {
			return fmt.Errorf("Unable to upload markdown file, %s: \n\t%s", markdownFile.Path, err)
		}
		fmt.Printf("%s: %s \n", markdownFile.Title, url)
	}
	return nil
}

func validateInput(s string, msg string) {
	if s == "" {
		fmt.Println(msg)
		os.Exit(1)
	}
}

func renderContent(s string) string {
	m := mark.New(s, nil)
	return m.Render()
}
