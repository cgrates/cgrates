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
	"fmt"
	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/history"
	"github.com/cgrates/cgrates/utils"
	"io/ioutil"
	"menteslibres.net/gosexy/redis"
	"strconv"
	"strings"
	"time"
)

type RedisStorage struct {
	dbNb int
	db   *redis.Client
	ms   Marshaler
}

func NewRedisStorage(address string, db int, pass, mrshlerStr string) (DataStorage, error) {
	addrSplit := strings.Split(address, ":")
	host := addrSplit[0]
	port := 6379
	if len(addrSplit) == 2 {
		port, _ = strconv.Atoi(addrSplit[1])
	}
	ndb := redis.New()
	err := ndb.Connect(host, uint(port))
	if err != nil {
		return nil, err
	}
	if pass != "" {
		if _, err = ndb.Auth(pass); err != nil {
			return nil, err
		}
	}
	if db > 0 {
		if _, err = ndb.Select(int64(db)); err != nil {
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
	rs.db.Quit()
}

func (rs *RedisStorage) Flush() (err error) {
	_, err = rs.db.FlushDB()
	return
}

func (rs *RedisStorage) PreCache(dKeys, rpKeys []string) (err error) {
	if dKeys == nil {
		if dKeys, err = rs.db.Keys(DESTINATION_PREFIX + "*"); err != nil {
			return
		}
	}
	for _, key := range dKeys {
		if _, err = rs.GetDestination(key[len(DESTINATION_PREFIX):]); err != nil {
			return err
		}
	}
	if rpKeys == nil {
		if rpKeys, err = rs.db.Keys(RATING_PLAN_PREFIX + "*"); err != nil {
			return
		}
	}
	for _, key := range rpKeys {
		if _, err = rs.GetRatingPlan(key[len(RATING_PLAN_PREFIX):]); err != nil {
			return err
		}
	}
	return
}

func (rs *RedisStorage) GetRatingPlan(key string) (rp *RatingPlan, err error) {
	if x, err := cache2go.GetCached(key); err == nil {
		return x.(*RatingPlan), nil
	}
	var values string
	if values, err = rs.db.Get(RATING_PLAN_PREFIX + key); err == nil {
		b := bytes.NewBufferString(values)
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
	_, err = rs.db.Set(RATING_PLAN_PREFIX+rp.Id, b.Bytes())
	if err == nil && historyScribe != nil {
		response := 0
		historyScribe.Record(&history.Record{RATING_PLAN_PREFIX + rp.Id, rp}, &response)
	}
	return
}

func (rs *RedisStorage) ExistsRatingPlan(rpId string) (bool, error) {
	return rs.db.Exists(RATING_PLAN_PREFIX + rpId)
}

func (rs *RedisStorage) GetRatingProfile(key string) (rp *RatingProfile, err error) {
	var values string
	if values, err = rs.db.Get(RATING_PROFILE_PREFIX + key); err == nil {
		rp = new(RatingProfile)
		err = rs.ms.Unmarshal([]byte(values), rp)
	}
	return
}

func (rs *RedisStorage) SetRatingProfile(rp *RatingProfile) (err error) {
	result, err := rs.ms.Marshal(rp)
	_, err = rs.db.Set(RATING_PROFILE_PREFIX+rp.Id, result)
	if err == nil && historyScribe != nil {
		response := 0
		historyScribe.Record(&history.Record{RATING_PROFILE_PREFIX + rp.Id, rp}, &response)
	}
	return
}

func (rs *RedisStorage) GetDestination(key string) (dest *Destination, err error) {
	if x, err := cache2go.GetCached(key); err == nil {
		return x.(*Destination), nil
	}
	var values string
	if values, err = rs.db.Get(DESTINATION_PREFIX + key); len(values) > 0 && err == nil {
		b := bytes.NewBufferString(values)
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

/*func (rs *RedisStorage) DestinationContainsPrefix(key string, prefix string) (precision int, err error) {
	if _, err := rs.db.SAdd(TEMP_DESTINATION_PREFIX+prefix, utils.SplitPrefixInterface(prefix)...); err != nil {
		return 0, err
	}
	var values []string
	if values, err = rs.db.SInter(DESTINATION_PREFIX+key, TEMP_DESTINATION_PREFIX+prefix); err == nil {
		for _, p := range values {
			if len(p) > precision {
				precision = len(p)
			}
		}
	}
	if _, err := rs.db.Del(TEMP_DESTINATION_PREFIX + prefix); err != nil {
		Logger.Err("Error removing temp ")
	}
	return
}*/

func (rs *RedisStorage) SetDestination(dest *Destination) (err error) {
	result, err := rs.ms.Marshal(dest)
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	_, err = rs.db.Set(DESTINATION_PREFIX+dest.Id, b.Bytes())
	if err == nil && historyScribe != nil {
		response := 0
		historyScribe.Record(&history.Record{DESTINATION_PREFIX + dest.Id, dest}, &response)
	}
	return
}

func (rs *RedisStorage) GetActions(key string) (as Actions, err error) {
	var values string
	if values, err = rs.db.Get(ACTION_PREFIX + key); err == nil {
		err = rs.ms.Unmarshal([]byte(values), &as)
	}
	return
}

func (rs *RedisStorage) SetActions(key string, as Actions) (err error) {
	result, err := rs.ms.Marshal(&as)
	_, err = rs.db.Set(ACTION_PREFIX+key, result)
	return
}

func (rs *RedisStorage) GetUserBalance(key string) (ub *UserBalance, err error) {
	var values string
	if values, err = rs.db.Get(USER_BALANCE_PREFIX + key); err == nil {
		ub = &UserBalance{Id: key}
		err = rs.ms.Unmarshal([]byte(values), ub)
	}

	return
}

func (rs *RedisStorage) SetUserBalance(ub *UserBalance) (err error) {
	result, err := rs.ms.Marshal(ub)
	_, err = rs.db.Set(USER_BALANCE_PREFIX+ub.Id, result)
	return
}

func (rs *RedisStorage) GetActionTimings(key string) (ats ActionTimings, err error) {
	var values string
	if values, err = rs.db.Get(ACTION_TIMING_PREFIX + key); err == nil {
		err = rs.ms.Unmarshal([]byte(values), &ats)
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
	_, err = rs.db.Set(ACTION_TIMING_PREFIX+key, result)
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
		err = rs.ms.Unmarshal([]byte(values), &tempAts)
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
	_, err = rs.db.Set(LOG_CALL_COST_PREFIX+source+"_"+uuid, result)
	return
}

func (rs *RedisStorage) GetCallCostLog(uuid, source string) (cc *CallCost, err error) {
	var values string
	if values, err = rs.db.Get(LOG_CALL_COST_PREFIX + source + "_" + uuid); err == nil {
		err = rs.ms.Unmarshal([]byte(values), cc)
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
	_, err = rs.db.Set(LOG_ACTION_TIMMING_PREFIX+source+"_"+time.Now().Format(time.RFC3339Nano), []byte(fmt.Sprintf("%v*%v", string(mat), string(mas))))
	return
}

func (rs *RedisStorage) LogError(uuid, source, errstr string) (err error) {
	_, err = rs.db.Set(LOG_ERR+source+"_"+uuid, errstr)
	return
}
