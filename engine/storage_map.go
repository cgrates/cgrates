/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

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
	"bytes"
	"compress/zlib"
	"errors"
	"io/ioutil"
	"sync"

	"strings"

	"github.com/cgrates/cgrates/utils"
)

type MapStorage struct {
	dict         map[string][]byte
	tasks        [][]byte
	ms           Marshaler
	mu           sync.RWMutex
	cacheDumpDir string
}

func NewMapStorage() (*MapStorage, error) {
	CacheSetDumperPath("/tmp/cgrates")
	return &MapStorage{dict: make(map[string][]byte), ms: NewCodecMsgpackMarshaler(), cacheDumpDir: "/tmp/cgrates"}, nil
}

func NewMapStorageJson() (*MapStorage, error) {
	CacheSetDumperPath("/tmp/cgrates")
	return &MapStorage{dict: make(map[string][]byte), ms: new(JSONBufMarshaler), cacheDumpDir: "/tmp/cgrates"}, nil
}

func (ms *MapStorage) Close() {}

func (ms *MapStorage) Flush(ignore string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.dict = make(map[string][]byte)
	return nil
}

func (ms *MapStorage) GetKeysForPrefix(prefix string) ([]string, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if x, ok := CacheGet(utils.GENERIC_PREFIX + "keys_" + prefix); ok {
		return x.([]string), nil
	}
	keysForPrefix := make([]string, 0)
	for key := range ms.dict {
		if strings.HasPrefix(key, prefix) {
			keysForPrefix = append(keysForPrefix, key)
		}
	}
	CacheSet(utils.GENERIC_PREFIX+"keys_"+prefix, r.List())
	return keysForPrefix, nil
}

// Used to check if specific subject is stored using prefix key attached to entity
func (ms *MapStorage) HasData(categ, subject string) (bool, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	switch categ {
	case utils.DESTINATION_PREFIX, utils.RATING_PLAN_PREFIX, utils.RATING_PROFILE_PREFIX, utils.ACTION_PREFIX, utils.ACTION_PLAN_PREFIX, utils.ACCOUNT_PREFIX, utils.DERIVEDCHARGERS_PREFIX:
		if x, ok := CacheGet(utils.GENERIC_PREFIX + "has_" + category + "_" + subject); ok {
			return x.(bool), nil
		}
		_, exists := ms.dict[categ+subject]
		CacheSet(utils.GENERIC_PREFIX+"has_"+category+"_"+subject, exists)
		return exists, nil
	}
	CacheSet(utils.GENERIC_PREFIX+"has_"+category+"_"+subject, false)
	return false, errors.New("Unsupported HasData category")
}

func (ms *MapStorage) GetRatingPlan(key string) (rp *RatingPlan, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	key = utils.RATING_PLAN_PREFIX + key
	if x, ok := CacheGet(key); ok {
		if x != nil {
			return x.(*RatingPlan), nil
		}
		return nil, utils.ErrNotFound
	}
	if values, ok := ms.dict[key]; ok {
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
		err = ms.ms.Unmarshal(out, rp)
	} else {
		CacheSet(key, nil)
		return nil, utils.ErrNotFound
	}
	CacheSet(key, rp)
	return
}

func (ms *MapStorage) SetRatingPlan(rp *RatingPlan, cache bool) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(rp)
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	ms.dict[utils.RATING_PLAN_PREFIX+rp.Id] = b.Bytes()
	response := 0
	if historyScribe != nil {
		go historyScribe.Call("HistoryV1.Record", rp.GetHistoryRecord(), &response)
	}
	if cache && err == nil {
		CacheSet(utils.RATING_PLAN_PREFIX+rp.Id, rp)
	}
	return
}

func (ms *MapStorage) GetRatingProfile(key string) (rpf *RatingProfile, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	key = utils.RATING_PROFILE_PREFIX + key
	if x, ok := CacheGet(key); ok {
		if x != nil {
			return x.(*RatingProfile), nil
		}
		return nil, utils.ErrNotFound
	}

	if values, ok := ms.dict[key]; ok {
		rpf = new(RatingProfile)

		err = ms.ms.Unmarshal(values, rpf)
	} else {
		CacheSet(key, nil)
		return nil, utils.ErrNotFound
	}
	CacheSet(key, rpf)
	return
}

