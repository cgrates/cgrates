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
	"time"

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
	colLogErr = "error_logs"
	colVer    = "versions"
)

var (
	CGRIDLow           = strings.ToLower(utils.CGRID)
	RunIDLow           = strings.ToLower(utils.MEDI_RUNID)
	OrderIDLow         = strings.ToLower(utils.ORDERID)
	OriginHostLow      = strings.ToLower(utils.CDRHOST)
	OriginIDLow        = strings.ToLower(utils.ACCID)
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
	session         *mgo.Session
	db              string
	ms              Marshaler
	cacheDumpDir    string
	loadHistorySize int
}

func (ms *MongoStorage) conn(col string) (*mgo.Session, *mgo.Collection) {
	sessionCopy := ms.session.Copy()
	return sessionCopy, sessionCopy.DB(ms.db).C(col)
}

func NewMongoStorage(host, port, db, user, pass string, cdrsIndexes []string, cacheDumpDir string, loadHistorySize int) (*MongoStorage, error) {

	// We need this object to establish a session to our MongoDB.
	/*address := fmt.Sprintf("%s:%s", host, port)
			mongoDBDialInfo := &mgo.DialInfo{
				Addrs:    []string{address},
				Timeout:  60 * time.Second,
				Database: db,
	Username: user,
	Password: pass,
			}

			// Create a session which maintains a pool of socket connections
			// to our MongoDB.
			session, err := mgo.DialWithInfo(mongoDBDialInfo)
			if err != nil {
				log.Printf("ERR: %v", err)
				return nil, err
			}*/

	address := fmt.Sprintf("%s:%s", host, port)
	if user != "" && pass != "" {
		address = fmt.Sprintf("%s:%s@%s", user, pass, address)
	}
	session, err := mgo.Dial(address)
	if err != nil {
		return nil, err
	}

	ndb := session.DB(db)
	session.SetMode(mgo.Strong, true)
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
		Key:        []string{CGRIDLow, RunIDLow, OriginIDLow},
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
	index = mgo.Index{
		Key:        []string{CGRIDLow, RunIDLow},
		Unique:     true,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	}
	if err = ndb.C(utils.TBLSMCosts).EnsureIndex(index); err != nil {
		return nil, err
	}
	index = mgo.Index{
		Key:        []string{OriginHostLow, OriginIDLow},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	}
	if err = ndb.C(utils.TBLSMCosts).EnsureIndex(index); err != nil {
		return nil, err
	}
	if cacheDumpDir != "" {
		if err := CacheSetDumperPath(cacheDumpDir); err != nil {
			utils.Logger.Info("<cache dumper> init error: " + err.Error())
		}
	}
	return &MongoStorage{db: db, session: session, ms: NewCodecMsgpackMarshaler(), cacheDumpDir: cacheDumpDir, loadHistorySize: loadHistorySize}, err
}

func (ms *MongoStorage) Close() {
	ms.session.Close()
}

func (ms *MongoStorage) GetKeysForPrefix(prefix string, skipCache bool) ([]string, error) {
	var category, subject string
	length := len(utils.DESTINATION_PREFIX)
	if len(prefix) >= length {
		category = prefix[:length] // prefix lenght
		subject = fmt.Sprintf("^%s", prefix[length:])
	} else {
		return nil, fmt.Errorf("unsupported prefix in GetKeysForPrefix: %s", prefix)
	}
	var result []string
	if skipCache {
		session := ms.session.Copy()
		defer session.Close()
		db := session.DB(ms.db)
		keyResult := struct{ Key string }{}
		idResult := struct{ Id string }{}
		switch category {
		case utils.DESTINATION_PREFIX:
			iter := db.C(colDst).Find(bson.M{"key": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"key": 1}).Iter()
			for iter.Next(&keyResult) {
				result = append(result, utils.DESTINATION_PREFIX+keyResult.Key)
			}
			return result, nil
		case utils.RATING_PLAN_PREFIX:
			iter := db.C(colRpl).Find(bson.M{"key": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"key": 1}).Iter()
			for iter.Next(&keyResult) {
				result = append(result, utils.RATING_PLAN_PREFIX+keyResult.Key)
			}
			return result, nil
		case utils.RATING_PROFILE_PREFIX:
			iter := db.C(colRpf).Find(bson.M{"id": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"id": 1}).Iter()
			for iter.Next(&idResult) {
				result = append(result, utils.RATING_PROFILE_PREFIX+idResult.Id)
			}
			return result, nil
		case utils.ACTION_PREFIX:
			iter := db.C(colAct).Find(bson.M{"key": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"key": 1}).Iter()
			for iter.Next(&keyResult) {
				result = append(result, utils.ACTION_PREFIX+keyResult.Key)
			}
			return result, nil
		case utils.ACTION_PLAN_PREFIX:
			iter := db.C(colApl).Find(bson.M{"key": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"key": 1}).Iter()
			for iter.Next(&keyResult) {
				result = append(result, utils.ACTION_PLAN_PREFIX+keyResult.Key)
			}
			return result, nil
		case utils.ACTION_TRIGGER_PREFIX:
			iter := db.C(colAtr).Find(bson.M{"key": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"key": 1}).Iter()
			for iter.Next(&keyResult) {
				result = append(result, utils.ACTION_TRIGGER_PREFIX+keyResult.Key)
			}
			return result, nil
		case utils.ACCOUNT_PREFIX:
			iter := db.C(colAcc).Find(bson.M{"id": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"id": 1}).Iter()
			for iter.Next(&idResult) {
				result = append(result, utils.ACCOUNT_PREFIX+idResult.Id)
			}
			return result, nil
		}
		return result, fmt.Errorf("unsupported prefix in GetKeysForPrefix: %s", prefix)
	}
	return CacheGetEntriesKeys(prefix), nil
}

