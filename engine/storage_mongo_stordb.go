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
along with this program. If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	"github.com/mongodb/mongo-go-driver/x/bsonx"
)

func (ms *MongoStorage) GetTpIds(colName string) (tpids []string, err error) {
	getTpIDs := func(ctx context.Context, col string, tpMap map[string]struct{}) (map[string]struct{}, error) {
		if strings.HasPrefix(col, "tp_") {
			result, err := ms.getCol(col).Distinct(ctx, "tpid", nil)
			if err != nil {
				return tpMap, err
			}
			for _, tpid := range result {
				tpMap[tpid.(string)] = struct{}{}
			}
		}
		return tpMap, nil
	}
	tpidMap := make(map[string]struct{})

	if colName == "" {
		if err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) error {
			col, err := ms.DB().ListCollections(sctx, nil, options.ListCollections().SetNameOnly(true))
			if err != nil {
				return err
			}
			for col.Next(sctx) {
				var elem struct{ Name string }
				if err := col.Decode(&elem); err != nil {
					return err
				}
				if tpidMap, err = getTpIDs(sctx, elem.Name, tpidMap); err != nil {
					return err
				}
			}
			return col.Close(sctx)
		}); err != nil {
			return nil, err
		}
	} else {
		if err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) error {
			tpidMap, err = getTpIDs(sctx, colName, tpidMap)
			return err
		}); err != nil {
			return nil, err
		}
	}
	for tpid := range tpidMap {
		tpids = append(tpids, tpid)
	}
	return tpids, nil
}

func (ms *MongoStorage) GetTpTableIds(tpid, table string, distinct utils.TPDistinctIds, filter map[string]string, pag *utils.Paginator) ([]string, error) {
	findMap := bson.M{}
	if tpid != "" {
		findMap["tpid"] = tpid
	}
	for k, v := range filter {
		findMap[k] = v
	}
	for k, v := range distinct { //fix for MongoStorage on TPUsers
		if v == "user_name" {
			distinct[k] = "username"
		}
	}
	if pag != nil && pag.SearchTerm != "" {
		var searchItems []bson.M
		for _, d := range distinct {
			searchItems = append(searchItems, bson.M{d: bsonx.Regex(".*"+regexp.QuoteMeta(pag.SearchTerm)+".*", "")})
		}
		// findMap["$and"] = []bson.M{{"$or": searchItems}} //before
		findMap["$or"] = searchItems // after
	}

	fop := options.Find()
	if pag != nil {
		if pag.Limit != nil {
			fop = fop.SetLimit(int64(*pag.Limit))
		}
		if pag.Offset != nil {
			fop = fop.SetSkip(int64(*pag.Offset))
		}
	}

	selectors := bson.M{"_id": 0}
	for i, d := range distinct {
		if d == "tag" { // convert the tag used in SQL into id used here
			distinct[i] = "id"
		}
		selectors[distinct[i]] = 1
	}
	fop.SetProjection(selectors)

	distinctIds := make(utils.StringMap)
	if err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(table).Find(sctx, findMap, fop)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var elem bson.D
			err := cur.Decode(&elem)
			if err != nil {
				return err
			}
			item := elem.Map()

			var id string
			last := len(distinct) - 1
			for i, d := range distinct {
				if distinctValue, ok := item[d]; ok {
					id += distinctValue.(string)
				}
				if i < last {
					id += utils.CONCATENATED_KEY_SEP
				}
			}
			distinctIds[id] = true
		}
		return cur.Close(sctx)
	}); err != nil {
		return nil, err
	}
	return distinctIds.Slice(), nil
}

func (ms *MongoStorage) GetTPTimings(tpid, id string) ([]*utils.ApierTPTiming, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.ApierTPTiming
	err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPTimings).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var el utils.ApierTPTiming
			err := cur.Decode(&el)
			if err != nil {
				return err
			}
			results = append(results, &el)
		}
		if len(results) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return results, err
}

func (ms *MongoStorage) GetTPDestinations(tpid, id string) ([]*utils.TPDestination, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPDestination
	err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPDestinations).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var el utils.TPDestination
			err := cur.Decode(&el)
			if err != nil {
				return err
			}
			results = append(results, &el)
		}
		if len(results) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return results, err
}

func (ms *MongoStorage) GetTPRates(tpid, id string) ([]*utils.TPRate, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPRate
	err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPRates).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var el utils.TPRate
			err := cur.Decode(&el)
			if err != nil {
				return err
			}
			for _, rs := range el.RateSlots {
				rs.SetDurations()
			}
			results = append(results, &el)
		}
		if len(results) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return results, err
}

