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
	"math"
	"strings"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

type AttrAcntAction struct {
	Tenant  string
	Account string
}

type AccountActionTiming struct {
	ActionPlanId string    // The id of the ActionPlanId profile attached to the account
	Uuid         string    // The id to reference this particular ActionTiming
	ActionsId    string    // The id of actions which will be executed
	NextExecTime time.Time // Next execution time
}

func (self *ApierV1) GetAccountActionPlan(attrs AttrAcntAction, reply *[]*AccountActionTiming) error {
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(strings.Join(missing, ","), "")
	}
	acntID := utils.AccountKey(attrs.Tenant, attrs.Account)
	acntATsIf, err := guardian.Guardian.Guard(func() (interface{}, error) {
		acntAPids, err := self.DataManager.DataDB().GetAccountActionPlans(acntID, false, utils.NonTransactional)
		if err != nil && err != utils.ErrNotFound {
			return nil, utils.NewErrServerError(err)
		}
		var acntAPs []*engine.ActionPlan
		for _, apID := range acntAPids {
			if ap, err := self.DataManager.DataDB().GetActionPlan(apID, false, utils.NonTransactional); err != nil {
				return nil, err
			} else if ap != nil {
				acntAPs = append(acntAPs, ap)
			}
		}

		accountATs := make([]*AccountActionTiming, 0) // needs to be initialized if remains empty
		for _, ap := range acntAPs {
			for _, at := range ap.ActionTimings {
				accountATs = append(accountATs, &AccountActionTiming{
					ActionPlanId: ap.Id,
					Uuid:         at.Uuid,
					ActionsId:    at.ActionsID,
					NextExecTime: at.GetNextStartTime(time.Now()),
				})
			}
		}
		return accountATs, nil
	}, 0, utils.ACTION_PLAN_PREFIX)
	if err != nil {
		return err
	}
	*reply = acntATsIf.([]*AccountActionTiming)
	return nil
}

type AttrRemActionTiming struct {
	ActionPlanId    string // Id identifying the ActionTimings profile
	ActionTimingId  string // Internal CGR id identifying particular ActionTiming, *all for all user related ActionTimings to be canceled
	Tenant          string // Tenant the account belongs to
	Account         string // Account name
	ReloadScheduler bool   // If set it will reload the scheduler after adding
}

// Removes an ActionTimings or parts of it depending on filters being set
func (self *ApierV1) RemActionTiming(attrs AttrRemActionTiming, reply *string) (err error) {
	if missing := utils.MissingStructFields(&attrs, []string{"ActionPlanId"}); len(missing) != 0 { // Only mandatory ActionPlanId
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	var accID string
	if len(attrs.Account) != 0 { // Presence of Account requires complete account details to be provided
		if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Account"}); len(missing) != 0 {
			return utils.NewErrMandatoryIeMissing(missing...)
		}
		accID = utils.AccountKey(attrs.Tenant, attrs.Account)
	}

	var remAcntAPids []string // list of accounts who's indexes need modification
	_, err = guardian.Guardian.Guard(func() (interface{}, error) {
		ap, err := self.DataManager.DataDB().GetActionPlan(attrs.ActionPlanId, false, utils.NonTransactional)
		if err != nil {
			return 0, err
		} else if ap == nil {
			return 0, utils.ErrNotFound
		}
		if accID != "" {
			delete(ap.AccountIDs, accID)
			remAcntAPids = append(remAcntAPids, accID)
			err = self.DataManager.DataDB().SetActionPlan(ap.Id, ap, true, utils.NonTransactional)
			goto UPDATE
		}
		if attrs.ActionTimingId != "" { // delete only a action timing from action plan
			for i, at := range ap.ActionTimings {
				if at.Uuid == attrs.ActionTimingId {
					ap.ActionTimings[i] = ap.ActionTimings[len(ap.ActionTimings)-1]
					ap.ActionTimings = ap.ActionTimings[:len(ap.ActionTimings)-1]
					break
				}
			}
			err = self.DataManager.DataDB().SetActionPlan(ap.Id, ap, true, utils.NonTransactional)
			goto UPDATE
		}
		if attrs.ActionPlanId != "" { // delete the entire action plan
			ap.ActionTimings = nil              // will delete the action plan
			for acntID := range ap.AccountIDs { // Make sure we clear indexes for all accounts
				remAcntAPids = append(remAcntAPids, acntID)
			}
			err = self.DataManager.DataDB().SetActionPlan(ap.Id, ap, true, utils.NonTransactional)
			goto UPDATE
		}

	UPDATE:
		if err != nil {
			return 0, err
		}
		if err = self.DataManager.CacheDataFromDB(utils.ACTION_PLAN_PREFIX, []string{attrs.ActionPlanId}, true); err != nil {
			return 0, err
		}
		for _, acntID := range remAcntAPids {
			if err = self.DataManager.DataDB().RemAccountActionPlans(acntID, []string{attrs.ActionPlanId}); err != nil {
				return 0, nil
			}
		}
		if len(remAcntAPids) != 0 {
			if err = self.DataManager.CacheDataFromDB(utils.AccountActionPlansPrefix, remAcntAPids, true); err != nil {
				return 0, nil
			}
		}
		return 0, nil
	}, 0, utils.ACTION_PLAN_PREFIX)
	if err != nil {
		*reply = err.Error()
		return utils.NewErrServerError(err)
	}
	if attrs.ReloadScheduler {
		sched := self.ServManager.GetScheduler()
		if sched == nil {
			return errors.New(utils.SchedulerNotRunningCaps)
		}
		sched.Reload()
	}
	*reply = OK
	return nil
}

