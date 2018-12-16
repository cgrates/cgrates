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
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/bsoncodec"
	"github.com/mongodb/mongo-go-driver/bson/bsonrw"
	"github.com/mongodb/mongo-go-driver/bson/bsontype"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	"github.com/mongodb/mongo-go-driver/x/bsonx"
	"github.com/mongodb/mongo-go-driver/x/mongo/driver/topology"
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
	colDcs   = "derived_chargers"
	colAls   = "aliases"
	colRCfgs = "reverse_aliases"
	colPbs   = "pubsub"
	colUsr   = "users"
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
	colSpp   = "supplier_profiles"
	colAttr  = "attribute_profiles"
	ColCDRs  = "cdrs"
	colCpp   = "charger_profiles"
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
	DirectionLow       = strings.ToLower(utils.Direction)
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
func NewMongoStorage(host, port, db, user, pass, storageType string, cdrsIndexes []string,
	cacheCfg config.CacheCfg, isDataDB bool) (ms *MongoStorage, err error) {
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
	url = "mongodb://" + url
	reg := bson.NewRegistryBuilder().RegisterDecoder(tTime, bsoncodec.ValueDecoderFunc(TimeDecodeValue1)).Build()
	opt := &options.ClientOptions{
		Registry: reg,
		TopologyOptions: []topology.Option{
			topology.WithServerOptions(func(opts ...topology.ServerOption) []topology.ServerOption {
				return []topology.ServerOption{
					topology.WithRegistry(func(r *bsoncodec.Registry) *bsoncodec.Registry { return reg }),
				}
			}),
		},
	}

	client, err := mongo.NewClientWithOptions(url, opt)
	// client, err := mongo.NewClient(url)

	if err != nil {
		return nil, err
	}
	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}
	ms = &MongoStorage{
		client:      client,
		ctx:         ctx,
		db:          dbName,
		storageType: storageType,
		ms:          NewCodecMsgpackMarshaler(),
		cacheCfg:    cacheCfg,
		cdrsIndexes: cdrsIndexes,
		isDataDB:    isDataDB,
	}
	if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) error {
		if col, err := ms.client.Database(dbName).ListCollections(sctx, nil, options.ListCollections().SetNameOnly(true)); err != nil {
			return err
		} else {
			empty := true
			for col.Next(sctx) { // create indexes only if database is empty or only version table is present
				var elem struct{ Name string }
				err := col.Decode(&elem)
				if err != nil {
					return err
				}
				if elem.Name != colVer {
					empty = false
					break
				}
			}
			col.Close(sctx)
			if empty {
				if err = ms.EnsureIndexes(); err != nil {
					return err
				}
			}
			return nil
		}
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
	db          string
	storageType string // datadb, stordb
	ms          Marshaler
	cacheCfg    config.CacheCfg
	cdrsIndexes []string
	cnter       *utils.Counter
	isDataDB    bool
}

func (ms *MongoStorage) IsDataDB() bool {
	return ms.isDataDB
}

func (ms *MongoStorage) EnusureIndex(colName string, uniq bool, keys ...string) error {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) error {
		col := ms.getCol(colName)
		io := mongo.NewIndexOptionsBuilder().Unique(uniq)
		var doc bsonx.Doc
		for _, k := range keys {
			doc = doc.Append(k, bsonx.Int32(1))
		}
		_, err := col.Indexes().CreateOne(sctx, mongo.IndexModel{
			Keys:    doc,
			Options: io.Build(),
		})
		return err
	})
}

func (ms *MongoStorage) getCol(col string) *mongo.Collection {
	return ms.client.Database(ms.db).Collection(col)
}

func (ms *MongoStorage) GetContext() context.Context {
	return ms.ctx
}

