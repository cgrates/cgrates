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
	"strconv"
	"strings"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

type APIerSv2 struct {
	v1.APIerSv1
}

// Call implements rpcclient.ClientConnector interface for internal RPC
func (apiv2 *APIerSv2) Call(serviceMethod string,
	args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(apiv2, serviceMethod, args, reply)
}

type AttrLoadRatingProfile struct {
	TPid            string
	RatingProfileID string
}

// Process dependencies and load a specific rating profile from storDb into dataDb.
func (apiv2 *APIerSv2) LoadRatingProfile(attrs AttrLoadRatingProfile, reply *string) error {
	if len(attrs.TPid) == 0 {
		return utils.NewErrMandatoryIeMissing("TPid")
	}
	tpRpf := &utils.TPRatingProfile{TPid: attrs.TPid}
	dbReader, err := engine.NewTpReader(apiv2.DataManager.DataDB(), apiv2.StorDb,
		attrs.TPid, apiv2.Config.GeneralCfg().DefaultTimezone,
		apiv2.Config.ApierCfg().CachesConns, apiv2.Config.ApierCfg().SchedulerConns)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if err := dbReader.LoadRatingProfilesFiltered(tpRpf); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrLoadAccountActions struct {
	TPid             string
	AccountActionsId string
}

// Process dependencies and load a specific AccountActions profile from storDb into dataDb.
func (apiv2 *APIerSv2) LoadAccountActions(attrs AttrLoadAccountActions, reply *string) error {
	if len(attrs.TPid) == 0 {
		return utils.NewErrMandatoryIeMissing("TPid")
	}
	dbReader, err := engine.NewTpReader(apiv2.DataManager.DataDB(), apiv2.StorDb,
		attrs.TPid, apiv2.Config.GeneralCfg().DefaultTimezone,
		apiv2.Config.ApierCfg().CachesConns, apiv2.Config.ApierCfg().SchedulerConns)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	tpAa := &utils.TPAccountActions{TPid: attrs.TPid}
	tpAa.SetAccountActionsId(attrs.AccountActionsId)
	if _, err := guardian.Guardian.Guard(func() (interface{}, error) {
		return 0, dbReader.LoadAccountActionsFiltered(tpAa)
	}, config.CgrConfig().GeneralCfg().LockingTimeout, attrs.AccountActionsId); err != nil {
		return utils.NewErrServerError(err)
	}
	sched := apiv2.SchedulerService.GetScheduler()
	if sched != nil {
		sched.Reload()
	}
	*reply = utils.OK
	return nil
}

func (apiv2 *APIerSv2) LoadTariffPlanFromFolder(attrs utils.AttrLoadTpFromFolder, reply *utils.LoadInstance) error {
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
	loader, err := engine.NewTpReader(apiv2.DataManager.DataDB(),
		engine.NewFileCSVStorage(utils.CSV_SEP, attrs.FolderPath, attrs.Recursive), "", apiv2.Config.GeneralCfg().DefaultTimezone,
		apiv2.Config.ApierCfg().CachesConns, apiv2.Config.ApierCfg().SchedulerConns)
	if err != nil {
		return utils.NewErrServerError(err)
	}
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

	if err := loader.WriteToDatabase(false, false); err != nil {
		return utils.NewErrServerError(err)
	}

	utils.Logger.Info("APIerSv2.LoadTariffPlanFromFolder, reloading cache.")
	//verify If Caching is present in arguments
	caching := config.CgrConfig().GeneralCfg().DefaultCaching
	if attrs.Caching != nil {
		caching = *attrs.Caching
	}
	if err := loader.ReloadCache(caching, true, attrs.ArgDispatcher); err != nil {
		return utils.NewErrServerError(err)
	}
	if len(apiv2.Config.ApierCfg().SchedulerConns) != 0 {
		utils.Logger.Info("APIerSv2.LoadTariffPlanFromFolder, reloading scheduler.")
		if err := loader.ReloadScheduler(true); err != nil {
			return utils.NewErrServerError(err)
		}
	}
	// release the reader with it's structures
	loader.Init()
	loadHistList, err := apiv2.DataManager.DataDB().GetLoadHistory(1, true, utils.NonTransactional)
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
func (apiv2 *APIerSv2) GetActions(attr AttrGetActions, reply *map[string]engine.Actions) error {
	var actionKeys []string
	var err error
	if len(attr.ActionIDs) == 0 {
		if actionKeys, err = apiv2.DataManager.DataDB().GetKeysForPrefix(utils.ACTION_PREFIX); err != nil {
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
		acts, err := apiv2.DataManager.GetActions(key, false, utils.NonTransactional)
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

// GetDestinations returns a list of destination based on the destinationIDs given
func (apiv2 *APIerSv2) GetDestinations(attr AttrGetDestinations, reply *[]*engine.Destination) (err error) {
	if len(attr.DestinationIDs) == 0 {
		// get all destination ids
		if attr.DestinationIDs, err = apiv2.DataManager.DataDB().GetKeysForPrefix(utils.DESTINATION_PREFIX); err != nil {
			return
		}
		for i, destID := range attr.DestinationIDs {
			attr.DestinationIDs[i] = destID[len(utils.DESTINATION_PREFIX):]
		}
	}
	dests := make([]*engine.Destination, len(attr.DestinationIDs))
	for i, destID := range attr.DestinationIDs {
		if dests[i], err = apiv2.DataManager.GetDestination(destID, false, utils.NonTransactional); err != nil {
			return
		}
	}
	*reply = dests
	return
}

func (apiv2 *APIerSv2) SetActions(attrs utils.AttrSetActions, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"ActionsId", "Actions"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	for _, action := range attrs.Actions {
		requiredFields := []string{"Identifier", "Weight"}
		if action.BalanceType != "" { // Add some inter-dependent parameters - if balanceType then we are not talking about simply calling actions
			requiredFields = append(requiredFields, "Units")
		}
		if missing := utils.MissingStructFields(action, requiredFields); len(missing) != 0 {
			return fmt.Errorf("%s:Action:%s:%v", utils.ErrMandatoryIeMissing.Error(), action.Identifier, missing)
		}
	}
	if !attrs.Overwrite {
		if exists, err := apiv2.DataManager.HasData(utils.ACTION_PREFIX, attrs.ActionsId, ""); err != nil {
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

		var blocker *bool
		if apiAct.BalanceBlocker != "" {
			if x, err := strconv.ParseBool(apiAct.BalanceBlocker); err == nil {
				blocker = &x
			} else {
				return err
			}
		}

		var disabled *bool
		if apiAct.BalanceDisabled != "" {
			if x, err := strconv.ParseBool(apiAct.BalanceDisabled); err == nil {
				disabled = &x
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
				DestinationIDs: utils.StringMapPointer(utils.ParseStringMap(apiAct.DestinationIds)),
				RatingSubject:  utils.StringPointer(apiAct.RatingSubject),
				SharedGroups:   utils.StringMapPointer(utils.ParseStringMap(apiAct.SharedGroups)),
				Categories:     utils.StringMapPointer(utils.ParseStringMap(apiAct.Categories)),
				TimingIDs:      utils.StringMapPointer(utils.ParseStringMap(apiAct.TimingTags)),
				Blocker:        blocker,
				Disabled:       disabled,
			},
		}
		storeActions[idx] = a
	}
	if err := apiv2.DataManager.SetActions(attrs.ActionsId, storeActions, utils.NonTransactional); err != nil {
		return utils.NewErrServerError(err)
	}
	//generate a loadID for CacheActions and store it in database
	if err := apiv2.DataManager.SetLoadIDs(map[string]int64{utils.CacheActions: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}
