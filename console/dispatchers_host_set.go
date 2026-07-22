// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSetDispatcherHost{
		name:      "dispatchers_host_set",
		rpcMethod: utils.APIerSv1SetDispatcherHost,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdSetDispatcherHost struct {
	name      string
	rpcMethod string
	rpcParams *engine.DispatcherHostWithAPIOpts
	*CommandExecuter
}

func (self *CmdSetDispatcherHost) Name() string {
	return self.name
}

func (self *CmdSetDispatcherHost) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetDispatcherHost) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.DispatcherHostWithAPIOpts{
			DispatcherHost: new(engine.DispatcherHost),
			APIOpts:        make(map[string]any),
		}
	}
	return self.rpcParams
}

func (self *CmdSetDispatcherHost) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetDispatcherHost) RpcResult() any {
	var s string
	return &s
}
