package dbconnector

import (
	"sync"
	"testing"
	"time"
)

func TestPutGet(t *testing.T) {

	connector, err := CreateEtcdConnector("127.0.0.1:2379")
	if err != nil {
		t.Errorf("Error:%s", err)
		return
	}

	key1 := "key-1"
	value1 := "value-1"

	err = connector.Put(key1, value1)
	if err != nil {
		t.Errorf("Error:%s", err)
		return
	}

	value, err := connector.Get(key1)

	valueString := string(value)

	if valueString != value1 {
		t.Errorf("expected value %s received value %s", value1, value)
		return
	}

	err = connector.Close()
	if err != nil {
		t.Errorf("close error: %s", err)
	}

}

func TestPutDeleteGet(t *testing.T) {
	connector, err := CreateEtcdConnector("127.0.0.1:2379")
	if err != nil {
		t.Errorf("Error:%s", err)
		return
	}

	key1 := "key-2"
	value1 := "value-to-delete"

	err = connector.Put(key1, value1)
	if err != nil {
		t.Errorf("Error:%s", err)
		return
	}

	err = connector.Delete(key1)
	if err != nil {
		t.Errorf("Delete Error:%s", err)
		return
	}

	value, err := connector.Get(key1)
	if err != nil {
		t.Errorf("Get Error:%s", err)
	}

	valueString := string(value)

	if valueString != "" {
		t.Errorf("Gets a deleted key's value Error:%s", err)
		return
	}

	err = connector.Close()
	if err != nil {
		t.Errorf("close error: %s", err)
	}

}

func TestLockUnlock(t *testing.T) {

	connector, err := CreateEtcdConnector("127.0.0.1:2379")
	if err != nil {
		t.Errorf("Error:%s", err)
		return
	}

	lockName := "ds-lock"
	err = connector.Lock(lockName)
	if err != nil {
		t.Errorf("Locking Error:%s", err)
		return
	}

	time.Sleep(5 * time.Second)

	err = connector.Unlock(lockName)
	if err != nil {
		t.Errorf("Unlocking Error:%s", err)
		return
	}

	err = connector.Close()
	if err != nil {
		t.Errorf("close error: %s", err)
	}

}

func TestTryLock(t *testing.T) {

	connector1, err := CreateEtcdConnector("127.0.0.1:2379")
	if err != nil {
		t.Errorf("Error:%s", err)
		return
	}

	connector2, err := CreateEtcdConnector("127.0.0.1:2379")
	if err != nil {
		t.Errorf("Error:%s", err)
		return
	}

	lockName := "hello-there"
	err = connector1.TryLock(lockName)
	if err != nil {
		t.Errorf("TryLock Error:%s", err)
		return
	}

	time.Sleep(1 * time.Second)

	err = connector2.TryLock(lockName)
	if err == nil {
		t.Errorf("TryLock Error: it must fail but it did not")
		return
	}

	err = connector1.Unlock(lockName)
	if err != nil {
		t.Errorf("Unlock Error:%s", err)
		return
	}

	err = connector2.TryLock(lockName)
	if err != nil {
		t.Errorf("TryLock Error:%s", err)
		return
	}

	time.Sleep(1 * time.Second)

	err = connector2.Unlock(lockName)
	if err != nil {
		t.Errorf("Unlock Error:%s", err)
		return
	}

	err = connector1.Close()
	if err != nil {
		t.Errorf("close error: %s", err)
	}

	err = connector2.Close()
	if err != nil {
		t.Errorf("close error: %s", err)
	}

}

func TestWatchPutEvents(t *testing.T) {

	connector1, err := CreateEtcdConnector("127.0.0.1:2379")
	if err != nil {
		t.Errorf("Error:%s", err)
		return
	}

	watchKeyName := "key-123"
	value := "hello-world"

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {

		result := connector1.WatchPutEvents(watchKeyName)
		resultString := string(result)

		if resultString != value {
			t.Errorf("Error: Could not read same value. Expected [%s], received [%s]", value, resultString)
		}

		wg.Done()

	}()

	connector2, err := CreateEtcdConnector("127.0.0.1:2379")
	if err != nil {
		t.Errorf("Error:%s", err)
		return
	}

	err = connector2.Put(watchKeyName, value)
	if err != nil {
		t.Errorf("Error:%s", err)
	}

	wg.Wait()

	err = connector1.Close()
	if err != nil {
		t.Errorf("close error: %s", err)
	}

	err = connector2.Close()
	if err != nil {
		t.Errorf("close error: %s", err)
	}

}

func TestInterface(t *testing.T) {
	connector, err := CreateEtcdConnector("127.0.0.1:2379")
	if err != nil {
		t.Errorf("Error:%s", err)
		return
	}

	doSomethingWithConnection(connector, t)

}

func doSomethingWithConnection(connector DBConnector, t *testing.T) {

}
