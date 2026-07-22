// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetDispatcherHostIDs{
		name:      "dispatchers_host_ids",
		rpcMethod: utils.APIerSv1GetDispatcherHostIDs,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetDispatcherHostIDs struct {
	name      string
	rpcMethod string
	rpcParams *utils.PaginatorWithTenant
	*CommandExecuter
}

func (self *CmdGetDispatcherHostIDs) Name() string {
	return self.name
}

func (self *CmdGetDispatcherHostIDs) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetDispatcherHostIDs) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.PaginatorWithTenant)
	}
	return self.rpcParams
}

func (self *CmdGetDispatcherHostIDs) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetDispatcherHostIDs) RpcResult() any {
	var s []string
	return &s
}
