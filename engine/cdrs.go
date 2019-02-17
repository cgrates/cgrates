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
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var cdrServer *CDRServer // Share the server so we can use it in http handlers

// cgrCdrHandler handles CDRs received over HTTP REST
func cgrCdrHandler(w http.ResponseWriter, r *http.Request) {
	cgrCdr, err := NewCgrCdrFromHttpReq(r,
		cdrServer.cgrCfg.GeneralCfg().DefaultTimezone)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> could not create CDR entry from http: %+v, err <%s>",
				utils.CDRs, r.Form, err.Error()))
		return
	}
	cdr := cgrCdr.AsCDR(cdrServer.cgrCfg.GeneralCfg().DefaultTimezone)
	var ignored string
	if err := cdrServer.V1ProcessCDR(cdr, &ignored); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> processing CDR: %s, err: <%s>",
				utils.CDRs, cdr, err.Error()))
	}
}

// fsCdrHandler will handle CDRs received from FreeSWITCH over HTTP-JSON
func fsCdrHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	fsCdr, err := NewFSCdr(body, cdrServer.cgrCfg)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<CDRS> Could not create CDR entry: %s", err.Error()))
		return
	}
	cdr := fsCdr.AsCDR(cdrServer.cgrCfg.GeneralCfg().DefaultTimezone)
	var ignored string
	if err := cdrServer.V1ProcessCDR(cdr, &ignored); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> processing CDR: %s, err: <%s>",
				utils.CDRs, cdr, err.Error()))
	}
}

// NewCDRServer is a constructor for CDRServer
func NewCDRServer(cgrCfg *config.CGRConfig, cdrDb CdrStorage, dm *DataManager, rater, pubsub,
	attrS, users, thdS, statS, chargerS rpcclient.RpcClientConnection, filterS *FilterS) *CDRServer {
	if rater != nil && reflect.ValueOf(rater).IsNil() {
		rater = nil
	}
	if pubsub != nil && reflect.ValueOf(pubsub).IsNil() {
		pubsub = nil
	}
	if attrS != nil && reflect.ValueOf(attrS).IsNil() {
		attrS = nil
	}
	if users != nil && reflect.ValueOf(users).IsNil() {
		users = nil
	}
	if thdS != nil && reflect.ValueOf(thdS).IsNil() {
		thdS = nil
	}
	if statS != nil && reflect.ValueOf(statS).IsNil() {
		statS = nil
	}
	if chargerS != nil && reflect.ValueOf(chargerS).IsNil() {
		chargerS = nil
	}
	return &CDRServer{cgrCfg: cgrCfg, cdrDb: cdrDb, dm: dm,
		rals: rater, pubsub: pubsub, attrS: attrS,
		users: users,
		statS: statS, thdS: thdS,
		chargerS: chargerS, guard: guardian.Guardian,
		respCache: utils.NewResponseCache(cgrCfg.GeneralCfg().ResponseCacheTTL),
		httpPoster: NewHTTPPoster(cgrCfg.GeneralCfg().HttpSkipTlsVerify,
			cgrCfg.GeneralCfg().ReplyTimeout), filterS: filterS}
}

// CDRServer stores and rates CDRs
type CDRServer struct {
	cgrCfg     *config.CGRConfig
	cdrDb      CdrStorage
	dm         *DataManager
	rals       rpcclient.RpcClientConnection
	pubsub     rpcclient.RpcClientConnection
	attrS      rpcclient.RpcClientConnection
	users      rpcclient.RpcClientConnection
	thdS       rpcclient.RpcClientConnection
	statS      rpcclient.RpcClientConnection
	chargerS   rpcclient.RpcClientConnection
	guard      *guardian.GuardianLocker
	respCache  *utils.ResponseCache
	httpPoster *HTTPPoster // used for replication
	filterS    *FilterS
}

