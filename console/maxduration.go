// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetMaxDuration{
		name:       "maxduration",
		rpcMethod:  utils.ResponderGetMaxSessionTime,
		clientArgs: []string{"Category", "ToR", "Tenant", "Subject", "Account", "Destination", "TimeStart", "TimeEnd", "CallDuration", "FallbackSubject"},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetMaxDuration struct {
	name       string
	rpcMethod  string
	rpcParams  *engine.CallDescriptorWithArgDispatcher
	clientArgs []string
	*CommandExecuter
}

func (self *CmdGetMaxDuration) Name() string {
	return self.name
}

func (self *CmdGetMaxDuration) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetMaxDuration) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.CallDescriptorWithArgDispatcher{
			CallDescriptor: new(engine.CallDescriptor),
			ArgDispatcher:  new(utils.ArgDispatcher),
		}
	}
	return self.rpcParams
}

func (self *CmdGetMaxDuration) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetMaxDuration) RpcResult() any {
	var d time.Duration
	return &d
}

func (self *CmdGetMaxDuration) ClientArgs() []string {
	return self.clientArgs
}

func (self *CmdGetMaxDuration) GetFormatedResult(result any) string {
	if tv, canCast := result.(*time.Duration); canCast {
		return fmt.Sprintf(`"%s"`, tv.String())
	}
	out, _ := json.MarshalIndent(result, "", " ")
	return string(out)
}
