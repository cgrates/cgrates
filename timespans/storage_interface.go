package timespans

/*
Interface for storage providers.
*/
type StorageGetter interface {
	Close()
	GetActivationPeriods(key string) ([]*ActivationPeriod, error)
	SetActivationPeriods(key string, aps []*ActivationPeriod)
}
