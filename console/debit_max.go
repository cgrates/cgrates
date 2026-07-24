// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdMaxDebit{
		name:       "debit_max",
		rpcMethod:  utils.ResponderMaxDebit,
		clientArgs: []string{"Category", "ToR", "Tenant", "Subject", "Account", "Destination", "TimeStart", "TimeEnd", "CallDuration", "FallbackSubject"},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdMaxDebit struct {
	name       string
	rpcMethod  string
	rpcParams  *engine.CallDescriptorWithArgDispatcher
	clientArgs []string
	*CommandExecuter
}

func (self *CmdMaxDebit) Name() string {
	return self.name
}

func (self *CmdMaxDebit) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdMaxDebit) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.CallDescriptorWithArgDispatcher{
			CallDescriptor: new(engine.CallDescriptor),
			ArgDispatcher:  new(utils.ArgDispatcher),
		}
	}
	return self.rpcParams
}

func (self *CmdMaxDebit) PostprocessRpcParams() error {
	return nil
}

func (self *CmdMaxDebit) RpcResult() any {
	return &engine.CallCost{}
}

func (self *CmdMaxDebit) ClientArgs() []string {
	return self.clientArgs
}
