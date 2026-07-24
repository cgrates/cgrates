// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSetDispatcherProfile{
		name:      "dispatcherprofile_set",
		rpcMethod: utils.APIerSv1SetDispatcherProfile,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdSetDispatcherProfile struct {
	name      string
	rpcMethod string
	rpcParams *v1.DispatcherWithCache
	*CommandExecuter
}

func (self *CmdSetDispatcherProfile) Name() string {
	return self.name
}

func (self *CmdSetDispatcherProfile) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetDispatcherProfile) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(v1.DispatcherWithCache)
	}
	return self.rpcParams
}

func (self *CmdSetDispatcherProfile) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetDispatcherProfile) RpcResult() any {
	var s string
	return &s
}
