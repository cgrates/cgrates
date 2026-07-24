// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSetFilter{
		name:      "filter_set",
		rpcMethod: utils.APIerSv1SetFilter,
		rpcParams: &v1.FilterWithCache{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdSetFilter struct {
	name      string
	rpcMethod string
	rpcParams *v1.FilterWithCache
	*CommandExecuter
}

func (self *CmdSetFilter) Name() string {
	return self.name
}

func (self *CmdSetFilter) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetFilter) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.FilterWithCache{Filter: new(engine.Filter)}
	}
	return self.rpcParams
}

func (self *CmdSetFilter) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetFilter) RpcResult() any {
	var s string
	return &s
}
