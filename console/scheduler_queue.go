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

import "github.com/cgrates/cgrates/apier/v1"

func init() {
	c := &CmdGetScheduledActions{
		name:      "scheduler_queue",
		rpcMethod: "ApierV1.GetScheduledActions",
		rpcParams: &v1.AttrsGetScheduledActions{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetScheduledActions struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrsGetScheduledActions
	*CommandExecuter
}

func (self *CmdGetScheduledActions) Name() string {
	return self.name
}

func (self *CmdGetScheduledActions) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetScheduledActions) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrsGetScheduledActions{}
	}
	return self.rpcParams
}

func (self *CmdGetScheduledActions) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetScheduledActions) RpcResult() interface{} {
	s := v1.ScheduledActions{}
	return &s
}
