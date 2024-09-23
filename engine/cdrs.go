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
	"fmt"
	"net/http"
	"slices"
	"time"

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

// cgrCdrHandler handles CDRs received over HTTP REST
func (cdrS *CDRServer) cgrCdrHandler(w http.ResponseWriter, r *http.Request) {
	cgrCDR, err := newMapEventFromReqForm(r)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> could not create CDR entry from http: %+v, err <%s>",
				utils.CDRs, r.Form, err.Error()))
		return
	}
	cdr, err := cgrCDR.AsCDR(cdrS.cgrCfg, cdrS.cgrCfg.GeneralCfg().DefaultTenant, cdrS.cgrCfg.GeneralCfg().DefaultTimezone)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> could not create CDR entry from rawCDR: %+v, err <%s>",
				utils.CDRs, cgrCDR, err.Error()))
		return
	}
	var ignored string
	if err := cdrS.V1ProcessCDR(context.TODO(), &CDRWithAPIOpts{CDR: cdr}, &ignored); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> processing CDR: %s, err: <%s>",
				utils.CDRs, cdr, err.Error()))
	}
}

// fsCdrHandler will handle CDRs received from FreeSWITCH over HTTP-JSON
func (cdrS *CDRServer) fsCdrHandler(w http.ResponseWriter, r *http.Request) {
	fsCdr, err := NewFSCdr(r.Body, cdrS.cgrCfg)
	r.Body.Close()
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<CDRS> Could not create CDR entry: %s", err.Error()))
		return
	}
	cdr, err := fsCdr.AsCDR(cdrS.cgrCfg.GeneralCfg().DefaultTimezone)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<CDRS> Could not create AsCDR entry: %s", err.Error()))
		return
	}
	var ignored string
	if err := cdrS.V1ProcessCDR(context.TODO(), &CDRWithAPIOpts{CDR: cdr}, &ignored); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> processing CDR: %s, err: <%s>",
				utils.CDRs, cdr, err.Error()))
	}
}

// NewCDRServer is a constructor for CDRServer
func NewCDRServer(cgrCfg *config.CGRConfig, storDBChan chan StorDB, dm *DataManager, filterS *FilterS,
	connMgr *ConnManager) *CDRServer {
	cdrDb := <-storDBChan
	return &CDRServer{
		cgrCfg:     cgrCfg,
		cdrDb:      cdrDb,
		dm:         dm,
		guard:      guardian.Guardian,
		filterS:    filterS,
		connMgr:    connMgr,
		storDBChan: storDBChan,
	}
}

// CDRServer stores and rates CDRs
type CDRServer struct {
	cgrCfg     *config.CGRConfig
	cdrDb      CdrStorage
	dm         *DataManager
	guard      *guardian.GuardianLocker
	filterS    *FilterS
	connMgr    *ConnManager
	storDBChan chan StorDB
}

// ListenAndServe listen for storbd reload
func (cdrS *CDRServer) ListenAndServe(stopChan chan struct{}) {
	for {
		select {
		case <-stopChan:
			return
		case stordb, ok := <-cdrS.storDBChan:
			if !ok { // the chanel was closed by the shutdown of stordbService
				return
			}
			cdrS.cdrDb = stordb
		}
	}
}

// RegisterHandlersToServer is called by cgr-engine to register HTTP URL handlers
func (cdrS *CDRServer) RegisterHandlersToServer(server utils.Server) {
	server.RegisterHttpFunc(cdrS.cgrCfg.HTTPCfg().HTTPCDRsURL, cdrS.cgrCdrHandler)
	server.RegisterHttpFunc(cdrS.cgrCfg.HTTPCfg().HTTPFreeswitchCDRsURL, cdrS.fsCdrHandler)
}

// storeSMCost will store a SMCost
func (cdrS *CDRServer) storeSMCost(smCost *SMCost, checkDuplicate bool) error {
	smCost.CostDetails.Compute()                                              // make sure the total cost reflect the increment
	lockKey := utils.MetaCDRs + smCost.CGRID + smCost.RunID + smCost.OriginID // Will lock on this ID
	if checkDuplicate {
		return cdrS.guard.Guard(func() error {
			smCosts, err := cdrS.cdrDb.GetSMCosts(smCost.CGRID, smCost.RunID, "", "")
			if err != nil && err.Error() != utils.NotFoundCaps {
				return err
			}
			if len(smCosts) != 0 {
				return utils.ErrExists
			}
			return cdrS.cdrDb.SetSMCost(smCost)
		}, config.CgrConfig().GeneralCfg().LockingTimeout, lockKey) // FixMe: Possible deadlock with Guard from SMG session close()
	}
	return cdrS.cdrDb.SetSMCost(smCost)
}

