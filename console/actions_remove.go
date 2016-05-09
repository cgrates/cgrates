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

import "github.com/cgrates/cgrates/apier/v1"

func init() {
	c := &CmdRemoveActions{
		name:      "actions_remove",
		rpcMethod: "ApierV1.RemoveActions",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdRemoveActions struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrRemoveActions
	*CommandExecuter
}

func (self *CmdRemoveActions) Name() string {
	return self.name
}

func (self *CmdRemoveActions) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRemoveActions) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrRemoveActions{}
	}
	return self.rpcParams
}

func (self *CmdRemoveActions) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRemoveActions) RpcResult() interface{} {
	var s string
	return &s
}
