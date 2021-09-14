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
	"context"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

// Mongo collections names
const (
	ColDst  = "destinations"
	ColRds  = "reverse_destinations"
	ColAct  = "actions"
	ColApl  = "action_plans"
	ColAAp  = "account_action_plans"
	ColTsk  = "tasks"
	ColAtr  = "action_triggers"
	ColRpl  = "rating_plans"
	ColRpf  = "rating_profiles"
	ColAcc  = "accounts"
	ColShg  = "shared_groups"
	ColLht  = "load_history"
	ColVer  = "versions"
	ColRsP  = "resource_profiles"
	ColRFI  = "request_filter_indexes"
	ColTmg  = "timings"
	ColRes  = "resources"
	ColSqs  = "statqueues"
	ColSqp  = "statqueue_profiles"
	ColTps  = "threshold_profiles"
	ColThs  = "thresholds"
	ColFlt  = "filters"
	ColSpp  = "supplier_profiles"
	ColAttr = "attribute_profiles"
	ColCDRs = "cdrs"
	ColCpp  = "charger_profiles"
	ColDpp  = "dispatcher_profiles"
	ColDph  = "dispatcher_hosts"
	ColLID  = "load_ids"
)

var (
	CGRIDLow           = strings.ToLower(utils.CGRID)
	RunIDLow           = strings.ToLower(utils.RunID)
	OrderIDLow         = strings.ToLower(utils.OrderID)
	OriginHostLow      = strings.ToLower(utils.OriginHost)
	OriginIDLow        = strings.ToLower(utils.OriginID)
	ToRLow             = strings.ToLower(utils.ToR)
	CDRHostLow         = strings.ToLower(utils.OriginHost)
	CDRSourceLow       = strings.ToLower(utils.Source)
	RequestTypeLow     = strings.ToLower(utils.RequestType)
	TenantLow          = strings.ToLower(utils.Tenant)
	CategoryLow        = strings.ToLower(utils.Category)
	AccountLow         = strings.ToLower(utils.Account)
	SubjectLow         = strings.ToLower(utils.Subject)
	SupplierLow        = strings.ToLower(utils.SUPPLIER)
	DisconnectCauseLow = strings.ToLower(utils.DISCONNECT_CAUSE)
	SetupTimeLow       = strings.ToLower(utils.SetupTime)
	AnswerTimeLow      = strings.ToLower(utils.AnswerTime)
	CreatedAtLow       = strings.ToLower(utils.CreatedAt)
	UpdatedAtLow       = strings.ToLower(utils.UpdatedAt)
	UsageLow           = strings.ToLower(utils.Usage)
	PDDLow             = strings.ToLower(utils.PDD)
	CostDetailsLow     = strings.ToLower(utils.CostDetails)
	DestinationLow     = strings.ToLower(utils.Destination)
	CostLow            = strings.ToLower(utils.COST)
	CostSourceLow      = strings.ToLower(utils.CostSource)

	tTime = reflect.TypeOf(time.Time{})
)

func TimeDecodeValue1(dc bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	if vr.Type() != bsontype.DateTime {
		return fmt.Errorf("cannot decode %v into a time.Time", vr.Type())
	}

	dt, err := vr.ReadDateTime()
	if err != nil {
		return err
	}

	if !val.CanSet() || val.Type() != tTime {
		return bsoncodec.ValueDecoderError{Name: "TimeDecodeValue", Types: []reflect.Type{tTime}, Received: val}
	}
	val.Set(reflect.ValueOf(time.Unix(dt/1000, dt%1000*1000000).UTC()))
	return nil
}

// NewMongoStorage givese new mongo driver
func NewMongoStorage(host, port, db, user, pass, mrshlerStr, storageType string, cdrsIndexes []string,
	isDataDB bool) (ms *MongoStorage, err error) {
	url := host
	if port != "" {
		url += ":" + port
	}
	if user != "" && pass != "" {
		url = fmt.Sprintf("%s:%s@%s", user, pass, url)
	}
	var dbName string
	if db != "" {
		url += "/" + db
		dbName = strings.Split(db, "?")[0] // remove extra info after ?
	}
	ctx := context.Background()
	ttl := config.CgrConfig().DataDbCfg().QueryTimeout
	if !isDataDB {
		ttl = config.CgrConfig().StorDbCfg().QueryTimeout
	}
	url = "mongodb://" + url
	reg := bson.NewRegistryBuilder().RegisterDecoder(tTime, bsoncodec.ValueDecoderFunc(TimeDecodeValue1)).Build()
	opt := options.Client().
		ApplyURI(url).
		SetRegistry(reg).
		SetServerSelectionTimeout(ttl).
		SetRetryWrites(false) // set this option to false because as default it is on true

	client, err := mongo.NewClient(opt)
	// client, err := mongo.NewClient(url)

	if err != nil {
		return nil, err
	}
	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}
	mrshler, err := NewMarshaler(mrshlerStr)
	if err != nil {
		return nil, err
	}

	ms = &MongoStorage{
		client:      client,
		ctx:         ctx,
		ctxTTL:      ttl,
		db:          dbName,
		storageType: storageType,
		ms:          mrshler,
		cdrsIndexes: cdrsIndexes,
		isDataDB:    isDataDB,
	}

	if err = ms.query(func(sctx mongo.SessionContext) error {
		cols, err := ms.client.Database(dbName).ListCollectionNames(sctx, bson.D{})
		if err != nil {
			return err
		}
		empty := true
		for _, col := range cols { // create indexes only if database is empty or only version table is present
			if col != ColVer {
				empty = false
				break
			}
		}
		if empty {
			return ms.EnsureIndexes()
		}
		return nil
	}); err != nil {
		return nil, err
	}
	ms.cnter = utils.NewCounter(time.Now().UnixNano(), 0)
	return
}

// MongoStorage struct for new mongo driver
type MongoStorage struct {
	client      *mongo.Client
	ctx         context.Context
	ctxTTL      time.Duration
	ctxTTLMutex sync.RWMutex // used for TTL reload
	db          string
	storageType string // datadb, stordb
	ms          Marshaler
	cdrsIndexes []string
	cnter       *utils.Counter
	isDataDB    bool
}

func (ms *MongoStorage) query(argfunc func(ctx mongo.SessionContext) error) (err error) {
	ms.ctxTTLMutex.RLock()
	ctxSession, ctxSessionCancel := context.WithTimeout(ms.ctx, ms.ctxTTL)
	ms.ctxTTLMutex.RUnlock()
	defer ctxSessionCancel()
	return ms.client.UseSession(ctxSession, argfunc)
}

// IsDataDB returns if the storeage is used for DataDb
func (ms *MongoStorage) IsDataDB() bool {
	return ms.isDataDB
}

// SetTTL set the context TTL used for queries (is thread safe)
func (ms *MongoStorage) SetTTL(ttl time.Duration) {
	ms.ctxTTLMutex.Lock()
	ms.ctxTTL = ttl
	ms.ctxTTLMutex.Unlock()
}

