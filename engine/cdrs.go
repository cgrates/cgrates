/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH
This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
This program is distributed in the hope that it will be useful,
but WITHOUT MetaAny WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.
You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
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
	if err := cdrS.V1ProcessCDR(&CDRWithOpts{CDR: cdr}, &ignored); err != nil {
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
	if err := cdrS.V1ProcessCDR(&CDRWithOpts{CDR: cdr}, &ignored); err != nil {
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
		_, err := cdrS.guard.Guard(func() (interface{}, error) {
			smCosts, err := cdrS.cdrDb.GetSMCosts(smCost.CGRID, smCost.RunID, "", "")
			if err != nil && err.Error() != utils.NotFoundCaps {
				return nil, err
			}
			if len(smCosts) != 0 {
				return nil, utils.ErrExists
			}
			return nil, cdrS.cdrDb.SetSMCost(smCost)
		}, config.CgrConfig().GeneralCfg().LockingTimeout, lockKey) // FixMe: Possible deadlock with Guard from SMG session close()
		return err
	}
	return cdrS.cdrDb.SetSMCost(smCost)
}

// rateCDR will populate cost field
// Returns more than one rated CDR in case of SMCost retrieved based on prefix
func (cdrS *CDRServer) rateCDR(cdr *CDRWithOpts) ([]*CDR, error) {
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
	if utils.SliceHasMember([]string{utils.MetaPrepaid, utils.Prepaid}, cdr.RequestType) &&
		(cdr.Usage != 0 || hasLastUsed) && cdr.CostDetails == nil {
		// ToDo: Get rid of Prepaid as soon as we don't want to support it backwards
		// Should be previously calculated and stored in DB
		fib := utils.Fib()
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
				time.Sleep(time.Duration(fib()) * time.Second)
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
					if qryCC, err = cdrS.getCostFromRater(&CDRWithOpts{CDR: cdrClone}); err != nil {
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
		if cdr.Usage == cdr.CostDetails.GetUsage() { // Costs were previously calculated, make sure they cover the full usage
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
func (cdrS *CDRServer) getCostFromRater(cdr *CDRWithOpts) (*CallCost, error) {
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
		TimeStart:       timeStart,
		TimeEnd:         timeStart.Add(cdr.Usage),
		DurationIndex:   cdr.Usage,
		PerformRounding: true,
	}
	if reqTypes.Has(cdr.RequestType) { // Prepaid - Cost can be recalculated in case of missing records from SM
		err = cdrS.connMgr.Call(cdrS.cgrCfg.CdrsCfg().RaterConns, nil,
			utils.ResponderDebit,
			&CallDescriptorWithOpts{
				CallDescriptor: cd,
				Opts:           cdr.Opts,
			}, cc)
		if err != nil && err.Error() == utils.ErrAccountNotFound.Error() &&
			cdr.RequestType == utils.MetaDynaprepaid {
			var reply string
			// execute the actionPlan configured in RalS
			if err = cdrS.connMgr.Call(cdrS.cgrCfg.CdrsCfg().SchedulerConns, nil,
				utils.SchedulerSv1ExecuteActionPlans, &utils.AttrsExecuteActionPlans{
					ActionPlanIDs: cdrS.cgrCfg.RalsCfg().DynaprepaidActionPlans,
					AccountID:     cdr.Account, Tenant: cdr.Tenant},
				&reply); err != nil {
				return cc, err
			}
			// execute again the Debit operation
			err = cdrS.connMgr.Call(cdrS.cgrCfg.CdrsCfg().RaterConns, nil,
				utils.ResponderDebit,
				&CallDescriptorWithOpts{
					CallDescriptor: cd,
					Opts:           cdr.Opts,
				}, cc)
		}
	} else {
		err = cdrS.connMgr.Call(cdrS.cgrCfg.CdrsCfg().RaterConns, nil,
			utils.ResponderGetCost,
			&CallDescriptorWithOpts{
				CallDescriptor: cd,
				Opts:           cdr.Opts,
			}, cc)
	}
	if err != nil {
		return cc, err
	}
	cdr.CostSource = utils.MetaCDRs
	return cc, nil
}

// rateCDRWithErr rates a CDR including errors
func (cdrS *CDRServer) rateCDRWithErr(cdr *CDRWithOpts) (ratedCDRs []*CDR) {
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
	if err = cdrS.connMgr.Call(cdrS.cgrCfg.CdrsCfg().RaterConns, nil,
		utils.ResponderRefundIncrements,
		&CallDescriptorWithOpts{CallDescriptor: cd}, &acnt); err != nil {
		return
	}
	return true, nil
}

// chrgrSProcessEvent forks CGREventWithOpts into multiples based on matching ChargerS profiles
func (cdrS *CDRServer) chrgrSProcessEvent(cgrEv *utils.CGREventWithOpts) (cgrEvs []*utils.CGREventWithOpts, err error) {
	var chrgrs []*ChrgSProcessEventReply
	if err = cdrS.connMgr.Call(cdrS.cgrCfg.CdrsCfg().ChargerSConns, nil,
		utils.ChargerSv1ProcessEvent,
		cgrEv, &chrgrs); err != nil {
		return
	}
	if len(chrgrs) == 0 {
		return
	}
	cgrEvs = make([]*utils.CGREventWithOpts, len(chrgrs))
	for i, cgrPrfl := range chrgrs {
		cgrEvs[i] = &utils.CGREventWithOpts{
			CGREvent: cgrPrfl.CGREvent,
			Opts:     cgrPrfl.Opts,
		}
	}
	return
}

// attrSProcessEvent will send the event to StatS if the connection is configured
func (cdrS *CDRServer) attrSProcessEvent(cgrEv *utils.CGREventWithOpts) (err error) {
	var rplyEv AttrSProcessEventReply
	if cgrEv.Opts == nil {
		cgrEv.Opts = make(map[string]interface{})
	}
	cgrEv.Opts[utils.Subsys] = utils.MetaCDRs
	var processRuns *int
	if val, has := cgrEv.Opts[utils.OptsAttributesProcessRuns]; has {
		if v, err := utils.IfaceAsTInt64(val); err == nil {
			processRuns = utils.IntPointer(int(v))
		}
	}
	attrArgs := &AttrArgsProcessEvent{
		Context: utils.StringPointer(utils.FirstNonEmpty(
			utils.IfaceAsString(cgrEv.Opts[utils.OptsContext]),
			utils.MetaCDRs)),
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: cgrEv.CGREvent,
			Opts:     cgrEv.Opts,
		},
		ProcessRuns: processRuns,
	}
	if err = cdrS.connMgr.Call(cdrS.cgrCfg.CdrsCfg().AttributeSConns, nil,
		utils.AttributeSv1ProcessEvent,
		attrArgs, &rplyEv); err == nil && len(rplyEv.AlteredFields) != 0 {
		cgrEv.CGREvent = rplyEv.CGREvent
		cgrEv.Opts = rplyEv.Opts
	} else if err != nil &&
		err.Error() == utils.ErrNotFound.Error() {
		err = nil // cancel ErrNotFound
	}
	return
}

// thdSProcessEvent will send the event to ThresholdS
func (cdrS *CDRServer) thdSProcessEvent(cgrEv *utils.CGREventWithOpts) (err error) {
	var tIDs []string
	// we clone the CGREvent so we can add EventType without being propagated
	thArgs := &ThresholdsArgsProcessEvent{
		CGREventWithOpts: cgrEv.Clone(),
	}
	if thArgs.Opts == nil {
		thArgs.Opts = make(map[string]interface{})
	}
	thArgs.Opts[utils.MetaEventType] = utils.CDR
	if err = cdrS.connMgr.Call(cdrS.cgrCfg.CdrsCfg().ThresholdSConns, nil,
		utils.ThresholdSv1ProcessEvent,
		thArgs, &tIDs); err != nil &&
		err.Error() == utils.ErrNotFound.Error() {
		err = nil // NotFound is not considered error
	}
	return
}

// statSProcessEvent will send the event to StatS
func (cdrS *CDRServer) statSProcessEvent(cgrEv *utils.CGREventWithOpts) (err error) {
	var reply []string
	statArgs := &StatsArgsProcessEvent{
		CGREventWithOpts: cgrEv.Clone(),
	}
	if err = cdrS.connMgr.Call(cdrS.cgrCfg.CdrsCfg().StatSConns, nil,
		utils.StatSv1ProcessEvent,
		statArgs, &reply); err != nil &&
		err.Error() == utils.ErrNotFound.Error() {
		err = nil // NotFound is not considered error
	}
	return
}

// eeSProcessEvent will process the event with the EEs component
func (cdrS *CDRServer) eeSProcessEvent(cgrEv *utils.CGREventWithEeIDs) (err error) {
	var reply map[string]map[string]interface{}
	if err = cdrS.connMgr.Call(cdrS.cgrCfg.CdrsCfg().EEsConns, nil,
		utils.EeSv1ProcessEvent,
		cgrEv, &reply); err != nil &&
		err.Error() == utils.ErrNotFound.Error() {
		err = nil // NotFound is not considered error
	}
	return
}

// processEvent processes a CGREvent based on arguments
// in case of partially executed, both error and evs will be returned
func (cdrS *CDRServer) processEvent(ev *utils.CGREventWithOpts,
	chrgS, attrS, refund, ralS, store, reRate, export, thdS, stS bool) (evs []*utils.EventWithFlags, err error) {
	if attrS {
		if err = cdrS.attrSProcessEvent(ev); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: <%s> processing event %+v with %s",
					utils.CDRs, err.Error(), utils.ToJSON(ev), utils.AttributeS))
			err = utils.ErrPartiallyExecuted
			return
		}
	}
	var cgrEvs []*utils.CGREventWithOpts
	if chrgS {
		if cgrEvs, err = cdrS.chrgrSProcessEvent(ev); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: <%s> processing event %+v with %s",
					utils.CDRs, err.Error(), utils.ToJSON(ev), utils.ChargerS))
			err = utils.ErrPartiallyExecuted
			return
		}
	} else { // ChargerS not requested, charge the original event
		cgrEvs = []*utils.CGREventWithOpts{ev}
	}
	// Check if the unique ID was not already processed
	if !refund {
		for _, cgrEv := range cgrEvs {
			me := MapEvent(cgrEv.CGREvent.Event)
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
			if Cache.HasItem(utils.CacheCDRIDs, uID) && !reRate {
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
	if refund || ralS || store || reRate || export {
		for i, cgrEv := range cgrEvs {
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
	if refund {
		for i, cdr := range cdrs {
			if rfnd, errRfd := cdrS.refundEventCost(cdr.CostDetails,
				cdr.RequestType, cdr.ToR); errRfd != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: <%s> refunding CDR %+v",
						utils.CDRs, errRfd.Error(), utils.ToJSON(cdr)))
			} else if rfnd {
				procFlgs[i].Add(utils.MetaRefund)
			}
		}
	}
	if ralS {
		for i, cdr := range cdrs {
			for j, rtCDR := range cdrS.rateCDRWithErr(
				&CDRWithOpts{
					CDR:  cdr,
					Opts: ev.Opts,
				}) {
				cgrEv := &utils.CGREventWithOpts{
					CGREvent: rtCDR.AsCGREvent(),
					Opts:     cgrEvs[i].Opts,
				}
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
	if store {
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
		for i, cdr := range cdrs {
			if err = cdrS.cdrDb.SetCDR(cdr, false); err != nil {
				if err != utils.ErrExists || !reRate {
					refundCDRCosts()
					return
				}
				// CDR was found in StorDB
				// reRate is allowed, refund the previous CDR
				var prevCDRs []*CDR // only one should be returned
				if prevCDRs, _, err = cdrS.cdrDb.GetCDRs(
					&utils.CDRsFilter{CGRIDs: []string{cdr.CGRID},
						RunIDs: []string{cdr.RunID}}, false); err != nil {
					refundCDRCosts()
					return
				}
				var rfnd bool
				if rfnd, err = cdrS.refundEventCost(prevCDRs[0].CostDetails,
					cdr.RequestType, cdr.ToR); err != nil {
					refundCDRCosts()
					return
				} else if rfnd {
					procFlgs[i].Add(utils.MetaRefund)
				}
				// after refund we can force update
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
	if export {
		if len(cdrS.cgrCfg.CdrsCfg().EEsConns) != 0 {
			for _, cgrEv := range cgrEvs {
				evWithOpts := &utils.CGREventWithEeIDs{
					CGREventWithOpts: &utils.CGREventWithOpts{
						CGREvent: cgrEv.CGREvent,
						Opts:     cgrEv.Opts,
					},
					EeIDs: cdrS.cgrCfg.CdrsCfg().OnlineCDRExports,
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
	if thdS {
		for _, cgrEv := range cgrEvs {
			if err = cdrS.thdSProcessEvent(cgrEv); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: <%s> processing event %+v with %s",
						utils.CDRs, err.Error(), utils.ToJSON(cgrEv), utils.ThresholdS))
				partiallyExecuted = true
			}
		}
	}
	if stS {
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
	evs = make([]*utils.EventWithFlags, len(cgrEvs))
	for i, cgrEv := range cgrEvs {
		evs[i] = &utils.EventWithFlags{
			Flags: procFlgs[i].AsSlice(),
			Event: cgrEv.CGREvent.Event,
		}
	}
	return
}

// Call implements the rpcclient.ClientConnector interface
func (cdrS *CDRServer) Call(serviceMethod string, args interface{}, reply interface{}) error {
	parts := strings.Split(serviceMethod, ".")
	if len(parts) != 2 {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	// get method
	method := reflect.ValueOf(cdrS).MethodByName(parts[0][len(parts[0])-2:] + parts[1]) // Inherit the version in the method
	if !method.IsValid() {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	// construct the params
	params := []reflect.Value{reflect.ValueOf(args), reflect.ValueOf(reply)}
	ret := method.Call(params)
	if len(ret) != 1 {
		return utils.ErrServerError
	}
	if ret[0].Interface() == nil {
		return nil
	}
	err, ok := ret[0].Interface().(error)
	if !ok {
		return utils.ErrServerError
	}
	return err
}

// V1ProcessCDR processes a CDR
func (cdrS *CDRServer) V1ProcessCDR(cdr *CDRWithOpts, reply *string) (err error) {
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
	cgrEv := &utils.CGREventWithOpts{
		CGREvent: cdr.AsCGREvent(),
		Opts:     cdr.Opts,
	}

	if _, err = cdrS.processEvent(cgrEv,
		len(cdrS.cgrCfg.CdrsCfg().ChargerSConns) != 0 && !cdr.PreRated,
		len(cdrS.cgrCfg.CdrsCfg().AttributeSConns) != 0,
		false,
		!cdr.PreRated, // rate the CDR if is not PreRated
		cdrS.cgrCfg.CdrsCfg().StoreCdrs,
		false, // no rerate
		len(cdrS.cgrCfg.CdrsCfg().OnlineCDRExports) != 0 || len(cdrS.cgrCfg.CdrsCfg().EEsConns) != 0,
		len(cdrS.cgrCfg.CdrsCfg().ThresholdSConns) != 0,
		len(cdrS.cgrCfg.CdrsCfg().StatSConns) != 0); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// ArgV1ProcessEvent is the CGREvent with proccesing Flags
type ArgV1ProcessEvent struct {
	Flags []string
	utils.CGREventWithOpts
	clnb bool //rpcclonable
}

// SetCloneable sets if the args should be clonned on internal connections
func (attr *ArgV1ProcessEvent) SetCloneable(rpcCloneable bool) {
	attr.clnb = rpcCloneable
}

// RPCClone implements rpcclient.RPCCloner interface
func (attr *ArgV1ProcessEvent) RPCClone() (interface{}, error) {
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
		for i, id := range attr.Flags {
			flags[i] = id
		}
	}
	return &ArgV1ProcessEvent{
		Flags:            flags,
		CGREventWithOpts: *attr.CGREventWithOpts.Clone(),
	}
}

// V1ProcessEvent will process the CGREvent
func (cdrS *CDRServer) V1ProcessEvent(arg *ArgV1ProcessEvent, reply *string) (err error) {
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

	// processing options
	flgs := utils.FlagsWithParamsFromSlice(arg.Flags)
	attrS := len(cdrS.cgrCfg.CdrsCfg().AttributeSConns) != 0
	if flgs.Has(utils.MetaAttributes) {
		attrS = flgs.GetBool(utils.MetaAttributes)
	}
	store := cdrS.cgrCfg.CdrsCfg().StoreCdrs
	if flgs.Has(utils.MetaStore) {
		store = flgs.GetBool(utils.MetaStore)
	}
	export := len(cdrS.cgrCfg.CdrsCfg().OnlineCDRExports) != 0 || len(cdrS.cgrCfg.CdrsCfg().EEsConns) != 0
	if flgs.Has(utils.MetaExport) {
		export = flgs.GetBool(utils.MetaExport)
	}
	thdS := len(cdrS.cgrCfg.CdrsCfg().ThresholdSConns) != 0
	if flgs.Has(utils.MetaThresholds) {
		thdS = flgs.GetBool(utils.MetaThresholds)
	}
	stS := len(cdrS.cgrCfg.CdrsCfg().StatSConns) != 0
	if flgs.Has(utils.MetaStats) {
		stS = flgs.GetBool(utils.MetaStats)
	}
	chrgS := len(cdrS.cgrCfg.CdrsCfg().ChargerSConns) != 0 // activate charging for the Event
	if flgs.Has(utils.MetaChargers) {
		chrgS = flgs.GetBool(utils.MetaChargers)
	}
	var ralS bool // activate single rating for the CDR
	if flgs.Has(utils.MetaRALs) {
		ralS = flgs.GetBool(utils.MetaRALs)
	}
	var reRate bool
	if flgs.Has(utils.MetaRerate) {
		if reRate = flgs.GetBool(utils.MetaRerate); reRate {
			ralS = true
		}
	}
	var refund bool
	if flgs.Has(utils.MetaRefund) {
		refund = flgs.GetBool(utils.MetaRefund)
	}
	// end of processing options

	if _, err = cdrS.processEvent(&arg.CGREventWithOpts, chrgS, attrS, refund,
		ralS, store, reRate, export, thdS, stS); err != nil {
		return
	}
	*reply = utils.OK
	return nil
}

// V2ProcessEvent has the same logic with V1ProcessEvent except it adds the proccessed events to the reply
func (cdrS *CDRServer) V2ProcessEvent(arg *ArgV1ProcessEvent, evs *[]*utils.EventWithFlags) (err error) {
	if arg.CGREvent.ID == "" {
		arg.CGREvent.ID = utils.GenUUID()
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

	// processing options
	flgs := utils.FlagsWithParamsFromSlice(arg.Flags)
	attrS := len(cdrS.cgrCfg.CdrsCfg().AttributeSConns) != 0
	if flgs.Has(utils.MetaAttributes) {
		attrS = flgs.GetBool(utils.MetaAttributes)
	}
	store := cdrS.cgrCfg.CdrsCfg().StoreCdrs
	if flgs.Has(utils.MetaStore) {
		store = flgs.GetBool(utils.MetaStore)
	}
	export := len(cdrS.cgrCfg.CdrsCfg().OnlineCDRExports) != 0 || len(cdrS.cgrCfg.CdrsCfg().EEsConns) != 0
	if flgs.Has(utils.MetaExport) {
		export = flgs.GetBool(utils.MetaExport)
	}
	thdS := len(cdrS.cgrCfg.CdrsCfg().ThresholdSConns) != 0
	if flgs.Has(utils.MetaThresholds) {
		thdS = flgs.GetBool(utils.MetaThresholds)
	}
	stS := len(cdrS.cgrCfg.CdrsCfg().StatSConns) != 0
	if flgs.Has(utils.MetaStats) {
		stS = flgs.GetBool(utils.MetaStats)
	}
	chrgS := len(cdrS.cgrCfg.CdrsCfg().ChargerSConns) != 0 // activate charging for the Event
	if flgs.Has(utils.MetaChargers) {
		chrgS = flgs.GetBool(utils.MetaChargers)
	}
	var ralS bool // activate single rating for the CDR
	if flgs.Has(utils.MetaRALs) {
		ralS = flgs.GetBool(utils.MetaRALs)
	}
	var reRate bool
	if flgs.Has(utils.MetaRerate) {
		reRate = flgs.GetBool(utils.MetaRerate)
		if reRate {
			ralS = true
		}
	}
	var refund bool
	if flgs.Has(utils.MetaRefund) {
		refund = flgs.GetBool(utils.MetaRefund)
	}
	// end of processing options

	var procEvs []*utils.EventWithFlags
	if procEvs, err = cdrS.processEvent(&arg.CGREventWithOpts, chrgS, attrS, refund,
		ralS, store, reRate, export, thdS, stS); err != nil {
		return
	}
	*evs = procEvs
	return nil
}

// V1StoreSessionCost handles storing of the cost into session_costs table
func (cdrS *CDRServer) V1StoreSessionCost(attr *AttrCDRSStoreSMCost, reply *string) (err error) {
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
func (cdrS *CDRServer) V2StoreSessionCost(args *ArgsV2CDRSStoreSMCost, reply *string) (err error) {
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
	cc.Round()
	roundIncrements := cc.GetRoundIncrements()
	if len(roundIncrements) != 0 {
		cd := cc.CreateCallDescriptor()
		cd.CgrID = args.Cost.CGRID
		cd.RunID = args.Cost.RunID
		cd.Increments = roundIncrements
		var response float64
		if err := cdrS.connMgr.Call(cdrS.cgrCfg.CdrsCfg().RaterConns, nil,
			utils.ResponderRefundRounding,
			&CallDescriptorWithOpts{CallDescriptor: cd},
			&response); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<CDRS> RefundRounding for cc: %+v, got error: %s",
					cc, err.Error()))
		}
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
	Tenant string
	Opts   map[string]interface{}
}

// V1RateCDRs is used for re-/rate CDRs which are already stored within StorDB
// FixMe: add RPC caching
func (cdrS *CDRServer) V1RateCDRs(arg *ArgRateCDRs, reply *string) (err error) {
	var cdrFltr *utils.CDRsFilter
	if cdrFltr, err = arg.RPCCDRsFilter.AsCDRsFilter(cdrS.cgrCfg.GeneralCfg().DefaultTimezone); err != nil {
		return utils.NewErrServerError(err)
	}
	var cdrs []*CDR
	if cdrs, _, err = cdrS.cdrDb.GetCDRs(cdrFltr, false); err != nil {
		return
	}
	flgs := utils.FlagsWithParamsFromSlice(arg.Flags)
	store := cdrS.cgrCfg.CdrsCfg().StoreCdrs
	if flgs.Has(utils.MetaStore) {
		store = flgs.GetBool(utils.MetaStore)
	}
	export := len(cdrS.cgrCfg.CdrsCfg().OnlineCDRExports) != 0 || len(cdrS.cgrCfg.CdrsCfg().EEsConns) != 0
	if flgs.Has(utils.MetaExport) {
		export = flgs.GetBool(utils.MetaExport)
	}
	thdS := len(cdrS.cgrCfg.CdrsCfg().ThresholdSConns) != 0
	if flgs.Has(utils.MetaThresholds) {
		thdS = flgs.GetBool(utils.MetaThresholds)
	}
	statS := len(cdrS.cgrCfg.CdrsCfg().StatSConns) != 0
	if flgs.Has(utils.MetaStatS) {
		statS = flgs.GetBool(utils.MetaStatS)
	}
	chrgS := len(cdrS.cgrCfg.CdrsCfg().ChargerSConns) != 0
	if flgs.Has(utils.MetaChargers) {
		chrgS = flgs.GetBool(utils.MetaChargers)
	}
	attrS := len(cdrS.cgrCfg.CdrsCfg().AttributeSConns) != 0
	if flgs.Has(utils.MetaAttributes) {
		attrS = flgs.GetBool(utils.MetaAttributes)
	}

	if chrgS && len(cdrS.cgrCfg.CdrsCfg().ChargerSConns) == 0 {
		return utils.NewErrNotConnected(utils.ChargerS)
	}
	for _, cdr := range cdrs {
		cdr.Cost = -1 // the cost will be recalculated
		cgrEv := &utils.CGREventWithOpts{
			CGREvent: cdr.AsCGREvent(),
			Opts:     arg.Opts,
		}
		if _, err = cdrS.processEvent(cgrEv, chrgS, attrS, false,
			true, store, true, export, thdS, statS); err != nil {
			return utils.NewErrServerError(err)
		}
	}
	*reply = utils.OK
	return
}

// V1ProcessExternalCDR is used to process external CDRs
func (cdrS *CDRServer) V1ProcessExternalCDR(eCDR *ExternalCDRWithOpts, reply *string) error {
	cdr, err := NewCDRFromExternalCDR(eCDR.ExternalCDR,
		cdrS.cgrCfg.GeneralCfg().DefaultTimezone)
	if err != nil {
		return err
	}
	return cdrS.V1ProcessCDR(&CDRWithOpts{
		CDR:  cdr,
		Opts: eCDR.Opts,
	}, reply)
}

// V1GetCDRs returns CDRs from DB
func (cdrS *CDRServer) V1GetCDRs(args utils.RPCCDRsFilterWithOpts, cdrs *[]*CDR) error {
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

// V1CountCDRs counts CDRs from DB
func (cdrS *CDRServer) V1CountCDRs(args *utils.RPCCDRsFilterWithOpts, cnt *int64) error {
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
