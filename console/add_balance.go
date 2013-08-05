package console

import (
	"fmt"
	"github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"
	"strconv"
)

func init() {
	commands["add_balance"] = &CmdAddBalance{}
}

// Commander implementation
type CmdAddBalance struct {
	rpcMethod string
	rpcParams *apier.AttrAddBalance
	rpcResult float64
}

// name should be exec's name
func (self *CmdAddBalance) Usage(name string) string {
	return fmt.Sprintf("\n\tUsage: cgr-console [cfg_opts...{-h}] get_balance <tenant> <account> <value> [<balanceid=monetary|sms|internet|internet_time|minutes> [<direction>]]")
}

// set param defaults
func (self *CmdAddBalance) defaults() error {
	self.rpcMethod = "ApierV1.AddBalance"
	self.rpcParams = &apier.AttrAddBalance{BalanceId: engine.CREDIT}
	self.rpcParams.Direction = "*out"
	return nil
}

// Parses command line args and builds CmdBalance value
func (self *CmdAddBalance) FromArgs(args []string) error {
	if len(args) < 5 {
		return fmt.Errorf(self.Usage(""))
	}
	// Args look OK, set defaults before going further
	self.defaults()
	self.rpcParams.Tenant = args[2]
	self.rpcParams.Account = args[3]
	value, err := strconv.ParseFloat(args[4], 64)
	if err != nil {
		return err
	}
	self.rpcParams.Value = value
	if len(args) > 5 {
		self.rpcParams.BalanceId = args[5]
	}
	if len(args) > 6 {
		self.rpcParams.Direction = args[6]
	}
	return nil
}

func (self *CmdAddBalance) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdAddBalance) RpcParams() interface{} {
	return self.rpcParams
}

func (self *CmdAddBalance) RpcResult() interface{} {
	return &self.rpcResult
}
