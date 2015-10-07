/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"log"
	"strings"
	"time"

	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/utils"

	"gopkg.in/mgo.v2"

	"gopkg.in/mgo.v2/bson"
)

const (
	colDst = "destinations"
	colAct = "actions"
	colApl = "actionplans"
	colAtr = "actiontriggers"
	colRpl = "ratingplans"
	colRpf = "ratingprofiles"
	colAcc = "accounts"
	colShg = "sharedgroups"
	colLcr = "lcrrules"
	colDcs = "derivedchargers"
	colAls = "aliases"
	colStq = "statsqeues"
	colPbs = "pubsub"
	colUsr = "users"
)

type MongoStorage struct {
	session *mgo.Session
	db      *mgo.Database
}

func NewMongoStorage(address, db, user, pass string) (*MongoStorage, error) {
	if user != "" && pass != "" {
		address = fmt.Sprintf("%s:%s@%s", user, pass, address)
	}
	session, err := mgo.Dial(address)
	if err != nil {
		log.Printf(fmt.Sprintf("Could not connect to logger database: %v", err))
		return nil, err
	}
	ndb := session.DB(db)
	session.SetMode(mgo.Monotonic, true)
	index := mgo.Index{Key: []string{"key"}, Background: true}
	err = ndb.C(colAct).EnsureIndex(index)
	err = ndb.C(colApl).EnsureIndex(index)
	index = mgo.Index{Key: []string{"id"}, Background: true}
	err = ndb.C(colRpf).EnsureIndex(index)
	err = ndb.C(colRpl).EnsureIndex(index)
	err = ndb.C(colDst).EnsureIndex(index)
	err = ndb.C(colAcc).EnsureIndex(index)

	return &MongoStorage{db: ndb, session: session}, nil
}

func (ms *MongoStorage) Close() {
	ms.session.Close()
}

func (ms *MongoStorage) GetKeysForPrefix(prefix string) ([]string, error) {
	return nil, nil
}

func (ms *MongoStorage) Flush(ignore string) (err error) {
	collections, err := ms.db.CollectionNames()
	if err != nil {
		return err
	}
	for _, c := range collections {
		if err = ms.db.C(c).DropCollection(); err != nil {
			return err
		}
	}
	return nil
}

type KeyedContainer struct {
	Key   string
	Value interface{}
}

type LogTimingEntry struct {
	ActionPlan *ActionPlan
	Actions    Actions
	LogTime    time.Time
	Source     string
}

type LogTriggerEntry struct {
	ubId          string
	ActionTrigger *ActionTrigger
	Actions       Actions
	LogTime       time.Time
	Source        string
}

type LogErrEntry struct {
	Id     string `bson:"_id,omitempty"`
	ErrStr string
	Source string
}

func (ms *MongoStorage) CacheRatingAll() error {
	return ms.cacheRating(nil, nil, nil, nil, nil, nil, nil)
}

func (ms *MongoStorage) CacheRatingPrefixes(prefixes ...string) error {
	pm := map[string][]string{
		utils.DESTINATION_PREFIX:     []string{},
		utils.RATING_PLAN_PREFIX:     []string{},
		utils.RATING_PROFILE_PREFIX:  []string{},
		utils.LCR_PREFIX:             []string{},
		utils.DERIVEDCHARGERS_PREFIX: []string{},
		utils.ACTION_PREFIX:          []string{},
		utils.SHARED_GROUP_PREFIX:    []string{},
	}
	for _, prefix := range prefixes {
		if _, found := pm[prefix]; !found {
			return utils.ErrNotFound
		}
		pm[prefix] = nil
	}
	return ms.cacheRating(pm[utils.DESTINATION_PREFIX], pm[utils.RATING_PLAN_PREFIX], pm[utils.RATING_PROFILE_PREFIX], pm[utils.LCR_PREFIX], pm[utils.DERIVEDCHARGERS_PREFIX], pm[utils.ACTION_PREFIX], pm[utils.SHARED_GROUP_PREFIX])
}

