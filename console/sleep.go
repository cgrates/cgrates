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
	rpcParams interface{}
	*CommandExecuter
}

func (cmd *CmdSleep) Name() string {
	return cmd.name
}

func (cmd *CmdSleep) RpcMethod() string {
	return cmd.rpcMethod
}

func (cmd *CmdSleep) RpcParams(reset bool) interface{} {
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

func (cmd *CmdSleep) RpcResult() interface{} {
	var s string
	return &s
}
