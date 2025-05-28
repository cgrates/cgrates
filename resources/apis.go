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

package resources

import (
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
)

// V1GetResourcesForEvent returns active resource configs matching the event
func (rS *ResourceS) V1GetResourcesForEvent(ctx *context.Context, args *utils.CGREvent, reply *Resources) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID, utils.Event}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	var usageID string
	if usageID, err = engine.GetStringOpts(ctx, args.Tenant, args.AsDataProvider(), nil, rS.fltrS, rS.cfg.ResourceSCfg().Opts.UsageID,
		utils.OptsResourcesUsageID); err != nil {
		return
	}

	var ttl time.Duration
	if ttl, err = engine.GetDurationOpts(ctx, args.Tenant, args.AsDataProvider(), nil, rS.fltrS, rS.cfg.ResourceSCfg().Opts.UsageTTL,
		utils.OptsResourcesUsageTTL); err != nil {
		return
	}
	usageTTL := utils.DurationPointer(ttl)

	if usageID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.UsageID)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = rS.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.ResourceSv1GetResourcesForEvent, utils.ConcatenatedKey(tnt, args.ID))
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)
		if itm, has := engine.Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*Resources)
			}
			return cachedResp.Error
		}
		defer engine.Cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	var mtcRLs Resources
	if mtcRLs, err = rS.matchingResourcesForEvent(ctx, tnt, args, usageID, usageTTL); err != nil {
		return err
	}
	*reply = mtcRLs
	mtcRLs.unlock()
	return
}

// V1AuthorizeResources queries service to find if an Usage is allowed
func (rS *ResourceS) V1AuthorizeResources(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID, utils.Event}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	var usageID string
	if usageID, err = engine.GetStringOpts(ctx, args.Tenant, args.AsDataProvider(), nil, rS.fltrS, rS.cfg.ResourceSCfg().Opts.UsageID,
		utils.OptsResourcesUsageID); err != nil {
		return
	}

	var units float64
	if units, err = engine.GetFloat64Opts(ctx, args.Tenant, args.AsDataProvider(), nil, rS.fltrS, rS.cfg.ResourceSCfg().Opts.Units,
		utils.OptsResourcesUnits); err != nil {
		return
	}

	var ttl time.Duration
	if ttl, err = engine.GetDurationOpts(ctx, args.Tenant, args.AsDataProvider(), nil, rS.fltrS, rS.cfg.ResourceSCfg().Opts.UsageTTL,
		utils.OptsResourcesUsageTTL); err != nil {
		return
	}
	usageTTL := utils.DurationPointer(ttl)

	if usageID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.UsageID)
	}

	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = rS.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.ResourceSv1AuthorizeResources, utils.ConcatenatedKey(tnt, args.ID))
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)
		if itm, has := engine.Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*string)
			}
			return cachedResp.Error
		}
		defer engine.Cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	var mtcRLs Resources
	if mtcRLs, err = rS.matchingResourcesForEvent(ctx, tnt, args, usageID, usageTTL); err != nil {
		return err
	}
	defer mtcRLs.unlock()

	var alcMessage string
	if alcMessage, err = mtcRLs.allocateResource(&utils.ResourceUsage{
		Tenant: tnt,
		ID:     usageID,
		Units:  units}, true); err != nil {
		if err == utils.ErrResourceUnavailable {
			err = utils.ErrResourceUnauthorized
		}
		return
	}
	*reply = alcMessage
	return
}

