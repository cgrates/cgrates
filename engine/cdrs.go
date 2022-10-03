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

package engine

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

func newMapEventFromReqForm(r *http.Request) (mp MapEvent, err error) {
	if r.Form == nil {
		if err = r.ParseForm(); err != nil {
			return
		}
	}
	mp = MapEvent{utils.Source: r.RemoteAddr}
	for k, vals := range r.Form {
		mp[k] = vals[0] // We only support the first value for now, if more are provided it is considered remote's fault
	}
	return
}

// NewCDRServer is a constructor for CDRServer
func NewCDRServer(cfg *config.CGRConfig, dm *DataManager, filterS *FilterS,
	connMgr *ConnManager) *CDRServer {
	return &CDRServer{
		cfg:     cfg,
		dm:      dm,
		guard:   guardian.Guardian,
		fltrS:   filterS,
		connMgr: connMgr,
	}
}

// CDRServer stores and rates CDRs
type CDRServer struct {
	cfg     *config.CGRConfig
	dm      *DataManager
	guard   *guardian.GuardianLocker
	fltrS   *FilterS
	connMgr *ConnManager
}

// ListenAndServe listen for storbd reload
func (cdrS *CDRServer) ListenAndServe(stopChan chan struct{}) {
	for {
		select {
		case <-stopChan:
			return
		}
	}
}

// chrgrSProcessEvent forks CGREventWithOpts into multiples based on matching ChargerS profiles
func (cdrS *CDRServer) chrgrSProcessEvent(ctx *context.Context, cgrEv *utils.CGREvent) (cgrEvs []*utils.CGREvent, err error) {
	var chrgrs []*ChrgSProcessEventReply
	if err = cdrS.connMgr.Call(ctx, cdrS.cfg.CdrsCfg().ChargerSConns,
		utils.ChargerSv1ProcessEvent,
		cgrEv, &chrgrs); err != nil {
		return
	}
	if len(chrgrs) == 0 {
		return
	}
	cgrEvs = make([]*utils.CGREvent, len(chrgrs))
	for i, cgrPrfl := range chrgrs {
		cgrEvs[i] = cgrPrfl.CGREvent
	}
	return
}

// attrSProcessEvent will send the event to StatS if the connection is configured
func (cdrS *CDRServer) attrSProcessEvent(ctx *context.Context, cgrEv *utils.CGREvent) (err error) {
	var rplyEv AttrSProcessEventReply
	if cgrEv.APIOpts == nil {
		cgrEv.APIOpts = make(map[string]interface{})
	}
	cgrEv.APIOpts[utils.MetaSubsys] = utils.MetaCDRs
	cgrEv.APIOpts[utils.OptsContext] = utils.FirstNonEmpty(
		utils.IfaceAsString(cgrEv.APIOpts[utils.OptsContext]),
		utils.MetaCDRs)
	if err = cdrS.connMgr.Call(ctx, cdrS.cfg.CdrsCfg().AttributeSConns,
		utils.AttributeSv1ProcessEvent,
		cgrEv, &rplyEv); err == nil && len(rplyEv.AlteredFields) != 0 {
		*cgrEv = *rplyEv.CGREvent
	} else if err != nil &&
		err.Error() == utils.ErrNotFound.Error() {
		err = nil // cancel ErrNotFound
	}
	return
}

// rateSProcessEvent will send the event to rateS and attach the cost received back to event
func (cdrS *CDRServer) rateSCostForEvent(ctx *context.Context, cgrEv *utils.CGREvent) (err error) {
	var rpCost utils.RateProfileCost
	if err = cdrS.connMgr.Call(ctx, cdrS.cfg.CdrsCfg().RateSConns,
		utils.RateSv1CostForEvent,
		cgrEv, &rpCost); err != nil {
		return
	}
	cgrEv.APIOpts[utils.MetaRateSCost] = rpCost
	return
}

// accountSDebitEvent will send the event to accountS and attach the cost received back to event
func (cdrS *CDRServer) accountSDebitEvent(ctx *context.Context, cgrEv *utils.CGREvent) (err error) {
	acntCost := new(utils.EventCharges)
	if err = cdrS.connMgr.Call(ctx, cdrS.cfg.CdrsCfg().AccountSConns,
		utils.AccountSv1DebitAbstracts, cgrEv, acntCost); err != nil {
		return
	}
	cgrEv.APIOpts[utils.MetaAccountSCost] = acntCost
	return
}

// accountSRefundCharges will refund the charges into Account
func (cdrS *CDRServer) accountSRefundCharges(ctx *context.Context, tnt string, eChrgs *utils.EventCharges,
	apiOpts map[string]interface{}) (err error) {
	var rply string
	argsRefund := &utils.APIEventCharges{
		Tenant:       tnt,
		APIOpts:      apiOpts,
		EventCharges: eChrgs,
	}
	if err = cdrS.connMgr.Call(ctx, cdrS.cfg.CdrsCfg().AccountSConns,
		utils.AccountSv1RefundCharges, argsRefund, &rply); err != nil {
		return
	}
	return
}

