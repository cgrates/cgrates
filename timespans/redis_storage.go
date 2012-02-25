/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/
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

func (rs *RedisStorage) SetActivationPeriods(key string, aps []*ActivationPeriod) error {
	//.db.Select(rs.dbNb)
	result := ""
	for _, ap := range aps {
		result += ap.store() + "\n"
	}
	return rs.db.Set(key, result)
}

func (rs *RedisStorage) GetDestination(key string) (dest *Destination, err error) {
	//rs.db.Select(rs.dbNb + 1)
	if values, err := rs.db.Get(key); err == nil {
		dest = &Destination{Id: key}
		dest.restore(values.String())
	}
	return
}

func (rs *RedisStorage) SetDestination(dest *Destination) error {
	//rs.db.Select(rs.dbNb + 1)
	return rs.db.Set(dest.Id, dest.store())
}

func (rs *RedisStorage) GetTariffPlan(key string) (tp *TariffPlan, err error) {
	//rs.db.Select(rs.dbNb + 2)
	if values, err := rs.db.Get(key); err == nil {
		tp = &TariffPlan{Id: key}
		tp.restore(values.String())
	}
	return
}

func (rs *RedisStorage) SetTariffPlan(tp *TariffPlan) error {
	//rs.db.Select(rs.dbNb + 2)
	return rs.db.Set(tp.Id, tp.store())
}

func (rs *RedisStorage) GetUserBudget(key string) (ub *UserBudget, err error) {
	//rs.db.Select(rs.dbNb + 3)
	if values, err := rs.db.Get(key); err == nil {
		ub = &UserBudget{Id: key}
		ub.restore(values.String())
	}
	return
}

func (rs *RedisStorage) SetUserBudget(ub *UserBudget) error {
	//rs.db.Select(rs.dbNb + 3)
	return rs.db.Set(ub.Id, ub.store())
}
