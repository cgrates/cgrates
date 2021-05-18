/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package console

import (
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

func (self *CmdSessionsProcessEvent) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &sessions.V1ProcessMessageArgs{
			CGREvent: new(utils.CGREvent),
		}
	}
	return self.rpcParams
}

func (self *CmdSessionsProcessEvent) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSessionsProcessEvent) RpcResult() interface{} {
	var atr sessions.V1ProcessMessageReply
	return &atr
}

func (self *CmdSessionsProcessEvent) GetFormatedResult(result interface{}) string {
	return GetFormatedResult(result, utils.StringSet{
		utils.Usage:       {},
		utils.CapMaxUsage: {},
	})
}
