// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSetStorDBVersions{
		name:      "set_stordb_versions",
		rpcMethod: utils.APIerSv1SetStorDBVersions,
		rpcParams: &v1.SetVersionsArg{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdSetStorDBVersions struct {
	name      string
	rpcMethod string
	rpcParams *v1.SetVersionsArg
	*CommandExecuter
}

func (self *CmdSetStorDBVersions) Name() string {
	return self.name
}

func (self *CmdSetStorDBVersions) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetStorDBVersions) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.SetVersionsArg{}
	}
	return self.rpcParams
}

func (self *CmdSetStorDBVersions) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetStorDBVersions) RpcResult() any {
	var atr string
	return &atr
}
