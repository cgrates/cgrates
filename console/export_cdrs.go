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
	"time"
)

func init() {
	commands["export_cdrs"] = &CmdExportCdrs{}
}

// Commander implementation
type CmdExportCdrs struct {
	rpcMethod string
	rpcParams *utils.AttrExpFileCdrs
	rpcResult utils.ExportedFileCdrs
}

// name should be exec's name
func (self *CmdExportCdrs) Usage(name string) string {
	return fmt.Sprintf("\n\tUsage: cgr-console [cfg_opts...{-h}] export_cdrs <dry_run|csv> [<start_time|*one_month> [<stop_time> [remove_from_db]]]")
}

// set param defaults
func (self *CmdExportCdrs) defaults() error {
	self.rpcMethod = "ApierV1.ExportCdrsToFile"
	self.rpcParams = &utils.AttrExpFileCdrs{CdrFormat:"csv"}
	return nil
}

func (self *CmdExportCdrs) FromArgs(args []string) error {
	self.defaults()
	var timeStart, timeEnd string
	if len(args) < 3 {
		return fmt.Errorf(self.Usage(""))
	}
	if !utils.IsSliceMember(utils.CdreCdrFormats, args[2]) {
		return fmt.Errorf(self.Usage(""))
	}
	self.rpcParams.CdrFormat = args[2]
	switch len(args) {
	case 4:
		timeStart = args[3]
		
	case 5:
		timeStart = args[3]
		timeEnd = args[4]
	case 6:
		timeStart = args[3]
		timeEnd = args[4]
		if args[5] == "remove_from_db" {
			self.rpcParams.RemoveFromDb = true
		}
	}
	if timeStart == "*one_month" {
		now := time.Now()
		self.rpcParams.TimeStart = now.AddDate(0,-1,0).String()
		self.rpcParams.TimeEnd = now.String()
	} else {
		self.rpcParams.TimeStart = timeStart
		self.rpcParams.TimeEnd = timeEnd
	}
	return nil
}

func (self *CmdExportCdrs) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdExportCdrs) RpcParams() interface{} {
	return self.rpcParams
}

func (self *CmdExportCdrs) RpcResult() interface{} {
	return &self.rpcResult
}