// RegisterHandlersToServer is called by cgr-engine to register HTTP URL handlers
func (cdrS *CDRServer) RegisterHandlersToServer(server *utils.Server) {
	cdrServer = cdrS // Share the server object for handlers
	server.RegisterHttpFunc(cdrS.cgrCfg.HTTPCfg().HTTPCDRsURL, cgrCdrHandler)
	server.RegisterHttpFunc(cdrS.cgrCfg.HTTPCfg().HTTPFreeswitchCDRsURL, fsCdrHandler)
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
		}, time.Duration(2*time.Second), lockKey) // FixMe: Possible deadlock with Guard from SMG session close()
		return err
	}
	return cdrS.cdrDb.SetSMCost(smCost)
}

// rateCDR will populate cost field
// Returns more than one rated CDR in case of SMCost retrieved based on prefix
func (cdrS *CDRServer) rateCDR(cdr *CDR) ([]*CDR, error) {
	var qryCC *CallCost
	var err error
	if cdr.RequestType == utils.META_NONE {
		return nil, nil
	}
	if cdr.Usage < 0 {
		cdr.Usage = time.Duration(0)
	}
	cdr.ExtraInfo = "" // Clean previous ExtraInfo, useful when re-rating
	var cdrsRated []*CDR
	_, hasLastUsed := cdr.ExtraFields[utils.LastUsed]
	if utils.IsSliceMember([]string{utils.META_PREPAID, utils.PREPAID}, cdr.RequestType) &&
		(cdr.Usage != 0 || hasLastUsed) { // ToDo: Get rid of PREPAID as soon as we don't want to support it backwards
		// Should be previously calculated and stored in DB
		fib := utils.Fib()
		var smCosts []*SMCost
		cgrID := cdr.CGRID
		if _, hasIT := cdr.ExtraFields[utils.OriginIDPrefix]; hasIT {
			cgrID = "" // for queries involving originIDPrefix we ignore CGRID
		}
		for i := 0; i < cdrS.cgrCfg.CdrsCfg().CDRSSMCostRetries; i++ {
			smCosts, err = cdrS.cdrDb.GetSMCosts(cgrID, cdr.RunID, cdr.OriginHost,
				cdr.ExtraFields[utils.OriginIDPrefix])
			if err == nil && len(smCosts) != 0 {
				break
			}
			if i != 3 {
				time.Sleep(time.Duration(fib()) * time.Second)
			}
		}
		if len(smCosts) != 0 { // Cost retrieved from SMCost table
			for _, smCost := range smCosts {
				cdrClone := cdr.Clone()
				cdrClone.OriginID = smCost.OriginID
				if cdr.Usage == 0 {
					cdrClone.Usage = smCost.Usage
				}
				cdrClone.Cost = smCost.CostDetails.GetCost()
				cdrClone.CostDetails = smCost.CostDetails
				cdrClone.CostSource = smCost.CostSource
				cdrsRated = append(cdrsRated, cdrClone)
			}
			return cdrsRated, nil
		} else { //calculate CDR as for pseudoprepaid
			utils.Logger.Warning(
				fmt.Sprintf("<Cdrs> WARNING: Could not find CallCostLog for cgrid: %s, source: %s, runid: %s, originID: %s originHost: %s, will recalculate",
					cdr.CGRID, utils.MetaSessionS, cdr.RunID, cdr.OriginID, cdr.OriginHost))
			qryCC, err = cdrS.getCostFromRater(cdr)
		}
	} else {
		qryCC, err = cdrS.getCostFromRater(cdr)
	}
	if err != nil {
		return nil, err
	} else if qryCC != nil {
		cdr.Cost = qryCC.Cost
		cdr.CostDetails = NewEventCostFromCallCost(qryCC, cdr.CGRID, cdr.RunID)
	}
	cdr.CostDetails.Compute()
	return []*CDR{cdr}, nil
}

