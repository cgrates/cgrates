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

	"github.com/cgrates/cgrates/cache"
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
	cgrCdr, err := NewCgrCdrFromHttpReq(r, cdrServer.cgrCfg.DefaultTimezone)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<CDRS> Could not create CDR entry: %s", err.Error()))
		return
	}
	if err := cdrServer.processCdr(cgrCdr.AsCDR(cdrServer.cgrCfg.DefaultTimezone)); err != nil {
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
	attrs, users, aliases, cdrstats, thdS, stats rpcclient.RpcClientConnection) (*CdrServer, error) {
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
	if cdrstats != nil && reflect.ValueOf(cdrstats).IsNil() {
		cdrstats = nil
	}
	if thdS != nil && reflect.ValueOf(thdS).IsNil() {
		thdS = nil
	}
	if stats != nil && reflect.ValueOf(stats).IsNil() {
		stats = nil
	}
	return &CdrServer{cgrCfg: cgrCfg, cdrDb: cdrDb, dm: dm,
		rals: rater, pubsub: pubsub, users: users, aliases: aliases,
		cdrstats: cdrstats, stats: stats, thdS: thdS, guard: guardian.Guardian,
		httpPoster: utils.NewHTTPPoster(cgrCfg.HttpSkipTlsVerify, cgrCfg.ReplyTimeout)}, nil
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
	cdrstats      rpcclient.RpcClientConnection
	thdS          rpcclient.RpcClientConnection
	stats         rpcclient.RpcClientConnection
	guard         *guardian.GuardianLock
	responseCache *cache.ResponseCache
	httpPoster    *utils.HTTPPoster // used for replication
}

func (self *CdrServer) Timezone() string {
	return self.cgrCfg.DefaultTimezone
}
func (self *CdrServer) SetTimeToLive(timeToLive time.Duration, out *int) error {
	self.responseCache = cache.NewResponseCache(timeToLive)
	return nil
}

func (self *CdrServer) getCache() *cache.ResponseCache {
	if self.responseCache == nil {
		self.responseCache = cache.NewResponseCache(0)
	}
	return self.responseCache
}

func (self *CdrServer) RegisterHandlersToServer(server *utils.Server) {
	cdrServer = self // Share the server object for handlers
	server.RegisterHttpFunc("/cdr_http", cgrCdrHandler)
	server.RegisterHttpFunc("/freeswitch_json", fsCdrHandler)
}

// Used to process external CDRs
func (self *CdrServer) ProcessExternalCdr(eCDR *ExternalCDR) error {
	cdr, err := NewCDRFromExternalCDR(eCDR, self.cgrCfg.DefaultTimezone)
	if err != nil {
		return err
	}
	return self.processCdr(cdr)
}

