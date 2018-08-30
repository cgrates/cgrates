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
)

type mapMigrator struct {
	dm       *engine.DataManager
	mp       *engine.MapStorage
	dataKeys []string
	qryIdx   *int
}

func newMapMigrator(dm *engine.DataManager) (mM *mapMigrator) {
	return &mapMigrator{
		dm: dm,
		mp: dm.DataDB().(*engine.MapStorage),
	}
}

func (mM *mapMigrator) DataManager() *engine.DataManager {
	return mM.dm
}

//Account methods
//V1
//get
func (mM *mapMigrator) getv1Account() (v1Acnt *v1Account, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (mM *mapMigrator) setV1Account(x *v1Account) (err error) {
	return utils.ErrNotImplemented
}

//rem
func (mM *mapMigrator) remV1Account(id string) (err error) {
	return utils.ErrNotImplemented
}

//V2
//get
func (mM *mapMigrator) getv2Account() (v2Acnt *v2Account, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (mM *mapMigrator) setV2Account(x *v2Account) (err error) {
	return utils.ErrNotImplemented
}

//rem
func (mM *mapMigrator) remV2Account(id string) (err error) {
	return utils.ErrNotImplemented
}

//ActionPlans methods
//get
func (mM *mapMigrator) getV1ActionPlans() (v1aps *v1ActionPlans, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (mM *mapMigrator) setV1ActionPlans(x *v1ActionPlans) (err error) {
	return utils.ErrNotImplemented
}

//Actions methods
//get
func (mM *mapMigrator) getV1Actions() (v1acs *v1Actions, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (mM *mapMigrator) setV1Actions(x *v1Actions) (err error) {
	return utils.ErrNotImplemented
}

//ActionTriggers methods
//get
func (mM *mapMigrator) getV1ActionTriggers() (v1acts *v1ActionTriggers, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (mM *mapMigrator) setV1ActionTriggers(x *v1ActionTriggers) (err error) {
	return utils.ErrNotImplemented
}

//SharedGroup methods
//get
func (mM *mapMigrator) getV1SharedGroup() (v1sg *v1SharedGroup, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (mM *mapMigrator) setV1SharedGroup(x *v1SharedGroup) (err error) {
	return utils.ErrNotImplemented
}

//Stats methods
//get
func (mM *mapMigrator) getV1Stats() (v1st *v1Stat, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (mM *mapMigrator) setV1Stats(x *v1Stat) (err error) {
	return utils.ErrNotImplemented
}

//Action  methods
//get
func (mM *mapMigrator) getV2ActionTrigger() (v2at *v2ActionTrigger, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (mM *mapMigrator) setV2ActionTrigger(x *v2ActionTrigger) (err error) {
	return utils.ErrNotImplemented
}

//AttributeProfile methods
//get
func (mM *mapMigrator) getV1AttributeProfile() (v1attrPrf *v1AttributeProfile, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (mM *mapMigrator) setV1AttributeProfile(x *v1AttributeProfile) (err error) {
	return utils.ErrNotImplemented
}

//ThresholdProfile methods
//get
func (mM *mapMigrator) getV2ThresholdProfile() (v2T *v2Threshold, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (mM *mapMigrator) setV2ThresholdProfile(x *v2Threshold) (err error) {
	return utils.ErrNotImplemented
}

//rem
func (mM *mapMigrator) remV2ThresholdProfile(tenant, id string) (err error) {
	return utils.ErrNotImplemented
}
