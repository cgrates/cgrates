package utils

import "testing"

type RpcStruct struct{}

func (rpc *RpcStruct) Hopa(normal string, out *float64) error {
	return nil
}

func TestRPCObjectSimple(t *testing.T) {

	RegisterRpcObject("", &RpcStruct{})
	if len(RpcObjects) != 1 {
		t.Errorf("error registering rpc object: %v", RpcObjects)
	}
	x, found := RpcObjects["RpcStruct.Hopa"]
	if found {
		t.Errorf("error getting rpcobject: %v (%+v)", RpcObjects, x)
	}
}
