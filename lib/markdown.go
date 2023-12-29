package lib

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/boltdb/bolt"
	"github.com/justmiles/go-markdown2confluence/lib/confluence"

	log "github.com/sirupsen/logrus"
)

const (
	// DefaultEndpoint provides an example endpoint for users
	DefaultEndpoint = "https://mydomain.atlassian.net/wiki"

	// Parallelism determines how many files to convert and upload at a time
	// TODO: fix race condition against m.files map
	Parallelism = 1
)

// Markdown2Confluence stores the settings for each run
type Markdown2Confluence struct {
	confluence.Client

	Space               string
	SpaceID             string
	Comment             string
	Title               string
	LocalStore          string
	UseDocumentTitle    bool
	ForceUpdates        bool
	WithHardWraps       bool
	APIToken            string
	Parent              string
	SourceMarkdown      []string
	ExcludeFilePatterns []string

	excludeFilePatterns []*regexp.Regexp
	files               map[string]*MarkdownFile
	db                  *bolt.DB
}

func (m *Markdown2Confluence) Init() error {

	level, err := log.ParseLevel(m.LogLevel)
	if err != nil {
		return err
	}
	log.SetLevel(level)

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

	m.db, err = bolt.Open(m.LocalStore, 0600, nil)
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

	if !isUrl(m.Endpoint) {
		return fmt.Errorf("--endpoint is not a valid URL: %s", m.Endpoint)
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
	return true
}

func (m *Markdown2Confluence) PurgeSpace() error {

	err := m.DeleteAllPagesInSpace(m.SpaceID)
	if err != nil {
		return err
	}
	return nil
}

// Import imports remote pages to local database
func (m *Markdown2Confluence) Import() error {
	// TODO: implement this!
	return fmt.Errorf("this feature not yet implemented")
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

		if page != nil && markdownFile.Status == "SYNCED" {
			fmt.Printf("%s%s - %s\n", m.Endpoint, page.Links.Tinyui, markdownFile.Path)
			markdownFile.Logger().Infof("%s%s", m.Endpoint, page.Links.Tinyui)
		}
	}
}
