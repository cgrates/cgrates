/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

import "github.com/cgrates/cgrates/engine"

func init() {
	c := &CmdGetDestination{
		name:      "destination",
		rpcMethod: "ApierV1.GetDestination",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetDestination struct {
	name      string
	rpcMethod string
	rpcParams *StringWrapper
	rpcResult engine.Destination
	*CommandExecuter
}

func (self *CmdGetDestination) Name() string {
	return self.name
}

func (self *CmdGetDestination) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetDestination) RpcParams() interface{} {
	if self.rpcParams == nil {
		self.rpcParams = &StringWrapper{}
	}
	return self.rpcParams
}

func (self *CmdGetDestination) RpcResult() interface{} {
	return &self.rpcResult
}
