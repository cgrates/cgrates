/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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
	"github.com/cgrates/cgrates/cache"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"strings"
	"time"
)

const (
	colDst   = "destinations"
	colRds   = "reverse_destinations"
	colAct   = "actions"
	colApl   = "action_plans"
	colAAp   = "account_action_plans"
	colTsk   = "tasks"
	colAtr   = "action_triggers"
	colRpl   = "rating_plans"
	colRpf   = "rating_profiles"
	colAcc   = "accounts"
	colShg   = "shared_groups"
	colLcr   = "lcr_rules"
	colDcs   = "derived_chargers"
	colAls   = "aliases"
	colRCfgs = "reverse_aliases"
	colStq   = "stat_qeues"
	colPbs   = "pubsub"
	colUsr   = "users"
	colCrs   = "cdr_stats"
	colLht   = "load_history"
	colVer   = "versions"
	colRsP   = "resource_profiles"
	colRFI   = "request_filter_indexes"
	colTmg   = "timings"
	colRes   = "resources"
	colSqs   = "statqueues"
	colSqp   = "statqueue_profiles"
	colTps   = "threshold_profiles"
	colThs   = "thresholds"
	colFlt   = "filters"
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

func NewMongoStorage(host, port, db, user, pass, storageType string, cdrsIndexes []string, cacheCfg config.CacheConfig, loadHistorySize int) (ms *MongoStorage, err error) {
	url := host
	if port != "" {
		url += ":" + port
	}
	if user != "" && pass != "" {
		url = fmt.Sprintf("%s:%s@%s", user, pass, url)
	}
	if db != "" {
		url += "/" + db
	}
	session, err := mgo.Dial(url)
	if err != nil {
		return nil, err
	}
	session.SetMode(mgo.Strong, true)
	ms = &MongoStorage{db: db, session: session, storageType: storageType, ms: NewCodecMsgpackMarshaler(),
		cacheCfg: cacheCfg, loadHistorySize: loadHistorySize, cdrsIndexes: cdrsIndexes}
	if cNames, err := session.DB(ms.db).CollectionNames(); err != nil {
		return nil, err
	} else if len(cNames) == 0 { // create indexes only if database is empty
		if err = ms.EnsureIndexes(); err != nil {
			return nil, err
		}
	}
	ms.cnter = utils.NewCounter(time.Now().UnixNano(), 0)
	return
}

type MongoStorage struct {
	session         *mgo.Session
	db              string
	storageType     string // datadb, stordb
	ms              Marshaler
	cacheCfg        config.CacheConfig
	loadHistorySize int
	cdrsIndexes     []string
	cnter           *utils.Counter
}

func (ms *MongoStorage) conn(col string) (*mgo.Session, *mgo.Collection) {
	sessionCopy := ms.session.Copy()
	return sessionCopy, sessionCopy.DB(ms.db).C(col)
}

// EnsureIndexes creates db indexes
func (ms *MongoStorage) EnsureIndexes() (err error) {
	dbSession := ms.session.Copy()
	defer dbSession.Close()
	db := dbSession.DB(ms.db)
	idx := mgo.Index{
		Key:        []string{"key"},
		Unique:     true,  // Prevent two documents from having the same index key
		DropDups:   false, // Drop documents with the same index key as a previously indexed one
		Background: false, // Build index in background and return immediately
		Sparse:     false, // Only index documents containing the Key fields
	}
	var colectNames []string // collection names containing this index
	if ms.storageType == utils.DataDB {
		colectNames = []string{colAct, colApl, colAAp, colAtr, colDcs, colRsP, colRpl, colLcr, colDst, colRds, colAls, colUsr, colLht, colRes}
	}
	for _, col := range colectNames {
		if err = db.C(col).EnsureIndex(idx); err != nil {
			return
		}
	}
	idx = mgo.Index{
		Key:        []string{"id"},
		Unique:     true,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	}
	if ms.storageType == utils.DataDB {
		colectNames = []string{colRpf, colShg, colCrs, colAcc}
	}
	for _, col := range colectNames {
		if err = db.C(col).EnsureIndex(idx); err != nil {
			return
		}
	}
	if ms.storageType == utils.StorDB {
		idx = mgo.Index{
			Key:        []string{"tpid", "id"},
			Unique:     true,
			DropDups:   false,
			Background: false,
			Sparse:     false,
		}
		for _, col := range []string{utils.TBLTPTimings, utils.TBLTPDestinations, utils.TBLTPDestinationRates, utils.TBLTPRatingPlans,
			utils.TBLTPSharedGroups, utils.TBLTPCdrStats, utils.TBLTPActions, utils.TBLTPActionPlans, utils.TBLTPActionTriggers, utils.TBLTPStats, utils.TBLTPResources} {
			if err = db.C(col).EnsureIndex(idx); err != nil {
				return
			}
		}
		idx = mgo.Index{
			Key:        []string{"tpid", "direction", "tenant", "category", "subject", "loadid"},
			Unique:     true,
			DropDups:   false,
			Background: false,
			Sparse:     false,
		}
		if err = db.C(utils.TBLTPRateProfiles).EnsureIndex(idx); err != nil {
			return
		}
		idx = mgo.Index{
			Key:        []string{"tpid", "direction", "tenant", "category", "account", "subject"},
			Unique:     true,
			DropDups:   false,
			Background: false,
			Sparse:     false,
		}
		if err = db.C(utils.TBLTPLcrs).EnsureIndex(idx); err != nil {
			return
		}
		idx = mgo.Index{
			Key:        []string{"tpid", "tenant", "username"},
			Unique:     true,
			DropDups:   false,
			Background: false,
			Sparse:     false,
		}
		if err = db.C(utils.TBLTPUsers).EnsureIndex(idx); err != nil {
			return
		}
		idx = mgo.Index{
			Key:        []string{"tpid", "direction", "tenant", "category", "account", "subject", "context"},
			Unique:     true,
			DropDups:   false,
			Background: false,
			Sparse:     false,
		}
		if err = db.C(utils.TBLTPLcrs).EnsureIndex(idx); err != nil {
			return
		}
		idx = mgo.Index{
			Key:        []string{"tpid", "direction", "tenant", "category", "subject", "account", "loadid"},
			Unique:     true,
			DropDups:   false,
			Background: false,
			Sparse:     false,
		}
		if err = db.C(utils.TBLTPDerivedChargers).EnsureIndex(idx); err != nil {
			return
		}
		idx = mgo.Index{
			Key:        []string{"tpid", "direction", "tenant", "account", "loadid"},
			Unique:     true,
			DropDups:   false,
			Background: false,
			Sparse:     false,
		}
		if err = db.C(utils.TBLTPDerivedChargers).EnsureIndex(idx); err != nil {
			return
		}
		idx = mgo.Index{
			Key:        []string{CGRIDLow, RunIDLow, OriginIDLow},
			Unique:     true,
			DropDups:   false,
			Background: false,
			Sparse:     false,
		}
		if err = db.C(utils.TBLCDRs).EnsureIndex(idx); err != nil {
			return
		}
		for _, idxKey := range ms.cdrsIndexes {
			idx = mgo.Index{
				Key:        []string{idxKey},
				Unique:     false,
				DropDups:   false,
				Background: false,
				Sparse:     false,
			}
			if err = db.C(utils.TBLCDRs).EnsureIndex(idx); err != nil {
				return
			}
		}
		idx = mgo.Index{
			Key:        []string{CGRIDLow, RunIDLow},
			Unique:     true,
			DropDups:   false,
			Background: false,
			Sparse:     false,
		}
		if err = db.C(utils.TBLSMCosts).EnsureIndex(idx); err != nil {
			return
		}
		idx = mgo.Index{
			Key:        []string{OriginHostLow, OriginIDLow},
			Unique:     false,
			DropDups:   false,
			Background: false,
			Sparse:     false,
		}
		if err = db.C(utils.TBLSMCosts).EnsureIndex(idx); err != nil {
			return
		}
		idx = mgo.Index{
			Key:        []string{RunIDLow, OriginIDLow},
			Unique:     false,
			DropDups:   false,
			Background: false,
			Sparse:     false,
		}
		if err = db.C(utils.TBLSMCosts).EnsureIndex(idx); err != nil {
			return
		}
	}
	return
}

func (ms *MongoStorage) getColNameForPrefix(prefix string) (name string, ok bool) {
	colMap := map[string]string{
		utils.DESTINATION_PREFIX:         colDst,
		utils.REVERSE_DESTINATION_PREFIX: colRds,
		utils.ACTION_PREFIX:              colAct,
		utils.ACTION_PLAN_PREFIX:         colApl,
		utils.AccountActionPlansPrefix:   colAAp,
		utils.TASKS_KEY:                  colTsk,
		utils.ACTION_TRIGGER_PREFIX:      colAtr,
		utils.RATING_PLAN_PREFIX:         colRpl,
		utils.RATING_PROFILE_PREFIX:      colRpf,
		utils.ACCOUNT_PREFIX:             colAcc,
		utils.SHARED_GROUP_PREFIX:        colShg,
		utils.LCR_PREFIX:                 colLcr,
		utils.DERIVEDCHARGERS_PREFIX:     colDcs,
		utils.ALIASES_PREFIX:             colAls,
		utils.REVERSE_ALIASES_PREFIX:     colRCfgs,
		utils.PUBSUB_SUBSCRIBERS_PREFIX:  colPbs,
		utils.USERS_PREFIX:               colUsr,
		utils.CDR_STATS_PREFIX:           colCrs,
		utils.LOADINST_KEY:               colLht,
		utils.VERSION_PREFIX:             colVer,
		//utils.CDR_STATS_QUEUE_PREFIX:            colStq,
		utils.TimingsPrefix:          colTmg,
		utils.ResourcesPrefix:        colRes,
		utils.ResourceProfilesPrefix: colRsP,
		utils.ThresholdProfilePrefix: colTps,
		utils.ThresholdPrefix:        colThs,
		utils.FilterPrefix:           colFlt,
	}
	name, ok = colMap[prefix]
	return
}

func (ms *MongoStorage) Close() {
	ms.session.Close()
}

func (ms *MongoStorage) Flush(ignore string) (err error) {
	dbSession := ms.session.Copy()
	defer dbSession.Close()
	return dbSession.DB(ms.db).DropDatabase()
}

func (ms *MongoStorage) Marshaler() Marshaler {
	return ms.ms
}

// DB returnes a database object with cloned session inside
func (ms *MongoStorage) DB() *mgo.Database {
	return ms.session.Copy().DB(ms.db)
}

func (ms *MongoStorage) SelectDatabase(dbName string) (err error) {
	ms.db = dbName
	return
}

func (ms *MongoStorage) RebuildReverseForPrefix(prefix string) (err error) {
	if !utils.IsSliceMember([]string{utils.REVERSE_DESTINATION_PREFIX, utils.REVERSE_ALIASES_PREFIX, utils.AccountActionPlansPrefix}, prefix) {
		return utils.ErrInvalidKey
	}
	colName, ok := ms.getColNameForPrefix(prefix)
	if !ok {
		return utils.ErrInvalidKey
	}
	session, col := ms.conn(colName)
	defer session.Close()
	if _, err := col.RemoveAll(bson.M{}); err != nil {
		return err
	}
	var keys []string
	switch prefix {
	case utils.REVERSE_DESTINATION_PREFIX:
		if keys, err = ms.GetKeysForPrefix(utils.DESTINATION_PREFIX); err != nil {
			return
		}
		for _, key := range keys {
			dest, err := ms.GetDestination(key[len(utils.DESTINATION_PREFIX):], true, utils.NonTransactional)
			if err != nil {
				return err
			}
			if err = ms.SetReverseDestination(dest, utils.NonTransactional); err != nil {
				return err
			}
		}
	case utils.REVERSE_ALIASES_PREFIX:
		if keys, err = ms.GetKeysForPrefix(utils.ALIASES_PREFIX); err != nil {
			return
		}
		for _, key := range keys {
			al, err := ms.GetAlias(key[len(utils.ALIASES_PREFIX):], true, utils.NonTransactional)
			if err != nil {
				return err
			}
			if err = ms.SetReverseAlias(al, utils.NonTransactional); err != nil {
				return err
			}
		}
	case utils.AccountActionPlansPrefix:
		if keys, err = ms.GetKeysForPrefix(utils.ACTION_PLAN_PREFIX); err != nil {
			return
		}
		for _, key := range keys {
			apl, err := ms.GetActionPlan(key[len(utils.ACTION_PLAN_PREFIX):], true, utils.NonTransactional)
			if err != nil {
				return err
			}
			for acntID := range apl.AccountIDs {
				if err = ms.SetAccountActionPlans(acntID, []string{apl.Id}, false); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (ms *MongoStorage) LoadDataDBCache(dstIDs, rvDstIDs, rplIDs, rpfIDs, actIDs, aplIDs, aaPlIDs, atrgIDs, sgIDs, lcrIDs, dcIDs, alsIDs, rvAlsIDs, rpIDs, resIDs []string) (err error) {
	for key, ids := range map[string][]string{
		utils.DESTINATION_PREFIX:         dstIDs,
		utils.REVERSE_DESTINATION_PREFIX: rvDstIDs,
		utils.RATING_PLAN_PREFIX:         rplIDs,
		utils.RATING_PROFILE_PREFIX:      rpfIDs,
		utils.ACTION_PREFIX:              actIDs,
		utils.ACTION_PLAN_PREFIX:         aplIDs,
		utils.AccountActionPlansPrefix:   aaPlIDs,
		utils.ACTION_TRIGGER_PREFIX:      atrgIDs,
		utils.SHARED_GROUP_PREFIX:        sgIDs,
		utils.LCR_PREFIX:                 lcrIDs,
		utils.DERIVEDCHARGERS_PREFIX:     dcIDs,
		utils.ALIASES_PREFIX:             alsIDs,
		utils.REVERSE_ALIASES_PREFIX:     rvAlsIDs,
		utils.ResourceProfilesPrefix:     rpIDs,
		utils.ResourcesPrefix:            resIDs,
	} {
		if err = ms.CacheDataFromDB(key, ids, false); err != nil {
			return
		}
	}
	return
}

// CacheDataFromDB loads data to cache
// prfx represents the cache prefix, ids should be nil if all available data should be loaded
// mustBeCached specifies that data needs to be cached in order to be retrieved from db
func (ms *MongoStorage) CacheDataFromDB(prfx string, ids []string, mustBeCached bool) (err error) {
	if !utils.IsSliceMember([]string{utils.DESTINATION_PREFIX,
		utils.REVERSE_DESTINATION_PREFIX,
		utils.RATING_PLAN_PREFIX,
		utils.RATING_PROFILE_PREFIX,
		utils.ACTION_PREFIX,
		utils.ACTION_PLAN_PREFIX,
		utils.AccountActionPlansPrefix,
		utils.ACTION_TRIGGER_PREFIX,
		utils.SHARED_GROUP_PREFIX,
		utils.DERIVEDCHARGERS_PREFIX,
		utils.LCR_PREFIX,
		utils.ALIASES_PREFIX,
		utils.REVERSE_ALIASES_PREFIX,
		utils.ResourceProfilesPrefix,
		utils.TimingsPrefix,
		utils.ResourcesPrefix}, prfx) {
		return utils.NewCGRError(utils.MONGO,
			utils.MandatoryIEMissingCaps,
			utils.UnsupportedCachePrefix,
			fmt.Sprintf("prefix <%s> is not a supported cache prefix", prfx))
	}
	if ids == nil {
		keyIDs, err := ms.GetKeysForPrefix(prfx)
		if err != nil {
			return utils.NewCGRError(utils.MONGO,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("mongo error <%s> querying keys for prefix: <%s>", prfx))
		}
		for _, keyID := range keyIDs {
			if mustBeCached { // Only consider loading ids which are already in cache
				if _, hasIt := cache.Get(keyID); !hasIt {
					continue
				}
			}
			ids = append(ids, keyID[len(prfx):])
		}
		var nrItems int
		if cCfg, has := ms.cacheCfg[utils.CachePrefixToInstance[prfx]]; has {
			nrItems = cCfg.Limit
		}
		if nrItems > 0 && nrItems < len(ids) { // More ids than cache config allows it, limit here
			ids = ids[:nrItems]
		}
	}
	for _, dataID := range ids {
		if mustBeCached {
			if _, hasIt := cache.Get(prfx + dataID); !hasIt { // only cache if previously there
				continue
			}
		}
		switch prfx {
		case utils.DESTINATION_PREFIX:
			_, err = ms.GetDestination(dataID, true, utils.NonTransactional)
		case utils.REVERSE_DESTINATION_PREFIX:
			_, err = ms.GetReverseDestination(dataID, true, utils.NonTransactional)
		case utils.RATING_PLAN_PREFIX:
			_, err = ms.GetRatingPlan(dataID, true, utils.NonTransactional)
		case utils.RATING_PROFILE_PREFIX:
			_, err = ms.GetRatingProfile(dataID, true, utils.NonTransactional)
		case utils.ACTION_PREFIX:
			_, err = ms.GetActions(dataID, true, utils.NonTransactional)
		case utils.ACTION_PLAN_PREFIX:
			_, err = ms.GetActionPlan(dataID, true, utils.NonTransactional)
		case utils.AccountActionPlansPrefix:
			_, err = ms.GetAccountActionPlans(dataID, true, utils.NonTransactional)
		case utils.ACTION_TRIGGER_PREFIX:
			_, err = ms.GetActionTriggers(dataID, true, utils.NonTransactional)
		case utils.SHARED_GROUP_PREFIX:
			_, err = ms.GetSharedGroup(dataID, true, utils.NonTransactional)
		case utils.DERIVEDCHARGERS_PREFIX:
			_, err = ms.GetDerivedChargers(dataID, true, utils.NonTransactional)
		case utils.LCR_PREFIX:
			_, err = ms.GetLCR(dataID, true, utils.NonTransactional)
		case utils.ALIASES_PREFIX:
			_, err = ms.GetAlias(dataID, true, utils.NonTransactional)
		case utils.REVERSE_ALIASES_PREFIX:
			_, err = ms.GetReverseAlias(dataID, true, utils.NonTransactional)
		case utils.ResourceProfilesPrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = ms.GetResourceProfile(tntID.Tenant, tntID.ID, true, utils.NonTransactional)
		case utils.ResourcesPrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = ms.GetResource(tntID.Tenant, tntID.ID, true, utils.NonTransactional)
		case utils.TimingsPrefix:
			_, err = ms.GetTiming(dataID, true, utils.NonTransactional)
		case utils.ThresholdProfilePrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = ms.GetThresholdProfile(tntID.Tenant, tntID.ID, true, utils.NonTransactional)
		case utils.ThresholdPrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = ms.GetThreshold(tntID.Tenant, tntID.ID, true, utils.NonTransactional)
		case utils.FilterPrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = ms.GetFilterDrv(tntID.Tenant, tntID.ID)
		}
		if err != nil {
			return utils.NewCGRError(utils.MONGO,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("error <%s> querying mongo for category: <%s>, dataID: <%s>", prfx, dataID))
		}
	}
	return
}

func (ms *MongoStorage) IsDBEmpty() (resp bool, err error) {
	session := ms.session.Copy()
	defer session.Close()
	db := session.DB(ms.db)
	cols, err := db.CollectionNames()
	if err != nil {
		return
	}
	return len(cols) == 0 || cols[0] == "cdrs", nil
}

func (ms *MongoStorage) GetKeysForPrefix(prefix string) (result []string, err error) {
	var category, subject string
	keyLen := len(utils.DESTINATION_PREFIX)
	if len(prefix) < keyLen {
		return nil, fmt.Errorf("unsupported prefix in GetKeysForPrefix: %s", prefix)
	}
	category = prefix[:keyLen] // prefix lenght
	tntID := utils.NewTenantID(prefix[keyLen:])
	subject = fmt.Sprintf("^%s", prefix[keyLen:]) // old way, no tenant support
	session := ms.session.Copy()
	defer session.Close()
	db := session.DB(ms.db)
	keyResult := struct{ Key string }{}
	idResult := struct{ Tenant, Id string }{}
	switch category {
	case utils.DESTINATION_PREFIX:
		iter := db.C(colDst).Find(bson.M{"key": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"key": 1}).Iter()
		for iter.Next(&keyResult) {
			result = append(result, utils.DESTINATION_PREFIX+keyResult.Key)
		}
	case utils.REVERSE_DESTINATION_PREFIX:
		iter := db.C(colRds).Find(bson.M{"key": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"key": 1}).Iter()
		for iter.Next(&keyResult) {
			result = append(result, utils.REVERSE_DESTINATION_PREFIX+keyResult.Key)
		}
	case utils.RATING_PLAN_PREFIX:
		iter := db.C(colRpl).Find(bson.M{"key": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"key": 1}).Iter()
		for iter.Next(&keyResult) {
			result = append(result, utils.RATING_PLAN_PREFIX+keyResult.Key)
		}
	case utils.RATING_PROFILE_PREFIX:
		iter := db.C(colRpf).Find(bson.M{"id": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"id": 1}).Iter()
		for iter.Next(&idResult) {
			result = append(result, utils.RATING_PROFILE_PREFIX+idResult.Id)
		}
	case utils.ACTION_PREFIX:
		iter := db.C(colAct).Find(bson.M{"key": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"key": 1}).Iter()
		for iter.Next(&keyResult) {
			result = append(result, utils.ACTION_PREFIX+keyResult.Key)
		}
	case utils.ACTION_PLAN_PREFIX:
		iter := db.C(colApl).Find(bson.M{"key": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"key": 1}).Iter()
		for iter.Next(&keyResult) {
			result = append(result, utils.ACTION_PLAN_PREFIX+keyResult.Key)
		}
	case utils.ACTION_TRIGGER_PREFIX:
		iter := db.C(colAtr).Find(bson.M{"key": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"key": 1}).Iter()
		for iter.Next(&keyResult) {
			result = append(result, utils.ACTION_TRIGGER_PREFIX+keyResult.Key)
		}
	case utils.SHARED_GROUP_PREFIX:
		iter := db.C(colShg).Find(bson.M{"id": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"id": 1}).Iter()
		for iter.Next(&idResult) {
			result = append(result, utils.SHARED_GROUP_PREFIX+idResult.Id)
		}
	case utils.DERIVEDCHARGERS_PREFIX:
		iter := db.C(colDcs).Find(bson.M{"key": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"key": 1}).Iter()
		for iter.Next(&keyResult) {
			result = append(result, utils.DERIVEDCHARGERS_PREFIX+keyResult.Key)
		}
	case utils.LCR_PREFIX:
		iter := db.C(colLcr).Find(bson.M{"key": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"key": 1}).Iter()
		for iter.Next(&keyResult) {
			result = append(result, utils.LCR_PREFIX+keyResult.Key)
		}
	case utils.ACCOUNT_PREFIX:
		iter := db.C(colAcc).Find(bson.M{"id": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"id": 1}).Iter()
		for iter.Next(&idResult) {
			result = append(result, utils.ACCOUNT_PREFIX+idResult.Id)
		}
	case utils.ALIASES_PREFIX:
		iter := db.C(colAls).Find(bson.M{"key": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"key": 1}).Iter()
		for iter.Next(&keyResult) {
			result = append(result, utils.ALIASES_PREFIX+keyResult.Key)
		}
	case utils.REVERSE_ALIASES_PREFIX:
		iter := db.C(colRCfgs).Find(bson.M{"key": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"key": 1}).Iter()
		for iter.Next(&keyResult) {
			result = append(result, utils.REVERSE_ALIASES_PREFIX+keyResult.Key)
		}
	case utils.ResourceProfilesPrefix:
		iter := db.C(colRsP).Find(bson.M{"id": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"tenant": 1, "id": 1}).Iter()
		for iter.Next(&idResult) {
			result = append(result, utils.ResourceProfilesPrefix+utils.ConcatenatedKey(idResult.Tenant, idResult.Id))
		}
	case utils.ResourcesPrefix:
		iter := db.C(colRes).Find(bson.M{"id": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"tenant": 1, "id": 1}).Iter()
		for iter.Next(&idResult) {
			result = append(result, utils.ResourcesPrefix+utils.ConcatenatedKey(idResult.Tenant, idResult.Id))
		}
	case utils.StatQueuePrefix:
		qry := bson.M{}
		if tntID.Tenant != "" {
			qry["tenant"] = tntID.Tenant
		}
		if tntID.ID != "" {
			qry["id"] = bson.M{"$regex": bson.RegEx{Pattern: subject}}
		}
		iter := db.C(colSqs).Find(qry).Select(bson.M{"tenant": 1, "id": 1}).Iter()
		for iter.Next(&idResult) {
			result = append(result, utils.StatQueuePrefix+utils.ConcatenatedKey(idResult.Tenant, idResult.Id))
		}
	case utils.StatQueueProfilePrefix:
		iter := db.C(colSqp).Find(bson.M{"id": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"tenant": 1, "id": 1}).Iter()
		for iter.Next(&idResult) {
			result = append(result, utils.StatQueueProfilePrefix+utils.ConcatenatedKey(idResult.Tenant, idResult.Id))
		}
	case utils.AccountActionPlansPrefix:
		iter := db.C(colAAp).Find(bson.M{"id": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"id": 1}).Iter()
		for iter.Next(&idResult) {
			result = append(result, utils.AccountActionPlansPrefix+idResult.Id)
		}
	case utils.TimingsPrefix:
		iter := db.C(colTmg).Find(bson.M{"id": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"id": 1}).Iter()
		for iter.Next(&idResult) {
			result = append(result, utils.TimingsPrefix+idResult.Id)
		}
	case utils.FilterPrefix:
		iter := db.C(colFlt).Find(bson.M{"id": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"tenant": 1, "id": 1}).Iter()
		for iter.Next(&idResult) {
			result = append(result, utils.FilterPrefix+utils.ConcatenatedKey(idResult.Tenant, idResult.Id))
		}
	case utils.ThresholdPrefix:
		qry := bson.M{}
		if tntID.Tenant != "" {
			qry["tenant"] = tntID.Tenant
		}
		if tntID.ID != "" {
			qry["id"] = bson.M{"$regex": bson.RegEx{Pattern: subject}}
		}
		iter := db.C(colThs).Find(qry).Select(bson.M{"tenant": 1, "id": 1}).Iter()
		for iter.Next(&idResult) {
			result = append(result, utils.ThresholdPrefix+utils.ConcatenatedKey(idResult.Tenant, idResult.Id))
		}
	default:
		err = fmt.Errorf("unsupported prefix in GetKeysForPrefix: %s", prefix)
	}
	return
}

func (ms *MongoStorage) HasData(category, subject string) (has bool, err error) {
	session := ms.session.Copy()
	defer session.Close()
	db := session.DB(ms.db)
	var count int
	switch category {
	case utils.DESTINATION_PREFIX:
		count, err = db.C(colDst).Find(bson.M{"key": subject}).Count()
		has = count > 0
	case utils.RATING_PLAN_PREFIX:
		count, err = db.C(colRpl).Find(bson.M{"key": subject}).Count()
		has = count > 0
	case utils.RATING_PROFILE_PREFIX:
		count, err = db.C(colRpf).Find(bson.M{"id": subject}).Count()
		has = count > 0
	case utils.ACTION_PREFIX:
		count, err = db.C(colAct).Find(bson.M{"key": subject}).Count()
		has = count > 0
	case utils.ACTION_PLAN_PREFIX:
		count, err = db.C(colApl).Find(bson.M{"key": subject}).Count()
		has = count > 0
	case utils.ACCOUNT_PREFIX:
		count, err = db.C(colAcc).Find(bson.M{"id": subject}).Count()
		has = count > 0
	case utils.ResourcesPrefix:
		count, err = db.C(colRes).Find(bson.M{"id": subject}).Count()
		has = count > 0
	case utils.StatQueuePrefix:
		count, err = db.C(colRes).Find(bson.M{"id": subject}).Count()
		has = count > 0
	case utils.ThresholdPrefix:
		count, err = db.C(colTps).Find(bson.M{"id": subject}).Count()
		has = count > 0
	case utils.FilterPrefix:
		count, err = db.C(colFlt).Find(bson.M{"id": subject}).Count()
		has = count > 0
	default:
		err = fmt.Errorf("unsupported category in HasData: %s", category)
	}
	return
}

func (ms *MongoStorage) GetRatingPlan(key string, skipCache bool, transactionID string) (rp *RatingPlan, err error) {
	cacheKey := utils.RATING_PLAN_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(cacheKey); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*RatingPlan), nil
		}
	}
	rp = new(RatingPlan)
	var kv struct {
		Key   string
		Value []byte
	}
	session, col := ms.conn(colRpl)
	defer session.Close()
	if err = col.Find(bson.M{"key": key}).One(&kv); err != nil {
		if err == mgo.ErrNotFound {
			cache.Set(cacheKey, nil, cacheCommit(transactionID), transactionID)
			err = utils.ErrNotFound
		}
		return nil, err
	}
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
	if err = ms.ms.Unmarshal(out, &rp); err != nil {
		return nil, err
	}
	cache.Set(cacheKey, rp, cacheCommit(transactionID), transactionID)
	return
}

func (ms *MongoStorage) SetRatingPlan(rp *RatingPlan, transactionID string) error {
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

func (ms *MongoStorage) GetRatingProfile(key string, skipCache bool, transactionID string) (rp *RatingProfile, err error) {
	cacheKey := utils.RATING_PROFILE_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(cacheKey); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*RatingProfile), nil
		}
	}
	session, col := ms.conn(colRpf)
	defer session.Close()
	rp = new(RatingProfile)
	if err = col.Find(bson.M{"id": key}).One(rp); err != nil {
		if err == mgo.ErrNotFound {
			cache.Set(cacheKey, nil, cacheCommit(transactionID), transactionID)
			err = utils.ErrNotFound
		}
		return nil, err // Make sure we don't return new object on error
	}
	cache.Set(cacheKey, rp, cacheCommit(transactionID), transactionID)
	return
}

