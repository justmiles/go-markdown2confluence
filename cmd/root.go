package cmd

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"

	lib "github.com/justmiles/go-markdown2confluence/lib"

	"github.com/spf13/cobra"
)

var m lib.Markdown2Confluence

func init() {
	log.SetFlags(0)

	rootCmd.Flags().SetInterspersed(false)
	rootCmd.PersistentFlags().StringVarP(&m.Space, "space", "s", "", "Space in which page should be created")
	rootCmd.PersistentFlags().StringVarP(&m.Comment, "comment", "c", "", "(Optional) Add comment to page")
	rootCmd.PersistentFlags().StringVarP(&m.Username, "username", "u", "", "Confluence username. (Alternatively set CONFLUENCE_USERNAME environment variable)")
	rootCmd.PersistentFlags().StringVarP(&m.Password, "password", "p", "", "Confluence password. (Alternatively set CONFLUENCE_PASSWORD environment variable)")
	rootCmd.PersistentFlags().StringVarP(&m.AccessToken, "access-token", "a", "", "Confluence access-token. (Alternatively set CONFLUENCE_ACCESS_TOKEN environment variable)")
	rootCmd.PersistentFlags().StringVarP(&m.Endpoint, "endpoint", "e", lib.DefaultEndpoint, "Confluence endpoint. (Alternatively set CONFLUENCE_ENDPOINT environment variable)")
	rootCmd.PersistentFlags().BoolVarP(&m.InsecureTLS, "insecuretls", "i", false, "Skip certificate validation. (e.g. for self-signed certificates)")
	rootCmd.PersistentFlags().StringVar(&m.Parent, "parent", "", "Optional parent page to next content under")
	rootCmd.PersistentFlags().BoolVarP(&m.Debug, "debug", "d", false, "Enable debug logging")
	rootCmd.PersistentFlags().BoolVarP(&m.UseDocumentTitle, "use-document-title", "", false, "Will use the Markdown document title (# Title) if available")
	rootCmd.PersistentFlags().BoolVarP(&m.WithHardWraps, "hardwraps", "w", false, "Render newlines as <br />")
	rootCmd.PersistentFlags().IntVarP(&m.Since, "modified-since", "m", 0, "Only upload files that have modifed in the past n minutes")
	rootCmd.PersistentFlags().StringVarP(&m.Title, "title", "t", "", "Set the page title on upload (defaults to filename without extension)")
	rootCmd.PersistentFlags().StringSliceVarP(&m.ExcludeFilePatterns, "exclude", "x", []string{}, "list of exclude file patterns (regex) for that will be applied on markdown file paths")
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
		if m.InsecureTLS {
			fmt.Println("Warning: TLS verification is disabled. This allows for man-in-the-middle-attacks.")
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}

		errors := m.Run()
		for _, err := range errors {
			fmt.Println()
			fmt.Println(err)
		}
		if len(errors) > 0 {
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	rootCmd.Version = version
	rootCmd.SetVersionTemplate(`{{with .Name}}{{printf "%s " .}}{{end}}{{printf "%s" .Version}}
`)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
