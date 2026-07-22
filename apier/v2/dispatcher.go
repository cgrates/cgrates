// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
