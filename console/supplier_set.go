// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSetSupplier{
		name:      "supplier_set",
		rpcMethod: utils.APIerSv1SetSupplierProfile,
		rpcParams: &v1.SupplierWithCache{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdSetSupplier struct {
	name      string
	rpcMethod string
	rpcParams *v1.SupplierWithCache
	*CommandExecuter
}

func (self *CmdSetSupplier) Name() string {
	return self.name
}

func (self *CmdSetSupplier) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetSupplier) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.SupplierWithCache{SupplierProfile: new(engine.SupplierProfile)}
	}
	return self.rpcParams
}

func (self *CmdSetSupplier) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetSupplier) RpcResult() any {
	var s string
	return &s
}
