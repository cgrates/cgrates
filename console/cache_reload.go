// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import "github.com/cgrates/cgrates/utils"

func init() {
	c := &CmdReloadCache{
		name:      "cache_reload",
		rpcMethod: utils.CacheSv1ReloadCache,
		rpcParams: &utils.AttrReloadCacheWithAPIOpts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdReloadCache struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrReloadCacheWithAPIOpts
	rpcResult string
	*CommandExecuter
}

func (self *CmdReloadCache) Name() string {
	return self.name
}

func (self *CmdReloadCache) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdReloadCache) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.AttrReloadCacheWithAPIOpts{}
	}
	return self.rpcParams
}

func (self *CmdReloadCache) PostprocessRpcParams() error {
	return nil
}

func (self *CmdReloadCache) RpcResult() any {
	var s string
	return &s
}
