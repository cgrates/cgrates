// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetFilter{
		name:      "filter",
		rpcMethod: utils.APIerSv1GetFilter,
		rpcParams: &utils.TenantID{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetFilter struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantID
	*CommandExecuter
}

func (self *CmdGetFilter) Name() string {
	return self.name
}

func (self *CmdGetFilter) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetFilter) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantID{}
	}
	return self.rpcParams
}

func (self *CmdGetFilter) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetFilter) RpcResult() any {
	var atr engine.Filter
	return &atr
}