func (ms *MongoStorage) SetRatingProfile(rp *RatingProfile, transactionID string) (err error) {
	session, col := ms.conn(colRpf)
	defer session.Close()
	if _, err = col.Upsert(bson.M{"id": rp.Id}, rp); err != nil {
		return
	}
	if historyScribe != nil {
		var response int
		historyScribe.Call("HistoryV1.Record", rp.GetHistoryRecord(false), &response)
	}
	return
}

func (ms *MongoStorage) RemoveRatingProfile(key, transactionID string) error {
	session, col := ms.conn(colRpf)
	defer session.Close()
	iter := col.Find(bson.M{"id": key}).Select(bson.M{"id": 1}).Iter()
	var result struct{ Id string }
	for iter.Next(&result) {
		if err := col.Remove(bson.M{"id": result.Id}); err != nil {
			return err
		}
		cache.RemKey(utils.RATING_PROFILE_PREFIX+key, cacheCommit(transactionID), transactionID)
		rpf := &RatingProfile{Id: result.Id}
		if historyScribe != nil {
			var response int
			go historyScribe.Call("HistoryV1.Record", rpf.GetHistoryRecord(true), &response)
		}
	}
	return iter.Close()
}

func (ms *MongoStorage) GetLCR(key string, skipCache bool, transactionID string) (lcr *LCR, err error) {
	cacheKey := utils.LCR_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(cacheKey); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*LCR), nil
		}
	}
	var result struct {
		Key   string
		Value *LCR
	}
	session, col := ms.conn(colLcr)
	defer session.Close()
	cCommit := cacheCommit(transactionID)
	if err = col.Find(bson.M{"key": key}).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			cache.Set(cacheKey, nil, cCommit, transactionID)
			err = utils.ErrNotFound
		}
		return nil, err
	}
	lcr = result.Value
	cache.Set(cacheKey, lcr, cCommit, transactionID)
	return
}

