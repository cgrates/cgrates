// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetStatQueueIDs{
		name:      "statqueue_ids",
		rpcMethod: utils.APIerSv1GetStatQueueProfileIDs,
		rpcParams: &utils.TenantArgWithPaginator{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetStatQueueIDs struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantArgWithPaginator
	*CommandExecuter
}

func (self *CmdGetStatQueueIDs) Name() string {
	return self.name
}

func (self *CmdGetStatQueueIDs) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetStatQueueIDs) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantArgWithPaginator{}
	}
	return self.rpcParams
}

func (self *CmdGetStatQueueIDs) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetStatQueueIDs) RpcResult() any {
	var atr []string
	return &atr
}
