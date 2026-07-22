// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdLoadTimes{
		name:      "get_load_times",
		rpcMethod: utils.APIerSv1GetLoadTimes,
		rpcParams: &v1.LoadTimeArgs{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdLoadTimes struct {
	name      string
	rpcMethod string
	rpcParams *v1.LoadTimeArgs
	*CommandExecuter
}

func (self *CmdLoadTimes) Name() string {
	return self.name
}

func (self *CmdLoadTimes) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdLoadTimes) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.LoadTimeArgs{}
	}
	return self.rpcParams
}

func (self *CmdLoadTimes) PostprocessRpcParams() error {
	return nil
}

func (self *CmdLoadTimes) RpcResult() any {
	a := make(map[string]string, 0)
	return &a
}
