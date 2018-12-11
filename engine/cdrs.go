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

var cdrServer *CdrServer // Share the server so we can use it in http handlers

type CallCostLog struct {
	CgrId          string
	Source         string
	RunId          string
	Usage          float64 // real usage (not increment rounded)
	CallCost       *CallCost
	CheckDuplicate bool
}

// Handler for generic cgr cdr http
func cgrCdrHandler(w http.ResponseWriter, r *http.Request) {
	cgrCdr, err := NewCgrCdrFromHttpReq(r, cdrServer.cgrCfg.GeneralCfg().DefaultTimezone)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<CDRS> Could not create CDR entry: %s", err.Error()))
		return
	}
	if err := cdrServer.processCdr(cgrCdr.AsCDR(cdrServer.cgrCfg.GeneralCfg().DefaultTimezone)); err != nil {
		utils.Logger.Err(fmt.Sprintf("<CDRS> Errors when storing CDR entry: %s", err.Error()))
	}
}

// Handler for fs http
func fsCdrHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	fsCdr, err := NewFSCdr(body, cdrServer.cgrCfg)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<CDRS> Could not create CDR entry: %s", err.Error()))
		return
	}
	if err := cdrServer.processCdr(fsCdr.AsCDR(cdrServer.Timezone())); err != nil {
		utils.Logger.Err(fmt.Sprintf("<CDRS> Errors when storing CDR entry: %s", err.Error()))
	}
}

func NewCdrServer(cgrCfg *config.CGRConfig, cdrDb CdrStorage, dm *DataManager, rater, pubsub,
	attrs, users, aliases, thdS, stats, chargerS rpcclient.RpcClientConnection, filterS *FilterS) (*CdrServer, error) {
	if rater != nil && reflect.ValueOf(rater).IsNil() { // Work around so we store actual nil instead of nil interface value, faster to check here than in CdrServer code
		rater = nil
	}
	if pubsub != nil && reflect.ValueOf(pubsub).IsNil() {
		pubsub = nil
	}
	if attrs != nil && reflect.ValueOf(attrs).IsNil() {
		attrs = nil
	}
	if users != nil && reflect.ValueOf(users).IsNil() {
		users = nil
	}
	if aliases != nil && reflect.ValueOf(aliases).IsNil() {
		aliases = nil
	}

	if thdS != nil && reflect.ValueOf(thdS).IsNil() {
		thdS = nil
	}
	if stats != nil && reflect.ValueOf(stats).IsNil() {
		stats = nil
	}
	if chargerS != nil && reflect.ValueOf(chargerS).IsNil() {
		chargerS = nil
	}
	return &CdrServer{cgrCfg: cgrCfg, cdrDb: cdrDb, dm: dm,
		rals: rater, pubsub: pubsub, attrS: attrs,
		users: users, aliases: aliases,
		stats: stats, thdS: thdS,
		chargerS: chargerS, guard: guardian.Guardian,
		httpPoster: NewHTTPPoster(cgrCfg.GeneralCfg().HttpSkipTlsVerify,
			cgrCfg.GeneralCfg().ReplyTimeout), filterS: filterS}, nil
}

type CdrServer struct {
	cgrCfg        *config.CGRConfig
	cdrDb         CdrStorage
	dm            *DataManager
	rals          rpcclient.RpcClientConnection
	pubsub        rpcclient.RpcClientConnection
	attrS         rpcclient.RpcClientConnection
	users         rpcclient.RpcClientConnection
	aliases       rpcclient.RpcClientConnection
	thdS          rpcclient.RpcClientConnection
	stats         rpcclient.RpcClientConnection
	chargerS      rpcclient.RpcClientConnection
	guard         *guardian.GuardianLocker
	responseCache *utils.ResponseCache
	httpPoster    *HTTPPoster // used for replication
	filterS       *FilterS
}

func (self *CdrServer) Timezone() string {
	return self.cgrCfg.GeneralCfg().DefaultTimezone
}
func (self *CdrServer) SetTimeToLive(timeToLive time.Duration, out *int) error {
	self.responseCache = utils.NewResponseCache(timeToLive)
	return nil
}

func (self *CdrServer) getCache() *utils.ResponseCache {
	if self.responseCache == nil {
		self.responseCache = utils.NewResponseCache(0)
	}
	return self.responseCache
}

