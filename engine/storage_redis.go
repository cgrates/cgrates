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
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"

	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/history"
	"github.com/cgrates/cgrates/utils"
	"github.com/hoisie/redis"

	"io/ioutil"
	"time"
)

type RedisStorage struct {
	dbNb int
	db   *redis.Client
	ms   Marshaler
}

func NewRedisStorage(address string, db int, pass, mrshlerStr string) (DataStorage, error) {
	ndb := &redis.Client{Addr: address, Db: db}
	if pass != "" {
		if err := ndb.Auth(pass); err != nil {
			return nil, err
		}
	}

	var mrshler Marshaler
	if mrshlerStr == utils.MSGPACK {
		mrshler = NewCodecMsgpackMarshaler()
	} else if mrshlerStr == utils.JSON {
		mrshler = new(JSONMarshaler)
	} else {
		return nil, fmt.Errorf("Unsupported marshaler: %v", mrshlerStr)
	}
	return &RedisStorage{db: ndb, dbNb: db, ms: mrshler}, nil
}

func (rs *RedisStorage) Close() {
	// no close for me
	//rs.db.Quit()
}

func (rs *RedisStorage) Flush() (err error) {
	err = rs.db.Flush(false)
	return
}

func (rs *RedisStorage) PreCache(dKeys, rpKeys, rpfKeys, actKeys []string) (err error) {
	if dKeys == nil {
		Logger.Info("Caching all destinations")
		if dKeys, err = rs.db.Keys(DESTINATION_PREFIX + "*"); err != nil {
			return
		}
	} else if len(dKeys) != 0 {
		Logger.Info(fmt.Sprintf("Caching destinations: %v", dKeys))
	}
	for _, key := range dKeys {
		cache2go.RemKey(key)
		if _, err = rs.GetDestination(key[len(DESTINATION_PREFIX):], true); err != nil {
			return err
		}
	}
	if len(dKeys) != 0 {
		Logger.Info("Finished destinations caching.")
	}
	if rpKeys == nil {
		Logger.Info("Caching all rating plans")
		if rpKeys, err = rs.db.Keys(RATING_PLAN_PREFIX + "*"); err != nil {
			return
		}
	} else if len(rpKeys) != 0 {
		Logger.Info(fmt.Sprintf("Caching rating plans: %v", rpKeys))
	}
	for _, key := range rpKeys {
		cache2go.RemKey(key)
		if _, err = rs.GetRatingPlan(key[len(RATING_PLAN_PREFIX):], true); err != nil {
			return err
		}
	}
	if len(rpKeys) != 0 {
		Logger.Info("Finished rating plans caching.")
	}
	if rpfKeys == nil {
		Logger.Info("Caching all rating profiles")
		if rpfKeys, err = rs.db.Keys(RATING_PROFILE_PREFIX + "*"); err != nil {
			return
		}
	} else if len(rpfKeys) != 0 {
		Logger.Info(fmt.Sprintf("Caching rating profile: %v", rpfKeys))
	}
	for _, key := range rpfKeys {
		cache2go.RemKey(key)
		if _, err = rs.GetRatingProfile(key[len(RATING_PROFILE_PREFIX):], true); err != nil {
			return err
		}
	}
	if len(rpfKeys) != 0 {
		Logger.Info("Finished actions caching.")
	}
	if actKeys == nil {
		Logger.Info("Caching all actions")
		if actKeys, err = rs.db.Keys(ACTION_PREFIX + "*"); err != nil {
			return
		}
	} else if len(actKeys) != 0 {
		Logger.Info(fmt.Sprintf("Caching actions: %v", actKeys))
	}
	for _, key := range actKeys {
		cache2go.RemKey(key)
		if _, err = rs.GetActions(key[len(ACTION_PREFIX):], true); err != nil {
			return err
		}
	}
	if len(actKeys) != 0 {
		Logger.Info("Finished actions caching.")
	}
	return
}

// Used to check if specific subject is stored using prefix key attached to entity
func (rs *RedisStorage) ExistsData(category, subject string) (bool, error) {
	switch category {
	case DESTINATION_PREFIX, RATING_PLAN_PREFIX, RATING_PROFILE_PREFIX, ACTION_PREFIX, ACTION_TIMING_PREFIX, USER_BALANCE_PREFIX:
		return rs.db.Exists(category + subject)
	}
	return false, errors.New("Unsupported category in ExistsData")
}

