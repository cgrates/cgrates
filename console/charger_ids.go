// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetChargerIDs{
		name:      "chargers_ids",
		rpcMethod: utils.APIerSv1GetChargerProfileIDs,
		rpcParams: &utils.TenantArgWithPaginator{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetChargerIDs struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantArgWithPaginator
	*CommandExecuter
}

func (self *CmdGetChargerIDs) Name() string {
	return self.name
}

func (self *CmdGetChargerIDs) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetChargerIDs) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantArgWithPaginator{}
	}
	return self.rpcParams
}

func (self *CmdGetChargerIDs) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetChargerIDs) RpcResult() any {
	var atr []string
	return &atr
}
