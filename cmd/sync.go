package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	autoApprove bool
)

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.PersistentFlags().StringVarP(&m.Comment, "comment", "c", "", "(Optional) Add comment to page")
	syncCmd.PersistentFlags().StringVar(&m.Parent, "parent", "", "Optional parent page to nest content under")
	syncCmd.PersistentFlags().BoolVarP(&m.UseDocumentTitle, "use-document-title", "", false, "Use Markdown document title (# Title) if available")
	syncCmd.PersistentFlags().BoolVarP(&m.WithHardWraps, "hardwraps", "w", false, "Render newlines as <br />")
	syncCmd.PersistentFlags().BoolVarP(&m.ForceUpdates, "force", "f", false, "Force an upload regardless of whether or not it changed locally")
	syncCmd.PersistentFlags().StringVarP(&m.Title, "title", "t", "", "Set the page title on upload (defaults to filename without extension)")
	syncCmd.PersistentFlags().StringSliceVarP(&m.ExcludeFilePatterns, "exclude", "x", []string{}, "Regex expressions to exclude matching files or file paths")
	syncCmd.PersistentFlags().BoolVar(&autoApprove, "auto-approve", false, "Automatically approve changes")
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

		fmt.Printf("%d to add, %d to change, %d to delete.\n", creates, updates, deletes)

		// if there's nothing to do, exit cleanly
		if creates+updates+deletes == 0 {
			os.Exit(0)
		}

		if !autoApprove {
			confirmation := askForConfirmation("Do you want to continue? ")
			if !confirmation {
				os.Exit(0)
			}
		}

		err = m.Sync()
		if err != nil {
			log.Fatal(err)
		}

	},
}
