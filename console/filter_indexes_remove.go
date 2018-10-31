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

import (
	"github.com/cgrates/cgrates/apier/v1"
)

func init() {
	c := &CmdRemoveFilterIndexes{
		name:      "filter_indexes_remove",
		rpcMethod: "ApierV1.RemoveFilterIndexes",
		rpcParams: &v1.AttrRemFilterIndexes{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdRemoveFilterIndexes struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrRemFilterIndexes
	*CommandExecuter
}

func (self *CmdRemoveFilterIndexes) Name() string {
	return self.name
}

func (self *CmdRemoveFilterIndexes) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRemoveFilterIndexes) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrRemFilterIndexes{}
	}
	return self.rpcParams
}

func (self *CmdRemoveFilterIndexes) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRemoveFilterIndexes) RpcResult() interface{} {
	var atr string
	return &atr
}