func (ms *MongoStorage) Flush(ignore string) (err error) {
	session := ms.session.Copy()
	defer session.Close()
	db := session.DB(ms.db)
	collections, err := db.CollectionNames()
	if err != nil {
		return err
	}
	for _, c := range collections {
		if err = db.C(c).DropCollection(); err != nil {
			return err
		}
	}
	return nil
}

func (ms *MongoStorage) CacheRatingAll(loadID string) error {
	return ms.cacheRating(loadID, nil, nil, nil, nil, nil, nil, nil, nil)
}

func (ms *MongoStorage) CacheRatingPrefixes(loadID string, prefixes ...string) error {
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
	return ms.cacheRating(loadID, pm[utils.DESTINATION_PREFIX], pm[utils.RATING_PLAN_PREFIX], pm[utils.RATING_PROFILE_PREFIX], pm[utils.LCR_PREFIX], pm[utils.DERIVEDCHARGERS_PREFIX], pm[utils.ACTION_PREFIX], pm[utils.ACTION_PLAN_PREFIX], pm[utils.SHARED_GROUP_PREFIX])
}

func (ms *MongoStorage) CacheRatingPrefixValues(loadID string, prefixes map[string][]string) error {
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
	return ms.cacheRating(loadID, pm[utils.DESTINATION_PREFIX], pm[utils.RATING_PLAN_PREFIX], pm[utils.RATING_PROFILE_PREFIX], pm[utils.LCR_PREFIX], pm[utils.DERIVEDCHARGERS_PREFIX], pm[utils.ACTION_PREFIX], pm[utils.ACTION_PLAN_PREFIX], pm[utils.SHARED_GROUP_PREFIX])
}

