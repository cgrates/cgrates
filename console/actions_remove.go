// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdRemoveActions{
		name:      "actions_remove",
		rpcMethod: utils.APIerSv1RemoveActions,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdRemoveActions struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrRemoveActions
	*CommandExecuter
}

func (self *CmdRemoveActions) Name() string {
	return self.name
}

func (self *CmdRemoveActions) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRemoveActions) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrRemoveActions{}
	}
	return self.rpcParams
}

func (self *CmdRemoveActions) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRemoveActions) RpcResult() any {
	var s string
	return &s
}
