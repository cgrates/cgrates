/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"path"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/jinzhu/gorm"
)

var cdrServer *CdrServer // Share the server so we can use it in http handlers

type CallCostLog struct {
	CgrId          string
	Source         string
	RunId          string
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
	if err := cdrServer.processCdr(cgrCdr.AsStoredCdr(cdrServer.cgrCfg.DefaultTimezone)); err != nil {
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
	if err := cdrServer.processCdr(fsCdr.AsStoredCdr(cdrServer.Timezone())); err != nil {
		utils.Logger.Err(fmt.Sprintf("<CDRS> Errors when storing CDR entry: %s", err.Error()))
	}
}

func NewCdrServer(cgrCfg *config.CGRConfig, cdrDb CdrStorage, rater Connector, pubsub PublisherSubscriber, users UserService, aliases AliasService, stats StatsInterface) (*CdrServer, error) {
	return &CdrServer{cgrCfg: cgrCfg, cdrDb: cdrDb, rater: rater, pubsub: pubsub, users: users, aliases: aliases, stats: stats, guard: &GuardianLock{queue: make(map[string]chan bool)}}, nil
}

type CdrServer struct {
	cgrCfg  *config.CGRConfig
	cdrDb   CdrStorage
	rater   Connector
	pubsub  PublisherSubscriber
	users   UserService
	aliases AliasService
	stats   StatsInterface
	guard   *GuardianLock
}

func (self *CdrServer) Timezone() string {
	return self.cgrCfg.DefaultTimezone
}

func (self *CdrServer) RegisterHandlersToServer(server *utils.Server) {
	cdrServer = self // Share the server object for handlers
	server.RegisterHttpFunc("/cdr_http", cgrCdrHandler)
	server.RegisterHttpFunc("/freeswitch_json", fsCdrHandler)
}

// RPC method, used to internally process CDR
func (self *CdrServer) ProcessCdr(cdr *StoredCdr) error {
	return self.processCdr(cdr)
}

// RPC method, used to process external CDRs
func (self *CdrServer) ProcessExternalCdr(cdr *ExternalCdr) error {
	storedCdr, err := NewStoredCdrFromExternalCdr(cdr, self.cgrCfg.DefaultTimezone)
	if err != nil {
		return err
	}
	return self.processCdr(storedCdr)
}

// RPC method, used to log callcosts to db
func (self *CdrServer) LogCallCost(ccl *CallCostLog) error {
	if ccl.CheckDuplicate {
		_, err := self.guard.Guard(func() (interface{}, error) {
			cc, err := self.cdrDb.GetCallCostLog(ccl.CgrId, ccl.Source, ccl.RunId)
			if err != nil && err != gorm.RecordNotFound {
				return nil, err
			}
			if cc != nil {
				return nil, utils.ErrExists
			}
			return nil, self.cdrDb.LogCallCost(ccl.CgrId, ccl.Source, ccl.RunId, ccl.CallCost)
		}, 0, ccl.CgrId)
		return err
	}
	return self.cdrDb.LogCallCost(ccl.CgrId, ccl.Source, ccl.RunId, ccl.CallCost)
}

// Called by rate/re-rate API
func (self *CdrServer) RateCdrs(cgrIds, runIds, tors, cdrHosts, cdrSources, reqTypes, directions, tenants, categories, accounts, subjects, destPrefixes, ratedAccounts, ratedSubjects []string,
	orderIdStart, orderIdEnd int64, timeStart, timeEnd time.Time, rerateErrors, rerateRated, sendToStats bool) error {
	var costStart, costEnd *float64
	if rerateErrors {
		costStart = utils.Float64Pointer(-1.0)
		if !rerateRated {
			costEnd = utils.Float64Pointer(0.0)
		}
	} else if rerateRated {
		costStart = utils.Float64Pointer(0.0)
	}
	cdrs, _, err := self.cdrDb.GetStoredCdrs(&utils.CdrsFilter{CgrIds: cgrIds, RunIds: runIds, Tors: tors, CdrHosts: cdrHosts, CdrSources: cdrSources,
		ReqTypes: reqTypes, Directions: directions, Tenants: tenants, Categories: categories, Accounts: accounts,
		Subjects: subjects, DestPrefixes: destPrefixes, RatedAccounts: ratedAccounts, RatedSubjects: ratedSubjects,
		OrderIdStart: orderIdStart, OrderIdEnd: orderIdEnd, AnswerTimeStart: &timeStart, AnswerTimeEnd: &timeEnd,
		MinCost: costStart, MaxCost: costEnd})
	if err != nil {
		return err
	}
	for _, cdr := range cdrs {
		if cdr.MediationRunId == "" { // raw CDRs which were not calculated before
			cdr.MediationRunId = utils.META_DEFAULT
		}
		// replace aliases for cases they were loaded after CDR received
		if err := LoadAlias(&AttrMatchingAlias{
			Destination: cdr.Destination,
			Direction:   cdr.Direction,
			Tenant:      cdr.Tenant,
			Category:    cdr.Category,
			Account:     cdr.Account,
			Subject:     cdr.Subject,
			Context:     utils.ALIAS_CONTEXT_RATING,
		}, cdr, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
			return err
		}
		// replace user profile fields
		if err := LoadUserProfile(cdr, utils.EXTRA_FIELDS); err != nil {
			return err
		}
		if err := self.rateStoreStatsReplicate(cdr); err != nil {
			utils.Logger.Err(fmt.Sprintf("<CDRS> Processing CDR %+v, got error: %s", cdr, err.Error()))
		}
	}
	return nil
}

// Returns error if not able to properly store the CDR, mediation is async since we can always recover offline
func (self *CdrServer) processCdr(storedCdr *StoredCdr) (err error) {
	if storedCdr.Direction == "" {
		storedCdr.Direction = utils.OUT
	}
	if storedCdr.ReqType == "" {
		storedCdr.ReqType = self.cgrCfg.DefaultReqType
	}
	if storedCdr.Tenant == "" {
		storedCdr.Tenant = self.cgrCfg.DefaultTenant
	}
	if storedCdr.Category == "" {
		storedCdr.Category = self.cgrCfg.DefaultCategory
	}
	if storedCdr.Subject == "" { // Use account information as rating subject if missing
		storedCdr.Subject = storedCdr.Account
	}
	// replace aliases
	if err := LoadAlias(&AttrMatchingAlias{
		Destination: storedCdr.Destination,
		Direction:   storedCdr.Direction,
		Tenant:      storedCdr.Tenant,
		Category:    storedCdr.Category,
		Account:     storedCdr.Account,
		Subject:     storedCdr.Subject,
		Context:     utils.ALIAS_CONTEXT_RATING,
	}, storedCdr, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		return err
	}
	// replace user profile fields
	if err := LoadUserProfile(storedCdr, utils.EXTRA_FIELDS); err != nil {
		return err
	}
	if self.cgrCfg.CDRSStoreCdrs { // Store RawCDRs, this we do sync so we can reply with the status
		if err := self.cdrDb.SetCdr(storedCdr); err != nil { // Only original CDR stored in primary table, no derived
			utils.Logger.Err(fmt.Sprintf("<CDRS> Storing primary CDR %+v, got error: %s", storedCdr, err.Error()))
			return err // Error is propagated back and we don't continue processing the CDR if we cannot store it
		}

	}
	go self.deriveRateStoreStatsReplicate(storedCdr)
	return nil
}

// Returns error if not able to properly store the CDR, mediation is async since we can always recover offline
func (self *CdrServer) deriveRateStoreStatsReplicate(storedCdr *StoredCdr) error {
	cdrRuns, err := self.deriveCdrs(storedCdr)
	if err != nil {
		return err
	}
	for _, cdr := range cdrRuns {
		if err := self.rateStoreStatsReplicate(cdr); err != nil {
			return err
		}
	}
	return nil
}

func (self *CdrServer) rateStoreStatsReplicate(cdr *StoredCdr) error {
	if cdr.MediationRunId != utils.META_DEFAULT { // Process Aliases and Users for derived CDRs
		if err := LoadAlias(&AttrMatchingAlias{
			Destination: cdr.Destination,
			Direction:   cdr.Direction,
			Tenant:      cdr.Tenant,
			Category:    cdr.Category,
			Account:     cdr.Account,
			Subject:     cdr.Subject,
			Context:     utils.ALIAS_CONTEXT_RATING,
		}, cdr, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
			return err
		}
		if err := LoadUserProfile(cdr, utils.EXTRA_FIELDS); err != nil {
			return err
		}
	}
	// Rate CDR
	if self.rater != nil && !cdr.Rated {
		if err := self.rateCDR(cdr); err != nil {
			cdr.Cost = -1.0 // If there was an error, mark the CDR
			cdr.ExtraInfo = err.Error()
		}
	}
	if cdr.MediationRunId == utils.META_SURETAX { // Request should be processed by SureTax
		if err := SureTaxProcessCdr(cdr); err != nil {
			cdr.Cost = -1.0
			cdr.ExtraInfo = err.Error() // Something failed, write the error in the ExtraInfo
		}
	}
	if self.cgrCfg.CDRSStoreCdrs { // Store CDRs
		// Store RatedCDR
		if err := self.cdrDb.SetRatedCdr(cdr); err != nil {
			utils.Logger.Err(fmt.Sprintf("<CDRS> Storing rated CDR %+v, got error: %s", cdr, err.Error()))
		}
		// Store CostDetails
		if cdr.Rated || utils.IsSliceMember([]string{utils.RATED, utils.META_RATED}, cdr.ReqType) { // Account related CDRs are saved automatically, so save the others here if requested
			if err := self.cdrDb.LogCallCost(cdr.CgrId, utils.CDRS_SOURCE, cdr.MediationRunId, cdr.CostDetails); err != nil {
				utils.Logger.Err(fmt.Sprintf("<CDRS> Storing costs for CDR %+v, costDetails: %+v, got error: %s", cdr, cdr.CostDetails, err.Error()))
			}
		}
	}
	// Attach CDR to stats
	if self.stats != nil { // Send CDR to stats
		if err := self.stats.AppendCDR(cdr, nil); err != nil {
			utils.Logger.Err(fmt.Sprintf("<CDRS> Could not append cdr to stats: %s", err.Error()))
		}
	}
	if len(self.cgrCfg.CDRSCdrReplication) != 0 {
		self.replicateCdr(cdr)
	}
	return nil
}

func (self *CdrServer) deriveCdrs(storedCdr *StoredCdr) ([]*StoredCdr, error) {
	if len(storedCdr.MediationRunId) == 0 {
		storedCdr.MediationRunId = utils.META_DEFAULT
	}
	cdrRuns := []*StoredCdr{storedCdr}
	if storedCdr.Rated { // Do not derive already rated CDRs since they should be already derived
		return cdrRuns, nil
	}
	attrsDC := &utils.AttrDerivedChargers{Tenant: storedCdr.Tenant, Category: storedCdr.Category, Direction: storedCdr.Direction,
		Account: storedCdr.Account, Subject: storedCdr.Subject, Destination: storedCdr.Destination}
	var dcs utils.DerivedChargers
	if err := self.rater.GetDerivedChargers(attrsDC, &dcs); err != nil {
		utils.Logger.Err(fmt.Sprintf("Could not get derived charging for cgrid %s, error: %s", storedCdr.CgrId, err.Error()))
		return nil, err
	}
	for _, dc := range dcs.Chargers {
		runFilters, _ := utils.ParseRSRFields(dc.RunFilters, utils.INFIELD_SEP)
		matchingAllFilters := true
		for _, dcRunFilter := range runFilters {
			if fltrPass, _ := storedCdr.PassesFieldFilter(dcRunFilter); !fltrPass {
				matchingAllFilters = false
				break
			}
		}
		if !matchingAllFilters { // Do not process the derived charger further if not all filters were matched

			continue
		}
		dcReqTypeFld, _ := utils.NewRSRField(dc.ReqTypeField)
		dcDirFld, _ := utils.NewRSRField(dc.DirectionField)
		dcTenantFld, _ := utils.NewRSRField(dc.TenantField)
		dcCategoryFld, _ := utils.NewRSRField(dc.CategoryField)
		dcAcntFld, _ := utils.NewRSRField(dc.AccountField)
		dcSubjFld, _ := utils.NewRSRField(dc.SubjectField)
		dcDstFld, _ := utils.NewRSRField(dc.DestinationField)
		dcSTimeFld, _ := utils.NewRSRField(dc.SetupTimeField)
		dcPddFld, _ := utils.NewRSRField(dc.PddField)
		dcATimeFld, _ := utils.NewRSRField(dc.AnswerTimeField)
		dcDurFld, _ := utils.NewRSRField(dc.UsageField)
		dcSupplFld, _ := utils.NewRSRField(dc.SupplierField)
		dcDCauseFld, _ := utils.NewRSRField(dc.DisconnectCauseField)
		dcRatedFld, _ := utils.NewRSRField(dc.RatedField)
		dcCostFld, _ := utils.NewRSRField(dc.CostField)
		forkedCdr, err := storedCdr.ForkCdr(dc.RunId, dcReqTypeFld, dcDirFld, dcTenantFld, dcCategoryFld, dcAcntFld, dcSubjFld, dcDstFld,
			dcSTimeFld, dcPddFld, dcATimeFld, dcDurFld, dcSupplFld, dcDCauseFld, dcRatedFld, dcCostFld, []*utils.RSRField{}, true, self.cgrCfg.DefaultTimezone)
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("Could not fork CGR with cgrid %s, run: %s, error: %s", storedCdr.CgrId, dc.RunId, err.Error()))
			continue // do not add it to the forked CDR list
		}
		if !forkedCdr.Rated {
			forkedCdr.Cost = -1.0 // Make sure that un-rated CDRs start with Cost -1
		}
		cdrRuns = append(cdrRuns, forkedCdr)
	}
	return cdrRuns, nil
}

