/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

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
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/cgrates/cgrates/engine"
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
	accountATs := make([]*AccountActionTiming, 0) // needs to be initialized if remains empty
	allAPs, err := self.RatingDb.GetAllActionPlans()
	if err != nil {
		return utils.NewErrServerError(err)
	}
	accID := utils.AccountKey(attrs.Tenant, attrs.Account)
	for _, ap := range allAPs {
		if _, exists := ap.AccountIDs[accID]; exists {
			for _, at := range ap.ActionTimings {
				accountATs = append(accountATs, &AccountActionTiming{
					ActionPlanId: ap.Id,
					Uuid:         at.Uuid,
					ActionsId:    at.ActionsID,
					NextExecTime: at.GetNextStartTime(time.Now()),
				})
			}
		}

	}
	*reply = accountATs
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
func (self *ApierV1) RemActionTiming(attrs AttrRemActionTiming, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"ActionPlanId"}); len(missing) != 0 { // Only mandatory ActionPlanId
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if len(attrs.Account) != 0 { // Presence of Account requires complete account details to be provided
		if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Account"}); len(missing) != 0 {
			return utils.NewErrMandatoryIeMissing(missing...)
		}
	}
	_, err := engine.Guardian.Guard(func() (interface{}, error) {
		ap, err := self.RatingDb.GetActionPlan(attrs.ActionPlanId, false)
		if err != nil {
			return 0, err
		} else if ap == nil {
			return 0, utils.ErrNotFound
		}

		if attrs.ActionPlanId != "" { // delete the entire action plan
			ap.ActionTimings = nil // will delete the action plan
			return 0, self.RatingDb.SetActionPlan(ap.Id, ap, true)
		}

		if attrs.ActionTimingId != "" { // delete only a action timing from action plan
			for i, at := range ap.ActionTimings {
				if at.Uuid == attrs.ActionTimingId {
					ap.ActionTimings[i] = ap.ActionTimings[len(ap.ActionTimings)-1]
					ap.ActionTimings = ap.ActionTimings[:len(ap.ActionTimings)-1]
					break
				}
			}
			return 0, self.RatingDb.SetActionPlan(ap.Id, ap, true)
		}

		if attrs.Tenant != "" && attrs.Account != "" {
			accID := utils.AccountKey(attrs.Tenant, attrs.Account)
			delete(ap.AccountIDs, accID)
			return 0, self.RatingDb.SetActionPlan(ap.Id, ap, true)
		}

		// update cache
		self.RatingDb.CacheRatingPrefixValues(map[string][]string{utils.ACTION_PLAN_PREFIX: []string{utils.ACTION_PLAN_PREFIX + attrs.ActionPlanId}})
		return 0, nil
	}, 0, utils.ACTION_PLAN_PREFIX)
	if err != nil {
		*reply = err.Error()
		return utils.NewErrServerError(err)
	}
	if attrs.ReloadScheduler && self.Sched != nil {
		self.Sched.Reload(true)
	}
	*reply = OK
	return nil
}

