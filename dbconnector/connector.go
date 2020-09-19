package dbconnector

/*dbConnector interface defines an
interface for key value store to use in danelion*/
type dbConnector interface {
	get(key string) ([]byte, error)
	put(key string, value string) error
	del(key string) error

	lock(name string) error
	unlock(name string) error
	tryLock(name string) error

	watchPutEvents(key string, function onPutEvent) error

	close() error
}

type onPutEvent func(string, string)
