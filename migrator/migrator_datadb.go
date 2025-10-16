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
)

type MigratorDataDB interface {
	getV1Stats() (v1st *v1Stat, err error)
	setV1Stats(x *v1Stat) (err error)
	getV2Stats() (v2 *engine.StatQueue, err error)
	setV2Stats(v2 *engine.StatQueue) (err error)

	getV1Filter() (v1Fltr *v1Filter, err error)
	setV1Filter(x *v1Filter) (err error)
	remV1Filter(tenant, id string) (err error)
	getV4Filter() (v1Fltr *engine.Filter, err error)

	getV1ChargerProfile() (v1chrPrf *utils.ChargerProfile, err error)

	getV3Stats() (v1st *engine.StatQueueProfile, err error)

	DataManager() *engine.DataManager
	close()
}
