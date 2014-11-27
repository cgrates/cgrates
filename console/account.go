/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2014 ITsysCOM

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
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetAccount{
		name:      "account",
		rpcMethod: "ApierV1.GetAccount",
		rpcParams: &utils.AttrGetAccount{Direction: "*out"},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetAccount struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrGetAccount
	*CommandExecuter
}

func (self *CmdGetAccount) Name() string {
	return self.name
}

func (self *CmdGetAccount) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetAccount) RpcParams() interface{} {
	if self.rpcParams == nil {
		self.rpcParams = &utils.AttrGetAccount{Direction: "*out"}
	}
	return self.rpcParams
}

func (self *CmdGetAccount) RpcResult() interface{} {
	return &engine.Account{}
}
