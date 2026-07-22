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

func TestCmdSharedGroup(t *testing.T) {
	// commands map is initiated in init function
	command := commands["sharedgroup"]
	// verify if ApierSv1 object has method on it
	m, ok := reflect.TypeOf(new(v1.APIerSv1)).MethodByName(strings.Split(command.RpcMethod(), utils.NestingSep)[1])
	if !ok {
		t.Fatal("method not found")
	}
	if m.Type.NumIn() != 4 { // expecting 4 inputs
		t.Fatalf("invalid number of input parameters ")
	}
	// for coverage purpose
	result := command.RpcParams(false)
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
