/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	c := &CmdGetMaxDuration{
		name:       "maxduration",
		rpcMethod:  "Responder.GetMaxSessionTime",
		clientArgs: []string{"Direction", "Category", "TOR", "Tenant", "Subject", "Account", "Destination", "TimeStart", "TimeEnd", "CallDuration", "FallbackSubject"},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetMaxDuration struct {
	name       string
	rpcMethod  string
	rpcParams  *engine.CallDescriptor
	clientArgs []string
	*CommandExecuter
}

func (self *CmdGetMaxDuration) Name() string {
	return self.name
}

func (self *CmdGetMaxDuration) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetMaxDuration) RpcParams(ptr bool) interface{} {
	if self.rpcParams == nil {
		self.rpcParams = &engine.CallDescriptor{Direction: "*out"}
	}
	if ptr {
		return self.rpcParams
	}
	return *self.rpcParams
}

func (self *CmdGetMaxDuration) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetMaxDuration) RpcResult() interface{} {
	var f float64
	return &f
}

func (self *CmdGetMaxDuration) ClientArgs() []string {
	return self.clientArgs
}
