// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetAccounts{
		name:      "accounts",
		rpcMethod: utils.APIerSv2GetAccounts,
		rpcParams: &utils.AttrGetAccounts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetAccounts struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrGetAccounts
	*CommandExecuter
}

func (self *CmdGetAccounts) Name() string {
	return self.name
}

func (self *CmdGetAccounts) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetAccounts) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.AttrGetAccounts{}
	}
	return self.rpcParams
}

func (self *CmdGetAccounts) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetAccounts) RpcResult() any {
	a := make([]*engine.Account, 0)
	return &a
}

func (self *CmdGetAccounts) GetFormatedResult(result any) string {
	return GetFormatedSliceResult(result, utils.StringSet{
		utils.MinSleep: {},
	})
}
