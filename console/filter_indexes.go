// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetFilterIndexes{
		name:      "filter_indexes",
		rpcMethod: utils.APIerSv1GetFilterIndexes,
		rpcParams: &v1.AttrGetFilterIndexes{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetFilterIndexes struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrGetFilterIndexes
	*CommandExecuter
}

func (self *CmdGetFilterIndexes) Name() string {
	return self.name
}

func (self *CmdGetFilterIndexes) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetFilterIndexes) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrGetFilterIndexes{}
	}
	return self.rpcParams
}

func (self *CmdGetFilterIndexes) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetFilterIndexes) RpcResult() any {
	var atr []string
	return &atr
}
