// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import "github.com/cgrates/cgrates/utils"

func init() {
	c := &CmdRemoveFilter{
		name:      "filter_remove",
		rpcMethod: utils.APIerSv1RemoveFilter,
		rpcParams: &utils.TenantIDWithAPIOpts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdRemoveFilter struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantIDWithAPIOpts
	*CommandExecuter
}

func (self *CmdRemoveFilter) Name() string {
	return self.name
}

func (self *CmdRemoveFilter) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRemoveFilter) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantIDWithAPIOpts{APIOpts: make(map[string]any)}
	}
	return self.rpcParams
}

func (self *CmdRemoveFilter) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRemoveFilter) RpcResult() any {
	var s string
	return &s
}
