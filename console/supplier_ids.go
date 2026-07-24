// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSuppliersIDs{
		name:      "supplier_ids",
		rpcMethod: utils.APIerSv1GetSupplierProfileIDs,
		rpcParams: &utils.TenantArgWithPaginator{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdSuppliersIDs struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantArgWithPaginator
	*CommandExecuter
}

func (self *CmdSuppliersIDs) Name() string {
	return self.name
}

func (self *CmdSuppliersIDs) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSuppliersIDs) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantArgWithPaginator{}
	}
	return self.rpcParams
}

func (self *CmdSuppliersIDs) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSuppliersIDs) RpcResult() any {
	var atr []string
	return &atr
}
