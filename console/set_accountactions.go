package console

import (
	"fmt"
	"github.com/cgrates/cgrates/apier/v1"
)

func init() {
	commands["set_accountactions"] = &CmdSetAccountActions{}
}

// Commander implementation
type CmdSetAccountActions struct {
	rpcMethod string
	rpcParams *apier.AttrSetAccountActions
	rpcResult string
}

// name should be exec's name
func (self *CmdSetAccountActions) Usage(name string) string {
	return fmt.Sprintf("\n\tUsage: cgr-console [cfg_opts...{-h}] set_accountactions <tpid> <accountactionsid>")
}

// set param defaults
func (self *CmdSetAccountActions) defaults() error {
	self.rpcMethod = "ApierV1.SetAccountActions"
	self.rpcParams = &apier.AttrSetAccountActions{}
	return nil
}

// Parses command line args and builds CmdBalance value
func (self *CmdSetAccountActions) FromArgs(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf(self.Usage(""))
	}
	// Args look OK, set defaults before going further
	self.defaults()
	self.rpcParams.TPid = args[2]
	self.rpcParams.AccountActionsId = args[3]
	return nil
}

func (self *CmdSetAccountActions) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetAccountActions) RpcParams() interface{} {
	return self.rpcParams
}

func (self *CmdSetAccountActions) RpcResult() interface{} {
	return &self.rpcResult
}