func (self *CdrServer) RegisterHandlersToServer(server *utils.Server) {
	cdrServer = self // Share the server object for handlers
	server.RegisterHttpFunc(self.cgrCfg.HTTPCfg().HTTPCDRsURL, cgrCdrHandler)
	server.RegisterHttpFunc(self.cgrCfg.HTTPCfg().HTTPFreeswitchCDRsURL, fsCdrHandler)
}

// Used to process external CDRs
func (self *CdrServer) ProcessExternalCdr(eCDR *ExternalCDR) error {
	cdr, err := NewCDRFromExternalCDR(eCDR, self.cgrCfg.GeneralCfg().DefaultTimezone)
	if err != nil {
		return err
	}
	return self.processCdr(cdr)
}

func (self *CdrServer) storeSMCost(smCost *SMCost, checkDuplicate bool) error {
	smCost.CostDetails.Compute()                                              // make sure the total cost reflect the increment
	lockKey := utils.MetaCDRs + smCost.CGRID + smCost.RunID + smCost.OriginID // Will lock on this ID
	if checkDuplicate {
		_, err := self.guard.Guard(func() (interface{}, error) {
			smCosts, err := self.cdrDb.GetSMCosts(smCost.CGRID, smCost.RunID, "", "")
			if err != nil && err.Error() != utils.NotFoundCaps {
				return nil, err
			}
			if len(smCosts) != 0 {
				return nil, utils.ErrExists
			}
			return nil, self.cdrDb.SetSMCost(smCost)
		}, time.Duration(2*time.Second), lockKey) // FixMe: Possible deadlock with Guard from SMG session close()
		return err
	}
	return self.cdrDb.SetSMCost(smCost)
}

// Returns error if not able to properly store the CDR, mediation is async since we can always recover offline
func (self *CdrServer) processCdr(cdr *CDR) (err error) {
	if cdr.RequestType == "" {
		cdr.RequestType = self.cgrCfg.GeneralCfg().DefaultReqType
	}
	if cdr.Tenant == "" {
		cdr.Tenant = self.cgrCfg.GeneralCfg().DefaultTenant
	}
	if cdr.Category == "" {
		cdr.Category = self.cgrCfg.GeneralCfg().DefaultCategory
	}
	if cdr.Subject == "" { // Use account information as rating subject if missing
		cdr.Subject = cdr.Account
	}
	if !cdr.PreRated { // Enforce the RunID if CDR is not rated
		cdr.RunID = utils.MetaRaw
	}
	if cdr.RunID == utils.MetaRaw {
		cdr.Cost = -1.0
	}
	if self.cgrCfg.CdrsCfg().CDRSStoreCdrs { // Store RawCDRs, this we do sync so we can reply with the status
		if err := self.cdrDb.SetCDR(cdr, false); err != nil {
			utils.Logger.Err(fmt.Sprintf("<CDRS> Storing primary CDR %+v, got error: %s", cdr, err.Error()))
			return err // Error is propagated back and we don't continue processing the CDR if we cannot store it
		}
	}
	if self.thdS != nil {
		// process CDR with thresholdS
		go self.thdSProcessEvent(cdr.AsCGREvent())
	}
	if self.stats != nil {
		var reply []string

		go self.stats.Call(utils.StatSv1ProcessEvent, &StatsArgsProcessEvent{CGREvent: *cdr.AsCGREvent()}, &reply)
	}
	if len(self.cgrCfg.CdrsCfg().CDRSOnlineCDRExports) != 0 { // Replicate raw CDR
		self.replicateCDRs([]*CDR{cdr})
	}
	if self.rals != nil && !cdr.PreRated { // CDRs not rated will be processed by Rating
		go self.deriveRateStoreStatsReplicate(cdr, self.cgrCfg.CdrsCfg().CDRSStoreCdrs,
			true, len(self.cgrCfg.CdrsCfg().CDRSOnlineCDRExports) != 0)
	}
	return nil
}

