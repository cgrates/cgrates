// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import "github.com/cgrates/cgrates/utils"

func init() {
	c := &CmdRemoveChargers{
		name:      "chargers_profile_remove",
		rpcMethod: utils.APIerSv1RemoveChargerProfile,
		rpcParams: &utils.TenantIDWithAPIOpts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdRemoveChargers struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantIDWithAPIOpts
	*CommandExecuter
}

func (self *CmdRemoveChargers) Name() string {
	return self.name
}

func (self *CmdRemoveChargers) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRemoveChargers) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantIDWithAPIOpts{}
	}
	return self.rpcParams
}

func (self *CmdRemoveChargers) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRemoveChargers) RpcResult() any {
	var s string
	return &s
}
