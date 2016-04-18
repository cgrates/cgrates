package utils

import "testing"

type RpcStruct struct{}

func (rpc *RpcStruct) Hopa(normal string, out *float64) error {
	return nil
}

func (rpc *RpcStruct) Tropa(pointer *string, out *float64) error {
	return nil
}

func (rpc *RpcStruct) Call(string, interface{}, interface{}) error {
	return nil
}

func TestRPCObjectPointer(t *testing.T) {
	RegisterRpcParams("", &RpcStruct{})
	if len(rpcParamsMap) != 2 {
		t.Errorf("error registering rpc object: %v", rpcParamsMap)
	}
	x, found := rpcParamsMap["RpcStruct.Hopa"]
	if !found {
		t.Errorf("error getting rpcobject: %v (%+v)", rpcParamsMap, x)
	}
	x, found = rpcParamsMap["RpcStruct.Tropa"]
	if !found {
		t.Errorf("error getting rpcobject: %v (%+v)", rpcParamsMap, x)
	}
}
