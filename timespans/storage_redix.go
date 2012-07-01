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
	"encoding/json"
	// "encoding/gob"
	// "bytes"
	"log"
)

type RedixStorage struct {
	db *redis.Client
	//net bytes.Buffer
}

func NewRedixStorage(address string, db int) (*RedixStorage, error) {
	ndb, err := redis.NewClient(redis.Configuration{Address: address, Database: db})
	if err != nil {
		log.Fatalf("Could not connect to redis server: %v", err)
	}
	return &RedixStorage{db: ndb}, nil
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
	//rs.net.Reset()
	//rs.net.Write(elem)
	//err = gob.NewDecoder(&rs.net).Decode(&aps)
	err = json.Unmarshal(elem, &aps)
	if err != nil {
		//err = gob.NewDecoder(&rs.net).Decode(&fallbackKey)
		err = json.Unmarshal(elem, &fallbackKey)
	}
	return
}

func (rs *RedixStorage) SetActivationPeriodsOrFallback(key string, aps []*ActivationPeriod, fallbackKey string) (err error) {
	var result []byte
	//rs.net.Reset()
	if len(aps) > 0 {
		//gob.NewEncoder(&rs.net).Encode(aps)
		result, err = json.Marshal(&aps)
	} else {
		//gob.NewEncoder(&rs.net).Encode(fallbackKey)
		result, err = json.Marshal(fallbackKey)
	}
	//result = rs.net.Bytes()
	return rs.db.Set(key, result).Err
}

func (rs *RedixStorage) GetDestination(key string) (dest *Destination, err error) {
	if values, err := rs.db.Get(key).Bytes(); err == nil {
		dest = &Destination{Id: key}
		err = json.Unmarshal(values, dest)
	}
	return
}
func (rs *RedixStorage) SetDestination(dest *Destination) (err error) {
	result, err := json.Marshal(dest)
	return rs.db.Set(dest.Id, result).Err
}

func (rs *RedixStorage) GetActions(key string) (as []*Action, err error) {
	if values, err := rs.db.Get(key).Bytes(); err == nil {
		err = json.Unmarshal(values, &as)
	}
	return
}

func (rs *RedixStorage) SetActions(key string, as []*Action) (err error) {
	result, err := json.Marshal(as)
	return rs.db.Set(key, result).Err
}

func (rs *RedixStorage) GetUserBalance(key string) (ub *UserBalance, err error) {
	if values, err := rs.db.Get(key).Bytes(); err == nil {
		ub = &UserBalance{Id: key}
		err = json.Unmarshal(values, ub)
	}
	return
}

func (rs *RedixStorage) SetUserBalance(ub *UserBalance) (err error) {
	result, err := json.Marshal(ub)
	return rs.db.Set(ub.Id, result).Err
}

func (rs *RedixStorage) GetActionTimings(key string) (ats []*ActionTiming, err error) {
	if values, err := rs.db.Get(key).Bytes(); err == nil {
		err = json.Unmarshal(values, ats)
	}
	return
}

func (rs *RedixStorage) SetActionTimings(key string, ats []*ActionTiming) (err error) {
	result, err := json.Marshal(ats)
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
		err = json.Unmarshal([]byte(v), &ats)
		ats = append(ats, tempAts...)
	}
	return
}
