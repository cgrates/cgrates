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
	"github.com/fzzy/radix/redis"
	"time"
)

type RadixStorage struct {
	dbNb int
	db   *redis.Client
	ms   Marshaler
}

func NewRadixStorage(address string, db int, pass string) (DataStorage, error) {
	ndb, err := redis.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	if pass != "" {
		if r := ndb.Cmd("auth", pass); r.Err != nil {
			return nil, r.Err
		}
	}
	if db > 0 {
		if r := ndb.Cmd("select", db); r.Err != nil {
			return nil, r.Err
		}
	}
	return &RadixStorage{db: ndb, dbNb: db, ms: NewCodecMsgpackMarshaler()}, nil
}

func (rs *RadixStorage) Close() {
	rs.db.Close()
}

func (rs *RadixStorage) Flush() (err error) {
	if r := rs.db.Cmd("flushdb"); r.Err != nil {
		return r.Err
	}
	return
}

func (rs *RadixStorage) GetRatingProfile(key string) (rp *RatingProfile, err error) {
	if values, err := rs.db.Cmd("get", RATING_PROFILE_PREFIX+key).Bytes(); err == nil {
		rp = new(RatingProfile)
		err = rs.ms.Unmarshal(values, rp)
	} else {
		return nil, err
	}
	return
}

func (rs *RadixStorage) SetRatingProfile(rp *RatingProfile) (err error) {
	result, err := rs.ms.Marshal(rp)
	if r := rs.db.Cmd("set", RATING_PROFILE_PREFIX+rp.Id, string(result)); r.Err != nil {
		return r.Err
	}
	if err == nil && historyScribe != nil {
		response := 0
		historyScribe.Record(&history.Record{RATING_PROFILE_PREFIX + rp.Id, rp}, &response)
	}
	return
}

func (rs *RadixStorage) GetDestination(key string) (dest *Destination, err error) {
	var values []string
	if values, err = rs.db.Cmd("hkeys", DESTINATION_PREFIX+key).List(); len(values) > 0 && err == nil {
		dest = &Destination{Id: key, Prefixes: values}
	}
	return
}

func (rs *RadixStorage) DestinationContainsPrefix(key string, prefix string) (precision int, err error) {
	var values []string
	var pfs []interface{}
	pfs = append(pfs, DESTINATION_PREFIX+key)
	pfs = append(pfs, utils.SplitPrefixInterface(prefix)...)
	if values, err = rs.db.Cmd("hmget", pfs...).List(); err == nil {
		for i, p := range values {
			if p != "" {
				return len(prefix) - i, nil
			}
		}
	}
	return
}

func (rs *RadixStorage) SetDestination(dest *Destination) (err error) {
	newPrefixes := make(map[string]string, len(dest.Prefixes))
	for _, p := range dest.Prefixes {
		newPrefixes[p] = "*"
	}
	if r := rs.db.Cmd("hmset", DESTINATION_PREFIX+dest.Id, newPrefixes); r.Err != nil {
		return r.Err
	}
	if err == nil && historyScribe != nil {
		response := 0
		historyScribe.Record(&history.Record{DESTINATION_PREFIX + dest.Id, dest}, &response)
	}
	return
}

func (rs *RadixStorage) GetActions(key string) (as Actions, err error) {
	var values []byte
	if values, err = rs.db.Cmd("get", ACTION_PREFIX+key).Bytes(); err == nil {
		err = rs.ms.Unmarshal(values, &as)
	}
	return
}

func (rs *RadixStorage) SetActions(key string, as Actions) (err error) {
	result, err := rs.ms.Marshal(&as)
	if r := rs.db.Cmd("set", ACTION_PREFIX+key, string(result)); r.Err != nil {
		return r.Err
	}
	return
}

func (rs *RadixStorage) GetUserBalance(key string) (ub *UserBalance, err error) {
	if values, err := rs.db.Cmd("get", USER_BALANCE_PREFIX+key).Bytes(); err == nil {
		ub = &UserBalance{Id: key}
		err = rs.ms.Unmarshal(values, ub)
	} else {
		return nil, err
	}
	return
}

