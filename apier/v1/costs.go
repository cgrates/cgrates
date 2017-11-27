/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type AttrGetCost struct {
	Tenant      string
	Category    string
	Subject     string
	AnswerTime  time.Time
	Destination string
	Usage       string
}

func (apier *ApierV1) GetCost(attrs AttrGetCost, ec *engine.EventCost) error {
	usage, err := utils.ParseDurationWithNanosecs(attrs.Usage)
	if err != nil {
		return err
	}
	cd := &engine.CallDescriptor{
		Direction:     utils.OUT,
		Category:      attrs.Category,
		Tenant:        attrs.Tenant,
		Subject:       attrs.Subject,
		Destination:   attrs.Destination,
		TimeStart:     attrs.AnswerTime,
		TimeEnd:       attrs.AnswerTime.Add(usage),
		DurationIndex: usage,
	}
	var cc engine.CallCost
	if err := apier.Responder.GetCost(cd, &cc); err != nil {
		return utils.NewErrServerError(err)
	}
	*ec = *engine.NewEventCostFromCallCost(&cc, "", "")
	ec.Compute()
	return nil
}

type AttrGetDataCost struct {
	Tenant     string
	Category   string
	Subject    string
	AnswerTime time.Time
	Usage      time.Duration // the call duration so far (till TimeEnd)
}

func (apier *ApierV1) GetDataCost(attrs AttrGetDataCost, reply *engine.DataCost) error {
	cd := &engine.CallDescriptor{
		Direction:     utils.OUT,
		Category:      attrs.Category,
		Tenant:        attrs.Tenant,
		Subject:       attrs.Subject,
		TimeStart:     attrs.AnswerTime,
		TimeEnd:       attrs.AnswerTime.Add(attrs.Usage),
		DurationIndex: attrs.Usage,
		TOR:           utils.DATA,
	}
	var cc engine.CallCost
	if err := apier.Responder.GetCost(cd, &cc); err != nil {
		return utils.NewErrServerError(err)
	}
	if dc, err := cc.ToDataCost(); err != nil {
		return utils.NewErrServerError(err)
	} else if dc != nil {
		*reply = *dc
	}
	return nil
}
