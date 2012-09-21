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
	"fmt"
	"github.com/simonz05/godis/redis"
	"time"
)

type RedisStorage struct {
	dbNb int
	db   *redis.Client
	ms   Marshaler
}

func NewRedisStorage(address string, db int, pass string) (DataStorage, error) {
	if address != "" {
		address = "tcp:" + address
	}
	ndb := redis.New(address, db, pass)
	ms := new(MyMarshaler)
	return &RedisStorage{db: ndb, dbNb: db, ms: ms}, nil
}

func (rs *RedisStorage) Close() {
	rs.db.Quit()
}

func (rs *RedisStorage) Flush() error {
	return rs.db.Flushdb()
}

func (rs *RedisStorage) GetRatingProfile(key string) (rp *RatingProfile, err error) {
	if values, err := rs.db.Get(RATING_PROFILE_PREFIX + key); err == nil {
		rp = new(RatingProfile)
		err = rs.ms.Unmarshal(values, rp)
	} else {
		return nil, err
	}
	return
}

func (rs *RedisStorage) SetRatingProfile(rp *RatingProfile) (err error) {
	result, err := rs.ms.Marshal(rp)
	return rs.db.Set(RATING_PROFILE_PREFIX+rp.Id, result)
}

func (rs *RedisStorage) GetDestination(key string) (dest *Destination, err error) {
	if values, err := rs.db.Get(DESTINATION_PREFIX + key); err == nil {
		dest = &Destination{Id: key}
		err = rs.ms.Unmarshal(values, dest)
	} else {
		return nil, err
	}
	return
}

func (rs *RedisStorage) SetDestination(dest *Destination) (err error) {
	result, err := rs.ms.Marshal(dest)
	return rs.db.Set(DESTINATION_PREFIX+dest.Id, result)
}

func (rs *RedisStorage) GetActions(key string) (as []*Action, err error) {
	if values, err := rs.db.Get(ACTION_PREFIX + key); err == nil {
		err = rs.ms.Unmarshal(values, &as)
	} else {
		return nil, err
	}
	return
}

func (rs *RedisStorage) SetActions(key string, as []*Action) (err error) {
	result, err := rs.ms.Marshal(as)
	return rs.db.Set(ACTION_PREFIX+key, result)
}

func (rs *RedisStorage) GetUserBalance(key string) (ub *UserBalance, err error) {
	if values, err := rs.db.Get(USER_BALANCE_PREFIX + key); err == nil {
		ub = &UserBalance{Id: key}
		err = rs.ms.Unmarshal(values, ub)
	} else {
		return nil, err
	}
	return
}

func (rs *RedisStorage) SetUserBalance(ub *UserBalance) (err error) {
	result, err := rs.ms.Marshal(ub)
	return rs.db.Set(USER_BALANCE_PREFIX+ub.Id, result)
}

func (rs *RedisStorage) GetActionTimings(key string) (ats []*ActionTiming, err error) {
	if values, err := rs.db.Get(ACTION_TIMING_PREFIX + key); err == nil {
		err = rs.ms.Unmarshal(values, &ats)
	} else {
		return nil, err
	}
	return
}

func (rs *RedisStorage) SetActionTimings(key string, ats []*ActionTiming) (err error) {
	if len(ats) == 0 {
		// delete the key
		_, err = rs.db.Del(ACTION_TIMING_PREFIX + key)
		return
	}
	result, err := rs.ms.Marshal(ats)
	return rs.db.Set(ACTION_TIMING_PREFIX+key, result)
}

func (rs *RedisStorage) GetAllActionTimings() (ats map[string][]*ActionTiming, err error) {
	keys, err := rs.db.Keys(ACTION_TIMING_PREFIX + "*")
	if err != nil {
		return
	}
	ats = make(map[string][]*ActionTiming, len(keys))
	for _, key := range keys {
		values, err := rs.db.Get(key)
		if err != nil {
			continue
		}
		var tempAts []*ActionTiming
		err = rs.ms.Unmarshal(values, &tempAts)
		ats[key[len(ACTION_TIMING_PREFIX):]] = tempAts
	}

	return
}

func (rs *RedisStorage) LogCallCost(uuid string, cc *CallCost) (err error) {
	result, err := rs.ms.Marshal(cc)
	if err != nil {
		return
	}
	return rs.db.Set(CALL_COST_LOG_PREFIX+uuid, result)
}

func (rs *RedisStorage) GetCallCostLog(uuid string) (cc *CallCost, err error) {
	if values, err := rs.db.Get(uuid); err == nil {
		err = rs.ms.Unmarshal(values, cc)
	} else {
		return nil, err
	}
	return
}

func (rs *RedisStorage) LogActionTrigger(ubId string, at *ActionTrigger, as []*Action) (err error) {
	mat, err := rs.ms.Marshal(at)
	if err != nil {
		return
	}
	mas, err := rs.ms.Marshal(as)
	if err != nil {
		return
	}
	rs.db.Set(LOG_PREFIX+time.Now().Format(time.RFC3339Nano), []byte(fmt.Sprintf("%s*%s*%s", ubId, string(mat), string(mas))))
	return
}

func (rs *RedisStorage) LogActionTiming(at *ActionTiming, as []*Action) (err error) {
	mat, err := rs.ms.Marshal(at)
	if err != nil {
		return
	}
	mas, err := rs.ms.Marshal(as)
	if err != nil {
		return
	}
	rs.db.Set(LOG_PREFIX+time.Now().Format(time.RFC3339Nano), []byte(fmt.Sprintf("%s*%s", string(mat), string(mas))))
	return
}
