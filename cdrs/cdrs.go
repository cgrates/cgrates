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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/attributes"
	"github.com/cgrates/cgrates/chargers"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
)

func newMapEventFromReqForm(r *http.Request) (mp engine.MapEvent, err error) {
	if r.Form == nil {
		if err = r.ParseForm(); err != nil {
			return
		}
	}
	mp = engine.MapEvent{utils.Source: r.RemoteAddr}
	for k, vals := range r.Form {
		mp[k] = vals[0] // We only support the first value for now, if more are provided it is considered remote's fault
	}
	return
}

// NewCDRServer is a constructor for CDRServer
func NewCDRServer(cfg *config.CGRConfig, dm *engine.DataManager, filterS *engine.FilterS, connMgr *engine.ConnManager,
	storDB engine.StorDB) *CDRServer {
	return &CDRServer{
		cfg:     cfg,
		dm:      dm,
		db:      storDB,
		guard:   guardian.Guardian,
		fltrS:   filterS,
		connMgr: connMgr,
	}
}

// CDRServer stores and rates CDRs
type CDRServer struct {
	cfg     *config.CGRConfig
	dm      *engine.DataManager
	db      engine.StorDB
	guard   *guardian.GuardianLocker
	fltrS   *engine.FilterS
	connMgr *engine.ConnManager
}

// chrgrSProcessEvent forks CGREventWithOpts into multiples based on matching ChargerS profiles
func (cdrS *CDRServer) chrgrSProcessEvent(ctx *context.Context, cgrEv *utils.CGREvent) (cgrEvs []*utils.CGREvent, err error) {
	var chrgrs []*chargers.ChrgSProcessEventReply
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
	var rplyEv attributes.AttrSProcessEventReply
	if cgrEv.APIOpts == nil {
		cgrEv.APIOpts = make(map[string]any)
	}
	cgrEv.APIOpts[utils.MetaSubsys] = utils.MetaCDRs
	cgrEv.APIOpts[utils.OptsContext] = utils.FirstNonEmpty(
		utils.IfaceAsString(cgrEv.APIOpts[utils.OptsContext]),
		utils.MetaCDRs)
	if err = cdrS.connMgr.Call(ctx, cdrS.cfg.CdrsCfg().AttributeSConns,
		utils.AttributeSv1ProcessEvent,
		cgrEv, &rplyEv); err != nil {
		if err.Error() == utils.ErrNotFound.Error() {
			err = nil
		}
		return
	}
	if len(rplyEv.AlteredFields) != 0 {
		*cgrEv = *rplyEv.CGREvent
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
	cgrEv.APIOpts[utils.MetaEventType] = utils.CDRKey
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
		attrS, err := engine.GetBoolOpts(ctx, ev.Tenant, ev.AsDataProvider(), nil, cdrS.fltrS, cdrS.cfg.CdrsCfg().Opts.Attributes,
			utils.MetaAttributes)
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
		chrgS, err := engine.GetBoolOpts(ctx, ev.Tenant, ev.AsDataProvider(), nil, cdrS.fltrS, cdrS.cfg.CdrsCfg().Opts.Chargers,
			utils.MetaChargers)
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
		rateS, err := engine.GetBoolOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(), nil, cdrS.fltrS, cdrS.cfg.CdrsCfg().Opts.Rates,
			utils.MetaRates)
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
		acntS, err := engine.GetBoolOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(), nil, cdrS.fltrS, cdrS.cfg.CdrsCfg().Opts.Accounts,
			utils.MetaAccounts)
		if err != nil {
			return nil, fmt.Errorf("retrieving %s option failed: %w", utils.MetaAccounts, err)
		}
		if !acntS {
			continue
		}
		if ecCostIface, wasCharged := cgrEv.APIOpts[utils.MetaAccountSCost]; wasCharged {
			ecCostMap, ok := ecCostIface.(*utils.EventCharges)
			if !ok {
				return nil, fmt.Errorf("expected %s to be a  *utils.EventCharges, got %T", utils.MetaAccountSCost, ecCostMap)
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
		store, err := engine.GetBoolOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(), nil, cdrS.fltrS, cdrS.cfg.CdrsCfg().Opts.Store,
			utils.MetaStore)
		if err != nil {
			return nil, fmt.Errorf("retrieving %s option failed: %w", utils.MetaStore, err)
		}
		if !store {
			continue
		}
		rerate, err := engine.GetBoolOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(), nil, cdrS.fltrS, cdrS.cfg.CdrsCfg().Opts.Rerate,
			utils.MetaRerate)
		if err != nil {
			return nil, fmt.Errorf("retrieving %s option failed: %w", utils.MetaRerate, err)
		}

		// Prevent 'assignment to entry in nil map' panic.
		if cgrEv.APIOpts == nil {
			cgrEv.APIOpts = make(map[string]any)
		}

		// Make sure *cdrID key exists in opts, as it's needed to identify CDRs during CRUD operations.
		if _, ok := cgrEv.APIOpts[utils.MetaCDRID]; !ok {
			cgrEv.APIOpts[utils.MetaCDRID] = utils.GetUniqueCDRID(cgrEv)
		}

		if err := cdrS.db.SetCDR(ctx, cgrEv, false); err != nil {
			if err != utils.ErrExists || !rerate {

				// ToDo: add refund logic
				return nil, fmt.Errorf("storing CDR %s failed: %w", utils.ToJSON(cgrEv), err)
			}
			if err = cdrS.db.SetCDR(ctx, cgrEv, true); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: <%s> updating CDR %+v",
						utils.CDRs, err.Error(), utils.ToJSON(cgrEv)))
				return nil, utils.ErrPartiallyExecuted
			}
		}
	}

	for _, cgrEv := range cgrEvs {
		export, err := engine.GetBoolOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(), nil, cdrS.fltrS, cdrS.cfg.CdrsCfg().Opts.Export,
			utils.OptsCDRsExport)
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
		thdS, err := engine.GetBoolOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(), nil, cdrS.fltrS, cdrS.cfg.CdrsCfg().Opts.Thresholds,
			utils.MetaThresholds)
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
		stS, err := engine.GetBoolOpts(ctx, cgrEv.Tenant, cgrEv.AsDataProvider(), nil, cdrS.fltrS, cdrS.cfg.CdrsCfg().Opts.Stats,
			utils.MetaStats)
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
