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

// This file deals with tp_destrates_timing management over APIs

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Creates a new DestinationRateTiming profile within a tariff plan
func (self *ApierV1) SetTPRatingPlan(attrs utils.TPRatingPlan, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RatingPlanId", "RatingPlanBindings"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	rp := engine.APItoModelRatingPlan(&attrs)
	if err := self.StorDb.SetTpRatingPlans(rp); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = "OK"
	return nil
}

type AttrGetTPRatingPlan struct {
	TPid         string // Tariff plan id
	RatingPlanId string // Rate id
	utils.Paginator
}

// Queries specific RatingPlan profile on tariff plan
func (self *ApierV1) GetTPRatingPlan(attrs AttrGetTPRatingPlan, reply *utils.TPRatingPlan) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RatingPlanId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if rps, err := self.StorDb.GetTpRatingPlans(attrs.TPid, attrs.RatingPlanId, &attrs.Paginator); err != nil {
		return utils.NewErrServerError(err)
	} else if len(rps) == 0 {
		return utils.ErrNotFound
	} else {
		rpsMap, err := engine.TpRatingPlans(rps).GetRatingPlans()
		if err != nil {
			return err
		}
		*reply = utils.TPRatingPlan{TPid: attrs.TPid, RatingPlanId: attrs.RatingPlanId, RatingPlanBindings: rpsMap[attrs.RatingPlanId]}
	}
	return nil
}

type AttrGetTPRatingPlanIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries RatingPlan identities on specific tariff plan.
func (self *ApierV1) GetTPRatingPlanIds(attrs AttrGetTPRatingPlanIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBL_TP_RATING_PLANS, utils.TPDistinctIds{"tag"}, nil, &attrs.Paginator); err != nil {
		return utils.NewErrServerError(err)
	} else if ids == nil {
		return utils.ErrNotFound
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific RatingPlan on Tariff plan
func (self *ApierV1) RemTPRatingPlan(attrs AttrGetTPRatingPlan, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RatingPlanId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBL_TP_RATING_PLANS, attrs.TPid, map[string]string{"tag": attrs.RatingPlanId}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = "OK"
	}
	return nil
}
