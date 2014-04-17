/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package console

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/cgrates/cgrates/apier"
	"github.com/cgrates/cgrates/engine"
)

func init() {
	commands["add_balance"] = &CmdAddBalance{
		rpcMethod: "ApierV1.AddBalance",
		rpcParams: &apier.AttrAddBalance{BalanceType: engine.CREDIT},
	}
}

// Commander implementation
type CmdAddBalance struct {
	rpcMethod string
	rpcParams *apier.AttrAddBalance
	rpcResult string
}

func (self *CmdAddBalance) Usage() string {
	jsn, _ := json.Marshal(apier.AttrAddBalance{Direction: "*out"})
	return "\n\tUsage: add_balance " + FromJSON(jsn, self.ClientArgs()) + "\n"
}

// Parses command line args and builds CmdBalance value
func (self *CmdAddBalance) FromArgs(args string, verbose bool) error {
	if len(args) == 0 {
		return fmt.Errorf(self.Usage())
	}
	// defaults
	self.rpcParams = &apier.AttrAddBalance{Direction: "*out"}

	if err := json.Unmarshal(ToJSON(args), &self.rpcParams); err != nil {
		return err
	}
	if verbose {
		jsn, _ := json.Marshal(self.rpcParams)
		fmt.Println("add_balance ", FromJSON(jsn, self.ClientArgs()))
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

func (self *CmdAddBalance) ClientArgs() (args []string) {
	val := reflect.ValueOf(&apier.AttrAddBalance{}).Elem()

	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		args = append(args, typeField.Name)
	}
	return
}
