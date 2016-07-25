/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

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
	"io/ioutil"

	"github.com/cgrates/cgrates/utils"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
)

var (
	ErrRedisNotFound = errors.New("RedisNotFound")
)

type RedisStorage struct {
	db              *pool.Pool
	ms              Marshaler
	cacheDumpDir    string
	loadHistorySize int
}

func NewRedisStorage(address string, db int, pass, mrshlerStr string, maxConns int, cacheDumpDir string, loadHistorySize int) (*RedisStorage, error) {
	df := func(network, addr string) (*redis.Client, error) {
		client, err := redis.Dial(network, addr)
		if err != nil {
			return nil, err
		}
		if len(pass) != 0 {
			if err = client.Cmd("AUTH", pass).Err; err != nil {
				client.Close()
				return nil, err
			}
		}
		if db != 0 {
			if err = client.Cmd("SELECT", db).Err; err != nil {
				client.Close()
				return nil, err
			}
		}
		return client, nil
	}
	p, err := pool.NewCustom("tcp", address, maxConns, df)
	if err != nil {
		return nil, err
	}
	var mrshler Marshaler
	if mrshlerStr == utils.MSGPACK {
		mrshler = NewCodecMsgpackMarshaler()
	} else if mrshlerStr == utils.JSON {
		mrshler = new(JSONMarshaler)
	} else {
		return nil, fmt.Errorf("Unsupported marshaler: %v", mrshlerStr)
	}
	if cacheDumpDir != "" {
		if err := CacheSetDumperPath(cacheDumpDir); err != nil {
			utils.Logger.Info("<cache dumper> init error: " + err.Error())
		}
	}
	return &RedisStorage{db: p, ms: mrshler, cacheDumpDir: cacheDumpDir, loadHistorySize: loadHistorySize}, nil
}

func (rs *RedisStorage) Close() {
	rs.db.Empty()
}

func (rs *RedisStorage) Flush(ignore string) error {
	return rs.db.Cmd("FLUSHDB").Err
}

func (rs *RedisStorage) GetKeysForPrefix(prefix string) ([]string, error) {
	if x, ok := CacheGet(utils.GENERIC_PREFIX + "keys_" + prefix); ok {
		if x != nil {
			return x.([]string), nil
		}
		return nil, utils.ErrNotFound
	}
	r := rs.db.Cmd("KEYS", prefix+"*")
	if r.Err != nil {
		CacheSet(utils.GENERIC_PREFIX+"keys_"+prefix, nil)
		return nil, r.Err
	}
	CacheSet(utils.GENERIC_PREFIX+"keys_"+prefix, r.List())
	return r.List()
}

// Used to check if specific subject is stored using prefix key attached to entity
func (rs *RedisStorage) HasData(category, subject string) (bool, error) {
	switch category {
	case utils.DESTINATION_PREFIX, utils.RATING_PLAN_PREFIX, utils.RATING_PROFILE_PREFIX, utils.ACTION_PREFIX, utils.ACTION_PLAN_PREFIX, utils.ACCOUNT_PREFIX, utils.DERIVEDCHARGERS_PREFIX:
		if x, ok := CacheGet(utils.GENERIC_PREFIX + "has_" + category + "_" + subject); ok {
			return x.(bool), nil
		}
		i, err := rs.db.Cmd("EXISTS", category+subject).Int()
		CacheSet(utils.GENERIC_PREFIX+"has_"+category+"_"+subject, i == 1)
		return i == 1, err
	}
	CacheSet(utils.GENERIC_PREFIX+"has_"+category+"_"+subject, false)
	return false, errors.New("unsupported HasData category")
}

func (rs *RedisStorage) GetRatingPlan(key string) (rp *RatingPlan, err error) {
	key = utils.RATING_PLAN_PREFIX + key

	if x, ok := CacheGet(key); ok {
		if x != nil {
			return x.(*RatingPlan), nil
		}
		return nil, utils.ErrNotFound
	}
	var values []byte
	if values, err = rs.db.Cmd("GET", key).Bytes(); err == nil {
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
	}
	CacheSet(key, rp)
	return
}

