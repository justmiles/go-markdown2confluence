package cmd

import (
	"fmt"
	"log"
	"os"

	lib "github.com/justmiles/go-markdown2confluence/lib"

	"github.com/spf13/cobra"
)

var m lib.Markdown2Confluence

func init() {
	log.SetFlags(0)

	rootCmd.Flags().SetInterspersed(false)
	rootCmd.PersistentFlags().StringVarP(&m.Space, "space", "s", "", "Space in which content should be created")
	rootCmd.PersistentFlags().StringVarP(&m.Username, "username", "u", "", "Confluence username. (Alternatively set CONFLUENCE_USERNAME environment variable)")
	rootCmd.PersistentFlags().StringVarP(&m.Password, "password", "p", "", "Confluence password. (Alternatively set CONFLUENCE_PASSWORD environment variable)")
	rootCmd.PersistentFlags().StringVar(&m.APIToken, "api-token", "", "api-token for Confluence Cloud. (Alternatively set CONFLUENCE_API_TOKEN environment variable)")
	rootCmd.PersistentFlags().StringVarP(&m.AccessToken, "access-token", "a", "", "access-token for Confluence Data Center. (Alternatively set CONFLUENCE_ACCESS_TOKEN environment variable)")
	rootCmd.PersistentFlags().StringVarP(&m.Endpoint, "endpoint", "e", lib.DefaultEndpoint, "Confluence endpoint. (Alternatively set CONFLUENCE_ENDPOINT environment variable)")
	rootCmd.PersistentFlags().BoolVarP(&m.InsecureTLS, "insecuretls", "i", false, "Skip certificate validation. (e.g. for self-signed certificates)")
	rootCmd.PersistentFlags().BoolVarP(&m.Debug, "debug", "d", false, "Enable debug logging")

	// source environment variables
	m.Endpoint = os.Getenv("CONFLUENCE_ENDPOINT")
	m.Username = os.Getenv("CONFLUENCE_USERNAME")
	m.Password = os.Getenv("CONFLUENCE_PASSWORD")
	m.APIToken = os.Getenv("CONFLUENCE_API_TOKEN")
	m.AccessToken = os.Getenv("CONFLUENCE_ACCESS_TOKEN")
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "markdown2confluence",
	Short: "Push markdown files to Confluence",
	Long:  "A fast and flexible tool to syncronize or migrate your markdown documents to Confluence",
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
