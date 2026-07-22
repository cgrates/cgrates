// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"time"

	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdCacheGetItemExpiryTime{
		name:      "cache_item_expiry_time",
		rpcMethod: utils.CacheSv1GetItemExpiryTime,
		rpcParams: &utils.ArgsGetCacheItemWithAPIOpts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdCacheGetItemExpiryTime struct {
	name      string
	rpcMethod string
	rpcParams *utils.ArgsGetCacheItemWithAPIOpts
	*CommandExecuter
}

func (self *CmdCacheGetItemExpiryTime) Name() string {
	return self.name
}

func (self *CmdCacheGetItemExpiryTime) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdCacheGetItemExpiryTime) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.ArgsGetCacheItemWithAPIOpts{}
	}
	return self.rpcParams
}

func (self *CmdCacheGetItemExpiryTime) PostprocessRpcParams() error {
	return nil
}

func (self *CmdCacheGetItemExpiryTime) RpcResult() any {
	var reply time.Time
	return &reply
}