// rateCDR will populate cost field
// Returns more than one rated CDR in case of SMCost retrieved based on prefix
func (cdrS *CDRServer) rateCDR(cdr *CDRWithAPIOpts) ([]*CDR, error) {
	var qryCC *CallCost
	var err error
	if cdr.RequestType == utils.MetaNone {
		return nil, nil
	}
	if cdr.Usage < 0 {
		cdr.Usage = time.Duration(0)
	}
	cdr.ExtraInfo = "" // Clean previous ExtraInfo, useful when re-rating
	var cdrsRated []*CDR
	_, hasLastUsed := cdr.ExtraFields[utils.LastUsed]
	if slices.Contains([]string{utils.MetaPrepaid, utils.Prepaid}, cdr.RequestType) &&
		(cdr.Usage != 0 || hasLastUsed) && cdr.CostDetails == nil {
		// ToDo: Get rid of Prepaid as soon as we don't want to support it backwards
		// Should be previously calculated and stored in DB
		fib := utils.FibDuration(time.Millisecond, 0)
		var smCosts []*SMCost
		cgrID := cdr.CGRID
		if _, hasIT := cdr.ExtraFields[utils.OriginIDPrefix]; hasIT {
			cgrID = "" // for queries involving originIDPrefix we ignore CGRID
		}
		for i := 0; i < cdrS.cgrCfg.CdrsCfg().SMCostRetries; i++ {
			smCosts, err = cdrS.cdrDb.GetSMCosts(cgrID, cdr.RunID, cdr.OriginHost,
				cdr.ExtraFields[utils.OriginIDPrefix])
			if err == nil && len(smCosts) != 0 {
				break
			}
			if i <= cdrS.cgrCfg.CdrsCfg().SMCostRetries-1 {
				time.Sleep(fib())
			}
		}
		if len(smCosts) != 0 { // Cost retrieved from SMCost table
			for _, smCost := range smCosts {
				cdrClone := cdr.CDR.Clone()
				cdrClone.OriginID = smCost.OriginID
				if cdr.Usage == 0 {
					cdrClone.Usage = smCost.Usage
				} else if smCost.Usage != cdr.Usage {
					if _, err = cdrS.refundEventCost(smCost.CostDetails, // ToDo: need to maybe mark it in the future in the processed flags
						cdrClone.RequestType, cdrClone.ToR); err != nil {
						return nil, err
					}
					cdrClone.CostDetails = nil
					if qryCC, err = cdrS.getCostFromRater(&CDRWithAPIOpts{CDR: cdrClone}); err != nil {
						return nil, err
					}
					smCost = &SMCost{
						CGRID:       cdrClone.CGRID,
						RunID:       cdrClone.RunID,
						OriginHost:  cdrClone.OriginID,
						CostSource:  utils.CDRs,
						Usage:       cdrClone.Usage,
						CostDetails: NewEventCostFromCallCost(qryCC, cdrClone.CGRID, cdrClone.RunID),
					}
				}
				cdrClone.Cost = smCost.CostDetails.GetCost()
				cdrClone.CostDetails = smCost.CostDetails
				cdrClone.CostSource = smCost.CostSource
				cdrsRated = append(cdrsRated, cdrClone)
			}
			return cdrsRated, nil
		}
		utils.Logger.Warning(
			fmt.Sprintf("<Cdrs> WARNING: Could not find CallCostLog for cgrid: %s, source: %s, runid: %s, originID: %s originHost: %s, will recalculate",
				cdr.CGRID, utils.MetaSessionS, cdr.RunID, cdr.OriginID, cdr.OriginHost))
	}
	if cdr.CostDetails != nil {
		if cdr.Usage <= cdr.CostDetails.GetUsage() { // Costs were previously calculated, make sure they cover the full usage
			cdr.Cost = cdr.CostDetails.GetCost()
			cdr.CostDetails.Compute()
			return []*CDR{cdr.CDR}, nil
		}
		// ToDo: need to maybe mark it in the future in the processed flags
		if _, err = cdrS.refundEventCost(cdr.CostDetails,
			cdr.RequestType, cdr.ToR); err != nil {
			return nil, err
		}
		cdr.CostDetails = nil
	}
	qryCC, err = cdrS.getCostFromRater(cdr)
	if err != nil {
		return nil, err
	}
	if qryCC != nil {
		cdr.Cost = qryCC.Cost
		cdr.CostDetails = NewEventCostFromCallCost(qryCC, cdr.CGRID, cdr.RunID)
	}
	cdr.CostDetails.Compute()
	return []*CDR{cdr.CDR}, nil
}

var reqTypes = utils.NewStringSet([]string{utils.MetaPseudoPrepaid, utils.MetaPostpaid, utils.MetaPrepaid,
	utils.PseudoPrepaid, utils.Postpaid, utils.Prepaid, utils.MetaDynaprepaid})

// getCostFromRater will retrieve the cost from RALs
func (cdrS *CDRServer) getCostFromRater(cdr *CDRWithAPIOpts) (*CallCost, error) {
	if len(cdrS.cgrCfg.CdrsCfg().RaterConns) == 0 {
		return nil, utils.NewErrNotConnected(utils.RALService)
	}
	cc := new(CallCost)
	var err error
	timeStart := cdr.AnswerTime
	if timeStart.IsZero() { // Fix for FreeSWITCH unanswered calls
		timeStart = cdr.SetupTime
	}
	cd := &CallDescriptor{
		ToR:             cdr.ToR,
		Tenant:          cdr.Tenant,
		Category:        cdr.Category,
		Subject:         cdr.Subject,
		Account:         cdr.Account,
		Destination:     cdr.Destination,
		ExtraFields:     cdr.ExtraFields,
		TimeStart:       timeStart,
		TimeEnd:         timeStart.Add(cdr.Usage),
		DurationIndex:   cdr.Usage,
		PerformRounding: true,
	}
	if reqTypes.Has(cdr.RequestType) { // Prepaid - Cost can be recalculated in case of missing records from SM
		err = cdrS.connMgr.Call(context.TODO(), cdrS.cgrCfg.CdrsCfg().RaterConns,
			utils.ResponderDebit,
			&CallDescriptorWithAPIOpts{
				CallDescriptor: cd,
				APIOpts:        cdr.APIOpts,
			}, cc)
		if err != nil && err.Error() == utils.ErrAccountNotFound.Error() &&
			cdr.RequestType == utils.MetaDynaprepaid {
			var reply string
			// execute the actionPlan configured in Scheduler
			if err = cdrS.connMgr.Call(context.TODO(), cdrS.cgrCfg.CdrsCfg().SchedulerConns,
				utils.SchedulerSv1ExecuteActionPlans, &utils.AttrsExecuteActionPlans{
					ActionPlanIDs: cdrS.cgrCfg.SchedulerCfg().DynaprepaidActionPlans,
					AccountID:     cdr.Account, Tenant: cdr.Tenant},
				&reply); err != nil {
				return cc, err
			}
			// execute again the Debit operation
			err = cdrS.connMgr.Call(context.TODO(), cdrS.cgrCfg.CdrsCfg().RaterConns,
				utils.ResponderDebit,
				&CallDescriptorWithAPIOpts{
					CallDescriptor: cd,
					APIOpts:        cdr.APIOpts,
				}, cc)
		}
	} else {
		err = cdrS.connMgr.Call(context.TODO(), cdrS.cgrCfg.CdrsCfg().RaterConns,
			utils.ResponderGetCost,
			&CallDescriptorWithAPIOpts{
				CallDescriptor: cd,
				APIOpts:        cdr.APIOpts,
			}, cc)
	}
	if err != nil {
		return cc, err
	}
	cdr.CostSource = utils.MetaCDRs
	return cc, nil
}

