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
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/ericlagergren/decimal"

	"github.com/cgrates/birpc/context"
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
	ColRes  = "resources"
	ColSqs  = "statqueues"
	ColSqp  = "statqueue_profiles"
	ColTps  = "threshold_profiles"
	ColThs  = "thresholds"
	ColTrs  = "trend_profiles"
	ColTrd  = "trends"
	ColRgp  = "ranking_profiles"
	ColRnk  = "rankings"
	ColFlt  = "filters"
	ColRts  = "route_profiles"
	ColAttr = "attribute_profiles"
	ColCDRs = "cdrs"
	ColCpp  = "charger_profiles"
	ColRpp  = "rate_profiles"
	ColApp  = "action_profiles"
	ColLID  = "load_ids"
	ColAnp  = "account_profiles"
)

var (
	MetaOriginLow  = strings.ToLower(utils.MetaOriginID)
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
	dBig := decimal.WithContext(utils.DecimalContext)
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
	ctx := context.TODO()
	mongoStorage.client, err = mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	mongoStorage.ms, err = utils.NewMarshaler(mrshlerStr)
	if err != nil {
		return nil, err
	}
	if db != "" {
		// Populate ms.db with the url path after trimming everything after '?'.
		mongoStorage.db = strings.Split(db, "?")[0]
	}

	err = mongoStorage.query(ctx, func(sctx mongo.SessionContext) error {
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

// MongoStorage struct for new mongo driver
type MongoStorage struct {
	client      *mongo.Client
	ctxTTL      time.Duration
	ctxTTLMutex sync.RWMutex // used for TTL reload
	db          string
	storageType string // DataDB/StorDB
	ms          utils.Marshaler
	cdrsIndexes []string
	counter     *utils.Counter
}

func (ms *MongoStorage) query(ctx *context.Context, argfunc func(ctx mongo.SessionContext) error) error {
	ms.ctxTTLMutex.RLock()
	ctxSession, ctxSessionCancel := context.WithTimeout(ctx, ms.ctxTTL)
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

func (ms *MongoStorage) ensureIndex(colName string, uniq bool, keys ...string) error {
	return ms.query(context.TODO(), func(sctx mongo.SessionContext) error {
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
	return ms.query(context.TODO(), func(sctx mongo.SessionContext) error {
		col := ms.getCol(colName)
		_, err := col.Indexes().DropAll(sctx)
		return err
	})
}

func (ms *MongoStorage) getCol(col string) *mongo.Collection {
	return ms.client.Database(ms.db).Collection(col)
}

// GetContext returns the context used for the current database.
func (ms *MongoStorage) GetContext() *context.Context {
	return context.TODO()
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
		err = ms.ensureIndex(col, true, "key")
	case ColRsP, ColRes, ColSqs, ColRgp, ColTrs, ColTrd, ColSqp, ColTps, ColThs, ColRts, ColAttr, ColFlt, ColCpp, ColRpp, ColApp, ColAnp:
		err = ms.ensureIndex(col, true, "tenant", "id")
	case ColRpf, ColShg, ColAcc:
		err = ms.ensureIndex(col, true, "id")
	case utils.CDRsTBL:
		err = ms.ensureIndex(col, true, "opts.*cdrID") // should probably create a constant for the key
		if err == nil {
			for _, idxKey := range ms.cdrsIndexes {
				err = ms.ensureIndex(col, false, idxKey)
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
				ColRsP, ColRes, ColSqs, ColSqp, ColTps, ColThs, ColRts, ColAttr, ColFlt,
				ColCpp, ColRpp, ColApp, ColRpf, ColShg, ColAcc, ColAnp, ColTrd, ColTrs,
			}
		} else {
			cols = []string{utils.CDRsTBL}
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
	if err := ms.client.Disconnect(context.TODO()); err != nil {
		utils.Logger.Err(fmt.Sprintf("<MongoStorage> Error on disconect:%s", err))
	}
}

// Flush drops the datatable and recreates the indexes.
func (ms *MongoStorage) Flush(_ string) error {
	return ms.query(context.TODO(), func(sctx mongo.SessionContext) error {
		if err := ms.client.Database(ms.db).Drop(sctx); err != nil {
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

// IsDBEmpty checks if the database is empty by verifying if each collection is empty.
func (ms *MongoStorage) IsDBEmpty() (isEmpty bool, err error) {
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) error {
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

func (ms *MongoStorage) getAllKeysMatchingTenantID(sctx mongo.SessionContext, col, prefix string, tntID *utils.TenantID) (result []string, err error) {
	idResult := struct{ Tenant, ID string }{}
	elem := bson.M{}
	if tntID.Tenant != "" {
		elem["tenant"] = tntID.Tenant
	}
	if tntID.ID != "" {
		elem["id"] = primitive.Regex{

			// Note: Before replacing subject with the ID within TenantID,
			// we used to prefix the pattern with a caret(^).
			Pattern: tntID.ID,
		}
	}

	iter, err := ms.getCol(col).Find(sctx, elem,
		options.Find().SetProjection(bson.M{"tenant": 1, "id": 1}))
	if err != nil {
		return
	}
	for iter.Next(sctx) {
		err = iter.Decode(&idResult)
		if err != nil {
			return
		}
		result = append(result, prefix+utils.ConcatenatedKey(idResult.Tenant, idResult.ID))
	}
	return result, iter.Close(sctx)
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

// GetKeysForPrefix implementation
func (ms *MongoStorage) GetKeysForPrefix(ctx *context.Context, prefix string) (keys []string, err error) {
	keyLen := len(utils.AccountPrefix)
	if len(prefix) < keyLen {
		return nil, fmt.Errorf("unsupported prefix in GetKeysForPrefix: %q", prefix)
	}
	category := prefix[:keyLen] // prefix length
	tntID := utils.NewTenantID(prefix[keyLen:])
	err = ms.query(ctx, func(sctx mongo.SessionContext) (qryErr error) {
		switch category {
		case utils.ResourceProfilesPrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColRsP, utils.ResourceProfilesPrefix, tntID)
		case utils.ResourcesPrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColRes, utils.ResourcesPrefix, tntID)
		case utils.StatQueuePrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColSqs, utils.StatQueuePrefix, tntID)
		case utils.StatQueueProfilePrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColSqp, utils.StatQueueProfilePrefix, tntID)
		case utils.FilterPrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColFlt, utils.FilterPrefix, tntID)
		case utils.ThresholdPrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColThs, utils.ThresholdPrefix, tntID)
		case utils.ThresholdProfilePrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColTps, utils.ThresholdProfilePrefix, tntID)
		case utils.RankingProfilePrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColRgp, utils.RankingProfilePrefix, tntID)
		case utils.RankingPrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColRnk, utils.RankingPrefix, tntID)
		case utils.TrendProfilePrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColTrs, utils.TrendProfilePrefix, tntID)
		case utils.TrendPrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColTrd, utils.TrendPrefix, tntID)
		case utils.RouteProfilePrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColRts, utils.RouteProfilePrefix, tntID)
		case utils.AttributeProfilePrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColAttr, utils.AttributeProfilePrefix, tntID)
		case utils.ChargerProfilePrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColCpp, utils.ChargerProfilePrefix, tntID)
		case utils.RateProfilePrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColRpp, utils.RateProfilePrefix, tntID)
		case utils.ActionProfilePrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColApp, utils.ActionProfilePrefix, tntID)
		case utils.AccountPrefix:
			keys, qryErr = ms.getAllKeysMatchingTenantID(sctx, ColAnp, utils.AccountPrefix, tntID)
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
		case utils.ActionPlanIndexes:
			keys, qryErr = ms.getAllIndexKeys(sctx, utils.ActionPlanIndexes)
		case utils.ActionProfilesFilterIndexPrfx:
			keys, qryErr = ms.getAllIndexKeys(sctx, utils.ActionProfilesFilterIndexPrfx)
		case utils.AccountFilterIndexPrfx:
			keys, qryErr = ms.getAllIndexKeys(sctx, utils.AccountFilterIndexPrfx)
		case utils.RateProfilesFilterIndexPrfx:
			keys, qryErr = ms.getAllIndexKeys(sctx, utils.RateProfilesFilterIndexPrfx)
		case utils.RateFilterIndexPrfx:
			keys, qryErr = ms.getAllIndexKeys(sctx, utils.RateFilterIndexPrfx)
		case utils.FilterIndexPrfx:
			keys, qryErr = ms.getAllIndexKeys(sctx, utils.FilterIndexPrfx)
		default:
			qryErr = fmt.Errorf("unsupported prefix in GetKeysForPrefix: %q", prefix)
		}
		return qryErr
	})
	return keys, err
}

