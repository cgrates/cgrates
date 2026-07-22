// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSetDataDBVersions{
		name:      "set_datadb_versions",
		rpcMethod: utils.APIerSv1SetDataDBVersions,
		rpcParams: &v1.SetVersionsArg{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdSetDataDBVersions struct {
	name      string
	rpcMethod string
	rpcParams *v1.SetVersionsArg
	*CommandExecuter
}

func (self *CmdSetDataDBVersions) Name() string {
	return self.name
}

func (self *CmdSetDataDBVersions) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetDataDBVersions) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.SetVersionsArg{}
	}
	return self.rpcParams
}

func (self *CmdSetDataDBVersions) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetDataDBVersions) RpcResult() any {
	var atr string
	return &atr
}
