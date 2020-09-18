package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.etcd.io/etcd/clientv3"
)

func init() {
	rootCmd.AddCommand(createExperimentCmd)
}

var createExperimentCmd = &cobra.Command{
	Use:   "create-experiment [network folder path] [experiment version] [number of nodes]",
	Short: "Creates a new experiment on etcd and uploads data folders for nodes.",
	Args:  createExperimentCmdValidateArgs,
	Run:   createExperimentCmdRun,
}

func createExperimentCmdValidateArgs(cmd *cobra.Command, args []string) error {

	if len(args) < 3 {
		return errors.New("requires at least 3 arguments: [network folder path] [experiment version] [number of nodes]")
	}

	networkFolder := args[0]
	_, err := os.Stat(networkFolder)
	if os.IsNotExist(err) {
		return fmt.Errorf("Error: Network folder does not exist: %s", networkFolder)
	}

	_, err = strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("Error: experiment version must be an integer: %s", args[1])
	}

	_, err = strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("Error: number of nodes version must be an integer: %s", args[2])
	}

	return nil
}

func createExperimentCmdRun(cmd *cobra.Command, args []string) {

	etcdAddres := getEtcdAddress()
	cli, err := getEtcdClient(etcdAddres)

	handleErrorWithPanic(err)
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := cli.Get(ctx, ExperimentKeyPrefix, clientv3.WithPrefix(), clientv3.WithCountOnly())
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	if response.Count > 0 {
		fmt.Println("Error: there is already an experiment registered. Remove the experiment properly before starting a new one ")
		return
	}

	experimentVersion := args[1]
	numberOfNodes := args[2]

	_, err = cli.Put(ctx, ExperimentVersion, experimentVersion)
	handleErrorWithPanic(err)

	_, err = cli.Put(ctx, ExperimentNumberOfNodes, numberOfNodes)
	handleErrorWithPanic(err)

	networkFolderPath := args[0]
	files, err := ioutil.ReadDir(networkFolderPath)
	handleErrorWithPanic(err)

	uploadedFileCount := 0
	for _, f := range files {

		fileName := f.Name()
		if strings.Contains(fileName, ".zip") {

			var reader io.Reader
			var err error

			reader, err = os.Open(networkFolderPath + fileName)
			if err != nil {
				panic(err)
			}

			uploadETCD(cli, f, reader)

			uploadedFileCount = uploadedFileCount + 1
		}

	}

	if uploadedFileCount == 0 {
		fmt.Println("Error: no data folder uploaded to etcd. Is this a correct network folder: ", networkFolderPath)
	}

}

func uploadETCD(etcdCli *clientv3.Client, fileInfo os.FileInfo, reader io.Reader) {

	key := fileInfo.Name()
	value := loadFile(reader)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	_, err := etcdCli.Put(ctx, fmt.Sprintf("%s/%s", ExperimentNodeFiles, key), string(value))
	cancel()

	handleErrorWithPanic(err)
}

func loadFile(reader io.Reader) []byte {

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}

	return data
}
