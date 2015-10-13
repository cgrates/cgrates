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

package v1

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Creates a new timing within a tariff plan
func (self *ApierV1) SetTPTiming(attrs utils.ApierTPTiming, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "TimingId", "Years", "Months", "MonthDays", "WeekDays", "Time"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tm := engine.APItoModelTiming(&attrs)
	if err := self.StorDb.SetTpTimings([]engine.TpTiming{*tm}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = "OK"
	return nil
}

type AttrGetTPTiming struct {
	TPid     string // Tariff plan id
	TimingId string // Timing id
}

// Queries specific Timing on Tariff plan
func (self *ApierV1) GetTPTiming(attrs AttrGetTPTiming, reply *utils.ApierTPTiming) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "TimingId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if tms, err := self.StorDb.GetTpTimings(attrs.TPid, attrs.TimingId); err != nil {
		return utils.NewErrServerError(err)
	} else if len(tms) == 0 {
		return utils.ErrNotFound
	} else {
		tmMap, err := engine.TpTimings(tms).GetApierTimings()
		if err != nil {
			return err
		}
		*reply = *tmMap[attrs.TimingId]
	}
	return nil
}

type AttrGetTPTimingIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries timing identities on specific tariff plan.
func (self *ApierV1) GetTPTimingIds(attrs AttrGetTPTimingIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBL_TP_TIMINGS, utils.TPDistinctIds{"tag"}, nil, &attrs.Paginator); err != nil {
		return utils.NewErrServerError(err)
	} else if ids == nil {
		return utils.ErrNotFound
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific Timing on Tariff plan
func (self *ApierV1) RemTPTiming(attrs AttrGetTPTiming, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "TimingId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBL_TP_TIMINGS, attrs.TPid, map[string]string{"tag": attrs.TimingId}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = "OK"
	}
	return nil
}