func (rs *RedisStorage) GetRatingPlan(key string, checkDb bool) (rp *RatingPlan, err error) {
	key = RATING_PLAN_PREFIX + key
	if x, err := cache2go.GetCached(key); err == nil {
		return x.(*RatingPlan), nil
	}
	if !checkDb {
		return nil, errors.New(utils.ERR_NOT_FOUND)
	}
	var values []byte
	if values, err = rs.db.Get(key); err == nil {
		b := bytes.NewBuffer(values)
		r, err := zlib.NewReader(b)
		if err != nil {
			return nil, err
		}
		out, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		r.Close()
		rp = new(RatingPlan)
		err = rs.ms.Unmarshal(out, rp)
		cache2go.Cache(key, rp)
	}
	return
}

func (rs *RedisStorage) SetRatingPlan(rp *RatingPlan) (err error) {
	result, err := rs.ms.Marshal(rp)
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	err = rs.db.Set(RATING_PLAN_PREFIX+rp.Id, b.Bytes())
	if err == nil && historyScribe != nil {
		response := 0
		go historyScribe.Record(&history.Record{RATING_PLAN_PREFIX + rp.Id, rp}, &response)
	}
	cache2go.Cache(RATING_PLAN_PREFIX+rp.Id, rp)
	return
}

func (rs *RedisStorage) GetRatingProfile(key string, checkDb bool) (rpf *RatingProfile, err error) {
	key = RATING_PROFILE_PREFIX + key
	if x, err := cache2go.GetCached(key); err == nil {
		return x.(*RatingProfile), nil
	}
	if !checkDb {
		return nil, errors.New(utils.ERR_NOT_FOUND)
	}
	var values []byte
	if values, err = rs.db.Get(key); err == nil {
		rpf = new(RatingProfile)
		err = rs.ms.Unmarshal(values, rpf)
		cache2go.Cache(key, rpf)
	}
	return
}

func (rs *RedisStorage) SetRatingProfile(rpf *RatingProfile) (err error) {
	result, err := rs.ms.Marshal(rpf)
	err = rs.db.Set(RATING_PROFILE_PREFIX+rpf.Id, result)
	if err == nil && historyScribe != nil {
		response := 0
		go historyScribe.Record(&history.Record{RATING_PROFILE_PREFIX + rpf.Id, rpf}, &response)
	}
	cache2go.Cache(RATING_PROFILE_PREFIX+rpf.Id, rpf)
	return
}

func (rs *RedisStorage) GetDestination(key string, checkDb bool) (dest *Destination, err error) {
	key = DESTINATION_PREFIX + key
	if x, err := cache2go.GetCached(key); err == nil {
		return x.(*Destination), nil
	}
	if !checkDb {
		return nil, errors.New(utils.ERR_NOT_FOUND)
	}
	var values []byte
	if values, err = rs.db.Get(key); len(values) > 0 && err == nil {
		b := bytes.NewBuffer(values)
		r, err := zlib.NewReader(b)
		if err != nil {
			return nil, err
		}
		out, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		r.Close()
		dest = new(Destination)
		err = rs.ms.Unmarshal(out, dest)
		cache2go.Cache(key, dest)
	}
	return
}

func (rs *RedisStorage) SetDestination(dest *Destination) (err error) {
	result, err := rs.ms.Marshal(dest)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	err = rs.db.Set(DESTINATION_PREFIX+dest.Id, b.Bytes())
	if err == nil && historyScribe != nil {
		response := 0
		go historyScribe.Record(&history.Record{DESTINATION_PREFIX + dest.Id, dest}, &response)
	}
	cache2go.Cache(DESTINATION_PREFIX+dest.Id, dest)
	return
}

func (rs *RedisStorage) GetActions(key string, checkDb bool) (as Actions, err error) {
	key = ACTION_PREFIX + key
	if x, err := cache2go.GetCached(key); err == nil {
		return x.(Actions), nil
	}
	if !checkDb {
		return nil, errors.New(utils.ERR_NOT_FOUND)
	}
	var values []byte
	if values, err = rs.db.Get(key); err == nil {
		err = rs.ms.Unmarshal(values, &as)
		cache2go.Cache(key, as)
	}
	return
}

