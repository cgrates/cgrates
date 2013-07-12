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
	"github.com/cgrates/cgrates/utils"
)

// Creates a new RateProfile within a tariff plan
func (self *Apier) SetTPRateProfile(attrs utils.TPRateProfile, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RateProfileId", "Tenant", "TOR", "Direction", "Subject", "RatingActivations"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if exists, err := self.StorDb.ExistsTPRateProfile(attrs.TPid, attrs.RateProfileId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if exists {
		return errors.New(utils.ERR_DUPLICATE)
	}
	if err := self.StorDb.SetTPRateProfile(&attrs); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = "OK"
	return nil
}

type AttrGetTPRateProfile struct {
	TPid             string // Tariff plan id
	RateProfileId    string // RateProfile id
}

// Queries specific RateProfile on tariff plan
func (self *Apier) GetTPRateProfile(attrs AttrGetTPRateProfile, reply *utils.TPRateProfile) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RateProfileId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if dr, err := self.StorDb.GetTPRateProfile(attrs.TPid, attrs.RateProfileId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if dr == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = *dr
	}
	return nil
}

// Queries RateProfile identities on specific tariff plan.
func (self *Apier) GetTPRateProfileIds(attrs utils.AttrTPRateProfileIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if ids, err := self.StorDb.GetTPRateProfileIds(&attrs); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if ids == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = ids
	}
	return nil
}
