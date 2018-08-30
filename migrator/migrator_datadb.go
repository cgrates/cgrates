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
	getv1Account() (v1Acnt *v1Account, err error)
	setV1Account(x *v1Account) (err error)
	remV1Account(id string) (err error)
	getV1ActionPlans() (v1aps *v1ActionPlans, err error)
	setV1ActionPlans(x *v1ActionPlans) (err error)
	getV1Actions() (v1acs *v1Actions, err error)
	setV1Actions(x *v1Actions) (err error)
	getV1ActionTriggers() (v1acts *v1ActionTriggers, err error)
	setV1ActionTriggers(x *v1ActionTriggers) (err error)
	getV1SharedGroup() (v1acts *v1SharedGroup, err error)
	setV1SharedGroup(x *v1SharedGroup) (err error)
	getV1Stats() (v1st *v1Stat, err error)
	setV1Stats(x *v1Stat) (err error)
	getV2ActionTrigger() (v2at *v2ActionTrigger, err error)
	setV2ActionTrigger(x *v2ActionTrigger) (err error)
	getv2Account() (v2Acnt *v2Account, err error)
	setV2Account(x *v2Account) (err error)
	remV2Account(id string) (err error)
	getV1AttributeProfile() (v1attrPrf *v1AttributeProfile, err error)
	setV1AttributeProfile(x *v1AttributeProfile) (err error)
	getV2ThresholdProfile() (v2T *v2Threshold, err error)
	setV2ThresholdProfile(x *v2Threshold) (err error)
	remV2ThresholdProfile(tenant, id string) (err error)
	DataManager() *engine.DataManager
}
