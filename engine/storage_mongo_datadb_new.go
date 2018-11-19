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
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/objectid"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/options"
	"github.com/mongodb/mongo-go-driver/x/bsonx"
)

// NewMongoStorageNew givese new mongo driver
func NewMongoStorageNew(host, port, db, user, pass, storageType string,
	cdrsIndexes []string, cacheCfg config.CacheCfg) (ms *MongoStorageNew, err error) {
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

	client, err := mongo.NewClient(url)
	if err != nil {
		return nil, err
	}
	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}
	ms = &MongoStorageNew{
		client:      client,
		ctx:         ctx,
		db:          dbName,
		storageType: storageType,
		ms:          NewCodecMsgpackMarshaler(),
		cacheCfg:    cacheCfg,
		cdrsIndexes: cdrsIndexes,
	}
	ms.cnter = utils.NewCounter(time.Now().UnixNano(), 0)
	return
}

// MongoStorageNew struct for new mongo driver
type MongoStorageNew struct {
	client      *mongo.Client
	ctx         context.Context
	db          string
	storageType string // datadb, stordb

	ms          Marshaler
	cacheCfg    config.CacheCfg
	cdrsIndexes []string
	cnter       *utils.Counter
}

func (ms *MongoStorageNew) getCol(col string) *mongo.Collection {
	return ms.client.Database(ms.db).Collection(col)
}

func (ms *MongoStorageNew) getColNameForPrefix(prefix string) (string, bool) {
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
		utils.StatQueueProfilePrefix: colSqp,
		utils.ThresholdPrefix:        colThs,
		utils.FilterPrefix:           colFlt,
		utils.SupplierProfilePrefix:  colSpp,
		utils.AttributeProfilePrefix: colAttr,
	}[prefix]
	return res, ok
}

// Close disconects the client
func (ms *MongoStorageNew) Close() {
	if err := ms.client.Disconnect(ms.ctx); err != nil {
		utils.Logger.Err(fmt.Sprintf("<MongoStorage> Error on disconect:%s", err))
	}
}

// Flush drops the datatable
func (ms *MongoStorageNew) Flush(ignore string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) error {
		return ms.client.Database(ms.db).Drop(sctx)
	})
}

// Marshaler returns the marshall
func (ms *MongoStorageNew) Marshaler() Marshaler {
	return ms.ms
}

// DB returnes a database object
func (ms *MongoStorageNew) DB() *mongo.Database {
	return ms.client.Database(ms.db)
}

// SelectDatabase selects the database
func (ms *MongoStorageNew) SelectDatabase(dbName string) (err error) {
	ms.db = dbName
	return
}

// RebuildReverseForPrefix implementation
func (ms *MongoStorageNew) RebuildReverseForPrefix(prefix string) (err error) {
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
func (ms *MongoStorageNew) RemoveReverseForPrefix(prefix string) (err error) {
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
func (ms *MongoStorageNew) IsDBEmpty() (resp bool, err error) {
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) error {
		col, err := ms.DB().ListCollections(sctx, bson.D{})
		if err != nil {
			return err
		}
		resp = !col.Next(sctx)
		elem := bson.D{}
		err = col.Decode(&elem)
		if err != nil {
			return err
		}
		resp = resp || (elem.Map()["name"] == "cdrs")
		col.Close(sctx)
		return nil
	})
	return resp, err
}

