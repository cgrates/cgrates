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

import (
	"fmt"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	commands["reload_cache"] = &CmdReloadCache{}
}

// Commander implementation
type CmdReloadCache struct {
	rpcMethod string
	rpcParams *utils.ApiReloadCache
	rpcResult string
}

// name should be exec's name
func (self *CmdReloadCache) Usage(name string) string {
	return fmt.Sprintf("\n\tUsage: cgr-console [cfg_opts...{-h}] reload_cache")
}

// set param defaults
func (self *CmdReloadCache) defaults() error {
	self.rpcMethod = "ApierV1.ReloadCache"
	return nil
}

// Parses command line args and builds CmdBalance value
func (self *CmdReloadCache) FromArgs(args []string) error {
	self.defaults()
	return nil
}

func (self *CmdReloadCache) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdReloadCache) RpcParams() interface{} {
	return self.rpcParams
}

func (self *CmdReloadCache) RpcResult() interface{} {
	return &self.rpcResult
}