// thdSProcessEvent will send the event to ThresholdS
func (cdrS *CDRServer) thdSProcessEvent(ctx *context.Context, cgrEv *utils.CGREvent) (err error) {
	var tIDs []string
	// we clone the CGREvent so we can add EventType without being propagated
	cgrEv = cgrEv.Clone()
	cgrEv.APIOpts[utils.MetaEventType] = utils.CDR
	if err = cdrS.connMgr.Call(ctx, cdrS.cfg.CdrsCfg().ThresholdSConns,
		utils.ThresholdSv1ProcessEvent,
		cgrEv, &tIDs); err != nil &&
		err.Error() == utils.ErrNotFound.Error() {
		err = nil // NotFound is not considered error
	}
	return
}

// statSProcessEvent will send the event to StatS
func (cdrS *CDRServer) statSProcessEvent(ctx *context.Context, cgrEv *utils.CGREvent) (err error) {
	var reply []string
	if err = cdrS.connMgr.Call(ctx, cdrS.cfg.CdrsCfg().StatSConns,
		utils.StatSv1ProcessEvent,
		cgrEv.Clone(), &reply); err != nil &&
		err.Error() == utils.ErrNotFound.Error() {
		err = nil // NotFound is not considered error
	}
	return
}

// eeSProcessEvent will process the event with the EEs component
func (cdrS *CDRServer) eeSProcessEvent(ctx *context.Context, cgrEv *utils.CGREventWithEeIDs) (err error) {
	var reply map[string]map[string]interface{}
	if err = cdrS.connMgr.Call(ctx, cdrS.cfg.CdrsCfg().EEsConns,
		utils.EeSv1ProcessEvent,
		cgrEv, &reply); err != nil &&
		err.Error() == utils.ErrNotFound.Error() {
		err = nil // NotFound is not considered error
	}
	return
}

