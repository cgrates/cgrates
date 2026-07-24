// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"time"

	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdResourceAllocate{
		name:      "resources_allocate",
		rpcMethod: utils.ResourceSv1AllocateResources,
		rpcParams: &utils.ArgRSv1ResourceUsage{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdResourceAllocate struct {
	name      string
	rpcMethod string
	rpcParams *utils.ArgRSv1ResourceUsage
	*CommandExecuter
}

func (self *CmdResourceAllocate) Name() string {
	return self.name
}

func (self *CmdResourceAllocate) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdResourceAllocate) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.ArgRSv1ResourceUsage{ArgDispatcher: new(utils.ArgDispatcher)}
	}
	return self.rpcParams
}

func (self *CmdResourceAllocate) PostprocessRpcParams() error {
	if self.rpcParams != nil && self.rpcParams.CGREvent != nil &&
		self.rpcParams.CGREvent.Time == nil {
		self.rpcParams.CGREvent.Time = utils.TimePointer(time.Now())
	}
	return nil
}

func (self *CmdResourceAllocate) RpcResult() any {
	var atr *string
	return &atr
}
