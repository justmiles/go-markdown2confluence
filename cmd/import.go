package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(importCmd)
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "import remote page IDs into the local database (based on document title)",
	Run: func(rootCmd *cobra.Command, args []string) {
		m.SourceMarkdown = args

		err := m.Init()
		if err != nil {
			log.Fatal(err)
		}

		err = m.Import()
		if err != nil {
			log.Fatal(err)
		}

	},
}
