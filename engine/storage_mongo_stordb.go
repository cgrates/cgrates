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

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
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

func (ms *MongoStorage) GetTpTableIds(tpid, table string, distinct []string,
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
			for _, d := range distinct {
				searchItems = append(searchItems, bson.M{d: bsonx.Regex(".*"+regexp.QuoteMeta(pag.Search)+".*", "")})
			}
			// findMap["$and"] = []bson.M{{"$or": searchItems}} //before
			findMap["$or"] = searchItems // after
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
	for i, d := range distinct {
		if d == "tag" { // convert the tag used in SQL into id used here
			distinct[i] = "id"
		}
		selectors[distinct[i]] = 1
	}
	fop.SetProjection(selectors)

	distinctIds := make(utils.StringSet)
	if err := ms.query(func(sctx mongo.SessionContext) (err error) {
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
					id += utils.ConcatenatedKeySep
				}
			}
			distinctIds.Add(id)
		}
		return cur.Close(sctx)
	}); err != nil {
		return nil, err
	}
	return distinctIds.AsSlice(), nil
}

func (ms *MongoStorage) GetTPTimings(tpid, id string) ([]*utils.ApierTPTiming, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	var results []*utils.ApierTPTiming
	err := ms.query(func(sctx mongo.SessionContext) (err error) {
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
	err := ms.query(func(sctx mongo.SessionContext) (err error) {
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

func (ms *MongoStorage) RemTpData(table, tpid string, args map[string]string) error {
	if len(table) == 0 { // Remove tpid out of all tables
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

func (ms *MongoStorage) SetCDR(cdr *CDR, allowUpdate bool) error {
	if cdr.OrderID == 0 {
		cdr.OrderID = ms.cnter.Next()
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
	//file, _ := os.TempFile(os.TempDir(), "debug")
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
			regexpRule += "^(" + regexp.QuoteMeta(prefix) + ")"
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
		err := ms.query(func(sctx mongo.SessionContext) (err error) {
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
			return nil, 0, fmt.Errorf("Invalid value : %s", separateVals[0])
		}
		fop = fop.SetSort(bson.M{orderVal: ordVal})
	}
	if qryFltr.Count {
		var cnt int64
		if err := ms.query(func(sctx mongo.SessionContext) (err error) {
			cnt, err = ms.getCol(ColCDRs).CountDocuments(sctx, filters, cop)
			return err
		}); err != nil {
			return nil, 0, err
		}
		return nil, cnt, nil
	}
	// Execute query
	var cdrs []*CDR
	err := ms.query(func(sctx mongo.SessionContext) (err error) {
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

func (ms *MongoStorage) GetTPRateProfiles(tpid, tenant, id string) ([]*utils.TPRateProfile, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	if tenant != "" {
		filter["tenant"] = tenant
	}
	var results []*utils.TPRateProfile
	err := ms.query(func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPRateProfiles).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var tp utils.TPRateProfile
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

func (ms *MongoStorage) SetTPRateProfiles(tpDPPs []*utils.TPRateProfile) (err error) {
	if len(tpDPPs) == 0 {
		return
	}

	return ms.query(func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tpDPPs {
			_, err = ms.getCol(utils.TBLTPRateProfiles).UpdateOne(sctx, bson.M{"tpid": tp.TPid, "id": tp.ID},
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

func (ms *MongoStorage) GetTPActionProfiles(tpid, tenant, id string) ([]*utils.TPActionProfile, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	if tenant != "" {
		filter["tenant"] = tenant
	}
	var results []*utils.TPActionProfile
	err := ms.query(func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPActionProfiles).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var tp utils.TPActionProfile
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

func (ms *MongoStorage) GetTPAccounts(tpid, tenant, id string) ([]*utils.TPAccount, error) {
	filter := bson.M{"tpid": tpid}
	if id != "" {
		filter["id"] = id
	}
	if tenant != "" {
		filter["tenant"] = tenant
	}
	var results []*utils.TPAccount
	err := ms.query(func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(utils.TBLTPAccounts).Find(sctx, filter)
		if err != nil {
			return err
		}
		for cur.Next(sctx) {
			var tp utils.TPAccount
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

func (ms *MongoStorage) SetTPActionProfiles(tpAps []*utils.TPActionProfile) (err error) {
	if len(tpAps) == 0 {
		return
	}
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tpAps {
			_, err = ms.getCol(utils.TBLTPActionProfiles).UpdateOne(sctx, bson.M{"tpid": tp.TPid, "id": tp.ID},
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

func (ms *MongoStorage) SetTPAccounts(tpAps []*utils.TPAccount) (err error) {
	if len(tpAps) == 0 {
		return
	}
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		for _, tp := range tpAps {
			_, err = ms.getCol(utils.TBLTPAccounts).UpdateOne(sctx, bson.M{"tpid": tp.TPid, "id": tp.ID},
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
	if err = ms.query(func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColVer).FindOne(sctx, bson.D{}, fop)
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
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColVer).UpdateOne(sctx, bson.D{}, bson.M{"$set": vrs},
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

func (ms *MongoStorage) RemoveVersions(vrs Versions) (err error) {
	if len(vrs) == 0 {
		return ms.query(func(sctx mongo.SessionContext) (err error) {
			var dr *mongo.DeleteResult
			dr, err = ms.getCol(ColVer).DeleteOne(sctx, bson.D{})
			if err != nil {
				return
			}
			if dr.DeletedCount == 0 {
				return utils.ErrNotFound
			}
			return
		})
	}
	return ms.query(func(sctx mongo.SessionContext) (err error) {
		for k := range vrs {
			if _, err = ms.getCol(ColVer).UpdateOne(sctx, bson.D{}, bson.M{"$unset": bson.M{k: 1}},
				options.Update().SetUpsert(true)); err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *MongoStorage) GetStorageType() string {
	return utils.Mongo
}
