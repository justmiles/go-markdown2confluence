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

		fmt.Printf("Press the Enter to continue. This will delete all the things in %s!\n", m.Space)
		fmt.Scanln() // wait for Enter Key
		fmt.Scanln() // wait for Enter Key, again :D
		fmt.Print("purging...")

		err = m.PurgeSpace()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(" done!\n")
	},
}
