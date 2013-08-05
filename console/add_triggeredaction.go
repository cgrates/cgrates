package console

import (
	"fmt"
	"github.com/cgrates/cgrates/apier/v1"
	"strconv"
)

func init() {
	commands["add_triggeredaction"] = &CmdAddTriggeredAction{}
}

// Commander implementation
type CmdAddTriggeredAction struct {
	rpcMethod string
	rpcParams *apier.AttrAddActionTrigger
	rpcResult string
}

// name should be exec's name
func (self *CmdAddTriggeredAction) Usage(name string) string {
	return fmt.Sprintf("\n\tUsage: cgr-console [cfg_opts...{-h}] add_triggeredaction <tenant> <account> <balanceid> <thresholdvalue> <destinationid> <weight> <actionsid> [<direction>]")
}

// set param defaults
func (self *CmdAddTriggeredAction) defaults() error {
	self.rpcMethod = "ApierV1.AddTriggeredAction"
	self.rpcParams = &apier.AttrAddActionTrigger{Direction: "*out"}
	return nil
}

// Parses command line args and builds CmdBalance value
func (self *CmdAddTriggeredAction) FromArgs(args []string) error {
	if len(args) < 9 {
		return fmt.Errorf(self.Usage(""))
	}
	// Args look OK, set defaults before going further
	self.defaults()
	self.rpcParams.Tenant = args[2]
	self.rpcParams.Account = args[3]
	self.rpcParams.BalanceId = args[4]
	thresholdvalue, err := strconv.ParseFloat(args[5], 64)
	if err != nil {
		return err
	}
	self.rpcParams.ThresholdValue = thresholdvalue
	self.rpcParams.DestinationId = args[6]
	weight, err := strconv.ParseFloat(args[7], 64)
	if err != nil {
		return err
	}
	self.rpcParams.Weight = weight
	self.rpcParams.ActionsId = args[8]

	if len(args) > 9 {
		self.rpcParams.Direction = args[9]
	}
	return nil
}

func (self *CmdAddTriggeredAction) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdAddTriggeredAction) RpcParams() interface{} {
	return self.rpcParams
}

func (self *CmdAddTriggeredAction) RpcResult() interface{} {
	return &self.rpcResult
}
