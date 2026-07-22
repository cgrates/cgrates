// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetDispatcherHost{
		name:      "dispatchers_host",
		rpcMethod: utils.APIerSv1GetDispatcherHost,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetDispatcherHost struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantID
	*CommandExecuter
}

func (self *CmdGetDispatcherHost) Name() string {
	return self.name
}

func (self *CmdGetDispatcherHost) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetDispatcherHost) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.TenantID)
	}
	return self.rpcParams
}

func (self *CmdGetDispatcherHost) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetDispatcherHost) RpcResult() any {
	var s engine.DispatcherHost
	return &s
}