// EnsureIndexes creates db indexes
func (ms *MongoStorage) EnsureIndexes() (err error) {
	if ms.storageType == utils.DataDB {
		for _, col := range []string{colAct, colApl, colAAp, colAtr,
			colDcs, colRpl, colDst, colRds, colAls, colUsr, colLht} {
			if err = ms.EnusureIndex(col, true, "key"); err != nil {
				return
			}
		}
		for _, col := range []string{colRsP, colRes, colSqs, colSqp,
			colTps, colThs, colSpp, colAttr, colFlt, colCpp} {
			if err = ms.EnusureIndex(col, true, "tenant", "id"); err != nil {
				return
			}
		}
		for _, col := range []string{colRpf, colShg, colAcc} {
			if err = ms.EnusureIndex(col, true, "id"); err != nil {
				return
			}
		}
	}
	if ms.storageType == utils.StorDB {
		for _, col := range []string{utils.TBLTPTimings, utils.TBLTPDestinations,
			utils.TBLTPDestinationRates, utils.TBLTPRatingPlans,
			utils.TBLTPSharedGroups, utils.TBLTPActions,
			utils.TBLTPActionPlans, utils.TBLTPActionTriggers,
			utils.TBLTPStats, utils.TBLTPResources} {
			if err = ms.EnusureIndex(col, true, "tpid", "id"); err != nil {
				return
			}
		}

		if err = ms.EnusureIndex(utils.TBLTPRateProfiles, true, "tpid", "direction",
			"tenant", "category", "subject", "loadid"); err != nil {
			return
		}

		if err = ms.EnusureIndex(utils.TBLTPUsers, true, "tpid", "tenant",
			"username"); err != nil {
			return
		}

		if err = ms.EnusureIndex(utils.TBLTPDerivedChargers, true, "tpid", "tenant",
			"category", "subject", "account", "loadid"); err != nil {
			return
		}

		if err = ms.EnusureIndex(utils.TBLTPDerivedChargers, true, "tpid", "direction",
			"tenant", "account", "loadid"); err != nil {
			return
		}

		if err = ms.EnusureIndex(utils.CDRsTBL, true, CGRIDLow, RunIDLow,
			OriginIDLow); err != nil {
			return
		}

		for _, idxKey := range ms.cdrsIndexes {
			if err = ms.EnusureIndex(utils.CDRsTBL, false, idxKey); err != nil {
				return
			}
		}

		if err = ms.EnusureIndex(utils.SessionsCostsTBL, true, CGRIDLow,
			RunIDLow); err != nil {
			return
		}

		if err = ms.EnusureIndex(utils.SessionsCostsTBL, false, OriginHostLow,
			OriginIDLow); err != nil {
			return
		}

		if err = ms.EnusureIndex(utils.SessionsCostsTBL, false, RunIDLow,
			OriginIDLow); err != nil {
			return
		}
	}
	return
}

func (ms *MongoStorage) getColNameForPrefix(prefix string) (string, bool) {
	res, ok := map[string]string{
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
		utils.DERIVEDCHARGERS_PREFIX:     colDcs,
		utils.ALIASES_PREFIX:             colAls,
		utils.REVERSE_ALIASES_PREFIX:     colRCfgs,
		utils.PUBSUB_SUBSCRIBERS_PREFIX:  colPbs,
		utils.USERS_PREFIX:               colUsr,
		utils.LOADINST_KEY:               colLht,
		utils.VERSION_PREFIX:             colVer,
		utils.TimingsPrefix:              colTmg,
		utils.ResourcesPrefix:            colRes,
		utils.ResourceProfilesPrefix:     colRsP,
		utils.ThresholdProfilePrefix:     colTps,
		utils.StatQueueProfilePrefix:     colSqp,
		utils.ThresholdPrefix:            colThs,
		utils.FilterPrefix:               colFlt,
		utils.SupplierProfilePrefix:      colSpp,
		utils.AttributeProfilePrefix:     colAttr,
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
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) error {
		return ms.client.Database(ms.db).Drop(sctx)
	})
}

