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

type internalMigrator struct {
	dm  *engine.DataManager
	iDB *engine.InternalDB
}

func newInternalMigrator(dm *engine.DataManager) (iDBMig *internalMigrator) {
	return &internalMigrator{
		dm:  dm,
		iDB: dm.DataDB().(*engine.InternalDB),
	}
}

func (iDBMig *internalMigrator) DataManager() *engine.DataManager {
	return iDBMig.dm
}

//Account methods
//V1
//get
func (iDBMig *internalMigrator) getv1Account() (v1Acnt *v1Account, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (iDBMig *internalMigrator) setV1Account(x *v1Account) (err error) {
	return utils.ErrNotImplemented
}

//rem
func (iDBMig *internalMigrator) remV1Account(id string) (err error) {
	return utils.ErrNotImplemented
}

//V2
//get
func (iDBMig *internalMigrator) getv2Account() (v2Acnt *v2Account, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (iDBMig *internalMigrator) setV2Account(x *v2Account) (err error) {
	return utils.ErrNotImplemented
}

//rem
func (iDBMig *internalMigrator) remV2Account(id string) (err error) {
	return utils.ErrNotImplemented
}

//ActionPlans methods
//get
func (iDBMig *internalMigrator) getV1ActionPlans() (v1aps *v1ActionPlans, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (iDBMig *internalMigrator) setV1ActionPlans(x *v1ActionPlans) (err error) {
	return utils.ErrNotImplemented
}

//rem
func (iDBMig *internalMigrator) remV1ActionPlans(x *v1ActionPlans) (err error) {
	return utils.ErrNotImplemented
}

//Actions methods
//get
func (iDBMig *internalMigrator) getV1Actions() (v1acs *v1Actions, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (iDBMig *internalMigrator) setV1Actions(x *v1Actions) (err error) {
	return utils.ErrNotImplemented
}

//rem
func (iDBMig *internalMigrator) remV1Actions(x v1Actions) (err error) {
	return utils.ErrNotImplemented
}

//ActionTriggers methods
//get
func (iDBMig *internalMigrator) getV1ActionTriggers() (v1acts *v1ActionTriggers, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (iDBMig *internalMigrator) setV1ActionTriggers(x *v1ActionTriggers) (err error) {
	return utils.ErrNotImplemented
}

//rem
func (iDBMig *internalMigrator) remV1ActionTriggers(x *v1ActionTriggers) (err error) {
	return utils.ErrNotImplemented
}

//SharedGroup methods
//get
func (iDBMig *internalMigrator) getV1SharedGroup() (v1sg *v1SharedGroup, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (iDBMig *internalMigrator) setV1SharedGroup(x *v1SharedGroup) (err error) {
	return utils.ErrNotImplemented
}

//Stats methods
//get
func (iDBMig *internalMigrator) getV1Stats() (v1st *v1Stat, err error) {
	return nil, utils.ErrNotImplemented
}

func (iDBMig *internalMigrator) getV3Stats() (v1st *engine.StatQueueProfile, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (iDBMig *internalMigrator) setV1Stats(x *v1Stat) (err error) {
	return utils.ErrNotImplemented
}

func (iDBMig *internalMigrator) getV2Stats() (v2 *engine.StatQueue, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (iDBMig *internalMigrator) setV2Stats(v2 *engine.StatQueue) (err error) {
	return utils.ErrNotImplemented
}

//Action  methods
//get
func (iDBMig *internalMigrator) getV2ActionTrigger() (v2at *v2ActionTrigger, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (iDBMig *internalMigrator) setV2ActionTrigger(x *v2ActionTrigger) (err error) {
	return utils.ErrNotImplemented
}

//AttributeProfile methods
//get
func (iDBMig *internalMigrator) getV1AttributeProfile() (v1attrPrf *v1AttributeProfile, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (iDBMig *internalMigrator) setV1AttributeProfile(x *v1AttributeProfile) (err error) {
	return utils.ErrNotImplemented
}

//ThresholdProfile methods
//get
func (iDBMig *internalMigrator) getV2ThresholdProfile() (v2T *v2Threshold, err error) {
	return nil, utils.ErrNotImplemented
}

func (iDBMig *internalMigrator) getV3ThresholdProfile() (v2T *engine.ThresholdProfile, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (iDBMig *internalMigrator) setV2ThresholdProfile(x *v2Threshold) (err error) {
	return utils.ErrNotImplemented
}

//rem
func (iDBMig *internalMigrator) remV2ThresholdProfile(tenant, id string) (err error) {
	return utils.ErrNotImplemented
}

//Alias methods
//get
func (iDBMig *internalMigrator) getV1Alias() (v1a *v1Alias, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (iDBMig *internalMigrator) setV1Alias(al *v1Alias) (err error) {
	return utils.ErrNotImplemented
}

//rem
func (iDBMig *internalMigrator) remV1Alias(key string) (err error) {
	return utils.ErrNotImplemented
}

// User methods
//get
func (iDBMig *internalMigrator) getV1User() (v1u *v1UserProfile, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (iDBMig *internalMigrator) setV1User(us *v1UserProfile) (err error) {
	return utils.ErrNotImplemented
}

//rem
func (iDBMig *internalMigrator) remV1User(key string) (err error) {
	return utils.ErrNotImplemented
}

// DerivedChargers methods
//get
func (iDBMig *internalMigrator) getV1DerivedChargers() (v1d *v1DerivedChargersWithKey, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (iDBMig *internalMigrator) setV1DerivedChargers(dc *v1DerivedChargersWithKey) (err error) {
	return utils.ErrNotImplemented
}

//rem
func (iDBMig *internalMigrator) remV1DerivedChargers(key string) (err error) {
	return utils.ErrNotImplemented
}

//AttributeProfile methods
//get
func (iDBMig *internalMigrator) getV2AttributeProfile() (v2attrPrf *v2AttributeProfile, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (iDBMig *internalMigrator) setV2AttributeProfile(x *v2AttributeProfile) (err error) {
	return utils.ErrNotImplemented
}

//rem
func (iDBMig *internalMigrator) remV2AttributeProfile(tenant, id string) (err error) {
	return utils.ErrNotImplemented
}

//AttributeProfile methods
//get
func (iDBMig *internalMigrator) getV3AttributeProfile() (v3attrPrf *v3AttributeProfile, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (iDBMig *internalMigrator) setV3AttributeProfile(x *v3AttributeProfile) (err error) {
	return utils.ErrNotImplemented
}

//rem
func (iDBMig *internalMigrator) remV3AttributeProfile(tenant, id string) (err error) {
	return utils.ErrNotImplemented
}

//AttributeProfile methods
//get
func (iDBMig *internalMigrator) getV4AttributeProfile() (v4attrPrf *v4AttributeProfile, err error) {
	return nil, utils.ErrNotImplemented
}
func (iDBMig *internalMigrator) getV5AttributeProfile() (v4attrPrf *engine.AttributeProfile, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (iDBMig *internalMigrator) setV4AttributeProfile(x *v4AttributeProfile) (err error) {
	return utils.ErrNotImplemented
}

//rem
func (iDBMig *internalMigrator) remV4AttributeProfile(tenant, id string) (err error) {
	return utils.ErrNotImplemented
}

// Filter Methods
//get
func (iDBMig *internalMigrator) getV1Filter() (v1Fltr *v1Filter, err error) {
	return nil, utils.ErrNotImplemented
}

func (iDBMig *internalMigrator) getV4Filter() (v1Fltr *engine.Filter, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (iDBMig *internalMigrator) setV1Filter(x *v1Filter) (err error) {
	return utils.ErrNotImplemented
}

//rem
func (iDBMig *internalMigrator) remV1Filter(tenant, id string) (err error) {
	return utils.ErrNotImplemented
}

// Supplier Methods
//get
func (iDBMig *internalMigrator) getSupplier() (spl *SupplierProfile, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (iDBMig *internalMigrator) setSupplier(spl *SupplierProfile) (err error) {
	return utils.ErrNotImplemented
}

//rem
func (iDBMig *internalMigrator) remSupplier(tenant, id string) (err error) {
	return utils.ErrNotImplemented
}

func (iDBMig *internalMigrator) close() {}

func (iDBMig *internalMigrator) getV1ChargerProfile() (v1chrPrf *engine.ChargerProfile, err error) {
	return nil, utils.ErrNotImplemented
}

func (iDBMig *internalMigrator) getV1DispatcherProfile() (v1chrPrf *engine.DispatcherProfile, err error) {
	return nil, utils.ErrNotImplemented
}

func (iDBMig *internalMigrator) getV1RouteProfile() (v1chrPrf *engine.RouteProfile, err error) {
	return nil, utils.ErrNotImplemented
}
