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
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var cdrServer *CdrServer // Share the server so we can use it in http handlers

// Handler for generic cgr cdr http
func cgrCdrHandler(w http.ResponseWriter, r *http.Request) {
	cgrCdr, err := NewCgrCdrFromHttpReq(r)
	if err != nil {
		Logger.Err(fmt.Sprintf("<CDRS> Could not create CDR entry: %s", err.Error()))
	}
	if err := cdrServer.rateStoreStatsReplicate(cgrCdr.AsStoredCdr()); err != nil {
		Logger.Err(fmt.Sprintf("<CDRS> Errors when storing CDR entry: %s", err.Error()))
	}
}

// Handler for fs http
func fsCdrHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	fsCdr, err := NewFSCdr(body, cdrServer.cgrCfg)
	if err != nil {
		Logger.Err(fmt.Sprintf("<CDRS> Could not create CDR entry: %s", err.Error()))
	}
	if err := cdrServer.rateStoreStatsReplicate(fsCdr.AsStoredCdr()); err != nil {
		Logger.Err(fmt.Sprintf("<CDRS> Errors when storing CDR entry: %s", err.Error()))
	}
}

func NewCdrServer(cgrCfg *config.CGRConfig, cdrDb CdrStorage, rater Connector, stats StatsInterface) (*CdrServer, error) {
	return &CdrServer{cgrCfg: cgrCfg, cdrDb: cdrDb, rater: rater, stats: stats}, nil
	/*
		if cfg.CDRSStats != "" {
			if cfg.CDRSStats != utils.INTERNAL {
				if s, err := NewProxyStats(cfg.CDRSStats); err == nil {
					stats = s
				} else {
					Logger.Err(fmt.Sprintf("<CDRS> Errors connecting to CDRS stats service : %s", err.Error()))
				}
			}
		} else {
			// disable stats for cdrs
			stats = nil
		}
	*/
}

type CdrServer struct {
	cgrCfg *config.CGRConfig
	cdrDb  CdrStorage
	rater  Connector
	stats  StatsInterface
}

func (self *CdrServer) RegisterHanlersToServer(server *Server) {
	cdrServer = self // Share the server object for handlers
	server.RegisterHttpFunc("/cdr_post", cgrCdrHandler)
	server.RegisterHttpFunc("/freeswitch_json", fsCdrHandler)
}

// RPC method, used to internally process CDR
func (self *CdrServer) ProcessCdr(cdr *StoredCdr) error {
	return self.rateStoreStatsReplicate(cdr)
}

// RPC method, used to process external CDRs
func (self *CdrServer) ProcessExternalCdr(cdr *ExternalCdr) error {
	storedCdr, err := NewStoredCdrFromExternalCdr(cdr)
	if err != nil {
		return err
	}
	return self.rateStoreStatsReplicate(storedCdr)
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
		CostStart: costStart, CostEnd: costEnd})
	if err != nil {
		return err
	}
	for _, cdr := range cdrs {
		if err := self.rateStoreStatsReplicate(cdr); err != nil {
			Logger.Err(fmt.Sprintf("<CDRS> Processing CDR %+v, got error: %s", cdr, err.Error()))
		}
	}
	return nil
}

