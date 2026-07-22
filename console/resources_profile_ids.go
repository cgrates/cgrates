// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetResourceIDs{
		name:      "resources_profile_ids",
		rpcMethod: utils.APIerSv1GetResourceProfileIDs,
		rpcParams: &utils.PaginatorWithTenant{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetResourceIDs struct {
	name      string
	rpcMethod string
	rpcParams *utils.PaginatorWithTenant
	*CommandExecuter
}

func (self *CmdGetResourceIDs) Name() string {
	return self.name
}

func (self *CmdGetResourceIDs) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetResourceIDs) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.PaginatorWithTenant{}
	}
	return self.rpcParams
}

func (self *CmdGetResourceIDs) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetResourceIDs) RpcResult() any {
	var atr []string
	return &atr
}
