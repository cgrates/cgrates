package timespans

/*
Interface for storage providers.
*/
type StorageGetter interface {
	Close()
	Get(key string) (string, error)
}
