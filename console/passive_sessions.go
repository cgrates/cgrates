// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdPassiveSessions{
		name:      "passive_sessions",
		rpcMethod: utils.SessionSv1GetPassiveSessions,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdPassiveSessions struct {
	name      string
	rpcMethod string
	rpcParams any
	*CommandExecuter
}

func (self *CmdPassiveSessions) Name() string {
	return self.name
}

func (self *CmdPassiveSessions) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdPassiveSessions) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.SessionFilter{ArgDispatcher: new(utils.ArgDispatcher)}
	}
	return self.rpcParams
}

func (self *CmdPassiveSessions) PostprocessRpcParams() error {
	param := self.rpcParams.(*utils.SessionFilter)
	self.rpcParams = param
	return nil
}

func (self *CmdPassiveSessions) RpcResult() any {
	var sessions *[]*sessions.ExternalSession
	return &sessions
}

func (self *CmdPassiveSessions) GetFormatedResult(result any) string {
	return GetFormatedSliceResult(result, map[string]struct{}{
		"Usage":         {},
		"DurationIndex": {},
		"MaxRateUnit":   {},
		"DebitInterval": {},
	})
}
