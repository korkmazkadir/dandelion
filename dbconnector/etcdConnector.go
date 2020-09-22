package dbconnector

import (
	"context"
	"fmt"
	"time"

	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
)

/*EtcdConnector keeps etcd client related objects*/
type EtcdConnector struct {
	cli     *clientv3.Client
	session *concurrency.Session
	lockMap map[string]*concurrency.Mutex
}

/*CreateEtcdConnector creates an EtcdConnector object*/
func CreateEtcdConnector(netAddress string) (*EtcdConnector, error) {

	etcdCli, err := createEtcdClient(netAddress)
	if err != nil {
		return nil, err
	}

	etcdSession, err := createEtcdSession(etcdCli)
	if err != nil {
		return nil, err
	}

	connector := EtcdConnector{
		cli:     etcdCli,
		session: etcdSession,
		lockMap: make(map[string]*concurrency.Mutex),
	}

	return &connector, nil
}

func createEtcdClient(netAddress string) (*clientv3.Client, error) {

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{netAddress},
		DialTimeout: 5 * time.Second,
	})

	return cli, err
}

func createEtcdSession(cli *clientv3.Client) (*concurrency.Session, error) {
	session, err := concurrency.NewSession(cli)
	return session, err
}

func (connector EtcdConnector) Get(key string) ([]byte, error) {

	resp, err := connector.cli.Get(context.TODO(), key)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) > 0 {
		return resp.Kvs[0].Value, nil
	}

	return nil, nil
}

func (connector EtcdConnector) GetWithPrefix(prefix string) ([][]byte, error) {

	resp, err := connector.cli.Get(context.TODO(), prefix, clientv3.WithPrefix())

	if err != nil {
		return nil, err
	}

	results := make([][]byte, len(resp.Kvs))
	for i := 0; i < len(resp.Kvs); i++ {
		results[i] = resp.Kvs[0].Value
	}

	return results, nil
}

func (connector EtcdConnector) Put(key string, value string) error {

	_, err := connector.cli.Put(context.TODO(), key, value)
	return err
}

func (connector EtcdConnector) Delete(key string) error {

	_, err := connector.cli.Delete(context.TODO(), key)
	return err
}

func (connector EtcdConnector) Lock(name string) error {

	mutex := concurrency.NewMutex(connector.session, name)
	err := mutex.Lock(context.TODO())
	if err == nil {
		connector.lockMap[name] = mutex
	}

	return err
}

func (connector EtcdConnector) Unlock(name string) error {

	mutex, isAvailable := connector.lockMap[name]
	if isAvailable == false {
		return fmt.Errorf("Error: there is no lock with name %s", name)
	}

	err := mutex.Unlock(context.TODO())
	if err == nil {
		delete(connector.lockMap, name)
	}

	return err
}

func (connector EtcdConnector) TryLock(name string) error {

	mutex := concurrency.NewMutex(connector.session, name)
	err := mutex.TryLock(context.TODO())
	if err == nil {
		connector.lockMap[name] = mutex
	}

	return err
}

/*This will return the first value. Currently it is good neough.
Consider using versions because this may cause the lost of a put event*/
func (connector EtcdConnector) WatchPutEvents(key string) []byte {

	rch := connector.cli.Watch(context.Background(), key)
	for wresp := range rch {
		for _, ev := range wresp.Events {

			if ev.Type == clientv3.EventTypePut {
				return ev.Kv.Value
			}
		}
	}

	return nil
}

func (connector EtcdConnector) Close() error {

	if connector.session != nil {
		err := connector.session.Close()
		if err != nil {
			return err
		}
	}

	if connector.cli != nil {
		err := connector.cli.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