func (ms *MongoStorage) cacheRating(loadID string, dKeys, rpKeys, rpfKeys, lcrKeys, dcsKeys, actKeys, aplKeys, shgKeys []string) (err error) {
	start := time.Now()
	CacheBeginTransaction()
	keyResult := struct{ Key string }{}
	idResult := struct{ Id string }{}
	session := ms.session.Copy()
	defer session.Close()
	db := session.DB(ms.db)
	if dKeys == nil || (float64(CacheCountEntries(utils.DESTINATION_PREFIX))*utils.DESTINATIONS_LOAD_THRESHOLD < float64(len(dKeys))) {
		// if need to load more than a half of exiting keys load them all
		utils.Logger.Info("Caching all destinations")
		iter := db.C(colDst).Find(nil).Select(bson.M{"key": 1}).Iter()
		dKeys = make([]string, 0)
		for iter.Next(&keyResult) {
			dKeys = append(dKeys, utils.DESTINATION_PREFIX+keyResult.Key)
		}
		if err := iter.Close(); err != nil {
			CacheRollbackTransaction()
			return fmt.Errorf("destinations: %s", err.Error())
		}
		CacheRemPrefixKey(utils.DESTINATION_PREFIX)
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
			CacheRollbackTransaction()
			return fmt.Errorf("destinations: %s", err.Error())
		}
	}
	if len(dKeys) != 0 {
		utils.Logger.Info("Finished destinations caching.")
	}
	if rpKeys == nil {
		utils.Logger.Info("Caching all rating plans")
		iter := db.C(colRpl).Find(nil).Select(bson.M{"key": 1}).Iter()
		rpKeys = make([]string, 0)
		for iter.Next(&keyResult) {
			rpKeys = append(rpKeys, utils.RATING_PLAN_PREFIX+keyResult.Key)
		}
		if err := iter.Close(); err != nil {
			CacheRollbackTransaction()
			return fmt.Errorf("rating plans: %s", err.Error())
		}
		CacheRemPrefixKey(utils.RATING_PLAN_PREFIX)
	} else if len(rpKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching rating plans: %v", rpKeys))
	}
	for _, key := range rpKeys {
		CacheRemKey(key)
		if _, err = ms.GetRatingPlan(key[len(utils.RATING_PLAN_PREFIX):], true); err != nil {
			CacheRollbackTransaction()
			return fmt.Errorf("rating plans: %s", err.Error())
		}
	}
	if len(rpKeys) != 0 {
		utils.Logger.Info("Finished rating plans caching.")
	}
	if rpfKeys == nil {
		utils.Logger.Info("Caching all rating profiles")
		iter := db.C(colRpf).Find(nil).Select(bson.M{"id": 1}).Iter()
		rpfKeys = make([]string, 0)
		for iter.Next(&idResult) {
			rpfKeys = append(rpfKeys, utils.RATING_PROFILE_PREFIX+idResult.Id)
		}
		if err := iter.Close(); err != nil {
			CacheRollbackTransaction()
			return fmt.Errorf("rating profiles: %s", err.Error())
		}
		CacheRemPrefixKey(utils.RATING_PROFILE_PREFIX)
	} else if len(rpfKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching rating profile: %v", rpfKeys))
	}
	for _, key := range rpfKeys {
		CacheRemKey(key)
		if _, err = ms.GetRatingProfile(key[len(utils.RATING_PROFILE_PREFIX):], true); err != nil {
			CacheRollbackTransaction()
			return fmt.Errorf("rating profiles: %s", err.Error())
		}
	}
	if len(rpfKeys) != 0 {
		utils.Logger.Info("Finished rating profile caching.")
	}
	if lcrKeys == nil {
		utils.Logger.Info("Caching LCR rules.")
		iter := db.C(colLcr).Find(nil).Select(bson.M{"key": 1}).Iter()
		lcrKeys = make([]string, 0)
		for iter.Next(&keyResult) {
			lcrKeys = append(lcrKeys, utils.LCR_PREFIX+keyResult.Key)
		}
		if err := iter.Close(); err != nil {
			CacheRollbackTransaction()
			return fmt.Errorf("lcr rules: %s", err.Error())
		}
		CacheRemPrefixKey(utils.LCR_PREFIX)
	} else if len(lcrKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching LCR rules: %v", lcrKeys))
	}
	for _, key := range lcrKeys {
		CacheRemKey(key)
		if _, err = ms.GetLCR(key[len(utils.LCR_PREFIX):], true); err != nil {
			CacheRollbackTransaction()
			return fmt.Errorf("lcr rules: %s", err.Error())
		}
	}
	if len(lcrKeys) != 0 {
		utils.Logger.Info("Finished LCR rules caching.")
	}
	// DerivedChargers caching
	if dcsKeys == nil {
		utils.Logger.Info("Caching all derived chargers")
		iter := db.C(colDcs).Find(nil).Select(bson.M{"key": 1}).Iter()
		dcsKeys = make([]string, 0)
		for iter.Next(&keyResult) {
			dcsKeys = append(dcsKeys, utils.DERIVEDCHARGERS_PREFIX+keyResult.Key)
		}
		if err := iter.Close(); err != nil {
			CacheRollbackTransaction()
			return fmt.Errorf("derived chargers: %s", err.Error())
		}
		CacheRemPrefixKey(utils.DERIVEDCHARGERS_PREFIX)
	} else if len(dcsKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching derived chargers: %v", dcsKeys))
	}
	for _, key := range dcsKeys {
		CacheRemKey(key)
		if _, err = ms.GetDerivedChargers(key[len(utils.DERIVEDCHARGERS_PREFIX):], true); err != nil {
			CacheRollbackTransaction()
			return fmt.Errorf("derived chargers: %s", err.Error())
		}
	}
	if len(dcsKeys) != 0 {
		utils.Logger.Info("Finished derived chargers caching.")
	}
	if actKeys == nil {
		CacheRemPrefixKey(utils.ACTION_PREFIX)
	}
	if actKeys == nil {
		utils.Logger.Info("Caching all actions")
		iter := db.C(colAct).Find(nil).Select(bson.M{"key": 1}).Iter()
		actKeys = make([]string, 0)
		for iter.Next(&keyResult) {
			actKeys = append(actKeys, utils.ACTION_PREFIX+keyResult.Key)
		}
		if err := iter.Close(); err != nil {
			CacheRollbackTransaction()
			return fmt.Errorf("actions: %s", err.Error())
		}
		CacheRemPrefixKey(utils.ACTION_PREFIX)
	} else if len(actKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching actions: %v", actKeys))
	}
	for _, key := range actKeys {
		CacheRemKey(key)
		if _, err = ms.GetActions(key[len(utils.ACTION_PREFIX):], true); err != nil {
			CacheRollbackTransaction()
			return fmt.Errorf("actions: %s", err.Error())
		}
	}
	if len(actKeys) != 0 {
		utils.Logger.Info("Finished actions caching.")
	}

	if aplKeys == nil {
		CacheRemPrefixKey(utils.ACTION_PLAN_PREFIX)
	}
	if aplKeys == nil {
		utils.Logger.Info("Caching all action plans")
		iter := db.C(colApl).Find(nil).Select(bson.M{"key": 1}).Iter()
		aplKeys = make([]string, 0)
		for iter.Next(&keyResult) {
			aplKeys = append(aplKeys, utils.ACTION_PLAN_PREFIX+keyResult.Key)
		}
		if err := iter.Close(); err != nil {
			CacheRollbackTransaction()
			return fmt.Errorf("action plans: %s", err.Error())
		}
		CacheRemPrefixKey(utils.ACTION_PLAN_PREFIX)
	} else if len(aplKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching action plans: %v", aplKeys))
	}
	for _, key := range aplKeys {
		CacheRemKey(key)
		if _, err = ms.GetActionPlan(key[len(utils.ACTION_PLAN_PREFIX):], true); err != nil {
			CacheRollbackTransaction()
			return fmt.Errorf("action plans: %s", err.Error())
		}
	}
	if len(aplKeys) != 0 {
		utils.Logger.Info("Finished action plans caching.")
	}

	if shgKeys == nil {
		CacheRemPrefixKey(utils.SHARED_GROUP_PREFIX)
	}
	if shgKeys == nil {
		utils.Logger.Info("Caching all shared groups")
		iter := db.C(colShg).Find(nil).Select(bson.M{"id": 1}).Iter()
		shgKeys = make([]string, 0)
		for iter.Next(&idResult) {
			shgKeys = append(shgKeys, utils.SHARED_GROUP_PREFIX+idResult.Id)
		}
		if err := iter.Close(); err != nil {
			CacheRollbackTransaction()
			return fmt.Errorf("shared groups: %s", err.Error())
		}
	} else if len(shgKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching shared groups: %v", shgKeys))
	}
	for _, key := range shgKeys {
		CacheRemKey(key)
		if _, err = ms.GetSharedGroup(key[len(utils.SHARED_GROUP_PREFIX):], true); err != nil {
			CacheRollbackTransaction()
			return fmt.Errorf("shared groups: %s", err.Error())
		}
	}
	if len(shgKeys) != 0 {
		utils.Logger.Info("Finished shared groups caching.")
	}
	CacheCommitTransaction()
	utils.Logger.Info(fmt.Sprintf("Cache rating creation time: %v", time.Since(start)))
	loadHistList, err := ms.GetLoadHistory(1, true)
	if err != nil || len(loadHistList) == 0 {
		utils.Logger.Info(fmt.Sprintf("could not get load history: %v (%v)", loadHistList, err))
	}
	var loadHist *utils.LoadInstance
	if len(loadHistList) == 0 {
		loadHist = &utils.LoadInstance{
			RatingLoadID:     utils.GenUUID(),
			AccountingLoadID: utils.GenUUID(),
			LoadID:           loadID,
			LoadTime:         time.Now(),
		}
	} else {
		loadHist = loadHistList[0]
		loadHist.RatingLoadID = utils.GenUUID()
		loadHist.LoadID = loadID
		loadHist.LoadTime = time.Now()
	}
	if err := ms.AddLoadHistory(loadHist, ms.loadHistorySize); err != nil {
		utils.Logger.Info(fmt.Sprintf("error saving load history: %v (%v)", loadHist, err))
		return err
	}
	ms.GetLoadHistory(1, true) // to load last instance in cache
	return utils.SaveCacheFileInfo(ms.cacheDumpDir, &utils.CacheFileInfo{Encoding: utils.MSGPACK, LoadInfo: loadHist})
}