// Retrive the cost from engine
func (self *CdrServer) getCostFromRater(storedCdr *StoredCdr) (*CallCost, error) {
	cc := new(CallCost)
	var err error
	timeStart := storedCdr.AnswerTime
	if timeStart.IsZero() { // Fix for FreeSWITCH unanswered calls
		timeStart = storedCdr.SetupTime
	}
	cd := &CallDescriptor{
		TOR:           storedCdr.TOR,
		Direction:     storedCdr.Direction,
		Tenant:        storedCdr.Tenant,
		Category:      storedCdr.Category,
		Subject:       storedCdr.Subject,
		Account:       storedCdr.Account,
		Destination:   storedCdr.Destination,
		TimeStart:     timeStart,
		TimeEnd:       timeStart.Add(storedCdr.Usage),
		DurationIndex: storedCdr.Usage,
	}
	if utils.IsSliceMember([]string{utils.META_PSEUDOPREPAID, utils.META_POSTPAID, utils.META_PREPAID, utils.PSEUDOPREPAID, utils.POSTPAID, utils.PREPAID}, storedCdr.ReqType) { // Prepaid - Cost can be recalculated in case of missing records from SM
		if err = self.rater.Debit(cd, cc); err == nil { // Debit has occured, we are forced to write the log, even if CDR store is disabled
			self.cdrDb.LogCallCost(storedCdr.CgrId, utils.CDRS_SOURCE, storedCdr.MediationRunId, cc)
		}
	} else {
		err = self.rater.GetCost(cd, cc)
	}
	if err != nil {
		return cc, err
	}
	return cc, nil
}

