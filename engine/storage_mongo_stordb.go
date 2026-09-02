// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"errors"
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

// isMongoDuplicateError checks if the provided error is a MongoDB duplicate key error.
func isMongoDuplicateError(err error) bool {
	var e mongo.WriteException
	if errors.As(err, &e) {
		for _, we := range e.WriteErrors {
			if we.Code == 11000 { // MongoDB error code for duplicate key.
				return true
			}
		}
	}
	return false
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
