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

// SetTPRatingPlan creates a new DestinationRateTiming profile within a tariff plan
func (apierSv1 *APIerSv1) SetTPRatingPlan(attrs *utils.TPRatingPlan, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID, utils.RatingPlanBindings}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierSv1.StorDb.SetTPRatingPlans([]*utils.TPRatingPlan{attrs}); err != nil {
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

// GetTPRatingPlan queries specific RatingPlan profile on tariff plan
func (apierSv1 *APIerSv1) GetTPRatingPlan(attrs *AttrGetTPRatingPlan, reply *utils.TPRatingPlan) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	rps, err := apierSv1.StorDb.GetTPRatingPlans(attrs.TPid, attrs.ID, &attrs.Paginator)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *rps[0]
	return nil
}

type AttrGetTPRatingPlanIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// GetTPRatingPlanIds queries RatingPlan identities on specific tariff plan.
func (apierSv1 *APIerSv1) GetTPRatingPlanIds(attrs *AttrGetTPRatingPlanIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := apierSv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPRatingPlans,
		utils.TPDistinctIds{utils.TagCfg}, nil, &attrs.PaginatorWithSearch)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = ids
	return nil
}

// RemoveTPRatingPlan removes specific RatingPlan on Tariff plan
func (apierSv1 *APIerSv1) RemoveTPRatingPlan(attrs *AttrGetTPRatingPlan, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierSv1.StorDb.RemTpData(utils.TBLTPRatingPlans, attrs.TPid, map[string]string{utils.TagCfg: attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
