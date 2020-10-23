package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"../dbconnector"
	"github.com/spf13/cobra"
)

var BlockRandomPayloadSize int
var TCDelay int
var TCBandwidth int

func init() {
	rootCmd.AddCommand(createExperimentCmd)
	createExperimentCmd.Flags().IntVar(&BlockRandomPayloadSize, "block-payload-size", 1000000, "Specifies the maximum block size in bytes")
	createExperimentCmd.Flags().IntVar(&TCDelay, "tc-delay", 0, "Specifies nodes outgoing communication delay in milliseconds")
	createExperimentCmd.Flags().IntVar(&TCBandwidth, "tc-bandwidth", 0, "Specifies nodes outgoing and incomming data rate in Mbps")
}

var createExperimentCmd = &cobra.Command{
	Use:   "create-experiment [network folder path] [number of nodes]",
	Short: "Creates a new experiment on etcd and uploads data folders for nodes.",
	Args:  cobra.MinimumNArgs(2),
	Run:   createExperimentCmdRun,
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

	numberOfNodes := args[1]

	err = dbConnector.Put(ExperimentNumberOfNodes, numberOfNodes)
	handleErrorWithPanic(err)

	err = dbConnector.Put(ExperimentMaxBlockSize, strconv.Itoa(BlockRandomPayloadSize))
	handleErrorWithPanic(err)

	err = dbConnector.Put(ExperimentNetworkDelay, strconv.Itoa(TCDelay))
	handleErrorWithPanic(err)

	err = dbConnector.Put(ExperimentNetworkBandwidth, strconv.Itoa(TCBandwidth))
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
