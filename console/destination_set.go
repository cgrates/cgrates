// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import "github.com/cgrates/cgrates/utils"

func init() {
	c := &CmdSetDestination{
		name:      "destination_set",
		rpcMethod: utils.APIerSv1SetDestination,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdSetDestination struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrSetDestination
	rpcResult string
	*CommandExecuter
}

func (self *CmdSetDestination) Name() string {
	return self.name
}

func (self *CmdSetDestination) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetDestination) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.AttrSetDestination{}
	}
	return self.rpcParams
}

func (self *CmdSetDestination) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetDestination) RpcResult() any {
	var s string
	return &s
}
