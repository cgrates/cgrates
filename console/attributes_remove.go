// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import "github.com/cgrates/cgrates/utils"

func init() {
	c := &CmdRemoveAttributes{
		name:      "attributes_remove",
		rpcMethod: utils.APIerSv1RemoveAttributeProfile,
		rpcParams: &utils.TenantIDWithCache{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdRemoveAttributes struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantIDWithCache
	*CommandExecuter
}

func (self *CmdRemoveAttributes) Name() string {
	return self.name
}

func (self *CmdRemoveAttributes) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRemoveAttributes) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantIDWithCache{}
	}
	return self.rpcParams
}

func (self *CmdRemoveAttributes) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRemoveAttributes) RpcResult() any {
	var s string
	return &s
}
