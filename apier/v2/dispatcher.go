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

package v2

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewDispatcherSCDRsV2(dps *dispatchers.DispatcherService) *DispatcherSCDRsV2 {
	return &DispatcherSCDRsV2{dS: dps}
}

// Exports RPC from CDRsV2
type DispatcherSCDRsV2 struct {
	dS *dispatchers.DispatcherService
}

func (dS *DispatcherSCDRsV2) StoreSessionCost(ctx *context.Context, args *engine.ArgsV2CDRSStoreSMCost, reply *string) error {
	return dS.dS.CDRsV2StoreSessionCost(ctx, args, reply)
}

func (dS *DispatcherSCDRsV2) ProcessEvent(ctx *context.Context, args *engine.ArgV1ProcessEvent, reply *[]*utils.EventWithFlags) error {
	return dS.dS.CDRsV2ProcessEvent(ctx, args, reply)
}
