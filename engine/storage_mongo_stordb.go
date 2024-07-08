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
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (ms *MongoStorage) GetTpIds(colName string) (tpids []string, err error) {
	getTpIDs := func(ctx context.Context, col string, tpMap utils.StringSet) (utils.StringSet, error) {
		if strings.HasPrefix(col, "tp_") {
			result, err := ms.getCol(col).Distinct(ctx, "tpid", bson.D{})
			if err != nil {
				return tpMap, err
			}
			for _, tpid := range result {
				tpMap.Add(tpid.(string))
			}
		}
		return tpMap, nil
	}
	tpidMap := make(utils.StringSet)

	if colName == "" {
		if err := ms.query(func(sctx mongo.SessionContext) error {
			col, err := ms.DB().ListCollections(sctx, bson.D{}, options.ListCollections().SetNameOnly(true))
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
		if err := ms.query(func(sctx mongo.SessionContext) error {
			tpidMap, err = getTpIDs(sctx, colName, tpidMap)
			return err
		}); err != nil {
			return nil, err
		}
	}
	tpids = tpidMap.AsSlice()
	return tpids, nil
}

func (ms *MongoStorage) GetTpTableIds(tpid, table string, distinctIDs utils.TPDistinctIds,
	filter map[string]string, pag *utils.PaginatorWithSearch) ([]string, error) {
	findMap := bson.M{}
	if tpid != "" {
		findMap["tpid"] = tpid
	}
	for k, v := range filter {
		if k != "" && v != "" {
			findMap[k] = v
		}
	}

	fop := options.Find()
	if pag != nil {
		if pag.Search != "" {
			var searchItems []bson.M
			for _, distinctID := range distinctIDs {
				searchItems = append(searchItems,
					bson.M{
						distinctID: primitive.Regex{
							Pattern: ".*" + regexp.QuoteMeta(pag.Search) + ".*",
						},
					},
				)
			}
			findMap["$or"] = searchItems
		}
		if pag.Paginator != nil {
			if pag.Limit != nil {
				fop = fop.SetLimit(int64(*pag.Limit))
			}
			if pag.Offset != nil {
				fop = fop.SetSkip(int64(*pag.Offset))
			}
		}
	}

	selectors := bson.M{"_id": 0}
	for i, distinctID := range distinctIDs {
		if distinctID == "tag" { // convert the tag used in SQL into id used here
			distinctIDs[i] = "id"
		}
		selectors[distinctIDs[i]] = 1
	}
	fop.SetProjection(selectors)

	distinctIds := make(utils.StringMap)
	err := ms.query(func(sctx mongo.SessionContext) (qryErr error) {
		cur, qryErr := ms.getCol(table).Find(sctx, findMap, fop)
		if qryErr != nil {
			return qryErr
		}
		for cur.Next(sctx) {
			var item bson.M
			err := cur.Decode(&item)
			if err != nil {
				return err
			}

			var id string
			last := len(distinctIDs) - 1
			for i, distinctID := range distinctIDs {
				if distinctValue, ok := item[distinctID]; ok {
					id += distinctValue.(string)
				}
				if i < last {
					id += utils.ConcatenatedKeySep
				}
			}
			distinctIds[id] = true
		}
		return cur.Close(sctx)
	})
	if err != nil {
		return nil, err
	}
	return distinctIds.Slice(), nil
}

func (ms *MongoStorage) GetTPTimings(tpid, id string) ([]*utils.ApierTPTiming, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	var tpTimings []*utils.ApierTPTiming
	err := ms.query(func(sctx mongo.SessionContext) (qryErr error) {
		cur, qryErr := ms.getCol(utils.TBLTPTimings).Find(sctx, filter)
		if qryErr != nil {
			return qryErr
		}
		for cur.Next(sctx) {
			var el utils.ApierTPTiming
			qryErr = cur.Decode(&el)
			if qryErr != nil {
				return qryErr
			}
			tpTimings = append(tpTimings, &el)
		}
		if len(tpTimings) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return tpTimings, err
}

func (ms *MongoStorage) GetTPDestinations(tpid, id string) ([]*utils.TPDestination, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	var tpDestinations []*utils.TPDestination
	err := ms.query(func(sctx mongo.SessionContext) (qryErr error) {
		cur, qryErr := ms.getCol(utils.TBLTPDestinations).Find(sctx, filter)
		if qryErr != nil {
			return qryErr
		}
		for cur.Next(sctx) {
			var el utils.TPDestination
			qryErr = cur.Decode(&el)
			if qryErr != nil {
				return qryErr
			}
			tpDestinations = append(tpDestinations, &el)
		}
		if len(tpDestinations) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return tpDestinations, err
}

func (ms *MongoStorage) GetTPRates(tpid, id string) ([]*utils.TPRateRALs, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	var tpRates []*utils.TPRateRALs
	err := ms.query(func(sctx mongo.SessionContext) (qryErr error) {
		cur, qryErr := ms.getCol(utils.TBLTPRates).Find(sctx, filter)
		if qryErr != nil {
			return qryErr
		}
		for cur.Next(sctx) {
			var el utils.TPRateRALs
			err := cur.Decode(&el)
			if err != nil {
				return err
			}
			for _, rs := range el.RateSlots {
				rs.SetDurations()
			}
			tpRates = append(tpRates, &el)
		}
		if len(tpRates) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return tpRates, err
}

func (ms *MongoStorage) GetTPDestinationRates(tpid, id string, pag *utils.Paginator) ([]*utils.TPDestinationRate, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	var tpDstRates []*utils.TPDestinationRate
	fop := options.Find()
	if pag != nil {
		if pag.Limit != nil {
			fop = fop.SetLimit(int64(*pag.Limit))
		}
		if pag.Offset != nil {
			fop = fop.SetSkip(int64(*pag.Offset))
		}
	}
	err := ms.query(func(sctx mongo.SessionContext) (qryErr error) {
		cur, qryErr := ms.getCol(utils.TBLTPDestinationRates).Find(sctx, filter, fop)
		if qryErr != nil {
			return qryErr
		}
		for cur.Next(sctx) {
			var el utils.TPDestinationRate
			qryErr = cur.Decode(&el)
			if qryErr != nil {
				return qryErr
			}
			tpDstRates = append(tpDstRates, &el)
		}
		if len(tpDstRates) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return tpDstRates, err
}

func (ms *MongoStorage) GetTPRatingPlans(tpid, id string, pag *utils.Paginator) ([]*utils.TPRatingPlan, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	var tpRatingPlans []*utils.TPRatingPlan
	fop := options.Find()
	if pag != nil {
		if pag.Limit != nil {
			fop = fop.SetLimit(int64(*pag.Limit))
		}
		if pag.Offset != nil {
			fop = fop.SetSkip(int64(*pag.Offset))
		}
	}
	err := ms.query(func(sctx mongo.SessionContext) (qryErr error) {
		cur, qryErr := ms.getCol(utils.TBLTPRatingPlans).Find(sctx, filter, fop)
		if qryErr != nil {
			return qryErr
		}
		for cur.Next(sctx) {
			var el utils.TPRatingPlan
			qryErr = cur.Decode(&el)
			if qryErr != nil {
				return qryErr
			}
			tpRatingPlans = append(tpRatingPlans, &el)
		}
		if len(tpRatingPlans) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return tpRatingPlans, err
}

func (ms *MongoStorage) GetTPRatingProfiles(tp *utils.TPRatingProfile) ([]*utils.TPRatingProfile, error) {
	filter := bson.M{"tpid": tp.TPid}
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
	var tpRatingProfiles []*utils.TPRatingProfile
	err := ms.query(func(sctx mongo.SessionContext) (qryErr error) {
		cur, qryErr := ms.getCol(utils.TBLTPRatingProfiles).Find(sctx, filter)
		if qryErr != nil {
			return qryErr
		}
		for cur.Next(sctx) {
			var el utils.TPRatingProfile
			qryErr = cur.Decode(&el)
			if qryErr != nil {
				return qryErr
			}
			tpRatingProfiles = append(tpRatingProfiles, &el)
		}
		if len(tpRatingProfiles) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return tpRatingProfiles, err
}

func (ms *MongoStorage) GetTPSharedGroups(tpid, id string) ([]*utils.TPSharedGroups, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.TPSharedGroups
	err := ms.query(func(sctx mongo.SessionContext) (err error) {
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

func (ms *MongoStorage) GetTPResources(tpid, tenant, id string) ([]*utils.TPResourceProfile, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	if tenant != "" {
		filter["tenant"] = tenant
	}
	var results []*utils.TPResourceProfile
	err := ms.query(func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPResources).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var el utils.TPResourceProfile
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

func (ms *MongoStorage) GetTPStats(tpid, tenant, id string) ([]*utils.TPStatProfile, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if id != "" {
		filter["id"] = id
	}
	if tenant != "" {
		filter["tenant"] = tenant
	}
	var results []*utils.TPStatProfile
	err := ms.query(func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPStats).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var el utils.TPStatProfile
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

func (ms *MongoStorage) GetTPTrends(tpid string, tenant string, id string) ([]*utils.TPTrendsProfile, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if id != "" {
		filter["id"] = id
	}
	if tenant != "" {
		filter["tenant"] = tenant
	}
	var results []*utils.TPTrendsProfile
	err := ms.query(func(sctx mongo.SessionContext) error {
		cur, err := ms.getCol(utils.TBLTPTrends).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var el utils.TPTrendsProfile
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

func (ms *MongoStorage) GetTPRankings(tpid string, tenant string, id string) ([]*utils.TPRankingProfile, error) {
	filter := bson.M{
		"tpid": tpid,
	}
	if id != "" {
		filter["id"] = id
	}
	if tenant != "" {
		filter["tenant"] = tenant
	}
	var results []*utils.TPRankingProfile
	err := ms.query(func(sctx mongo.SessionContext) error {
		cur, err := ms.getCol(utils.TBLTPRankings).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var el utils.TPRankingProfile
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
	err := ms.query(func(sctx mongo.SessionContext) (err error) {
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
	err := ms.query(func(sctx mongo.SessionContext) (err error) {
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
	err := ms.query(func(sctx mongo.SessionContext) (err error) {
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
	err := ms.query(func(sctx mongo.SessionContext) (err error) {
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
	if table == utils.EmptyString { // Remove tpid out of all tables
		return ms.query(func(sctx mongo.SessionContext) error {
			col, err := ms.DB().ListCollections(sctx, bson.D{}, options.ListCollections().SetNameOnly(true))
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

	if _, has := args["tag"]; has { // API uses tag to be compatible with SQL models, fix it here
		args["id"] = args["tag"]
		delete(args, "tag")
	}
	if tpid != "" {
		args["tpid"] = tpid
	}
	return ms.query(func(sctx mongo.SessionContext) (err error) {
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
	return ms.query(func(sctx mongo.SessionContext) (err error) {
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
	return ms.query(func(sctx mongo.SessionContext) (err error) {
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

func (ms *MongoStorage) SetTPRates(tps []*utils.TPRateRALs) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tps {
			if !m[tp.ID] {
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
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tps {
			if !m[tp.ID] {
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
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tps {
			if !m[tp.ID] {
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
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tps {
			_, err = ms.getCol(utils.TBLTPRatingProfiles).UpdateOne(sctx, bson.M{
				"tpid":     tp.TPid,
				"loadid":   tp.LoadId,
				"tenant":   tp.Tenant,
				"category": tp.Category,
				"subject":  tp.Subject,
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
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tps {
			if !m[tp.ID] {
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

func (ms *MongoStorage) SetTPActions(tps []*utils.TPActions) error {
	if len(tps) == 0 {
		return nil
	}
	m := make(map[string]bool)
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tps {
			if !m[tp.ID] {
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
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tps {
			if !m[tp.ID] {
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
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tps {
			if !m[tp.ID] {
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
	return ms.query(func(sctx mongo.SessionContext) (err error) {
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

func (ms *MongoStorage) SetTPResources(tpRLs []*utils.TPResourceProfile) (err error) {
	if len(tpRLs) == 0 {
		return
	}
	return ms.query(func(sctx mongo.SessionContext) (err error) {
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

func (ms *MongoStorage) SetTPRStats(tps []*utils.TPStatProfile) (err error) {
	if len(tps) == 0 {
		return
	}
	return ms.query(func(sctx mongo.SessionContext) (err error) {
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
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(utils.SessionCostsTBL).InsertOne(sctx, smc)
		return err
	})
}

func (ms *MongoStorage) RemoveSMCost(smc *SMCost) error {
	remParams := bson.M{}
	if smc != nil {
		remParams = bson.M{"cgrid": smc.CGRID, "runid": smc.RunID}
	}
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(utils.SessionCostsTBL).DeleteMany(sctx, remParams)
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
		filter[OriginIDLow] = primitive.Regex{
			Pattern: "^" + originIDPrefix,
		}
	}
	err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.SessionCostsTBL).Find(sctx, filter)
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
			clone.CostDetails.initCache()
			smcs = append(smcs, &clone)
		}
		if len(smcs) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return smcs, err
}

func (ms *MongoStorage) RemoveSMCosts(qryFltr *utils.SMCostFilter) error {
	filters := bson.M{
		CGRIDLow:      bson.M{"$in": qryFltr.CGRIDs, "$nin": qryFltr.NotCGRIDs},
		RunIDLow:      bson.M{"$in": qryFltr.RunIDs, "$nin": qryFltr.NotRunIDs},
		OriginHostLow: bson.M{"$in": qryFltr.OriginHosts, "$nin": qryFltr.NotOriginHosts},
		OriginIDLow:   bson.M{"$in": qryFltr.OriginIDs, "$nin": qryFltr.NotOriginIDs},
		CostSourceLow: bson.M{"$in": qryFltr.CostSources, "$nin": qryFltr.NotCostSources},
		UsageLow:      bson.M{"$gte": qryFltr.Usage.Min, "$lt": qryFltr.Usage.Max},
		CreatedAtLow:  bson.M{"$gte": qryFltr.CreatedAt.Begin, "$lt": qryFltr.CreatedAt.End},
	}
	ms.cleanEmptyFilters(filters)
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(utils.SessionCostsTBL).DeleteMany(sctx, filters)
		return err
	})
}

func (ms *MongoStorage) SetCDR(cdr *CDR, allowUpdate bool) error {
	if cdr.OrderID == 0 {
		cdr.OrderID = ms.counter.Next()
	}
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		if allowUpdate {
			_, err = ms.getCol(ColCDRs).UpdateOne(sctx,
				bson.M{CGRIDLow: cdr.CGRID, RunIDLow: cdr.RunID},
				bson.M{"$set": cdr}, options.Update().SetUpsert(true))
			return
		}
		_, err = ms.getCol(ColCDRs).InsertOne(sctx, cdr)
		if err != nil && strings.Contains(err.Error(), "E11000") { // Mongo returns E11000 when key is duplicated
			err = utils.ErrExists
		}
		return
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

// _, err := col(ColCDRs).UpdateAll(bson.M{CGRIDLow: bson.M{"$in": cgrIds}}, bson.M{"$set": bson.M{"deleted_at": time.Now()}})
func (ms *MongoStorage) GetCDRs(qryFltr *utils.CDRsFilter, remove bool) (cdrs []*CDR, n int64, err error) {
	var minUsage, maxUsage *time.Duration
	if qryFltr.MinUsage != utils.EmptyString {
		parsedDur, err := utils.ParseDurationWithNanosecs(qryFltr.MinUsage)
		if err != nil {
			return nil, 0, err
		} else {
			minUsage = &parsedDur
		}
	}
	if qryFltr.MaxUsage != utils.EmptyString {
		parsedDur, err := utils.ParseDurationWithNanosecs(qryFltr.MaxUsage)
		if err != nil {
			return nil, 0, err
		} else {
			maxUsage = &parsedDur
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
		// CostDetailsLow + "." + AccountLow: bson.M{"$in": qryFltr.RatedAccounts, "$nin": qryFltr.NotRatedAccounts},
		// CostDetailsLow + "." + SubjectLow: bson.M{"$in": qryFltr.RatedSubjects, "$nin": qryFltr.NotRatedSubjects},
	}
	// file, _ := os.TempFile(os.TempDir(), "debug")
	// file.WriteString(fmt.Sprintf("FILTER: %v\n", utils.ToIJSON(qryFltr)))
	// file.WriteString(fmt.Sprintf("BEFORE: %v\n", utils.ToIJSON(filters)))
	ms.cleanEmptyFilters(filters)
	if len(qryFltr.DestinationPrefixes) != 0 {
		var regexpRule string
		for _, prefix := range qryFltr.DestinationPrefixes {
			if prefix == utils.EmptyString {
				continue
			}
			if regexpRule != utils.EmptyString {
				regexpRule += "|"
			}
			regexpRule += "^(" + regexp.QuoteMeta(prefix) + ")"
		}
		if _, hasIt := filters["$and"]; !hasIt {
			filters["$and"] = make([]bson.M, 0)
		}
		// The "$and" operator is used to include additional query conditions that cannot be
		// represented at the top level of the query.
		filters["$and"] = append(filters["$and"].([]bson.M), bson.M{
			DestinationLow: primitive.Regex{
				Pattern: regexpRule,
			},
		})
	}
	if len(qryFltr.NotDestinationPrefixes) != 0 {
		if _, hasIt := filters["$and"]; !hasIt {
			filters["$and"] = make([]bson.M, 0)
		}
		for _, prefix := range qryFltr.NotDestinationPrefixes {
			if prefix == utils.EmptyString {
				continue
			}
			filters["$and"] = append(filters["$and"].([]bson.M),
				bson.M{
					DestinationLow: primitive.Regex{
						Pattern: "^(?!" + prefix + ")",
					},
				},
			)
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
	// file.WriteString(fmt.Sprintf("AFTER: %v\n", utils.ToIJSON(filters)))
	// file.Close()
	if remove {
		err := ms.query(func(sctx mongo.SessionContext) (qryErr error) {
			dr, qryErr := ms.getCol(ColCDRs).DeleteMany(sctx, filters)
			if qryErr != nil {
				return qryErr
			}
			n = dr.DeletedCount
			return qryErr
		})
		return nil, n, err
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
		separateVals := strings.Split(qryFltr.OrderBy, utils.InfieldSep)
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
			return nil, 0, fmt.Errorf("invalid value : %s", separateVals[0])
		}
		fop = fop.SetSort(bson.M{orderVal: ordVal})
	}
	if qryFltr.Count {
		var cnt int64
		err := ms.query(func(sctx mongo.SessionContext) (qryErr error) {
			cnt, qryErr = ms.getCol(ColCDRs).CountDocuments(sctx, filters, cop)
			return qryErr
		})
		if err != nil {
			return nil, 0, err
		}
		return nil, cnt, nil
	}
	// Execute query
	err = ms.query(func(sctx mongo.SessionContext) (qryErr error) {
		cur, qryErr := ms.getCol(ColCDRs).Find(sctx, filters, fop)
		if qryErr != nil {
			return qryErr
		}
		for cur.Next(sctx) {
			cdr := CDR{}
			err := cur.Decode(&cdr)
			if err != nil {
				return err
			}
			clone := cdr
			clone.CostDetails.initCache()
			cdrs = append(cdrs, &clone)
		}
		if len(cdrs) == 0 {
			return utils.ErrNotFound
		}
		return cur.Close(sctx)
	})
	return cdrs, 0, err
}

func (ms *MongoStorage) SetTPStats(tpSTs []*utils.TPStatProfile) (err error) {
	if len(tpSTs) == 0 {
		return
	}
	return ms.query(func(sctx mongo.SessionContext) (err error) {
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

func (ms *MongoStorage) SetTPTrends(tpTrends []*utils.TPTrendsProfile) (err error) {
	if len(tpTrends) == 0 {
		return
	}
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tpTrends {
			_, err := ms.getCol(utils.TBLTPTrends).UpdateOne(sctx, bson.M{"tpid": tp.TPid, "id": tp.ID},
				bson.M{"$set": tp}, options.Update().SetUpsert(true))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) SetTPRankings(tpRankings []*utils.TPRankingProfile) (err error) {
	if len(tpRankings) == 0 {
		return
	}
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tpRankings {
			_, err := ms.getCol(utils.TBLTPRankings).UpdateOne(sctx, bson.M{"tpid": tp.TPid, "id": tp.ID},
				bson.M{"$set": tp}, options.Update().SetUpsert(true))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) GetTPThresholds(tpid, tenant, id string) ([]*utils.TPThresholdProfile, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	if tenant != "" {
		filter["tenant"] = tenant
	}
	var results []*utils.TPThresholdProfile
	err := ms.query(func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPThresholds).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var tp utils.TPThresholdProfile
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

func (ms *MongoStorage) SetTPThresholds(tpTHs []*utils.TPThresholdProfile) (err error) {
	if len(tpTHs) == 0 {
		return
	}
	return ms.query(func(sctx mongo.SessionContext) (err error) {
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

func (ms *MongoStorage) GetTPFilters(tpid, tenant, id string) ([]*utils.TPFilterProfile, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	if tenant != "" {
		filter["tenant"] = tenant
	}
	results := []*utils.TPFilterProfile{}
	err := ms.query(func(sctx mongo.SessionContext) (err error) {
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
	return ms.query(func(sctx mongo.SessionContext) (err error) {
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

func (ms *MongoStorage) GetTPRoutes(tpid, tenant, id string) ([]*utils.TPRouteProfile, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	if tenant != "" {
		filter["tenant"] = tenant
	}
	var results []*utils.TPRouteProfile
	err := ms.query(func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPRoutes).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var tp utils.TPRouteProfile
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

func (ms *MongoStorage) SetTPRoutes(tpRoutes []*utils.TPRouteProfile) (err error) {
	if len(tpRoutes) == 0 {
		return
	}
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tpRoutes {
			_, err = ms.getCol(utils.TBLTPRoutes).UpdateOne(sctx, bson.M{"tpid": tp.TPid, "id": tp.ID},
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

func (ms *MongoStorage) GetTPAttributes(tpid, tenant, id string) ([]*utils.TPAttributeProfile, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	if tenant != "" {
		filter["tenant"] = tenant
	}
	var results []*utils.TPAttributeProfile
	err := ms.query(func(sctx mongo.SessionContext) (err error) {
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

func (ms *MongoStorage) SetTPAttributes(tpRoutes []*utils.TPAttributeProfile) (err error) {
	if len(tpRoutes) == 0 {
		return
	}
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tpRoutes {
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

func (ms *MongoStorage) GetTPChargers(tpid, tenant, id string) ([]*utils.TPChargerProfile, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	if tenant != "" {
		filter["tenant"] = tenant
	}
	var results []*utils.TPChargerProfile
	err := ms.query(func(sctx mongo.SessionContext) (err error) {
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
	return ms.query(func(sctx mongo.SessionContext) (err error) {
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

func (ms *MongoStorage) GetTPDispatcherProfiles(tpid, tenant, id string) ([]*utils.TPDispatcherProfile, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	if tenant != "" {
		filter["tenant"] = tenant
	}
	var results []*utils.TPDispatcherProfile
	err := ms.query(func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPDispatchers).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var tp utils.TPDispatcherProfile
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

func (ms *MongoStorage) SetTPDispatcherProfiles(tpDPPs []*utils.TPDispatcherProfile) (err error) {
	if len(tpDPPs) == 0 {
		return
	}
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tpDPPs {
			_, err = ms.getCol(utils.TBLTPDispatchers).UpdateOne(sctx, bson.M{"tpid": tp.TPid, "id": tp.ID},
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

func (ms *MongoStorage) GetTPDispatcherHosts(tpid, tenant, id string) ([]*utils.TPDispatcherHost, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	if tenant != "" {
		filter["tenant"] = tenant
	}
	var results []*utils.TPDispatcherHost
	err := ms.query(func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPDispatcherHosts).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var tp utils.TPDispatcherHost
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

func (ms *MongoStorage) SetTPDispatcherHosts(tpDPPs []*utils.TPDispatcherHost) (err error) {
	if len(tpDPPs) == 0 {
		return
	}
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tpDPPs {
			_, err = ms.getCol(utils.TBLTPDispatcherHosts).UpdateOne(sctx, bson.M{"tpid": tp.TPid, "id": tp.ID},
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

func (ms *MongoStorage) GetVersions(itm string) (Versions, error) {
	fop := options.FindOne()
	if itm != "" {
		fop.SetProjection(bson.M{itm: 1, "_id": 0})
	} else {
		fop.SetProjection(bson.M{"_id": 0})
	}
	var vrs Versions
	err := ms.query(func(sctx mongo.SessionContext) (err error) {
		sr := ms.getCol(ColVer).FindOne(sctx, bson.D{}, fop)
		decodeErr := sr.Decode(&vrs)
		if errors.Is(decodeErr, mongo.ErrNoDocuments) {
			return utils.ErrNotFound
		}
		return decodeErr
	})
	if err != nil {
		return nil, err
	}
	if len(vrs) == 0 {
		return nil, utils.ErrNotFound
	}
	return vrs, nil
}

func (ms *MongoStorage) SetVersions(vrs Versions, overwrite bool) error {
	if overwrite {
		err := ms.RemoveVersions(nil)
		if err != nil && !errors.Is(err, utils.ErrNotFound) {
			return err
		}
	}
	return ms.query(func(sctx mongo.SessionContext) error {
		_, err := ms.getCol(ColVer).UpdateOne(sctx, bson.D{}, bson.M{"$set": vrs},
			options.Update().SetUpsert(true),
		)
		return err
	})
	// }
	// return ms.query( func(sctx mongo.SessionContext) error {
	// 	_, err := ms.getCol(ColVer).InsertOne(sctx, vrs)
	// 	return err
	// })
	// _, err = col.Upsert(bson.M{}, bson.M{"$set": &vrs})
}

func (ms *MongoStorage) RemoveVersions(vrs Versions) error {
	if len(vrs) == 0 {
		return ms.query(func(sctx mongo.SessionContext) error {
			dr, err := ms.getCol(ColVer).DeleteOne(sctx, bson.D{})
			if err != nil {
				return err
			}
			if dr.DeletedCount == 0 {
				return utils.ErrNotFound
			}
			return nil
		})
	}
	return ms.query(func(sctx mongo.SessionContext) error {
		for k := range vrs {
			if _, err := ms.getCol(ColVer).UpdateOne(sctx, bson.D{}, bson.M{"$unset": bson.M{k: 1}},
				options.Update().SetUpsert(true)); err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) GetStorageType() string {
	return utils.MetaMongo
}