// Marshaler returns the marshall
func (ms *MongoStorage) Marshaler() Marshaler {
	return ms.ms
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

// RebuildReverseForPrefix implementation
func (ms *MongoStorage) RebuildReverseForPrefix(prefix string) (err error) {
	if !utils.IsSliceMember([]string{utils.REVERSE_DESTINATION_PREFIX,
		utils.REVERSE_ALIASES_PREFIX, utils.AccountActionPlansPrefix}, prefix) {
		return utils.ErrInvalidKey
	}
	colName, ok := ms.getColNameForPrefix(prefix)
	if !ok {
		return utils.ErrInvalidKey
	}
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) error {
		col := ms.getCol(colName)
		if _, err := col.DeleteMany(sctx, bson.M{}); err != nil {
			return err
		}
		var keys []string
		switch prefix {
		case utils.REVERSE_DESTINATION_PREFIX:
			if keys, err = ms.GetKeysForPrefix(utils.DESTINATION_PREFIX); err != nil {
				return err
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
				return err
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
				return err
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
	})
}

// RemoveReverseForPrefix implementation
func (ms *MongoStorage) RemoveReverseForPrefix(prefix string) (err error) {
	if !utils.IsSliceMember([]string{utils.REVERSE_DESTINATION_PREFIX,
		utils.REVERSE_ALIASES_PREFIX, utils.AccountActionPlansPrefix}, prefix) {
		return utils.ErrInvalidKey
	}
	colName, ok := ms.getColNameForPrefix(prefix)
	if !ok {
		return utils.ErrInvalidKey
	}
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) error {
		col := ms.getCol(colName)

		if dr, err := col.DeleteMany(sctx, bson.M{}); err != nil {
			return err
		} else if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}

		var keys []string
		switch prefix {
		case utils.REVERSE_DESTINATION_PREFIX:
			if keys, err = ms.GetKeysForPrefix(utils.DESTINATION_PREFIX); err != nil {
				return err
			}
			for _, key := range keys {
				dest, err := ms.GetDestination(key[len(utils.DESTINATION_PREFIX):], true, utils.NonTransactional)
				if err != nil {
					return err
				}
				if err := ms.RemoveDestination(dest.Id, utils.NonTransactional); err != nil {
					return err
				}
			}
		case utils.REVERSE_ALIASES_PREFIX:
			if keys, err = ms.GetKeysForPrefix(utils.ALIASES_PREFIX); err != nil {
				return err
			}
			for _, key := range keys {
				al, err := ms.GetAlias(key[len(utils.ALIASES_PREFIX):], true, utils.NonTransactional)
				if err != nil {
					return err
				}
				if err := ms.RemoveAlias(al.GetId(), utils.NonTransactional); err != nil {
					return err
				}
			}
		case utils.AccountActionPlansPrefix:
			if keys, err = ms.GetKeysForPrefix(utils.ACTION_PLAN_PREFIX); err != nil {
				return err
			}
			for _, key := range keys {
				apl, err := ms.GetActionPlan(key[len(utils.ACTION_PLAN_PREFIX):], true, utils.NonTransactional)
				if err != nil {
					return err
				}
				for acntID := range apl.AccountIDs {
					if err = ms.RemAccountActionPlans(acntID, []string{apl.Id}); err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
}

// IsDBEmpty implementation
func (ms *MongoStorage) IsDBEmpty() (resp bool, err error) {
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) error {
		col, err := ms.DB().ListCollections(sctx, nil)
		if err != nil {
			return err
		}
		if resp = !col.Next(sctx); resp {
			return nil
		}
		elem := bson.D{}
		err = col.Decode(&elem)
		if err != nil {
			return err
		}
		resp = (elem.Map()["name"] == "cdrs")
		col.Close(sctx)
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

	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		switch category {
		case utils.DESTINATION_PREFIX:
			result, err = ms.getField(sctx, colDst, utils.DESTINATION_PREFIX, subject, "key")
		case utils.REVERSE_DESTINATION_PREFIX:
			result, err = ms.getField(sctx, colRds, utils.REVERSE_DESTINATION_PREFIX, subject, "key")
		case utils.RATING_PLAN_PREFIX:
			result, err = ms.getField(sctx, colRpl, utils.RATING_PLAN_PREFIX, subject, "key")
		case utils.RATING_PROFILE_PREFIX:
			result, err = ms.getField(sctx, colRpf, utils.RATING_PROFILE_PREFIX, subject, "id")
		case utils.ACTION_PREFIX:
			result, err = ms.getField(sctx, colAct, utils.ACTION_PREFIX, subject, "key")
		case utils.ACTION_PLAN_PREFIX:
			result, err = ms.getField(sctx, colApl, utils.ACTION_PLAN_PREFIX, subject, "key")
		case utils.ACTION_TRIGGER_PREFIX:
			result, err = ms.getField(sctx, colAtr, utils.ACTION_TRIGGER_PREFIX, subject, "key")
		case utils.SHARED_GROUP_PREFIX:
			result, err = ms.getField(sctx, colShg, utils.SHARED_GROUP_PREFIX, subject, "id")
		case utils.DERIVEDCHARGERS_PREFIX:
			result, err = ms.getField(sctx, colDcs, utils.DERIVEDCHARGERS_PREFIX, subject, "key")
		case utils.ACCOUNT_PREFIX:
			result, err = ms.getField(sctx, colAcc, utils.ACCOUNT_PREFIX, subject, "id")
		case utils.ALIASES_PREFIX:
			result, err = ms.getField(sctx, colAls, utils.ALIASES_PREFIX, subject, "key")
		case utils.REVERSE_ALIASES_PREFIX:
			result, err = ms.getField(sctx, colRCfgs, utils.REVERSE_ALIASES_PREFIX, subject, "key")
		case utils.ResourceProfilesPrefix:
			result, err = ms.getField2(sctx, colRsP, utils.ResourceProfilesPrefix, subject, tntID)
		case utils.ResourcesPrefix:
			result, err = ms.getField2(sctx, colRes, utils.ResourcesPrefix, subject, tntID)
		case utils.StatQueuePrefix:
			result, err = ms.getField2(sctx, colSqs, utils.StatQueuePrefix, subject, tntID)
		case utils.StatQueueProfilePrefix:
			result, err = ms.getField2(sctx, colSqp, utils.StatQueueProfilePrefix, subject, tntID)
		case utils.AccountActionPlansPrefix:
			result, err = ms.getField(sctx, colAAp, utils.AccountActionPlansPrefix, subject, "key")
		case utils.TimingsPrefix:
			result, err = ms.getField(sctx, colTmg, utils.TimingsPrefix, subject, "id")
		case utils.FilterPrefix:
			result, err = ms.getField2(sctx, colFlt, utils.FilterPrefix, subject, tntID)
		case utils.ThresholdPrefix:
			result, err = ms.getField2(sctx, colThs, utils.ThresholdPrefix, subject, tntID)
		case utils.ThresholdProfilePrefix:
			result, err = ms.getField2(sctx, colTps, utils.ThresholdProfilePrefix, subject, tntID)
		case utils.SupplierProfilePrefix:
			result, err = ms.getField2(sctx, colSpp, utils.SupplierProfilePrefix, subject, tntID)
		case utils.AttributeProfilePrefix:
			result, err = ms.getField2(sctx, colAttr, utils.AttributeProfilePrefix, subject, tntID)
		case utils.ChargerProfilePrefix:
			result, err = ms.getField2(sctx, colCpp, utils.ChargerProfilePrefix, subject, tntID)
		default:
			err = fmt.Errorf("unsupported prefix in GetKeysForPrefix: %s", prefix)
		}
		return err
	})
	return
}

func (ms *MongoStorage) HasDataDrv(category, subject, tenant string) (has bool, err error) {
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		var count int64
		switch category {
		case utils.DESTINATION_PREFIX:
			count, err = ms.getCol(colDst).Count(sctx, bson.M{"key": subject})
		case utils.RATING_PLAN_PREFIX:
			count, err = ms.getCol(colRpl).Count(sctx, bson.M{"key": subject})
		case utils.RATING_PROFILE_PREFIX:
			count, err = ms.getCol(colRpf).Count(sctx, bson.M{"key": subject})
		case utils.ACTION_PREFIX:
			count, err = ms.getCol(colAct).Count(sctx, bson.M{"key": subject})
		case utils.ACTION_PLAN_PREFIX:
			count, err = ms.getCol(colApl).Count(sctx, bson.M{"key": subject})
		case utils.ACCOUNT_PREFIX:
			count, err = ms.getCol(colAcc).Count(sctx, bson.M{"id": subject})
		case utils.ResourcesPrefix:
			count, err = ms.getCol(colRes).Count(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.ResourceProfilesPrefix:
			count, err = ms.getCol(colRsP).Count(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.StatQueuePrefix:
			count, err = ms.getCol(colSqs).Count(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.StatQueueProfilePrefix:
			count, err = ms.getCol(colSqp).Count(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.ThresholdPrefix:
			count, err = ms.getCol(colTps).Count(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.FilterPrefix:
			count, err = ms.getCol(colFlt).Count(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.SupplierProfilePrefix:
			count, err = ms.getCol(colSpp).Count(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.AttributeProfilePrefix:
			count, err = ms.getCol(colAttr).Count(sctx, bson.M{"tenant": tenant, "id": subject})
		case utils.ChargerProfilePrefix:
			count, err = ms.getCol(colCpp).Count(sctx, bson.M{"tenant": tenant, "id": subject})
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
	if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colRpl).FindOne(sctx, bson.M{"key": key})
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
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colRpl).UpdateOne(sctx, bson.M{"key": rp.Id},
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
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colRpl).DeleteMany(sctx, bson.M{"key": key})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetRatingProfileDrv(key string) (rp *RatingProfile, err error) {
	rp = new(RatingProfile)
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colRpf).FindOne(sctx, bson.M{"id": key})
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
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colRpf).UpdateOne(sctx, bson.M{"id": rp.Id},
			bson.M{"$set": rp},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveRatingProfileDrv(key string) error {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colRpf).DeleteMany(sctx, bson.M{"id": key})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetDestination(key string, skipCache bool,
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
	if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colDst).FindOne(sctx, bson.M{"key": key})
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

func (ms *MongoStorage) SetDestination(dest *Destination, transactionID string) (err error) {
	result, err := ms.ms.Marshal(dest)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()

	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colDst).UpdateOne(sctx, bson.M{"key": dest.Id},
			bson.M{"$set": struct {
				Key   string
				Value []byte
			}{Key: dest.Id, Value: b.Bytes()}},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveDestination(destID string,
	transactionID string) (err error) {
	// get destination for prefix list
	d, err := ms.GetDestination(destID, false, transactionID)
	if err != nil {
		return
	}
	if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colDst).DeleteOne(sctx, bson.M{"key": destID})
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
		if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
			_, err = ms.getCol(colRds).UpdateOne(sctx, bson.M{"key": prefix},
				bson.M{"$pull": bson.M{"value": destID}})
			return err
		}); err != nil {
			return err
		}
		ms.GetReverseDestination(prefix, true, transactionID) // it will recache the destination
	}
	return
}

func (ms *MongoStorage) GetReverseDestination(prefix string, skipCache bool,
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
	if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colRds).FindOne(sctx, bson.M{"key": prefix})
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

func (ms *MongoStorage) SetReverseDestination(dest *Destination,
	transactionID string) (err error) {
	for _, p := range dest.Prefixes {
		if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
			_, err = ms.getCol(colRds).UpdateOne(sctx, bson.M{"key": p},
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

func (ms *MongoStorage) UpdateReverseDestination(oldDest, newDest *Destination,
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
		if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
			_, err = ms.getCol(colRds).UpdateOne(sctx, bson.M{"key": obsoletePrefix},
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
		if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
			_, err = ms.getCol(colRds).UpdateOne(sctx, bson.M{"key": addedPrefix},
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
	if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colAct).FindOne(sctx, bson.M{"key": key})
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
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colAct).UpdateOne(sctx, bson.M{"key": key},
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
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colAct).DeleteOne(sctx, bson.M{"key": key})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetSharedGroupDrv(key string) (sg *SharedGroup, err error) {
	sg = new(SharedGroup)
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colShg).FindOne(sctx, bson.M{"id": key})
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
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colShg).UpdateOne(sctx, bson.M{"id": sg.Id},
			bson.M{"$set": sg},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveSharedGroupDrv(id, transactionID string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colShg).DeleteOne(sctx, bson.M{"id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetAccount(key string) (result *Account, err error) {
	result = new(Account)
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colAcc).FindOne(sctx, bson.M{"id": key})
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
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colAcc).UpdateOne(sctx, bson.M{"id": acc.ID},
			bson.M{"$set": acc},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveAccount(key string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colAcc).DeleteOne(sctx, bson.M{"id": key})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetSubscribersDrv() (result map[string]*SubscriberData, err error) {
	result = make(map[string]*SubscriberData)
	var kv struct {
		Key   string
		Value *SubscriberData
	}
	if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(colPbs).Find(sctx, nil)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			err := cur.Decode(&kv)
			if err != nil {
				return err
			}
			result[kv.Key] = kv.Value
		}
		return cur.Close(sctx)
	}); err != nil {
		return nil, err
	}
	return
}

func (ms *MongoStorage) SetSubscriberDrv(key string, sub *SubscriberData) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colPbs).UpdateOne(sctx, bson.M{"key": key},
			bson.M{"$set": struct {
				Key   string
				Value *SubscriberData
			}{Key: key, Value: sub}},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveSubscriberDrv(key string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colPbs).DeleteOne(sctx, bson.M{"key": key})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetUserDrv(key string) (up *UserProfile, err error) {
	var kv struct {
		Key   string
		Value *UserProfile
	}
	if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colUsr).FindOne(sctx, bson.M{"key": key})
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
	return kv.Value, nil
}

func (ms *MongoStorage) GetUsersDrv() (result []*UserProfile, err error) {
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(colUsr).Find(sctx, nil)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var kv struct {
				Key   string
				Value *UserProfile
			}
			err := cur.Decode(&kv)
			if err != nil {
				return err
			}
			result = append(result, kv.Value)
		}
		return cur.Close(sctx)
	})
	return
}

func (ms *MongoStorage) SetUserDrv(up *UserProfile) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colUsr).UpdateOne(sctx, bson.M{"key": up.GetId()},
			bson.M{"$set": struct {
				Key   string
				Value *UserProfile
			}{Key: up.GetId(), Value: up}},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveUserDrv(key string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colUsr).DeleteOne(sctx, bson.M{"key": key})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetAlias(key string, skipCache bool,
	transactionID string) (al *Alias, err error) {
	if !skipCache {
		if x, ok := Cache.Get(utils.CacheAliases, key); ok {
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
	cCommit := cacheCommit(transactionID)

	if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colAls).FindOne(sctx, bson.M{"key": key})
		if err := cur.Decode(&kv); err != nil {
			if err == mongo.ErrNoDocuments {
				Cache.Set(utils.CacheAliases, key, nil, nil,
					cacheCommit(transactionID), transactionID)
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	al = &Alias{Values: kv.Value}
	al.SetId(key)
	Cache.Set(utils.CacheAliases, key, al, nil,
		cCommit, transactionID)
	return
}

func (ms *MongoStorage) SetAlias(al *Alias, transactionID string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colAls).UpdateOne(sctx, bson.M{"key": al.GetId()},
			bson.M{"$set": struct {
				Key   string
				Value AliasValues
			}{Key: al.GetId(), Value: al.Values}},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveAlias(key, transactionID string) (err error) {
	al := new(Alias)
	al.SetId(key)
	origKey := key
	key = utils.ALIASES_PREFIX + key
	var kv struct {
		Key   string
		Value AliasValues
	}
	if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colAls).FindOne(sctx, bson.M{"key": origKey})
		if err := cur.Decode(&kv); err != nil {
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	al.Values = kv.Value
	if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colAls).DeleteOne(sctx, bson.M{"key": origKey})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	}); err != nil {
		return err
	}
	cCommit := cacheCommit(transactionID)
	Cache.Remove(utils.CacheAliases, key, cCommit, transactionID)

	for _, value := range al.Values {
		tmpKey := utils.ConcatenatedKey(al.GetId(), value.DestinationId)
		for target, pairs := range value.Pairs {
			for _, alias := range pairs {
				rKey := alias + target + al.Context
				if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
					_, err = ms.getCol(colAls).UpdateOne(sctx, bson.M{"key": rKey},
						bson.M{"$pull": bson.M{"value": tmpKey}})
					return err
				}); err != nil {
					return err
				}
				Cache.Remove(utils.CacheReverseAliases, rKey, cCommit, transactionID)
			}
		}
	}
	return
}

func (ms *MongoStorage) GetReverseAlias(reverseID string, skipCache bool,
	transactionID string) (ids []string, err error) {
	if !skipCache {
		if x, ok := Cache.Get(utils.CacheReverseAliases, reverseID); ok {
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
	if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colRCfgs).FindOne(sctx, bson.M{"key": reverseID})
		if err := cur.Decode(&result); err != nil {
			if err == mongo.ErrNoDocuments {
				Cache.Set(utils.CacheReverseAliases, reverseID, nil, nil,
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
	Cache.Set(utils.CacheReverseAliases, reverseID, ids, nil,
		cacheCommit(transactionID), transactionID)
	return
}

func (ms *MongoStorage) SetReverseAlias(al *Alias, transactionID string) (err error) {
	for _, value := range al.Values {
		for target, pairs := range value.Pairs {
			for _, alias := range pairs {
				rKey := strings.Join([]string{alias, target, al.Context}, "")
				id := utils.ConcatenatedKey(al.GetId(), value.DestinationId)
				if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
					_, err = ms.getCol(colRCfgs).UpdateOne(sctx, bson.M{"key": rKey},
						bson.M{"$addToSet": bson.M{"value": id}},
						options.Update().SetUpsert(true),
					)
					return err
				}); err != nil {
					return err
				}
			}
		}
	}
	return
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
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colLht).FindOne(sctx, bson.M{"key": utils.LOADINST_KEY})
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
	if err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colLht).FindOne(sctx, bson.M{"key": utils.LOADINST_KEY})
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

		return nil, ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
			_, err = ms.getCol(colLht).UpdateOne(sctx, bson.M{"key": utils.LOADINST_KEY},
				bson.M{"$set": struct {
					Key   string
					Value []*utils.LoadInstance
				}{Key: utils.LOADINST_KEY, Value: existingLoadHistory}},
				options.Update().SetUpsert(true),
			)
			return err
		})
	}, 0, utils.LOADINST_KEY)

	Cache.Remove(utils.LOADINST_KEY, "",
		cacheCommit(transactionID), transactionID)
	return err
}

