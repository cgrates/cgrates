// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetDispatcherProfileIDs{
		name:      "dispatchers_profile_ids",
		rpcMethod: utils.APIerSv1GetDispatcherProfileIDs,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetDispatcherProfileIDs struct {
	name      string
	rpcMethod string
	rpcParams *utils.PaginatorWithTenant
	*CommandExecuter
}

func (self *CmdGetDispatcherProfileIDs) Name() string {
	return self.name
}

func (self *CmdGetDispatcherProfileIDs) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetDispatcherProfileIDs) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.PaginatorWithTenant)
	}
	return self.rpcParams
}

func (self *CmdGetDispatcherProfileIDs) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetDispatcherProfileIDs) RpcResult() any {
	var s []string
	return &s
}