func (rs *RedisStorage) SetRatingPlan(rp *RatingPlan, cache bool) (err error) {
	result, err := rs.ms.Marshal(rp)
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	err = rs.db.Cmd("SET", utils.RATING_PLAN_PREFIX+rp.Id, b.Bytes()).Err
	if err == nil && historyScribe != nil {
		response := 0
		go historyScribe.Call("HistoryV1.Record", rp.GetHistoryRecord(), &response)
	}
	if cache && err == nil {
		CacheSet(utils.RATING_PLAN_PREFIX+rp.Id, rp)
	}
	return
}

func (rs *RedisStorage) GetRatingProfile(key string) (rpf *RatingProfile, err error) {
	key = utils.RATING_PROFILE_PREFIX + key

	if x, ok := CacheGet(key); ok {
		if x != nil {
			return x.(*RatingProfile), nil
		}
		return nil, utils.ErrNotFound
	}
	var values []byte
	if values, err = rs.db.Cmd("GET", key).Bytes(); err == nil {
		rpf = new(RatingProfile)
		err = rs.ms.Unmarshal(values, rpf)
	}
	CacheSet(key, rpf)
	return
}

func (rs *RedisStorage) SetRatingProfile(rpf *RatingProfile, cache bool) (err error) {
	result, err := rs.ms.Marshal(rpf)
	err = rs.db.Cmd("SET", utils.RATING_PROFILE_PREFIX+rpf.Id, result).Err
	if err == nil && historyScribe != nil {
		response := 0
		go historyScribe.Call("HistoryV1.Record", rpf.GetHistoryRecord(false), &response)
	}
	if cache && err == nil {
		CacheSet(utils.RATING_PROFILE_PREFIX+rpf.Id, rpf)
	}
	return
}

func (rs *RedisStorage) RemoveRatingProfile(key string) error {
	conn, err := rs.db.Get()
	if err != nil {
		return err
	}
	defer rs.db.Put(conn)
	keys, err := conn.Cmd("KEYS", utils.RATING_PROFILE_PREFIX+key+"*").List()
	if err != nil {
		return err
	}
	for _, key := range keys {
		if err = conn.Cmd("DEL", key).Err; err != nil {
			return err
		}
		CacheRemKey(key)
		rpf := &RatingProfile{Id: key}
		if historyScribe != nil {
			response := 0
			go historyScribe.Call("HistoryV1.Record", rpf.GetHistoryRecord(true), &response)
		}
	}
	return nil
}

func (rs *RedisStorage) GetLCR(key string) (lcr *LCR, err error) {
	key = utils.LCR_PREFIX + key
	if x, ok := CacheGet(key); ok {
		if x != nil {
			return x.(*LCR), nil
		}
		return nil, utils.ErrNotFound
	}
	var values []byte
	if values, err = rs.db.Cmd("GET", key).Bytes(); err == nil {
		err = rs.ms.Unmarshal(values, &lcr)
	}
	CacheSet(key, lcr)
	return
}

