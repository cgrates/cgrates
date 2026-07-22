// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import "github.com/cgrates/cgrates/utils"

func init() {
	c := &CmdRemoveRoute{
		name:      "routes_profile_remove",
		rpcMethod: utils.APIerSv1RemoveRouteProfile,
		rpcParams: &utils.TenantIDWithAPIOpts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdRemoveRoute struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantIDWithAPIOpts
	*CommandExecuter
}

func (self *CmdRemoveRoute) Name() string {
	return self.name
}

func (self *CmdRemoveRoute) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRemoveRoute) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantIDWithAPIOpts{APIOpts: make(map[string]any)}
	}
	return self.rpcParams
}

func (self *CmdRemoveRoute) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRemoveRoute) RpcResult() any {
	var s string
	return &s
}
