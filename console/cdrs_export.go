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

import "github.com/cgrates/cgrates/utils"

func init() {
	c := &CmdExportCdrs{
		name:      "cdrs_export",
		rpcMethod: "ApierV2.ExportCdrsToFile",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdExportCdrs struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrExportCdrsToFile
	*CommandExecuter
}

func (self *CmdExportCdrs) Name() string {
	return self.name
}

func (self *CmdExportCdrs) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdExportCdrs) RpcParams(ptr bool) interface{} {
	if self.rpcParams == nil {
		self.rpcParams = &utils.AttrExportCdrsToFile{}
	}
	if ptr {
		return self.rpcParams
	}
	return *self.rpcParams
}

func (self *CmdExportCdrs) PostprocessRpcParams() error {
	return nil
}

func (self *CmdExportCdrs) RpcResult() interface{} {
	return &utils.AttrExportCdrsToFile{}
}
