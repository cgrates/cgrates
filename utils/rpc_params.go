/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package utils

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/cgrates/birpc"
)

var (
	rpcParamsMap  = make(map[string]*RpcParams)
	rpcParamsLock sync.Mutex
)

// RpcParams holds the parameters for an RPC method, including the object, input, and output parameters.
type RpcParams struct {
	Object   any
	InParam  any
	OutParam any
}

// RegisterRpcParams takes a receiver of any type and if it is of type *birpc.Service or can be wrapped within one, it attempts to
// populate the rpcParamsMap. Each entry in the map associates a method name with an RpcParams struct containing the parameters
// for that method.
// The name parameter is taken into consideration only if the receiver is not already of type *birpc.Service.
func RegisterRpcParams(name string, rcvr any) {

	// Attempt to cast the receiver to a *birpc.Service.
	srv, isService := rcvr.(*birpc.Service)
	if !isService {
		useName := name != EmptyString

		// If the cast fails, create a new service instance.
		var err error
		srv, err = birpc.NewService(rcvr, name, useName)
		if err != nil {
			Logger.Err(fmt.Sprintf("failed to register rpc parameters, service initialization error: %s", err))
			return
		}
	}
	rpcParamsLock.Lock()
	defer rpcParamsLock.Unlock()
	for mName, mValue := range srv.Methods {
		params := &RpcParams{
			Object: srv,

			// ReplyType will always be a pointer, therefore it is safe to be dereferenced
			// and then create a new pointer to the underlying value.
			OutParam: reflect.New(mValue.ReplyType.Elem()).Interface(),
		}
		if mValue.ArgType.Kind() == reflect.Ptr {
			params.InParam = reflect.New(mValue.ArgType.Elem()).Interface()
		} else {
			params.InParam = reflect.New(mValue.ArgType).Elem().Interface()
		}
		rpcParamsMap[srv.Name+"."+mName] = params
	}
}

// GetRpcParams retrieves the RpcParams for a given method name.
func GetRpcParams(method string) (params *RpcParams, err error) {
	var found bool
	rpcParamsLock.Lock()
	defer rpcParamsLock.Unlock()
	if params, found = rpcParamsMap[method]; !found {
		return nil, ErrNotFound
	}
	return
}

func UnregisterRpcParams(name string) {
	rpcParamsLock.Lock()
	defer rpcParamsLock.Unlock()
	for method := range rpcParamsMap {
		if strings.HasPrefix(method, name) {
			delete(rpcParamsMap, method)
		}
	}
	delete(rpcParamsMap, name)
}
