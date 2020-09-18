package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type applicationFlags struct {
	etcdAddress string
}

var AppFlags applicationFlags

func init() {
	rootCmd.PersistentFlags().StringVarP(&AppFlags.etcdAddress, "etcdt-address", "e", "", "inet address of etcd keyvalue store(default value is 127.0.0.1:2379)")
}

var rootCmd = &cobra.Command{
	Use:   "dandelion",
	Short: "dandelion is a configuration tool for private algorand networks",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
