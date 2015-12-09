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
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

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
	dbReader := engine.NewTpReader(self.RatingDb, self.AccountDb, self.StorDb, attrs.TPid, self.Config.DefaultTimezone, self.Config.LoadHistorySize)
	if err := dbReader.LoadRatingProfilesFiltered(&rpf[0]); err != nil {
		return utils.NewErrServerError(err)
	}
	//Automatic cache of the newly inserted rating profile
	var ratingProfile []string
	if tpRpf.KeyId() != ":::" { // if has some filters
		ratingProfile = []string{utils.RATING_PROFILE_PREFIX + tpRpf.KeyId()}
	}
	if err := self.RatingDb.CacheRatingPrefixValues(map[string][]string{utils.RATING_PROFILE_PREFIX: ratingProfile}); err != nil {
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
	dbReader := engine.NewTpReader(self.RatingDb, self.AccountDb, self.StorDb, attrs.TPid, self.Config.DefaultTimezone, self.Config.LoadHistorySize)
	tpAa := &utils.TPAccountActions{TPid: attrs.TPid}
	tpAa.SetAccountActionsId(attrs.AccountActionsId)
	aa := engine.APItoModelAccountAction(tpAa)
	if _, err := engine.Guardian.Guard(func() (interface{}, error) {
		if err := dbReader.LoadAccountActionsFiltered(aa); err != nil {
			return 0, err
		}
		return 0, nil
	}, 0, attrs.AccountActionsId); err != nil {
		return utils.NewErrServerError(err)
	}
	// ToDo: Get the action keys loaded by dbReader so we reload only these in cache
	// Need to do it before scheduler otherwise actions to run will be unknown
	if err := self.RatingDb.CacheRatingPrefixes(utils.DERIVEDCHARGERS_PREFIX, utils.ACTION_PREFIX, utils.SHARED_GROUP_PREFIX); err != nil {
		return err
	}
	if self.Sched != nil {
		self.Sched.Reload(true)
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
	dbReader := engine.NewTpReader(self.RatingDb, self.AccountDb, self.StorDb, attrs.TPid, self.Config.DefaultTimezone, self.Config.LoadHistorySize)
	if err := dbReader.LoadDerivedChargersFiltered(&dc[0], true); err != nil {
		return utils.NewErrServerError(err)
	}
	//Automatic cache of the newly inserted rating plan
	var dcsChanged []string
	if len(attrs.DerivedChargersId) != 0 {
		dcsChanged = []string{utils.DERIVEDCHARGERS_PREFIX + attrs.DerivedChargersId}
	}
	if err := self.RatingDb.CacheRatingPrefixValues(map[string][]string{utils.DERIVEDCHARGERS_PREFIX: dcsChanged}); err != nil {
		return err
	}
	*reply = v1.OK
	return nil
}

func (self *ApierV2) LoadTariffPlanFromFolder(attrs utils.AttrLoadTpFromFolder, reply *engine.LoadInstance) error {
	if len(attrs.FolderPath) == 0 {
		return fmt.Errorf("%s:%s", utils.ErrMandatoryIeMissing.Error(), "FolderPath")
	}
	if fi, err := os.Stat(attrs.FolderPath); err != nil {
		if strings.HasSuffix(err.Error(), "no such file or directory") {
			return utils.ErrInvalidPath
		}
		return utils.NewErrServerError(err)
	} else if !fi.IsDir() {
		return utils.ErrInvalidPath
	}
	loader := engine.NewTpReader(self.RatingDb, self.AccountDb, engine.NewFileCSVStorage(utils.CSV_SEP,
		path.Join(attrs.FolderPath, utils.DESTINATIONS_CSV),
		path.Join(attrs.FolderPath, utils.TIMINGS_CSV),
		path.Join(attrs.FolderPath, utils.RATES_CSV),
		path.Join(attrs.FolderPath, utils.DESTINATION_RATES_CSV),
		path.Join(attrs.FolderPath, utils.RATING_PLANS_CSV),
		path.Join(attrs.FolderPath, utils.RATING_PROFILES_CSV),
		path.Join(attrs.FolderPath, utils.SHARED_GROUPS_CSV),
		path.Join(attrs.FolderPath, utils.LCRS_CSV),
		path.Join(attrs.FolderPath, utils.ACTIONS_CSV),
		path.Join(attrs.FolderPath, utils.ACTION_PLANS_CSV),
		path.Join(attrs.FolderPath, utils.ACTION_TRIGGERS_CSV),
		path.Join(attrs.FolderPath, utils.ACCOUNT_ACTIONS_CSV),
		path.Join(attrs.FolderPath, utils.DERIVED_CHARGERS_CSV),
		path.Join(attrs.FolderPath, utils.CDR_STATS_CSV),
		path.Join(attrs.FolderPath, utils.USERS_CSV),
		path.Join(attrs.FolderPath, utils.ALIASES_CSV),
	), "", self.Config.DefaultTimezone, self.Config.LoadHistorySize)
	if err := loader.LoadAll(); err != nil {
		return utils.NewErrServerError(err)
	}
	if attrs.DryRun {
		*reply = engine.LoadInstance{LoadId: utils.DRYRUN}
		return nil // Mission complete, no errors
	}

	if attrs.Validate {
		if !loader.IsValid() {
			return errors.New("invalid data")
		}
	}

	if err := loader.WriteToDatabase(attrs.FlushDb, false); err != nil {
		return utils.NewErrServerError(err)
	}
	// Make sure the items are in the cache
	dstIds, _ := loader.GetLoadedIds(utils.DESTINATION_PREFIX)
	dstKeys := make([]string, len(dstIds))
	for idx, dId := range dstIds {
		dstKeys[idx] = utils.DESTINATION_PREFIX + dId // Cache expects them as redis keys
	}
	rplIds, _ := loader.GetLoadedIds(utils.RATING_PLAN_PREFIX)
	rpKeys := make([]string, len(rplIds))
	for idx, rpId := range rplIds {
		rpKeys[idx] = utils.RATING_PLAN_PREFIX + rpId
	}
	rpfIds, _ := loader.GetLoadedIds(utils.RATING_PROFILE_PREFIX)
	rpfKeys := make([]string, len(rpfIds))
	for idx, rpfId := range rpfIds {
		rpfKeys[idx] = utils.RATING_PROFILE_PREFIX + rpfId
	}
	actIds, _ := loader.GetLoadedIds(utils.ACTION_PREFIX)
	actKeys := make([]string, len(actIds))
	for idx, actId := range actIds {
		actKeys[idx] = utils.ACTION_PREFIX + actId
	}
	aplIds, _ := loader.GetLoadedIds(utils.ACTION_PLAN_PREFIX)
	aplKeys := make([]string, len(aplIds))
	for idx, aplId := range aplIds {
		aplKeys[idx] = utils.ACTION_PLAN_PREFIX + aplId
	}
	shgIds, _ := loader.GetLoadedIds(utils.SHARED_GROUP_PREFIX)
	shgKeys := make([]string, len(shgIds))
	for idx, shgId := range shgIds {
		shgKeys[idx] = utils.SHARED_GROUP_PREFIX + shgId
	}
	aliases, _ := loader.GetLoadedIds(utils.ALIASES_PREFIX)
	alsKeys := make([]string, len(aliases))
	for idx, alias := range aliases {
		alsKeys[idx] = utils.ALIASES_PREFIX + alias
	}
	lcrIds, _ := loader.GetLoadedIds(utils.LCR_PREFIX)
	lcrKeys := make([]string, len(lcrIds))
	for idx, lcrId := range lcrIds {
		lcrKeys[idx] = utils.LCR_PREFIX + lcrId
	}
	dcs, _ := loader.GetLoadedIds(utils.DERIVEDCHARGERS_PREFIX)
	dcsKeys := make([]string, len(dcs))
	for idx, dc := range dcs {
		dcsKeys[idx] = utils.DERIVEDCHARGERS_PREFIX + dc
	}
	aps, _ := loader.GetLoadedIds(utils.ACTION_PLAN_PREFIX)
	utils.Logger.Info("ApierV1.LoadTariffPlanFromFolder, reloading cache.")

	cstKeys, _ := loader.GetLoadedIds(utils.CDR_STATS_PREFIX)
	userKeys, _ := loader.GetLoadedIds(utils.USERS_PREFIX)
	li := loader.GetLoadInstance()

	// release the tp data
	loader.Init()

	if err := self.RatingDb.CacheRatingPrefixValues(map[string][]string{
		utils.DESTINATION_PREFIX:     dstKeys,
		utils.RATING_PLAN_PREFIX:     rpKeys,
		utils.RATING_PROFILE_PREFIX:  rpfKeys,
		utils.LCR_PREFIX:             lcrKeys,
		utils.DERIVEDCHARGERS_PREFIX: dcsKeys,
		utils.ACTION_PREFIX:          actKeys,
		utils.ACTION_PLAN_PREFIX:     aplKeys,
		utils.SHARED_GROUP_PREFIX:    shgKeys,
	}); err != nil {
		return err
	}
	if err := self.AccountDb.CacheAccountingPrefixValues(map[string][]string{
		utils.ALIASES_PREFIX: alsKeys,
	}); err != nil {
		return err
	}
	if len(aps) != 0 && self.Sched != nil {
		utils.Logger.Info("ApierV1.LoadTariffPlanFromFolder, reloading scheduler.")
		self.Sched.Reload(true)
	}
	if len(cstKeys) != 0 && self.CdrStatsSrv != nil {
		var out int
		if err := self.CdrStatsSrv.Call("CDRStatsV1.ReloadQueues", cstKeys, &out); err != nil {
			return err
		}
	}

	if len(userKeys) != 0 && self.Users != nil {
		var r string
		if err := self.Users.Call("UsersV1.ReloadUsers", "", &r); err != nil {
			return err
		}
	}
	*reply = *li
	return nil
}
