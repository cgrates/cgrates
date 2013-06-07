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

package rater

import (
	"fmt"
	"menteslibres.net/gosexy/redis"
	//"log"
	"github.com/cgrates/cgrates/utils"
	"strconv"
	"strings"
	"time"
)

type GosexyStorage struct {
	dbNb int
	db   *redis.Client
	ms   Marshaler
}

func NewGosexyStorage(address string, db int, pass string) (DataStorage, error) {
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
	return &GosexyStorage{db: ndb, dbNb: db, ms: new(MyMarshaler)}, nil
}

func (rs *GosexyStorage) Close() {
	rs.db.Quit()
}

func (rs *GosexyStorage) Flush() (err error) {
	_, err = rs.db.FlushDB()
	return
}

func (rs *GosexyStorage) GetRatingProfile(key string) (rp *RatingProfile, err error) {
	var values string
	if values, err = rs.db.Get(RATING_PROFILE_PREFIX + key); err == nil {
		rp = new(RatingProfile)
		err = rs.ms.Unmarshal([]byte(values), rp)
	}
	return
}

func (rs *GosexyStorage) SetRatingProfile(rp *RatingProfile) (err error) {
	result, err := rs.ms.Marshal(rp)
	_, err = rs.db.Set(RATING_PROFILE_PREFIX+rp.Id, result)
	return
}

func (rs *GosexyStorage) GetDestination(key string) (dest *Destination, err error) {
	var values string
	if values, err = rs.db.Get(DESTINATION_PREFIX + key); err == nil {
		dest = &Destination{Id: key}
		err = rs.ms.Unmarshal([]byte(values), dest)
	}
	return
}

func (rs *GosexyStorage) SetDestination(dest *Destination) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(dest); err != nil {
		return
	}
	_, err = rs.db.Set(DESTINATION_PREFIX+dest.Id, result)
	return
}

func (rs *GosexyStorage) GetActions(key string) (as Actions, err error) {
	var values string
	if values, err = rs.db.Get(ACTION_PREFIX + key); err == nil {
		err = rs.ms.Unmarshal([]byte(values), &as)
	}
	return
}

func (rs *GosexyStorage) SetActions(key string, as Actions) (err error) {
	result, err := rs.ms.Marshal(&as)
	_, err = rs.db.Set(ACTION_PREFIX+key, result)
	return
}

func (rs *GosexyStorage) GetUserBalance(key string) (ub *UserBalance, err error) {
	var values string
	if values, err = rs.db.Get(USER_BALANCE_PREFIX + key); err == nil {
		ub = &UserBalance{Id: key}
		err = rs.ms.Unmarshal([]byte(values), ub)
	}

	return
}

func (rs *GosexyStorage) SetUserBalance(ub *UserBalance) (err error) {
	result, err := rs.ms.Marshal(ub)
	_, err = rs.db.Set(USER_BALANCE_PREFIX+ub.Id, result)
	return
}

func (rs *GosexyStorage) GetActionTimings(key string) (ats ActionTimings, err error) {
	var values string
	if values, err = rs.db.Get(ACTION_TIMING_PREFIX + key); err == nil {
		err = rs.ms.Unmarshal([]byte(values), &ats)
	}
	return
}

func (rs *GosexyStorage) SetActionTimings(key string, ats ActionTimings) (err error) {
	if len(ats) == 0 {
		// delete the key
		_, err = rs.db.Del(ACTION_TIMING_PREFIX + key)
		return err
	}
	result, err := rs.ms.Marshal(&ats)
	_, err = rs.db.Set(ACTION_TIMING_PREFIX+key, result)
	return
}

func (rs *GosexyStorage) GetAllActionTimings(tpid string) (ats map[string]ActionTimings, err error) {
	keys, err := rs.db.Keys(ACTION_TIMING_PREFIX + tpid + "*")
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

func (rs *GosexyStorage) LogCallCost(uuid, source string, cc *CallCost) (err error) {
	var result []byte
	result, err = rs.ms.Marshal(cc)
	if err != nil {
		return
	}
	_, err = rs.db.Set(LOG_CALL_COST_PREFIX+source+"_"+uuid, result)
	return
}

func (rs *GosexyStorage) GetCallCostLog(uuid, source string) (cc *CallCost, err error) {
	var values string
	if values, err = rs.db.Get(LOG_CALL_COST_PREFIX + source + "_" + uuid); err == nil {
		err = rs.ms.Unmarshal([]byte(values), cc)
	}
	return
}

func (rs *GosexyStorage) LogActionTrigger(ubId, source string, at *ActionTrigger, as Actions) (err error) {
	mat, err := rs.ms.Marshal(at)
	if err != nil {
		return
	}
	mas, err := rs.ms.Marshal(&as)
	if err != nil {
		return
	}
	rs.db.Set(LOG_ACTION_TRIGGER_PREFIX+source+"_"+time.Now().Format(time.RFC3339Nano), []byte(fmt.Sprintf("%s*%s*%s", ubId, string(mat), string(mas))))
	return
}

func (rs *GosexyStorage) LogActionTiming(source string, at *ActionTiming, as Actions) (err error) {
	mat, err := rs.ms.Marshal(at)
	if err != nil {
		return
	}
	mas, err := rs.ms.Marshal(&as)
	if err != nil {
		return
	}
	_, err = rs.db.Set(LOG_ACTION_TIMMING_PREFIX+source+"_"+time.Now().Format(time.RFC3339Nano), []byte(fmt.Sprintf("%s*%s", string(mat), string(mas))))
	return
}

func (rs *GosexyStorage) LogError(uuid, source, errstr string) (err error) {
	_, err = rs.db.Set(LOG_ERR+source+"_"+uuid, errstr)
	return
}

func (rs *GosexyStorage) SetCdr(utils.CDR) error {
	return nil
}

func (rs *GosexyStorage) SetRatedCdr(utils.CDR, *CallCost) error {
	return nil
}

func (rs *GosexyStorage) GetAllDestinations(tpid string) ([]*Destination, error) {
	return nil, nil
}

func (rs *GosexyStorage) GetAllRates(string) (map[string][]*Rate, error) {
	return nil, nil
}
func (rs *GosexyStorage) GetAllTimings(string) (map[string][]*Timing, error) {
	return nil, nil
}
func (rs *GosexyStorage) GetAllRateTimings(string) ([]*RateTiming, error) {
	return nil, nil
}
func (rs *GosexyStorage) GetAllRatingProfiles(string) (map[string]*RatingProfile, error) {
	return nil, nil
}
func (rs *GosexyStorage) GetAllActions(string) (map[string][]*Action, error) {
	return nil, nil
}
func (rs *GosexyStorage) GetAllActionTriggers(string) (map[string][]*ActionTrigger, error) {
	return nil, nil
}
