// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetStatQueueStringMetrics{
		name:      "stats_metrics",
		rpcMethod: utils.StatSv1GetQueueStringMetrics,
		rpcParams: &utils.TenantIDWithArgDispatcher{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetStatQueueStringMetrics struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantIDWithArgDispatcher
	*CommandExecuter
}

func (self *CmdGetStatQueueStringMetrics) Name() string {
	return self.name
}

func (self *CmdGetStatQueueStringMetrics) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetStatQueueStringMetrics) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantIDWithArgDispatcher{
			TenantID:      new(utils.TenantID),
			ArgDispatcher: new(utils.ArgDispatcher),
		}
	}
	return self.rpcParams
}

func (self *CmdGetStatQueueStringMetrics) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetStatQueueStringMetrics) RpcResult() any {
	var atr *map[string]string
	return &atr
}