func (ms *MongoStorage) GetTPDestinationRates(tpid, id string, pag *utils.Paginator) ([]*utils.TPDestinationRate, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPDestinationRate
	fop := options.Find()
	if pag != nil {
		if pag.Limit != nil {
			fop = fop.SetLimit(int64(*pag.Limit))
		}
		if pag.Offset != nil {
			fop = fop.SetSkip(int64(*pag.Offset))
		}
	}
	err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPDestinationRates).Find(sctx, filter, fop)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var el utils.TPDestinationRate
			err := cur.Decode(&el)
			if err != nil {
				return err
			}
			results = append(results, &el)
		}
		if len(results) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return results, err
}

func (ms *MongoStorage) GetTPRatingPlans(tpid, id string, pag *utils.Paginator) ([]*utils.TPRatingPlan, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPRatingPlan
	fop := options.Find()
	if pag != nil {
		if pag.Limit != nil {
			fop = fop.SetLimit(int64(*pag.Limit))
		}
		if pag.Offset != nil {
			fop = fop.SetSkip(int64(*pag.Offset))
		}
	}
	err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPRatingPlans).Find(sctx, filter, fop)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var el utils.TPRatingPlan
			err := cur.Decode(&el)
			if err != nil {
				return err
			}
			results = append(results, &el)
		}
		if len(results) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return results, err
}

func (ms *MongoStorage) GetTPRatingProfiles(tp *utils.TPRatingProfile) ([]*utils.TPRatingProfile, error) {
	filter := bson.M{"tpid": tp.TPid}
	if tp.Direction != "" {
		filter["direction"] = tp.Direction
	}
	if tp.Tenant != "" {
		filter["tenant"] = tp.Tenant
	}
	if tp.Category != "" {
		filter["category"] = tp.Category
	}
	if tp.Subject != "" {
		filter["subject"] = tp.Subject
	}
	if tp.LoadId != "" {
		filter["loadid"] = tp.LoadId
	}
	var results []*utils.TPRatingProfile
	err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPRateProfiles).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var el utils.TPRatingProfile
			err := cur.Decode(&el)
			if err != nil {
				return err
			}
			results = append(results, &el)
		}
		if len(results) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return results, err
}

func (ms *MongoStorage) GetTPSharedGroups(tpid, id string) ([]*utils.TPSharedGroups, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPSharedGroups
	err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPSharedGroups).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var el utils.TPSharedGroups
			err := cur.Decode(&el)
			if err != nil {
				return err
			}
			results = append(results, &el)
		}
		if len(results) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return results, err
}

func (ms *MongoStorage) GetTPUsers(tp *utils.TPUsers) ([]*utils.TPUsers, error) {
	filter := bson.M{"tpid": tp.TPid}
	if tp.Tenant != "" {
		filter["tenant"] = tp.Tenant
	}
	if tp.UserName != "" {
		filter["username"] = tp.UserName
	}
	var results []*utils.TPUsers
	err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPUsers).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var el utils.TPUsers
			err := cur.Decode(&el)
			if err != nil {
				return err
			}
			results = append(results, &el)
		}
		if len(results) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return results, err
}

func (ms *MongoStorage) GetTPAliases(tp *utils.TPAliases) ([]*utils.TPAliases, error) {
	filter := bson.M{"tpid": tp.TPid}
	if tp.Direction != "" {
		filter["direction"] = tp.Direction
	}
	if tp.Tenant != "" {
		filter["tenant"] = tp.Tenant
	}
	if tp.Category != "" {
		filter["category"] = tp.Category
	}
	if tp.Account != "" {
		filter["account"] = tp.Account
	}
	if tp.Subject != "" {
		filter["subject"] = tp.Subject
	}
	if tp.Context != "" {
		filter["context"] = tp.Context
	}
	var results []*utils.TPAliases
	err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPAliases).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var el utils.TPAliases
			err := cur.Decode(&el)
			if err != nil {
				return err
			}
			results = append(results, &el)
		}
		if len(results) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return results, err
}

func (ms *MongoStorage) GetTPResources(tpid, id string) ([]*utils.TPResource, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPResource
	err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPResources).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var el utils.TPResource
			err := cur.Decode(&el)
			if err != nil {
				return err
			}
			results = append(results, &el)
		}
		if len(results) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return results, err
}

func (ms *MongoStorage) GetTPStats(tpid, id string) ([]*utils.TPStats, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPStats
	err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPStats).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var el utils.TPStats
			err := cur.Decode(&el)
			if err != nil {
				return err
			}
			results = append(results, &el)
		}
		if len(results) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return results, err
}
func (ms *MongoStorage) GetTPDerivedChargers(tp *utils.TPDerivedChargers) ([]*utils.TPDerivedChargers, error) {
	filter := bson.M{"tpid": tp.TPid}
	if tp.Direction != "" {
		filter["direction"] = tp.Direction
	}
	if tp.Tenant != "" {
		filter["tenant"] = tp.Tenant
	}
	if tp.Category != "" {
		filter["category"] = tp.Category
	}
	if tp.Subject != "" {
		filter["subject"] = tp.Subject
	}
	if tp.Account != "" {
		filter["account"] = tp.Account
	}
	if tp.LoadId != "" {
		filter["loadid"] = tp.LoadId
	}
	var results []*utils.TPDerivedChargers
	err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPDerivedChargers).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var el utils.TPDerivedChargers
			err := cur.Decode(&el)
			if err != nil {
				return err
			}
			results = append(results, &el)
		}
		if len(results) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return results, err
}

