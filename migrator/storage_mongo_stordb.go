// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func newMongoStorDBMigrator(stor engine.StorDB) (mgoMig *mongoStorDBMigrator) {
	return &mongoStorDBMigrator{
		storDB: &stor,
		mgoDB:  stor.(*engine.MongoStorage),
		cursor: nil,
	}
}

type mongoStorDBMigrator struct {
	storDB *engine.StorDB
	mgoDB  *engine.MongoStorage
	cursor *mongo.Cursor
}

func (mgoMig *mongoStorDBMigrator) close() {
	mgoMig.mgoDB.Close()
}

func (mgoMig *mongoStorDBMigrator) StorDB() engine.StorDB {
	return *mgoMig.storDB
}

// CDR methods
// get
func (v1ms *mongoStorDBMigrator) getV1CDR() (v1Cdr *v1Cdrs, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(engine.ColCDRs).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v1Cdr = new(v1Cdrs)
	if err := (*v1ms.cursor).Decode(v1Cdr); err != nil {
		return nil, err
	}
	return v1Cdr, nil
}

// set
func (v1ms *mongoStorDBMigrator) setV1CDR(v1Cdr *v1Cdrs) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(engine.ColCDRs).InsertOne(v1ms.mgoDB.GetContext(), v1Cdr)
	return
}

// SMCost methods
// rename
func (v1ms *mongoStorDBMigrator) renameV1SMCosts() (err error) {
	if err = v1ms.mgoDB.DB().Collection(utils.OldSMCosts).Drop(v1ms.mgoDB.GetContext()); err != nil {
		return err
	}
	return v1ms.mgoDB.DB().RunCommand(v1ms.mgoDB.GetContext(),
		bson.D{{Key: "create", Value: utils.SessionCostsTBL}}).Err()
}

func (v1ms *mongoStorDBMigrator) createV1SMCosts() error {
	err := v1ms.mgoDB.DB().Collection(utils.OldSMCosts).Drop(v1ms.mgoDB.GetContext())
	if err != nil {
		return err
	}
	err = v1ms.mgoDB.DB().Collection(utils.SessionCostsTBL).Drop(v1ms.mgoDB.GetContext())
	if err != nil {
		return err
	}
	return v1ms.mgoDB.DB().RunCommand(v1ms.mgoDB.GetContext(),
		bson.D{
			{Key: "create", Value: utils.OldSMCosts},
			{Key: "size", Value: 1024},
			{Key: "capped", Value: true}}).Err()
}

// get
func (v1ms *mongoStorDBMigrator) getV2SMCost() (v2Cost *v2SessionsCost, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(utils.SessionCostsTBL).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v2Cost = new(v2SessionsCost)
	if err := (*v1ms.cursor).Decode(v2Cost); err != nil {
		return nil, err
	}
	return v2Cost, nil
}

// set
func (v1ms *mongoStorDBMigrator) setV2SMCost(v2Cost *v2SessionsCost) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(utils.SessionCostsTBL).InsertOne(v1ms.mgoDB.GetContext(), v2Cost)
	return
}

// remove
func (v1ms *mongoStorDBMigrator) remV2SMCost(v2Cost *v2SessionsCost) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(utils.SessionCostsTBL).DeleteMany(v1ms.mgoDB.GetContext(), bson.D{})
	return
}