// Ads a new account into dataDb. If already defined, returns success.
func (self *ApierV1) SetAccount(attr utils.AttrSetAccount, reply *string) (err error) {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	accID := utils.AccountKey(attr.Tenant, attr.Account)
	dirtyActionPlans := make(map[string]*engine.ActionPlan)
	_, err = guardian.Guardian.Guard(func() (interface{}, error) {
		var ub *engine.Account
		if bal, _ := self.DataManager.DataDB().GetAccount(accID); bal != nil {
			ub = bal
		} else { // Not found in db, create it here
			ub = &engine.Account{
				ID: accID,
			}
		}
		if attr.ActionPlanId != "" {
			_, err := guardian.Guardian.Guard(func() (interface{}, error) {
				acntAPids, err := self.DataManager.DataDB().GetAccountActionPlans(accID, false, utils.NonTransactional)
				if err != nil && err != utils.ErrNotFound {
					return 0, err
				}
				// clean previous action plans
				for i := 0; i < len(acntAPids); {
					apID := acntAPids[i]
					if apID == attr.ActionPlanId {
						i++ // increase index since we don't remove from slice
						continue
					}
					ap, err := self.DataManager.DataDB().GetActionPlan(apID, false, utils.NonTransactional)
					if err != nil {
						return 0, err
					}
					delete(ap.AccountIDs, accID)
					dirtyActionPlans[apID] = ap
					acntAPids = append(acntAPids[:i], acntAPids[i+1:]...) // remove the item from the list so we can overwrite the real list
				}
				if !utils.IsSliceMember(acntAPids, attr.ActionPlanId) { // Account not yet attached to action plan, do it here
					ap, err := self.DataManager.DataDB().GetActionPlan(attr.ActionPlanId, false, utils.NonTransactional)
					if err != nil {
						return 0, err
					}
					if ap.AccountIDs == nil {
						ap.AccountIDs = make(utils.StringMap)
					}
					ap.AccountIDs[accID] = true
					dirtyActionPlans[attr.ActionPlanId] = ap
					acntAPids = append(acntAPids, attr.ActionPlanId)
					// create tasks
					for _, at := range ap.ActionTimings {
						if at.IsASAP() {
							t := &engine.Task{
								Uuid:      utils.GenUUID(),
								AccountID: accID,
								ActionsID: at.ActionsID,
							}
							if err = self.DataManager.DataDB().PushTask(t); err != nil {
								return 0, err
							}
						}
					}
				}
				apIDs := make([]string, len(dirtyActionPlans))
				i := 0
				for actionPlanID, ap := range dirtyActionPlans {
					if err := self.DataManager.DataDB().SetActionPlan(actionPlanID, ap, true, utils.NonTransactional); err != nil {
						return 0, err
					}
					apIDs[i] = actionPlanID
					i++
				}
				if err := self.DataManager.CacheDataFromDB(utils.ACTION_PLAN_PREFIX, apIDs, true); err != nil {
					return 0, err
				}
				if err := self.DataManager.DataDB().SetAccountActionPlans(accID, acntAPids, true); err != nil {
					return 0, err
				}
				if err = self.DataManager.CacheDataFromDB(utils.AccountActionPlansPrefix, []string{accID}, true); err != nil {
					return 0, err
				}
				return 0, nil
			}, 0, utils.ACTION_PLAN_PREFIX)
			if err != nil {
				return 0, err
			}
		}

		if attr.ActionTriggersId != "" {
			atrs, err := self.DataManager.GetActionTriggers(attr.ActionTriggersId, false, utils.NonTransactional)
			if err != nil {
				return 0, err
			}
			ub.ActionTriggers = atrs
			ub.InitCounters()
		}

		if attr.AllowNegative != nil {
			ub.AllowNegative = *attr.AllowNegative
		}
		if attr.Disabled != nil {
			ub.Disabled = *attr.Disabled
		}
		// All prepared, save account
		if err := self.DataManager.DataDB().SetAccount(ub); err != nil {
			return 0, err
		}
		return 0, nil
	}, 0, accID)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if attr.ReloadScheduler && len(dirtyActionPlans) != 0 {
		sched := self.ServManager.GetScheduler()
		if sched == nil {
			return errors.New(utils.SchedulerNotRunningCaps)
		}
		sched.Reload()
	}
	*reply = utils.OK // This will mark saving of the account, error still can show up in actionTimingsId
	return nil
}

