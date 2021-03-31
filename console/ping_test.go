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
	m, ok := reflect.TypeOf(new(v1.ChargerSv1)).MethodByName(strings.Split(command.RpcMethod(), utils.NestingSep)[1])
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

func TestCmdPingResourcesLow(t *testing.T) {
	// commands map is initiated in init function
	command := commands["ping"]
	castCommand, canCast := command.(*CmdApierPing)
	if !canCast {
		t.Fatalf("cannot cast")
	}
	castCommand.item = utils.ResourcesLow
	result2 := command.RpcMethod()
	if !reflect.DeepEqual(result2, utils.ResourceSv1Ping) {
		t.Errorf("Expected <%+v>, Received <%+v>", utils.ResourceSv1Ping, result2)
	}
	m, ok := reflect.TypeOf(new(v1.ResourceSv1)).MethodByName(strings.Split(command.RpcMethod(), utils.NestingSep)[1])
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

func TestCmdPingStatServiceLow(t *testing.T) {
	// commands map is initiated in init function
	command := commands["ping"]
	castCommand, canCast := command.(*CmdApierPing)
	if !canCast {
		t.Fatalf("cannot cast")
	}
	castCommand.item = utils.StatServiceLow
	result2 := command.RpcMethod()
	if !reflect.DeepEqual(result2, utils.StatSv1Ping) {
		t.Errorf("Expected <%+v>, Received <%+v>", utils.StatSv1Ping, result2)
	}
	m, ok := reflect.TypeOf(new(v1.StatSv1)).MethodByName(strings.Split(command.RpcMethod(), utils.NestingSep)[1])
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

func TestCmdPingThresholdsLow(t *testing.T) {
	// commands map is initiated in init function
	command := commands["ping"]
	castCommand, canCast := command.(*CmdApierPing)
	if !canCast {
		t.Fatalf("cannot cast")
	}
	castCommand.item = utils.ThresholdsLow
	result2 := command.RpcMethod()
	if !reflect.DeepEqual(result2, utils.ThresholdSv1Ping) {
		t.Errorf("Expected <%+v>, Received <%+v>", utils.ThresholdSv1Ping, result2)
	}
	m, ok := reflect.TypeOf(new(v1.ThresholdSv1)).MethodByName(strings.Split(command.RpcMethod(), utils.NestingSep)[1])
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

func TestCmdPingSessionsLow(t *testing.T) {
	// commands map is initiated in init function
	command := commands["ping"]
	castCommand, canCast := command.(*CmdApierPing)
	if !canCast {
		t.Fatalf("cannot cast")
	}
	castCommand.item = utils.SessionsLow
	result2 := command.RpcMethod()
	if !reflect.DeepEqual(result2, utils.SessionSv1Ping) {
		t.Errorf("Expected <%+v>, Received <%+v>", utils.SessionSv1Ping, result2)
	}
	m, ok := reflect.TypeOf(new(v1.SessionSv1)).MethodByName(strings.Split(command.RpcMethod(), utils.NestingSep)[1])
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

func TestCmdPingLoaderSLow(t *testing.T) {
	// commands map is initiated in init function
	command := commands["ping"]
	castCommand, canCast := command.(*CmdApierPing)
	if !canCast {
		t.Fatalf("cannot cast")
	}
	castCommand.item = utils.LoaderSLow
	result2 := command.RpcMethod()
	if !reflect.DeepEqual(result2, utils.LoaderSv1Ping) {
		t.Errorf("Expected <%+v>, Received <%+v>", utils.LoaderSv1Ping, result2)
	}
	m, ok := reflect.TypeOf(new(v1.LoaderSv1)).MethodByName(strings.Split(command.RpcMethod(), utils.NestingSep)[1])
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

func TestCmdPingDispatcherSLow(t *testing.T) {
	// commands map is initiated in init function
	command := commands["ping"]
	castCommand, canCast := command.(*CmdApierPing)
	if !canCast {
		t.Fatalf("cannot cast")
	}
	castCommand.item = utils.DispatcherSLow
	result2 := command.RpcMethod()
	if !reflect.DeepEqual(result2, utils.DispatcherSv1Ping) {
		t.Errorf("Expected <%+v>, Received <%+v>", utils.DispatcherSv1Ping, result2)
	}
	m, ok := reflect.TypeOf(new(v1.DispatcherSv1)).MethodByName(strings.Split(command.RpcMethod(), utils.NestingSep)[1])
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

func TestCmdPingAnalyzerSLow(t *testing.T) {
	// commands map is initiated in init function
	command := commands["ping"]
	castCommand, canCast := command.(*CmdApierPing)
	if !canCast {
		t.Fatalf("cannot cast")
	}
	castCommand.item = utils.AnalyzerSLow
	result2 := command.RpcMethod()
	if !reflect.DeepEqual(result2, utils.AnalyzerSv1Ping) {
		t.Errorf("Expected <%+v>, Received <%+v>", utils.AnalyzerSv1Ping, result2)
	}
	m, ok := reflect.TypeOf(new(v1.AnalyzerSv1)).MethodByName(strings.Split(command.RpcMethod(), utils.NestingSep)[1])
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

func TestCmdPingSchedulerSLow(t *testing.T) {
	// commands map is initiated in init function
	command := commands["ping"]
	castCommand, canCast := command.(*CmdApierPing)
	if !canCast {
		t.Fatalf("cannot cast")
	}
	castCommand.item = utils.SchedulerSLow
	result2 := command.RpcMethod()
	if !reflect.DeepEqual(result2, utils.SchedulerSv1Ping) {
		t.Errorf("Expected <%+v>, Received <%+v>", utils.SchedulerSv1Ping, result2)
	}
	m, ok := reflect.TypeOf(new(v1.SchedulerSv1)).MethodByName(strings.Split(command.RpcMethod(), utils.NestingSep)[1])
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

func TestCmdPingRALsLow(t *testing.T) {
	// commands map is initiated in init function
	command := commands["ping"]
	castCommand, canCast := command.(*CmdApierPing)
	if !canCast {
		t.Fatalf("cannot cast")
	}
	castCommand.item = utils.RALsLow
	result2 := command.RpcMethod()
	if !reflect.DeepEqual(result2, utils.RALsV1Ping) {
		t.Errorf("Expected <%+v>, Received <%+v>", utils.RALsV1Ping, result2)
	}
	m, ok := reflect.TypeOf(new(v1.RALsV1)).MethodByName(strings.Split(command.RpcMethod(), utils.NestingSep)[1])
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

func TestCmdPingReplicatorLow(t *testing.T) {
	// commands map is initiated in init function
	command := commands["ping"]
	castCommand, canCast := command.(*CmdApierPing)
	if !canCast {
		t.Fatalf("cannot cast")
	}
	castCommand.item = utils.ReplicatorLow
	result2 := command.RpcMethod()
	if !reflect.DeepEqual(result2, utils.ReplicatorSv1Ping) {
		t.Errorf("Expected <%+v>, Received <%+v>", utils.RALsV1Ping, result2)
	}
	m, ok := reflect.TypeOf(new(v1.ReplicatorSv1)).MethodByName(strings.Split(command.RpcMethod(), utils.NestingSep)[1])
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

func TestCmdPingApierSLow(t *testing.T) {
	// commands map is initiated in init function
	command := commands["ping"]
	castCommand, canCast := command.(*CmdApierPing)
	if !canCast {
		t.Fatalf("cannot cast")
	}
	castCommand.item = utils.ApierSLow
	result2 := command.RpcMethod()
	if !reflect.DeepEqual(result2, utils.APIerSv1Ping) {
		t.Errorf("Expected <%+v>, Received <%+v>", utils.APIerSv1Ping, result2)
	}
	m, ok := reflect.TypeOf(new(v1.APIerSv1)).MethodByName(strings.Split(command.RpcMethod(), utils.NestingSep)[1])
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

func TestCmdPingEEsLow(t *testing.T) {
	// commands map is initiated in init function
	command := commands["ping"]
	castCommand, canCast := command.(*CmdApierPing)
	if !canCast {
		t.Fatalf("cannot cast")
	}
	castCommand.item = utils.EEsLow
	result2 := command.RpcMethod()
	if !reflect.DeepEqual(result2, utils.EeSv1Ping) {
		t.Errorf("Expected <%+v>, Received <%+v>", utils.EeSv1Ping, result2)
	}
	m, ok := reflect.TypeOf(new(v1.EeSv1)).MethodByName(strings.Split(command.RpcMethod(), utils.NestingSep)[1])
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

func TestCmdPingTestDefault(t *testing.T) {
	// commands map is initiated in init function
	command := commands["ping"]
	castCommand, canCast := command.(*CmdApierPing)
	if !canCast {
		t.Fatalf("cannot cast")
	}
	castCommand.item = "test_item"
	result2 := command.RpcMethod()
	if !reflect.DeepEqual(result2, "") {
		t.Errorf("Expected <%+v>, Received <%+v>", "", result2)
	}
}
