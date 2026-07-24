// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdCacheGetItemIDs{
		name:      "cache_item_ids",
		rpcMethod: utils.CacheSv1GetItemIDs,
		rpcParams: &utils.ArgsGetCacheItemIDsWithArgDispatcher{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdCacheGetItemIDs struct {
	name      string
	rpcMethod string
	rpcParams *utils.ArgsGetCacheItemIDsWithArgDispatcher
	*CommandExecuter
}

func (self *CmdCacheGetItemIDs) Name() string {
	return self.name
}

func (self *CmdCacheGetItemIDs) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdCacheGetItemIDs) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.ArgsGetCacheItemIDsWithArgDispatcher{}
	}
	return self.rpcParams
}

func (self *CmdCacheGetItemIDs) PostprocessRpcParams() error {
	return nil
}

func (self *CmdCacheGetItemIDs) RpcResult() any {
	var reply []string
	return &reply
}
