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
	colDst = "destinations"
	colRst = "reverse_destinations"
	colAct = "actions"
	colApl = "action_plans"
	colTsk = "tasks"
	colAtr = "action_triggers"
	colRpl = "rating_plans"
	colRpf = "rating_profiles"
	colAcc = "accounts"
	colShg = "shared_groups"
	colLcr = "lcr_rules"
	colDcs = "derived_chargers"
	colAls = "aliases"
	colRls = "reverse_aliases"
	colStq = "stat_qeues"
	colPbs = "pubsub"
	colUsr = "users"
	colCrs = "cdr_stats"
	colLht = "load_history"
	colVer = "versions"
	colRL  = "resource_limits"
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
	collections := []string{colAct, colApl, colAtr, colDcs, colAls, colRls, colUsr, colLcr, colLht, colRpl, colDst, colRst}
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
	return &MongoStorage{db: db, session: session, ms: NewCodecMsgpackMarshaler(), cacheDumpDir: cacheDumpDir, loadHistorySize: loadHistorySize}, err
}

func (ms *MongoStorage) getColNameForPrefix(prefix string) (name string, ok bool) {
	colMap := map[string]string{
		utils.DESTINATION_PREFIX:         colDst,
		utils.REVERSE_DESTINATION_PREFIX: colRst,
		utils.ACTION_PREFIX:              colAct,
		utils.ACTION_PLAN_PREFIX:         colApl,
		utils.TASKS_KEY:                  colTsk,
		utils.ACTION_TRIGGER_PREFIX:      colAtr,
		utils.RATING_PLAN_PREFIX:         colRpl,
		utils.RATING_PROFILE_PREFIX:      colRpf,
		utils.ACCOUNT_PREFIX:             colAcc,
		utils.SHARED_GROUP_PREFIX:        colShg,
		utils.LCR_PREFIX:                 colLcr,
		utils.DERIVEDCHARGERS_PREFIX:     colDcs,
		utils.ALIASES_PREFIX:             colAls,
		utils.REVERSE_ALIASES_PREFIX:     colRls,
		utils.PUBSUB_SUBSCRIBERS_PREFIX:  colPbs,
		utils.USERS_PREFIX:               colUsr,
		utils.CDR_STATS_PREFIX:           colCrs,
		utils.LOADINST_KEY:               colLht,
		utils.VERSION_PREFIX:             colVer,
		utils.ResourceLimitsPrefix:       colRL,
	}
	name, ok = colMap[prefix]
	return
}

func (ms *MongoStorage) Close() {
	ms.session.Close()
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
		if _, err = db.C(c).RemoveAll(bson.M{}); err != nil {
			return err
		}
	}
	return nil
}

func (ms *MongoStorage) RebuildReverseForPrefix(prefix string) error {
	colName, ok := ms.getColNameForPrefix(prefix)
	if !ok {
		return utils.ErrInvalidKey
	}

	session, col := ms.conn(colName)
	defer session.Close()

	if _, err := col.RemoveAll(bson.M{}); err != nil {
		return err
	}

	switch prefix {
	case utils.REVERSE_DESTINATION_PREFIX:
		keys, err := ms.GetKeysForPrefix(utils.DESTINATION_PREFIX)
		if err != nil {
			return err
		}
		for _, key := range keys {
			dest, err := ms.GetDestination(key[len(utils.DESTINATION_PREFIX):], false)
			if err != nil {
				return err
			}
			if err := ms.SetReverseDestination(dest); err != nil {
				return err
			}
		}
	case utils.REVERSE_ALIASES_PREFIX:
		keys, err := ms.GetKeysForPrefix(utils.ALIASES_PREFIX)
		if err != nil {
			return err
		}
		for _, key := range keys {
			al, err := ms.GetAlias(key[len(utils.ALIASES_PREFIX):], false)
			if err != nil {
				return err
			}
			if err := ms.SetReverseAlias(al); err != nil {
				return err
			}
		}
	default:
		return utils.ErrInvalidKey
	}
	return nil
}

func (ms *MongoStorage) PreloadRatingCache() error {
	err := ms.PreloadCacheForPrefix(utils.RATING_PLAN_PREFIX)
	if err != nil {
		return err
	}
	// add more prefixes if needed
	return nil
}

