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
	"github.com/algorand/go-algorand-sdk/crypto"
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

	algorandCmd, nodeNetAddress, nodeEndpointAddress := startAlgorandProcess(nodeID, IPAddress, basePortNumber, nodeList)

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

	err = createAlgorandAccountAddFound(nodeID, dbConnector)
	if err != nil {
		fmt.Println("Create account add found error is ", err)
	}

	saveEndpointAddressAndAlgodTokenToDB(nodeID, nodeEndpointAddress, dbConnector)

	setTCRules(dbConnector)

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
		return fmt.Errorf("node lisy is already empty")
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

func startAlgorandProcess(nodeID int, IPAddress string, basePortNumber int, relayNodeList string) (*exec.Cmd, string, string) {

	goalExecutable, err := exec.LookPath("goal")
	if err != nil {
		panic(fmt.Errorf("Error: could not find goal in path"))
	}

	//goal node start -d data -p "ipaddress-1:4161;ipaddress-2:4161"

	directory := getDataDirectory()
	dataFolderName := fmt.Sprintf("%sNode-%d", directory, nodeID)

	nodeNetAddress, nodeEndpointAddress := configureNodeNetAndEndPointAddress(dataFolderName, nodeID, IPAddress, basePortNumber)

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

	return algorandCmd, nodeNetAddress, nodeEndpointAddress
}

func configureNodeNetAndEndPointAddress(dataFolderName string, nodeID int, IPAddress string, basePortNumber int) (string, string) {

	netAddress := getNetAddress(nodeID, IPAddress, basePortNumber)
	endPointAddress := getEndpointAddress(nodeID, IPAddress, basePortNumber)

	nodeConfig := getNodeConfig(dataFolderName)

	nodeConfig.NetAddress = netAddress
	nodeConfig.EndpointAddress = endPointAddress
	nodeConfig.TxPoolExponentialIncreaseFactor = 1
	nodeConfig.TxPoolSize = 1000000

	writeNodeConfig(dataFolderName, nodeConfig)

	return netAddress, endPointAddress
}

func getNetAddress(nodeID int, IPAddress string, basePortNumber int) string {
	return fmt.Sprintf("%s:%d", IPAddress, basePortNumber+nodeID)
}

func getEndpointAddress(nodeID int, IPAddress string, basePortNumber int) string {
	return fmt.Sprintf("%s:%d", IPAddress, basePortNumber+nodeID+5000)
}

/*********************************************************************************************************************/

