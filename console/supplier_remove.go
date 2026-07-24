// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import "github.com/cgrates/cgrates/utils"

func init() {
	c := &CmdRemoveSupplier{
		name:      "supplier_remove",
		rpcMethod: utils.APIerSv1RemoveSupplierProfile,
		rpcParams: &utils.TenantIDWithCache{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdRemoveSupplier struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantIDWithCache
	*CommandExecuter
}

func (self *CmdRemoveSupplier) Name() string {
	return self.name
}

func (self *CmdRemoveSupplier) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRemoveSupplier) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantIDWithCache{}
	}
	return self.rpcParams
}

func (self *CmdRemoveSupplier) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRemoveSupplier) RpcResult() any {
	var s string
	return &s
}
