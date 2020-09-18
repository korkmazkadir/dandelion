package cmd

import (
	"time"

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

func getEtcdClient(etcdAddress string) (*clientv3.Client, error) {

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdAddress},
		DialTimeout: 5 * time.Second,
	})

	return cli, err
}