func (ms *MongoStorage) GetTPActions(tpid, id string) ([]*utils.TPActions, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPActions
	err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPActions).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var el utils.TPActions
			err := cur.Decode(&el)
			if err != nil {
				return err
			}
			results = append(results, &el)
		}
		if len(results) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return results, err
}

func (ms *MongoStorage) GetTPActionPlans(tpid, id string) ([]*utils.TPActionPlan, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPActionPlan
	err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPActionPlans).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var el utils.TPActionPlan
			err := cur.Decode(&el)
			if err != nil {
				return err
			}
			results = append(results, &el)
		}
		if len(results) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return results, err
}

func (ms *MongoStorage) GetTPActionTriggers(tpid, id string) ([]*utils.TPActionTriggers, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPActionTriggers
	err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPActionTriggers).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var el utils.TPActionTriggers
			err := cur.Decode(&el)
			if err != nil {
				return err
			}
			results = append(results, &el)
		}
		if len(results) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return results, err
}

func (ms *MongoStorage) GetTPAccountActions(tp *utils.TPAccountActions) ([]*utils.TPAccountActions, error) {
	filter := bson.M{"tpid": tp.TPid}
	if tp.Tenant != "" {
		filter["tenant"] = tp.Tenant
	}
	if tp.Account != "" {
		filter["account"] = tp.Account
	}
	if tp.LoadId != "" {
		filter["loadid"] = tp.LoadId
	}
	var results []*utils.TPAccountActions
	err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPAccountActions).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var el utils.TPAccountActions
			err := cur.Decode(&el)
			if err != nil {
				return err
			}
			results = append(results, &el)
		}
		if len(results) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return results, err
}

func (ms *MongoStorage) RemTpData(table, tpid string, args map[string]string) error {
	if len(table) == 0 { // Remove tpid out of all tables
		return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) error {
			col, err := ms.DB().ListCollections(sctx, nil, options.ListCollections().SetNameOnly(true))
			if err != nil {
				return err
			}
			for col.Next(sctx) {
				var elem struct{ Name string }
				if err := col.Decode(&elem); err != nil {
					return err
				}
				if strings.HasPrefix(elem.Name, "tp_") {
					_, err = ms.getCol(elem.Name).DeleteMany(sctx, bson.M{"tpid": tpid})
					if err != nil {
						return err
					}
				}
			}
			return col.Close(sctx)
		})
	}
	// Remove from a single table
	if args == nil {
		args = make(map[string]string)
	}
	for arg, val := range args { //fix for Mongo TPUsers tables
		if arg == "user_name" {
			delete(args, arg)
			args["username"] = val
		}
	}

	if _, has := args["tag"]; has { // API uses tag to be compatible with SQL models, fix it here
		args["id"] = args["tag"]
		delete(args, "tag")
	}
	if tpid != "" {
		args["tpid"] = tpid
	}
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		dr, err := ms.getCol(table).DeleteOne(sctx, args)
		if dr.DeletedCount == 0 {
			return utils.ErrNotFound
		}
		return err
	})
}

