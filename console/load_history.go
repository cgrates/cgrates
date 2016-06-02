/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

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

import "github.com/cgrates/cgrates/utils"

func init() {
	c := &CmdGetLoadHistory{
		name:      "load_history",
		rpcMethod: "ApierV1.GetLoadHistory",
		rpcParams: new(utils.Paginator),
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Returns the list of load items from the history, in reverse order
type CmdGetLoadHistory struct {
	name      string
	rpcMethod string
	rpcParams *utils.Paginator
	*CommandExecuter
}

func (self *CmdGetLoadHistory) Name() string {
	return self.name
}

func (self *CmdGetLoadHistory) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetLoadHistory) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.Paginator)
	}
	return self.rpcParams
}

func (self *CmdGetLoadHistory) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetLoadHistory) RpcResult() interface{} {
	a := make([]*utils.LoadInstance, 0)
	return &a
}