func (ms *MongoStorage) SetLCR(lcr *LCR, transactionID string) (err error) {
	session, col := ms.conn(colLcr)
	defer session.Close()
	if _, err = col.Upsert(bson.M{"key": lcr.GetId()}, &struct {
		Key   string
		Value *LCR
	}{lcr.GetId(), lcr}); err != nil {
		return
	}
	return
}

func (ms *MongoStorage) GetDestination(key string, skipCache bool, transactionID string) (result *Destination, err error) {
	cacheKey := utils.DESTINATION_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(cacheKey); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*Destination), nil
		}
	}
	var kv struct {
		Key   string
		Value []byte
	}
	session, col := ms.conn(colDst)
	defer session.Close()
	if err = col.Find(bson.M{"key": key}).One(&kv); err != nil {
		if err == mgo.ErrNotFound {
			cache.Set(cacheKey, nil, cacheCommit(transactionID), transactionID)
			err = utils.ErrNotFound
		}
		return
	}
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
	cache.Set(cacheKey, result, cacheCommit(transactionID), transactionID)
	return
}

func (ms *MongoStorage) SetDestination(dest *Destination, transactionID string) (err error) {
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
	if _, err = col.Upsert(bson.M{"key": dest.Id}, &struct {
		Key   string
		Value []byte
	}{Key: dest.Id, Value: b.Bytes()}); err != nil {
		return
	}
	if historyScribe != nil {
		var response int
		historyScribe.Call("HistoryV1.Record", dest.GetHistoryRecord(false), &response)
	}
	return
}

