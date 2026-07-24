// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdCacheHasGroup{
		name:      "cache_has_group",
		rpcMethod: utils.CacheSv1HasGroup,
		rpcParams: &utils.ArgsGetGroupWithArgDispatcher{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdCacheHasGroup struct {
	name      string
	rpcMethod string
	rpcParams *utils.ArgsGetGroupWithArgDispatcher
	*CommandExecuter
}

func (self *CmdCacheHasGroup) Name() string {
	return self.name
}

func (self *CmdCacheHasGroup) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdCacheHasGroup) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.ArgsGetGroupWithArgDispatcher{}
	}
	return self.rpcParams
}

func (self *CmdCacheHasGroup) PostprocessRpcParams() error {
	return nil
}

func (self *CmdCacheHasGroup) RpcResult() any {
	var reply bool
	return &reply
}