func (ms *MapStorage) SetRatingProfile(rpf *RatingProfile, cache bool) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(rpf)
	ms.dict[utils.RATING_PROFILE_PREFIX+rpf.Id] = result
	response := 0
	if historyScribe != nil {
		go historyScribe.Call("HistoryV1.Record", rpf.GetHistoryRecord(false), &response)
	}
	if cache && err == nil {
		CacheSet(utils.RATING_PROFILE_PREFIX+rpf.Id, rpf)
	}
	return
}

func (ms *MapStorage) RemoveRatingProfile(key string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for k := range ms.dict {
		if strings.HasPrefix(k, key) {
			delete(ms.dict, key)
			CacheRemKey(k)
			response := 0
			rpf := &RatingProfile{Id: key}
			if historyScribe != nil {
				go historyScribe.Call("HistoryV1.Record", rpf.GetHistoryRecord(true), &response)
			}
		}
	}
	return
}

func (ms *MapStorage) GetLCR(key string) (lcr *LCR, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	key = utils.LCR_PREFIX + key
	if x, ok := CacheGet(key); ok {
		if x != nil {
			return x.(*LCR), nil
		}
		return nil, utils.ErrNotFound
	}
	if values, ok := ms.dict[key]; ok {
		err = ms.ms.Unmarshal(values, &lcr)
	} else {
		CacheSet(key, nil)
		return nil, utils.ErrNotFound
	}
	CacheSet(key, lcr)
	return
}

func (ms *MapStorage) SetLCR(lcr *LCR, cache bool) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(lcr)
	ms.dict[utils.LCR_PREFIX+lcr.GetId()] = result
	if cache && err == nil {
		CacheSet(utils.LCR_PREFIX+lcr.GetId(), lcr)
	}
	return
}

func (ms *MapStorage) GetDestination(key string) (dest *Destination, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	key = utils.DESTINATION_PREFIX + key
	if values, ok := ms.dict[key]; ok {
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
		err = ms.ms.Unmarshal(out, dest)
		// create optimized structure
		for _, p := range dest.Prefixes {
			CachePush(utils.DESTINATION_PREFIX+p, dest.Id)
		}
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) SetDestination(dest *Destination, cache bool) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(dest)
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	ms.dict[utils.DESTINATION_PREFIX+dest.Id] = b.Bytes()
	response := 0
	if historyScribe != nil {
		go historyScribe.Call("HistoryV1.Record", dest.GetHistoryRecord(false), &response)
	}
	if cache && err == nil {
		CacheSet(utils.DESTINATION_PREFIX+dest.Id, dest)
	}
	return
}

func (ms *MapStorage) RemoveDestination(destID string) (err error) {
	return
}

func (ms *MapStorage) GetActions(key string) (as Actions, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	key = utils.ACTION_PREFIX + key
	if x, ok := CacheGet(key); ok {
		if x != nil {
			return x.(Actions), nil
		}
		return nil, utils.ErrNotFound
	}
	if values, ok := ms.dict[key]; ok {
		err = ms.ms.Unmarshal(values, &as)
	} else {
		CacheSet(key, nil)
		return nil, utils.ErrNotFound
	}
	CacheSet(key, as)
	return
}

func (ms *MapStorage) SetActions(key string, as Actions, cache bool) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(&as)
	ms.dict[utils.ACTION_PREFIX+key] = result
	if cache && err == nil {
		CacheSet(utils.ACTION_PREFIX+key, as)
	}
	return
}

func (ms *MapStorage) RemoveActions(key string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.dict, utils.ACTION_PREFIX+key)
	CacheRemKey(utils.ACTION_PREFIX + key)
	return
}

func (ms *MapStorage) GetSharedGroup(key string) (sg *SharedGroup, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	key = utils.SHARED_GROUP_PREFIX + key
	if x, ok := CacheGet(key); ok {
		if x != nil {
			return x.(*SharedGroup), nil
		}
		return nil, utils.ErrNotFound
	}
	if values, ok := ms.dict[key]; ok {
		err = ms.ms.Unmarshal(values, &sg)
		if err == nil {
			CacheSet(key, sg)
		}
	} else {
		CacheSet(key, nil)
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) SetSharedGroup(sg *SharedGroup, cache bool) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(sg)
	ms.dict[utils.SHARED_GROUP_PREFIX+sg.Id] = result
	if cache && err == nil {
		CacheSet(utils.SHARED_GROUP_PREFIX+sg.Id, sg)
	}
	return
}

