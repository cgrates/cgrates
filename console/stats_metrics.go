/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package console

import (
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetStatQueueStringMetrics{
		name:      "stats_metrics",
		rpcMethod: utils.StatSv1GetQueueStringMetrics,
		rpcParams: &dispatchers.TntIDWithApiKey{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetStatQueueStringMetrics struct {
	name      string
	rpcMethod string
	rpcParams *dispatchers.TntIDWithApiKey
	*CommandExecuter
}

func (self *CmdGetStatQueueStringMetrics) Name() string {
	return self.name
}

func (self *CmdGetStatQueueStringMetrics) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetStatQueueStringMetrics) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &dispatchers.TntIDWithApiKey{}
	}
	return self.rpcParams
}

func (self *CmdGetStatQueueStringMetrics) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetStatQueueStringMetrics) RpcResult() interface{} {
	var atr *map[string]string
	return &atr
}
