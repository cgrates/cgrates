package utils

import "reflect"

var RpcObjects map[string]interface{}

type RpcObject struct {
	Object   interface{}
	InParam  interface{}
	OutParam interface{}
}

func init() {
	RpcObjects = make(map[string]interface{})
}

func RegisterRpcObject(name string, rpcObject interface{}) {
	objType := reflect.TypeOf(rpcObject)
	if name == "" {
		val := reflect.ValueOf(rpcObject)
		name = objType.Name()
		if val.Kind() == reflect.Ptr {
			name = objType.Elem().Name()
		}
	}
	for i := 0; i < objType.NumMethod(); i++ {
		method := objType.Method(i)
		methodType := method.Type
		if methodType.NumIn() == 3 { // if it has three parameters (one is self and two are rpc params)
			RpcObjects[name+"."+method.Name] = &RpcObject{
				Object:   objType,
				InParam:  reflect.New(methodType.In(1)).Interface(),
				OutParam: reflect.New(methodType.In(2).Elem()).Interface(),
			}
		}

	}
}
