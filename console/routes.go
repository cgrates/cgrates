// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdRoutesSort{
		name:      "routes",
		rpcMethod: utils.RouteSv1GetRoutes,
		rpcParams: &utils.CGREvent{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdRoutesSort struct {
	name      string
	rpcMethod string
	rpcParams *utils.CGREvent
	*CommandExecuter
}

func (self *CmdRoutesSort) Name() string {
	return self.name
}

func (self *CmdRoutesSort) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRoutesSort) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.CGREvent)
	}
	return self.rpcParams
}

func (self *CmdRoutesSort) PostprocessRpcParams() error {
	if self.rpcParams != nil && self.rpcParams.Time == nil {
		self.rpcParams.Time = utils.TimePointer(time.Now())
	}
	return nil
}

func (self *CmdRoutesSort) RpcResult() any {
	var atr engine.SortedRoutesList
	return &atr
}
