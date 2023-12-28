package lib

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/justmiles/go-markdown2confluence/lib/confluence"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	renderer "github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"

	e "github.com/justmiles/go-markdown2confluence/lib/extension"
	log "github.com/sirupsen/logrus"
)

const (
	// DefaultEndpoint provides an example endpoint for users
	DefaultEndpoint = "https://mydomain.atlassian.net/wiki"

	// Parallelism determines how many files to convert and upload at a time
	// TODO: fix race condition against m.files map
	Parallelism = 1
)

var now = time.Now()

// Markdown2Confluence stores the settings for each run
type Markdown2Confluence struct {
	confluence.Client

	Space               string
	SpaceID             string
	Comment             string
	Title               string
	File                string
	UseDocumentTitle    bool
	ForceUpdates        bool
	WithHardWraps       bool
	Since               int
	APIToken            string
	Parent              string
	SourceMarkdown      []string
	ExcludeFilePatterns []string

	excludeFilePatterns []*regexp.Regexp
	files               map[string]*MarkdownFile
	db                  *bolt.DB
}

func (m *Markdown2Confluence) Init() error {

	// get the space ID
	space, err := m.GetSpaceByKey(m.Space)
	if err != nil {
		return err
	}

	m.SpaceID = space.ID

	// get the parent page ID
	if m.Parent != "" {
		getPagesInSpaceResponse, err := m.GetPagesInSpace(m.SpaceID, &confluence.GetPagesInSpaceQueryParameters{
			Title: m.Parent,
		})
		if err != nil {
			return fmt.Errorf("could not retrieve parent page: %s", err)
		}

		if len(getPagesInSpaceResponse.Results) == 0 {
			return fmt.Errorf("no pages match provided --parent in this space")
		}

		if len(getPagesInSpaceResponse.Results) > 1 {
			return fmt.Errorf("--parent filter matched more than one page")
		}

		m.Parent = getPagesInSpaceResponse.Results[0].ID
	}

	m.db, err = bolt.Open("confluence.db", 0600, nil)
	if err != nil {
		return err
	}

	// init the db fbucket
	err = m.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("pages"))
		return err
	})
	if err != nil {
		return err
	}

	m.files = make(map[string]*MarkdownFile)

	// pre-compile regex patterns
	for _, pattern := range m.ExcludeFilePatterns {
		r := regexp.MustCompile(pattern)
		m.excludeFilePatterns = append(m.excludeFilePatterns, r)
	}

	return nil
}

func (m *Markdown2Confluence) Close() func() error {
	return m.db.Close
}

// Validate required configs are set
func (m Markdown2Confluence) Validate() error {
	if m.Space == "" {
		return fmt.Errorf("--space is not defined")
	}
	if m.Username == "" && m.AccessToken == "" {
		return fmt.Errorf("--username is not defined")
	}
	if m.Password == "" && m.AccessToken == "" {
		return fmt.Errorf("--password is not defined")
	}
	if m.Endpoint == "" || m.Endpoint == DefaultEndpoint {
		return fmt.Errorf("--endpoint is not defined")
	}
	if len(m.SourceMarkdown) == 0 {
		return fmt.Errorf("please pass a markdown file or directory of markdown files")
	}
	if len(m.SourceMarkdown) > 1 && m.Title != "" {
		return fmt.Errorf("you can not set the title for multiple files")
	}
	if m.AccessToken == "" && m.Username == "" {
		return fmt.Errorf("--access-token is not defined")
	}
	return nil
}

func (m *Markdown2Confluence) IsExcluded(p string) bool {
	for _, r := range m.excludeFilePatterns {
		if r.MatchString(p) {
			log.Debugf(`excluding %s - matches pattern "%s"`, p, r.String())
			return true
		}
	}

	return false
}

func (m *Markdown2Confluence) IsIncluded(info os.FileInfo) bool {
	if !strings.HasSuffix(info.Name(), ".md") {
		return false
	}
	if m.IsExcluded(info.Name()) {
		return false
	}
	if m.Since != 0 {
		if info.ModTime().Unix() < now.Add(time.Duration(m.Since*-1)*time.Minute).Unix() {
			log.Debug(fmt.Sprintf("skipping %s: last modified %s\n", info.Name(), info.ModTime()))
			return false
		}
	}
	return true
}

func (m *Markdown2Confluence) GetSpaceID() error {

	space, err := m.GetSpaceByKey(m.Space)
	if err != nil {
		return err
	}

	m.SpaceID = space.ID

	return nil
}

func (m *Markdown2Confluence) PurgeSpace() error {

	err := m.DeleteAllPagesInSpace(m.SpaceID)
	if err != nil {
		return err
	}
	return nil
}

func (m *Markdown2Confluence) save() error {
	for _, file := range m.files {
		err := file.updateDB(m.db)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Markdown2Confluence) queueProcessor(wg *sync.WaitGroup, queue *chan *MarkdownFile, errors *[]error) {
	defer wg.Done()

	for markdownFile := range *queue {
		page, err := markdownFile.Upload(m)
		if err != nil {
			markdownFile.Logger().Error(err)
			markdownFile.Status = "ERRORED"
		}

		if page != nil {
			markdownFile.Logger().Infof("%s%s", m.Endpoint, page.Links.Tinyui)
		}
	}
}

// Import imports remote pages to local database
func (m *Markdown2Confluence) Import() error {
	// TODO: implement this!
	return fmt.Errorf("this feature not yet implemented")
}

func renderContent(filePath, s string, withHardWraps bool) (content string, images []string, err error) {
	confluenceExtension := e.NewConfluenceExtension(filePath)

	renderOptions := []renderer.Option{
		html.WithXHTML(),
	}

	if withHardWraps {
		renderOptions = append(renderOptions, html.WithHardWraps())
	}

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.DefinitionList),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(renderOptions...),
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

func getDocumentTitle(p string) string {
	// Read file to check for the content
	file_content, err := os.ReadFile(p)
	if err != nil {
		log.Fatal(err)
	}
	// Convert []byte to string and print to screen
	text := string(file_content)

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
