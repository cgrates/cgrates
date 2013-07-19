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
func (self *Apier) SetTPActionTimings(attrs utils.ApiTPActionTimings, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ActionsId", "ActionTimings"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	for _, at := range attrs.ActionTimings {
		requiredFields := []string{"ActionsId", "TimingId", "Weight"}
		if missing := utils.MissingStructFields(&at, requiredFields); len(missing) != 0 {
			return fmt.Errorf("%s:Action:%s:%v", utils.ERR_MANDATORY_IE_MISSING, at.ActionsId, missing)
		}
	}
	if exists, err := self.StorDb.ExistsTPActionTimings(attrs.TPid, attrs.ActionTimingsId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if exists {
		return errors.New(utils.ERR_DUPLICATE)
	}
	ats := make(map[string][]*utils.TPActionTimingsRow, 1) // Only one id will be stored in the map
	for _,at := range attrs.ActionTimings {
		ats[attrs.ActionTimingsId] = append( ats[attrs.ActionTimingsId], &utils.TPActionTimingsRow{at.ActionsId, at.TimingId, at.Weight} )
	}
	if err := self.StorDb.SetTPActionTimings(attrs.TPid, ats); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = "OK"
	return nil
}

type AttrGetTPActionTimings struct {
	TPid      string // Tariff plan id
	ActionTimingsId string // ActionTimings id
}

// Queries specific ActionTimings profile on tariff plan
func (self *Apier) GetTPActionTimings(attrs AttrGetTPActionTimings, reply *utils.ApiTPActionTimings) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ActionTimingsId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if ats, err := self.StorDb.GetTPActionTimings(attrs.TPid, attrs.ActionTimingsId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if len(ats) == 0 {
		return errors.New(utils.ERR_NOT_FOUND)
	} else { // Got the data we need, convert it from []TPActionTimingsRow into ApiTPActionTimings
		atRply := &utils.ApiTPActionTimings{ attrs.TPid, attrs.ActionTimingsId, make([]utils.ApiActionTiming, len(ats[attrs.ActionTimingsId])) }
		for idx,row := range ats[attrs.ActionTimingsId] {
			atRply.ActionTimings[idx] = utils.ApiActionTiming{ row.ActionsId, row.TimingId, row.Weight }
		}
	*reply = *atRply
	}
	return nil
}

type AttrGetTPActionTimingIds struct {
	TPid string // Tariff plan id
}

// Queries ActionTimings identities on specific tariff plan.
func (self *Apier) GetTPActionTimingIds(attrs AttrGetTPActionTimingIds, reply *[]string) error {
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


