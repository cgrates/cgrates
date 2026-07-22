// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetResourceProfile{
		name:      "resources_profile",
		rpcMethod: utils.APIerSv1GetResourceProfile,
		rpcParams: &utils.TenantID{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetResourceProfile struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantID
	*CommandExecuter
}

func (self *CmdGetResourceProfile) Name() string {
	return self.name
}

func (self *CmdGetResourceProfile) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetResourceProfile) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantID{}
	}
	return self.rpcParams
}

func (self *CmdGetResourceProfile) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetResourceProfile) RpcResult() any {
	var atr engine.ResourceProfile
	return &atr
}

func (self *CmdGetResourceProfile) GetFormatedResult(result any) string {
	return GetFormatedResult(result, utils.StringSet{
		utils.UsageTTL: {},
	})
}
