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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Creates a new RatingProfile within a tariff plan
func (self *ApierV1) SetTPRatingProfile(attrs utils.TPRatingProfile, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "LoadId", "Tenant", "Category", "Direction", "Subject", "RatingPlanActivations"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.SetTPRatingProfiles([]*utils.TPRatingProfile{&attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPRatingProfileByLoadId struct {
	TPid   string // Tariff plan id
	LoadId string // RatingProfile id
}

// Queries specific RatingProfile on tariff plan
func (self *ApierV1) GetTPRatingProfilesByLoadId(attrs utils.TPRatingProfile, reply *[]*utils.TPRatingProfile) error {
	mndtryFlds := []string{"TPid", "LoadId"}
	if len(attrs.Subject) != 0 { // If Subject provided as filter, make all related fields mandatory
		mndtryFlds = append(mndtryFlds, "Tenant", "Category", "Direction", "Subject")
	}
	if missing := utils.MissingStructFields(&attrs, mndtryFlds); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if rps, err := self.StorDb.GetTPRatingProfiles(&attrs); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = rps
	}
	return nil
}

// Queries RatingProfile identities on specific tariff plan.
func (self *ApierV1) GetTPRatingProfileLoadIds(attrs utils.AttrTPRatingProfileIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPRateProfiles, utils.TPDistinctIds{"loadid"}, map[string]string{
		"tenant":    attrs.Tenant,
		"category":  attrs.Category,
		"direction": attrs.Direction,
		"subject":   attrs.Subject,
	}, new(utils.Paginator)); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = ids
	}
	return nil
}

type AttrGetTPRatingProfile struct {
	TPid            string // Tariff plan id
	RatingProfileId string // RatingProfile id
}

// Queries specific RatingProfile on tariff plan
func (self *ApierV1) GetTPRatingProfile(attrs AttrGetTPRatingProfile, reply *utils.TPRatingProfile) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RatingProfileId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tmpRpf := &utils.TPRatingProfile{TPid: attrs.TPid}
	if err := tmpRpf.SetRatingProfilesId(attrs.RatingProfileId); err != nil {
		return err
	}
	if rpfs, err := self.StorDb.GetTPRatingProfiles(tmpRpf); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *rpfs[0]
	}
	return nil
}

type AttrGetTPRatingProfileIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries RatingProfiles identities on specific tariff plan.
func (self *ApierV1) GetTPRatingProfileIds(attrs AttrGetTPRatingProfileIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPRateProfiles, utils.TPDistinctIds{"loadid", "direction", "tenant", "category", "subject"}, nil, &attrs.Paginator); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific RatingProfiles on Tariff plan
func (self *ApierV1) RemTPRatingProfile(attrs AttrGetTPRatingProfile, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RatingProfileId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tmpRpf := engine.TpRatingProfile{}
	if err := tmpRpf.SetRatingProfileId(attrs.RatingProfileId); err != nil {
		return err
	}
	if err := self.StorDb.RemTpData(utils.TBLTPRateProfiles, attrs.TPid, map[string]string{"loadid": tmpRpf.Loadid, "direction": tmpRpf.Direction, "tenant": tmpRpf.Tenant, "category": tmpRpf.Category, "subject": tmpRpf.Subject}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = utils.OK
	}
	return nil
}
