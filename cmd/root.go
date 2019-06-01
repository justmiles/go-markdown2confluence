package cmd

import (
	"fmt"
	"log"
	"os"

	markdown2confluence "github.com/justmiles/go-markdown2confluence/markdown2confluence"

	"github.com/spf13/cobra"
)

var m markdown2confluence.Markdown2Confluence

func init() {
	log.SetFlags(0)

	rootCmd.Flags().SetInterspersed(false)
	rootCmd.PersistentFlags().StringVarP(&m.Space, "space", "s", "", "Space in which page should be created")
	rootCmd.PersistentFlags().StringVarP(&m.Username, "username", "u", "", "Confluence username. (Alternatively set CONFLUENCE_USERNAME environment variable)")
	rootCmd.PersistentFlags().StringVarP(&m.Password, "password", "p", "", "Confluence password. (Alternatively set CONFLUENCE_PASSWORD environment variable)")
	rootCmd.PersistentFlags().StringVarP(&m.Endpoint, "endpoint", "e", markdown2confluence.DefaultEndpoint, "Confluence endpoint. (Alternatively set CONFLUENCE_ENDPOINT environment variable)")
	rootCmd.PersistentFlags().BoolVarP(&m.Debug, "debug", "d", false, "Enable debug logging")

	m.SourceEnvironmentVariables()

}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "markdown2confluence",
	Short: "Push markdown files to Confluence Cloud",
	Run: func(rootCmd *cobra.Command, args []string) {
		m.SourceMarkdown = args
		// Validate the arguments
		err := m.Validate()
		if err != nil {
			log.Fatal(err)
		}

		err = m.Run()
		if err != nil {
			log.Fatal(err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	rootCmd.Version = version
	rootCmd.SetVersionTemplate(`{{with .Name}}{{printf "%s " .}}{{end}}{{printf "v%s" .Version}}
`)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
