// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdCacheHasItem{
		name:      "cache_has_item",
		rpcMethod: utils.CacheSv1HasItem,
		rpcParams: &utils.ArgsGetCacheItemWithArgDispatcher{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdCacheHasItem struct {
	name      string
	rpcMethod string
	rpcParams *utils.ArgsGetCacheItemWithArgDispatcher
	*CommandExecuter
}

func (self *CmdCacheHasItem) Name() string {
	return self.name
}

func (self *CmdCacheHasItem) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdCacheHasItem) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.ArgsGetCacheItemWithArgDispatcher{}
	}
	return self.rpcParams
}

func (self *CmdCacheHasItem) PostprocessRpcParams() error {
	return nil
}

func (self *CmdCacheHasItem) RpcResult() any {
	var reply bool
	return &reply
}
