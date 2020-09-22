package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"../dbconnector"
	"github.com/spf13/cobra"
)

func init() {
	runNodeCmd.PersistentFlags().StringVarP(&AppFlags.dataDirectory, "data-directory", "d", "./", "enclosing directory for node data")
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

	dbConnector := getDBConnector()
	defer dbConnector.Close()

	experimentVersionBytes, err := dbConnector.Get(ExperimentVersion)
	handleErrorWithPanic(err)

	experimentVersion, _ := strconv.Atoi(string(experimentVersionBytes))

	numberOfNodesBytes, err := dbConnector.Get(ExperimentNumberOfNodes)
	handleErrorWithPanic(err)

	numberOfNodes, _ := strconv.Atoi(string(numberOfNodesBytes))

	fmt.Println(fmt.Sprintf("experiment version: %d number of nodes: %d \n", experimentVersion, numberOfNodes))

	nodeID, mutexNodeFile, err := accuireLockOnNodeFile(numberOfNodes, dbConnector)
	if err != nil {
		fmt.Println(err)
		return
	}

	zipFileName := downloadNodeFolder(dbConnector, nodeID)
	extractZipFile(zipFileName)

	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

	mutexNodeList := accuireLockOnNodeList(dbConnector)

	nodeListResponse, err := dbConnector.Get(ExperimentNodeList)
	handleErrorWithPanic(err)

	nodeList := ""
	if nodeListResponse != nil {
		nodeList = string(nodeListResponse)
	}

	IPAddress := GetOutboundIP().String()
	basePortNumber := 9373

	algorandCmd, nodeNetAddress := startAlgorandProcess(nodeID, IPAddress, basePortNumber, nodeList)

	if nodeList == "" {
		nodeList = nodeNetAddress
	} else {
		nodeList = fmt.Sprintf("%s;%s", nodeList, nodeNetAddress)
	}

	err = dbConnector.Put(ExperimentNodeList, nodeList)
	handleErrorWithPanic(err)
	dbConnector.Unlock(mutexNodeList)

	err = algorandCmd.Wait()
	handleErrorWithPanic(err)

	fmt.Println("Node started successfuly. Waiting for commands...")

	dbConnector.WatchPutEvents(ExperimentNodeCommandStop)
	err = killAlgodProcess(nodeID)
	if err != nil {
		panic(err)
	}

	// Removes in use key
	nodeFileInUseKey := fmt.Sprintf("%s/%d", ExperimentNodeFilesInUse, nodeID)
	err = dbConnector.Delete(nodeFileInUseKey)
	handleErrorWithPanic(err)

	// Remove node information from node list
	mutexNodeList = accuireLockOnNodeList(dbConnector)

	err = removeNodeInfoFromNodeList(nodeID, IPAddress, basePortNumber, dbConnector)
	handleErrorWithPanic(err)

	dbConnector.Unlock(mutexNodeList)

	// Remove lock on node file
	dbConnector.Unlock(mutexNodeFile)

	fmt.Println("Node stoped successfuly.")

}

func removeNodeInfoFromNodeList(nodeID int, IPAddress string, basePortNumber int, dbConnector dbconnector.DBConnector) error {

	netAddress := getNetAddress(nodeID, IPAddress, basePortNumber)
	nodeListResp, err := dbConnector.Get(ExperimentNodeList)
	if err != nil {
		return err
	}

	if len(nodeListResp) == 0 {
		return fmt.Errorf("Node lisy is already empty.")
	}

	nodeListString := string(nodeListResp)

	fmt.Println("(BEF) Node list string:", nodeListString)

	nodeInfos := strings.Split(nodeListString, ";")

	var netAddressIndex = -1
	for i, n := range nodeInfos {
		if n == netAddress {
			netAddressIndex = i
			break
		}
	}

	if netAddressIndex == -1 {
		return fmt.Errorf("Could not find net address in the node list")
	}

	nodeInfos = append(nodeInfos[:netAddressIndex], nodeInfos[netAddressIndex+1:]...)

	nodeListString = strings.Join(nodeInfos[:], ";")

	fmt.Println("(AFT) Node list string:", nodeListString)

	err = dbConnector.Put(ExperimentNodeList, nodeListString)
	if err != nil {
		return err
	}

	return nil
}

func killAlgodProcess(nodeID int) error {

	dataFolderName := fmt.Sprintf("Node-%d", nodeID)
	pidFileName := fmt.Sprintf("%s/algod.pid", dataFolderName)
	pidBytes, err := ioutil.ReadFile(pidFileName)
	if err != nil {
		return fmt.Errorf("Error: Could not read algod pid file. Error message %s", err)
	}

	trimmedPIDString := strings.TrimSpace(string(pidBytes))
	algodPID, err := strconv.Atoi(trimmedPIDString)
	if err != nil {
		return fmt.Errorf("Error: Could not convert algod pid from string to int. Error message %s", err)
	}

	killExecutable, err := exec.LookPath("kill")
	if err != nil {
		return fmt.Errorf("Error: could not find kill in path")
	}

	killCmd := &exec.Cmd{
		Path: killExecutable,
		Args: []string{killExecutable, "-9", strconv.Itoa(algodPID)},
	}

	err = killCmd.Run()
	if err != nil {
		return fmt.Errorf("Error: Could not kill algod process. Error message %s", err)
	}

	return nil
}

