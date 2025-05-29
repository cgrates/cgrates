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

package ips

import (
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
)

// V1GetIPsForEvent returns active IPs matching the event.
func (s *IPService) V1GetIPsForEvent(ctx *context.Context, args *utils.CGREvent, reply *IPs) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID, utils.Event}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	var usageID string
	if usageID, err = engine.GetStringOpts(ctx, args.Tenant, args.AsDataProvider(), nil, s.fltrs, s.cfg.IPsCfg().Opts.UsageID,
		utils.OptsIPsUsageID); err != nil {
		return
	}

	var ttl time.Duration
	if ttl, err = engine.GetDurationOpts(ctx, args.Tenant, args.AsDataProvider(), nil, s.fltrs, s.cfg.IPsCfg().Opts.TTL,
		utils.OptsIPsTTL); err != nil {
		return
	}
	usageTTL := utils.DurationPointer(ttl)

	if usageID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.UsageID)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.IPsV1GetIPsForEvent, utils.ConcatenatedKey(tnt, args.ID))
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)
		if itm, has := engine.Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*IPs)
			}
			return cachedResp.Error
		}
		defer engine.Cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	var mtcRLs IPs
	if mtcRLs, err = s.matchingIPsForEvent(ctx, tnt, args, usageID, usageTTL); err != nil {
		return err
	}
	*reply = mtcRLs
	mtcRLs.unlock()
	return
}

// V1AuthorizeIPs queries service to find if an Usage is allowed
func (s *IPService) V1AuthorizeIPs(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID, utils.Event}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	var usageID string
	if usageID, err = engine.GetStringOpts(ctx, args.Tenant, args.AsDataProvider(), nil, s.fltrs, s.cfg.IPsCfg().Opts.UsageID,
		utils.OptsIPsUsageID); err != nil {
		return
	}

	if _, err = engine.GetFloat64Opts(ctx, args.Tenant, args.AsDataProvider(), nil, s.fltrs, s.cfg.IPsCfg().Opts.Units,
		utils.OptsIPsUnits); err != nil {
		return
	}

	var ttl time.Duration
	if ttl, err = engine.GetDurationOpts(ctx, args.Tenant, args.AsDataProvider(), nil, s.fltrs, s.cfg.IPsCfg().Opts.TTL,
		utils.OptsIPsTTL); err != nil {
		return
	}
	usageTTL := utils.DurationPointer(ttl)

	if usageID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.UsageID)
	}

	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.IPsV1AuthorizeIPs, utils.ConcatenatedKey(tnt, args.ID))
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

	var mtcRLs IPs
	if mtcRLs, err = s.matchingIPsForEvent(ctx, tnt, args, usageID, usageTTL); err != nil {
		return err
	}
	defer mtcRLs.unlock()

	/*
		authorize logic
		...
	*/

	*reply = utils.OK
	return
}

// V1AllocateIPs is called when an IP requires allocation.
func (s *IPService) V1AllocateIPs(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID, utils.Event}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	var usageID string
	if usageID, err = engine.GetStringOpts(ctx, args.Tenant, args.AsDataProvider(), nil, s.fltrs, s.cfg.IPsCfg().Opts.UsageID,
		utils.OptsIPsUsageID); err != nil {
		return
	}

	if _, err = engine.GetFloat64Opts(ctx, args.Tenant, args.AsDataProvider(), nil, s.fltrs, s.cfg.IPsCfg().Opts.Units,
		utils.OptsIPsUnits); err != nil {
		return
	}

	var ttl time.Duration
	if ttl, err = engine.GetDurationOpts(ctx, args.Tenant, args.AsDataProvider(), nil, s.fltrs, s.cfg.IPsCfg().Opts.TTL,
		utils.OptsIPsTTL); err != nil {
		return
	}
	usageTTL := utils.DurationPointer(ttl)

	if usageID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.UsageID)
	}

	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.IPsV1AllocateIPs, utils.ConcatenatedKey(tnt, args.ID))
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

	var mtcRLs IPs
	if mtcRLs, err = s.matchingIPsForEvent(ctx, tnt, args, usageID,
		usageTTL); err != nil {
		return err
	}
	defer mtcRLs.unlock()

	/*
		allocate logic
		...
	*/

	// index it for storing
	if err = s.storeMatchedIPs(ctx, mtcRLs); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// V1ReleaseIPs is called when we need to clear an allocation
func (s *IPService) V1ReleaseIPs(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID, utils.Event}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	var usageID string
	if usageID, err = engine.GetStringOpts(ctx, args.Tenant, args.AsDataProvider(), nil, s.fltrs, s.cfg.IPsCfg().Opts.UsageID,
		utils.OptsIPsUsageID); err != nil {
		return
	}

	var ttl time.Duration
	if ttl, err = engine.GetDurationOpts(ctx, args.Tenant, args.AsDataProvider(), nil, s.fltrs, s.cfg.IPsCfg().Opts.TTL,
		utils.OptsIPsTTL); err != nil {
		return
	}
	usageTTL := utils.DurationPointer(ttl)

	if usageID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.UsageID)
	}

	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.IPsV1ReleaseIPs, utils.ConcatenatedKey(tnt, args.ID))
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

	var mtcRLs IPs
	if mtcRLs, err = s.matchingIPsForEvent(ctx, tnt, args, usageID,
		usageTTL); err != nil {
		return
	}
	defer mtcRLs.unlock()

	/*
		release logic
		...
	*/

	// Handle storing
	if err = s.storeMatchedIPs(ctx, mtcRLs); err != nil {
		return
	}

	*reply = utils.OK
	return
}

// V1GetIP returns a resource configuration
func (s *IPService) V1GetIP(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *utils.IP) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	// make sure resource is locked at process level
	lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout,
		utils.IPLockKey(tnt, arg.ID))
	defer guardian.Guardian.UnguardIDs(lkID)

	ip, err := s.dm.GetIP(ctx, tnt, arg.ID, true, true, utils.NonTransactional)
	if err != nil {
		return err
	}
	*reply = *ip
	return nil
}
