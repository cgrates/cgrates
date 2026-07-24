// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdDispatcherProfile{
		name:      "dispatcherprofile",
		rpcMethod: utils.DispatcherSv1GetProfileForEvent,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdDispatcherProfile struct {
	name      string
	rpcMethod string
	rpcParams *dispatchers.DispatcherEvent
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
		self.rpcParams = &dispatchers.DispatcherEvent{ArgDispatcher: new(utils.ArgDispatcher)}
	}
	return self.rpcParams
}

func (self *CmdDispatcherProfile) PostprocessRpcParams() error {
	return nil
}

func (self *CmdDispatcherProfile) RpcResult() any {
	var s engine.DispatcherProfile
	return &s
}
