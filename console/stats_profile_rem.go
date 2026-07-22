// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import "github.com/cgrates/cgrates/utils"

func init() {
	c := &CmdRemoveStatQueue{
		name:      "stats_profile_remove",
		rpcMethod: utils.APIerSv1RemoveStatQueueProfile,
		rpcParams: &utils.TenantIDWithAPIOpts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdRemoveStatQueue struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantIDWithAPIOpts
	*CommandExecuter
}

func (self *CmdRemoveStatQueue) Name() string {
	return self.name
}

func (self *CmdRemoveStatQueue) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRemoveStatQueue) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantIDWithAPIOpts{APIOpts: make(map[string]any)}
	}
	return self.rpcParams
}

func (self *CmdRemoveStatQueue) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRemoveStatQueue) RpcResult() any {
	var s string
	return &s
}
