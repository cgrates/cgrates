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

type MediatorV1 struct {
	Medi *engine.Mediator
}

// Remotely start mediation with specific runid, runs asynchronously, it's status will be displayed in syslog
func (self *MediatorV1) RateCdrs(attrs utils.AttrRateCdrs, reply *string) error {
	if self.Medi == nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, "MediatorNotRunning")
	}
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
	//RateCdrs(cgrIds, runIds, tors, cdrHosts, cdrSources, reqTypes, directions, tenants, categories, accounts, subjects, destPrefixes []string,
	//orderIdStart, orderIdEnd int64, timeStart, timeEnd time.Time, rerateErrors, rerateRated bool)
	if err := self.Medi.RateCdrs(attrs.CgrIds, attrs.MediationRunIds, attrs.TORs, attrs.CdrHosts, attrs.CdrSources, attrs.ReqTypes, attrs.Directions,
		attrs.Tenants, attrs.Categories, attrs.Accounts, attrs.Subjects, attrs.DestinationPrefixes, attrs.RatedAccounts, attrs.RatedSubjects,
		attrs.OrderIdStart, attrs.OrderIdEnd, tStart, tEnd, attrs.RerateErrors, attrs.RerateRated); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = utils.OK
	return nil
}
