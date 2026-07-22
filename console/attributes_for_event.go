// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetAttributeForEvent{
		name:      "attributes_for_event",
		rpcMethod: utils.AttributeSv1GetAttributeForEvent,
		rpcParams: &utils.CGREvent{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdGetAttributeForEvent struct {
	name      string
	rpcMethod string
	rpcParams *utils.CGREvent
	*CommandExecuter
}

func (self *CmdGetAttributeForEvent) Name() string {
	return self.name
}

func (self *CmdGetAttributeForEvent) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetAttributeForEvent) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.CGREvent)
	}
	return self.rpcParams
}

func (self *CmdGetAttributeForEvent) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetAttributeForEvent) RpcResult() any {
	var atr engine.AttributeProfile
	return &atr
}