func (ms *MongoStorage) GetActionTriggersDrv(key string) (atrs ActionTriggers, err error) {
	var kv struct {
		Key   string
		Value ActionTriggers
	}
	if err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colAtr).FindOne(sctx, bson.M{"key": key})
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
		return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
			_, err = ms.getCol(colAtr).DeleteOne(sctx, bson.M{"key": key})
			return err
		})
	}
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colAtr).UpdateOne(sctx, bson.M{"key": key},
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
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colAtr).DeleteOne(sctx, bson.M{"key": key})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetActionPlan(key string, skipCache bool,
	transactionID string) (ats *ActionPlan, err error) {
	if !skipCache {
		if x, err := Cache.GetCloned(utils.CacheActionPlans, key); err != nil {
			if err != ltcache.ErrNotFound { // Only consider cache if item was found
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
	if err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colApl).FindOne(sctx, bson.M{"key": key})
		if err := cur.Decode(&kv); err != nil {
			if err == mongo.ErrNoDocuments {
				Cache.Set(utils.CacheActionPlans, key, nil, nil,
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
	if err = ms.ms.Unmarshal(out, &ats); err != nil {
		return nil, err
	}
	Cache.Set(utils.CacheActionPlans, key, ats, nil,
		cacheCommit(transactionID), transactionID)
	return
}

func (ms *MongoStorage) SetActionPlan(key string, ats *ActionPlan,
	overwrite bool, transactionID string) (err error) {
	// clean dots from account ids map
	cCommit := cacheCommit(transactionID)
	if len(ats.ActionTimings) == 0 {
		err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
			_, err = ms.getCol(colApl).DeleteOne(sctx, bson.M{"key": key})
			return err
		})
		Cache.Remove(utils.CacheActionPlans, key,
			cCommit, transactionID)
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

	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colApl).UpdateOne(sctx, bson.M{"key": key},
			bson.M{"$set": struct {
				Key   string
				Value []byte
			}{Key: key, Value: b.Bytes()}},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveActionPlan(key string, transactionID string) error {
	cCommit := cacheCommit(transactionID)
	Cache.Remove(utils.CacheActionPlans, key, cCommit, transactionID)
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colApl).DeleteOne(sctx, bson.M{"key": key})
		return err
	})
}

func (ms *MongoStorage) GetAllActionPlans() (ats map[string]*ActionPlan, err error) {
	keys, err := ms.GetKeysForPrefix(utils.ACTION_PLAN_PREFIX)
	if err != nil {
		return nil, err
	}
	ats = make(map[string]*ActionPlan, len(keys))
	for _, key := range keys {
		ap, err := ms.GetActionPlan(key[len(utils.ACTION_PLAN_PREFIX):],
			false, utils.NonTransactional)
		if err != nil {
			return nil, err
		}
		ats[key[len(utils.ACTION_PLAN_PREFIX):]] = ap
	}
	return
}

func (ms *MongoStorage) GetAccountActionPlans(acntID string, skipCache bool, transactionID string) (aPlIDs []string, err error) {
	if !skipCache {
		if x, ok := Cache.Get(utils.CacheAccountActionPlans, acntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.([]string), nil
		}
	}
	var kv struct {
		Key   string
		Value []string
	}
	if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colAAp).FindOne(sctx, bson.M{"key": acntID})
		if err := cur.Decode(&kv); err != nil {
			if err == mongo.ErrNoDocuments {
				Cache.Set(utils.CacheAccountActionPlans, acntID, nil, nil,
					cacheCommit(transactionID), transactionID)
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	aPlIDs = kv.Value
	Cache.Set(utils.CacheAccountActionPlans, acntID, aPlIDs, nil,
		cacheCommit(transactionID), transactionID)
	return
}

func (ms *MongoStorage) SetAccountActionPlans(acntID string, aPlIDs []string, overwrite bool) (err error) {
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

	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colAAp).UpdateOne(sctx, bson.M{"key": acntID},
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
func (ms *MongoStorage) RemAccountActionPlans(acntID string, aPlIDs []string) (err error) {
	if len(aPlIDs) == 0 {
		return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
			dr, err := ms.getCol(colAAp).DeleteOne(sctx, bson.M{"key": acntID})
			if dr.DeletedCount == 0 {
				return utils.ErrNotFound
			}
			return err
		})
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
		return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
			dr, err := ms.getCol(colAAp).DeleteOne(sctx, bson.M{"key": acntID})
			if dr.DeletedCount == 0 {
				return utils.ErrNotFound
			}
			return err
		})
	}

	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colAAp).UpdateOne(sctx, bson.M{"key": acntID},
			bson.M{"$set": struct {
				Key   string
				Value []string
			}{Key: acntID, Value: oldAPlIDs}},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) PushTask(t *Task) error {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(colTsk).InsertOne(sctx, bson.M{"_id": primitive.NewObjectID(), "task": t})
		return err
	})
}

