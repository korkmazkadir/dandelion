package cmd

import (
	"log"
	"net"
	"time"

	"../dbconnector"
	"go.etcd.io/etcd/clientv3"
)

func handleErrorWithPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func getEtcdAddress() string {

	etcdAddres := "127.0.0.1:2379"
	if len(AppFlags.etcdAddress) > 0 {
		etcdAddres = AppFlags.etcdAddress
	}

	return etcdAddres
}

func getDataDirectory() string {
	return AppFlags.dataDirectory
}

func getEtcdClient(etcdAddress string) (*clientv3.Client, error) {

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdAddress},
		DialTimeout: 5 * time.Second,
	})

	return cli, err
}

func getDBConnector() dbconnector.DBConnector {

	etcdAddress := getEtcdAddress()
	connector, err := dbconnector.CreateEtcdConnector(etcdAddress)
	if err != nil {
		handleErrorWithPanic(err)
	}

	return connector
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
