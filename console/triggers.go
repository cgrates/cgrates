// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetTriggers{
		name:      "triggers",
		rpcMethod: utils.APIerSv1GetActionTriggers,
		rpcParams: &v1.AttrGetActionTriggers{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetTriggers struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrGetActionTriggers
	*CommandExecuter
}

func (self *CmdGetTriggers) Name() string {
	return self.name
}

func (self *CmdGetTriggers) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetTriggers) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrGetActionTriggers{}
	}
	return self.rpcParams
}

func (self *CmdGetTriggers) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetTriggers) RpcResult() any {
	var atr engine.ActionTriggers
	return &atr
}

func (self *CmdGetTriggers) GetFormatedResult(result any) string {
	return GetFormatedSliceResult(result, utils.StringSet{
		utils.MinSleep: {},
	})
}