func (ms *MongoStorage) enusureIndex(colName string, uniq bool, keys ...string) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		col := ms.getCol(colName)
		io := options.Index().SetUnique(uniq)
		doc := make(bson.M)
		for _, k := range keys {
			doc[k] = 1
		}
		_, err := col.Indexes().CreateOne(sctx, mongo.IndexModel{
			Keys:    doc,
			Options: io,
		})
		return err
	})
}

func (ms *MongoStorage) dropAllIndexesForCol(colName string) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		col := ms.getCol(colName)
		_, err := col.Indexes().DropAll(sctx)
		return err
	})
}

func (ms *MongoStorage) getCol(col string) *mongo.Collection {
	return ms.client.Database(ms.db).Collection(col)
}

// GetContext returns the context used for the current DB
func (ms *MongoStorage) GetContext() context.Context {
	return ms.ctx
}

func isNotFound(err error) bool {
	de, ok := err.(mongo.CommandError)
	if !ok { // if still can't converted to the mongo.CommandError check if error do not contains message
		return strings.Contains(err.Error(), "ns not found")
	}
	return de.Code == 26 || de.Message == "ns not found"
}

func (ms *MongoStorage) ensureIndexesForCol(col string) (err error) { // exported for migrator
	if err = ms.dropAllIndexesForCol(col); err != nil && !isNotFound(err) { // make sure you do not have indexes
		return
	}
	err = nil
	switch col {
	case ColAct, ColApl, ColAAp, ColAtr, ColRpl, ColDst, ColRds, ColLht, ColRFI:
		if err = ms.enusureIndex(col, true, "key"); err != nil {
			return
		}
	case ColRsP, ColRes, ColSqs, ColSqp, ColTps, ColThs, ColSpp, ColAttr, ColFlt, ColCpp, ColDpp, ColDph:
		if err = ms.enusureIndex(col, true, "tenant", "id"); err != nil {
			return
		}
	case ColRpf, ColShg, ColAcc:
		if err = ms.enusureIndex(col, true, "id"); err != nil {
			return
		}
		//StorDB
	case utils.TBLTPTimings, utils.TBLTPDestinations,
		utils.TBLTPDestinationRates, utils.TBLTPRatingPlans,
		utils.TBLTPSharedGroups, utils.TBLTPActions,
		utils.TBLTPActionPlans, utils.TBLTPActionTriggers,
		utils.TBLTPStats, utils.TBLTPResources, utils.TBLTPDispatchers,
		utils.TBLTPDispatcherHosts, utils.TBLTPChargers,
		utils.TBLTPSuppliers, utils.TBLTPThresholds:
		if err = ms.enusureIndex(col, true, "tpid", "id"); err != nil {
			return
		}
	case utils.TBLTPRateProfiles:
		if err = ms.enusureIndex(col, true, "tpid", "tenant",
			"category", "subject", "loadid"); err != nil {
			return
		}
	case utils.CDRsTBL:
		if err = ms.enusureIndex(col, true, CGRIDLow, RunIDLow,
			OriginIDLow); err != nil {
			return
		}
		for _, idxKey := range ms.cdrsIndexes {
			if err = ms.enusureIndex(col, false, idxKey); err != nil {
				return
			}
		}
	case utils.SessionCostsTBL:
		if err = ms.enusureIndex(col, true, CGRIDLow,
			RunIDLow); err != nil {
			return
		}
		if err = ms.enusureIndex(col, false, OriginHostLow,
			OriginIDLow); err != nil {
			return
		}
		if err = ms.enusureIndex(col, false, RunIDLow,
			OriginIDLow); err != nil {
			return
		}
	}
	return
}

// EnsureIndexes creates db indexes
func (ms *MongoStorage) EnsureIndexes(cols ...string) (err error) {
	if len(cols) != 0 {
		for _, col := range cols {
			if err = ms.ensureIndexesForCol(col); err != nil {
				return
			}
		}
		return
	}
	if ms.storageType == utils.DataDB {
		for _, col := range []string{ColAct, ColApl, ColAAp, ColAtr,
			ColRpl, ColDst, ColRds, ColLht, ColRFI, ColRsP, ColRes, ColSqs, ColSqp,
			ColTps, ColThs, ColSpp, ColAttr, ColFlt, ColCpp, ColDpp,
			ColRpf, ColShg, ColAcc} {
			if err = ms.ensureIndexesForCol(col); err != nil {
				return
			}
		}
	}
	if ms.storageType == utils.StorDB {
		for _, col := range []string{utils.TBLTPTimings, utils.TBLTPDestinations,
			utils.TBLTPDestinationRates, utils.TBLTPRatingPlans,
			utils.TBLTPSharedGroups, utils.TBLTPActions,
			utils.TBLTPActionPlans, utils.TBLTPActionTriggers,
			utils.TBLTPStats, utils.TBLTPResources,
			utils.TBLTPRateProfiles, utils.CDRsTBL, utils.SessionCostsTBL} {
			if err = ms.ensureIndexesForCol(col); err != nil {
				return
			}
		}
	}
	return
}

func (ms *MongoStorage) getColNameForPrefix(prefix string) (string, bool) {
	res, ok := map[string]string{
		utils.DESTINATION_PREFIX:         ColDst,
		utils.REVERSE_DESTINATION_PREFIX: ColRds,
		utils.ACTION_PREFIX:              ColAct,
		utils.ACTION_PLAN_PREFIX:         ColApl,
		utils.AccountActionPlansPrefix:   ColAAp,
		utils.TASKS_KEY:                  ColTsk,
		utils.ACTION_TRIGGER_PREFIX:      ColAtr,
		utils.RATING_PLAN_PREFIX:         ColRpl,
		utils.RATING_PROFILE_PREFIX:      ColRpf,
		utils.ACCOUNT_PREFIX:             ColAcc,
		utils.SHARED_GROUP_PREFIX:        ColShg,
		utils.LOADINST_KEY:               ColLht,
		utils.VERSION_PREFIX:             ColVer,
		utils.TimingsPrefix:              ColTmg,
		utils.ResourcesPrefix:            ColRes,
		utils.ResourceProfilesPrefix:     ColRsP,
		utils.ThresholdProfilePrefix:     ColTps,
		utils.StatQueueProfilePrefix:     ColSqp,
		utils.ThresholdPrefix:            ColThs,
		utils.FilterPrefix:               ColFlt,
		utils.SupplierProfilePrefix:      ColSpp,
		utils.AttributeProfilePrefix:     ColAttr,
	}[prefix]
	return res, ok
}

// Close disconects the client
func (ms *MongoStorage) Close() {
	if err := ms.client.Disconnect(ms.ctx); err != nil {
		utils.Logger.Err(fmt.Sprintf("<MongoStorage> Error on disconect:%s", err))
	}
}

// Flush drops the datatable
func (ms *MongoStorage) Flush(ignore string) (err error) {
	return ms.query(func(sctx mongo.SessionContext) error {
		if err = ms.client.Database(ms.db).Drop(sctx); err != nil {
			return err
		}
		return ms.EnsureIndexes() // recreate the indexes
	})
}

