package console

import (
	"fmt"
	"github.com/cgrates/cgrates/apier/v1"
)

func init() {
	commands["add_account"] = &CmdAddAccount{}
}

// Commander implementation
type CmdAddAccount struct {
	rpcMethod string
	rpcParams *apier.AttrAddAccount
	rpcResult string
}

// name should be exec's name
func (self *CmdAddAccount) Usage(name string) string {
	return fmt.Sprintf("\n\tUsage: cgr-console [cfg_opts...{-h}] add_account <tenant> <account> <type=prepaid|postpaid> <actiontimingsid> [<direction>]")
}

// set param defaults
func (self *CmdAddAccount) defaults() error {
	self.rpcMethod = "ApierV1.AddAccount"
	self.rpcParams = &apier.AttrAddAccount{Direction: "*out"}
	return nil
}

// Parses command line args and builds CmdBalance value
func (self *CmdAddAccount) FromArgs(args []string) error {
	if len(args) < 6 {
		return fmt.Errorf(self.Usage(""))
	}
	// Args look OK, set defaults before going further
	self.defaults()
	self.rpcParams.Tenant = args[2]
	self.rpcParams.Account = args[3]
	self.rpcParams.Type = args[4]
	self.rpcParams.ActionTimingsId = args[5]
	if len(args) > 6 {
		self.rpcParams.Direction = args[6]
	}
	return nil
}

func (self *CmdAddAccount) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdAddAccount) RpcParams() interface{} {
	return self.rpcParams
}

func (self *CmdAddAccount) RpcResult() interface{} {
	return &self.rpcResult
}
