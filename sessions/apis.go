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
package sessions

import (
	"fmt"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/attributes"
	"github.com/cgrates/cgrates/chargers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/rates"
	"github.com/cgrates/cgrates/routes"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
	"github.com/ericlagergren/decimal"
)

// BiRPCv1AuthorizeEvent performs authorization for CGREvent based on specific subsystems
func (sS *SessionS) BiRPCv1AuthorizeEvent(ctx *context.Context,
	args *utils.CGREvent, authReply *V1AuthorizeReply) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if args.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if args.APIOpts == nil {
		args.APIOpts = make(map[string]any)
	}
	if args.ID == "" {
		args.ID = utils.GenUUID()
	}
	if args.Tenant == "" {
		args.Tenant = sS.cfg.GeneralCfg().DefaultTenant
	}
	// RPC caching
	if sS.cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.SessionSv1AuthorizeEvent, args.ID)
		refID := guardian.Guardian.GuardIDs("",
			sS.cfg.GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)

		if itm, has := sS.cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*authReply = *cachedResp.Result.(*V1AuthorizeReply)
			}
			return cachedResp.Error
		}
		defer sS.cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: authReply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching
	dP := args.AsDataProvider()
	var attrS bool
	if attrS, err = engine.GetBoolOpts(ctx, args.Tenant, dP, nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.Attributes,
		utils.MetaAttributes); err != nil {
		return
	}
	var acntS bool
	if acntS, err = engine.GetBoolOpts(ctx, args.Tenant, dP, nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.Accounts,
		utils.MetaAccounts); err != nil {
		return
	}
	var routeS bool
	if routeS, err = engine.GetBoolOpts(ctx, args.Tenant, dP, nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.Routes,
		utils.MetaRoutes); err != nil {
		return
	}
	var resourceS bool
	if resourceS, err = engine.GetBoolOpts(ctx, args.Tenant, dP, nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.ResourcesAuthorize,
		utils.MetaResources); err != nil {
		return
	}
	var ipS bool
	if ipS, err = engine.GetBoolOpts(ctx, args.Tenant, dP, nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.IPsAuthorize,
		utils.MetaIPs); err != nil {
		return
	}
	if !attrS && !acntS && !resourceS && !ipS && !routeS {
		return // Nothing to do
	}
	if attrS {
		rplyAttr, err := attributes.AttributeScProcessEvent(ctx, sS.fltrS,
			sS.cfg.SessionSCfg().Conns[utils.MetaAttributes], sS.connMgr, utils.MetaSessionS, args)
		if err == nil {
			args = rplyAttr.CGREvent
			authReply.Attributes = rplyAttr
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrAttributeS(err)
		}
	}

	runEvents := make(map[string]*utils.CGREvent)
	var chrgS bool
	if chrgS, err = engine.GetBoolOpts(ctx, args.Tenant, dP, nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.Chargers,
		utils.MetaChargers); err != nil {
		return
	}
	if chrgS {
		var chrgrs []*chargers.ChrgSProcessEventReply
		if chrgrs, err = chargers.ChargerScProcessEvent(ctx, sS.fltrS,
			sS.cfg.SessionSCfg().Conns[utils.MetaChargers], sS.connMgr, sS.cache,
			utils.MetaSessionS, args); err != nil {
			return
		}
		for _, chrgr := range chrgrs {
			runEvents[chrgr.ChargerSProfile] = chrgr.CGREvent
		}
	} else {
		runEvents[utils.MetaRaw] = args
	}
	if acntS {
		var maxAbstracts map[string]*utils.Decimal
		if maxAbstracts, err = sS.accountSMaxAbstracts(ctx, runEvents); err != nil {
			return utils.NewErrAccountS(err)
		}
		authReply.MaxUsage = getMaxUsageFromRuns(maxAbstracts)
	}
	if resourceS {
		resSConns, errConn := engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns[utils.MetaResources], args.Tenant, dP, sS.fltrS)
		if errConn != nil {
			return errConn
		}
		if len(resSConns) == 0 {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		originID, _ := args.OptAsString(utils.MetaOriginID)
		if originID == "" {
			originID = utils.UUIDSha1Prefix()
		}
		args.APIOpts[utils.OptsResourcesUsageID] = originID
		args.APIOpts[utils.OptsResourcesUnits] = 1
		var allocMsg string
		if err = sS.connMgr.Call(ctx, resSConns, utils.ResourceSv1AuthorizeResources,
			args, &allocMsg); err != nil {
			return utils.NewErrResourceS(err)
		}
		authReply.ResourceAllocation = &allocMsg
	}
	if ipS {
		ipsConns, errConn := engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns[utils.MetaIPs], args.Tenant, dP, sS.fltrS)
		if errConn != nil {
			return errConn
		}
		if len(ipsConns) == 0 {
			return utils.NewErrNotConnected(utils.IPs)
		}
		originID, _ := args.OptAsString(utils.MetaOriginID)
		if originID == "" {
			originID = utils.UUIDSha1Prefix()
		}
		args.APIOpts[utils.OptsIPsAllocationID] = originID
		var allocIP utils.AllocatedIP
		if err = sS.connMgr.Call(ctx, ipsConns,
			utils.IPsV1AuthorizeIP, args, &allocIP); err != nil {
			return utils.NewErrIPs(err)
		}
		authReply.AllocatedIP = &allocIP
	}
	if routeS {
		routesReply, err := sS.getRoutes(ctx, args.Clone())
		if err != nil {
			return err
		}
		if routesReply != nil {
			authReply.RouteProfiles = routesReply
		}
	}

	var withErrors bool
	var thdS bool
	if thdS, err = engine.GetBoolOpts(ctx, args.Tenant, dP, nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.Thresholds,
		utils.MetaThresholds); err != nil {
		return
	}
	if thdS {
		tIDs, err := sS.processThreshold(ctx, args, true)
		if err != nil && err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with ThresholdS.",
					utils.SessionS, err.Error(), args))
			withErrors = true
		}
		authReply.ThresholdIDs = &tIDs
	}
	var stS bool
	if stS, err = engine.GetBoolOpts(ctx, args.Tenant, dP, nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.Stats,
		utils.MetaStats); err != nil {
		return
	}
	if stS {
		sIDs, err := sS.processStats(ctx, args, false)
		if err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with StatS.",
					utils.SessionS, err.Error(), args))
			withErrors = true
		}
		authReply.StatQueueIDs = &sIDs
	}
	if withErrors {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// BiRPCv1AuthorizeEventWithDigest performs authorization for CGREvent based on specific subsystem
// returning one level fields instead of multiple ones returned by BiRPCv1AuthorizeEvent
func (sS *SessionS) BiRPCv1AuthorizeEventWithDigest(ctx *context.Context,
	args *utils.CGREvent, authReply *V1AuthorizeReplyWithDigest) (err error) {
	var initAuthRply V1AuthorizeReply
	if err = sS.BiRPCv1AuthorizeEvent(ctx, args, &initAuthRply); err != nil {
		return
	}
	if initAuthRply.Attributes != nil && len(initAuthRply.Attributes.AlteredFields) != 0 {
		authReply.AttributesDigest = utils.StringPointer(initAuthRply.Attributes.Digest())
	}
	if initAuthRply.ResourceAllocation != nil && len(*initAuthRply.ResourceAllocation) != 0 {
		authReply.ResourceAllocation = initAuthRply.ResourceAllocation
	}
	if initAuthRply.MaxUsage != nil {
		maxDur, _ := initAuthRply.MaxUsage.Duration()
		authReply.MaxUsage = maxDur.Nanoseconds()
	}
	if initAuthRply.RouteProfiles != nil && len(initAuthRply.RouteProfiles) != 0 {
		authReply.RoutesDigest = utils.StringPointer(initAuthRply.RouteProfiles.Digest())
	}
	if initAuthRply.ThresholdIDs != nil && len(*initAuthRply.ThresholdIDs) != 0 {
		authReply.Thresholds = utils.StringPointer(
			strings.Join(*initAuthRply.ThresholdIDs, utils.FieldsSep))
	}
	if initAuthRply.StatQueueIDs != nil && len(*initAuthRply.StatQueueIDs) != 0 {
		authReply.StatQueues = utils.StringPointer(
			strings.Join(*initAuthRply.StatQueueIDs, utils.FieldsSep))
	}
	return
}

// BiRPCv1InitiateSession initiates a new session
func (sS *SessionS) BiRPCv1InitiateSession(ctx *context.Context,
	args *utils.CGREvent, rply *V1InitSessionReply) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if args.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if args.ID == "" {
		args.ID = utils.GenUUID()
	}
	if args.Tenant == "" {
		args.Tenant = sS.cfg.GeneralCfg().DefaultTenant
	}
	if args.APIOpts == nil {
		args.APIOpts = make(map[string]any)
	}

	// RPC caching
	if sS.cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.SessionSv1InitiateSession, args.ID)
		refID := guardian.Guardian.GuardIDs("",
			sS.cfg.GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)

		if itm, has := sS.cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*rply = *cachedResp.Result.(*V1InitSessionReply)
			}
			return cachedResp.Error
		}
		defer sS.cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: rply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	originID, err := args.OptAsString(utils.MetaOriginID)
	if err != nil {
		return err
	}
	if originID == "" {
		return utils.NewErrMandatoryIeMissing(utils.OriginID)
	}

	rply.MaxUsage = utils.DurationPointer(time.Duration(utils.InvalidUsage)) // temp

	dP := args.AsDataProvider()

	// TODO: accounting not yet functional for InitiateSession API.
	// var acntS bool
	// if acntS, err = engine.GetBoolOpts(ctx, args.Tenant, dP, nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.Accounts,
	// 	utils.MetaAccounts); err != nil {
	// 	return
	// }

	var initS bool
	if initS, err = engine.GetBoolOpts(ctx, args.Tenant, dP, nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.Initiate,
		utils.MetaInitiate); err != nil {
		return
	}
	var resourceS bool
	if resourceS, err = engine.GetBoolOpts(ctx, args.Tenant, dP, nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.ResourcesAllocate,
		utils.MetaResources); err != nil {
		return
	}
	var ipS bool
	if ipS, err = engine.GetBoolOpts(ctx, args.Tenant, dP, nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.IPsAllocate,
		utils.MetaIPs); err != nil {
		return
	}
	if !initS && !resourceS && !ipS {
		return // Nothing to do
	}

	var attrS bool
	if attrS, err = engine.GetBoolOpts(ctx, args.Tenant, dP, nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.Attributes,
		utils.MetaAttributes); err != nil {
		return
	}
	if attrS {
		rplyAttr, err := attributes.AttributeScProcessEvent(ctx, sS.fltrS,
			sS.cfg.SessionSCfg().Conns[utils.MetaAttributes], sS.connMgr, utils.MetaSessionS, args)
		if err == nil {
			args = rplyAttr.CGREvent
			rply.Attributes = rplyAttr
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrAttributeS(err)
		}
	}

	runEvents := make(map[string]*utils.CGREvent)

	var chrgS bool
	if chrgS, err = engine.GetBoolOpts(ctx, args.Tenant, dP, nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.Chargers,
		utils.MetaChargers); err != nil {
		return
	}
	if chrgS {
		var chrgrs []*chargers.ChrgSProcessEventReply
		if chrgrs, err = chargers.ChargerScProcessEvent(ctx, sS.fltrS,
			sS.cfg.SessionSCfg().Conns[utils.MetaChargers], sS.connMgr, sS.cache,
			utils.MetaSessionS, args); err != nil {
			return
		}
		for _, chrgr := range chrgrs {
			runEvents[chrgr.ChargerSProfile] = chrgr.CGREvent
		}
	} else {
		runEvents[utils.MetaRaw] = args
	}

	if resourceS {
		resSConns, errConn := engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns[utils.MetaResources], args.Tenant, dP, sS.fltrS)
		if errConn != nil {
			return errConn
		}
		if len(resSConns) == 0 {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		args.APIOpts[utils.OptsResourcesUsageID] = originID
		args.APIOpts[utils.OptsResourcesUnits] = 1
		var allocMessage string
		if err = sS.connMgr.Call(ctx, resSConns,
			utils.ResourceSv1AllocateResources, args, &allocMessage); err != nil {
			return utils.NewErrResourceS(err)
		}
		rply.ResourceAllocation = &allocMessage
		defer func() { // we need to release the resources back in case of errors
			if err != nil {
				var reply string
				if err = sS.connMgr.Call(ctx, resSConns, utils.ResourceSv1ReleaseResources,
					args, &reply); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> error: %s releasing resources for event %+v.",
							utils.SessionS, err.Error(), args))
				}
			}
		}()

	}
	if ipS {
		ipsConns, errConn := engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns[utils.MetaIPs], args.Tenant, dP, sS.fltrS)
		if errConn != nil {
			return errConn
		}
		if len(ipsConns) == 0 {
			return utils.NewErrNotConnected(utils.IPs)
		}
		args.APIOpts[utils.OptsIPsAllocationID] = originID
		var allocIP utils.AllocatedIP
		if err = sS.connMgr.Call(ctx, ipsConns,
			utils.IPsV1AllocateIP, args, &allocIP); err != nil {
			return utils.NewErrIPs(err)
		}
		rply.AllocatedIP = &allocIP
		defer func() { // we need to release the IPs back in case of errors
			if err != nil {
				var reply string
				if err = sS.connMgr.Call(ctx, ipsConns, utils.IPsV1ReleaseIP,
					args, &reply); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> error: %s releasing IPs for event %+v.",
							utils.SessionS, err.Error(), args))
				}
			}
		}()

	}
	if initS {
		var s *Session
		if s, err = sS.initSession(ctx, args, sS.biJClntID(ctx.Client), false); err != nil {
			return
		}

		var sRunsUsage map[string]time.Duration
		if sRunsUsage, err = sS.updateSession(ctx, s, nil, args.APIOpts, time.Duration(0)); err != nil {
			return //utils.NewErrRALs(err)
		}

		var maxUsage time.Duration
		var maxUsageSet bool // so we know if we have set the 0 on purpose
		for _, rplyMaxUsage := range sRunsUsage {
			if !maxUsageSet || rplyMaxUsage < maxUsage {
				maxUsage = rplyMaxUsage
				maxUsageSet = true
			}
		}
		rply.MaxUsage = &maxUsage

	}

	var withErrors bool
	var thdS bool
	if thdS, err = engine.GetBoolOpts(ctx, args.Tenant, dP, nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.Thresholds,
		utils.MetaThresholds); err != nil {
		return
	}
	if thdS {
		tIDs, err := sS.processThreshold(ctx, args, true)
		if err != nil && err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with ThresholdS.",
					utils.SessionS, err.Error(), args))
			withErrors = true
		}
		rply.ThresholdIDs = &tIDs
	}
	var stS bool
	if stS, err = engine.GetBoolOpts(ctx, args.Tenant, dP, nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.Stats,
		utils.MetaStats); err != nil {
		return
	}
	if stS {
		sIDs, err := sS.processStats(ctx, args, false)
		if err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with StatS.",
					utils.SessionS, err.Error(), args))
			withErrors = true
		}
		rply.StatQueueIDs = &sIDs
	}
	if withErrors {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// BiRPCv1UpdateSession updates an existing session, returning the duration which the session can still last
func (sS *SessionS) BiRPCv1UpdateSession(ctx *context.Context,
	args *utils.CGREvent, rply *V1UpdateSessionReply) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if args.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if args.ID == "" {
		args.ID = utils.GenUUID()
	}
	if args.Tenant == "" {
		args.Tenant = sS.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if sS.cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.SessionSv1UpdateSession, args.ID)
		refID := guardian.Guardian.GuardIDs("",
			sS.cfg.GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)

		if itm, has := sS.cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*rply = *cachedResp.Result.(*V1UpdateSessionReply)
			}
			return cachedResp.Error
		}
		defer sS.cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: rply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	dP := args.AsDataProvider()
	var attrS bool
	if attrS, err = engine.GetBoolOpts(ctx, args.Tenant, dP, nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.Attributes,
		utils.MetaAttributes); err != nil {
		return
	}
	var updS bool
	if updS, err = engine.GetBoolOpts(ctx, args.Tenant, dP, nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.Update,
		utils.MetaUpdate); err != nil {
		return
	}
	if !attrS && !updS {
		return // nothing to do
	}
	if args.APIOpts == nil {
		args.APIOpts = make(map[string]any)
	}

	if attrS {
		rplyAttr, err := attributes.AttributeScProcessEvent(ctx, sS.fltrS,
			sS.cfg.SessionSCfg().Conns[utils.MetaAttributes], sS.connMgr, utils.MetaSessionS, args)
		if err == nil {
			args = rplyAttr.CGREvent
			rply.Attributes = rplyAttr
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrAttributeS(err)
		}
	}
	if updS {
		var originID string
		if originID, err = args.OptAsString(utils.MetaOriginID); err != nil {
			return err
		}
		if originID == "" {
			return utils.NewErrMandatoryIeMissing(utils.OriginID)
		}

		s := sS.getActivateSession(originID)
		if s == nil {
			if s, err = sS.initSession(ctx, args, sS.biJClntID(ctx.Client), true); err != nil {
				return
			}
		}
		var sRunsUsage map[string]time.Duration
		if sRunsUsage, err = sS.updateSession(ctx, s, engine.MapEvent(args.Event), engine.MapEvent(args.APIOpts), time.Duration(0)); err != nil {
			return err
		}

		var maxUsage time.Duration
		var maxUsageSet bool // so we know if we have set the 0 on purpose
		for _, rplyMaxUsage := range sRunsUsage {
			if !maxUsageSet || rplyMaxUsage < maxUsage {
				maxUsage = rplyMaxUsage
				maxUsageSet = true
			}
		}
		rply.MaxUsage = &maxUsage
	}

	return
}

