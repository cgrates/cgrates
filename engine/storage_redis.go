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
	"errors"
	"fmt"
	"github.com/cgrates/cgrates/history"
	"github.com/cgrates/cgrates/utils"
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

func NewRedisStorage(address string, db int, pass string) (DataStorage, error) {
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
	return &RedisStorage{db: ndb, dbNb: db, ms: new(MsgpackMarshaler)}, nil
}

func (rs *RedisStorage) Close() {
	rs.db.Quit()
}

func (rs *RedisStorage) Flush() (err error) {
	_, err = rs.db.FlushDB()
	return
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
	var values string
	if values, err = rs.db.Get(DESTINATION_PREFIX + key); err == nil {
		dest = &Destination{Id: key}
		err = rs.ms.Unmarshal([]byte(values), dest)
	}
	return
}

func (rs *RedisStorage) SetDestination(dest *Destination) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(dest); err != nil {
		return
	}
	_, err = rs.db.Set(DESTINATION_PREFIX+dest.Id, result)
	if err == nil && historyScribe != nil {
		response := 0
		historyScribe.Record(&history.Record{DESTINATION_PREFIX + dest.Id, dest}, &response)
	}
	return
}

func (rs *RedisStorage) GetTPIds() ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) SetTPTiming(tpid string, tm *Timing) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) ExistsTPTiming(tpid, tmId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) GetTPTiming(tpid, tmId string) (*Timing, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) GetTPTimingIds(tpid string) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) GetTPDestinationIds(tpid string) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) ExistsTPDestination(tpid, destTag string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

// Extracts destinations from StorDB on specific tariffplan id
func (rs *RedisStorage) GetTPDestination(tpid, destTag string) (*Destination, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) SetTPDestination(tpid string, dest *Destination) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) ExistsTPRate(tpid, rtId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) SetTPRates(tpid string, rts map[string][]*Rate) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) GetTPRate(tpid, rtId string) (*utils.TPRate, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) GetTPRateIds(tpid string) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) ExistsTPDestinationRate(tpid, drId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) SetTPDestinationRates(tpid string, drs map[string][]*DestinationRate) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) GetTPDestinationRate(tpid, drId string) (*utils.TPDestinationRate, error) {
	return nil, nil
}

func (rs *RedisStorage) GetTPDestinationRateIds(tpid string) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) ExistsTPDestRateTiming(tpid, drtId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) SetTPDestRateTimings(tpid string, drts map[string][]*DestinationRateTiming) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) GetTPDestRateTiming(tpid, drtId string) (*utils.TPDestRateTiming, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) GetTPDestRateTimingIds(tpid string) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) ExistsTPRatingProfile(tpid, rpId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) SetTPRatingProfiles(tpid string, rps map[string][]*RatingProfile) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) GetTPRatingProfile(tpid, rpId string) (*utils.TPRatingProfile, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) GetTPRatingProfileIds(filters *utils.AttrTPRatingProfileIds) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) ExistsTPActions(tpid, aId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) SetTPActions(tpid string, acts map[string][]*Action) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) GetTPActions(tpid, aId string) (*utils.TPActions, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) GetTPActionIds(tpid string) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) ExistsTPActionTimings(tpid, atId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) SetTPActionTimings(tpid string, ats map[string][]*ActionTiming) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) GetTPActionTimings(tpid, atId string) (map[string][]*utils.TPActionTimingsRow, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) GetTPActionTimingIds(tpid string) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) ExistsTPActionTriggers(tpid, atId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) SetTPActionTriggers(tpid string, ats map[string][]*ActionTrigger) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) GetTPActionTriggerIds(tpid string) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) ExistsTPAccountActions(tpid, aaId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) SetTPAccountActions(tpid string, aa map[string]*AccountAction) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) GetTPAccountActionIds(tpid string) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
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
	mas, err := rs.ms.Marshal(&as)
	if err != nil {
		return
	}
	rs.db.Set(LOG_ACTION_TRIGGER_PREFIX+source+"_"+time.Now().Format(time.RFC3339Nano), []byte(fmt.Sprintf("%s*%s*%s", ubId, string(mat), string(mas))))
	return
}

func (rs *RedisStorage) LogActionTiming(source string, at *ActionTiming, as Actions) (err error) {
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

func (rs *RedisStorage) LogError(uuid, source, errstr string) (err error) {
	_, err = rs.db.Set(LOG_ERR+source+"_"+uuid, errstr)
	return
}

func (rs *RedisStorage) SetCdr(utils.CDR) error {
	return nil
}

func (rs *RedisStorage) SetRatedCdr(cdr utils.CDR, cc *CallCost, extraInfo string) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (rs *RedisStorage) GetAllRatedCdr() ([]utils.CDR, error) {
	return nil, nil
}

func (rs *RedisStorage) GetTpDestinations(tpid, tag string) ([]*Destination, error) {
	return nil, nil
}

func (rs *RedisStorage) GetTpRates(tpid, tag string) (map[string]*Rate, error) {
	return nil, nil
}

func (ms *RedisStorage) GetTpDestinationRates(tpid, tag string) (map[string][]*DestinationRate, error) {
	return nil, nil
}
func (rs *RedisStorage) GetTpTimings(tpid, tag string) (map[string]*Timing, error) {
	return nil, nil
}
func (rs *RedisStorage) GetTpDestinationRateTimings(tpid, tag string) ([]*DestinationRateTiming, error) {
	return nil, nil
}
func (rs *RedisStorage) GetTpRatingProfiles(tpid, tag string) (map[string]*RatingProfile, error) {
	return nil, nil
}
func (rs *RedisStorage) GetTpActions(tpid, tag string) (map[string][]*Action, error) {
	return nil, nil
}
func (rs *RedisStorage) GetTpActionTimings(tpid, tag string) (map[string][]*ActionTiming, error) {
	return nil, nil
}
func (rs *RedisStorage) GetTpActionTriggers(tpid, tag string) (map[string][]*ActionTrigger, error) {
	return nil, nil
}
func (rs *RedisStorage) GetTpAccountActions(tpid, tag string) (map[string]*AccountAction, error) {
	return nil, nil
}
