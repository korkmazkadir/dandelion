package cmd

import (
	"context"
	"fmt"
	"time"

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

	etcdAddres := getEtcdAddress()
	cli, err := getEtcdClient(etcdAddres)

	handleErrorWithPanic(err)
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	_, err = cli.Put(ctx, ExperimentNodeCommandStop, "true")
	if err != nil {
		fmt.Println("Error: Could not set command/kill key on etcd. The error is ", err)
	}

}
