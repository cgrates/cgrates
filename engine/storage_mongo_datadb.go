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
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/ericlagergren/decimal"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Collection names in MongoDB.
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
	ColIndx = "indexes"
	ColTmg  = "timings"
	ColRes  = "resources"
	ColSqs  = "statqueues"
	ColTrp  = "trend_profiles"
	ColTrd  = "trends"
	ColSqp  = "statqueue_profiles"
	ColRgp  = "ranking_profiles"
	ColRnk  = "rankings"
	ColTps  = "threshold_profiles"
	ColThs  = "thresholds"
	ColFlt  = "filters"
	ColRts  = "route_profiles"
	ColAttr = "attribute_profiles"
	ColCDRs = "cdrs"
	ColCpp  = "charger_profiles"
	ColDpp  = "dispatcher_profiles"
	ColDph  = "dispatcher_hosts"
	ColLID  = "load_ids"
	ColBkup = "sessions_backup"
)

var (
	CGRIDLow       = strings.ToLower(utils.CGRID)
	RunIDLow       = strings.ToLower(utils.RunID)
	OrderIDLow     = strings.ToLower(utils.OrderID)
	OriginHostLow  = strings.ToLower(utils.OriginHost)
	OriginIDLow    = strings.ToLower(utils.OriginID)
	ToRLow         = strings.ToLower(utils.ToR)
	CDRHostLow     = strings.ToLower(utils.OriginHost)
	CDRSourceLow   = strings.ToLower(utils.Source)
	RequestTypeLow = strings.ToLower(utils.RequestType)
	TenantLow      = strings.ToLower(utils.Tenant)
	CategoryLow    = strings.ToLower(utils.Category)
	AccountLow     = strings.ToLower(utils.AccountField)
	SubjectLow     = strings.ToLower(utils.Subject)
	SetupTimeLow   = strings.ToLower(utils.SetupTime)
	AnswerTimeLow  = strings.ToLower(utils.AnswerTime)
	CreatedAtLow   = strings.ToLower(utils.CreatedAt)
	UpdatedAtLow   = strings.ToLower(utils.UpdatedAt)
	UsageLow       = strings.ToLower(utils.Usage)
	DestinationLow = strings.ToLower(utils.Destination)
	CostLow        = strings.ToLower(utils.Cost)
	CostSourceLow  = strings.ToLower(utils.CostSource)
)

func decimalEncoder(ec bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	decimalType := reflect.TypeOf(utils.Decimal{})

	// All encoder implementations should check that val is valid and is of
	// the correct type before proceeding.
	if !val.IsValid() || val.Type() != decimalType {
		return bsoncodec.ValueEncoderError{
			Name:     "decimalEncoder",
			Types:    []reflect.Type{decimalType},
			Received: val,
		}
	}

	sls, err := val.Interface().(utils.Decimal).MarshalText()
	if err != nil {
		return err
	}

	return vw.WriteBinary(sls)
}

func decimalDecoder(dc bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	decimalType := reflect.TypeOf(utils.Decimal{})

	// All decoder implementations should check that val is valid, settable,
	// and is of the correct kind before proceeding.
	if !val.IsValid() || !val.CanSet() || val.Type() != decimalType {
		return bsoncodec.ValueDecoderError{
			Name:     "decimalDecoder",
			Types:    []reflect.Type{decimalType},
			Received: val,
		}
	}

	data, _, err := vr.ReadBinary()
	if err != nil {
		return err
	}
	dBig := new(decimal.Big)
	if err := dBig.UnmarshalText(data); err != nil {
		return err
	}
	val.Set(reflect.ValueOf(utils.Decimal{Big: dBig}))
	return nil
}

