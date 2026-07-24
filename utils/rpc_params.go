// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"reflect"
	"sync"
)

var (
	rpcParamsMap  = make(map[string]*RpcParams)
	rpcParamsLock sync.RWMutex
)

type RpcParams struct {
	Object   any
	InParam  any
	OutParam any
}

func RegisterRpcParams(name string, obj any) {
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
			rpcParamsLock.Lock()
			rpcParamsMap[name+"."+method.Name] = &RpcParams{
				Object:   obj,
				InParam:  reflect.New(methodType.In(1)).Interface(),
				OutParam: reflect.New(out).Interface(),
			}
			rpcParamsLock.Unlock()
		}
	}
}

func GetRpcParams(method string) (*RpcParams, error) {
	rpcParamsLock.Lock()
	x, found := rpcParamsMap[method]
	rpcParamsLock.Unlock()
	if !found {
		return nil, ErrNotFound
	}
	return x, nil
}
