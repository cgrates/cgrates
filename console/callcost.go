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

import (
	"github.com/cgrates/cgrates/apier"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetCallCost{
		name:      "callcost",
		rpcMethod: "ApierV1.GetCallCostLog",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetCallCost struct {
	name      string
	rpcMethod string
	rpcParams *apier.AttrGetCallCost
	rpcResult string
	*CommandExecuter
}

func (self *CmdGetCallCost) Name() string {
	return self.name
}

func (self *CmdGetCallCost) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetCallCost) RpcParams() interface{} {
	if self.rpcParams == nil {
		self.rpcParams = &apier.AttrGetCallCost{RunId: utils.DEFAULT_RUNID}
	}
	return self.rpcParams
}

func (self *CmdGetCallCost) RpcResult() interface{} {
	var s string
	return &s
}
