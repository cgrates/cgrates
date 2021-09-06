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

package apis

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/utils"
)

func NewEeSv1(eesV1 *ees.EventExporterS) *EeSv1 {
	return &EeSv1{ees: eesV1}
}

// EeSv1 exports RPC from RLs
type EeSv1 struct {
	ees *ees.EventExporterS
	ping
}

func (cS *EeSv1) ProcessEvent(ctx *context.Context, cgrEv *utils.CGREventWithEeIDs, rply *map[string]map[string]interface{}) error {
	return cS.ees.V1ProcessEvent(ctx, cgrEv, rply)
}
