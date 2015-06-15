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

import (
	"github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetDataCost{
		name:       "datacost",
		rpcMethod:  "ApierV1.GetDataCost",
		clientArgs: []string{"Direction", "Category", "Tenant", "Account", "Subject", "StartTime", "Usage"},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetDataCost struct {
	name       string
	rpcMethod  string
	rpcParams  *v1.AttrGetDataCost
	clientArgs []string
	*CommandExecuter
}

func (self *CmdGetDataCost) Name() string {
	return self.name
}

func (self *CmdGetDataCost) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetDataCost) RpcParams(ptr bool) interface{} {
	if self.rpcParams == nil {
		self.rpcParams = &v1.AttrGetDataCost{Direction: utils.OUT}
	}
	if ptr {
		return self.rpcParams
	}
	return *self.rpcParams
}

func (self *CmdGetDataCost) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetDataCost) RpcResult() interface{} {
	return &engine.DataCost{}
}

func (self *CmdGetDataCost) ClientArgs() []string {
	return self.clientArgs
}
