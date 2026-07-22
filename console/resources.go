// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetResource{
		name:      "resources",
		rpcMethod: utils.ResourceSv1GetResource,
		rpcParams: &utils.TenantIDWithAPIOpts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetResource struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantIDWithAPIOpts
	*CommandExecuter
}

func (self *CmdGetResource) Name() string {
	return self.name
}

func (self *CmdGetResource) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetResource) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantIDWithAPIOpts{
			TenantID: new(utils.TenantID),
			APIOpts:  map[string]any{},
		}
	}
	return self.rpcParams
}

func (self *CmdGetResource) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetResource) RpcResult() any {
	var atr engine.Resource
	return &atr
}
