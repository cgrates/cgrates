package timespans

/*
Interface for storage providers.
*/
type StorageGetter interface {
	Close()
	GetActivationPeriods(string) ([]*ActivationPeriod, error)
	SetActivationPeriods(string, []*ActivationPeriod) error
	GetDestination(string) (*Destination, error)
	SetDestination(*Destination) error
	GetTariffPlan(string) (*TariffPlan, error)
	SetTariffPlan(*TariffPlan) error
	GetUserBudget(string) (*UserBudget, error)
	SetUserBudget(*UserBudget) error
}
