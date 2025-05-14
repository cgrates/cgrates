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
	"fmt"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewIPsV1(rls *engine.IPService) *IPsV1 {
	return &IPsV1{rls: rls}
}

type IPsV1 struct {
	rls *engine.IPService
}

// GetIPsForEvent returns IPs matching a specific event.
func (ip *IPsV1) GetIPsForEvent(ctx *context.Context, args *utils.CGREvent, reply *engine.IPs) error {
	return ip.rls.V1GetIPsForEvent(ctx, args, reply)
}

// AuthorizeIPs checks if there are limits imposed for event.
func (ip *IPsV1) AuthorizeIPs(ctx *context.Context, args *utils.CGREvent, reply *string) error {
	return ip.rls.V1AuthorizeIPs(ctx, args, reply)
}

// AllocateIPs records usage for an event.
func (ip *IPsV1) AllocateIPs(ctx *context.Context, args *utils.CGREvent, reply *string) error {
	return ip.rls.V1AllocateIPs(ctx, args, reply)
}

// V1TerminateIPUsage releases usage for an event
func (ip *IPsV1) ReleaseIPs(ctx *context.Context, args *utils.CGREvent, reply *string) error {
	return ip.rls.V1ReleaseIPs(ctx, args, reply)
}

// GetIP retrieves the specified IP from data_db.
func (ip *IPsV1) GetIP(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.IP) error {
	return ip.rls.V1GetIP(ctx, args, reply)
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
			utils.CacheIPs: loadID}); err != nil {
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
	if err := a.DataManager.SetLoadIDs(map[string]int64{utils.CacheIPProfiles: loadID, utils.CacheIPs: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}
