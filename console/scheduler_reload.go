// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import "github.com/cgrates/cgrates/utils"

func init() {
	c := &CmdReloadScheduler{
		name:      "scheduler_reload",
		rpcMethod: utils.SchedulerSv1Reload,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdReloadScheduler struct {
	name      string
	rpcMethod string
	rpcParams *utils.CGREvent
	*CommandExecuter
}

func (self *CmdReloadScheduler) Name() string {
	return self.name
}

func (self *CmdReloadScheduler) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdReloadScheduler) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.CGREvent{}
	}
	return self.rpcParams
}

func (self *CmdReloadScheduler) PostprocessRpcParams() error {
	return nil
}

func (self *CmdReloadScheduler) RpcResult() any {
	var s string
	return &s
}