// DB returnes a database object
func (ms *MongoStorage) DB() *mongo.Database {
	return ms.client.Database(ms.db)
}

// SelectDatabase selects the database
func (ms *MongoStorage) SelectDatabase(dbName string) (err error) {
	ms.db = dbName
	return
}

func (ms *MongoStorage) RemoveKeysForPrefix(prefix string) (err error) {
	var colName string
	switch prefix {
	case utils.DestinationPrefix:
		colName = ColDst
	case utils.REVERSE_DESTINATION_PREFIX:
		colName = ColRds
	case utils.ACTION_PREFIX:
		colName = ColAct
	case utils.ACTION_PLAN_PREFIX:
		colName = ColApl
	case utils.AccountActionPlansPrefix:
		colName = ColAAp
	case utils.TASKS_KEY:
		colName = ColTsk
	case utils.ACTION_TRIGGER_PREFIX:
		colName = ColAtr
	case utils.RATING_PLAN_PREFIX:
		colName = ColRpl
	case utils.RATING_PROFILE_PREFIX:
		colName = ColRpf
	case utils.ACCOUNT_PREFIX:
		colName = ColAcc
	case utils.SHARED_GROUP_PREFIX:
		colName = ColShg
	case utils.LOADINST_KEY:
		colName = ColLht
	case utils.VERSION_PREFIX:
		colName = ColVer
	case utils.TimingsPrefix:
		colName = ColTmg
	case utils.ResourcesPrefix:
		colName = ColRes
	case utils.ResourceProfilesPrefix:
		colName = ColRsP
	case utils.ThresholdProfilePrefix:
		colName = ColTps
	case utils.StatQueueProfilePrefix:
		colName = ColSqp
	case utils.ThresholdPrefix:
		colName = ColThs
	case utils.FilterPrefix:
		colName = ColFlt
	case utils.SupplierProfilePrefix:
		colName = ColSpp
	case utils.AttributeProfilePrefix:
		colName = ColAttr
	default:
		return utils.ErrInvalidKey
	}

	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(colName).DeleteMany(sctx, bson.M{})
		return err
	})
}

// IsDBEmpty implementation
func (ms *MongoStorage) IsDBEmpty() (resp bool, err error) {
	err = ms.query(func(sctx mongo.SessionContext) error {
		cols, err := ms.DB().ListCollectionNames(sctx, bson.D{})
		if err != nil {
			return err
		}
		for _, col := range cols {
			if col == utils.CDRsTBL { // ignore cdrs collection
				continue
			}
			var count int64
			if count, err = ms.getCol(col).CountDocuments(sctx, bson.D{}, options.Count().SetLimit(1)); err != nil { // check if collection is empty so limit the count to 1
				return err
			}
			if count != 0 {
				return nil
			}
		}
		resp = true
		return nil
	})
	return resp, err
}

func (ms *MongoStorage) getField(sctx mongo.SessionContext, col, prefix, subject, field string) (result []string, err error) {
	fieldResult := bson.D{}
	iter, err := ms.getCol(col).Find(sctx,
		bson.M{field: bsonx.Regex(subject, "")},
		options.Find().SetProjection(
			bson.M{field: 1},
		),
	)
	if err != nil {
		return
	}
	for iter.Next(sctx) {
		err = iter.Decode(&fieldResult)
		if err != nil {
			return
		}
		result = append(result, prefix+fieldResult.Map()[field].(string))
	}
	return result, iter.Close(sctx)
}

func (ms *MongoStorage) getField2(sctx mongo.SessionContext, col, prefix, subject string, tntID *utils.TenantID) (result []string, err error) {
	idResult := struct{ Tenant, Id string }{}
	elem := bson.M{}
	if tntID.Tenant != "" {
		elem["tenant"] = tntID.Tenant
	}
	if tntID.ID != "" {
		elem["id"] = bsonx.Regex(subject, "")
	}
	iter, err := ms.getCol(col).Find(sctx, elem,
		options.Find().SetProjection(bson.M{"tenant": 1, "id": 1}),
	)
	if err != nil {
		return
	}
	for iter.Next(sctx) {
		err = iter.Decode(&idResult)
		if err != nil {
			return
		}
		result = append(result, prefix+utils.ConcatenatedKey(idResult.Tenant, idResult.Id))
	}
	return result, iter.Close(sctx)
}

func (ms *MongoStorage) getField3(sctx mongo.SessionContext, col, prefix, field string) (result []string, err error) {
	fieldResult := bson.D{}
	iter, err := ms.getCol(col).Find(sctx,
		bson.M{field: bsonx.Regex(fmt.Sprintf("^%s", prefix), "")},
		options.Find().SetProjection(
			bson.M{field: 1},
		),
	)
	if err != nil {
		return
	}
	for iter.Next(sctx) {
		err = iter.Decode(&fieldResult)
		if err != nil {
			return
		}
		result = append(result, fieldResult.Map()[field].(string))
	}
	return result, iter.Close(sctx)
}