func (ms *MongoStorage) SetTPTimings(tps []*utils.ApierTPTiming) error {
	if len(tps) == 0 {
		return nil
	}
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tps {
			_, err = ms.getCol(utils.TBLTPTimings).UpdateOne(sctx, bson.M{"tpid": tp.TPid, "id": tp.ID},
				bson.M{"$set": tp},
				options.Update().SetUpsert(true),
			)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) SetTPDestinations(tpDsts []*utils.TPDestination) (err error) {
	if len(tpDsts) == 0 {
		return nil
	}
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tpDsts {
			_, err = ms.getCol(utils.TBLTPDestinations).UpdateOne(sctx, bson.M{"tpid": tp.TPid, "id": tp.ID},
				bson.M{"$set": tp},
				options.Update().SetUpsert(true),
			)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) SetTPRates(tps []*utils.TPRate) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tps {
			if found, _ := m[tp.ID]; !found {
				m[tp.ID] = true
				_, err := ms.getCol(utils.TBLTPRates).DeleteMany(sctx, bson.M{"tpid": tp.TPid, "id": tp.ID})
				if err != nil {
					return err
				}
			}
			_, err := ms.getCol(utils.TBLTPRates).InsertOne(sctx, tp)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) SetTPDestinationRates(tps []*utils.TPDestinationRate) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tps {
			if found, _ := m[tp.ID]; !found {
				m[tp.ID] = true
				_, err := ms.getCol(utils.TBLTPDestinationRates).DeleteMany(sctx, bson.M{"tpid": tp.TPid, "id": tp.ID})
				if err != nil {
					return err
				}
			}
			_, err := ms.getCol(utils.TBLTPDestinationRates).InsertOne(sctx, tp)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) SetTPRatingPlans(tps []*utils.TPRatingPlan) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tps {
			if found, _ := m[tp.ID]; !found {
				m[tp.ID] = true
				_, err := ms.getCol(utils.TBLTPRatingPlans).DeleteMany(sctx, bson.M{"tpid": tp.TPid, "id": tp.ID})
				if err != nil {
					return err
				}
			}
			_, err := ms.getCol(utils.TBLTPRatingPlans).InsertOne(sctx, tp)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) SetTPRatingProfiles(tps []*utils.TPRatingProfile) error {
	if len(tps) == 0 {
		return nil
	}
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tps {
			_, err = ms.getCol(utils.TBLTPRateProfiles).UpdateOne(sctx, bson.M{
				"tpid":      tp.TPid,
				"loadid":    tp.LoadId,
				"direction": tp.Direction,
				"tenant":    tp.Tenant,
				"category":  tp.Category,
				"subject":   tp.Subject,
			}, bson.M{"$set": tp}, options.Update().SetUpsert(true))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) SetTPSharedGroups(tps []*utils.TPSharedGroups) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tps {
			if found, _ := m[tp.ID]; !found {
				m[tp.ID] = true
				_, err := ms.getCol(utils.TBLTPSharedGroups).DeleteMany(sctx, bson.M{"tpid": tp.TPid, "id": tp.ID})
				if err != nil {
					return err
				}
			}
			_, err := ms.getCol(utils.TBLTPSharedGroups).InsertOne(sctx, tp)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) SetTPUsers(tps []*utils.TPUsers) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tps {
			if found, _ := m[tp.GetId()]; !found {
				m[tp.GetId()] = true
				if _, err := ms.getCol(utils.TBLTPUsers).DeleteMany(sctx, bson.M{
					"tpid":     tp.TPid,
					"tenant":   tp.Tenant,
					"username": tp.UserName,
				}); err != nil {
					return err
				}
			}
			if _, err := ms.getCol(utils.TBLTPUsers).InsertOne(sctx, tp); err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) SetTPAliases(tps []*utils.TPAliases) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tps {
			if found, _ := m[tp.Direction]; !found {
				m[tp.Direction] = true
				if _, err := ms.getCol(utils.TBLTPAliases).DeleteMany(sctx, bson.M{
					"tpid":      tp.TPid,
					"direction": tp.Direction,
					"tenant":    tp.Tenant,
					"category":  tp.Category,
					"account":   tp.Account,
					"subject":   tp.Subject,
					"context":   tp.Context}); err != nil {
					return err
				}
			}
			if _, err := ms.getCol(utils.TBLTPAliases).InsertOne(sctx, tp); err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) SetTPDerivedChargers(tps []*utils.TPDerivedChargers) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tps {
			if found, _ := m[tp.Direction]; !found {
				m[tp.Direction] = true
				if _, err := ms.getCol(utils.TBLTPDerivedChargers).DeleteMany(sctx, bson.M{
					"tpid":      tp.TPid,
					"direction": tp.Direction,
					"tenant":    tp.Tenant,
					"category":  tp.Category,
					"account":   tp.Account,
					"subject":   tp.Subject}); err != nil {
					return err
				}
			}
			if _, err := ms.getCol(utils.TBLTPDerivedChargers).InsertOne(sctx, tp); err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) SetTPActions(tps []*utils.TPActions) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tps {
			if found, _ := m[tp.ID]; !found {
				m[tp.ID] = true
				if _, err := ms.getCol(utils.TBLTPActions).DeleteMany(sctx, bson.M{"tpid": tp.TPid, "id": tp.ID}); err != nil {
					return err
				}
			}
			if _, err := ms.getCol(utils.TBLTPActions).InsertOne(sctx, tp); err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) SetTPActionPlans(tps []*utils.TPActionPlan) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tps {
			if found, _ := m[tp.ID]; !found {
				m[tp.ID] = true
				if _, err := ms.getCol(utils.TBLTPActionPlans).DeleteMany(sctx, bson.M{"tpid": tp.TPid, "id": tp.ID}); err != nil {
					return err
				}
			}
			if _, err := ms.getCol(utils.TBLTPActionPlans).InsertOne(sctx, tp); err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) SetTPActionTriggers(tps []*utils.TPActionTriggers) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tps {
			if found, _ := m[tp.ID]; !found {
				m[tp.ID] = true
				if _, err := ms.getCol(utils.TBLTPActionTriggers).DeleteMany(sctx, bson.M{"tpid": tp.TPid, "id": tp.ID}); err != nil {
					return err
				}
			}
			if _, err := ms.getCol(utils.TBLTPActionTriggers).InsertOne(sctx, tp); err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) SetTPAccountActions(tps []*utils.TPAccountActions) error {
	if len(tps) == 0 {
		return nil
	}
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tps {
			_, err = ms.getCol(utils.TBLTPAccountActions).UpdateOne(sctx, bson.M{
				"tpid":    tp.TPid,
				"loadid":  tp.LoadId,
				"tenant":  tp.Tenant,
				"account": tp.Account,
			}, bson.M{"$set": tp}, options.Update().SetUpsert(true))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) SetTPResources(tpRLs []*utils.TPResource) (err error) {
	if len(tpRLs) == 0 {
		return
	}
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tpRLs {
			_, err = ms.getCol(utils.TBLTPResources).UpdateOne(sctx, bson.M{"tpid": tp.TPid, "id": tp.ID},
				bson.M{"$set": tp}, options.Update().SetUpsert(true))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) SetTPRStats(tps []*utils.TPStats) (err error) {
	if len(tps) == 0 {
		return
	}
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tps {
			_, err = ms.getCol(utils.TBLTPStats).UpdateOne(sctx, bson.M{"tpid": tp.TPid, "id": tp.ID},
				bson.M{"$set": tp}, options.Update().SetUpsert(true))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) SetSMCost(smc *SMCost) error {
	if smc.CostDetails == nil {
		return nil
	}
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(utils.SessionsCostsTBL).InsertOne(sctx, smc)
		return err
	})
}

func (ms *MongoStorage) RemoveSMCost(smc *SMCost) error {
	remParams := bson.M{}
	if smc != nil {
		remParams = bson.M{"cgrid": smc.CGRID, "runid": smc.RunID}
	}
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(utils.SessionsCostsTBL).DeleteMany(sctx, remParams)
		return err
	})
}

func (ms *MongoStorage) GetSMCosts(cgrid, runid, originHost, originIDPrefix string) (smcs []*SMCost, err error) {
	filter := bson.M{}
	if cgrid != "" {
		filter[CGRIDLow] = cgrid
	}
	if runid != "" {
		filter[RunIDLow] = runid
	}
	if originHost != "" {
		filter[OriginHostLow] = originHost
	}
	if originIDPrefix != "" {
		filter[OriginIDLow] = bsonx.Regex(fmt.Sprintf("^%s", originIDPrefix), "")
	}
	err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.SessionsCostsTBL).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var smCost SMCost
			err := cur.Decode(&smCost)
			if err != nil {
				return err
			}
			clone := smCost
			smcs = append(smcs, &clone)
		}
		if len(smcs) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return smcs, err
}

