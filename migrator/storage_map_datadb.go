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
	dm       *engine.DataManager
	iDB      *engine.InternalDB
	dataKeys []string
	qryIdx   *int
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
func (iDBMig *internalMigrator) getV5AttributeProfile() (v4attrPrf *v6AttributeProfile, err error) {
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
