// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdResetStatQueue{
		name:      "stats_reset_queue",
		rpcMethod: utils.StatSv1ResetStatQueue,
		rpcParams: &utils.TenantIDWithAPIOpts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdResetStatQueue struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantIDWithAPIOpts
	*CommandExecuter
}

func (self *CmdResetStatQueue) Name() string {
	return self.name
}

func (self *CmdResetStatQueue) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdResetStatQueue) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantIDWithAPIOpts{
			TenantID: new(utils.TenantID),
			APIOpts:  make(map[string]any),
		}
	}
	return self.rpcParams
}

func (self *CmdResetStatQueue) PostprocessRpcParams() error {
	return nil
}

func (self *CmdResetStatQueue) RpcResult() any {
	var s string
	return &s
}
