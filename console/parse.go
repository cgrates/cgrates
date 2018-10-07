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
	c := &CmdParse{
		name:      "parse",
		rpcParams: &AttrParse{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type AttrParse struct {
	Expression string
	Value      string
}

type CmdParse struct {
	name      string
	rpcMethod string
	rpcParams *AttrParse
	*CommandExecuter
}

func (self *CmdParse) Name() string {
	return self.name
}

func (self *CmdParse) RpcMethod() string {
	return ""
}

func (self *CmdParse) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &AttrParse{}
	}
	return self.rpcParams
}

func (self *CmdParse) RpcResult() interface{} {
	return nil
}

func (self *CmdParse) PostprocessRpcParams() error {
	return nil
}

func (self *CmdParse) LocalExecute() string {
	if self.rpcParams.Expression == "" {
		return "Empty expression error"
	}
	if self.rpcParams.Value == "" {
		return "Empty value error"
	}
	if rsrField, err := utils.NewRSRField(self.rpcParams.Expression); err != nil {
		return err.Error()
	} else if parsed, err := rsrField.Parse(self.rpcParams.Value); err != nil {
		return err.Error()
	} else {
		return parsed
	}
}
