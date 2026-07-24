// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"time"

	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdResourceRelease{
		name:      "resources_release",
		rpcMethod: utils.ResourceSv1ReleaseResources,
		rpcParams: &utils.ArgRSv1ResourceUsage{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdResourceRelease struct {
	name      string
	rpcMethod string
	rpcParams *utils.ArgRSv1ResourceUsage
	*CommandExecuter
}

func (self *CmdResourceRelease) Name() string {
	return self.name
}

func (self *CmdResourceRelease) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdResourceRelease) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.ArgRSv1ResourceUsage{ArgDispatcher: new(utils.ArgDispatcher)}
	}
	return self.rpcParams
}

func (self *CmdResourceRelease) PostprocessRpcParams() error {
	if self.rpcParams != nil && self.rpcParams.CGREvent != nil &&
		self.rpcParams.CGREvent.Time == nil {
		self.rpcParams.CGREvent.Time = utils.TimePointer(time.Now())
	}
	return nil
}

func (self *CmdResourceRelease) RpcResult() any {
	var atr *string
	return &atr
}
