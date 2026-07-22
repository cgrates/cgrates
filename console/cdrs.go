// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetCDRs{
		name:      "cdrs",
		rpcMethod: utils.CDRsV1GetCDRs,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetCDRs struct {
	name      string
	rpcMethod string
	rpcParams *utils.RPCCDRsFilterWithAPIOpts
	*CommandExecuter
}

func (self *CmdGetCDRs) Name() string {
	return self.name
}

func (self *CmdGetCDRs) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetCDRs) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.RPCCDRsFilterWithAPIOpts{
			RPCCDRsFilter: new(utils.RPCCDRsFilter),
		}
	}
	return self.rpcParams
}

func (self *CmdGetCDRs) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetCDRs) RpcResult() any {
	a := make([]*engine.CDR, 0)
	return &a
}
