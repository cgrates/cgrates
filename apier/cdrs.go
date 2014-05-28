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

package apier

import (
	"fmt"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"time"
)

type AttrGetCallCost struct {
	CgrId string // Unique id of the CDR
	RunId string // Run Id
}

// Retrieves the callCost out of CGR logDb
func (apier *ApierV1) GetCallCostLog(attrs AttrGetCallCost, reply *engine.CallCost) error {
	if missing := utils.MissingStructFields(&attrs, []string{"CgrId", "RunId"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if cc, err := apier.LogDb.GetCallCostLog(attrs.CgrId, "", attrs.RunId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if cc == nil {
		return fmt.Errorf("NOT_FOUND")
	} else {
		*reply = *cc
	}
	return nil
}

// Retrieves CDRs based on the filters
func (apier *ApierV1) GetCdrs(attrs utils.AttrGetCdrs, reply *[]*utils.CgrCdrOut) error {
	var tStart, tEnd time.Time
	var err error
	if len(attrs.TimeStart) != 0 {
		if tStart, err = utils.ParseTimeDetectLayout(attrs.TimeStart); err != nil {
			return err
		}
	}
	if len(attrs.TimeEnd) != 0 {
		if tEnd, err = utils.ParseTimeDetectLayout(attrs.TimeEnd); err != nil {
			return err
		}
	}
	if cdrs, err := apier.CdrDb.GetStoredCdrs(attrs.CgrIds, attrs.MediationRunId, attrs.TOR, attrs.CdrHost, attrs.CdrSource, attrs.ReqType, attrs.Direction,
		attrs.Tenant, attrs.Category, attrs.Account, attrs.Subject, attrs.DestinationPrefix,
		attrs.OrderIdStart, attrs.OrderIdEnd, tStart, tEnd, attrs.SkipErrors, attrs.SkipRated, false); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else {
		for _, cdr := range cdrs {
			*reply = append(*reply, cdr.AsCgrCdrOut())
		}
	}
	return nil
}