func (ms *MongoStorageNew) getField(sctx mongo.SessionContext, col, prefix, subject, field string) (result []string, err error) {
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
func (ms *MongoStorageNew) getField2(sctx mongo.SessionContext, col, prefix, subject string, tntID *utils.TenantID) (result []string, err error) {
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
func (ms *MongoStorageNew) GetKeysForPrefix(prefix string) (result []string, err error) {
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
		case utils.LCR_PREFIX:
			result, err = ms.getField(sctx, colLcr, utils.LCR_PREFIX, subject, "key")
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

func (ms *MongoStorageNew) HasDataDrv(category, subject, tenant string) (has bool, err error) {
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

func (ms *MongoStorageNew) GetRatingPlanDrv(key string) (rp *RatingPlan, err error) {
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

func (ms *MongoStorageNew) SetRatingPlanDrv(rp *RatingPlan) error {
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

func (ms *MongoStorageNew) RemoveRatingPlanDrv(key string) error {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colRpl).DeleteMany(sctx, bson.M{"key": key})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorageNew) GetRatingProfileDrv(key string) (rp *RatingProfile, err error) {
	rp = new(RatingProfile)
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colRpf).FindOne(sctx, bson.M{"id": key})
		if err := cur.Decode(rp); err != nil {
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	return
}

func (ms *MongoStorageNew) SetRatingProfileDrv(rp *RatingProfile) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colRpf).UpdateOne(sctx, bson.M{"id": rp.Id},
			bson.M{"$set": rp},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorageNew) RemoveRatingProfileDrv(key string) error {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colRpf).DeleteMany(sctx, bson.M{"id": key})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorageNew) GetLCRDrv(key string) (lcr *LCR, err error) {
	var result struct {
		Key   string
		Value *LCR
	}
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colLcr).FindOne(sctx, bson.M{"key": key})
		if err := cur.Decode(&result); err != nil {
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	return result.Value, err
}

func (ms *MongoStorageNew) SetLCRDrv(lcr *LCR) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colLcr).UpdateOne(sctx, bson.M{"key": lcr.GetId()},
			bson.M{"$set": struct {
				Key   string
				Value *LCR
			}{lcr.GetId(), lcr}},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorageNew) RemoveLCRDrv(key, transactionID string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colLcr).DeleteOne(sctx, bson.M{"key": key})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorageNew) GetDestination(key string, skipCache bool,
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

func (ms *MongoStorageNew) SetDestination(dest *Destination, transactionID string) (err error) {
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

func (ms *MongoStorageNew) RemoveDestination(destID string,
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

func (ms *MongoStorageNew) GetReverseDestination(prefix string, skipCache bool,
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

func (ms *MongoStorageNew) SetReverseDestination(dest *Destination,
	transactionID string) (err error) {
	for _, p := range dest.Prefixes {
		if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
			_, err = ms.getCol(colRds).UpdateOne(sctx, bson.M{"key": p},
				bson.M{"$addToSet": bson.M{"value": dest.Id}},
			)
			return err
		}); err != nil {
			return err
		}
	}
	return nil
}

func (ms *MongoStorageNew) UpdateReverseDestination(oldDest, newDest *Destination,
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
			)
			return err
		}); err != nil {
			return err
		}
	}
	return nil
}

func (ms *MongoStorageNew) GetActionsDrv(key string) (as Actions, err error) {
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

func (ms *MongoStorageNew) SetActionsDrv(key string, as Actions) error {
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

func (ms *MongoStorageNew) RemoveActionsDrv(key string) error {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colAct).DeleteOne(sctx, bson.M{"key": key})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorageNew) GetSharedGroupDrv(key string) (sg *SharedGroup, err error) {
	sg = new(SharedGroup)
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colShg).FindOne(sctx, bson.M{"id": key})
		if err := cur.Decode(sg); err != nil {
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	return
}

func (ms *MongoStorageNew) SetSharedGroupDrv(sg *SharedGroup) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colShg).UpdateOne(sctx, bson.M{"id": sg.Id},
			bson.M{"$set": sg},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorageNew) RemoveSharedGroupDrv(id, transactionID string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colShg).DeleteOne(sctx, bson.M{"id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorageNew) GetAccount(key string) (result *Account, err error) {
	result = new(Account)
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colAcc).FindOne(sctx, bson.M{"id": key})
		if err := cur.Decode(result); err != nil {
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	return
}

func (ms *MongoStorageNew) SetAccount(acc *Account) error {
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

func (ms *MongoStorageNew) RemoveAccount(key string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colAcc).DeleteOne(sctx, bson.M{"id": key})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorageNew) GetCdrStatsQueueDrv(key string) (sq *CDRStatsQueue, err error) {
	var result struct {
		Key   string
		Value *CDRStatsQueue
	}
	if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colStq).FindOne(sctx, bson.M{"key": key})
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
	return result.Value, nil
}