func (ms *MongoStorage) PreloadAccountingCache() error {
	return nil
}

func (ms *MongoStorage) PreloadCacheForPrefix(prefix string) error {
	cache2go.BeginTransaction()
	cache2go.RemPrefixKey(prefix)
	keyList, err := ms.GetKeysForPrefix(prefix)
	if err != nil {
		cache2go.RollbackTransaction()
		return err
	}
	switch prefix {
	case utils.RATING_PLAN_PREFIX:
		for _, key := range keyList {
			_, err := ms.GetRatingPlan(key[len(utils.RATING_PLAN_PREFIX):], true)
			if err != nil {
				cache2go.RollbackTransaction()
				return err
			}
		}
	default:
		return utils.ErrInvalidKey
	}
	cache2go.CommitTransaction()
	return nil
}

func (ms *MongoStorage) GetKeysForPrefix(prefix string) ([]string, error) {
	var category, subject string
	length := len(utils.DESTINATION_PREFIX)
	if len(prefix) >= length {
		category = prefix[:length] // prefix lenght
		subject = fmt.Sprintf("^%s", prefix[length:])
	} else {
		return nil, fmt.Errorf("unsupported prefix in GetKeysForPrefix: %s", prefix)
	}
	var result []string
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
	case utils.ALIASES_PREFIX:
		iter := db.C(colAls).Find(bson.M{"key": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"key": 1}).Iter()
		for iter.Next(&keyResult) {
			result = append(result, utils.ACTION_PLAN_PREFIX+keyResult.Key)
		}
		return result, nil
	}
	return result, fmt.Errorf("unsupported prefix in GetKeysForPrefix: %s", prefix)
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
		if x, ok := cache2go.Get(utils.RATING_PLAN_PREFIX + key); ok {
			if x != nil {
				return x.(*RatingPlan), nil
			}
			return nil, utils.ErrNotFound
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
	}
	cache2go.Set(utils.RATING_PLAN_PREFIX+key, rp)
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
	cache2go.Set(utils.RATING_PLAN_PREFIX+rp.Id, rp)
	return err
}

