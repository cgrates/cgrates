// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSetStatQueue{
		name:      "stats_profile_set",
		rpcMethod: utils.APIerSv1SetStatQueueProfile,
		rpcParams: &engine.StatQueueProfileWithAPIOpts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdSetStatQueue struct {
	name      string
	rpcMethod string
	rpcParams *engine.StatQueueProfileWithAPIOpts
	*CommandExecuter
}

func (self *CmdSetStatQueue) Name() string {
	return self.name
}

func (self *CmdSetStatQueue) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetStatQueue) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.StatQueueProfileWithAPIOpts{
			StatQueueProfile: new(engine.StatQueueProfile),
			APIOpts:          make(map[string]any),
		}
	}
	return self.rpcParams
}

func (self *CmdSetStatQueue) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetStatQueue) RpcResult() any {
	var s string
	return &s
}
