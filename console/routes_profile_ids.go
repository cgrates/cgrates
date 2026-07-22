// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdRouteIDs{
		name:      "route_profile_ids",
		rpcMethod: utils.APIerSv1GetRouteProfileIDs,
		rpcParams: &utils.PaginatorWithTenant{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdRouteIDs struct {
	name      string
	rpcMethod string
	rpcParams *utils.PaginatorWithTenant
	*CommandExecuter
}

func (self *CmdRouteIDs) Name() string {
	return self.name
}

func (self *CmdRouteIDs) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRouteIDs) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.PaginatorWithTenant{}
	}
	return self.rpcParams
}

func (self *CmdRouteIDs) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRouteIDs) RpcResult() any {
	var atr []string
	return &atr
}
