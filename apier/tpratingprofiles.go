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
func (self *Apier) SetTPRatingProfile(attrs utils.TPRatingProfile, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RatingProfileId", "Tenant", "TOR", "Direction", "Subject", "RatingActivations"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if exists, err := self.StorDb.ExistsTPRatingProfile(attrs.TPid, attrs.RatingProfileId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if exists {
		return errors.New(utils.ERR_DUPLICATE)
	}
	rps := make([]*engine.RatingProfile, len(attrs.RatingActivations))
	for idx, ra := range attrs.RatingActivations {
		rps[idx] = &engine.RatingProfile{Tag: attrs.RatingProfileId,
			Tenant:               attrs.Tenant,
			TOR:                  attrs.TOR,
			Direction:            attrs.Direction,
			Subject:              attrs.Subject,
			ActivationTime:       ra.ActivationTime,
			DestRatesTimingTag:   ra.DestRateTimingId,
			RatesFallbackSubject: attrs.RatesFallbackSubject,
		}
	}
	if err := self.StorDb.SetTPRatingProfiles(attrs.TPid, map[string][]*engine.RatingProfile{attrs.RatingProfileId: rps}); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = "OK"
	return nil
}

type AttrGetTPRatingProfile struct {
	TPid            string // Tariff plan id
	RatingProfileId string // RatingProfile id
}

// Queries specific RatingProfile on tariff plan
func (self *Apier) GetTPRatingProfile(attrs AttrGetTPRatingProfile, reply *utils.TPRatingProfile) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RatingProfileId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if dr, err := self.StorDb.GetTPRatingProfile(attrs.TPid, attrs.RatingProfileId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if dr == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = *dr
	}
	return nil
}

// Queries RatingProfile identities on specific tariff plan.
func (self *Apier) GetTPRatingProfileIds(attrs utils.AttrTPRatingProfileIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if ids, err := self.StorDb.GetTPRatingProfileIds(&attrs); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if ids == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = ids
	}
	return nil
}
