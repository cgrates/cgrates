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
	"github.com/cgrates/birpc/context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	ColCfg = "config"
)

func (ms *MongoStorage) GetSection(ctx *context.Context, section string, val interface{}) error {
	return ms.query(context.TODO(), func(sctx mongo.SessionContext) (err error) {
		cur := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": section},
			options.FindOne().SetProjection(bson.M{"cfg": 1, "_id": 0 /*"section": 0, */}))
		tmp := map[string]bson.Raw{}
		if err = cur.Decode(&tmp); err != nil {
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return
		}
		return bson.UnmarshalWithRegistry(mongoReg, tmp["cfg"], val)
	})
}

func (ms *MongoStorage) SetSection(ctx *context.Context, section string, jsn interface{}) (err error) {
	return ms.query(ctx, func(sctx mongo.SessionContext) (err error) {
		_, err = ms.getCol(ColCfg).UpdateOne(sctx, bson.M{"section": section},
			bson.M{"$set": bson.M{
				"section": section,
				"cfg":     jsn}},
			options.Update().SetUpsert(true),
		)
		return err
	})
}
