// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/scheduler"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetScheduledActions{
		name:      "scheduler_queue",
		rpcMethod: utils.APIerSv1GetScheduledActions,
		rpcParams: &scheduler.ArgsGetScheduledActions{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetScheduledActions struct {
	name      string
	rpcMethod string
	rpcParams *scheduler.ArgsGetScheduledActions
	*CommandExecuter
}

func (self *CmdGetScheduledActions) Name() string {
	return self.name
}

func (self *CmdGetScheduledActions) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetScheduledActions) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &scheduler.ArgsGetScheduledActions{}
	}
	return self.rpcParams
}

func (self *CmdGetScheduledActions) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetScheduledActions) RpcResult() any {
	s := make([]*scheduler.ScheduledAction, 0)
	return &s
}
