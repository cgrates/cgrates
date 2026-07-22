// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdStatus{
		name:      "status",
		rpcMethod: utils.CoreSv1Status,
		rpcParams: &cores.V1StatusParams{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdStatus struct {
	name      string
	rpcMethod string
	rpcParams *cores.V1StatusParams
	*CommandExecuter
}

func (self *CmdStatus) Name() string {
	return self.name
}

func (self *CmdStatus) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdStatus) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &cores.V1StatusParams{
			APIOpts: make(map[string]any),
		}
	}
	return self.rpcParams
}

func (self *CmdStatus) PostprocessRpcParams() error {
	return nil
}

func (self *CmdStatus) RpcResult() any {
	var s map[string]any
	return &s
}
