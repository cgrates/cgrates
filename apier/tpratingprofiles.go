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

// This file deals with tp_rate_profiles management over APIs

import (
	"errors"
	"fmt"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Creates a new RatingProfile within a tariff plan
func (self *ApierV1) SetTPRatingProfile(attrs utils.TPRatingProfile, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "LoadId", "Tenant", "Category", "Direction", "Subject", "RatingPlanActivations"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if err := self.StorDb.SetTPRatingProfiles(attrs.TPid, map[string]*utils.TPRatingProfile{attrs.KeyId(): &attrs}); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
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
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if dr, err := self.StorDb.GetTpRatingProfiles(&attrs); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if dr == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		var rpfs []*utils.TPRatingProfile
		if len(attrs.Subject) != 0 {
			rpfs = []*utils.TPRatingProfile{dr[attrs.KeyId()]}
		} else {
			for _, rpfLst := range dr {
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
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if ids, err := self.StorDb.GetTPTableIds(attrs.TPid, utils.TBL_TP_RATE_PROFILES, utils.TPDistinctIds{"loadid"}, map[string]string{
		"tenant":    attrs.Tenant,
		"tor":       attrs.Category,
		"direction": attrs.Direction,
		"subject":   attrs.Subject,
	}, new(utils.TPPagination)); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if ids == nil {
		return errors.New(utils.ERR_NOT_FOUND)
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
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	tmpRpf := &utils.TPRatingProfile{TPid: attrs.TPid}
	if err := tmpRpf.SetRatingProfilesId(attrs.RatingProfileId); err != nil {
		return err
	}
	if rpfs, err := self.StorDb.GetTpRatingProfiles(tmpRpf); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if len(rpfs) == 0 {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		rpf := rpfs[tmpRpf.KeyId()]
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
	TPid         string // Tariff plan id
	Page         int
	ItemsPerPage int
	SearchTerm   string
}

// Queries RatingProfiles identities on specific tariff plan.
func (self *ApierV1) GetTPRatingProfileIds(attrs AttrGetTPRatingProfileIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if ids, err := self.StorDb.GetTPTableIds(attrs.TPid, utils.TBL_TP_RATE_PROFILES, utils.TPDistinctIds{"loadid", "direction", "tenant", "category", "subject"}, nil, &utils.TPPagination{Page: attrs.Page, ItemsPerPage: attrs.ItemsPerPage, SearchTerm: attrs.SearchTerm}); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if ids == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific RatingProfiles on Tariff plan
func (self *ApierV1) RemTPRatingProfile(attrs AttrGetTPRatingProfile, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RatingProfileId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	tmpRpf := engine.TpRatingProfile{}
	if err := tmpRpf.SetRatingProfileId(attrs.RatingProfileId); err != nil {
		return err
	}
	if err := self.StorDb.RemTPData(utils.TBL_TP_RATE_PROFILES, attrs.TPid, tmpRpf.Loadid, tmpRpf.Direction, tmpRpf.Tenant, tmpRpf.Category, tmpRpf.Subject); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else {
		*reply = "OK"
	}
	return nil
}
