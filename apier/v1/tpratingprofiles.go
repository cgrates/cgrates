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

// This file deals with tp_rate_profiles management over APIs

import (
	"github.com/cgrates/cgrates/utils"
)

// SetTPRatingProfile creates a new RatingProfile within a tariff plan
func (apierSv1 *APIerSv1) SetTPRatingProfile(attrs *utils.TPRatingProfile, reply *string) error {
	if missing := utils.MissingStructFields(attrs,
		[]string{"TPid", "LoadId", "Tenant", "Category", "Subject", "RatingPlanActivations"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierSv1.StorDb.SetTPRatingProfiles([]*utils.TPRatingProfile{attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// GetTPRatingProfilesByLoadID queries specific RatingProfile on tariff plan
func (apierSv1 *APIerSv1) GetTPRatingProfilesByLoadID(attrs *utils.TPRatingProfile, reply *[]*utils.TPRatingProfile) error {
	mndtryFlds := []string{"TPid", "LoadId"}
	if len(attrs.Subject) != 0 { // If Subject provided as filter, make all related fields mandatory
		mndtryFlds = append(mndtryFlds, "Tenant", "Category", "Subject")
	}
	if missing := utils.MissingStructFields(attrs, mndtryFlds); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	rps, err := apierSv1.StorDb.GetTPRatingProfiles(attrs)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = rps
	return nil
}

// GetTPRatingProfileLoadIds queries RatingProfile identities on specific tariff plan.
func (apierSv1 *APIerSv1) GetTPRatingProfileLoadIds(attrs *utils.AttrTPRatingProfileIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := apierSv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPRatingProfiles,
		utils.TPDistinctIds{"loadid"}, map[string]string{
			"tenant":   attrs.Tenant,
			"category": attrs.Category,
			"subject":  attrs.Subject,
		}, new(utils.PaginatorWithSearch))
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = ids
	return nil
}

// AttrGetTPRatingProfile arguments used by GetTPRatingProfile and RemoveTPRatingProfile
type AttrGetTPRatingProfile struct {
	TPid            string // Tariff plan id
	RatingProfileID string // RatingProfile id
}

// GetTPRatingProfile queries specific RatingProfile on tariff plan
func (apierSv1 *APIerSv1) GetTPRatingProfile(attrs *AttrGetTPRatingProfile, reply *utils.TPRatingProfile) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid", "RatingProfileID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tmpRpf := &utils.TPRatingProfile{TPid: attrs.TPid}
	if err := tmpRpf.SetRatingProfileID(attrs.RatingProfileID); err != nil {
		return err
	}
	rpfs, err := apierSv1.StorDb.GetTPRatingProfiles(tmpRpf)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *rpfs[0]
	return nil
}

// AttrGetTPRatingProfileIds arguments used by GetTPRatingProfileIds
type AttrGetTPRatingProfileIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// GetTPRatingProfileIds queries RatingProfiles identities on specific tariff plan.
func (apierSv1 *APIerSv1) GetTPRatingProfileIds(attrs *AttrGetTPRatingProfileIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := apierSv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPRatingProfiles,
		utils.TPDistinctIds{"loadid", "tenant", "category", "subject"},
		nil, &attrs.PaginatorWithSearch)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = ids
	return nil
}

// RemoveTPRatingProfile removes specific RatingProfiles on Tariff plan
func (apierSv1 *APIerSv1) RemoveTPRatingProfile(attrs *AttrGetTPRatingProfile, reply *string) (err error) {
	if missing := utils.MissingStructFields(attrs, []string{"TPid", "RatingProfileID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tmpRpf := new(utils.TPRatingProfile)
	if err = tmpRpf.SetRatingProfileID(attrs.RatingProfileID); err != nil {
		return
	}
	err = apierSv1.StorDb.RemTpData(utils.TBLTPRatingProfiles,
		attrs.TPid, map[string]string{
			"loadid":   tmpRpf.LoadId,
			"tenant":   tmpRpf.Tenant,
			"category": tmpRpf.Category,
			"subject":  tmpRpf.Subject,
		})
	if err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return
}
