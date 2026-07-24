// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdExportCdrs{
		name:      "cdrs_export",
		rpcMethod: utils.APIerSv1ExportCDRs,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdExportCdrs struct {
	name      string
	rpcMethod string
	rpcParams *v1.ArgExportCDRs
	*CommandExecuter
}

func (self *CmdExportCdrs) Name() string {
	return self.name
}

func (self *CmdExportCdrs) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdExportCdrs) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(v1.ArgExportCDRs)
	}
	return self.rpcParams
}

func (self *CmdExportCdrs) PostprocessRpcParams() error {
	return nil
}

func (self *CmdExportCdrs) RpcResult() any {
	var reply *v1.RplExportedCDRs
	return &reply
}
