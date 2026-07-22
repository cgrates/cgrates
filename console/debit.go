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
		clientArgs: []string{utils.Category, utils.ToR, utils.Tenant, utils.Subject, utils.AccountField, utils.Destination, utils.TimeStart, utils.TimeEnd, utils.CallDuration, utils.FallbackSubject, utils.DryRun},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdDebit struct {
	name       string
	rpcMethod  string
	rpcParams  *engine.CallDescriptorWithAPIOpts
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
		self.rpcParams = &engine.CallDescriptorWithAPIOpts{
			CallDescriptor: new(engine.CallDescriptor),
			APIOpts:        make(map[string]any),
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
