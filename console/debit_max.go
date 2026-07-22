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
		clientArgs: []string{utils.Category, utils.ToR, utils.Tenant, utils.Subject, utils.AccountField, utils.Destination, utils.TimeStart, utils.TimeEnd, utils.CallDuration, utils.FallbackSubject},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdMaxDebit struct {
	name       string
	rpcMethod  string
	rpcParams  *engine.CallDescriptorWithAPIOpts
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
		self.rpcParams = &engine.CallDescriptorWithAPIOpts{
			CallDescriptor: new(engine.CallDescriptor),
			APIOpts:        make(map[string]any),
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
