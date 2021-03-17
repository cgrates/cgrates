/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

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

func (self *CmdMaxDebit) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.CallDescriptorWithAPIOpts{
			CallDescriptor: new(engine.CallDescriptor),
			APIOpts:        make(map[string]interface{}),
		}
	}
	return self.rpcParams
}

func (self *CmdMaxDebit) PostprocessRpcParams() error {
	return nil
}

func (self *CmdMaxDebit) RpcResult() interface{} {
	return &engine.CallCost{}
}

func (self *CmdMaxDebit) ClientArgs() []string {
	return self.clientArgs
}
