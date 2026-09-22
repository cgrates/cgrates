// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package sessions

import (
	"fmt"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/attributes"
	"github.com/cgrates/cgrates/chargers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

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
		cacheKey := utils.ConcatenatedKey(utils.SessionSv1ProcessEvent, apiArgs.ID)
		unlock := sS.cache.LockRPCResponse(cacheKey) // RPC caching needs to be atomic
		defer unlock()

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
	if utils.OptAsBool(cch, utils.MetaAttributes) {
		if rplyAttr, errProc := attributes.AttributeScProcessEvent(ctx, sS.fltrS,
			sS.cfg.SessionSCfg().Conns, sS.connMgr, utils.MetaSessionS, apiArgs); errProc != nil {
			if errProc.Error() != utils.ErrNotFound.Error() {
				return utils.NewErrAttributeS(errProc)
			}
		} else if len(rplyAttr.AlteredFields) != 0 { // at least one change was performed
			*apiArgs = *rplyAttr.CGREvent
			if apiRply.Attributes == nil {
				apiRply.Attributes = make(map[string]*utils.AttributesProcessEventReply)
			}
			apiRply.Attributes[utils.MetaPrimary] = rplyAttr
		}
	}

	cgrEvs := map[string]*utils.CGREvent{
		cch[utils.MetaRunID].(string): apiArgs,
	}

	// Set *interimConsumed
	if interimConsumed, errUsage := engine.GetDecimalOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cch,
		sS.fltrS, sS.cfg.SessionSCfg().Opts.InterimConsumed, utils.MetaInterimConsumed); errUsage != nil {
		return errUsage
	} else if interimConsumed != nil {
		cch[utils.MetaInterimConsumed] = interimConsumed
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
	if utils.OptAsBool(cch, utils.MetaSession) {
		if s, err = sS.setSession(ctx, apiArgs, cch,
			sS.biJClntID(ctx.Client)); err != nil {
			return
		}
	}
	// extracting *terminate
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
	if utils.OptAsBool(cch, utils.MetaChargers) &&
		utils.IfaceAsString(cch[utils.MetaRunID]) == utils.MetaPrimary &&
		(s == nil || len(s.sRuns) == 0) { // initial event, sRuns not yet initialized
		var chrgrs []*chargers.ChrgSProcessEventReply
		if chrgrs, err = chargers.ChargerScProcessEvent(ctx, sS.fltrS,
			sS.cfg.SessionSCfg().Conns, sS.connMgr, sS.cache,
			utils.MetaSessionS, apiArgs); err != nil {
			return
		}
		delete(cgrEvs, utils.MetaPrimary) // it becomes entirely ChargerS responsibility to provide events
		if s != nil {
			s.lk.Lock()
			delete(s.sRuns, utils.MetaPrimary) // overwrite the primary event, empty chargers will mean no further charging applied
			s.lk.Unlock()
		}
		for _, chrgr := range chrgrs {
			runID := utils.IfaceAsString(chrgr.CGREvent.APIOpts[utils.MetaRunID]) // should be prepopulated always with check above
			cgrEvs[runID] = chrgr.CGREvent
			if len(chrgr.AlteredFields) != len(chargers.ChargerSDefaultAlteredFields) {
				if apiRply.Attributes == nil {
					apiRply.Attributes = make(map[string]*utils.AttributesProcessEventReply)
				}
				apiRply.Attributes[runID] = &utils.AttributesProcessEventReply{
					AlteredFields: chrgr.AlteredFields,
					CGREvent:      chrgr.CGREvent,
				}
			}
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

		// Set *usage
		if usage, errUsage := engine.GetDecimalOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cch,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.Usage, utils.MetaUsage); errUsage != nil {
			return errUsage
		} else if usage != nil {
			apiArgs.APIOpts[utils.MetaUsage] = usage // populated for Event at least
			cchEv[utils.MetaUsage] = usage
			cchEv[utils.MetaInterimUsage] = usage // interimUsage defaults to usage from event
		}

		// Set *interimConsumed
		var interimConsumed *utils.Decimal
		if interimConsumed, err = engine.GetDecimalOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cch,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.InterimConsumed, utils.MetaInterimConsumed); err != nil {
			return
		} else if interimConsumed != nil {
			cch[utils.MetaInterimConsumed] = interimConsumed
		}

		// Set *interimUsage
		var interimUsage *utils.Decimal
		if interimUsage, err = engine.GetDecimalOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cch,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.InterimUsage, utils.MetaInterimUsage, utils.MetaUsage); err != nil {
			return
		} else if interimUsage != nil {
			cchEv[utils.MetaInterimUsage] = interimUsage
		}

		// Set *totalUsage
		var totalUsage *utils.Decimal
		if totalUsage, err = engine.GetDecimalOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cch,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.TotalUsage, utils.MetaTotalUsage); err != nil {
			return
		} else if totalUsage != nil {
			cchEv[utils.MetaTotalUsage] = totalUsage
		}

		// setSRun
		if s != nil {
			s.lk.Lock()
			if has, errSet := s.setSRun(runID, cgrEv, sS.cfg.SessionSCfg().AlterableFields, cchEv,
				interimConsumed, interimUsage, totalUsage); errSet != nil {
				if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
					return errSet
				}
				withErrors = true
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s processing event: %+v for SRun set",
						utils.SessionS, errSet.Error(), cgrEv))
			} else if !has {
				sr := s.sRuns[runID] // index the new SRun
				sS.indexSRuns(s.ID, false, sr)
			}

			s.lk.Unlock()
		}

		// RouteS Enabled
		if rous, errRous := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.Routes,
			utils.MetaRoutes); errRous != nil {
			if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
				return errRous
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errRous.Error(), cgrEv, utils.RouteS))
		} else if rous {
			var rous utils.SortedRoutesList
			if rous, err = sS.getRoutes(ctx, cgrEv); err != nil {
				if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
					return
				}
				withErrors = true
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s processing event: %+v with %s",
						utils.SessionS, err.Error(), cgrEv, utils.RouteS))
			}
			if apiRply.RouteProfiles == nil {
				apiRply.RouteProfiles = make(map[string]utils.SortedRoutesList)
			}
			apiRply.RouteProfiles[runID] = rous
		}

		// StatS Enabled
		if sts, errSts := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.Stats,
			utils.MetaStats); errSts != nil {
			if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
				return errSts
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errSts.Error(), cgrEv, utils.StatS))
		} else if sts {
			var statIDs []string
			if statIDs, err = sS.processStats(ctx, cgrEv, true); err != nil {
				if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
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
			if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
				return errThds
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errThds.Error(), cgrEv, utils.ThresholdS))
		} else if thds {
			var thdIDs []string
			if thdIDs, err = sS.processThreshold(ctx, cgrEv, true); err != nil {
				if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
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
			if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
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
			if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
				return errRTs
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errRTs.Error(), cgrEv, utils.RateS))
		} else if rtS {
			var rtsCost *utils.RateProfileCost
			if rtsCost, err = sS.ratesCost(ctx, cgrEv); err != nil {
				if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
					return
				}
				withErrors = true
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s processing event: %+v with %s",
						utils.SessionS, err.Error(), cgrEv, utils.RateS))
			} else {
				if apiRply.RatesCost == nil {
					apiRply.RatesCost = make(map[string]float64)
				}
				costFlt, _ := rtsCost.Cost.Float64()
				apiRply.RatesCost[runID] = costFlt
				if s == nil || utils.OptAsBool(cch, utils.MetaTerminate) {
					cgrEv.APIOpts[utils.MetaRatesCost] = rtsCost
				}
			}
		}

		// AccountS Enabled
		if acntS, errAcnts := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.Accounts,
			utils.MetaAccounts); errAcnts != nil {
			if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
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
			if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
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
			if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
				return errAuth
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errAuth.Error(), cgrEv, utils.MetaAuthorize))
		} else {
			cchEv[utils.MetaAuthorize] = auth
		}

		// Refund
		if rfnd, errBool := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(),
			cchEv, sS.fltrS, sS.cfg.SessionSCfg().Opts.Refund,
			utils.MetaRefund); errBool != nil {
			if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
				return errBool
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errBool.Error(), cgrEv, utils.MetaRefund))
		} else {
			cchEv[utils.MetaRefund] = rfnd
		}

		// Debit
		if debit, errBool := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(),
			cchEv, sS.fltrS, sS.cfg.SessionSCfg().Opts.Debit,
			utils.MetaDebit); errBool != nil {
			if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
				return errBool
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errBool.Error(), cgrEv, utils.MetaDebit))
		} else {
			cchEv[utils.MetaDebit] = debit
		}

		// IPs Authorization
		if ipsAuthBool, errBool := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.IPsAuthorize,
			utils.MetaIPsAuthorizeCfg); errBool != nil {
			if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
				return errBool
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errBool.Error(), cgrEv, utils.MetaIPsAuthorizeCfg))
		} else {
			cchEv[utils.MetaIPsAuthorizeCfg] = ipsAuthBool
		}
		if utils.OptAsBool(cchEv, utils.MetaIPsAuthorizeCfg) ||
			(utils.OptAsBool(cchEv, utils.MetaAuthorize) && utils.OptAsBool(cchEv, utils.MetaIPs)) {
			var authIP *utils.AllocatedIP
			if authIP, err = sS.ipsAuthorize(ctx, cgrEv); err != nil {
				if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
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

		// IPs Allocation
		if ipsAllocBool, errBool := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.IPsAllocate,
			utils.MetaIPsAllocateCfg); errBool != nil {
			if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
				return errBool
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errBool.Error(), cgrEv, utils.MetaIPsAllocateCfg))
		} else {
			cchEv[utils.MetaIPsAllocateCfg] = ipsAllocBool
		}
		if utils.OptAsBool(cchEv, utils.MetaIPsAllocateCfg) {
			var allocatedIP *utils.AllocatedIP
			if allocatedIP, err = sS.ipsAllocate(ctx, cgrEv); err != nil {
				if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
					return err
				}
				withErrors = true
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s processing event: %+v with %s",
						utils.SessionS, err.Error(), cgrEv, utils.IPs))
			}
			if apiRply.IPsAllocation == nil {
				apiRply.IPsAllocation = make(map[string]*utils.AllocatedIP)
			}
			apiRply.IPsAllocation[runID] = allocatedIP
		}

		// IPs Release
		if ipsReleaseBool, errBool := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.IPsRelease,
			utils.MetaIPsReleaseCfg); errBool != nil {
			if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
				return errBool
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errBool.Error(), cgrEv, utils.MetaIPsReleaseCfg))
		} else {
			cchEv[utils.MetaIPsReleaseCfg] = ipsReleaseBool
		}
		if utils.OptAsBool(cchEv, utils.MetaIPsReleaseCfg) {
			if err = sS.ipsRelease(ctx, cgrEv); err != nil {
				if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
					return err
				}
				withErrors = true
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s processing event: %+v with %s",
						utils.SessionS, err.Error(), cgrEv, utils.IPs))
			}
		}

		// ResourceS Authorization
		if resAuthBool, errBool := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.ResourcesAuthorize,
			utils.MetaResourcesAuthorizeCfg); errBool != nil {
			if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
				return errBool
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errBool.Error(), cgrEv, utils.MetaResourcesAuthorizeCfg))
		} else {
			cchEv[utils.MetaResourcesAuthorizeCfg] = resAuthBool
		}
		if utils.OptAsBool(cchEv, utils.MetaResourcesAuthorizeCfg) ||
			(utils.OptAsBool(cchEv, utils.MetaAuthorize) && utils.OptAsBool(cchEv, utils.MetaResources)) {
			var resID string
			if resID, err = sS.resourcesAuthorize(ctx, cgrEv); err != nil {
				if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
					return
				}
				withErrors = true
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s processing event: %+v with %s",
						utils.SessionS, err.Error(), cgrEv, utils.ResourceS))
			}
			if apiRply.ResourceAllocation == nil {
				apiRply.ResourceAllocation = make(map[string]string)
			}
			apiRply.ResourceAllocation[runID] = resID
		}

		// ResourceS Allocation
		if resAllocBool, errBool := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.ResourcesAllocate,
			utils.MetaResourcesAllocateCfg); errBool != nil {
			if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
				return errBool
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errBool.Error(), cgrEv, utils.MetaResourcesAllocateCfg))
		} else {
			cchEv[utils.MetaResourcesAllocateCfg] = resAllocBool
		}
		if utils.OptAsBool(cchEv, utils.MetaResourcesAllocateCfg) {
			var resMessage string
			if resMessage, err = sS.resourcesAllocate(ctx, cgrEv); err != nil {
				if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
					return
				}
				withErrors = true
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s processing event: %+v with %s",
						utils.SessionS, err.Error(), cgrEv, utils.ResourceS))
			}
			if apiRply.ResourceAllocation == nil {
				apiRply.ResourceAllocation = make(map[string]string)
			}
			apiRply.ResourceAllocation[runID] = resMessage
		}

		// ResourceS Release
		if resReleaseBool, errBool := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.ResourcesRelease,
			utils.MetaResourcesReleaseCfg); errBool != nil {
			if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
				return errBool
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errBool.Error(), cgrEv, utils.MetaResourcesReleaseCfg))
		} else {
			cchEv[utils.MetaResourcesReleaseCfg] = resReleaseBool
		}
		if utils.OptAsBool(cchEv, utils.MetaResourcesReleaseCfg) {
			if err = sS.resourcesRelease(ctx, cgrEv); err != nil {
				if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
					return
				}
				withErrors = true
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s processing event: %+v with %s",
						utils.SessionS, err.Error(), cgrEv, utils.ResourceS))
			}
		}

		// AccountS Authorization
		if acntsAuthBool, errBool := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.AccountsAuthorize,
			utils.MetaAccountsAuthorizeCfg); errBool != nil {
			if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
				return errBool
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errBool.Error(), cgrEv, utils.MetaAccountsAuthorizeCfg))
		} else {
			cchEv[utils.MetaAccountsAuthorizeCfg] = acntsAuthBool
		}
		if utils.OptAsBool(cchEv, utils.MetaAccountsAuthorizeCfg) ||
			(utils.OptAsBool(cchEv, utils.MetaAuthorize) && utils.OptAsBool(cchEv, utils.MetaAccounts)) {
			var acntCost *utils.EventCharges
			if acntCost, err = sS.accountsMaxAbstracts(ctx, cgrEv); err != nil {
				if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
					return
				}
				withErrors = true
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s processing event: %+v with %s for MaxAbstracts",
						utils.SessionS, err.Error(), cgrEv, utils.AccountS))
			} else {
				maxDur, _ := acntCost.Abstracts.Duration()
				if apiRply.AccountsUsage == nil {
					apiRply.AccountsUsage = make(map[string]time.Duration)
				}
				apiRply.AccountsUsage[runID] = maxDur
				if s == nil { // add it for export or ur, only for non session since sessions are written in terminate method
					cgrEv.APIOpts[utils.MetaAccountsCost] = acntCost
				}
			}

		}

		// AccountsRefund
		if acntsRfndBool, errBool := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.AccountsRefund,
			utils.MetaAccountsRefundCfg); errBool != nil {
			if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
				return errBool
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errBool.Error(), cgrEv, utils.MetaAccountsRefundCfg))
		} else {
			cchEv[utils.MetaAccountsRefundCfg] = acntsRfndBool
		}
		if utils.OptAsBool(cchEv, utils.MetaRefund) ||
			utils.OptAsBool(cchEv, utils.MetaAccountsRefundCfg) {
			var refundCharges *utils.EventCharges
			var errRefund error
			accountsCost, has := cgrEv.APIOpts[utils.MetaAccountsCost]
			if !has {
				errRefund = utils.NewErrMandatoryIeMissing(utils.MetaAccountsCost)
			} else {
				refundCharges, errRefund = utils.IfaceAsEventCharges(accountsCost)
			}
			if errRefund == nil {
				errRefund = sS.accountSRefundCharges(ctx, refundCharges, cgrEv)
			}
			if errRefund != nil {
				if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
					return errRefund
				}
				withErrors = true
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
						utils.SessionS, errRefund.Error(), cgrEv, utils.MetaRefund))
				continue
			}
		}
		// AccountS Debit
		if acntsDebitBool, errBool := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.AccountsDebit,
			utils.MetaAccountsDebitCfg); errBool != nil {
			if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
				return errBool
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errBool.Error(), cgrEv, utils.MetaAccountsDebitCfg))
		} else {
			cchEv[utils.MetaAccountsDebitCfg] = acntsDebitBool
		}
		if utils.OptAsBool(cchEv, utils.MetaAccountsDebitCfg) ||
			(utils.OptAsBool(cchEv, utils.MetaAccounts) && utils.OptAsBool(cchEv, utils.MetaDebit)) {
			if s != nil {
				cgrEv.APIOpts[utils.MetaUsage] = s.sRuns[runID].nextDebit
			}
			var acntCost *utils.EventCharges
			if acntCost, err = sS.accountSDebitEvent(ctx, cgrEv); err != nil {
				if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
					return
				}
				withErrors = true
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s processing event: %+v with %s for Debit",
						utils.SessionS, err.Error(), cgrEv, utils.AccountS))
			} else {
				acntDbt := acntCost.Abstracts
				if s != nil {
					s.lk.Lock()
					if s.sRuns[runID].Charges == nil {
						s.sRuns[runID].Charges = utils.NewEventCharges()
					}
					s.sRuns[runID].Charges.Merge(acntCost)
					if s.sRuns[runID].lclDebit != nil {
						acntDbt = utils.SumDecimal(acntDbt, s.sRuns[runID].lclDebit)
					}
					s.lk.Unlock()
				} else { // add it for export or ur, only for non session since sessions are written in terminate method
					cgrEv.APIOpts[utils.MetaAccountsCost] = acntCost
				}
				if apiRply.AccountsUsage == nil {
					apiRply.AccountsUsage = make(map[string]time.Duration)
				}
				maxDur, _ := acntDbt.Duration()
				apiRply.AccountsUsage[runID] = maxDur
			}
		}

	}

	// TerminateSession
	if utils.OptAsBool(cch, utils.MetaTerminate) && s != nil {
		if errTerminate := sS.terminateSessionNew(ctx, s); errTerminate != nil {
			return errTerminate
		}
	}
	for runID, cgrEv := range cgrEvs { // need to run here in the eventuality of terminate being called
		cchEv := make(map[string]any)
		// terminate corrections
		if utils.OptAsBool(cch, utils.MetaTerminate) && s != nil {
			if _, has := apiRply.AccountsUsage[runID]; has { // correct usage to reflect total usage instead of interm one
				totalUsage, _ := s.sRuns[runID].TotalUsage.Duration()
				apiRply.AccountsUsage[runID] = totalUsage
			}
			if s.sRuns[runID].Charges != nil {
				cgrEv.APIOpts[utils.MetaAccountsCost] = s.sRuns[runID].Charges
			}
		}
		// UsageRecords generation
		if ur, errUR := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.UR,
			utils.MetaUR); errUR != nil {
			if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
				return errUR
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errUR.Error(), cgrEv, utils.MetaUR))
		} else if ur {
			if apiRply.UsageRecords == nil {
				apiRply.UsageRecords = make(map[string]*utils.CGREvent)
			}
			apiRply.UsageRecords[runID] = cgrEv
		}
		// Event Exports
		if ees, errEEs := engine.GetBoolOpts(ctx, apiArgs.Tenant, apiArgs.AsDataProvider(), cchEv,
			sS.fltrS, sS.cfg.SessionSCfg().Opts.EEs,
			utils.MetaEEs); errEEs != nil {
			if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
				return errEEs
			}
			withErrors = true
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event: %+v flag for %s",
					utils.SessionS, errEEs.Error(), cgrEv, utils.EEs))
		} else if ees {
			var eesIDs []string
			if eesIDs, err = sS.eesProcessEvent(ctx, cgrEv); err != nil {
				if utils.OptAsBool(cch, utils.OptsSesBlockerError) {
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
