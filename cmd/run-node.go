package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
)

func init() {
	rootCmd.AddCommand(runNodeCmd)
}

var runNodeCmd = &cobra.Command{
	Use:   "run-node",
	Short: "Accuires a lock, downloads data folder and runs a algorand node.",
	Long:  "Accuires a lock, downloads data folder and runs a algorand node.",
	//Args:  cobra.MinimumNArgs(2),
	Run: runNodeCmdRun,
}

func runNodeCmdRun(cmd *cobra.Command, args []string) {

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 2 * time.Second,
	})
	handleErrorWithPanic(err)
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	experimentVersionResp, err := cli.Get(ctx, ExperimentVersion)
	handleErrorWithPanic(err)

	experimentVersion, _ := strconv.Atoi(string(experimentVersionResp.Kvs[0].Value))

	numberOfNodesResp, err := cli.Get(ctx, ExperimentNumberOfNodes)
	handleErrorWithPanic(err)

	numberOfNodes, _ := strconv.Atoi(string(numberOfNodesResp.Kvs[0].Value))

	fmt.Println(fmt.Sprintf("experiment version: %d number of nodes: %d \n", experimentVersion, numberOfNodes))

	session, _ := concurrency.NewSession(cli)
	defer session.Close()

	nodeID, mutex, err := accuireLock(numberOfNodes, session)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	fmt.Println("Mutex locked: ", nodeID)

	zipFileName := downloadNodeFolder(cli, nodeID)
	extractZipFile(zipFileName)

	time.Sleep(2 * time.Second)
	mutex.Unlock(context.TODO())

}

func runAlgorandNode(dataFolder string) {

}

func extractZipFile(zipFile string) {

	unzipExecutable, err := exec.LookPath("unzip")
	if err != nil {
		panic(fmt.Errorf("Error: could not find unzip in path"))
	}

	unzipCmd := &exec.Cmd{
		Path:   unzipExecutable,
		Args:   []string{unzipExecutable, zipFile},
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	err = unzipCmd.Run()
	handleErrorWithPanic(err)

}

func downloadNodeFolder(etcdClient *clientv3.Client, nodeID int) string {

	fileName := fmt.Sprintf("Node-%d.zip", nodeID)
	keyName := fmt.Sprintf("%s/%s", ExperimentNodeFiles, fileName)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	zipFile, err := etcdClient.Get(ctx, keyName)
	handleErrorWithPanic(err)

	zipFileBytes := zipFile.Kvs[0].Value

	err = ioutil.WriteFile(fmt.Sprintf("./%s", fileName), zipFileBytes, 0644)
	handleErrorWithPanic(err)

	return fileName
}

func accuireLock(numberOfNodes int, etcdSession *concurrency.Session) (int, *concurrency.Mutex, error) {

	var err error

	for i := 0; i < numberOfNodes; i++ {
		mutexName := fmt.Sprintf("%s%d", ExperimentNodeLockKeyPrefix, i)
		mutex := concurrency.NewMutex(etcdSession, mutexName)

		err = mutex.TryLock(context.TODO())
		if err == concurrency.ErrLocked {
			continue
		}

		if err != nil {
			return 0, nil, err
		}

		return i, mutex, nil
	}

	return 0, nil, fmt.Errorf("Could not accuired lock. Did you run too many node?")
}