// Returns error if not able to properly store the CDR, mediation is async since we can always recover offline
func (self *CdrServer) rateStoreStatsReplicate(storedCdr *StoredCdr) (err error) {
	if storedCdr.ReqType == utils.META_NONE {
		return nil
	}
	cdrs := []*StoredCdr{storedCdr}
	if self.rater != nil && !storedCdr.Rated { // Rate CDR
		if cdrs, err = self.deriveAndRateCdr(storedCdr); err != nil {
			return err
		}
	}
	if self.cgrCfg.CDRSStoreCdrs { // Store CDRs
		// Store RawCdr
		if err := self.cdrDb.SetCdr(storedCdr); err != nil { // Only original CDR stored in primary table, no derived
			Logger.Err(fmt.Sprintf("<CDRS> Storing primary CDR %+v, got error: %s", storedCdr, err.Error()))
		}
		// Store rated CDRs (including derived)
		for _, cdr := range cdrs {
			if len(cdr.MediationRunId) == 0 { // Do not store rating info for rawCDRs
				continue
			}
			if err := self.cdrDb.SetRatedCdr(cdr); err != nil {
				Logger.Err(fmt.Sprintf("<CDRS> Storing rated CDR %+v, got error: %s", cdr, err.Error()))
			}
			// Store CostDetails
			if cdr.Rated || utils.IsSliceMember([]string{utils.RATED, utils.META_RATED}, cdr.ReqType) { // Account related CDRs are saved automatically, so save the others here if requested
				if err := self.cdrDb.LogCallCost(cdr.CgrId, utils.CDRS_SOURCE, cdr.MediationRunId, storedCdr.CostDetails); err != nil {
					Logger.Err(fmt.Sprintf("<CDRS> Storing costs for CDR %+v, costDetails: %+v, got error: %s", cdr, cdr.CostDetails, err.Error()))
				}
			}
		}
	}
	if self.stats != nil { // Send CDR to stats
		for _, cdr := range cdrs {
			go func(storedCdr *StoredCdr) {
				if err := self.stats.AppendCDR(storedCdr, nil); err != nil {
					Logger.Err(fmt.Sprintf("<CDRS> Could not append cdr to stats: %s", err.Error()))
				}
			}(cdr)
		}
	}
	if self.cgrCfg.CDRSCdrReplication != nil {
		for _, cdr := range cdrs {
			self.replicateCdr(cdr)
		}
	}
	return nil
}

// Derive the original CDR based on derivedCharging rules and calculate costs for each. Returns the results
func (self *CdrServer) deriveAndRateCdr(storedCdr *StoredCdr) ([]*StoredCdr, error) {
	cdrRuns, err := self.deriveCdrs(storedCdr)
	if err != nil {
		return nil, err
	}
	for _, cdr := range cdrRuns {
		if err := self.rateCDR(cdr); err != nil {
			cdr.Cost = -1.0 // If there was an error, mark the CDR
			cdr.ExtraInfo = err.Error()
		}
	}
	return cdrRuns, nil
}

// Retrive the cost from logging database, nil in case of no log
func (self *CdrServer) getCostsFromDB(cgrid, runId string) (cc *CallCost, err error) {
	for i := 0; i < 3; i++ { // Mechanism to avoid concurrency between SessionManager writing the costs and mediator picking them up
		cc, err = self.cdrDb.GetCallCostLog(cgrid, SESSION_MANAGER_SOURCE, runId)
		if cc != nil {
			break
		}
		time.Sleep(time.Duration((i+1)*10) * time.Millisecond)
	}
	return
}

// Retrive the cost from engine
func (self *CdrServer) getCostFromRater(storedCdr *StoredCdr) (*CallCost, error) {
	//if storedCdr.Usage == time.Duration(0) { // failed call,  nil cost
	//	return nil, nil // No costs present, better than empty call cost since could lead us to 0 costs
	//}
	cc := new(CallCost)
	var err error
	cd := CallDescriptor{
		TOR:           storedCdr.TOR,
		Direction:     storedCdr.Direction,
		Tenant:        storedCdr.Tenant,
		Category:      storedCdr.Category,
		Subject:       storedCdr.Subject,
		Account:       storedCdr.Account,
		Destination:   storedCdr.Destination,
		TimeStart:     storedCdr.AnswerTime,
		TimeEnd:       storedCdr.AnswerTime.Add(storedCdr.Usage),
		DurationIndex: storedCdr.Usage,
	}
	if utils.IsSliceMember([]string{utils.META_PSEUDOPREPAID, utils.META_POSTPAID, utils.PSEUDOPREPAID, utils.POSTPAID}, storedCdr.ReqType) {
		if err = self.rater.Debit(cd, cc); err == nil { // Debit has occured, we are forced to write the log, even if CDR store is disabled
			self.cdrDb.LogCallCost(storedCdr.CgrId, MEDIATOR_SOURCE, storedCdr.MediationRunId, cc)
		}
	} else {
		err = self.rater.GetCost(cd, cc)
	}
	if err != nil {
		return nil, err
	}
	return cc, nil
}