func (ms *MongoStorage) CacheRatingPrefixValues(prefixes map[string][]string) error {
	pm := map[string][]string{
		utils.DESTINATION_PREFIX:     []string{},
		utils.RATING_PLAN_PREFIX:     []string{},
		utils.RATING_PROFILE_PREFIX:  []string{},
		utils.LCR_PREFIX:             []string{},
		utils.DERIVEDCHARGERS_PREFIX: []string{},
		utils.ACTION_PREFIX:          []string{},
		utils.SHARED_GROUP_PREFIX:    []string{},
	}
	for prefix, ids := range prefixes {
		if _, found := pm[prefix]; !found {
			return utils.ErrNotFound
		}
		pm[prefix] = ids
	}
	return ms.cacheRating(pm[utils.DESTINATION_PREFIX], pm[utils.RATING_PLAN_PREFIX], pm[utils.RATING_PROFILE_PREFIX], pm[utils.LCR_PREFIX], pm[utils.DERIVEDCHARGERS_PREFIX], pm[utils.ACTION_PREFIX], pm[utils.SHARED_GROUP_PREFIX])
}

func (ms *MongoStorage) cacheRating(dKeys, rpKeys, rpfKeys, lcrKeys, dcsKeys, actKeys, shgKeys []string) (err error) {
	cache2go.BeginTransaction()
	var result struct{ Key string }
	if dKeys == nil || (float64(cache2go.CountEntries(utils.DESTINATION_PREFIX))*utils.DESTINATIONS_LOAD_THRESHOLD < float64(len(dKeys))) {
		// if need to load more than a half of exiting keys load them all
		utils.Logger.Info("Caching all destinations")
		iter := ms.db.C(colDst).Find(nil).Select(bson.M{"key": 1}).Iter()
		dKeys = make([]string, 0)
		for iter.Next(&result) {
			dKeys = append(dKeys, result.Key)
		}
		if err := iter.Close(); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
		cache2go.RemPrefixKey(utils.DESTINATION_PREFIX)
	} else if len(dKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching destinations: %v", dKeys))
		CleanStalePrefixes(dKeys)
	}
	for _, key := range dKeys {
		if len(key) <= len(utils.DESTINATION_PREFIX) {
			utils.Logger.Warning(fmt.Sprintf("Got malformed destination id: %s", key))
			continue
		}
		if _, err = ms.GetDestination(key[len(utils.DESTINATION_PREFIX):]); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	}
	if len(dKeys) != 0 {
		utils.Logger.Info("Finished destinations caching.")
	}
	if rpKeys == nil {
		utils.Logger.Info("Caching all rating plans")
		iter := ms.db.C(colRpl).Find(nil).Select(bson.M{"key": 1}).Iter()
		rpKeys = make([]string, 0)
		for iter.Next(&result) {
			rpKeys = append(rpKeys, result.Key)
		}
		if err := iter.Close(); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
		cache2go.RemPrefixKey(utils.RATING_PLAN_PREFIX)
	} else if len(rpKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching rating plans: %v", rpKeys))
	}
	for _, key := range rpKeys {
		cache2go.RemKey(key)
		if _, err = ms.GetRatingPlan(key[len(utils.RATING_PLAN_PREFIX):], true); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	}
	if len(rpKeys) != 0 {
		utils.Logger.Info("Finished rating plans caching.")
	}
	if rpfKeys == nil {
		utils.Logger.Info("Caching all rating profiles")
		iter := ms.db.C(colRpf).Find(nil).Select(bson.M{"key": 1}).Iter()
		rpfKeys = make([]string, 0)
		for iter.Next(&result) {
			rpfKeys = append(rpfKeys, result.Key)
		}
		if err := iter.Close(); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
		cache2go.RemPrefixKey(utils.RATING_PROFILE_PREFIX)
	} else if len(rpfKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching rating profile: %v", rpfKeys))
	}
	for _, key := range rpfKeys {
		cache2go.RemKey(key)
		if _, err = ms.GetRatingProfile(key[len(utils.RATING_PROFILE_PREFIX):], true); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	}
	if len(rpfKeys) != 0 {
		utils.Logger.Info("Finished rating profile caching.")
	}
	if lcrKeys == nil {
		utils.Logger.Info("Caching LCR rules.")
		iter := ms.db.C(colLcr).Find(nil).Select(bson.M{"key": 1}).Iter()
		lcrKeys = make([]string, 0)
		for iter.Next(&result) {
			lcrKeys = append(lcrKeys, result.Key)
		}
		if err := iter.Close(); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
		cache2go.RemPrefixKey(utils.LCR_PREFIX)
	} else if len(lcrKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching LCR rules: %v", lcrKeys))
	}
	for _, key := range lcrKeys {
		cache2go.RemKey(key)
		if _, err = ms.GetLCR(key[len(utils.LCR_PREFIX):], true); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	}
	if len(lcrKeys) != 0 {
		utils.Logger.Info("Finished LCR rules caching.")
	}
	// DerivedChargers caching
	if dcsKeys == nil {
		utils.Logger.Info("Caching all derived chargers")
		iter := ms.db.C(colDcs).Find(nil).Select(bson.M{"key": 1}).Iter()
		dcsKeys = make([]string, 0)
		for iter.Next(&result) {
			dcsKeys = append(dcsKeys, result.Key)
		}
		if err := iter.Close(); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
		cache2go.RemPrefixKey(utils.DERIVEDCHARGERS_PREFIX)
	} else if len(dcsKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching derived chargers: %v", dcsKeys))
	}
	for _, key := range dcsKeys {
		cache2go.RemKey(key)
		if _, err = ms.GetDerivedChargers(key[len(utils.DERIVEDCHARGERS_PREFIX):], true); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	}
	if len(dcsKeys) != 0 {
		utils.Logger.Info("Finished derived chargers caching.")
	}
	if actKeys == nil {
		cache2go.RemPrefixKey(utils.ACTION_PREFIX)
	}
	if actKeys == nil {
		utils.Logger.Info("Caching all actions")
		iter := ms.db.C(colAct).Find(nil).Select(bson.M{"key": 1}).Iter()
		actKeys = make([]string, 0)
		for iter.Next(&result) {
			actKeys = append(actKeys, result.Key)
		}
		if err := iter.Close(); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
		cache2go.RemPrefixKey(utils.ACTION_PREFIX)
	} else if len(actKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching actions: %v", actKeys))
	}
	for _, key := range actKeys {
		cache2go.RemKey(key)
		if _, err = rs.GetActions(key[len(utils.ACTION_PREFIX):], true); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	}
	if len(actKeys) != 0 {
		utils.Logger.Info("Finished actions caching.")
	}

	if shgKeys == nil {
		cache2go.RemPrefixKey(utils.SHARED_GROUP_PREFIX)
	}
	if shgKeys == nil {
		utils.Logger.Info("Caching all shared groups")
		iter := ms.db.C(colShg).Find(nil).Select(bson.M{"key": 1}).Iter()
		shgKeys = make([]string, 0)
		for iter.Next(&result) {
			shgKeys = append(shgKeys, result.Key)
		}
		if err := iter.Close(); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	} else if len(shgKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching shared groups: %v", shgKeys))
	}
	for _, key := range shgKeys {
		cache2go.RemKey(key)
		if _, err = rs.GetSharedGroup(key[len(utils.SHARED_GROUP_PREFIX):], true); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	}
	if len(shgKeys) != 0 {
		utils.Logger.Info("Finished shared groups caching.")
	}

	cache2go.CommitTransaction()
	return nil
}

