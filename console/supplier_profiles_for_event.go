// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetSupplierForEvent{
		name:      "supplier_profiles_for_event",
		rpcMethod: utils.SupplierSv1GetSupplierProfilesForEvent,
		rpcParams: &utils.CGREventWithArgDispatcher{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdGetSupplierForEvent struct {
	name      string
	rpcMethod string
	rpcParams *utils.CGREventWithArgDispatcher
	*CommandExecuter
}

func (self *CmdGetSupplierForEvent) Name() string {
	return self.name
}

func (self *CmdGetSupplierForEvent) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetSupplierForEvent) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.CGREventWithArgDispatcher{
			CGREvent:      new(utils.CGREvent),
			ArgDispatcher: new(utils.ArgDispatcher)}
	}
	return self.rpcParams
}

func (self *CmdGetSupplierForEvent) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetSupplierForEvent) RpcResult() any {
	var atr []*engine.SupplierProfile
	return &atr
}
