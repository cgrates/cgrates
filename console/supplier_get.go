// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetSupplier{
		name:      "supplier_get",
		rpcMethod: utils.APIerSv1GetSupplierProfile,
		rpcParams: &utils.TenantID{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdGetSupplier struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantID
	*CommandExecuter
}

func (self *CmdGetSupplier) Name() string {
	return self.name
}

func (self *CmdGetSupplier) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetSupplier) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantID{}
	}
	return self.rpcParams
}

func (self *CmdGetSupplier) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetSupplier) RpcResult() any {
	var atr engine.SupplierProfile
	return &atr
}
