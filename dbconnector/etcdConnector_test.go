package dbconnector

import (
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

	err = connector.put(key1, value1)
	if err != nil {
		t.Errorf("Error:%s", err)
		return
	}

	value, err := connector.get(key1)

	valueString := string(value)

	if valueString != value1 {
		t.Errorf("expected value %s received value %s", value1, value)
		return
	}

	err = connector.close()
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

	err = connector.put(key1, value1)
	if err != nil {
		t.Errorf("Error:%s", err)
		return
	}

	err = connector.delete(key1)
	if err != nil {
		t.Errorf("Delete Error:%s", err)
		return
	}

	value, err := connector.get(key1)
	if err != nil {
		t.Errorf("Get Error:%s", err)
	}

	valueString := string(value)

	if valueString != "" {
		t.Errorf("Gets a deleted key's value Error:%s", err)
		return
	}

	err = connector.close()
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
	err = connector.lock(lockName)
	if err != nil {
		t.Errorf("Locking Error:%s", err)
		return
	}

	time.Sleep(5 * time.Second)

	err = connector.unlock(lockName)
	if err != nil {
		t.Errorf("Unlocking Error:%s", err)
		return
	}

	err = connector.close()
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
	err = connector1.tryLock(lockName)
	if err != nil {
		t.Errorf("TryLock Error:%s", err)
		return
	}

	time.Sleep(1 * time.Second)

	err = connector2.tryLock(lockName)
	if err == nil {
		t.Errorf("TryLock Error: it must fail but it did not")
		return
	}

	err = connector1.unlock(lockName)
	if err != nil {
		t.Errorf("Unlock Error:%s", err)
		return
	}

	err = connector2.tryLock(lockName)
	if err != nil {
		t.Errorf("TryLock Error:%s", err)
		return
	}

	time.Sleep(1 * time.Second)

	err = connector2.unlock(lockName)
	if err != nil {
		t.Errorf("Unlock Error:%s", err)
		return
	}

	err = connector1.close()
	if err != nil {
		t.Errorf("close error: %s", err)
	}

	err = connector2.close()
	if err != nil {
		t.Errorf("close error: %s", err)
	}

}
