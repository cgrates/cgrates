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
)

type MigratorDataDB interface {
	getV1Stats() (v1st *v1Stat, err error)
	setV1Stats(x *v1Stat) (err error)
	getV2Stats() (v2 *engine.StatQueue, err error)
	setV2Stats(v2 *engine.StatQueue) (err error)
	getV1AttributeProfile() (v1attrPrf *v1AttributeProfile, err error)
	setV1AttributeProfile(x *v1AttributeProfile) (err error)
	getV2ThresholdProfile() (v2T *v2Threshold, err error)
	setV2ThresholdProfile(x *v2Threshold) (err error)
	remV2ThresholdProfile(tenant, id string) (err error)
	getV2AttributeProfile() (v2attrPrf *v2AttributeProfile, err error)
	setV2AttributeProfile(x *v2AttributeProfile) (err error)
	remV2AttributeProfile(tenant, id string) (err error)
	getV3AttributeProfile() (v3attrPrf *v3AttributeProfile, err error)
	setV3AttributeProfile(x *v3AttributeProfile) (err error)
	remV3AttributeProfile(tenant, id string) (err error)

	getV4AttributeProfile() (v4attrPrf *v4AttributeProfile, err error)
	setV4AttributeProfile(x *v4AttributeProfile) (err error)
	remV4AttributeProfile(tenant, id string) (err error)
	getV5AttributeProfile() (v5attrPrf *v6AttributeProfile, err error)

	getV1Filter() (v1Fltr *v1Filter, err error)
	setV1Filter(x *v1Filter) (err error)
	remV1Filter(tenant, id string) (err error)
	getV4Filter() (v1Fltr *engine.Filter, err error)

	getSupplier() (spl *SupplierProfile, err error)
	setSupplier(spl *SupplierProfile) (err error)
	remSupplier(tenant, id string) (err error)

	getV1ChargerProfile() (v1chrPrf *engine.ChargerProfile, err error)
	getV1DispatcherProfile() (v1chrPrf *engine.DispatcherProfile, err error)
	getV1RouteProfile() (v1chrPrf *engine.RouteProfile, err error)

	getV3Stats() (v1st *engine.StatQueueProfile, err error)
	getV3ThresholdProfile() (v2T *engine.ThresholdProfile, err error)

	DataManager() *engine.DataManager
	close()
}