// Returns error if not able to properly store the CDR, mediation is async since we can always recover offline
func (self *CdrServer) deriveRateStoreStatsReplicate(cdr *CDR, store, cdrstats, replicate bool) (err error) {
	cdrRuns, err := self.deriveCdrs(cdr)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<CDRS> Deriving CDR %+v, got error: %s", cdr, err.Error()))
		return err
	}
	var ratedCDRs []*CDR // Gather all CDRs received from rating subsystem
	for _, cdrRun := range cdrRuns {
		if self.attrS != nil {
			var rplyEv AttrSProcessEventReply
			cgrEv := cdrRun.AsCGREvent()
			cgrEv.Context = utils.StringPointer(utils.MetaCDRs)
			if err = self.attrS.Call(utils.AttributeSv1ProcessEvent,
				cgrEv, &rplyEv); err == nil {
				if err = cdrRun.UpdateFromCGREvent(rplyEv.CGREvent,
					rplyEv.AlteredFields); err != nil {
					return
				}
			} else if err.Error() != utils.ErrNotFound.Error() {
				return
			}
		}
		if err := LoadUserProfile(cdrRun, utils.EXTRA_FIELDS); err != nil {
			utils.Logger.Err(fmt.Sprintf("<CDRS> UserS handling for CDR %+v, got error: %s", cdrRun, err.Error()))
			continue
		}
		if err := LoadAlias(&AttrMatchingAlias{
			Destination: cdrRun.Destination,
			Direction:   utils.OUT,
			Tenant:      cdrRun.Tenant,
			Category:    cdrRun.Category,
			Account:     cdrRun.Account,
			Subject:     cdrRun.Subject,
			Context:     utils.MetaRating,
		}, cdrRun, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
			utils.Logger.Err(fmt.Sprintf("<CDRS> Aliasing CDR %+v, got error: %s", cdrRun, err.Error()))
			continue
		}
		rcvRatedCDRs, err := self.rateCDR(cdrRun)
		if err != nil {
			cdrRun.Cost = -1.0 // If there was an error, mark the CDR
			cdrRun.ExtraInfo = err.Error()
			rcvRatedCDRs = []*CDR{cdrRun}
		}
		ratedCDRs = append(ratedCDRs, rcvRatedCDRs...)
	}
	// Request should be processed by SureTax
	for _, ratedCDR := range ratedCDRs {
		if ratedCDR.RunID == utils.META_SURETAX {
			if err := SureTaxProcessCdr(ratedCDR); err != nil {
				ratedCDR.Cost = -1.0
				ratedCDR.ExtraInfo = err.Error() // Something failed, write the error in the ExtraInfo
			}
		}
	}
	// Store rated CDRs
	if store {
		for _, ratedCDR := range ratedCDRs {
			if err := self.cdrDb.SetCDR(ratedCDR, true); err != nil {
				utils.Logger.Err(fmt.Sprintf("<CDRS> Storing rated CDR %+v, got error: %s", ratedCDR, err.Error()))
			}
		}
	}
	// Attach CDR to stats
	if cdrstats { // Send CDR to stats
		for _, ratedCDR := range ratedCDRs {
			if self.stats != nil {
				var reply []string
				go self.stats.Call(utils.StatSv1ProcessEvent, &StatsArgsProcessEvent{CGREvent: *ratedCDR.AsCGREvent()}, &reply)
			}
		}
	}
	if replicate {
		self.replicateCDRs(ratedCDRs)
	}
	return nil
}

