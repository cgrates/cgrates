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
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdPassiveSessions{
		name:      "passive_sessions",
		rpcMethod: utils.SessionSv1GetPassiveSessions,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdPassiveSessions struct {
	name      string
	rpcMethod string
	rpcParams interface{}
	*CommandExecuter
}

func (cmd *CmdPassiveSessions) Name() string {
	return cmd.name
}

func (cmd *CmdPassiveSessions) RpcMethod() string {
	return cmd.rpcMethod
}

func (cmd *CmdPassiveSessions) RpcParams(reset bool) interface{} {
	if reset || cmd.rpcParams == nil {
		cmd.rpcParams = &utils.SessionFilter{APIOpts: make(map[string]interface{})}
	}
	return cmd.rpcParams
}

func (cmd *CmdPassiveSessions) PostprocessRpcParams() error {
	param := cmd.rpcParams.(*utils.SessionFilter)
	cmd.rpcParams = param
	return nil
}

func (cmd *CmdPassiveSessions) RpcResult() interface{} {
	var sessions []*sessions.ExternalSession
	return &sessions
}

func (cmd *CmdPassiveSessions) GetFormatedResult(result interface{}) string {
	return GetFormatedSliceResult(result, utils.StringSet{
		utils.Usage:         {},
		utils.DurationIndex: {},
		utils.MaxRateUnit:   {},
		utils.DebitInterval: {},
	})
}