func (ms *MongoStorage) GetReverseDestination(prefix string, skipCache bool, transactionID string) (ids []string, err error) {
	cacheKey := utils.REVERSE_DESTINATION_PREFIX + prefix
	if !skipCache {
		if x, ok := cache.Get(cacheKey); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.([]string), nil
		}
	}
	var result struct {
		Key   string
		Value []string
	}
	session, col := ms.conn(colRds)
	defer session.Close()
	if err = col.Find(bson.M{"key": prefix}).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			cache.Set(cacheKey, nil, cacheCommit(transactionID), transactionID)
			err = utils.ErrNotFound
		}
		return
	}
	ids = result.Value
	cache.Set(cacheKey, ids, cacheCommit(transactionID), transactionID)
	return
}

func (ms *MongoStorage) SetReverseDestination(dest *Destination, transactionID string) (err error) {
	session, col := ms.conn(colRds)
	defer session.Close()
	for _, p := range dest.Prefixes {
		if _, err = col.Upsert(bson.M{"key": p}, bson.M{"$addToSet": bson.M{"value": dest.Id}}); err != nil {
			break
		}
	}
	return
}

func (ms *MongoStorage) RemoveDestination(destID string, transactionID string) (err error) {
	session, col := ms.conn(colDst)
	key := utils.DESTINATION_PREFIX + destID
	// get destination for prefix list
	d, err := ms.GetDestination(destID, false, transactionID)
	if err != nil {
		if err == mgo.ErrNotFound {
			err = nil
		}
		return
	}
	err = col.Remove(bson.M{"key": destID})
	if err != nil {
		return err
	}
	cache.RemKey(key, cacheCommit(transactionID), transactionID)
	session.Close()

	session, col = ms.conn(colRds)
	defer session.Close()
	for _, prefix := range d.Prefixes {
		err = col.Update(bson.M{"key": prefix}, bson.M{"$pull": bson.M{"value": destID}})
		if err != nil {
			return err
		}
		ms.GetReverseDestination(prefix, true, transactionID) // it will recache the destination
	}
	return
}

