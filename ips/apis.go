// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ips

import (
	"errors"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// V1GetIPAllocationForEvent returns the IPAllocations object matching the event.
func (s *IPs) V1GetIPAllocationForEvent(ctx *context.Context, args *utils.CGREvent, reply *utils.IPAllocations) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID, utils.Event}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	var allocID string
	if allocID, err = engine.GetStringOpts(ctx, args.Tenant, args.AsDataProvider(), nil, s.filters, s.cfg.IPsCfg().Opts.AllocationID,
		utils.OptsIPsAllocationID); err != nil {
		return err
	}

	if allocID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.AllocationID)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if s.cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.IPsV1GetIPAllocationForEvent, utils.ConcatenatedKey(tnt, args.ID))
		unlock := s.cache.LockRPCResponse(cacheKey) // RPC caching needs to be atomic
		defer unlock()
		if itm, has := s.cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*utils.IPAllocations)
			}
			return cachedResp.Error
		}
		defer func() {
			s.cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
				&utils.CachedRPCResponse{Result: reply, Error: err},
				nil, true, utils.NonTransactional)
		}()
	}
	// end of RPC caching

	matched, unlock, err := s.matchingIPAllocationsForEvent(ctx, tnt, args, allocID)
	if err != nil {
		return err
	}
	defer unlock()
	*reply = *matched.allocs
	return nil
}

// V1AuthorizeIP checks if it's able to allocate an IP address for the given event.
func (s *IPs) V1AuthorizeIP(ctx *context.Context, args *utils.CGREvent, reply *utils.AllocatedIP) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID, utils.Event}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	var allocID string
	if allocID, err = engine.GetStringOpts(ctx, args.Tenant, args.AsDataProvider(), nil, s.filters,
		s.cfg.IPsCfg().Opts.AllocationID, utils.OptsIPsAllocationID); err != nil {
		return err
	}
	if allocID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.AllocationID)
	}

	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if s.cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.IPsV1AuthorizeIP, utils.ConcatenatedKey(tnt, args.ID))
		unlock := s.cache.LockRPCResponse(cacheKey) // RPC caching needs to be atomic
		defer unlock()
		if itm, has := s.cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*utils.AllocatedIP)
			}
			return cachedResp.Error
		}
		defer func() {
			s.cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
				&utils.CachedRPCResponse{Result: reply, Error: err},
				nil, true, utils.NonTransactional)
		}()
	}
	// end of RPC caching

	matched, unlock, err := s.matchingIPAllocationsForEvent(ctx, tnt, args, allocID)
	if err != nil {
		return err
	}
	defer unlock()

	var poolIDs []string
	if poolIDs, err = filterAndSortPools(ctx, tnt, matched.profile.Pools, s.filters,
		args.AsDataProvider()); err != nil {
		return err
	}

	var allocIP *utils.AllocatedIP
	if allocIP, err = matched.allocateFromPools(allocID, poolIDs, true); err != nil {
		if errors.Is(err, utils.ErrIPAlreadyAllocated) {
			return utils.ErrIPUnauthorized
		}
		return err
	}

	*reply = *allocIP
	return nil
}