func (ms *MongoStorage) PopTask() (t *Task, err error) {
	v := struct {
		ID   primitive.ObjectID `bson:"_id"`
		Task *Task
	}{}
	if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colTsk).FindOneAndDelete(sctx, nil)
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

func (ms *MongoStorage) GetDerivedChargersDrv(key string) (dcs *utils.DerivedChargers, err error) {
	var kv struct {
		Key   string
		Value *utils.DerivedChargers
	}
	if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colDcs).FindOne(sctx, bson.M{"key": key})
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
	return kv.Value, nil
}

func (ms *MongoStorage) SetDerivedChargers(key string,
	dcs *utils.DerivedChargers, transactionID string) (err error) {
	cCommit := cacheCommit(transactionID)
	if dcs == nil || len(dcs.Chargers) == 0 {
		if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
			_, err = ms.getCol(colDcs).DeleteOne(sctx, bson.M{"key": key})
			return err
		}); err != nil {
			return err
		}
		Cache.Remove(utils.CacheDerivedChargers, key, cCommit, transactionID)
		return nil
	}
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colDcs).UpdateOne(sctx, bson.M{"key": key},
			bson.M{"$set": struct {
				Key   string
				Value *utils.DerivedChargers
			}{Key: key, Value: dcs}},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveDerivedChargersDrv(id, transactionID string) (err error) {
	cCommit := cacheCommit(transactionID)
	if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colDcs).DeleteOne(sctx, bson.M{"key": id})
		return err
	}); err != nil {
		return err
	}
	Cache.Remove(utils.CacheDerivedChargers, id, cCommit, transactionID)
	return nil
}