func (ms *MongoStorage) UpdateReverseDestination(oldDest, newDest *Destination, transactionID string) error {
	session, col := ms.conn(colRds)
	defer session.Close()
	//log.Printf("Old: %+v, New: %+v", oldDest, newDest)
	var obsoletePrefixes []string
	var addedPrefixes []string
	var found bool
	if oldDest == nil {
		oldDest = new(Destination) // so we can process prefixes
	}
	for _, oldPrefix := range oldDest.Prefixes {
		found = false
		for _, newPrefix := range newDest.Prefixes {
			if oldPrefix == newPrefix {
				found = true
				break
			}
		}
		if !found {
			obsoletePrefixes = append(obsoletePrefixes, oldPrefix)
		}
	}

	for _, newPrefix := range newDest.Prefixes {
		found = false
		for _, oldPrefix := range oldDest.Prefixes {
			if newPrefix == oldPrefix {
				found = true
				break
			}
		}
		if !found {
			addedPrefixes = append(addedPrefixes, newPrefix)
		}
	}
	//log.Print("Obsolete prefixes: ", obsoletePrefixes)
	//log.Print("Added prefixes: ", addedPrefixes)
	// remove id for all obsolete prefixes
	cCommit := cacheCommit(transactionID)
	var err error
	for _, obsoletePrefix := range obsoletePrefixes {
		err = col.Update(bson.M{"key": obsoletePrefix}, bson.M{"$pull": bson.M{"value": oldDest.Id}})
		if err != nil {
			return err
		}
		cache.RemKey(utils.REVERSE_DESTINATION_PREFIX+obsoletePrefix, cCommit, transactionID)
	}

	// add the id to all new prefixes
	for _, addedPrefix := range addedPrefixes {
		_, err = col.Upsert(bson.M{"key": addedPrefix}, bson.M{"$addToSet": bson.M{"value": newDest.Id}})
		if err != nil {
			return err
		}
	}
	return nil
}

func (ms *MongoStorage) GetActions(key string, skipCache bool, transactionID string) (as Actions, err error) {
	cacheKey := utils.ACTION_PREFIX + key
	if !skipCache {
		if x, err := cache.GetCloned(cacheKey); err != nil {
			if err.Error() != utils.ItemNotFound {
				return nil, err
			}
		} else if x == nil {
			return nil, utils.ErrNotFound
		} else {
			return x.(Actions), nil
		}
	}
	var result struct {
		Key   string
		Value Actions
	}
	session, col := ms.conn(colAct)
	defer session.Close()
	if err = col.Find(bson.M{"key": key}).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			cache.Set(cacheKey, nil, cacheCommit(transactionID), transactionID)
			err = utils.ErrNotFound
		}
		return nil, err
	}
	as = result.Value
	cache.Set(utils.ACTION_PREFIX+key, as, cacheCommit(transactionID), transactionID)
	return
}

func (ms *MongoStorage) SetActions(key string, as Actions, transactionID string) error {
	session, col := ms.conn(colAct)
	defer session.Close()
	_, err := col.Upsert(bson.M{"key": key}, &struct {
		Key   string
		Value Actions
	}{Key: key, Value: as})
	return err
}

func (ms *MongoStorage) RemoveActions(key string, transactionID string) error {
	session, col := ms.conn(colAct)
	defer session.Close()
	err := col.Remove(bson.M{"key": key})
	cache.RemKey(utils.ACTION_PREFIX+key, cacheCommit(transactionID), transactionID)
	return err
}

func (ms *MongoStorage) GetSharedGroup(key string, skipCache bool, transactionID string) (sg *SharedGroup, err error) {
	cacheKey := utils.SHARED_GROUP_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(cacheKey); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*SharedGroup), nil
		}
	}
	session, col := ms.conn(colShg)
	defer session.Close()
	sg = new(SharedGroup)
	if err = col.Find(bson.M{"id": key}).One(sg); err != nil {
		if err == mgo.ErrNotFound {
			cache.Set(cacheKey, nil, cacheCommit(transactionID), transactionID)
			err = utils.ErrNotFound
		}
		return nil, err
	}
	cache.Set(cacheKey, sg, cacheCommit(transactionID), transactionID)
	return
}

func (ms *MongoStorage) SetSharedGroup(sg *SharedGroup, transactionID string) (err error) {
	session, col := ms.conn(colShg)
	defer session.Close()
	if _, err = col.Upsert(bson.M{"id": sg.Id}, sg); err != nil {
		return
	}
	return
}

func (ms *MongoStorage) GetAccount(key string) (result *Account, err error) {
	result = new(Account)
	session, col := ms.conn(colAcc)
	defer session.Close()
	err = col.Find(bson.M{"id": key}).One(result)
	if err == mgo.ErrNotFound {
		err = utils.ErrNotFound
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

func (ms *MongoStorage) GetCdrStatsQueue(key string) (sq *CDRStatsQueue, err error) {
	var result struct {
		Key   string
		Value *CDRStatsQueue
	}
	session, col := ms.conn(colStq)
	defer session.Close()
	if err = col.Find(bson.M{"key": key}).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			err = utils.ErrNotFound
		}
		return nil, err
	}
	sq = result.Value
	return
}

func (ms *MongoStorage) SetCdrStatsQueue(sq *CDRStatsQueue) (err error) {
	session, col := ms.conn(colStq)
	defer session.Close()
	_, err = col.Upsert(bson.M{"key": sq.GetId()}, &struct {
		Key   string
		Value *CDRStatsQueue
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
	if err = col.Find(bson.M{"key": key}).One(&kv); err != nil {
		if err == mgo.ErrNotFound {
			err = utils.ErrNotFound
		}
		return nil, err
	}
	up = kv.Value
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

func (ms *MongoStorage) GetAlias(key string, skipCache bool, transactionID string) (al *Alias, err error) {
	cacheKey := utils.ALIASES_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(cacheKey); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			al = x.(*Alias)
			return
		}
	}
	var kv struct {
		Key   string
		Value AliasValues
	}
	session, col := ms.conn(colAls)
	defer session.Close()
	cCommit := cacheCommit(transactionID)
	if err = col.Find(bson.M{"key": key}).One(&kv); err != nil {
		if err == mgo.ErrNotFound {
			cache.Set(cacheKey, nil, cacheCommit(transactionID), transactionID)
			err = utils.ErrNotFound
		}
		return
	}
	al = &Alias{Values: kv.Value}
	al.SetId(key)
	cache.Set(cacheKey, al, cCommit, transactionID)
	return
}

func (ms *MongoStorage) SetAlias(al *Alias, transactionID string) (err error) {
	session, col := ms.conn(colAls)
	defer session.Close()
	if _, err = col.Upsert(bson.M{"key": al.GetId()}, &struct {
		Key   string
		Value AliasValues
	}{Key: al.GetId(), Value: al.Values}); err != nil {
		return
	}
	return err
}

func (ms *MongoStorage) GetReverseAlias(reverseID string, skipCache bool, transactionID string) (ids []string, err error) {
	cacheKey := utils.REVERSE_ALIASES_PREFIX + reverseID
	if !skipCache {
		if x, ok := cache.Get(cacheKey); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.([]string), nil
		}
	}
	var result struct {
		Key   string
		Value []string
	}
	session, col := ms.conn(colRCfgs)
	defer session.Close()
	if err = col.Find(bson.M{"key": reverseID}).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			cache.Set(cacheKey, nil, cacheCommit(transactionID), transactionID)
			err = utils.ErrNotFound
		}
		return
	}
	ids = result.Value
	cache.Set(cacheKey, ids, cacheCommit(transactionID), transactionID)
	return
}