// NewMongoStorage initializes a new MongoDB storage instance with provided connection parameters and settings.
// Returns an error if the setup fails.
func NewMongoStorage(scheme, host, port, db, user, pass, mrshlerStr string, storageType string,
	cdrsIndexes []string, ttl time.Duration) (*MongoStorage, error) {
	mongoStorage := &MongoStorage{
		ctx:         context.TODO(),
		ctxTTL:      ttl,
		cdrsIndexes: cdrsIndexes,
		storageType: storageType,
		counter:     utils.NewCounter(time.Now().UnixNano(), 0),
	}
	uri := composeMongoURI(scheme, host, port, db, user, pass)
	reg := bson.NewRegistry()
	decimalType := reflect.TypeOf(utils.Decimal{})
	reg.RegisterTypeEncoder(decimalType, bsoncodec.ValueEncoderFunc(decimalEncoder))
	reg.RegisterTypeDecoder(decimalType, bsoncodec.ValueDecoderFunc(decimalDecoder))
	// serverAPI := options.ServerAPI(options.ServerAPIVersion1).SetStrict(true).SetDeprecationErrors(true)
	opts := options.Client().
		ApplyURI(uri).
		SetRegistry(reg).
		SetServerSelectionTimeout(mongoStorage.ctxTTL).
		SetRetryWrites(false) // default is true
		// SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server
	var err error
	mongoStorage.client, err = mongo.Connect(mongoStorage.ctx, opts)
	if err != nil {
		return nil, err
	}

	mongoStorage.ms, err = NewMarshaler(mrshlerStr)
	if err != nil {
		return nil, err
	}
	if db != "" {
		// Populate ms.db with the url path after trimming everything after '?'.
		mongoStorage.db = strings.Split(db, "?")[0]
	}

	err = mongoStorage.query(func(sctx mongo.SessionContext) error {
		// Create indexes only if the database is empty or only the version table is present.
		cols, err := mongoStorage.client.Database(mongoStorage.db).
			ListCollectionNames(sctx, bson.D{})
		if err != nil {
			return err
		}
		empty := true
		for _, col := range cols {
			if col != ColVer {
				empty = false
				break
			}
		}
		if empty {
			return mongoStorage.EnsureIndexes()
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return mongoStorage, nil
}

// MongoStorage represents a storage interface for the new MongoDB driver.
type MongoStorage struct {
	client      *mongo.Client
	ctx         context.Context
	ctxTTL      time.Duration
	ctxTTLMutex sync.RWMutex // used for TTL reload
	db          string
	storageType string // DataDB/StorDB
	ms          Marshaler
	cdrsIndexes []string
	counter     *utils.Counter
}

func (ms *MongoStorage) query(argfunc func(ctx mongo.SessionContext) error) error {
	ms.ctxTTLMutex.RLock()
	ctxSession, ctxSessionCancel := context.WithTimeout(ms.ctx, ms.ctxTTL)
	ms.ctxTTLMutex.RUnlock()
	defer ctxSessionCancel()
	return ms.client.UseSession(ctxSession, argfunc)
}

// IsDataDB returns whether or not the storage is used for DataDB.
func (ms *MongoStorage) IsDataDB() bool {
	return ms.storageType == utils.DataDB
}

// SetTTL sets the context TTL used for queries (Thread-safe).
func (ms *MongoStorage) SetTTL(ttl time.Duration) {
	ms.ctxTTLMutex.Lock()
	ms.ctxTTL = ttl
	ms.ctxTTLMutex.Unlock()
}

func (ms *MongoStorage) enusureIndex(colName string, uniq bool, keys ...string) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		col := ms.getCol(colName)
		indexOptions := options.Index().SetUnique(uniq)
		doc := make(bson.D, 0)
		for _, k := range keys {
			doc = append(doc, bson.E{Key: k, Value: 1})
		}
		_, err := col.Indexes().CreateOne(sctx, mongo.IndexModel{
			Keys:    doc,
			Options: indexOptions,
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

// GetContext returns the context used for the current database.
func (ms *MongoStorage) GetContext() context.Context {
	return ms.ctx
}

func isNotFound(err error) bool {
	var de *mongo.CommandError

	if errors.As(err, &de) {
		return de.Code == 26 || de.Message == "ns not found"
	}

	// If the error cannot be converted to mongo.CommandError
	// check if the error message contains "ns not found"
	return strings.Contains(err.Error(), "ns not found")
}

func (ms *MongoStorage) ensureIndexesForCol(col string) error { // exported for migrator
	err := ms.dropAllIndexesForCol(col)
	if err != nil && !isNotFound(err) { // make sure you do not have indexes
		return err
	}
	switch col {
	case ColAct, ColApl, ColAAp, ColAtr, ColRpl, ColDst, ColRds, ColLht, ColIndx:
		err = ms.enusureIndex(col, true, "key")
	case ColRsP, ColRes, ColSqs, ColRgp, ColTrp, ColRnk, ColSqp, ColTps, ColThs, ColTrd, ColRts, ColAttr, ColFlt, ColCpp, ColDpp, ColDph:
		err = ms.enusureIndex(col, true, "tenant", "id")
	case ColRpf, ColShg, ColAcc:
		err = ms.enusureIndex(col, true, "id")
		// StorDB
	case utils.TBLTPTimings, utils.TBLTPDestinations,
		utils.TBLTPDestinationRates, utils.TBLTPRatingPlans,
		utils.TBLTPSharedGroups, utils.TBLTPActions, utils.TBLTPRankings,
		utils.TBLTPActionPlans, utils.TBLTPActionTriggers,
		utils.TBLTPStats, utils.TBLTPResources, utils.TBLTPDispatchers,
		utils.TBLTPDispatcherHosts, utils.TBLTPChargers,
		utils.TBLTPRoutes, utils.TBLTPThresholds:
		err = ms.enusureIndex(col, true, "tpid", "id")
	case utils.TBLTPRatingProfiles:
		err = ms.enusureIndex(col, true, "tpid", "tenant",
			"category", "subject", "loadid")
	case utils.SessionCostsTBL:
		err = ms.enusureIndex(col, true, CGRIDLow, RunIDLow)
		if err == nil {
			err = ms.enusureIndex(col, false, OriginHostLow, OriginIDLow)
		}
		if err == nil {
			err = ms.enusureIndex(col, false, RunIDLow, OriginIDLow)
		}
	case utils.CDRsTBL:
		err = ms.enusureIndex(col, true, CGRIDLow, RunIDLow,
			OriginIDLow)
		if err == nil {
			for _, idxKey := range ms.cdrsIndexes {
				err = ms.enusureIndex(col, false, idxKey)
				if err != nil {
					break
				}
			}
		}
	}
	return err
}

// EnsureIndexes creates database indexes for the specified collections.
func (ms *MongoStorage) EnsureIndexes(cols ...string) error {
	if len(cols) == 0 {
		if ms.IsDataDB() {
			cols = []string{
				ColAct, ColApl, ColAAp, ColAtr, ColRpl, ColDst, ColRds, ColLht, ColIndx,
				ColRsP, ColRes, ColSqs, ColSqp, ColTps, ColThs, ColRts, ColAttr, ColFlt, ColCpp,
				ColDpp, ColRpf, ColShg, ColAcc, ColRgp, ColTrp, ColTrd, ColRnk,
			}
		} else {
			cols = []string{
				utils.TBLTPTimings, utils.TBLTPDestinations, utils.TBLTPDestinationRates,
				utils.TBLTPRatingPlans, utils.TBLTPSharedGroups, utils.TBLTPActions, utils.TBLTPActionPlans,
				utils.TBLTPActionTriggers, utils.TBLTPRankings, utils.TBLTPStats, utils.TBLTPResources, utils.TBLTPRatingProfiles,
				utils.CDRsTBL, utils.SessionCostsTBL,
			}
		}
	}
	for _, col := range cols {
		if err := ms.ensureIndexesForCol(col); err != nil {
			return err
		}
	}
	return nil
}

// Close disconnects the MongoDB client.
func (ms *MongoStorage) Close() {
	if err := ms.client.Disconnect(ms.ctx); err != nil {
		utils.Logger.Err(fmt.Sprintf("<MongoStorage> Error on disconnect:%s", err))
	}
}

// Flush drops the datatable and recreates the indexes.
func (ms *MongoStorage) Flush(_ string) (err error) {
	return ms.query(func(sctx mongo.SessionContext) error {
		err := ms.client.Database(ms.db).Drop(sctx)
		if err != nil {
			return err
		}
		return ms.EnsureIndexes()
	})
}

// DB returns the database object associated with the MongoDB client.
func (ms *MongoStorage) DB() *mongo.Database {
	return ms.client.Database(ms.db)
}

// SelectDatabase selects the specified database.
func (ms *MongoStorage) SelectDatabase(dbName string) error {
	ms.db = dbName
	return nil
}

func (ms *MongoStorage) RemoveKeysForPrefix(prefix string) error {
	var colName string
	switch prefix {
	case utils.DestinationPrefix:
		colName = ColDst
	case utils.ReverseDestinationPrefix:
		colName = ColRds
	case utils.ActionPrefix:
		colName = ColAct
	case utils.ActionPlanPrefix:
		colName = ColApl
	case utils.AccountActionPlansPrefix:
		colName = ColAAp
	case utils.TasksKey:
		colName = ColTsk
	case utils.ActionTriggerPrefix:
		colName = ColAtr
	case utils.RatingPlanPrefix:
		colName = ColRpl
	case utils.RatingProfilePrefix:
		colName = ColRpf
	case utils.AccountPrefix:
		colName = ColAcc
	case utils.SharedGroupPrefix:
		colName = ColShg
	case utils.LoadInstKey:
		colName = ColLht
	case utils.VersionPrefix:
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
	case utils.RankingsProfilePrefix:
		colName = ColRgp
	case utils.TrendsProfilePrefix:
		colName = ColTrp
	case utils.TrendPrefix:
		colName = ColTrd
	case utils.RankingPrefix:
		colName = ColRnk
	case utils.ThresholdPrefix:
		colName = ColThs
	case utils.FilterPrefix:
		colName = ColFlt
	case utils.RouteProfilePrefix:
		colName = ColRts
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

// IsDBEmpty checks if the database is empty by verifying if each collection is empty.
func (ms *MongoStorage) IsDBEmpty() (isEmpty bool, err error) {
	err = ms.query(func(sctx mongo.SessionContext) error {
		cols, err := ms.DB().ListCollectionNames(sctx, bson.D{})
		if err != nil {
			return err
		}
		for _, col := range cols {
			if col == utils.CDRsTBL { // ignore cdrs collection
				continue
			}
			count, err := ms.getCol(col).CountDocuments(sctx, bson.D{}, options.Count().SetLimit(1)) // limiting the count to 1 since we are only checking if the collection is empty
			if err != nil {
				return err
			}
			if count != 0 {
				return nil
			}
		}
		isEmpty = true
		return nil
	})
	return isEmpty, err
}

func (ms *MongoStorage) getAllKeysMatchingField(sctx mongo.SessionContext, col, prefix,
	subject, field string) (keys []string, err error) {
	fieldResult := bson.M{}
	iter, err := ms.getCol(col).Find(sctx,
		bson.M{
			field: primitive.Regex{
				Pattern: subject,
			},
		},
		options.Find().SetProjection(
			bson.M{
				field: 1,
			},
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
		keys = append(keys, prefix+fieldResult[field].(string))
	}
	return keys, iter.Close(sctx)
}

func (ms *MongoStorage) getAllKeysMatchingTenantID(sctx mongo.SessionContext, col, prefix,
	subject string, tntID *utils.TenantID) (keys []string, err error) {
	idResult := struct{ Tenant, ID string }{}
	elem := bson.M{}
	if tntID.Tenant != "" {
		elem["tenant"] = tntID.Tenant
	}
	if tntID.ID != "" {
		elem["id"] = primitive.Regex{
			Pattern: subject,
		}
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
		keys = append(keys, prefix+utils.ConcatenatedKey(idResult.Tenant, idResult.ID))
	}
	return keys, iter.Close(sctx)
}

func (ms *MongoStorage) getAllIndexKeys(sctx mongo.SessionContext, prefix string) (keys []string, err error) {
	fieldResult := bson.M{}
	iter, err := ms.getCol(ColIndx).Find(sctx,
		bson.M{
			"key": primitive.Regex{
				Pattern: "^" + prefix,
			},
		},
		options.Find().SetProjection(
			bson.M{"key": 1},
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
		keys = append(keys, fieldResult["key"].(string))
	}
	return keys, iter.Close(sctx)
}

// GetKeysForPrefix retrieves keys matching the specified prefix across different categories.
func (ms *MongoStorage) GetKeysForPrefix(prefix string) (keys []string, err error) {
	var category, subject string
	keyLen := len(utils.DestinationPrefix)
	if len(prefix) < keyLen {
		return nil, fmt.Errorf("unsupported prefix in GetKeysForPrefix: %s", prefix)
	}
	category = prefix[:keyLen] // prefix length
	tntID := utils.NewTenantID(prefix[keyLen:])
	subject = "^" + prefix[keyLen:] // old way, no tenant support
	err = ms.query(func(sctx mongo.SessionContext) error {
		var qryErr error
		switch category {
		case utils.DestinationPrefix:
			keys, qryErr = ms.getAllKeysMatchingField(sctx, ColDst, utils.DestinationPrefix, subject, "key")
		case utils.ReverseDestinationPrefix:
			keys, qryErr = ms.getAllKeysMatchingField(sctx, ColRds, utils.ReverseDestinationPrefix, subject, "key")
		case utils.RatingPlanPrefix:
			keys, qryErr = ms.getAllKeysMatchingField(sctx, ColRpl, utils.RatingPlanPrefix, subject, "key")
		case utils.RatingProfilePrefix:
			if strings.HasPrefix(prefix[keyLen:], utils.MetaOut) {
				// Rewrite the id as it starts with '*' (from "*out").
				subject = "^\\" + prefix[keyLen:]
			}
			keys, qryErr = ms.getAllKeysMatchingField(sctx, ColRpf, utils.RatingProfilePrefix, subject, "id")
		case utils.ActionPrefix:
			keys, qryErr = ms.getAllKeysMatchingField(sctx, ColAct, utils.ActionPrefix, subject, "key")
		case utils.ActionPlanPrefix:
			keys, qryErr = ms.getAllKeysMatchingField(sctx, ColApl, utils.ActionPlanPrefix, subject, "key")
		case utils.ActionTriggerPrefix:
			keys, qryErr = ms.getAllKeysMatchingField(sctx, ColAtr, utils.ActionTriggerPrefix, subject, "key")
		case utils.SharedGroupPrefix:
			keys, qryErr = ms.getAllKeysMatchingField(sctx, ColShg, utils.SharedGroupPrefix, subject, "id")
		case utils.AccountPrefix:
			keys, qryErr = ms.getAllKeysMatchingField(sctx, ColAcc, utils.AccountPrefix, subject, "id")
		case utils.ResourceProfilesPrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColRsP, utils.ResourceProfilesPrefix, subject, tntID)
		case utils.ResourcesPrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColRes, utils.ResourcesPrefix, subject, tntID)
		case utils.StatQueuePrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColSqs, utils.StatQueuePrefix, subject, tntID)
		case utils.RankingsProfilePrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColRgp, utils.RankingsProfilePrefix, subject, tntID)
		case utils.TrendsProfilePrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColTrp, utils.TrendsProfilePrefix, subject, tntID)
		case utils.StatQueueProfilePrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColSqp, utils.StatQueueProfilePrefix, subject, tntID)
		case utils.AccountActionPlansPrefix:
			keys, qryErr = ms.getAllKeysMatchingField(sctx, ColAAp, utils.AccountActionPlansPrefix, subject, "key")
		case utils.TimingsPrefix:
			keys, qryErr = ms.getAllKeysMatchingField(sctx, ColTmg, utils.TimingsPrefix, subject, "id")
		case utils.TrendPrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColTrd, utils.TrendPrefix, subject, tntID)
		case utils.RankingPrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColRnk, utils.RankingPrefix, subject, tntID)
		case utils.FilterPrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColFlt, utils.FilterPrefix, subject, tntID)
		case utils.ThresholdPrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColThs, utils.ThresholdPrefix, subject, tntID)
		case utils.ThresholdProfilePrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColTps, utils.ThresholdProfilePrefix, subject, tntID)
		case utils.RouteProfilePrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColRts, utils.RouteProfilePrefix, subject, tntID)
		case utils.AttributeProfilePrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColAttr, utils.AttributeProfilePrefix, subject, tntID)
		case utils.ChargerProfilePrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColCpp, utils.ChargerProfilePrefix, subject, tntID)
		case utils.DispatcherProfilePrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColDpp, utils.DispatcherProfilePrefix, subject, tntID)
		case utils.DispatcherHostPrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColDph, utils.DispatcherHostPrefix, subject, tntID)
		case utils.AttributeFilterIndexes:
			keys, qryErr = ms.getAllIndexKeys(sctx, utils.AttributeFilterIndexes)
		case utils.ResourceFilterIndexes:
			keys, qryErr = ms.getAllIndexKeys(sctx, utils.ResourceFilterIndexes)
		case utils.StatFilterIndexes:
			keys, qryErr = ms.getAllIndexKeys(sctx, utils.StatFilterIndexes)
		case utils.ThresholdFilterIndexes:
			keys, qryErr = ms.getAllIndexKeys(sctx, utils.ThresholdFilterIndexes)
		case utils.RouteFilterIndexes:
			keys, qryErr = ms.getAllIndexKeys(sctx, utils.RouteFilterIndexes)
		case utils.ChargerFilterIndexes:
			keys, qryErr = ms.getAllIndexKeys(sctx, utils.ChargerFilterIndexes)
		case utils.DispatcherFilterIndexes:
			keys, qryErr = ms.getAllIndexKeys(sctx, utils.DispatcherFilterIndexes)
		case utils.ActionPlanIndexes:
			keys, qryErr = ms.getAllIndexKeys(sctx, utils.ActionPlanIndexes)
		case utils.FilterIndexPrfx:
			keys, qryErr = ms.getAllIndexKeys(sctx, utils.FilterIndexPrfx)
		default:
			qryErr = fmt.Errorf("unsupported prefix in GetKeysForPrefix: %s", prefix)
		}
		return qryErr
	})
	return keys, err
}