// BiRPCv1TerminateSession will stop debit loops as well as release any used resources
func (sS *SessionS) BiRPCv1TerminateSession(ctx *context.Context,
	args *utils.CGREvent, rply *string) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if args.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if args.ID == "" {
		args.ID = utils.GenUUID()
	}
	if args.Tenant == "" {
		args.Tenant = sS.cfg.GeneralCfg().DefaultTenant
	}
	// RPC caching
	if sS.cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.SessionSv1TerminateSession, args.ID)
		refID := guardian.Guardian.GuardIDs("",
			sS.cfg.GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)

		if itm, has := sS.cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*rply = *cachedResp.Result.(*string)
			}
			return cachedResp.Error
		}
		defer sS.cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: rply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	var withErrors bool
	dP := args.AsDataProvider()
	var ipsRelease bool
	if ipsRelease, err = engine.GetBoolOpts(ctx, args.Tenant, dP, nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.IPsRelease,
		utils.MetaIPs); err != nil {
		return
	}
	var resourcesRelease bool
	if resourcesRelease, err = engine.GetBoolOpts(ctx, args.Tenant, dP, nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.ResourcesRelease,
		utils.MetaResources); err != nil {
		return
	}
	var termS bool
	if termS, err = engine.GetBoolOpts(ctx, args.Tenant, dP, nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.Terminate,
		utils.MetaTerminate); err != nil {
		return
	}
	if !ipsRelease && !resourcesRelease && !termS {
		return // nothing to do
	}
	if args.APIOpts == nil {
		args.APIOpts = make(map[string]any)
	}

	originID, err := args.OptAsString(utils.MetaOriginID)
	if err != nil {
		return err
	}
	if originID == "" {
		return utils.NewErrMandatoryIeMissing(utils.OriginID)
	}

	if termS {
		var dbtItvl time.Duration

		ev := engine.MapEvent(args.Event)
		// cgrID := utils.Sha1(ev.GetStringIgnoreErrors(utils.OriginID),
		// 	ev.GetStringIgnoreErrors(utils.OriginHost))
		var s *Session
		fib := utils.FibDuration(time.Millisecond, 0)
		var isInstantEvent bool // one time charging, do not perform indexing and sTerminator

		for i := 0; i < sS.cfg.SessionSCfg().TerminateAttempts; i++ {
			if s = sS.getActivateSession(originID); s != nil {
				break
			}
			if i+1 < sS.cfg.SessionSCfg().TerminateAttempts { // not last iteration
				time.Sleep(fib())
				continue
			}
			isInstantEvent = true
			if s, err = sS.initSession(ctx, args, sS.biJClntID(ctx.Client), isInstantEvent); err != nil {
				return err
			}
			if _, err = sS.updateSession(ctx, s, nil, engine.MapEvent(args.APIOpts), dbtItvl); err != nil {
				return err
			}
			break
		}
		if !isInstantEvent {
			s.lk.Lock()
			s.updateSRuns(ev, sS.cfg.SessionSCfg().AlterableFields)
			s.lk.Unlock()
		}
		if err = sS.terminateSession(ctx, s,
			ev.GetDurationPtrIgnoreErrors(utils.Usage),
			ev.GetDurationPtrIgnoreErrors(utils.LastUsed),
			ev.GetTimePtrIgnoreErrors(utils.AnswerTime, ""),
			isInstantEvent); err != nil {
			return err
		}
	}
	if resourcesRelease {
		resSConns, errConn := engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns[utils.MetaResources], args.Tenant, dP, sS.fltrS)
		if errConn != nil {
			return errConn
		}
		if len(resSConns) == 0 {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		args.APIOpts[utils.OptsResourcesUsageID] = originID
		args.APIOpts[utils.OptsResourcesUnits] = 1
		var reply string
		if err = sS.connMgr.Call(ctx, resSConns, utils.ResourceSv1ReleaseResources,
			args, &reply); err != nil {
			return utils.NewErrResourceS(err)
		}
	}

	if ipsRelease {
		ipsConns, errConn := engine.GetConnIDs(ctx, sS.cfg.SessionSCfg().Conns[utils.MetaIPs], args.Tenant, dP, sS.fltrS)
		if errConn != nil {
			return errConn
		}
		if len(ipsConns) == 0 {
			return utils.NewErrNotConnected(utils.IPs)
		}
		args.APIOpts[utils.OptsIPsAllocationID] = originID
		var reply string
		if err = sS.connMgr.Call(ctx, ipsConns, utils.IPsV1ReleaseIP,
			args, &reply); err != nil {
			return utils.NewErrIPs(err)
		}
	}

	var thdS bool
	if thdS, err = engine.GetBoolOpts(ctx, args.Tenant, dP, nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.Thresholds,
		utils.MetaThresholds); err != nil {
		return
	}
	if thdS {
		_, err := sS.processThreshold(ctx, args, true)
		if err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with ThresholdS.",
					utils.SessionS, err.Error(), args))
			withErrors = true
		}
	}
	var stS bool
	if stS, err = engine.GetBoolOpts(ctx, args.Tenant, dP, nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.Stats,
		utils.MetaStats); err != nil {
		return
	}
	if stS {
		_, err := sS.processStats(ctx, args, false)
		if err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with StatS.",
					utils.SessionS, err.Error(), args))
			withErrors = true
		}
	}
	if withErrors {
		err = utils.ErrPartiallyExecuted
	}
	*rply = utils.OK
	return
}

