// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetDispatcherProfile{
		name:      "dispatcherprofile_get",
		rpcMethod: utils.APIerSv1GetDispatcherProfile,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetDispatcherProfile struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantID
	*CommandExecuter
}

func (self *CmdGetDispatcherProfile) Name() string {
	return self.name
}

func (self *CmdGetDispatcherProfile) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetDispatcherProfile) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.TenantID)
	}
	return self.rpcParams
}

func (self *CmdGetDispatcherProfile) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetDispatcherProfile) RpcResult() any {
	var s engine.DispatcherProfile
	return &s
}