func (ms *MongoStorage) HasDataDrv(category, subject, tenant string) (has bool, err error) {
	err = ms.query(func(sctx mongo.SessionContext) (err error) {
		var count int64
		switch category {
		case utils.DestinationPrefix:
			count, err = ms.getCol(ColDst).CountDocuments(sctx, bson.M{"key": subject})
		case utils.RatingPlanPrefix:
			count, err = ms.getCol(ColRpl).CountDocuments(sctx, bson.M{"key": subject})
		case utils.RatingProfilePrefix:
			count, err = ms.getCol(ColRpf).CountDocuments(sctx, bson.M{"key": subject})
		case utils.ActionPrefix:
			count, err = ms.getCol(ColAct).CountDocuments(sctx, bson.M{"key": subject})
		case utils.ActionPlanPrefix:
			count, err = ms.getCol(ColApl).CountDocuments(sctx, bson.M{"key": subject})
		case utils.AccountPrefix:
			count, err = ms.getCol(ColAcc).CountDocuments(sctx, bson.M{"id": subject})
		case utils.ResourcesPrefix:
			count, err = ms.getCol(ColRes).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.ResourceProfilesPrefix:
			count, err = ms.getCol(ColRsP).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.StatQueuePrefix:
			count, err = ms.getCol(ColSqs).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.StatQueueProfilePrefix:
			count, err = ms.getCol(ColSqp).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.RankingPrefix:
			count, err = ms.getCol(ColRnk).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.RankingsProfilePrefix:
			count, err = ms.getCol(ColSqp).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.TrendPrefix:
			count, err = ms.getCol(ColTrd).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.TrendsProfilePrefix:
			count, err = ms.getCol(ColTrp).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.ThresholdPrefix:
			count, err = ms.getCol(ColThs).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.ThresholdProfilePrefix:
			count, err = ms.getCol(ColTps).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.FilterPrefix:
			count, err = ms.getCol(ColFlt).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.RouteProfilePrefix:
			count, err = ms.getCol(ColRts).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
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

func (ms *MongoStorage) GetRatingPlanDrv(key string) (*RatingPlan, error) {
	var kv struct {
		Key   string
		Value []byte
	}
	err := ms.query(func(sctx mongo.SessionContext) (err error) {
		sr := ms.getCol(ColRpl).FindOne(sctx, bson.M{"key": key})
		decodeErr := sr.Decode(&kv)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	if err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(kv.Value)
	r, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}
	out, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	err = r.Close()
	if err != nil {
		return nil, err
	}
	var ratingPlan *RatingPlan
	err = ms.ms.Unmarshal(out, &ratingPlan)
	return ratingPlan, err
}

func (ms *MongoStorage) SetRatingPlanDrv(rp *RatingPlan) error {
	result, err := ms.ms.Marshal(rp)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	_, err = w.Write(result)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColRpl).UpdateOne(sctx, bson.M{"key": rp.Id},
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
	return ms.query(func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColRpl).DeleteMany(sctx, bson.M{"key": key})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetRatingProfileDrv(key string) (*RatingProfile, error) {
	rtProfile := new(RatingProfile)
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColRpf).FindOne(sctx, bson.M{"id": key})
		decodeErr := sr.Decode(rtProfile)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return rtProfile, err
}

func (ms *MongoStorage) SetRatingProfileDrv(rp *RatingProfile) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColRpf).UpdateOne(sctx, bson.M{"id": rp.Id},
			bson.M{"$set": rp},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveRatingProfileDrv(key string) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColRpf).DeleteMany(sctx, bson.M{"id": key})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetDestinationDrv(key, transactionID string) (*Destination, error) {
	var kv struct {
		Key   string
		Value []byte
	}
	err := ms.query(func(sctx mongo.SessionContext) (err error) {
		sr := ms.getCol(ColDst).FindOne(sctx, bson.M{"key": key})
		decodeErr := sr.Decode(&kv)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			Cache.Set(utils.CacheDestinations, key, nil, nil,
				cacheCommit(transactionID), transactionID)
			return utils.ErrNotFound
		}
		return decodeErr
	})
	if err != nil {
		return nil, err
	}
	b := bytes.NewBuffer(kv.Value)
	r, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}
	out, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	err = r.Close()
	if err != nil {
		return nil, err
	}
	var dst *Destination
	err = ms.ms.Unmarshal(out, &dst)
	return dst, err
}

func (ms *MongoStorage) SetDestinationDrv(dest *Destination, _ string) error {
	result, err := ms.ms.Marshal(dest)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	_, err = w.Write(result)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return ms.query(func(sctx mongo.SessionContext) error {
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
	transactionID string) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColDst).DeleteOne(sctx, bson.M{"key": destID})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) RemoveReverseDestinationDrv(dstID, prfx, transactionID string) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColRds).UpdateOne(sctx, bson.M{"key": prfx},
			bson.M{"$pull": bson.M{"value": dstID}})
		return err
	})
}