func (ms *MongoStorage) CacheAccountingAll(loadID string) error {
	return ms.cacheAccounting(loadID, nil)
}

func (ms *MongoStorage) CacheAccountingPrefixes(loadID string, prefixes ...string) error {
	pm := map[string][]string{
		utils.ALIASES_PREFIX: []string{},
	}
	for _, prefix := range prefixes {
		if _, found := pm[prefix]; !found {
			return utils.ErrNotFound
		}
		pm[prefix] = nil
	}
	return ms.cacheAccounting(loadID, pm[utils.ALIASES_PREFIX])
}

func (ms *MongoStorage) CacheAccountingPrefixValues(loadID string, prefixes map[string][]string) error {
	pm := map[string][]string{
		utils.ALIASES_PREFIX: []string{},
	}
	for prefix, ids := range prefixes {
		if _, found := pm[prefix]; !found {
			return utils.ErrNotFound
		}
		pm[prefix] = ids
	}
	return ms.cacheAccounting(loadID, pm[utils.ALIASES_PREFIX])
}

func (ms *MongoStorage) cacheAccounting(loadID string, alsKeys []string) (err error) {
	start := time.Now()
	CacheBeginTransaction()
	var keyResult struct{ Key string }
	if alsKeys == nil {
		CacheRemPrefixKey(utils.ALIASES_PREFIX)
	}
	session := ms.session.Copy()
	defer session.Close()
	db := session.DB(ms.db)
	if alsKeys == nil {
		utils.Logger.Info("Caching all aliases")
		iter := db.C(colAls).Find(nil).Select(bson.M{"key": 1}).Iter()
		alsKeys = make([]string, 0)
		for iter.Next(&keyResult) {
			alsKeys = append(alsKeys, utils.ALIASES_PREFIX+keyResult.Key)
		}
		if err := iter.Close(); err != nil {
			CacheRollbackTransaction()
			return fmt.Errorf("aliases: %s", err.Error())
		}
	} else if len(alsKeys) != 0 {
		utils.Logger.Info(fmt.Sprintf("Caching aliases: %v", alsKeys))
	}
	for _, key := range alsKeys {
		// check if it already exists
		// to remove reverse cache keys
		if avs, err := CacheGet(key); err == nil && avs != nil {
			al := &Alias{Values: avs.(AliasValues)}
			al.SetId(key[len(utils.ALIASES_PREFIX):])
			al.RemoveReverseCache()
		}
		CacheRemKey(key)
		if _, err = ms.GetAlias(key[len(utils.ALIASES_PREFIX):], true); err != nil {
			CacheRollbackTransaction()
			return fmt.Errorf("aliases: %s", err.Error())
		}
	}
	if len(alsKeys) != 0 {
		utils.Logger.Info("Finished aliases caching.")
	}
	utils.Logger.Info("Caching load history")
	loadHistList, err := ms.GetLoadHistory(1, true)
	if err != nil {
		CacheRollbackTransaction()
		return err
	}
	utils.Logger.Info("Finished load history caching.")
	utils.Logger.Info(fmt.Sprintf("Cache accounting creation time: %v", time.Since(start)))

	var loadHist *utils.LoadInstance
	if len(loadHistList) == 0 {
		loadHist = &utils.LoadInstance{
			RatingLoadID:     utils.GenUUID(),
			AccountingLoadID: utils.GenUUID(),
			LoadID:           loadID,
			LoadTime:         time.Now(),
		}
	} else {
		loadHist = loadHistList[0]
		loadHist.AccountingLoadID = utils.GenUUID()
		loadHist.LoadID = loadID
		loadHist.LoadTime = time.Now()
	}
	if err := ms.AddLoadHistory(loadHist, ms.loadHistorySize); err != nil { //FIXME replace 100 with cfg
		utils.Logger.Info(fmt.Sprintf("error saving load history: %v (%v)", loadHist, err))
		return err
	}
	ms.GetLoadHistory(1, true) // to load last instance in cache
	return utils.SaveCacheFileInfo(ms.cacheDumpDir, &utils.CacheFileInfo{Encoding: utils.MSGPACK, LoadInfo: loadHist})
}