// getCostFromRater will retrieve the cost from RALs
func (cdrS *CDRServer) getCostFromRater(cdr *CDR) (*CallCost, error) {
	cc := new(CallCost)
	var err error
	timeStart := cdr.AnswerTime
	if timeStart.IsZero() { // Fix for FreeSWITCH unanswered calls
		timeStart = cdr.SetupTime
	}
	cd := &CallDescriptor{
		TOR:             cdr.ToR,
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
	if utils.IsSliceMember([]string{utils.META_PSEUDOPREPAID, utils.META_POSTPAID, utils.META_PREPAID,
		utils.PSEUDOPREPAID, utils.POSTPAID, utils.PREPAID}, cdr.RequestType) { // Prepaid - Cost can be recalculated in case of missing records from SM
		err = cdrS.rals.Call("Responder.Debit", cd, cc)
	} else {
		err = cdrS.rals.Call("Responder.GetCost", cd, cc)
	}
	if err != nil {
		return cc, err
	}
	cdr.CostSource = utils.MetaCDRs
	return cc, nil
}

// chrgRaStoReThStaCDR will process the CGREvent with ChargerS subsystem
// it is designed to run in it's own goroutine
func (cdrS *CDRServer) chrgRaStoReThStaCDR(cgrEv *utils.CGREvent,
	store, export, thdS, statS *bool) (err error) {
	var chrgrs []*ChrgSProcessEventReply
	if err = cdrS.chargerS.Call(utils.ChargerSv1ProcessEvent,
		cgrEv, &chrgrs); err != nil &&
		err.Error() != utils.ErrNotFound.Error() {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s processing CGR event %+v with %s.",
				utils.CDRs, err.Error(), cgrEv, utils.ChargerS))
		return
	}
	var partExec bool
	for _, chrgr := range chrgrs {
		cdr, errCdr := NewMapEvent(chrgr.CGREvent.Event).AsCDR(cdrS.cgrCfg,
			cgrEv.Tenant, cdrS.cgrCfg.GeneralCfg().DefaultTimezone)
		if errCdr != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s converting CDR event %+v with %s.",
					utils.CDRs, errCdr.Error(), cgrEv, utils.ChargerS))
			partExec = true
			continue
		}
		cdrS.raStoReThStaCDR(cdr, store, export, thdS, statS)
	}
	if partExec {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// raStoReThStaCDR will RAte/STOtore/REplicate/THresholds/STAts the CDR received
// used by both chargerS as well as re-/rating
func (cdrS *CDRServer) raStoReThStaCDR(cdr *CDR,
	store, export, thdS, statS *bool) {
	ratedCDRs, err := cdrS.rateCDR(cdr)
	if err != nil {
		cdr.Cost = -1.0 // If there was an error, mark the CDR
		cdr.ExtraInfo = err.Error()
		ratedCDRs = []*CDR{cdr}
	}
	for _, rtCDR := range ratedCDRs {
		shouldStore := cdrS.cgrCfg.CdrsCfg().CDRSStoreCdrs
		if store != nil {
			shouldStore = *store
		}
		if shouldStore { // Store CDR
			go func(rtCDR *CDR) {
				if err := cdrS.cdrDb.SetCDR(rtCDR, true); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> error: %s storing CDR  %+v.",
							utils.CDRs, err.Error(), rtCDR))
				}
			}(rtCDR)
		}
		shouldExport := len(cdrS.cgrCfg.CdrsCfg().CDRSOnlineCDRExports) != 0
		if export != nil {
			shouldExport = *export
		}
		if shouldExport {
			go cdrS.exportCDRs([]*CDR{rtCDR})
		}
		cgrEv := rtCDR.AsCGREvent()
		shouldThdS := cdrS.thdS != nil
		if thdS != nil {
			shouldThdS = *thdS
		}
		if shouldThdS {
			go cdrS.thdSProcessEvent(cgrEv)
		}
		shouldStatS := cdrS.statS != nil
		if statS != nil {
			shouldStatS = *statS
		}
		if shouldStatS {
			go cdrS.statSProcessEvent(cgrEv)
		}
	}
}

// statSProcessEvent will send the event to StatS if the connection is configured
func (cdrS *CDRServer) attrSProcessEvent(cgrEv *utils.CGREvent) (err error) {
	if cgrEv.Context == nil { // populate if not already in
		cgrEv.Context = utils.StringPointer(utils.MetaCDRs)
	}
	var rplyEv AttrSProcessEventReply
	if err = cdrS.attrS.Call(utils.AttributeSv1ProcessEvent,
		&AttrArgsProcessEvent{
			CGREvent: *cgrEv},
		&rplyEv); err == nil && len(rplyEv.AlteredFields) != 0 {
		*cgrEv = *rplyEv.CGREvent
	} else if err.Error() == utils.ErrNotFound.Error() {
		err = nil // cancel ErrNotFound
	}
	return
}