// V1AllocateResources is called when a resource requires allocation
func (rS *ResourceS) V1AllocateResources(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID, utils.Event}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	var usageID string
	if usageID, err = engine.GetStringOpts(ctx, args.Tenant, args.AsDataProvider(), nil, rS.fltrS, rS.cfg.ResourceSCfg().Opts.UsageID,
		utils.OptsResourcesUsageID); err != nil {
		return
	}

	var units float64
	if units, err = engine.GetFloat64Opts(ctx, args.Tenant, args.AsDataProvider(), nil, rS.fltrS, rS.cfg.ResourceSCfg().Opts.Units,
		utils.OptsResourcesUnits); err != nil {
		return
	}

	var ttl time.Duration
	if ttl, err = engine.GetDurationOpts(ctx, args.Tenant, args.AsDataProvider(), nil, rS.fltrS, rS.cfg.ResourceSCfg().Opts.UsageTTL,
		utils.OptsResourcesUsageTTL); err != nil {
		return
	}
	usageTTL := utils.DurationPointer(ttl)

	if usageID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.UsageID)
	}

	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = rS.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.ResourceSv1AllocateResources, utils.ConcatenatedKey(tnt, args.ID))
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)
		if itm, has := engine.Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*string)
			}
			return cachedResp.Error
		}
		defer engine.Cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	var mtcRLs Resources
	if mtcRLs, err = rS.matchingResourcesForEvent(ctx, tnt, args, usageID,
		usageTTL); err != nil {
		return err
	}
	defer mtcRLs.unlock()

	var alcMsg string
	if alcMsg, err = mtcRLs.allocateResource(&utils.ResourceUsage{Tenant: tnt, ID: usageID,
		Units: units}, false); err != nil {
		return
	}

	// index it for storing
	if err = rS.storeMatchedResources(ctx, mtcRLs); err != nil {
		return
	}
	if err = rS.processThresholds(ctx, mtcRLs, args.APIOpts); err != nil {
		return
	}
	*reply = alcMsg
	return
}

// V1ReleaseResources is called when we need to clear an allocation
func (rS *ResourceS) V1ReleaseResources(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID, utils.Event}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	var usageID string
	if usageID, err = engine.GetStringOpts(ctx, args.Tenant, args.AsDataProvider(), nil, rS.fltrS, rS.cfg.ResourceSCfg().Opts.UsageID,
		utils.OptsResourcesUsageID); err != nil {
		return
	}

	var ttl time.Duration
	if ttl, err = engine.GetDurationOpts(ctx, args.Tenant, args.AsDataProvider(), nil, rS.fltrS, rS.cfg.ResourceSCfg().Opts.UsageTTL,
		utils.OptsResourcesUsageTTL); err != nil {
		return
	}
	usageTTL := utils.DurationPointer(ttl)

	if usageID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.UsageID)
	}

	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = rS.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.ResourceSv1ReleaseResources, utils.ConcatenatedKey(tnt, args.ID))
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)
		if itm, has := engine.Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*string)
			}
			return cachedResp.Error
		}
		defer engine.Cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	var mtcRLs Resources
	if mtcRLs, err = rS.matchingResourcesForEvent(ctx, tnt, args, usageID,
		usageTTL); err != nil {
		return
	}
	defer mtcRLs.unlock()

	if err = mtcRLs.clearUsage(usageID); err != nil {
		return
	}

	// Handle storing
	if err = rS.storeMatchedResources(ctx, mtcRLs); err != nil {
		return
	}
	if err = rS.processThresholds(ctx, mtcRLs, args.APIOpts); err != nil {
		return
	}

	*reply = utils.OK
	return
}

// V1GetResource returns a resource configuration
func (rS *ResourceS) V1GetResource(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *utils.Resource) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = rS.cfg.GeneralCfg().DefaultTenant
	}

	// make sure resource is locked at process level
	lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout,
		utils.ResourceLockKey(tnt, arg.ID))
	defer guardian.Guardian.UnguardIDs(lkID)

	res, err := rS.dm.GetResource(ctx, tnt, arg.ID, true, true, utils.NonTransactional)
	if err != nil {
		return err
	}
	*reply = *res
	return nil
}

func (rS *ResourceS) V1GetResourceWithConfig(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *utils.ResourceWithConfig) (err error) {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = rS.cfg.GeneralCfg().DefaultTenant
	}

	// make sure resource is locked at process level
	lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout,
		utils.ResourceLockKey(tnt, arg.ID))
	defer guardian.Guardian.UnguardIDs(lkID)

	var res *utils.Resource
	res, err = rS.dm.GetResource(ctx, tnt, arg.ID, true, true, utils.NonTransactional)
	if err != nil {
		return
	}

	// make sure resourceProfile is locked at process level
	lkPrflID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout,
		utils.ResourceProfileLockKey(tnt, arg.ID))
	defer guardian.Guardian.UnguardIDs(lkPrflID)

	var cfg *utils.ResourceProfile
	cfg, err = rS.dm.GetResourceProfile(ctx, tnt, arg.ID, true, true, utils.NonTransactional)
	if err != nil {
		return
	}

	*reply = utils.ResourceWithConfig{
		Resource: res,
		Config:   cfg,
	}

	return
}
