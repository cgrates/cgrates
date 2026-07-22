// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"reflect"
	"strings"
	"testing"

	v2 "github.com/cgrates/cgrates/apier/v2"

	"github.com/cgrates/cgrates/utils"
)

func TestCmdActionPlanSet(t *testing.T) {
	// commands map is initiated in init function
	command := commands["actionplan_set"]
	// verify if ApierSv1 object has method on it
	m, ok := reflect.TypeOf(new(v2.APIerSv2)).MethodByName(strings.Split(command.RpcMethod(), utils.NestingSep)[1])
	if !ok {
		t.Fatal("method not found")
	}
	if m.Type.NumIn() != 4 { // expecting 4 inputs
		t.Fatalf("invalid number of input parameters ")
	}
	// verify the type of input parameter
	if ok := m.Type.In(2).AssignableTo(reflect.TypeOf(command.RpcParams(true))); !ok {
		t.Fatalf("cannot assign input parameter")
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