// GetKeysForPrefix implementation
func (ms *MongoStorage) GetKeysForPrefix(prefix string) (result []string, err error) {
	var category, subject string
	keyLen := len(utils.DESTINATION_PREFIX)
	if len(prefix) < keyLen {
		return nil, fmt.Errorf("unsupported prefix in GetKeysForPrefix: %s", prefix)
	}
	category = prefix[:keyLen] // prefix length
	tntID := utils.NewTenantID(prefix[keyLen:])
	subject = fmt.Sprintf("^%s", prefix[keyLen:]) // old way, no tenant support
	err = ms.query(func(sctx mongo.SessionContext) (err error) {
		switch category {
		case utils.DESTINATION_PREFIX:
			result, err = ms.getField(sctx, ColDst, utils.DESTINATION_PREFIX, subject, "key")
		case utils.REVERSE_DESTINATION_PREFIX:
			result, err = ms.getField(sctx, ColRds, utils.REVERSE_DESTINATION_PREFIX, subject, "key")
		case utils.RATING_PLAN_PREFIX:
			result, err = ms.getField(sctx, ColRpl, utils.RATING_PLAN_PREFIX, subject, "key")
		case utils.RATING_PROFILE_PREFIX:
			if strings.HasPrefix(prefix[keyLen:], utils.META_OUT) {
				subject = fmt.Sprintf("^\\%s", prefix[keyLen:]) // rewrite the id cause it start with * from `*out`
			}
			result, err = ms.getField(sctx, ColRpf, utils.RATING_PROFILE_PREFIX, subject, "id")
		case utils.ACTION_PREFIX:
			result, err = ms.getField(sctx, ColAct, utils.ACTION_PREFIX, subject, "key")
		case utils.ACTION_PLAN_PREFIX:
			result, err = ms.getField(sctx, ColApl, utils.ACTION_PLAN_PREFIX, subject, "key")
		case utils.ACTION_TRIGGER_PREFIX:
			result, err = ms.getField(sctx, ColAtr, utils.ACTION_TRIGGER_PREFIX, subject, "key")
		case utils.SHARED_GROUP_PREFIX:
			result, err = ms.getField(sctx, ColShg, utils.SHARED_GROUP_PREFIX, subject, "id")
		case utils.ACCOUNT_PREFIX:
			result, err = ms.getField(sctx, ColAcc, utils.ACCOUNT_PREFIX, subject, "id")
		case utils.ResourceProfilesPrefix:
			result, err = ms.getField2(sctx, ColRsP, utils.ResourceProfilesPrefix, subject, tntID)
		case utils.ResourcesPrefix:
			result, err = ms.getField2(sctx, ColRes, utils.ResourcesPrefix, subject, tntID)
		case utils.StatQueuePrefix:
			result, err = ms.getField2(sctx, ColSqs, utils.StatQueuePrefix, subject, tntID)
		case utils.StatQueueProfilePrefix:
			result, err = ms.getField2(sctx, ColSqp, utils.StatQueueProfilePrefix, subject, tntID)
		case utils.AccountActionPlansPrefix:
			result, err = ms.getField(sctx, ColAAp, utils.AccountActionPlansPrefix, subject, "key")
		case utils.TimingsPrefix:
			result, err = ms.getField(sctx, ColTmg, utils.TimingsPrefix, subject, "id")
		case utils.FilterPrefix:
			result, err = ms.getField2(sctx, ColFlt, utils.FilterPrefix, subject, tntID)
		case utils.ThresholdPrefix:
			result, err = ms.getField2(sctx, ColThs, utils.ThresholdPrefix, subject, tntID)
		case utils.ThresholdProfilePrefix:
			result, err = ms.getField2(sctx, ColTps, utils.ThresholdProfilePrefix, subject, tntID)
		case utils.SupplierProfilePrefix:
			result, err = ms.getField2(sctx, ColSpp, utils.SupplierProfilePrefix, subject, tntID)
		case utils.AttributeProfilePrefix:
			result, err = ms.getField2(sctx, ColAttr, utils.AttributeProfilePrefix, subject, tntID)
		case utils.ChargerProfilePrefix:
			result, err = ms.getField2(sctx, ColCpp, utils.ChargerProfilePrefix, subject, tntID)
		case utils.DispatcherProfilePrefix:
			result, err = ms.getField2(sctx, ColDpp, utils.DispatcherProfilePrefix, subject, tntID)
		case utils.DispatcherHostPrefix:
			result, err = ms.getField2(sctx, ColDph, utils.DispatcherHostPrefix, subject, tntID)
		case utils.AttributeFilterIndexes:
			result, err = ms.getField3(sctx, ColRFI, utils.AttributeFilterIndexes, "key")
		case utils.ResourceFilterIndexes:
			result, err = ms.getField3(sctx, ColRFI, utils.ResourceFilterIndexes, "key")
		case utils.StatFilterIndexes:
			result, err = ms.getField3(sctx, ColRFI, utils.StatFilterIndexes, "key")
		case utils.ThresholdFilterIndexes:
			result, err = ms.getField3(sctx, ColRFI, utils.ThresholdFilterIndexes, "key")
		case utils.SupplierFilterIndexes:
			result, err = ms.getField3(sctx, ColRFI, utils.SupplierFilterIndexes, "key")
		case utils.ChargerFilterIndexes:
			result, err = ms.getField3(sctx, ColRFI, utils.ChargerFilterIndexes, "key")
		case utils.DispatcherFilterIndexes:
			result, err = ms.getField3(sctx, ColRFI, utils.DispatcherFilterIndexes, "key")
		default:
			err = fmt.Errorf("unsupported prefix in GetKeysForPrefix: %s", prefix)
		}
		return err
	})
	return
}

func (ms *MongoStorage) HasDataDrv(category, subject, tenant string) (has bool, err error) {
	err = ms.query(func(sctx mongo.SessionContext) (err error) {
		var count int64
		switch category {
		case utils.DESTINATION_PREFIX:
			count, err = ms.getCol(ColDst).CountDocuments(sctx, bson.M{"key": subject})
		case utils.RATING_PLAN_PREFIX:
			count, err = ms.getCol(ColRpl).CountDocuments(sctx, bson.M{"key": subject})
		case utils.RATING_PROFILE_PREFIX:
			count, err = ms.getCol(ColRpf).CountDocuments(sctx, bson.M{"key": subject})
		case utils.ACTION_PREFIX:
			count, err = ms.getCol(ColAct).CountDocuments(sctx, bson.M{"key": subject})
		case utils.ACTION_PLAN_PREFIX:
			count, err = ms.getCol(ColApl).CountDocuments(sctx, bson.M{"key": subject})
		case utils.ACCOUNT_PREFIX:
			count, err = ms.getCol(ColAcc).CountDocuments(sctx, bson.M{"id": subject})
		case utils.ResourcesPrefix:
			count, err = ms.getCol(ColRes).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.ResourceProfilesPrefix:
			count, err = ms.getCol(ColRsP).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.StatQueuePrefix:
			count, err = ms.getCol(ColSqs).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.StatQueueProfilePrefix:
			count, err = ms.getCol(ColSqp).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.ThresholdPrefix:
			count, err = ms.getCol(ColThs).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.ThresholdProfilePrefix:
			count, err = ms.getCol(ColTps).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.FilterPrefix:
			count, err = ms.getCol(ColFlt).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.SupplierProfilePrefix:
			count, err = ms.getCol(ColSpp).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.AttributeProfilePrefix:
			count, err = ms.getCol(ColAttr).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.ChargerProfilePrefix:
			count, err = ms.getCol(ColCpp).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.DispatcherProfilePrefix:
			count, err = ms.getCol(ColDpp).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.DispatcherHostPrefix:
			count, err = ms.getCol(ColDph).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		default:
			err = fmt.Errorf("unsupported category in HasData: %s", category)
		}
		has = count > 0
		return err
	})
	return has, err
}

func (ms *MongoStorage) GetRatingPlanDrv(key string) (rp *RatingPlan, err error) {
	var kv struct {
		Key   string
		Value []byte
	}
	if err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColRpl).FindOne(sctx, bson.M{"key": key})
		if err := cur.Decode(&kv); err != nil {
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	}); err != nil {
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
	return
}

