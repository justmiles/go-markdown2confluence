package lib

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/justmiles/go-confluence"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"

	e "github.com/justmiles/go-markdown2confluence/lib/extension"
)

const (
	// DefaultEndpoint provides an example endpoint for users
	DefaultEndpoint = "https://mydomain.atlassian.net/wiki"

	// Parallelism determines how many files to convert and upload at a time
	Parallelism = 5
)

// Markdown2Confluence stores the settings for each run
type Markdown2Confluence struct {
	Space               string
	Comment             string
	Title               string
	File                string
	Ancestor            string
	Debug               bool
	UseDocumentTitle    bool
	FollowLinks         bool
	WithHardWraps       bool
	Since               int
	Username            string
	Password            string
	Endpoint            string
	Parent              string
	SourceMarkdown      []string
	ExcludeFilePatterns []string
	client              *confluence.Client
	url                 string
}

// CreateClient returns a new markdown clietn
func (m *Markdown2Confluence) CreateClient() {
	m.client = new(confluence.Client)
	m.client.Username = m.Username
	m.client.Password = m.Password
	m.client.Endpoint = m.Endpoint
	m.client.Debug = m.Debug
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

func (m *Markdown2Confluence) IsExcluded(p string) bool {
	for _, pattern := range m.ExcludeFilePatterns {
		r := regexp.MustCompile(pattern)
		if r.MatchString(p) {
			fmt.Printf("excluding markdown file '%s': exclude pattern '%s'\n", p, pattern)
			return true
		}
	}

	return false
}

// Run the sync
func (m *Markdown2Confluence) Run() []error {
	var markdownFiles []MarkdownFile
	var now = time.Now()
	m.CreateClient()

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

					if strings.HasSuffix(path, ".md") && !m.IsExcluded(path) {

						// Only include this file if it was modified m.Since minutes ago
						if m.Since != 0 {
							if info.ModTime().Unix() < now.Add(time.Duration(m.Since*-1)*time.Minute).Unix() {
								if m.Debug {
									fmt.Printf("skipping %s: last modified %s\n", info.Name(), info.ModTime())
								}
								return nil
							}
						}

						var tempTitle string
						var tempParents []string

						if strings.HasSuffix(path, "README.md") {
							tempTitle = strings.Split(path, "/")[len(strings.Split(path, "/"))-2]
							tempParents = deleteFromSlice(deleteFromSlice(strings.Split(filepath.Dir(strings.TrimPrefix(filepath.ToSlash(path), filepath.ToSlash(f))), "/"), "."), tempTitle)
						} else {
							tempTitle = strings.TrimSuffix(filepath.Base(path), ".md")
							tempParents = deleteFromSlice(strings.Split(filepath.Dir(strings.TrimPrefix(filepath.ToSlash(path), filepath.ToSlash(f))), "/"), ".")
						}

						if m.UseDocumentTitle == true {
							docTitle := getDocumentTitle(path)
							if docTitle != "" {
								tempTitle = docTitle
							}
						}

						md = MarkdownFile{
							Path:    path,
							Parents: tempParents,
							Title:   tempTitle,
						}

						if m.Parent != "" {
							parents := strings.Split(m.Parent, "/")
							md.Parents = append(parents, md.Parents...)
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
				if m.UseDocumentTitle == true {
					md.Title = getDocumentTitle(f)
				}
				if md.Title == "" {
					md.Title = strings.TrimSuffix(filepath.Base(f), ".md")
				}
			}
			if m.Parent != "" {
				parents := strings.Split(m.Parent, "/")
				md.Parents = append(parents, md.Parents...)
				md.Parents = deleteEmpty(md.Parents)
			}
			markdownFiles = append(markdownFiles, md)
		}
	}

	if m.FollowLinks == true {
		fmt.Print("Detecting linked markdown documents...\n")
		linkedDocuments := map[string]MarkdownFile{}

		// initialize already captures files
		for _, markdownFile := range markdownFiles {
			linkedDocuments[markdownFile.Path] = markdownFile
		}

		for _, markdownFile := range markdownFiles {
			getLinkedDocuments(markdownFile, m, linkedDocuments)
		}

		// append all md objects to upload list
		markdownFiles = []MarkdownFile{}
		for _, md := range linkedDocuments {
			markdownFiles = append(markdownFiles, md)
		}
	}

	// upload markdown files to confluence
	fmt.Print("Uploading markdown files...")
	errors, urls := m.upload(markdownFiles)

	if m.FollowLinks == true {
		fmt.Print("Patching relative paths in markdown documents...\n")
		// patch markdownfile with retrieved urls and upload again
		var markdownFilesPatched []MarkdownFile
		for _, md := range markdownFiles {
			mdp, patched := replaceRelativeLinks(md, urls)
			if patched {
				if m.Debug {
					fmt.Printf("File %s: patched successfully with Confluence links\n", md.Path)
				}
				markdownFilesPatched = append(markdownFilesPatched, mdp)
			}
		}
		errors2, _ := m.upload(markdownFilesPatched)
		errors = append(errors, errors2...)

		// clean up temporary files
		for _, md := range markdownFilesPatched {
			if m.Debug {
				fmt.Printf("Removing tempfile %s\n", md.Path)
			}
			defer os.Remove(md.Path)
		}
	}

	return errors
}

