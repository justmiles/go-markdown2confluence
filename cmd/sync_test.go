package cmd

import (
	"fmt"
	"log"
	"testing"

	"github.com/justmiles/go-markdown2confluence/lib"
)

func TestSync(t *testing.T) {
	t.Skip()

	m := lib.Markdown2Confluence{
		Space:            "TEST",
		SourceMarkdown:   []string{"test-docs/mkdocs-1.5.3/docs"},
		UseDocumentTitle: true,
	}

	err := m.Init()
	if err != nil {
		log.Fatal(err)
	}

	creates, updates, deletes, err := m.PrepareSync()
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("Sync Status: %d to add, %d to change, %d to delete.\n", creates, updates, deletes)

	err = m.Sync()
	if err != nil {
		t.Error(err)
	}

}
