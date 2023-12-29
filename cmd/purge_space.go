package cmd

import (
	"fmt"
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

		confirmation := askForConfirmation(fmt.Sprintf("Are you sure you want delete all pages in Space %s. This will delete all the things!\n", m.Space))

		if confirmation {
			fmt.Print("Deleting...")

			err = m.PurgeSpace()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Print(" done!\n")
		}

	},
}
