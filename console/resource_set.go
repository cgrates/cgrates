// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSetResource{
		name:      "resource_set",
		rpcMethod: utils.APIerSv1SetResourceProfile,
		rpcParams: &v1.ResourceWithCache{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdSetResource struct {
	name      string
	rpcMethod string
	rpcParams *v1.ResourceWithCache
	*CommandExecuter
}

func (self *CmdSetResource) Name() string {
	return self.name
}

func (self *CmdSetResource) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetResource) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.ResourceWithCache{}
	}
	return self.rpcParams
}

func (self *CmdSetResource) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetResource) RpcResult() any {
	var s string
	return &s
}
