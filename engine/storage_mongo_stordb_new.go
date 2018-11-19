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
	// "fmt"
	"context"
	"regexp"
	"strings"
	// "time"

	"github.com/cgrates/cgrates/utils"

	"github.com/mongodb/mongo-go-driver/bson"
	// "github.com/mongodb/mongo-go-driver/bson/objectid"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/options"
	"github.com/mongodb/mongo-go-driver/x/bsonx"
)

func (ms *MongoStorageNew) GetTpIds(colName string) (tpids []string, err error) {
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
			col.Close(sctx)
			return nil
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

func (ms *MongoStorageNew) GetTpTableIds(tpid, table string, distinct utils.TPDistinctIds, filter map[string]string, pag *utils.Paginator) ([]string, error) {
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
