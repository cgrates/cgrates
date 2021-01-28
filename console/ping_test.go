/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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
	"reflect"
	"strings"
	"testing"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/utils"
)

func TestCmdPingRoutesLow(t *testing.T) {
	// commands map is initiated in init function
	command := commands["ping"]
	castCommand, canCast := command.(*CmdApierPing)
	if !canCast {
		t.Fatalf("cannot cast")
	}
	castCommand.item = utils.RoutesLow
	result2 := command.RpcMethod()
	if !reflect.DeepEqual(result2, utils.RouteSv1Ping) {
		t.Errorf("Expected <%+v>, Received <%+v>", utils.RouteSv1Ping, result2)
	}
	m, ok := reflect.TypeOf(new(v1.RouteSv1)).MethodByName(strings.Split(command.RpcMethod(), utils.NestingSep)[1])
	if !ok {
		t.Fatal("method not found")
	}
	if m.Type.NumIn() != 3 { // ApierSv1 is consider and we expect 3 inputs
		t.Fatalf("invalid number of input parameters ")
	}
	// for coverage purpose
	result := command.RpcParams(true)
	if !reflect.DeepEqual(result, new(StringWrapper)) {
		t.Errorf("Expected <%T>, Received <%T>", new(StringWrapper), result)
	}
	// verify the type of output parameter
	if ok := m.Type.In(2).AssignableTo(reflect.TypeOf(command.RpcResult())); !ok {
		t.Fatalf("cannot assign output parameter")
	}
	// for coverage purpose
	if err := command.PostprocessRpcParams(); err != nil {
		t.Fatal(err)
	}
}

func TestCmdPingAttributesLow(t *testing.T) {
	// commands map is initiated in init function
	command := commands["ping"]
	castCommand, canCast := command.(*CmdApierPing)
	if !canCast {
		t.Fatalf("cannot cast")
	}
	castCommand.item = utils.AttributesLow
	result2 := command.RpcMethod()
	if !reflect.DeepEqual(result2, utils.AttributeSv1Ping) {
		t.Errorf("Expected <%+v>, Received <%+v>", utils.AttributeSv1Ping, result2)
	}
	m, ok := reflect.TypeOf(new(v1.AttributeSv1)).MethodByName(strings.Split(command.RpcMethod(), utils.NestingSep)[1])
	if !ok {
		t.Fatal("method not found")
	}
	if m.Type.NumIn() != 3 { // ApierSv1 is consider and we expect 3 inputs
		t.Fatalf("invalid number of input parameters ")
	}
	// for coverage purpose
	result := command.RpcParams(true)
	if !reflect.DeepEqual(result, new(StringWrapper)) {
		t.Errorf("Expected <%T>, Received <%T>", new(StringWrapper), result)
	}
	// verify the type of output parameter
	if ok := m.Type.In(2).AssignableTo(reflect.TypeOf(command.RpcResult())); !ok {
		t.Fatalf("cannot assign output parameter")
	}
	// for coverage purpose
	if err := command.PostprocessRpcParams(); err != nil {
		t.Fatal(err)
	}

}

func TestCmdPingChargerSLow(t *testing.T) {
	// commands map is initiated in init function
	command := commands["ping"]
	castCommand, canCast := command.(*CmdApierPing)
	if !canCast {
		t.Fatalf("cannot cast")
	}
	castCommand.item = utils.ChargerSLow
	result2 := command.RpcMethod()
	if !reflect.DeepEqual(result2, utils.ChargerSv1Ping) {
		t.Errorf("Expected <%+v>, Received <%+v>", utils.ChargerSv1Ping, result2)
	}
	m, ok := reflect.TypeOf(new(v1.AttributeSv1)).MethodByName(strings.Split(command.RpcMethod(), utils.NestingSep)[1])
	if !ok {
		t.Fatal("method not found")
	}
	if m.Type.NumIn() != 3 { // ApierSv1 is consider and we expect 3 inputs
		t.Fatalf("invalid number of input parameters ")
	}
	// for coverage purpose
	result := command.RpcParams(true)
	if !reflect.DeepEqual(result, new(StringWrapper)) {
		t.Errorf("Expected <%T>, Received <%T>", new(StringWrapper), result)
	}
	// verify the type of output parameter
	if ok := m.Type.In(2).AssignableTo(reflect.TypeOf(command.RpcResult())); !ok {
		t.Fatalf("cannot assign output parameter")
	}
	// for coverage purpose
	if err := command.PostprocessRpcParams(); err != nil {
		t.Fatal(err)
	}

}
