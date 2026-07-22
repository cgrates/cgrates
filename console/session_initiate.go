// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"time"

	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSessionsInitiate{
		name:      "session_initiate",
		rpcMethod: utils.SessionSv1InitiateSessionWithDigest,
		rpcParams: &sessions.V1InitSessionArgs{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdSessionsInitiate struct {
	name      string
	rpcMethod string
	rpcParams *sessions.V1InitSessionArgs
	*CommandExecuter
}

func (self *CmdSessionsInitiate) Name() string {
	return self.name
}

func (self *CmdSessionsInitiate) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSessionsInitiate) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &sessions.V1InitSessionArgs{
			CGREvent: new(utils.CGREvent),
		}
	}
	return self.rpcParams
}

func (self *CmdSessionsInitiate) PostprocessRpcParams() error {
	if self.rpcParams != nil && self.rpcParams.CGREvent != nil &&
		self.rpcParams.CGREvent.Time == nil {
		self.rpcParams.CGREvent.Time = utils.TimePointer(time.Now())
	}
	return nil
}

func (self *CmdSessionsInitiate) RpcResult() any {
	var atr sessions.V1InitReplyWithDigest
	return &atr
}

func (self *CmdSessionsInitiate) GetFormatedResult(result any) string {
	return GetFormatedResult(result, utils.StringSet{
		utils.Usage:       {},
		utils.CapMaxUsage: {},
	})
}
