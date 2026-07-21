// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"errors"
	"reflect"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	ColCfg = "config"
)

func (ms *MongoStorage) GetSection(ctx *context.Context, section string, val any) error {
	return ms.query(context.TODO(), func(sctx mongo.SessionContext) error {
		sr := ms.getCol(ColCfg).FindOne(sctx, bson.M{"section": section},
			options.FindOne().SetProjection(bson.M{"cfg": 1, "_id": 0 /*"section": 0, */}))
		tmp := map[string]bson.Raw{}
		decodeErr := sr.Decode(&tmp)
		if decodeErr != nil {
			if errors.Is(decodeErr, mongo.ErrNoDocuments) {
				return nil
			}
			return decodeErr
		}
		reg := bson.NewRegistry()
		decimalType := reflect.TypeOf(utils.Decimal{})
		reg.RegisterTypeEncoder(decimalType, bsoncodec.ValueEncoderFunc(decimalEncoder))
		reg.RegisterTypeDecoder(decimalType, bsoncodec.ValueDecoderFunc(decimalDecoder))

		dec, err := bson.NewDecoder(bsonrw.NewBSONDocumentReader(tmp["cfg"]))
		if err != nil {
			return err
		}
		if err = dec.SetRegistry(reg); err != nil {
			return err
		}
		return dec.Decode(val)
	})
}

func (ms *MongoStorage) SetSection(ctx *context.Context, section string, jsn any) (err error) {
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

// Only intended for InternalDB
func (ms *MongoStorage) DumpConfigDB() (err error) {
	return utils.ErrNotImplemented
}

// Only intended for InternalDB
func (ms *MongoStorage) RewriteConfigDB() (err error) {
	return utils.ErrNotImplemented
}

// Only intended for InternalDB
func (ms *MongoStorage) BackupConfigDB(backupFolderPath string, zip bool) (err error) {
	return utils.ErrNotImplemented
}
