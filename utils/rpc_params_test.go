package utils

import (
	"reflect"
	"testing"
)

type RpcStruct struct{}

type Attr struct {
	Name    string
	Surname string
	Age     float64
}

func (rpc *RpcStruct) Hopa(normal Attr, out *float64) error {
	return nil
}

func (rpc *RpcStruct) Tropa(pointer *Attr, out *float64) error {
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
	a := x.InParam
	if v, err := FromMapStringInterfaceValue(map[string]interface{}{"Name": "a", "Surname": "b", "Age": 10.2}, reflect.ValueOf(a)); err != nil || v.(Attr).Name != "a" || v.(Attr).Surname != "b" || v.(Attr).Age != 10.2 {
		t.Errorf("error converting to struct: %+v (%v)", v, err)
	}
	//TODO: make pointer in arguments usable
	/*x, found = rpcParamsMap["RpcStruct.Tropa"]
	if !found {
		t.Errorf("error getting rpcobject: %v (%+v)", rpcParamsMap, x)
	}
	b := x.InParam
	log.Printf("T: %+v", b)
	if v, err := FromMapStringInterfaceValue(map[string]interface{}{"Name": "a", "Surname": "b", "Age": 10.2}, b); err != nil || v.(Attr).Name != "a" || v.(Attr).Surname != "b" || v.(Attr).Age != 10.2 {
		t.Errorf("error converting to struct: %+v (%v)", v, err)
	}*/
}
