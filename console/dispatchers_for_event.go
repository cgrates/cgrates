// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdDispatcherProfile{
		name:      "dispatches_for_event",
		rpcMethod: utils.DispatcherSv1GetProfilesForEvent,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdDispatcherProfile struct {
	name      string
	rpcMethod string
	rpcParams *utils.CGREvent
	*CommandExecuter
}

func (self *CmdDispatcherProfile) Name() string {
	return self.name
}

func (self *CmdDispatcherProfile) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdDispatcherProfile) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.CGREvent)
	}
	return self.rpcParams
}

func (self *CmdDispatcherProfile) PostprocessRpcParams() error {
	return nil
}

func (self *CmdDispatcherProfile) RpcResult() any {
	var s engine.DispatcherProfiles
	return &s
}