func (self *CdrServer) storeSMCost(smCost *SMCost, checkDuplicate bool) error {
	smCost.CostDetails.UpdateCost()                                              // make sure the total cost reflect the increments
	smCost.CostDetails.UpdateRatedUsage()                                        // make sure rated usage is updated
	lockKey := utils.CDRS_SOURCE + smCost.CGRID + smCost.RunID + smCost.OriginID // Will lock on this ID
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
		cdr.RequestType = self.cgrCfg.DefaultReqType
	}
	if cdr.Tenant == "" {
		cdr.Tenant = self.cgrCfg.DefaultTenant
	}
	if cdr.Category == "" {
		cdr.Category = self.cgrCfg.DefaultCategory
	}
	if cdr.Subject == "" { // Use account information as rating subject if missing
		cdr.Subject = cdr.Account
	}
	if !cdr.Rated { // Enforce the RunID if CDR is not rated
		cdr.RunID = utils.MetaRaw
	}
	if cdr.RunID == utils.MetaRaw {
		cdr.Cost = -1.0
	}
	if self.cgrCfg.CDRSStoreCdrs { // Store RawCDRs, this we do sync so we can reply with the status
		if cdr.CostDetails != nil {
			cdr.CostDetails.UpdateCost()
			cdr.CostDetails.UpdateRatedUsage()
		}
		if err := self.cdrDb.SetCDR(cdr, false); err != nil {
			utils.Logger.Err(fmt.Sprintf("<CDRS> Storing primary CDR %+v, got error: %s", cdr, err.Error()))
			return err // Error is propagated back and we don't continue processing the CDR if we cannot store it
		}
	}
	if self.thdS != nil {
		var hits int
		thEv := &ArgsProcessEvent{
			CGREvent: *cdr.AsCGREvent()}
		if err := self.thdS.Call(utils.ThresholdSv1ProcessEvent, thEv, &hits); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<CDRS> error: %s processing CDR event %+v with thdS.", err.Error(), thEv))
		}
	}
	// Attach raw CDR to stats
	if self.cdrstats != nil { // Send raw CDR to stats
		var out int
		go self.cdrstats.Call("CDRStatsV1.AppendCDR", cdr, &out)
	}
	if self.stats != nil {
		var reply string
		go self.stats.Call(utils.StatSv1ProcessEvent, cdr.AsCGREvent(), &reply)
	}
	if len(self.cgrCfg.CDRSOnlineCDRExports) != 0 { // Replicate raw CDR
		self.replicateCDRs([]*CDR{cdr})
	}
	if self.rals != nil && !cdr.Rated { // CDRs not rated will be processed by Rating
		go self.deriveRateStoreStatsReplicate(cdr, self.cgrCfg.CDRSStoreCdrs,
			self.cdrstats != nil, len(self.cgrCfg.CDRSOnlineCDRExports) != 0)
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
			if err = self.attrS.Call(utils.AttributeSv1ProcessEvent,
				cdrRun.AsCGREvent(), &rplyEv); err != nil {
				return
			}
			if err = cdrRun.UpdateFromCGREvent(rplyEv.CGREvent,
				rplyEv.AlteredFields); err != nil {
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
			if ratedCDR.CostDetails != nil {
				ratedCDR.CostDetails.UpdateCost()
				ratedCDR.CostDetails.UpdateRatedUsage()
			}
			if err := self.cdrDb.SetCDR(ratedCDR, true); err != nil {
				utils.Logger.Err(fmt.Sprintf("<CDRS> Storing rated CDR %+v, got error: %s", ratedCDR, err.Error()))
			}
		}
	}
	// Attach CDR to stats
	if cdrstats { // Send CDR to stats
		for _, ratedCDR := range ratedCDRs {
			if self.cdrstats != nil {
				var out int
				if err := self.cdrstats.Call("CDRStatsV1.AppendCDR", ratedCDR, &out); err != nil {
					utils.Logger.Err(fmt.Sprintf("<CDRS> Could not send CDR to cdrstats: %s", err.Error()))
				}
			}
			if self.stats != nil {
				var reply string
				go self.stats.Call(utils.StatSv1ProcessEvent, ratedCDR.AsCGREvent(), &reply)
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
			if !dcRunFilter.FilterPasses(cdr.FieldAsString(dcRunFilter)) {
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
		dcRatedFld, _ := utils.NewRSRField(dc.RatedField)
		dcCostFld, _ := utils.NewRSRField(dc.CostField)

		dcExtraFields := []*utils.RSRField{}
		for key, _ := range cdr.ExtraFields {
			dcExtraFields = append(dcExtraFields, &utils.RSRField{Id: key})
		}

		forkedCdr, err := cdr.ForkCdr(dc.RunID, dcRequestTypeFld, dcTenantFld, dcCategoryFld, dcAcntFld, dcSubjFld, dcDstFld,
			dcSTimeFld, dcATimeFld, dcDurFld, dcRatedFld, dcCostFld, dcExtraFields, true, self.cgrCfg.DefaultTimezone)
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("Could not fork CGR with cgrid %s, run: %s, error: %s", cdr.CGRID, dc.RunID, err.Error()))
			continue // do not add it to the forked CDR list
		}
		if !forkedCdr.Rated {
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
		for i := 0; i < self.cgrCfg.CDRSSMCostRetries; i++ {
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
				cdrClone.Cost = smCost.CostDetails.Cost
				cdrClone.CostDetails = smCost.CostDetails
				cdrClone.CostSource = smCost.CostSource
				cdrsRated = append(cdrsRated, cdrClone)
			}
			return cdrsRated, nil
		} else { //calculate CDR as for pseudoprepaid
			utils.Logger.Warning(
				fmt.Sprintf("<Cdrs> WARNING: Could not find CallCostLog for cgrid: %s, source: %s, runid: %s, originID: %s, will recalculate",
					cdr.CGRID, utils.SESSION_MANAGER_SOURCE, cdr.RunID, cdr.OriginID))
			qryCC, err = self.getCostFromRater(cdr)
		}
	} else {
		qryCC, err = self.getCostFromRater(cdr)
	}
	if err != nil {
		return nil, err
	} else if qryCC != nil {
		cdr.Cost = qryCC.Cost
		cdr.CostDetails = qryCC
	}
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
	if utils.IsSliceMember([]string{utils.META_PSEUDOPREPAID, utils.META_POSTPAID, utils.META_PREPAID, utils.PSEUDOPREPAID, utils.POSTPAID, utils.PREPAID}, cdr.RequestType) { // Prepaid - Cost can be recalculated in case of missing records from SM
		err = self.rals.Call("Responder.Debit", cd, cc)
	} else {
		err = self.rals.Call("Responder.GetCost", cd, cc)
	}
	if err != nil {
		return cc, err
	}
	cdr.CostSource = utils.CDRS_SOURCE
	return cc, nil
}

