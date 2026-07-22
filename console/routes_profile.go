// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetRoute{
		name:      "routes_profile",
		rpcMethod: utils.APIerSv1GetRouteProfile,
		rpcParams: &utils.TenantID{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdGetRoute struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantID
	*CommandExecuter
}

func (self *CmdGetRoute) Name() string {
	return self.name
}

func (self *CmdGetRoute) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetRoute) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantID{}
	}
	return self.rpcParams
}

func (self *CmdGetRoute) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetRoute) RpcResult() any {
	var atr engine.RouteProfile
	return &atr
}
