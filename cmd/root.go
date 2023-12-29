package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	lib "github.com/justmiles/go-markdown2confluence/lib"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	m lib.Markdown2Confluence
)

func init() {

	rootCmd.Flags().SetInterspersed(false)
	rootCmd.PersistentFlags().StringVarP(&m.Space, "space", "s", "", "Space in which content should be created")
	rootCmd.PersistentFlags().StringVarP(&m.Username, "username", "u", "", "Confluence username (CONFLUENCE_USERNAME environment variable can be used as an alternative)")
	rootCmd.PersistentFlags().StringVarP(&m.Password, "password", "p", "", "Confluence password (CONFLUENCE_PASSWORD environment variable can be used as an alternative)")
	rootCmd.PersistentFlags().StringVar(&m.APIToken, "api-token", "", "API token for Confluence Cloud (CONFLUENCE_API_TOKEN environment variable can be used as an alternative)")
	rootCmd.PersistentFlags().StringVarP(&m.AccessToken, "access-token", "a", "", "Access token for Confluence Data Center (CONFLUENCE_ACCESS_TOKEN environment variable can be used as an alternative)")
	rootCmd.PersistentFlags().StringVarP(&m.Endpoint, "endpoint", "e", lib.DefaultEndpoint, "Confluence endpoint (CONFLUENCE_ENDPOINT environment variable can be used as an alternative)")
	rootCmd.PersistentFlags().BoolVarP(&m.InsecureTLS, "insecuretls", "i", false, "Skip certificate validation (e.g., for self-signed certificates)")
	rootCmd.PersistentFlags().StringVarP(&m.LocalStore, "local-store", "l", "markdown2confluence.db", "Path to the local storage database")
	rootCmd.PersistentFlags().StringVar(&m.LogLevel, "log-level", "error", "Verbosity log level (error, info, debug, or trace)")

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

func askForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}