func (ms *MapStorage) GetAccount(key string) (ub *Account, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if values, ok := ms.dict[utils.ACCOUNT_PREFIX+key]; ok {
		ub = &Account{ID: key}
		err = ms.ms.Unmarshal(values, ub)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) SetAccount(ub *Account) (err error) {
	// never override existing account with an empty one
	// UPDATE: if all balances expired and were cleaned it makes
	// sense to write empty balance map
	if len(ub.BalanceMap) == 0 {
		if ac, err := ms.GetAccount(ub.ID); err == nil && !ac.allBalancesExpired() {
			ac.ActionTriggers = ub.ActionTriggers
			ac.UnitCounters = ub.UnitCounters
			ac.AllowNegative = ub.AllowNegative
			ac.Disabled = ub.Disabled
			ub = ac
		}
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(ub)
	ms.dict[utils.ACCOUNT_PREFIX+ub.ID] = result
	return
}

func (ms *MapStorage) RemoveAccount(key string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.dict, utils.ACCOUNT_PREFIX+key)
	return
}

func (ms *MapStorage) GetCdrStatsQueue(key string) (sq *StatsQueue, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if values, ok := ms.dict[utils.CDR_STATS_QUEUE_PREFIX+key]; ok {
		sq = &StatsQueue{}
		err = ms.ms.Unmarshal(values, sq)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) SetCdrStatsQueue(sq *StatsQueue) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(sq)
	ms.dict[utils.CDR_STATS_QUEUE_PREFIX+sq.GetId()] = result
	return
}

func (ms *MapStorage) GetSubscribers() (result map[string]*SubscriberData, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	result = make(map[string]*SubscriberData)
	for key, value := range ms.dict {
		if strings.HasPrefix(key, utils.PUBSUB_SUBSCRIBERS_PREFIX) {
			sub := &SubscriberData{}
			if err = ms.ms.Unmarshal(value, sub); err == nil {
				result[key[len(utils.PUBSUB_SUBSCRIBERS_PREFIX):]] = sub
			}
		}
	}
	return
}
func (ms *MapStorage) SetSubscriber(key string, sub *SubscriberData) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(sub)
	ms.dict[utils.PUBSUB_SUBSCRIBERS_PREFIX+key] = result
	return
}

func (ms *MapStorage) RemoveSubscriber(key string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.dict, utils.PUBSUB_SUBSCRIBERS_PREFIX+key)
	return
}

func (ms *MapStorage) SetUser(up *UserProfile) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(up)
	if err != nil {
		return err
	}
	ms.dict[utils.USERS_PREFIX+up.GetId()] = result
	return nil
}
func (ms *MapStorage) GetUser(key string) (up *UserProfile, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	up = &UserProfile{}
	if values, ok := ms.dict[utils.USERS_PREFIX+key]; ok {
		err = ms.ms.Unmarshal(values, &up)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) GetUsers() (result []*UserProfile, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	for key, value := range ms.dict {
		if strings.HasPrefix(key, utils.USERS_PREFIX) {
			up := &UserProfile{}
			if err = ms.ms.Unmarshal(value, up); err == nil {
				result = append(result, up)
			}
		}
	}
	return
}

func (ms *MapStorage) RemoveUser(key string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.dict, utils.USERS_PREFIX+key)
	return nil
}

func (ms *MapStorage) GetAlias(key string) (al *Alias, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	key = utils.ALIASES_PREFIX + key
	if x, ok := CacheGet(key); ok {
		if x != nil {
			al = &Alias{Values: x.(AliasValues)}
			al.SetId(origKey)
			return al, nil
		}
		return nil, utils.ErrNotFound
	}
	if values, ok := ms.dict[key]; ok {
		al = &Alias{Values: make(AliasValues, 0)}
		al.SetId(key[len(utils.ALIASES_PREFIX):])
		err = ms.ms.Unmarshal(values, &al.Values)
		if err == nil {
			CacheSet(key, al.Values)
			al.SetReverseCache()
		}
	} else {
		CacheSet(key, nil)
		return nil, utils.ErrNotFound
	}
	return al, nil
}

