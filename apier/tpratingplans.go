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

// This file deals with tp_destrates_timing management over APIs

import (
	"errors"
	"fmt"

	"github.com/cgrates/cgrates/utils"
)

// Creates a new DestinationRateTiming profile within a tariff plan
func (self *ApierV1) SetTPRatingPlan(attrs utils.TPRatingPlan, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RatingPlanId", "RatingPlanBindings"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if err := self.StorDb.SetTPRatingPlans(attrs.TPid, map[string][]*utils.TPRatingPlanBinding{attrs.RatingPlanId: attrs.RatingPlanBindings}); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = "OK"
	return nil
}

type AttrGetTPRatingPlan struct {
	TPid         string // Tariff plan id
	RatingPlanId string // Rate id
}

// Queries specific RatingPlan profile on tariff plan
func (self *ApierV1) GetTPRatingPlan(attrs AttrGetTPRatingPlan, reply *utils.TPRatingPlan) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RatingPlanId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if rps, err := self.StorDb.GetTpRatingPlans(attrs.TPid, attrs.RatingPlanId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if len(rps) == 0 {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = utils.TPRatingPlan{TPid: attrs.TPid, RatingPlanId: attrs.RatingPlanId, RatingPlanBindings: rps[attrs.RatingPlanId]}
	}
	return nil
}

type AttrGetTPRatingPlanIds struct {
	TPid string // Tariff plan id
}

// Queries RatingPlan identities on specific tariff plan.
func (self *ApierV1) GetTPRatingPlanIds(attrs AttrGetTPRatingPlanIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if ids, err := self.StorDb.GetTPTableIds(attrs.TPid, utils.TBL_TP_RATING_PLANS, "tag", nil); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if ids == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific RatingPlan on Tariff plan
func (self *ApierV1) RemTPRatingPlan(attrs AttrGetTPRatingPlan, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RatingPlanId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if err := self.StorDb.RemTPData(utils.TBL_TP_RATING_PLANS, attrs.TPid, attrs.RatingPlanId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else {
		*reply = "OK"
	}
	return nil
}
