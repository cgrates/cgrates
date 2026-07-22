// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package cdrs

import (
	"errors"
	"fmt"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// V1ProcessEvent will process the CGREvent
func (cdrS *CDRServer) V1ProcessEvent(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	if args.ID == utils.EmptyString {
		args.ID = utils.GenUUID()
	}
	if args.Tenant == utils.EmptyString {
		args.Tenant = cdrS.cfg.GeneralCfg().DefaultTenant
	}
	// RPC caching
	if cdrS.cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.CDRsV1ProcessEvent, args.ID)
		unlock := cdrS.cache.LockRPCResponse(cacheKey) // RPC caching needs to be atomic
		defer unlock()

		if itm, has := cdrS.cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*string)
			}
			return cachedResp.Error
		}
		defer cdrS.cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	if _, err = cdrS.processEvents(ctx, []*utils.CGREvent{args}); err != nil {
		return
	}
	*reply = utils.OK
	return nil
}

// V1ProcessEventWithGet has the same logic with V1ProcessEvent except it adds the proccessed events to the reply
func (cdrS *CDRServer) V1ProcessEventWithGet(ctx *context.Context, args *utils.CGREvent, evs *[]*utils.EventsWithOpts) (err error) {
	if args.ID == utils.EmptyString {
		args.ID = utils.GenUUID()
	}
	if args.Tenant == utils.EmptyString {
		args.Tenant = cdrS.cfg.GeneralCfg().DefaultTenant
	}
	// RPC caching
	if cdrS.cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.CDRsV1ProcessEventWithGet, args.ID)
		unlock := cdrS.cache.LockRPCResponse(cacheKey) // RPC caching needs to be atomic
		defer unlock()

		if itm, has := cdrS.cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*evs = *cachedResp.Result.(*[]*utils.EventsWithOpts)
			}
			return cachedResp.Error
		}
		defer cdrS.cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: evs, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching
	var procEvs []*utils.EventsWithOpts
	if procEvs, err = cdrS.processEvents(ctx, []*utils.CGREvent{args}); err != nil {
		return
	}
	*evs = procEvs
	return nil
}

// V1ProcessStoredEvents processes stored events based on provided filters.
func (cdrS *CDRServer) V1ProcessStoredEvents(ctx *context.Context, args *utils.CDRFilters, reply *string) (err error) {
	if args.ID == utils.EmptyString {
		args.ID = utils.GenUUID()
	}
	if args.Tenant == utils.EmptyString {
		args.Tenant = cdrS.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if cdrS.cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.CDRsV1ProcessStoredEvents, args.ID)
		unlock := cdrS.cache.LockRPCResponse(cacheKey)
		defer unlock()

		if itm, has := cdrS.cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*string)
			}
			return cachedResp.Error
		}
		defer cdrS.cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}

	fltrs, err := engine.GetFilters(ctx, args.FilterIDs, args.Tenant, cdrS.dm)
	if err != nil {
		return fmt.Errorf("preparing filters failed: %w", err)
	}
	cdrs, err := cdrS.dm.GetCDRs(ctx, fltrs, args.APIOpts)
	if err != nil {
		return fmt.Errorf("retrieving CDRs failed: %w", err)
	}
	_, err = cdrS.processEvents(ctx, utils.CDRsToCGREvents(cdrs))
	if err != nil && !errors.Is(err, utils.ErrPartiallyExecuted) {
		return fmt.Errorf("processing events failed: %w", err)
	}
	*reply = utils.OK
	return err
}
