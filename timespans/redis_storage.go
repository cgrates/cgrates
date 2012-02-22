package timespans

import (
	"github.com/simonz05/godis"
	"strings"
)

type RedisStorage struct {
	dbNb int
	db   *godis.Client
}

func NewRedisStorage(address string, db int) (*RedisStorage, error) {
	ndb := godis.New(address, db, "")
	return &RedisStorage{db: ndb, dbNb: db}, nil
}

func (rs *RedisStorage) Close() {
	rs.db.Quit()
}

func (rs *RedisStorage) GetActivationPeriods(key string) (aps []*ActivationPeriod, err error) {
	//rs.db.Select(rs.dbNb)
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
	//.db.Select(rs.dbNb)
	result := ""
	for _, ap := range aps {
		result += ap.store() + "\n"
	}
	rs.db.Set(key, result)
}

func (rs *RedisStorage) GetDestination(key string) (dest *Destination, err error) {
	//rs.db.Select(rs.dbNb + 1)
	if values, err := rs.db.Get(key); err == nil {
		dest = &Destination{Id: key}
		dest.restore(values.String())
	}
	return
}

func (rs *RedisStorage) SetDestination(dest *Destination) {
	//rs.db.Select(rs.dbNb + 1)
	rs.db.Set(dest.Id, dest.store())
}

func (rs *RedisStorage) GetTariffPlan(key string) (tp *TariffPlan, err error) {
	//rs.db.Select(rs.dbNb + 2)
	if values, err := rs.db.Get(key); err == nil {
		tp = &TariffPlan{Id: key}
		tp.restore(values.String())
	}
	return
}

func (rs *RedisStorage) SetTariffPlan(tp *TariffPlan) {
	//rs.db.Select(rs.dbNb + 2)
	rs.db.Set(tp.Id, tp.store())
}

func (rs *RedisStorage) GetUserBudget(key string) (ub *UserBudget, err error) {
	//rs.db.Select(rs.dbNb + 3)
	if values, err := rs.db.Get(key); err == nil {
		ub = &UserBudget{Id: key}
		ub.restore(values.String())
	}
	return
}

func (rs *RedisStorage) SetUserBudget(ub *UserBudget) {
	//rs.db.Select(rs.dbNb + 3)
	rs.db.Set(ub.Id, ub.store())
}