func (ms *MongoStorageNew) SetCdrStatsQueueDrv(sq *CDRStatsQueue) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colStq).UpdateOne(sctx, bson.M{"key": sq.GetId()},
			bson.M{"$set": struct {
				Key   string
				Value *CDRStatsQueue
			}{Key: sq.GetId(), Value: sq}},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorageNew) RemoveCdrStatsQueueDrv(id string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colStq).DeleteOne(sctx, bson.M{"key": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorageNew) GetSubscribersDrv() (result map[string]*SubscriberData, err error) {
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

func (ms *MongoStorageNew) SetSubscriberDrv(key string, sub *SubscriberData) (err error) {
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

func (ms *MongoStorageNew) RemoveSubscriberDrv(key string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colPbs).DeleteOne(sctx, bson.M{"key": key})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorageNew) GetUserDrv(key string) (up *UserProfile, err error) {
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

func (ms *MongoStorageNew) SetUserDrv(up *UserProfile) (err error) {
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

func (ms *MongoStorageNew) RemoveUserDrv(key string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colUsr).DeleteOne(sctx, bson.M{"key": key})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorageNew) GetAlias(key string, skipCache bool,
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

func (ms *MongoStorageNew) SetAlias(al *Alias, transactionID string) (err error) {
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

func (ms *MongoStorageNew) RemoveAlias(key, transactionID string) (err error) {
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

func (ms *MongoStorageNew) GetReverseAlias(reverseID string, skipCache bool,
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

func (ms *MongoStorageNew) SetReverseAlias(al *Alias, transactionID string) (err error) {
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
func (ms *MongoStorageNew) GetLoadHistory(limit int, skipCache bool,
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
	if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colLht).FindOne(sctx, bson.M{"key": utils.LOADINST_KEY})
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
func (ms *MongoStorageNew) AddLoadHistory(ldInst *utils.LoadInstance,
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
				return utils.ErrNotFound
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

func (ms *MongoStorageNew) GetActionTriggersDrv(key string) (atrs ActionTriggers, err error) {
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

func (ms *MongoStorageNew) SetActionTriggersDrv(key string, atrs ActionTriggers) (err error) {
	if len(atrs) == 0 {
		return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
			_, err = ms.getCol(colStq).DeleteOne(sctx, bson.M{"key": key})
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

func (ms *MongoStorageNew) RemoveActionTriggersDrv(key string) error {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colStq).DeleteOne(sctx, bson.M{"key": key})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorageNew) GetActionPlan(key string, skipCache bool,
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

func (ms *MongoStorageNew) SetActionPlan(key string, ats *ActionPlan,
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

func (ms *MongoStorageNew) RemoveActionPlan(key string, transactionID string) error {
	cCommit := cacheCommit(transactionID)
	Cache.Remove(utils.CacheActionPlans, key, cCommit, transactionID)
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colApl).DeleteOne(sctx, bson.M{"key": key})
		return err
	})
}

func (ms *MongoStorageNew) GetAllActionPlans() (ats map[string]*ActionPlan, err error) {
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

func (ms *MongoStorageNew) GetAccountActionPlans(acntID string, skipCache bool, transactionID string) (aPlIDs []string, err error) {
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

func (ms *MongoStorageNew) SetAccountActionPlans(acntID string, aPlIDs []string, overwrite bool) (err error) {
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
func (ms *MongoStorageNew) RemAccountActionPlans(acntID string, aPlIDs []string) (err error) {
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

func (ms *MongoStorageNew) PushTask(t *Task) error {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(colTsk).InsertOne(sctx, bson.M{"_id": objectid.New(), "task": t})
		return err
	})
}

func (ms *MongoStorageNew) PopTask() (t *Task, err error) {
	v := struct {
		ID   objectid.ObjectID `bson:"_id"`
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

func (ms *MongoStorageNew) GetDerivedChargersDrv(key string) (dcs *utils.DerivedChargers, err error) {
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

func (ms *MongoStorageNew) SetDerivedChargers(key string,
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

func (ms *MongoStorageNew) RemoveDerivedChargersDrv(id, transactionID string) (err error) {
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

func (ms *MongoStorageNew) GetCdrStatsDrv(key string) (cs *CdrStats, err error) {
	cs = new(CdrStats)
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colCrs).FindOne(sctx, bson.M{"id": key})
		if err := cur.Decode(cs); err != nil {
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	return
}

func (ms *MongoStorageNew) GetAllCdrStatsDrv() (css []*CdrStats, err error) {
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(colCrs).Find(sctx, nil)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var cs CdrStats
			err := cur.Decode(&cs)
			if err != nil {
				return err
			}
			clone := cs // avoid using the same pointer in append
			css = append(css, &clone)
		}
		return cur.Close(sctx)
	})
	return
}

func (ms *MongoStorageNew) SetCdrStatsDrv(cs *CdrStats) error {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colCrs).UpdateOne(sctx, bson.M{"id": cs.Id},
			bson.M{"$set": cs},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorageNew) GetResourceProfileDrv(tenant, id string) (rp *ResourceProfile, err error) {
	rp = new(ResourceProfile)
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colRsP).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		if err := cur.Decode(rp); err != nil {
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	return
}

func (ms *MongoStorageNew) SetResourceProfileDrv(rp *ResourceProfile) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colRsP).UpdateOne(sctx, bson.M{"tenant": rp.Tenant, "id": rp.ID},
			bson.M{"$set": rp},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorageNew) RemoveResourceProfileDrv(tenant, id string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colDcs).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorageNew) GetResourceDrv(tenant, id string) (r *Resource, err error) {
	r = new(Resource)
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colRes).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		if err := cur.Decode(r); err != nil {
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	return
}

func (ms *MongoStorageNew) SetResourceDrv(r *Resource) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colRes).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorageNew) RemoveResourceDrv(tenant, id string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colRes).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorageNew) GetTimingDrv(id string) (t *utils.TPTiming, err error) {
	t = new(utils.TPTiming)
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colTmg).FindOne(sctx, bson.M{"id": id})
		if err := cur.Decode(t); err != nil {
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	return
}

func (ms *MongoStorageNew) SetTimingDrv(t *utils.TPTiming) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colTmg).UpdateOne(sctx, bson.M{"id": t.ID},
			bson.M{"$set": t},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorageNew) RemoveTimingDrv(id string) (err error) {
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
func (ms *MongoStorageNew) GetFilterIndexesDrv(cacheID, itemIDPrefix, filterType string,
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
func (ms *MongoStorageNew) SetFilterIndexesDrv(cacheID, itemIDPrefix string,
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
		pairs := []interface{}{}
		var lastErr error
		for key, itmMp := range indexes {
			pairs = append(pairs, bson.M{"key": utils.ConcatenatedKey(dbKey, key)})
			if len(itmMp) == 0 {
				pairs = append(pairs, bson.M{"$unset": bson.M{"value": 1}})
			} else {
				pairs = append(pairs, bson.M{"$set": bson.M{"key": utils.ConcatenatedKey(dbKey, key), "value": itmMp.Slice()}})
			}

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

func (ms *MongoStorageNew) RemoveFilterIndexesDrv(cacheID, itemIDPrefix string) (err error) {
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

func (ms *MongoStorageNew) MatchFilterIndexDrv(cacheID, itemIDPrefix,
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
func (ms *MongoStorageNew) GetStatQueueProfileDrv(tenant string, id string) (sq *StatQueueProfile, err error) {
	sq = new(StatQueueProfile)
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colSqp).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		if err := cur.Decode(sq); err != nil {
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
func (ms *MongoStorageNew) SetStatQueueProfileDrv(sq *StatQueueProfile) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colSqp).UpdateOne(sctx, bson.M{"tenant": sq.Tenant, "id": sq.ID},
			bson.M{"$set": sq},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

// RemStatQueueProfileDrv removes a StatsQueue from dataDB
func (ms *MongoStorageNew) RemStatQueueProfileDrv(tenant, id string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colSqp).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

// GetStoredStatQueueDrv retrieves a StoredStatQueue
func (ms *MongoStorageNew) GetStoredStatQueueDrv(tenant, id string) (sq *StoredStatQueue, err error) {
	sq = new(StoredStatQueue)
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colSqs).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		if err := cur.Decode(sq); err != nil {
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
func (ms *MongoStorageNew) SetStoredStatQueueDrv(sq *StoredStatQueue) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colSqs).UpdateOne(sctx, bson.M{"tenant": sq.Tenant, "id": sq.ID},
			bson.M{"$set": sq},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

// RemStoredStatQueueDrv removes stored metrics for a StoredStatQueue
func (ms *MongoStorageNew) RemStoredStatQueueDrv(tenant, id string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colSqs).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

// GetThresholdProfileDrv retrieves a ThresholdProfile from dataDB
func (ms *MongoStorageNew) GetThresholdProfileDrv(tenant, ID string) (tp *ThresholdProfile, err error) {
	tp = new(ThresholdProfile)
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colTps).FindOne(sctx, bson.M{"tenant": tenant, "id": ID})
		if err := cur.Decode(tp); err != nil {
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
func (ms *MongoStorageNew) SetThresholdProfileDrv(tp *ThresholdProfile) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colTps).UpdateOne(sctx, bson.M{"tenant": tp.Tenant, "id": tp.ID},
			bson.M{"$set": tp},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

// RemoveThresholdProfile removes a ThresholdProfile from dataDB/cache
func (ms *MongoStorageNew) RemThresholdProfileDrv(tenant, id string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colTps).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorageNew) GetThresholdDrv(tenant, id string) (r *Threshold, err error) {
	r = new(Threshold)
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colThs).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		if err := cur.Decode(r); err != nil {
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	return
}

func (ms *MongoStorageNew) SetThresholdDrv(r *Threshold) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colThs).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorageNew) RemoveThresholdDrv(tenant, id string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colThs).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorageNew) GetFilterDrv(tenant, id string) (r *Filter, err error) {
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

func (ms *MongoStorageNew) SetFilterDrv(r *Filter) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colFlt).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorageNew) RemoveFilterDrv(tenant, id string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colFlt).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorageNew) GetSupplierProfileDrv(tenant, id string) (r *SupplierProfile, err error) {
	r = new(SupplierProfile)
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colSpp).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		if err := cur.Decode(r); err != nil {
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	return
}

func (ms *MongoStorageNew) SetSupplierProfileDrv(r *SupplierProfile) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colSpp).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorageNew) RemoveSupplierProfileDrv(tenant, id string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colSpp).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorageNew) GetAttributeProfileDrv(tenant, id string) (r *AttributeProfile, err error) {
	r = new(AttributeProfile)
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colAttr).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		if err := cur.Decode(r); err != nil {
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	return
}

func (ms *MongoStorageNew) SetAttributeProfileDrv(r *AttributeProfile) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colAttr).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorageNew) RemoveAttributeProfileDrv(tenant, id string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colAttr).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorageNew) GetChargerProfileDrv(tenant, id string) (r *ChargerProfile, err error) {
	r = new(ChargerProfile)
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colCpp).FindOne(sctx, bson.M{"tenant": tenant, "id": id})
		if err := cur.Decode(r); err != nil {
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	})
	return
}

func (ms *MongoStorageNew) SetChargerProfileDrv(r *ChargerProfile) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colCpp).UpdateOne(sctx, bson.M{"tenant": r.Tenant, "id": r.ID},
			bson.M{"$set": r},
			options.Update().SetUpsert(true),
		)
		return err
	})
}

func (ms *MongoStorageNew) RemoveChargerProfileDrv(tenant, id string) (err error) {
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(colCpp).DeleteOne(sctx, bson.M{"tenant": tenant, "id": id})
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}
