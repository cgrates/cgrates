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
	"errors"
	"fmt"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type ApierTPTiming struct {
	TPid      string // Tariff plan id
	TimingId  string // Timing id
	Years     string // semicolon separated list of years this timing is valid on, *all supported
	Months    string // semicolon separated list of months this timing is valid on, *none and *all supported
	MonthDays string // semicolon separated list of month's days this timing is valid on, *none and *all supported
	WeekDays  string // semicolon separated list of week day names this timing is valid on *none and *all supported
	Time      string // String representing the time this timing starts on
}

// Creates a new timing within a tariff plan
func (self *ApierV1) SetTPTiming(attrs ApierTPTiming, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "TimingId", "Years", "Months", "MonthDays", "WeekDays", "Time"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	tm := engine.NewTiming(attrs.TimingId, attrs.Years, attrs.Months, attrs.MonthDays, attrs.WeekDays, attrs.Time)
	if err := self.StorDb.SetTPTiming(attrs.TPid, tm); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = "OK"
	return nil
}

type AttrGetTPTiming struct {
	TPid     string // Tariff plan id
	TimingId string // Timing id
}

// Queries specific Timing on Tariff plan
func (self *ApierV1) GetTPTiming(attrs AttrGetTPTiming, reply *ApierTPTiming) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "TimingId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if tm, err := self.StorDb.GetTPTiming(attrs.TPid, attrs.TimingId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if tm == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = ApierTPTiming{attrs.TPid, tm.Id, tm.Years.Serialize(";"),
			tm.Months.Serialize(";"), tm.MonthDays.Serialize(";"), tm.WeekDays.Serialize(";"), tm.StartTime}
	}
	return nil
}

type AttrGetTPTimingIds struct {
	TPid string // Tariff plan id
}

// Queries timing identities on specific tariff plan.
func (self *ApierV1) GetTPTimingIds(attrs AttrGetTPTimingIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if ids, err := self.StorDb.GetTPTimingIds(attrs.TPid); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if ids == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific Timing on Tariff plan
func (self *ApierV1) RemTPTiming(attrs AttrGetTPTiming, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "TimingId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if err := self.StorDb.RemTPData(utils.TBL_TP_TIMINGS, attrs.TPid, attrs.TimingId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else {
		*reply = "OK"
	}
	return nil
}
