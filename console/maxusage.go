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

import "github.com/cgrates/cgrates/engine"

func init() {
	c := &CmdGetMaxUsage{
		name:      "maxusage",
		rpcMethod: "ApierV1.GetMaxUsage",
		clientArgs: []string{"ToR", "RequestType", "Tenant",
			"Category", "Account", "Subject", "Destination",
			"SetupTime", "AnswerTime", "Usage", "ExtraFields"},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetMaxUsage struct {
	name       string
	rpcMethod  string
	rpcParams  *engine.UsageRecord
	clientArgs []string
	*CommandExecuter
}

func (self *CmdGetMaxUsage) Name() string {
	return self.name
}

func (self *CmdGetMaxUsage) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetMaxUsage) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(engine.UsageRecord)
	}
	return self.rpcParams
}

func (self *CmdGetMaxUsage) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetMaxUsage) RpcResult() interface{} {
	var f int64
	return &f
}

func (self *CmdGetMaxUsage) ClientArgs() []string {
	return self.clientArgs
}
