package utils

import (
	"reflect"
)

var rpcParamsMap map[string]*RpcParams

type RpcParams struct {
	Object   interface{}
	InParam  reflect.Value
	OutParam interface{}
}

func init() {
	rpcParamsMap = make(map[string]*RpcParams)
}

func RegisterRpcParams(name string, obj interface{}) {
	objType := reflect.TypeOf(obj)
	if name == "" {
		val := reflect.ValueOf(obj)
		name = objType.Name()
		if val.Kind() == reflect.Ptr {
			name = objType.Elem().Name()
		}
	}
	for i := 0; i < objType.NumMethod(); i++ {
		method := objType.Method(i)
		methodType := method.Type
		if methodType.NumIn() == 3 { // if it has three parameters (one is self and two are rpc params)
			out := methodType.In(2)
			if out.Kind() == reflect.Ptr {
				out = out.Elem()
			}
			rpcParamsMap[name+"."+method.Name] = &RpcParams{
				Object:   obj,
				InParam:  reflect.New(methodType.In(1)),
				OutParam: reflect.New(out).Interface(),
			}
		}

	}
}

func GetRpcParams(method string) (*RpcParams, error) {
	Logger.Info("REGISTERED: " + ToJSON(rpcParamsMap))
	x, found := rpcParamsMap[method]
	if !found {
		return nil, ErrNotFound
	}
	return x, nil
}