func (ms *MongoStorage) GetReverseDestinationDrv(prefix, transactionID string) ([]string, error) {
	var result struct {
		Key   string
		Value []string
	}
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColRds).FindOne(sctx, bson.M{"key": prefix})
		decodeErr := sr.Decode(&result)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	if err != nil {
		return nil, err
	}
	return result.Value, nil
}

func (ms *MongoStorage) SetReverseDestinationDrv(destID string, prefixes []string, _ string) error {
	for _, p := range prefixes {
		err := ms.query(func(sctx mongo.SessionContext) error {
			_, qryErr := ms.getCol(ColRds).UpdateOne(sctx, bson.M{"key": p},
				bson.M{"$addToSet": bson.M{"value": destID}},
				options.Update().SetUpsert(true),
			)
			return qryErr
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (ms *MongoStorage) GetActionsDrv(key string) (Actions, error) {
	var result struct {
		Key   string
		Value Actions
	}
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColAct).FindOne(sctx, bson.M{"key": key})
		decodeErr := sr.Decode(&result)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return result.Value, err
}

func (ms *MongoStorage) SetActionsDrv(key string, as Actions) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColAct).UpdateOne(sctx, bson.M{"key": key},
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
	return ms.query(func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColAct).DeleteOne(sctx, bson.M{"key": key})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetSharedGroupDrv(key string) (*SharedGroup, error) {
	sg := new(SharedGroup)
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColShg).FindOne(sctx, bson.M{"id": key})
		decodeErr := sr.Decode(sg)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return sg, err
}

func (ms *MongoStorage) SetSharedGroupDrv(sg *SharedGroup) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColShg).UpdateOne(sctx, bson.M{"id": sg.Id},
			bson.M{"$set": sg},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveSharedGroupDrv(id string) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColShg).DeleteOne(sctx, bson.M{"id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetAccountDrv(key string) (*Account, error) {
	acc := new(Account)
	err := ms.query(func(sctx mongo.SessionContext) (err error) {
		sr := ms.getCol(ColAcc).FindOne(sctx, bson.M{"id": key})
		decodeErr := sr.Decode(acc)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return acc, err
}

func (ms *MongoStorage) SetAccountDrv(acc *Account) error {
	// never override existing account with an empty one
	// UPDATE: if all balances expired and were cleaned it makes
	// sense to write empty balance map
	if len(acc.BalanceMap) == 0 {
		ac, err := ms.GetAccountDrv(acc.ID)
		if err == nil && !ac.allBalancesExpired() {
			ac.ActionTriggers = acc.ActionTriggers
			ac.UnitCounters = acc.UnitCounters
			ac.AllowNegative = acc.AllowNegative
			ac.Disabled = acc.Disabled
			acc = ac
		}
	}
	acc.UpdateTime = time.Now()
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColAcc).UpdateOne(sctx, bson.M{"id": acc.ID},
			bson.M{"$set": acc},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveAccountDrv(key string) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColAcc).DeleteOne(sctx, bson.M{"id": key})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

// GetLoadHistory retrieves the last n items from the load history, newest first.
func (ms *MongoStorage) GetLoadHistory(limit int, skipCache bool,
	transactionID string) ([]*utils.LoadInstance, error) {
	if limit == 0 {
		return nil, nil
	}
	if !skipCache {
		x, ok := Cache.Get(utils.LoadInstKey, "")
		if ok {
			if x != nil {
				items, ok := x.([]*utils.LoadInstance)
				if !ok {
					return nil, utils.ErrCastFailed
				}
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
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColLht).FindOne(sctx, bson.M{"key": utils.LoadInstKey})
		decodeErr := sr.Decode(&kv)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	cCommit := cacheCommit(transactionID)
	if err == nil {
		if errCh := Cache.Remove(utils.LoadInstKey, "", cCommit, transactionID); errCh != nil {
			return nil, errCh
		}
		if errCh := Cache.Set(utils.LoadInstKey, "", kv.Value, nil, cCommit, transactionID); errCh != nil {
			return nil, errCh
		}
	}
	if len(kv.Value) < limit || limit == -1 {
		return kv.Value, nil
	}
	return kv.Value[:limit], nil
}

// AddLoadHistory adds a single load instance to the load history.
func (ms *MongoStorage) AddLoadHistory(ldInst *utils.LoadInstance,
	loadHistSize int, transactionID string) error {
	if loadHistSize == 0 { // Load history disabled
		return nil
	}
	// Get existing load history.
	var existingLoadHistory []*utils.LoadInstance
	var kv struct {
		Key   string
		Value []*utils.LoadInstance
	}
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColLht).FindOne(sctx, bson.M{"key": utils.LoadInstKey})
		decodeErr := sr.Decode(&kv)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return nil // utils.ErrNotFound
		}
		return decodeErr
	})
	if kv.Value != nil {
		existingLoadHistory = kv.Value
	}

	// Make sure we do it locked since other instances can modify the history while we read it.
	err = guardian.Guardian.Guard(func() error {

		// Insert at the first position.
		existingLoadHistory = append(existingLoadHistory, nil)
		copy(existingLoadHistory[1:], existingLoadHistory[0:])
		existingLoadHistory[0] = ldInst

		histLen := len(existingLoadHistory)
		if histLen >= loadHistSize { // Have hit maximum history allowed, remove oldest element in order to add new one
			existingLoadHistory = existingLoadHistory[:loadHistSize]
		}
		return ms.query(func(sctx mongo.SessionContext) (err error) {
			_, err = ms.getCol(ColLht).UpdateOne(sctx, bson.M{"key": utils.LoadInstKey},
				bson.M{"$set": struct {
					Key   string
					Value []*utils.LoadInstance
				}{Key: utils.LoadInstKey, Value: existingLoadHistory}},
				options.Update().SetUpsert(true),
			)
			return err
		})
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.LoadInstKey)

	if errCh := Cache.Remove(utils.LoadInstKey, "",
		cacheCommit(transactionID), transactionID); errCh != nil {
		return errCh
	}
	return err
}

func (ms *MongoStorage) GetActionTriggersDrv(key string) (ActionTriggers, error) {
	var kv struct {
		Key   string
		Value ActionTriggers
	}
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColAtr).FindOne(sctx, bson.M{"key": key})
		decodeErr := sr.Decode(&kv)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return kv.Value, err
}

func (ms *MongoStorage) SetActionTriggersDrv(key string, atrs ActionTriggers) error {
	if len(atrs) == 0 {
		return ms.query(func(sctx mongo.SessionContext) error {
			_, err := ms.getCol(ColAtr).DeleteOne(sctx, bson.M{"key": key})
			return err
		})
	}
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColAtr).UpdateOne(sctx, bson.M{"key": key},
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

	return ms.query(func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColAtr).DeleteOne(sctx, bson.M{"key": key})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetActionPlanDrv(key string) (*ActionPlan, error) {
	var kv struct {
		Key   string
		Value []byte
	}
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColApl).FindOne(sctx, bson.M{"key": key})
		decodeErr := sr.Decode(&kv)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	if err != nil {
		return nil, err
	}
	b := bytes.NewBuffer(kv.Value)
	r, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}
	out, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	err = r.Close()
	if err != nil {
		return nil, err
	}
	var ap *ActionPlan
	err = ms.ms.Unmarshal(out, &ap)
	return ap, err
}

func (ms *MongoStorage) SetActionPlanDrv(key string, ats *ActionPlan) error {
	result, err := ms.ms.Marshal(ats)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	_, err = w.Write(result)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColApl).UpdateOne(sctx, bson.M{"key": key},
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
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColApl).DeleteOne(sctx, bson.M{"key": key})
		return err
	})
}