func (ms *MongoStorage) GetResourceProfileDrv(tenant, id string) (rp *ResourceProfile, err error) {
	rp = new(ResourceProfile)
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colRsP).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
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
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colRsP).UpdateOne(sctx, bson.M{"tenant": rp.Tenant, "id": rp.ID},
			bson.M{"$set": rp},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveResourceProfileDrv(tenant, id string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colRsP).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetResourceDrv(tenant, id string) (r *Resource, err error) {
	r = new(Resource)
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colRes).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
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
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colRes).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveResourceDrv(tenant, id string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colRes).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetTimingDrv(id string) (t *utils.TPTiming, err error) {
	t = new(utils.TPTiming)
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colTmg).FindOne(sctx, bson.M{"id": id})
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
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colTmg).UpdateOne(sctx, bson.M{"id": t.ID},
			bson.M{"$set": t},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveTimingDrv(id string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colTmg).DeleteOne(sctx, bson.M{"id": id})
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
			if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
				cur, err := ms.getCol(colRFI).Find(sctx, bson.M{"key": utils.ConcatenatedKey(dbKey, filterType, fldName, fldValue)})
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
		if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
			cur, err := ms.getCol(colRFI).Find(sctx, bson.M{"key": bsonx.Regex("^"+dbKey, "")})
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
		if _, hasIt := indexes[indexKey]; !hasIt {
			indexes[indexKey] = make(utils.StringMap)
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
		if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
			_, err = ms.getCol(colRFI).DeleteMany(sctx, bson.M{"key": bsonx.Regex("^"+regexKey, "")})
			return err
		}); err != nil {
			return err
		}
		var lastErr error
		for key, itmMp := range indexes {
			if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
				_, err = ms.getCol(colRFI).UpdateOne(sctx, bson.M{"key": utils.ConcatenatedKey(originKey, key)},
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
		return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
			_, err = ms.getCol(colRFI).DeleteMany(sctx, bson.M{"key": bsonx.Regex("^"+oldKey, "")})
			return err
		})
	} else {
		var lastErr error
		for key, itmMp := range indexes {
			if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
				var action bson.M
				if len(itmMp) == 0 {
					action = bson.M{"$unset": bson.M{"value": 1}}
				} else {
					action = bson.M{"$set": bson.M{"key": utils.ConcatenatedKey(dbKey, key), "value": itmMp.Slice()}}
				}
				_, err = ms.getCol(colRFI).UpdateOne(sctx, bson.M{"key": utils.ConcatenatedKey(dbKey, key)},
					action, options.Update().SetUpsert(true),
				)
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
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colRFI).DeleteMany(sctx, bson.M{"key": bsonx.Regex("^"+regexKey, "")})
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
	if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colRFI).FindOne(sctx, bson.M{"key": utils.ConcatenatedKey(dbKey, filterType, fldName, fldVal)})
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
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colSqp).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
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
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colSqp).UpdateOne(sctx, bson.M{"tenant": sq.Tenant, "id": sq.ID},
			bson.M{"$set": sq},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

// RemStatQueueProfileDrv removes a StatsQueue from dataDB
func (ms *MongoStorage) RemStatQueueProfileDrv(tenant, id string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colSqp).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

// GetStoredStatQueueDrv retrieves a StoredStatQueue
func (ms *MongoStorage) GetStoredStatQueueDrv(tenant, id string) (sq *StoredStatQueue, err error) {
	sq = new(StoredStatQueue)
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colSqs).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
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

// SetStoredStatQueueDrv stores the metrics for a StoredStatQueue
func (ms *MongoStorage) SetStoredStatQueueDrv(sq *StoredStatQueue) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colSqs).UpdateOne(sctx, bson.M{"tenant": sq.Tenant, "id": sq.ID},
			bson.M{"$set": sq},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

// RemStoredStatQueueDrv removes stored metrics for a StoredStatQueue
func (ms *MongoStorage) RemStoredStatQueueDrv(tenant, id string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colSqs).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

// GetThresholdProfileDrv retrieves a ThresholdProfile from dataDB
func (ms *MongoStorage) GetThresholdProfileDrv(tenant, ID string) (tp *ThresholdProfile, err error) {
	tp = new(ThresholdProfile)
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colTps).FindOne(sctx, bson.M{"tenant": tenant, "id": ID})
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
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colTps).UpdateOne(sctx, bson.M{"tenant": tp.Tenant, "id": tp.ID},
			bson.M{"$set": tp}, options.Update().SetUpsert(true),
		)
		return err
	})
}