func (self *CdrServer) deriveCdrs(cdr *CDR) (drvdCDRs []*CDR, err error) {
	dfltCDRRun := cdr.Clone()
	cdrRuns := []*CDR{dfltCDRRun}
	if cdr.RunID != utils.MetaRaw { // Only derive *raw CDRs
		return cdrRuns, nil
	}
	dfltCDRRun.RunID = utils.META_DEFAULT // Rewrite *raw with *default since we have it as first run
	if self.attrS != nil {
		var rplyEv AttrSProcessEventReply
		if err = self.attrS.Call(utils.AttributeSv1ProcessEvent,
			cdr.AsCGREvent(), &rplyEv); err != nil {
			return
		}
		if err = cdr.UpdateFromCGREvent(rplyEv.CGREvent,
			rplyEv.AlteredFields); err != nil {
			return
		}
	}
	if err := LoadUserProfile(cdr, utils.EXTRA_FIELDS); err != nil {
		return nil, err
	}
	if err := LoadAlias(&AttrMatchingAlias{
		Destination: cdr.Destination,
		Direction:   utils.OUT,
		Tenant:      cdr.Tenant,
		Category:    cdr.Category,
		Account:     cdr.Account,
		Subject:     cdr.Subject,
		Context:     utils.MetaRating,
	}, cdr, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		return nil, err
	}
	attrsDC := &utils.AttrDerivedChargers{Tenant: cdr.Tenant, Category: cdr.Category, Direction: utils.OUT,
		Account: cdr.Account, Subject: cdr.Subject, Destination: cdr.Destination}
	var dcs utils.DerivedChargers
	if err := self.rals.Call("Responder.GetDerivedChargers", attrsDC, &dcs); err != nil {
		utils.Logger.Err(fmt.Sprintf("Could not get derived charging for cgrid %s, error: %s", cdr.CGRID, err.Error()))
		return nil, err
	}
	for _, dc := range dcs.Chargers {
		runFilters, _ := utils.ParseRSRFields(dc.RunFilters, utils.INFIELD_SEP)
		matchingAllFilters := true
		for _, dcRunFilter := range runFilters {
			if _, err := cdr.FieldAsStringWithRSRField(dcRunFilter); err != nil {
				matchingAllFilters = false
				break
			}
		}
		if !matchingAllFilters { // Do not process the derived charger further if not all filters were matched
			continue
		}
		dcRequestTypeFld, _ := utils.NewRSRField(dc.RequestTypeField)
		dcTenantFld, _ := utils.NewRSRField(dc.TenantField)
		dcCategoryFld, _ := utils.NewRSRField(dc.CategoryField)
		dcAcntFld, _ := utils.NewRSRField(dc.AccountField)
		dcSubjFld, _ := utils.NewRSRField(dc.SubjectField)
		dcDstFld, _ := utils.NewRSRField(dc.DestinationField)
		dcSTimeFld, _ := utils.NewRSRField(dc.SetupTimeField)
		dcATimeFld, _ := utils.NewRSRField(dc.AnswerTimeField)
		dcDurFld, _ := utils.NewRSRField(dc.UsageField)
		dcRatedFld, _ := utils.NewRSRField(dc.PreRatedField)
		dcCostFld, _ := utils.NewRSRField(dc.CostField)

		dcExtraFields := []*utils.RSRField{}
		for key := range cdr.ExtraFields {
			dcExtraFields = append(dcExtraFields, &utils.RSRField{Id: key})
		}

		forkedCdr, err := cdr.ForkCdr(dc.RunID, dcRequestTypeFld, dcTenantFld,
			dcCategoryFld, dcAcntFld, dcSubjFld, dcDstFld, dcSTimeFld, dcATimeFld,
			dcDurFld, dcRatedFld, dcCostFld, dcExtraFields, true,
			self.cgrCfg.GeneralCfg().DefaultTimezone)
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("Could not fork CGR with cgrid %s, run: %s, error: %s", cdr.CGRID, dc.RunID, err.Error()))
			continue // do not add it to the forked CDR list
		}
		if !forkedCdr.PreRated {
			forkedCdr.Cost = -1.0 // Make sure that un-rated CDRs start with Cost -1
		}
		cdrRuns = append(cdrRuns, forkedCdr)
	}
	return cdrRuns, nil
}