func (self *CdrServer) replicateCDRs(cdrs []*CDR) (err error) {
	for _, exportID := range self.cgrCfg.CDRSOnlineCDRExports {
		expTpl := self.cgrCfg.CdreProfiles[exportID] // not checking for existence of profile since this should be done in a higher layer
		var cdre *CDRExporter
		if cdre, err = NewCDRExporter(cdrs, expTpl, expTpl.ExportFormat, expTpl.ExportPath, self.cgrCfg.FailedPostsDir, "CDRSReplication",
			expTpl.Synchronous, expTpl.Attempts, expTpl.FieldSeparator, expTpl.UsageMultiplyFactor,
			expTpl.CostMultiplyFactor, self.cgrCfg.RoundingDecimals, self.cgrCfg.HttpSkipTlsVerify, self.httpPoster); err != nil {
			utils.Logger.Err(fmt.Sprintf("<CDRS> Building CDRExporter for online exports got error: <%s>", err.Error()))
			continue
		}
		if err = cdre.ExportCDRs(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<CDRS> Replicating CDR: %+v, got error: <%s>", err.Error()))
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
		if err := self.deriveRateStoreStatsReplicate(cdr, self.cgrCfg.CDRSStoreCdrs, sendToStats, len(self.cgrCfg.CDRSOnlineCDRExports) != 0); err != nil {
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
		self.getCache().Cache(cacheKey, &cache.CacheItem{Err: err})
		return utils.NewErrServerError(err)
	}
	self.getCache().Cache(cacheKey, &cache.CacheItem{Value: utils.OK})
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
		self.getCache().Cache(cacheKey, &cache.CacheItem{Err: err})
		return utils.NewErrServerError(err)
	}
	self.getCache().Cache(cacheKey, &cache.CacheItem{Value: utils.OK})
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
		CostDetails: cc,
	}, args.CheckDuplicate); err != nil {
		cdrs.getCache().Cache(cacheKey, &cache.CacheItem{Err: err})
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	cdrs.getCache().Cache(cacheKey, &cache.CacheItem{Value: *reply})
	return nil

}

// Called by rate/re-rate API, RPC method
func (self *CdrServer) V1RateCDRs(attrs utils.AttrRateCDRs, reply *string) error {
	cdrFltr, err := attrs.RPCCDRsFilter.AsCDRsFilter(self.cgrCfg.DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	cdrs, _, err := self.cdrDb.GetCDRs(cdrFltr, false)
	if err != nil {
		return err
	}
	storeCDRs := self.cgrCfg.CDRSStoreCdrs
	if attrs.StoreCDRs != nil {
		storeCDRs = *attrs.StoreCDRs
	}
	sendToStats := self.cdrstats != nil
	if attrs.SendToStatS != nil {
		sendToStats = *attrs.SendToStatS
	}
	replicate := len(self.cgrCfg.CDRSOnlineCDRExports) != 0
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
