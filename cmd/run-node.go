package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
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

	etcdAddres := getEtcdAddress()
	cli, err := getEtcdClient(etcdAddres)

	handleErrorWithPanic(err)
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//ctx, cancel := context.WithCancel(context.TODO())
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

	nodeID, mutexNodeFile, err := accuireLockOnNodeFile(numberOfNodes, session)
	if err != nil {
		fmt.Println(err)
		return
	}

	zipFileName := downloadNodeFolder(cli, nodeID)
	extractZipFile(zipFileName)

	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

	mutexNodeList := accuireLockOnNodeList(session)

	nodeListResponse, err := cli.Get(ctx, ExperimentNodeList)
	handleErrorWithPanic(err)

	nodeList := ""
	if len(nodeListResponse.Kvs) > 0 {
		nodeList = string(nodeListResponse.Kvs[0].Value)
	}

	algorandCmd, nodeNetAddress := startAlgorandProcess(nodeID, nodeList)

	if nodeList == "" {
		nodeList = nodeNetAddress
	} else {
		nodeList = fmt.Sprintf("%s;%s", nodeList, nodeNetAddress)
	}

	_, err = cli.Put(ctx, ExperimentNodeList, nodeList)
	handleErrorWithPanic(err)
	mutexNodeList.Unlock(context.TODO())

	err = algorandCmd.Wait()
	handleErrorWithPanic(err)

	fmt.Println("Node started successfuly. Waiting for commands...")

	waitForStopCommand(cli)
	err = killAlgodProcess(nodeID)
	if err != nil {
		panic(err)
	}

	// Removes in use key
	nodeFileInUseKey := fmt.Sprintf("%s/%d", ExperimentNodeFilesInUse, nodeID)
	_, err = cli.Delete(context.TODO(), nodeFileInUseKey)
	handleErrorWithPanic(err)

	// Remove node information from node list
	mutexNodeList = accuireLockOnNodeList(session)

	IPAddress := "127.0.0.1"
	basePortNumber := 9373
	err = removeNodeInfoFromNodeList(nodeID, IPAddress, basePortNumber, session)
	handleErrorWithPanic(err)

	mutexNodeList.Unlock(context.TODO())

	// Remove lock on node file
	mutexNodeFile.Unlock(context.TODO())

	fmt.Println("Node stoped successfuly.")

}

func removeNodeInfoFromNodeList(nodeID int, IPAddress string, basePortNumber int, etcdSession *concurrency.Session) error {

	netAddress := getNetAddress(nodeID, IPAddress, basePortNumber)
	nodeListResp, err := etcdSession.Client().Get(context.TODO(), ExperimentNodeList)
	if err != nil {
		return err
	}

	if len(nodeListResp.Kvs) == 0 {
		return nil
	}

	nodeListString := string(nodeListResp.Kvs[0].Value)

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

	_, err = etcdSession.Client().Put(context.TODO(), ExperimentNodeList, nodeListString)
	if err != nil {
		return err
	}

	return nil
}

func waitForStopCommand(etcdClient *clientv3.Client) {

	rch := etcdClient.Watch(context.Background(), ExperimentNodeCommandStop)
	for wresp := range rch {
		for _, ev := range wresp.Events {

			if ev.Type == clientv3.EventTypePut {
				return
			}

		}
	}

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
		Args: []string{unzipExecutable, zipFile},
		//Stdout: os.Stdout,
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

	nodeFileInUseKey := fmt.Sprintf("%s/%d", ExperimentNodeFilesInUse, nodeID)
	_, err = etcdClient.Put(ctx, nodeFileInUseKey, "true")
	handleErrorWithPanic(err)

	return fileName
}

func accuireLockOnNodeFile(numberOfNodes int, etcdSession *concurrency.Session) (int, *concurrency.Mutex, error) {

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

		nodeFileInUseKey := fmt.Sprintf("%s/%d", ExperimentNodeFilesInUse, i)
		nodeFileInUseResponse, err := etcdSession.Client().Get(context.TODO(), nodeFileInUseKey)
		handleErrorWithPanic(err)

		if nodeFileInUseResponse.Count != 0 {
			err = mutex.Unlock(context.TODO())
			handleErrorWithPanic(err)
			continue
		}

		return i, mutex, nil
	}

	return 0, nil, fmt.Errorf("Could not accuired lock. Did you run too many node?")
}

func accuireLockOnNodeList(etcdSession *concurrency.Session) *concurrency.Mutex {
	mutexName := ExperimentNodeListLock
	mutex := concurrency.NewMutex(etcdSession, mutexName)
	//TODO: Handle error here!!!!
	mutex.Lock(context.TODO())

	return mutex
}

func startAlgorandProcess(nodeID int, relayNodeList string) (*exec.Cmd, string) {

	goalExecutable, err := exec.LookPath("goal")
	if err != nil {
		panic(fmt.Errorf("Error: could not find goal in path"))
	}

	//goal node start -d data -p "ipaddress-1:4161;ipaddress-2:4161"

	dataFolderName := fmt.Sprintf("Node-%d", nodeID)

	IPAddress := "127.0.0.1"
	basePortNumber := 9373

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
