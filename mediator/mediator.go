/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package mediator

import (
	"errors"
	"fmt"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewMediator(connector engine.Connector, logDb engine.LogStorage, cdrDb engine.CdrStorage, cfg *config.CGRConfig) (m *Mediator, err error) {
	m = &Mediator{
		connector: connector,
		logDb:     logDb,
		cdrDb:     cdrDb,
		cgrCfg:    cfg,
	}
	return m, nil
}

type Mediator struct {
	connector engine.Connector
	logDb     engine.LogStorage
	cdrDb     engine.CdrStorage
	cgrCfg    *config.CGRConfig
}

// Retrive the cost from logging database, nil in case of no log
func (self *Mediator) getCostsFromDB(cgrid, runId string) (cc *engine.CallCost, err error) {
	for i := 0; i < 3; i++ { // Mechanism to avoid concurrency between SessionManager writing the costs and mediator picking them up
		cc, err = self.logDb.GetCallCostLog(cgrid, engine.SESSION_MANAGER_SOURCE, runId)
		if cc != nil {
			break
		}
		time.Sleep(time.Duration(i) * time.Second)
	}
	return
}

// Retrive the cost from engine
func (self *Mediator) getCostFromRater(storedCdr *utils.StoredCdr) (*engine.CallCost, error) {
	cc := &engine.CallCost{}
	var err error
	if storedCdr.Usage == time.Duration(0) { // failed call,  returning empty callcost, no error
		return cc, nil
	}
	cd := engine.CallDescriptor{
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
	if utils.IsSliceMember([]string{utils.PSEUDOPREPAID, utils.POSTPAID}, storedCdr.ReqType) {
		err = self.connector.Debit(cd, cc)
	} else {
		err = self.connector.GetCost(cd, cc)
	}
	if err != nil {
		self.logDb.LogError(storedCdr.CgrId, engine.MEDIATOR_SOURCE, storedCdr.MediationRunId, err.Error())
	} else {
		// If the mediator calculated a price it will write it to logdb
		self.logDb.LogCallCost(storedCdr.CgrId, engine.MEDIATOR_SOURCE, storedCdr.MediationRunId, cc)
	}
	return cc, err
}

func (self *Mediator) rateCDR(storedCdr *utils.StoredCdr) error {
	var qryCC *engine.CallCost
	var errCost error
	if storedCdr.ReqType == utils.PREPAID {
		// Should be previously calculated and stored in DB
		qryCC, errCost = self.getCostsFromDB(storedCdr.CgrId, storedCdr.MediationRunId)
	} else {
		qryCC, errCost = self.getCostFromRater(storedCdr)
	}
	if errCost != nil {
		return errCost
	} else if qryCC == nil {
		return errors.New("No cost returned from rater")
	}
	storedCdr.Cost = qryCC.Cost
	return nil
}

func (self *Mediator) RateCdr(storedCdr *utils.StoredCdr) error {
	storedCdr.MediationRunId = utils.DEFAULT_RUNID
	cdrRuns := []*utils.StoredCdr{storedCdr}     // Start with initial storCdr, will add here all to be mediated
	attrsDC := utils.AttrDerivedChargers{Tenant: storedCdr.Tenant, Category: storedCdr.Category, Direction: storedCdr.Direction,
		Account: storedCdr.Account, Subject: storedCdr.Subject}
	var dcs utils.DerivedChargers
	if err := self.connector.GetDerivedChargers(attrsDC, &dcs); err != nil {
		errText := fmt.Sprintf("Could not get derived charging for cgrid %s, error: %s", storedCdr.CgrId, err.Error())
		engine.Logger.Err(errText)
		return errors.New(errText)
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
		forkedCdr, err := storedCdr.ForkCdr(dc.RunId, dcReqTypeFld, dcDirFld, dcTenantFld, dcCategoryFld, dcAcntFld, dcSubjFld, dcDstFld, dcSTimeFld, dcATimeFld, dcDurFld,
			[]*utils.RSRField{}, true)
		if err != nil { // Errors on fork, cannot calculate further, write that into db for later analysis
			self.cdrDb.SetRatedCdr(&utils.StoredCdr{CgrId: storedCdr.CgrId, CdrSource: utils.FORKED_CDR, MediationRunId: dc.RunId, Cost: -1},
				err.Error()) // Cannot fork CDR, important just runid and error
			continue
		}
		engine.Logger.Debug(fmt.Sprintf("Appending CdrRun: %+v\n", forkedCdr))
		cdrRuns = append(cdrRuns, forkedCdr)
	}
	for _, cdr := range cdrRuns {
		extraInfo := ""
		if err := self.rateCDR(cdr); err != nil {
			extraInfo = err.Error()
		}
		if err := self.cdrDb.SetRatedCdr(cdr, extraInfo); err != nil {
			engine.Logger.Err(fmt.Sprintf("<Mediator> Could not record cost for cgrid: <%s>, ERROR: <%s>, cost: %f, extraInfo: %s",
				cdr.CgrId, err.Error(), cdr.Cost, extraInfo))
		}
	}
	return nil
}

func (self *Mediator) RateCdrs(timeStart, timeEnd time.Time, rerateErrors, rerateRated bool) error {
	cdrs, err := self.cdrDb.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, timeStart, timeEnd, !rerateErrors, !rerateRated, true)
	if err != nil {
		return err
	}
	for _, cdr := range cdrs {
		if err := self.RateCdr(cdr); err != nil {
			return err
		}
	}
	return nil
}
