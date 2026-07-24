// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import "github.com/cgrates/cgrates/engine"

func init() {
	c := &CmdGetMaxUsage{
		name:      "maxusage",
		rpcMethod: "APIerSv1.GetMaxUsage",
		clientArgs: []string{"ToR", "RequestType", "Tenant",
			"Category", "Account", "Subject", "Destination",
			"SetupTime", "AnswerTime", "Usage", "ExtraFields"},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetMaxUsage struct {
	name       string
	rpcMethod  string
	rpcParams  *engine.UsageRecord
	clientArgs []string
	*CommandExecuter
}

func (self *CmdGetMaxUsage) Name() string {
	return self.name
}

func (self *CmdGetMaxUsage) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetMaxUsage) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(engine.UsageRecord)
	}
	return self.rpcParams
}

func (self *CmdGetMaxUsage) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetMaxUsage) RpcResult() any {
	var f int64
	return &f
}

func (self *CmdGetMaxUsage) ClientArgs() []string {
	return self.clientArgs
}