func (ms *MongoStorage) CacheAccountingAll() error {
	return ms.cacheAccounting(nil)
}

func (ms *MongoStorage) CacheAccountingPrefixes(prefixes ...string) error {
	pm := map[string][]string{
		utils.ALIASES_PREFIX: []string{},
	}
	for _, prefix := range prefixes {
		if _, found := pm[prefix]; !found {
			return utils.ErrNotFound
		}
		pm[prefix] = nil
	}
	return ms.cacheAccounting(pm[utils.ALIASES_PREFIX])
}

func (ms *MongoStorage) CacheAccountingPrefixValues(prefixes map[string][]string) error {
	pm := map[string][]string{
		utils.ALIASES_PREFIX: []string{},
	}
	for prefix, ids := range prefixes {
		if _, found := pm[prefix]; !found {
			return utils.ErrNotFound
		}
		pm[prefix] = ids
	}
	return ms.cacheAccounting(pm[utils.ALIASES_PREFIX])
}

func (ms *MongoStorage) cacheAccounting(alsKeys []string) (err error) {
	cache2go.BeginTransaction()
	var result struct{ Key string }
	if alsKeys == nil {
		cache2go.RemPrefixKey(utils.ALIASES_PREFIX)
	}
	if alsKeys == nil {
		utils.Logger.Info("Caching all aliases")
		iter := ms.db.C(colAls).Find(nil).Select(bson.M{"key": 1}).Iter()
		alsKeys = make([]string, 0)
		for iter.Next(&result) {
			alsKeys = append(alsKeys, result.Key)
		}
		if err := iter.Close(); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	} else if len(alsKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching aliases: %v", alsKeys))
	}
	for _, key := range alsKeys {
		cache2go.RemKey(key)
		if _, err = ms.GetAlias(key[len(utils.ALIASES_PREFIX):], true); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	}
	if len(alsKeys) != 0 {
		utils.Logger.Info("Finished aliases caching.")
	}
	utils.Logger.Info("Caching load history")
	if _, err = ms.GetLoadHistory(1, true); err != nil {
		cache2go.RollbackTransaction()
		return err
	}
	utils.Logger.Info("Finished load history caching.")
	cache2go.CommitTransaction()
	return nil
}

