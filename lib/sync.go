package lib

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

func (m *Markdown2Confluence) Sync() error {

	// create the queue to be processed
	var (
		wg     = sync.WaitGroup{}
		queue  = make(chan *MarkdownFile)
		errors []error
		err    error
	)

	// Process the queue
	for worker := 0; worker < Parallelism; worker++ {
		wg.Add(1)
		go m.queueProcessor(&wg, &queue, &errors)
	}

	var waitingOnParentPages = true
	maxIterations := 300 // Prevent infinite loops (5 minutes with 1 second sleep)
	iteration := 0

	for waitingOnParentPages && iteration < maxIterations {
		waitingOnParentPages = false
		queuedCount := 0
		waitingCount := 0

		for _, markdownFile := range m.files {

			if markdownFile.Status == "SYNCED" || markdownFile.Status == "ERRORED" || markdownFile.Status == "DELETED" {
				continue
			}

			// If the parent is root (/ or .) or RemoteParentID is already set, queue the file
			isRootParent := markdownFile.Parent == "/" || markdownFile.Parent == "."

			if markdownFile.ID != "/" && markdownFile.RemoteParentID == "" && !isRootParent {
				waitingOnParentPages = true
				waitingCount++

				// Update the child pages with the new parent ID
				if val, ok := m.files[markdownFile.Parent]; ok {
					if val.RemoteID != "" {
						markdownFile.RemoteParentID = val.RemoteID
						m.files[markdownFile.ID] = markdownFile
						markdownFile.Logger().Infof("resolved parent ID for %s", markdownFile.Path)
					}

					// If we can't create the parent page, we can't create the child pages
					if val.Status == "ERRORED" {
						markdownFile.Logger().Errorf("cannot sync (parent errored): %s", markdownFile.Path)
						markdownFile.Status = "ERRORED"
						m.files[markdownFile.ID] = markdownFile
						waitingCount--
					}
				} else {
					markdownFile.Logger().Warnf("parent page not found in files map: %s (parent: %s)", markdownFile.Path, markdownFile.Parent)
				}
			} else {
				queue <- markdownFile
				queuedCount++
			}
		}

		iteration++
		if waitingOnParentPages {
			log.Infof("sync iteration %d: queued %d files, %d waiting on parents", iteration, queuedCount, waitingCount)
			time.Sleep(1 * time.Second)
		}
	}

	if iteration >= maxIterations {
		log.Errorf("sync timeout: reached maximum iterations (%d), some files may not have synced", maxIterations)
		// Mark remaining files as errored
		for _, markdownFile := range m.files {
			if markdownFile.Status != "SYNCED" && markdownFile.Status != "ERRORED" && markdownFile.Status != "DELETED" {
				markdownFile.Logger().Errorf("timeout waiting for parent page")
				markdownFile.Status = "ERRORED"
				m.files[markdownFile.ID] = markdownFile
			}
		}
	}

	close(queue)

	wg.Wait()

	m.save()

	return err
}
