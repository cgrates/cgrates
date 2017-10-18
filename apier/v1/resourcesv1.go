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
	"github.com/cgrates/cgrates/cache"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
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
	return utils.APIerRPCCall(rsv1, serviceMethod, args, reply)
}

// GetResourcesForEvent returns Resources matching a specific event
func (rsv1 *ResourceSV1) GetResourcesForEvent(args utils.ArgRSv1ResourceUsage, reply *engine.Resources) error {
	return rsv1.rls.V1ResourcesForEvent(args, reply)
}

// AllowUsage checks if there are limits imposed for event
func (rsv1 *ResourceSV1) AllowUsage(args utils.ArgRSv1ResourceUsage, allowed *bool) error {
	return rsv1.rls.V1AllowUsage(args, allowed)
}

// V1InitiateResourceUsage records usage for an event
func (rsv1 *ResourceSV1) AllocateResource(args utils.ArgRSv1ResourceUsage, reply *string) error {
	return rsv1.rls.V1AllocateResource(args, reply)
}

// V1TerminateResourceUsage releases usage for an event
func (rsv1 *ResourceSV1) ReleaseResource(args utils.ArgRSv1ResourceUsage, reply *string) error {
	return rsv1.rls.V1ReleaseResource(args, reply)
}

// GetResourceProfile returns a resource configuration
func (apierV1 *ApierV1) GetResourceProfile(arg utils.TenantID, reply *engine.ResourceProfile) error {
	if missing := utils.MissingStructFields(&arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if rcfg, err := apierV1.DataManager.GetResourceProfile(arg.Tenant, arg.ID, false, utils.NonTransactional); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *rcfg
	}
	return nil
}

//SetResourceProfile add a new resource configuration
func (apierV1 *ApierV1) SetResourceProfile(res *engine.ResourceProfile, reply *string) error {
	if missing := utils.MissingStructFields(res, []string{"Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierV1.DataManager.SetResourceProfile(res); err != nil {
		return utils.APIErrorHandler(err)
	}
	cache.RemKey(utils.ResourceProfilesPrefix+utils.ConcatenatedKey(res.Tenant, res.ID), true, "") // ToDo: Remove here with autoreload
	*reply = utils.OK
	return nil
}

//RemResourceProfile remove a specific resource configuration
func (apierV1 *ApierV1) RemResourceProfile(arg utils.TenantID, reply *string) error {
	if missing := utils.MissingStructFields(&arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierV1.DataManager.RemoveResourceProfile(arg.Tenant, arg.ID, utils.NonTransactional); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = utils.OK
	return nil
}
