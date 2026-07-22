// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

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
	return utils.EmptyString
}

func (self *CmdParse) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &AttrParse{}
	}
	return self.rpcParams
}

func (self *CmdParse) RpcResult() any {
	return nil
}

func (self *CmdParse) PostprocessRpcParams() error {
	return nil
}

func (self *CmdParse) LocalExecute() string {
	if self.rpcParams.Expression == utils.EmptyString {
		return "Empty expression error"
	}
	if self.rpcParams.Value == utils.EmptyString {
		return "Empty value error"
	}
	if rsrField, err := config.NewRSRParser(self.rpcParams.Expression); err != nil {
		return err.Error()
	} else if parsed, err := rsrField.ParseValue(self.rpcParams.Value); err != nil {
		return err.Error()
	} else {
		return parsed
	}
}
