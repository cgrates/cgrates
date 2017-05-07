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
package v1

import (
	"reflect"
	"strings"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func NewRLsV1(rls *engine.ResourceLimiterService) *RLsV1 {
	return &RLsV1{rls: rls}
}

// Exports RPC from RLs
type RLsV1 struct {
	rls *engine.ResourceLimiterService
}

// Call implements rpcclient.RpcClientConnection interface for internal RPC
func (rlsv1 *RLsV1) Call(serviceMethod string, args interface{}, reply interface{}) error {
	methodSplit := strings.Split(serviceMethod, ".")
	if len(methodSplit) != 2 {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	method := reflect.ValueOf(rlsv1).MethodByName(methodSplit[1])
	if !method.IsValid() {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	params := []reflect.Value{reflect.ValueOf(args), reflect.ValueOf(reply)}
	ret := method.Call(params)
	if len(ret) != 1 {
		return utils.ErrServerError
	}
	if ret[0].Interface() == nil {
		return nil
	}
	err, ok := ret[0].Interface().(error)
	if !ok {
		return utils.ErrServerError
	}
	return err
}

// GetLimitsForEvent returns ResourceLimits matching a specific event
func (rlsv1 *RLsV1) GetLimitsForEvent(ev map[string]interface{}, reply *[]*engine.ResourceLimit) error {
	return rlsv1.rls.V1ResourceLimitsForEvent(ev, reply)
}

// AllowUsage checks if there are limits imposed for event
func (rlsv1 *RLsV1) AllowUsage(args utils.AttrRLsResourceUsage, allowed *bool) error {
	return rlsv1.rls.V1AllowUsage(args, allowed)
}

// V1InitiateResourceUsage records usage for an event
func (rlsv1 *RLsV1) AllocateResource(args utils.AttrRLsResourceUsage, reply *string) error {
	return rlsv1.rls.V1AllocateResource(args, reply)
}

// V1TerminateResourceUsage releases usage for an event
func (rlsv1 *RLsV1) ReleaseResource(args utils.AttrRLsResourceUsage, reply *string) error {
	return rlsv1.rls.V1ReleaseResource(args, reply)
}
