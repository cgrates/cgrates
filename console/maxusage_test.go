// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"reflect"
	"strings"
	"testing"

	v1 "github.com/cgrates/cgrates/apier/v1"

	"github.com/cgrates/cgrates/utils"
)

func TestCmdMaxUsage(t *testing.T) {
	// commands map is initiated in init function
	command := commands["maxusage"]
	// verify if ApierSv1 object has method on it
	m, ok := reflect.TypeOf(new(v1.APIerSv1)).MethodByName(strings.Split(command.RpcMethod(), utils.NestingSep)[1])
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
	// for coverage purpose
	formatedResult := command.GetFormatedResult(command.RpcResult())

	if formatedResult != `"0s"` {
		t.Errorf("Expected 0s, Received %s", formatedResult)
	}
	expected2 := []string{utils.ToR, utils.RequestType, utils.Tenant,
		utils.Category, utils.AccountField, utils.Subject, utils.Destination,
		utils.SetupTime, utils.AnswerTime, utils.Usage, utils.ExtraFields}

	if !reflect.DeepEqual(command.ClientArgs(), expected2) {
		t.Errorf("Expected <%+v>, Received <%+v>", expected2, command.ClientArgs())
	}

}