func (rs *RadixStorage) SetUserBalance(ub *UserBalance) (err error) {
	result, err := rs.ms.Marshal(ub)
	r := rs.db.Cmd("set", USER_BALANCE_PREFIX+ub.Id, string(result))
	if r.Err != nil {
		return r.Err
	}
	return
}

func (rs *RadixStorage) GetActionTimings(key string) (ats ActionTimings, err error) {
	if values, err := rs.db.Cmd("get", ACTION_TIMING_PREFIX+key).Bytes(); err == nil {
		err = rs.ms.Unmarshal(values, &ats)
	} else {
		return nil, err
	}
	return
}

func (rs *RadixStorage) SetActionTimings(key string, ats ActionTimings) (err error) {
	if len(ats) == 0 {
		// delete the key
		r := rs.db.Cmd("del", ACTION_TIMING_PREFIX+key)
		if r.Err != nil {
			return r.Err
		}
		return
	}
	result, err := rs.ms.Marshal(&ats)
	if r := rs.db.Cmd("set", ACTION_TIMING_PREFIX+key, string(result)); r.Err != nil {
		return r.Err
	}
	return
}

func (rs *RadixStorage) GetAllActionTimings() (ats map[string]ActionTimings, err error) {
	keys, err := rs.db.Cmd("keys", ACTION_TIMING_PREFIX+"*").List()
	if err != nil {
		return
	}
	ats = make(map[string]ActionTimings, len(keys))
	for _, key := range keys {
		values, err := rs.db.Cmd("get", key).Bytes()
		if err != nil {
			continue
		}
		var tempAts ActionTimings
		err = rs.ms.Unmarshal(values, &tempAts)
		ats[key[len(ACTION_TIMING_PREFIX):]] = tempAts
	}

	return
}

func (rs *RadixStorage) LogCallCost(uuid, source string, cc *CallCost) (err error) {
	result, err := rs.ms.Marshal(cc)
	if err != nil {
		return
	}
	if r := rs.db.Cmd("set", LOG_CALL_COST_PREFIX+source+"_"+uuid, string(result)); r.Err != nil {
		return r.Err
	}
	return
}

func (rs *RadixStorage) GetCallCostLog(uuid, source string) (cc *CallCost, err error) {
	if values, err := rs.db.Cmd("get", LOG_CALL_COST_PREFIX+source+"_"+uuid).Bytes(); err == nil {
		err = rs.ms.Unmarshal(values, cc)
	} else {
		return nil, err
	}
	return
}

func (rs *RadixStorage) LogActionTrigger(ubId, source string, at *ActionTrigger, as Actions) (err error) {
	mat, err := rs.ms.Marshal(at)
	if err != nil {
		return
	}
	mas, err := rs.ms.Marshal(&as)
	if err != nil {
		return
	}
	rs.db.Cmd("set", LOG_ACTION_TRIGGER_PREFIX+source+"_"+time.Now().Format(time.RFC3339Nano), fmt.Sprintf("%s*%s*%s", ubId, string(mat), string(mas)))
	return
}

func (rs *RadixStorage) LogActionTiming(source string, at *ActionTiming, as Actions) (err error) {
	mat, err := rs.ms.Marshal(at)
	if err != nil {
		return
	}
	mas, err := rs.ms.Marshal(&as)
	if err != nil {
		return
	}
	rs.db.Cmd("set", LOG_ACTION_TIMMING_PREFIX+source+"_"+time.Now().Format(time.RFC3339Nano), fmt.Sprintf("%s*%s", string(mat), string(mas)))
	return
}

func (rs *RadixStorage) LogError(uuid, source, errstr string) (err error) {
	if r := rs.db.Cmd("set", LOG_ERR+source+"_"+uuid, errstr); r.Err != nil {
		return r.Err
	}
	return
}
