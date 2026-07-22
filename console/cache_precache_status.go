// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetPrecacheStatus{
		name:      "cache_precache_status",
		rpcMethod: utils.CacheSv1PrecacheStatus,
		rpcParams: &utils.AttrCacheIDsWithAPIOpts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetPrecacheStatus struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrCacheIDsWithAPIOpts
	*CommandExecuter
}

func (self *CmdGetPrecacheStatus) Name() string {
	return self.name
}

func (self *CmdGetPrecacheStatus) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetPrecacheStatus) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.AttrCacheIDsWithAPIOpts)
	}
	return self.rpcParams
}

func (self *CmdGetPrecacheStatus) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetPrecacheStatus) RpcResult() any {
	reply := make(map[string]string)
	return &reply
}
