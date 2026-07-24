// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSetDispatcherHost{
		name:      "dispatcher_host_set",
		rpcMethod: utils.APIerSv1SetDispatcherHost,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdSetDispatcherHost struct {
	name      string
	rpcMethod string
	rpcParams *v1.DispatcherHostWithCache
	*CommandExecuter
}

func (self *CmdSetDispatcherHost) Name() string {
	return self.name
}

func (self *CmdSetDispatcherHost) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetDispatcherHost) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(v1.DispatcherHostWithCache)
	}
	return self.rpcParams
}

func (self *CmdSetDispatcherHost) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetDispatcherHost) RpcResult() any {
	var s string
	return &s
}