func (ms *MongoStorage) GetRatingPlan(key string, skipCache bool) (rp *RatingPlan, err error) {
	if !skipCache {
		if x, err := cache2go.Get(utils.RATING_PLAN_PREFIX + key); err == nil {
			return x.(*RatingPlan), nil
		} else {
			return nil, err
		}
	}
	rp = new(RatingPlan)
	err = ms.db.C(colRpl).Find(bson.M{"id": key}).One(&rp)
	if err == nil {
		cache2go.Cache(utils.RATING_PLAN_PREFIX+key, rp)
	}
	return
}

func (ms *MongoStorage) SetRatingPlan(rp *RatingPlan) error {
	err := ms.db.C(colRpl).Insert(rp)
	if err == nil && historyScribe != nil {
		response := 0
		historyScribe.Record(rp.GetHistoryRecord(), &response)
	}
	return err
}

func (ms *MongoStorage) GetRatingProfile(key string, skipCache bool) (rp *RatingProfile, err error) {
	if !skipCache {
		if x, err := cache2go.Get(utils.RATING_PROFILE_PREFIX + key); err == nil {
			return x.(*RatingProfile), nil
		} else {
			return nil, err
		}
	}
	rp = new(RatingProfile)
	err = ms.db.C(colRpf).Find(bson.M{"id": key}).One(&rp)
	if err == nil {
		cache2go.Cache(utils.RATING_PROFILE_PREFIX+key, rp)
	}
	return
}

func (ms *MongoStorage) SetRatingProfile(rp *RatingProfile) error {
	err := ms.db.C("ratingprofiles").Insert(rp)
	if err == nil && historyScribe != nil {
		response := 0
		historyScribe.Record(rp.GetHistoryRecord(false), &response)
	}
	return err
}

func (ms *MongoStorage) RemoveRatingProfile(key string) error {
	iter := ms.db.C(colRpf).Find(bson.M{"id": bson.RegEx{Pattern: key + ".*", Options: ""}}).Select(bson.M{"id": 1}).Iter()
	var result struct{ Id string }
	for iter.Next(&result) {
		if err := ms.db.C(colRpf).Remove(bson.M{"id": result.Id}); err != nil {
			return err
		}
		cache2go.RemKey(utils.RATING_PROFILE_PREFIX + key)
		rpf := &RatingProfile{Id: result.Id}
		if historyScribe != nil {
			response := 0
			go historyScribe.Record(rpf.GetHistoryRecord(true), &response)
		}
	}
	if err := iter.Close(); err != nil {
		return err
	}
	return nil
}

func (ms *MongoStorage) GetLCR(key string, skipCache bool) (lcr *LCR, err error) {
	if !skipCache {
		if x, err := cache2go.Get(utils.LCR_PREFIX + key); err == nil {
			return x.(*LCR), nil
		} else {
			return nil, err
		}
	}
	result := KeyedContainer{}
	err = ms.db.C(colLcr).Find(bson.M{"key": key}).One(&result)
	if err == nil {
		lcr = result.Value.(*LCR)
		cache2go.Cache(key, lcr)
	}
	return
}

func (ms *MongoStorage) SetLCR(lcr *LCR) (err error) {
	return ms.db.C(colLcr).Insert(&KeyedValue{Key: key, Value: lcr})
}