// thdSProcessEvent will send the event to ThresholdS if the connection is configured
func (cdrS *CDRServer) thdSProcessEvent(cgrEv *utils.CGREvent) {
	var tIDs []string
	if err := cdrS.thdS.Call(utils.ThresholdSv1ProcessEvent,
		&ArgsProcessEvent{CGREvent: *cgrEv}, &tIDs); err != nil &&
		err.Error() != utils.ErrNotFound.Error() {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s processing CDR event %+v with thdS.",
				utils.CDRs, err.Error(), cgrEv))
		return
	}
}

// statSProcessEvent will send the event to StatS if the connection is configured
func (cdrS *CDRServer) statSProcessEvent(cgrEv *utils.CGREvent) {
	var reply []string
	if err := cdrS.statS.Call(utils.StatSv1ProcessEvent,
		&StatsArgsProcessEvent{CGREvent: *cgrEv}, &reply); err != nil &&
		err.Error() != utils.ErrNotFound.Error() {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s processing CDR event %+v with %s.",
				utils.CDRs, err.Error(), cgrEv, utils.StatS))
		return
	}
}

// exportCDRs will export the CDRs received
func (cdrS *CDRServer) exportCDRs(cdrs []*CDR) (err error) {
	for _, exportID := range cdrS.cgrCfg.CdrsCfg().CDRSOnlineCDRExports {
		expTpl := cdrS.cgrCfg.CdreProfiles[exportID] // not checking for existence of profile since this should be done in a higher layer
		var cdre *CDRExporter
		if cdre, err = NewCDRExporter(cdrs, expTpl, expTpl.ExportFormat,
			expTpl.ExportPath, cdrS.cgrCfg.GeneralCfg().FailedPostsDir,
			"CDRSReplication", expTpl.Synchronous, expTpl.Attempts,
			expTpl.FieldSeparator, expTpl.UsageMultiplyFactor,
			expTpl.CostMultiplyFactor, cdrS.cgrCfg.GeneralCfg().RoundingDecimals,
			cdrS.cgrCfg.GeneralCfg().HttpSkipTlsVerify, cdrS.httpPoster,
			cdrS.filterS); err != nil {
			utils.Logger.Err(fmt.Sprintf("<CDRS> Building CDRExporter for online exports got error: <%s>", err.Error()))
			continue
		}
		if err = cdre.ExportCDRs(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<CDRS> Replicating CDR: %+v, got error: <%s>", cdrs, err.Error()))
			continue
		}
	}
	return
}

// Call implements the rpcclient.RpcClientConnection interface
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
func (cdrS *CDRServer) V1ProcessCDR(cdr *CDR, reply *string) (err error) {
	if cdr.CGRID == utils.EmptyString { // Populate CGRID if not present
		cdr.ComputeCGRID()
	}
	cacheKey := "V1ProcessCDR" + cdr.CGRID + cdr.RunID
	if item, err := cdrS.respCache.Get(cacheKey); err == nil && item != nil {
		if item.Err == nil {
			*reply = *item.Value.(*string)
		}
		return item.Err
	}
	defer cdrS.respCache.Cache(cacheKey,
		&utils.ResponseCacheItem{Value: reply, Err: err})

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
	if !cdr.PreRated { // Enforce the RunID if CDR is not rated
		cdr.RunID = utils.MetaRaw
	}
	if utils.IsSliceMember([]string{"", utils.MetaRaw}, cdr.RunID) {
		cdr.Cost = -1.0
	}
	cgrEv := &utils.CGREvent{
		Tenant: cdr.Tenant,
		ID:     utils.UUIDSha1Prefix(),
		Event:  cdr.AsMapStringIface(),
	}
	if cdrS.attrS != nil {
		if err = cdrS.attrSProcessEvent(cgrEv); err != nil {
			return utils.NewErrServerError(err)
		}
	}
	if cdrS.cgrCfg.CdrsCfg().CDRSStoreCdrs { // Store *raw CDR
		if err = cdrS.cdrDb.SetCDR(cdr, false); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> storing primary CDR %+v, got error: %s",
					utils.CDRs, cdr, err.Error()))
			return utils.NewErrServerError(err) // Cannot store CDR
		}
	}
	if len(cdrS.cgrCfg.CdrsCfg().CDRSOnlineCDRExports) != 0 {
		cdrS.exportCDRs([]*CDR{cdr}) // Replicate raw CDR
	}
	if cdrS.thdS != nil {
		go cdrS.thdSProcessEvent(cgrEv)
	}
	if cdrS.statS != nil {
		go cdrS.statSProcessEvent(cgrEv)
	}
	if cdrS.chargerS != nil &&
		utils.IsSliceMember([]string{"", utils.MetaRaw}, cdr.RunID) {
		go cdrS.chrgRaStoReThStaCDR(cgrEv, nil, nil, nil, nil)
	}
	*reply = utils.OK

	return
}

