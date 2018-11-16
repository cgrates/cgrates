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
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
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

func (ms *MongoStorageNew) getCon(col string) *mongo.Collection {
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
		col := ms.getCon(colName)
		if _, err := col.DeleteMany(sctx, bson.M{}); err != nil {
			return err
		} /*
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
			}*/
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
		col := ms.getCon(colName)

		if _, err := col.DeleteMany(sctx, bson.M{}); err != nil {
			return err
		}
		/*
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
					if err := ms.RemoveDestination(dest.Id, utils.NonTransactional); err != nil {
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
					if err := ms.RemoveAlias(al.GetId(), utils.NonTransactional); err != nil {
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
						if err = ms.RemAccountActionPlans(acntID, []string{apl.Id}); err != nil {
							return err
						}
					}
				}
			}*/
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

/* TODO: implement
func (ms *MongoStorageNew) GetKeysForPrefix(prefix string) (result []string, err error) {
	var category, subject string
	keyLen := len(utils.DESTINATION_PREFIX)
	if len(prefix) < keyLen {
		return nil, fmt.Errorf("unsupported prefix in GetKeysForPrefix: %s", prefix)
	}
	category = prefix[:keyLen] // prefix length
	tntID := utils.NewTenantID(prefix[keyLen:])
	subject = fmt.Sprintf("^%s", prefix[keyLen:]) // old way, no tenant support

	err = ms.client.UseSession(ctx, func(sctx mongo.SessionContext) error {

		db := ms.DB()
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
			qry := bson.M{}
			if tntID.Tenant != "" {
				qry["tenant"] = tntID.Tenant
			}
			if tntID.ID != "" {
				qry["id"] = bson.M{"$regex": bson.RegEx{Pattern: subject}}
			}
			iter := db.C(colRsP).Find(qry).Select(bson.M{"tenant": 1, "id": 1}).Iter()
			for iter.Next(&idResult) {
				result = append(result, utils.ResourceProfilesPrefix+utils.ConcatenatedKey(idResult.Tenant, idResult.Id))
			}
		case utils.ResourcesPrefix:
			qry := bson.M{}
			if tntID.Tenant != "" {
				qry["tenant"] = tntID.Tenant
			}
			if tntID.ID != "" {
				qry["id"] = bson.M{"$regex": bson.RegEx{Pattern: subject}}
			}
			iter := db.C(colRes).Find(qry).Select(bson.M{"tenant": 1, "id": 1}).Iter()
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
			qry := bson.M{}
			if tntID.Tenant != "" {
				qry["tenant"] = tntID.Tenant
			}
			if tntID.ID != "" {
				qry["id"] = bson.M{"$regex": bson.RegEx{Pattern: subject}}
			}
			iter := db.C(colSqp).Find(qry).Select(bson.M{"tenant": 1, "id": 1}).Iter()
			for iter.Next(&idResult) {
				result = append(result, utils.StatQueueProfilePrefix+utils.ConcatenatedKey(idResult.Tenant, idResult.Id))
			}
		case utils.AccountActionPlansPrefix:
			iter := db.C(colAAp).Find(bson.M{"key": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"key": 1}).Iter()
			for iter.Next(&keyResult) {
				result = append(result, utils.AccountActionPlansPrefix+keyResult.Key)
			}
		case utils.TimingsPrefix:
			iter := db.C(colTmg).Find(bson.M{"id": bson.M{"$regex": bson.RegEx{Pattern: subject}}}).Select(bson.M{"id": 1}).Iter()
			for iter.Next(&idResult) {
				result = append(result, utils.TimingsPrefix+idResult.Id)
			}
		case utils.FilterPrefix:
			qry := bson.M{}
			if tntID.Tenant != "" {
				qry["tenant"] = tntID.Tenant
			}
			if tntID.ID != "" {
				qry["id"] = bson.M{"$regex": bson.RegEx{Pattern: subject}}
			}
			iter := db.C(colFlt).Find(qry).Select(bson.M{"tenant": 1, "id": 1}).Iter()
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
		case utils.ThresholdProfilePrefix:
			qry := bson.M{}
			if tntID.Tenant != "" {
				qry["tenant"] = tntID.Tenant
			}
			if tntID.ID != "" {
				qry["id"] = bson.M{"$regex": bson.RegEx{Pattern: subject}}
			}
			iter := db.C(colTps).Find(qry).Select(bson.M{"tenant": 1, "id": 1}).Iter()
			for iter.Next(&idResult) {
				result = append(result, utils.ThresholdProfilePrefix+utils.ConcatenatedKey(idResult.Tenant, idResult.Id))
			}
		case utils.SupplierProfilePrefix:
			qry := bson.M{}
			if tntID.Tenant != "" {
				qry["tenant"] = tntID.Tenant
			}
			if tntID.ID != "" {
				qry["id"] = bson.M{"$regex": bson.RegEx{Pattern: subject}}
			}
			iter := db.C(colSpp).Find(qry).Select(bson.M{"tenant": 1, "id": 1}).Iter()
			for iter.Next(&idResult) {
				result = append(result, utils.SupplierProfilePrefix+utils.ConcatenatedKey(idResult.Tenant, idResult.Id))
			}
		case utils.AttributeProfilePrefix:
			qry := bson.M{}
			if tntID.Tenant != "" {
				qry["tenant"] = tntID.Tenant
			}
			if tntID.ID != "" {
				qry["id"] = bson.M{"$regex": bson.RegEx{Pattern: subject}}
			}
			iter := db.C(colAttr).Find(qry).Select(bson.M{"tenant": 1, "id": 1}).Iter()
			for iter.Next(&idResult) {
				result = append(result, utils.AttributeProfilePrefix+utils.ConcatenatedKey(idResult.Tenant, idResult.Id))
			}
		case utils.ChargerProfilePrefix:
			qry := bson.M{}
			if tntID.Tenant != "" {
				qry["tenant"] = tntID.Tenant
			}
			if tntID.ID != "" {
				qry["id"] = bson.M{"$regex": bson.RegEx{Pattern: subject}}
			}
			iter := db.C(colCpp).Find(qry).Select(bson.M{"tenant": 1, "id": 1}).Iter()
			for iter.Next(&idResult) {
				result = append(result, utils.ChargerProfilePrefix+utils.ConcatenatedKey(idResult.Tenant, idResult.Id))
			}
		default:
			err = fmt.Errorf("unsupported prefix in GetKeysForPrefix: %s", prefix)
		}
		return
	})
	return
}
//*/
