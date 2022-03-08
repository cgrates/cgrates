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
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSessionsForceDisconnect{
		name:      "session_force_disconnect",
		rpcMethod: utils.SessionSv1ForceDisconnect,
		rpcParams: &utils.SessionFilter{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdSessionsForceDisconnect struct {
	name      string
	rpcMethod string
	rpcParams *utils.SessionFilter
	*CommandExecuter
}

func (cmd *CmdSessionsForceDisconnect) Name() string {
	return cmd.name
}

func (cmd *CmdSessionsForceDisconnect) RpcMethod() string {
	return cmd.rpcMethod
}

func (cmd *CmdSessionsForceDisconnect) RpcParams(reset bool) interface{} {
	if reset || cmd.rpcParams == nil {
		cmd.rpcParams = &utils.SessionFilter{APIOpts: make(map[string]interface{})}
	}
	return cmd.rpcParams
}

func (cmd *CmdSessionsForceDisconnect) PostprocessRpcParams() error {
	param := cmd.rpcParams
	cmd.rpcParams = param
	return nil
}

func (cmd *CmdSessionsForceDisconnect) RpcResult() interface{} {
	var sessions string
	return &sessions
}

func (cmd *CmdSessionsForceDisconnect) GetFormatedResult(result interface{}) string {
	return GetFormatedSliceResult(result, utils.StringSet{
		utils.Usage:         {},
		utils.DurationIndex: {},
		utils.MaxRateUnit:   {},
		utils.DebitInterval: {},
	})
}
