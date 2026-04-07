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

package resources

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
)

// V1GetResourcesForEvent returns active resource configs matching the event
func (s *ResourceS) V1GetResourcesForEvent(ctx *context.Context, args *utils.CGREvent, reply *Resources) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID, utils.Event}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	usageID, err := engine.GetStringOpts(ctx, args.Tenant, args.AsDataProvider(), nil, s.filters, s.cfg.ResourceSCfg().Opts.UsageID,
		utils.OptsResourcesUsageID)
	if err != nil {
		return err
	}

	ttl, err := engine.GetDurationOpts(ctx, args.Tenant, args.AsDataProvider(), nil, s.filters, s.cfg.ResourceSCfg().Opts.UsageTTL,
		utils.OptsResourcesUsageTTL)
	if err != nil {
		return err
	}
	usageTTL := utils.DurationPointer(ttl)

	if usageID == "" {
		return utils.NewErrMandatoryIeMissing(utils.UsageID)
	}
	tnt := args.Tenant
	if tnt == "" {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if s.cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.ResourceSv1GetResourcesForEvent, utils.ConcatenatedKey(tnt, args.ID))
		refID := guardian.Guardian.GuardIDs("",
			s.cfg.GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
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

	mtcRLs, unlock, err := s.matchingResourcesForEvent(ctx, tnt, args, usageID, usageTTL)
	if err != nil {
		return err
	}
	defer unlock()
	*reply = mtcRLs
	return nil
}

// V1AuthorizeResources queries service to find if an Usage is allowed
func (s *ResourceS) V1AuthorizeResources(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID, utils.Event}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	usageID, err := engine.GetStringOpts(ctx, args.Tenant, args.AsDataProvider(), nil, s.filters, s.cfg.ResourceSCfg().Opts.UsageID,
		utils.OptsResourcesUsageID)
	if err != nil {
		return err
	}

	units, err := engine.GetFloat64Opts(ctx, args.Tenant, args.AsDataProvider(), nil, s.filters, s.cfg.ResourceSCfg().Opts.Units,
		utils.OptsResourcesUnits)
	if err != nil {
		return err
	}

	ttl, err := engine.GetDurationOpts(ctx, args.Tenant, args.AsDataProvider(), nil, s.filters, s.cfg.ResourceSCfg().Opts.UsageTTL,
		utils.OptsResourcesUsageTTL)
	if err != nil {
		return err
	}
	usageTTL := utils.DurationPointer(ttl)

	if usageID == "" {
		return utils.NewErrMandatoryIeMissing(utils.UsageID)
	}

	tnt := args.Tenant
	if tnt == "" {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if s.cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.ResourceSv1AuthorizeResources, utils.ConcatenatedKey(tnt, args.ID))
		refID := guardian.Guardian.GuardIDs("",
			s.cfg.GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
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

	mtcRLs, unlock, err := s.matchingResourcesForEvent(ctx, tnt, args, usageID, usageTTL)
	if err != nil {
		return err
	}
	defer unlock()

	allocMsg, err := mtcRLs.allocateResource(&utils.ResourceUsage{
		Tenant: tnt,
		ID:     usageID,
		Units:  units}, true)
	if err != nil {
		if err == utils.ErrResourceUnavailable {
			return utils.ErrResourceUnauthorized
		}
		return err
	}
	*reply = allocMsg
	return nil
}

