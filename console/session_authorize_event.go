// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"time"

	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSessionsAuthorize{
		name:      "session_authorize_event",
		rpcMethod: utils.SessionSv1AuthorizeEventWithDigest,
		rpcParams: &sessions.V1AuthorizeArgs{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdSessionsAuthorize struct {
	name      string
	rpcMethod string
	rpcParams *sessions.V1AuthorizeArgs
	*CommandExecuter
}

func (self *CmdSessionsAuthorize) Name() string {
	return self.name
}

func (self *CmdSessionsAuthorize) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSessionsAuthorize) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &sessions.V1AuthorizeArgs{ArgDispatcher: new(utils.ArgDispatcher)}
	}
	return self.rpcParams
}

func (self *CmdSessionsAuthorize) PostprocessRpcParams() error {
	if self.rpcParams != nil && self.rpcParams.CGREvent != nil &&
		self.rpcParams.CGREvent.Time == nil {
		self.rpcParams.CGREvent.Time = utils.TimePointer(time.Now())
	}
	return nil
}

func (self *CmdSessionsAuthorize) RpcResult() any {
	var atr *sessions.V1AuthorizeReplyWithDigest
	return &atr
}