func (ms *MongoStorage) SetCDR(cdr *CDR, allowUpdate bool) (err error) {
	if cdr.OrderID == 0 {
		cdr.OrderID = ms.cnter.Next()
	}
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		if allowUpdate {
			_, err = ms.getCol(ColCDRs).UpdateOne(sctx,
				bson.M{CGRIDLow: cdr.CGRID, RunIDLow: cdr.RunID},
				bson.M{"$set": cdr}, options.Update().SetUpsert(true))
			// return err
		} else {
			_, err = ms.getCol(ColCDRs).InsertOne(sctx, cdr)
		}
		return err
	})
}

func (ms *MongoStorage) cleanEmptyFilters(filters bson.M) {
	for k, v := range filters {
		switch value := v.(type) {
		case *int64:
			if value == nil {
				delete(filters, k)
			}
		case *float64:
			if value == nil {
				delete(filters, k)
			}
		case *time.Time:
			if value == nil {
				delete(filters, k)
			}
		case *time.Duration:
			if value == nil {
				delete(filters, k)
			}
		case []string:
			if len(value) == 0 {
				delete(filters, k)
			}
		case bson.M:
			ms.cleanEmptyFilters(value)
			if len(value) == 0 {
				delete(filters, k)
			}
		}
	}
}

//  _, err := col(ColCDRs).UpdateAll(bson.M{CGRIDLow: bson.M{"$in": cgrIds}}, bson.M{"$set": bson.M{"deleted_at": time.Now()}})
func (ms *MongoStorage) GetCDRs(qryFltr *utils.CDRsFilter, remove bool) ([]*CDR, int64, error) {
	var minUsage, maxUsage *time.Duration
	if len(qryFltr.MinUsage) != 0 {
		if parsed, err := utils.ParseDurationWithNanosecs(qryFltr.MinUsage); err != nil {
			return nil, 0, err
		} else {
			minUsage = &parsed
		}
	}
	if len(qryFltr.MaxUsage) != 0 {
		if parsed, err := utils.ParseDurationWithNanosecs(qryFltr.MaxUsage); err != nil {
			return nil, 0, err
		} else {
			maxUsage = &parsed
		}
	}
	filters := bson.M{
		CGRIDLow:       bson.M{"$in": qryFltr.CGRIDs, "$nin": qryFltr.NotCGRIDs},
		RunIDLow:       bson.M{"$in": qryFltr.RunIDs, "$nin": qryFltr.NotRunIDs},
		OriginIDLow:    bson.M{"$in": qryFltr.OriginIDs, "$nin": qryFltr.NotOriginIDs},
		OrderIDLow:     bson.M{"$gte": qryFltr.OrderIDStart, "$lt": qryFltr.OrderIDEnd},
		ToRLow:         bson.M{"$in": qryFltr.ToRs, "$nin": qryFltr.NotToRs},
		CDRHostLow:     bson.M{"$in": qryFltr.OriginHosts, "$nin": qryFltr.NotOriginHosts},
		CDRSourceLow:   bson.M{"$in": qryFltr.Sources, "$nin": qryFltr.NotSources},
		RequestTypeLow: bson.M{"$in": qryFltr.RequestTypes, "$nin": qryFltr.NotRequestTypes},
		TenantLow:      bson.M{"$in": qryFltr.Tenants, "$nin": qryFltr.NotTenants},
		CategoryLow:    bson.M{"$in": qryFltr.Categories, "$nin": qryFltr.NotCategories},
		AccountLow:     bson.M{"$in": qryFltr.Accounts, "$nin": qryFltr.NotAccounts},
		SubjectLow:     bson.M{"$in": qryFltr.Subjects, "$nin": qryFltr.NotSubjects},
		SetupTimeLow:   bson.M{"$gte": qryFltr.SetupTimeStart, "$lt": qryFltr.SetupTimeEnd},
		AnswerTimeLow:  bson.M{"$gte": qryFltr.AnswerTimeStart, "$lt": qryFltr.AnswerTimeEnd},
		CreatedAtLow:   bson.M{"$gte": qryFltr.CreatedAtStart, "$lt": qryFltr.CreatedAtEnd},
		UpdatedAtLow:   bson.M{"$gte": qryFltr.UpdatedAtStart, "$lt": qryFltr.UpdatedAtEnd},
		UsageLow:       bson.M{"$gte": minUsage, "$lt": maxUsage},
		//CostDetailsLow + "." + AccountLow: bson.M{"$in": qryFltr.RatedAccounts, "$nin": qryFltr.NotRatedAccounts},
		//CostDetailsLow + "." + SubjectLow: bson.M{"$in": qryFltr.RatedSubjects, "$nin": qryFltr.NotRatedSubjects},
	}
	//file, _ := ioutil.TempFile(os.TempDir(), "debug")
	//file.WriteString(fmt.Sprintf("FILTER: %v\n", utils.ToIJSON(qryFltr)))
	//file.WriteString(fmt.Sprintf("BEFORE: %v\n", utils.ToIJSON(filters)))
	ms.cleanEmptyFilters(filters)
	if len(qryFltr.DestinationPrefixes) != 0 {
		var regexpRule string
		for _, prefix := range qryFltr.DestinationPrefixes {
			if len(prefix) == 0 {
				continue
			}
			if len(regexpRule) != 0 {
				regexpRule += "|"
			}
			regexpRule += "^(" + prefix + ")"
		}
		if _, hasIt := filters["$and"]; !hasIt {
			filters["$and"] = make([]bson.M, 0)
		}
		filters["$and"] = append(filters["$and"].([]bson.M), bson.M{DestinationLow: bsonx.Regex(regexpRule, "")}) // $and gathers all rules not fitting top level query
	}
	if len(qryFltr.NotDestinationPrefixes) != 0 {
		if _, hasIt := filters["$and"]; !hasIt {
			filters["$and"] = make([]bson.M, 0)
		}
		for _, prefix := range qryFltr.NotDestinationPrefixes {
			if len(prefix) == 0 {
				continue
			}
			filters["$and"] = append(filters["$and"].([]bson.M), bson.M{DestinationLow: bsonx.Regex("^(?!"+prefix+")", "")})
		}
	}

	if len(qryFltr.ExtraFields) != 0 {
		var extrafields []bson.M
		for field, value := range qryFltr.ExtraFields {
			if value == utils.MetaExists {
				extrafields = append(extrafields, bson.M{"extrafields." + field: bson.M{"$exists": true}})
			} else {
				extrafields = append(extrafields, bson.M{"extrafields." + field: value})
			}
		}
		filters["$and"] = extrafields
	}

	if len(qryFltr.NotExtraFields) != 0 {
		var extrafields []bson.M
		for field, value := range qryFltr.NotExtraFields {
			if value == utils.MetaExists {
				extrafields = append(extrafields, bson.M{"extrafields." + field: bson.M{"$exists": false}})
			} else {
				extrafields = append(extrafields, bson.M{"extrafields." + field: bson.M{"$ne": value}})
			}
		}
		filters["$and"] = extrafields
	}

	if qryFltr.MinCost != nil {
		if qryFltr.MaxCost == nil {
			filters[CostLow] = bson.M{"$gte": *qryFltr.MinCost}
		} else if *qryFltr.MinCost == 0.0 && *qryFltr.MaxCost == -1.0 { // Special case when we want to skip errors
			filters["$or"] = []bson.M{
				{CostLow: bson.M{"$gte": 0.0}},
			}
		} else {
			filters[CostLow] = bson.M{"$gte": *qryFltr.MinCost, "$lt": *qryFltr.MaxCost}
		}
	} else if qryFltr.MaxCost != nil {
		if *qryFltr.MaxCost == -1.0 { // Non-rated CDRs
			filters[CostLow] = 0.0 // Need to include it otherwise all CDRs will be returned
		} else { // Above limited CDRs, since MinCost is empty, make sure we query also NULL cost
			filters[CostLow] = bson.M{"$lt": *qryFltr.MaxCost}
		}
	}
	//file.WriteString(fmt.Sprintf("AFTER: %v\n", utils.ToIJSON(filters)))
	//file.Close()
	if remove {
		var chgd int64
		err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
			dr, err := ms.getCol(ColCDRs).DeleteMany(sctx, filters)
			chgd = dr.DeletedCount
			return err
		})
		return nil, chgd, err
	}
	fop := options.Find()
	cop := options.Count()
	if qryFltr.Paginator.Limit != nil {
		fop = fop.SetLimit(int64(*qryFltr.Paginator.Limit))
		cop = cop.SetLimit(int64(*qryFltr.Paginator.Limit))
	}
	if qryFltr.Paginator.Offset != nil {
		fop = fop.SetSkip(int64(*qryFltr.Paginator.Offset))
		cop = cop.SetSkip(int64(*qryFltr.Paginator.Offset))
	}

	if qryFltr.OrderBy != "" {
		var orderVal string
		separateVals := strings.Split(qryFltr.OrderBy, utils.INFIELD_SEP)
		ordVal := 1
		if len(separateVals) == 2 && separateVals[1] == "desc" {
			ordVal = -1
			// orderVal += "-"
		}
		switch separateVals[0] {
		case utils.OrderID:
			orderVal += "orderid"
		case utils.AnswerTime:
			orderVal += "answertime"
		case utils.SetupTime:
			orderVal += "setuptime"
		case utils.Usage:
			orderVal += "usage"
		case utils.Cost:
			orderVal += "cost"
		default:
			return nil, 0, fmt.Errorf("Invalid value : %s", separateVals[0])
		}
		fop = fop.SetSort(bson.M{orderVal: ordVal})
	}
	if qryFltr.Count {
		var cnt int64
		if err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
			cnt, err = ms.getCol(ColCDRs).Count(sctx, filters, cop)
			return err
		}); err != nil {
			return nil, 0, err
		}
		return nil, cnt, nil
	}
	// Execute query
	var cdrs []*CDR
	err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(ColCDRs).Find(sctx, filters, fop)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			cdr := CDR{}
			err := cur.Decode(&cdr)
			if err != nil {
				return err
			}
			clone := cdr
			cdrs = append(cdrs, &clone)
		}
		if len(cdrs) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return cdrs, 0, err
}