func (ms *MongoStorage) GetDestination(key string) (result *Destination, err error) {
	result = new(Destination)
	err = ms.db.C(colDst).Find(bson.M{"id": key}).One(result)
	if err != nil {
		result = nil
		return
	}
	// create optimized structure
	for _, p := range result.Prefixes {
		cache2go.CachePush(utils.DESTINATION_PREFIX+p, result.Id)
	}
	return
}

func (ms *MongoStorage) SetDestination(dest *Destination) error {
	err := ms.db.C(colDst).Insert(dest)
	if err == nil && historyScribe != nil {
		response := 0
		historyScribe.Record(dest.GetHistoryRecord(), &response)
	}
	return
}

func (ms *MongoStorage) GetActions(key string, skipCache bool) (as Actions, err error) {
	if !skipCache {
		if x, err := cache2go.Get(utils.ACTION_PREFIX + key); err == nil {
			return x.(Actions), nil
		} else {
			return nil, err
		}
	}
	result := KeyedContainer{}
	err = ms.db.C(colAct).Find(bson.M{"key": key}).One(&result)
	if err == nil {
		as = result.Value.(Actions)
		cache2go.Cache(utils.ACTION_PREFIX+key, as)
	}
	return
}

func (ms *MongoStorage) SetActions(key string, as Actions) error {
	return ms.db.C(colAct).Insert(&KeyedContainer{Key: key, Value: as})
}

func (ms *MongoStorage) GetSharedGroup(key string, skipCache bool) (sg *SharedGroup, err error) {
	if !skipCache {
		if x, err := cache2go.Get(utils.SHARED_GROUP_PREFIX + key); err == nil {
			return x.(*SharedGroup), nil
		} else {
			return nil, err
		}
	}
	sg = &SharedGroup{}
	err = ms.db.C(colShg).Find(bson.M{"id": key}).One(sg)
	if err == nil {
		cache2go.Cache(key, sg)
	}
	return
}

func (ms *MongoStorage) SetSharedGroup(sg *SharedGroup) (err error) {
	return ms.db.C(colShg).Insert(sg)
}

func (ms *MongoStorage) GetAccount(key string) (result *Account, err error) {
	result = new(Account)
	err = ms.db.C(colAcc).Find(bson.M{"id": key}).One(result)
	return
}

func (ms *MongoStorage) SetAccount(acc *Account) error {
	// never override existing account with an empty one
	// UPDATE: if all balances expired and were cleaned it makes
	// sense to write empty balance map
	if len(acc.BalanceMap) == 0 {
		if ac, err := ms.GetAccount(acc.Id); err == nil && !ac.allBalancesExpired() {
			ac.ActionTriggers = acc.ActionTriggers
			ac.UnitCounters = acc.UnitCounters
			ac.AllowNegative = acc.AllowNegative
			ac.Disabled = acc.Disabled
			acc = ac
		}
	}
	return ms.db.C(colAcc).Insert(acc)
}

func (ms *MongoStorage) RemoveAccount(key string) error {
	return ms.db.C(colAcc).Remove(bson.M{"id": key})

}

func (ms *MongoStorage) GetCdrStatsQueue(key string) (sq *StatsQueue, err error) {
	result := KeyedContainer{}
	err = ms.db.C(colStq).Find(bson.M{"key": key}).One(result)
	if err == nil {
		sq = result.Value.(*StatsQueue)
	}
	return
}

func (ms *MongoStorage) SetCdrStatsQueue(sq *StatsQueue) (err error) {
	return ms.db.C(colShg).Insert(&KeyedContainer{Key: sq.GetId(), Value: sq})
}

func (ms *MongoStorage) GetSubscribers() (result map[string]*SubscriberData, err error) {
	iter := ms.db.C(colPbs).Find(nil).Iter()
	result = make(map[string]*SubscriberData)
	kv := KeyedContainer{}
	for iter.Next(&kv) {
		result[kv.Key] = kv.Value.(*SubscriberData)
	}
	err = iter.Close()
	return
}

func (ms *MongoStorage) SetSubscriber(key string, sub *SubscriberData) (err error) {
	return ms.db.C(colPbs).Insert(&KeyedContainer{Key: key, Value: sub})
}

func (ms *MongoStorage) RemoveSubscriber(key string) (err error) {
	return ms.db.C(colPbs).Remove(bson.M{"key": key})
}

