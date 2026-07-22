// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSetRoute{
		name:      "routes_profile_set",
		rpcMethod: utils.APIerSv1SetRouteProfile,
		rpcParams: &engine.RouteWithAPIOpts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdSetRoute struct {
	name      string
	rpcMethod string
	rpcParams *engine.RouteWithAPIOpts
	*CommandExecuter
}

func (self *CmdSetRoute) Name() string {
	return self.name
}

func (self *CmdSetRoute) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetRoute) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.RouteWithAPIOpts{
			RouteProfile: new(engine.RouteProfile),
			APIOpts:      map[string]any{},
		}
	}
	return self.rpcParams
}

func (self *CmdSetRoute) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetRoute) RpcResult() any {
	var s string
	return &s
}
