package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"../dbconnector"
	"github.com/spf13/cobra"
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

	dbConnector := getDBConnector()
	defer dbConnector.Close()

	response, err := dbConnector.GetWithPrefix(ExperimentKeyPrefix)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	if len(response) > 0 {
		fmt.Println("Error: there is already an experiment registered. Remove the experiment properly before starting a new one ")
		return
	}

	experimentVersion := args[1]
	numberOfNodes := args[2]

	err = dbConnector.Put(ExperimentVersion, experimentVersion)
	handleErrorWithPanic(err)

	err = dbConnector.Put(ExperimentNumberOfNodes, numberOfNodes)
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

			uploadToDB(dbConnector, f, reader)

			uploadedFileCount = uploadedFileCount + 1
		}

	}

	if uploadedFileCount == 0 {
		fmt.Println("Error: no data folder uploaded to etcd. Is this a correct network folder: ", networkFolderPath)
	}

}

func uploadToDB(dbConnector dbconnector.DBConnector, fileInfo os.FileInfo, reader io.Reader) {

	key := fileInfo.Name()
	value := loadFile(reader)

	err := dbConnector.Put(fmt.Sprintf("%s/%s", ExperimentNodeFiles, key), string(value))
	handleErrorWithPanic(err)
}

func loadFile(reader io.Reader) []byte {

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}

	return data
}
