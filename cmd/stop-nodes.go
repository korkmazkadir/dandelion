package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(stopNodesCmd)
}

var stopNodesCmd = &cobra.Command{
	Use:   "stop-nodes",
	Short: "Stop all running nodes. It does not collect logs.",
	Run:   stopNodesCmdRun,
}

func stopNodesCmdRun(cmd *cobra.Command, args []string) {

	dbConnector := getDBConnector()
	defer dbConnector.Close()

	err := dbConnector.Put(ExperimentNodeCommandStop, "true")
	if err != nil {
		fmt.Println("Error: Could not set command/kill key on etcd. The error is ", err)
	}

}
