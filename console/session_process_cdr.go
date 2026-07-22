// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"time"

	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSessionsProcessCDR{
		name:      "session_process_cdr",
		rpcMethod: utils.SessionSv1ProcessCDR,
		rpcParams: &utils.CGREvent{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdSessionsProcessCDR struct {
	name      string
	rpcMethod string
	rpcParams *utils.CGREvent
	*CommandExecuter
}

func (self *CmdSessionsProcessCDR) Name() string {
	return self.name
}

func (self *CmdSessionsProcessCDR) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSessionsProcessCDR) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.CGREvent)
	}
	return self.rpcParams
}

func (self *CmdSessionsProcessCDR) PostprocessRpcParams() error {
	if self.rpcParams.Time == nil {
		self.rpcParams.Time = utils.TimePointer(time.Now())
	}
	return nil
}

func (self *CmdSessionsProcessCDR) RpcResult() any {
	var atr string
	return &atr
}
