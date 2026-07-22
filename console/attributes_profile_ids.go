// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetAttributeIDs{
		name:      "attributes_profile_ids",
		rpcMethod: utils.APIerSv1GetAttributeProfileIDs,
		rpcParams: &utils.PaginatorWithTenant{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetAttributeIDs struct {
	name      string
	rpcMethod string
	rpcParams *utils.PaginatorWithTenant
	*CommandExecuter
}

func (self *CmdGetAttributeIDs) Name() string {
	return self.name
}

func (self *CmdGetAttributeIDs) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetAttributeIDs) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.PaginatorWithTenant{}
	}
	return self.rpcParams
}

func (self *CmdGetAttributeIDs) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetAttributeIDs) RpcResult() any {
	var atr []string
	return &atr
}
