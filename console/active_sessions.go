// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdActiveSessions{
		name:      "active_sessions",
		rpcMethod: utils.SessionSv1GetActiveSessions,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdActiveSessions struct {
	name      string
	rpcMethod string
	rpcParams any
	*CommandExecuter
}

func (self *CmdActiveSessions) Name() string {
	return self.name
}

func (self *CmdActiveSessions) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdActiveSessions) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.SessionFilter{ArgDispatcher: new(utils.ArgDispatcher)}

	}
	return self.rpcParams
}

func (self *CmdActiveSessions) PostprocessRpcParams() error {
	param := self.rpcParams.(*utils.SessionFilter)
	self.rpcParams = param
	return nil
}

func (self *CmdActiveSessions) RpcResult() any {
	var sessions *[]*sessions.ExternalSession
	return &sessions
}

func (self *CmdActiveSessions) GetFormatedResult(result any) string {
	return GetFormatedSliceResult(result, map[string]struct{}{
		"Usage":         {},
		"DurationIndex": {},
		"MaxRateUnit":   {},
		"DebitInterval": {},
	})
}