// rateCDRWithErr rates a CDR including errors
func (cdrS *CDRServer) rateCDRWithErr(cdr *CDRWithAPIOpts) (ratedCDRs []*CDR) {
	var err error
	ratedCDRs, err = cdrS.rateCDR(cdr)
	if err != nil {
		cdr.Cost = -1.0 // If there was an error, mark the CDR
		cdr.ExtraInfo = err.Error()
		ratedCDRs = []*CDR{cdr.CDR}
	}
	return
}

// refundEventCost will refund the EventCost using RefundIncrements
func (cdrS *CDRServer) refundEventCost(ec *EventCost, reqType, tor string) (rfnd bool, err error) {
	if len(cdrS.cgrCfg.CdrsCfg().RaterConns) == 0 {
		return false, utils.NewErrNotConnected(utils.RALService)
	}
	if ec == nil || !utils.AccountableRequestTypes.Has(reqType) {
		return // non refundable
	}
	cd := ec.AsRefundIncrements(tor)
	if cd == nil || len(cd.Increments) == 0 {
		return
	}
	var acnt Account
	if err = cdrS.connMgr.Call(context.TODO(), cdrS.cgrCfg.CdrsCfg().RaterConns,
		utils.ResponderRefundIncrements,
		&CallDescriptorWithAPIOpts{CallDescriptor: cd}, &acnt); err != nil {
		return
	}
	return true, nil
}

