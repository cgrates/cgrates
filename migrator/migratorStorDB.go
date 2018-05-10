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

type MigratorStorDB interface {
	getV1CDR() (v1Cdr *v1Cdrs, err error)
	setV1CDR(v1Cdr *v1Cdrs) (err error)
	createV1SMCosts() (err error)
	renameV1SMCosts() (err error)
	getV2SMCost() (v2Cost *v2SessionsCost, err error)
	setV2SMCost(v2Cost *v2SessionsCost) (err error)
	remV2SMCost(v2Cost *v2SessionsCost) (err error)
	StorDB() engine.StorDB
}
