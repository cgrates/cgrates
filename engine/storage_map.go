/*
Rating system designed to be used in VoIP Carriems World
Copyright (C) 2013 ITsysCOM

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either vemsion 3 of the License, or
(at your option) any later vemsion.

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
	"strings"
	"time"
)

type MapStorage struct {
	dict map[string][]byte
	ms   Marshaler
}

func NewMapStorage() (DataStorage, error) {
	return &MapStorage{dict: make(map[string][]byte), ms: new(MyMarshaler)}, nil
}

func (ms *MapStorage) Close() {}

func (ms *MapStorage) Flush() error {
	ms.dict = make(map[string][]byte)
	return nil
}

func (ms *MapStorage) GetRatingProfile(key string) (rp *RatingProfile, err error) {
	if values, ok := ms.dict[RATING_PROFILE_PREFIX+key]; ok {
		rp = new(RatingProfile)
		err = ms.ms.Unmarshal(values, rp)
	} else {
		return nil, errors.New("not found")
	}
	return
}

func (ms *MapStorage) SetRatingProfile(rp *RatingProfile) (err error) {
	result, err := ms.ms.Marshal(rp)
	ms.dict[RATING_PROFILE_PREFIX+rp.Id] = result
	response := 0
	go historyScribe.Record(&history.Record{RATING_PROFILE_PREFIX + rp.Id, rp}, &response)
	return
}

func (ms *MapStorage) GetDestination(key string) (dest *Destination, err error) {
	if values, ok := ms.dict[DESTINATION_PREFIX+key]; ok {
		dest = &Destination{Id: key}
		err = ms.ms.Unmarshal(values, dest)
	} else {
		return nil, errors.New("not found")
	}
	return
}
func (ms *MapStorage) SetDestination(dest *Destination) (err error) {
	result, err := ms.ms.Marshal(dest)
	ms.dict[DESTINATION_PREFIX+dest.Id] = result
	response := 0
	go historyScribe.Record(&history.Record{DESTINATION_PREFIX + dest.Id, dest}, &response)
	return
}

func (ms *MapStorage) GetTPIds() ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) SetTPTiming(tpid string, tm *Timing) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) ExistsTPTiming(tpid, tmId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) GetTPTiming(tpid, tmId string) (*Timing, error) {
	return nil, nil
}

func (ms *MapStorage) GetTPTimingIds(tpid string) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) SetTPDestination(tpid string, dest *Destination) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) ExistsTPDestination(tpid, destTag string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) GetTPDestination(tpid, destTag string) (*Destination, error) {
	return nil, nil
}

func (ms *MapStorage) GetTPDestinationIds(tpid string) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) ExistsTPRate(tpid, rtId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) SetTPRates(tpid string, rts map[string][]*Rate) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) GetTPRate(tpid, rtId string) (*utils.TPRate, error) {
	return nil, nil
}

func (ms *MapStorage) GetTPRateIds(tpid string) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) ExistsTPDestinationRate(tpid, drId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) SetTPDestinationRates(tpid string, drs map[string][]*DestinationRate) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) GetTPDestinationRate(tpid, drId string) (*utils.TPDestinationRate, error) {
	return nil, nil
}

func (ms *MapStorage) GetTPDestinationRateIds(tpid string) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) ExistsTPDestRateTiming(tpid, drtId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) SetTPDestRateTimings(tpid string, drts map[string][]*DestinationRateTiming) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) GetTPDestRateTiming(tpid, drtId string) (*utils.TPDestRateTiming, error) {
	return nil, nil
}

func (ms *MapStorage) GetTPDestRateTimingIds(tpid string) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) ExistsTPRatingProfile(tpid, rpId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) SetTPRatingProfiles(tpid string, rps map[string][]*RatingProfile) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) GetTPRatingProfile(tpid, rpId string) (*utils.TPRatingProfile, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) GetTPRatingProfileIds(filters *utils.AttrTPRatingProfileIds) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) ExistsTPActions(tpid, aId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) SetTPActions(tpid string, acts map[string][]*Action) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) GetTPActions(tpid, aId string) (*utils.TPActions, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) GetTPActionIds(tpid string) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) ExistsTPActionTimings(tpid, atId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) SetTPActionTimings(tpid string, ats map[string][]*ActionTiming) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) GetTPActionTimings(tpid, atId string) (map[string][]*utils.TPActionTimingsRow, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) GetTPActionTimingIds(tpid string) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) ExistsTPActionTriggers(tpid, atId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) SetTPActionTriggers(tpid string, ats map[string][]*ActionTrigger) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) GetTPActionTriggerIds(tpid string) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) ExistsTPAccountActions(tpid, aaId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) SetTPAccountActions(tpid string, aa map[string]*AccountAction) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) GetTPAccountActionIds(tpid string) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) GetActions(key string) (as Actions, err error) {
	if values, ok := ms.dict[ACTION_PREFIX+key]; ok {
		err = ms.ms.Unmarshal(values, &as)
	} else {
		return nil, errors.New("not found")
	}
	return
}

func (ms *MapStorage) SetActions(key string, as Actions) (err error) {
	result, err := ms.ms.Marshal(&as)
	ms.dict[ACTION_PREFIX+key] = result
	return
}

func (ms *MapStorage) GetUserBalance(key string) (ub *UserBalance, err error) {
	if values, ok := ms.dict[USER_BALANCE_PREFIX+key]; ok {
		ub = &UserBalance{Id: key}
		err = ms.ms.Unmarshal(values, ub)
	} else {
		return nil, errors.New("not found")
	}
	return
}

func (ms *MapStorage) SetUserBalance(ub *UserBalance) (err error) {
	result, err := ms.ms.Marshal(ub)
	ms.dict[USER_BALANCE_PREFIX+ub.Id] = result
	return
}

func (ms *MapStorage) GetActionTimings(key string) (ats ActionTimings, err error) {
	if values, ok := ms.dict[ACTION_TIMING_PREFIX+key]; ok {
		err = ms.ms.Unmarshal(values, &ats)
	} else {
		return nil, errors.New("not found")
	}
	return
}

func (ms *MapStorage) SetActionTimings(key string, ats ActionTimings) (err error) {
	if len(ats) == 0 {
		// delete the key
		delete(ms.dict, ACTION_TIMING_PREFIX+key)
		return
	}
	result, err := ms.ms.Marshal(&ats)
	ms.dict[ACTION_TIMING_PREFIX+key] = result
	return
}

func (ms *MapStorage) GetAllActionTimings() (ats map[string]ActionTimings, err error) {
	ats = make(map[string]ActionTimings)
	for key, value := range ms.dict {
		if !strings.Contains(key, ACTION_TIMING_PREFIX) {
			continue
		}
		var tempAts ActionTimings
		err = ms.ms.Unmarshal(value, &tempAts)
		ats[key[len(ACTION_TIMING_PREFIX):]] = tempAts
	}

	return
}

func (ms *MapStorage) LogCallCost(uuid, source string, cc *CallCost) error {
	result, err := ms.ms.Marshal(cc)
	ms.dict[LOG_CALL_COST_PREFIX+source+"_"+uuid] = result
	return err
}

func (ms *MapStorage) GetCallCostLog(uuid, source string) (cc *CallCost, err error) {
	if values, ok := ms.dict[LOG_CALL_COST_PREFIX+source+"_"+uuid]; ok {
		err = ms.ms.Unmarshal(values, &cc)
	} else {
		return nil, errors.New("not found")
	}
	return
}

func (ms *MapStorage) LogActionTrigger(ubId, source string, at *ActionTrigger, as Actions) (err error) {
	mat, err := ms.ms.Marshal(at)
	if err != nil {
		return
	}
	mas, err := ms.ms.Marshal(&as)
	if err != nil {
		return
	}
	ms.dict[LOG_ACTION_TRIGGER_PREFIX+source+"_"+time.Now().Format(time.RFC3339Nano)] = []byte(fmt.Sprintf("%s*%s*%s", ubId, string(mat), string(mas)))
	return
}

func (ms *MapStorage) LogActionTiming(source string, at *ActionTiming, as Actions) (err error) {
	mat, err := ms.ms.Marshal(at)
	if err != nil {
		return
	}
	mas, err := ms.ms.Marshal(&as)
	if err != nil {
		return
	}
	ms.dict[LOG_ACTION_TIMMING_PREFIX+source+"_"+time.Now().Format(time.RFC3339Nano)] = []byte(fmt.Sprintf("%s*%s", string(mat), string(mas)))
	return
}

func (ms *MapStorage) LogError(uuid, source, errstr string) (err error) {
	ms.dict[LOG_ERR+source+"_"+uuid] = []byte(errstr)
	return nil
}

func (ms *MapStorage) SetCdr(utils.CDR) error {
	return nil
}

func (ms *MapStorage) SetRatedCdr(cdr utils.CDR, cc *CallCost, extraInfo string) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MapStorage) GetAllRatedCdr() ([]utils.CDR, error) {
	return nil, nil
}

func (ms *MapStorage) GetTpDestinations(tpid, tag string) ([]*Destination, error) {
	return nil, nil
}

func (ms *MapStorage) GetTpRates(tpid, tag string) (map[string]*Rate, error) {
	return nil, nil
}
func (ms *MapStorage) GetTpDestinationRates(tpid, tag string) (map[string][]*DestinationRate, error) {
	return nil, nil
}
func (ms *MapStorage) GetTpTimings(tpid, tag string) (map[string]*Timing, error) {
	return nil, nil
}
func (ms *MapStorage) GetTpDestinationRateTimings(tpid, tag string) ([]*DestinationRateTiming, error) {
	return nil, nil
}

func (ms *MapStorage) GetTpRatingProfiles(tpid, tag string) (map[string]*RatingProfile, error) {
	return nil, nil
}
func (ms *MapStorage) GetTpActions(tpid, tag string) (map[string][]*Action, error) {
	return nil, nil
}
func (ms *MapStorage) GetTpActionTimings(tpid, tag string) (map[string][]*ActionTiming, error) {
	return nil, nil
}
func (ms *MapStorage) GetTpActionTriggers(tpid, tag string) (map[string][]*ActionTrigger, error) {
	return nil, nil
}
func (ms *MapStorage) GetTpAccountActions(tpid, tag string) (map[string]*AccountAction, error) {
	return nil, nil
}
