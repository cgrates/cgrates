/*
Real-time Charging System for Telecom & ISP environments
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

package v2

import (
	"fmt"
	"log"

	"github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type ApierV2 struct {
	v1.ApierV1
}

type AttrLoadRatingProfile struct {
	TPid            string
	RatingProfileId string
}

// Process dependencies and load a specific rating profile from storDb into dataDb.
func (self *ApierV2) LoadRatingProfile(attrs AttrLoadRatingProfile, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RatingProfileId"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}

	if attrs.RatingProfileId == utils.EMPTY {
		attrs.RatingProfileId = ""
	}
	tpRpf := &utils.TPRatingProfile{TPid: attrs.TPid}
	tpRpf.SetRatingProfilesId(attrs.RatingProfileId)
	dbReader := engine.NewDbReader(self.StorDb, self.RatingDb, self.AccountDb, attrs.TPid)
	if err := dbReader.LoadRatingProfileFiltered(tpRpf); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	//Automatic cache of the newly inserted rating profile
	didNotChange := []string{}
	var ratingProfile []string
	if tpRpf.KeyId() != ":::" { // if has some filters
		ratingProfile = []string{engine.RATING_PROFILE_PREFIX + tpRpf.KeyId()}
	}
	if err := self.RatingDb.CacheRating(didNotChange, didNotChange, ratingProfile, didNotChange, didNotChange); err != nil {
		return err
	}
	*reply = v1.OK
	return nil
}

type AttrLoadAccountActions struct {
	TPid             string
	AccountActionsId string
}

// Process dependencies and load a specific AccountActions profile from storDb into dataDb.
func (self *ApierV2) LoadAccountActions(attrs AttrLoadAccountActions, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "AccountActionsId"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	dbReader := engine.NewDbReader(self.StorDb, self.RatingDb, self.AccountDb, attrs.TPid)

	if attrs.AccountActionsId == utils.EMPTY {
		attrs.AccountActionsId = ""
	}

	tpAa := &utils.TPAccountActions{TPid: attrs.TPid}
	tpAa.SetAccountActionsId(attrs.AccountActionsId)

	if _, err := engine.AccLock.Guard(attrs.AccountActionsId, func() (float64, error) {
		if err := dbReader.LoadAccountActionsFiltered(tpAa); err != nil {
			return 0, err
		}
		return 0, nil
	}); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	// ToDo: Get the action keys loaded by dbReader so we reload only these in cache
	// Need to do it before scheduler otherwise actions to run will be unknown
	if err := self.AccountDb.CacheAccounting(nil, nil, nil, []string{}); err != nil {
		return err
	}
	if self.Sched != nil {
		self.Sched.LoadActionTimings(self.AccountDb)
		self.Sched.Restart()
	}
	*reply = v1.OK
	return nil
}

type AttrLoadDerivedChargers struct {
	TPid              string
	DerivedChargersId string
}

// Load derived chargers from storDb into dataDb.
func (self *ApierV2) LoadDerivedChargers(attrs AttrLoadDerivedChargers, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "DerivedChargersId"}); len(missing) != 0 {
		log.Printf("ATTR: %+v", attrs)
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if attrs.DerivedChargersId == utils.EMPTY {
		attrs.DerivedChargersId = ""
	}
	tpDc := &utils.TPDerivedChargers{TPid: attrs.TPid}
	tpDc.SetDerivedChargersId(attrs.DerivedChargersId)

	dbReader := engine.NewDbReader(self.StorDb, self.RatingDb, self.AccountDb, attrs.TPid)
	if err := dbReader.LoadDerivedChargersFiltered(tpDc); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	//Automatic cache of the newly inserted rating plan
	didNotChange := []string{}
	if err := self.AccountDb.CacheAccounting(didNotChange, didNotChange, didNotChange, nil); err != nil {
		return err
	}
	*reply = v1.OK
	return nil
}
