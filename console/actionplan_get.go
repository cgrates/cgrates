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

import (
	"github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"
)

func init() {
	c := &CmdGetActionPlan{
		name:      "actionplan_get",
		rpcMethod: "ApierV1.GetActionPlan",
		rpcParams: &v1.AttrGetActionPlan{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetActionPlan struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrGetActionPlan
	*CommandExecuter
}

func (self *CmdGetActionPlan) Name() string {
	return self.name
}

func (self *CmdGetActionPlan) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetActionPlan) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrGetActionPlan{}
	}
	return self.rpcParams
}

func (self *CmdGetActionPlan) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetActionPlan) RpcResult() interface{} {
	s := make([]*engine.ActionPlan, 0)
	return &s
}
