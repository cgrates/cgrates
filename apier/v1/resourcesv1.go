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

func NewResourceSV1(rls *engine.ResourceService) *ResourceSV1 {
	return &ResourceSV1{rls: rls}
}

// Exports RPC from RLs
type ResourceSV1 struct {
	rls *engine.ResourceService
}

// Call implements rpcclient.RpcClientConnection interface for internal RPC
func (rsv1 *ResourceSV1) Call(serviceMethod string, args interface{}, reply interface{}) error {
	methodSplit := strings.Split(serviceMethod, ".")
	if len(methodSplit) != 2 {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	method := reflect.ValueOf(rsv1).MethodByName(methodSplit[1])
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

// GetResourcesForEvent returns Resources matching a specific event
func (rsv1 *ResourceSV1) GetResourcesForEvent(ev map[string]interface{}, reply *[]*engine.ResourceCfg) error {
	return rsv1.rls.V1ResourcesForEvent(ev, reply)
}

// AllowUsage checks if there are limits imposed for event
func (rsv1 *ResourceSV1) AllowUsage(args utils.AttrRLsResourceUsage, allowed *bool) error {
	return rsv1.rls.V1AllowUsage(args, allowed)
}

// V1InitiateResourceUsage records usage for an event
func (rsv1 *ResourceSV1) AllocateResource(args utils.AttrRLsResourceUsage, reply *string) error {
	return rsv1.rls.V1AllocateResource(args, reply)
}

// V1TerminateResourceUsage releases usage for an event
func (rsv1 *ResourceSV1) ReleaseResource(args utils.AttrRLsResourceUsage, reply *string) error {
	return rsv1.rls.V1ReleaseResource(args, reply)
}

type AttrGetResCfg struct {
	ID string
}

// GetResourceConfig returns a resource configuration
func (apierV1 *ApierV1) GetResourceConfig(attr AttrGetResCfg, reply *engine.ResourceCfg) error {
	if missing := utils.MissingStructFields(&attr, []string{"ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if rcfg, err := apierV1.DataDB.GetResourceCfg(attr.ID, true, utils.NonTransactional); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *rcfg
	}
	return nil
}

//SetResourceConfig add a new resource configuration
func (apierV1 *ApierV1) SetResourceConfig(attr *engine.ResourceCfg, reply *string) error {
	if missing := utils.MissingStructFields(attr, []string{"ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierV1.DataDB.SetResourceCfg(attr, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

//RemResourceConfig remove a specific resource configuration
func (apierV1 *ApierV1) RemResourceConfig(attrs AttrGetResCfg, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierV1.DataDB.RemoveResourceCfg(attrs.ID, utils.NonTransactional); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = utils.OK
	return nil
}