// V1AllocateResources is called when a resource requires allocation
func (s *ResourceS) V1AllocateResources(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID, utils.Event}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	usageID, err := engine.GetStringOpts(ctx, args.Tenant, args.AsDataProvider(), nil, s.filters, s.cfg.ResourceSCfg().Opts.UsageID,
		utils.OptsResourcesUsageID)
	if err != nil {
		return err
	}

	units, err := engine.GetFloat64Opts(ctx, args.Tenant, args.AsDataProvider(), nil, s.filters, s.cfg.ResourceSCfg().Opts.Units,
		utils.OptsResourcesUnits)
	if err != nil {
		return err
	}

	ttl, err := engine.GetDurationOpts(ctx, args.Tenant, args.AsDataProvider(), nil, s.filters, s.cfg.ResourceSCfg().Opts.UsageTTL,
		utils.OptsResourcesUsageTTL)
	if err != nil {
		return err
	}
	usageTTL := utils.DurationPointer(ttl)

	if usageID == "" {
		return utils.NewErrMandatoryIeMissing(utils.UsageID)
	}

	tnt := args.Tenant
	if tnt == "" {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if s.cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.ResourceSv1AllocateResources, utils.ConcatenatedKey(tnt, args.ID))
		refID := guardian.Guardian.GuardIDs("",
			s.cfg.GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
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

	mtcRLs, unlock, err := s.matchingResourcesForEvent(ctx, tnt, args, usageID, usageTTL)
	if err != nil {
		return err
	}
	defer unlock()

	allocMsg, err := mtcRLs.allocateResource(&utils.ResourceUsage{Tenant: tnt, ID: usageID,
		Units: units}, false)
	if err != nil {
		return err
	}

	if err := s.storeMatchedResources(ctx, mtcRLs); err != nil {
		return err
	}
	if err := s.processThresholds(ctx, mtcRLs, args.APIOpts); err != nil {
		return err
	}
	*reply = allocMsg
	return nil
}

// V1ReleaseResources is called when we need to clear an allocation
func (s *ResourceS) V1ReleaseResources(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID, utils.Event}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	usageID, err := engine.GetStringOpts(ctx, args.Tenant, args.AsDataProvider(), nil, s.filters, s.cfg.ResourceSCfg().Opts.UsageID,
		utils.OptsResourcesUsageID)
	if err != nil {
		return err
	}

	ttl, err := engine.GetDurationOpts(ctx, args.Tenant, args.AsDataProvider(), nil, s.filters, s.cfg.ResourceSCfg().Opts.UsageTTL,
		utils.OptsResourcesUsageTTL)
	if err != nil {
		return err
	}
	usageTTL := utils.DurationPointer(ttl)

	if usageID == "" {
		return utils.NewErrMandatoryIeMissing(utils.UsageID)
	}

	tnt := args.Tenant
	if tnt == "" {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if s.cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.ResourceSv1ReleaseResources, utils.ConcatenatedKey(tnt, args.ID))
		refID := guardian.Guardian.GuardIDs("",
			s.cfg.GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
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

	mtcRLs, unlock, err := s.matchingResourcesForEvent(ctx, tnt, args, usageID, usageTTL)
	if err != nil {
		return err
	}
	defer unlock()

	if err := mtcRLs.clearUsage(usageID); err != nil {
		return err
	}

	if err := s.storeMatchedResources(ctx, mtcRLs); err != nil {
		return err
	}
	if err := s.processThresholds(ctx, mtcRLs, args.APIOpts); err != nil {
		return err
	}

	*reply = utils.OK
	return nil
}

// V1GetResource returns a resource configuration
func (s *ResourceS) V1GetResource(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *utils.Resource) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == "" {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	// make sure resource is locked at process level
	lkID := guardian.Guardian.GuardIDs("",
		s.cfg.GeneralCfg().LockingTimeout,
		utils.ResourceLockKey(tnt, arg.ID))
	defer guardian.Guardian.UnguardIDs(lkID)

	res, err := s.dm.GetResource(ctx, tnt, arg.ID, true, true, utils.NonTransactional)
	if err != nil {
		return err
	}
	*reply = *res
	return nil
}

func (s *ResourceS) V1GetResourceWithConfig(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *utils.ResourceWithConfig) (err error) {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == "" {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	lkID := guardian.Guardian.GuardIDs("",
		s.cfg.GeneralCfg().LockingTimeout,
		utils.ResourceLockKey(tnt, arg.ID))
	defer guardian.Guardian.UnguardIDs(lkID)

	res, err := s.dm.GetResource(ctx, tnt, arg.ID, true, true, utils.NonTransactional)
	if err != nil {
		return err
	}

	cfg, err := s.dm.GetResourceProfile(ctx, tnt, arg.ID, true, true, utils.NonTransactional)
	if err != nil {
		return err
	}

	*reply = utils.ResourceWithConfig{
		Resource: res,
		Config:   cfg,
	}
	return nil
}