// chrgrSProcessEvent forks CGREventWithOpts into multiples based on matching ChargerS profiles
func (cdrS *CDRServer) chrgrSProcessEvent(cgrEv *utils.CGREvent) (cgrEvs []*utils.CGREvent, err error) {
	var chrgrs []*ChrgSProcessEventReply
	if err = cdrS.connMgr.Call(context.TODO(), cdrS.cgrCfg.CdrsCfg().ChargerSConns,
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
func (cdrS *CDRServer) attrSProcessEvent(cgrEv *utils.CGREvent) (err error) {
	var rplyEv AttrSProcessEventReply
	if cgrEv.APIOpts == nil {
		cgrEv.APIOpts = make(map[string]any)
	}
	cgrEv.APIOpts[utils.MetaSubsys] = utils.MetaCDRs
	ctx, has := cgrEv.APIOpts[utils.OptsContext]
	cgrEv.APIOpts[utils.OptsContext] = utils.FirstNonEmpty(
		utils.IfaceAsString(ctx),
		utils.MetaCDRs)
	if err = cdrS.connMgr.Call(context.TODO(), cdrS.cgrCfg.CdrsCfg().AttributeSConns,
		utils.AttributeSv1ProcessEvent,
		cgrEv, &rplyEv); err == nil && len(rplyEv.AlteredFields) != 0 {
		*cgrEv = *rplyEv.CGREvent
		if !has && utils.IfaceAsString(cgrEv.APIOpts[utils.OptsContext]) == utils.MetaCDRs {
			delete(cgrEv.APIOpts, utils.OptsContext)
		}
	} else if err != nil &&
		err.Error() == utils.ErrNotFound.Error() {
		err = nil // cancel ErrNotFound
	}
	return
}

// thdSProcessEvent will send the event to ThresholdS
func (cdrS *CDRServer) thdSProcessEvent(cgrEv *utils.CGREvent) (err error) {
	var tIDs []string
	// we clone the CGREvent so we can add EventType without being propagated
	thArgs := cgrEv.Clone()
	if thArgs.APIOpts == nil {
		thArgs.APIOpts = make(map[string]any)
	}
	thArgs.APIOpts[utils.MetaEventType] = utils.CDR
	if err = cdrS.connMgr.Call(context.TODO(), cdrS.cgrCfg.CdrsCfg().ThresholdSConns,
		utils.ThresholdSv1ProcessEvent,
		thArgs, &tIDs); err != nil &&
		err.Error() == utils.ErrNotFound.Error() {
		err = nil // NotFound is not considered error
	}
	return
}

// statSProcessEvent will send the event to StatS
func (cdrS *CDRServer) statSProcessEvent(cgrEv *utils.CGREvent) (err error) {
	var reply []string
	statArgs := cgrEv.Clone()
	if err = cdrS.connMgr.Call(context.TODO(), cdrS.cgrCfg.CdrsCfg().StatSConns,
		utils.StatSv1ProcessEvent,
		statArgs, &reply); err != nil &&
		err.Error() == utils.ErrNotFound.Error() {
		err = nil // NotFound is not considered error
	}
	return
}

// eeSProcessEvent will process the event with the EEs component
func (cdrS *CDRServer) eeSProcessEvent(cgrEv *CGREventWithEeIDs) (err error) {
	var reply map[string]map[string]any
	if err = cdrS.connMgr.Call(context.TODO(), cdrS.cgrCfg.CdrsCfg().EEsConns,
		utils.EeSv1ProcessEvent,
		cgrEv, &reply); err != nil &&
		err.Error() == utils.ErrNotFound.Error() {
		err = nil // NotFound is not considered error
	}
	return
}

// cdrProcessingArgs holds the arguments for processing CDR events.
type cdrProcessingArgs struct {
	attrS     bool
	chrgS     bool
	refund    bool
	ralS      bool
	store     bool
	reRate    bool
	export    bool
	thdS      bool
	stS       bool
	reprocess bool
}

// newCDRProcessingArgs initializes processing arguments from config and overrides them with provided flags.
func newCDRProcessingArgs(cfg *config.CdrsCfg, flags utils.FlagsWithParams, opts map[string]any) (*cdrProcessingArgs, error) {
	args := &cdrProcessingArgs{
		attrS:  len(cfg.AttributeSConns) != 0,
		chrgS:  len(cfg.ChargerSConns) != 0,
		store:  cfg.StoreCdrs,
		export: len(cfg.OnlineCDRExports) != 0 || len(cfg.EEsConns) != 0,
		thdS:   len(cfg.ThresholdSConns) != 0,
		stS:    len(cfg.StatSConns) != 0,
		ralS:   len(cfg.RaterConns) != 0,
	}
	var err error
	if v, has := opts[utils.OptsAttributeS]; has {
		if args.attrS, err = utils.IfaceAsBool(v); err != nil {
			return nil, err
		}
	}
	if flags.Has(utils.MetaAttributes) {
		args.attrS = flags.GetBool(utils.MetaAttributes)
	}
	if v, has := opts[utils.OptsChargerS]; has {
		if args.chrgS, err = utils.IfaceAsBool(v); err != nil {
			return nil, err
		}
	}
	if flags.Has(utils.MetaChargers) {
		args.chrgS = flags.GetBool(utils.MetaChargers)
	}
	if flags.Has(utils.MetaStore) {
		args.store = flags.GetBool(utils.MetaStore)
	}
	if flags.Has(utils.MetaExport) {
		args.export = flags.GetBool(utils.MetaExport)
	}
	if v, has := opts[utils.OptsThresholdS]; has {
		if args.thdS, err = utils.IfaceAsBool(v); err != nil {
			return nil, err
		}
	}
	if flags.Has(utils.MetaThresholds) {
		args.thdS = flags.GetBool(utils.MetaThresholds)
	}
	if v, has := opts[utils.OptsStatS]; has {
		if args.stS, err = utils.IfaceAsBool(v); err != nil {
			return nil, err
		}
	}
	if flags.Has(utils.MetaStats) {
		args.stS = flags.GetBool(utils.MetaStats)
	}
	if v, has := opts[utils.OptsRerate]; has {
		if args.reRate, err = utils.IfaceAsBool(v); err != nil {
			return nil, err
		}
	}
	if flags.Has(utils.MetaRerate) {
		args.reRate = flags.GetBool(utils.MetaRerate)
	}
	if args.reRate {
		args.ralS = true
		args.refund = true
	}
	if v, has := opts[utils.OptsRefund]; has {
		if args.refund, err = utils.IfaceAsBool(v); err != nil {
			return nil, err
		}
	}
	if flags.Has(utils.MetaRefund) {
		args.refund = flags.GetBool(utils.MetaRefund)
	}
	if args.refund && !args.reRate {
		args.ralS = false
	}
	if args.refund {
		args.reprocess = true
	}
	if flags.Has(utils.MetaReprocess) {
		args.reprocess = flags.GetBool(utils.MetaReprocess)
	}
	if v, has := opts[utils.OptsRALs]; has {
		if args.ralS, err = utils.IfaceAsBool(v); err != nil {
			return nil, err
		}
	}
	if flags.Has(utils.MetaRALs) {
		args.ralS = flags.GetBool(utils.MetaRALs)
	}
	return args, nil
}

// newCDRProcessingArgsNoCfg initializes processing arguments by taking flags from provided FlagsWithParams and APIOpts, without taking them from configs.
func newCDRProcessingArgsNoCfg(flags utils.FlagsWithParams, opts map[string]any) (*cdrProcessingArgs, error) {
	args := new(cdrProcessingArgs)
	var err error
	if v, has := opts[utils.OptsAttributeS]; has {
		if args.attrS, err = utils.IfaceAsBool(v); err != nil {
			return nil, err
		}
	}
	if flags.Has(utils.MetaAttributes) {
		args.attrS = flags.GetBool(utils.MetaAttributes)
	}
	if v, has := opts[utils.OptsChargerS]; has {
		if args.chrgS, err = utils.IfaceAsBool(v); err != nil {
			return nil, err
		}
	}
	if flags.Has(utils.MetaChargers) {
		args.chrgS = flags.GetBool(utils.MetaChargers)
	}
	if flags.Has(utils.MetaStore) {
		args.store = flags.GetBool(utils.MetaStore)
	}
	if flags.Has(utils.MetaExport) {
		args.export = flags.GetBool(utils.MetaExport)
	}
	if v, has := opts[utils.OptsThresholdS]; has {
		if args.thdS, err = utils.IfaceAsBool(v); err != nil {
			return nil, err
		}
	}
	if flags.Has(utils.MetaThresholds) {
		args.thdS = flags.GetBool(utils.MetaThresholds)
	}
	if v, has := opts[utils.OptsStatS]; has {
		if args.stS, err = utils.IfaceAsBool(v); err != nil {
			return nil, err
		}
	}
	if flags.Has(utils.MetaStats) {
		args.stS = flags.GetBool(utils.MetaStats)
	}
	if v, has := opts[utils.OptsRerate]; has {
		if args.reRate, err = utils.IfaceAsBool(v); err != nil {
			return nil, err
		}
	}
	if flags.Has(utils.MetaRerate) {
		args.reRate = flags.GetBool(utils.MetaRerate)
	}
	if args.reRate {
		args.ralS = true
		args.refund = true
	}
	if v, has := opts[utils.OptsRefund]; has {
		if args.refund, err = utils.IfaceAsBool(v); err != nil {
			return nil, err
		}
	}
	if flags.Has(utils.MetaRefund) {
		args.refund = flags.GetBool(utils.MetaRefund)
	}
	if args.refund && !args.reRate {
		args.ralS = false
	}
	if args.refund {
		args.reprocess = true
	}
	if flags.Has(utils.MetaReprocess) {
		args.reprocess = flags.GetBool(utils.MetaReprocess)
	}
	if v, has := opts[utils.OptsRALs]; has {
		if args.ralS, err = utils.IfaceAsBool(v); err != nil {
			return nil, err
		}
	}
	if flags.Has(utils.MetaRALs) {
		args.ralS = flags.GetBool(utils.MetaRALs)
	}
	return args, nil
}

// processEvent processes a CGREvent based on arguments
// in case of partially executed, both error and evs will be returned
func (cdrS *CDRServer) processEvents(evs []*utils.CGREvent, args *cdrProcessingArgs) (outEvs []*utils.EventWithFlags, err error) {
	if args.attrS {
		for _, ev := range evs {
			if err = cdrS.attrSProcessEvent(ev); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: <%s> processing event %+v with %s",
						utils.CDRs, err.Error(), utils.ToJSON(ev), utils.AttributeS))
				err = utils.ErrPartiallyExecuted
				return
			}
		}
	}
	var cgrEvs []*utils.CGREvent
	if args.chrgS {
		for _, ev := range evs {
			var chrgEvs []*utils.CGREvent
			if chrgEvs, err = cdrS.chrgrSProcessEvent(ev); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: <%s> processing event %+v with %s",
						utils.CDRs, err.Error(), utils.ToJSON(ev), utils.ChargerS))
				err = utils.ErrPartiallyExecuted
				return
			} else {
				cgrEvs = append(cgrEvs, chrgEvs...)
			}
		}
	} else { // ChargerS not requested, charge the original event
		cgrEvs = evs
	}
	// Check if the unique ID was not already processed
	if !args.reprocess {
		for _, cgrEv := range cgrEvs {
			me := MapEvent(cgrEv.Event)
			if !me.HasField(utils.CGRID) { // try to compute the CGRID if missing
				me[utils.CGRID] = utils.Sha1(
					me.GetStringIgnoreErrors(utils.OriginID),
					me.GetStringIgnoreErrors(utils.OriginHost),
				)
			}
			uID := utils.ConcatenatedKey(
				me.GetStringIgnoreErrors(utils.CGRID),
				me.GetStringIgnoreErrors(utils.RunID),
			)
			if Cache.HasItem(utils.CacheCDRIDs, uID) && !args.reRate {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: <%s> processing event %+v with %s",
						utils.CDRs, utils.ErrExists, utils.ToJSON(cgrEv), utils.CacheS))
				return nil, utils.ErrExists
			}
			if errCh := Cache.Set(utils.CacheCDRIDs, uID, true, nil,
				cacheCommit(utils.NonTransactional), utils.NonTransactional); errCh != nil {
				return nil, errCh
			}
		}
	}
	// Populate CDR list out of events
	cdrs := make([]*CDR, len(cgrEvs))
	if args.refund || args.ralS || args.store || args.reRate || args.export {
		for i, cgrEv := range cgrEvs {
			if args.refund {
				if _, has := cgrEv.Event[utils.CostDetails]; !has {
					// if CostDetails is not populated or is nil, look for it inside the previously stored cdr
					var cgrID string // prepare CGRID to filter for previous CDR
					if val, has := cgrEv.Event[utils.CGRID]; !has {
						cgrID = utils.Sha1(utils.IfaceAsString(cgrEv.Event[utils.OriginID]),
							utils.IfaceAsString(cgrEv.Event[utils.OriginHost]))
					} else {
						cgrID = utils.IfaceAsString(val)
					}
					var prevCDRs []*CDR // only one should be returned
					if prevCDRs, _, err = cdrS.cdrDb.GetCDRs(
						&utils.CDRsFilter{CGRIDs: []string{cgrID},
							RunIDs: []string{utils.IfaceAsString(cgrEv.Event[utils.RunID])}}, false); err != nil {
						utils.Logger.Err(
							fmt.Sprintf("<%s> could not retrieve previously stored CDR, error: <%s>",
								utils.CDRs, err.Error()))
						err = utils.ErrPartiallyExecuted
						return
					} else {
						cgrEv.Event[utils.CostDetails] = prevCDRs[0].CostDetails
					}
				}
			} else if args.reRate {
				// Force rerate by removing CostDetails to avoid marking as already rated.
				delete(cgrEv.Event, utils.CostDetails)
			}
			if cdrs[i], err = NewMapEvent(cgrEv.Event).AsCDR(cdrS.cgrCfg,
				cgrEv.Tenant, cdrS.cgrCfg.GeneralCfg().DefaultTimezone); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: <%s> converting event %+v to CDR",
						utils.CDRs, err.Error(), utils.ToJSON(cgrEv)))
				err = utils.ErrPartiallyExecuted
				return
			}
		}
	}
	procFlgs := make([]utils.StringSet, len(cgrEvs)) // will save the flags for the reply here
	for i := range cgrEvs {
		procFlgs[i] = utils.NewStringSet(nil)
	}
	if args.refund {
		for i, cdr := range cdrs {
			if rfnd, errRfd := cdrS.refundEventCost(cdr.CostDetails,
				cdr.RequestType, cdr.ToR); errRfd != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: <%s> refunding CDR %+v",
						utils.CDRs, errRfd.Error(), utils.ToJSON(cdr)))
			} else if rfnd {
				cdr.CostDetails = nil // this makes sure that the rater will recalculate (and debit) the cost
				procFlgs[i].Add(utils.MetaRefund)
			}
		}
	}
	if args.ralS {
		for i, cdr := range cdrs {
			for j, rtCDR := range cdrS.rateCDRWithErr(
				&CDRWithAPIOpts{
					CDR:     cdr,
					APIOpts: cgrEvs[i].APIOpts,
				}) {
				cgrEv := rtCDR.AsCGREvent()
				cgrEv.APIOpts = cgrEvs[i].APIOpts
				if j == 0 { // the first CDR will replace the events we got already as a small optimization
					cdrs[i] = rtCDR
					cgrEvs[i] = cgrEv
				} else {
					cdrs = append(cdrs, cdr)
					cgrEvs = append(cgrEvs, cgrEv)
				}
			}
		}
	}
	if args.store {
		refundCDRCosts := func() { // will be used to refund all CDRs on errors
			for _, cdr := range cdrs { // refund what we have charged since duplicates are not allowed
				if _, errRfd := cdrS.refundEventCost(cdr.CostDetails,
					cdr.RequestType, cdr.ToR); errRfd != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> error: <%s> refunding CDR %+v",
							utils.CDRs, errRfd.Error(), utils.ToJSON(cdr)))
				}
			}
		}
		for _, cdr := range cdrs {
			if err = cdrS.cdrDb.SetCDR(cdr, false); err != nil {
				if err != utils.ErrExists || !args.reRate {
					refundCDRCosts()
					return
				}
				if err = cdrS.cdrDb.SetCDR(cdr, true); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> error: <%s> updating CDR %+v",
							utils.CDRs, err.Error(), utils.ToJSON(cdr)))
					err = utils.ErrPartiallyExecuted
					return
				}
			}
		}
	}
	var partiallyExecuted bool // from here actions are optional and a general error is returned
	if args.export {
		if len(cdrS.cgrCfg.CdrsCfg().EEsConns) != 0 {
			for _, cgrEv := range cgrEvs {
				evWithOpts := &CGREventWithEeIDs{
					CGREvent: cgrEv,
					EeIDs:    cdrS.cgrCfg.CdrsCfg().OnlineCDRExports,
				}
				if err = cdrS.eeSProcessEvent(evWithOpts); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> error: <%s> exporting cdr %+v",
							utils.CDRs, err.Error(), utils.ToJSON(evWithOpts)))
					partiallyExecuted = true
				}
			}
		}
	}
	if args.thdS {
		for _, cgrEv := range cgrEvs {
			if err = cdrS.thdSProcessEvent(cgrEv); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: <%s> processing event %+v with %s",
						utils.CDRs, err.Error(), utils.ToJSON(cgrEv), utils.ThresholdS))
				partiallyExecuted = true
			}
		}
	}
	if args.stS {
		for _, cgrEv := range cgrEvs {
			if err = cdrS.statSProcessEvent(cgrEv); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: <%s> processing event %+v with %s",
						utils.CDRs, err.Error(), utils.ToJSON(cgrEv), utils.StatS))
				partiallyExecuted = true
			}
		}
	}
	if partiallyExecuted {
		err = utils.ErrPartiallyExecuted
	}
	outEvs = make([]*utils.EventWithFlags, len(cgrEvs))
	for i, cgrEv := range cgrEvs {
		outEvs[i] = &utils.EventWithFlags{
			Flags: procFlgs[i].AsSlice(),
			Event: cgrEv.Event,
		}
	}
	return
}

