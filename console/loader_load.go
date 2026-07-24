// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdLoaderLoad{
		name:      "loader_load",
		rpcMethod: utils.LoaderSv1Load,
		rpcParams: &loaders.ArgsProcessFolder{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdLoaderLoad struct {
	name      string
	rpcMethod string
	rpcParams *loaders.ArgsProcessFolder
	*CommandExecuter
}

func (self *CmdLoaderLoad) Name() string {
	return self.name
}

func (self *CmdLoaderLoad) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdLoaderLoad) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &loaders.ArgsProcessFolder{}
	}
	return self.rpcParams
}

func (self *CmdLoaderLoad) PostprocessRpcParams() error {
	return nil
}

func (self *CmdLoaderLoad) RpcResult() any {
	var s string
	return &s
}
