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

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewCDRsV1 constructs the RPC Object for CDRsV1
func NewCDRsV1(cdrS *engine.CDRServer) *CDRsV1 {
	return &CDRsV1{cdrS: cdrS}
}

// CDRsV1 Exports RPC from CDRs
type CDRsV1 struct {
	ping
	cdrS *engine.CDRServer
}

// ProcessEvent
func (cdrSv1 *CDRsV1) ProcessEvent(ctx *context.Context, args *utils.CGREvent,
	reply *string) error {
	return cdrSv1.cdrS.V1ProcessEvent(ctx, args, reply)
}