func (ms *MongoStorage) GetAllActionPlansDrv() (map[string]*ActionPlan, error) {
	keys, err := ms.GetKeysForPrefix(utils.ActionPlanPrefix)
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, utils.ErrNotFound
	}
	actionPlans := make(map[string]*ActionPlan, len(keys))
	for _, key := range keys {
		ap, err := ms.GetActionPlanDrv(key[len(utils.ActionPlanPrefix):])
		if err != nil {
			return nil, err
		}
		actionPlans[key[len(utils.ActionPlanPrefix):]] = ap
	}
	return actionPlans, nil
}

func (ms *MongoStorage) GetAccountActionPlansDrv(acntID string) ([]string, error) {
	var kv struct {
		Key   string
		Value []string
	}
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColAAp).FindOne(sctx, bson.M{"key": acntID})
		decodeErr := sr.Decode(&kv)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return kv.Value, err
}

func (ms *MongoStorage) SetAccountActionPlansDrv(acntID string, aPlIDs []string) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColAAp).UpdateOne(sctx, bson.M{"key": acntID},
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
func (ms *MongoStorage) RemAccountActionPlansDrv(acntID string) error {
	return ms.query(func(sctx mongo.SessionContext) error {
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

func (ms *MongoStorage) PopTask() (*Task, error) {
	v := struct {
		ID   primitive.ObjectID `bson:"_id"`
		Task *Task
	}{}
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColTsk).FindOneAndDelete(sctx, bson.D{})
		decodeErr := sr.Decode(&v)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return v.Task, err
}

func (ms *MongoStorage) GetResourceProfileDrv(tenant, id string) (*ResourceProfile, error) {
	rsProfile := new(ResourceProfile)
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColRsP).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(rsProfile)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return rsProfile, err
}

func (ms *MongoStorage) SetResourceProfileDrv(rp *ResourceProfile) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColRsP).UpdateOne(sctx, bson.M{"tenant": rp.Tenant, "id": rp.ID},
			bson.M{"$set": rp},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveResourceProfileDrv(tenant, id string) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColRsP).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetResourceDrv(tenant, id string) (*Resource, error) {
	resource := new(Resource)
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColRes).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(resource)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return resource, err
}