func (ms *MongoStorage) SetRatingPlanDrv(rp *RatingPlan) error {
	result, err := ms.ms.Marshal(rp)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColRpl).UpdateOne(sctx, bson.M{"key": rp.Id},
			bson.M{"$set": struct {
				Key   string
				Value []byte
			}{Key: rp.Id, Value: b.Bytes()}},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveRatingPlanDrv(key string) error {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(ColRpl).DeleteMany(sctx, bson.M{"key": key})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetRatingProfileDrv(key string) (rp *RatingProfile, err error) {
	rp = new(RatingProfile)
	err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColRpf).FindOne(sctx, bson.M{"id": key})
		if err := cur.Decode(rp); err != nil {
			rp = nil
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	return
}

func (ms *MongoStorage) SetRatingProfileDrv(rp *RatingProfile) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColRpf).UpdateOne(sctx, bson.M{"id": rp.Id},
			bson.M{"$set": rp},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveRatingProfileDrv(key string) error {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(ColRpf).DeleteMany(sctx, bson.M{"id": key})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetDestinationDrv(key string, skipCache bool,
	transactionID string) (result *Destination, err error) {
	if !skipCache {
		if x, ok := Cache.Get(utils.CacheDestinations, key); ok {
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
	if err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColDst).FindOne(sctx, bson.M{"key": key})
		if err := cur.Decode(&kv); err != nil {
			if err == mongo.ErrNoDocuments {
				Cache.Set(utils.CacheDestinations, key, nil, nil,
					cacheCommit(transactionID), transactionID)
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	}); err != nil {
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
	err = ms.ms.Unmarshal(out, &result)
	if err != nil {
		return nil, err
	}
	Cache.Set(utils.CacheDestinations, key, result, nil,
		cacheCommit(transactionID), transactionID)
	return
}

func (ms *MongoStorage) SetDestinationDrv(dest *Destination, transactionID string) (err error) {
	result, err := ms.ms.Marshal(dest)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColDst).UpdateOne(sctx, bson.M{"key": dest.Id},
			bson.M{"$set": struct {
				Key   string
				Value []byte
			}{Key: dest.Id, Value: b.Bytes()}},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveDestinationDrv(destID string,
	transactionID string) (err error) {
	// get destination for prefix list
	d, err := ms.GetDestinationDrv(destID, false, transactionID)
	if err != nil {
		return
	}
	if err = ms.query(func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(ColDst).DeleteOne(sctx, bson.M{"key": destID})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	}); err != nil {
		return err
	}
	Cache.Remove(utils.CacheDestinations, destID,
		cacheCommit(transactionID), transactionID)

	for _, prefix := range d.Prefixes {
		if err = ms.query(func(sctx mongo.SessionContext) (err error) {
			_, err = ms.getCol(ColRds).UpdateOne(sctx, bson.M{"key": prefix},
				bson.M{"$pull": bson.M{"value": destID}})
			return err
		}); err != nil {
			return err
		}
		ms.GetReverseDestinationDrv(prefix, true, transactionID) // it will recache the destination
	}
	return
}

func (ms *MongoStorage) GetReverseDestinationDrv(prefix string, skipCache bool,
	transactionID string) (ids []string, err error) {
	if !skipCache {
		if x, ok := Cache.Get(utils.CacheReverseDestinations, prefix); ok {
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
	if err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColRds).FindOne(sctx, bson.M{"key": prefix})
		if err := cur.Decode(&result); err != nil {
			if err == mongo.ErrNoDocuments {
				Cache.Set(utils.CacheReverseDestinations, prefix, nil, nil,
					cacheCommit(transactionID), transactionID)
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	ids = result.Value
	Cache.Set(utils.CacheReverseDestinations, prefix, ids, nil,
		cacheCommit(transactionID), transactionID)
	return
}

func (ms *MongoStorage) SetReverseDestinationDrv(dest *Destination,
	transactionID string) (err error) {
	for _, p := range dest.Prefixes {
		if err = ms.query(func(sctx mongo.SessionContext) (err error) {
			_, err = ms.getCol(ColRds).UpdateOne(sctx, bson.M{"key": p},
				bson.M{"$addToSet": bson.M{"value": dest.Id}},
				options.Update().SetUpsert(true),
			)
			return err
		}); err != nil {

			return err
		}
	}
	return nil
}

func (ms *MongoStorage) UpdateReverseDestinationDrv(oldDest, newDest *Destination,
	transactionID string) error {
	//log.Printf("Old: %+v, New: %+v", oldDest, newDest)
	var obsoletePrefixes []string
	var addedPrefixes []string
	if oldDest == nil {
		oldDest = new(Destination) // so we can process prefixes
	}
	for _, oldPrefix := range oldDest.Prefixes {
		found := false
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
		found := false
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
		if err = ms.query(func(sctx mongo.SessionContext) (err error) {
			_, err = ms.getCol(ColRds).UpdateOne(sctx, bson.M{"key": obsoletePrefix},
				bson.M{"$pull": bson.M{"value": oldDest.Id}})
			return err
		}); err != nil {
			return err
		}
		Cache.Remove(utils.CacheReverseDestinations, obsoletePrefix,
			cCommit, transactionID)
	}

	// add the id to all new prefixes
	for _, addedPrefix := range addedPrefixes {
		if err = ms.query(func(sctx mongo.SessionContext) (err error) {
			_, err = ms.getCol(ColRds).UpdateOne(sctx, bson.M{"key": addedPrefix},
				bson.M{"$addToSet": bson.M{"value": newDest.Id}},
				options.Update().SetUpsert(true),
			)
			return err
		}); err != nil {
			return err
		}
	}
	return nil
}

func (ms *MongoStorage) GetActionsDrv(key string) (as Actions, err error) {
	var result struct {
		Key   string
		Value Actions
	}
	if err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColAct).FindOne(sctx, bson.M{"key": key})
		if err := cur.Decode(&result); err != nil {
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	as = result.Value
	return
}

func (ms *MongoStorage) SetActionsDrv(key string, as Actions) error {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColAct).UpdateOne(sctx, bson.M{"key": key},
			bson.M{"$set": struct {
				Key   string
				Value Actions
			}{Key: key, Value: as}},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveActionsDrv(key string) error {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(ColAct).DeleteOne(sctx, bson.M{"key": key})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetSharedGroupDrv(key string) (sg *SharedGroup, err error) {
	sg = new(SharedGroup)
	err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColShg).FindOne(sctx, bson.M{"id": key})
		if err := cur.Decode(sg); err != nil {
			sg = nil
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	return
}

func (ms *MongoStorage) SetSharedGroupDrv(sg *SharedGroup) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColShg).UpdateOne(sctx, bson.M{"id": sg.Id},
			bson.M{"$set": sg},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveSharedGroupDrv(id string) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(ColShg).DeleteOne(sctx, bson.M{"id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetAccountDrv(key string) (result *Account, err error) {
	result = new(Account)
	err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColAcc).FindOne(sctx, bson.M{"id": key})
		if err := cur.Decode(result); err != nil {
			result = nil
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	return
}

func (ms *MongoStorage) SetAccountDrv(acc *Account) error {
	// never override existing account with an empty one
	// UPDATE: if all balances expired and were cleaned it makes
	// sense to write empty balance map
	if len(acc.BalanceMap) == 0 {
		if ac, err := ms.GetAccountDrv(acc.ID); err == nil && !ac.allBalancesExpired() {
			ac.ActionTriggers = acc.ActionTriggers
			ac.UnitCounters = acc.UnitCounters
			ac.AllowNegative = acc.AllowNegative
			ac.Disabled = acc.Disabled
			acc = ac
		}
	}
	acc.UpdateTime = time.Now()
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColAcc).UpdateOne(sctx, bson.M{"id": acc.ID},
			bson.M{"$set": acc},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveAccountDrv(key string) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(ColAcc).DeleteOne(sctx, bson.M{"id": key})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

// Limit will only retrieve the last n items out of history, newest first
func (ms *MongoStorage) GetLoadHistory(limit int, skipCache bool,
	transactionID string) (loadInsts []*utils.LoadInstance, err error) {
	if limit == 0 {
		return nil, nil
	}
	if !skipCache {
		if x, ok := Cache.Get(utils.LOADINST_KEY, ""); ok {
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
	err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColLht).FindOne(sctx, bson.M{"key": utils.LOADINST_KEY})
		if err := cur.Decode(&kv); err != nil {
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	cCommit := cacheCommit(transactionID)
	if err == nil {
		loadInsts = kv.Value
		Cache.Remove(utils.LOADINST_KEY, "", cCommit, transactionID)
		Cache.Set(utils.LOADINST_KEY, "", loadInsts, nil, cCommit, transactionID)
	}
	if len(loadInsts) < limit || limit == -1 {
		return loadInsts, nil
	}
	return loadInsts[:limit], nil
}

// Adds a single load instance to load history
func (ms *MongoStorage) AddLoadHistory(ldInst *utils.LoadInstance,
	loadHistSize int, transactionID string) error {
	if loadHistSize == 0 { // Load history disabled
		return nil
	}
	// get existing load history
	var existingLoadHistory []*utils.LoadInstance
	var kv struct {
		Key   string
		Value []*utils.LoadInstance
	}
	if err := ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColLht).FindOne(sctx, bson.M{"key": utils.LOADINST_KEY})
		if err := cur.Decode(&kv); err != nil {
			if err == mongo.ErrNoDocuments {
				return nil // utils.ErrNotFound
			}
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	if kv.Value != nil {
		existingLoadHistory = kv.Value
	}
	_, err := guardian.Guardian.Guard(func() (interface{}, error) { // Make sure we do it locked since other instance can modify history while we read it
		// insert on first position
		existingLoadHistory = append(existingLoadHistory, nil)
		copy(existingLoadHistory[1:], existingLoadHistory[0:])
		existingLoadHistory[0] = ldInst

		//check length
		histLen := len(existingLoadHistory)
		if histLen >= loadHistSize { // Have hit maximum history allowed, remove oldest element in order to add new one
			existingLoadHistory = existingLoadHistory[:loadHistSize]
		}
		return nil, ms.query(func(sctx mongo.SessionContext) (err error) {
			_, err = ms.getCol(ColLht).UpdateOne(sctx, bson.M{"key": utils.LOADINST_KEY},
				bson.M{"$set": struct {
					Key   string
					Value []*utils.LoadInstance
				}{Key: utils.LOADINST_KEY, Value: existingLoadHistory}},
				options.Update().SetUpsert(true),
			)
			return err
		})
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.LOADINST_KEY)

	Cache.Remove(utils.LOADINST_KEY, "",
		cacheCommit(transactionID), transactionID)
	return err
}

func (ms *MongoStorage) GetActionTriggersDrv(key string) (atrs ActionTriggers, err error) {
	var kv struct {
		Key   string
		Value ActionTriggers
	}
	if err := ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColAtr).FindOne(sctx, bson.M{"key": key})
		if err := cur.Decode(&kv); err != nil {
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	atrs = kv.Value
	return
}

func (ms *MongoStorage) SetActionTriggersDrv(key string, atrs ActionTriggers) (err error) {
	if len(atrs) == 0 {
		return ms.query(func(sctx mongo.SessionContext) (err error) {
			_, err = ms.getCol(ColAtr).DeleteOne(sctx, bson.M{"key": key})
			return err
		})
	}
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColAtr).UpdateOne(sctx, bson.M{"key": key},
			bson.M{"$set": struct {
				Key   string
				Value ActionTriggers
			}{Key: key, Value: atrs}},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveActionTriggersDrv(key string) error {

	return ms.query(func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(ColAtr).DeleteOne(sctx, bson.M{"key": key})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetActionPlanDrv(key string) (ats *ActionPlan, err error) {
	var kv struct {
		Key   string
		Value []byte
	}
	if err := ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColApl).FindOne(sctx, bson.M{"key": key})
		if err := cur.Decode(&kv); err != nil {
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	}); err != nil {
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
	err = ms.ms.Unmarshal(out, &ats)
	return
}

func (ms *MongoStorage) SetActionPlanDrv(key string, ats *ActionPlan) (err error) {
	result, err := ms.ms.Marshal(ats)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColApl).UpdateOne(sctx, bson.M{"key": key},
			bson.M{"$set": struct {
				Key   string
				Value []byte
			}{Key: key, Value: b.Bytes()}},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveActionPlanDrv(key string) error {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColApl).DeleteOne(sctx, bson.M{"key": key})
		return err
	})
}

func (ms *MongoStorage) GetAllActionPlansDrv() (ats map[string]*ActionPlan, err error) {
	keys, err := ms.GetKeysForPrefix(utils.ACTION_PLAN_PREFIX)
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, utils.ErrNotFound
	}
	ats = make(map[string]*ActionPlan, len(keys))
	for _, key := range keys {
		ap, err := ms.GetActionPlanDrv(key[len(utils.ACTION_PLAN_PREFIX):])
		if err != nil {
			return nil, err
		}
		ats[key[len(utils.ACTION_PLAN_PREFIX):]] = ap
	}
	return
}

func (ms *MongoStorage) GetAccountActionPlansDrv(acntID string) (aPlIDs []string, err error) {
	var kv struct {
		Key   string
		Value []string
	}
	if err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColAAp).FindOne(sctx, bson.M{"key": acntID})
		if err := cur.Decode(&kv); err != nil {
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	aPlIDs = kv.Value
	return
}

func (ms *MongoStorage) SetAccountActionPlansDrv(acntID string, aPlIDs []string) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColAAp).UpdateOne(sctx, bson.M{"key": acntID},
			bson.M{"$set": struct {
				Key   string
				Value []string
			}{Key: acntID, Value: aPlIDs}},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

// ToDo: check return len(aPlIDs) == 0
func (ms *MongoStorage) RemAccountActionPlansDrv(acntID string) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(ColAAp).DeleteOne(sctx, bson.M{"key": acntID})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) PushTask(t *Task) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColTsk).InsertOne(sctx, bson.M{"_id": primitive.NewObjectID(), "task": t})
		return err
	})
}

func (ms *MongoStorage) PopTask() (t *Task, err error) {
	v := struct {
		ID   primitive.ObjectID `bson:"_id"`
		Task *Task
	}{}
	if err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColTsk).FindOneAndDelete(sctx, bson.D{})
		if err := cur.Decode(&v); err != nil {
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return v.Task, nil
}

func (ms *MongoStorage) GetResourceProfileDrv(tenant, id string) (rp *ResourceProfile, err error) {
	rp = new(ResourceProfile)
	err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColRsP).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		if err := cur.Decode(rp); err != nil {
			rp = nil
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	return
}

func (ms *MongoStorage) SetResourceProfileDrv(rp *ResourceProfile) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColRsP).UpdateOne(sctx, bson.M{"tenant": rp.Tenant, "id": rp.ID},
			bson.M{"$set": rp},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveResourceProfileDrv(tenant, id string) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(ColRsP).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetResourceDrv(tenant, id string) (r *Resource, err error) {
	r = new(Resource)
	err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColRes).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		if err := cur.Decode(r); err != nil {
			r = nil
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	return
}

func (ms *MongoStorage) SetResourceDrv(r *Resource) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColRes).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveResourceDrv(tenant, id string) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(ColRes).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetTimingDrv(id string) (t *utils.TPTiming, err error) {
	t = new(utils.TPTiming)
	err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColTmg).FindOne(sctx, bson.M{"id": id})
		if err := cur.Decode(t); err != nil {
			t = nil
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	return
}

func (ms *MongoStorage) SetTimingDrv(t *utils.TPTiming) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColTmg).UpdateOne(sctx, bson.M{"id": t.ID},
			bson.M{"$set": t},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveTimingDrv(id string) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(ColTmg).DeleteOne(sctx, bson.M{"id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

// GetFilterIndexesDrv retrieves Indexes from dataDB
//filterType is used togheter with fieldName:Val
func (ms *MongoStorage) GetFilterIndexesDrv(cacheID, itemIDPrefix, filterType string,
	fldNameVal map[string]string) (indexes map[string]utils.StringMap, err error) {
	type result struct {
		Key   string
		Value []string
	}
	var results []result
	dbKey := utils.CacheInstanceToPrefix[cacheID] + itemIDPrefix
	if len(fldNameVal) != 0 {
		for fldName, fldValue := range fldNameVal {
			if err = ms.query(func(sctx mongo.SessionContext) (err error) {
				cur, err := ms.getCol(ColRFI).Find(sctx, bson.M{"key": utils.ConcatenatedKey(dbKey, filterType, fldName, fldValue)})
				if err != nil {
					return err
				}
				for cur.Next(sctx) {
					var elem result
					if err := cur.Decode(&elem); err != nil {
						return err
					}
					results = append(results, elem)
				}
				return cur.Close(sctx)
			}); err != nil {
				return nil, err
			}
			if len(results) == 0 {
				return nil, utils.ErrNotFound
			}
		}
	} else {
		for _, character := range []string{".", "*"} {
			dbKey = strings.Replace(dbKey, character, `\`+character, strings.Count(dbKey, character))
		}
		//inside bson.RegEx add carrot to match the prefix (optimization)
		if err = ms.query(func(sctx mongo.SessionContext) (err error) {
			cur, err := ms.getCol(ColRFI).Find(sctx, bson.M{"key": bsonx.Regex("^"+dbKey, "")})
			if err != nil {
				return err
			}
			for cur.Next(sctx) {
				var elem result
				if err := cur.Decode(&elem); err != nil {
					return err
				}
				results = append(results, elem)
			}
			return cur.Close(sctx)
		}); err != nil {
			return nil, err
		}
		if len(results) == 0 {
			return nil, utils.ErrNotFound
		}
	}
	indexes = make(map[string]utils.StringMap)

	for _, res := range results {
		if len(res.Value) == 0 {
			continue
		}
		keys := strings.Split(res.Key, ":")
		indexKey := utils.ConcatenatedKey(keys[1], keys[2], keys[3])
		//check here if itemIDPrefix has context
		if len(strings.Split(itemIDPrefix, ":")) == 2 {
			indexKey = utils.ConcatenatedKey(keys[2], keys[3], keys[4])
		}
		indexes[indexKey] = utils.StringMapFromSlice(res.Value)
	}
	if len(indexes) == 0 {
		return nil, utils.ErrNotFound
	}
	return indexes, nil
}

// SetFilterIndexesDrv stores Indexes into DataDB
func (ms *MongoStorage) SetFilterIndexesDrv(cacheID, itemIDPrefix string,
	indexes map[string]utils.StringMap, commit bool, transactionID string) (err error) {
	originKey := utils.CacheInstanceToPrefix[cacheID] + itemIDPrefix
	dbKey := originKey
	if transactionID != "" {
		dbKey = "tmp_" + utils.ConcatenatedKey(originKey, transactionID)
	}
	if commit && transactionID != "" {
		regexKey := originKey
		for _, character := range []string{".", "*"} {
			regexKey = strings.Replace(regexKey, character, `\`+character, strings.Count(regexKey, character))
		}
		//inside bson.RegEx add carrot to match the prefix (optimization)
		if err = ms.query(func(sctx mongo.SessionContext) (err error) {
			_, err = ms.getCol(ColRFI).DeleteMany(sctx, bson.M{"key": bsonx.Regex("^"+regexKey, "")})
			return err
		}); err != nil {
			return err
		}
		var lastErr error
		for key, itmMp := range indexes {
			if err = ms.query(func(sctx mongo.SessionContext) (err error) {
				_, err = ms.getCol(ColRFI).UpdateOne(sctx, bson.M{"key": utils.ConcatenatedKey(originKey, key)},
					bson.M{"$set": bson.M{"key": utils.ConcatenatedKey(originKey, key), "value": itmMp.Slice()}},
					options.Update().SetUpsert(true),
				)
				return err
			}); err != nil {
				lastErr = err
			}
		}
		if lastErr != nil {
			return lastErr
		}
		oldKey := "tmp_" + utils.ConcatenatedKey(originKey, transactionID)
		for _, character := range []string{".", "*"} {
			oldKey = strings.Replace(oldKey, character, `\`+character, strings.Count(oldKey, character))
		}
		//inside bson.RegEx add carrot to match the prefix (optimization)
		return ms.query(func(sctx mongo.SessionContext) (err error) {
			_, err = ms.getCol(ColRFI).DeleteMany(sctx, bson.M{"key": bsonx.Regex("^"+oldKey, "")})
			return err
		})
	} else {
		var lastErr error
		for key, itmMp := range indexes {
			if err = ms.query(func(sctx mongo.SessionContext) (err error) {
				idxDbkey := utils.ConcatenatedKey(dbKey, key)
				if len(itmMp) == 0 { // remove from DB if we set it with empty indexes
					_, err = ms.getCol(ColRFI).DeleteOne(sctx,
						bson.M{"key": idxDbkey})
				} else {
					_, err = ms.getCol(ColRFI).UpdateOne(sctx, bson.M{"key": idxDbkey},
						bson.M{"$set": bson.M{"key": idxDbkey, "value": itmMp.Slice()}},
						options.Update().SetUpsert(true),
					)
				}
				return err
			}); err != nil {
				lastErr = err
			}
		}
		return lastErr
	}
}

func (ms *MongoStorage) RemoveFilterIndexesDrv(cacheID, itemIDPrefix string) (err error) {
	regexKey := utils.CacheInstanceToPrefix[cacheID] + itemIDPrefix
	for _, character := range []string{".", "*"} {
		regexKey = strings.Replace(regexKey, character, `\`+character, strings.Count(regexKey, character))
	}
	//inside bson.RegEx add carrot to match the prefix (optimization)
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColRFI).DeleteMany(sctx, bson.M{"key": bsonx.Regex("^"+regexKey, "")})
		return err
	})
}

func (ms *MongoStorage) MatchFilterIndexDrv(cacheID, itemIDPrefix,
	filterType, fldName, fldVal string) (itemIDs utils.StringMap, err error) {
	var result struct {
		Key   string
		Value []string
	}
	dbKey := utils.CacheInstanceToPrefix[cacheID] + itemIDPrefix
	if err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColRFI).FindOne(sctx, bson.M{"key": utils.ConcatenatedKey(dbKey, filterType, fldName, fldVal)})
		if err := cur.Decode(&result); err != nil {
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return utils.StringMapFromSlice(result.Value), nil
}

// GetStatQueueProfileDrv retrieves a StatQueueProfile from dataDB
func (ms *MongoStorage) GetStatQueueProfileDrv(tenant string, id string) (sq *StatQueueProfile, err error) {
	sq = new(StatQueueProfile)
	err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColSqp).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		if err := cur.Decode(sq); err != nil {
			sq = nil
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	return
}

// SetStatQueueProfileDrv stores a StatsQueue into DataDB
func (ms *MongoStorage) SetStatQueueProfileDrv(sq *StatQueueProfile) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColSqp).UpdateOne(sctx, bson.M{"tenant": sq.Tenant, "id": sq.ID},
			bson.M{"$set": sq},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

// RemStatQueueProfileDrv removes a StatsQueue from dataDB
func (ms *MongoStorage) RemStatQueueProfileDrv(tenant, id string) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(ColSqp).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

// GetStatQueueDrv retrieves a StoredStatQueue
func (ms *MongoStorage) GetStatQueueDrv(tenant, id string) (sq *StatQueue, err error) {
	ssq := new(StoredStatQueue)
	if err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColSqs).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		if err := cur.Decode(ssq); err != nil {
			sq = nil
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	}); err != nil {
		return
	}
	sq, err = ssq.AsStatQueue(ms.ms)
	return
}

// SetStatQueueDrv stores the metrics for a StoredStatQueue
func (ms *MongoStorage) SetStatQueueDrv(ssq *StoredStatQueue, sq *StatQueue) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColSqs).UpdateOne(sctx, bson.M{"tenant": ssq.Tenant, "id": ssq.ID},
			bson.M{"$set": ssq},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

// RemStatQueueDrv removes stored metrics for a StoredStatQueue
func (ms *MongoStorage) RemStatQueueDrv(tenant, id string) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(ColSqs).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

// GetThresholdProfileDrv retrieves a ThresholdProfile from dataDB
func (ms *MongoStorage) GetThresholdProfileDrv(tenant, ID string) (tp *ThresholdProfile, err error) {
	tp = new(ThresholdProfile)
	err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColTps).FindOne(sctx, bson.M{"tenant": tenant, "id": ID})
		if err := cur.Decode(tp); err != nil {
			tp = nil
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	return
}

// SetThresholdProfileDrv stores a ThresholdProfile into DataDB
func (ms *MongoStorage) SetThresholdProfileDrv(tp *ThresholdProfile) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColTps).UpdateOne(sctx, bson.M{"tenant": tp.Tenant, "id": tp.ID},
			bson.M{"$set": tp}, options.Update().SetUpsert(true),
		)
		return err
	})
}

// RemoveThresholdProfile removes a ThresholdProfile from dataDB/cache
func (ms *MongoStorage) RemThresholdProfileDrv(tenant, id string) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(ColTps).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetThresholdDrv(tenant, id string) (r *Threshold, err error) {
	r = new(Threshold)
	err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColThs).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		if err := cur.Decode(r); err != nil {
			r = nil
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	return
}

func (ms *MongoStorage) SetThresholdDrv(r *Threshold) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColThs).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveThresholdDrv(tenant, id string) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(ColThs).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetFilterDrv(tenant, id string) (r *Filter, err error) {
	r = new(Filter)
	if err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColFlt).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		if err := cur.Decode(r); err != nil {
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return
}

func (ms *MongoStorage) SetFilterDrv(r *Filter) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColFlt).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveFilterDrv(tenant, id string) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(ColFlt).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetSupplierProfileDrv(tenant, id string) (r *SupplierProfile, err error) {
	r = new(SupplierProfile)
	err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColSpp).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		if err := cur.Decode(r); err != nil {
			r = nil
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	return
}

func (ms *MongoStorage) SetSupplierProfileDrv(r *SupplierProfile) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColSpp).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveSupplierProfileDrv(tenant, id string) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(ColSpp).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetAttributeProfileDrv(tenant, id string) (r *AttributeProfile, err error) {
	r = new(AttributeProfile)
	err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColAttr).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		if err := cur.Decode(r); err != nil {
			r = nil
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	return
}

func (ms *MongoStorage) SetAttributeProfileDrv(r *AttributeProfile) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColAttr).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveAttributeProfileDrv(tenant, id string) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(ColAttr).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetChargerProfileDrv(tenant, id string) (r *ChargerProfile, err error) {
	r = new(ChargerProfile)
	err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCpp).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		if err := cur.Decode(r); err != nil {
			r = nil
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	return
}

func (ms *MongoStorage) SetChargerProfileDrv(r *ChargerProfile) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColCpp).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveChargerProfileDrv(tenant, id string) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(ColCpp).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetDispatcherProfileDrv(tenant, id string) (r *DispatcherProfile, err error) {
	r = new(DispatcherProfile)
	err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColDpp).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		if err := cur.Decode(r); err != nil {
			r = nil
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	return
}

func (ms *MongoStorage) SetDispatcherProfileDrv(r *DispatcherProfile) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColDpp).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveDispatcherProfileDrv(tenant, id string) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(ColDpp).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetDispatcherHostDrv(tenant, id string) (r *DispatcherHost, err error) {
	r = new(DispatcherHost)
	err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColDph).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		if err := cur.Decode(r); err != nil {
			r = nil
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	return
}

func (ms *MongoStorage) SetDispatcherHostDrv(r *DispatcherHost) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColDph).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveDispatcherHostDrv(tenant, id string) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(ColDph).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetItemLoadIDsDrv(itemIDPrefix string) (loadIDs map[string]int64, err error) {
	fop := options.FindOne()
	if itemIDPrefix != "" {
		fop.SetProjection(bson.M{itemIDPrefix: 1, "_id": 0})
	} else {
		fop.SetProjection(bson.M{"_id": 0})
	}
	if err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColLID).FindOne(sctx, bson.D{}, fop)
		if err := cur.Decode(&loadIDs); err != nil {
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if len(loadIDs) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MongoStorage) SetLoadIDsDrv(loadIDs map[string]int64) (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColLID).UpdateOne(sctx, bson.D{}, bson.M{"$set": loadIDs},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveLoadIDsDrv() (err error) {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColLID).DeleteMany(sctx, bson.M{})
		return err
	})
}