func extractZipFile(zipFile string) {

	unzipExecutable, err := exec.LookPath("unzip")
	if err != nil {
		panic(fmt.Errorf("Error: could not find unzip in path"))
	}

	unzipCmd := &exec.Cmd{
		Path: unzipExecutable,
		Args: []string{unzipExecutable, zipFile, "-d", getDataDirectory()},
		//Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	err = unzipCmd.Run()
	handleErrorWithPanic(err)

}

func downloadNodeFolder(dbConnector dbconnector.DBConnector, nodeID int) string {

	directory := getDataDirectory()

	fileName := fmt.Sprintf("%sNode-%d.zip", directory, nodeID)
	keyName := fmt.Sprintf("%s/%s", ExperimentNodeFiles, fmt.Sprintf("Node-%d.zip", nodeID))

	zipFileBytes, err := dbConnector.Get(keyName)
	handleErrorWithPanic(err)

	err = ioutil.WriteFile(fmt.Sprintf("%s", fileName), zipFileBytes, 0644)
	handleErrorWithPanic(err)

	nodeFileInUseKey := fmt.Sprintf("%s/%d", ExperimentNodeFilesInUse, nodeID)
	err = dbConnector.Put(nodeFileInUseKey, "true")
	handleErrorWithPanic(err)

	return fileName
}

func accuireLockOnNodeFile(numberOfNodes int, dbConnector dbconnector.DBConnector) (int, string, error) {

	var err error

	for i := 0; i < numberOfNodes; i++ {
		mutexName := fmt.Sprintf("%s%d", ExperimentNodeLockKeyPrefix, i)

		err = dbConnector.TryLock(mutexName)
		if err != nil {
			continue
		}

		nodeFileInUseKey := fmt.Sprintf("%s/%d", ExperimentNodeFilesInUse, i)
		nodeFileInUseBytes, err := dbConnector.Get(nodeFileInUseKey)
		handleErrorWithPanic(err)

		if nodeFileInUseBytes != nil {
			err = dbConnector.Unlock(mutexName)
			handleErrorWithPanic(err)
			continue
		}

		return i, mutexName, nil
	}

	return 0, "", fmt.Errorf("Could not accuired lock. Did you run too many node?")
}

func accuireLockOnNodeList(dbConnector dbconnector.DBConnector) string {
	mutexName := ExperimentNodeListLock
	err := dbConnector.Lock(mutexName)
	handleErrorWithPanic(err)

	return mutexName
}

func startAlgorandProcess(nodeID int, IPAddress string, basePortNumber int, relayNodeList string) (*exec.Cmd, string) {

	goalExecutable, err := exec.LookPath("goal")
	if err != nil {
		panic(fmt.Errorf("Error: could not find goal in path"))
	}

	//goal node start -d data -p "ipaddress-1:4161;ipaddress-2:4161"

	directory := getDataDirectory()
	dataFolderName := fmt.Sprintf("%sNode-%d", directory, nodeID)

	nodeNetAddress := configureNodeNetAddress(dataFolderName, nodeID, IPAddress, basePortNumber)

	var args []string
	if len(relayNodeList) > 0 {
		args = []string{goalExecutable, "node", "start", "-d", dataFolderName, "-p", relayNodeList}
	} else {
		args = []string{goalExecutable, "node", "start", "-d", dataFolderName}
	}

	algorandCmd := &exec.Cmd{
		Path: goalExecutable,
		Args: args,
		//Stdout: os.Stdout,
		//Stderr: os.Stdout,
	}

	//fmt.Println("Command: ", algorandCmd.String())

	err = algorandCmd.Start()
	if err != nil {
		panic(fmt.Errorf("Error: could not start an algorand process"))
	}

	return algorandCmd, nodeNetAddress
}

func configureNodeNetAddress(dataFolderName string, nodeID int, IPAddress string, basePortNumber int) string {
	netAddress := getNetAddress(nodeID, IPAddress, basePortNumber)
	nodeConfig := getNodeConfig(dataFolderName)
	nodeConfig.NetAddress = netAddress

	writeNodeConfig(dataFolderName, nodeConfig)

	return netAddress
}

func getNetAddress(nodeID int, IPAddress string, basePortNumber int) string {
	return fmt.Sprintf("%s:%d", IPAddress, basePortNumber+nodeID)
}

/*********************************************************************************************************************/

type nodeConfig struct {
	GossipFanout    int
	EndpointAddress string
	DNSBootstrapID  string
	EnableProfiler  bool
	NetAddress      string
}

func getConfigFileName(dataFolderName string) string {
	return fmt.Sprintf("%s/config.json", dataFolderName)
}

func getNodeConfig(dataFolderName string) nodeConfig {

	configFileName := getConfigFileName(dataFolderName)

	jsonFile, err := os.Open(configFileName)
	if err != nil {
		panic(fmt.Errorf("Error: could not open node config file to READ: %s", configFileName))
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var config nodeConfig
	json.Unmarshal(byteValue, &config)

	jsonFile.Close()

	return config
}

func writeNodeConfig(dataFolderName string, config nodeConfig) {

	configFileName := getConfigFileName(dataFolderName)
	byteValue, err := json.Marshal(config)
	if err != nil {
		panic(fmt.Errorf("Error: could not marshal node config file: %v", err))
	}

	err = ioutil.WriteFile(configFileName, byteValue, 0666)
	handleErrorWithPanic(err)

}
