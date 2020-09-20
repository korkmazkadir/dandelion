package dbconnector

/*dbConnector interface defines an
interface for key value store to use in danelion*/
type dbConnector interface {
	get(key string) ([]byte, error)
	put(key string, value string) error
	delete(key string) error

	lock(name string) error
	tryLock(name string) error
	unlock(name string) error

	watchPutEvents(key string) []byte

	close() error
}
