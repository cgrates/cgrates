// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdLoaderRemove{
		name:      "loader_remove",
		rpcMethod: utils.LoaderSv1Remove,
		rpcParams: &loaders.ArgsProcessFolder{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdLoaderRemove struct {
	name      string
	rpcMethod string
	rpcParams *loaders.ArgsProcessFolder
	*CommandExecuter
}

func (self *CmdLoaderRemove) Name() string {
	return self.name
}

func (self *CmdLoaderRemove) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdLoaderRemove) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &loaders.ArgsProcessFolder{}
	}
	return self.rpcParams
}

func (self *CmdLoaderRemove) PostprocessRpcParams() error {
	return nil
}

func (self *CmdLoaderRemove) RpcResult() any {
	var s string
	return &s
}
