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
	"fmt"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (ms *MongoStorage) GetVersions(itm string) (vrs Versions, err error) {
	fop := options.FindOne()
	if itm != "" {
		fop.SetProjection(bson.M{itm: 1, "_id": 0})
	} else {
		fop.SetProjection(bson.M{"_id": 0})
	}
	if err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
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
	return ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
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
		return ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
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
	return ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
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
	return utils.MetaMongo
}

func (ms *MongoStorage) SetCDR(cdr *utils.CGREvent, allowUpdate bool) error {
	if val, has := cdr.Event[utils.OrderID]; has && val == 0 {
		cdr.Event[utils.OrderID] = ms.counter.Next()
	}
	cdrTable := &CDR{
		Tenant:    cdr.Tenant,
		Opts:      cdr.APIOpts,
		Event:     cdr.Event,
		CreatedAt: time.Now(),
	}
	return ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		/*
			if allowUpdate {
				cdrTable.UpdatedAt = time.Now()
				_, err = ms.getCol(ColCDRs).UpdateOne(sctx,
					//bson.M{"_id": cdrTable.}
					//bson.M{CGRIDLow: utils.IfaceAsString(cdr.Event[utils.CGRID])},
					bson.M{"$set": cdrTable}, options.Update().SetUpsert(true))
				return
			}
		*/
		_, err = ms.getCol(ColCDRs).InsertOne(sctx, cdrTable)
		if err != nil && strings.Contains(err.Error(), "E11000") { // Mongo returns E11000 when key is duplicated
			err = utils.ErrExists
		}
		return
	})
}

func (ms *MongoStorage) GetCDRs(_ *context.Context, qryFltr []*Filter, opts map[string]interface{}) (cdrs []*CDR, err error) {
	fltrs := make(bson.M)
	for _, fltr := range qryFltr {
		for _, rule := range fltr.Rules {
			if !cdrQueryFilterTypes.Has(rule.Type) {
				continue
			}
			var elem string
			if strings.HasPrefix(rule.Element, utils.DynamicDataPrefix+utils.MetaReq) {
				elem = "event." + strings.TrimPrefix(rule.Element, utils.DynamicDataPrefix+utils.MetaReq+".")
			} else {
				elem = "opts." + strings.TrimPrefix(rule.Element, utils.DynamicDataPrefix+utils.MetaOpts+".")
			}
			fltrs[elem] = ms.valueQry(fltrs, elem, rule.Type, rule.Values, strings.HasPrefix(rule.Type, utils.MetaNot))
		}
	}
	ms.cleanEmptyFilters(fltrs)

	fop := options.Find()
	// cop := options.Count()

	limit, offset, maxItems, err := utils.GetPaginateOpts(opts)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve paginator opts: %w", err)
	}
	if maxItems < limit+offset {
		return nil, fmt.Errorf("sum of limit and offset exceeds maxItems")
	}
	fop.SetLimit(int64(limit))
	// cop.SetLimit(int64(limit))
	fop.SetSkip(int64(offset))
	// cop.SetSkip(int64(offset))

	// Execute query
	err = ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur, err := ms.getCol(ColCDRs).Find(sctx, fltrs, fop)
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
	if err != nil {
		return
	}
	cdrs, err = utils.Paginate(cdrs, 0, 0, int(maxItems))
	return
}

func (ms *MongoStorage) valueQry(fltrs bson.M, elem, ruleType string, values []string, not bool) (m bson.M) {
	msQuery, valChanged := getQueryType(ruleType, not, values)
	v, has := fltrs[elem]
	if !has {
		m = make(bson.M)
		fltrs[elem] = m
	} else {
		m = v.(bson.M)
	}
	if valChanged != nil {
		if val, has := m[msQuery]; has {
			m[msQuery] = append(val.([]primitive.Regex), valChanged.([]primitive.Regex)...)
		} else {
			m[msQuery] = valChanged
		}
		return
	}
	if val, has := m[msQuery]; has {
		m[msQuery] = append(val.([]string), values...)
	} else {
		m[msQuery] = values
	}
	return
}

func getQueryType(ruleType string, not bool, values []string) (msQuery string, valChanged any) {
	switch ruleType {
	case utils.MetaString, utils.MetaNotString, utils.MetaEqual, utils.MetaNotEqual:
		msQuery = "$in"
		if not {
			msQuery = "$nin"
		}
	case utils.MetaLessThan, utils.MetaLessOrEqual, utils.MetaGreaterThan, utils.MetaGreaterOrEqual:
		if ruleType == utils.MetaGreaterOrEqual {
			msQuery = "$gte"
		} else if ruleType == utils.MetaGreaterThan {
			msQuery = "$gt"
		} else if ruleType == utils.MetaLessOrEqual {
			msQuery = "$lte"
		} else if ruleType == utils.MetaLessThan {
			msQuery = "$lt"
		}
	case utils.MetaPrefix, utils.MetaNotPrefix, utils.MetaSuffix, utils.MetaNotSuffix:
		msQuery = "$in"
		if not {
			msQuery = "$nin"
		}
		regex := make([]primitive.Regex, 0, len(values))
		if ruleType == utils.MetaPrefix || ruleType == utils.MetaNotPrefix {
			for _, val := range values {
				regex = append(regex, primitive.Regex{
					Pattern: "/^" + val + "/",
				})
			}
		} else {
			for _, val := range values {
				regex = append(regex, primitive.Regex{
					Pattern: "/" + val + "$/",
				})
			}
		}
		valChanged = regex
	}
	return
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

func (ms *MongoStorage) RemoveCDRs(_ *context.Context, qryFltr []*Filter) (err error) {
	return utils.ErrNotImplemented
}
