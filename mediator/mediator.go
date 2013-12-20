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
	"bufio"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

func NewMediator(connector engine.Connector, logDb engine.LogStorage, cdrDb engine.CdrStorage, cfg *config.CGRConfig) (m *Mediator, err error) {
	m = &Mediator{
		connector: connector,
		logDb:     logDb,
		cdrDb:     cdrDb,
		cgrCfg:    cfg,
	}
	m.fieldNames = make(map[string][]string)
	m.fieldIdxs = make(map[string][]int)
	// Load config fields
	if errLoad := m.loadConfig(); errLoad != nil {
		return nil, errLoad
	}
	return m, nil
}

type Mediator struct {
	connector           engine.Connector
	logDb               engine.LogStorage
	cdrDb               engine.CdrStorage
	cgrCfg              *config.CGRConfig
	cdrInDir, cdrOutDir string
	accIdField          string
	accIdIdx            int // Populated only for csv files where we have no names but indexes for the fields
	fieldNames          map[string][]string
	fieldIdxs           map[string][]int // Populated only for csv files where we have no names but indexes for the fields
}

/*
Responsible for loading configuration fields out of CGRConfig instance, doing the necessary pre-checks.
@param fieldKeys: stores the keys which will be referenced inside fieldNames/fieldIdxs
@param cfgVals: keep ordered references to configuration fields from CGRConfig instance.
fieldKeys and cfgVals are directly related through index.
Method logic:
 * Make sure the field used as reference in mediation process loop is not empty.
 * All other fields should match the length of reference field.
 * Accounting id field should not be empty.
 * If we run mediation on csv file:
  * Make sure cdrInDir and cdrOutDir are valid paths.
  * Populate accIdIdx by converting accIdField into integer.
  * Populate fieldIdxs by converting fieldNames into integers
*/
func (self *Mediator) loadConfig() error {
	fieldKeys := []string{"subject", "reqtype", "direction", "tenant", "tor", "account", "destination", "time_start", "duration"}
	cfgVals := [][]string{self.cgrCfg.MediatorSubjectFields, self.cgrCfg.MediatorReqTypeFields, self.cgrCfg.MediatorDirectionFields,
		self.cgrCfg.MediatorTenantFields, self.cgrCfg.MediatorTORFields, self.cgrCfg.MediatorAccountFields, self.cgrCfg.MediatorDestFields,
		self.cgrCfg.MediatorTimeAnswerFields, self.cgrCfg.MediatorDurationFields}

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

	// AccIdField has no special requirements, should just exist
	if len(self.cgrCfg.MediatorAccIdField) == 0 {
		return errors.New("Undefined mediator accid field")
	}
	self.accIdField = self.cgrCfg.MediatorAccIdField

	var errConv error
	// Specific settings of CSV style CDRS
	if self.cgrCfg.MediatorCDRType == utils.FSCDR_FILE_CSV {
		// Check paths to be valid before adding as configuration
		if _, err := os.Stat(self.cgrCfg.MediatorCDRInDir); err != nil {
			return fmt.Errorf("The input path for mediator does not exist: %v", self.cgrCfg.MediatorCDRInDir)
		} else {
			self.cdrInDir = self.cgrCfg.MediatorCDRInDir
		}
		if _, err := os.Stat(self.cgrCfg.MediatorCDROutDir); err != nil {
			return fmt.Errorf("The output path for mediator does not exist: %v", self.cgrCfg.MediatorCDROutDir)
		} else {
			self.cdrOutDir = self.cgrCfg.MediatorCDROutDir
		}
		if self.accIdIdx, errConv = strconv.Atoi(self.cgrCfg.MediatorAccIdField); errConv != nil {
			return errors.New("AccIdIndex must be integer.")
		}
	}

	// Load here field names and convert to integers in case of unamed cdrs like CSV
	for idx, key := range fieldKeys {
		self.fieldNames[key] = cfgVals[idx]
		if self.cgrCfg.MediatorCDRType == utils.FSCDR_FILE_CSV { // Special case when field names represent indexes of their location in file
			self.fieldIdxs[key] = make([]int, len(cfgVals[idx]))
			for iStr, cfgStr := range cfgVals[idx] {
				if self.fieldIdxs[key][iStr], errConv = strconv.Atoi(cfgStr); errConv != nil {
					return fmt.Errorf("All mediator index members (%s) must be ints", key)
				}
			}
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