type ArgV2ProcessCDR struct {
	utils.CGREvent
	AttributeS *bool // control AttributeS processing
	ChargerS   *bool // control ChargerS processing
	Store      *bool // control storing of the CDR
	Export     *bool // control online exports for the CDR
	ThresholdS *bool // control ThresholdS
	StatS      *bool // control sending the CDR to StatS for aggregation
}

// V2ProcessCDR will process the CDR out of CGREvent
func (cdrS *CDRServer) V2ProcessCDR(arg *ArgV2ProcessCDR, reply *string) (err error) {
	attrS := cdrS.attrS != nil
	if arg.AttributeS != nil {
		attrS = *arg.AttributeS
	}
	cgrEv := &arg.CGREvent
	if attrS {
		if err := cdrS.attrSProcessEvent(cgrEv); err != nil {
			return utils.NewErrServerError(err)
		}
	}
	rawCDR, err := NewMapEvent(cgrEv.Event).AsCDR(cdrS.cgrCfg,
		cgrEv.Tenant, cdrS.cgrCfg.GeneralCfg().DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	store := cdrS.cgrCfg.CdrsCfg().CDRSStoreCdrs
	if arg.Store != nil {
		store = *arg.Store
	}
	if store { // Store *raw CDR
		if err = cdrS.cdrDb.SetCDR(rawCDR, false); err != nil {
			return utils.NewErrServerError(err) // Cannot store CDR
		}
	}
	export := len(cdrS.cgrCfg.CdrsCfg().CDRSOnlineCDRExports) != 0
	if arg.Export != nil {
		export = *arg.Export
	}
	if export {
		cdrS.exportCDRs([]*CDR{rawCDR}) // Replicate raw CDR
	}
	thrdS := cdrS.thdS != nil
	if arg.ThresholdS != nil {
		thrdS = *arg.ThresholdS
	}
	if thrdS {
		go cdrS.thdSProcessEvent(cgrEv)
	}
	statS := cdrS.statS != nil
	if arg.StatS != nil {
		statS = *arg.StatS
	}
	if statS {
		go cdrS.statSProcessEvent(cgrEv)
	}
	chrgS := cdrS.chargerS != nil
	if arg.ChargerS != nil {
		chrgS = *arg.ChargerS
	}
	if chrgS {
		go cdrS.chrgRaStoReThStaCDR(cgrEv,
			arg.Store, arg.Export, arg.ThresholdS, arg.StatS)
	}
	*reply = utils.OK
	return nil
}

// V1StoreSMCost handles storing of the cost into session_costs table
func (cdrS *CDRServer) V1StoreSessionCost(attr *AttrCDRSStoreSMCost, reply *string) error {
	if attr.Cost.CGRID == "" {
		return utils.NewCGRError(utils.CDRSCtx,
			utils.MandatoryIEMissingCaps, fmt.Sprintf("%s: CGRID", utils.MandatoryInfoMissing),
			"SMCost: %+v with empty CGRID")
	}
	if err := cdrS.storeSMCost(attr.Cost, attr.CheckDuplicate); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// V2StoreSessionCost will store the SessionCost into session_costs table
func (cdrS *CDRServer) V2StoreSessionCost(args *ArgsV2CDRSStoreSMCost, reply *string) error {
	if args.Cost.CGRID == "" {
		return utils.NewCGRError(utils.CDRSCtx,
			utils.MandatoryIEMissingCaps, fmt.Sprintf("%s: CGRID", utils.MandatoryInfoMissing),
			"SMCost: %+v with empty CGRID")
	}
	cc := args.Cost.CostDetails.AsCallCost()
	cc.Round()
	roundIncrements := cc.GetRoundIncrements()
	if len(roundIncrements) != 0 {
		cd := cc.CreateCallDescriptor()
		cd.CgrID = args.Cost.CGRID
		cd.RunID = args.Cost.RunID
		cd.Increments = roundIncrements
		var response float64
		if err := cdrS.rals.Call("Responder.RefundRounding",
			cd, &response); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<CDRS> RefundRounding for cc: %+v, got error: %s",
					cc, err.Error()))
		}
	}
	if err := cdrS.storeSMCost(
		&SMCost{
			CGRID:       args.Cost.CGRID,
			RunID:       args.Cost.RunID,
			OriginHost:  args.Cost.OriginHost,
			OriginID:    args.Cost.OriginID,
			CostSource:  args.Cost.CostSource,
			Usage:       args.Cost.Usage,
			CostDetails: args.Cost.CostDetails},
		args.CheckDuplicate); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil

}

