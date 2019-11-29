package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/justmiles/mark"

	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"

	r "github.com/justmiles/go-markdown2confluence/lib/renderer"
)

const (
	// DefaultEndpoint provides an example endpoint for users
	DefaultEndpoint = "https://mydomain.atlassian.net/wiki"

	// Parallelism determines how many files to convert and upload at a time
	Parallelism = 5
)

// Markdown2Confluence stores the settings for each run
type Markdown2Confluence struct {
	Space          string
	Title          string
	File           string
	Ancestor       string
	Debug          bool
	Since          int
	Username       string
	Password       string
	Endpoint       string
	Parent         string
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
		return fmt.Errorf("--space is not defined")
	}
	if m.Username == "" {
		return fmt.Errorf("--username is not defined")
	}
	if m.Password == "" {
		return fmt.Errorf("--password is not defined")
	}
	if m.Endpoint == "" {
		return fmt.Errorf("--endpoint is not defined")
	}
	if m.Endpoint == DefaultEndpoint {
		return fmt.Errorf("--endpoint is not defined")
	}
	if len(m.SourceMarkdown) == 0 {
		return fmt.Errorf("please pass a markdown file or directory of markdown files")
	}
	if len(m.SourceMarkdown) > 1 && m.Title != "" {
		return fmt.Errorf("You can not set the title for multiple files")
	}
	return nil
}

// Run the sync
func (m *Markdown2Confluence) Run() []error {
	var markdownFiles []MarkdownFile
	var now = time.Now()

	for _, f := range m.SourceMarkdown {
		file, err := os.Open(f)
		defer file.Close()
		if err != nil {
			return []error{fmt.Errorf("Error opening file %s", err)}
		}

		stat, err := file.Stat()
		if err != nil {
			return []error{fmt.Errorf("Error reading file meta %s", err)}
		}

		var md MarkdownFile

		if stat.IsDir() {

			// prevent someone from accidently uploading everything under the same title
			if m.Title != "" {
				return []error{fmt.Errorf("--title not supported for directories")}
			}

			err := filepath.Walk(f,
				func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}

					if strings.HasSuffix(path, ".md") {

						// Only include this file if it was modified m.Since minutes ago
						if m.Since != 0 {
							if info.ModTime().Unix() < now.Add(time.Duration(m.Since*-1)*time.Minute).Unix() {
								if m.Debug {
									fmt.Printf("skipping %s: last modified %s\n", info.Name(), info.ModTime())
								}
								return nil
							}
						}

						md := MarkdownFile{
							Path:    path,
							Parents: deleteFromSlice(strings.Split(filepath.Dir(strings.TrimPrefix(filepath.ToSlash(path), filepath.ToSlash(f))), "/"), "."),
							Title:   strings.TrimSuffix(filepath.Base(path), ".md"),
						}

						if m.Parent != "" {
							md.Parents = append([]string{m.Parent}, md.Parents...)
							md.Parents = deleteEmpty(md.Parents)
						}

						markdownFiles = append(markdownFiles, md)

					}
					return nil
				})
			if err != nil {
				return []error{fmt.Errorf("Unable to walk path: %s", f)}
			}

		} else {
			md = MarkdownFile{
				Path:  f,
				Title: m.Title,
			}

			if md.Title == "" {
				md.Title = strings.TrimSuffix(filepath.Base(f), ".md")
			}

			if m.Parent != "" {
				md.Parents = append([]string{m.Parent}, md.Parents...)
				md.Parents = deleteEmpty(md.Parents)
			}

			markdownFiles = append(markdownFiles, md)
		}

	}

	var (
		wg    = sync.WaitGroup{}
		queue = make(chan MarkdownFile)
	)

	var errors []error

	// Process the queue
	for worker := 0; worker < Parallelism; worker++ {
		wg.Add(1)
		go m.queueProcessor(&wg, &queue, &errors)
	}

	for _, markdownFile := range markdownFiles {
		queue <- markdownFile
	}

	close(queue)

	wg.Wait()

	return errors
}

func (m *Markdown2Confluence) queueProcessor(wg *sync.WaitGroup, queue *chan MarkdownFile, errors *[]error) {
	defer wg.Done()

	for markdownFile := range *queue {
		url, err := markdownFile.Upload(m)
		if err != nil {
			*errors = append(*errors, fmt.Errorf("Unable to upload markdown file, %s: \n\t%s", markdownFile.Path, err))
		}
		fmt.Printf("%s: %s\n", markdownFile.FormattedPath(), url)
	}
}

func validateInput(s string, msg string) {
	if s == "" {
		fmt.Println(msg)
		os.Exit(1)
	}
}

func renderContent(s string) (string, error) {

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.DefinitionList),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
		goldmark.WithRendererOptions(renderer.WithNodeRenderers(
			util.Prioritized(r.NewConfluenceFencedCodeBlockHTMLRender(), 100),
			util.Prioritized(r.NewConfluenceCodeBlockHTMLRender(), 100),
		)),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(s), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func renderContentLegacy(s string) string {
	m := mark.New(s, nil)
	return m.Render()
}

func deleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func deleteFromSlice(s []string, del string) []string {
	for i, v := range s {
		if v == del {
			s = append(s[:i], s[i+1:]...)
			break
		}
	}
	return s
}