func (ms *MongoStorage) HasDataDrv(ctx *context.Context, category, subject, tenant string) (has bool, err error) {
	err = ms.query(ctx, func(sctx mongo.SessionContext) (err error) {
		var count int64
		switch category {
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
		case utils.RouteProfilePrefix:
			count, err = ms.getCol(ColRts).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.AttributeProfilePrefix:
			count, err = ms.getCol(ColAttr).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.ChargerProfilePrefix:
			count, err = ms.getCol(ColCpp).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.TrendPrefix:
			count, err = ms.getCol(ColTrd).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.TrendProfilePrefix:
			count, err = ms.getCol(ColTrs).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.RateProfilePrefix:
			count, err = ms.getCol(ColRpp).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.ActionProfilePrefix:
			count, err = ms.getCol(ColApp).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.AccountPrefix:
			count, err = ms.getCol(ColAnp).CountDocuments(sctx, bson.M{"tenant": tenant, "id": subject})
		default:
			err = fmt.Errorf("unsupported category in HasData: %s", category)
		}
		has = count > 0
		return err
	})
	return has, err
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
	err := ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		sr := ms.getCol(ColLht).FindOne(sctx, bson.M{"key": utils.LoadInstKey})
		decodeErr := sr.Decode(&kv)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	cCommit := cacheCommit(transactionID)
	if err == nil {
		if errCh := Cache.Remove(context.TODO(), utils.LoadInstKey, "", cCommit, transactionID); errCh != nil {
			return nil, errCh
		}
		if errCh := Cache.Set(context.TODO(), utils.LoadInstKey, "", kv.Value, nil, cCommit, transactionID); errCh != nil {
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
	err := ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		sr := ms.getCol(ColLht).FindOne(sctx, bson.M{"key": utils.LoadInstKey})
		decodeErr := sr.Decode(&kv)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return nil // utils.ErrNotFound
		}
		return decodeErr
	})
	if err != nil {
		return err
	}
	if kv.Value != nil {
		existingLoadHistory = kv.Value
	}

	// Make sure we do it locked since other instances can modify the history while we read it.
	err = guardian.Guardian.Guard(context.TODO(), func(ctx *context.Context) error {

		// Insert at the first position.
		existingLoadHistory = append(existingLoadHistory, nil)
		copy(existingLoadHistory[1:], existingLoadHistory[0:])
		existingLoadHistory[0] = ldInst

		histLen := len(existingLoadHistory)
		if histLen >= loadHistSize { // Have hit maximum history allowed, remove oldest element in order to add new one
			existingLoadHistory = existingLoadHistory[:loadHistSize]
		}
		return ms.query(ctx, func(sctx mongo.SessionContext) (err error) {
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

	if errCh := Cache.Remove(context.TODO(), utils.LoadInstKey, "",
		cacheCommit(transactionID), transactionID); errCh != nil {
		return errCh
	}
	return err
}

func (ms *MongoStorage) GetResourceProfileDrv(ctx *context.Context, tenant, id string) (*ResourceProfile, error) {
	rsProfile := new(ResourceProfile)
	err := ms.query(ctx, func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColRsP).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(rsProfile)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return rsProfile, err
}

func (ms *MongoStorage) SetResourceProfileDrv(ctx *context.Context, rp *ResourceProfile) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColRsP).UpdateOne(sctx, bson.M{"tenant": rp.Tenant, "id": rp.ID},
			bson.M{"$set": rp},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveResourceProfileDrv(ctx *context.Context, tenant, id string) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColRsP).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetResourceDrv(ctx *context.Context, tenant, id string) (*Resource, error) {
	resource := new(Resource)
	err := ms.query(ctx, func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColRes).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(resource)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return resource, err
}

func (ms *MongoStorage) SetResourceDrv(ctx *context.Context, r *Resource) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColRes).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveResourceDrv(ctx *context.Context, tenant, id string) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColRes).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetRankingProfileDrv(ctx *context.Context, tenant, id string) (*utils.RankingProfile, error) {
	rgProfile := new(utils.RankingProfile)
	err := ms.query(ctx, func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColRgp).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(rgProfile)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return rgProfile, err
}

