/*
Real-time Charging System for Telecom & ISP environments
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
	"github.com/cgrates/cgrates/engine"
)

func init() {
	c := &CmdGetLcr{
		name:      "lcr",
		rpcMethod: "ApierV1.GetLcr",
		rpcParams: &engine.LcrRequest{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetLcr struct {
	name      string
	rpcMethod string
	rpcParams *engine.LcrRequest
	*CommandExecuter
}

func (self *CmdGetLcr) Name() string {
	return self.name
}

func (self *CmdGetLcr) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetLcr) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.LcrRequest{}
	}
	return self.rpcParams
}

func (self *CmdGetLcr) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetLcr) RpcResult() interface{} {
	return &engine.LcrReply{}
}
