// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"time"

	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSessionsTerminate{
		name:      "session_terminate",
		rpcMethod: utils.SessionSv1TerminateSession,
		rpcParams: &sessions.V1TerminateSessionArgs{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdSessionsTerminate struct {
	name      string
	rpcMethod string
	rpcParams *sessions.V1TerminateSessionArgs
	*CommandExecuter
}

func (self *CmdSessionsTerminate) Name() string {
	return self.name
}

func (self *CmdSessionsTerminate) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSessionsTerminate) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &sessions.V1TerminateSessionArgs{ArgDispatcher: new(utils.ArgDispatcher)}
	}
	return self.rpcParams
}

func (self *CmdSessionsTerminate) PostprocessRpcParams() error {
	if self.rpcParams != nil && self.rpcParams.CGREvent != nil &&
		self.rpcParams.CGREvent.Time == nil {
		self.rpcParams.CGREvent.Time = utils.TimePointer(time.Now())
	}
	return nil
}

func (self *CmdSessionsTerminate) RpcResult() any {
	var atr *string
	return &atr
}