func (self *ApierV1) RemoveAccount(attr utils.AttrRemoveAccount, reply *string) (err error) {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	dirtyActionPlans := make(map[string]*engine.ActionPlan)
	accID := utils.AccountKey(attr.Tenant, attr.Account)
	_, err = guardian.Guardian.Guard(func() (interface{}, error) {
		// remove it from all action plans
		_, err := guardian.Guardian.Guard(func() (interface{}, error) {
			actionPlansMap, err := self.DataManager.DataDB().GetAllActionPlans()
			if err == utils.ErrNotFound {
				// no action plans
				return 0, nil
			}
			if err != nil {
				return 0, err
			}

			for actionPlanID, ap := range actionPlansMap {
				if _, exists := ap.AccountIDs[accID]; exists {
					delete(ap.AccountIDs, accID)
					dirtyActionPlans[actionPlanID] = ap
				}
			}

			for actionPlanID, ap := range dirtyActionPlans {
				if err := self.DataManager.DataDB().SetActionPlan(actionPlanID, ap, true,
					utils.NonTransactional); err != nil {
					return 0, err
				}
			}
			return 0, nil
		}, 0, utils.ACTION_PLAN_PREFIX)
		if err != nil {
			return 0, err
		}
		if err := self.DataManager.DataDB().RemoveAccount(accID); err != nil {
			return 0, err
		}
		return 0, nil
	}, 0, accID)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if err = self.DataManager.DataDB().RemAccountActionPlans(accID, nil); err != nil &&
		err.Error() != utils.ErrNotFound.Error() {
		return err
	}
	if err = self.DataManager.CacheDataFromDB(utils.AccountActionPlansPrefix,
		[]string{accID}, true); err != nil &&
		err.Error() != utils.ErrNotFound.Error() {
		return err
	}
	*reply = OK
	return nil
}

func (self *ApierV1) GetAccounts(attr utils.AttrGetAccounts, reply *[]interface{}) error {
	if len(attr.Tenant) == 0 {
		return utils.NewErrMandatoryIeMissing("Tenant")
	}
	var accountKeys []string
	var err error
	if len(attr.AccountIds) == 0 {
		if accountKeys, err = self.DataManager.DataDB().GetKeysForPrefix(utils.ACCOUNT_PREFIX + attr.Tenant); err != nil {
			return err
		}
	} else {
		for _, acntId := range attr.AccountIds {
			if len(acntId) == 0 { // Source of error returned from redis (key not found)
				continue
			}
			accountKeys = append(accountKeys, utils.ACCOUNT_PREFIX+utils.ConcatenatedKey(attr.Tenant, acntId))
		}
	}
	if len(accountKeys) == 0 {
		return nil
	}
	var limitedAccounts []string
	if attr.Limit != 0 {
		max := math.Min(float64(attr.Offset+attr.Limit), float64(len(accountKeys)))
		limitedAccounts = accountKeys[attr.Offset:int(max)]
	} else {
		limitedAccounts = accountKeys[attr.Offset:]
	}
	retAccounts := make([]interface{}, 0)
	for _, acntKey := range limitedAccounts {
		if acnt, err := self.DataManager.DataDB().GetAccount(acntKey[len(utils.ACCOUNT_PREFIX):]); err != nil && err != utils.ErrNotFound { // Not found is not an error here
			return err
		} else if acnt != nil {
			retAccounts = append(retAccounts, acnt.AsOldStructure())
		}
	}
	*reply = retAccounts
	return nil
}

