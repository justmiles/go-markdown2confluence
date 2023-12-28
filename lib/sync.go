package lib

import (
	"sync"
	"time"
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
	for waitingOnParentPages {
		waitingOnParentPages = false

		for _, markdownFile := range m.files {

			if markdownFile.Status == "SYNCED" || markdownFile.Status == "ERRORED" || markdownFile.Status == "DELETED" {
				continue
			}

			if markdownFile.ID != "/" && markdownFile.RemoteParentID == "" {
				waitingOnParentPages = true
				// Update the child pages with the new parent ID
				if val, ok := m.files[markdownFile.Parent]; ok {
					if val.RemoteID != "" {
						markdownFile.RemoteParentID = val.RemoteID
						m.files[markdownFile.ID] = markdownFile
					}

					// If we can't create the parent page, we can't create the child pages
					if val.Status == "ERRORED" {
						markdownFile.Logger().Error("cannot sync (parent errored) ")
						markdownFile.Status = "ERRORED"
						m.files[markdownFile.ID] = markdownFile
					}
				}
			} else {
				queue <- markdownFile
			}
		}

		time.Sleep(1 * time.Second)
	}

	close(queue)

	wg.Wait()

	m.save()

	return err
}
