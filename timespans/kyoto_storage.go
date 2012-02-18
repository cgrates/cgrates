package timespans

import (
	"strings"
	"github.com/fsouza/gokabinet/kc"
)

type KyotoStorage struct {
	db *kc.DB
}

func NewKyotoStorage(filaName string) (*KyotoStorage, error) {
	ndb, err := kc.Open(filaName, kc.READ)
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

func (ks *KyotoStorage) SetActivationPeriods(key string, aps []*ActivationPeriod){
	result := ""
	for _, ap := range aps {
		result += ap.store() + "\n"
	}
	ks.db.Set(key, result)
}
