// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"time"

	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSessionsProcessEvent{
		name:      "session_process_message",
		rpcMethod: utils.SessionSv1ProcessMessage,
		rpcParams: &sessions.V1ProcessMessageArgs{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdSessionsProcessEvent struct {
	name      string
	rpcMethod string
	rpcParams *sessions.V1ProcessMessageArgs
	*CommandExecuter
}

func (self *CmdSessionsProcessEvent) Name() string {
	return self.name
}

func (self *CmdSessionsProcessEvent) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSessionsProcessEvent) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &sessions.V1ProcessMessageArgs{
			CGREvent: new(utils.CGREvent),
		}
	}
	return self.rpcParams
}

func (self *CmdSessionsProcessEvent) PostprocessRpcParams() error {
	if self.rpcParams != nil && self.rpcParams.CGREvent != nil &&
		self.rpcParams.CGREvent.Time == nil {
		self.rpcParams.CGREvent.Time = utils.TimePointer(time.Now())
	}
	return nil
}

func (self *CmdSessionsProcessEvent) RpcResult() any {
	var atr sessions.V1ProcessMessageReply
	return &atr
}

func (self *CmdSessionsProcessEvent) GetFormatedResult(result any) string {
	return GetFormatedResult(result, utils.StringSet{
		utils.Usage:       {},
		utils.CapMaxUsage: {},
	})
}
