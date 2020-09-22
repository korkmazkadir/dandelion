package dbconnector

/*DBConnector interface defines an
interface for key value store to use in danelion*/
type DBConnector interface {
	Get(key string) ([]byte, error)
	GetWithPrefix(prefix string) ([][]byte, error)
	Put(key string, value string) error
	Delete(key string) error

	Lock(name string) error
	TryLock(name string) error
	Unlock(name string) error

	WatchPutEvents(key string) []byte

	Close() error
}
