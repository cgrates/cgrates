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
	c := &CmdSessionsAuthorize{
		name:      "session_authorize_event",
		rpcMethod: utils.SessionSv1AuthorizeEventWithDigest,
		rpcParams: &dispatchers.AuthorizeArgsWithApiKey{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdSessionsAuthorize struct {
	name      string
	rpcMethod string
	rpcParams *dispatchers.AuthorizeArgsWithApiKey
	*CommandExecuter
}

func (self *CmdSessionsAuthorize) Name() string {
	return self.name
}

func (self *CmdSessionsAuthorize) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSessionsAuthorize) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &dispatchers.AuthorizeArgsWithApiKey{}
	}
	return self.rpcParams
}

func (self *CmdSessionsAuthorize) PostprocessRpcParams() error {
	if self.rpcParams.CGREvent.Time == nil {
		self.rpcParams.CGREvent.Time = utils.TimePointer(time.Now())
	}
	return nil
}

func (self *CmdSessionsAuthorize) RpcResult() interface{} {
	var atr *sessions.V1AuthorizeReplyWithDigest
	return &atr
}