func (ms *MongoStorage) SetRankingProfileDrv(ctx *context.Context, sgp *utils.RankingProfile) (err error) {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColRgp).UpdateOne(sctx, bson.M{"tenant": sgp.Tenant, "id": sgp.ID},
			bson.M{"$set": sgp},
			options.Update().SetUpsert(true))
		return err
	})
}

func (ms *MongoStorage) RemRankingProfileDrv(ctx *context.Context, tenant, id string) (err error) {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColRgp).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})

}

func (ms *MongoStorage) GetRankingDrv(ctx *context.Context, tenant, id string) (*utils.Ranking, error) {
	rn := new(utils.Ranking)
	err := ms.query(ctx, func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColRnk).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(rn)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return rn, err
}
func (ms *MongoStorage) SetRankingDrv(ctx *context.Context, rn *utils.Ranking) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColRnk).UpdateOne(sctx, bson.M{"tenant": rn.Tenant, "id": rn.ID},
			bson.M{"$set": rn},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveRankingDrv(ctx *context.Context, tenant, id string) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColRnk).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetTrendProfileDrv(ctx *context.Context, tenant, id string) (*utils.TrendProfile, error) {
	srProfile := new(utils.TrendProfile)
	err := ms.query(ctx, func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColTrs).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(srProfile)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return srProfile, err
}

