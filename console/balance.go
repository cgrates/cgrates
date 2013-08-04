/* Implementing balance related console commands.
 */
package console

import (
	"fmt"
	"github.com/cgrates/cgrates/engine"
)

func init() {
	commands["get_balance"] = &CmdGetBalance{}
}

// Commander implementation
type CmdGetBalance struct {
	rpcMethod string
	rpcParams *engine.CallDescriptor
	rpcResult *engine.CallCost
}

// name should be exec's name
func (self *CmdGetBalance) Usage(name string) string {
	return fmt.Sprintf("\n\tUsage: cgr-console [cfg_opts...{-h}] get_balance <tenant> <user> [<balanceid> [<direction>]]")
}

// set param defaults
func (self *CmdGetBalance) defaults() error {
	self.rpcMethod = "Responder.GetMonetary"
	self.rpcParams = &engine.CallDescriptor{Direction: "*out"}	
	return nil
}

// Parses command line args and builds CmdBalance value
func (self *CmdGetBalance) FromArgs(args []string) error {
	if len(args) < 4 {
		return fmt.Errorf(self.Usage(""))
	}
	// Args look OK, set defaults before going further	
	self.defaults()
	self.rpcParams.Tenant = args[2]
	self.rpcParams.Account = args[3]
	if len(args) > 4 {
		switch args[4] {
		case "MONETARY":
			self.rpcMethod = "Responder.GetMonetary"
		case "SMS":
			self.rpcMethod = "Responder.GetSMS"
		case "INETRNET":
			self.rpcMethod = "Responder.GetInternet"
		case "INTERNET_TIME":
			self.rpcMethod = "Responder.GetInternetTime"
		case "MINUTES":
			self.rpcMethod = "Responder.GetMonetary"
		}
	}
	if len(args) > 5 {
		self.rpcParams.Direction = args[5]
	}
	return nil
}

func (self *CmdGetBalance) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetBalance) RpcParams() interface{} {
	return self.rpcParams
}

func (self *CmdGetBalance) RpcResult() interface{} {
	self.rpcResult = &engine.CallCost{}
	return self.rpcResult
}
