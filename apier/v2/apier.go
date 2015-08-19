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
	if len(attrs.TPid) == 0 {
		return utils.NewErrMandatoryIeMissing("TPid")
	}
	tpRpf := &utils.TPRatingProfile{TPid: attrs.TPid}
	tpRpf.SetRatingProfilesId(attrs.RatingProfileId)
	rpf := engine.APItoModelRatingProfile(tpRpf)
	dbReader := engine.NewTpReader(self.RatingDb, self.AccountDb, self.StorDb, attrs.TPid, self.Config.DefaultTimezone)
	if err := dbReader.LoadRatingProfilesFiltered(&rpf[0]); err != nil {
		return utils.NewErrServerError(err)
	}
	//Automatic cache of the newly inserted rating profile
	var ratingProfile []string
	if tpRpf.KeyId() != ":::" { // if has some filters
		ratingProfile = []string{utils.RATING_PROFILE_PREFIX + tpRpf.KeyId()}
	}
	if err := self.RatingDb.CachePrefixValues(map[string][]string{utils.RATING_PROFILE_PREFIX: ratingProfile}); err != nil {
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
	if len(attrs.TPid) == 0 {
		return utils.NewErrMandatoryIeMissing("TPid")
	}
	dbReader := engine.NewTpReader(self.RatingDb, self.AccountDb, self.StorDb, attrs.TPid, self.Config.DefaultTimezone)
	tpAa := &utils.TPAccountActions{TPid: attrs.TPid}
	tpAa.SetAccountActionsId(attrs.AccountActionsId)
	aa := engine.APItoModelAccountAction(tpAa)
	if _, err := engine.Guardian.Guard(func() (interface{}, error) {
		if err := dbReader.LoadAccountActionsFiltered(aa); err != nil {
			return 0, err
		}
		return 0, nil
	}, attrs.AccountActionsId); err != nil {
		return utils.NewErrServerError(err)
	}
	// ToDo: Get the action keys loaded by dbReader so we reload only these in cache
	// Need to do it before scheduler otherwise actions to run will be unknown
	if err := self.RatingDb.CachePrefixes(utils.DERIVED_CHARGERS_CSV, utils.ACTION_PREFIX, utils.SHARED_GROUP_PREFIX, utils.ACC_ALIAS_PREFIX); err != nil {
		return err
	}
	if self.Sched != nil {
		self.Sched.LoadActionPlans(self.RatingDb)
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
	if len(attrs.TPid) == 0 {
		return utils.NewErrMandatoryIeMissing("TPid")
	}
	tpDc := &utils.TPDerivedChargers{TPid: attrs.TPid}
	tpDc.SetDerivedChargersId(attrs.DerivedChargersId)
	dc := engine.APItoModelDerivedCharger(tpDc)
	dbReader := engine.NewTpReader(self.RatingDb, self.AccountDb, self.StorDb, attrs.TPid, self.Config.DefaultTimezone)
	if err := dbReader.LoadDerivedChargersFiltered(&dc[0], true); err != nil {
		return utils.NewErrServerError(err)
	}
	//Automatic cache of the newly inserted rating plan
	var dcsChanged []string
	if len(attrs.DerivedChargersId) != 0 {
		dcsChanged = []string{utils.DERIVEDCHARGERS_PREFIX + attrs.DerivedChargersId}
	}
	if err := self.RatingDb.CachePrefixValues(map[string][]string{utils.DERIVEDCHARGERS_PREFIX: dcsChanged}); err != nil {
		return err
	}
	*reply = v1.OK
	return nil
}
