/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	// Parse config
	if err := m.parseConfig(); err != nil {
		return nil, err
	}
	return m, nil
}

type Mediator struct {
	connector engine.Connector
	logDb     engine.LogStorage
	cdrDb     engine.CdrStorage
	cgrCfg    *config.CGRConfig
}

func (self *Mediator) parseConfig() error {
	cfgVals := [][]string{self.cgrCfg.MediatorSubjectFields, self.cgrCfg.MediatorReqTypeFields, self.cgrCfg.MediatorDirectionFields,
		self.cgrCfg.MediatorTenantFields, self.cgrCfg.MediatorTORFields, self.cgrCfg.MediatorAccountFields, self.cgrCfg.MediatorDestFields,
		self.cgrCfg.MediatorSetupTimeFields, self.cgrCfg.MediatorAnswerTimeFields, self.cgrCfg.MediatorDurationFields}

	// All other configured fields must match the length of reference fields
	for iCfgVal := range cfgVals {
		if len(self.cgrCfg.MediatorRunIds) != len(cfgVals[iCfgVal]) {
			// Make sure we have everywhere the length of runIds
			return errors.New("Inconsistent lenght of mediator fields.")
		}
	}

	return nil
}

// Retrive the cost from logging database
func (self *Mediator) getCostsFromDB(cgrid string) (cc *engine.CallCost, err error) {
	for i := 0; i < 3; i++ { // Mechanism to avoid concurrency between SessionManager writing the costs and mediator picking them up
		cc, err = self.logDb.GetCallCostLog(cgrid, engine.SESSION_MANAGER_SOURCE, utils.DEFAULT_RUNID) //ToDo: What are we getting when there is no log?
		if cc != nil {                                                                                 // There were no errors, chances are that we got what we are looking for
			break
		}
		time.Sleep(time.Duration(i) * time.Second)
	}
	return
}

// Retrive the cost from engine
func (self *Mediator) getCostsFromRater(cdr *utils.StoredCdr) (*engine.CallCost, error) {
	cc := &engine.CallCost{}
	var err error
	if cdr.Duration == time.Duration(0) { // failed call,  returning empty callcost, no error
		return cc, nil
	}
	cd := engine.CallDescriptor{
		Direction:    "*out", //record[m.directionFields[runIdx]] TODO: fix me
		Tenant:       cdr.Tenant,
		TOR:          cdr.TOR,
		Subject:      cdr.Subject,
		Account:      cdr.Account,
		Destination:  cdr.Destination,
		TimeStart:    cdr.AnswerTime,
		TimeEnd:      cdr.AnswerTime.Add(cdr.Duration),
		LoopIndex:    0,
		CallDuration: cdr.Duration,
	}
	if cdr.ReqType == utils.PSEUDOPREPAID {
		err = self.connector.Debit(cd, cc)
	} else {
		err = self.connector.GetCost(cd, cc)
	}
	if err != nil {
		self.logDb.LogError(cdr.CgrId, engine.MEDIATOR_SOURCE, cdr.MediationRunId, err.Error())
	} else {
		// If the mediator calculated a price it will write it to logdb
		self.logDb.LogCallCost(cdr.AccId, engine.MEDIATOR_SOURCE, cdr.MediationRunId, cc)
	}
	return cc, err
}

func (self *Mediator) rateCDR(cdr *utils.StoredCdr) error {
	var qryCC *engine.CallCost
	var errCost error
	if cdr.ReqType == utils.PREPAID || cdr.ReqType == utils.POSTPAID {
		// Should be previously calculated and stored in DB
		qryCC, errCost = self.getCostsFromDB(cdr.CgrId)
	} else {
		qryCC, errCost = self.getCostsFromRater(cdr)
	}
	if errCost != nil {
		return errCost
	} else if qryCC == nil {
		return errors.New("No cost returned from rater")
	}
	cdr.Cost = qryCC.Cost
	return nil
}

// Forks original CDR based on original request plus runIds for extra mediation
func (self *Mediator) RateCdr(dbcdr utils.RawCDR) error {
	rtCdr, err := utils.NewStoredCdrFromRawCDR(dbcdr)
	if err != nil {
		return err
	}
	cdrs := []*utils.StoredCdr{rtCdr} // Start with initial dbcdr, will add here all to be mediated
	for runIdx, runId := range self.cgrCfg.MediatorRunIds {
		forkedCdr, err := dbcdr.AsStoredCdr(self.cgrCfg.MediatorRunIds[runIdx], self.cgrCfg.MediatorReqTypeFields[runIdx], self.cgrCfg.MediatorDirectionFields[runIdx],
			self.cgrCfg.MediatorTenantFields[runIdx], self.cgrCfg.MediatorTORFields[runIdx], self.cgrCfg.MediatorAccountFields[runIdx],
			self.cgrCfg.MediatorSubjectFields[runIdx], self.cgrCfg.MediatorDestFields[runIdx],
			self.cgrCfg.MediatorSetupTimeFields[runIdx], self.cgrCfg.MediatorAnswerTimeFields[runIdx],
			self.cgrCfg.MediatorDurationFields[runIdx], []string{}, true)
		if err != nil { // Errors on fork, cannot calculate further, write that into db for later analysis
			self.cdrDb.SetRatedCdr(&utils.StoredCdr{CgrId: dbcdr.GetCgrId(), MediationRunId: runId, Cost: -1.0}, err.Error()) // Cannot fork CDR, important just runid and error
			continue
		}
		cdrs = append(cdrs, forkedCdr)
	}
	for _, cdr := range cdrs {
		extraInfo := ""
		if err = self.rateCDR(cdr); err != nil {
			extraInfo = err.Error()
		}
		if err := self.cdrDb.SetRatedCdr(cdr, extraInfo); err != nil {
			engine.Logger.Err(fmt.Sprintf("<Mediator> Could not record cost for cgrid: <%s>, err: <%s>, cost: %f, extraInfo: %s",
				cdr.CgrId, err.Error(), cdr.Cost, extraInfo))
		}
	}
	return nil
}

func (self *Mediator) RateCdrs(timeStart, timeEnd time.Time, rerateErrors, rerateRated bool) error {
	cdrs, err := self.cdrDb.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, timeStart, timeEnd, !rerateErrors, !rerateRated)
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
