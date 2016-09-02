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
	rpf := engine.APItoModelRatingProfile(&attrs)
	if err := self.StorDb.SetTpRatingProfiles(rpf); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = "OK"
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
		mndtryFlds = append(mndtryFlds, "Tenant", "TOR", "Direction", "Subject")
	}
	if missing := utils.MissingStructFields(&attrs, mndtryFlds); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	rpf := engine.APItoModelRatingProfile(&attrs)
	if dr, err := self.StorDb.GetTpRatingProfiles(&rpf[0]); err != nil {
		return utils.NewErrServerError(err)
	} else if dr == nil {
		return utils.ErrNotFound
	} else {
		rpfMap, err := engine.TpRatingProfiles(dr).GetRatingProfiles()
		if err != nil {
			return err
		}
		var rpfs []*utils.TPRatingProfile
		if len(attrs.Subject) != 0 {
			rpfs = []*utils.TPRatingProfile{rpfMap[attrs.KeyId()]}
		} else {
			for _, rpfLst := range rpfMap {
				rpfs = append(rpfs, rpfLst)
			}
		}
		*reply = rpfs
	}
	return nil
}

// Queries RatingProfile identities on specific tariff plan.
func (self *ApierV1) GetTPRatingProfileLoadIds(attrs utils.AttrTPRatingProfileIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBL_TP_RATE_PROFILES, utils.TPDistinctIds{"loadid"}, map[string]string{
		"tenant":    attrs.Tenant,
		"tor":       attrs.Category,
		"direction": attrs.Direction,
		"subject":   attrs.Subject,
	}, new(utils.Paginator)); err != nil {
		return utils.NewErrServerError(err)
	} else if ids == nil {
		return utils.ErrNotFound
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
	rpf := engine.APItoModelRatingProfile(tmpRpf)
	if rpfs, err := self.StorDb.GetTpRatingProfiles(&rpf[0]); err != nil {
		return utils.NewErrServerError(err)
	} else if len(rpfs) == 0 {
		return utils.ErrNotFound
	} else {
		rpfMap, err := engine.TpRatingProfiles(rpfs).GetRatingProfiles()
		if err != nil {
			return err
		}
		rpf := rpfMap[tmpRpf.KeyId()]
		tpdc := utils.TPRatingProfile{
			TPid: attrs.TPid,
			RatingPlanActivations: rpf.RatingPlanActivations,
		}
		if err := tpdc.SetRatingProfilesId(attrs.RatingProfileId); err != nil {
			return err
		}
		*reply = tpdc
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
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBL_TP_RATE_PROFILES, utils.TPDistinctIds{"loadid", "direction", "tenant", "category", "subject"}, nil, &attrs.Paginator); err != nil {
		return utils.NewErrServerError(err)
	} else if ids == nil {
		return utils.ErrNotFound
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
	if err := self.StorDb.RemTpData(utils.TBL_TP_RATE_PROFILES, attrs.TPid, map[string]string{"loadid": tmpRpf.Loadid, "direction": tmpRpf.Direction, "tenant": tmpRpf.Tenant, "category": tmpRpf.Category, "subject": tmpRpf.Subject}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = "OK"
	}
	return nil
}
