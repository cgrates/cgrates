// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/mitchellh/mapstructure"
)

type RpcStruct struct{}

type Attr struct {
	Name    string
	Surname string
	Age     float64
}

func (rpc *RpcStruct) Method1(ctx *context.Context, normal Attr, out *float64) error {
	return nil
}

func (rpc *RpcStruct) Method2(ctx *context.Context, pointer *Attr, out *float64) error {
	return nil
}

func TestRPCObjectPointer(t *testing.T) {
	RegisterRpcParams("", &RpcStruct{})
	if len(rpcParamsMap) != 2 {
		t.Errorf("error registering rpc object: %v", rpcParamsMap)
	}
	x, found := rpcParamsMap["RpcStruct.Method1"]
	if !found {
		t.Errorf("error getting rpcobject: %v (%+v)", rpcParamsMap, x)
	}
	a := x.InParam
	if err := mapstructure.Decode(map[string]any{"Name": "a", "Surname": "b", "Age": 10.2}, &a); err != nil || a.(Attr).Name != "a" || a.(Attr).Surname != "b" || a.(Attr).Age != 10.2 {
		t.Errorf("error converting to struct: %+v (%v)", a, err)
	}
}

func TestGetRpcParamsError(t *testing.T) {
	_, err := GetRpcParams("exampleTest")
	if err == nil || err.Error() != "NOT_FOUND" {
		t.Errorf("Expected <NOT_FOUND>, received <%+v>", err)
	}
}

func TestGetRpcParams(t *testing.T) {
	testStruct := Attr{"", "", 0}
	RegisterRpcParams("", &RpcStruct{})
	if result, err := GetRpcParams("RpcStruct.Method1"); err != nil {
		t.Errorf("Expected <nil>, received <%+v>", err)
	} else if !reflect.DeepEqual(result.InParam, testStruct) {
		t.Errorf("Expected <%+v>, received <%+v>", testStruct, result.InParam)
	}
}

func TestUnregisterRpcParams(t *testing.T) {
	tests := []struct {
		name  string
		param string
	}{
		{
			name:  "Empty name",
			param: "",
		},
		{
			name:  "Test name",
			param: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UnregisterRpcParams(tt.param)
			if rpcParamsMap == nil {
				t.Errorf("Error getting the rpc object: %v", rpcParamsMap)
			}
		})
	}
}

func TestRegisterRpcParamsServiceErr(t *testing.T) {
	in := len(rpcParamsMap)
	RegisterRpcParams(EmptyString, "")

	if in != len(rpcParamsMap) {
		t.Errorf("Expected map length %d, but got map with length: %d", in, len(rpcParamsMap))
	}

}
