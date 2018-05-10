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
	"github.com/cgrates/mgo"
	"github.com/cgrates/mgo/bson"
)

func newMongoStorDBMigrator(stor engine.StorDB) (mgoMig *mongoStorDBMigrator) {
	return &mongoStorDBMigrator{
		storDB:  &stor,
		mgoDB:   stor.(*engine.MongoStorage),
		qryIter: nil,
	}
}

type mongoStorDBMigrator struct {
	storDB  *engine.StorDB
	mgoDB   *engine.MongoStorage
	qryIter *mgo.Iter
}

func (mgoMig *mongoStorDBMigrator) StorDB() engine.StorDB {
	return *mgoMig.storDB
}

//CDR methods
//get
func (v1ms *mongoStorDBMigrator) getV1CDR() (v1Cdr *v1Cdrs, err error) {
	if v1ms.qryIter == nil {
		v1ms.qryIter = v1ms.mgoDB.DB().C(engine.ColCDRs).Find(nil).Iter()
	}
	v1ms.qryIter.Next(&v1Cdr)

	if v1Cdr == nil {
		v1ms.qryIter = nil
		return nil, utils.ErrNoMoreData

	}
	return v1Cdr, nil
}

//set
func (v1ms *mongoStorDBMigrator) setV1CDR(v1Cdr *v1Cdrs) (err error) {
	if err = v1ms.mgoDB.DB().C(engine.ColCDRs).Insert(v1Cdr); err != nil {
		return err
	}
	return
}

//SMCost methods
//rename
func (v1ms *mongoStorDBMigrator) renameV1SMCosts() (err error) {
	if err = v1ms.mgoDB.DB().C(utils.OldSMCosts).DropCollection(); err != nil {
		return err
	}
	result := make(map[string]string)
	return v1ms.mgoDB.DB().Run(bson.D{{"create", utils.SessionsCostsTBL}}, result)
}

func (v1ms *mongoStorDBMigrator) createV1SMCosts() (err error) {
	err = v1ms.mgoDB.DB().C(utils.OldSMCosts).DropCollection()
	err = v1ms.mgoDB.DB().C(utils.SessionsCostsTBL).DropCollection()
	result := make(map[string]string)
	return v1ms.mgoDB.DB().Run(bson.D{{"create", utils.OldSMCosts},
		{"size", 1024}}, result)
}

//get
func (v1ms *mongoStorDBMigrator) getV2SMCost() (v2Cost *v2SessionsCost, err error) {
	if v1ms.qryIter == nil {
		v1ms.qryIter = v1ms.mgoDB.DB().C(utils.SessionsCostsTBL).Find(nil).Iter()
	}
	v1ms.qryIter.Next(&v2Cost)

	if v2Cost == nil {
		v1ms.qryIter = nil
		return nil, utils.ErrNoMoreData

	}
	return v2Cost, nil
}

//set
func (v1ms *mongoStorDBMigrator) setV2SMCost(v2Cost *v2SessionsCost) (err error) {
	if err = v1ms.mgoDB.DB().C(utils.SessionsCostsTBL).Insert(v2Cost); err != nil {
		return err
	}
	return
}

//remove
func (v1ms *mongoStorDBMigrator) remV2SMCost(v2Cost *v2SessionsCost) (err error) {
	if err = v1ms.mgoDB.DB().C(utils.SessionsCostsTBL).Remove(nil); err != nil {
		return err
	}
	return
}
