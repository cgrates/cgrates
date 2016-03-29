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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
	"github.com/jinzhu/gorm"
	mgov2 "gopkg.in/mgo.v2"
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

func NewCdrServer(cgrCfg *config.CGRConfig, cdrDb CdrStorage, rater Connector, pubsub rpcclient.RpcClientConnection, users UserService, aliases AliasService, stats StatsInterface) (*CdrServer, error) {
	return &CdrServer{cgrCfg: cgrCfg, cdrDb: cdrDb, rater: rater, pubsub: pubsub, users: users, aliases: aliases, stats: stats, guard: &GuardianLock{locksMap: make(map[string]chan bool)}}, nil
}

type CdrServer struct {
	cgrCfg  *config.CGRConfig
	cdrDb   CdrStorage
	rater   Connector
	pubsub  rpcclient.RpcClientConnection
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
func (self *CdrServer) ProcessCdr(cdr *CDR) error {
	return self.processCdr(cdr)
}

// RPC method, used to process external CDRs
func (self *CdrServer) ProcessExternalCdr(eCDR *ExternalCDR) error {
	cdr, err := NewCDRFromExternalCDR(eCDR, self.cgrCfg.DefaultTimezone)
	if err != nil {
		return err
	}
	return self.processCdr(cdr)
}

// RPC method, used to log callcosts to db
func (self *CdrServer) StoreSMCost(smCost *SMCost, checkDuplicate bool) error {
	smCost.CostDetails.UpdateCost()        // make sure the total cost reflect the increments
	smCost.CostDetails.UpdateRatedUsage()  // make sure rated usage is updated
	lockKey := smCost.CGRID + smCost.RunID // Will lock on this ID
	if smCost.CGRID == "" && smCost.OriginID != "" {
		lockKey = smCost.OriginHost + smCost.OriginID
	}
	if checkDuplicate {
		_, err := self.guard.Guard(func() (interface{}, error) {
			smCosts, err := self.cdrDb.GetSMCosts(smCost.CGRID, smCost.RunID, "", "")
			if err != nil {
				return nil, err
			}
			if len(smCosts) != 0 {
				return nil, utils.ErrExists
			}
			return nil, self.cdrDb.SetSMCost(smCost)
		}, 0, lockKey) // FixMe: Possible deadlock with Guard from SMG session close()
		return err
	}
	return self.cdrDb.SetSMCost(smCost)
}

// Called by rate/re-rate API
func (self *CdrServer) RateCDRs(cdrFltr *utils.CDRsFilter, sendToStats bool) error {
	cdrs, _, err := self.cdrDb.GetCDRs(cdrFltr, false)
	if err != nil {
		return err
	}
	for _, cdr := range cdrs {
		// replace user profile fields
		if err := LoadUserProfile(cdr, utils.EXTRA_FIELDS); err != nil {
			return err
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
		if err := self.rateStoreStatsReplicate(cdr, sendToStats); err != nil {
			utils.Logger.Err(fmt.Sprintf("<CDRS> Processing CDR %+v, got error: %s", cdr, err.Error()))
		}
	}
	return nil
}

// Returns error if not able to properly store the CDR, mediation is async since we can always recover offline
func (self *CdrServer) processCdr(cdr *CDR) (err error) {
	if cdr.Direction == "" {
		cdr.Direction = utils.OUT
	}
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
	if !cdr.Rated {
		cdr.RunID = utils.MetaRaw
	}
	if self.cgrCfg.CDRSStoreCdrs { // Store RawCDRs, this we do sync so we can reply with the status
		if cdr.CostDetails != nil {
			cdr.CostDetails.UpdateCost()
			cdr.CostDetails.UpdateRatedUsage()
		}
		if err := self.cdrDb.SetCDR(cdr, false); err != nil { // Only original CDR stored in primary table, no derived
			utils.Logger.Err(fmt.Sprintf("<CDRS> Storing primary CDR %+v, got error: %s", cdr, err.Error()))
			return err // Error is propagated back and we don't continue processing the CDR if we cannot store it
		}
	}
	go self.deriveRateStoreStatsReplicate(cdr)
	return nil
}

// Returns error if not able to properly store the CDR, mediation is async since we can always recover offline
func (self *CdrServer) deriveRateStoreStatsReplicate(cdr *CDR) error {
	cdrRuns, err := self.deriveCdrs(cdr)
	if err != nil {
		return err
	}
	for _, cdrRun := range cdrRuns {
		if err := self.rateStoreStatsReplicate(cdrRun, true); err != nil {
			return err
		}
	}
	return nil
}

func (self *CdrServer) rateStoreStatsReplicate(cdr *CDR, sendToStats bool) error {
	if cdr.RunID == utils.MetaRaw { // Overwrite *raw with *default for rating
		cdr.RunID = utils.META_DEFAULT
	}
	if err := LoadUserProfile(cdr, utils.EXTRA_FIELDS); err != nil {
		return err
	}
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

	// Rate CDR
	if self.rater != nil && !cdr.Rated {
		if err := self.rateCDR(cdr); err != nil {
			cdr.Cost = -1.0 // If there was an error, mark the CDR
			cdr.ExtraInfo = err.Error()
		}
	}
	if cdr.RunID == utils.META_SURETAX { // Request should be processed by SureTax
		if err := SureTaxProcessCdr(cdr); err != nil {
			cdr.Cost = -1.0
			cdr.ExtraInfo = err.Error() // Something failed, write the error in the ExtraInfo
		}
	}
	if self.cgrCfg.CDRSStoreCdrs { // Store CDRs
		// Store RatedCDR
		if cdr.CostDetails != nil {
			cdr.CostDetails.UpdateCost()
			cdr.CostDetails.UpdateRatedUsage()
		}
		if err := self.cdrDb.SetCDR(cdr, true); err != nil {
			utils.Logger.Err(fmt.Sprintf("<CDRS> Storing rated CDR %+v, got error: %s", cdr, err.Error()))
		}
	}
	// Attach CDR to stats
	if self.stats != nil && sendToStats { // Send CDR to stats
		if err := self.stats.AppendCDR(cdr, nil); err != nil {
			utils.Logger.Err(fmt.Sprintf("<CDRS> Could not append CDR to stats: %s", err.Error()))
		}
	}
	if len(self.cgrCfg.CDRSCdrReplication) != 0 {
		self.replicateCdr(cdr)
	}
	return nil
}

func (self *CdrServer) deriveCdrs(cdr *CDR) ([]*CDR, error) {
	cdrRuns := []*CDR{cdr}
	if cdr.RunID != utils.MetaRaw { // Only derive *raw CDRs
		return cdrRuns, nil
	}
	if err := LoadUserProfile(cdr, utils.EXTRA_FIELDS); err != nil {
		return nil, err
	}
	if err := LoadAlias(&AttrMatchingAlias{
		Destination: cdr.Destination,
		Direction:   cdr.Direction,
		Tenant:      cdr.Tenant,
		Category:    cdr.Category,
		Account:     cdr.Account,
		Subject:     cdr.Subject,
		Context:     utils.ALIAS_CONTEXT_RATING,
	}, cdr, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		return nil, err
	}

	attrsDC := &utils.AttrDerivedChargers{Tenant: cdr.Tenant, Category: cdr.Category, Direction: cdr.Direction,
		Account: cdr.Account, Subject: cdr.Subject, Destination: cdr.Destination}
	var dcs utils.DerivedChargers
	if err := self.rater.GetDerivedChargers(attrsDC, &dcs); err != nil {
		utils.Logger.Err(fmt.Sprintf("Could not get derived charging for cgrid %s, error: %s", cdr.CGRID, err.Error()))
		return nil, err
	}
	for _, dc := range dcs.Chargers {
		runFilters, _ := utils.ParseRSRFields(dc.RunFilters, utils.INFIELD_SEP)
		matchingAllFilters := true
		for _, dcRunFilter := range runFilters {
			if fltrPass, _ := cdr.PassesFieldFilter(dcRunFilter); !fltrPass {
				matchingAllFilters = false
				break
			}
		}
		if !matchingAllFilters { // Do not process the derived charger further if not all filters were matched
			continue
		}
		dcRequestTypeFld, _ := utils.NewRSRField(dc.RequestTypeField)
		dcDirFld, _ := utils.NewRSRField(dc.DirectionField)
		dcTenantFld, _ := utils.NewRSRField(dc.TenantField)
		dcCategoryFld, _ := utils.NewRSRField(dc.CategoryField)
		dcAcntFld, _ := utils.NewRSRField(dc.AccountField)
		dcSubjFld, _ := utils.NewRSRField(dc.SubjectField)
		dcDstFld, _ := utils.NewRSRField(dc.DestinationField)
		dcSTimeFld, _ := utils.NewRSRField(dc.SetupTimeField)
		dcPddFld, _ := utils.NewRSRField(dc.PDDField)
		dcATimeFld, _ := utils.NewRSRField(dc.AnswerTimeField)
		dcDurFld, _ := utils.NewRSRField(dc.UsageField)
		dcSupplFld, _ := utils.NewRSRField(dc.SupplierField)
		dcDCauseFld, _ := utils.NewRSRField(dc.DisconnectCauseField)
		dcRatedFld, _ := utils.NewRSRField(dc.RatedField)
		dcCostFld, _ := utils.NewRSRField(dc.CostField)
		forkedCdr, err := cdr.ForkCdr(dc.RunID, dcRequestTypeFld, dcDirFld, dcTenantFld, dcCategoryFld, dcAcntFld, dcSubjFld, dcDstFld,
			dcSTimeFld, dcPddFld, dcATimeFld, dcDurFld, dcSupplFld, dcDCauseFld, dcRatedFld, dcCostFld, []*utils.RSRField{}, true, self.cgrCfg.DefaultTimezone)
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

func (self *CdrServer) rateCDR(cdr *CDR) error {
	var qryCC *CallCost
	var err error
	if cdr.RequestType == utils.META_NONE {
		return nil
	}
	_, hasLastUsed := cdr.ExtraFields[utils.LastUsed]
	if utils.IsSliceMember([]string{utils.META_PREPAID, utils.PREPAID}, cdr.RequestType) && (cdr.Usage != 0 || hasLastUsed) { // ToDo: Get rid of PREPAID as soon as we don't want to support it backwards
		// Should be previously calculated and stored in DB
		delay := utils.Fib()
		var usage float64
		for i := 0; i < 4; i++ {
			smCosts, err := self.cdrDb.GetSMCosts(cdr.CGRID, cdr.RunID, cdr.OriginHost, cdr.ExtraFields[utils.OriginIDPrefix])
			if err == nil && len(smCosts) != 0 {
				qryCC = smCosts[0].CostDetails
				usage = smCosts[0].Usage
				break
			}
			time.Sleep(delay())
		}
		if err != nil && (err == gorm.RecordNotFound || err == mgov2.ErrNotFound) { //calculate CDR as for pseudoprepaid
			utils.Logger.Warning(fmt.Sprintf("<Cdrs> WARNING: Could not find CallCostLog for cgrid: %s, source: %s, runid: %s, will recalculate", cdr.CGRID, utils.SESSION_MANAGER_SOURCE, cdr.RunID))
			qryCC, err = self.getCostFromRater(cdr)
		}
		if cdr.Usage == 0 {
			cdr.Usage = time.Duration(usage)
		}

	} else {
		qryCC, err = self.getCostFromRater(cdr)
	}
	if err != nil {
		return err
	} else if qryCC != nil {
		cdr.Cost = qryCC.Cost
		cdr.CostDetails = qryCC
	}
	return nil
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
		Direction:       cdr.Direction,
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
		err = self.rater.Debit(cd, cc)
	} else {
		err = self.rater.GetCost(cd, cc)
	}
	if err != nil {
		return cc, err
	}
	return cc, nil
}

// ToDo: Add websocket support
func (self *CdrServer) replicateCdr(cdr *CDR) error {
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
			jsn, err := json.Marshal(cdr)
			if err != nil {
				return err
			}
			body = jsn
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
