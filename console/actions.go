// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v2 "github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetActions{
		name:      "actions",
		rpcMethod: utils.APIerSv2GetActions,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetActions struct {
	name      string
	rpcMethod string
	rpcParams *v2.AttrGetActions
	*CommandExecuter
}

func (self *CmdGetActions) Name() string {
	return self.name
}

func (self *CmdGetActions) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetActions) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v2.AttrGetActions{}
	}
	return self.rpcParams
}

func (self *CmdGetActions) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetActions) RpcResult() any {
	a := make(map[string]engine.Actions, 0)
	return &a
}
