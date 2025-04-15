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

package cdrs

import (
	"errors"
	"fmt"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
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
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.CDRsV1ProcessEvent, args.ID)
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
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.CDRsV1ProcessEventWithGet, args.ID)
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)

		if itm, has := engine.Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*evs = *cachedResp.Result.(*[]*utils.EventsWithOpts)
			}
			return cachedResp.Error
		}
		defer engine.Cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
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
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.CDRsV1ProcessStoredEvents, args.ID)
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey)
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

	fltrs, err := engine.GetFilters(ctx, args.FilterIDs, args.Tenant, cdrS.dm)
	if err != nil {
		return fmt.Errorf("preparing filters failed: %w", err)
	}
	cdrs, err := cdrS.db.GetCDRs(ctx, fltrs, args.APIOpts)
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
