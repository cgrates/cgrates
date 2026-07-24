// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewCoreSv1(cS *engine.CoreService) *CoreSv1 {
	return &CoreSv1{cS: cS}
}

// Exports RPC from RLs
type CoreSv1 struct {
	cS *engine.CoreService
}

// Call implements birpc.ClientConnector interface for internal RPC
func (cS *CoreSv1) Call(ctx *context.Context, serviceMethod string,
	args any, reply any) error {
	return utils.APIerRPCCall(cS, serviceMethod, args, reply)
}

func (cS *CoreSv1) Status(arg *utils.TenantWithArgDispatcher, reply *map[string]any) error {
	return cS.cS.Status(arg, reply)
}

// Ping used to determinate if component is active
func (cS *CoreSv1) Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error {
	*reply = utils.Pong
	return nil
}

// Sleep is used to test the concurrent requests mechanism
func (cS *CoreSv1) Sleep(arg *utils.DurationArgs, reply *string) error {
	time.Sleep(arg.DurationTime)
	*reply = utils.OK
	return nil
}
