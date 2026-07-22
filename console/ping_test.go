// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"reflect"
	"testing"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"
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
	srv, err := engine.NewService(&v1.RouteSv1{})
	if err != nil {
		t.Fatal(err)
	}
	mType, ok := srv.Methods["Ping"]
	if !ok {
		t.Fatal("method not found")
	}
	m := mType.Method
	if m.Type.NumIn() != 4 { // expecting 4 inputs
		t.Fatalf("invalid number of input parameters ")
	}
	// for coverage purpose
	result := command.RpcParams(true)
	if !reflect.DeepEqual(result, new(StringWrapper)) {
		t.Errorf("Expected <%T>, Received <%T>", new(StringWrapper), result)
	}
	// verify the type of output parameter
	if ok := m.Type.In(3).AssignableTo(reflect.TypeOf(command.RpcResult())); !ok {
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
	srv, err := engine.NewService(&v1.AttributeSv1{})
	if err != nil {
		t.Fatal(err)
	}
	mType, ok := srv.Methods["Ping"]
	if !ok {
		t.Fatal("method not found")
	}
	m := mType.Method
	if m.Type.NumIn() != 4 { // expecting 4 inputs
		t.Fatalf("invalid number of input parameters ")
	}
	// for coverage purpose
	result := command.RpcParams(true)
	if !reflect.DeepEqual(result, new(StringWrapper)) {
		t.Errorf("Expected <%T>, Received <%T>", new(StringWrapper), result)
	}
	// verify the type of output parameter
	if ok := m.Type.In(3).AssignableTo(reflect.TypeOf(command.RpcResult())); !ok {
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
	srv, err := engine.NewService(&v1.ChargerSv1{})
	if err != nil {
		t.Fatal(err)
	}
	mType, ok := srv.Methods["Ping"]
	if !ok {
		t.Fatal("method not found")
	}
	m := mType.Method
	if m.Type.NumIn() != 4 { // expecting 4 inputs
		t.Fatalf("invalid number of input parameters ")
	}
	// for coverage purpose
	result := command.RpcParams(true)
	if !reflect.DeepEqual(result, new(StringWrapper)) {
		t.Errorf("Expected <%T>, Received <%T>", new(StringWrapper), result)
	}
	// verify the type of output parameter
	if ok := m.Type.In(3).AssignableTo(reflect.TypeOf(command.RpcResult())); !ok {
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
	srv, err := engine.NewService(&v1.ResourceSv1{})
	if err != nil {
		t.Fatal(err)
	}
	mType, ok := srv.Methods["Ping"]
	if !ok {
		t.Fatal("method not found")
	}
	m := mType.Method
	if m.Type.NumIn() != 4 { // expecting 4 inputs
		t.Fatalf("invalid number of input parameters ")
	}
	// for coverage purpose
	result := command.RpcParams(true)
	if !reflect.DeepEqual(result, new(StringWrapper)) {
		t.Errorf("Expected <%T>, Received <%T>", new(StringWrapper), result)
	}
	// verify the type of output parameter
	if ok := m.Type.In(3).AssignableTo(reflect.TypeOf(command.RpcResult())); !ok {
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
	srv, err := engine.NewService(&v1.StatSv1{})
	if err != nil {
		t.Fatal(err)
	}
	mType, ok := srv.Methods["Ping"]
	if !ok {
		t.Fatal("method not found")
	}
	m := mType.Method
	if m.Type.NumIn() != 4 { // expecting 4 inputs
		t.Fatalf("invalid number of input parameters ")
	}
	// for coverage purpose
	result := command.RpcParams(true)
	if !reflect.DeepEqual(result, new(StringWrapper)) {
		t.Errorf("Expected <%T>, Received <%T>", new(StringWrapper), result)
	}
	// verify the type of output parameter
	if ok := m.Type.In(3).AssignableTo(reflect.TypeOf(command.RpcResult())); !ok {
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
	srv, err := engine.NewService(&v1.ThresholdSv1{})
	if err != nil {
		t.Fatal(err)
	}
	mType, ok := srv.Methods["Ping"]
	if !ok {
		t.Fatal("method not found")
	}
	m := mType.Method
	if m.Type.NumIn() != 4 { // expecting 4 inputs
		t.Fatalf("invalid number of input parameters ")
	}
	// for coverage purpose
	result := command.RpcParams(true)
	if !reflect.DeepEqual(result, new(StringWrapper)) {
		t.Errorf("Expected <%T>, Received <%T>", new(StringWrapper), result)
	}
	// verify the type of output parameter
	if ok := m.Type.In(3).AssignableTo(reflect.TypeOf(command.RpcResult())); !ok {
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
	srv, err := engine.NewService(&v1.SessionSv1{})
	if err != nil {
		t.Fatal(err)
	}
	mType, ok := srv.Methods["Ping"]
	if !ok {
		t.Fatal("method not found")
	}
	m := mType.Method
	if m.Type.NumIn() != 4 { // expecting 4 inputs
		t.Fatalf("invalid number of input parameters ")
	}
	// for coverage purpose
	result := command.RpcParams(true)
	if !reflect.DeepEqual(result, new(StringWrapper)) {
		t.Errorf("Expected <%T>, Received <%T>", new(StringWrapper), result)
	}
	// verify the type of output parameter
	if ok := m.Type.In(3).AssignableTo(reflect.TypeOf(command.RpcResult())); !ok {
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
	srv, err := engine.NewService(&v1.DispatcherSv1{})
	if err != nil {
		t.Fatal(err)
	}
	mType, ok := srv.Methods["Ping"]
	if !ok {
		t.Fatal("method not found")
	}
	m := mType.Method
	if m.Type.NumIn() != 4 { // expecting 4 inputs
		t.Fatalf("invalid number of input parameters ")
	}
	// for coverage purpose
	result := command.RpcParams(true)
	if !reflect.DeepEqual(result, new(StringWrapper)) {
		t.Errorf("Expected <%T>, Received <%T>", new(StringWrapper), result)
	}
	// verify the type of output parameter
	if ok := m.Type.In(3).AssignableTo(reflect.TypeOf(command.RpcResult())); !ok {
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
	srv, err := engine.NewService(&v1.AnalyzerSv1{})
	if err != nil {
		t.Fatal(err)
	}
	mType, ok := srv.Methods["Ping"]
	if !ok {
		t.Fatal("method not found")
	}
	m := mType.Method
	if m.Type.NumIn() != 4 { // expecting 4 inputs
		t.Fatalf("invalid number of input parameters ")
	}
	// for coverage purpose
	result := command.RpcParams(true)
	if !reflect.DeepEqual(result, new(StringWrapper)) {
		t.Errorf("Expected <%T>, Received <%T>", new(StringWrapper), result)
	}
	// verify the type of output parameter
	if ok := m.Type.In(3).AssignableTo(reflect.TypeOf(command.RpcResult())); !ok {
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
	srv, err := engine.NewService(&v1.SchedulerSv1{})
	if err != nil {
		t.Fatal(err)
	}
	mType, ok := srv.Methods["Ping"]
	if !ok {
		t.Fatal("method not found")
	}
	m := mType.Method
	if m.Type.NumIn() != 4 { // expecting 4 inputs
		t.Fatalf("invalid number of input parameters ")
	}
	// for coverage purpose
	result := command.RpcParams(true)
	if !reflect.DeepEqual(result, new(StringWrapper)) {
		t.Errorf("Expected <%T>, Received <%T>", new(StringWrapper), result)
	}
	// verify the type of output parameter
	if ok := m.Type.In(3).AssignableTo(reflect.TypeOf(command.RpcResult())); !ok {
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
	srv, err := engine.NewService(&v1.RALsV1{})
	if err != nil {
		t.Fatal(err)
	}
	mType, ok := srv.Methods["Ping"]
	if !ok {
		t.Fatal("method not found")
	}
	m := mType.Method
	if m.Type.NumIn() != 4 { // expecting 4 inputs
		t.Fatalf("invalid number of input parameters ")
	}
	// for coverage purpose
	result := command.RpcParams(true)
	if !reflect.DeepEqual(result, new(StringWrapper)) {
		t.Errorf("Expected <%T>, Received <%T>", new(StringWrapper), result)
	}
	// verify the type of output parameter
	if ok := m.Type.In(3).AssignableTo(reflect.TypeOf(command.RpcResult())); !ok {
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
	srv, err := engine.NewService(&v1.ReplicatorSv1{})
	if err != nil {
		t.Fatal(err)
	}
	mType, ok := srv.Methods["Ping"]
	if !ok {
		t.Fatal("method not found")
	}
	m := mType.Method
	if m.Type.NumIn() != 4 { // expecting 4 inputs
		t.Fatalf("invalid number of input parameters ")
	}
	// for coverage purpose
	result := command.RpcParams(true)
	if !reflect.DeepEqual(result, new(StringWrapper)) {
		t.Errorf("Expected <%T>, Received <%T>", new(StringWrapper), result)
	}
	// verify the type of output parameter
	if ok := m.Type.In(3).AssignableTo(reflect.TypeOf(command.RpcResult())); !ok {
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
	srv, err := engine.NewService(&v1.APIerSv1{})
	if err != nil {
		t.Fatal(err)
	}
	mType, ok := srv.Methods["Ping"]
	if !ok {
		t.Fatal("method not found")
	}
	m := mType.Method
	if m.Type.NumIn() != 4 { // expecting 4 inputs
		t.Fatalf("invalid number of input parameters ")
	}
	// for coverage purpose
	result := command.RpcParams(true)
	if !reflect.DeepEqual(result, new(StringWrapper)) {
		t.Errorf("Expected <%T>, Received <%T>", new(StringWrapper), result)
	}
	// verify the type of output parameter
	if ok := m.Type.In(3).AssignableTo(reflect.TypeOf(command.RpcResult())); !ok {
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
	srv, err := engine.NewService(&v1.EeSv1{})
	if err != nil {
		t.Fatal(err)
	}
	mType, ok := srv.Methods["Ping"]
	if !ok {
		t.Fatal("method not found")
	}
	m := mType.Method
	if m.Type.NumIn() != 4 { // expecting 4 inputs
		t.Fatalf("invalid number of input parameters ")
	}
	// for coverage purpose
	result := command.RpcParams(true)
	if !reflect.DeepEqual(result, new(StringWrapper)) {
		t.Errorf("Expected <%T>, Received <%T>", new(StringWrapper), result)
	}
	// verify the type of output parameter
	if ok := m.Type.In(3).AssignableTo(reflect.TypeOf(command.RpcResult())); !ok {
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