// V1AllocateIP allocates an IP address for the given event.
func (s *IPs) V1AllocateIP(ctx *context.Context, args *utils.CGREvent, reply *utils.AllocatedIP) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID, utils.Event}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	var allocID string
	if allocID, err = engine.GetStringOpts(ctx, args.Tenant, args.AsDataProvider(), nil, s.filters, s.cfg.IPsCfg().Opts.AllocationID,
		utils.OptsIPsAllocationID); err != nil {
		return err
	}
	if allocID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.AllocationID)
	}

	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if s.cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.IPsV1AllocateIP, utils.ConcatenatedKey(tnt, args.ID))
		unlock := s.cache.LockRPCResponse(cacheKey) // RPC caching needs to be atomic
		defer unlock()
		if itm, has := s.cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*utils.AllocatedIP)
			}
			return cachedResp.Error
		}
		defer func() {
			s.cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
				&utils.CachedRPCResponse{Result: reply, Error: err},
				nil, true, utils.NonTransactional)
		}()
	}
	// end of RPC caching

	matched, unlock, err := s.matchingIPAllocationsForEvent(ctx, tnt, args, allocID)
	if err != nil {
		return err
	}
	defer unlock()

	var poolIDs []string
	if poolIDs, err = filterAndSortPools(ctx, tnt, matched.profile.Pools, s.filters,
		args.AsDataProvider()); err != nil {
		return err
	}

	var allocIP *utils.AllocatedIP
	if allocIP, err = matched.allocateFromPools(allocID, poolIDs, false); err != nil {
		return err
	}

	// index it for storing
	if err = s.storeMatchedIPAllocations(ctx, matched.allocs); err != nil {
		return err
	}
	*reply = *allocIP
	return nil
}

// V1ReleaseIP releases an allocated IP address for the given event.
func (s *IPs) V1ReleaseIP(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID, utils.Event}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	var allocID string
	if allocID, err = engine.GetStringOpts(ctx, args.Tenant, args.AsDataProvider(), nil, s.filters, s.cfg.IPsCfg().Opts.AllocationID,
		utils.OptsIPsAllocationID); err != nil {
		return err
	}
	if allocID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.AllocationID)
	}

	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if s.cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.IPsV1ReleaseIP, utils.ConcatenatedKey(tnt, args.ID))
		unlock := s.cache.LockRPCResponse(cacheKey) // RPC caching needs to be atomic
		defer unlock()
		if itm, has := s.cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*string)
			}
			return cachedResp.Error
		}
		defer func() {
			s.cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
				&utils.CachedRPCResponse{Result: reply, Error: err},
				nil, true, utils.NonTransactional)
		}()
	}
	// end of RPC caching

	matched, unlock, err := s.matchingIPAllocationsForEvent(ctx, tnt, args, allocID)
	if err != nil {
		return err
	}
	defer unlock()

	if err = matched.releaseAllocation(allocID); err != nil {
		return err
	}

	// Handle storing
	if err = s.storeMatchedIPAllocations(ctx, matched.allocs); err != nil {
		return err
	}

	*reply = utils.OK
	return nil
}

// V1GetIPAllocations returns all IP allocations for a tenantID.
func (s *IPs) V1GetIPAllocations(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *utils.IPAllocations) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	// make sure resource is locked at process level
	unlock := s.dm.Lock(utils.IPAllocationsLockKey(tnt, arg.ID))
	defer unlock()

	ip, err := s.dm.GetIPAllocations(ctx, tnt, arg.ID, true, true, utils.NonTransactional)
	if err != nil {
		return err
	}
	*reply = *ip
	return nil
}

// V1ClearIPAllocations clears IP allocations from an IPAllocations object.
// If args.AllocationIDs is empty or nil, all allocations will be cleared.
func (s *IPs) V1ClearIPAllocations(ctx *context.Context, args *utils.ClearIPAllocationsArgs, reply *string) error {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	unlock := s.dm.Lock(utils.IPAllocationsLockKey(tnt, args.ID))
	defer unlock()

	allocs, err := s.dm.GetIPAllocations(ctx, tnt, args.ID, true, true, utils.NonTransactional)
	if err != nil {
		return err
	}
	prfl, err := s.dm.GetIPProfile(ctx, tnt, args.ID, true, true, utils.NonTransactional)
	if err != nil {
		return err
	}
	matched, err := newMatchedIPAllocs(allocs, prfl)
	if err != nil {
		return err
	}
	if err := matched.clearAllocations(args.AllocationIDs); err != nil {
		return err
	}
	if err := s.storeIPAllocations(ctx, matched.allocs); err != nil {
		return err
	}

	*reply = utils.OK
	return nil
}
