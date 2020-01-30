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

// This file deals with tp_destrates_timing management over APIs

import (
	"github.com/cgrates/cgrates/utils"
)

// Creates a new DestinationRateTiming profile within a tariff plan
func (self *APIerSv1) SetTPRatingPlan(attrs utils.TPRatingPlan, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID", "RatingPlanBindings"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.SetTPRatingPlans([]*utils.TPRatingPlan{&attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPRatingPlan struct {
	TPid string // Tariff plan id
	ID   string // Rate id
	utils.Paginator
}

// Queries specific RatingPlan profile on tariff plan
func (self *APIerSv1) GetTPRatingPlan(attrs AttrGetTPRatingPlan, reply *utils.TPRatingPlan) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if rps, err := self.StorDb.GetTPRatingPlans(attrs.TPid, attrs.ID, &attrs.Paginator); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *rps[0]
	}
	return nil
}

type AttrGetTPRatingPlanIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// Queries RatingPlan identities on specific tariff plan.
func (self *APIerSv1) GetTPRatingPlanIds(attrs AttrGetTPRatingPlanIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPRatingPlans,
		utils.TPDistinctIds{"tag"}, nil, &attrs.PaginatorWithSearch); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific RatingPlan on Tariff plan
func (self *APIerSv1) RemoveTPRatingPlan(attrs AttrGetTPRatingPlan, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBLTPRatingPlans, attrs.TPid, map[string]string{"tag": attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = utils.OK
	}
	return nil
}
