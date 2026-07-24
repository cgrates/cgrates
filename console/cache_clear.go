// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdClear{
		name:      "cache_clear",
		rpcMethod: utils.CacheSv1Clear,
		rpcParams: &utils.AttrCacheIDsWithArgDispatcher{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdClear struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrCacheIDsWithArgDispatcher
	*CommandExecuter
}

func (self *CmdClear) Name() string {
	return self.name
}

func (self *CmdClear) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdClear) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.AttrCacheIDsWithArgDispatcher)
	}
	return self.rpcParams
}

func (self *CmdClear) PostprocessRpcParams() error {
	return nil
}

func (self *CmdClear) RpcResult() any {
	var reply string
	return &reply
}
