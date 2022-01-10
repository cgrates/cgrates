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

//SMCost methods
//rename
func (v1ms *mongoStorDBMigrator) renameV1SMCosts() (err error) {
	if err = v1ms.mgoDB.DB().Collection(utils.OldSMCosts).Drop(v1ms.mgoDB.GetContext()); err != nil {
		return err
	}
	return v1ms.mgoDB.DB().RunCommand(v1ms.mgoDB.GetContext(),
		bson.D{{Key: "create", Value: utils.SessionCostsTBL}}).Err()
}

func (v1ms *mongoStorDBMigrator) createV1SMCosts() (err error) {
	v1ms.mgoDB.DB().Collection(utils.OldSMCosts).Drop(v1ms.mgoDB.GetContext())
	v1ms.mgoDB.DB().Collection(utils.SessionCostsTBL).Drop(v1ms.mgoDB.GetContext())
	return v1ms.mgoDB.DB().RunCommand(v1ms.mgoDB.GetContext(),
		bson.D{{Key: "create", Value: utils.OldSMCosts}, {Key: "size", Value: 1024}, {Key: "capped", Value: true}}).Err()
}