// Get balance
func (self *ApierV1) GetAccount(attr *utils.AttrGetAccount, reply *interface{}) error {
	tag := fmt.Sprintf("%s:%s", attr.Tenant, attr.Account)
	userBalance, err := self.DataManager.DataDB().GetAccount(tag)
	if err != nil {
		return err
	}

	*reply = userBalance.AsOldStructure()
	return nil
}

type AttrAddBalance struct {
	Tenant         string
	Account        string
	BalanceUuid    *string
	BalanceId      *string
	BalanceType    string
	Directions     *string
	Value          float64
	ExpiryTime     *string
	RatingSubject  *string
	Categories     *string
	DestinationIds *string
	TimingIds      *string
	Weight         *float64
	SharedGroups   *string
	Overwrite      bool // When true it will reset if the balance is already there
	Blocker        *bool
	Disabled       *bool
}

func (self *ApierV1) AddBalance(attr *AttrAddBalance, reply *string) error {
	return self.modifyBalance(engine.TOPUP, attr, reply)
}
func (self *ApierV1) DebitBalance(attr *AttrAddBalance, reply *string) error {
	return self.modifyBalance(engine.DEBIT, attr, reply)
}

func (self *ApierV1) modifyBalance(aType string, attr *AttrAddBalance, reply *string) error {
	if missing := utils.MissingStructFields(attr, []string{"Tenant", "Account", "BalanceType", "Value"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	var expTime *time.Time
	if attr.ExpiryTime != nil {
		expTimeVal, err := utils.ParseTimeDetectLayout(*attr.ExpiryTime,
			self.Config.GeneralCfg().DefaultTimezone)
		if err != nil {
			*reply = err.Error()
			return err
		}
		expTime = &expTimeVal
	}
	accID := utils.AccountKey(attr.Tenant, attr.Account)
	if _, err := self.DataManager.DataDB().GetAccount(accID); err != nil {
		// create account if does not exist
		account := &engine.Account{
			ID: accID,
		}
		if err := self.DataManager.DataDB().SetAccount(account); err != nil {
			*reply = err.Error()
			return err
		}
	}
	at := &engine.ActionTiming{}
	at.SetAccountIDs(utils.StringMap{accID: true})

	if attr.Overwrite {
		aType += "_reset" // => *topup_reset/*debit_reset
	}
	a := &engine.Action{
		ActionType: aType,
		Balance: &engine.BalanceFilter{
			Uuid:           attr.BalanceUuid,
			ID:             attr.BalanceId,
			Type:           utils.StringPointer(attr.BalanceType),
			Value:          &utils.ValueFormula{Static: attr.Value},
			ExpirationDate: expTime,
			RatingSubject:  attr.RatingSubject,
			Weight:         attr.Weight,
			Blocker:        attr.Blocker,
			Disabled:       attr.Disabled,
		},
	}
	if attr.Directions != nil {
		a.Balance.Directions = utils.StringMapPointer(utils.ParseStringMap(*attr.Directions))
	}
	if attr.DestinationIds != nil {
		a.Balance.DestinationIDs = utils.StringMapPointer(utils.ParseStringMap(*attr.DestinationIds))
	}
	if attr.Categories != nil {
		a.Balance.Categories = utils.StringMapPointer(utils.ParseStringMap(*attr.Categories))
	}
	if attr.SharedGroups != nil {
		a.Balance.SharedGroups = utils.StringMapPointer(utils.ParseStringMap(*attr.SharedGroups))
	}
	if attr.TimingIds != nil {
		a.Balance.TimingIDs = utils.StringMapPointer(utils.ParseStringMap(*attr.TimingIds))
	}
	publishAction := &engine.Action{
		ActionType: engine.MetaPublishBalance,
	}
	at.SetActions(engine.Actions{a, publishAction})
	if err := at.Execute(nil, nil); err != nil {
		return err
	}
	*reply = OK
	return nil
}

func (self *ApierV1) SetBalance(attr *utils.AttrSetBalance, reply *string) error {
	if missing := utils.MissingStructFields(attr, []string{"Tenant", "Account", "BalanceType"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if (attr.BalanceID == nil || *attr.BalanceID == "") &&
		(attr.BalanceUUID == nil || *attr.BalanceUUID == "") {
		return utils.NewErrMandatoryIeMissing("BalanceID", "or", "BalanceUUID")
	}
	var expTime *time.Time
	if attr.ExpiryTime != nil {
		expTimeVal, err := utils.ParseTimeDetectLayout(*attr.ExpiryTime,
			self.Config.GeneralCfg().DefaultTimezone)
		if err != nil {
			*reply = err.Error()
			return err
		}
		expTime = &expTimeVal
	}
	accID := utils.AccountKey(attr.Tenant, attr.Account)
	if _, err := self.DataManager.DataDB().GetAccount(accID); err != nil {
		// create account if not exists
		account := &engine.Account{
			ID: accID,
		}
		if err := self.DataManager.DataDB().SetAccount(account); err != nil {
			*reply = err.Error()
			return err
		}
	}
	at := &engine.ActionTiming{}
	at.SetAccountIDs(utils.StringMap{accID: true})

	a := &engine.Action{
		ActionType: engine.SET_BALANCE,
		Balance: &engine.BalanceFilter{
			Uuid:           attr.BalanceUUID,
			ID:             attr.BalanceID,
			Type:           utils.StringPointer(attr.BalanceType),
			ExpirationDate: expTime,
			RatingSubject:  attr.RatingSubject,
			Weight:         attr.Weight,
			Blocker:        attr.Blocker,
			Disabled:       attr.Disabled,
		},
	}
	if attr.Value != nil {
		a.Balance.Value = &utils.ValueFormula{Static: *attr.Value}
	}
	if attr.Directions != nil {
		a.Balance.Directions = utils.StringMapPointer(utils.ParseStringMap(*attr.Directions))
	}
	if attr.DestinationIds != nil {
		a.Balance.DestinationIDs = utils.StringMapPointer(utils.ParseStringMap(*attr.DestinationIds))
	}
	if attr.Categories != nil {
		a.Balance.Categories = utils.StringMapPointer(utils.ParseStringMap(*attr.Categories))
	}
	if attr.SharedGroups != nil {
		a.Balance.SharedGroups = utils.StringMapPointer(utils.ParseStringMap(*attr.SharedGroups))
	}
	if attr.TimingIds != nil {
		a.Balance.TimingIDs = utils.StringMapPointer(utils.ParseStringMap(*attr.TimingIds))
	}
	publishAction := &engine.Action{
		ActionType: engine.MetaPublishBalance,
	}
	at.SetActions(engine.Actions{a, publishAction})
	if err := at.Execute(nil, nil); err != nil {
		*reply = err.Error()
		return err
	}
	*reply = OK
	return nil
}

func (self *ApierV1) RemoveBalances(attr *utils.AttrSetBalance, reply *string) error {
	if missing := utils.MissingStructFields(attr, []string{"Tenant", "Account", "BalanceType"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	var expTime *time.Time
	if attr.ExpiryTime != nil {
		expTimeVal, err := utils.ParseTimeDetectLayout(*attr.ExpiryTime,
			self.Config.GeneralCfg().DefaultTimezone)
		if err != nil {
			*reply = err.Error()
			return err
		}
		expTime = &expTimeVal
	}
	accID := utils.AccountKey(attr.Tenant, attr.Account)
	if _, err := self.DataManager.DataDB().GetAccount(accID); err != nil {
		return utils.ErrNotFound
	}

	at := &engine.ActionTiming{}
	at.SetAccountIDs(utils.StringMap{accID: true})
	a := &engine.Action{
		ActionType: engine.REMOVE_BALANCE,
		Balance: &engine.BalanceFilter{
			Uuid:           attr.BalanceUUID,
			ID:             attr.BalanceID,
			Type:           utils.StringPointer(attr.BalanceType),
			ExpirationDate: expTime,
			RatingSubject:  attr.RatingSubject,
			Weight:         attr.Weight,
			Blocker:        attr.Blocker,
			Disabled:       attr.Disabled,
		},
	}
	if attr.Value != nil {
		a.Balance.Value = &utils.ValueFormula{Static: *attr.Value}
	}
	if attr.Directions != nil {
		a.Balance.Directions = utils.StringMapPointer(utils.ParseStringMap(*attr.Directions))
	}
	if attr.DestinationIds != nil {
		a.Balance.DestinationIDs = utils.StringMapPointer(utils.ParseStringMap(*attr.DestinationIds))
	}
	if attr.Categories != nil {
		a.Balance.Categories = utils.StringMapPointer(utils.ParseStringMap(*attr.Categories))
	}
	if attr.SharedGroups != nil {
		a.Balance.SharedGroups = utils.StringMapPointer(utils.ParseStringMap(*attr.SharedGroups))
	}
	if attr.TimingIds != nil {
		a.Balance.TimingIDs = utils.StringMapPointer(utils.ParseStringMap(*attr.TimingIds))
	}
	at.SetActions(engine.Actions{a})
	if err := at.Execute(nil, nil); err != nil {
		*reply = err.Error()
		return err
	}
	*reply = OK
	return nil
}