// RemoveThresholdProfile removes a ThresholdProfile from dataDB/cache
func (ms *MongoStorage) RemThresholdProfileDrv(tenant, id string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colTps).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetThresholdDrv(tenant, id string) (r *Threshold, err error) {
	r = new(Threshold)
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colThs).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
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
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colThs).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveThresholdDrv(tenant, id string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colThs).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetFilterDrv(tenant, id string) (r *Filter, err error) {
	r = new(Filter)
	if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colFlt).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
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
	for _, fltr := range r.Rules {
		if err = fltr.CompileValues(); err != nil {
			return
		}
	}
	return
}

func (ms *MongoStorage) SetFilterDrv(r *Filter) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colFlt).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveFilterDrv(tenant, id string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colFlt).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetSupplierProfileDrv(tenant, id string) (r *SupplierProfile, err error) {
	r = new(SupplierProfile)
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colSpp).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
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
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colSpp).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveSupplierProfileDrv(tenant, id string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colSpp).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetAttributeProfileDrv(tenant, id string) (r *AttributeProfile, err error) {
	r = new(AttributeProfile)
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colAttr).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
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
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colAttr).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveAttributeProfileDrv(tenant, id string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colAttr).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) GetChargerProfileDrv(tenant, id string) (r *ChargerProfile, err error) {
	r = new(ChargerProfile)
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colCpp).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
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
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colCpp).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorage) RemoveChargerProfileDrv(tenant, id string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colCpp).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}