type ArgRateCDRs struct {
	utils.RPCCDRsFilter
	ChargerS   *bool
	Store      *bool
	Export     *bool // Replicate results
	ThresholdS *bool
	StatS      *bool // Set to true if the CDRs should be sent to stats server
}

// V1RateCDRs is used for re-/rate CDRs which are already stored within StorDB
func (cdrS *CDRServer) V1RateCDRs(arg *ArgRateCDRs, reply *string) (err error) {
	cdrFltr, err := arg.RPCCDRsFilter.AsCDRsFilter(cdrS.cgrCfg.GeneralCfg().DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	cdrs, _, err := cdrS.cdrDb.GetCDRs(cdrFltr, false)
	if err != nil {
		return err
	}
	for _, cdr := range cdrs {
		if arg.ChargerS != nil && *arg.ChargerS {
			if cdrS.chargerS == nil {
				return utils.NewErrNotConnected(utils.ChargerS)
			}
			if err = cdrS.chrgRaStoReThStaCDR(cdr.AsCGREvent(),
				arg.Store, arg.Export, arg.ThresholdS, arg.StatS); err != nil {
				return utils.NewErrServerError(err)
			}
		} else {
			cdrS.raStoReThStaCDR(cdr, arg.Store,
				arg.Export, arg.ThresholdS, arg.StatS)
		}
	}
	*reply = utils.OK
	return nil
}

// Used to process external CDRs
func (cdrS *CDRServer) V1ProcessExternalCDR(eCDR *ExternalCDR, reply *string) error {
	cdr, err := NewCDRFromExternalCDR(eCDR, cdrS.cgrCfg.GeneralCfg().DefaultTimezone)
	if err != nil {
		return err
	}
	return cdrS.V1ProcessCDR(cdr, reply)
}

// V1GetCDRs returns CDRs from DB
func (cdrS *CDRServer) V1GetCDRs(args utils.RPCCDRsFilter, cdrs *[]*CDR) error {
	cdrsFltr, err := args.AsCDRsFilter(cdrS.cgrCfg.GeneralCfg().DefaultTimezone)
	if err != nil {
		if err.Error() != utils.NotFoundCaps {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	if qryCDRs, _, err := cdrS.cdrDb.GetCDRs(cdrsFltr, false); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*cdrs = qryCDRs
	}
	return nil
}

// V1CountCDRs counts CDRs from DB
func (cdrS *CDRServer) V1CountCDRs(args *utils.RPCCDRsFilter, cnt *int64) error {
	cdrsFltr, err := args.AsCDRsFilter(cdrS.cgrCfg.GeneralCfg().DefaultTimezone)
	if err != nil {
		if err.Error() != utils.NotFoundCaps {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	cdrsFltr.Count = true
	if _, qryCnt, err := cdrS.cdrDb.GetCDRs(cdrsFltr, false); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*cnt = qryCnt
	}
	return nil
}