func (m *Markdown2Confluence) upload(markdownFiles []MarkdownFile) ([]error, map[string]string) {
	var (
		wg    = sync.WaitGroup{}
		queue = make(chan MarkdownFile)
	)

	var errors []error
	urls := map[string]string{}

	// Process the queue
	for worker := 0; worker < Parallelism; worker++ {
		wg.Add(1)
		go m.queueProcessor(&wg, &queue, &errors, urls)
	}

	for _, markdownFile := range markdownFiles {
		// Create parent pages synchronously
		if len(markdownFile.Parents) > 0 {
			var err error
			markdownFile.Ancestor, err = markdownFile.FindOrCreateAncestors(m)
			if err != nil {
				errors = append(errors, err)
				continue
			}
		}
		queue <- markdownFile
	}
	close(queue)
	wg.Wait()

	return errors, urls
}

func (m *Markdown2Confluence) queueProcessor(wg *sync.WaitGroup, queue *chan MarkdownFile, errors *[]error, urls map[string]string) {
	defer wg.Done()

	for markdownFile := range *queue {
		url, err := markdownFile.Upload(m)
		if err != nil {
			*errors = append(*errors, fmt.Errorf("Unable to upload markdown file %s: \n\t%s", markdownFile.Path, err))
		}
		urls[markdownFile.Path] = url
		fmt.Printf("--> %s: %s\n", markdownFile.FormattedPath(), url)
	}
}

func validateInput(s string, msg string) {
	if s == "" {
		fmt.Println(msg)
		os.Exit(1)
	}
}

func renderContent(filePath, s string, withHardWraps bool) (content string, images []string, err error) {
	confluenceExtension := e.NewConfluenceExtension(filePath)
	ro := goldmark.WithRendererOptions(
		html.WithXHTML(),
	)
	if withHardWraps {
		ro = goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		)
	}
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.DefinitionList),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		ro,
		goldmark.WithExtensions(
			confluenceExtension,
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(s), &buf); err != nil {
		return "", nil, err
	}

	return buf.String(), confluenceExtension.Images(), nil
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

func getDocumentTitle(p string) string {
	// Read file to check for the content
	fileContent, err := ioutil.ReadFile(p)
	if err != nil {
		log.Fatal(err)
	}
	// Convert []byte to string and print to screen
	text := string(fileContent)

	// check if there is a
	e := `^#\s+(.+)`
	r := regexp.MustCompile(e)
	result := r.FindStringSubmatch(text)
	if len(result) > 1 {
		// assign the Title to the matching group
		return result[1]
	}

	return ""
}

