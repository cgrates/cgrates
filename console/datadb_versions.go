// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetDataDBVersions{
		name:      "datadb_versions",
		rpcMethod: utils.APIerSv1GetDataDBVersions,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetDataDBVersions struct {
	name      string
	rpcMethod string
	rpcParams *EmptyWrapper
	*CommandExecuter
}

func (self *CmdGetDataDBVersions) Name() string {
	return self.name
}

func (self *CmdGetDataDBVersions) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetDataDBVersions) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &EmptyWrapper{}
	}
	return self.rpcParams
}

func (self *CmdGetDataDBVersions) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetDataDBVersions) RpcResult() any {
	var s engine.Versions
	return &s
}

func (self *CmdGetDataDBVersions) ClientArgs() (args []string) {
	return
}