func (self *CdrServer) deriveCdrs(storedCdr *StoredCdr) ([]*StoredCdr, error) {
	if len(storedCdr.MediationRunId) == 0 {
		storedCdr.MediationRunId = utils.META_DEFAULT
	}
	cdrRuns := []*StoredCdr{storedCdr}
	if storedCdr.Rated { // Do not derive already rated CDRs since they should be already derived
		return cdrRuns, nil
	}
	attrsDC := utils.AttrDerivedChargers{Tenant: storedCdr.Tenant, Category: storedCdr.Category, Direction: storedCdr.Direction,
		Account: storedCdr.Account, Subject: storedCdr.Subject}
	var dcs utils.DerivedChargers
	if err := self.rater.GetDerivedChargers(attrsDC, &dcs); err != nil {
		Logger.Err(fmt.Sprintf("Could not get derived charging for cgrid %s, error: %s", storedCdr.CgrId, err.Error()))
		return nil, err
	}
	for _, dc := range dcs {
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
		dcATimeFld, _ := utils.NewRSRField(dc.AnswerTimeField)
		dcDurFld, _ := utils.NewRSRField(dc.UsageField)
		dcSupplFld, _ := utils.NewRSRField(dc.SupplierField)
		dcDCausseld, _ := utils.NewRSRField(dc.DisconnectCauseField)
		dcPddFld, _ := utils.NewRSRField("0") // FixMe
		forkedCdr, err := storedCdr.ForkCdr(dc.RunId, dcReqTypeFld, dcDirFld, dcTenantFld, dcCategoryFld, dcAcntFld, dcSubjFld, dcDstFld,
			dcSTimeFld, dcATimeFld, dcDurFld, dcPddFld, dcSupplFld, dcDCausseld, []*utils.RSRField{}, true)
		if err != nil {
			Logger.Err(fmt.Sprintf("Could not fork CGR with cgrid %s, run: %s, error: %s", storedCdr.CgrId, dc.RunId, err.Error()))
			continue // do not add it to the forked CDR list
		}
		cdrRuns = append(cdrRuns, forkedCdr)
	}
	return cdrRuns, nil
}

func (self *CdrServer) rateCDR(storedCdr *StoredCdr) error {
	var qryCC *CallCost
	var errCost error
	if utils.IsSliceMember([]string{utils.META_PREPAID, utils.PREPAID}, storedCdr.ReqType) { // ToDo: Get rid of PREPAID as soon as we don't want to support it backwards
		// Should be previously calculated and stored in DB
		qryCC, errCost = self.getCostsFromDB(storedCdr.CgrId, storedCdr.MediationRunId)
	} else {
		qryCC, errCost = self.getCostFromRater(storedCdr)
	}
	if errCost != nil {
		return errCost
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
		switch rplCfg.Transport {
		case utils.META_HTTP_POST:
			httpClient := new(http.Client)
			errChan := make(chan error)
			go func(cdr *StoredCdr, rplCfg *config.CdrReplicationCfg, errChan chan error) {
				if _, err := httpClient.PostForm(fmt.Sprintf("http://%s/cdr_post", rplCfg.Server), cdr.AsHttpForm()); err != nil {
					Logger.Err(fmt.Sprintf("<CDRReplicator> Replicating CDR: %+v, got error: %s", cdr, err.Error()))
					errChan <- err
				}
				errChan <- nil
			}(cdr, rplCfg, errChan)
			if rplCfg.Synchronous { // Synchronize here
				<-errChan
			}
		}
	}
	return nil
}
