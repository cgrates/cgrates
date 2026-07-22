// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetThresholdIDs{
		name:      "thresholds_profile_ids",
		rpcMethod: utils.APIerSv1GetThresholdProfileIDs,
		rpcParams: &utils.PaginatorWithTenant{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdGetThresholdIDs struct {
	name      string
	rpcMethod string
	rpcParams *utils.PaginatorWithTenant
	*CommandExecuter
}

func (self *CmdGetThresholdIDs) Name() string {
	return self.name
}

func (self *CmdGetThresholdIDs) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetThresholdIDs) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.PaginatorWithTenant{}
	}
	return self.rpcParams
}

func (self *CmdGetThresholdIDs) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetThresholdIDs) RpcResult() any {
	var atr []string
	return &atr
}
