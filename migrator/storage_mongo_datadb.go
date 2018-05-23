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

const (
	v2AccountsCol          = "accounts"
	v1ActionTriggersCol    = "action_triggers"
	v1AttributeProfilesCol = "attribute_profiles"
	v2ThresholdProfileCol  = "threshold_profiles"
)

type mongoMigrator struct {
	dm      *engine.DataManager
	mgoDB   *engine.MongoStorage
	qryIter *mgo.Iter
}

type AcKeyValue struct {
	Key   string
	Value v1Actions
}
type AtKeyValue struct {
	Key   string
	Value v1ActionPlans
}

func newMongoMigrator(dm *engine.DataManager) (mgoMig *mongoMigrator) {
	return &mongoMigrator{
		dm:      dm,
		mgoDB:   dm.DataDB().(*engine.MongoStorage),
		qryIter: nil,
	}
}

func (mgoMig *mongoMigrator) DataManager() *engine.DataManager {
	return mgoMig.dm
}

//Account methods
//V1
//get
func (v1ms *mongoMigrator) getv1Account() (v1Acnt *v1Account, err error) {
	if v1ms.qryIter == nil {
		v1ms.qryIter = v1ms.mgoDB.DB().C(v1AccountDBPrefix).Find(nil).Iter()
	}
	v1ms.qryIter.Next(&v1Acnt)

	if v1Acnt == nil {
		v1ms.qryIter = nil
		return nil, utils.ErrNoMoreData

	}
	return v1Acnt, nil
}

//set
func (v1ms *mongoMigrator) setV1Account(x *v1Account) (err error) {
	if err := v1ms.mgoDB.DB().C(v1AccountDBPrefix).Insert(x); err != nil {
		return err
	}
	return
}

//V2
//get
func (v1ms *mongoMigrator) getv2Account() (v2Acnt *v2Account, err error) {
	if v1ms.qryIter == nil {
		v1ms.qryIter = v1ms.mgoDB.DB().C(v2AccountsCol).Find(nil).Iter()
	}
	v1ms.qryIter.Next(&v2Acnt)

	if v2Acnt == nil {
		v1ms.qryIter = nil
		return nil, utils.ErrNoMoreData

	}
	return v2Acnt, nil
}

//set
func (v1ms *mongoMigrator) setV2Account(x *v2Account) (err error) {
	if err := v1ms.mgoDB.DB().C(v2AccountsCol).Insert(x); err != nil {
		return err
	}
	return
}

//Action methods
//get
func (v1ms *mongoMigrator) getV1ActionPlans() (v1aps *v1ActionPlans, err error) {
	var strct *AtKeyValue
	if v1ms.qryIter == nil {
		v1ms.qryIter = v1ms.mgoDB.DB().C("actiontimings").Find(nil).Iter()
	}
	v1ms.qryIter.Next(&strct)
	if strct == nil {
		v1ms.qryIter = nil
		return nil, utils.ErrNoMoreData
	}
	v1aps = &strct.Value
	return v1aps, nil
}

//set
func (v1ms *mongoMigrator) setV1ActionPlans(x *v1ActionPlans) (err error) {
	key := utils.ACTION_PLAN_PREFIX + (*x)[0].Id
	if err := v1ms.mgoDB.DB().C("actiontimings").Insert(&AtKeyValue{key, *x}); err != nil {
		return err
	}
	return
}

//Actions methods
//get
func (v1ms *mongoMigrator) getV1Actions() (v1acs *v1Actions, err error) {
	var strct *AcKeyValue
	if v1ms.qryIter == nil {
		v1ms.qryIter = v1ms.mgoDB.DB().C("actions").Find(nil).Iter()
	}
	v1ms.qryIter.Next(&strct)
	if strct == nil {
		v1ms.qryIter = nil
		return nil, utils.ErrNoMoreData
	}

	v1acs = &strct.Value
	return v1acs, nil
}

//set
func (v1ms *mongoMigrator) setV1Actions(x *v1Actions) (err error) {
	key := utils.ACTION_PREFIX + (*x)[0].Id
	if err := v1ms.mgoDB.DB().C("actions").Insert(&AcKeyValue{key, *x}); err != nil {
		return err
	}
	return
}

