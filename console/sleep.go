// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import "github.com/cgrates/cgrates/utils"

func init() {
	c := &CmdSleep{
		name:      "sleep",
		rpcMethod: utils.CoreSv1Sleep,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdSleep struct {
	name      string
	rpcMethod string
	rpcParams any
	*CommandExecuter
}

func (cmd *CmdSleep) Name() string {
	return cmd.name
}

func (cmd *CmdSleep) RpcMethod() string {
	return cmd.rpcMethod
}

func (cmd *CmdSleep) RpcParams(reset bool) any {
	if reset || cmd.rpcParams == nil {
		cmd.rpcParams = &utils.DurationArgs{}
	}
	return cmd.rpcParams
}

func (cmd *CmdSleep) PostprocessRpcParams() (err error) {
	params := new(utils.DurationArgs)
	if val, can := cmd.rpcParams.(*StringWrapper); can {
		params.Duration, err = utils.ParseDurationWithNanosecs(val.Item)
		if err != nil {
			return
		}
	}
	cmd.rpcParams = params
	return
}

func (cmd *CmdSleep) RpcResult() any {
	var s string
	return &s
}
