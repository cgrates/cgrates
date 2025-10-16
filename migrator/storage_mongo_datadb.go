/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package migrator

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	v2AccountsCol          = "accounts"
	v1ActionTriggersCol    = "action_triggers"
	v1AttributeProfilesCol = "attribute_profiles"
	v2ThresholdProfileCol  = "threshold_profiles"
	v1AliasCol             = "aliases"
	v1UserCol              = "users"
	v1DerivedChargersCol   = "derived_chargers"
	v2StatsCol             = "statqueues"
)

type mongoMigrator struct {
	dm     *engine.DataManager
	mgoDB  *engine.MongoStorage
	cursor *mongo.Cursor
}

func newMongoMigrator(dm *engine.DataManager) (mgoMig *mongoMigrator) {
	var mgoDB *engine.MongoStorage
	for _, dbInf := range dm.DataDB() {
		var canCast bool
		if mgoDB, canCast = dbInf.(*engine.MongoStorage); canCast {
			return &mongoMigrator{
				dm:     dm,
				mgoDB:  mgoDB,
				cursor: nil,
			}
		}
	}
	return nil
}

func (v1ms *mongoMigrator) close() {
	v1ms.mgoDB.Close()
}

func (v1ms *mongoMigrator) DataManager() *engine.DataManager {
	return v1ms.dm
}

// Stats methods
// get
func (v1ms *mongoMigrator) getV1Stats() (v1st *v1Stat, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(utils.CDRsStatsPrefix).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v1st = new(v1Stat)
	if err := (*v1ms.cursor).Decode(v1st); err != nil {
		return nil, err
	}
	return v1st, nil
}

func (v1ms *mongoMigrator) getV3Stats() (v1st *engine.StatQueueProfile, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(engine.ColSqp).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v1st = new(engine.StatQueueProfile)
	if err := (*v1ms.cursor).Decode(v1st); err != nil {
		return nil, err
	}
	return v1st, nil
}

// set
func (v1ms *mongoMigrator) setV1Stats(x *v1Stat) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(utils.CDRsStatsPrefix).InsertOne(v1ms.mgoDB.GetContext(), x)
	return
}

// get V2
func (v1ms *mongoMigrator) getV2Stats() (v2 *engine.StatQueue, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(v2StatsCol).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v2 = new(engine.StatQueue)
	if err := (*v1ms.cursor).Decode(v2); err != nil {
		return nil, err
	}
	return v2, nil
}

// set v2
func (v1ms *mongoMigrator) setV2Stats(v2 *engine.StatQueue) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(v2StatsCol).InsertOne(v1ms.mgoDB.GetContext(), v2)
	return
}

// Filter Methods
// get
func (v1ms *mongoMigrator) getV1Filter() (v1Fltr *v1Filter, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(engine.ColFlt).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v1Fltr = new(v1Filter)
	if err := (*v1ms.cursor).Decode(v1Fltr); err != nil {
		return nil, err
	}
	return
}

func (v1ms *mongoMigrator) getV4Filter() (v4Fltr *engine.Filter, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(engine.ColFlt).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v4Fltr = new(engine.Filter)
	if err := (*v1ms.cursor).Decode(v4Fltr); err != nil {
		return nil, err
	}
	return
}

// set
func (v1ms *mongoMigrator) setV1Filter(x *v1Filter) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(engine.ColFlt).InsertOne(v1ms.mgoDB.GetContext(), x)
	return
}

// rem
func (v1ms *mongoMigrator) remV1Filter(tenant, id string) (err error) {
	_, err = v1ms.mgoDB.DB().Collection(engine.ColFlt).DeleteOne(v1ms.mgoDB.GetContext(), bson.M{"tenant": tenant, "id": id})
	return
}

func (v1ms *mongoMigrator) getV1ChargerProfile() (v1chrPrf *utils.ChargerProfile, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(engine.ColCpp).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v1chrPrf = new(utils.ChargerProfile)
	if err := (*v1ms.cursor).Decode(v1chrPrf); err != nil {
		return nil, err
	}
	return
}

func (v1ms *mongoMigrator) getV1RouteProfile() (v1dppPrf *utils.RouteProfile, err error) {
	if v1ms.cursor == nil {
		v1ms.cursor, err = v1ms.mgoDB.DB().Collection(engine.ColRpp).Find(v1ms.mgoDB.GetContext(), bson.D{})
		if err != nil {
			return nil, err
		}
	}
	if !(*v1ms.cursor).Next(v1ms.mgoDB.GetContext()) {
		(*v1ms.cursor).Close(v1ms.mgoDB.GetContext())
		v1ms.cursor = nil
		return nil, utils.ErrNoMoreData
	}
	v1dppPrf = new(utils.RouteProfile)
	if err := (*v1ms.cursor).Decode(v1dppPrf); err != nil {
		return nil, err
	}
	return
}
