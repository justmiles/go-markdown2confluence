package lib

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/boltdb/bolt"
)

func (m *Markdown2Confluence) PrepareSync() (int, int, int, error) {

	// Determine what files are in scope for updates
	for _, sourceMarkdown := range m.SourceMarkdown {

		file, err := os.Open(sourceMarkdown)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("error opening file %s", err)
		}
		defer file.Close()

		err = filepath.Walk(sourceMarkdown,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if m.IsIncluded(info) {
					markdownFiles, err := resolveSyncStatus(m.db, path, sourceMarkdown, m.UseDocumentTitle, m.ForceUpdates)
					if err != nil {
						return err
					}
					for _, markdownFile := range markdownFiles {
						m.files[markdownFile.ID] = markdownFile
					}
				}
				return nil
			})
		if err != nil {
			return 0, 0, 0, fmt.Errorf("unable to walk path: %s", sourceMarkdown)
		}
	}

	// Determine what pages no longer exist on disk and should be deleted
	err := m.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("pages"))

		return bucket.ForEach(func(key, value []byte) error {
			var mf MarkdownFile
			err := json.Unmarshal(value, &mf)
			if err != nil {
				return err
			}

			_, err = os.Stat(mf.Path)
			if err != nil {
				if os.IsNotExist(err) {
					mf.Status = "DELETE"
					m.files[mf.ID] = &mf
				}
			}
			return nil
		})
	})
	if err != nil {
		log.Fatal(err)
	}

	var creates, updates, deletes = 0, 0, 0

	for _, mf := range m.files {
		switch mf.Status {
		case "CREATE":
			creates += 1
		case "UPDATE":
			updates += 1
		case "DELETE":
			deletes += 1
		}

	}

	return creates, updates, deletes, nil
}

func resolveSyncStatus(db *bolt.DB, path, sourceMarkdown string, useDocumentTitle, forceUpdate bool) ([]*MarkdownFile, error) {

	// Ensure directories that contain README or INDEX files are not duplicated
	if isDir(path) && hasIndexFiles(path) {
		return nil, nil
	}

	isIndex := (strings.HasSuffix(strings.ToUpper(path), "README.MD") || strings.HasSuffix(strings.ToUpper(path), "INDEX.MD"))

	var markdownFiles []*MarkdownFile
	var ID string
	if isIndex {
		ID = filepath.Dir(strings.Replace(path, sourceMarkdown, "", 1))
	} else {
		ID = strings.Replace(path, sourceMarkdown, "", 1)

	}

	contextLogger := log.WithFields(log.Fields{
		"file": ID,
	})

	mf, err := GetStoredMarkdownFile(db, ID)
	if err != nil {
		return nil, err
	}

	if mf.RemoteID == "" {
		mf = &MarkdownFile{
			ID: ID,
		}
	}

	mf.Path = path

	if isIndex {
		mf.Title = filepath.Base(filepath.Dir(mf.ID))
		mf.Parent = filepath.Dir(filepath.Dir(mf.ID))
	} else {
		mf.Title = strings.TrimSuffix(filepath.Base(mf.ID), ".md")
		mf.Parent = filepath.Dir(mf.ID)
	}

	if useDocumentTitle || isIndex {
		docTitle := getDocumentTitle(path)
		if docTitle != "" {
			mf.Title = docTitle
		}
	}

	if mf.Parent != "/" && mf.Parent != "." {
		mff, err := resolveSyncStatus(db, filepath.Dir(mf.Path), sourceMarkdown, false, forceUpdate)
		if err != nil {
			return nil, err
		}

		markdownFiles = append(markdownFiles, mff...)
	}

	markdownFiles = append(markdownFiles, mf)

	if mf.RemoteID == "" {
		mf.Status = "CREATE"
	}

	mdfSum, err := md5sum(path)
	if err != nil {
		return nil, err
	}

	if !isDir(path) {
		if forceUpdate && mf.MD5Sum != "" {
			mf.Status = "UPDATE"
		}
		// File has changed on disk since we last created it, let's update
		if mf.MD5Sum != mdfSum && mf.MD5Sum != "" {
			contextLogger.Debug(fmt.Sprintf("md5sum changed (%s - %s)", mdfSum, mf.MD5Sum))
			mf.Status = "UPDATE"
		}

		mf.MD5Sum = mdfSum

	}

	return markdownFiles, nil
}

func md5sum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	if fileInfo.IsDir() {
		return "DIRECTORY", nil
	}

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func isDir(filePath string) bool {
	file, _ := os.Open(filePath)
	defer file.Close()

	fileInfo, _ := file.Stat()
	return fileInfo.IsDir()
}

func hasIndexFiles(filePath string) bool {
	file, _ := os.Open(filePath)
	defer file.Close()

	names, _ := file.Readdirnames(0)

	for _, name := range names {
		if strings.HasSuffix(strings.ToUpper(name), "README.MD") || strings.HasSuffix(strings.ToUpper(name), "INDEX.MD") {
			return true
		}
	}
	return false
}