func (ms *MongoStorage) GetRatingProfile(key string, skipCache bool) (rp *RatingProfile, err error) {
	if !skipCache {
		if x, ok := cache2go.Get(utils.RATING_PROFILE_PREFIX + key); ok {
			if x != nil {
				return x.(*RatingProfile), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	rp = new(RatingProfile)
	session, col := ms.conn(colRpf)
	defer session.Close()
	err = col.Find(bson.M{"id": key}).One(rp)
	if err == nil {
		cache2go.Set(utils.RATING_PROFILE_PREFIX+key, rp)
	} else {
		cache2go.Set(utils.RATING_PROFILE_PREFIX+key, nil)
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
	cache2go.RemKey(utils.RATING_PROFILE_PREFIX + rp.Id)
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
		cache2go.RemKey(utils.RATING_PROFILE_PREFIX + key)
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
		if x, ok := cache2go.Get(utils.LCR_PREFIX + key); ok {
			if x != nil {
				return x.(*LCR), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	var result struct {
		Key   string
		Value *LCR
	}
	session, col := ms.conn(colLcr)
	defer session.Close()
	if err = col.Find(bson.M{"key": key}).One(&result); err == nil {
		lcr = result.Value
	} else {
		cache2go.Set(utils.LCR_PREFIX+key, nil)
		return nil, utils.ErrNotFound
	}
	cache2go.Set(utils.LCR_PREFIX+key, lcr)
	return
}

func (ms *MongoStorage) SetLCR(lcr *LCR) error {
	session, col := ms.conn(colLcr)
	defer session.Close()
	_, err := col.Upsert(bson.M{"key": lcr.GetId()}, &struct {
		Key   string
		Value *LCR
	}{lcr.GetId(), lcr})
	cache2go.RemKey(utils.LCR_PREFIX + lcr.GetId())
	return err
}

func (ms *MongoStorage) GetDestination(key string, skipCache bool) (result *Destination, err error) {
	if !skipCache {
		if x, ok := cache2go.Get(utils.DESTINATION_PREFIX + key); ok {
			if x != nil {
				return x.(*Destination), nil
			}
			return nil, utils.ErrNotFound
		}
	}
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
	}
	if err != nil {
		result = nil
	}
	cache2go.Set(utils.DESTINATION_PREFIX+key, result)
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
	cache2go.RemKey(utils.DESTINATION_PREFIX + dest.Id)
	if err == nil && historyScribe != nil {
		var response int
		historyScribe.Call("HistoryV1.Record", dest.GetHistoryRecord(false), &response)
	}
	return
}

func (ms *MongoStorage) GetReverseDestination(prefix string, skipCache bool) (ids []string, err error) {
	if !skipCache {
		if x, ok := cache2go.Get(utils.REVERSE_DESTINATION_PREFIX + prefix); ok {
			if x != nil {
				return x.([]string), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	var result struct {
		Key   string
		Value []string
	}
	session, col := ms.conn(colRst)
	defer session.Close()
	err = col.Find(bson.M{"key": prefix}).One(&result)
	if err == nil {
		ids = result.Value
	}
	cache2go.Set(utils.REVERSE_DESTINATION_PREFIX+prefix, ids)
	return
}

func (ms *MongoStorage) SetReverseDestination(dest *Destination) (err error) {
	session, col := ms.conn(colRst)
	defer session.Close()
	for _, p := range dest.Prefixes {
		_, err = col.Upsert(bson.M{"key": p}, bson.M{"$addToSet": bson.M{"value": dest.Id}})
		if err != nil {
			break
		}
		cache2go.RemKey(utils.REVERSE_DESTINATION_PREFIX + p)
	}
	return
}

func (ms *MongoStorage) RemoveDestination(destID string) (err error) {
	session, col := ms.conn(colDst)
	key := utils.DESTINATION_PREFIX + destID
	// get destination for prefix list
	d, err := ms.GetDestination(destID, false)
	if err != nil {
		return
	}
	err = col.Remove(bson.M{"key": key})
	if err != nil {
		return err
	}
	cache2go.RemKey(key)
	session.Close()

	session, col = ms.conn(colRst)
	defer session.Close()
	for _, prefix := range d.Prefixes {
		err = col.Update(bson.M{"key": prefix}, bson.M{"$pull": bson.M{"value": destID}})
		if err != nil {
			return err
		}
		ms.GetReverseDestination(prefix, true) // it will recache the destination
	}
	return
}

func (ms *MongoStorage) UpdateReverseDestination(oldDest, newDest *Destination) error {
	session, col := ms.conn(colRst)
	defer session.Close()
	//log.Printf("Old: %+v, New: %+v", oldDest, newDest)
	var obsoletePrefixes []string
	var addedPrefixes []string
	var found bool
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
	var err error
	for _, obsoletePrefix := range obsoletePrefixes {
		err = col.Update(bson.M{"key": obsoletePrefix}, bson.M{"$pull": bson.M{"value": oldDest.Id}})
		if err != nil {
			return err
		}
		cache2go.RemKey(utils.REVERSE_DESTINATION_PREFIX + obsoletePrefix)
	}

	// add the id to all new prefixes
	for _, addedPrefix := range addedPrefixes {
		_, err = col.Upsert(bson.M{"key": addedPrefix}, bson.M{"$addToSet": bson.M{"value": newDest.Id}})
		if err != nil {
			return err
		}
		cache2go.RemKey(utils.REVERSE_DESTINATION_PREFIX + addedPrefix)
	}
	return nil
}

func (ms *MongoStorage) GetActions(key string, skipCache bool) (as Actions, err error) {
	if !skipCache {
		if x, ok := cache2go.Get(utils.ACTION_PREFIX + key); ok {
			if x != nil {
				return x.(Actions), nil
			}
			return nil, utils.ErrNotFound
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
	}
	cache2go.Set(utils.ACTION_PREFIX+key, as)
	return
}

func (ms *MongoStorage) SetActions(key string, as Actions) error {
	session, col := ms.conn(colAct)
	defer session.Close()
	_, err := col.Upsert(bson.M{"key": key}, &struct {
		Key   string
		Value Actions
	}{Key: key, Value: as})
	cache2go.RemKey(utils.ACTION_PREFIX + key)
	return err
}

func (ms *MongoStorage) RemoveActions(key string) error {
	session, col := ms.conn(colAct)
	defer session.Close()
	return col.Remove(bson.M{"key": key})
}

func (ms *MongoStorage) GetSharedGroup(key string, skipCache bool) (sg *SharedGroup, err error) {
	if !skipCache {
		if x, ok := cache2go.Get(utils.SHARED_GROUP_PREFIX + key); ok {
			if x != nil {
				return x.(*SharedGroup), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	session, col := ms.conn(colShg)
	defer session.Close()
	sg = &SharedGroup{}
	err = col.Find(bson.M{"id": key}).One(sg)
	if err == nil {
		cache2go.Set(utils.SHARED_GROUP_PREFIX+key, sg)
	} else {
		cache2go.Set(utils.SHARED_GROUP_PREFIX+key, nil)
	}
	return
}

func (ms *MongoStorage) SetSharedGroup(sg *SharedGroup) (err error) {
	session, col := ms.conn(colShg)
	defer session.Close()
	_, err = col.Upsert(bson.M{"id": sg.Id}, sg)
	cache2go.RemKey(utils.SHARED_GROUP_PREFIX + sg.Id)
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

func (ms *MongoStorage) GetAlias(key string, skipCache bool) (al *Alias, err error) {
	origKey := key
	key = utils.ALIASES_PREFIX + key
	if !skipCache {
		if x, ok := cache2go.Get(key); ok {
			if x != nil {
				al = &Alias{Values: x.(AliasValues)}
				al.SetId(origKey)
				return al, nil
			}
			return nil, utils.ErrNotFound
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
			cache2go.Set(key, al.Values)
		}
	} else {
		cache2go.Set(key, nil)
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MongoStorage) SetAlias(al *Alias) (err error) {
	session, col := ms.conn(colAls)
	defer session.Close()
	_, err = col.Upsert(bson.M{"key": al.GetId()}, &struct {
		Key   string
		Value AliasValues
	}{Key: al.GetId(), Value: al.Values})
	cache2go.RemKey(utils.ALIASES_PREFIX + al.GetId())
	return err
}

func (ms *MongoStorage) GetReverseAlias(reverseID string, skipCache bool) (ids []string, err error) {
	if !skipCache {
		if x, ok := cache2go.Get(utils.REVERSE_ALIASES_PREFIX + reverseID); ok {
			if x != nil {
				return x.([]string), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	var result struct {
		Key   string
		Value []string
	}
	session, col := ms.conn(colRls)
	defer session.Close()
	if err = col.Find(bson.M{"key": reverseID}).One(&result); err == nil {
		ids = result.Value
		cache2go.Set(utils.REVERSE_ALIASES_PREFIX+reverseID, ids)
	} else {
		cache2go.Set(utils.REVERSE_ALIASES_PREFIX+reverseID, nil)
		return nil, utils.ErrNotFound
	}

	return
}

func (ms *MongoStorage) SetReverseAlias(al *Alias) (err error) {
	session, col := ms.conn(colRls)
	defer session.Close()
	for _, value := range al.Values {
		for target, pairs := range value.Pairs {
			for _, alias := range pairs {
				rKey := strings.Join([]string{alias, target, al.Context}, "")
				id := utils.ConcatenatedKey(al.GetId(), value.DestinationId)
				_, err = col.Upsert(bson.M{"key": rKey}, bson.M{"$addToSet": bson.M{"value": id}})
				if err != nil {
					break
				}
				cache2go.RemKey(rKey)
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
	var kv struct {
		Key   string
		Value AliasValues
	}
	session, col := ms.conn(colAls)
	if err := col.Find(bson.M{"key": origKey}).One(&kv); err == nil {
		al.Values = kv.Value
	}
	err = col.Remove(bson.M{"key": origKey})
	if err != nil {
		return err
	}
	cache2go.RemKey(key)
	session.Close()

	session, col = ms.conn(colRls)
	defer session.Close()
	for _, value := range al.Values {
		tmpKey := utils.ConcatenatedKey(al.GetId(), value.DestinationId)
		for target, pairs := range value.Pairs {
			for _, alias := range pairs {
				rKey := alias + target + al.Context
				err = col.Update(bson.M{"key": rKey}, bson.M{"$pull": bson.M{"value": tmpKey}})
				if err != nil {
					return err
				}
				cache2go.RemKey(utils.REVERSE_ALIASES_PREFIX + rKey)
			}
		}
	}
	return
}

func (ms *MongoStorage) UpdateReverseAlias(oldAl, newAl *Alias) error {
	return nil
}

// Limit will only retrieve the last n items out of history, newest first
func (ms *MongoStorage) GetLoadHistory(limit int, skipCache bool) (loadInsts []*utils.LoadInstance, err error) {
	if limit == 0 {
		return nil, nil
	}
	if !skipCache {
		if x, ok := cache2go.Get(utils.LOADINST_KEY); ok {
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
	if err == nil {
		loadInsts = kv.Value
		cache2go.RemKey(utils.LOADINST_KEY)
		cache2go.Set(utils.LOADINST_KEY, loadInsts)
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

	cache2go.RemKey(utils.LOADINST_KEY)
	return err
}

func (ms *MongoStorage) GetActionTriggers(key string, skipCache bool) (atrs ActionTriggers, err error) {
	if !skipCache {
		if x, ok := cache2go.Get(utils.ACTION_TRIGGER_PREFIX + key); ok {
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
	err = col.Find(bson.M{"key": key}).One(&kv)
	if err == nil {
		atrs = kv.Value
	}
	cache2go.Set(utils.ACTION_TRIGGER_PREFIX+key, atrs)
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
	cache2go.RemKey(utils.ACTION_TRIGGER_PREFIX + key)
	return err
}

func (ms *MongoStorage) RemoveActionTriggers(key string) error {
	session, col := ms.conn(colAtr)
	defer session.Close()
	err := col.Remove(bson.M{"key": key})
	if err == nil {
		cache2go.RemKey(key)
	}
	return err
}

func (ms *MongoStorage) GetActionPlan(key string, skipCache bool) (ats *ActionPlan, err error) {
	if !skipCache {
		if x, ok := cache2go.Get(utils.ACTION_PLAN_PREFIX + key); ok {
			if x != nil {
				return x.(*ActionPlan), nil
			}
			return nil, utils.ErrNotFound
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
	}
	cache2go.Set(utils.ACTION_PLAN_PREFIX+key, ats)
	return
}

func (ms *MongoStorage) SetActionPlan(key string, ats *ActionPlan, overwrite bool) error {
	session, col := ms.conn(colApl)
	defer session.Close()
	// clean dots from account ids map
	if len(ats.ActionTimings) == 0 {
		cache2go.RemKey(utils.ACTION_PLAN_PREFIX + key)
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
	cache2go.RemKey(utils.ACTION_PLAN_PREFIX + key)
	return err
}

func (ms *MongoStorage) GetAllActionPlans() (ats map[string]*ActionPlan, err error) {
	apls := cache2go.GetAllEntries(utils.ACTION_PLAN_PREFIX)

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
		if x, ok := cache2go.Get(utils.DERIVEDCHARGERS_PREFIX + key); ok {
			if x != nil {
				return x.(*utils.DerivedChargers), nil
			}
			return nil, utils.ErrNotFound
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
	} else {
		cache2go.Set(utils.DERIVEDCHARGERS_PREFIX+key, nil)
		return nil, utils.ErrNotFound
	}
	cache2go.Set(utils.DERIVEDCHARGERS_PREFIX+key, dcs)
	return
}

func (ms *MongoStorage) SetDerivedChargers(key string, dcs *utils.DerivedChargers) (err error) {
	if dcs == nil || len(dcs.Chargers) == 0 {
		cache2go.RemKey(utils.DERIVEDCHARGERS_PREFIX + key)
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
	cache2go.RemKey(utils.DERIVEDCHARGERS_PREFIX + key)
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

func (ms *MongoStorage) GetResourceLimit(id string, skipCache bool) (*ResourceLimit, error) {
	return nil, nil
}
func (ms *MongoStorage) SetResourceLimit(rl *ResourceLimit) error {
	return nil
}
func (ms *MongoStorage) RemoveResourceLimit(id string) error {
	return nil
}
