/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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

func init() {
	c := &CmdCdrStatsQueueIds{
		name:      "cdrstats_queueids",
		rpcMethod: "CDRStatsV1.GetQueueIds",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdCdrStatsQueueIds struct {
	name      string
	rpcMethod string
	rpcParams *StringWrapper
	*CommandExecuter
}

func (self *CmdCdrStatsQueueIds) Name() string {
	return self.name
}

func (self *CmdCdrStatsQueueIds) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdCdrStatsQueueIds) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &StringWrapper{}
	}
	return self.rpcParams
}

func (self *CmdCdrStatsQueueIds) PostprocessRpcParams() error {
	return nil
}

func (self *CmdCdrStatsQueueIds) RpcResult() interface{} {
	var s []string
	return &s
}

func (self *CmdCdrStatsQueueIds) ClientArgs() (args []string) {
	return
}