func (ms *MongoStorage) GetTPStat(tpid, id string) ([]*utils.TPStats, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPStats
	err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPStats).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var tp utils.TPStats
			err := cur.Decode(&tp)
			if err != nil {
				return err
			}
			results = append(results, &tp)
		}
		if len(results) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return results, err
}

func (ms *MongoStorage) SetTPStats(tpSTs []*utils.TPStats) (err error) {
	if len(tpSTs) == 0 {
		return
	}
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tpSTs {
			_, err = ms.getCol(utils.TBLTPStats).UpdateOne(sctx, bson.M{"tpid": tp.TPid, "id": tp.ID},
				bson.M{"$set": tp},
				options.Update().SetUpsert(true),
			)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) GetTPThresholds(tpid, id string) ([]*utils.TPThreshold, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPThreshold
	err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPThresholds).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var tp utils.TPThreshold
			err := cur.Decode(&tp)
			if err != nil {
				return err
			}
			results = append(results, &tp)
		}
		if len(results) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return results, err
}

func (ms *MongoStorage) SetTPThresholds(tpTHs []*utils.TPThreshold) (err error) {
	if len(tpTHs) == 0 {
		return
	}
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tpTHs {
			_, err = ms.getCol(utils.TBLTPThresholds).UpdateOne(sctx, bson.M{"tpid": tp.TPid, "id": tp.ID},
				bson.M{"$set": tp},
				options.Update().SetUpsert(true),
			)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) GetTPFilters(tpid, id string) ([]*utils.TPFilterProfile, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	results := []*utils.TPFilterProfile{}
	err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPFilters).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var tp utils.TPFilterProfile
			err := cur.Decode(&tp)
			if err != nil {
				return err
			}
			results = append(results, &tp)
		}
		if len(results) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return results, err
}

