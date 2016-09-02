/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/
package utils

import "reflect"

var rpcParamsMap map[string]*RpcParams

type RpcParams struct {
	Object   interface{}
	InParam  interface{}
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
				InParam:  reflect.New(methodType.In(1)).Interface(),
				OutParam: reflect.New(out).Interface(),
			}
		}
	}
}

func GetRpcParams(method string) (*RpcParams, error) {
	x, found := rpcParamsMap[method]
	if !found {
		return nil, ErrNotFound
	}
	return x, nil
}
