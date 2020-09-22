package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type applicationFlags struct {
	etcdAddress   string
	dataDirectory string
}

var AppFlags applicationFlags

func init() {
	rootCmd.PersistentFlags().StringVarP(&AppFlags.etcdAddress, "etcdt-address", "e", "127.0.0.1:2379", "inet address of etcd keyvalue store")
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