func (ms *MongoStorage) SetTPFilters(tpTHs []*utils.TPFilterProfile) (err error) {
	if len(tpTHs) == 0 {
		return
	}
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tpTHs {
			_, err = ms.getCol(utils.TBLTPFilters).UpdateOne(sctx, bson.M{"tpid": tp.TPid, "id": tp.ID},
				bson.M{"$set": tp},
				options.Update().SetUpsert(true),
			)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) GetTPSuppliers(tpid, id string) ([]*utils.TPSupplierProfile, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPSupplierProfile
	err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPSuppliers).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var tp utils.TPSupplierProfile
			err := cur.Decode(&tp)
			if err != nil {
				return err
			}
			results = append(results, &tp)
		}
		if len(results) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return results, err
}

func (ms *MongoStorage) SetTPSuppliers(tpSPs []*utils.TPSupplierProfile) (err error) {
	if len(tpSPs) == 0 {
		return
	}
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tpSPs {
			_, err = ms.getCol(utils.TBLTPSuppliers).UpdateOne(sctx, bson.M{"tpid": tp.TPid, "id": tp.ID},
				bson.M{"$set": tp},
				options.Update().SetUpsert(true),
			)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) GetTPAttributes(tpid, id string) ([]*utils.TPAttributeProfile, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPAttributeProfile
	err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPAttributes).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var tp utils.TPAttributeProfile
			err := cur.Decode(&tp)
			if err != nil {
				return err
			}
			results = append(results, &tp)
		}
		if len(results) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return results, err
}