// V1ProcessCDR processes a CDR
func (cdrS *CDRServer) V1ProcessCDR(ctx *context.Context, cdr *CDRWithAPIOpts, reply *string) (err error) {
	if cdr.CGRID == utils.EmptyString { // Populate CGRID if not present
		cdr.ComputeCGRID()
	}
	// RPC caching
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.CDRsV1ProcessCDR, cdr.CGRID, cdr.RunID)
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
		defer Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	if cdr.RequestType == utils.EmptyString {
		cdr.RequestType = cdrS.cgrCfg.GeneralCfg().DefaultReqType
	}
	if cdr.Tenant == utils.EmptyString {
		cdr.Tenant = cdrS.cgrCfg.GeneralCfg().DefaultTenant
	}
	if cdr.Category == utils.EmptyString {
		cdr.Category = cdrS.cgrCfg.GeneralCfg().DefaultCategory
	}
	if cdr.Subject == utils.EmptyString { // Use account information as rating subject if missing
		cdr.Subject = cdr.Account
	}
	if cdr.RunID == utils.EmptyString {
		cdr.RunID = utils.MetaDefault
	}
	cgrEv := cdr.AsCGREvent()
	cgrEv.APIOpts = cdr.APIOpts

	procArgs, err := newCDRProcessingArgs(cdrS.cgrCfg.CdrsCfg(), nil, nil)
	if err != nil {
		return fmt.Errorf("failed to configure processing args: %v", err)
	}
	procArgs.chrgS = procArgs.chrgS && !cdr.PreRated
	procArgs.ralS = !cdr.PreRated // rate the CDR if it's not PreRated

	if _, err = cdrS.processEvents([]*utils.CGREvent{cgrEv}, procArgs); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// ArgV1ProcessEvent is the CGREvent with proccesing Flags
