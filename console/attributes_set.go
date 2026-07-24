// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v2 "github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSetAttributes{
		name:      "attributes_set",
		rpcMethod: utils.APIerSv2SetAttributeProfile,
		rpcParams: &v2.AttributeWithCache{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdSetAttributes struct {
	name      string
	rpcMethod string
	rpcParams *v2.AttributeWithCache
	*CommandExecuter
}

func (self *CmdSetAttributes) Name() string {
	return self.name
}

func (self *CmdSetAttributes) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetAttributes) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v2.AttributeWithCache{ExternalAttributeProfile: new(engine.ExternalAttributeProfile)}
	}
	return self.rpcParams
}

func (self *CmdSetAttributes) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetAttributes) RpcResult() any {
	var s string
	return &s
}
