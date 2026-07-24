// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdRemoveDispatcherHost{
		name:      "dispatcher_host_remove",
		rpcMethod: utils.APIerSv1RemoveDispatcherHost,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdRemoveDispatcherHost struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantIDWithCache
	*CommandExecuter
}

func (self *CmdRemoveDispatcherHost) Name() string {
	return self.name
}

func (self *CmdRemoveDispatcherHost) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRemoveDispatcherHost) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.TenantIDWithCache)
	}
	return self.rpcParams
}

func (self *CmdRemoveDispatcherHost) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRemoveDispatcherHost) RpcResult() any {
	var s string
	return &s
}