//ActionTriggers methods
//get
func (v1ms *mongoMigrator) getV1ActionTriggers() (v1acts *v1ActionTriggers, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (v1ms *mongoMigrator) setV1ActionTriggers(x *v1ActionTriggers) (err error) {
	return utils.ErrNotImplemented
}

//Actions methods
//get
func (v1ms *mongoMigrator) getV1SharedGroup() (v1sg *v1SharedGroup, err error) {
	if v1ms.qryIter == nil {
		v1ms.qryIter = v1ms.mgoDB.DB().C(utils.SHARED_GROUP_PREFIX).Find(nil).Iter()
	}
	v1ms.qryIter.Next(&v1sg)
	if v1sg == nil {
		v1ms.qryIter = nil
		return nil, utils.ErrNoMoreData

	}
	return v1sg, nil
}

//set
func (v1ms *mongoMigrator) setV1SharedGroup(x *v1SharedGroup) (err error) {
	if err := v1ms.mgoDB.DB().C(utils.SHARED_GROUP_PREFIX).Insert(x); err != nil {
		return err
	}
	return
}

//Stats methods
//get
func (v1ms *mongoMigrator) getV1Stats() (v1st *v1Stat, err error) {
	if v1ms.qryIter == nil {
		v1ms.qryIter = v1ms.mgoDB.DB().C(utils.CDR_STATS_PREFIX).Find(nil).Iter()
	}
	v1ms.qryIter.Next(&v1st)
	if v1st == nil {
		v1ms.qryIter = nil
		return nil, utils.ErrNoMoreData

	}
	return v1st, nil
}

//set
func (v1ms *mongoMigrator) setV1Stats(x *v1Stat) (err error) {
	if err := v1ms.mgoDB.DB().C(utils.CDR_STATS_PREFIX).Insert(x); err != nil {
		return err
	}
	return
}

//Stats methods
//get
func (v1ms *mongoMigrator) getV2ActionTrigger() (v2at *v2ActionTrigger, err error) {
	if v1ms.qryIter == nil {
		v1ms.qryIter = v1ms.mgoDB.DB().C(v1ActionTriggersCol).Find(nil).Iter()
	}
	v1ms.qryIter.Next(&v2at)
	if v2at == nil {
		v1ms.qryIter = nil
		return nil, utils.ErrNoMoreData

	}
	return v2at, nil
}

//set
func (v1ms *mongoMigrator) setV2ActionTrigger(x *v2ActionTrigger) (err error) {
	if err := v1ms.mgoDB.DB().C(v1ActionTriggersCol).Insert(x); err != nil {
		return err
	}
	return
}

//AttributeProfile methods
//get
func (v1ms *mongoMigrator) getV1AttributeProfile() (v1attrPrf *v1AttributeProfile, err error) {
	if v1ms.qryIter == nil {
		v1ms.qryIter = v1ms.mgoDB.DB().C(v1AttributeProfilesCol).Find(nil).Iter()
	}
	v1ms.qryIter.Next(&v1attrPrf)
	if v1attrPrf == nil {
		v1ms.qryIter = nil
		return nil, utils.ErrNoMoreData

	}
	return v1attrPrf, nil
}

//set
func (v1ms *mongoMigrator) setV1AttributeProfile(x *v1AttributeProfile) (err error) {
	if err := v1ms.mgoDB.DB().C(v1AttributeProfilesCol).Insert(x); err != nil {
		return err
	}
	return
}

//ThresholdProfile methods
//get
func (v1ms *mongoMigrator) getV2ThresholdProfile() (v2T *v2Threshold, err error) {
	if v1ms.qryIter == nil {
		v1ms.qryIter = v1ms.mgoDB.DB().C(v2ThresholdProfileCol).Find(nil).Iter()
	}
	v1ms.qryIter.Next(&v2T)
	if v2T == nil {
		v1ms.qryIter = nil
		return nil, utils.ErrNoMoreData

	}
	return v2T, nil
}

//set
func (v1ms *mongoMigrator) setV2ThresholdProfile(x *v2Threshold) (err error) {
	if err := v1ms.mgoDB.DB().C(v2ThresholdProfileCol).Insert(x); err != nil {
		return err
	}
	return
}

//rem
func (v1ms *mongoMigrator) remV2ThresholdProfile(tenant, id string) (err error) {
	return v1ms.mgoDB.DB().C(v2ThresholdProfileCol).Remove(bson.M{"tenant": tenant, "id": id})
}
