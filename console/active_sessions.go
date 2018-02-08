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
	rpcParams interface{}
	*CommandExecuter
}

func (self *CmdActiveSessions) Name() string {
	return self.name
}

func (self *CmdActiveSessions) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdActiveSessions) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &map[string]string{}
	}
	return self.rpcParams
}

func (self *CmdActiveSessions) PostprocessRpcParams() error {
	param := self.rpcParams.(*map[string]string)
	self.rpcParams = param
	return nil
}

func (self *CmdActiveSessions) RpcResult() interface{} {
	var sessions *[]*sessions.ActiveSession
	return &sessions
}