func getLinkedDocuments(p MarkdownFile, m *Markdown2Confluence, docs map[string]MarkdownFile) map[string]MarkdownFile {
	// Read file to check for the content
	fileRoot, _ := filepath.Abs(filepath.Dir(p.Path))
	fileContent, err := ioutil.ReadFile(p.Path)
	if err != nil {
		log.Fatal(err)
	}
	// Convert []byte to string and print to screen
	text := string(fileContent)

	// get all links to md files
	e1 := `\[.*\]\((.*\.md)\)`
	e2 := `\[.*\]:(.*)$`
	r1 := regexp.MustCompile(e1)
	r2 := regexp.MustCompile(e2)
	matches1 := r1.FindAllStringSubmatch(text, -1)
	matches2 := r2.FindAllStringSubmatch(text, -1)

	matches := append(matches1, matches2...)

	for i := range matches {
		mdFile := filepath.Join(fileRoot, strings.TrimSpace(matches[i][1]))
		// check if file is already captured
		if _, ok := docs[mdFile]; ok {
			// file already captured, continue
			continue
		}

		if _, err := os.Stat(mdFile); err == nil {
			if m.Debug {
				fmt.Printf("Found linked file '%s' in '%s'\n", mdFile, p.Path)
			}

			// get relative path to root file
			dirReferencedFile, _ := filepath.Abs(filepath.Dir(mdFile))
			dirBaseFile, _ := filepath.Abs(filepath.Dir(p.Path))

			// capture out of tree files
			var parents []string
			if len(dirReferencedFile) >= len(dirBaseFile) {
				relPath := strings.Replace(dirReferencedFile, dirBaseFile, "", -1)
				relPathComponents := deleteEmpty(strings.Split(filepath.ToSlash(relPath), "/"))
				parents = append(p.Parents, relPathComponents...)
			} else {
				relPath := strings.Replace(dirBaseFile, dirReferencedFile, "", -1)
				relPathComponents := deleteEmpty(strings.Split(filepath.ToSlash(relPath), "/"))

				// make sure we do not run out of tree
				if len(relPathComponents) > len(p.Parents) {
					fmt.Printf("WARNING: Referenced file '%s' cannot be caputered by parents '%s'. Skipping\n", mdFile, strings.Join(p.Parents[:], "/"))
					continue
				}

				parents = p.Parents[:len(p.Parents)-len(relPathComponents)]
			}

			// md file exists exists
			md := MarkdownFile{
				Path:    mdFile,
				Parents: parents,
				Title:   "",
			}

			if m.UseDocumentTitle == true {
				md.Title = getDocumentTitle(mdFile)
			}
			if md.Title == "" {
				md.Title = strings.TrimSuffix(filepath.Base(mdFile), ".md")
			}

			docs[mdFile] = md
			docs = getLinkedDocuments(docs[mdFile], m, docs)
		}
	}

	return docs
}

func replaceRelativeLinks(p MarkdownFile, urls map[string]string) (MarkdownFile, bool) {
	// Read file to check for the content
	fileRoot, _ := filepath.Abs(filepath.Dir(p.Path))
	fileContent, err := ioutil.ReadFile(p.Path)
	if err != nil {
		log.Fatal(err)
	}
	// Convert []byte to string and print to screen
	textOriginal := string(fileContent)
	textPatched := string(fileContent)

	// get all links to md files
	e1 := `\[.*\]\((.*\.md)\)`
	e2 := `\[.*\]:(.*)$`
	r1 := regexp.MustCompile(e1)
	r2 := regexp.MustCompile(e2)
	matches1 := r1.FindAllStringSubmatch(textOriginal, -1)
	matches2 := r2.FindAllStringSubmatch(textOriginal, -1)

	matches := append(matches1, matches2...)

	for i := range matches {
		mdFile := filepath.Join(fileRoot, strings.TrimSpace(matches[i][1]))
		// check if file is already captured
		if _, ok := urls[mdFile]; ok {
			// found confluence page
			textPatched = strings.ReplaceAll(textPatched, matches[i][1], urls[mdFile])
		}
	}

	// check if there were any changes
	patched := (textOriginal != textPatched)
	mdFile := p.Path
	if patched {
		file, err := ioutil.TempFile("", "md")
		if err != nil {
			log.Fatal(err)
		}
		_, err2 := file.WriteString(textPatched)
		if err2 != nil {
			log.Fatal(err)
		}
		mdFile = file.Name()
	}

	md := p
	md.Path = mdFile

	return md, patched
}

func isRelative(path string) bool {
	return false
}
