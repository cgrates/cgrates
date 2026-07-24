// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdCacheRemoveItem{
		name:      "cache_remove_item",
		rpcMethod: utils.CacheSv1RemoveItem,
		rpcParams: &utils.ArgsGetCacheItemWithArgDispatcher{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdCacheRemoveItem struct {
	name      string
	rpcMethod string
	rpcParams *utils.ArgsGetCacheItemWithArgDispatcher
	*CommandExecuter
}

func (self *CmdCacheRemoveItem) Name() string {
	return self.name
}

func (self *CmdCacheRemoveItem) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdCacheRemoveItem) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.ArgsGetCacheItemWithArgDispatcher{}
	}
	return self.rpcParams
}

func (self *CmdCacheRemoveItem) PostprocessRpcParams() error {
	return nil
}

func (self *CmdCacheRemoveItem) RpcResult() any {
	var reply string
	return &reply
}
