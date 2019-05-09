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
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdCacheRemoveItem{
		name:      "cache_remove_item",
		rpcMethod: utils.CacheSv1RemoveItem,
		rpcParams: &utils.ArgsGetCacheItem{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdCacheRemoveItem struct {
	name      string
	rpcMethod string
	rpcParams *utils.ArgsGetCacheItem
	*CommandExecuter
}

func (self *CmdCacheRemoveItem) Name() string {
	return self.name
}

func (self *CmdCacheRemoveItem) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdCacheRemoveItem) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.ArgsGetCacheItem{}
	}
	return self.rpcParams
}

func (self *CmdCacheRemoveItem) PostprocessRpcParams() error {
	return nil
}

func (self *CmdCacheRemoveItem) RpcResult() interface{} {
	var reply string
	return &reply
}
