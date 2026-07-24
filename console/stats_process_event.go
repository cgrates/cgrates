// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdStatQueueProcessEvent{
		name:      "stats_process_event",
		rpcMethod: utils.StatSv1ProcessEvent,
		rpcParams: &engine.StatsArgsProcessEvent{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdStatQueueProcessEvent struct {
	name      string
	rpcMethod string
	rpcParams *engine.StatsArgsProcessEvent
	*CommandExecuter
}

func (self *CmdStatQueueProcessEvent) Name() string {
	return self.name
}

func (self *CmdStatQueueProcessEvent) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdStatQueueProcessEvent) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.StatsArgsProcessEvent{ArgDispatcher: new(utils.ArgDispatcher)}
	}
	return self.rpcParams
}

func (self *CmdStatQueueProcessEvent) PostprocessRpcParams() error {
	if self.rpcParams != nil && self.rpcParams.CGREvent != nil &&
		self.rpcParams.CGREvent.Time == nil {
		self.rpcParams.CGREvent.Time = utils.TimePointer(time.Now())
	}
	return nil
}

func (self *CmdStatQueueProcessEvent) RpcResult() any {
	var atr []string
	return &atr
}