func (ms *MongoStorage) SetUser(up *UserProfile) (err error) {
	return ms.db.C(colUsr).Insert(&KeyedContainer{Key: up.GetId(), Value: up})
}

func (ms *MongoStorage) GetUser(key string) (up *UserProfile, err error) {
	kv := KeyedContainer{}
	err = ms.db.C(colUsr).Find(bson.M{"key": key}).One(&kv)
	if err == nil {
		up = kv.Value.(*UserProfile)
	}
	return
}

func (ms *MongoStorage) GetUsers() (result []*UserProfile, err error) {
	iter := ms.db.C(colUsr).Find(nil).Iter()
	kv := KeyedContainer{}
	for iter.Next(&kv) {
		result = append(result, kv.Value.(*UserProfile))
	}
	err = iter.Close()
	return
}

func (ms *MongoStorage) RemoveUser(key string) (err error) {
	return ms.db.C(colUsr).Remove(bson.M{"key": key})
}

func (ms *MongoStorage) SetAlias(al *Alias) (err error) {
	return ms.db.C(colAls).Insert(&KeyedContainer{Key: al.GetId(), Value: al.Values})
}

func (ms *MongoStorage) GetAlias(key string, skipCache bool) (al *Alias, err error) {
	origKey := key
	key = utils.ALIASES_PREFIX + key
	if !skipCache {
		if x, err := cache2go.Get(key); err == nil {
			al = &Alias{Values: x.(AliasValues)}
			al.SetId(key[len(utils.ALIASES_PREFIX):])
			return al, nil
		} else {
			return nil, err
		}
	}
	kv := KeyedContainer{}

	if err = ms.db.C(colAls).Find(bson.M{"key": origKey}).One(&kv); err == nil {
		al = &Alias{Value: kv.Value.(AliasValues)}
		al.SetId(origKey)
		if err == nil {
			cache2go.Cache(key, al.Values)
			// cache reverse alias
			for _, value := range al.Values {
				for target, pairs := range value.Pairs {
					for _, alias := range pairs {
						var existingKeys map[string]bool
						rKey := utils.REVERSE_ALIASES_PREFIX + alias + target + al.Context
						if x, err := cache2go.Get(rKey); err == nil {
							existingKeys = x.(map[string]bool)
						} else {
							existingKeys = make(map[string]bool)
						}
						existingKeys[utils.ConcatenatedKey(origKey, value.DestinationId)] = true
						cache2go.Cache(rKey, existingKeys)
					}
				}
			}
		}
	}
	return
}

func (ms *MongoStorage) RemoveAlias(key string) (err error) {
	al := &Alias{}
	al.SetId(key)
	origKey := key
	key = utils.ALIASES_PREFIX + key
	var aliasValues AliasValues

	kv := KeyedContainer{}
	if err := ms.db.C(colAls).Find(bson.M{"key": origKey}).One(&kv); err == nil {
		aliasValues = kv.Value.(AliasValues)
	}
	err = ms.db.C(colAls).Remove(bson.M{"key": origKey})
	if err == nil {
		for _, value := range aliasValues {
			for target, pairs := range value.Pairs {
				for _, alias := range pairs {
					var existingKeys map[string]bool
					rKey := utils.REVERSE_ALIASES_PREFIX + alias + target + al.Context
					if x, err := cache2go.Get(rKey); err == nil {
						existingKeys = x.(map[string]bool)
					}
					for eKey := range existingKeys {
						if strings.HasPrefix(origKey, eKey) {
							delete(existingKeys, eKey)
						}
					}
					if len(existingKeys) == 0 {
						cache2go.RemKey(rKey)
					} else {
						cache2go.Cache(rKey, existingKeys)
					}
				}
				cache2go.RemKey(key)
			}
		}
	}
	return
}

