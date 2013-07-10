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

// This file deals with tp_rates management over APIs

import (
	"errors"
	"fmt"
	"github.com/cgrates/cgrates/utils"
)



// Creates a new rate within a tariff plan
func (self *Apier) SetTPRate(attrs utils.TPRate, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RateId", "ConnectFee", "RateSlots"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if exists, err := self.StorDb.ExistsTPRate(attrs.TPid, attrs.RateId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if exists {
		return errors.New(utils.ERR_DUPLICATE)
	}
	if err := self.StorDb.SetTPRate(&attrs); err!=nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = "OK"
	return nil
}

/*
type AttrGetTPRate struct {
	TPid     string // Tariff plan id
	RateId string // Rate id
}

// Queries specific Rate on tariff plan
func (self *Apier) GetTPRate(attrs AttrGetTPRate, reply *ApierTPRate) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RateId"}); len(missing) != 0 { //Params missing
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
*/