func (ms *MongoStorage) SetReverseAlias(al *Alias, transactionID string) (err error) {
	session, col := ms.conn(colRCfgs)
	defer session.Close()
	for _, value := range al.Values {
		for target, pairs := range value.Pairs {
			for _, alias := range pairs {
				rKey := strings.Join([]string{alias, target, al.Context}, "")
				id := utils.ConcatenatedKey(al.GetId(), value.DestinationId)
				if _, err = col.Upsert(bson.M{"key": rKey}, bson.M{"$addToSet": bson.M{"value": id}}); err != nil {
					break
				}
			}
		}
	}
	return
}

func (ms *MongoStorage) RemoveAlias(key, transactionID string) (err error) {
	al := &Alias{}
	al.SetId(key)
	origKey := key
	key = utils.ALIASES_PREFIX + key
	var kv struct {
		Key   string
		Value AliasValues
	}
	session, col := ms.conn(colAls)
	if err := col.Find(bson.M{"key": origKey}).One(&kv); err != nil {
		if err == mgo.ErrNotFound {
			err = nil
		}
		return err
	}
	al.Values = kv.Value
	if err = col.Remove(bson.M{"key": origKey}); err != nil {
		return
	}
	cCommit := cacheCommit(transactionID)
	cache.RemKey(key, cCommit, transactionID)
	session.Close()
	session, col = ms.conn(colRCfgs)
	defer session.Close()
	for _, value := range al.Values {
		tmpKey := utils.ConcatenatedKey(al.GetId(), value.DestinationId)
		for target, pairs := range value.Pairs {
			for _, alias := range pairs {
				rKey := alias + target + al.Context
				if err = col.Update(bson.M{"key": rKey}, bson.M{"$pull": bson.M{"value": tmpKey}}); err != nil {
					if err == mgo.ErrNotFound {
						err = nil // cancel the error not to be propagated with return bellow
					} else {
						return err
					}
				}
				cache.RemKey(utils.REVERSE_ALIASES_PREFIX+rKey, cCommit, transactionID)
			}
		}
	}
	return
}

// Limit will only retrieve the last n items out of history, newest first
func (ms *MongoStorage) GetLoadHistory(limit int, skipCache bool, transactionID string) (loadInsts []*utils.LoadInstance, err error) {
	if limit == 0 {
		return nil, nil
	}
	if !skipCache {
		if x, ok := cache.Get(utils.LOADINST_KEY); ok {
			if x != nil {
				items := x.([]*utils.LoadInstance)
				if len(items) < limit || limit == -1 {
					return items, nil
				}
				return items[:limit], nil
			}
			return nil, utils.ErrNotFound
		}
	}
	var kv struct {
		Key   string
		Value []*utils.LoadInstance
	}
	session, col := ms.conn(colLht)
	defer session.Close()
	err = col.Find(bson.M{"key": utils.LOADINST_KEY}).One(&kv)
	cCommit := cacheCommit(transactionID)
	if err == nil {
		loadInsts = kv.Value
		cache.RemKey(utils.LOADINST_KEY, cCommit, transactionID)
		cache.Set(utils.LOADINST_KEY, loadInsts, cCommit, transactionID)
	}
	if len(loadInsts) < limit || limit == -1 {
		return loadInsts, nil
	}
	return loadInsts[:limit], nil
}

// Adds a single load instance to load history
func (ms *MongoStorage) AddLoadHistory(ldInst *utils.LoadInstance, loadHistSize int, transactionID string) error {
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

	_, err = guardian.Guardian.Guard(func() (interface{}, error) { // Make sure we do it locked since other instance can modify history while we read it
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

	cache.RemKey(utils.LOADINST_KEY, cacheCommit(transactionID), transactionID)
	return err
}

func (ms *MongoStorage) GetActionTriggers(key string, skipCache bool, transactionID string) (atrs ActionTriggers, err error) {
	cacheKey := utils.ACTION_TRIGGER_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(cacheKey); ok {
			if x != nil {
				return x.(ActionTriggers), nil
			}
			return nil, utils.ErrNotFound
		}
	}

	var kv struct {
		Key   string
		Value ActionTriggers
	}
	session, col := ms.conn(colAtr)
	defer session.Close()
	if err = col.Find(bson.M{"key": key}).One(&kv); err != nil {
		if err == mgo.ErrNotFound {
			cache.Set(cacheKey, nil, cacheCommit(transactionID), transactionID)
			err = utils.ErrNotFound
		}
		return
	}
	atrs = kv.Value
	cache.Set(cacheKey, atrs, cacheCommit(transactionID), transactionID)
	return
}

func (ms *MongoStorage) SetActionTriggers(key string, atrs ActionTriggers, transactionID string) (err error) {
	session, col := ms.conn(colAtr)
	defer session.Close()
	if len(atrs) == 0 {
		err = col.Remove(bson.M{"key": key})
		if err == mgo.ErrNotFound { // Overwrite not found since it is not really mandatory here to be returned
			err = nil
		}
		return
	}
	_, err = col.Upsert(bson.M{"key": key}, &struct {
		Key   string
		Value ActionTriggers
	}{Key: key, Value: atrs})
	return err
}

func (ms *MongoStorage) RemoveActionTriggers(key string, transactionID string) error {
	session, col := ms.conn(colAtr)
	defer session.Close()
	err := col.Remove(bson.M{"key": key})
	if err == nil {
		cache.RemKey(key, cacheCommit(transactionID), transactionID)
	}
	return err
}

func (ms *MongoStorage) GetActionPlan(key string, skipCache bool, transactionID string) (ats *ActionPlan, err error) {
	cacheKey := utils.ACTION_PLAN_PREFIX + key
	if !skipCache {
		if x, err := cache.GetCloned(cacheKey); err != nil {
			if err.Error() != utils.ItemNotFound { // Only consider cache if item was found
				return nil, err
			}
		} else if x == nil { // item was placed nil in cache
			return nil, utils.ErrNotFound
		} else {
			return x.(*ActionPlan), nil
		}
	}
	var kv struct {
		Key   string
		Value []byte
	}
	session, col := ms.conn(colApl)
	defer session.Close()
	if err = col.Find(bson.M{"key": key}).One(&kv); err != nil {
		if err == mgo.ErrNotFound {
			cache.Set(cacheKey, nil, cacheCommit(transactionID), transactionID)
			err = utils.ErrNotFound
		}
		return nil, err
	}
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
	if err = ms.ms.Unmarshal(out, &ats); err != nil {
		return nil, err
	}
	cache.Set(cacheKey, ats, cacheCommit(transactionID), transactionID)
	return
}

func (ms *MongoStorage) SetActionPlan(key string, ats *ActionPlan, overwrite bool, transactionID string) (err error) {
	session, col := ms.conn(colApl)
	defer session.Close()
	dbKey := utils.ACTION_PLAN_PREFIX + key
	// clean dots from account ids map
	cCommit := cacheCommit(transactionID)
	if len(ats.ActionTimings) == 0 {
		cache.RemKey(dbKey, cCommit, transactionID)
		if err = col.Remove(bson.M{"key": key}); err != nil && err == mgo.ErrNotFound {
			err = nil // NotFound is good
		}
		return
	}
	if !overwrite {
		// get existing action plan to merge the account ids
		if existingAts, _ := ms.GetActionPlan(key, true, transactionID); existingAts != nil {
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
	if _, err = col.Upsert(bson.M{"key": key}, &struct {
		Key   string
		Value []byte
	}{Key: key, Value: b.Bytes()}); err != nil {
		return
	}
	return err
}

func (ms *MongoStorage) GetAllActionPlans() (ats map[string]*ActionPlan, err error) {
	keys, err := ms.GetKeysForPrefix(utils.ACTION_PLAN_PREFIX)
	if err != nil {
		return nil, err
	}

	ats = make(map[string]*ActionPlan, len(keys))
	for _, key := range keys {
		ap, err := ms.GetActionPlan(key[len(utils.ACTION_PLAN_PREFIX):], false, utils.NonTransactional)
		if err != nil {
			return nil, err
		}
		ats[key[len(utils.ACTION_PLAN_PREFIX):]] = ap
	}

	return
}

func (ms *MongoStorage) GetAccountActionPlans(acntID string, skipCache bool, transactionID string) (aPlIDs []string, err error) {
	cacheKey := utils.AccountActionPlansPrefix + acntID
	if !skipCache {
		if x, ok := cache.Get(cacheKey); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.([]string), nil
		}
	}
	session, col := ms.conn(colAAp)
	defer session.Close()
	var kv struct {
		Key   string
		Value []string
	}
	if err = col.Find(bson.M{"key": acntID}).One(&kv); err != nil {
		if err == mgo.ErrNotFound {
			cache.Set(cacheKey, nil, cacheCommit(transactionID), transactionID)
			err = utils.ErrNotFound
		}
		return
	}
	aPlIDs = kv.Value
	cache.Set(cacheKey, aPlIDs, cacheCommit(transactionID), transactionID)
	return
}

func (ms *MongoStorage) SetAccountActionPlans(acntID string, aPlIDs []string, overwrite bool) (err error) {
	session, col := ms.conn(colAAp)
	defer session.Close()
	if !overwrite {
		if oldaPlIDs, err := ms.GetAccountActionPlans(acntID, false, utils.NonTransactional); err != nil && err != utils.ErrNotFound {
			return err
		} else {
			for _, oldAPid := range oldaPlIDs {
				if !utils.IsSliceMember(aPlIDs, oldAPid) {
					aPlIDs = append(aPlIDs, oldAPid)
				}
			}
		}
	}
	_, err = col.Upsert(bson.M{"key": acntID}, &struct {
		Key   string
		Value []string
	}{Key: acntID, Value: aPlIDs})
	return
}

func (ms *MongoStorage) RemAccountActionPlans(acntID string, aPlIDs []string) (err error) {
	session, col := ms.conn(colAAp)
	defer session.Close()
	if len(aPlIDs) == 0 {
		return col.Remove(bson.M{"key": acntID})
	}
	oldAPlIDs, err := ms.GetAccountActionPlans(acntID, true, utils.NonTransactional)
	if err != nil {
		return err
	}
	for i := 0; i < len(oldAPlIDs); {
		if utils.IsSliceMember(aPlIDs, oldAPlIDs[i]) {
			oldAPlIDs = append(oldAPlIDs[:i], oldAPlIDs[i+1:]...)
			continue // if we have stripped, don't increase index so we can check next element by next run
		}
		i++
	}
	if len(oldAPlIDs) == 0 { // no more elements, remove the reference
		return col.Remove(bson.M{"key": acntID})
	}
	_, err = col.Upsert(bson.M{"key": acntID}, &struct {
		Key   string
		Value []string
	}{Key: acntID, Value: oldAPlIDs})
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

func (ms *MongoStorage) GetDerivedChargers(key string, skipCache bool, transactionID string) (dcs *utils.DerivedChargers, err error) {
	cacheKey := utils.DERIVEDCHARGERS_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(cacheKey); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*utils.DerivedChargers), nil
		}
	}
	var kv struct {
		Key   string
		Value *utils.DerivedChargers
	}
	session, col := ms.conn(colDcs)
	defer session.Close()
	cCommit := cacheCommit(transactionID)
	if err = col.Find(bson.M{"key": key}).One(&kv); err != nil {
		if err == mgo.ErrNotFound {
			cache.Set(cacheKey, nil, cCommit, transactionID)
			err = utils.ErrNotFound
		}
		return
	}
	dcs = kv.Value
	cache.Set(cacheKey, dcs, cCommit, transactionID)
	return
}

