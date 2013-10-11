/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package engine

import (
	"fmt"
	"github.com/cgrates/cgrates/history"
	"github.com/cgrates/cgrates/utils"
	"github.com/garyburd/redigo/redis"
	"time"
)

type RedigoStorage struct {
	dbNb int
	db   redis.Conn
	ms   Marshaler
}

func NewRedigoStorage(address string, db int, pass string) (DataStorage, error) {
	ndb, err := redis.DialTimeout("tcp", address, 5*time.Second, time.Second, time.Second)
	if err != nil {
		return nil, err
	}
	if pass != "" {
		if _, err = ndb.Do("auth", pass); err != nil {
			return nil, err
		}
	}
	if db > 0 {
		if _, err = ndb.Do("select", db); err != nil {
			return nil, err
		}
	}
	return &RedigoStorage{db: ndb, dbNb: db, ms: NewCodecMsgpackMarshaler()}, nil
}

func (rs *RedigoStorage) Close() {
	rs.db.Close()
}

func (rs *RedigoStorage) Flush() (err error) {
	_, err = rs.db.Do("flushdb")
	return
}

func (rs *RedigoStorage) GetRatingProfile(key string) (rp *RatingProfile, err error) {
	var values []byte
	if values, err = redis.Bytes(rs.db.Do("get", RATING_PROFILE_PREFIX+key)); err == nil {
		rp = new(RatingProfile)
		err = rs.ms.Unmarshal(values, rp)
	}
	return
}

func (rs *RedigoStorage) SetRatingProfile(rp *RatingProfile) (err error) {
	result, err := rs.ms.Marshal(rp)
	_, err = rs.db.Do("set", RATING_PROFILE_PREFIX+rp.Id, result)
	if err == nil && historyScribe != nil {
		response := 0
		historyScribe.Record(&history.Record{RATING_PROFILE_PREFIX + rp.Id, rp}, &response)
	}
	return
}

func (rs *RedigoStorage) GetDestination(key string) (dest *Destination, err error) {
	var values []string
	if values, err = redis.Strings(rs.db.Do("smembers", DESTINATION_PREFIX+key)); len(values) > 0 && err == nil {
		dest = &Destination{Id: key, Prefixes: values}
	}
	return
}

func (rs *RedigoStorage) DestinationContainsPrefix(key string, prefix string) (precision int, err error) {
	if _, err := rs.db.Do("sadd", redis.Args{}.Add(TEMP_DESTINATION_PREFIX+prefix).AddFlat(utils.SplitPrefixInterface(prefix))...); err != nil {
		return 0, err
	}
	var values []string
	if values, err = redis.Strings(rs.db.Do("sinter", DESTINATION_PREFIX+key, TEMP_DESTINATION_PREFIX+prefix)); err == nil {
		for _, p := range values {
			if len(p) > precision {
				precision = len(p)
			}
		}
	}
	if _, err := rs.db.Do("del", TEMP_DESTINATION_PREFIX+prefix); err != nil {
		Logger.Err("Error removing temp ")
	}
	return
}

func (rs *RedigoStorage) SetDestination(dest *Destination) (err error) {
	_, err = rs.db.Do("sadd", redis.Args{}.Add(DESTINATION_PREFIX+dest.Id).AddFlat(dest.Prefixes)...)
	if err == nil && historyScribe != nil {
		response := 0
		historyScribe.Record(&history.Record{DESTINATION_PREFIX + dest.Id, dest}, &response)
	}
	return
}

func (rs *RedigoStorage) GetActions(key string) (as Actions, err error) {
	var values []byte
	if values, err = redis.Bytes(rs.db.Do("get", ACTION_PREFIX+key)); err == nil {
		err = rs.ms.Unmarshal(values, &as)
	}
	return
}

func (rs *RedigoStorage) SetActions(key string, as Actions) (err error) {
	result, err := rs.ms.Marshal(&as)
	_, err = rs.db.Do("set", ACTION_PREFIX+key, result)
	return
}

