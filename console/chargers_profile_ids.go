// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetChargerIDs{
		name:      "chargers_profile_ids",
		rpcMethod: utils.APIerSv1GetChargerProfileIDs,
		rpcParams: &utils.PaginatorWithTenant{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetChargerIDs struct {
	name      string
	rpcMethod string
	rpcParams *utils.PaginatorWithTenant
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
		self.rpcParams = &utils.PaginatorWithTenant{}
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