// rateCDR will populate cost field
// Returns more than one rated CDR in case of SMCost retrieved based on prefix
func (self *CdrServer) rateCDR(cdr *CDR) ([]*CDR, error) {
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
		for i := 0; i < self.cgrCfg.CdrsCfg().CDRSSMCostRetries; i++ {
			smCosts, err = self.cdrDb.GetSMCosts(cgrID, cdr.RunID, cdr.OriginHost,
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
			qryCC, err = self.getCostFromRater(cdr)
		}
	} else {
		qryCC, err = self.getCostFromRater(cdr)
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

// Retrive the cost from engine
func (self *CdrServer) getCostFromRater(cdr *CDR) (*CallCost, error) {
	cc := new(CallCost)
	var err error
	timeStart := cdr.AnswerTime
	if timeStart.IsZero() { // Fix for FreeSWITCH unanswered calls
		timeStart = cdr.SetupTime
	}
	cd := &CallDescriptor{
		TOR:             cdr.ToR,
		Direction:       utils.OUT,
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
		err = self.rals.Call("Responder.Debit", cd, cc)
	} else {
		err = self.rals.Call("Responder.GetCost", cd, cc)
	}
	if err != nil {
		return cc, err
	}
	cdr.CostSource = utils.MetaCDRs
	return cc, nil
}

func (self *CdrServer) replicateCDRs(cdrs []*CDR) (err error) {
	for _, exportID := range self.cgrCfg.CdrsCfg().CDRSOnlineCDRExports {
		expTpl := self.cgrCfg.CdreProfiles[exportID] // not checking for existence of profile since this should be done in a higher layer
		var cdre *CDRExporter
		if cdre, err = NewCDRExporter(cdrs, expTpl, expTpl.ExportFormat,
			expTpl.ExportPath, self.cgrCfg.GeneralCfg().FailedPostsDir,
			"CDRSReplication", expTpl.Synchronous, expTpl.Attempts,
			expTpl.FieldSeparator, expTpl.UsageMultiplyFactor,
			expTpl.CostMultiplyFactor, self.cgrCfg.GeneralCfg().RoundingDecimals,
			self.cgrCfg.GeneralCfg().HttpSkipTlsVerify, self.httpPoster,
			self.filterS); err != nil {
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

// Called by rate/re-rate API, FixMe: deprecate it once new APIer structure is operational
func (self *CdrServer) RateCDRs(cdrFltr *utils.CDRsFilter, sendToStats bool) error {
	cdrs, _, err := self.cdrDb.GetCDRs(cdrFltr, false)
	if err != nil {
		return err
	}
	for _, cdr := range cdrs {
		if err := self.deriveRateStoreStatsReplicate(cdr, self.cgrCfg.CdrsCfg().CDRSStoreCdrs,
			sendToStats, len(self.cgrCfg.CdrsCfg().CDRSOnlineCDRExports) != 0); err != nil {
			utils.Logger.Err(fmt.Sprintf("<CDRS> Processing CDR %+v, got error: %s", cdr, err.Error()))
		}
	}
	return nil
}

// Internally used and called from CDRSv1
// Cached requests for HA setups
func (self *CdrServer) V1ProcessCDR(cdr *CDR, reply *string) error {
	if len(cdr.CGRID) == 0 { // Populate CGRID if not present
		cdr.ComputeCGRID()
	}
	cacheKey := "V1ProcessCDR" + cdr.CGRID + cdr.RunID
	if item, err := self.getCache().Get(cacheKey); err == nil && item != nil {
		if item.Value != nil {
			*reply = item.Value.(string)
		}
		return item.Err
	}
	if err := self.processCdr(cdr); err != nil {
		self.getCache().Cache(cacheKey, &utils.ResponseCacheItem{Err: err})
		return utils.NewErrServerError(err)
	}
	self.getCache().Cache(cacheKey, &utils.ResponseCacheItem{Value: utils.OK})
	*reply = utils.OK
	return nil
}

// RPC method, differs from storeSMCost through it's signature
func (self *CdrServer) V1StoreSMCost(attr AttrCDRSStoreSMCost, reply *string) error {
	if attr.Cost.CGRID == "" {
		return utils.NewCGRError(utils.CDRSCtx,
			utils.MandatoryIEMissingCaps, fmt.Sprintf("%s: CGRID", utils.MandatoryInfoMissing),
			"SMCost: %+v with empty CGRID")
	}
	cacheKey := "V1StoreSMCost" + attr.Cost.CGRID + attr.Cost.RunID + attr.Cost.OriginID
	if item, err := self.getCache().Get(cacheKey); err == nil && item != nil {
		if item.Value != nil {
			*reply = item.Value.(string)
		}
		return item.Err
	}
	if err := self.storeSMCost(attr.Cost, attr.CheckDuplicate); err != nil {
		self.getCache().Cache(cacheKey, &utils.ResponseCacheItem{Err: err})
		return utils.NewErrServerError(err)
	}
	self.getCache().Cache(cacheKey, &utils.ResponseCacheItem{Value: utils.OK})
	*reply = utils.OK
	return nil
}

func (cdrs *CdrServer) V2StoreSMCost(args ArgsV2CDRSStoreSMCost, reply *string) error {
	if args.Cost.CGRID == "" {
		return utils.NewCGRError(utils.CDRSCtx,
			utils.MandatoryIEMissingCaps, fmt.Sprintf("%s: CGRID", utils.MandatoryInfoMissing),
			"SMCost: %+v with empty CGRID")
	}
	cacheKey := "V2StoreSMCost" + args.Cost.CGRID + args.Cost.RunID + args.Cost.OriginID
	if item, err := cdrs.getCache().Get(cacheKey); err == nil && item != nil {
		if item.Value != nil {
			*reply = item.Value.(string)
		}
		return item.Err
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
		if err := cdrs.rals.Call("Responder.RefundRounding", cd, &response); err != nil {
			utils.Logger.Err(fmt.Sprintf("<CDRS> RefundRounding for cc: %+v, got error: %s", cc, err.Error()))
		}
	}
	if err := cdrs.storeSMCost(&SMCost{
		CGRID:       args.Cost.CGRID,
		RunID:       args.Cost.RunID,
		OriginHost:  args.Cost.OriginHost,
		OriginID:    args.Cost.OriginID,
		CostSource:  args.Cost.CostSource,
		Usage:       args.Cost.Usage,
		CostDetails: args.Cost.CostDetails,
	}, args.CheckDuplicate); err != nil {
		cdrs.getCache().Cache(cacheKey, &utils.ResponseCacheItem{Err: err})
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	cdrs.getCache().Cache(cacheKey, &utils.ResponseCacheItem{Value: *reply})
	return nil

}

// Called by rate/re-rate API, RPC method
func (self *CdrServer) V1RateCDRs(attrs utils.AttrRateCDRs, reply *string) error {
	cdrFltr, err := attrs.RPCCDRsFilter.AsCDRsFilter(self.cgrCfg.GeneralCfg().DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	cdrs, _, err := self.cdrDb.GetCDRs(cdrFltr, false)
	if err != nil {
		return err
	}
	storeCDRs := self.cgrCfg.CdrsCfg().CDRSStoreCdrs
	if attrs.StoreCDRs != nil {
		storeCDRs = *attrs.StoreCDRs
	}
	sendToStats := self.stats != nil
	if attrs.SendToStatS != nil {
		sendToStats = *attrs.SendToStatS
	}
	replicate := len(self.cgrCfg.CdrsCfg().CDRSOnlineCDRExports) != 0
	if attrs.ReplicateCDRs != nil {
		replicate = *attrs.ReplicateCDRs
	}
	for _, cdr := range cdrs {
		if err := self.deriveRateStoreStatsReplicate(cdr, storeCDRs, sendToStats, replicate); err != nil {
			utils.Logger.Err(fmt.Sprintf("<CDRS> Processing CDR %+v, got error: %s", cdr, err.Error()))
		}
	}
	return nil
}

func (cdrsrv *CdrServer) Call(serviceMethod string, args interface{}, reply interface{}) error {
	parts := strings.Split(serviceMethod, ".")
	if len(parts) != 2 {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	// get method
	method := reflect.ValueOf(cdrsrv).MethodByName(parts[0][len(parts[0])-2:] + parts[1]) // Inherit the version in the method
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

// thdSProcessEvent will send the event to ThresholdS if the connection is configured
func (cdrS *CdrServer) thdSProcessEvent(cgrEv *utils.CGREvent) {
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
func (cdrS *CdrServer) statSProcessEvent(cgrEv *utils.CGREvent) {
	var reply []string
	if err := cdrS.stats.Call(utils.StatSv1ProcessEvent, &StatsArgsProcessEvent{CGREvent: *cgrEv}, &reply); err != nil &&
		err.Error() != utils.ErrNotFound.Error() {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s processing CDR event %+v with %s.",
				utils.CDRs, err.Error(), cgrEv, utils.StatS))
		return
	}
}

// rarethsta will RAte/STOtore/REplicate/THresholds/STAts the CDR received
// used by both chargerS as well as re-/rating
func (cdrS *CdrServer) raStoReThStaCDR(cdr *CDR) {
	ratedCDRs, err := cdrS.rateCDR(cdr)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s rating CDR  %+v.",
				utils.CDRs, err.Error(), cdr))
		return
	}
	for _, rtCDR := range ratedCDRs {
		if cdrS.cgrCfg.CdrsCfg().CDRSStoreCdrs { // Store CDR
			go func(rtCDR *CDR) {
				if err := cdrS.cdrDb.SetCDR(rtCDR, true); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> error: %s storing CDR  %+v.",
							utils.CDRs, err.Error(), rtCDR))
				}
			}(rtCDR)
		}
		if len(cdrS.cgrCfg.CdrsCfg().CDRSOnlineCDRExports) != 0 {
			go cdrS.replicateCDRs([]*CDR{rtCDR})
		}
		cgrEv := rtCDR.AsCGREvent()
		if cdrS.thdS != nil {
			go cdrS.thdSProcessEvent(cgrEv)
		}
		if cdrS.stats != nil {
			go cdrS.statSProcessEvent(cgrEv)
		}
	}
}

// chrgrSProcessEvent will process the CGREvent with ChargerS subsystem
func (cdrS *CdrServer) chrgrSProcessEvent(cgrEv *utils.CGREvent) {
	var chrgrs []*ChrgSProcessEventReply
	if err := cdrS.chargerS.Call(utils.ChargerSv1ProcessEvent, cgrEv, &chrgrs); err != nil &&
		err.Error() != utils.ErrNotFound.Error() {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s processing CGR event %+v with %s.",
				utils.CDRs, err.Error(), cgrEv, utils.ChargerS))
		return
	}
	for _, chrgr := range chrgrs {
		cdr, err := NewMapEvent(chrgr.CGREvent.Event).AsCDR(cdrS.cgrCfg,
			cgrEv.Tenant, cdrS.Timezone())
		if err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s converting CDR event %+v with %s.",
					utils.CDRs, err.Error(), cgrEv, utils.ChargerS))
			continue
		}
		cdrS.raStoReThStaCDR(cdr)

	}
}