func (rs *RedisStorage) SetActions(key string, as Actions) (err error) {
	result, err := rs.ms.Marshal(&as)
	err = rs.db.Set(ACTION_PREFIX+key, result)
	cache2go.Cache(ACTION_PREFIX+key, as)
	return
}

func (rs *RedisStorage) GetUserBalance(key string) (ub *UserBalance, err error) {
	var values []byte
	if values, err = rs.db.Get(USER_BALANCE_PREFIX + key); err == nil {
		ub = &UserBalance{Id: key}
		err = rs.ms.Unmarshal(values, ub)
	}

	return
}

func (rs *RedisStorage) SetUserBalance(ub *UserBalance) (err error) {
	result, err := rs.ms.Marshal(ub)
	err = rs.db.Set(USER_BALANCE_PREFIX+ub.Id, result)
	return
}

func (rs *RedisStorage) GetActionTimings(key string) (ats ActionTimings, err error) {
	var values []byte
	if values, err = rs.db.Get(ACTION_TIMING_PREFIX + key); err == nil {
		err = rs.ms.Unmarshal(values, &ats)
	}
	return
}

func (rs *RedisStorage) SetActionTimings(key string, ats ActionTimings) (err error) {
	if len(ats) == 0 {
		// delete the key
		_, err = rs.db.Del(ACTION_TIMING_PREFIX + key)
		return err
	}
	result, err := rs.ms.Marshal(&ats)
	err = rs.db.Set(ACTION_TIMING_PREFIX+key, result)
	return
}

func (rs *RedisStorage) GetAllActionTimings() (ats map[string]ActionTimings, err error) {
	keys, err := rs.db.Keys(ACTION_TIMING_PREFIX + "*")
	if err != nil {
		return nil, err
	}
	ats = make(map[string]ActionTimings, len(keys))
	for _, key := range keys {
		values, err := rs.db.Get(key)
		if err != nil {
			continue
		}
		var tempAts ActionTimings
		err = rs.ms.Unmarshal(values, &tempAts)
		ats[key[len(ACTION_TIMING_PREFIX):]] = tempAts
	}

	return
}

func (rs *RedisStorage) LogCallCost(uuid, source string, cc *CallCost) (err error) {
	var result []byte
	result, err = rs.ms.Marshal(cc)
	if err != nil {
		return
	}
	err = rs.db.Set(LOG_CALL_COST_PREFIX+source+"_"+uuid, result)
	return
}

func (rs *RedisStorage) GetCallCostLog(uuid, source string) (cc *CallCost, err error) {
	var values []byte
	if values, err = rs.db.Get(LOG_CALL_COST_PREFIX + source + "_" + uuid); err == nil {
		err = rs.ms.Unmarshal(values, cc)
	}
	return
}

func (rs *RedisStorage) LogActionTrigger(ubId, source string, at *ActionTrigger, as Actions) (err error) {
	mat, err := rs.ms.Marshal(at)
	if err != nil {
		return
	}
	mas, err := rs.ms.Marshal(as)
	if err != nil {
		return
	}
	rs.db.Set(LOG_ACTION_TRIGGER_PREFIX+source+"_"+time.Now().Format(time.RFC3339Nano), []byte(fmt.Sprintf("%v*%v*%v", ubId, string(mat), string(mas))))
	return
}

func (rs *RedisStorage) LogActionTiming(source string, at *ActionTiming, as Actions) (err error) {
	mat, err := rs.ms.Marshal(at)
	if err != nil {
		return
	}
	mas, err := rs.ms.Marshal(as)
	if err != nil {
		return
	}
	err = rs.db.Set(LOG_ACTION_TIMMING_PREFIX+source+"_"+time.Now().Format(time.RFC3339Nano), []byte(fmt.Sprintf("%v*%v", string(mat), string(mas))))
	return
}

func (rs *RedisStorage) LogError(uuid, source, errstr string) (err error) {
	err = rs.db.Set(LOG_ERR+source+"_"+uuid, []byte(errstr))
	return
}
