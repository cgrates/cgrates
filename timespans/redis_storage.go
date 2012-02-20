package timespans

import (
	"github.com/simonz05/godis"
	"strings"
)

type RedisStorage struct {
	db *godis.Client
}

func NewRedisStorage(address string, db int) (*RedisStorage, error) {
	ndb := godis.New(address, db, "")
	return &RedisStorage{db: ndb}, nil
}

func (rs *RedisStorage) Close() {
	rs.db.Quit()
}

func (rs *RedisStorage) GetActivationPeriods(key string) (aps []*ActivationPeriod, err error) {
	elem, err := rs.db.Get(key)
	values := elem.String()
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

func (rs *RedisStorage) SetActivationPeriods(key string, aps []*ActivationPeriod) {
	result := ""
	for _, ap := range aps {
		result += ap.store() + "\n"
	}
	rs.db.Set(key, result)
}

func (rs *RedisStorage) GetDestination(key string) (dest *Destination, err error) {
	values, err := rs.db.Get(key)
	dest = &Destination{Id:key}
	dest.restore(values.String())
	return
}

func (rs *RedisStorage) SetDestination(dest *Destination) {
	rs.db.Set(dest.Id, dest.store())
}
