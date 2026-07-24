// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSuppliersSort{
		name:      "suppliers",
		rpcMethod: utils.SupplierSv1GetSuppliers,
		rpcParams: &engine.ArgsGetSuppliers{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdSuppliersSort struct {
	name      string
	rpcMethod string
	rpcParams *engine.ArgsGetSuppliers
	*CommandExecuter
}

func (self *CmdSuppliersSort) Name() string {
	return self.name
}

func (self *CmdSuppliersSort) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSuppliersSort) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.ArgsGetSuppliers{ArgDispatcher: new(utils.ArgDispatcher)}
	}
	return self.rpcParams
}

func (self *CmdSuppliersSort) PostprocessRpcParams() error {
	if self.rpcParams != nil && self.rpcParams.CGREvent != nil &&
		self.rpcParams.CGREvent.Time == nil {
		self.rpcParams.CGREvent.Time = utils.TimePointer(time.Now())
	}
	return nil
}

func (self *CmdSuppliersSort) RpcResult() any {
	var atr *engine.SortedSuppliers
	return &atr
}
