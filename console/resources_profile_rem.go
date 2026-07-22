// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import "github.com/cgrates/cgrates/utils"

func init() {
	c := &CmdRemoveResource{
		name:      "resources_profile_remove",
		rpcMethod: utils.APIerSv1RemoveResourceProfile,
		rpcParams: &utils.TenantIDWithAPIOpts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdRemoveResource struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantIDWithAPIOpts
	*CommandExecuter
}

func (self *CmdRemoveResource) Name() string {
	return self.name
}

func (self *CmdRemoveResource) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRemoveResource) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantIDWithAPIOpts{APIOpts: make(map[string]any)}
	}
	return self.rpcParams
}

func (self *CmdRemoveResource) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRemoveResource) RpcResult() any {
	var s string
	return &s
}
