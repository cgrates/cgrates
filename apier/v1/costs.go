/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

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

package v1

import (
	"fmt"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"time"
)

type AttrGetDataCost struct {
	Direction                string
	Category                 string
	Tenant, Account, Subject string
	StartTime                time.Time
	Usage                    int64 // the call duration so far (till TimeEnd)
}

func (apier *ApierV1) GetDataCost(attrs AttrGetDataCost, reply *engine.DataCost) error {
	usageAsDuration := time.Duration(attrs.Usage) * time.Second // Convert to seconds to match the loaded rates
	cd := engine.CallDescriptor{
		Direction:     attrs.Direction,
		Category:      attrs.Category,
		Tenant:        attrs.Tenant,
		Account:       attrs.Account,
		Subject:       attrs.Subject,
		TimeStart:     attrs.StartTime,
		TimeEnd:       attrs.StartTime.Add(usageAsDuration),
		DurationIndex: usageAsDuration,
		TOR:           utils.DATA,
	}
	var cc engine.CallCost
	if err := apier.Responder.GetCost(cd, &cc); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	if dc, err := cc.ToDataCost(); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if dc != nil {
		*reply = *dc
	}
	return nil
}
