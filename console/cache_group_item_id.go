// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdCacheGetGroupItemIDs{
		name:      "cache_group_item_ids",
		rpcMethod: utils.CacheSv1GetGroupItemIDs,
		rpcParams: &utils.ArgsGetGroupWithArgDispatcher{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdCacheGetGroupItemIDs struct {
	name      string
	rpcMethod string
	rpcParams *utils.ArgsGetGroupWithArgDispatcher
	*CommandExecuter
}

func (self *CmdCacheGetGroupItemIDs) Name() string {
	return self.name
}

func (self *CmdCacheGetGroupItemIDs) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdCacheGetGroupItemIDs) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.ArgsGetGroupWithArgDispatcher{}
	}
	return self.rpcParams
}

func (self *CmdCacheGetGroupItemIDs) PostprocessRpcParams() error {
	return nil
}

func (self *CmdCacheGetGroupItemIDs) RpcResult() any {
	var reply []string
	return &reply
}
