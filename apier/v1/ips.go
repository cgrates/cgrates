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

package v1

import (
	"fmt"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewIPsV1 initializes the IPSv1 object.
func NewIPsV1(ipS *engine.IPService) *IPsV1 {
	return &IPsV1{ips: ipS}
}

// IPsV1 represents the RPC object to register for ips v1 APIs.
type IPsV1 struct {
	ips *engine.IPService
}

// GetIPAllocationForEvent returns the IPAllocations object matching the event.
func (s *IPsV1) GetIPAllocationForEvent(ctx *context.Context, args *utils.CGREvent, reply *engine.IPAllocations) error {
	return s.ips.V1GetIPAllocationForEvent(ctx, args, reply)
}

// AuthorizeIP checks if it's able to allocate an IP address for the given event.
func (s *IPsV1) AuthorizeIP(ctx *context.Context, args *utils.CGREvent, reply *engine.AllocatedIP) error {
	return s.ips.V1AuthorizeIP(ctx, args, reply)
}

// AllocateIP allocates an IP address for the given event.
func (s *IPsV1) AllocateIP(ctx *context.Context, args *utils.CGREvent, reply *engine.AllocatedIP) error {
	return s.ips.V1AllocateIP(ctx, args, reply)
}

// ReleaseIP releases an allocated IP address for the given event.
func (s *IPsV1) ReleaseIP(ctx *context.Context, args *utils.CGREvent, reply *string) error {
	return s.ips.V1ReleaseIP(ctx, args, reply)
}

// GetIPAllocations returns all IP allocations for a tenantID.
func (s *IPsV1) GetIPAllocations(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *engine.IPAllocations) error {
	return s.ips.V1GetIPAllocations(ctx, arg, reply)
}

// ClearIPAllocations clears IP allocations from an IPAllocations object.
// If args.AllocationIDs is empty or nil, all allocations will be cleared.
func (s *IPsV1) ClearIPAllocations(ctx *context.Context, arg *engine.ClearIPAllocationsArgs, reply *string) error {
	return s.ips.V1ClearIPAllocations(ctx, arg, reply)
}

// GetIPProfile retrieves the specificed IPProfile from data_db.
func (a *APIerSv1) GetIPProfile(ctx *context.Context, arg *utils.TenantID, reply *engine.IPProfile) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = a.Config.GeneralCfg().DefaultTenant
	}
	if rcfg, err := a.DataManager.GetIPProfile(tnt, arg.ID, true, true, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	} else {
		*reply = *rcfg
	}
	return nil
}

// GetIPProfileIDs returns list of IPProfile IDs registered for a tenant.
func (a *APIerSv1) GetIPProfileIDs(ctx *context.Context, args *utils.PaginatorWithTenant, rsPrfIDs *[]string) error {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = a.Config.GeneralCfg().DefaultTenant
	}
	prfx := utils.IPProfilesPrefix + tnt + utils.ConcatenatedKeySep
	keys, err := a.DataManager.DataDB().GetKeysForPrefix(prfx)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	retIDs := make([]string, len(keys))
	for i, key := range keys {
		retIDs[i] = key[len(prfx):]
	}
	*rsPrfIDs = args.PaginateStringSlice(retIDs)
	return nil
}

// SetIPProfile persists the passed IPProfile to data_db.
func (a *APIerSv1) SetIPProfile(ctx *context.Context, arg *engine.IPProfileWithAPIOpts, reply *string) (err error) {
	if missing := utils.MissingStructFields(arg.IPProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = a.Config.GeneralCfg().DefaultTenant
	}
	if err = a.DataManager.SetIPProfile(arg.IPProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheIPProfiles and CacheIPs and store it in database
	//make 1 insert for both IPProfile and IPs instead of 2
	loadID := time.Now().UnixNano()
	if err = a.DataManager.SetLoadIDs(
		map[string]int64{utils.CacheIPProfiles: loadID,
			utils.CacheIPAllocations: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if a.Config.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<SetIPProfile> Delaying cache call for %v", a.Config.GeneralCfg().CachingDelay))
		time.Sleep(a.Config.GeneralCfg().CachingDelay)
	}
	//handle caching for IPProfile
	if err = a.CallCache(utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]), arg.Tenant, utils.CacheIPProfiles,
		arg.TenantID(), utils.EmptyString, &arg.FilterIDs, nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveIPProfile removes the specified IPProfile from data_db.
func (a *APIerSv1) RemoveIPProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = a.Config.GeneralCfg().DefaultTenant
	}
	if err := a.DataManager.RemoveIPProfile(tnt, arg.ID, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if a.Config.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<RemoveIPProfile> Delaying cache call for %v", a.Config.GeneralCfg().CachingDelay))
		time.Sleep(a.Config.GeneralCfg().CachingDelay)
	}
	//handle caching for IPProfile
	if err := a.CallCache(utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]), tnt, utils.CacheIPProfiles,
		utils.ConcatenatedKey(tnt, arg.ID), utils.EmptyString, nil, nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheIPProfiles and CacheIPs and store it in database
	//make 1 insert for both IPProfile and IPs instead of 2
	loadID := time.Now().UnixNano()
	if err := a.DataManager.SetLoadIDs(map[string]int64{utils.CacheIPProfiles: loadID, utils.CacheIPAllocations: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}
