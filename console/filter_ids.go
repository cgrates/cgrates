// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetFilterIDs{
		name:      "filter_ids",
		rpcMethod: utils.APIerSv1GetFilterIDs,
		rpcParams: &utils.TenantArgWithPaginator{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetFilterIDs struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantArgWithPaginator
	*CommandExecuter
}

func (self *CmdGetFilterIDs) Name() string {
	return self.name
}

func (self *CmdGetFilterIDs) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetFilterIDs) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantArgWithPaginator{}
	}
	return self.rpcParams
}

func (self *CmdGetFilterIDs) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetFilterIDs) RpcResult() any {
	var atr []string
	return &atr
}