func (ms *MongoStorage) SetResourceDrv(r *Resource) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColRes).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveResourceDrv(tenant, id string) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColRes).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetTimingDrv(id string) (*utils.TPTiming, error) {
	timing := new(utils.TPTiming)
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColTmg).FindOne(sctx, bson.M{"id": id})
		decodeErr := sr.Decode(timing)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return timing, err
}

func (ms *MongoStorage) SetTimingDrv(t *utils.TPTiming) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColTmg).UpdateOne(sctx, bson.M{"id": t.ID},
			bson.M{"$set": t},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveTimingDrv(id string) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColTmg).DeleteOne(sctx, bson.M{"id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

// GetStatQueueProfileDrv retrieves a StatQueueProfile from dataDB
func (ms *MongoStorage) GetStatQueueProfileDrv(tenant string, id string) (*StatQueueProfile, error) {
	sqProfile := new(StatQueueProfile)
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColSqp).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(sqProfile)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return sqProfile, err
}

// SetStatQueueProfileDrv stores a StatsQueue into DataDB
func (ms *MongoStorage) SetStatQueueProfileDrv(sq *StatQueueProfile) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColSqp).UpdateOne(sctx, bson.M{"tenant": sq.Tenant, "id": sq.ID},
			bson.M{"$set": sq},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

// RemStatQueueProfileDrv removes a StatsQueue from dataDB
func (ms *MongoStorage) RemStatQueueProfileDrv(tenant, id string) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColSqp).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

// GetStatQueueDrv retrieves a StoredStatQueue
func (ms *MongoStorage) GetStatQueueDrv(tenant, id string) (*StatQueue, error) {
	ssq := new(StoredStatQueue)
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColSqs).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(ssq)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	if err != nil {
		return nil, err
	}
	return ssq.AsStatQueue(ms.ms)
}

// SetStatQueueDrv stores the metrics for a StoredStatQueue
func (ms *MongoStorage) SetStatQueueDrv(ssq *StoredStatQueue, sq *StatQueue) (err error) {
	if ssq == nil {
		if ssq, err = NewStoredStatQueue(sq, ms.ms); err != nil {
			return err
		}
	}
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColSqs).UpdateOne(sctx, bson.M{"tenant": ssq.Tenant, "id": ssq.ID},
			bson.M{"$set": ssq},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

// RemStatQueueDrv removes stored metrics for a StoredStatQueue
func (ms *MongoStorage) RemStatQueueDrv(tenant, id string) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColSqs).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetRankingProfileDrv(tenant, id string) (*RankingProfile, error) {
	rgProfile := new(RankingProfile)
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColRgp).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(rgProfile)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return rgProfile, err
}

func (ms *MongoStorage) SetRankingProfileDrv(sgp *RankingProfile) (err error) {
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColRgp).UpdateOne(sctx, bson.M{"tenant": sgp.Tenant, "id": sgp.ID},
			bson.M{"$set": sgp},
			options.Update().SetUpsert(true))
		return err
	})
}

func (ms *MongoStorage) RemRankingProfileDrv(tenant, id string) (err error) {
	return ms.query(func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColRgp).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetRankingDrv(tenant, id string) (*Ranking, error) {
	rn := new(Ranking)
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColRnk).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(rn)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return rn, err
}
func (ms *MongoStorage) SetRankingDrv(rn *Ranking) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColRnk).UpdateOne(sctx, bson.M{"tenant": rn.Tenant, "id": rn.ID},
			bson.M{"$set": rn},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveRankingDrv(tenant, id string) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColRnk).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetTrendProfileDrv(tenant, id string) (*TrendProfile, error) {
	srProfile := new(TrendProfile)
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColTrp).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(srProfile)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return srProfile, err
}

func (ms *MongoStorage) SetTrendProfileDrv(srp *TrendProfile) (err error) {
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColTrp).UpdateOne(sctx, bson.M{"tenant": srp.Tenant, "id": srp.ID},
			bson.M{"$set": srp},
			options.Update().SetUpsert(true))
		return err
	})
}

func (ms *MongoStorage) RemTrendProfileDrv(tenant, id string) (err error) {
	return ms.query(func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColTrp).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetTrendDrv(tenant, id string) (*Trend, error) {
	tr := new(Trend)
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColTrd).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(tr)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})

	return tr, err
}

func (ms *MongoStorage) SetTrendDrv(tr *Trend) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColTrd).UpdateOne(sctx, bson.M{"tenant": tr.Tenant, "id": tr.ID},
			bson.M{"$set": tr},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveTrendDrv(tenant, id string) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColTrd).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

