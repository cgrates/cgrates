// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v2 "github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetDestination{
		name:      "destinations",
		rpcMethod: utils.APIerSv2GetDestinations,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetDestination struct {
	name      string
	rpcMethod string
	rpcParams *v2.AttrGetDestinations
	*CommandExecuter
}

func (self *CmdGetDestination) Name() string {
	return self.name
}

func (self *CmdGetDestination) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetDestination) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v2.AttrGetDestinations{}
	}
	return self.rpcParams
}

func (self *CmdGetDestination) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetDestination) RpcResult() any {
	a := make([]*engine.Destination, 0)
	return &a
}
