package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deleteExperimentCmd)
}

var deleteExperimentCmd = &cobra.Command{
	Use:   "delete-experiment",
	Short: "Deletes all keys from the etcd",
	Run:   deleteExperimentCmdRun,
}

func deleteExperimentCmdRun(cmd *cobra.Command, args []string) {

	dbConnector := getDBConnector()
	defer dbConnector.Close()

	err := dbConnector.DeleteWithPrefix(ExperimentKeyPrefix)
	if err != nil {
		fmt.Println("Error: Could not delete experiment from etcd. The error is ", err)
	}

}
