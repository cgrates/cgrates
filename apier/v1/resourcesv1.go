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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewResourceSv1(rls *engine.ResourceService) *ResourceSv1 {
	return &ResourceSv1{rls: rls}
}

// Exports RPC from RLs
type ResourceSv1 struct {
	rls *engine.ResourceService
}

// Call implements rpcclient.RpcClientConnection interface for internal RPC
func (rsv1 *ResourceSv1) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(rsv1, serviceMethod, args, reply)
}

// GetResourcesForEvent returns Resources matching a specific event
func (rsv1 *ResourceSv1) GetResourcesForEvent(args utils.ArgRSv1ResourceUsage, reply *engine.Resources) error {
	return rsv1.rls.V1ResourcesForEvent(args, reply)
}

// AuthorizeResources checks if there are limits imposed for event
func (rsv1 *ResourceSv1) AuthorizeResources(args utils.ArgRSv1ResourceUsage, reply *string) error {
	return rsv1.rls.V1AuthorizeResources(args, reply)
}

// V1InitiateResourceUsage records usage for an event
func (rsv1 *ResourceSv1) AllocateResources(args utils.ArgRSv1ResourceUsage, reply *string) error {
	return rsv1.rls.V1AllocateResource(args, reply)
}

// V1TerminateResourceUsage releases usage for an event
func (rsv1 *ResourceSv1) ReleaseResources(args utils.ArgRSv1ResourceUsage, reply *string) error {
	return rsv1.rls.V1ReleaseResource(args, reply)
}

// GetResourceProfile returns a resource configuration
func (apierV1 *ApierV1) GetResourceProfile(arg utils.TenantID, reply *engine.ResourceProfile) error {
	if missing := utils.MissingStructFields(&arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if rcfg, err := apierV1.DataManager.GetResourceProfile(arg.Tenant, arg.ID, true, true, utils.NonTransactional); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *rcfg
	}
	return nil
}

// GetResourceProfileIDs returns list of resourceProfile IDs registered for a tenant
func (apierV1 *ApierV1) GetResourceProfileIDs(tenant string, rsPrfIDs *[]string) error {
	prfx := utils.ResourceProfilesPrefix + tenant + ":"
	keys, err := apierV1.DataManager.DataDB().GetKeysForPrefix(prfx)
	if err != nil {
		return err
	}
	retIDs := make([]string, len(keys))
	for i, key := range keys {
		retIDs[i] = key[len(prfx):]
	}
	*rsPrfIDs = retIDs
	return nil
}

//SetResourceProfile add a new resource configuration
func (apierV1 *ApierV1) SetResourceProfile(res *engine.ResourceProfile, reply *string) error {
	if missing := utils.MissingStructFields(res, []string{"Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierV1.DataManager.SetResourceProfile(res, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := apierV1.DataManager.SetResource(
		&engine.Resource{Tenant: res.Tenant,
			ID:     res.ID,
			Usages: make(map[string]*engine.ResourceUsage)}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

//RemoveResourceProfile remove a specific resource configuration
func (apierV1 *ApierV1) RemoveResourceProfile(arg utils.TenantID, reply *string) error {
	if missing := utils.MissingStructFields(&arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierV1.DataManager.RemoveResourceProfile(arg.Tenant, arg.ID, utils.NonTransactional, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := apierV1.DataManager.RemoveResource(arg.Tenant, arg.ID, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

func (rsv1 *ResourceSv1) Ping(ign string, reply *string) error {
	*reply = utils.Pong
	return nil
}
