// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdRemoveFilterIndexes{
		name:      "filter_indexes_remove",
		rpcMethod: utils.APIerSv1RemoveFilterIndexes,
		rpcParams: &v1.AttrRemFilterIndexes{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdRemoveFilterIndexes struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrRemFilterIndexes
	*CommandExecuter
}

func (self *CmdRemoveFilterIndexes) Name() string {
	return self.name
}

func (self *CmdRemoveFilterIndexes) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRemoveFilterIndexes) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrRemFilterIndexes{}
	}
	return self.rpcParams
}

func (self *CmdRemoveFilterIndexes) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRemoveFilterIndexes) RpcResult() any {
	var atr string
	return &atr
}
