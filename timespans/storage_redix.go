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
	"github.com/fzzbt/radix/redis"
	"log"
)

type RedixStorage struct {
	db *redis.Client
	ms Marshaler
}

func NewRedixStorage(address string, db int) (*RedixStorage, error) {
	ndb, err := redis.NewClient(redis.Configuration{Address: address, Database: db})
	if err != nil {
		log.Fatalf("Could not connect to redis server: %v", err)
	}
	ms := &MyMarshaler{}
	return &RedixStorage{db: ndb, ms: ms}, nil
}

func (rs *RedixStorage) Close() {
	rs.db.Close()
}

func (rs *RedixStorage) Flush() error {
	rs.db.Flushdb()
	return nil
}

func (rs *RedixStorage) GetActivationPeriodsOrFallback(key string) (aps []*ActivationPeriod, fallbackKey string, err error) {
	elem, err := rs.db.Get(key).Bytes()
	if err != nil {
		return
	}
	err = rs.ms.Unmarshal(elem, &aps)
	if err != nil {
		err = rs.ms.Unmarshal(elem, &fallbackKey)
	}
	return
}

func (rs *RedixStorage) SetActivationPeriodsOrFallback(key string, aps []*ActivationPeriod, fallbackKey string) (err error) {
	var result []byte
	if len(aps) > 0 {
		result, err = rs.ms.Marshal(&aps)
	} else {
		result, err = rs.ms.Marshal(fallbackKey)
	}
	return rs.db.Set(key, result).Err
}

func (rs *RedixStorage) GetDestination(key string) (dest *Destination, err error) {
	if values, err := rs.db.Get(key).Bytes(); err == nil {
		dest = &Destination{Id: key}
		err = rs.ms.Unmarshal(values, dest)
	}
	return
}
func (rs *RedixStorage) SetDestination(dest *Destination) (err error) {
	result, err := rs.ms.Marshal(dest)
	return rs.db.Set(dest.Id, result).Err
}

func (rs *RedixStorage) GetActions(key string) (as []*Action, err error) {
	if values, err := rs.db.Get(key).Bytes(); err == nil {
		err = rs.ms.Unmarshal(values, &as)
	}
	return
}

func (rs *RedixStorage) SetActions(key string, as []*Action) (err error) {
	result, err := rs.ms.Marshal(as)
	return rs.db.Set(key, result).Err
}

func (rs *RedixStorage) GetUserBalance(key string) (ub *UserBalance, err error) {
	if values, err := rs.db.Get(key).Bytes(); err == nil {
		ub = &UserBalance{Id: key}
		err = rs.ms.Unmarshal(values, ub)
	}
	return
}

func (rs *RedixStorage) SetUserBalance(ub *UserBalance) (err error) {
	result, err := rs.ms.Marshal(ub)
	return rs.db.Set(ub.Id, result).Err
}

func (rs *RedixStorage) GetActionTimings(key string) (ats []*ActionTiming, err error) {
	if values, err := rs.db.Get(key).Bytes(); err == nil {
		err = rs.ms.Unmarshal(values, ats)
	}
	return
}

func (rs *RedixStorage) SetActionTimings(key string, ats []*ActionTiming) (err error) {
	result, err := rs.ms.Marshal(ats)
	return rs.db.Set(key, result).Err
}

func (rs *RedixStorage) GetAllActionTimings() (ats []*ActionTiming, err error) {
	keys, err := rs.db.Keys(ACTION_TIMING_PREFIX + "*").List()
	if err != nil {
		return
	}
	values, err := rs.db.Mget(keys).List()
	if err != nil {
		return
	}
	for _, v := range values {
		var tempAts []*ActionTiming
		err = rs.ms.Unmarshal([]byte(v), &ats)
		ats = append(ats, tempAts...)
	}
	return
}
