// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetSharedGroup{
		name:      "sharedgroup",
		rpcMethod: utils.APIerSv1GetSharedGroup,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetSharedGroup struct {
	name      string
	rpcMethod string
	rpcParams *StringWrapper
	*CommandExecuter
}

func (self *CmdGetSharedGroup) Name() string {
	return self.name
}

func (self *CmdGetSharedGroup) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetSharedGroup) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &StringWrapper{}
	}
	return self.rpcParams
}

func (self *CmdGetSharedGroup) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetSharedGroup) RpcResult() any {
	return &engine.SharedGroup{}
}
