// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSetResource{
		name:      "resources_profile_set",
		rpcMethod: utils.APIerSv1SetResourceProfile,
		rpcParams: &engine.ResourceProfileWithAPIOpts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdSetResource struct {
	name      string
	rpcMethod string
	rpcParams *engine.ResourceProfileWithAPIOpts
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
		self.rpcParams = &engine.ResourceProfileWithAPIOpts{
			ResourceProfile: new(engine.ResourceProfile),
			APIOpts:         make(map[string]any),
		}
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