type ArgV1ProcessEvent struct {
	Flags []string
	utils.CGREvent
	clnb bool //rpcclonable
}

// SetCloneable sets if the args should be clonned on internal connections
func (attr *ArgV1ProcessEvent) SetCloneable(rpcCloneable bool) {
	attr.clnb = rpcCloneable
}

// RPCClone implements rpcclient.RPCCloner interface
func (attr *ArgV1ProcessEvent) RPCClone() (any, error) {
	if !attr.clnb {
		return attr, nil
	}
	return attr.Clone(), nil
}

// Clone creates a clone of the object
func (attr *ArgV1ProcessEvent) Clone() *ArgV1ProcessEvent {
	var flags []string
	if attr.Flags != nil {
		flags = make([]string, len(attr.Flags))
		copy(flags, attr.Flags)

	}
	return &ArgV1ProcessEvent{
		Flags:    flags,
		CGREvent: *attr.CGREvent.Clone(),
	}
}

// V1ProcessEvent will process the CGREvent
func (cdrS *CDRServer) V1ProcessEvent(ctx *context.Context, arg *ArgV1ProcessEvent, reply *string) (err error) {
	if arg.CGREvent.ID == utils.EmptyString {
		arg.CGREvent.ID = utils.GenUUID()
	}
	if arg.CGREvent.Tenant == utils.EmptyString {
		arg.CGREvent.Tenant = cdrS.cgrCfg.GeneralCfg().DefaultTenant
	}
	// RPC caching
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.CDRsV1ProcessEvent, arg.CGREvent.ID)
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
		defer Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	// Compute processing arguments based on flags and configuration.
	flgs := utils.FlagsWithParamsFromSlice(arg.Flags)
	procArgs, err := newCDRProcessingArgs(cdrS.cgrCfg.CdrsCfg(), flgs, arg.APIOpts)
	if err != nil {
		return fmt.Errorf("failed to configure processing args: %v", err)
	}

	if _, err = cdrS.processEvents([]*utils.CGREvent{&arg.CGREvent}, procArgs); err != nil {
		return
	}
	*reply = utils.OK
	return nil
}

