package timespans

/*
Interface for storage providers.
*/
type StorageGetter interface {
	Close()
	GetActivationPeriods(string) ([]*ActivationPeriod, error)
	SetActivationPeriods(string, []*ActivationPeriod)
	GetDestination(string) (*Destination, error)
	SetDestination(*Destination)
	GetTariffPlan(string) (*TariffPlan, error)
	SetTariffPlan(*TariffPlan)
}
