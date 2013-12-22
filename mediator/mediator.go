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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"strconv"
	"time"
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
	connector           engine.Connector
	logDb               engine.LogStorage
	cdrDb               engine.CdrStorage
	cgrCfg              *config.CGRConfig
}

/*
Responsible for parsing configuration fields out of CGRConfig instance, doing the necessary pre-checks.
@param cfgVals: keep ordered references to configuration fields from CGRConfig instance.
fieldKeys and cfgVals are directly related through index.
Method logic:
 * Make sure the field used as reference in mediation process loop is not empty.
 * All other fields should match the length of reference field.
 * Accounting id field should not be empty.
 * If we run mediation on csv file:
  * Make sure cdrInDir and cdrOutDir are valid paths.
  * Populate fieldIdxs by converting fieldNames into integers
*/
func (self *Mediator) parseConfig() error {
	cfgVals := [][]string{self.cgrCfg.MediatorSubjectFields, self.cgrCfg.MediatorReqTypeFields, self.cgrCfg.MediatorDirectionFields,
		self.cgrCfg.MediatorTenantFields, self.cgrCfg.MediatorTORFields, self.cgrCfg.MediatorAccountFields, self.cgrCfg.MediatorDestFields,
		self.cgrCfg.MediatorAnswerTimeFields, self.cgrCfg.MediatorDurationFields}

	if len(self.cgrCfg.MediatorRunIds) == 0 {
		return errors.New("Unconfigured mediator run_ids")
	}
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
func (self *Mediator) getCostsFromDB(cdr utils.CDR) (cc *engine.CallCost, err error) {
	for i := 0; i < 3; i++ { // Mechanism to avoid concurrency between SessionManager writing the costs and mediator picking them up
		cc, err = self.logDb.GetCallCostLog(cdr.GetCgrId(), engine.SESSION_MANAGER_SOURCE) //ToDo: What are we getting when there is no log?
		if cc != nil {                                                                     // There were no errors, chances are that we got what we are looking for
			break
		}
		time.Sleep(time.Duration(i) * time.Second)
	}
	return
}

// Retrive the cost from engine
func (self *Mediator) getCostsFromRater(cdr utils.CDR) (*engine.CallCost, error) {
	cc := &engine.CallCost{}
	d, err := time.ParseDuration(strconv.FormatInt(cdr.GetDuration(), 10) + "s")
	if err != nil {
		return nil, err
	}
	if d.Seconds() == 0 { // failed call,  returning empty callcost, no error
		return cc, nil
	}
	t1, err := cdr.GetAnswerTime()
	if err != nil {
		return nil, err
	}
	cd := engine.CallDescriptor{
		Direction:    "*out", //record[m.directionFields[runIdx]] TODO: fix me
		Tenant:       cdr.GetTenant(),
		TOR:          cdr.GetTOR(),
		Subject:      cdr.GetSubject(),
		Account:      cdr.GetAccount(),
		Destination:  cdr.GetDestination(),
		TimeStart:    t1,
		TimeEnd:      t1.Add(d),
		LoopIndex:    0,
		CallDuration: d,
	}
	if cdr.GetReqType() == utils.PSEUDOPREPAID {
		err = self.connector.Debit(cd, cc)
	} else {
		err = self.connector.GetCost(cd, cc)
	}
	if err != nil {
		self.logDb.LogError(cdr.GetCgrId(), engine.MEDIATOR_SOURCE, err.Error())
	} else {
		// If the mediator calculated a price it will write it to logdb
		self.logDb.LogCallCost(cdr.GetAccId(), engine.MEDIATOR_SOURCE, cc)
	}
	return cc, err
}



func (self *Mediator) MediateDBCDR(cdr utils.CDR) error {
	var qryCC *engine.CallCost
	cc := &engine.CallCost{Cost: -1}
	var errCost error
	if cdr.GetReqType() == utils.PREPAID || cdr.GetReqType() == utils.POSTPAID {
		// Should be previously calculated and stored in DB
		qryCC, errCost = self.getCostsFromDB(cdr)
	} else {
		qryCC, errCost = self.getCostsFromRater(cdr)
	}
	if errCost != nil || qryCC == nil {
		engine.Logger.Err(fmt.Sprintf("<Mediator> Could not calculate price for cgrid: <%s>, err: <%s>, cost: <%v>", cdr.GetCgrId(), errCost.Error(), qryCC))
	} else {
		cc = qryCC
		engine.Logger.Debug(fmt.Sprintf("<Mediator> Calculated for cgrid:%s, cost: %f", cdr.GetCgrId(), cc.ConnectFee+cc.Cost))
	}
	extraInfo := ""
	if errCost != nil {
		extraInfo = errCost.Error()
	}
	return self.cdrDb.SetRatedCdr(cdr, cc, extraInfo)
}