// processEvent processes a CGREvent based on arguments
// in case of partially executed, both error and evs will be returned
func (cdrS *CDRServer) processEvent(ctx *context.Context, ev *utils.CGREvent) (evs []*utils.EventsWithOpts, err error) {
	// making the options
	var attrS bool
	if attrS, err = GetBoolOpts(ctx, ev.Tenant, ev.AsDataProvider(), cdrS.fltrS, cdrS.cfg.CdrsCfg().Opts.Attributes,
		config.CDRsAttributesDftOpt, utils.MetaAttributes); err != nil {
		return
	}
	if attrS {
		if err = cdrS.attrSProcessEvent(ctx, ev); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: <%s> processing event %+v with %s",
					utils.CDRs, err.Error(), utils.ToJSON(ev), utils.AttributeS))
			err = utils.NewErrAttributeS(err)
			return
		}
	}

	var cgrEvs []*utils.CGREvent
	var chrgS bool
	if chrgS, err = GetBoolOpts(ctx, ev.Tenant, ev.AsDataProvider(), cdrS.fltrS, cdrS.cfg.CdrsCfg().Opts.Chargers,
		config.CDRsChargersDftOpt, utils.MetaChargers); err != nil {
		return
	}
	if chrgS {
		if cgrEvs, err = cdrS.chrgrSProcessEvent(ctx, ev); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: <%s> processing event %+v with %s",
					utils.CDRs, err.Error(), utils.ToJSON(ev), utils.ChargerS))
			err = utils.NewErrChargerS(err)
			return
		}
	} else { // ChargerS not requested, charge the original event
		cgrEvs = []*utils.CGREvent{ev}
	}

	var partiallyExecuted bool // from here actions are optional and a general error is returned

	var rateS bool
	for _, cgrEv := range cgrEvs {
		if rateS, err = GetBoolOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(), cdrS.fltrS, cdrS.cfg.CdrsCfg().Opts.Rates,
			config.CDRsRatesDftOpt, utils.MetaRates); err != nil {
			return
		}
		if rateS {
			if err := cdrS.rateSCostForEvent(ctx, cgrEv); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: <%s> processing event %+v with %s",
						utils.CDRs, err.Error(), utils.ToJSON(cgrEv), utils.RateS))
				partiallyExecuted = true
			}
		}
	}

	var acntS bool
	for _, cgrEv := range cgrEvs {
		if acntS, err = GetBoolOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(), cdrS.fltrS, cdrS.cfg.CdrsCfg().Opts.Accounts,
			config.CDRsAccountsDftOpt, utils.MetaAccounts); err != nil {
			return
		}
		if acntS {
			if ecCostIface, wasCharged := cgrEv.APIOpts[utils.MetaAccountSCost]; wasCharged {
				// before converting into EventChargers, we must get the JSON encoding and Unmarshal it into an EventChargers
				var btsEvCh []byte
				btsEvCh, err = json.Marshal(ecCostIface.(map[string]interface{}))
				if err != nil {
					return
				}
				ecCost := new(utils.EventCharges)
				if err = json.Unmarshal(btsEvCh, &ecCost); err != nil {
					return
				}
				// call the refund
				if err := cdrS.accountSRefundCharges(ctx, cgrEv.Tenant, ecCost, cgrEv.APIOpts); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> error: <%s> processing event %+v with %s",
							utils.CDRs, err.Error(), utils.ToJSON(cgrEv), utils.AccountS))
					partiallyExecuted = true
				}
			}
			if err := cdrS.accountSDebitEvent(ctx, cgrEv); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: <%s> processing event %+v with %s",
						utils.CDRs, err.Error(), utils.ToJSON(cgrEv), utils.AccountS))
				partiallyExecuted = true
			}
		}
	}
	// populate cost from accounts or rates for every event
	for _, cgrEv := range cgrEvs {
		if cost := populateCost(cgrEv.APIOpts); cost != nil {
			cgrEv.APIOpts[utils.MetaCost] = cost
		}
	}
	var export bool
	for _, cgrEv := range cgrEvs {
		if export, err = GetBoolOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(), cdrS.fltrS, cdrS.cfg.CdrsCfg().Opts.Export,
			config.CDRsExportDftOpt, utils.OptsCDRsExport); err != nil {
			return
		}
		if export {
			evWithOpts := &utils.CGREventWithEeIDs{
				CGREvent: cgrEv,
				EeIDs:    cdrS.cfg.CdrsCfg().OnlineCDRExports,
			}
			if err := cdrS.eeSProcessEvent(ctx, evWithOpts); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: <%s> exporting cdr %+v",
						utils.CDRs, err.Error(), utils.ToJSON(evWithOpts)))
				partiallyExecuted = true
			}
		}
	}

	var thdS bool
	for _, cgrEv := range cgrEvs {
		if thdS, err = GetBoolOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(), cdrS.fltrS, cdrS.cfg.CdrsCfg().Opts.Thresholds,
			config.CDRsThresholdsDftOpt, utils.MetaThresholds); err != nil {
			return
		}
		if thdS {
			if err := cdrS.thdSProcessEvent(ctx, cgrEv); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: <%s> processing event %+v with %s",
						utils.CDRs, err.Error(), utils.ToJSON(cgrEv), utils.ThresholdS))
				partiallyExecuted = true
			}
		}
	}

	var stS bool
	for _, cgrEv := range cgrEvs {
		if stS, err = GetBoolOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(), cdrS.fltrS, cdrS.cfg.CdrsCfg().Opts.Stats,
			config.CDRsStatsDftOpt, utils.MetaStats); err != nil {
			return
		}
		if stS {
			if err := cdrS.statSProcessEvent(ctx, cgrEv); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: <%s> processing event %+v with %s",
						utils.CDRs, err.Error(), utils.ToJSON(cgrEv), utils.StatS))
				partiallyExecuted = true
			}
		}
	}

	// now that we did all the requested processed events, we have to build our EventsWithOpts
	evs = make([]*utils.EventsWithOpts, len(cgrEvs))
	for i, cgrEv := range cgrEvs {
		evs[i] = &utils.EventsWithOpts{
			Event: cgrEv.Event,
			Opts:  cgrEv.APIOpts,
		}
	}

	if partiallyExecuted {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// V1ProcessEvent will process the CGREvent
func (cdrS *CDRServer) V1ProcessEvent(ctx *context.Context, arg *utils.CGREvent, reply *string) (err error) {
	if arg.ID == utils.EmptyString {
		arg.ID = utils.GenUUID()
	}
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = cdrS.cfg.GeneralCfg().DefaultTenant
	}
	// RPC caching
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.CDRsV1ProcessEvent, arg.ID)
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)

		if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*string)
			}
			return cachedResp.Error
		}
		defer Cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	if _, err = cdrS.processEvent(ctx, arg); err != nil {
		return
	}
	*reply = utils.OK
	return nil
}

// V1ProcessEventWithGet has the same logic with V1ProcessEvent except it adds the proccessed events to the reply
func (cdrS *CDRServer) V1ProcessEventWithGet(ctx *context.Context, arg *utils.CGREvent, evs *[]*utils.EventsWithOpts) (err error) {
	if arg.ID == utils.EmptyString {
		arg.ID = utils.GenUUID()
	}
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = cdrS.cfg.GeneralCfg().DefaultTenant
	}
	// RPC caching
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.CDRsV1ProcessEventWithGet, arg.ID)
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)

		if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*evs = *cachedResp.Result.(*[]*utils.EventsWithOpts)
			}
			return cachedResp.Error
		}
		defer Cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: evs, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching
	var procEvs []*utils.EventsWithOpts
	if procEvs, err = cdrS.processEvent(ctx, arg); err != nil {
		return
	}
	*evs = procEvs
	return nil
}

func populateCost(cgrOpts map[string]interface{}) *utils.Decimal {
	// if the cost is already present, get out
	if _, has := cgrOpts[utils.MetaCost]; has {
		return nil
	}
	// check firstly in accounts
	if accCost, has := cgrOpts[utils.MetaAccountSCost]; has {
		return accCost.(*utils.EventCharges).Concretes
	}
	// after check in rates
	if rtCost, has := cgrOpts[utils.MetaRateSCost]; has {
		return rtCost.(utils.RateProfileCost).Cost
	}
	return nil
}
