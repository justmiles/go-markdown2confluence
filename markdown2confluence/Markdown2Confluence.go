package markdown2confluence

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/justmiles/mark"
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
func (m *Markdown2Confluence) Run() []error {
	var markdownFiles []MarkdownFile

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
		if stat.IsDir() {
			err := filepath.Walk(f,
				func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if strings.HasSuffix(path, ".md") {
						md := MarkdownFile{
							Path:    path,
							Parents: strings.Split(filepath.Dir(strings.TrimPrefix(filepath.ToSlash(path), filepath.ToSlash(f))), "/"),
							Title:   strings.TrimSuffix(filepath.Base(path), ".md"),
						}
						if m.Ancestor != "" {
							md.Parents = append([]string{m.Ancestor}, md.Parents...)
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
			md := MarkdownFile{
				Path:  f,
				Title: strings.TrimSuffix(filepath.Base(f), ".md"),
			}

			if m.Ancestor != "" {
				md.Parents = append([]string{m.Ancestor}, md.Parents...)
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
		fmt.Println(strings.TrimPrefix(fmt.Sprintf("%s - %s: %s", strings.TrimPrefix(strings.Join(markdownFile.Parents, "/"), "/"), markdownFile.Title, url), " - "))
	}
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

func deleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}