func (ms *MongoStorage) HasData(category, subject string) (bool, error) {
	session := ms.session.Copy()
	defer session.Close()
	db := session.DB(ms.db)
	switch category {
	case utils.DESTINATION_PREFIX:
		count, err := db.C(colDst).Find(bson.M{"key": subject}).Count()
		return count > 0, err
	case utils.RATING_PLAN_PREFIX:
		count, err := db.C(colRpl).Find(bson.M{"key": subject}).Count()
		return count > 0, err
	case utils.RATING_PROFILE_PREFIX:
		count, err := db.C(colRpf).Find(bson.M{"id": subject}).Count()
		return count > 0, err
	case utils.ACTION_PREFIX:
		count, err := db.C(colAct).Find(bson.M{"key": subject}).Count()
		return count > 0, err
	case utils.ACTION_PLAN_PREFIX:
		count, err := db.C(colApl).Find(bson.M{"key": subject}).Count()
		return count > 0, err
	case utils.ACCOUNT_PREFIX:
		count, err := db.C(colAcc).Find(bson.M{"id": subject}).Count()
		return count > 0, err
	}
	return false, errors.New("unsupported category in HasData")
}

func (ms *MongoStorage) GetRatingPlan(key string, skipCache bool) (rp *RatingPlan, err error) {
	if !skipCache {
		if x, err := CacheGet(utils.RATING_PLAN_PREFIX + key); err == nil {
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
	session, col := ms.conn(colRpl)
	defer session.Close()
	err = col.Find(bson.M{"key": key}).One(&kv)
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
		CacheSet(utils.RATING_PLAN_PREFIX+key, rp)
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
	session, col := ms.conn(colRpl)
	defer session.Close()
	_, err = col.Upsert(bson.M{"key": rp.Id}, &struct {
		Key   string
		Value []byte
	}{Key: rp.Id, Value: b.Bytes()})
	if err == nil && historyScribe != nil {
		var response int
		historyScribe.Call("HistoryV1.Record", rp.GetHistoryRecord(), &response)
	}
	return err
}

func (ms *MongoStorage) GetRatingProfile(key string, skipCache bool) (rp *RatingProfile, err error) {
	if !skipCache {
		if x, err := CacheGet(utils.RATING_PROFILE_PREFIX + key); err == nil {
			return x.(*RatingProfile), nil
		} else {
			return nil, err
		}
	}
	rp = new(RatingProfile)
	session, col := ms.conn(colRpf)
	defer session.Close()
	err = col.Find(bson.M{"id": key}).One(rp)
	if err == nil {
		CacheSet(utils.RATING_PROFILE_PREFIX+key, rp)
	}
	return
}

func (ms *MongoStorage) SetRatingProfile(rp *RatingProfile) error {
	session, col := ms.conn(colRpf)
	defer session.Close()
	_, err := col.Upsert(bson.M{"id": rp.Id}, rp)
	if err == nil && historyScribe != nil {
		var response int
		historyScribe.Call("HistoryV1.Record", rp.GetHistoryRecord(false), &response)
	}
	return err
}

func (ms *MongoStorage) RemoveRatingProfile(key string) error {
	session, col := ms.conn(colRpf)
	defer session.Close()
	iter := col.Find(bson.M{"id": bson.RegEx{Pattern: key + ".*", Options: ""}}).Select(bson.M{"id": 1}).Iter()
	var result struct{ Id string }
	for iter.Next(&result) {
		if err := col.Remove(bson.M{"id": result.Id}); err != nil {
			return err
		}
		CacheRemKey(utils.RATING_PROFILE_PREFIX + key)
		rpf := &RatingProfile{Id: result.Id}
		if historyScribe != nil {
			var response int
			go historyScribe.Call("HistoryV1.Record", rpf.GetHistoryRecord(true), &response)
		}
	}
	return iter.Close()
}

func (ms *MongoStorage) GetLCR(key string, skipCache bool) (lcr *LCR, err error) {
	if !skipCache {
		if x, err := CacheGet(utils.LCR_PREFIX + key); err == nil {
			return x.(*LCR), nil
		} else {
			return nil, err
		}
	}
	var result struct {
		Key   string
		Value *LCR
	}
	session, col := ms.conn(colLcr)
	defer session.Close()
	err = col.Find(bson.M{"key": key}).One(&result)
	if err == nil {
		lcr = result.Value
		CacheSet(utils.LCR_PREFIX+key, lcr)
	}
	return
}

func (ms *MongoStorage) SetLCR(lcr *LCR) error {
	session, col := ms.conn(colLcr)
	defer session.Close()
	_, err := col.Upsert(bson.M{"key": lcr.GetId()}, &struct {
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
	session, col := ms.conn(colDst)
	defer session.Close()
	err = col.Find(bson.M{"key": key}).One(&kv)
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
			CachePush(utils.DESTINATION_PREFIX+p, result.Id)
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
	session, col := ms.conn(colDst)
	defer session.Close()
	_, err = col.Upsert(bson.M{"key": dest.Id}, &struct {
		Key   string
		Value []byte
	}{Key: dest.Id, Value: b.Bytes()})
	if err == nil && historyScribe != nil {
		var response int
		historyScribe.Call("HistoryV1.Record", dest.GetHistoryRecord(false), &response)
	}
	return
}

func (ms *MongoStorage) RemoveDestination(destID string) (err error) {
	return
}

func (ms *MongoStorage) GetActions(key string, skipCache bool) (as Actions, err error) {
	if !skipCache {
		if x, err := CacheGet(utils.ACTION_PREFIX + key); err == nil {
			return x.(Actions), nil
		} else {
			return nil, err
		}
	}
	var result struct {
		Key   string
		Value Actions
	}
	session, col := ms.conn(colAct)
	defer session.Close()
	err = col.Find(bson.M{"key": key}).One(&result)
	if err == nil {
		as = result.Value
		CacheSet(utils.ACTION_PREFIX+key, as)
	}
	return
}

func (ms *MongoStorage) SetActions(key string, as Actions) error {
	session, col := ms.conn(colAct)
	defer session.Close()
	_, err := col.Upsert(bson.M{"key": key}, &struct {
		Key   string
		Value Actions
	}{Key: key, Value: as})
	return err
}

func (ms *MongoStorage) RemoveActions(key string) error {
	session, col := ms.conn(colAct)
	defer session.Close()
	return col.Remove(bson.M{"key": key})
}

func (ms *MongoStorage) GetSharedGroup(key string, skipCache bool) (sg *SharedGroup, err error) {
	if !skipCache {
		if x, err := CacheGet(utils.SHARED_GROUP_PREFIX + key); err == nil {
			return x.(*SharedGroup), nil
		} else {
			return nil, err
		}
	}
	session, col := ms.conn(colShg)
	defer session.Close()
	sg = &SharedGroup{}
	err = col.Find(bson.M{"id": key}).One(sg)
	if err == nil {
		CacheSet(utils.SHARED_GROUP_PREFIX+key, sg)
	}
	return
}

func (ms *MongoStorage) SetSharedGroup(sg *SharedGroup) (err error) {
	session, col := ms.conn(colShg)
	defer session.Close()
	_, err = col.Upsert(bson.M{"id": sg.Id}, sg)
	return err
}

func (ms *MongoStorage) GetAccount(key string) (result *Account, err error) {
	result = new(Account)
	session, col := ms.conn(colAcc)
	defer session.Close()
	err = col.Find(bson.M{"id": key}).One(result)
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
	session, col := ms.conn(colAcc)
	defer session.Close()
	_, err := col.Upsert(bson.M{"id": acc.ID}, acc)
	return err
}

func (ms *MongoStorage) RemoveAccount(key string) error {
	session, col := ms.conn(colAcc)
	defer session.Close()
	return col.Remove(bson.M{"id": key})

}

func (ms *MongoStorage) GetCdrStatsQueue(key string) (sq *StatsQueue, err error) {
	var result struct {
		Key   string
		Value *StatsQueue
	}
	session, col := ms.conn(colStq)
	defer session.Close()
	err = col.Find(bson.M{"key": key}).One(&result)
	if err == nil {
		sq = result.Value
	}
	return
}

func (ms *MongoStorage) SetCdrStatsQueue(sq *StatsQueue) (err error) {
	session, col := ms.conn(colStq)
	defer session.Close()
	_, err = col.Upsert(bson.M{"key": sq.GetId()}, &struct {
		Key   string
		Value *StatsQueue
	}{Key: sq.GetId(), Value: sq})
	return
}

func (ms *MongoStorage) GetSubscribers() (result map[string]*SubscriberData, err error) {
	session, col := ms.conn(colPbs)
	defer session.Close()
	iter := col.Find(nil).Iter()
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
	session, col := ms.conn(colPbs)
	defer session.Close()
	_, err = col.Upsert(bson.M{"key": key}, &struct {
		Key   string
		Value *SubscriberData
	}{Key: key, Value: sub})
	return err
}

func (ms *MongoStorage) RemoveSubscriber(key string) (err error) {
	session, col := ms.conn(colPbs)
	defer session.Close()
	return col.Remove(bson.M{"key": key})
}

func (ms *MongoStorage) SetUser(up *UserProfile) (err error) {
	session, col := ms.conn(colUsr)
	defer session.Close()
	_, err = col.Upsert(bson.M{"key": up.GetId()}, &struct {
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
	session, col := ms.conn(colUsr)
	defer session.Close()
	err = col.Find(bson.M{"key": key}).One(&kv)
	if err == nil {
		up = kv.Value
	}
	return
}

func (ms *MongoStorage) GetUsers() (result []*UserProfile, err error) {
	session, col := ms.conn(colUsr)
	defer session.Close()
	iter := col.Find(nil).Iter()
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
	session, col := ms.conn(colUsr)
	defer session.Close()
	return col.Remove(bson.M{"key": key})
}

func (ms *MongoStorage) SetAlias(al *Alias) (err error) {
	session, col := ms.conn(colAls)
	defer session.Close()
	_, err = col.Upsert(bson.M{"key": al.GetId()}, &struct {
		Key   string
		Value AliasValues
	}{Key: al.GetId(), Value: al.Values})
	return err
}

func (ms *MongoStorage) GetAlias(key string, skipCache bool) (al *Alias, err error) {
	origKey := key
	key = utils.ALIASES_PREFIX + key
	if !skipCache {
		if x, err := CacheGet(key); err == nil {
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
	session, col := ms.conn(colAls)
	defer session.Close()
	if err = col.Find(bson.M{"key": origKey}).One(&kv); err == nil {
		al = &Alias{Values: kv.Value}
		al.SetId(origKey)
		if err == nil {
			CacheSet(key, al.Values)
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
	session, col := ms.conn(colAls)
	defer session.Close()
	if err := col.Find(bson.M{"key": origKey}).One(&kv); err == nil {
		al.Values = kv.Value
	}
	err = col.Remove(bson.M{"key": origKey})
	if err == nil {
		al.RemoveReverseCache()
		CacheRemKey(key)
	}
	return
}

// Limit will only retrieve the last n items out of history, newest first
func (ms *MongoStorage) GetLoadHistory(limit int, skipCache bool) (loadInsts []*utils.LoadInstance, err error) {
	if limit == 0 {
		return nil, nil
	}
	if !skipCache {
		if x, err := CacheGet(utils.LOADINST_KEY); err != nil {
			return nil, err
		} else {
			items := x.([]*utils.LoadInstance)
			if len(items) < limit || limit == -1 {
				return items, nil
			}
			return items[:limit], nil
		}
	}
	var kv struct {
		Key   string
		Value []*utils.LoadInstance
	}
	session, col := ms.conn(colLht)
	defer session.Close()
	err = col.Find(bson.M{"key": utils.LOADINST_KEY}).One(&kv)
	if err == nil {
		loadInsts = kv.Value
		CacheRemKey(utils.LOADINST_KEY)
		CacheSet(utils.LOADINST_KEY, loadInsts)
	}
	if len(loadInsts) < limit || limit == -1 {
		return loadInsts, nil
	}
	return loadInsts[:limit], nil
}

// Adds a single load instance to load history
func (ms *MongoStorage) AddLoadHistory(ldInst *utils.LoadInstance, loadHistSize int) error {
	if loadHistSize == 0 { // Load history disabled
		return nil
	}
	// get existing load history
	var existingLoadHistory []*utils.LoadInstance
	var kv struct {
		Key   string
		Value []*utils.LoadInstance
	}
	session, col := ms.conn(colLht)
	defer session.Close()
	err := col.Find(bson.M{"key": utils.LOADINST_KEY}).One(&kv)

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
		session, col := ms.conn(colLht)
		defer session.Close()
		_, err = col.Upsert(bson.M{"key": utils.LOADINST_KEY}, &struct {
			Key   string
			Value []*utils.LoadInstance
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
	session, col := ms.conn(colAtr)
	defer session.Close()
	err = col.Find(bson.M{"key": key}).One(&kv)
	if err == nil {
		atrs = kv.Value
	}
	return
}

func (ms *MongoStorage) SetActionTriggers(key string, atrs ActionTriggers) (err error) {
	session, col := ms.conn(colAtr)
	defer session.Close()
	if len(atrs) == 0 {
		err = col.Remove(bson.M{"key": key}) // delete the key
		if err != mgo.ErrNotFound {
			return err
		}
		return nil
	}
	_, err = col.Upsert(bson.M{"key": key}, &struct {
		Key   string
		Value ActionTriggers
	}{Key: key, Value: atrs})
	return err
}

func (ms *MongoStorage) RemoveActionTriggers(key string) error {
	session, col := ms.conn(colAtr)
	defer session.Close()
	return col.Remove(bson.M{"key": key})
}

func (ms *MongoStorage) GetActionPlan(key string, skipCache bool) (ats *ActionPlan, err error) {
	if !skipCache {
		if x, err := CacheGet(utils.ACTION_PLAN_PREFIX + key); err == nil {
			return x.(*ActionPlan), nil
		} else {
			return nil, err
		}
	}
	var kv struct {
		Key   string
		Value []byte
	}
	session, col := ms.conn(colApl)
	defer session.Close()
	err = col.Find(bson.M{"key": key}).One(&kv)
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
		CacheSet(utils.ACTION_PLAN_PREFIX+key, ats)
	}
	return
}

func (ms *MongoStorage) SetActionPlan(key string, ats *ActionPlan, overwrite bool) error {
	session, col := ms.conn(colApl)
	defer session.Close()
	// clean dots from account ids map
	if len(ats.ActionTimings) == 0 {
		CacheRemKey(utils.ACTION_PLAN_PREFIX + key)
		err := col.Remove(bson.M{"key": key})
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
	_, err = col.Upsert(bson.M{"key": key}, &struct {
		Key   string
		Value []byte
	}{Key: key, Value: b.Bytes()})
	return err
}

func (ms *MongoStorage) GetAllActionPlans() (ats map[string]*ActionPlan, err error) {
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

func (ms *MongoStorage) PushTask(t *Task) error {
	session, col := ms.conn(colTsk)
	defer session.Close()
	return col.Insert(bson.M{"_id": bson.NewObjectId(), "task": t})
}

func (ms *MongoStorage) PopTask() (t *Task, err error) {
	v := struct {
		ID   bson.ObjectId `bson:"_id"`
		Task *Task
	}{}
	session, col := ms.conn(colTsk)
	defer session.Close()
	if err = col.Find(nil).One(&v); err == nil {
		if remErr := col.Remove(bson.M{"_id": v.ID}); remErr != nil {
			return nil, remErr
		}
		t = v.Task
	}

	return
}

func (ms *MongoStorage) GetDerivedChargers(key string, skipCache bool) (dcs *utils.DerivedChargers, err error) {
	if !skipCache {
		if x, err := CacheGet(utils.DERIVEDCHARGERS_PREFIX + key); err == nil {
			return x.(*utils.DerivedChargers), nil
		} else {
			return nil, err
		}
	}
	var kv struct {
		Key   string
		Value *utils.DerivedChargers
	}
	session, col := ms.conn(colDcs)
	defer session.Close()
	err = col.Find(bson.M{"key": key}).One(&kv)
	if err == nil {
		dcs = kv.Value
		CacheSet(utils.DERIVEDCHARGERS_PREFIX+key, dcs)
	}
	return
}

func (ms *MongoStorage) SetDerivedChargers(key string, dcs *utils.DerivedChargers) (err error) {
	if dcs == nil || len(dcs.Chargers) == 0 {
		CacheRemKey(utils.DERIVEDCHARGERS_PREFIX + key)
		session, col := ms.conn(colDcs)
		defer session.Close()
		err = col.Remove(bson.M{"key": key})
		if err != mgo.ErrNotFound {
			return err
		}
		return nil
	}
	session, col := ms.conn(colDcs)
	defer session.Close()
	_, err = col.Upsert(bson.M{"key": key}, &struct {
		Key   string
		Value *utils.DerivedChargers
	}{Key: key, Value: dcs})
	return err
}

func (ms *MongoStorage) SetCdrStats(cs *CdrStats) error {
	session, col := ms.conn(colCrs)
	defer session.Close()
	_, err := col.Upsert(bson.M{"id": cs.Id}, cs)
	return err
}

func (ms *MongoStorage) GetCdrStats(key string) (cs *CdrStats, err error) {
	cs = &CdrStats{}
	session, col := ms.conn(colCrs)
	defer session.Close()
	err = col.Find(bson.M{"id": key}).One(cs)
	return
}

func (ms *MongoStorage) GetAllCdrStats() (css []*CdrStats, err error) {
	session, col := ms.conn(colCrs)
	defer session.Close()
	iter := col.Find(nil).Iter()
	var cs CdrStats
	for iter.Next(&cs) {
		clone := cs // avoid using the same pointer in append
		css = append(css, &clone)
	}
	err = iter.Close()
	return
}

func (ms *MongoStorage) SetStructVersion(v *StructVersion) (err error) {
	session, col := ms.conn(colVer)
	defer session.Close()
	_, err = col.Upsert(bson.M{"key": utils.VERSION_PREFIX + "struct"}, &struct {
		Key   string
		Value *StructVersion
	}{utils.VERSION_PREFIX + "struct", v})
	return
}

func (ms *MongoStorage) GetStructVersion() (rsv *StructVersion, err error) {
	var result struct {
		Key   string
		Value StructVersion
	}
	session, col := ms.conn(colVer)
	defer session.Close()
	err = col.Find(bson.M{"key": utils.VERSION_PREFIX + "struct"}).One(&result)
	if err == mgo.ErrNotFound {
		rsv = nil
	}
	rsv = &result.Value
	return
}
