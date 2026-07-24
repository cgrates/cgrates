// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdDebit{
		name:       "debit",
		rpcMethod:  utils.ResponderDebit,
		clientArgs: []string{"Category", "ToR", "Tenant", "Subject", "Account", "Destination", "TimeStart", "TimeEnd", "CallDuration", "FallbackSubject", "DryRun"},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdDebit struct {
	name       string
	rpcMethod  string
	rpcParams  *engine.CallDescriptorWithArgDispatcher
	clientArgs []string
	*CommandExecuter
}

func (self *CmdDebit) Name() string {
	return self.name
}

func (self *CmdDebit) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdDebit) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.CallDescriptorWithArgDispatcher{
			CallDescriptor: new(engine.CallDescriptor),
			ArgDispatcher:  new(utils.ArgDispatcher),
		}
	}
	return self.rpcParams
}

func (self *CmdDebit) PostprocessRpcParams() error {
	return nil
}

func (self *CmdDebit) RpcResult() any {
	return &engine.CallCost{}
}

func (self *CmdDebit) ClientArgs() []string {
	return self.clientArgs
}
