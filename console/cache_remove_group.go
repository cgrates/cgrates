// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdCacheRemoveGroup{
		name:      "cache_remove_group",
		rpcMethod: utils.CacheSv1RemoveGroup,
		rpcParams: &utils.ArgsGetGroupWithAPIOpts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdCacheRemoveGroup struct {
	name      string
	rpcMethod string
	rpcParams *utils.ArgsGetGroupWithAPIOpts
	*CommandExecuter
}

func (self *CmdCacheRemoveGroup) Name() string {
	return self.name
}

func (self *CmdCacheRemoveGroup) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdCacheRemoveGroup) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.ArgsGetGroupWithAPIOpts{}
	}
	return self.rpcParams
}

func (self *CmdCacheRemoveGroup) PostprocessRpcParams() error {
	return nil
}

func (self *CmdCacheRemoveGroup) RpcResult() any {
	var reply string
	return &reply
}
