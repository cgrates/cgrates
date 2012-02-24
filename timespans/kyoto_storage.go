package timespans

import (
	"github.com/fsouza/gokabinet/kc"
	//"log"
	"strings"
)

type KyotoStorage struct {
	db *kc.DB
}

func NewKyotoStorage(filaName string) (*KyotoStorage, error) {
	ndb, err := kc.Open(filaName, kc.WRITE)
	return &KyotoStorage{db: ndb}, err
}

func (ks *KyotoStorage) Close() {
	ks.db.Close()
}

func (ks *KyotoStorage) GetActivationPeriods(key string) (aps []*ActivationPeriod, err error) {
	values, err := ks.db.Get(key)

	if err == nil {
		for _, ap_string := range strings.Split(values, "\n") {
			if len(ap_string) > 0 {
				ap := &ActivationPeriod{}
				ap.restore(ap_string)
				aps = append(aps, ap)
			}
		}
	}
	return aps, err
}

func (ks *KyotoStorage) SetActivationPeriods(key string, aps []*ActivationPeriod) error {
	result := ""
	for _, ap := range aps {
		result += ap.store() + "\n"
	}
	return ks.db.Set(key, result)
}

func (ks *KyotoStorage) GetDestination(key string) (dest *Destination, err error) {
	if values, err := ks.db.Get(key); err == nil {
		dest = &Destination{Id: key}
		dest.restore(values)
	}
	return
}

func (ks *KyotoStorage) SetDestination(dest *Destination) error {
	return ks.db.Set(dest.Id, dest.store())
}

func (ks *KyotoStorage) GetTariffPlan(key string) (tp *TariffPlan, err error) {
	if values, err := ks.db.Get(key); err == nil {
		tp = &TariffPlan{Id: key}
		tp.restore(values)
	}
	return
}

func (ks *KyotoStorage) SetTariffPlan(tp *TariffPlan) error {
	return ks.db.Set(tp.Id, tp.store())
}

func (ks *KyotoStorage) GetUserBudget(key string) (ub *UserBudget, err error) {
	if values, err := ks.db.Get(key); err == nil {
		ub = &UserBudget{Id: key}
		ub.restore(values)
	}
	return
}

func (ks *KyotoStorage) SetUserBudget(ub *UserBudget) error {
	return ks.db.Set(ub.Id, ub.store())
}
