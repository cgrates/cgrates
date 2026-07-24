// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdRemoveDispatcherProfile{
		name:      "dispatcherprofile_remove",
		rpcMethod: utils.APIerSv1RemoveDispatcherProfile,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdRemoveDispatcherProfile struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantIDWithCache
	*CommandExecuter
}

func (self *CmdRemoveDispatcherProfile) Name() string {
	return self.name
}

func (self *CmdRemoveDispatcherProfile) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRemoveDispatcherProfile) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.TenantIDWithCache)
	}
	return self.rpcParams
}

func (self *CmdRemoveDispatcherProfile) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRemoveDispatcherProfile) RpcResult() any {
	var s string
	return &s
}