func (ms *MongoStorage) SetTrendProfileDrv(ctx *context.Context, srp *utils.TrendProfile) (err error) {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColTrs).UpdateOne(sctx, bson.M{"tenant": srp.Tenant, "id": srp.ID},
			bson.M{"$set": srp},
			options.Update().SetUpsert(true))
		return err
	})
}

func (ms *MongoStorage) RemTrendProfileDrv(ctx *context.Context, tenant, id string) (err error) {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColTrs).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetTrendDrv(ctx *context.Context, tenant, id string) (*utils.Trend, error) {
	tr := new(utils.Trend)
	err := ms.query(ctx, func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColTrd).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(tr)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return tr, err
}

func (ms *MongoStorage) SetTrendDrv(ctx *context.Context, tr *utils.Trend) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColTrd).UpdateOne(sctx, bson.M{"tenant": tr.Tenant, "id": tr.ID},
			bson.M{"$set": tr},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveTrendDrv(ctx *context.Context, tenant, id string) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColTrd).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

// GetStatQueueProfileDrv retrieves a StatQueueProfile from dataDB
func (ms *MongoStorage) GetStatQueueProfileDrv(ctx *context.Context, tenant string, id string) (*StatQueueProfile, error) {
	sqProfile := new(StatQueueProfile)
	err := ms.query(ctx, func(sctx mongo.SessionContext) error {
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
func (ms *MongoStorage) SetStatQueueProfileDrv(ctx *context.Context, sq *StatQueueProfile) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColSqp).UpdateOne(sctx, bson.M{"tenant": sq.Tenant, "id": sq.ID},
			bson.M{"$set": sq},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

// RemStatQueueProfileDrv removes a StatsQueue from dataDB
func (ms *MongoStorage) RemStatQueueProfileDrv(ctx *context.Context, tenant, id string) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColSqp).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

// GetStatQueueDrv retrieves a StoredStatQueue
func (ms *MongoStorage) GetStatQueueDrv(ctx *context.Context, tenant, id string) (*StatQueue, error) {
	ssq := new(StoredStatQueue)
	err := ms.query(ctx, func(sctx mongo.SessionContext) error {
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
func (ms *MongoStorage) SetStatQueueDrv(ctx *context.Context, ssq *StoredStatQueue, sq *StatQueue) (err error) {
	if ssq == nil {
		if ssq, err = NewStoredStatQueue(sq, ms.ms); err != nil {
			return err
		}
	}
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		_, err = ms.getCol(ColSqs).UpdateOne(sctx, bson.M{"tenant": ssq.Tenant, "id": ssq.ID},
			bson.M{"$set": ssq},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

// RemStatQueueDrv removes stored metrics for a StoredStatQueue
func (ms *MongoStorage) RemStatQueueDrv(ctx *context.Context, tenant, id string) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColSqs).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

// GetThresholdProfileDrv retrieves a ThresholdProfile from dataDB
func (ms *MongoStorage) GetThresholdProfileDrv(ctx *context.Context, tenant, ID string) (*ThresholdProfile, error) {
	thProfile := new(ThresholdProfile)
	err := ms.query(ctx, func(sctx mongo.SessionContext) error {
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
func (ms *MongoStorage) SetThresholdProfileDrv(ctx *context.Context, tp *ThresholdProfile) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColTps).UpdateOne(sctx, bson.M{"tenant": tp.Tenant, "id": tp.ID},
			bson.M{"$set": tp}, options.Update().SetUpsert(true),
		)
		return err
	})
}

// RemoveThresholdProfile removes a ThresholdProfile from dataDB/cache
func (ms *MongoStorage) RemThresholdProfileDrv(ctx *context.Context, tenant, id string) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColTps).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetThresholdDrv(ctx *context.Context, tenant, id string) (*Threshold, error) {
	th := new(Threshold)
	err := ms.query(ctx, func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColThs).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(th)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return th, err
}

func (ms *MongoStorage) SetThresholdDrv(ctx *context.Context, r *Threshold) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColThs).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveThresholdDrv(ctx *context.Context, tenant, id string) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColThs).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetFilterDrv(ctx *context.Context, tenant, id string) (*Filter, error) {
	fltr := new(Filter)
	err := ms.query(ctx, func(sctx mongo.SessionContext) error {
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

func (ms *MongoStorage) SetFilterDrv(ctx *context.Context, r *Filter) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColFlt).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveFilterDrv(ctx *context.Context, tenant, id string) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColFlt).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetRouteProfileDrv(ctx *context.Context, tenant, id string) (*RouteProfile, error) {
	routeProfile := new(RouteProfile)
	err := ms.query(ctx, func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColRts).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(routeProfile)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return routeProfile, err
}

func (ms *MongoStorage) SetRouteProfileDrv(ctx *context.Context, r *RouteProfile) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColRts).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveRouteProfileDrv(ctx *context.Context, tenant, id string) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColRts).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetAttributeProfileDrv(ctx *context.Context, tenant, id string) (*AttributeProfile, error) {
	attrProfile := new(AttributeProfile)
	err := ms.query(ctx, func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColAttr).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(attrProfile)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return attrProfile, err
}

func (ms *MongoStorage) SetAttributeProfileDrv(ctx *context.Context, r *AttributeProfile) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColAttr).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveAttributeProfileDrv(ctx *context.Context, tenant, id string) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColAttr).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetChargerProfileDrv(ctx *context.Context, tenant, id string) (*ChargerProfile, error) {
	chargerProfile := new(ChargerProfile)
	err := ms.query(ctx, func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColCpp).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(chargerProfile)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	return chargerProfile, err
}

