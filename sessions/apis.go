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
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
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

		if itm, has := engine.Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*authReply = *cachedResp.Result.(*V1AuthorizeReply)
			}
			return cachedResp.Error
		}
		defer engine.Cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
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
		rplyAttr, err := sS.processAttributes(ctx, args)
		if err == nil {
			args = rplyAttr.CGREvent
			authReply.Attributes = &rplyAttr
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
		if chrgrs, err = sS.processChargerS(ctx, args); err != nil {
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
		if len(sS.cfg.SessionSCfg().ResourceSConns) == 0 {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		originID, _ := args.OptAsString(utils.MetaOriginID)
		if originID == "" {
			originID = utils.UUIDSha1Prefix()
		}
		args.APIOpts[utils.OptsResourcesUsageID] = originID
		args.APIOpts[utils.OptsResourcesUnits] = 1
		var allocMsg string
		if err = sS.connMgr.Call(ctx, sS.cfg.SessionSCfg().ResourceSConns, utils.ResourceSv1AuthorizeResources,
			args, &allocMsg); err != nil {
			return utils.NewErrResourceS(err)
		}
		authReply.ResourceAllocation = &allocMsg
	}
	if ipS {
		if len(sS.cfg.SessionSCfg().IPsConns) == 0 {
			return utils.NewErrNotConnected(utils.IPs)
		}
		originID, _ := args.OptAsString(utils.MetaOriginID)
		if originID == "" {
			originID = utils.UUIDSha1Prefix()
		}
		args.APIOpts[utils.OptsIPsAllocationID] = originID
		var allocIP utils.AllocatedIP
		if err = sS.connMgr.Call(ctx, sS.cfg.SessionSCfg().IPsConns,
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

		if itm, has := engine.Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*rply = *cachedResp.Result.(*V1InitSessionReply)
			}
			return cachedResp.Error
		}
		defer engine.Cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
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
		rplyAttr, err := sS.processAttributes(ctx, args)
		if err == nil {
			args = rplyAttr.CGREvent
			rply.Attributes = &rplyAttr
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
		if chrgrs, err = sS.processChargerS(ctx, args); err != nil {
			return
		}
		for _, chrgr := range chrgrs {
			runEvents[chrgr.ChargerSProfile] = chrgr.CGREvent
		}
	} else {
		runEvents[utils.MetaRaw] = args
	}

	if resourceS {
		if len(sS.cfg.SessionSCfg().ResourceSConns) == 0 {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		args.APIOpts[utils.OptsResourcesUsageID] = originID
		args.APIOpts[utils.OptsResourcesUnits] = 1
		var allocMessage string
		if err = sS.connMgr.Call(ctx, sS.cfg.SessionSCfg().ResourceSConns,
			utils.ResourceSv1AllocateResources, args, &allocMessage); err != nil {
			return utils.NewErrResourceS(err)
		}
		rply.ResourceAllocation = &allocMessage
		defer func() { // we need to release the resources back in case of errors
			if err != nil {
				var reply string
				if err = sS.connMgr.Call(ctx, sS.cfg.SessionSCfg().ResourceSConns, utils.ResourceSv1ReleaseResources,
					args, &reply); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> error: %s releasing resources for event %+v.",
							utils.SessionS, err.Error(), args))
				}
			}
		}()

	}
	if ipS {
		if len(sS.cfg.SessionSCfg().IPsConns) == 0 {
			return utils.NewErrNotConnected(utils.IPs)
		}
		args.APIOpts[utils.OptsIPsAllocationID] = originID
		var allocIP utils.AllocatedIP
		if err = sS.connMgr.Call(ctx, sS.cfg.SessionSCfg().IPsConns,
			utils.IPsV1AllocateIP, args, &allocIP); err != nil {
			return utils.NewErrIPs(err)
		}
		rply.AllocatedIP = &allocIP
		defer func() { // we need to release the IPs back in case of errors
			if err != nil {
				var reply string
				if err = sS.connMgr.Call(ctx, sS.cfg.SessionSCfg().IPsConns, utils.IPsV1ReleaseIP,
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
		var dbtItvl time.Duration
		if dbtItvl, err = engine.GetDurationOpts(ctx, args.Tenant, args.AsDataProvider(), nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.DebitInterval,
			utils.OptsSesDebitInterval); err != nil {
			return
		}
		if dbtItvl > 0 { //active debit
			rply.MaxUsage = utils.DurationPointer(sS.cfg.SessionSCfg().GetDefaultUsage(utils.IfaceAsString(args.Event[utils.ToR])))
		} else {
			var sRunsUsage map[string]time.Duration
			if sRunsUsage, err = sS.updateSession(ctx, s, nil, args.APIOpts, dbtItvl); err != nil {
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

		if itm, has := engine.Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*rply = *cachedResp.Result.(*V1UpdateSessionReply)
			}
			return cachedResp.Error
		}
		defer engine.Cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
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
		rplyAttr, err := sS.processAttributes(ctx, args)
		if err == nil {
			args = rplyAttr.CGREvent
			rply.Attributes = &rplyAttr
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

		var dbtItvl time.Duration
		if dbtItvl, err = engine.GetDurationOpts(ctx, args.Tenant, args.AsDataProvider(), nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.DebitInterval,
			utils.OptsSesDebitInterval); err != nil {
			return err
		}

		s := sS.getActivateSession(originID)
		if s == nil {
			if s, err = sS.initSession(ctx, args, sS.biJClntID(ctx.Client), true); err != nil {
				return
			}
		}
		var sRunsUsage map[string]time.Duration
		if sRunsUsage, err = sS.updateSession(ctx, s, engine.MapEvent(args.Event), engine.MapEvent(args.APIOpts), dbtItvl); err != nil {
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

		if itm, has := engine.Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*rply = *cachedResp.Result.(*string)
			}
			return cachedResp.Error
		}
		defer engine.Cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
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
		if dbtItvl, err = engine.GetDurationOpts(ctx, args.Tenant, args.AsDataProvider(), nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.DebitInterval,
			utils.OptsSesDebitInterval); err != nil {
			return err
		}

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
		if len(sS.cfg.SessionSCfg().ResourceSConns) == 0 {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		args.APIOpts[utils.OptsResourcesUsageID] = originID
		args.APIOpts[utils.OptsResourcesUnits] = 1
		var reply string
		if err = sS.connMgr.Call(ctx, sS.cfg.SessionSCfg().ResourceSConns, utils.ResourceSv1ReleaseResources,
			args, &reply); err != nil {
			return utils.NewErrResourceS(err)
		}
	}

	if ipsRelease {
		if len(sS.cfg.SessionSCfg().IPsConns) == 0 {
			return utils.NewErrNotConnected(utils.IPs)
		}
		args.APIOpts[utils.OptsIPsAllocationID] = originID
		var reply string
		if err = sS.connMgr.Call(ctx, sS.cfg.SessionSCfg().IPsConns, utils.IPsV1ReleaseIP,
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

		if itm, has := engine.Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*rply = *cachedResp.Result.(*string)
			}
			return cachedResp.Error
		}
		defer engine.Cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
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

		if itm, has := engine.Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*apiRply = *cachedResp.Result.(*V1ProcessEventReply)
			}
			return cachedResp.Error
		}
		defer engine.Cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: apiRply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	cgrEvs := map[string]*utils.CGREvent{
		utils.MetaDefault: apiArgs,
	}

	cch := make(map[string]any)

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
		apiRply.Attributes = make(map[string]*attributes.AttrSProcessEventReply)
		if rplyAttr, errProc := sS.processAttributes(ctx, apiArgs); errProc != nil {
			if errProc.Error() != utils.ErrNotFound.Error() {
				return utils.NewErrAttributeS(errProc)
			}
		} else {
			*apiArgs = *rplyAttr.CGREvent
			apiRply.Attributes[utils.MetaDefault] = &rplyAttr
		}
	}

	// ChargerS will multiply/alter the event before any auth/accounting/cdr taking place
	if chrgS, errChrg := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cch,
		sS.fltrS, sS.cfg.SessionSCfg().Opts.Chargers,
		utils.MetaChargers); errChrg != nil {
		return errChrg
	} else {
		cch[utils.MetaChargers] = chrgS
	}
	if cch[utils.MetaChargers].(bool) {
		delete(cgrEvs, utils.MetaDefault) // ChargerS becomes responsive of charging
		var chrgrs []*chargers.ChrgSProcessEventReply
		if chrgrs, err = sS.processChargerS(ctx, apiArgs); err != nil {
			return
		}
		for _, chrgr := range chrgrs {
			cgrEvs[utils.IfaceAsString(chrgr.CGREvent.APIOpts[utils.MetaRunID])] = chrgr.CGREvent
		}
	}

	//var partiallyExecuted bool // will be	 added to the final answer if true
	if blkrErr, errBlkr := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(),
		cch, sS.fltrS, sS.cfg.SessionSCfg().Opts.Authorize,
		utils.OptsSesBlockerError, utils.MetaBlockerErrorCfg); errBlkr != nil {
		return errBlkr
	} else {
		cch[utils.OptsSesBlockerError] = blkrErr
	}

	// same processing for each event
	for runID, cgrEv := range cgrEvs {
		cchEv := make(map[string]any)

		// RateS Enabled
		if rtS, errRTs := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.Rates,
			utils.MetaRates); errRTs != nil {
			if cch[utils.OptsSesBlockerError].(bool) {
				return errRTs
			}
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v with %s",
					utils.SessionS, err.Error(), cgrEv, utils.RateS))
		} else if rtS {
			var rtsCost *utils.Decimal
			if rtsCost, err = sS.ratesCost(ctx, cgrEv); err != nil {
				return
			}
			if apiRply.RateSCost == nil {
				apiRply.RateSCost = make(map[string]float64)
			}
			costFlt, _ := rtsCost.Float64()
			apiRply.RateSCost[runID] = costFlt
		}

		// IPs Enabled
		if ipS, errIPs := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.IPs,
			utils.MetaIPs); errIPs != nil {
			if cch[utils.OptsSesBlockerError].(bool) {
				return errIPs
			}
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s authorizing event: %+v with %s",
					utils.SessionS, err.Error(), cgrEv, utils.IPs))
		} else {
			cchEv[utils.MetaIPs] = ipS
		}

		// AccountS Enabled
		if acntS, errAcnts := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.Accounts,
			utils.MetaAccounts); errAcnts != nil {
			return errAcnts
		} else {
			cchEv[utils.MetaAccounts] = acntS
		}

		// ResourceS Enabled
		if rscS, errRscS := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.Resources,
			utils.MetaResources); errRscS != nil {
			return errRscS
		} else {
			cchEv[utils.MetaResources] = rscS
		}

		// Auth the events
		if auth, errAuth := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(),
			cchEv, sS.fltrS, sS.cfg.SessionSCfg().Opts.IPsAuthorize,
			utils.MetaAuthorize); errAuth != nil {
			return errAuth
		} else {
			cchEv[utils.MetaAuthorize] = auth
		}

		// IPs Authorization
		if ipsAuthBool, errBool := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.IPsAuthorize,
			utils.MetaIPsAuthorizeCfg); errBool != nil {
			return errBool
		} else {
			cchEv[utils.MetaIPsAuthorizeCfg] = ipsAuthBool
		}
		if cchEv[utils.MetaIPsAuthorizeCfg].(bool) ||
			(cchEv[utils.MetaAuthorize].(bool) && cchEv[utils.MetaIPs].(bool)) {
			var authIP *utils.AllocatedIP
			if authIP, err = sS.ipsAuthorize(ctx, cgrEv); err != nil {
				return
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
			return errBool
		} else {
			cchEv[utils.MetaResourcesAuthorizeCfg] = resAuthBool
		}
		if cchEv[utils.MetaResourcesAuthorizeCfg].(bool) ||
			(cchEv[utils.MetaAuthorize].(bool) && cchEv[utils.MetaResources].(bool)) {
			var resID string
			if resID, err = sS.resourcesAuthorize(ctx, cgrEv); err != nil {
				return
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
			return errBool
		} else {
			cchEv[utils.MetaAccountsAuthorizeCfg] = acntsAuthBool
		}
		if cchEv[utils.MetaAccountsAuthorizeCfg].(bool) ||
			(cchEv[utils.MetaAuthorize].(bool) && cchEv[utils.MetaAccounts].(bool)) {
			var acntCost *utils.EventCharges
			if acntCost, err = sS.accountsMaxAbstracts(ctx, cgrEv); err != nil {
				return
			}
			maxDur, _ := acntCost.Abstracts.Duration()
			apiRply.AccountSUsage[runID] = maxDur
		}

	}
	return
}
