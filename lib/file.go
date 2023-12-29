package lib

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/boltdb/bolt"
	"github.com/justmiles/go-markdown2confluence/lib/confluence"
	log "github.com/sirupsen/logrus"
)

const defaultAncestorPage = `
<p>
   <ac:structured-macro ac:name="children" ac:schema-version="2" ac:macro-id="a93cdc19-61cd-4c21-8da7-0af3c6b76c07">
      <ac:parameter ac:name="all">true</ac:parameter>
      <ac:parameter ac:name="sort">title</ac:parameter>
   </ac:structured-macro>
</p>
`

// MarkdownFile contains information about the file to upload
type MarkdownFile struct {
	ID             string `json:"id"`
	Path           string `json:"path"`
	Title          string `json:"title"`
	Parent         string `json:"parent"`
	RemoteID       string `json:"remote_id"`
	RemoteParentID string `json:"parent_id"`
	Status         string `json:"status"`
	MD5Sum         string `json:"md5sum"`
	Link           string `json:"link:"`
}

func (f *MarkdownFile) String() (urlPath string) {
	return fmt.Sprintf("ID: %s, Title: %s, Parent: %s, Path: %s", f.ID, f.Title, f.Parent, f.Path)
}
func (f *MarkdownFile) Logger() *log.Entry {
	return log.WithFields(log.Fields{
		"Title":          f.Title,
		"ID":             f.ID,
		"ParentDir":      f.Parent,
		"RemoteID":       f.RemoteID,
		"RemoteParentID": f.RemoteParentID,
	})
}

// Upload a markdown file
func (mf *MarkdownFile) Upload(m *Markdown2Confluence) (*confluence.Page, error) {

	if mf.Status == "SYNCED" {
		return nil, nil
	}

	var err error

	if mf.Status == "DELETE" {
		mf.Logger().Debug("deleting page")
		err = m.DeletePage(mf.RemoteID, &confluence.DeletePageQueryParameters{})
		if err != nil {
			return nil, err
		}
		mf.Status = "DELETED"
		return nil, nil
	}

	var wikiContent = defaultAncestorPage
	var images []string

	if !isDir(mf.Path) {
		// Content of Wiki
		dat, err := os.ReadFile(mf.Path)
		if err != nil {
			return nil, fmt.Errorf("could not open file %s:\n\t%s", mf.Path, err)
		}

		wikiContent = string(dat)
		wikiContent, images, err = m.renderContent(mf.Path, wikiContent)
		if err != nil {
			return nil, fmt.Errorf("unable to render content from %s: %s", mf.Path, err)
		}
	}

	mf.Logger().Trace("---- RENDERED CONTENT START ---- ")
	log.Trace(wikiContent)
	mf.Logger().Trace("---- RENDERED CONTENT END ---- ")

	var page confluence.Page

	var createPageBody = &confluence.CreatePageBody{
		Title:          mf.Title,
		SpaceID:        m.SpaceID,
		Status:         "current",
		RemoteParentID: mf.RemoteParentID,
		Body: confluence.ContentBody{
			Representation: "storage",
			Value:          wikiContent,
		},
	}

	// set the parent supplied from --parent option
	if mf.Parent == "/" && m.Parent != "" {
		mf.RemoteParentID = m.Parent
	}

	if mf.Status == "CREATE" {
		mf.Logger().Debug("creating page")
		var queryParameters = &confluence.CreatePageQueryParameters{}
		if mf.Parent == "/" && mf.ID == "/" {
			queryParameters.RootLevel = true
		}

		page, err = m.CreatePage(createPageBody, queryParameters)
		if err != nil {
			return nil, err
		}

		mf.RemoteID = page.ID

	}

	if mf.Status == "UPDATE" {
		mf.Logger().Debug("updating page")
		var queryParameters = &confluence.CreatePageQueryParameters{}
		if mf.Parent == "/" && mf.ID == "/" {
			queryParameters.RootLevel = true
		}

		// createPageBody.RemoteParentID
		createPageBody.ID = mf.RemoteID
		createPageBody.Version = confluence.Version{
			Message: "sync from markdown2confluence",
		}

		page, err = m.UpdatePageWithVersionBump(mf.RemoteID, createPageBody)
		if err != nil {
			return nil, err
		}
	}

	mf.Status = "SYNCED"
	mf.Link = page.Links.Webui

	// Set the homepage
	if mf.ID == "/" && m.Parent == "" {
		err = m.UpdateSpaceHomePage(m.Space, mf.RemoteID)
		if err != nil {
			return nil, err
		}
	}

	_, errors := m.AddUpdateAttachments(mf.RemoteID, images)
	if len(errors) > 0 {
		err = errors[0]
	}

	return &page, err
}

func (mf *MarkdownFile) updateDB(db *bolt.DB) error {
	return db.Batch(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("pages"))
		value, err := json.Marshal(mf)
		if err != nil {
			return err
		}

		if mf.Status == "DELETED" {
			err = bucket.Delete([]byte(mf.ID))
		} else {
			err = bucket.Put([]byte(mf.ID), value)
		}
		if err != nil {
			return err
		}

		return nil
	})
}

func GetStoredMarkdownFile(db *bolt.DB, ID string) (*MarkdownFile, error) {
	var mf MarkdownFile

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("pages"))

		value := bucket.Get([]byte(ID))
		if value == nil {
			return nil
		}

		err := json.Unmarshal(value, &mf)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &mf, nil
}
