// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	rpcParams any
	*CommandExecuter
}

func (cmd *CmdPassiveSessions) Name() string {
	return cmd.name
}

func (cmd *CmdPassiveSessions) RpcMethod() string {
	return cmd.rpcMethod
}

func (cmd *CmdPassiveSessions) RpcParams(reset bool) any {
	if reset || cmd.rpcParams == nil {
		cmd.rpcParams = &utils.SessionFilter{APIOpts: make(map[string]any)}
	}
	return cmd.rpcParams
}

func (cmd *CmdPassiveSessions) PostprocessRpcParams() error {
	param := cmd.rpcParams.(*utils.SessionFilter)
	cmd.rpcParams = param
	return nil
}

func (cmd *CmdPassiveSessions) RpcResult() any {
	var sessions []*sessions.ExternalSession
	return &sessions
}

func (cmd *CmdPassiveSessions) GetFormatedResult(result any) string {
	return GetFormatedSliceResult(result, utils.StringSet{
		utils.Usage:         {},
		utils.DurationIndex: {},
		utils.MaxRateUnit:   {},
		utils.DebitInterval: {},
	})
}
