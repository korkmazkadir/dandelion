package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dandelion",
	Short: "dandelion is a configuration tool for private algorand networks",
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
