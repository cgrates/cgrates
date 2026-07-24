// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

func init() {
	c := &CmdGetCacheStats{
		name:      "cache_stats",
		rpcMethod: utils.CacheSv1GetCacheStats,
		rpcParams: &utils.AttrCacheIDsWithArgDispatcher{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetCacheStats struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrCacheIDsWithArgDispatcher
	*CommandExecuter
}

func (self *CmdGetCacheStats) Name() string {
	return self.name
}

func (self *CmdGetCacheStats) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetCacheStats) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.AttrCacheIDsWithArgDispatcher)
	}
	return self.rpcParams
}

func (self *CmdGetCacheStats) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetCacheStats) RpcResult() any {
	reply := make(map[string]*ltcache.CacheStats)
	return &reply
}
