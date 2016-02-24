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
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	colDst    = "destinations"
	colAct    = "actions"
	colApl    = "action_plans"
	colTsk    = "tasks"
	colAtr    = "action_triggers"
	colRpl    = "rating_plans"
	colRpf    = "rating_profiles"
	colAcc    = "accounts"
	colShg    = "shared_groups"
	colLcr    = "lcr_rules"
	colDcs    = "derived_chargers"
	colAls    = "aliases"
	colStq    = "stat_qeues"
	colPbs    = "pubsub"
	colUsr    = "users"
	colCrs    = "cdr_stats"
	colLht    = "load_history"
	colLogAtr = "action_trigger_logs"
	colLogApl = "action_plan_logs"
	colLogErr = "error_logs"
)

var (
	CGRIDLow           = strings.ToLower(utils.CGRID)
	RunIDLow           = strings.ToLower(utils.MEDI_RUNID)
	OrderIDLow         = strings.ToLower(utils.ORDERID)
	ToRLow             = strings.ToLower(utils.TOR)
	CDRHostLow         = strings.ToLower(utils.CDRHOST)
	CDRSourceLow       = strings.ToLower(utils.CDRSOURCE)
	RequestTypeLow     = strings.ToLower(utils.REQTYPE)
	DirectionLow       = strings.ToLower(utils.DIRECTION)
	TenantLow          = strings.ToLower(utils.TENANT)
	CategoryLow        = strings.ToLower(utils.CATEGORY)
	AccountLow         = strings.ToLower(utils.ACCOUNT)
	SubjectLow         = strings.ToLower(utils.SUBJECT)
	SupplierLow        = strings.ToLower(utils.SUPPLIER)
	DisconnectCauseLow = strings.ToLower(utils.DISCONNECT_CAUSE)
	SetupTimeLow       = strings.ToLower(utils.SETUP_TIME)
	AnswerTimeLow      = strings.ToLower(utils.ANSWER_TIME)
	CreatedAtLow       = strings.ToLower(utils.CreatedAt)
	UpdatedAtLow       = strings.ToLower(utils.UpdatedAt)
	UsageLow           = strings.ToLower(utils.USAGE)
	PDDLow             = strings.ToLower(utils.PDD)
	CostDetailsLow     = strings.ToLower(utils.COST_DETAILS)
	DestinationLow     = strings.ToLower(utils.DESTINATION)
	CostLow            = strings.ToLower(utils.COST)
)

type MongoStorage struct {
	session *mgo.Session
	db      *mgo.Database
	ms      Marshaler
}

