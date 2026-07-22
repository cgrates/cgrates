// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetStorDBVersions{
		name:      "stordb_versions",
		rpcMethod: utils.APIerSv1GetStorDBVersions,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetStorDBVersions struct {
	name      string
	rpcMethod string
	rpcParams *EmptyWrapper
	*CommandExecuter
}

func (self *CmdGetStorDBVersions) Name() string {
	return self.name
}

func (self *CmdGetStorDBVersions) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetStorDBVersions) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &EmptyWrapper{}
	}
	return self.rpcParams
}

func (self *CmdGetStorDBVersions) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetStorDBVersions) RpcResult() any {
	s := engine.Versions{}
	return &s
}

func (self *CmdGetStorDBVersions) ClientArgs() (args []string) {
	return
}