// Ads a new account into dataDb. If already defined, returns success.
func (self *ApierV1) SetAccount(attr utils.AttrSetAccount, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	var schedulerReloadNeeded = false
	accID := utils.AccountKey(attr.Tenant, attr.Account)
	var ub *engine.Account
	_, err := engine.Guardian.Guard(func() (interface{}, error) {
		if bal, _ := self.AccountDb.GetAccount(accID); bal != nil {
			ub = bal
		} else { // Not found in db, create it here
			ub = &engine.Account{
				Id: accID,
			}
		}
		if len(attr.ActionPlanId) != 0 {
			_, err := engine.Guardian.Guard(func() (interface{}, error) {
				// clean previous action plans
				actionPlansMap, err := self.RatingDb.GetAllActionPlans()
				if err != nil {
					return 0, err
				}
				var dirtyAps []string
				for actionPlanID, ap := range actionPlansMap {
					if actionPlanID == attr.ActionPlanId {
						// don't remove it if it's the current one
						continue
					}
					if _, exists := ap.AccountIDs[accID]; exists {
						delete(ap.AccountIDs, accID)
						dirtyAps = append(dirtyAps, utils.ACTION_PLAN_PREFIX+actionPlanID)
					}
				}

				if len(dirtyAps) > 0 {
					// update cache
					self.RatingDb.CacheRatingPrefixValues(map[string][]string{utils.ACTION_PLAN_PREFIX: dirtyAps})
					schedulerReloadNeeded = true
				}

				var ap *engine.ActionPlan
				ap, err = self.RatingDb.GetActionPlan(attr.ActionPlanId, false)
				if err != nil {
					return 0, err
				}
				if _, exists := ap.AccountIDs[accID]; !exists {
					if ap.AccountIDs == nil {
						ap.AccountIDs = make(utils.StringMap)
					}
					ap.AccountIDs[accID] = true
					schedulerReloadNeeded = true
					// create tasks
					for _, at := range ap.ActionTimings {
						if at.IsASAP() {
							t := &engine.Task{
								Uuid:      utils.GenUUID(),
								AccountID: accID,
								ActionsID: at.ActionsID,
							}
							if err = self.RatingDb.PushTask(t); err != nil {
								return 0, err
							}
						}
					}
					if err := self.RatingDb.SetActionPlan(attr.ActionPlanId, ap, true); err != nil {
						return 0, err
					}
					// update cache
					self.RatingDb.CacheRatingPrefixValues(map[string][]string{utils.ACTION_PLAN_PREFIX: []string{utils.ACTION_PLAN_PREFIX + attr.ActionPlanId}})
				}
				return 0, nil
			}, 0, utils.ACTION_PLAN_PREFIX)
			if err != nil {
				return 0, err
			}
		}

		if len(attr.ActionTriggersId) != 0 {
			atrs, err := self.RatingDb.GetActionTriggers(attr.ActionTriggersId)
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
		if err := self.AccountDb.SetAccount(ub); err != nil {
			return 0, err
		}
		return 0, nil
	}, 0, accID)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if attr.ReloadScheduler && schedulerReloadNeeded {
		// reload scheduler
		if self.Sched != nil {
			self.Sched.Reload(true)
		}
	}
	*reply = OK // This will mark saving of the account, error still can show up in actionTimingsId
	return nil
}

func (self *ApierV1) RemoveAccount(attr utils.AttrRemoveAccount, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	dirtyActionPlans := make(map[string]*engine.ActionPlan)
	accID := utils.AccountKey(attr.Tenant, attr.Account)
	_, err := engine.Guardian.Guard(func() (interface{}, error) {
		// remove it from all action plans
		_, err := engine.Guardian.Guard(func() (interface{}, error) {
			actionPlansMap, err := self.RatingDb.GetAllActionPlans()
			if err != nil {
				return 0, err
			}

			for actionPlanID, ap := range actionPlansMap {
				if _, exists := ap.AccountIDs[accID]; exists {
					delete(ap.AccountIDs, accID)
					dirtyActionPlans[actionPlanID] = ap
				}
			}

			var actionPlansCacheIds []string
			for actionPlanID, ap := range dirtyActionPlans {
				if err := self.RatingDb.SetActionPlan(actionPlanID, ap, true); err != nil {
					return 0, err
				}
				actionPlansCacheIds = append(actionPlansCacheIds, utils.ACTION_PLAN_PREFIX+actionPlanID)
			}
			if len(actionPlansCacheIds) > 0 {
				// update cache
				self.RatingDb.CacheRatingPrefixValues(map[string][]string{
					utils.ACTION_PLAN_PREFIX: actionPlansCacheIds})
			}
			return 0, nil
		}, 0, utils.ACTION_PLAN_PREFIX)
		if err != nil {
			return 0, err
		}

		if err := self.AccountDb.RemoveAccount(accID); err != nil {
			return 0, err
		}
		return 0, nil
	}, 0, accID)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if attr.ReloadScheduler && len(dirtyActionPlans) > 0 {
		// reload scheduler
		if self.Sched != nil {
			self.Sched.Reload(true)
		}
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
		if accountKeys, err = self.AccountDb.GetKeysForPrefix(utils.ACCOUNT_PREFIX + attr.Tenant); err != nil {
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
		if acnt, err := self.AccountDb.GetAccount(acntKey[len(utils.ACCOUNT_PREFIX):]); err != nil && err != utils.ErrNotFound { // Not found is not an error here
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
	userBalance, err := self.AccountDb.GetAccount(tag)
	if err != nil {
		return err
	}

	*reply = userBalance.AsOldStructure()
	return nil
}

type AttrAddBalance struct {
	Tenant         string
	Account        string
	BalanceUuid    string
	BalanceId      string
	BalanceType    string
	Directions     string
	Value          float64
	ExpiryTime     string
	RatingSubject  string
	Categories     string
	DestinationIds string
	TimingIds      string
	Weight         float64
	SharedGroups   string
	Overwrite      bool // When true it will reset if the balance is already there
	Blocker        bool
	Disabled       bool
}

func (self *ApierV1) AddBalance(attr *AttrAddBalance, reply *string) error {
	return self.modifyBalance(engine.TOPUP, attr, reply)
}
func (self *ApierV1) DebitBalance(attr *AttrAddBalance, reply *string) error {
	return self.modifyBalance(engine.DEBIT, attr, reply)
}

func (self *ApierV1) modifyBalance(aType string, attr *AttrAddBalance, reply *string) error {
	if missing := utils.MissingStructFields(attr, []string{"Tenant", "Account", "BalanceType"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	expTime, err := utils.ParseTimeDetectLayout(attr.ExpiryTime, self.Config.DefaultTimezone)
	if err != nil {
		*reply = err.Error()
		return err
	}
	accID := utils.AccountKey(attr.Tenant, attr.Account)
	if _, err := self.AccountDb.GetAccount(accID); err != nil {
		// create account if not exists
		account := &engine.Account{
			Id: accID,
		}
		if err := self.AccountDb.SetAccount(account); err != nil {
			*reply = err.Error()
			return err
		}
	}
	at := &engine.ActionTiming{}
	at.SetAccountIDs(utils.StringMap{accID: true})

	if attr.Overwrite {
		aType += "_reset" // => *topup_reset/*debit_reset
	}
	at.SetActions(engine.Actions{
		&engine.Action{
			ActionType:  aType,
			BalanceType: attr.BalanceType,
			Balance: &engine.Balance{
				Uuid:           attr.BalanceUuid,
				Id:             attr.BalanceId,
				Value:          attr.Value,
				ExpirationDate: expTime,
				RatingSubject:  attr.RatingSubject,
				Directions:     utils.ParseStringMap(attr.Directions),
				DestinationIds: utils.ParseStringMap(attr.DestinationIds),
				Categories:     utils.ParseStringMap(attr.Categories),
				Weight:         attr.Weight,
				SharedGroups:   utils.ParseStringMap(attr.SharedGroups),
				TimingIDs:      utils.ParseStringMap(attr.TimingIds),
				Blocker:        attr.Blocker,
				Disabled:       attr.Disabled,
			},
		},
	})
	if err := at.Execute(); err != nil {
		*reply = err.Error()
		return err
	}
	*reply = OK
	return nil
}

func (self *ApierV1) EnableDisableBalance(attr *AttrAddBalance, reply *string) error {
	if missing := utils.MissingStructFields(attr, []string{"Tenant", "Account", "BalanceType"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	expTime, err := utils.ParseTimeDetectLayout(attr.ExpiryTime, self.Config.DefaultTimezone)
	if err != nil {
		*reply = err.Error()
		return err
	}
	accID := utils.ConcatenatedKey(attr.Tenant, attr.Account)
	if _, err := self.AccountDb.GetAccount(accID); err != nil {
		return utils.ErrNotFound
	}
	at := &engine.ActionTiming{}
	at.SetAccountIDs(utils.StringMap{accID: true})

	at.SetActions(engine.Actions{
		&engine.Action{
			ActionType:  engine.ENABLE_DISABLE_BALANCE,
			BalanceType: attr.BalanceType,
			Balance: &engine.Balance{
				Uuid:           attr.BalanceUuid,
				Id:             attr.BalanceId,
				Value:          attr.Value,
				ExpirationDate: expTime,
				RatingSubject:  attr.RatingSubject,
				Categories:     utils.ParseStringMap(attr.Categories),
				Directions:     utils.ParseStringMap(attr.Directions),
				DestinationIds: utils.ParseStringMap(attr.DestinationIds),
				Weight:         attr.Weight,
				SharedGroups:   utils.ParseStringMap(attr.SharedGroups),
				TimingIDs:      utils.ParseStringMap(attr.TimingIds),
				Blocker:        attr.Blocker,
				Disabled:       attr.Disabled,
			},
		},
	})
	if err := at.Execute(); err != nil {
		*reply = err.Error()
		return err
	}
	*reply = OK
	return nil
}

func (self *ApierV1) RemoveBalances(attr *AttrAddBalance, reply *string) error {
	if missing := utils.MissingStructFields(attr, []string{"Tenant", "Account", "BalanceType"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	expTime, err := utils.ParseTimeDetectLayout(attr.ExpiryTime, self.Config.DefaultTimezone)
	if err != nil {
		*reply = err.Error()
		return err
	}
	accID := utils.AccountKey(attr.Tenant, attr.Account)
	if _, err := self.AccountDb.GetAccount(accID); err != nil {
		return utils.ErrNotFound
	}

	at := &engine.ActionTiming{}
	at.SetAccountIDs(utils.StringMap{accID: true})
	at.SetActions(engine.Actions{
		&engine.Action{
			ActionType:  engine.REMOVE_BALANCE,
			BalanceType: attr.BalanceType,
			Balance: &engine.Balance{
				Uuid:           attr.BalanceUuid,
				Id:             attr.BalanceId,
				Value:          attr.Value,
				ExpirationDate: expTime,
				RatingSubject:  attr.RatingSubject,
				Directions:     utils.ParseStringMap(attr.Directions),
				DestinationIds: utils.ParseStringMap(attr.DestinationIds),
				Categories:     utils.ParseStringMap(attr.Categories),
				Weight:         attr.Weight,
				SharedGroups:   utils.ParseStringMap(attr.SharedGroups),
				TimingIDs:      utils.ParseStringMap(attr.TimingIds),
				Blocker:        attr.Blocker,
				Disabled:       attr.Disabled,
			},
		},
	})
	if err := at.Execute(); err != nil {
		*reply = err.Error()
		return err
	}
	*reply = OK
	return nil
}

type AttrSetBalance struct {
	Tenant         string
	Account        string
	BalanceType    string
	BalanceUUID    *string
	BalanceID      *string
	Directions     *[]string
	Value          *float64
	ExpiryTime     *string
	RatingSubject  *string
	Categories     *[]string
	DestinationIDs *[]string
	SharedGroups   *[]string
	TimingIDs      *[]string
	Weight         *float64
	Blocker        *bool
	Disabled       *bool
	expTime        time.Time
}

func (attr *AttrSetBalance) SetBalance(b *engine.Balance) {
	if b == nil {
		return
	}
	if attr.Directions != nil {
		b.Directions = utils.StringMapFromSlice(*attr.Directions)
	}
	if attr.Value != nil {
		b.Value = *attr.Value
	}
	if attr.ExpiryTime != nil {
		b.ExpirationDate = attr.expTime
	}
	if attr.RatingSubject != nil {
		b.RatingSubject = *attr.RatingSubject
	}
	if attr.Categories != nil {
		b.Categories = utils.StringMapFromSlice(*attr.Categories)
	}
	if attr.DestinationIDs != nil {
		b.DestinationIds = utils.StringMapFromSlice(*attr.DestinationIDs)
	}
	if attr.SharedGroups != nil {
		b.SharedGroups = utils.StringMapFromSlice(*attr.SharedGroups)
	}
	if attr.TimingIDs != nil {
		b.TimingIDs = utils.StringMapFromSlice(*attr.TimingIDs)
	}
	if attr.Weight != nil {
		b.Weight = *attr.Weight
	}
	if attr.Blocker != nil {
		b.Blocker = *attr.Blocker
	}
	if attr.Disabled != nil {
		b.Disabled = *attr.Disabled
	}
	b.SetDirty() // Mark the balance as dirty since we have modified and it should be checked by action triggers
}

/* // SetAccount api using action and action timing to set balance,
//to be uncommented when using pointers in action.balance
func (self *ApierV1) SetBalance(attr *AttrAddBalance, reply *string) error {
	if missing := utils.MissingStructFields(attr, []string{"Tenant", "Account", "BalanceType"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attr.BalanceID == "" && attr.BalanceUUID == "" {
		return utils.NewErrMandatoryIeMissing("BalanceID", "or", "BalanceUUID")
	}
	expTime, err := utils.ParseTimeDetectLayout(attr.ExpiryTime, self.Config.DefaultTimezone)
	if err != nil {
		*reply = err.Error()
		return err
	}
	accID := utils.ConcatenatedKey(attr.Tenant, attr.Account)
	if _, err := self.AccountDb.GetAccount(accID); err != nil {
		return utils.ErrNotFound
	}

	at := &engine.ActionTiming{}
	at.SetAccountIDs(utils.StringMap{accID: true})

	at.SetActions(engine.Actions{
		&engine.Action{
			ActionType:  engine.SET_BALANCE,
			BalanceType: attr.BalanceType,
			Balance: &engine.Balance{
				Uuuid:           attr.BalanceUUID,
				ID:             attr.BalanceID,
				Value:          attr.Value,
				ExpirationDate: expTime,
				RatingSubject:  attr.RatingSubject,
				Directions:     utils.ParseStringMap(attr.Directions),
				DestinationIDs: utils.ParseStringMap(attr.DestinationIDs),
				Categories:     utils.ParseStringMap(attr.Categories),
				Weight:         attr.Weight,
				SharedGroups:   utils.ParseStringMap(attr.SharedGroups),
                TimingIDs:      utils.ParseStringMap(attr.TimingIDs),
				Blocker:        true,
				Disabled:       attr.Disabled,
			},
		},
	})
	if err := at.Execute(); err != nil {
		*reply = err.Error()
		return err
	}
	*reply = OK
	return nil
}
*/

func (self *ApierV1) SetBalance(attr *AttrSetBalance, reply *string) error {
	if missing := utils.MissingStructFields(attr, []string{"Tenant", "Account", "BalanceType"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if (attr.BalanceID == nil || *attr.BalanceID == "") &&
		(attr.BalanceUUID == nil || *attr.BalanceUUID == "") {
		return utils.NewErrMandatoryIeMissing("BalanceID", "or", "BalanceUUID")
	}
	var err error
	if attr.ExpiryTime != nil {
		attr.expTime, err = utils.ParseTimeDetectLayout(*attr.ExpiryTime, self.Config.DefaultTimezone)
		if err != nil {
			*reply = err.Error()
			return err
		}
	}
	accID := utils.ConcatenatedKey(attr.Tenant, attr.Account)
	_, err = engine.Guardian.Guard(func() (interface{}, error) {
		account, err := self.AccountDb.GetAccount(accID)
		if err != nil {
			return 0, utils.ErrNotFound
		}

		if account.BalanceMap == nil {
			account.BalanceMap = make(map[string]engine.BalanceChain, 1)
		}
		var previousSharedGroups utils.StringMap // kept for comparison
		var balance *engine.Balance
		var found bool
		for _, b := range account.BalanceMap[attr.BalanceType] {
			if b.IsExpired() {
				continue
			}
			if (attr.BalanceUUID != nil && b.Uuid == *attr.BalanceUUID) ||
				(attr.BalanceID != nil && b.Id == *attr.BalanceID) {
				previousSharedGroups = b.SharedGroups
				balance = b
				found = true
				break // only set one balance
			}
		}

		// if it is not found then we add it to the list
		if balance == nil {
			balance = &engine.Balance{}
			balance.Uuid = utils.GenUUID() // alway overwrite the uuid for consistency
			account.BalanceMap[attr.BalanceType] = append(account.BalanceMap[attr.BalanceType], balance)
		}

		if attr.BalanceID != nil && *attr.BalanceID == utils.META_DEFAULT {
			balance.Id = utils.META_DEFAULT
			if attr.Value != nil {
				balance.Value = *attr.Value
			}
		} else {
			attr.SetBalance(balance)
		}

		if !found || !previousSharedGroups.Equal(balance.SharedGroups) {
			_, err = engine.Guardian.Guard(func() (interface{}, error) {
				for sgID := range balance.SharedGroups {
					// add shared group member
					sg, err := self.RatingDb.GetSharedGroup(sgID, false)
					if err != nil || sg == nil {
						//than is problem
						utils.Logger.Warning(fmt.Sprintf("Could not get shared group: %v", sgID))
					} else {
						if _, found := sg.MemberIds[account.Id]; !found {
							// add member and save
							if sg.MemberIds == nil {
								sg.MemberIds = make(utils.StringMap)
							}
							sg.MemberIds[account.Id] = true
							self.RatingDb.SetSharedGroup(sg)
						}
					}
				}
				return 0, nil
			}, 0, balance.SharedGroups.Slice()...)
		}

		account.InitCounters()
		account.ExecuteActionTriggers(nil)
		self.AccountDb.SetAccount(account)
		return 0, nil
	}, 0, accID)
	if err != nil {
		*reply = err.Error()
		return err
	}
	*reply = utils.OK
	return nil
}
