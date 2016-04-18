package utils

import "testing"

type RpcStruct struct{}

func (rpc *RpcStruct) Hopa(normal string, out *float64) error {
	return nil
}

func (rpc *RpcStruct) Tropa(pointer *string, out *float64) error {
	return nil
}

func TestRPCObjectPointer(t *testing.T) {
	RegisterRpcObject("", &RpcStruct{})
	if len(RpcObjects) != 2 {
		t.Errorf("error registering rpc object: %v", RpcObjects)
	}
	x, found := RpcObjects["RpcStruct.Hopa"]
	if !found {
		t.Errorf("error getting rpcobject: %v (%+v)", RpcObjects, x)
	}
	x, found = RpcObjects["RpcStruct.Tropa"]
	if !found {
		t.Errorf("error getting rpcobject: %v (%+v)", RpcObjects, x)
	}
}
