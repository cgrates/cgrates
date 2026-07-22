// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetRouteForEvent{
		name:      "routes_profiles_for_event",
		rpcMethod: utils.RouteSv1GetRouteProfilesForEvent,
		rpcParams: &utils.CGREvent{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdGetRouteForEvent struct {
	name      string
	rpcMethod string
	rpcParams *utils.CGREvent
	*CommandExecuter
}

func (self *CmdGetRouteForEvent) Name() string {
	return self.name
}

func (self *CmdGetRouteForEvent) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetRouteForEvent) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.CGREvent)
	}
	return self.rpcParams
}

func (self *CmdGetRouteForEvent) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetRouteForEvent) RpcResult() any {
	var atr []*engine.RouteProfile
	return &atr
}
