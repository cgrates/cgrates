// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestCmdParse(t *testing.T) {
	// commands map is initiated in init function
	command := commands["parse"]

	// for coverage purpose
	expected := command.Name()
	if !reflect.DeepEqual(expected, "parse") {
		t.Errorf("Expected <%+v>, Received <%+v>", "parse", expected)
	}

	// for coverage purpose
	expected = command.RpcMethod()
	if !reflect.DeepEqual(expected, utils.EmptyString) {
		t.Errorf("Expected <%+v>, Received <%+v>", utils.EmptyString, expected)
	}

	// for coverage purpose
	expected2 := command.RpcParams(true)
	if !reflect.DeepEqual(expected2, &AttrParse{}) {
		t.Errorf("Expected <%+v>, Received <%+v>", &AttrParse{}, expected2)
	}

	// for coverage purpose
	if err := command.RpcResult(); err != nil {
		t.Fatal(err)
	}

	// for coverage purpose
	if err := command.PostprocessRpcParams(); err != nil {
		t.Fatal(err)
	}

}
func TestCmdParseLocalExecuteCase1(t *testing.T) {
	// for coverage purpose
	testStruct := &CmdParse{
		rpcParams: &AttrParse{
			Expression: "",
			Value:      "",
		},
	}

	result := testStruct.LocalExecute()
	expected := "Empty expression error"
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected <%+v>, Received <%+v>", expected, result)
	}

}
func TestCmdParseLocalExecuteCase2(t *testing.T) {
	// for coverage purpose
	testStruct := &CmdParse{
		rpcParams: &AttrParse{
			Expression: "test_exp",
			Value:      "",
		},
	}

	result := testStruct.LocalExecute()
	expected := "Empty value error"
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected <%+v>, Received <%+v>", expected, result)
	}
}

func TestCmdParseLocalExecuteCase3(t *testing.T) {
	// for coverage purpose
	testStruct := &CmdParse{
		rpcParams: &AttrParse{
			Expression: "test_exp",
			Value:      "test_val",
		},
	}

	result := testStruct.LocalExecute()
	expected := "test_exp"
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected <%+v>, Received <%+v>", expected, result)
	}
}

func TestCmdParseLocalExecuteCase4(t *testing.T) {
	// for coverage purpose
	testStruct := &CmdParse{
		rpcParams: &AttrParse{
			Expression: "~*req.Field{*}",
			Value:      "~*req.Field{*}",
		},
	}
	err := testStruct.LocalExecute()
	expected := "invalid converter value in string: <*>, err: unsupported converter definition: <*>"
	if !reflect.DeepEqual(err, expected) {
		t.Errorf("Expected <%+v>, Received <%+v>", expected, err)
	}
}

func TestCmdParseLocalExecuteCase5(t *testing.T) {
	// for coverage purpose
	testStruct := &CmdParse{
		rpcParams: &AttrParse{
			Expression: "~*req.Field{*duration}",
			Value:      "a",
		},
	}
	expected := "time: invalid duration \"a\""
	received := testStruct.LocalExecute()
	if !reflect.DeepEqual(received, expected) {
		t.Errorf("Expected <%+v>, Received <%+v>", expected, received)
	}
}
