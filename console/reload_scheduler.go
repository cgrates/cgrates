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
)

func init() {
	commands["reload_scheduler"] = &CmdReloadScheduler{}
}

// Commander implementation
type CmdReloadScheduler struct {
	rpcMethod string
	rpcParams string
	rpcResult string
}

// name should be exec's name
func (self *CmdReloadScheduler) Usage(name string) string {
	return fmt.Sprintf("\n\tUsage: cgr-console [cfg_opts...{-h}] reload_scheduler")
}

// set param defaults
func (self *CmdReloadScheduler) defaults() error {
	self.rpcMethod = "ApierV1.ReloadScheduler"
	return nil
}

// Parses command line args and builds CmdBalance value
func (self *CmdReloadScheduler) FromArgs(args []string) error {
	self.defaults()
	return nil
}

func (self *CmdReloadScheduler) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdReloadScheduler) RpcParams() interface{} {
	return self.rpcParams
}

func (self *CmdReloadScheduler) RpcResult() interface{} {
	return &self.rpcResult
}
