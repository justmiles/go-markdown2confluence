package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.PersistentFlags().StringVarP(&m.Comment, "comment", "c", "", "(Optional) Add comment to page")
	syncCmd.PersistentFlags().StringVar(&m.Parent, "parent", "", "Optional parent page to nest content under")
	syncCmd.PersistentFlags().BoolVarP(&m.UseDocumentTitle, "use-document-title", "", false, "Will use the Markdown document title (# Title) if available")
	syncCmd.PersistentFlags().BoolVarP(&m.WithHardWraps, "hardwraps", "w", false, "Render newlines as <br />")
	syncCmd.PersistentFlags().BoolVarP(&m.ForceUpdates, "force", "f", false, "force an upload regardless of whether or not it changed locally")
	syncCmd.PersistentFlags().StringVarP(&m.Title, "title", "t", "", "Set the page title on upload (defaults to filename without extension)")
	syncCmd.PersistentFlags().StringSliceVarP(&m.ExcludeFilePatterns, "exclude", "x", []string{}, "regex expression to exclude matching files or file paths")
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync markdown files to Confluence",
	Run: func(rootCmd *cobra.Command, args []string) {
		m.SourceMarkdown = args
		// Validate the arguments
		err := m.Validate()
		if err != nil {
			log.Fatal(err)
		}

		err = m.Init()
		if err != nil {
			log.Fatal(err)
		}

		defer m.Close()

		creates, updates, deletes, err := m.PrepareSync()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		fmt.Printf("Sync Status: %d to add, %d to change, %d to delete.\n", creates, updates, deletes)
		// fmt.Println("Press the Enter to continue (skip this prompt with --auto-approve)")
		// fmt.Scanln() // wait for Enter Key

		err = m.Sync()
		if err != nil {
			log.Fatal(err)
		}

	},
}
