package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(clearExperimentCmd)
}

var clearExperimentCmd = &cobra.Command{
	Use:   "clear-experiment",
	Short: "Deletes in-use keys from the etcd",
	Run:   clearExperimentCmdRun,
}

func clearExperimentCmdRun(cmd *cobra.Command, args []string) {

	dbConnector := getDBConnector()
	defer dbConnector.Close()

	err := dbConnector.DeleteWithPrefix(ExperimentNodeFilesInUse)
	if err != nil {
		fmt.Println("Error: Could not delete in-use keys from etcd. The error is ", err)
	}

}
