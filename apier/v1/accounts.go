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
	"math"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

type AccountActionTiming struct {
	ActionPlanId string    // The id of the ActionPlanId profile attached to the account
	Uuid         string    // The id to reference this particular ActionTiming
	ActionsId    string    // The id of actions which will be executed
	NextExecTime time.Time // Next execution time
}

func (api *APIerSv1) GetAccountActionPlan(attrs utils.TenantAccount, reply *[]*AccountActionTiming) error {
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(strings.Join(missing, ","), "")
	}
	acntID := utils.ConcatenatedKey(attrs.Tenant, attrs.Account)
	acntATsIf, err := guardian.Guardian.Guard(func() (interface{}, error) {
		acntAPids, err := api.DataManager.GetAccountActionPlans(acntID, true, true, utils.NonTransactional)
		if err != nil && err != utils.ErrNotFound {
			return nil, utils.NewErrServerError(err)
		}
		var acntAPs []*engine.ActionPlan
		for _, apID := range acntAPids {
			if ap, err := api.DataManager.GetActionPlan(apID, true, true, utils.NonTransactional); err != nil {
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
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.ACTION_PLAN_PREFIX)
	if err != nil {
		return err
	}
	*reply = acntATsIf.([]*AccountActionTiming)
	return nil
}

type AttrRemoveActionTiming struct {
	ActionPlanId    string // Id identifying the ActionTimings profile
	ActionTimingId  string // Internal CGR id identifying particular ActionTiming, *all for all user related ActionTimings to be canceled
	Tenant          string // Tenant the account belongs to
	Account         string // Account name
	ReloadScheduler bool   // If set it will reload the scheduler after adding
}

// Removes an ActionTimings or parts of it depending on filters being set
func (api *APIerSv1) RemoveActionTiming(attrs AttrRemoveActionTiming, reply *string) (err error) {
	if missing := utils.MissingStructFields(&attrs, []string{"ActionPlanId"}); len(missing) != 0 { // Only mandatory ActionPlanId
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	var accID string
	if len(attrs.Account) != 0 { // Presence of Account requires complete account details to be provided
		if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Account"}); len(missing) != 0 {
			return utils.NewErrMandatoryIeMissing(missing...)
		}
		accID = utils.ConcatenatedKey(attrs.Tenant, attrs.Account)
	}

	var remAcntAPids []string // list of accounts who's indexes need modification
	_, err = guardian.Guardian.Guard(func() (interface{}, error) {
		ap, err := api.DataManager.GetActionPlan(attrs.ActionPlanId, true, true, utils.NonTransactional)
		if err != nil {
			return 0, err
		} else if ap == nil {
			return 0, utils.ErrNotFound
		}
		if accID != "" {
			delete(ap.AccountIDs, accID)
			remAcntAPids = append(remAcntAPids, accID)
			err = api.DataManager.SetActionPlan(ap.Id, ap, true, utils.NonTransactional)
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
			err = api.DataManager.SetActionPlan(ap.Id, ap, true, utils.NonTransactional)
			goto UPDATE
		}
		if attrs.ActionPlanId != "" { // delete the entire action plan
			ap.ActionTimings = nil              // will delete the action plan
			for acntID := range ap.AccountIDs { // Make sure we clear indexes for all accounts
				remAcntAPids = append(remAcntAPids, acntID)
			}
			err = api.DataManager.SetActionPlan(ap.Id, ap, true, utils.NonTransactional)
			goto UPDATE
		}

	UPDATE:
		if err != nil {
			return 0, err
		}
		if err := api.ConnMgr.Call(api.Config.ApierCfg().CachesConns, nil,
			utils.CacheSv1ReloadCache, utils.AttrReloadCacheWithArgDispatcher{
				TenantArg: utils.TenantArg{
					Tenant: attrs.Tenant,
				},
				AttrReloadCache: utils.AttrReloadCache{
					ArgsCache: utils.ArgsCache{ActionPlanIDs: []string{attrs.ActionPlanId}},
				},
			}, reply); err != nil {
			return 0, err
		}
		for _, acntID := range remAcntAPids {
			if err = api.DataManager.RemAccountActionPlans(acntID, []string{attrs.ActionPlanId}); err != nil {
				return 0, nil
			}
		}
		if len(remAcntAPids) != 0 {
			if err := api.ConnMgr.Call(api.Config.ApierCfg().CachesConns, nil,
				utils.CacheSv1ReloadCache, utils.AttrReloadCacheWithArgDispatcher{
					TenantArg: utils.TenantArg{
						Tenant: attrs.Tenant,
					},
					AttrReloadCache: utils.AttrReloadCache{
						ArgsCache: utils.ArgsCache{AccountActionPlanIDs: remAcntAPids},
					},
				}, reply); err != nil {
				return 0, err
			}
		}
		return 0, nil
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.ACTION_PLAN_PREFIX)
	if err != nil {
		*reply = err.Error()
		return utils.NewErrServerError(err)
	}
	if attrs.ReloadScheduler {
		sched := api.SchedulerService.GetScheduler()
		if sched == nil {
			return errors.New(utils.SchedulerNotRunningCaps)
		}
		sched.Reload()
	}
	*reply = utils.OK
	return nil
}

// Ads a new account into dataDb. If already defined, returns success.
func (api *APIerSv1) SetAccount(attr utils.AttrSetAccount, reply *string) (err error) {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	accID := utils.ConcatenatedKey(attr.Tenant, attr.Account)
	dirtyActionPlans := make(map[string]*engine.ActionPlan)
	_, err = guardian.Guardian.Guard(func() (interface{}, error) {
		var ub *engine.Account
		if bal, _ := api.DataManager.GetAccount(accID); bal != nil {
			ub = bal
		} else { // Not found in db, create it here
			ub = &engine.Account{
				ID: accID,
			}
		}
		if attr.ActionPlanID != "" {
			_, err := guardian.Guardian.Guard(func() (interface{}, error) {
				acntAPids, err := api.DataManager.GetAccountActionPlans(accID, true, true, utils.NonTransactional)
				if err != nil && err != utils.ErrNotFound {
					return 0, err
				}
				// clean previous action plans
				for i := 0; i < len(acntAPids); {
					apID := acntAPids[i]
					if apID == attr.ActionPlanID {
						i++ // increase index since we don't remove from slice
						continue
					}
					ap, err := api.DataManager.GetActionPlan(apID, true, true, utils.NonTransactional)
					if err != nil {
						return 0, err
					}
					delete(ap.AccountIDs, accID)
					dirtyActionPlans[apID] = ap
					acntAPids = append(acntAPids[:i], acntAPids[i+1:]...) // remove the item from the list so we can overwrite the real list
				}
				if !utils.IsSliceMember(acntAPids, attr.ActionPlanID) { // Account not yet attached to action plan, do it here
					ap, err := api.DataManager.GetActionPlan(attr.ActionPlanID, true, true, utils.NonTransactional)
					if err != nil {
						return 0, err
					}
					if ap.AccountIDs == nil {
						ap.AccountIDs = make(utils.StringMap)
					}
					ap.AccountIDs[accID] = true
					dirtyActionPlans[attr.ActionPlanID] = ap
					acntAPids = append(acntAPids, attr.ActionPlanID)
					// create tasks
					for _, at := range ap.ActionTimings {
						if at.IsASAP() {
							t := &engine.Task{
								Uuid:      utils.GenUUID(),
								AccountID: accID,
								ActionsID: at.ActionsID,
							}
							if err = api.DataManager.DataDB().PushTask(t); err != nil {
								return 0, err
							}
						}
					}
				}
				apIDs := make([]string, len(dirtyActionPlans))
				i := 0
				for actionPlanID, ap := range dirtyActionPlans {
					if err := api.DataManager.SetActionPlan(actionPlanID, ap, true, utils.NonTransactional); err != nil {
						return 0, err
					}
					apIDs[i] = actionPlanID
					i++
				}
				if err := api.DataManager.SetAccountActionPlans(accID, acntAPids, true); err != nil {
					return 0, err
				}
				if err := api.ConnMgr.Call(api.Config.ApierCfg().CachesConns, nil,
					utils.CacheSv1ReloadCache, utils.AttrReloadCacheWithArgDispatcher{
						TenantArg: utils.TenantArg{
							Tenant: attr.Tenant,
						},
						AttrReloadCache: utils.AttrReloadCache{
							ArgsCache: utils.ArgsCache{AccountActionPlanIDs: []string{accID}, ActionPlanIDs: apIDs},
						},
					}, reply); err != nil {
					return 0, err
				}
				return 0, nil
			}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.ACTION_PLAN_PREFIX)
			if err != nil {
				return 0, err
			}
		}

		if attr.ActionTriggersID != "" {
			atrs, err := api.DataManager.GetActionTriggers(attr.ActionTriggersID, false, utils.NonTransactional)
			if err != nil {
				return 0, err
			}
			ub.ActionTriggers = atrs
			ub.InitCounters()
		}

		if alNeg, has := attr.ExtraOptions[utils.AllowNegative]; has {
			ub.AllowNegative = alNeg
		}
		if dis, has := attr.ExtraOptions[utils.Disabled]; has {
			ub.Disabled = dis
		}
		// All prepared, save account
		if err := api.DataManager.SetAccount(ub); err != nil {
			return 0, err
		}
		return 0, nil
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.ACCOUNT_PREFIX+accID)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if attr.ReloadScheduler && len(dirtyActionPlans) != 0 {
		sched := api.SchedulerService.GetScheduler()
		if sched == nil {
			return errors.New(utils.SchedulerNotRunningCaps)
		}
		sched.Reload()
	}
	*reply = utils.OK // This will mark saving of the account, error still can show up in actionTimingsId
	return nil
}

func (api *APIerSv1) RemoveAccount(attr utils.AttrRemoveAccount, reply *string) (err error) {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	dirtyActionPlans := make(map[string]*engine.ActionPlan)
	accID := utils.ConcatenatedKey(attr.Tenant, attr.Account)
	_, err = guardian.Guardian.Guard(func() (interface{}, error) {
		// remove it from all action plans
		_, err := guardian.Guardian.Guard(func() (interface{}, error) {
			actionPlansMap, err := api.DataManager.GetAllActionPlans()
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
				if err := api.DataManager.SetActionPlan(actionPlanID, ap, true,
					utils.NonTransactional); err != nil {
					return 0, err
				}
			}
			return 0, nil
		}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.ACTION_PLAN_PREFIX)
		if err != nil {
			return 0, err
		}
		if err := api.DataManager.RemoveAccount(accID); err != nil {
			return 0, err
		}
		return 0, nil
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.ACCOUNT_PREFIX+accID)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if err = api.DataManager.RemAccountActionPlans(accID, nil); err != nil &&
		err.Error() != utils.ErrNotFound.Error() {
		return err
	}
	if err = api.ConnMgr.Call(api.Config.ApierCfg().CachesConns, nil,
		utils.CacheSv1ReloadCache, utils.AttrReloadCacheWithArgDispatcher{
			TenantArg: utils.TenantArg{
				Tenant: attr.Tenant,
			},
			AttrReloadCache: utils.AttrReloadCache{
				ArgsCache: utils.ArgsCache{AccountActionPlanIDs: []string{accID}},
			},
		}, reply); err != nil {
		return
	}
	*reply = utils.OK
	return nil
}

func (api *APIerSv1) GetAccounts(attr utils.AttrGetAccounts, reply *[]interface{}) error {
	if len(attr.Tenant) == 0 {
		return utils.NewErrMandatoryIeMissing("Tenant")
	}
	var accountKeys []string
	var err error
	if len(attr.AccountIDs) == 0 {
		if accountKeys, err = api.DataManager.DataDB().GetKeysForPrefix(utils.ACCOUNT_PREFIX + attr.Tenant); err != nil {
			return err
		}
	} else {
		for _, acntID := range attr.AccountIDs {
			if len(acntID) == 0 { // Source of error returned from redis (key not found)
				continue
			}
			accountKeys = append(accountKeys, utils.ACCOUNT_PREFIX+utils.ConcatenatedKey(attr.Tenant, acntID))
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
		if acnt, err := api.DataManager.GetAccount(acntKey[len(utils.ACCOUNT_PREFIX):]); err != nil && err != utils.ErrNotFound { // Not found is not an error here
			return err
		} else if acnt != nil {
			if alNeg, has := attr.Filter[utils.AllowNegative]; has && alNeg != acnt.AllowNegative {
				continue
			}
			if dis, has := attr.Filter[utils.Disabled]; has && dis != acnt.Disabled {
				continue
			}
			retAccounts = append(retAccounts, acnt.AsOldStructure())
		}
	}
	*reply = retAccounts
	return nil
}

// GetAccount returns the account
func (api *APIerSv1) GetAccount(attr *utils.AttrGetAccount, reply *interface{}) error {
	tag := utils.ConcatenatedKey(attr.Tenant, attr.Account)
	userBalance, err := api.DataManager.GetAccount(tag)
	if err != nil {
		return err
	}

	*reply = userBalance.AsOldStructure()
	return nil
}

type AttrAddBalance struct {
	Tenant          string
	Account         string
	BalanceType     string
	Value           float64
	Balance         map[string]interface{}
	ActionExtraData *map[string]interface{}
	Overwrite       bool // When true it will reset if the balance is already there
	Cdrlog          bool
}

func (api *APIerSv1) AddBalance(attr *AttrAddBalance, reply *string) error {
	return api.modifyBalance(utils.TOPUP, attr, reply)
}
func (api *APIerSv1) DebitBalance(attr *AttrAddBalance, reply *string) error {
	return api.modifyBalance(utils.DEBIT, attr, reply)
}

func (api *APIerSv1) modifyBalance(aType string, attr *AttrAddBalance, reply *string) (err error) {
	if missing := utils.MissingStructFields(attr, []string{"Tenant", "Account", "BalanceType", "Value"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	var balance *engine.BalanceFilter
	if balance, err = engine.NewBalanceFilter(attr.Balance, api.Config.GeneralCfg().DefaultTimezone); err != nil {
		return
	}
	balance.Type = utils.StringPointer(attr.BalanceType)
	if attr.Value != 0 {
		balance.Value = &utils.ValueFormula{Static: math.Abs(attr.Value)}
	}

	accID := utils.ConcatenatedKey(attr.Tenant, attr.Account)
	if _, err = api.DataManager.GetAccount(accID); err != nil {
		// create account if does not exist
		account := &engine.Account{
			ID: accID,
		}
		if err = api.DataManager.SetAccount(account); err != nil {
			return
		}
	}
	at := &engine.ActionTiming{}
	//check if we have extra data
	if attr.ActionExtraData != nil && len(*attr.ActionExtraData) != 0 {
		at.ExtraData = *attr.ActionExtraData
	}
	at.SetAccountIDs(utils.StringMap{accID: true})

	if attr.Overwrite {
		aType += "_reset" // => *topup_reset/*debit_reset
	}
	if balance.TimingIDs != nil {
		for _, timingID := range balance.TimingIDs.Slice() {
			var tmg *utils.TPTiming
			if tmg, err = api.DataManager.GetTiming(timingID, false, utils.NonTransactional); err != nil {
				return
			}
			balance.Timings = append(balance.Timings, &engine.RITiming{
				Years:     tmg.Years,
				Months:    tmg.Months,
				MonthDays: tmg.MonthDays,
				WeekDays:  tmg.WeekDays,
				StartTime: tmg.StartTime,
				EndTime:   tmg.EndTime,
			})
		}
	}

	a := &engine.Action{
		ActionType: aType,
		Balance:    balance,
	}
	publishAction := &engine.Action{
		ActionType: utils.MetaPublishBalance,
	}
	acts := engine.Actions{a, publishAction}
	if attr.Cdrlog {
		acts = engine.Actions{a, publishAction, &engine.Action{
			ActionType: utils.CDRLOG,
		}}
	}
	at.SetActions(acts)
	if err := at.Execute(nil, nil); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetBalance sets the balance for the given account
// if the account is not already created it will create the account also
func (api *APIerSv1) SetBalance(attr *utils.AttrSetBalance, reply *string) (err error) {
	if missing := utils.MissingStructFields(attr, []string{"Tenant", "Account", "BalanceType"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	var balance *engine.BalanceFilter
	if balance, err = engine.NewBalanceFilter(attr.Balance, api.Config.GeneralCfg().DefaultTimezone); err != nil {
		return
	}
	balance.Type = utils.StringPointer(attr.BalanceType)
	if attr.Value != 0 {
		balance.Value = &utils.ValueFormula{Static: math.Abs(attr.Value)}
	}
	if (balance.ID == nil || *balance.ID == "") &&
		(balance.Uuid == nil || *balance.Uuid == "") {
		return utils.NewErrMandatoryIeMissing("BalanceID", "or", "BalanceUUID")
	}

	accID := utils.ConcatenatedKey(attr.Tenant, attr.Account)
	if _, err = api.DataManager.GetAccount(accID); err != nil {
		// create account if not exists
		account := &engine.Account{
			ID: accID,
		}
		if err = api.DataManager.SetAccount(account); err != nil {
			return
		}
	}
	at := &engine.ActionTiming{}
	//check if we have extra data
	if attr.ActionExtraData != nil && len(*attr.ActionExtraData) != 0 {
		at.ExtraData = *attr.ActionExtraData
	}
	at.SetAccountIDs(utils.StringMap{accID: true})
	if balance.TimingIDs != nil {
		for _, timingID := range balance.TimingIDs.Slice() {
			var tmg *utils.TPTiming
			if tmg, err = api.DataManager.GetTiming(timingID, false, utils.NonTransactional); err != nil {
				return
			}
			balance.Timings = append(balance.Timings, &engine.RITiming{
				Years:     tmg.Years,
				Months:    tmg.Months,
				MonthDays: tmg.MonthDays,
				WeekDays:  tmg.WeekDays,
				StartTime: tmg.StartTime,
				EndTime:   tmg.EndTime,
			})
		}
	}

	a := &engine.Action{
		ActionType: utils.SET_BALANCE,
		Balance:    balance,
	}
	publishAction := &engine.Action{
		ActionType: utils.MetaPublishBalance,
	}
	acts := engine.Actions{a, publishAction}
	if attr.Cdrlog {
		acts = engine.Actions{a, publishAction, &engine.Action{
			ActionType: utils.CDRLOG,
		}}
	}
	at.SetActions(acts)
	if err = at.Execute(nil, nil); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveBalances remove the matching balances for the account
func (api *APIerSv1) RemoveBalances(attr *utils.AttrSetBalance, reply *string) (err error) {
	if missing := utils.MissingStructFields(attr, []string{"Tenant", "Account", "BalanceType"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	var balance *engine.BalanceFilter
	if balance, err = engine.NewBalanceFilter(attr.Balance, api.Config.GeneralCfg().DefaultTimezone); err != nil {
		return
	}
	balance.Type = utils.StringPointer(attr.BalanceType)

	accID := utils.ConcatenatedKey(attr.Tenant, attr.Account)
	if _, err := api.DataManager.GetAccount(accID); err != nil {
		return utils.ErrNotFound
	}

	at := &engine.ActionTiming{}
	//check if we have extra data
	if attr.ActionExtraData != nil && len(*attr.ActionExtraData) != 0 {
		at.ExtraData = *attr.ActionExtraData
	}
	at.SetAccountIDs(utils.StringMap{accID: true})
	a := &engine.Action{
		ActionType: utils.REMOVE_BALANCE,
		Balance:    balance,
	}
	at.SetActions(engine.Actions{a})
	if err := at.Execute(nil, nil); err != nil {
		*reply = err.Error()
		return err
	}
	*reply = utils.OK
	return nil
}

func (api *APIerSv1) GetAccountsCount(attr utils.TenantArg, reply *int) (err error) {
	if len(attr.Tenant) == 0 {
		return utils.NewErrMandatoryIeMissing("Tenant")
	}
	var accountKeys []string
	if accountKeys, err = api.DataManager.DataDB().GetKeysForPrefix(utils.ACCOUNT_PREFIX + attr.Tenant); err != nil {
		return
	}
	*reply = len(accountKeys)
	return
}