// GetThresholdProfileDrv retrieves a ThresholdProfile from dataDB
func (ms *MongoStorage) GetThresholdProfileDrv(tenant, ID string) (*ThresholdProfile, error) {
	thProfile := new(ThresholdProfile)
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColTps).FindOne(sctx, bson.M{"tenant": tenant, "id": ID})
		decodeErr := sr.Decode(thProfile)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return thProfile, err
}

// SetThresholdProfileDrv stores a ThresholdProfile into DataDB
func (ms *MongoStorage) SetThresholdProfileDrv(tp *ThresholdProfile) error {
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColTps).UpdateOne(sctx, bson.M{"tenant": tp.Tenant, "id": tp.ID},
			bson.M{"$set": tp}, options.Update().SetUpsert(true),
		)
		return err
	})
}

// RemoveThresholdProfile removes a ThresholdProfile from dataDB/cache
func (ms *MongoStorage) RemThresholdProfileDrv(tenant, id string) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColTps).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetThresholdDrv(tenant, id string) (*Threshold, error) {
	th := new(Threshold)
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColThs).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(th)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return th, err
}

func (ms *MongoStorage) SetThresholdDrv(r *Threshold) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColThs).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveThresholdDrv(tenant, id string) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColThs).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetFilterDrv(tenant, id string) (*Filter, error) {
	fltr := new(Filter)
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColFlt).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(fltr)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	if err != nil {
		return nil, err
	}
	return fltr, err
}

func (ms *MongoStorage) SetFilterDrv(r *Filter) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColFlt).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveFilterDrv(tenant, id string) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColFlt).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetRouteProfileDrv(tenant, id string) (*RouteProfile, error) {
	routeProfile := new(RouteProfile)
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColRts).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(routeProfile)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return routeProfile, err
}

func (ms *MongoStorage) SetRouteProfileDrv(r *RouteProfile) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColRts).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveRouteProfileDrv(tenant, id string) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColRts).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetAttributeProfileDrv(tenant, id string) (*AttributeProfile, error) {
	attrProfile := new(AttributeProfile)
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColAttr).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(attrProfile)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return attrProfile, err
}

func (ms *MongoStorage) SetAttributeProfileDrv(r *AttributeProfile) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColAttr).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveAttributeProfileDrv(tenant, id string) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColAttr).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetChargerProfileDrv(tenant, id string) (*ChargerProfile, error) {
	chargerProfile := new(ChargerProfile)
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColCpp).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(chargerProfile)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return chargerProfile, err
}

func (ms *MongoStorage) SetChargerProfileDrv(r *ChargerProfile) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColCpp).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveChargerProfileDrv(tenant, id string) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColCpp).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetDispatcherProfileDrv(tenant, id string) (*DispatcherProfile, error) {
	dspProfile := new(DispatcherProfile)
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColDpp).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(dspProfile)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrDSPProfileNotFound
		}
		return decodeErr
	})
	return dspProfile, err
}

func (ms *MongoStorage) SetDispatcherProfileDrv(r *DispatcherProfile) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColDpp).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveDispatcherProfileDrv(tenant, id string) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColDpp).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		return err
	})
}

func (ms *MongoStorage) GetDispatcherHostDrv(tenant, id string) (*DispatcherHost, error) {
	dspHost := new(DispatcherHost)
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColDph).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(dspHost)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrDSPHostNotFound
		}
		return decodeErr
	})
	return dspHost, err
}

func (ms *MongoStorage) SetDispatcherHostDrv(r *DispatcherHost) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColDph).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveDispatcherHostDrv(tenant, id string) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColDph).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})

		return err
	})
}

func (ms *MongoStorage) GetItemLoadIDsDrv(itemIDPrefix string) (map[string]int64, error) {
	fop := options.FindOne()
	if itemIDPrefix != "" {
		fop.SetProjection(bson.M{itemIDPrefix: 1, "_id": 0})
	} else {
		fop.SetProjection(bson.M{"_id": 0})
	}
	loadIDs := make(map[string]int64)
	err := ms.query(func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColLID).FindOne(sctx, bson.D{}, fop)
		decodeErr := sr.Decode(&loadIDs)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	if err != nil {
		return nil, err
	}
	if len(loadIDs) == 0 {
		return nil, utils.ErrNotFound
	}
	return loadIDs, nil
}

func (ms *MongoStorage) SetLoadIDsDrv(loadIDs map[string]int64) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColLID).UpdateOne(sctx, bson.D{}, bson.M{"$set": loadIDs},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveLoadIDsDrv() error {
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColLID).DeleteMany(sctx, bson.M{})
		return err
	})
}