func (ms *MapStorage) SetAlias(al *Alias, cache bool) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(al.Values)
	if err != nil {
		return err
	}
	ms.dict[utils.ALIASES_PREFIX+al.GetId()] = result
	if cache && err == nil {
		CacheSet(key, al.Values)
		al.SetReverseCache()
	}
	return nil
}

func (ms *MapStorage) RemoveAlias(key string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	al := &Alias{}
	al.SetId(key)
	key = utils.ALIASES_PREFIX + key
	aliasValues := make(AliasValues, 0)
	if values, ok := ms.dict[key]; ok {
		ms.ms.Unmarshal(values, &aliasValues)
	}
	al.Values = aliasValues
	delete(ms.dict, key)
	al.RemoveReverseCache()
	CacheRemKey(key)
	return nil
}

func (ms *MapStorage) GetLoadHistory(limitItems int) ([]*utils.LoadInstance, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return nil, nil
}

func (ms *MapStorage) AddLoadHistory(*utils.LoadInstance, int, bool) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return nil
}

func (ms *MapStorage) GetActionTriggers(key string) (atrs ActionTriggers, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	key = utils.ACTION_TRIGGER_PREFIX + key
	if x, ok := CacheGet(key); ok {
		if x != nil {
			return x.(ActionTriggers), nil
		}
		return nil, utils.ErrNotFound
	}
	if values, ok := ms.dict[key]; ok {
		err = ms.ms.Unmarshal(values, &atrs)
	} else {
		CacheSet(key, nil)
		return nil, utils.ErrNotFound
	}
	CacheSet(key, atrs)
	return
}

func (ms *MapStorage) SetActionTriggers(key string, atrs ActionTriggers, cache bool) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if len(atrs) == 0 {
		// delete the key
		delete(ms.dict, utils.ACTION_TRIGGER_PREFIX+key)
		return
	}
	result, err := ms.ms.Marshal(&atrs)
	ms.dict[utils.ACTION_TRIGGER_PREFIX+key] = result
	if cache && err == nil {
		CacheSet(utils.ACTION_TRIGGER_PREFIX+key, atrs)
	}
	return
}

func (ms *MapStorage) RemoveActionTriggers(key string) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.dict, utils.ACTION_TRIGGER_PREFIX+key)
	CacheRemKey(key)
	return
}

func (ms *MapStorage) GetActionPlan(key string) (ats *ActionPlan, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	key = utils.ACTION_PLAN_PREFIX + key
	if x, ok := CacheGet(key); ok {
		if x != nil {
			return x.(*ActionPlan), nil
		}
		return nil, utils.ErrNotFond
	}
	if values, ok := ms.dict[key]; ok {
		err = ms.ms.Unmarshal(values, &ats)
	} else {
		CacheSet(key, nil)
		return nil, utils.ErrNotFound
	}
	CacheSet(key, ats)
	return
}

func (ms *MapStorage) SetActionPlan(key string, ats *ActionPlan, overwrite, cache bool) (err error) {
	if len(ats.ActionTimings) == 0 {
		ms.mu.Lock()
		defer ms.mu.Unlock()
		// delete the key
		delete(ms.dict, utils.ACTION_PLAN_PREFIX+key)
		CacheRemKey(utils.ACTION_PLAN_PREFIX + key)
		return
	}
	if !overwrite {
		// get existing action plan to merge the account ids
		if existingAts, _ := ms.GetActionPlan(key, true); existingAts != nil {
			if ats.AccountIDs == nil && len(existingAts.AccountIDs) > 0 {
				ats.AccountIDs = make(utils.StringMap)
			}
			for accID := range existingAts.AccountIDs {
				ats.AccountIDs[accID] = true
			}
		}
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(&ats)
	ms.dict[utils.ACTION_PLAN_PREFIX+key] = result
	if cache && err == nil {
		CacheSet(utils.ACTION_PLAN_PREFIX+key, at)
	}
	return
}

func (ms *MapStorage) GetAllActionPlans() (ats map[string]*ActionPlan, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	apls, err := CacheGetAllEntries(utils.ACTION_PLAN_PREFIX)
	if err != nil {
		return nil, err
	}

	ats = make(map[string]*ActionPlan, len(apls))
	for key, value := range apls {
		apl := value.(*ActionPlan)
		ats[key] = apl
	}

	return
}

func (ms *MapStorage) PushTask(t *Task) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(t)
	if err != nil {
		return err
	}
	ms.tasks = append(ms.tasks, result)
	return nil
}