// Limit will only retrieve the last n items out of history, newest first
func (ms *MongoStorage) GetLoadHistory(limit int, skipCache bool) ([]*LoadInstance, error) {
	if limit == 0 {
		return nil, nil
	}
	if !skipCache {
		if x, err := cache2go.Get(utils.LOADINST_KEY); err != nil {
			return nil, err
		} else {
			items := x.([]*LoadInstance)
			if len(items) < limit || limit == -1 {
				return items, nil
			}
			return items[:limit], nil
		}
	}
	if limit != -1 {
		limit -= -1 // Decrease limit to match redis approach on lrange
	}
	marshaleds, err := ms.db.Lrange(utils.LOADINST_KEY, 0, limit)
	if err != nil {
		return nil, err
	}
	loadInsts := make([]*LoadInstance, len(marshaleds))
	for idx, marshaled := range marshaleds {
		var lInst LoadInstance
		err = rs.ms.Unmarshal(marshaled, &lInst)
		if err != nil {
			return nil, err
		}
		loadInsts[idx] = &lInst
	}
	cache2go.RemKey(utils.LOADINST_KEY)
	cache2go.Cache(utils.LOADINST_KEY, loadInsts)
	return loadInsts, nil
}

// Adds a single load instance to load history
func (ms *MongoStorage) AddLoadHistory(ldInst *LoadInstance, loadHistSize int) error {
	if loadHistSize == 0 { // Load history disabled
		return nil
	}
	marshaled, err := rs.ms.Marshal(&ldInst)
	if err != nil {
		return err
	}
	_, err = Guardian.Guard(func() (interface{}, error) { // Make sure we do it locked since other instance can modify history while we read it
		histLen, err := ms.db.Llen(utils.LOADINST_KEY)
		if err != nil {
			return nil, err
		}
		if histLen >= loadHistSize { // Have hit maximum history allowed, remove oldest element in order to add new one
			if _, err := ms.db.Rpop(utils.LOADINST_KEY); err != nil {
				return nil, err
			}
		}
		err = ms.db.Lpush(utils.LOADINST_KEY, marshaled)
		return nil, err
	}, 0, utils.LOADINST_KEY)
	return err
}

func (ms *MongoStorage) GetActionTriggers(key string) (atrs ActionTriggers, err error) {
	var values []byte
	if values, err = ms.db.Get(utils.ACTION_TRIGGER_PREFIX + key); err == nil {
		err = rs.ms.Unmarshal(values, &atrs)
	}
	return
}

func (ms *MongoStorage) SetActionTriggers(key string, atrs ActionTriggers) (err error) {
	if len(atrs) == 0 {
		// delete the key
		_, err = ms.db.Del(utils.ACTION_TRIGGER_PREFIX + key)
		return err
	}
	result, err := rs.ms.Marshal(&atrs)
	if err != nil {
		return err
	}
	err = ms.db.Set(utils.ACTION_TRIGGER_PREFIX+key, result)
	return
}

func (ms *MongoStorage) GetActionPlans(key string) (ats ActionPlan, err error) {
	result := AtKeyValue{}
	err = ms.db.C("actiontimings").Find(bson.M{"key": key}).One(&result)
	return result.Value, err
}

func (ms *MongoStorage) SetActionPlans(key string, ats ActionPlan) error {
	return ms.db.C("actiontimings").Insert(&AtKeyValue{key, ats})
}

func (ms *MongoStorage) GetAllActionPlans() (ats map[string]ActionPlan, err error) {
	result := AtKeyValue{}
	iter := ms.db.C("actiontimings").Find(nil).Iter()
	ats = make(map[string]ActionPlan)
	for iter.Next(&result) {
		ats[result.Key] = result.Value
	}
	return
}

func (ms *MongoStorage) LogCallCost(cgrid, source string, cc *CallCost) error {
	return ms.db.C("cclog").Insert(&LogCostEntry{cgrid, cc, source})
}

func (ms *MongoStorage) GetCallCostLog(cgrid, source string) (cc *CallCost, err error) {
	result := new(LogCostEntry)
	err = ms.db.C("cclog").Find(bson.M{"_id": cgrid, "source": source}).One(result)
	cc = result.CallCost
	return
}

func (ms *MongoStorage) LogActionTrigger(ubId, source string, at *ActionTrigger, as Actions) (err error) {
	return ms.db.C("actlog").Insert(&LogTriggerEntry{ubId, at, as, time.Now(), source})
}

func (ms *MongoStorage) LogActionPlan(source string, at *ActionPlan, as Actions) (err error) {
	return ms.db.C("actlog").Insert(&LogTimingEntry{at, as, time.Now(), source})
}

func (ms *MongoStorage) LogError(uuid, source, errstr string) (err error) {
	return ms.db.C("errlog").Insert(&LogErrEntry{uuid, errstr, source})
}