// GetIndexesDrv retrieves Indexes from dataDB
// the key is the tenant of the item or in case of context dependent profiles is a concatenatedKey between tenant and context
// id is used as a concatenated key in case of filterIndexes the id will be filterType:fieldName:fieldVal
func (ms *MongoStorage) GetIndexesDrv(idxItmType, tntCtx, idxKey string) (map[string]utils.StringSet, error) {
	type result struct {
		Key   string
		Value []string
	}
	dbKey := utils.CacheInstanceToPrefix[idxItmType] + tntCtx
	var q bson.M
	if len(idxKey) != 0 {
		q = bson.M{"key": utils.ConcatenatedKey(dbKey, idxKey)}
	} else {
		for _, character := range []string{".", "*"} {
			dbKey = strings.Replace(dbKey, character, `\`+character, strings.Count(dbKey, character))
		}
		// For optimization, use a caret (^) in the regex pattern.
		q = bson.M{"key": primitive.Regex{Pattern: "^" + dbKey}}
	}
	indexes := make(map[string]utils.StringSet)
	err := ms.query(func(sctx mongo.SessionContext) (qryErr error) {
		cur, qryErr := ms.getCol(ColIndx).Find(sctx, q)
		if qryErr != nil {
			return qryErr
		}
		defer func() {
			closeErr := cur.Close(sctx)
			if closeErr != nil && qryErr == nil {
				qryErr = closeErr
			}
		}()
		for cur.Next(sctx) {
			var elem result
			qryErr = cur.Decode(&elem)
			if qryErr != nil {
				return qryErr
			}
			if len(elem.Value) == 0 {
				continue
			}
			indexKey := strings.TrimPrefix(elem.Key, utils.CacheInstanceToPrefix[idxItmType]+tntCtx+utils.ConcatenatedKeySep)
			indexes[indexKey] = utils.NewStringSet(elem.Value)
		}
		return cur.Err()
	})
	if err != nil {
		return nil, err
	}
	if len(indexes) == 0 {
		return nil, utils.ErrNotFound
	}
	return indexes, nil
}

// SetIndexesDrv stores Indexes into DataDB
// the key is the tenant of the item or in case of context dependent profiles is a concatenatedKey between tenant and context
func (ms *MongoStorage) SetIndexesDrv(idxItmType, tntCtx string,
	indexes map[string]utils.StringSet, commit bool, transactionID string) error {
	originKey := utils.CacheInstanceToPrefix[idxItmType] + tntCtx
	dbKey := originKey
	if transactionID != utils.EmptyString {
		dbKey = "tmp_" + utils.ConcatenatedKey(originKey, transactionID)
	}
	if commit && transactionID != utils.EmptyString {
		regexKey := dbKey
		for _, character := range []string{".", "*"} {
			regexKey = strings.ReplaceAll(regexKey, character, `\`+character)
		}
		err := ms.query(func(sctx mongo.SessionContext) error {
			result, qryErr := ms.getAllIndexKeys(sctx, regexKey)
			for _, key := range result {
				idxKey := strings.TrimPrefix(key, dbKey)
				if _, qryErr = ms.getCol(ColIndx).DeleteOne(sctx,
					bson.M{"key": originKey + idxKey}); qryErr != nil { //ensure we do not have dup
					return qryErr
				}
				if _, qryErr = ms.getCol(ColIndx).UpdateOne(sctx, bson.M{"key": key},
					bson.M{"$set": bson.M{"key": originKey + idxKey}}, // only update the key
				); qryErr != nil {
					return qryErr
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	var lastErr error
	for idxKey, itmMp := range indexes {
		err := ms.query(func(sctx mongo.SessionContext) (qryErr error) {
			idxDbkey := utils.ConcatenatedKey(dbKey, idxKey)
			if len(itmMp) == 0 { // remove from DB if we set it with empty indexes
				_, qryErr = ms.getCol(ColIndx).DeleteOne(sctx,
					bson.M{"key": idxDbkey})
			} else {
				_, qryErr = ms.getCol(ColIndx).UpdateOne(sctx, bson.M{"key": idxDbkey},
					bson.M{"$set": bson.M{"key": idxDbkey, "value": itmMp.AsSlice()}},
					options.Update().SetUpsert(true),
				)
			}
			return qryErr
		})
		if err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// RemoveIndexesDrv removes the indexes
func (ms *MongoStorage) RemoveIndexesDrv(idxItmType, tntCtx, idxKey string) error {
	if len(idxKey) != 0 {
		return ms.query(func(sctx mongo.SessionContext) error {
			dr, err := ms.getCol(ColIndx).DeleteOne(sctx,
				bson.M{"key": utils.ConcatenatedKey(utils.CacheInstanceToPrefix[idxItmType]+tntCtx, idxKey)})
			if dr.DeletedCount == 0 {
				return utils.ErrNotFound
			}
			return err
		})
	}
	regexKey := utils.CacheInstanceToPrefix[idxItmType] + tntCtx
	for _, character := range []string{".", "*"} {
		regexKey = strings.ReplaceAll(regexKey, character, `\`+character)
	}
	// For optimization, use a caret (^) in the regex pattern.
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColIndx).DeleteMany(sctx, bson.M{
			"key": primitive.Regex{
				Pattern: "^" + regexKey,
			},
		})
		return err
	})
}

// used to "mold" the structure so that it appears in the first level of the mongo document and to make the conversion from mongo back to cgrates simpler
type mongoStoredSession struct {
	NodeID        string
	CGRID         string
	Tenant        string
	ResourceID    string
	ClientConnID  string
	EventStart    MapEvent
	DebitInterval time.Duration
	Chargeable    bool
	SRuns         []*StoredSRun
	OptsStart     MapEvent
	UpdatedAt     time.Time
}

// Will backup active sessions in DataDB
func (ms *MongoStorage) SetBackupSessionsDrv(nodeID, tnt string, storedSessions []*StoredSession) error {
	return ms.query(func(sctx mongo.SessionContext) error {
		for i := 0; i < len(storedSessions); i += 1000 {
			end := i + 1000
			if end > len(storedSessions) {
				end = len(storedSessions)
			}
			// split sessions into batches of 1001 sessons
			batch := storedSessions[i:end] //  if sessions < 1001, puts all sessions in 1 batch
			var models []mongo.WriteModel
			for _, sess := range batch {
				doc := bson.M{"$set": mongoStoredSession{
					NodeID:        nodeID,
					CGRID:         sess.CGRID,
					Tenant:        sess.Tenant,
					ResourceID:    sess.ResourceID,
					ClientConnID:  sess.ClientConnID,
					EventStart:    sess.EventStart,
					DebitInterval: sess.DebitInterval,
					Chargeable:    sess.Chargeable,
					SRuns:         sess.SRuns,
					OptsStart:     sess.OptsStart,
					UpdatedAt:     sess.UpdatedAt,
				}}
				model := mongo.NewUpdateOneModel().SetUpdate(doc).SetUpsert(true).SetFilter(bson.M{"nodeid": nodeID, "cgrid": sess.CGRID})
				models = append(models, model)
			}
			if len(models) != 0 {
				_, err := ms.getCol(ColBkup).BulkWrite(sctx, models)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// Will restore sessions that were active from dataDB backup
func (ms *MongoStorage) GetSessionsBackupDrv(nodeID, tnt string) ([]*StoredSession, error) {
	var storeSessions []*StoredSession
	if err := ms.query(func(sctx mongo.SessionContext) (qryErr error) {
		cur, qryErr := ms.getCol(ColBkup).Find(sctx, bson.M{"nodeid": nodeID})
		if qryErr != nil {
			return qryErr
		}
		defer func() {
			closeErr := cur.Close(sctx)
			if closeErr != nil && qryErr == nil {
				qryErr = closeErr
			}
		}()
		for cur.Next(sctx) {
			var result mongoStoredSession
			qryErr := cur.Decode(&result)
			if errors.Is(qryErr, mongo.ErrNoDocuments) {
				return utils.ErrNoBackupFound
			} else if qryErr != nil {
				return qryErr
			}
			oneStSession := &StoredSession{
				CGRID:         result.CGRID,
				Tenant:        result.Tenant,
				ResourceID:    result.ResourceID,
				ClientConnID:  result.ClientConnID,
				EventStart:    result.EventStart,
				DebitInterval: result.DebitInterval,
				Chargeable:    result.Chargeable,
				SRuns:         result.SRuns,
				OptsStart:     result.OptsStart,
				UpdatedAt:     result.UpdatedAt,
			}
			storeSessions = append(storeSessions, oneStSession)
		}
		if len(storeSessions) == 0 {
			return utils.ErrNoBackupFound
		}
		return
	}); err != nil {
		return nil, err
	}
	return storeSessions, nil
}

// Will remove one or all sessions from dataDB Backup
func (ms *MongoStorage) RemoveSessionsBackupDrv(nodeID, tnt, cgrid string) error {
	if cgrid == utils.EmptyString {
		return ms.query(func(sctx mongo.SessionContext) error {
			_, err := ms.getCol(ColBkup).DeleteMany(sctx, bson.M{"nodeid": nodeID})
			return err
		})
	}
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColBkup).DeleteOne(sctx, bson.M{"nodeid": nodeID, "cgrid": cgrid})
		return err
	})
}