// statSProcessEvent will send the event to StatS if the connection is configured
func (cdrS *CdrServer) attrSProcessEvent(cgrEv *utils.CGREvent) (err error) {
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

// V2ProcessCDR will process the CDR out of CGREvent
func (cdrS *CdrServer) V2ProcessCDR(cgrEv *utils.CGREvent, reply *string) (err error) {
	if cdrS.attrS != nil {
		if err := cdrS.attrSProcessEvent(cgrEv); err != nil {
			return utils.NewErrServerError(err)
		}
	}
	rawCDR, err := NewMapEvent(cgrEv.Event).AsCDR(cdrS.cgrCfg, cgrEv.Tenant, cdrS.Timezone())
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if cdrS.chargerS == nil { // backwards compatibility for DerivedChargers
		return cdrS.V1ProcessCDR(rawCDR, reply)
	}
	if cdrS.cgrCfg.CdrsCfg().CDRSStoreCdrs { // Store *raw CDR
		if err = cdrS.cdrDb.SetCDR(rawCDR, false); err != nil {
			return utils.NewErrServerError(err) // Cannot store CDR
		}
	}
	if len(cdrS.cgrCfg.CdrsCfg().CDRSOnlineCDRExports) != 0 {
		cdrS.replicateCDRs([]*CDR{rawCDR}) // Replicate raw CDR
	}
	if cdrS.thdS != nil {
		go cdrS.thdSProcessEvent(cgrEv)
	}
	if cdrS.stats != nil {
		go cdrS.statSProcessEvent(cgrEv)
	}
	if cdrS.chargerS != nil {
		go cdrS.chrgrSProcessEvent(cgrEv)
	}

	*reply = utils.OK
	return nil
}

// Called by rate/re-rate API, RPC method
func (cdrS *CdrServer) V2RateCDRs(attrs *utils.RPCCDRsFilter, reply *string) error {
	if cdrS.chargerS == nil {
		return utils.NewErrNotConnected(utils.ChargerS)
	}
	cdrFltr, err := attrs.AsCDRsFilter(cdrS.cgrCfg.GeneralCfg().DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	cdrs, _, err := cdrS.cdrDb.GetCDRs(cdrFltr, false)
	if err != nil {
		return err
	}
	for _, cdr := range cdrs {
		go cdrS.raStoReThStaCDR(cdr)
	}
	*reply = utils.OK
	return nil
}

// V1GetCDRs returns CDRs from DB
func (self *CdrServer) V1GetCDRs(args utils.RPCCDRsFilter, cdrs *[]*CDR) error {
	cdrsFltr, err := args.AsCDRsFilter(self.Timezone())
	if err != nil {
		if err.Error() != utils.NotFoundCaps {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	if qryCDRs, _, err := self.cdrDb.GetCDRs(cdrsFltr, false); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*cdrs = qryCDRs
	}
	return nil
}

// V1CountCDRs counts CDRs from DB
func (self *CdrServer) V1CountCDRs(args utils.RPCCDRsFilter, cnt *int64) error {
	cdrsFltr, err := args.AsCDRsFilter(self.Timezone())
	if err != nil {
		if err.Error() != utils.NotFoundCaps {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	cdrsFltr.Count = true
	if _, qryCnt, err := self.cdrDb.GetCDRs(cdrsFltr, false); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*cnt = qryCnt
	}
	return nil
}
