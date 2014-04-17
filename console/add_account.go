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
	commands["add_account"] = &CmdAddAccount{
		rpcMethod: "ApierV1.SetAccount",
	}
}

// Commander implementation
type CmdAddAccount struct {
	rpcMethod string
	rpcParams *apier.AttrSetAccount
	rpcResult string
}

func (self *CmdAddAccount) Usage() string {
	jsn, _ := json.Marshal(engine.CallDescriptor{Direction: "*out"})
	return "\n\tUsage: add_account " + FromJSON(jsn, self.ClientArgs()) + "\n"
}

// Parses command line args and builds CmdBalance value
func (self *CmdAddAccount) FromArgs(args string, verbose bool) error {
	if len(args) == 0 {
		return fmt.Errorf(self.Usage())
	}
	// defaults
	self.rpcParams = &apier.AttrSetAccount{Direction: "*out"}

	if err := json.Unmarshal(ToJSON(args), &self.rpcParams); err != nil {
		return err
	}
	if verbose {
		jsn, _ := json.Marshal(self.rpcParams)
		fmt.Println("add_account ", FromJSON(jsn, self.ClientArgs()))
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

func (self *CmdAddAccount) ClientArgs() (args []string) {
	val := reflect.ValueOf(apier.AttrSetAccount{}).Elem()

	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		args = append(args, typeField.Name)
	}
	return
}