func (ms *MongoStorage) SetChargerProfileDrv(ctx *context.Context, r *ChargerProfile) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColCpp).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveChargerProfileDrv(ctx *context.Context, tenant, id string) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColCpp).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetItemLoadIDsDrv(ctx *context.Context, itemIDPrefix string) (map[string]int64, error) {
	fop := options.FindOne()
	if itemIDPrefix != "" {
		fop.SetProjection(bson.M{itemIDPrefix: 1, "_id": 0})
	} else {
		fop.SetProjection(bson.M{"_id": 0})
	}
	loadIDs := make(map[string]int64)
	err := ms.query(ctx, func(sctx mongo.SessionContext) error {
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

func (ms *MongoStorage) SetLoadIDsDrv(ctx *context.Context, loadIDs map[string]int64) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColLID).UpdateOne(sctx, bson.D{}, bson.M{"$set": loadIDs},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveLoadIDsDrv() error {
	return ms.query(context.TODO(), func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColLID).DeleteMany(sctx, bson.M{})
		return err
	})
}

func (ms *MongoStorage) GetRateProfileDrv(ctx *context.Context, tenant, id string) (*utils.RateProfile, error) {
	mapRP := make(map[string]any)
	err := ms.query(ctx, func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColRpp).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(mapRP)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	if err != nil {
		return nil, err
	}
	return utils.NewRateProfileFromMapDataDBMap(tenant, id, mapRP, ms.ms)
}

func (ms *MongoStorage) GetRateProfileRatesDrv(ctx *context.Context, tenant, profileID, rtPrfx string, needIDs bool) (rateIDs []string, rates []*utils.Rate, err error) {
	prefix := utils.Rates + utils.ConcatenatedKeySep
	if rtPrfx != utils.EmptyString {
		prefix = utils.ConcatenatedKey(utils.Rates, rtPrfx)
	}

	matchStage, queryStage := newAggregateStages(profileID, tenant, prefix)
	var result []bson.M
	if err = ms.query(ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(ColRpp).Aggregate(sctx,
			mongo.Pipeline{matchStage, queryStage}, options.Aggregate().SetMaxTime(2*time.Second))
		if err != nil {
			return
		}
		if err = cur.All(sctx, &result); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return
	}
	for _, doc := range result {
		for key, rate := range doc {
			if needIDs {
				rateIDs = append(rateIDs, key[6:])
				continue
			}
			rtToAppend := new(utils.Rate)
			if err = ms.ms.Unmarshal([]byte(utils.IfaceAsString(rate)), rtToAppend); err != nil {
				return nil, nil, err
			}
			rates = append(rates, rtToAppend)
		}
	}
	return
}

func newAggregateStages(profileID, tenant, prefix string) (match, query bson.D) {
	match = bson.D{{
		Key: "$match", Value: bson.M{
			"id":     profileID,
			"tenant": tenant,
		}},
	}
	query = bson.D{{
		Key: "$replaceRoot", Value: bson.D{{
			Key: "newRoot", Value: bson.D{{
				Key: "$arrayToObject", Value: bson.D{{
					Key: "$filter", Value: bson.D{
						{Key: "input", Value: bson.M{
							"$objectToArray": "$$ROOT",
						}},
						{Key: "cond", Value: bson.D{{
							Key: "$regexFind", Value: bson.M{
								"input": "$$this.k",
								"regex": prefix,
							},
						}}},
					},
				}},
			}},
		}},
	}}
	return
}

func (ms *MongoStorage) SetRateProfileDrv(ctx *context.Context, rpp *utils.RateProfile, optOverwrite bool) error {
	rpMap, err := rpp.AsDataDBMap(ms.ms)
	if err != nil {
		return err
	}
	return ms.query(ctx, func(sctx mongo.SessionContext) (err error) {
		if optOverwrite {
			if _, err = ms.getCol(ColRpp).DeleteOne(sctx, bson.M{"tenant": rpp.Tenant, "id": rpp.ID}); err != nil {
				return
			}
		}
		_, err = ms.getCol(ColRpp).UpdateOne(sctx, bson.M{"tenant": rpp.Tenant, "id": rpp.ID},
			bson.M{"$set": rpMap},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveRateProfileDrv(ctx *context.Context, tenant, id string, rateIDs *[]string) (err error) {
	if rateIDs != nil {
		// if we want to remove just some rates from our profile, we will remove by their key Rates:rateID
		return ms.query(ctx, func(sctx mongo.SessionContext) (err error) {
			for _, rateID := range *rateIDs {
				_, err = ms.getCol(ColRpp).UpdateOne(ctx, bson.M{"tenant": tenant, "id": id}, bson.A{bson.M{"$unset": utils.Rates + utils.InInFieldSep + rateID}})
				if err != nil {
					return
				}
			}
			return
		})
	}
	return ms.query(ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(ColRpp).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetActionProfileDrv(ctx *context.Context, tenant, id string) (*ActionProfile, error) {
	ap := new(ActionProfile)
	err := ms.query(ctx, func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColApp).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(ap)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	if err != nil {
		return nil, err
	}
	return ap, nil
}

func (ms *MongoStorage) SetActionProfileDrv(ctx *context.Context, ap *ActionProfile) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColApp).UpdateOne(sctx, bson.M{"tenant": ap.Tenant, "id": ap.ID},
			bson.M{"$set": ap},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveActionProfileDrv(ctx *context.Context, tenant, id string) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColApp).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

// GetIndexesDrv retrieves Indexes from dataDB
// the key is the tenant of the item or in case of context dependent profiles is a concatenatedKey between tenant and context
// id is used as a concatenated key in case of filterIndexes the id will be filterType:fieldName:fieldVal
func (ms *MongoStorage) GetIndexesDrv(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (map[string]utils.StringSet, error) {
	type result struct {
		Key   string
		Value []string
	}
	originKey := utils.CacheInstanceToPrefix[idxItmType] + tntCtx
	if transactionID != utils.NonTransactional {
		originKey = "tmp_" + utils.ConcatenatedKey(originKey, transactionID)
	}
	dbKey := originKey
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
	err := ms.query(ctx, func(sctx mongo.SessionContext) (qryErr error) {
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
			if qryErr := cur.Decode(&elem); qryErr != nil {
				return qryErr
			}
			if len(elem.Value) == 0 {
				continue
			}
			indexKey := strings.TrimPrefix(elem.Key, originKey+utils.ConcatenatedKeySep)
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
func (ms *MongoStorage) SetIndexesDrv(ctx *context.Context, idxItmType, tntCtx string,
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
		err := ms.query(ctx, func(sctx mongo.SessionContext) error {
			result, qryErr := ms.getAllIndexKeys(sctx, regexKey)
			if qryErr != nil {
				return qryErr
			}
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
		err := ms.query(ctx, func(sctx mongo.SessionContext) (qryErr error) {
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
func (ms *MongoStorage) RemoveIndexesDrv(ctx *context.Context, idxItmType, tntCtx, idxKey string) error {
	if len(idxKey) != 0 {
		return ms.query(ctx, func(sctx mongo.SessionContext) error {
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
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColIndx).DeleteMany(sctx, bson.M{
			"key": primitive.Regex{
				Pattern: "^" + regexKey,
			},
		})
		return err
	})
}

func (ms *MongoStorage) GetAccountDrv(ctx *context.Context, tenant, id string) (*utils.Account, error) {
	ap := new(utils.Account)
	err := ms.query(ctx, func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColAnp).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		decodeErr := sr.Decode(ap)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	if err != nil {
		return nil, err
	}
	return ap, nil
}

func (ms *MongoStorage) SetAccountDrv(ctx *context.Context, ap *utils.Account) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColAnp).UpdateOne(sctx, bson.M{"tenant": ap.Tenant, "id": ap.ID},
			bson.M{"$set": ap},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveAccountDrv(ctx *context.Context, tenant, id string) error {
	return ms.query(ctx, func(sctx mongo.SessionContext) error {
		dr, err := ms.getCol(ColAnp).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetConfigSectionsDrv(ctx *context.Context, nodeID string, sectionIDs []string) (map[string][]byte, error) {
	sectionMap := make(map[string][]byte)
	for _, sectionID := range sectionIDs {
		err := ms.query(ctx, func(sctx mongo.SessionContext) error {
			cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{
				"nodeID":  nodeID,
				"section": sectionID,
			}, options.FindOne().SetProjection(bson.M{"cfgData": 1, "_id": 0}))
			cfgMap := make(map[string][]byte)
			decodeErr := cur.Decode(&cfgMap)
			if decodeErr != nil {
				if errors.Is(decodeErr, mongo.ErrNoDocuments) {
					return nil
				}
				return decodeErr
			}
			sectionMap[sectionID] = cfgMap["cfgData"]
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	if len(sectionMap) == 0 {
		return nil, utils.ErrNotFound
	}
	return sectionMap, nil
}

func (ms *MongoStorage) SetConfigSectionsDrv(ctx *context.Context, nodeID string, sectionsData map[string][]byte) error {
	for sectionID, sectionData := range sectionsData {
		err := ms.query(ctx, func(sctx mongo.SessionContext) error {
			_, qryErr := ms.getCol(ColCfg).UpdateOne(sctx, bson.M{
				"nodeID":  nodeID,
				"section": sectionID,
			}, bson.M{"$set": bson.M{
				"nodeID":  nodeID,
				"section": sectionID,
				"cfgData": sectionData}},
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

func (ms *MongoStorage) RemoveConfigSectionsDrv(ctx *context.Context, nodeID string, sectionIDs []string) error {
	for _, sectionID := range sectionIDs {
		err := ms.query(ctx, func(sctx mongo.SessionContext) error {
			_, err := ms.getCol(ColCfg).DeleteOne(sctx, bson.M{
				"nodeID":  nodeID,
				"section": sectionID,
			})
			return err
		})
		if err != nil {
			return err
		}
	}
	return nil
}
