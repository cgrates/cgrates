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

func (rs *RedisStorage) GetActivationPeriodsOrFallback(key string) (aps []*ActivationPeriod, fallbackKey string, err error) {
	//rs.db.Select(rs.dbNb)
	elem, err := rs.db.Get(key)
	if err != nil {
		return
	}
	valuesString := elem.String()
	values := strings.Split(valuesString, "\n")
	if len(values) > 1 {
		for _, ap_string := range values {
			if len(ap_string) > 0 {
				ap := &ActivationPeriod{}
				ap.restore(ap_string)
				aps = append(aps, ap)
			}
		}
	} else { // fallback case
		fallbackKey = valuesString
	}
	return
}

func (rs *RedisStorage) SetActivationPeriodsOrFallback(key string, aps []*ActivationPeriod, fallbackKey string) error {
	//.db.Select(rs.dbNb)
	result := ""
	if len(aps) > 0 {
		for _, ap := range aps {
			result += ap.store() + "\n"
		}
	} else {
		result = fallbackKey
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

func (rs *RedisStorage) GetUserBalance(key string) (ub *UserBalance, err error) {
	//rs.db.Select(rs.dbNb + 3)
	if values, err := rs.db.Get(key); err == nil {
		ub = &UserBalance{Id: key}
		ub.restore(values.String())
	}
	return
}

func (rs *RedisStorage) SetUserBalance(ub *UserBalance) error {
	//rs.db.Select(rs.dbNb + 3)
	return rs.db.Set(ub.Id, ub.store())
}