func (self *CdrServer) rateCDR(storedCdr *StoredCdr) error {
	var qryCC *CallCost
	var err error
	if storedCdr.ReqType == utils.META_NONE {
		return nil
	}
	if utils.IsSliceMember([]string{utils.META_PREPAID, utils.PREPAID}, storedCdr.ReqType) && storedCdr.Usage != 0 { // ToDo: Get rid of PREPAID as soon as we don't want to support it backwards
		// Should be previously calculated and stored in DB
		delay := utils.Fib()
		for i := 0; i < 4; i++ {
			qryCC, err = self.cdrDb.GetCallCostLog(storedCdr.CgrId, utils.SESSION_MANAGER_SOURCE, storedCdr.MediationRunId)
			if err == nil {
				break
			}
			time.Sleep(delay())
		}
		if err != nil && err == gorm.RecordNotFound { //calculate CDR as for pseudoprepaid
			utils.Logger.Warning(fmt.Sprintf("<Cdrs> WARNING: Could not find CallCostLog for cgrid: %s, source: %s, runid: %s, will recalculate", storedCdr.CgrId, utils.SESSION_MANAGER_SOURCE, storedCdr.MediationRunId))
			qryCC, err = self.getCostFromRater(storedCdr)
		}

	} else {
		qryCC, err = self.getCostFromRater(storedCdr)
	}
	if err != nil {
		return err
	} else if qryCC != nil {
		storedCdr.Cost = qryCC.Cost
		storedCdr.CostDetails = qryCC
	}
	return nil
}

