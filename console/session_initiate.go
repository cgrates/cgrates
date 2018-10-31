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
	"time"

	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSessionsInitiate{
		name:      "session_initiate",
		rpcMethod: utils.SessionSv1InitiateSessionWithDigest,
		rpcParams: &dispatchers.InitArgsWithApiKey{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdSessionsInitiate struct {
	name      string
	rpcMethod string
	rpcParams *dispatchers.InitArgsWithApiKey
	*CommandExecuter
}

func (self *CmdSessionsInitiate) Name() string {
	return self.name
}

func (self *CmdSessionsInitiate) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSessionsInitiate) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &dispatchers.InitArgsWithApiKey{}
	}
	return self.rpcParams
}

func (self *CmdSessionsInitiate) PostprocessRpcParams() error {
	if self.rpcParams.CGREvent.Time == nil {
		self.rpcParams.CGREvent.Time = utils.TimePointer(time.Now())
	}
	return nil
}

func (self *CmdSessionsInitiate) RpcResult() interface{} {
	var atr *sessions.V1InitSessionReply
	return &atr
}
