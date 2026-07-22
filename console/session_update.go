// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"time"

	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSessionsUpdate{
		name:      "session_update",
		rpcMethod: utils.SessionSv1UpdateSession,
		rpcParams: &sessions.V1UpdateSessionArgs{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdSessionsUpdate struct {
	name      string
	rpcMethod string
	rpcParams *sessions.V1UpdateSessionArgs
	*CommandExecuter
}

func (self *CmdSessionsUpdate) Name() string {
	return self.name
}

func (self *CmdSessionsUpdate) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSessionsUpdate) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &sessions.V1UpdateSessionArgs{
			CGREvent: new(utils.CGREvent),
		}
	}
	return self.rpcParams
}

func (self *CmdSessionsUpdate) PostprocessRpcParams() error {
	if self.rpcParams != nil && self.rpcParams.CGREvent != nil &&
		self.rpcParams.CGREvent.Time == nil {
		self.rpcParams.CGREvent.Time = utils.TimePointer(time.Now())
	}
	return nil
}

func (self *CmdSessionsUpdate) RpcResult() any {
	var atr sessions.V1UpdateSessionReply
	return &atr
}

func (self *CmdSessionsUpdate) GetFormatedResult(result any) string {
	return GetFormatedResult(result, utils.StringSet{
		utils.Usage:       {},
		utils.CapMaxUsage: {},
	})
}
