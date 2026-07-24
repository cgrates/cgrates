// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"time"

	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdResourceAuthorize{
		name:      "resources_authorize",
		rpcMethod: utils.ResourceSv1AuthorizeResources,
		rpcParams: &utils.ArgRSv1ResourceUsage{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdResourceAuthorize struct {
	name      string
	rpcMethod string
	rpcParams *utils.ArgRSv1ResourceUsage
	*CommandExecuter
}

func (self *CmdResourceAuthorize) Name() string {
	return self.name
}

func (self *CmdResourceAuthorize) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdResourceAuthorize) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.ArgRSv1ResourceUsage{ArgDispatcher: new(utils.ArgDispatcher)}
	}
	return self.rpcParams
}

func (self *CmdResourceAuthorize) PostprocessRpcParams() error {
	if self.rpcParams != nil && self.rpcParams.CGREvent != nil &&
		self.rpcParams.CGREvent.Time == nil {
		self.rpcParams.CGREvent.Time = utils.TimePointer(time.Now())
	}
	return nil
}

func (self *CmdResourceAuthorize) RpcResult() any {
	var atr *string
	return &atr
}
