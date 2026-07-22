// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"strings"

	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdApierPing{
		name: "ping",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdApierPing struct {
	name      string
	rpcMethod string
	rpcParams any
	item      string
	*CommandExecuter
}

type ArgsPing struct {
	MethodName string
}

func (self *CmdApierPing) Name() string {
	return self.name
}

func (self *CmdApierPing) RpcMethod() string {
	switch strings.ToLower(self.item) {
	case utils.RoutesLow:
		return utils.RouteSv1Ping
	case utils.AttributesLow:
		return utils.AttributeSv1Ping
	case utils.ChargerSLow:
		return utils.ChargerSv1Ping
	case utils.ResourcesLow:
		return utils.ResourceSv1Ping
	case utils.StatServiceLow:
		return utils.StatSv1Ping
	case utils.ThresholdsLow:
		return utils.ThresholdSv1Ping
	case utils.SessionsLow:
		return utils.SessionSv1Ping
	case utils.DispatcherSLow:
		return utils.DispatcherSv1Ping
	case utils.AnalyzerSLow:
		return utils.AnalyzerSv1Ping
	case utils.SchedulerSLow:
		return utils.SchedulerSv1Ping
	case utils.RALsLow:
		return utils.RALsV1Ping
	case utils.ReplicatorLow:
		return utils.ReplicatorSv1Ping
	case utils.ApierSLow:
		return utils.APIerSv1Ping
	case utils.EEsLow:
		return utils.EeSv1Ping
	default:
	}
	return self.rpcMethod
}

func (self *CmdApierPing) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &StringWrapper{}
	}
	return self.rpcParams
}

func (self *CmdApierPing) PostprocessRpcParams() error {
	if val, can := self.rpcParams.(*StringWrapper); can {
		self.item = val.Item
	}
	self.rpcParams = &utils.CGREvent{}
	return nil
}

func (self *CmdApierPing) RpcResult() any {
	var s string
	return &s
}
