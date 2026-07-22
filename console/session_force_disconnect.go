// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSessionsForceDisconnect{
		name:      "session_force_disconnect",
		rpcMethod: utils.SessionSv1ForceDisconnect,
		rpcParams: utils.SessionFilterWithEvent{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdSessionsForceDisconnect struct {
	name      string
	rpcMethod string
	rpcParams utils.SessionFilterWithEvent
	*CommandExecuter
}

func (cmd *CmdSessionsForceDisconnect) Name() string {
	return cmd.name
}

func (cmd *CmdSessionsForceDisconnect) RpcMethod() string {
	return cmd.rpcMethod
}

func (cmd *CmdSessionsForceDisconnect) RpcParams(reset bool) any {
	if reset || cmd.rpcParams.SessionFilter == nil {
		cmd.rpcParams.SessionFilter = &utils.SessionFilter{
			APIOpts: make(map[string]any),
		}
	}
	return cmd.rpcParams
}

func (cmd *CmdSessionsForceDisconnect) PostprocessRpcParams() error {
	param := cmd.rpcParams
	cmd.rpcParams = param
	return nil
}

func (cmd *CmdSessionsForceDisconnect) RpcResult() any {
	var sessions string
	return &sessions
}

func (cmd *CmdSessionsForceDisconnect) GetFormatedResult(result any) string {
	return GetFormatedSliceResult(result, utils.StringSet{
		utils.Usage:         {},
		utils.DurationIndex: {},
		utils.MaxRateUnit:   {},
		utils.DebitInterval: {},
	})
}