// BiRPCv1ProcessCDR sends the CDR to CDRs
func (sS *SessionS) BiRPCv1ProcessCDR(ctx *context.Context,
	cgrEv *utils.CGREvent, rply *string) (err error) {
	if cgrEv.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if cgrEv.ID == "" {
		cgrEv.ID = utils.GenUUID()
	}
	if cgrEv.Tenant == "" {
		cgrEv.Tenant = sS.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if sS.cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.SessionSv1ProcessCDR, cgrEv.ID)
		refID := guardian.Guardian.GuardIDs("",
			sS.cfg.GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)

		if itm, has := sS.cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*rply = *cachedResp.Result.(*string)
			}
			return cachedResp.Error
		}
		defer sS.cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: rply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	return sS.processCDR(ctx, cgrEv, rply)
}

// BiRPCv1ProcessEvent processes an CGREvent with various subsystems
func (sS *SessionS) BiRPCv1ProcessEvent(ctx *context.Context,
	apiArgs *utils.CGREvent, apiRply *V1ProcessEventReply) (err error) {
	if apiArgs == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if apiArgs.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if apiArgs.APIOpts == nil {
		apiArgs.APIOpts = make(map[string]any)
	}
	if apiArgs.ID == "" {
		apiArgs.ID = utils.GenUUID()
	}
	if apiArgs.Tenant == "" {
		apiArgs.Tenant = sS.cfg.GeneralCfg().DefaultTenant
	}
	// RPC caching
	if sS.cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.SessionSv1AuthorizeEvent, apiArgs.ID)
		refID := guardian.Guardian.GuardIDs("",
			sS.cfg.GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)

		if itm, has := sS.cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*apiRply = *cachedResp.Result.(*V1ProcessEventReply)
			}
			return cachedResp.Error
		}
		defer sS.cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: apiRply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	cch := make(map[string]any) // cached opts

	// Only  *runID *primary can be derived
	if _, hasRunID := apiArgs.APIOpts[utils.MetaRunID]; !hasRunID {
		cch[utils.MetaRunID] = utils.MetaPrimary
		apiArgs.APIOpts[utils.MetaRunID] = utils.MetaPrimary // also save it for session creation with single run
	} else if optRunID, _ := apiArgs.OptAsString(utils.MetaRunID); optRunID != utils.EmptyString {
		cch[utils.MetaRunID] = optRunID
	}
	// Set cgrID of the event, most important for the session
	if cch[utils.MetaCGRid], err = engine.GetComputeCGRid(ctx, apiArgs, cch, sS.fltrS,
		sS.cfg.SessionSCfg().Opts.CGRid, sS.cfg.SessionSCfg().Opts.OriginID, sS.cfg.SessionSCfg().Opts.HostID); err != nil {
		return
	}
	apiArgs.APIOpts[utils.MetaCGRid] = cch[utils.MetaCGRid].(string)

	// processing AttributeS first gives us the opportunity of enhancing all the other flags
	// check for *attribute
	if attrS, errAttrS := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cch,
		sS.fltrS, sS.cfg.SessionSCfg().Opts.Attributes,
		utils.MetaAttributes); errAttrS != nil {
		return errAttrS
	} else {
		cch[utils.MetaAttributes] = attrS
	}
	if cch[utils.MetaAttributes].(bool) {
		if rplyAttr, errProc := attributes.AttributeScProcessEvent(ctx, sS.fltrS,
			sS.cfg.SessionSCfg().Conns[utils.MetaAttributes], sS.connMgr, utils.MetaSessionS, apiArgs); errProc != nil {
			if errProc.Error() != utils.ErrNotFound.Error() {
				return utils.NewErrAttributeS(errProc)
			}
		} else if len(rplyAttr.AlteredFields) != 0 { // at least one change was performed
			*apiArgs = *rplyAttr.CGREvent
			if apiRply.Attributes == nil {
				apiRply.Attributes = make(map[string]*attributes.ProcessEventReply)
			}
			apiRply.Attributes[utils.MetaPrimary] = rplyAttr
		}
	}

	cgrEvs := map[string]*utils.CGREvent{
		cch[utils.MetaRunID].(string): apiArgs,
	}

	// Set *interimUsage
	if interimUsage, errUsage := engine.GetDecimalBigOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cch,
		sS.fltrS, sS.cfg.SessionSCfg().Opts.InterimUsage, utils.MetaInterimUsage); errUsage != nil {
		return errUsage
	} else if interimUsage != nil && interimUsage.Cmp(decimal.New(0, 0)) == 1 { // >0
		cch[utils.MetaInterimUsage] = interimUsage
	}

	// Set *totalUsage
	if totalUsage, errUsage := engine.GetDecimalBigOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cch,
		sS.fltrS, sS.cfg.SessionSCfg().Opts.TotalUsage, utils.MetaTotalUsage); errUsage != nil {
		return errUsage
	} else if totalUsage != nil && totalUsage.Cmp(decimal.New(0, 0)) == 1 { // >0
		cch[utils.MetaTotalUsage] = totalUsage
	}

	// *session will set/add a session
	if sesBool, errBool := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cch,
		sS.fltrS, sS.cfg.SessionSCfg().Opts.Session,
		utils.MetaSession); errBool != nil {
		return errBool
	} else {
		cch[utils.MetaSession] = sesBool
	}
	var s *Session
	if cch[utils.MetaSession].(bool) {
		if s, err = sS.setSession(ctx, apiArgs, cch,
			sS.biJClntID(ctx.Client)); err != nil {
			return
		}
		cgrEvs = s.asCGREventsMap() // inherit session events to process
	}

	// extracting *terminate informs if the event should be attached to a session so we do not fork later
	if terminateBool, errBool := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cch,
		sS.fltrS, sS.cfg.SessionSCfg().Opts.Terminate,
		utils.MetaTerminate); errBool != nil {
		return errBool
	} else {
		cch[utils.MetaTerminate] = terminateBool
	}

	// ChargerS will multiply/alter the event before any auth/accounting/cdr taking place
	if chrgS, errChrg := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cch,
		sS.fltrS, sS.cfg.SessionSCfg().Opts.Chargers,
		utils.MetaChargers); errChrg != nil {
		return errChrg
	} else {
		cch[utils.MetaChargers] = chrgS
	}
	// Apply ChargerS, but only if *primary *runID
	if cch[utils.MetaChargers].(bool) &&
		cch[utils.MetaRunID].(string) == utils.MetaPrimary && len(cgrEvs) < 2 { // initial event, not inherited from Session
		var chrgrs []*chargers.ChrgSProcessEventReply
		if chrgrs, err = chargers.ChargerScProcessEvent(ctx, sS.fltrS,
			sS.cfg.SessionSCfg().Conns[utils.MetaChargers], sS.connMgr, sS.cache,
			utils.MetaSessionS, apiArgs); err != nil {
			return
		}
		delete(cgrEvs, utils.MetaPrimary) // it becomes entirely ChargerS responsibility to provide events
		if s != nil {
			s.lk.Lock()
			delete(s.sRuns, utils.MetaPrimary) // overwrite the primary event, empty chargers will mean no further charging applied
		}
		for _, chrgr := range chrgrs {
			runID := utils.IfaceAsString(chrgr.CGREvent.APIOpts[utils.MetaRunID]) // should be prepopulated always with check above
			if s != nil {                                                         // Append the SRuns
				s.sRuns[runID] = NewSRun(chrgr.CGREvent)
			}
			cgrEvs[runID] = chrgr.CGREvent
			if len(chrgr.AlteredFields) != len(chargers.ChargerSDefaultAlteredFields) {
				if apiRply.Attributes == nil {
					apiRply.Attributes = make(map[string]*attributes.ProcessEventReply)
				}
				apiRply.Attributes[runID] = &attributes.ProcessEventReply{
					AlteredFields: chrgr.AlteredFields,
					CGREvent:      chrgr.CGREvent,
				}
			}
		}
		if s != nil {
			s.lk.Unlock()
		}
	}

	if blkrErr, errBlkr := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(),
		cch, sS.fltrS, sS.cfg.SessionSCfg().Opts.BlockerError,
		utils.OptsSesBlockerError, utils.MetaBlockerErrorCfg); errBlkr != nil {
		return errBlkr
	} else {
		cch[utils.OptsSesBlockerError] = blkrErr
	}

	var withErrors bool // populate in case of non blocking errors

	// same processing for each event
	for runID, cgrEv := range cgrEvs {
		cchEv := make(map[string]any)

		// RouteS Enabled
		if rous, errRous := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.Routes,
			utils.MetaRoutes); errRous != nil {
			if cch[utils.OptsSesBlockerError].(bool) {
				return errRous
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errRous.Error(), cgrEv, utils.RouteS))
		} else if rous {
			var rous routes.SortedRoutesList
			if rous, err = sS.getRoutes(ctx, cgrEv); err != nil {
				if cch[utils.OptsSesBlockerError].(bool) {
					return
				}
				withErrors = true
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s processing event: %+v with %s",
						utils.SessionS, err.Error(), cgrEv, utils.RouteS))
			}
			if apiRply.RouteProfiles == nil {
				apiRply.RouteProfiles = make(map[string]routes.SortedRoutesList)
			}
			apiRply.RouteProfiles[runID] = rous
		}

		// StatS Enabled
		if sts, errSts := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.Stats,
			utils.MetaStats); errSts != nil {
			if cch[utils.OptsSesBlockerError].(bool) {
				return errSts
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errSts.Error(), cgrEv, utils.StatS))
		} else if sts {
			var statIDs []string
			if statIDs, err = sS.processStats(ctx, cgrEv, true); err != nil {
				if cch[utils.OptsSesBlockerError].(bool) {
					return
				}
				withErrors = true
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s processing event: %+v with %s",
						utils.SessionS, err.Error(), cgrEv, utils.StatS))
			}
			if apiRply.StatQueueIDs == nil {
				apiRply.StatQueueIDs = make(map[string][]string)
			}
			apiRply.StatQueueIDs[runID] = statIDs
		}

		// ThresholdS Enabled
		if thds, errThds := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.Thresholds,
			utils.MetaThresholds); errThds != nil {
			if cch[utils.OptsSesBlockerError].(bool) {
				return errThds
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errThds.Error(), cgrEv, utils.ThresholdS))
		} else if thds {
			var thdIDs []string
			if thdIDs, err = sS.processThreshold(ctx, cgrEv, true); err != nil {
				if cch[utils.OptsSesBlockerError].(bool) {
					return
				}
				withErrors = true
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s processing event: %+v with %s",
						utils.SessionS, err.Error(), cgrEv, utils.ThresholdS))
			}
			if apiRply.ThresholdIDs == nil {
				apiRply.ThresholdIDs = make(map[string][]string)
			}
			apiRply.ThresholdIDs[runID] = thdIDs
		}

		// IPs Enabled
		if ipS, errIPs := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.IPs,
			utils.MetaIPs); errIPs != nil {
			if cch[utils.OptsSesBlockerError].(bool) {
				return errIPs
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s authorizing event: %+v with %s",
					utils.SessionS, errIPs.Error(), cgrEv, utils.IPs))
			cchEv[utils.MetaIPs] = false
		} else {
			cchEv[utils.MetaIPs] = ipS
		}

		// RateS Enabled
		if rtS, errRTs := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.Rates,
			utils.MetaRates); errRTs != nil {
			if cch[utils.OptsSesBlockerError].(bool) {
				return errRTs
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errRTs.Error(), cgrEv, utils.RateS))
		} else if rtS {
			var rtsCost *utils.RateProfileCost
			if rtsCost, err = rates.RateScCostForEvent(ctx, sS.fltrS,
				sS.cfg.SessionSCfg().Conns[utils.MetaRates], sS.connMgr,
				utils.MetaSessionS, cgrEv); err != nil {
				if cch[utils.OptsSesBlockerError].(bool) {
					return
				}
				withErrors = true
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s processing event: %+v with %s",
						utils.SessionS, err.Error(), cgrEv, utils.RateS))
			} else {
				if apiRply.RateSCost == nil {
					apiRply.RateSCost = make(map[string]float64)
				}
				costFlt, _ := rtsCost.Cost.Float64()
				apiRply.RateSCost[runID] = costFlt
			}
		}

		// AccountS Enabled
		if acntS, errAcnts := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.Accounts,
			utils.MetaAccounts); errAcnts != nil {
			if cch[utils.OptsSesBlockerError].(bool) {
				return errAcnts
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errAcnts.Error(), cgrEv, utils.AccountS))
		} else {
			cchEv[utils.MetaAccounts] = acntS
		}

		// ResourceS Enabled
		if rscS, errRscS := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.Resources,
			utils.MetaResources); errRscS != nil {
			if cch[utils.OptsSesBlockerError].(bool) {
				return errRscS
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errRscS.Error(), cgrEv, utils.RouteS))
		} else {
			cchEv[utils.MetaResources] = rscS
		}

		// Auth the events
		if auth, errAuth := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(),
			cchEv, sS.fltrS, sS.cfg.SessionSCfg().Opts.Authorize,
			utils.MetaAuthorize); errAuth != nil {
			if cch[utils.OptsSesBlockerError].(bool) {
				return errAuth
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errAuth.Error(), cgrEv, utils.MetaAuthorize))
		} else {
			cchEv[utils.MetaAuthorize] = auth
		}

		// IPs Authorization
		if ipsAuthBool, errBool := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.IPsAuthorize,
			utils.MetaIPsAuthorizeCfg); errBool != nil {
			if cch[utils.OptsSesBlockerError].(bool) {
				return errBool
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errBool.Error(), cgrEv, utils.MetaIPsAuthorizeCfg))
		} else {
			cchEv[utils.MetaIPsAuthorizeCfg] = ipsAuthBool
		}
		if cchEv[utils.MetaIPsAuthorizeCfg].(bool) ||
			(cchEv[utils.MetaAuthorize].(bool) && cchEv[utils.MetaIPs].(bool)) {
			var authIP *utils.AllocatedIP
			if authIP, err = sS.ipsAuthorize(ctx, cgrEv); err != nil {
				if cch[utils.OptsSesBlockerError].(bool) {
					return
				}
				withErrors = true
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s processing event: %+v with %s",
						utils.SessionS, err.Error(), cgrEv, utils.IPs))
			}
			if apiRply.IPsAllocation == nil {
				apiRply.IPsAllocation = make(map[string]*utils.AllocatedIP)
			}
			apiRply.IPsAllocation[runID] = authIP
		}

		// ResourceS Authorization
		if resAuthBool, errBool := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.ResourcesAuthorize,
			utils.MetaResourcesAuthorizeCfg); errBool != nil {
			if cch[utils.OptsSesBlockerError].(bool) {
				return errBool
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errBool.Error(), cgrEv, utils.MetaResourcesAuthorizeCfg))
		} else {

			cchEv[utils.MetaResourcesAuthorizeCfg] = resAuthBool
		}
		if cchEv[utils.MetaResourcesAuthorizeCfg].(bool) ||
			(cchEv[utils.MetaAuthorize].(bool) && cchEv[utils.MetaResources].(bool)) {
			var resID string
			if resID, err = sS.resourcesAuthorize(ctx, cgrEv); err != nil {
				if cch[utils.OptsSesBlockerError].(bool) {
					return
				}
				withErrors = true
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s processing event: %+v with %s",
						utils.SessionS, err.Error(), cgrEv, utils.IPs))
			}
			if apiRply.ResourceAllocation == nil {
				apiRply.ResourceAllocation = make(map[string]string)
			}
			apiRply.ResourceAllocation[runID] = resID
		}

		// AccountS Authorization
		if acntsAuthBool, errBool := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.AccountsAuthorize,
			utils.MetaAccountsAuthorizeCfg); errBool != nil {
			if cch[utils.OptsSesBlockerError].(bool) {
				return errBool
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errBool.Error(), cgrEv, utils.MetaAccountsAuthorizeCfg))
		} else {
			cchEv[utils.MetaAccountsAuthorizeCfg] = acntsAuthBool
		}
		if !cchEv[utils.MetaAccountsAuthorizeCfg].(bool) &&
			cchEv[utils.MetaAccounts].(bool) && cchEv[utils.MetaAuthorize].(bool) {
			cchEv[utils.MetaAccountsAuthorizeCfg] = true // mark it so we can use it later
		}
		if cchEv[utils.MetaAccountsAuthorizeCfg].(bool) {
			var acntCost *utils.EventCharges
			if acntCost, err = sS.accountsMaxAbstracts(ctx, cgrEv); err != nil {
				if cch[utils.OptsSesBlockerError].(bool) {
					return
				}
				withErrors = true
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s processing event: %+v with %s for MaxAbstracts",
						utils.SessionS, err.Error(), cgrEv, utils.AccountS))
			}
			maxDur, _ := acntCost.Abstracts.Duration()
			if apiRply.AccountSUsage == nil {
				apiRply.AccountSUsage = make(map[string]time.Duration)
			}
			apiRply.AccountSUsage[runID] = maxDur
		}
		// AccountS Debit
		if acntsDebitBool, errBool := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.AccountsDebit,
			utils.MetaAccountsDebitCfg); errBool != nil {
			if cch[utils.OptsSesBlockerError].(bool) {
				return errBool
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errBool.Error(), cgrEv, utils.MetaAccountsDebitCfg))
		} else {
			cchEv[utils.MetaAccountsDebitCfg] = acntsDebitBool
		}
		if cchEv[utils.MetaAccountsDebitCfg].(bool) {
			var acntCost *utils.EventCharges
			if acntCost, err = sS.accountSDebitEvent(ctx, cgrEv, s); err != nil {
				if cch[utils.OptsSesBlockerError].(bool) {
					return
				}
				withErrors = true
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s processing event: %+v with %s for Debit",
						utils.SessionS, err.Error(), cgrEv, utils.AccountS))
			}
			maxDur, _ := acntCost.Abstracts.Duration()
			if apiRply.AccountSUsage == nil {
				apiRply.AccountSUsage = make(map[string]time.Duration)
			}
			apiRply.AccountSUsage[runID] = maxDur
		}
		// UsageRecords generation
		if ees, errEEs := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.EEs,
			utils.MetaEEs); errEEs != nil {
			if cch[utils.OptsSesBlockerError].(bool) {
				return errEEs
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errEEs.Error(), cgrEv, utils.EEs))
		} else if ees {
			var eesIDs []string
			if eesIDs, err = sS.eesProcessEvent(ctx, cgrEv); err != nil {
				if cch[utils.OptsSesBlockerError].(bool) {
					return
				}
				withErrors = true
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s processing event: %+v with %s",
						utils.SessionS, err.Error(), cgrEv, utils.EEs))
			}
			if apiRply.EventExporters == nil {
				apiRply.EventExporters = make(map[string][]string)
			}
			apiRply.EventExporters[runID] = eesIDs
		}
	}
	if withErrors {
		err = utils.ErrPartiallyExecuted
	}
	return
}
