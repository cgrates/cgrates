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

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

// Returns a list of ActionTriggers on an account
func (apierSv1 *APIerSv1) GetAccountActionTriggers(attrs *utils.TenantAccount, reply *engine.ActionTriggers) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.AccountField}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := attrs.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if account, err := apierSv1.DataManager.GetAccount(utils.ConcatenatedKey(tnt, attrs.Account)); err != nil {
		return utils.NewErrServerError(err)
	} else {
		ats := account.ActionTriggers
		if ats == nil {
			ats = engine.ActionTriggers{}
		}
		*reply = ats
	}
	return nil
}

type AttrAddAccountActionTriggers struct {
	Tenant                 string
	Account                string
	ActionTriggerIDs       []string
	ActionTriggerOverwrite bool
	ActivationDate         string
	Executed               bool
}

func (apierSv1 *APIerSv1) AddAccountActionTriggers(attr *AttrAddAccountActionTriggers, reply *string) (err error) {
	if missing := utils.MissingStructFields(attr, []string{utils.AccountField}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := attr.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	var actTime time.Time
	if actTime, err = utils.ParseTimeDetectLayout(attr.ActivationDate,
		apierSv1.Config.GeneralCfg().DefaultTimezone); err != nil {
		return
	}
	accID := utils.ConcatenatedKey(tnt, attr.Account)
	var account *engine.Account
	_, err = guardian.Guardian.Guard(func() (interface{}, error) {
		if account, err = apierSv1.DataManager.GetAccount(accID); err != nil {
			return 0, err
		}
		if attr.ActionTriggerOverwrite {
			account.ActionTriggers = make(engine.ActionTriggers, 0)
		}
		for _, actionTriggerID := range attr.ActionTriggerIDs {
			atrs, err := apierSv1.DataManager.GetActionTriggers(actionTriggerID, false, utils.NonTransactional)
			if err != nil {
				return 0, err
			}
			for _, at := range atrs {
				var found bool
				for _, existingAt := range account.ActionTriggers {
					if existingAt.Equals(at) {
						found = true
						break
					}
				}
				at.ActivationDate = actTime
				at.Executed = attr.Executed
				if !found {
					account.ActionTriggers = append(account.ActionTriggers, at)
				}
			}
		}
		account.InitCounters()
		return 0, apierSv1.DataManager.SetAccount(account)
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.AccountPrefix+accID)
	if err != nil {
		return
	}
	*reply = utils.OK
	return
}

type AttrRemoveAccountActionTriggers struct {
	Tenant   string
	Account  string
	GroupID  string
	UniqueID string
}

func (apierSv1 *APIerSv1) RemoveAccountActionTriggers(attr *AttrRemoveAccountActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{utils.AccountField}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := attr.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	accID := utils.ConcatenatedKey(tnt, attr.Account)
	_, err := guardian.Guardian.Guard(func() (interface{}, error) {
		var account *engine.Account
		if acc, err := apierSv1.DataManager.GetAccount(accID); err == nil {
			account = acc
		} else {
			return 0, err
		}
		var newActionTriggers engine.ActionTriggers
		for _, at := range account.ActionTriggers {
			if (attr.UniqueID == "" || at.UniqueID == attr.UniqueID) &&
				(attr.GroupID == "" || at.ID == attr.GroupID) {
				// remove action trigger
				continue
			}
			newActionTriggers = append(newActionTriggers, at)
		}
		account.ActionTriggers = newActionTriggers
		account.InitCounters()
		return 0, apierSv1.DataManager.SetAccount(account)
	}, config.CgrConfig().GeneralCfg().LockingTimeout, accID)
	if err != nil {
		*reply = err.Error()
		return err
	}
	*reply = utils.OK
	return nil
}

type AttrResetAccountActionTriggers struct {
	Tenant   string
	Account  string
	GroupID  string
	UniqueID string
	Executed bool
}

func (apierSv1 *APIerSv1) ResetAccountActionTriggers(attr *AttrResetAccountActionTriggers, reply *string) error {
	fmt.Println("yay")
	if missing := utils.MissingStructFields(&attr, []string{utils.AccountField}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := attr.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	accID := utils.ConcatenatedKey(tnt, attr.Account)
	var account *engine.Account
	_, err := guardian.Guardian.Guard(func() (interface{}, error) {
		if acc, err := apierSv1.DataManager.GetAccount(accID); err == nil {
			account = acc
		} else {
			return 0, err
		}
		for _, at := range account.ActionTriggers {
			if (attr.UniqueID == "" || at.UniqueID == attr.UniqueID) &&
				(attr.GroupID == "" || at.ID == attr.GroupID) {
				// reset action trigger
				at.Executed = attr.Executed
			}

		}
		if attr.Executed == false {
			account.ExecuteActionTriggers(nil)
		}
		return 0, apierSv1.DataManager.SetAccount(account)
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.AccountPrefix+accID)
	if err != nil {
		*reply = err.Error()
		return err
	}
	*reply = utils.OK
	return nil
}

type AttrSetAccountActionTriggers struct {
	Tenant  string
	Account string
	AttrSetActionTrigger
}
type AttrSetActionTrigger struct {
	GroupID       string
	UniqueID      string
	ActionTrigger map[string]interface{}
}

// UpdateActionTrigger updates the ActionTrigger if is matching
func (attr *AttrSetActionTrigger) UpdateActionTrigger(at *engine.ActionTrigger, timezone string) (updated bool, err error) {
	if at == nil {
		return false, errors.New("Empty ActionTrigger")
	}
	if at.ID == utils.EmptyString { // New AT, update it's data
		if attr.GroupID == utils.EmptyString {
			return false, utils.NewErrMandatoryIeMissing(utils.GroupID)
		}
		if missing := utils.MissingMapFields(attr.ActionTrigger, []string{"ThresholdType", "ThresholdValue"}); len(missing) != 0 {
			return false, utils.NewErrMandatoryIeMissing(missing...)
		}
		at.ID = attr.GroupID
		if attr.UniqueID != utils.EmptyString {
			at.UniqueID = attr.UniqueID
		}
	}
	if attr.GroupID != utils.EmptyString && attr.GroupID != at.ID {
		return
	}
	if attr.UniqueID != utils.EmptyString && attr.UniqueID != at.UniqueID {
		return
	}
	// at matches
	updated = true
	if thr, has := attr.ActionTrigger[utils.ThresholdType]; has {
		at.ThresholdType = utils.IfaceAsString(thr)
	}
	if thr, has := attr.ActionTrigger[utils.ThresholdValue]; has {
		if at.ThresholdValue, err = utils.IfaceAsFloat64(thr); err != nil {
			return
		}
	}
	if rec, has := attr.ActionTrigger[utils.Recurrent]; has {
		if at.Recurrent, err = utils.IfaceAsBool(rec); err != nil {
			return
		}
	}
	if exec, has := attr.ActionTrigger[utils.Executed]; has {
		if at.Executed, err = utils.IfaceAsBool(exec); err != nil {
			return
		}
	}
	if minS, has := attr.ActionTrigger[utils.MinSleep]; has {
		if at.MinSleep, err = utils.IfaceAsDuration(minS); err != nil {
			return
		}
	}
	if exp, has := attr.ActionTrigger[utils.ExpirationDate]; has {
		if at.ExpirationDate, err = utils.IfaceAsTime(exp, timezone); err != nil {
			return
		}
	}
	if act, has := attr.ActionTrigger[utils.ActivationDate]; has {
		if at.ActivationDate, err = utils.IfaceAsTime(act, timezone); err != nil {
			return
		}
	}
	if at.Balance == nil {
		at.Balance = &engine.BalanceFilter{}
	}
	if bid, has := attr.ActionTrigger[utils.BalanceID]; has {
		at.Balance.ID = utils.StringPointer(utils.IfaceAsString(bid))
	}
	if btype, has := attr.ActionTrigger[utils.BalanceType]; has {
		at.Balance.Type = utils.StringPointer(utils.IfaceAsString(btype))
	}
	if bdest, has := attr.ActionTrigger[utils.BalanceDestinationIds]; has {
		var bdIds []string
		if bdIds, err = utils.IfaceAsSliceString(bdest); err != nil {
			return
		}
		at.Balance.DestinationIDs = utils.StringMapPointer(utils.NewStringMap(bdIds...))
	}
	if bweight, has := attr.ActionTrigger[utils.BalanceWeight]; has {
		var bw float64
		if bw, err = utils.IfaceAsFloat64(bweight); err != nil {
			return
		}
		at.Balance.Weight = utils.Float64Pointer(bw)
	}
	if exp, has := attr.ActionTrigger[utils.BalanceExpirationDate]; has {
		var balanceExpTime time.Time
		if balanceExpTime, err = utils.IfaceAsTime(exp, timezone); err != nil {
			return
		}
		at.Balance.ExpirationDate = utils.TimePointer(balanceExpTime)
	}
	if bTimeTag, has := attr.ActionTrigger[utils.BalanceTimingTags]; has {
		var timeTag []string
		if timeTag, err = utils.IfaceAsSliceString(bTimeTag); err != nil {
			return
		}
		at.Balance.TimingIDs = utils.StringMapPointer(utils.NewStringMap(timeTag...))
	}
	if brs, has := attr.ActionTrigger[utils.BalanceRatingSubject]; has {
		at.Balance.RatingSubject = utils.StringPointer(utils.IfaceAsString(brs))
	}
	if bcat, has := attr.ActionTrigger[utils.BalanceCategories]; has {
		var cat []string
		if cat, err = utils.IfaceAsSliceString(bcat); err != nil {
			return
		}
		at.Balance.Categories = utils.StringMapPointer(utils.NewStringMap(cat...))
	}
	if bsg, has := attr.ActionTrigger[utils.BalanceSharedGroups]; has {
		var shrgrps []string
		if shrgrps, err = utils.IfaceAsSliceString(bsg); err != nil {
			return
		}
		at.Balance.SharedGroups = utils.StringMapPointer(utils.NewStringMap(shrgrps...))
	}
	if bb, has := attr.ActionTrigger[utils.BalanceBlocker]; has {
		var bBlocker bool
		if bBlocker, err = utils.IfaceAsBool(bb); err != nil {
			return
		}
		at.Balance.Blocker = utils.BoolPointer(bBlocker)
	}
	if bd, has := attr.ActionTrigger[utils.BalanceDisabled]; has {
		var bDis bool
		if bDis, err = utils.IfaceAsBool(bd); err != nil {
			return
		}
		at.Balance.Disabled = utils.BoolPointer(bDis)
	}
	if minQ, has := attr.ActionTrigger[utils.MinQueuedItems]; has {
		var mQ int64
		if mQ, err = utils.IfaceAsInt64(minQ); err != nil {
			return
		}
		at.MinQueuedItems = int(mQ)
	}
	if accID, has := attr.ActionTrigger[utils.ActionsID]; has {
		at.ActionsID = utils.IfaceAsString(accID)
	}
	return
}

// SetAccountActionTriggers updates or creates if not present the ActionTrigger for an Account
func (apierSv1 *APIerSv1) SetAccountActionTriggers(attr *AttrSetAccountActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{utils.AccountField}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := attr.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	accID := utils.ConcatenatedKey(tnt, attr.Account)
	var account *engine.Account
	_, err := guardian.Guardian.Guard(func() (interface{}, error) {
		if acc, err := apierSv1.DataManager.GetAccount(accID); err == nil {
			account = acc
		} else {
			return 0, err
		}
		var foundOne bool
		for _, at := range account.ActionTriggers {
			if updated, err := attr.UpdateActionTrigger(at,
				apierSv1.Config.GeneralCfg().DefaultTimezone); err != nil {
				return 0, err
			} else if updated && !foundOne {
				foundOne = true
			}
		}
		if !foundOne { // Did not find one to update, create a new AT
			at := new(engine.ActionTrigger)
			if updated, err := attr.UpdateActionTrigger(at,
				apierSv1.Config.GeneralCfg().DefaultTimezone); err != nil {
				return 0, err
			} else if updated { // Adding a new AT
				account.ActionTriggers = append(account.ActionTriggers, at)
			}
		}
		account.ExecuteActionTriggers(nil)
		return 0, apierSv1.DataManager.SetAccount(account)
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.AccountPrefix+accID)
	if err != nil {
		*reply = err.Error()
		return err
	}
	*reply = utils.OK
	return nil
}

type AttrRemoveActionTrigger struct {
	GroupID  string
	UniqueID string
}

func (apierSv1 *APIerSv1) RemoveActionTrigger(attr AttrRemoveActionTrigger, reply *string) (err error) {
	if missing := utils.MissingStructFields(&attr, []string{"GroupID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attr.UniqueID == "" {
		err = apierSv1.DataManager.RemoveActionTriggers(attr.GroupID, utils.NonTransactional)
		if err != nil {
			return
		}
		*reply = utils.OK
		return
	}
	var atrs engine.ActionTriggers
	if atrs, err = apierSv1.DataManager.GetActionTriggers(attr.GroupID, false, utils.NonTransactional); err != nil {
		return
	}
	remainingAtrs := make(engine.ActionTriggers, 0, len(atrs))
	for _, atr := range atrs {
		if atr.UniqueID != attr.UniqueID {
			remainingAtrs = append(remainingAtrs, atr)
		}
	}
	// set the cleared list back
	if err = apierSv1.DataManager.SetActionTriggers(attr.GroupID, remainingAtrs, utils.NonTransactional); err != nil {
		return
	}
	// CacheReload
	if err = apierSv1.ConnMgr.Call(apierSv1.Config.ApierCfg().CachesConns, nil,
		utils.CacheSv1ReloadCache, utils.AttrReloadCacheWithOpts{
			ArgsCache: map[string][]string{utils.ActionTriggerIDs: {attr.GroupID}},
		}, reply); err != nil {
		return
	}
	// generate a loadID for CacheActionTriggers and store it in database
	if err = apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheActionTriggers: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return
}

// SetActionTrigger updates a ActionTrigger
func (apierSv1 *APIerSv1) SetActionTrigger(attr AttrSetActionTrigger, reply *string) (err error) {
	if missing := utils.MissingStructFields(&attr, []string{"GroupID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	atrs, _ := apierSv1.DataManager.GetActionTriggers(attr.GroupID, false, utils.NonTransactional)
	var newAtr *engine.ActionTrigger
	if attr.UniqueID != utils.EmptyString {
		//search for exiting one
		for _, atr := range atrs {
			if atr.UniqueID == attr.UniqueID {
				newAtr = atr
				break
			}
		}
	}

	if newAtr == nil {
		newAtr = &engine.ActionTrigger{}
		atrs = append(atrs, newAtr)
	}
	if attr.UniqueID == utils.EmptyString {
		attr.UniqueID = utils.GenUUID()
	}
	newAtr.ID = attr.GroupID
	newAtr.UniqueID = attr.UniqueID
	if _, err = attr.UpdateActionTrigger(newAtr,
		apierSv1.Config.GeneralCfg().DefaultTimezone); err != nil {
		return
	}

	if err = apierSv1.DataManager.SetActionTriggers(attr.GroupID, atrs, utils.NonTransactional); err != nil {
		return
	}
	// CacheReload
	if err = apierSv1.ConnMgr.Call(apierSv1.Config.ApierCfg().CachesConns, nil,
		utils.CacheSv1ReloadCache, utils.AttrReloadCacheWithOpts{
			ArgsCache: map[string][]string{utils.ActionTriggerIDs: {attr.GroupID}},
		}, reply); err != nil {
		return
	}
	// generate a loadID for CacheActionTriggers and store it in database
	if err = apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheActionTriggers: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return
}

type AttrGetActionTriggers struct {
	GroupIDs []string
}

func (apierSv1 *APIerSv1) GetActionTriggers(attr *AttrGetActionTriggers, atrs *engine.ActionTriggers) error {
	var allAttrs engine.ActionTriggers
	if len(attr.GroupIDs) > 0 {
		for _, key := range attr.GroupIDs {
			getAttrs, err := apierSv1.DataManager.GetActionTriggers(key, false, utils.NonTransactional)
			if err != nil {
				return err
			}
			allAttrs = append(allAttrs, getAttrs...)
		}

	} else {
		keys, err := apierSv1.DataManager.DataDB().GetKeysForPrefix(utils.ActionTriggerPrefix)
		if err != nil {
			return err
		}
		if len(keys) == 0 {
			return utils.ErrNotFound
		}
		for _, key := range keys {
			getAttrs, err := apierSv1.DataManager.GetActionTriggers(key[len(utils.ActionTriggerPrefix):], false, utils.NonTransactional)
			if err != nil {
				return err
			}
			allAttrs = append(allAttrs, getAttrs...)
		}
	}
	*atrs = allAttrs
	return nil
}

type AttrAddActionTrigger struct {
	ActionTriggersId      string
	Tenant                string
	Account               string
	ThresholdType         string
	ThresholdValue        float64
	BalanceId             string
	BalanceType           string
	BalanceDestinationIds string
	BalanceRatingSubject  string
	BalanceWeight         float64
	BalanceExpiryTime     string
	BalanceSharedGroup    string
	Weight                float64
	ActionsId             string
}

// Deprecated in rc8, replaced by AddAccountActionTriggers
func (apierSv1 *APIerSv1) AddTriggeredAction(attr AttrAddActionTrigger, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{utils.AccountField}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := attr.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	at := &engine.ActionTrigger{
		ID:             attr.ActionTriggersId,
		ThresholdType:  attr.ThresholdType,
		ThresholdValue: attr.ThresholdValue,
		Balance:        new(engine.BalanceFilter),
		Weight:         attr.Weight,
		ActionsID:      attr.ActionsId,
	}
	if attr.BalanceId != "" {
		at.Balance.ID = utils.StringPointer(attr.BalanceId)
	}
	if attr.BalanceType != "" {
		at.Balance.Type = utils.StringPointer(attr.BalanceType)
	}
	if attr.BalanceDestinationIds != "" {
		dstIDsMp := utils.StringMapFromSlice(strings.Split(attr.BalanceDestinationIds, utils.InfieldSep))
		at.Balance.DestinationIDs = &dstIDsMp
	}
	if attr.BalanceRatingSubject != "" {
		at.Balance.RatingSubject = utils.StringPointer(attr.BalanceRatingSubject)
	}
	if attr.BalanceWeight != 0.0 {
		at.Balance.Weight = utils.Float64Pointer(attr.BalanceWeight)
	}
	if balExpiryTime, err := utils.ParseTimeDetectLayout(attr.BalanceExpiryTime,
		apierSv1.Config.GeneralCfg().DefaultTimezone); err != nil {
		return utils.NewErrServerError(err)
	} else {
		at.Balance.ExpirationDate = &balExpiryTime
	}
	if attr.BalanceSharedGroup != "" {
		at.Balance.SharedGroups = &utils.StringMap{attr.BalanceSharedGroup: true}
	}
	acntID := utils.ConcatenatedKey(tnt, attr.Account)
	_, err := guardian.Guardian.Guard(func() (interface{}, error) {
		acnt, err := apierSv1.DataManager.GetAccount(acntID)
		if err != nil {
			return 0, err
		}
		acnt.ActionTriggers = append(acnt.ActionTriggers, at)

		return 0, apierSv1.DataManager.SetAccount(acnt)
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.AccountPrefix+acntID)
	if err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}
