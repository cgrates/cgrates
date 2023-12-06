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
	"errors"
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
func NewCDRServer(cfg *config.CGRConfig, dm *DataManager, filterS *FilterS, connMgr *ConnManager,
	storDBChan chan StorDB) *CDRServer {
	cdrDB := <-storDBChan
	return &CDRServer{
		cfg:     cfg,
		dm:      dm,
		db:      cdrDB,
		guard:   guardian.Guardian,
		fltrS:   filterS,
		connMgr: connMgr,
		dbChan:  storDBChan,
	}
}

// CDRServer stores and rates CDRs
type CDRServer struct {
	cfg     *config.CGRConfig
	dm      *DataManager
	db      StorDB
	guard   *guardian.GuardianLocker
	fltrS   *FilterS
	connMgr *ConnManager
	dbChan  chan StorDB
}

// ListenAndServe listen for storbd reload
func (cdrS *CDRServer) ListenAndServe(stopChan chan struct{}) {
	for {
		select {
		case <-stopChan:
			return
		case _, ok := <-cdrS.dbChan:
			if !ok { // the channel was closed by the shutdown of the StorDB Service
				return
			}
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
		cgrEv.APIOpts = make(map[string]any)
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
	apiOpts map[string]any) (err error) {
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
	var reply map[string]map[string]any
	if err = cdrS.connMgr.Call(ctx, cdrS.cfg.CdrsCfg().EEsConns,
		utils.EeSv1ProcessEvent,
		cgrEv, &reply); err != nil &&
		err.Error() == utils.ErrNotFound.Error() {
		err = nil // NotFound is not considered error
	}
	return
}

// processEvents processes CGREvents based on arguments.
// In case of partially executed, both the error and the events will be returned.
func (cdrS *CDRServer) processEvents(ctx *context.Context, evs []*utils.CGREvent) ([]*utils.EventsWithOpts, error) {
	for _, ev := range evs {
		attrS, err := GetBoolOpts(ctx, ev.Tenant, ev.AsDataProvider(), cdrS.fltrS, cdrS.cfg.CdrsCfg().Opts.Attributes,
			config.CDRsAttributesDftOpt, utils.MetaAttributes)
		if err != nil {
			return nil, fmt.Errorf("retrieving %s option failed: %w", utils.MetaAttributes, err)
		}
		if !attrS {
			continue
		}
		if err = cdrS.attrSProcessEvent(ctx, ev); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: <%v> processing event %s with %s",
					utils.CDRs, err, utils.ToJSON(ev), utils.AttributeS))
			return nil, utils.NewErrAttributeS(err)
		}
	}

	// Allocate capacity equal to the number of events for slight optimization.
	cgrEvs := make([]*utils.CGREvent, 0, len(evs))

	for _, ev := range evs {
		chrgS, err := GetBoolOpts(ctx, ev.Tenant, ev.AsDataProvider(), cdrS.fltrS, cdrS.cfg.CdrsCfg().Opts.Chargers,
			config.CDRsChargersDftOpt, utils.MetaChargers)
		if err != nil {
			return nil, fmt.Errorf("retrieving %s option failed: %w", utils.MetaChargers, err)
		}
		if !chrgS {
			cgrEvs = append(cgrEvs, ev)
			continue
		}
		chrgEvs, err := cdrS.chrgrSProcessEvent(ctx, ev)
		if err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: <%v> processing event %s with %s",
					utils.CDRs, err, utils.ToJSON(ev), utils.ChargerS))
			return nil, utils.NewErrChargerS(err)
		}
		cgrEvs = append(cgrEvs, chrgEvs...)
	}

	var partiallyExecuted bool // from here actions are optional and a general error is returned

	for _, cgrEv := range cgrEvs {
		rateS, err := GetBoolOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(), cdrS.fltrS, cdrS.cfg.CdrsCfg().Opts.Rates,
			config.CDRsRatesDftOpt, utils.MetaRates)
		if err != nil {
			return nil, fmt.Errorf("retrieving %s option failed: %w", utils.MetaRates, err)
		}
		if !rateS {
			continue
		}
		if err := cdrS.rateSCostForEvent(ctx, cgrEv); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: <%v> processing event %s with %s",
					utils.CDRs, err, utils.ToJSON(cgrEv), utils.RateS))
			partiallyExecuted = true
		}
	}

	for _, cgrEv := range cgrEvs {
		acntS, err := GetBoolOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(), cdrS.fltrS, cdrS.cfg.CdrsCfg().Opts.Accounts,
			config.CDRsAccountsDftOpt, utils.MetaAccounts)
		if err != nil {
			return nil, fmt.Errorf("retrieving %s option failed: %w", utils.MetaAccounts, err)
		}
		if !acntS {
			continue
		}
		if ecCostIface, wasCharged := cgrEv.APIOpts[utils.MetaAccountSCost]; wasCharged {
			ecCostMap, ok := ecCostIface.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("expected %s to be a map[string]any, got %T", utils.MetaAccountSCost, ecCostIface)
			}

			// before converting into EventChargers, we must get the JSON encoding and Unmarshal it into an EventChargers
			btsEvCh, err := json.Marshal(ecCostMap)
			if err != nil {
				return nil, err
			}

			var ecCost utils.EventCharges
			if err = json.Unmarshal(btsEvCh, &ecCost); err != nil {
				return nil, err
			}

			if err := cdrS.accountSRefundCharges(ctx, cgrEv.Tenant, &ecCost, cgrEv.APIOpts); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: <%v> processing event %s with %s",
						utils.CDRs, err, utils.ToJSON(cgrEv), utils.AccountS))
				partiallyExecuted = true
			}
		}
		if err := cdrS.accountSDebitEvent(ctx, cgrEv); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: <%v> processing event %s with %s",
					utils.CDRs, err, utils.ToJSON(cgrEv), utils.AccountS))
			partiallyExecuted = true
		}
	}

	// populate cost from accounts or rates for every event
	for _, cgrEv := range cgrEvs {
		if cost := populateCost(cgrEv.APIOpts); cost != nil {
			cgrEv.APIOpts[utils.MetaCost] = cost
		}
	}

	for _, cgrEv := range cgrEvs {
		store, err := GetBoolOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(), cdrS.fltrS, cdrS.cfg.CdrsCfg().Opts.Store,
			config.CDRsStoreDftOpt, utils.MetaStore)
		if err != nil {
			return nil, fmt.Errorf("retrieving %s option failed: %w", utils.MetaStore, err)
		}
		if !store {
			continue
		}
		rerate, err := GetBoolOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(), cdrS.fltrS, cdrS.cfg.CdrsCfg().Opts.Rerate,
			config.CDRsRerateDftOpt, utils.MetaRerate)
		if err != nil {
			return nil, fmt.Errorf("retrieving %s option failed: %w", utils.MetaRerate, err)
		}
		if err := cdrS.db.SetCDR(cgrEv, false); err != nil {
			if err != utils.ErrExists || !rerate {

				// ToDo: add refund logic
				return nil, fmt.Errorf("storing CDR %s failed: %w", utils.ToJSON(cgrEv), err)
			}
			if err = cdrS.db.SetCDR(cgrEv, true); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: <%s> updating CDR %+v",
						utils.CDRs, err.Error(), utils.ToJSON(cgrEv)))
				return nil, utils.ErrPartiallyExecuted
			}
		}
	}

	for _, cgrEv := range cgrEvs {
		export, err := GetBoolOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(), cdrS.fltrS, cdrS.cfg.CdrsCfg().Opts.Export,
			config.CDRsExportDftOpt, utils.OptsCDRsExport)
		if err != nil {
			return nil, fmt.Errorf("retrieving %s option failed: %w", utils.OptsCDRsExport, err)
		}
		if !export {
			continue
		}
		evWithOpts := &utils.CGREventWithEeIDs{
			CGREvent: cgrEv,
			EeIDs:    cdrS.cfg.CdrsCfg().OnlineCDRExports,
		}
		if err := cdrS.eeSProcessEvent(ctx, evWithOpts); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: <%v> exporting cdr %s",
					utils.CDRs, err, utils.ToJSON(evWithOpts)))
			partiallyExecuted = true
		}
	}

	for _, cgrEv := range cgrEvs {
		thdS, err := GetBoolOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(), cdrS.fltrS, cdrS.cfg.CdrsCfg().Opts.Thresholds,
			config.CDRsThresholdsDftOpt, utils.MetaThresholds)
		if err != nil {
			return nil, fmt.Errorf("retrieving %s option failed: %w", utils.MetaThresholds, err)
		}
		if !thdS {
			continue
		}
		if err := cdrS.thdSProcessEvent(ctx, cgrEv); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: <%v> processing event %s with %s",
					utils.CDRs, err, utils.ToJSON(cgrEv), utils.ThresholdS))
			partiallyExecuted = true
		}
	}

	for _, cgrEv := range cgrEvs {
		stS, err := GetBoolOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(), cdrS.fltrS, cdrS.cfg.CdrsCfg().Opts.Stats,
			config.CDRsStatsDftOpt, utils.MetaStats)
		if err != nil {
			return nil, fmt.Errorf("retrieving %s option failed: %w", utils.MetaStats, err)
		}
		if !stS {
			continue
		}
		if err := cdrS.statSProcessEvent(ctx, cgrEv); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: <%v> processing event %s with %s",
					utils.CDRs, err, utils.ToJSON(cgrEv), utils.StatS))
			partiallyExecuted = true
		}
	}

	// now that we did all the requested processed events, we have to build our EventsWithOpts
	outEvs := make([]*utils.EventsWithOpts, len(cgrEvs))
	for i, cgrEv := range cgrEvs {
		outEvs[i] = &utils.EventsWithOpts{
			Event: cgrEv.Event,
			Opts:  cgrEv.APIOpts,
		}
	}

	if partiallyExecuted {
		return outEvs, utils.ErrPartiallyExecuted
	}
	return outEvs, nil
}

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
	if procEvs, err = cdrS.processEvents(ctx, []*utils.CGREvent{args}); err != nil {
		return
	}
	*evs = procEvs
	return nil
}

func populateCost(cgrOpts map[string]any) *utils.Decimal {
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

// V1ProcessStoredEvents processes stored events based on provided filters.
func (cdrS *CDRServer) V1ProcessStoredEvents(ctx *context.Context, args *CDRFilters, reply *string) (err error) {
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

	fltrs, err := GetFilters(ctx, args.FilterIDs, args.Tenant, cdrS.dm)
	if err != nil {
		return fmt.Errorf("preparing filters failed: %w", err)
	}
	cdrs, err := cdrS.db.GetCDRs(ctx, fltrs, args.APIOpts)
	if err != nil {
		return fmt.Errorf("retrieving CDRs failed: %w", err)
	}
	_, err = cdrS.processEvents(ctx, CDRsToCGREvents(cdrs))
	if err != nil && !errors.Is(err, utils.ErrPartiallyExecuted) {
		return fmt.Errorf("processing events failed: %w", err)
	}
	*reply = utils.OK
	return err
}
