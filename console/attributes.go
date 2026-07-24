// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetAttributes{
		name:      "attributes",
		rpcMethod: utils.APIerSv1GetAttributeProfile,
		rpcParams: &utils.TenantID{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetAttributes struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantID
	*CommandExecuter
}

func (self *CmdGetAttributes) Name() string {
	return self.name
}

func (self *CmdGetAttributes) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetAttributes) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantID{}
	}
	return self.rpcParams
}

func (self *CmdGetAttributes) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetAttributes) RpcResult() any {
	var atr engine.AttributeProfile
	return &atr
}