// V2ProcessEvent has the same logic with V1ProcessEvent except it adds the proccessed events to the reply
func (cdrS *CDRServer) V2ProcessEvent(ctx *context.Context, arg *ArgV1ProcessEvent, evs *[]*utils.EventWithFlags) (err error) {
	if arg.ID == "" {
		arg.ID = utils.GenUUID()
	}
	// RPC caching
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.CDRsV2ProcessEvent, arg.CGREvent.ID)
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)

		if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*evs = *cachedResp.Result.(*[]*utils.EventWithFlags)
			}
			return cachedResp.Error
		}
		defer Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: evs, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	// Compute processing arguments based on flags and configuration.
	flgs := utils.FlagsWithParamsFromSlice(arg.Flags)
	procArgs, err := newCDRProcessingArgs(cdrS.cgrCfg.CdrsCfg(), flgs, arg.APIOpts)
	if err != nil {
		return fmt.Errorf("failed to configure processing args: %v", err)
	}

	var procEvs []*utils.EventWithFlags
	if procEvs, err = cdrS.processEvents([]*utils.CGREvent{&arg.CGREvent}, procArgs); err != nil {
		return
	}
	*evs = procEvs
	return nil
}

// ArgV1ProcessEvents defines the structure for holding CGREvents about to be processed by CDRsV1.ProcessEvents.
type ArgV1ProcessEvents struct {
	Flags     []string
	CGREvents []*utils.CGREvent
	APIOpts   map[string]any
	clnb      bool //rpcclonable
}

// SetCloneable sets if the args should be clonned on internal connections
func (attr *ArgV1ProcessEvents) SetCloneable(rpcCloneable bool) {
	attr.clnb = rpcCloneable
}

// RPCClone implements rpcclient.RPCCloner interface
func (attr *ArgV1ProcessEvents) RPCClone() (any, error) {
	if !attr.clnb {
		return attr, nil
	}
	return attr.Clone(), nil
}

// Clone creates a clone of the object
func (attr *ArgV1ProcessEvents) Clone() *ArgV1ProcessEvents {
	cln := &ArgV1ProcessEvents{
		Flags:     slices.Clone(attr.Flags),
		CGREvents: make([]*utils.CGREvent, 0, len(attr.CGREvents)),
	}
	for _, cgrEv := range attr.CGREvents {
		cln.CGREvents = append(cln.CGREvents, cgrEv.Clone())
	}
	if attr.APIOpts == nil {
		return cln
	}
	cln.APIOpts = make(map[string]any)
	for key, value := range attr.APIOpts {
		cln.APIOpts[key] = value
	}
	return cln
}

// V1ProcessEvents is similar to V1ProcessEvent but it can accept multiple CGREvents inside the parameter.
func (cdrS *CDRServer) V1ProcessEvents(ctx *context.Context, arg *ArgV1ProcessEvents, reply *string) (err error) {

	// Populate ID and Tenant of CGREvents if missing.
	for _, cgrEv := range arg.CGREvents {
		if cgrEv.ID == "" {
			cgrEv.ID = utils.GenUUID()
		}
		if cgrEv.Tenant == "" {
			cgrEv.Tenant = cdrS.cgrCfg.GeneralCfg().DefaultTenant
		}
	}

	// Compute processing arguments based on flags and configuration.
	flgs := utils.FlagsWithParamsFromSlice(arg.Flags)
	procArgs, err := newCDRProcessingArgs(cdrS.cgrCfg.CdrsCfg(), flgs, arg.APIOpts)
	if err != nil {
		return fmt.Errorf("failed to configure processing args: %v", err)
	}

	if _, err = cdrS.processEvents(arg.CGREvents, procArgs); err != nil {
		return
	}
	*reply = utils.OK
	return nil
}

// V1StoreSessionCost handles storing of the cost into session_costs table
func (cdrS *CDRServer) V1StoreSessionCost(ctx *context.Context, attr *AttrCDRSStoreSMCost, reply *string) (err error) {
	if attr.Cost.CGRID == "" {
		return utils.NewCGRError(utils.CDRsCtx,
			utils.MandatoryIEMissingCaps, fmt.Sprintf("%s: CGRID", utils.MandatoryInfoMissing),
			"SMCost: %+v with empty CGRID")
	}
	// RPC caching
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.CDRsV1StoreSessionCost, attr.Cost.CGRID, attr.Cost.RunID)
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
		defer Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching
	if err = cdrS.storeSMCost(attr.Cost, attr.CheckDuplicate); err != nil {
		err = utils.NewErrServerError(err)
		return
	}
	*reply = utils.OK
	return nil
}

// V2StoreSessionCost will store the SessionCost into session_costs table
func (cdrS *CDRServer) V2StoreSessionCost(ctx *context.Context, args *ArgsV2CDRSStoreSMCost, reply *string) (err error) {
	if args.Cost.CGRID == "" {
		return utils.NewCGRError(utils.CDRsCtx,
			utils.MandatoryIEMissingCaps, fmt.Sprintf("%s: CGRID", utils.MandatoryInfoMissing),
			"SMCost: %+v with empty CGRID")
	}
	// RPC caching
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.CDRsV1StoreSessionCost, args.Cost.CGRID, args.Cost.RunID)
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
		defer Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching
	cc := args.Cost.CostDetails.AsCallCost(utils.EmptyString)
	if args.Cost.CostDetails.AccountSummary != nil {
		cc.Tenant = args.Cost.CostDetails.AccountSummary.Tenant
		cc.Account = args.Cost.CostDetails.AccountSummary.ID
	}
	cc.Round()
	roundIncrements := cc.GetRoundIncrements()
	if len(roundIncrements) != 0 {
		cd := cc.CreateCallDescriptor()
		cd.CgrID = args.Cost.CGRID
		cd.RunID = args.Cost.RunID
		cd.Increments = roundIncrements
		response := new(Account)
		if err := cdrS.connMgr.Call(context.TODO(), cdrS.cgrCfg.CdrsCfg().RaterConns,
			utils.ResponderRefundRounding,
			&CallDescriptorWithAPIOpts{CallDescriptor: cd},
			response); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<CDRS> RefundRounding for cc: %+v, got error: %s",
					cc, err.Error()))
		}
		accSum := response.AsAccountSummary()
		accSum.UpdateInitialValue(cc.AccountSummary)
		cc.AccountSummary = accSum

	}
	if err = cdrS.storeSMCost(
		&SMCost{
			CGRID:       args.Cost.CGRID,
			RunID:       args.Cost.RunID,
			OriginHost:  args.Cost.OriginHost,
			OriginID:    args.Cost.OriginID,
			CostSource:  args.Cost.CostSource,
			Usage:       args.Cost.Usage,
			CostDetails: NewEventCostFromCallCost(cc, args.Cost.CGRID, args.Cost.RunID)},
		args.CheckDuplicate); err != nil {
		err = utils.NewErrServerError(err)
		return
	}
	*reply = utils.OK
	return

}

