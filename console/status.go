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
	commands["status"] = &CmdStatus{}
}

type CmdStatus struct {
	rpcMethod string
	rpcParams string
	rpcResult string
}

func (self *CmdStatus) Usage(name string) string {
	return fmt.Sprintf("\n\tUsage: %s [cfg_opts...{-h}] status", name)
}

func (self *CmdStatus) defaults() error {
	self.rpcMethod = "Responder.Status"
	return nil
}

func (self *CmdStatus) FromArgs(args []string) error {
	self.defaults()
	return nil
}

func (self *CmdStatus) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdStatus) RpcParams() interface{} {
	return &self.rpcParams
}

func (self *CmdStatus) RpcResult() interface{} {
	return &self.rpcResult
}
