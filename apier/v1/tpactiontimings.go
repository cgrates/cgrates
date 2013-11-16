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
	"github.com/cgrates/cgrates/utils"
)

// Creates a new ActionTimings profile within a tariff plan
func (self *ApierV1) SetTPActionTimings(attrs utils.TPActionTimings, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ActionTimingsId", "ActionTimings"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	for _, at := range attrs.ActionTimings {
		requiredFields := []string{"ActionsId", "TimingId", "Weight"}
		if missing := utils.MissingStructFields(at, requiredFields); len(missing) != 0 {
			return fmt.Errorf("%s:Action:%s:%v", utils.ERR_MANDATORY_IE_MISSING, at.ActionsId, missing)
		}
	}
	if err := self.StorDb.SetTPActionTimings(attrs.TPid, map[string][]*utils.TPActionTiming{attrs.ActionTimingsId: attrs.ActionTimings}); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = "OK"
	return nil
}

type AttrGetTPActionTimings struct {
	TPid            string // Tariff plan id
	ActionTimingsId string // ActionTimings id
}

// Queries specific ActionTimings profile on tariff plan
func (self *ApierV1) GetTPActionTimings(attrs AttrGetTPActionTimings, reply *utils.TPActionTimings) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ActionTimingsId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if ats, err := self.StorDb.GetTPActionTimings(attrs.TPid, attrs.ActionTimingsId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if len(ats) == 0 {
		return errors.New(utils.ERR_NOT_FOUND)
	} else { // Got the data we need, convert it
		atRply := &utils.TPActionTimings{attrs.TPid, attrs.ActionTimingsId, ats[attrs.ActionTimingsId]}
		*reply = *atRply
	}
	return nil
}

type AttrGetTPActionTimingIds struct {
	TPid string // Tariff plan id
}

// Queries ActionTimings identities on specific tariff plan.
func (self *ApierV1) GetTPActionTimingIds(attrs AttrGetTPActionTimingIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if ids, err := self.StorDb.GetTPActionTimingIds(attrs.TPid); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if ids == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific ActionTimings on Tariff plan
func (self *ApierV1) RemTPActionTimings(attrs AttrGetTPActionTimings, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ActionTimingsId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if err := self.StorDb.RemTPData(utils.TBL_TP_ACTION_TIMINGS, attrs.TPid, attrs.ActionTimingsId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else {
		*reply = "OK"
	}
	return nil
}
