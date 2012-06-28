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
	// "bytes"
	// "encoding/gob"
	"encoding/json"
	"github.com/simonz05/godis"
)

const (
	ACTION_TIMING_PREFIX = "acttmg"
)

type RedisStorage struct {
	dbNb int
	db   *godis.Client
	//net  bytes.Buffer
}

func NewRedisStorage(address string, db int) (*RedisStorage, error) {
	ndb := godis.New(address, db, "")
	return &RedisStorage{db: ndb, dbNb: db}, nil
}

func (rs *RedisStorage) Close() {
	rs.db.Quit()
}

func (rs *RedisStorage) Flush() error {
	return rs.db.Flushdb()
}

func (rs *RedisStorage) GetActivationPeriodsOrFallback(key string) (aps []*ActivationPeriod, fallbackKey string, err error) {
	//rs.db.Select(rs.dbNb)
	elem, err := rs.db.Get(key)
	if err != nil {
		return
	}
	// rs.net.Reset()
	// rs.net.Write(elem)
	// err = gob.NewDecoder(&rs.net).Decode(&aps)
	err = json.Unmarshal(elem, &aps)
	if err != nil {
		// err = gob.NewDecoder(&rs.net).Decode(&fallbackKey)
		err = json.Unmarshal(elem, &fallbackKey)
	}
	return
}

func (rs *RedisStorage) SetActivationPeriodsOrFallback(key string, aps []*ActivationPeriod, fallbackKey string) (err error) {
	//.db.Select(rs.dbNb)
	var result []byte
	//rs.net.Reset()
	if len(aps) > 0 {
		//gob.NewEncoder(&rs.net).Encode(aps)
		result, err = json.Marshal(aps)
	} else {
		//gob.NewEncoder(&rs.net).Encode(fallbackKey)
		result, err = json.Marshal(fallbackKey)
	}
	//result = rs.net.Bytes()
	return rs.db.Set(key, result)
}

func (rs *RedisStorage) GetDestination(key string) (dest *Destination, err error) {
	//rs.db.Select(rs.dbNb + 1)
	if values, err := rs.db.Get(key); err == nil {
		dest = &Destination{Id: key}
		err = json.Unmarshal(values, dest)
	}
	return
}
func (rs *RedisStorage) SetDestination(dest *Destination) (err error) {
	//rs.db.Select(rs.dbNb + 1)
	result, err := json.Marshal(dest)
	return rs.db.Set(dest.Id, result)
}

func (rs *RedisStorage) GetActions(key string) (as []*Action, err error) {
	//rs.db.Select(rs.dbNb + 2)
	if values, err := rs.db.Get(key); err == nil {
		err = json.Unmarshal(values, as)
	}
	return
}

func (rs *RedisStorage) SetActions(key string, as []*Action) (err error) {
	//rs.db.Select(rs.dbNb + 2)
	result, err := json.Marshal(as)
	return rs.db.Set(key, result)
}

func (rs *RedisStorage) GetUserBalance(key string) (ub *UserBalance, err error) {
	//rs.db.Select(rs.dbNb + 3)
	if values, err := rs.db.Get(key); err == nil {
		ub = &UserBalance{Id: key}
		err = json.Unmarshal(values, ub)
	}
	return
}

func (rs *RedisStorage) SetUserBalance(ub *UserBalance) (err error) {
	//rs.db.Select(rs.dbNb + 3)
	result, err := json.Marshal(ub)
	return rs.db.Set(ub.Id, result)
}

func (rs *RedisStorage) GetActionTimings(key string) (ats []*ActionTiming, err error) {
	if values, err := rs.db.Get(key); err == nil {
		err = json.Unmarshal(values, ats)
	}
	return
}

func (rs *RedisStorage) SetActionTimings(key string, ats []*ActionTiming) (err error) {
	result, err := json.Marshal(ats)
	return rs.db.Set(key, result)
}

func (rs *RedisStorage) GetAllActionTimings() (ats []*ActionTiming, err error) {
	keys, err := rs.db.Keys(ACTION_TIMING_PREFIX + "*")
	if err != nil {
		return
	}
	values, err := rs.db.Mget(keys...)
	if err != nil {
		return
	}
	for _, v := range values.BytesArray() {
		var tempAts []*ActionTiming
		err = json.Unmarshal(v, &tempAts)
		ats = append(ats, tempAts...)
	}
	return
}
