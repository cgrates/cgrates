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

import "github.com/cgrates/cgrates/utils"

func init() {
	c := &CmdGetCacheAge{
		name:      "get_cache_age",
		rpcMethod: "ApierV1.GetCachedItemAge",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetCacheAge struct {
	name      string
	rpcMethod string
	rpcParams string
	rpcResult *utils.CachedItemAge
	*CommandExecuter
}

func (self *CmdGetCacheAge) Name() string {
	return self.name
}

func (self *CmdGetCacheAge) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetCacheAge) RpcParams() interface{} {
	return self.rpcParams
}

func (self *CmdGetCacheAge) RpcResult() interface{} {
	return &self.rpcResult
}