func NewMongoStorage(host, port, db, user, pass string, cdrsIndexes []string) (*MongoStorage, error) {
	address := fmt.Sprintf("%s:%s", host, port)
	if user != "" && pass != "" {
		address = fmt.Sprintf("%s:%s@%s", user, pass, address)
	}
	session, err := mgo.Dial(address)
	if err != nil {
		return nil, err
	}
	ndb := session.DB(db)
	//session.SetMode(mgo.Monotonic, true)
	index := mgo.Index{
		Key:        []string{"key"},
		Unique:     true,  // Prevent two documents from having the same index key
		DropDups:   false, // Drop documents with the same index key as a previously indexed one
		Background: false, // Build index in background and return immediately
		Sparse:     false, // Only index documents containing the Key fields
	}
	collections := []string{colAct, colApl, colAtr, colDcs, colAls, colUsr, colLcr, colLht, colRpl, colDst}
	for _, col := range collections {
		if err = ndb.C(col).EnsureIndex(index); err != nil {
			return nil, err
		}
	}
	index = mgo.Index{
		Key:        []string{"id"},
		Unique:     true,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	}
	collections = []string{colRpf, colShg, colAcc, colCrs}
	for _, col := range collections {
		if err = ndb.C(col).EnsureIndex(index); err != nil {
			return nil, err
		}
	}
	index = mgo.Index{
		Key:        []string{"tpid", "tag"},
		Unique:     true,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	}
	collections = []string{utils.TBL_TP_TIMINGS, utils.TBL_TP_DESTINATIONS, utils.TBL_TP_DESTINATION_RATES, utils.TBL_TP_RATING_PLANS, utils.TBL_TP_SHARED_GROUPS, utils.TBL_TP_CDR_STATS, utils.TBL_TP_ACTIONS, utils.TBL_TP_ACTION_PLANS, utils.TBL_TP_ACTION_TRIGGERS}
	for _, col := range collections {
		if err = ndb.C(col).EnsureIndex(index); err != nil {
			return nil, err
		}
	}
	index = mgo.Index{
		Key:        []string{"tpid", "direction", "tenant", "category", "subject", "loadid"},
		Unique:     true,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	}
	collections = []string{utils.TBL_TP_RATE_PROFILES}
	for _, col := range collections {
		if err = ndb.C(col).EnsureIndex(index); err != nil {
			return nil, err
		}
	}
	index = mgo.Index{
		Key:        []string{"tpid", "direction", "tenant", "category", "account", "subject"},
		Unique:     true,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	}
	collections = []string{utils.TBL_TP_LCRS}
	for _, col := range collections {
		if err = ndb.C(col).EnsureIndex(index); err != nil {
			return nil, err
		}
	}
	index = mgo.Index{
		Key:        []string{"tpid", "tenant", "username"},
		Unique:     true,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	}
	collections = []string{utils.TBL_TP_USERS}
	for _, col := range collections {
		if err = ndb.C(col).EnsureIndex(index); err != nil {
			return nil, err
		}
	}
	index = mgo.Index{
		Key:        []string{"tpid", "direction", "tenant", "category", "account", "subject", "context"},
		Unique:     true,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	}
	collections = []string{utils.TBL_TP_LCRS}
	for _, col := range collections {
		if err = ndb.C(col).EnsureIndex(index); err != nil {
			return nil, err
		}
	}
	index = mgo.Index{
		Key:        []string{"tpid", "direction", "tenant", "category", "subject", "account", "loadid"},
		Unique:     true,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	}
	collections = []string{utils.TBL_TP_DERIVED_CHARGERS}
	for _, col := range collections {
		if err = ndb.C(col).EnsureIndex(index); err != nil {
			return nil, err
		}
	}
	index = mgo.Index{
		Key:        []string{"tpid", "direction", "tenant", "account", "loadid"},
		Unique:     true,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	}
	collections = []string{utils.TBL_TP_DERIVED_CHARGERS}
	for _, col := range collections {
		if err = ndb.C(col).EnsureIndex(index); err != nil {
			return nil, err
		}
	}
	index = mgo.Index{
		Key:        []string{CGRIDLow, RunIDLow},
		Unique:     true,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	}
	if err = ndb.C(utils.TBL_CDRS).EnsureIndex(index); err != nil {
		return nil, err
	}
	for _, idxKey := range cdrsIndexes {
		index = mgo.Index{
			Key:        []string{idxKey},
			Unique:     false,
			DropDups:   false,
			Background: false,
			Sparse:     false,
		}
		if err = ndb.C(utils.TBL_CDRS).EnsureIndex(index); err != nil {
			return nil, err
		}
	}
	return &MongoStorage{db: ndb, session: session, ms: NewCodecMsgpackMarshaler()}, err
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

func (ms *MongoStorage) CacheRatingAll() error {
	return ms.cacheRating(nil, nil, nil, nil, nil, nil, nil, nil)
}

func (ms *MongoStorage) CacheRatingPrefixes(prefixes ...string) error {
	pm := map[string][]string{
		utils.DESTINATION_PREFIX:     []string{},
		utils.RATING_PLAN_PREFIX:     []string{},
		utils.RATING_PROFILE_PREFIX:  []string{},
		utils.LCR_PREFIX:             []string{},
		utils.DERIVEDCHARGERS_PREFIX: []string{},
		utils.ACTION_PREFIX:          []string{},
		utils.ACTION_PLAN_PREFIX:     []string{},
		utils.SHARED_GROUP_PREFIX:    []string{},
	}
	for _, prefix := range prefixes {
		if _, found := pm[prefix]; !found {
			return utils.ErrNotFound
		}
		pm[prefix] = nil
	}
	return ms.cacheRating(pm[utils.DESTINATION_PREFIX], pm[utils.RATING_PLAN_PREFIX], pm[utils.RATING_PROFILE_PREFIX], pm[utils.LCR_PREFIX], pm[utils.DERIVEDCHARGERS_PREFIX], pm[utils.ACTION_PREFIX], pm[utils.ACTION_PLAN_PREFIX], pm[utils.SHARED_GROUP_PREFIX])
}

func (ms *MongoStorage) CacheRatingPrefixValues(prefixes map[string][]string) error {
	pm := map[string][]string{
		utils.DESTINATION_PREFIX:     []string{},
		utils.RATING_PLAN_PREFIX:     []string{},
		utils.RATING_PROFILE_PREFIX:  []string{},
		utils.LCR_PREFIX:             []string{},
		utils.DERIVEDCHARGERS_PREFIX: []string{},
		utils.ACTION_PREFIX:          []string{},
		utils.ACTION_PLAN_PREFIX:     []string{},
		utils.SHARED_GROUP_PREFIX:    []string{},
	}
	for prefix, ids := range prefixes {
		if _, found := pm[prefix]; !found {
			return utils.ErrNotFound
		}
		pm[prefix] = ids
	}
	return ms.cacheRating(pm[utils.DESTINATION_PREFIX], pm[utils.RATING_PLAN_PREFIX], pm[utils.RATING_PROFILE_PREFIX], pm[utils.LCR_PREFIX], pm[utils.DERIVEDCHARGERS_PREFIX], pm[utils.ACTION_PREFIX], pm[utils.ACTION_PLAN_PREFIX], pm[utils.SHARED_GROUP_PREFIX])
}

func (ms *MongoStorage) cacheRating(dKeys, rpKeys, rpfKeys, lcrKeys, dcsKeys, actKeys, aplKeys, shgKeys []string) (err error) {
	cache2go.BeginTransaction()
	keyResult := struct{ Key string }{}
	idResult := struct{ Id string }{}
	if dKeys == nil || (float64(cache2go.CountEntries(utils.DESTINATION_PREFIX))*utils.DESTINATIONS_LOAD_THRESHOLD < float64(len(dKeys))) {
		// if need to load more than a half of exiting keys load them all
		utils.Logger.Info("Caching all destinations")
		iter := ms.db.C(colDst).Find(nil).Select(bson.M{"key": 1}).Iter()
		dKeys = make([]string, 0)
		for iter.Next(&keyResult) {
			dKeys = append(dKeys, utils.DESTINATION_PREFIX+keyResult.Key)
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
		for iter.Next(&keyResult) {
			rpKeys = append(rpKeys, utils.RATING_PLAN_PREFIX+keyResult.Key)
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
		iter := ms.db.C(colRpf).Find(nil).Select(bson.M{"id": 1}).Iter()
		rpfKeys = make([]string, 0)
		for iter.Next(&idResult) {
			rpfKeys = append(rpfKeys, utils.RATING_PROFILE_PREFIX+idResult.Id)
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
		for iter.Next(&keyResult) {
			lcrKeys = append(lcrKeys, utils.LCR_PREFIX+keyResult.Key)
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
		for iter.Next(&keyResult) {
			dcsKeys = append(dcsKeys, utils.DERIVEDCHARGERS_PREFIX+keyResult.Key)
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
		for iter.Next(&keyResult) {
			actKeys = append(actKeys, utils.ACTION_PREFIX+keyResult.Key)
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
		if _, err = ms.GetActions(key[len(utils.ACTION_PREFIX):], true); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	}
	if len(actKeys) != 0 {
		utils.Logger.Info("Finished actions caching.")
	}

	if aplKeys == nil {
		cache2go.RemPrefixKey(utils.ACTION_PLAN_PREFIX)
	}
	if aplKeys == nil {
		utils.Logger.Info("Caching all action plans")
		iter := ms.db.C(colApl).Find(nil).Select(bson.M{"key": 1}).Iter()
		aplKeys = make([]string, 0)
		for iter.Next(&keyResult) {
			aplKeys = append(aplKeys, utils.ACTION_PLAN_PREFIX+keyResult.Key)
		}
		if err := iter.Close(); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
		cache2go.RemPrefixKey(utils.ACTION_PLAN_PREFIX)
	} else if len(aplKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching action plans: %v", aplKeys))
	}
	for _, key := range aplKeys {
		cache2go.RemKey(key)
		if _, err = ms.GetActionPlan(key[len(utils.ACTION_PLAN_PREFIX):], true); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	}
	if len(aplKeys) != 0 {
		utils.Logger.Info("Finished action plans caching.")
	}

	if shgKeys == nil {
		cache2go.RemPrefixKey(utils.SHARED_GROUP_PREFIX)
	}
	if shgKeys == nil {
		utils.Logger.Info("Caching all shared groups")
		iter := ms.db.C(colShg).Find(nil).Select(bson.M{"id": 1}).Iter()
		shgKeys = make([]string, 0)
		for iter.Next(&idResult) {
			shgKeys = append(shgKeys, utils.SHARED_GROUP_PREFIX+idResult.Id)
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
		if _, err = ms.GetSharedGroup(key[len(utils.SHARED_GROUP_PREFIX):], true); err != nil {
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
	var keyResult struct{ Key string }
	if alsKeys == nil {
		cache2go.RemPrefixKey(utils.ALIASES_PREFIX)
	}
	if alsKeys == nil {
		utils.Logger.Info("Caching all aliases")
		iter := ms.db.C(colAls).Find(nil).Select(bson.M{"key": 1}).Iter()
		alsKeys = make([]string, 0)
		for iter.Next(&keyResult) {
			alsKeys = append(alsKeys, utils.ALIASES_PREFIX+keyResult.Key)
		}
		if err := iter.Close(); err != nil {
			cache2go.RollbackTransaction()
			return err
		}
	} else if len(alsKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching aliases: %v", alsKeys))
	}
	for _, key := range alsKeys {
		// check if it already exists
		// to remove reverse cache keys
		if avs, err := cache2go.Get(key); err == nil && avs != nil {
			al := &Alias{Values: avs.(AliasValues)}
			al.SetId(key[len(utils.ALIASES_PREFIX):])
			al.RemoveReverseCache()
		}
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

func (ms *MongoStorage) HasData(category, subject string) (bool, error) {
	switch category {
	case utils.DESTINATION_PREFIX:
		count, err := ms.db.C(colDst).Find(bson.M{"key": subject}).Count()
		return count > 0, err
	case utils.RATING_PLAN_PREFIX:
		count, err := ms.db.C(colRpl).Find(bson.M{"key": subject}).Count()
		return count > 0, err
	case utils.RATING_PROFILE_PREFIX:
		count, err := ms.db.C(colRpf).Find(bson.M{"id": subject}).Count()
		return count > 0, err
	case utils.ACTION_PREFIX:
		count, err := ms.db.C(colAct).Find(bson.M{"key": subject}).Count()
		return count > 0, err
	case utils.ACTION_PLAN_PREFIX:
		count, err := ms.db.C(colApl).Find(bson.M{"key": subject}).Count()
		return count > 0, err
	case utils.ACCOUNT_PREFIX:
		count, err := ms.db.C(colAcc).Find(bson.M{"id": subject}).Count()
		return count > 0, err
	}
	return false, errors.New("Unsupported category in HasData")
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
	var kv struct {
		Key   string
		Value []byte
	}
	err = ms.db.C(colRpl).Find(bson.M{"key": key}).One(&kv)
	if err == nil {
		b := bytes.NewBuffer(kv.Value)
		r, err := zlib.NewReader(b)
		if err != nil {
			return nil, err
		}
		out, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		r.Close()
		err = ms.ms.Unmarshal(out, &rp)
		if err != nil {
			return nil, err
		}
		cache2go.Cache(utils.RATING_PLAN_PREFIX+key, rp)
	}
	return
}

func (ms *MongoStorage) SetRatingPlan(rp *RatingPlan) error {
	result, err := ms.ms.Marshal(rp)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	_, err = ms.db.C(colRpl).Upsert(bson.M{"key": rp.Id}, &struct {
		Key   string
		Value []byte
	}{Key: rp.Id, Value: b.Bytes()})
	if err == nil && historyScribe != nil {
		var response int
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
	err = ms.db.C(colRpf).Find(bson.M{"id": key}).One(rp)
	if err == nil {
		cache2go.Cache(utils.RATING_PROFILE_PREFIX+key, rp)
	}
	return
}

func (ms *MongoStorage) SetRatingProfile(rp *RatingProfile) error {
	_, err := ms.db.C(colRpf).Upsert(bson.M{"id": rp.Id}, rp)
	if err == nil && historyScribe != nil {
		var response int
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
			var response int
			go historyScribe.Record(rpf.GetHistoryRecord(true), &response)
		}
	}
	return iter.Close()
}

func (ms *MongoStorage) GetLCR(key string, skipCache bool) (lcr *LCR, err error) {
	if !skipCache {
		if x, err := cache2go.Get(utils.LCR_PREFIX + key); err == nil {
			return x.(*LCR), nil
		} else {
			return nil, err
		}
	}
	var result struct {
		Key   string
		Value *LCR
	}
	err = ms.db.C(colLcr).Find(bson.M{"key": key}).One(&result)
	if err == nil {
		lcr = result.Value
		cache2go.Cache(utils.LCR_PREFIX+key, lcr)
	}
	return
}

func (ms *MongoStorage) SetLCR(lcr *LCR) error {
	_, err := ms.db.C(colLcr).Upsert(bson.M{"key": lcr.GetId()}, &struct {
		Key   string
		Value *LCR
	}{lcr.GetId(), lcr})
	return err
}

func (ms *MongoStorage) GetDestination(key string) (result *Destination, err error) {
	result = new(Destination)
	var kv struct {
		Key   string
		Value []byte
	}
	err = ms.db.C(colDst).Find(bson.M{"key": key}).One(&kv)
	if err == nil {
		b := bytes.NewBuffer(kv.Value)
		r, err := zlib.NewReader(b)
		if err != nil {
			return nil, err
		}
		out, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		r.Close()
		err = ms.ms.Unmarshal(out, &result)
		if err != nil {
			return nil, err
		}
		// create optimized structure
		for _, p := range result.Prefixes {
			cache2go.Push(utils.DESTINATION_PREFIX+p, result.Id)
		}
	}
	if err != nil {
		result = nil
	}
	return
}

func (ms *MongoStorage) SetDestination(dest *Destination) (err error) {
	result, err := ms.ms.Marshal(dest)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	_, err = ms.db.C(colDst).Upsert(bson.M{"key": dest.Id}, &struct {
		Key   string
		Value []byte
	}{Key: dest.Id, Value: b.Bytes()})
	if err == nil && historyScribe != nil {
		var response int
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
	var result struct {
		Key   string
		Value Actions
	}
	err = ms.db.C(colAct).Find(bson.M{"key": key}).One(&result)
	if err == nil {
		as = result.Value
		cache2go.Cache(utils.ACTION_PREFIX+key, as)
	}
	return
}

func (ms *MongoStorage) SetActions(key string, as Actions) error {
	_, err := ms.db.C(colAct).Upsert(bson.M{"key": key}, &struct {
		Key   string
		Value Actions
	}{Key: key, Value: as})
	return err
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
		cache2go.Cache(utils.SHARED_GROUP_PREFIX+key, sg)
	}
	return
}

func (ms *MongoStorage) SetSharedGroup(sg *SharedGroup) (err error) {
	_, err = ms.db.C(colShg).Upsert(bson.M{"id": sg.Id}, sg)
	return err
}

func (ms *MongoStorage) GetAccount(key string) (result *Account, err error) {
	result = new(Account)
	err = ms.db.C(colAcc).Find(bson.M{"id": key}).One(result)
	if err == mgo.ErrNotFound {
		result = nil
	}
	return
}

func (ms *MongoStorage) SetAccount(acc *Account) error {
	// never override existing account with an empty one
	// UPDATE: if all balances expired and were cleaned it makes
	// sense to write empty balance map
	if len(acc.BalanceMap) == 0 {
		if ac, err := ms.GetAccount(acc.ID); err == nil && !ac.allBalancesExpired() {
			ac.ActionTriggers = acc.ActionTriggers
			ac.UnitCounters = acc.UnitCounters
			ac.AllowNegative = acc.AllowNegative
			ac.Disabled = acc.Disabled
			acc = ac
		}
	}
	_, err := ms.db.C(colAcc).Upsert(bson.M{"id": acc.ID}, acc)
	return err
}

func (ms *MongoStorage) RemoveAccount(key string) error {
	return ms.db.C(colAcc).Remove(bson.M{"id": key})

}

func (ms *MongoStorage) GetCdrStatsQueue(key string) (sq *StatsQueue, err error) {
	var result struct {
		Key   string
		Value *StatsQueue
	}
	err = ms.db.C(colStq).Find(bson.M{"key": key}).One(&result)
	if err == nil {
		sq = result.Value
	}
	return
}

func (ms *MongoStorage) SetCdrStatsQueue(sq *StatsQueue) (err error) {
	_, err = ms.db.C(colStq).Upsert(bson.M{"key": sq.GetId()}, &struct {
		Key   string
		Value *StatsQueue
	}{Key: sq.GetId(), Value: sq})
	return
}

func (ms *MongoStorage) GetSubscribers() (result map[string]*SubscriberData, err error) {
	iter := ms.db.C(colPbs).Find(nil).Iter()
	result = make(map[string]*SubscriberData)
	var kv struct {
		Key   string
		Value *SubscriberData
	}
	for iter.Next(&kv) {
		result[kv.Key] = kv.Value
	}
	err = iter.Close()
	return
}

func (ms *MongoStorage) SetSubscriber(key string, sub *SubscriberData) (err error) {
	_, err = ms.db.C(colPbs).Upsert(bson.M{"key": key}, &struct {
		Key   string
		Value *SubscriberData
	}{Key: key, Value: sub})
	return err
}

func (ms *MongoStorage) RemoveSubscriber(key string) (err error) {
	return ms.db.C(colPbs).Remove(bson.M{"key": key})
}

func (ms *MongoStorage) SetUser(up *UserProfile) (err error) {
	_, err = ms.db.C(colUsr).Upsert(bson.M{"key": up.GetId()}, &struct {
		Key   string
		Value *UserProfile
	}{Key: up.GetId(), Value: up})
	return err
}

func (ms *MongoStorage) GetUser(key string) (up *UserProfile, err error) {
	var kv struct {
		Key   string
		Value *UserProfile
	}
	err = ms.db.C(colUsr).Find(bson.M{"key": key}).One(&kv)
	if err == nil {
		up = kv.Value
	}
	return
}

func (ms *MongoStorage) GetUsers() (result []*UserProfile, err error) {
	iter := ms.db.C(colUsr).Find(nil).Iter()
	var kv struct {
		Key   string
		Value *UserProfile
	}
	for iter.Next(&kv) {
		result = append(result, kv.Value)
	}
	err = iter.Close()
	return
}

func (ms *MongoStorage) RemoveUser(key string) (err error) {
	return ms.db.C(colUsr).Remove(bson.M{"key": key})
}

func (ms *MongoStorage) SetAlias(al *Alias) (err error) {
	_, err = ms.db.C(colAls).Upsert(bson.M{"key": al.GetId()}, &struct {
		Key   string
		Value AliasValues
	}{Key: al.GetId(), Value: al.Values})
	return err
}

func (ms *MongoStorage) GetAlias(key string, skipCache bool) (al *Alias, err error) {
	origKey := key
	key = utils.ALIASES_PREFIX + key
	if !skipCache {
		if x, err := cache2go.Get(key); err == nil {
			al = &Alias{Values: x.(AliasValues)}
			al.SetId(origKey)
			return al, nil
		} else {
			return nil, err
		}
	}
	var kv struct {
		Key   string
		Value AliasValues
	}
	if err = ms.db.C(colAls).Find(bson.M{"key": origKey}).One(&kv); err == nil {
		al = &Alias{Values: kv.Value}
		al.SetId(origKey)
		if err == nil {
			cache2go.Cache(key, al.Values)
			// cache reverse alias
			al.SetReverseCache()
		}
	}
	return
}

func (ms *MongoStorage) RemoveAlias(key string) (err error) {
	al := &Alias{}
	al.SetId(key)
	origKey := key
	key = utils.ALIASES_PREFIX + key
	var kv struct {
		Key   string
		Value AliasValues
	}
	if err := ms.db.C(colAls).Find(bson.M{"key": origKey}).One(&kv); err == nil {
		al.Values = kv.Value
	}
	err = ms.db.C(colAls).Remove(bson.M{"key": origKey})
	if err == nil {
		al.RemoveReverseCache()
		cache2go.RemKey(key)
	}
	return
}

// Limit will only retrieve the last n items out of history, newest first
func (ms *MongoStorage) GetLoadHistory(limit int, skipCache bool) (loadInsts []*LoadInstance, err error) {
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
	var kv struct {
		Key   string
		Value []*LoadInstance
	}
	err = ms.db.C(colLht).Find(bson.M{"key": utils.LOADINST_KEY}).One(&kv)
	if err == nil {
		loadInsts = kv.Value
		cache2go.RemKey(utils.LOADINST_KEY)
		cache2go.Cache(utils.LOADINST_KEY, loadInsts)
	}
	return loadInsts, nil
}

// Adds a single load instance to load history
func (ms *MongoStorage) AddLoadHistory(ldInst *LoadInstance, loadHistSize int) error {
	if loadHistSize == 0 { // Load history disabled
		return nil
	}
	// get existing load history
	var existingLoadHistory []*LoadInstance
	var kv struct {
		Key   string
		Value []*LoadInstance
	}
	err := ms.db.C(colLht).Find(bson.M{"key": utils.LOADINST_KEY}).One(&kv)

	if err != nil && err != mgo.ErrNotFound {
		return err
	} else {
		if kv.Value != nil {
			existingLoadHistory = kv.Value
		}
	}

	_, err = Guardian.Guard(func() (interface{}, error) { // Make sure we do it locked since other instance can modify history while we read it
		// insert on first position
		existingLoadHistory = append(existingLoadHistory, nil)
		copy(existingLoadHistory[1:], existingLoadHistory[0:])
		existingLoadHistory[0] = ldInst

		//check length
		histLen := len(existingLoadHistory)
		if histLen >= loadHistSize { // Have hit maximum history allowed, remove oldest element in order to add new one
			existingLoadHistory = existingLoadHistory[:loadHistSize]
		}
		_, err = ms.db.C(colLht).Upsert(bson.M{"key": utils.LOADINST_KEY}, &struct {
			Key   string
			Value []*LoadInstance
		}{Key: utils.LOADINST_KEY, Value: existingLoadHistory})
		return nil, err
	}, 0, utils.LOADINST_KEY)
	return err
}

func (ms *MongoStorage) GetActionTriggers(key string) (atrs ActionTriggers, err error) {
	var kv struct {
		Key   string
		Value ActionTriggers
	}
	err = ms.db.C(colAtr).Find(bson.M{"key": key}).One(&kv)
	if err == nil {
		atrs = kv.Value
	}
	return
}

func (ms *MongoStorage) SetActionTriggers(key string, atrs ActionTriggers) (err error) {
	if len(atrs) == 0 {
		err = ms.db.C(colAtr).Remove(bson.M{"key": key}) // delete the key
		if err != mgo.ErrNotFound {
			return err
		}
		return nil
	}
	_, err = ms.db.C(colAtr).Upsert(bson.M{"key": key}, &struct {
		Key   string
		Value ActionTriggers
	}{Key: key, Value: atrs})
	return err
}

func (ms *MongoStorage) GetActionPlan(key string, skipCache bool) (ats *ActionPlan, err error) {
	if !skipCache {
		if x, err := cache2go.Get(utils.ACTION_PLAN_PREFIX + key); err == nil {
			return x.(*ActionPlan), nil
		} else {
			return nil, err
		}
	}
	var kv struct {
		Key   string
		Value []byte
	}
	err = ms.db.C(colApl).Find(bson.M{"key": key}).One(&kv)
	if err == nil {
		b := bytes.NewBuffer(kv.Value)
		r, err := zlib.NewReader(b)
		if err != nil {
			return nil, err
		}
		out, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		r.Close()
		err = ms.ms.Unmarshal(out, &ats)
		if err != nil {
			return nil, err
		}
		cache2go.Cache(utils.ACTION_PLAN_PREFIX+key, ats)
	}
	return
}

func (ms *MongoStorage) SetActionPlan(key string, ats *ActionPlan, overwrite bool) error {
	// clean dots from account ids map
	if len(ats.ActionTimings) == 0 {
		cache2go.RemKey(utils.ACTION_PLAN_PREFIX + key)
		err := ms.db.C(colApl).Remove(bson.M{"key": key})
		if err != mgo.ErrNotFound {
			return err
		}
		return nil
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
	result, err := ms.ms.Marshal(ats)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	_, err = ms.db.C(colApl).Upsert(bson.M{"key": key}, &struct {
		Key   string
		Value []byte
	}{Key: key, Value: b.Bytes()})
	return err
}

func (ms *MongoStorage) GetAllActionPlans() (ats map[string]*ActionPlan, err error) {
	apls, err := cache2go.GetAllEntries(utils.ACTION_PLAN_PREFIX)
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

func (ms *MongoStorage) PushTask(t *Task) error {
	return ms.db.C(colTsk).Insert(bson.M{"_id": bson.NewObjectId(), "task": t})
}

func (ms *MongoStorage) PopTask() (t *Task, err error) {
	v := struct {
		ID   bson.ObjectId `bson:"_id"`
		Task *Task
	}{}
	if err = ms.db.C(colTsk).Find(nil).One(&v); err == nil {
		if remErr := ms.db.C(colTsk).Remove(bson.M{"_id": v.ID}); remErr != nil {
			return nil, remErr
		}
		t = v.Task
	}

	return
}

func (ms *MongoStorage) GetDerivedChargers(key string, skipCache bool) (dcs *utils.DerivedChargers, err error) {
	if !skipCache {
		if x, err := cache2go.Get(utils.DERIVEDCHARGERS_PREFIX + key); err == nil {
			return x.(*utils.DerivedChargers), nil
		} else {
			return nil, err
		}
	}
	var kv struct {
		Key   string
		Value *utils.DerivedChargers
	}
	err = ms.db.C(colDcs).Find(bson.M{"key": key}).One(&kv)
	if err == nil {
		dcs = kv.Value
		cache2go.Cache(utils.DERIVEDCHARGERS_PREFIX+key, dcs)
	}
	return
}

func (ms *MongoStorage) SetDerivedChargers(key string, dcs *utils.DerivedChargers) (err error) {
	if dcs == nil || len(dcs.Chargers) == 0 {
		cache2go.RemKey(utils.DERIVEDCHARGERS_PREFIX + key)
		err = ms.db.C(colDcs).Remove(bson.M{"key": key})
		if err != mgo.ErrNotFound {
			return err
		}
		return nil
	}
	_, err = ms.db.C(colDcs).Upsert(bson.M{"key": key}, &struct {
		Key   string
		Value *utils.DerivedChargers
	}{Key: key, Value: dcs})
	return err
}

func (ms *MongoStorage) SetCdrStats(cs *CdrStats) error {
	_, err := ms.db.C(colCrs).Upsert(bson.M{"id": cs.Id}, cs)
	return err
}

func (ms *MongoStorage) GetCdrStats(key string) (cs *CdrStats, err error) {
	cs = &CdrStats{}
	err = ms.db.C(colCrs).Find(bson.M{"id": key}).One(cs)
	return
}

func (ms *MongoStorage) GetAllCdrStats() (css []*CdrStats, err error) {
	iter := ms.db.C(colCrs).Find(nil).Iter()
	var cs CdrStats
	for iter.Next(&cs) {
		clone := cs // avoid using the same pointer in append
		css = append(css, &clone)
	}
	err = iter.Close()
	return
}