func (ms *MongoStorage) SetDerivedChargers(key string, dcs *utils.DerivedChargers, transactionID string) (err error) {
	cCommit := cacheCommit(transactionID)
	cacheKey := utils.DERIVEDCHARGERS_PREFIX + key
	if dcs == nil || len(dcs.Chargers) == 0 {

		session, col := ms.conn(colDcs)
		defer session.Close()
		if err = col.Remove(bson.M{"key": key}); err != nil && err != mgo.ErrNotFound {
			return
		}
		cache.RemKey(cacheKey, cCommit, transactionID)
		return nil
	}
	session, col := ms.conn(colDcs)
	defer session.Close()
	if _, err = col.Upsert(bson.M{"key": key}, &struct {
		Key   string
		Value *utils.DerivedChargers
	}{Key: key, Value: dcs}); err != nil {
		return
	}
	return err
}

func (ms *MongoStorage) SetCdrStats(cs *CdrStats) error {
	session, col := ms.conn(colCrs)
	defer session.Close()
	_, err := col.Upsert(bson.M{"id": cs.Id}, cs)
	return err
}

func (ms *MongoStorage) GetCdrStats(key string) (cs *CdrStats, err error) {
	cs = new(CdrStats)
	session, col := ms.conn(colCrs)
	defer session.Close()
	if err = col.Find(bson.M{"id": key}).One(cs); err != nil {
		if err == mgo.ErrNotFound {
			err = utils.ErrNotFound
		}
		return nil, err
	}
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

func (ms *MongoStorage) GetResourceProfile(tenant, id string, skipCache bool, transactionID string) (rp *ResourceProfile, err error) {
	key := utils.ResourceProfilesPrefix + utils.ConcatenatedKey(tenant, id)
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*ResourceProfile), nil
		}
	}
	session, col := ms.conn(colRsP)
	defer session.Close()
	rp = new(ResourceProfile)
	if err = col.Find(bson.M{"tenant": tenant, "id": id}).One(rp); err != nil {
		if err == mgo.ErrNotFound {
			err = utils.ErrNotFound
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	for _, fltr := range rp.Filters {
		if err = fltr.CompileValues(); err != nil {
			return
		}
	}
	cache.Set(key, rp, cacheCommit(transactionID), transactionID)
	return
}

func (ms *MongoStorage) SetResourceProfile(rp *ResourceProfile) (err error) {
	session, col := ms.conn(colRsP)
	defer session.Close()
	_, err = col.Upsert(bson.M{"tenant": rp.Tenant, "id": rp.ID}, rp)
	return
}

func (ms *MongoStorage) RemoveResourceProfile(tenant, id string, transactionID string) (err error) {
	session, col := ms.conn(colRsP)
	defer session.Close()
	if err = col.Remove(bson.M{"tenant": tenant, "id": id}); err != nil {
		return
	}
	cache.RemKey(utils.ResourceProfilesPrefix+utils.ConcatenatedKey(tenant, id),
		cacheCommit(transactionID), transactionID)
	return nil
}

func (ms *MongoStorage) GetResource(tenant, id string, skipCache bool, transactionID string) (r *Resource, err error) {
	key := utils.ResourcesPrefix + utils.ConcatenatedKey(tenant, id)
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*Resource), nil
		}
	}
	session, col := ms.conn(colRes)
	defer session.Close()
	r = new(Resource)
	if err = col.Find(bson.M{"tenant": tenant, "id": id}).One(r); err != nil {
		if err == mgo.ErrNotFound {
			err = utils.ErrNotFound
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	cache.Set(key, r, cacheCommit(transactionID), transactionID)
	return
}

func (ms *MongoStorage) SetResource(r *Resource) (err error) {
	session, col := ms.conn(colRes)
	defer session.Close()
	_, err = col.Upsert(bson.M{"tenant": r.Tenant, "id": r.ID}, r)
	return
}

func (ms *MongoStorage) RemoveResource(tenant, id string, transactionID string) (err error) {
	session, col := ms.conn(colRes)
	defer session.Close()
	if err = col.Remove(bson.M{"tenant": tenant, "id": id}); err != nil {
		return
	}
	cache.RemKey(utils.ResourcesPrefix+utils.ConcatenatedKey(tenant, id),
		cacheCommit(transactionID), transactionID)
	return nil
}

func (ms *MongoStorage) GetTiming(id string, skipCache bool, transactionID string) (t *utils.TPTiming, err error) {
	key := utils.TimingsPrefix + id
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*utils.TPTiming), nil
		}
	}
	session, col := ms.conn(colTmg)
	defer session.Close()
	t = new(utils.TPTiming)
	if err = col.Find(bson.M{"id": id}).One(t); err != nil {
		if err == mgo.ErrNotFound {
			err = utils.ErrNotFound
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	cache.Set(key, t, cacheCommit(transactionID), transactionID)
	return
}

func (ms *MongoStorage) SetTiming(t *utils.TPTiming, transactionID string) (err error) {
	session, col := ms.conn(colTmg)
	defer session.Close()
	_, err = col.Upsert(bson.M{"id": t.ID}, t)
	return
}

func (ms *MongoStorage) RemoveTiming(id string, transactionID string) (err error) {
	session, col := ms.conn(colTmg)
	defer session.Close()
	if err = col.Remove(bson.M{"id": id}); err != nil {
		return
	}
	cache.RemKey(utils.TimingsPrefix+id, cacheCommit(transactionID), transactionID)
	return nil
}

func (ms *MongoStorage) GetReqFilterIndexes(dbKey string) (indexes map[string]map[string]utils.StringMap, err error) {
	session, col := ms.conn(colRFI)
	defer session.Close()
	var result struct {
		Key   string
		Value map[string]map[string]utils.StringMap
	}
	if err = col.Find(bson.M{"key": dbKey}).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			err = utils.ErrNotFound
		}
		return nil, err
	}
	return result.Value, nil
}