func (ms *MapStorage) PopTask() (t *Task, err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if len(ms.tasks) > 0 {
		var values []byte
		values, ms.tasks = ms.tasks[0], ms.tasks[1:]
		t = &Task{}
		err = ms.ms.Unmarshal(values, t)
	} else {
		err = utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) GetDerivedChargers(key string) (dcs *utils.DerivedChargers, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	key = utils.DERIVEDCHARGERS_PREFIX + key
	if x, ok := CacheGet(key); ok {
		if x != nil {
			return x.(*utils.DerivedChargers), nil
		}
		return nil, utils.ErrNotFound
	}
	if values, ok := ms.dict[key]; ok {
		err = ms.ms.Unmarshal(values, &dcs)
	} else {
		CacheSet(key, nil)
		return nil, utils.ErrNotFound
	}
	CacheSet(key, dcs)
	return
}

func (ms *MapStorage) SetDerivedChargers(key string, dcs *utils.DerivedChargers, cache bool) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if dcs == nil || len(dcs.Chargers) == 0 {
		delete(ms.dict, utils.DERIVEDCHARGERS_PREFIX+key)
		CacheRemKey(utils.DERIVEDCHARGERS_PREFIX + key)
		return nil
	}
	result, err := ms.ms.Marshal(dcs)
	ms.dict[utils.DERIVEDCHARGERS_PREFIX+key] = result
	if cache && err == nil {
		CacheSet(key, dcs)
	}
	return err
}

func (ms *MapStorage) SetCdrStats(cs *CdrStats) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(cs)
	ms.dict[utils.CDR_STATS_PREFIX+cs.Id] = result
	return err
}

func (ms *MapStorage) GetCdrStats(key string) (cs *CdrStats, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if values, ok := ms.dict[utils.CDR_STATS_PREFIX+key]; ok {
		err = ms.ms.Unmarshal(values, &cs)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) GetAllCdrStats() (css []*CdrStats, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	for key, value := range ms.dict {
		if !strings.HasPrefix(key, utils.CDR_STATS_PREFIX) {
			continue
		}
		cs := &CdrStats{}
		err = ms.ms.Unmarshal(value, cs)
		css = append(css, cs)
	}
	return
}

func (ms *MapStorage) SetSMCost(smCost *SMCost) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	result, err := ms.ms.Marshal(smCost)
	ms.dict[utils.LOG_CALL_COST_PREFIX+smCost.CostSource+smCost.RunID+"_"+smCost.CGRID] = result
	return err
}

func (ms *MapStorage) GetSMCost(cgrid, source, runid, originHost, originID string) (smCost *SMCost, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if values, ok := ms.dict[utils.LOG_CALL_COST_PREFIX+source+runid+"_"+cgrid]; ok {
		err = ms.ms.Unmarshal(values, &smCost)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) SetStructVersion(v *StructVersion) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	var result []byte
	result, err = ms.ms.Marshal(v)
	if err != nil {
		return
	}
	ms.dict[utils.VERSION_PREFIX+"struct"] = result
	return
}

func (ms *MapStorage) GetStructVersion() (rsv *StructVersion, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	rsv = &StructVersion{}
	if values, ok := ms.dict[utils.VERSION_PREFIX+"struct"]; ok {
		err = ms.ms.Unmarshal(values, &rsv)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) GetResourceLimit(id string, skipCache bool) (*ResourceLimit, error) {
	return nil, nil
}
func (ms *MapStorage) SetResourceLimit(rl *ResourceLimit) error {
	return nil
}
func (ms *MapStorage) RemoveResourceLimit(id string) error {
	return nil
}