func (ms *MongoStorage) SetTPAttributes(tpSPs []*utils.TPAttributeProfile) (err error) {
	if len(tpSPs) == 0 {
		return
	}
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tpSPs {
			_, err = ms.getCol(utils.TBLTPAttributes).UpdateOne(sctx, bson.M{"tpid": tp.TPid, "id": tp.ID},
				bson.M{"$set": tp},
				options.Update().SetUpsert(true),
			)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) GetTPChargers(tpid, id string) ([]*utils.TPChargerProfile, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPChargerProfile
	err := ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPChargers).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var tp utils.TPChargerProfile
			err := cur.Decode(&tp)
			if err != nil {
				return err
			}
			results = append(results, &tp)
		}
		if len(results) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return results, err
}

func (ms *MongoStorage) SetTPChargers(tpCPP []*utils.TPChargerProfile) (err error) {
	if len(tpCPP) == 0 {
		return
	}
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tpCPP {
			_, err = ms.getCol(utils.TBLTPChargers).UpdateOne(sctx, bson.M{"tpid": tp.TPid, "id": tp.ID},
				bson.M{"$set": tp},
				options.Update().SetUpsert(true),
			)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) GetVersions(itm string) (vrs Versions, err error) {
	fop := options.FindOne()
	if itm != "" {
		fop.SetProjection(bson.M{itm: 1, "_id": 0})
	} else {
		fop.SetProjection(bson.M{"_id": 0})
	}
	if err = ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(colVer).FindOne(sctx, nil, fop)
		if err := cur.Decode(&vrs); err != nil {
			if err == mongo.ErrNoDocuments {
				return utils.ErrNotFound
			}
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if len(vrs) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MongoStorage) SetVersions(vrs Versions, overwrite bool) (err error) {
	if overwrite {
		ms.RemoveVersions(nil)
	}
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(colVer).UpdateOne(sctx, nil, bson.M{"$set": vrs},
			options.Update().SetUpsert(true),
		)
		return err
	})
	// }
	// return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) error {
	// 	_, err := ms.getCol(colVer).InsertOne(sctx, vrs)
	// 	return err
	// })
	// _, err = col.Upsert(bson.M{}, bson.M{"$set": &vrs})
}

func (ms *MongoStorage) RemoveVersions(vrs Versions) (err error) {
	if len(vrs) == 0 {
		return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
			dr, err := ms.getCol(colVer).DeleteOne(sctx, nil)
			if dr.DeletedCount == 0 {
				return utils.ErrNotFound
			}
			return err
		})
	}
	return ms.client.UseSession(ms.ctx, func(sctx mongo.SessionContext) (err error) {
		for k := range vrs {
			if _, err = ms.getCol(colVer).UpdateOne(sctx, nil, bson.M{"$unset": bson.M{k: 1}},
				options.Update().SetUpsert(true)); err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) GetStorageType() string {
	return utils.MONGO
}