// ToDo: Add websocket support
func (self *CdrServer) replicateCdr(cdr *StoredCdr) error {
	for _, rplCfg := range self.cgrCfg.CDRSCdrReplication {
		passesFilters := true
		for _, cdfFltr := range rplCfg.CdrFilter {
			if fltrPass, _ := cdr.PassesFieldFilter(cdfFltr); !fltrPass {
				passesFilters = false
				break
			}
		}
		if !passesFilters { // Not passes filters, ignore this replication
			continue
		}
		var body interface{}
		var content = ""
		switch rplCfg.Transport {
		case utils.META_HTTP_POST:
			content = utils.CONTENT_FORM
			body = cdr.AsHttpForm()
		case utils.META_HTTP_JSON:
			content = utils.CONTENT_JSON
			body = cdr
		}

		errChan := make(chan error)
		go func(body interface{}, rplCfg *config.CdrReplicationCfg, content string, errChan chan error) {
			fallbackPath := path.Join(
				self.cgrCfg.HttpFailedDir,
				rplCfg.FallbackFileName())
			_, err := utils.HttpPoster(
				rplCfg.Server, self.cgrCfg.HttpSkipTlsVerify, body,
				content, rplCfg.Attempts, fallbackPath)
			if err != nil {
				utils.Logger.Err(fmt.Sprintf(
					"<CDRReplicator> Replicating CDR: %+v, got error: %s", cdr, err.Error()))
				errChan <- err
			}
			errChan <- nil

		}(body, rplCfg, content, errChan)
		if rplCfg.Synchronous { // Synchronize here
			<-errChan
		}

	}
	return nil
}