func (rs *RedisStorage) SetLCR(lcr *LCR, cache bool) (err error) {
	result, err := rs.ms.Marshal(lcr)
	err = rs.db.Cmd("SET", utils.LCR_PREFIX+lcr.GetId(), result).Err
	if cache && err == nil {
		CacheSet(utils.LCR_PREFIX+lcr.GetId(), lcr)
	}
	return
}
func (rs *RedisStorage) GetDestination(key string) (dest *Destination, err error) {
	key = utils.DESTINATION_PREFIX + key
	var values []byte
	if values, err = rs.db.Cmd("GET", key).Bytes(); len(values) > 0 && err == nil {
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
		// create optimized structure
		for _, p := range dest.Prefixes {
			CachePush(utils.DESTINATION_PREFIX+p, dest.Id)
		}
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (rs *RedisStorage) SetDestination(dest *Destination, cache bool) (err error) {
	result, err := rs.ms.Marshal(dest)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	err = rs.db.Cmd("SET", utils.DESTINATION_PREFIX+dest.Id, b.Bytes()).Err
	if err == nil && historyScribe != nil {
		response := 0
		go historyScribe.Call("HistoryV1.Record", dest.GetHistoryRecord(false), &response)
	}
	if cache && err == nil {
		CacheSet(utils.DESTINATION_PREFIX+dest.Id, dest)
	}
	return
}

func (rs *RedisStorage) GetDestinationIDs(prefix string) (ids []string, err error) {
	prefix = utils.DESTINATION_PREFIX + prefix
	var values []string
	if values, err = rs.db.Cmd("SMEMBERS", prefix).List(); len(values) > 0 && err == nil {
		CachePush(utils.DESTINATION_PREFIX+prefix, values...)
	} else {
		return nil, utils.ErrNotFound
	}
	return
}

func (rs *RedisStorage) SetDestinationIDs(dest *Destination) (err error) {
	for _, p := range dest.Prefixes {
		err = rs.db.Cmd("SADD", utils.DESTINATION_PREFIX+p, dest.Id).Err
		if err != nil {
			break
		}
	}

	if err == nil && historyScribe != nil {
		response := 0
		go historyScribe.Call("HistoryV1.Record", dest.GetHistoryRecord(false), &response)
	}
	return
}

func (rs *RedisStorage) RemoveDestination(destID string) (err error) {
	/*conn, err := rs.db.Get()
	if err != nil {
		return err
	}
	var dest *Destination
	defer rs.db.Put(conn)
	if values, err = conn.Cmd("GET", key).Bytes(); len(values) > 0 && err == nil {
		b := bytes.NewBuffer(values)
		r, err := zlib.NewReader(b)
		if err != nil {
			return err
		}
		out, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		r.Close()
		dest = new(Destination)
		err = rs.ms.Unmarshal(out, dest)
	} else {
		return utils.ErrNotFound
	}
	key := utils.DESTINATION_PREFIX + destID
	if err = conn.Cmd("DEL", key).Err; err != nil {
		return err
	}
	if dest != nil {
		for _, prefix := range dest.Prefixes {
			changed := false
			if idIDs, err := CacheGet(utils.DESTINATION_PREFIX + prefix); err == nil {
				dIDs := idIDs.(map[interface{}]struct{})
				if len(dIDs) == 1 {
					// remove de prefix from cache
					CacheRemKey(utils.DESTINATION_PREFIX + prefix)
				} else {
					// delete the destination from list and put the new list in chache
					delete(dIDs, searchedDID)
					changed = true
				}
			}
			if changed {
				CacheSet(utils.DESTINATION_PREFIX+prefix, dIDs)
			}
		}
	}
	dest := &Destination{Id: key}
	if historyScribe != nil {
		response := 0
		go historyScribe.Call("HistoryV1.Record", dest.GetHistoryRecord(true), &response)
	}*/

	return
}

func (rs *RedisStorage) GetActions(key string) (as Actions, err error) {
	key = utils.ACTION_PREFIX + key
	if x, ok := CacheGet(key); ok {
		if x != nil {
			return x.(Actions), nil
		}
		return nil, utils.ErrNotFound
	}
	var values []byte
	if values, err = rs.db.Cmd("GET", key).Bytes(); err == nil {
		err = rs.ms.Unmarshal(values, &as)
	}
	CacheSet(key, as)
	return
}

func (rs *RedisStorage) SetActions(key string, as Actions, cache bool) (err error) {
	result, err := rs.ms.Marshal(&as)
	err = rs.db.Cmd("SET", utils.ACTION_PREFIX+key, result).Err
	if cache && err == nil {
		CacheSet(utils.ACTION_PREFIX+key, as)
	}
	return
}

func (rs *RedisStorage) RemoveActions(key string) (err error) {
	err = rs.db.Cmd("DEL", utils.ACTION_PREFIX+key).Err
	CacheRemKey(utils.ACTION_PREFIX + key)
	return
}

func (rs *RedisStorage) GetSharedGroup(key string) (sg *SharedGroup, err error) {
	key = utils.SHARED_GROUP_PREFIX + key
	if x, ok := CacheGet(key); ok {
		if x != nil {
			return x.(*SharedGroup), nil
		}
		return nil, utils.ErrNotFound
	}
	var values []byte
	if values, err = rs.db.Cmd("GET", key).Bytes(); err == nil {
		err = rs.ms.Unmarshal(values, &sg)
	}
	CacheSet(key, sg)
	return
}

func (rs *RedisStorage) SetSharedGroup(sg *SharedGroup, cache bool) (err error) {
	result, err := rs.ms.Marshal(sg)
	err = rs.db.Cmd("SET", utils.SHARED_GROUP_PREFIX+sg.Id, result).Err
	if cache && err == nil {
		CacheSet(utils.SHARED_GROUP_PREFIX+sg.Id, sg)
	}
	return
}

func (rs *RedisStorage) GetAccount(key string) (*Account, error) {
	rpl := rs.db.Cmd("GET", utils.ACCOUNT_PREFIX+key)
	if rpl.Err != nil {
		return nil, rpl.Err
	} else if rpl.IsType(redis.Nil) {
		return nil, ErrRedisNotFound
	}
	values, err := rpl.Bytes()
	if err != nil {
		return nil, err
	}
	ub := &Account{ID: key}
	if err = rs.ms.Unmarshal(values, ub); err != nil {
		return nil, err
	}
	return ub, nil
}

func (rs *RedisStorage) SetAccount(ub *Account) (err error) {
	// never override existing account with an empty one
	// UPDATE: if all balances expired and were cleaned it makes
	// sense to write empty balance map
	if len(ub.BalanceMap) == 0 {
		if ac, err := rs.GetAccount(ub.ID); err == nil && !ac.allBalancesExpired() {
			ac.ActionTriggers = ub.ActionTriggers
			ac.UnitCounters = ub.UnitCounters
			ac.AllowNegative = ub.AllowNegative
			ac.Disabled = ub.Disabled
			ub = ac
		}
	}
	result, err := rs.ms.Marshal(ub)
	err = rs.db.Cmd("SET", utils.ACCOUNT_PREFIX+ub.ID, result).Err
	return
}

func (rs *RedisStorage) RemoveAccount(key string) (err error) {
	return rs.db.Cmd("DEL", utils.ACCOUNT_PREFIX+key).Err

}

func (rs *RedisStorage) GetCdrStatsQueue(key string) (sq *StatsQueue, err error) {
	var values []byte
	if values, err = rs.db.Cmd("GET", utils.CDR_STATS_QUEUE_PREFIX+key).Bytes(); err == nil {
		sq = &StatsQueue{}
		err = rs.ms.Unmarshal(values, &sq)
	}
	return
}

func (rs *RedisStorage) SetCdrStatsQueue(sq *StatsQueue) (err error) {
	result, err := rs.ms.Marshal(sq)
	err = rs.db.Cmd("SET", utils.CDR_STATS_QUEUE_PREFIX+sq.GetId(), result).Err
	return
}

func (rs *RedisStorage) GetSubscribers() (result map[string]*SubscriberData, err error) {
	conn, err := rs.db.Get()
	if err != nil {
		return nil, err
	}
	defer rs.db.Put(conn)
	keys, err := conn.Cmd("KEYS", utils.PUBSUB_SUBSCRIBERS_PREFIX+"*").List()
	if err != nil {
		return nil, err
	}
	result = make(map[string]*SubscriberData)
	for _, key := range keys {
		if values, err := conn.Cmd("GET", key).Bytes(); err == nil {
			sub := &SubscriberData{}
			err = rs.ms.Unmarshal(values, sub)
			result[key[len(utils.PUBSUB_SUBSCRIBERS_PREFIX):]] = sub
		} else {
			return nil, utils.ErrNotFound
		}
	}
	return
}

func (rs *RedisStorage) SetSubscriber(key string, sub *SubscriberData) (err error) {
	result, err := rs.ms.Marshal(sub)
	if err != nil {
		return err
	}
	return rs.db.Cmd("SET", utils.PUBSUB_SUBSCRIBERS_PREFIX+key, result).Err
}

func (rs *RedisStorage) RemoveSubscriber(key string) (err error) {
	err = rs.db.Cmd("DEL", utils.PUBSUB_SUBSCRIBERS_PREFIX+key).Err
	return
}

func (rs *RedisStorage) SetUser(up *UserProfile) (err error) {
	result, err := rs.ms.Marshal(up)
	if err != nil {
		return err
	}
	return rs.db.Cmd("SET", utils.USERS_PREFIX+up.GetId(), result).Err
}

func (rs *RedisStorage) GetUser(key string) (up *UserProfile, err error) {
	var values []byte
	if values, err = rs.db.Cmd("GET", utils.USERS_PREFIX+key).Bytes(); err == nil {
		up = &UserProfile{}
		err = rs.ms.Unmarshal(values, &up)
	}
	return
}

func (rs *RedisStorage) GetUsers() (result []*UserProfile, err error) {
	conn, err := rs.db.Get()
	if err != nil {
		return nil, err
	}
	defer rs.db.Put(conn)
	keys, err := conn.Cmd("KEYS", utils.USERS_PREFIX+"*").List()
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		if values, err := conn.Cmd("GET", key).Bytes(); err == nil {
			up := &UserProfile{}
			err = rs.ms.Unmarshal(values, up)
			result = append(result, up)
		} else {
			return nil, utils.ErrNotFound
		}
	}
	return
}

func (rs *RedisStorage) RemoveUser(key string) (err error) {
	return rs.db.Cmd("DEL", utils.USERS_PREFIX+key).Err
}

func (rs *RedisStorage) GetAlias(key string) (al *Alias, err error) {
	origKey := key
	key = utils.ALIASES_PREFIX + key

	if x, ok := CacheGet(key); ok {
		if x != nil {
			al = &Alias{Values: x.(AliasValues)}
			al.SetId(origKey)
			return al, nil
		}
		return nil, utils.ErrNotFound
	}
	var values []byte
	if values, err = rs.db.Cmd("GET", key).Bytes(); err == nil {
		al = &Alias{Values: make(AliasValues, 0)}
		al.SetId(origKey)
		err = rs.ms.Unmarshal(values, &al.Values)
		if err == nil {
			CacheSet(key, al.Values)
			// cache reverse alias
			al.SetReverseCache()
		}
	} else {
		CacheSet(key, nil)
	}
	return
}

func (rs *RedisStorage) SetAlias(al *Alias, cache bool) (err error) {
	result, err := rs.ms.Marshal(al.Values)
	if err != nil {
		return err
	}
	err = rs.db.Cmd("SET", utils.ALIASES_PREFIX+al.GetId(), result).Err
	if cache && err == nil {
		CacheSet(key, al.Values)
		al.SetReverseCache()
	}
	return
}

func (rs *RedisStorage) RemoveAlias(key string) (err error) {
	conn, err := rs.db.Get()
	if err != nil {
		return err
	}
	defer rs.db.Put(conn)
	al := &Alias{}
	al.SetId(key)
	key = utils.ALIASES_PREFIX + key
	aliasValues := make(AliasValues, 0)
	if values, err := conn.Cmd("GET", key).Bytes(); err == nil {
		rs.ms.Unmarshal(values, &aliasValues)
	}
	al.Values = aliasValues
	err = conn.Cmd("DEL", key).Err
	if err == nil {
		al.RemoveReverseCache()
		CacheRemKey(key)
	}
	return
}

// Limit will only retrieve the last n items out of history, newest first
func (rs *RedisStorage) GetLoadHistory(limit int) ([]*utils.LoadInstance, error) {
	if limit == 0 {
		return nil, nil
	}

	if x, ok := CacheGet(utils.LOADINST_KEY); ok {
		if x != nil {
			items := x.([]*utils.LoadInstance)
			if len(items) < limit || limit == -1 {
				return items, nil
			}
			return items[:limit], nil
		}
		return nil, utils.ErrNotFound
	}
	if limit != -1 {
		limit -= -1 // Decrease limit to match redis approach on lrange
	}
	marshaleds, err := rs.db.Cmd("LRANGE", utils.LOADINST_KEY, 0, limit).ListBytes()
	if err != nil {
		CacheSet(utils.LOADINST_KEY, nil)
		return nil, err
	}
	loadInsts := make([]*utils.LoadInstance, len(marshaleds))
	for idx, marshaled := range marshaleds {
		var lInst utils.LoadInstance
		err = rs.ms.Unmarshal(marshaled, &lInst)
		if err != nil {
			return nil, err
		}
		loadInsts[idx] = &lInst
	}
	CacheRemKey(utils.LOADINST_KEY)
	CacheSet(utils.LOADINST_KEY, loadInsts)
	if len(loadInsts) < limit || limit == -1 {
		return loadInsts, nil
	}
	return loadInsts[:limit], nil
}

// Adds a single load instance to load history
func (rs *RedisStorage) AddLoadHistory(ldInst *utils.LoadInstance, loadHistSize int, cache bool) error {
	conn, err := rs.db.Get()
	if err != nil {
		return err
	}
	defer rs.db.Put(conn)
	if loadHistSize == 0 { // Load history disabled
		return nil
	}
	marshaled, err := rs.ms.Marshal(&ldInst)
	if err != nil {
		return err
	}
	_, err = Guardian.Guard(func() (interface{}, error) { // Make sure we do it locked since other instance can modify history while we read it
		histLen, err := conn.Cmd("LLEN", utils.LOADINST_KEY).Int()
		if err != nil {
			return nil, err
		}
		if histLen >= loadHistSize { // Have hit maximum history allowed, remove oldest element in order to add new one
			if err := conn.Cmd("RPOP", utils.LOADINST_KEY).Err; err != nil {
				return nil, err
			}
		}
		err = conn.Cmd("LPUSH", utils.LOADINST_KEY, marshaled).Err
		return nil, err
	}, 0, utils.LOADINST_KEY)

	if cache && err == nil {
		CacheSet(utils.LOADINST_KEY, loadInsts)
	}
	return err
}

func (rs *RedisStorage) GetActionTriggers(key string) (atrs ActionTriggers, err error) {
	key = utils.ACTION_TRIGGER_PREFIX + key
	if x, ok := CacheGet(key); ok {
		if x != nil {
			return x.(ActionTriggers), nil
		}
		return nil, utils.ErrNotFound
	}
	var values []byte
	if values, err = rs.db.Cmd("GET", key).Bytes(); err == nil {
		err = rs.ms.Unmarshal(values, &atrs)
	}
	CacheSet(key, atrs)
	return
}

func (rs *RedisStorage) SetActionTriggers(key string, atrs ActionTriggers, cache bool) (err error) {
	conn, err := rs.db.Get()
	if err != nil {
		return err
	}
	defer rs.db.Put(conn)
	if len(atrs) == 0 {
		// delete the key
		return conn.Cmd("DEL", utils.ACTION_TRIGGER_PREFIX+key).Err
	}
	result, err := rs.ms.Marshal(atrs)
	if err != nil {
		return err
	}
	err = conn.Cmd("SET", utils.ACTION_TRIGGER_PREFIX+key, result).Err
	if cache && err == nil {
		CacheSet(utils.ACTION_TRIGGER_PREFIX+key, atrs)
	}
	return
}

func (rs *RedisStorage) RemoveActionTriggers(key string) (err error) {
	key = utils.ACTION_TRIGGER_PREFIX + key
	err = rs.db.Cmd("DEL", key).Err
	if err == nil {
		CacheRemKey(key)
	}
	return
}

func (rs *RedisStorage) GetActionPlan(key string) (ats *ActionPlan, err error) {
	key = utils.ACTION_PLAN_PREFIX + key

	if x, ok := CacheGet(key); ok {
		if x != nil {
			return x.(*ActionPlan), nil
		}
		return nil, utils.ErrNotFond
	}
	var values []byte
	if values, err = rs.db.Cmd("GET", key).Bytes(); err == nil {
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
		err = rs.ms.Unmarshal(out, &ats)
	}
	CacheSet(key, ats)
	return
}

func (rs *RedisStorage) SetActionPlan(key string, ats *ActionPlan, overwrite, cache bool) (err error) {
	if len(ats.ActionTimings) == 0 {
		// delete the key
		err = rs.db.Cmd("DEL", utils.ACTION_PLAN_PREFIX+key).Err
		CacheRemKey(utils.ACTION_PLAN_PREFIX + key)
		return err
	}
	if !overwrite {
		// get existing action plan to merge the account ids
		if existingAts, _ := rs.GetActionPlan(key, true); existingAts != nil {
			if ats.AccountIDs == nil && len(existingAts.AccountIDs) > 0 {
				ats.AccountIDs = make(utils.StringMap)
			}
			for accID := range existingAts.AccountIDs {
				ats.AccountIDs[accID] = true
			}
		}
	}

	result, err := rs.ms.Marshal(ats)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	err = rs.db.Cmd("SET", utils.ACTION_PLAN_PREFIX+key, b.Bytes()).Err
	if cache && err == nil {
		CacheSet(utils.ACTION_PLAN_PREFIX+key, at)
	}
	return
}

func (rs *RedisStorage) GetAllActionPlans() (ats map[string]*ActionPlan, err error) {
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

func (rs *RedisStorage) PushTask(t *Task) error {
	result, err := rs.ms.Marshal(t)
	if err != nil {
		return err
	}
	return rs.db.Cmd("RPUSH", utils.TASKS_KEY, result).Err
}

func (rs *RedisStorage) PopTask() (t *Task, err error) {
	var values []byte
	if values, err = rs.db.Cmd("LPOP", utils.TASKS_KEY).Bytes(); err == nil {
		t = &Task{}
		err = rs.ms.Unmarshal(values, t)
	}
	return
}

func (rs *RedisStorage) GetDerivedChargers(key string) (dcs *utils.DerivedChargers, err error) {
	key = utils.DERIVEDCHARGERS_PREFIX + key
	if x, ok := CacheGet(key); ok {
		if x != nil {
			return x.(*utils.DerivedChargers), nil
		}
		return nil, utils.ErrNotFound
	}
	var values []byte
	if values, err = rs.db.Cmd("GET", key).Bytes(); err == nil {
		err = rs.ms.Unmarshal(values, &dcs)
	}
	CacheSet(key, dcs)
	return dcs, err
}

func (rs *RedisStorage) SetDerivedChargers(key string, dcs *utils.DerivedChargers, cache bool) (err error) {
	key = utils.DERIVEDCHARGERS_PREFIX + key
	if dcs == nil || len(dcs.Chargers) == 0 {
		err = rs.db.Cmd("DEL").Err
		CacheRemKey(key)
		return err
	}
	marshaled, err := rs.ms.Marshal(dcs)
	if err != nil {
		return err
	}
	err = rs.db.Cmd("SET", key, marshaled).Err
	if cache && err == nil {
		CacheSet(key, dcs)
	}
	return
}

func (rs *RedisStorage) SetCdrStats(cs *CdrStats) error {
	marshaled, err := rs.ms.Marshal(cs)
	if err != nil {
		return err
	}
	return rs.db.Cmd("SET", utils.CDR_STATS_PREFIX+cs.Id, marshaled).Err
}

func (rs *RedisStorage) GetCdrStats(key string) (cs *CdrStats, err error) {
	var values []byte
	if values, err = rs.db.Cmd("GET", utils.CDR_STATS_PREFIX+key).Bytes(); err == nil {
		err = rs.ms.Unmarshal(values, &cs)
	}
	return
}

func (rs *RedisStorage) GetAllCdrStats() (css []*CdrStats, err error) {
	conn, err := rs.db.Get()
	if err != nil {
		return nil, err
	}
	defer rs.db.Put(conn)
	keys, err := conn.Cmd("KEYS", utils.CDR_STATS_PREFIX+"*").List()
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		value, err := conn.Cmd("GET", key).Bytes()
		if err != nil {
			continue
		}
		cs := &CdrStats{}
		err = rs.ms.Unmarshal(value, cs)
		css = append(css, cs)
	}
	return
}

func (rs *RedisStorage) SetStructVersion(v *StructVersion) (err error) {
	var result []byte
	result, err = rs.ms.Marshal(v)
	if err != nil {
		return
	}
	return rs.db.Cmd("SET", utils.VERSION_PREFIX+"struct", result).Err
}

func (rs *RedisStorage) GetStructVersion() (rsv *StructVersion, err error) {
	var values []byte
	rsv = &StructVersion{}
	if values, err = rs.db.Cmd("GET", utils.VERSION_PREFIX+"struct").Bytes(); err == nil {
		err = rs.ms.Unmarshal(values, &rsv)
	}
	return
}