type nodeConfig struct {
	GossipFanout                    int
	EndpointAddress                 string
	DNSBootstrapID                  string
	EnableProfiler                  bool
	NetAddress                      string
	TxPoolExponentialIncreaseFactor int
	TxPoolSize                      int
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

/*****************************************************/

func createAlgorandAccountAddFound(nodeID int, dbConnector dbconnector.DBConnector) error {

	account := crypto.GenerateAccount()

	fmt.Println("Account Address: ", account.Address)
	fmt.Println("Account Public KEY: ", string(account.PublicKey))
	fmt.Println("Account Private KEY: ", string(account.PrivateKey))

	err := addFound(nodeID, account)
	if err != nil {
		return fmt.Errorf("Error: could not add found to account. Error message is %s", err)
	}

	fmt.Println("Found added to account successfuly :)")

	err = saveAccountInfoToDB(nodeID, dbConnector, account)
	if err != nil {
		return fmt.Errorf("Could notsave account info to db %s", err)
	}

	return nil
}

func saveAccountInfoToDB(nodeID int, dbConnector dbconnector.DBConnector, account crypto.Account) error {

	algorandAccount := AlgorandAccount{
		Address:    account.Address.String(),
		PublicKey:  account.PublicKey,
		PrivateKey: account.PrivateKey}

	algorandAccountJSON, _ := json.Marshal(algorandAccount)

	accountKey := fmt.Sprintf("%s/%d", ExperimentAccountPrefix, nodeID)

	err := dbConnector.Put(accountKey, string(algorandAccountJSON))

	return err
}

func addFound(nodeID int, account crypto.Account) error {

	goalExecutable, err := exec.LookPath("goal")
	if err != nil {
		return fmt.Errorf("Error: could not find goal in path %s", err)
	}

	//$ goal clerk send --from=<my-account> --to=GD64YIY3TWGDMCNPP553DZPPR6LDUSFQOIJVFDPPXWEG3FVOJCCDBBHU5A --fee=1000 --amount=1000000 --note="Hello World"

	directory := getDataDirectory()
	dataFolderName := fmt.Sprintf("%sNode-%d", directory, nodeID)

	walletAddress, err := getWalletAddress(nodeID)
	if err != nil {
		return err
	}

	from := fmt.Sprintf("--from=%s", walletAddress)
	to := fmt.Sprintf("--to=%s", account.Address)
	fee := fmt.Sprintf("--fee=%d", 1000)
	amount := fmt.Sprintf("--amount=%d", 1000000000000)
	note := fmt.Sprintf("--note=%s", "\"Initial founding.\"")

	algorandCmd := &exec.Cmd{
		Path: goalExecutable,
		Args: []string{goalExecutable, "clerk", "send", "-d", dataFolderName, from, to, fee, amount, note, "-N"},
	}

	fmt.Println("Command is: ", algorandCmd.String())

	err = algorandCmd.Start()
	if err != nil {
		return err
	}

	return algorandCmd.Wait()
}

func getWalletAddress(nodeID int) (string, error) {

	walletAddress := ""
	goalExecutable, err := exec.LookPath("goal")
	if err != nil {
		return walletAddress, fmt.Errorf("Error: could not find goal in path %s", err)
	}

	//kadir@rita:~/Git/dandelion/my-network-16$ goal account -d Node-0 list
	//[n/a]	HISKEQ3DCHFARLYKGTNTVWXAIGGBIO2MS7ZIODEV7OIIFYWAJM5XS36HXM	HISKEQ3DCHFARLYKGTNTVWXAIGGBIO2MS7ZIODEV7OIIFYWAJM5XS36HXM	[n/a] microAlgos

	directory := getDataDirectory()
	dataFolderName := fmt.Sprintf("%sNode-%d", directory, nodeID)

	algorandCmd := &exec.Cmd{
		Path: goalExecutable,
		Args: []string{goalExecutable, "account", "list", "-d", dataFolderName},
	}

	commandOutput, err := algorandCmd.Output()
	if err != nil {
		return walletAddress, err
	}

	commandResult := string(commandOutput)
	fmt.Println("Goal account list command output: ", commandResult)

	walletAddressTokens := strings.Split(string(commandResult), "\t")

	if len(walletAddressTokens) != 4 {
		fmt.Println("Goal account list: ", commandResult)
		return walletAddress, fmt.Errorf("Expecting 4 tokens but received %d", len(walletAddressTokens))
	}

	walletAddress = walletAddressTokens[1]

	fmt.Println(fmt.Sprintf("Goal account list result ready: %s Wallet address: %s ", commandResult, walletAddress))

	return walletAddress, nil
}

func setTCRules(dbConnector dbconnector.DBConnector) {

	bandwidthBytes, err := dbConnector.Get(ExperimentNetworkBandwidth)
	if err != nil {
		panic(err)
	}

	delayBytes, err := dbConnector.Get(ExperimentNetworkDelay)
	if err != nil {
		panic(err)
	}

	if len(bandwidthBytes) == 0 && len(delayBytes) == 0 {
		return
	}

	bandwidth := string(bandwidthBytes)
	delay := string(delayBytes)

	tcSetExecutable, err := exec.LookPath("tcset")
	if err != nil {
		panic(fmt.Errorf("Error: could not find tcset in path"))
	}

	args1 := strings.Fields(fmt.Sprintf("eth0 --rate %sMbps --direction incoming", bandwidth))
	args1 = append([]string{tcSetExecutable}, args1...)

	tcSetIncommingCmd := &exec.Cmd{
		Path:   tcSetExecutable,
		Args:   args1,
		Stderr: os.Stdout,
	}

	err = tcSetIncommingCmd.Run()
	if err != nil {
		panic(fmt.Errorf("tcset incomming error: %s", err))
	}

	args2 := strings.Fields(fmt.Sprintf("eth0 --rate %sMbps --delay %sms --direction outgoing", bandwidth, delay))
	args2 = append([]string{tcSetExecutable}, args2...)

	tcSetOutgoingCmd := &exec.Cmd{
		Path:   tcSetExecutable,
		Args:   args2,
		Stderr: os.Stdout,
	}

	err = tcSetOutgoingCmd.Run()
	if err != nil {
		panic(fmt.Errorf("tcset outgoing error: %s", err))
	}

}

/***************************************************/

func saveEndpointAddressAndAlgodTokenToDB(nodeID int, endPointAddress string, dbConnector dbconnector.DBConnector) {

	directory := getDataDirectory()
	dataFolderName := fmt.Sprintf("%sNode-%d", directory, nodeID)

	//Get token from file
	algodTokenFile := fmt.Sprintf("%s/algod.token", dataFolderName)
	algodTokenBytes, err := ioutil.ReadFile(algodTokenFile)

	algodInfo := AlgodInfo{EndPointAddress: endPointAddress, Token: string(algodTokenBytes)}

	algodInfoKey := fmt.Sprintf("%s/%d", ExperimentAlgodInfoPrefix, nodeID)
	algodInfoJSON, _ := json.Marshal(algodInfo)

	err = dbConnector.Put(algodInfoKey, string(algodInfoJSON))
	if err != nil {
		panic(fmt.Errorf("Could not put algod info to db. Error: %s", err))
	}
}
