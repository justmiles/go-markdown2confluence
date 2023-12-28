package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(purgeSpaceCmd)
}

var purgeSpaceCmd = &cobra.Command{
	Use:   "purge-space",
	Short: "Delete all pages from a Space - useful for a fresh sync",
	Run: func(rootCmd *cobra.Command, args []string) {

		err := m.Init()
		if err != nil {
			log.Fatal(err)
		}

		err = m.PurgeSpace()
		if err != nil {
			log.Fatal(err)
		}
	},
}
