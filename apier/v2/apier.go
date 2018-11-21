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

package v2

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/guardian"
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
	dbReader := engine.NewTpReader(self.DataManager.DataDB(), self.StorDb,
		attrs.TPid, self.Config.GeneralCfg().DefaultTimezone)
	if err := dbReader.LoadRatingProfilesFiltered(tpRpf); err != nil {
		return utils.NewErrServerError(err)
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
	dbReader := engine.NewTpReader(self.DataManager.DataDB(), self.StorDb,
		attrs.TPid, self.Config.GeneralCfg().DefaultTimezone)
	tpAa := &utils.TPAccountActions{TPid: attrs.TPid}
	tpAa.SetAccountActionsId(attrs.AccountActionsId)
	if _, err := guardian.Guardian.Guard(func() (interface{}, error) {
		if err := dbReader.LoadAccountActionsFiltered(tpAa); err != nil {
			return 0, err
		}
		return 0, nil
	}, 0, attrs.AccountActionsId); err != nil {
		return utils.NewErrServerError(err)
	}
	sched := self.ServManager.GetScheduler()
	if sched != nil {
		sched.Reload()
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
	dbReader := engine.NewTpReader(self.DataManager.DataDB(), self.StorDb,
		attrs.TPid, self.Config.GeneralCfg().DefaultTimezone)
	if err := dbReader.LoadDerivedChargersFiltered(tpDc, true); err != nil {
		return utils.NewErrServerError(err)
	}
	if err := self.DataManager.CacheDataFromDB(utils.DERIVEDCHARGERS_PREFIX,
		[]string{attrs.DerivedChargersId}, true); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = v1.OK
	return nil
}

func (self *ApierV2) LoadTariffPlanFromFolder(attrs utils.AttrLoadTpFromFolder, reply *utils.LoadInstance) error {
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
	loader := engine.NewTpReader(self.DataManager.DataDB(),
		engine.NewFileCSVStorage(utils.CSV_SEP,
			path.Join(attrs.FolderPath, utils.DESTINATIONS_CSV),
			path.Join(attrs.FolderPath, utils.TIMINGS_CSV),
			path.Join(attrs.FolderPath, utils.RATES_CSV),
			path.Join(attrs.FolderPath, utils.DESTINATION_RATES_CSV),
			path.Join(attrs.FolderPath, utils.RATING_PLANS_CSV),
			path.Join(attrs.FolderPath, utils.RATING_PROFILES_CSV),
			path.Join(attrs.FolderPath, utils.SHARED_GROUPS_CSV),
			path.Join(attrs.FolderPath, utils.ACTIONS_CSV),
			path.Join(attrs.FolderPath, utils.ACTION_PLANS_CSV),
			path.Join(attrs.FolderPath, utils.ACTION_TRIGGERS_CSV),
			path.Join(attrs.FolderPath, utils.ACCOUNT_ACTIONS_CSV),
			path.Join(attrs.FolderPath, utils.DERIVED_CHARGERS_CSV),
			path.Join(attrs.FolderPath, utils.USERS_CSV),
			path.Join(attrs.FolderPath, utils.ALIASES_CSV),
			path.Join(attrs.FolderPath, utils.ResourcesCsv),
			path.Join(attrs.FolderPath, utils.StatsCsv),
			path.Join(attrs.FolderPath, utils.ThresholdsCsv),
			path.Join(attrs.FolderPath, utils.FiltersCsv),
			path.Join(attrs.FolderPath, utils.SuppliersCsv),
			path.Join(attrs.FolderPath, utils.AttributesCsv),
			path.Join(attrs.FolderPath, utils.ChargersCsv),
		), "", self.Config.GeneralCfg().DefaultTimezone)
	if err := loader.LoadAll(); err != nil {
		return utils.NewErrServerError(err)
	}
	if attrs.DryRun {
		*reply = utils.LoadInstance{RatingLoadID: utils.DRYRUN, AccountingLoadID: utils.DRYRUN}
		return nil // Mission complete, no errors
	}

	if attrs.Validate {
		if !loader.IsValid() {
			return errors.New("invalid data")
		}
	}

	if err := loader.WriteToDatabase(attrs.FlushDb, false, false); err != nil {
		return utils.NewErrServerError(err)
	}

	utils.Logger.Info("ApierV2.LoadTariffPlanFromFolder, reloading cache.")
	for _, prfx := range []string{
		utils.DESTINATION_PREFIX,
		utils.REVERSE_DESTINATION_PREFIX,
		utils.ACTION_PLAN_PREFIX,
		utils.AccountActionPlansPrefix,
		utils.DERIVEDCHARGERS_PREFIX,
		utils.ALIASES_PREFIX,
		utils.REVERSE_ALIASES_PREFIX} {
		loadedIDs, _ := loader.GetLoadedIds(prfx)
		if err := self.DataManager.CacheDataFromDB(prfx, loadedIDs, true); err != nil {
			return utils.NewErrServerError(err)
		}
	}
	aps, _ := loader.GetLoadedIds(utils.ACTION_PLAN_PREFIX)
	userKeys, _ := loader.GetLoadedIds(utils.USERS_PREFIX)

	// relase tp data
	loader.Init()

	if len(aps) != 0 {
		sched := self.ServManager.GetScheduler()
		if sched != nil {
			utils.Logger.Info("ApierV2.LoadTariffPlanFromFolder, reloading scheduler.")
			sched.Reload()
		}
	}
	if len(userKeys) != 0 && self.Users != nil {
		var r string
		if err := self.Users.Call("UsersV1.ReloadUsers", "", &r); err != nil {
			return err
		}
	}
	loadHistList, err := self.DataManager.DataDB().GetLoadHistory(1, true, utils.NonTransactional)
	if err != nil {
		return err
	}
	if len(loadHistList) > 0 {
		*reply = *loadHistList[0]
	}
	return nil
}

type AttrGetActions struct {
	ActionIDs []string
	Offset    int // Set the item offset
	Limit     int // Limit number of items retrieved
}

// Retrieves actions attached to specific ActionsId within cache
func (self *ApierV2) GetActions(attr AttrGetActions, reply *map[string]engine.Actions) error {
	var actionKeys []string
	var err error
	if len(attr.ActionIDs) == 0 {
		if actionKeys, err = self.DataManager.DataDB().GetKeysForPrefix(utils.ACTION_PREFIX); err != nil {
			return err
		}
	} else {
		for _, accID := range attr.ActionIDs {
			if len(accID) == 0 { // Source of error returned from redis (key not found)
				continue
			}
			actionKeys = append(actionKeys, utils.ACCOUNT_PREFIX+accID)
		}
	}
	if len(actionKeys) == 0 {
		return nil
	}
	if attr.Offset > len(actionKeys) {
		attr.Offset = len(actionKeys)
	}
	if attr.Offset < 0 {
		attr.Offset = 0
	}
	var limitedActions []string
	if attr.Limit != 0 {
		max := math.Min(float64(attr.Offset+attr.Limit), float64(len(actionKeys)))
		limitedActions = actionKeys[attr.Offset:int(max)]
	} else {
		limitedActions = actionKeys[attr.Offset:]
	}
	retActions := make(map[string]engine.Actions)
	for _, accKey := range limitedActions {
		key := accKey[len(utils.ACTION_PREFIX):]
		acts, err := self.DataManager.GetActions(key, false, utils.NonTransactional)
		if err != nil {
			return utils.NewErrServerError(err)
		}
		if len(acts) > 0 {
			retActions[key] = acts

		}
	}

	*reply = retActions
	return nil
}

type AttrGetDestinations struct {
	DestinationIDs []string
}

func (self *ApierV2) GetDestinations(attr AttrGetDestinations, reply *[]*engine.Destination) error {
	dests := make([]*engine.Destination, 0)
	if attr.DestinationIDs == nil {
		return utils.NewErrMandatoryIeMissing("DestIDs")
	}
	if len(attr.DestinationIDs) == 0 {
		// get all destination ids
		destIDs, err := self.DataManager.DataDB().GetKeysForPrefix(utils.DESTINATION_PREFIX)
		if err != nil {
			return err
		}
		for _, destID := range destIDs {
			attr.DestinationIDs = append(attr.DestinationIDs, destID[len(utils.DESTINATION_PREFIX):])
		}
	}
	for _, destID := range attr.DestinationIDs {
		dst, err := self.DataManager.DataDB().GetDestination(destID, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		dests = append(dests, dst)
	}

	*reply = dests
	return nil
}

func (self *ApierV2) SetActions(attrs utils.AttrSetActions, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"ActionsId", "Actions"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	for _, action := range attrs.Actions {
		requiredFields := []string{"Identifier", "Weight"}
		if action.BalanceType != "" { // Add some inter-dependent parameters - if balanceType then we are not talking about simply calling actions
			requiredFields = append(requiredFields, "Direction", "Units")
		}
		if missing := utils.MissingStructFields(action, requiredFields); len(missing) != 0 {
			return fmt.Errorf("%s:Action:%s:%v", utils.ErrMandatoryIeMissing.Error(), action.Identifier, missing)
		}
	}
	if !attrs.Overwrite {
		if exists, err := self.DataManager.HasData(utils.ACTION_PREFIX, attrs.ActionsId, ""); err != nil {
			return utils.NewErrServerError(err)
		} else if exists {
			return utils.ErrExists
		}
	}
	storeActions := make(engine.Actions, len(attrs.Actions))
	for idx, apiAct := range attrs.Actions {
		var vf *utils.ValueFormula
		if apiAct.Units != "" {
			if x, err := utils.ParseBalanceFilterValue(apiAct.BalanceType, apiAct.Units); err == nil {
				vf = x
			} else {
				return err
			}
		}

		var weight *float64
		if apiAct.BalanceWeight != "" {
			if x, err := strconv.ParseFloat(apiAct.BalanceWeight, 64); err == nil {
				weight = &x
			} else {
				return err
			}
		}

		a := &engine.Action{
			Id:               attrs.ActionsId,
			ActionType:       apiAct.Identifier,
			Weight:           apiAct.Weight,
			ExpirationString: apiAct.ExpiryTime,
			ExtraParameters:  apiAct.ExtraParameters,
			Filter:           apiAct.Filter,
			Balance: &engine.BalanceFilter{ // TODO: update this part
				Uuid:           utils.StringPointer(apiAct.BalanceUuid),
				ID:             utils.StringPointer(apiAct.BalanceId),
				Type:           utils.StringPointer(apiAct.BalanceType),
				Value:          vf,
				Weight:         weight,
				Directions:     utils.StringMapPointer(utils.ParseStringMap(apiAct.Directions)),
				DestinationIDs: utils.StringMapPointer(utils.ParseStringMap(apiAct.DestinationIds)),
				RatingSubject:  utils.StringPointer(apiAct.RatingSubject),
				SharedGroups:   utils.StringMapPointer(utils.ParseStringMap(apiAct.SharedGroups)),
			},
		}
		storeActions[idx] = a
	}
	if err := self.DataManager.SetActions(attrs.ActionsId, storeActions, utils.NonTransactional); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
