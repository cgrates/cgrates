// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import "github.com/cgrates/cgrates/utils"

func init() {
	c := &CmdRemoveThreshold{
		name:      "thresholds_profile_remove",
		rpcMethod: utils.APIerSv1RemoveThresholdProfile,
		rpcParams: &utils.TenantIDWithAPIOpts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdRemoveThreshold struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantIDWithAPIOpts
	*CommandExecuter
}

func (self *CmdRemoveThreshold) Name() string {
	return self.name
}

func (self *CmdRemoveThreshold) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRemoveThreshold) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantIDWithAPIOpts{APIOpts: make(map[string]any)}
	}
	return self.rpcParams
}

func (self *CmdRemoveThreshold) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRemoveThreshold) RpcResult() any {
	var s string
	return &s
}
