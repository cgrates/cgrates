/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

import "github.com/cgrates/cgrates/engine"

func init() {
	c := &CmdDebitBalance{
		name:       "balance_debit",
		rpcMethod:  "Responder.Debit",
		rpcParams:  &engine.CallDescriptor{Direction: "*out"},
		clientArgs: []string{"Direction", "TOR", "Tenant", "Subject", "Account", "Destination", "TimeStart", "TimeEnd", "CallDuration", "FallbackSubject"},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdDebitBalance struct {
	name       string
	rpcMethod  string
	rpcParams  *engine.CallDescriptor
	rpcResult  string
	clientArgs []string
	*CommandExecuter
}

func (self *CmdDebitBalance) Name() string {
	return self.name
}

func (self *CmdDebitBalance) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdDebitBalance) RpcParams() interface{} {
    if self.rpcParams == nil {
		self.rpcParams = &engine.CallDescriptor{Direction: "*out"}
	}
	return self.rpcParams
}

func (self *CmdDebitBalance) RpcResult() interface{} {
	return &self.rpcResult
}

func (self *CmdDebitBalance) ClientArgs() []string {
	return self.clientArgs
}