// ArgRateCDRs a cdr with extra flags
type ArgRateCDRs struct {
	Flags []string
	utils.RPCCDRsFilter
	Tenant  string
	APIOpts map[string]any
}

// V1RateCDRs is used for re-/rate CDRs which are already stored within StorDB
// FixMe: add RPC caching
func (cdrS *CDRServer) V1RateCDRs(ctx *context.Context, arg *ArgRateCDRs, reply *string) (err error) {
	var cdrFltr *utils.CDRsFilter
	if cdrFltr, err = arg.RPCCDRsFilter.AsCDRsFilter(cdrS.cgrCfg.GeneralCfg().DefaultTimezone); err != nil {
		return utils.NewErrServerError(err)
	}
	var cdrs []*CDR
	if cdrs, _, err = cdrS.cdrDb.GetCDRs(cdrFltr, false); err != nil {
		return
	}

	// Compute processing arguments based on flags and configuration.
	flgs := utils.FlagsWithParamsFromSlice(arg.Flags)
	procArgs, err := newCDRProcessingArgs(cdrS.cgrCfg.CdrsCfg(), flgs, arg.APIOpts)
	if err != nil {
		return fmt.Errorf("failed to configure processing args: %v", err)
	}
	procArgs.ralS = true

	cgrEvs := make([]*utils.CGREvent, len(cdrs))
	for i, cdr := range cdrs {
		cdr.Cost = -1 // the cost will be recalculated
		cgrEvs[i] = cdr.AsCGREvent()
		cgrEvs[i].APIOpts = arg.APIOpts
	}
	if _, err = cdrS.processEvents(cgrEvs, procArgs); err != nil {
		return utils.NewErrServerError(err)
	}

	*reply = utils.OK
	return
}

// V1ReprocessCDRs is used to reprocess CDRs which are already stored within StorDB
func (cdrS *CDRServer) V1ReprocessCDRs(ctx *context.Context, arg *ArgRateCDRs, reply *string) (err error) {
	var cdrFltr *utils.CDRsFilter
	if cdrFltr, err = arg.RPCCDRsFilter.AsCDRsFilter(cdrS.cgrCfg.GeneralCfg().DefaultTimezone); err != nil {
		return utils.NewErrServerError(err)
	}
	var cdrs []*CDR
	if cdrs, _, err = cdrS.cdrDb.GetCDRs(cdrFltr, false); err != nil {
		return
	}

	// Compute processing arguments based on flags and configuration.
	flgs := utils.FlagsWithParamsFromSlice(arg.Flags)
	procArgs, err := newCDRProcessingArgsNoCfg(flgs, arg.APIOpts)
	if err != nil {
		return fmt.Errorf("failed to configure processing args: %v", err)
	}
	procArgs.reprocess = true

	cgrEvs := make([]*utils.CGREvent, len(cdrs))
	for i, cdr := range cdrs {
		cgrEvs[i] = cdr.AsCGREvent()
		cgrEvs[i].APIOpts = arg.APIOpts
	}
	if _, err = cdrS.processEvents(cgrEvs, procArgs); err != nil {
		return utils.NewErrServerError(err)
	}

	*reply = utils.OK
	return
}

// V1ProcessExternalCDR is used to process external CDRs
func (cdrS *CDRServer) V1ProcessExternalCDR(ctx *context.Context, eCDR *ExternalCDRWithAPIOpts, reply *string) error {
	cdr, err := NewCDRFromExternalCDR(eCDR.ExternalCDR,
		cdrS.cgrCfg.GeneralCfg().DefaultTimezone)
	if err != nil {
		return err
	}
	return cdrS.V1ProcessCDR(ctx,
		&CDRWithAPIOpts{
			CDR:     cdr,
			APIOpts: eCDR.APIOpts,
		}, reply)
}

// V1GetCDRs returns CDRs from DB
func (cdrS *CDRServer) V1GetCDRs(ctx *context.Context, args utils.RPCCDRsFilterWithAPIOpts, cdrs *[]*CDR) error {
	cdrsFltr, err := args.AsCDRsFilter(cdrS.cgrCfg.GeneralCfg().DefaultTimezone)
	if err != nil {
		if err.Error() != utils.NotFoundCaps {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	qryCDRs, _, err := cdrS.cdrDb.GetCDRs(cdrsFltr, false)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	*cdrs = qryCDRs
	return nil
}

// V1GetCDRsCount counts CDRs from DB
func (cdrS *CDRServer) V1GetCDRsCount(ctx *context.Context, args *utils.RPCCDRsFilterWithAPIOpts, cnt *int64) error {
	cdrsFltr, err := args.AsCDRsFilter(cdrS.cgrCfg.GeneralCfg().DefaultTimezone)
	if err != nil {
		if err.Error() != utils.NotFoundCaps {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	cdrsFltr.Count = true
	_, qryCnt, err := cdrS.cdrDb.GetCDRs(cdrsFltr, false)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	*cnt = qryCnt
	return nil
}
