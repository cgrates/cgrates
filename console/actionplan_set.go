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

import "github.com/cgrates/cgrates/apier/v1"

func init() {
	c := &CmdSetActionPlan{
		name:      "actionplan_set",
		rpcMethod: "ApierV1.SetActionPlan",
		rpcParams: &v1.AttrSetActionPlan{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdSetActionPlan struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrSetActionPlan
	*CommandExecuter
}

func (self *CmdSetActionPlan) Name() string {
	return self.name
}

func (self *CmdSetActionPlan) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetActionPlan) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrSetActionPlan{}
	}
	return self.rpcParams
}

func (self *CmdSetActionPlan) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetActionPlan) RpcResult() interface{} {
	var s string
	return &s
}