func (rs *RedigoStorage) GetUserBalance(key string) (ub *UserBalance, err error) {
	var values []byte
	if values, err = redis.Bytes(rs.db.Do("get", USER_BALANCE_PREFIX+key)); err == nil {
		ub = &UserBalance{Id: key}
		err = rs.ms.Unmarshal(values, ub)
	}

	return
}

func (rs *RedigoStorage) SetUserBalance(ub *UserBalance) (err error) {
	result, err := rs.ms.Marshal(ub)
	_, err = rs.db.Do("set", USER_BALANCE_PREFIX+ub.Id, result)
	return
}

func (rs *RedigoStorage) GetActionTimings(key string) (ats ActionTimings, err error) {
	var values []byte
	if values, err = redis.Bytes(rs.db.Do("get", ACTION_TIMING_PREFIX+key)); err == nil {
		err = rs.ms.Unmarshal(values, &ats)
	}
	return
}

func (rs *RedigoStorage) SetActionTimings(key string, ats ActionTimings) (err error) {
	if len(ats) == 0 {
		// delete the key
		_, err = rs.db.Do("del", ACTION_TIMING_PREFIX+key)
		return err
	}
	result, err := rs.ms.Marshal(&ats)
	_, err = rs.db.Do("set", ACTION_TIMING_PREFIX+key, result)
	return
}

func (rs *RedigoStorage) GetAllActionTimings() (ats map[string]ActionTimings, err error) {
	reply, err := redis.Values(rs.db.Do("keys", ACTION_TIMING_PREFIX+"*"))
	if err != nil {
		return nil, err
	}
	var keys []string
	for _, x := range reply {
		if v, ok := x.([]byte); ok {
			keys = append(keys, string(v))
		}
	}
	ats = make(map[string]ActionTimings, len(keys))
	for _, key := range keys {
		values, err := redis.Bytes(rs.db.Do("get", key))
		if err != nil {
			continue
		}
		var tempAts ActionTimings
		err = rs.ms.Unmarshal(values, &tempAts)
		ats[key[len(ACTION_TIMING_PREFIX):]] = tempAts
	}

	return
}

func (rs *RedigoStorage) LogCallCost(uuid, source string, cc *CallCost) (err error) {
	var result []byte
	result, err = rs.ms.Marshal(cc)
	if err != nil {
		return
	}
	_, err = rs.db.Do("set", LOG_CALL_COST_PREFIX+source+"_"+uuid, result)
	return
}

func (rs *RedigoStorage) GetCallCostLog(uuid, source string) (cc *CallCost, err error) {
	var values []byte
	if values, err = redis.Bytes(rs.db.Do("get", LOG_CALL_COST_PREFIX+source+"_"+uuid)); err == nil {
		err = rs.ms.Unmarshal(values, cc)
	}
	return
}

func (rs *RedigoStorage) LogActionTrigger(ubId, source string, at *ActionTrigger, as Actions) (err error) {
	mat, err := rs.ms.Marshal(at)
	if err != nil {
		return
	}
	mas, err := rs.ms.Marshal(as)
	if err != nil {
		return
	}
	_, err = rs.db.Do("set", LOG_ACTION_TRIGGER_PREFIX+source+"_"+time.Now().Format(time.RFC3339Nano), []byte(fmt.Sprintf("%s*%s*%s", ubId, string(mat), string(mas))))
	return
}

func (rs *RedigoStorage) LogActionTiming(source string, at *ActionTiming, as Actions) (err error) {
	mat, err := rs.ms.Marshal(at)
	if err != nil {
		return
	}
	mas, err := rs.ms.Marshal(as)
	if err != nil {
		return
	}
	_, err = rs.db.Do("set", LOG_ACTION_TIMMING_PREFIX+source+"_"+time.Now().Format(time.RFC3339Nano), []byte(fmt.Sprintf("%s*%s", string(mat), string(mas))))
	return
}

func (rs *RedigoStorage) LogError(uuid, source, errstr string) (err error) {
	_, err = rs.db.Do("set", LOG_ERR+source+"_"+uuid, errstr)
	return
}
