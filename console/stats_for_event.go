// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"time"

	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdStatsQueueForEvent{
		name:      "stats_for_event",
		rpcMethod: utils.StatSv1GetStatQueuesForEvent,
		rpcParams: &utils.CGREvent{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdStatsQueueForEvent struct {
	name      string
	rpcMethod string
	rpcParams *utils.CGREvent
	*CommandExecuter
}

func (self *CmdStatsQueueForEvent) Name() string {
	return self.name
}

func (self *CmdStatsQueueForEvent) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdStatsQueueForEvent) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.CGREvent)
	}
	return self.rpcParams
}

func (self *CmdStatsQueueForEvent) PostprocessRpcParams() error {
	if self.rpcParams.Time == nil {
		self.rpcParams.Time = utils.TimePointer(time.Now())
	}
	return nil
}

func (self *CmdStatsQueueForEvent) RpcResult() any {
	var atr []string
	return &atr
}
