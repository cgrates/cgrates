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

package migrator

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
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

func (mgoMig *mongoStorDBMigrator) StorDB() engine.StorDB {
	return *mgoMig.storDB
}

//CDR methods
//get
func (v1ms *mongoStorDBMigrator) getV1CDR() (v1Cdr *v1Cdrs, err error) {
	if v1ms.cursor == nil {
		var cursor mongo.Cursor
		cursor, err = v1ms.mgoDB.DB().Collection(engine.ColCDRs).Find(v1ms.mgoDB.GetContext(), nil)
		if err != nil {
			return nil, err
		}
		v1ms.cursor = &cursor
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

//set
func (v1ms *mongoStorDBMigrator) setV1CDR(v1Cdr *v1Cdrs) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(engine.ColCDRs).InsertOne(v1ms.mgoDB.GetContext(), v1Cdr)
	return
}

//SMCost methods
//rename
func (v1ms *mongoStorDBMigrator) renameV1SMCosts() (err error) {
	if err = v1ms.mgoDB.DB().Collection(utils.OldSMCosts).Drop(v1ms.mgoDB.GetContext()); err != nil {
		return err
	}
	return v1ms.mgoDB.DB().RunCommand(v1ms.mgoDB.GetContext(),
		bson.D{{"create", utils.SessionsCostsTBL}}).Err()
}

func (v1ms *mongoStorDBMigrator) createV1SMCosts() (err error) {
	v1ms.mgoDB.DB().Collection(utils.OldSMCosts).Drop(v1ms.mgoDB.GetContext())
	v1ms.mgoDB.DB().Collection(utils.SessionsCostsTBL).Drop(v1ms.mgoDB.GetContext())
	return v1ms.mgoDB.DB().RunCommand(v1ms.mgoDB.GetContext(),
		bson.D{{"create", utils.OldSMCosts}, {"size", 1024}}).Err()
}

//get
func (v1ms *mongoStorDBMigrator) getV2SMCost() (v2Cost *v2SessionsCost, err error) {
	if v1ms.cursor == nil {
		var cursor mongo.Cursor
		cursor, err = v1ms.mgoDB.DB().Collection(utils.SessionsCostsTBL).Find(v1ms.mgoDB.GetContext(), nil)
		if err != nil {
			return nil, err
		}
		v1ms.cursor = &cursor
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

//set
func (v1ms *mongoStorDBMigrator) setV2SMCost(v2Cost *v2SessionsCost) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(utils.SessionsCostsTBL).InsertOne(v1ms.mgoDB.GetContext(), v2Cost)
	return
}

//remove
func (v1ms *mongoStorDBMigrator) remV2SMCost(v2Cost *v2SessionsCost) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(utils.SessionsCostsTBL).DeleteMany(v1ms.mgoDB.GetContext(), nil)
	return
}