func (ms *MongoStorage) SetReqFilterIndexes(dbKey string, indexes map[string]map[string]utils.StringMap) (err error) {
	session, col := ms.conn(colRFI)
	defer session.Close()
	_, err = col.Upsert(bson.M{"key": dbKey}, &struct {
		Key   string
		Value map[string]map[string]utils.StringMap
	}{dbKey, indexes})
	return
}

func (ms *MongoStorage) MatchReqFilterIndex(dbKey, fldName, fldVal string) (itemIDs utils.StringMap, err error) {
	fieldValKey := utils.ConcatenatedKey(fldName, fldVal)
	cacheKey := dbKey + fieldValKey
	if x, ok := cache.Get(cacheKey); ok { // Attempt to find in cache first
		if x == nil {
			return nil, utils.ErrNotFound
		}
		return x.(utils.StringMap), nil
	}
	session, col := ms.conn(colRFI)
	defer session.Close()
	var result struct {
		Key   string
		Value map[string]map[string]utils.StringMap
	}
	fldKey := fmt.Sprintf("value.%s.%s", fldName, fldVal)
	if err = col.Find(
		bson.M{"key": dbKey, fldKey: bson.M{"$exists": true}}).Select(
		bson.M{fldKey: true}).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			err = utils.ErrNotFound
			cache.Set(cacheKey, nil, true, utils.NonTransactional)
		}
		return nil, err
	}
	itemIDs = result.Value[fldName][fldVal]
	cache.Set(cacheKey, itemIDs, true, utils.NonTransactional)
	return
}

// GetStatQueueProfile retrieves a StatQueueProfile from dataDB
func (ms *MongoStorage) GetStatQueueProfile(tenant string, id string,
	skipCache bool, transactionID string) (sq *StatQueueProfile, err error) {
	key := utils.StatQueueProfilePrefix + utils.ConcatenatedKey(tenant, id)
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*StatQueueProfile), nil
		}
	}
	session, col := ms.conn(colSqp)
	defer session.Close()
	sq = new(StatQueueProfile)
	if err = col.Find(bson.M{"tenant": tenant, "id": id}).One(&sq); err != nil {
		if err == mgo.ErrNotFound {
			err = utils.ErrNotFound
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	for _, fltr := range sq.Filters {
		if err = fltr.CompileValues(); err != nil {
			return
		}
	}
	cache.Set(key, sq, cacheCommit(transactionID), transactionID)
	return
}

// SetStatsQueue stores a StatsQueue into DataDB
func (ms *MongoStorage) SetStatQueueProfile(sq *StatQueueProfile) (err error) {
	session, col := ms.conn(colSqp)
	defer session.Close()
	_, err = col.UpsertId(bson.M{"tennat": sq.Tenant, "id": sq.ID}, sq)
	return
}

// RemStatsQueue removes a StatsQueue from dataDB
func (ms *MongoStorage) RemStatQueueProfile(tenant, id string, transactionID string) (err error) {
	session, col := ms.conn(colSqp)
	err = col.Remove(bson.M{"tenant": tenant, "id": id})
	if err != nil {
		return err
	}
	session.Close()
	cache.RemKey(utils.StatQueueProfilePrefix+utils.ConcatenatedKey(tenant, id),
		cacheCommit(transactionID), transactionID)
	return
}

// GetStoredStatQueue retrieves a StoredStatQueue
func (ms *MongoStorage) GetStoredStatQueue(tenant, id string) (sq *StoredStatQueue, err error) {
	session, col := ms.conn(colSqs)
	defer session.Close()
	if err = col.Find(bson.M{"tenant": tenant, "id": id}).One(&sq); err != nil {
		if err == mgo.ErrNotFound {
			err = utils.ErrNotFound
		}
		return nil, err
	}
	return
}

// SetStoredStatQueue stores the metrics for a StoredStatQueue
func (ms *MongoStorage) SetStoredStatQueue(sq *StoredStatQueue) (err error) {
	session, col := ms.conn(colSqs)
	defer session.Close()
	_, err = col.Upsert(bson.M{"tenant": sq.Tenant, "id": sq.ID}, sq)
	return
}

// RemStatQueue removes stored metrics for a StoredStatQueue
func (ms *MongoStorage) RemStoredStatQueue(tenant, id string) (err error) {
	session, col := ms.conn(colSqs)
	defer session.Close()
	err = col.Remove(bson.M{"tenant": tenant, "id": id})
	if err == mgo.ErrNotFound {
		err = utils.ErrNotFound
	}
	return err
}

// GetThresholdProfile retrieves a ThresholdProfile from dataDB/cache
func (ms *MongoStorage) GetThresholdProfile(tenant, ID string,
	skipCache bool, transactionID string) (tp *ThresholdProfile, err error) {
	cacheKey := utils.ThresholdProfilePrefix + utils.ConcatenatedKey(tenant, ID)
	if !skipCache {
		if x, ok := cache.Get(cacheKey); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*ThresholdProfile), nil
		}
	}
	session, col := ms.conn(colTps)
	defer session.Close()
	tp = new(ThresholdProfile)
	cCommit := cacheCommit(transactionID)
	if err = col.Find(bson.M{"tenant": tenant, "id": ID}).One(&tp); err != nil {
		if err == mgo.ErrNotFound {
			cache.Set(cacheKey, nil, cCommit, transactionID)
			err = utils.ErrNotFound
		}
		return nil, err
	}
	for _, fltr := range tp.Filters {
		if err = fltr.CompileValues(); err != nil {
			return
		}
	}
	cache.Set(cacheKey, tp, cCommit, transactionID)
	return
}

// SetThresholdProfile stores a ThresholdProfile into DataDB
func (ms *MongoStorage) SetThresholdProfile(tp *ThresholdProfile) (err error) {
	session, col := ms.conn(colTps)
	defer session.Close()
	_, err = col.UpsertId(bson.M{"tenant": tp.Tenant, "id": tp.ID}, tp)
	return
}

// RemThresholdProfile removes a ThresholdProfile from dataDB/cache
func (ms *MongoStorage) RemThresholdProfile(tenant, id, transactionID string) (err error) {
	session, col := ms.conn(colTps)
	defer session.Close()
	err = col.Remove(bson.M{"tenant": tenant, "id": id})
	if err != nil {
		return err
	}
	cache.RemKey(utils.ThresholdProfilePrefix+utils.ConcatenatedKey(tenant, id),
		cacheCommit(transactionID), transactionID)
	return
}

func (ms *MongoStorage) GetThreshold(tenant, id string, skipCache bool, transactionID string) (r *Threshold, err error) {
	key := utils.ThresholdPrefix + utils.ConcatenatedKey(tenant, id)
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*Threshold), nil
		}
	}
	session, col := ms.conn(colThs)
	defer session.Close()
	r = new(Threshold)
	if err = col.Find(bson.M{"tenant": tenant, "id": id}).One(r); err != nil {
		if err == mgo.ErrNotFound {
			err = utils.ErrNotFound
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	cache.Set(key, r, cacheCommit(transactionID), transactionID)
	return
}

func (ms *MongoStorage) SetThreshold(r *Threshold) (err error) {
	session, col := ms.conn(colThs)
	defer session.Close()
	_, err = col.Upsert(bson.M{"tenant": r.Tenant, "id": r.ID}, r)
	return
}

func (ms *MongoStorage) RemoveThreshold(tenant, id string, transactionID string) (err error) {
	session, col := ms.conn(colThs)
	defer session.Close()
	if err = col.Remove(bson.M{"tenant": tenant, "id": id}); err != nil {
		return
	}
	cache.RemKey(utils.ThresholdPrefix+utils.ConcatenatedKey(tenant, id),
		cacheCommit(transactionID), transactionID)
	return nil
}

func (ms *MongoStorage) GetFilterDrv(tenant, id string) (r *Filter, err error) {
	session, col := ms.conn(colFlt)
	defer session.Close()
	if err = col.Find(bson.M{"tenant": tenant, "id": id}).One(&r); err != nil {
		if err == mgo.ErrNotFound {
			err = utils.ErrNotFound
		}
		return nil, err
	}
	return
}

func (ms *MongoStorage) SetFilterDrv(r *Filter) (err error) {
	session, col := ms.conn(colFlt)
	defer session.Close()
	_, err = col.Upsert(bson.M{"tenant": r.Tenant, "id": r.ID}, r)
	return
}

func (ms *MongoStorage) RemoveFilterDrv(tenant, id string) (err error) {
	session, col := ms.conn(colFlt)
	defer session.Close()
	if err = col.Remove(bson.M{"tenant": tenant, "id": id}); err != nil {
		return
	}
	return nil
}
