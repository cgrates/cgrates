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
package sessions

import (
	"fmt"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
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
	var withErrors bool
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
	if acntS, err = engine.GetBoolOpts(ctx, args.Tenant, dP, nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.MaxUsage,
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
	if args.APIOpts == nil {
		args.APIOpts = make(map[string]any)
	}
	if attrS {
		if args.APIOpts == nil {
			args.APIOpts = make(map[string]any)
		}
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
		if maxAbstracts, err = sS.accounSMaxAbstracts(ctx, runEvents); err != nil {
			return utils.NewErrAccountS(err)
		}
		authReply.MaxUsage = getMaxUsageFromRuns(maxAbstracts)
	}
	if resourceS {
		if len(sS.cfg.SessionSCfg().ResourceSConns) == 0 {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		originID, _ := args.FieldAsString(utils.OriginID)
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
		originID, _ := args.FieldAsString(utils.OriginID)
		if originID == "" {
			originID = utils.UUIDSha1Prefix()
		}
		args.APIOpts[utils.OptsIPsAllocationID] = originID
		var allocIP utils.AllocatedIP
		if err = sS.connMgr.Call(ctx, sS.cfg.SessionSCfg().IPsConns, utils.IPsV1AuthorizeIP,
			args, &allocIP); err != nil {
			return utils.NewErrIPs(err)
		}
		authReply.IPAllocation = &allocIP.Message
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
	var withErrors bool
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

	var originID string
	dP := args.AsDataProvider()
	if originID, err = engine.GetStringOpts(ctx, args.Tenant, dP, nil, sS.fltrS, nil,
		utils.MetaOriginID); err != nil {
		return
	} else if originID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.OriginID)
	}
	if _, has := args.APIOpts[utils.MetaOriginID]; !has {
		args.APIOpts[utils.MetaOriginID] = originID
	}

	rply.MaxUsage = utils.DurationPointer(time.Duration(utils.InvalidUsage)) // temp

	var acntS bool
	if acntS, err = engine.GetBoolOpts(ctx, args.Tenant, dP, nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.MaxUsage,
		utils.MetaAccounts); err != nil {
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
	if !acntS && !resourceS && !ipS {
		return // Nothing to do
	}

	var attrS bool
	if attrS, err = engine.GetBoolOpts(ctx, args.Tenant, dP, nil, sS.fltrS, sS.cfg.SessionSCfg().Opts.Attributes,
		utils.MetaAttributes); err != nil {
		return
	}
	if attrS {
		if args.APIOpts == nil {
			args.APIOpts = make(map[string]any)
		}
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
		if originID == utils.EmptyString {
			return utils.NewErrMandatoryIeMissing(utils.OriginID)
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
		if originID == utils.EmptyString {
			return utils.NewErrMandatoryIeMissing(utils.OriginID)
		}
		args.APIOpts[utils.OptsIPsAllocationID] = originID
		var allocIP utils.AllocatedIP
		if err = sS.connMgr.Call(ctx, sS.cfg.SessionSCfg().IPsConns,
			utils.IPsV1AllocateIP, args, &allocIP); err != nil {
			return utils.NewErrIPs(err)
		}
		rply.IPAllocation = &allocIP.Message
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
	if acntS {

		var s *Session
		if s, err = sS.initSession(ctx, args, sS.biJClntID(ctx.Client), true); err != nil {
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

// BiRPCv1ProcessCDR sends the CDR to CDRs
func (sS *SessionS) BiRPCv1ProcessCDR(ctx *context.Context,
	cgrEv *utils.CGREvent, rply *string) (err error) {
	if cgrEv.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if cgrEv.ID == utils.EmptyString {
		cgrEv.ID = utils.GenUUID()
	}
	if cgrEv.Tenant == utils.EmptyString {
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
