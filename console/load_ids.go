// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import "github.com/cgrates/cgrates/utils"

func init() {
	c := &CmdCacheVersions{
		name:      "get_load_ids",
		rpcMethod: utils.APIerSv1GetLoadIDs,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdCacheVersions struct {
	name      string
	rpcMethod string
	rpcParams *StringWrapper
	*CommandExecuter
}

func (self *CmdCacheVersions) Name() string {
	return self.name
}

func (self *CmdCacheVersions) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdCacheVersions) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &StringWrapper{}
	}
	return self.rpcParams
}

func (self *CmdCacheVersions) PostprocessRpcParams() error {
	return nil
}

func (self *CmdCacheVersions) RpcResult() any {
	a := make(map[string]int64, 0)
	return &a
}
