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

package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

type ActionTrigger struct {
	ID             string // original csv tag
	UniqueID       string // individual id
	ThresholdType  string //*min_event_counter, *max_event_counter, *min_balance_counter, *max_balance_counter, *min_balance, *max_balance, *balance_expired
	ThresholdValue float64
	Recurrent      bool          // reset excuted flag each run
	MinSleep       time.Duration // Minimum duration between two executions in case of recurrent triggers
	ExpirationDate time.Time
	ActivationDate time.Time
	//BalanceType       string // *monetary/*voice etc
	Balance           *BalanceFilter
	Weight            float64
	ActionsID         string
	MinQueuedItems    int // Trigger actions only if this number is hit (stats only)
	Executed          bool
	LastExecutionTime time.Time
}

func (at *ActionTrigger) Execute(acc *Account, fltrS *FilterS) (err error) {
	// check for min sleep time
	if at.Recurrent && !at.LastExecutionTime.IsZero() && time.Since(at.LastExecutionTime) < at.MinSleep {
		return
	}
	at.LastExecutionTime = time.Now()
	if acc != nil && acc.Disabled {
		return fmt.Errorf("User %s is disabled and there are triggers in action!", acc.ID)
	}
	// does NOT need to Lock() because it is triggered from a method that took the Lock
	var acts Actions
	acts, err = dm.GetActions(at.ActionsID, false, utils.NonTransactional)
	if err != nil {
		utils.Logger.Err(
			fmt.Sprintf("Failed to get actions: %v",
				err))
		return
	}
	acts.Sort()
	at.Executed = true
	transactionFailed := false
	removeAccountActionFound := false
	sharedData := NewSharedActionsData(acts)
	for i, act := range acts {
		// check action filter
		if len(act.Filters) > 0 {
			if pass, err := fltrS.Pass(utils.NewTenantID(act.Id).Tenant, act.Filters,
				utils.MapStorage{utils.MetaReq: acc}); err != nil {
				return err
			} else if !pass {
				continue
			}
		}
		if act.Balance == nil {
			act.Balance = &BalanceFilter{}
		}
		if act.ExpirationString != "" { // if it's *unlimited then it has to be zero time'
			if expDate, parseErr := utils.ParseTimeDetectLayout(act.ExpirationString,
				config.CgrConfig().GeneralCfg().DefaultTimezone); parseErr == nil {
				act.Balance.ExpirationDate = &time.Time{}
				*act.Balance.ExpirationDate = expDate
			}
		}

		actionFunction, exists := getActionFunc(act.ActionType)
		if !exists {
			utils.Logger.Err(
				fmt.Sprintf("Function type %v not available, aborting execution!",
					act.ActionType))
			transactionFailed = false
			break
		}
		sharedData.idx = i // set the current action index in shared data
		if err := actionFunction(acc, act, acts, fltrS, nil, sharedData,
			newActionConnCfg(utils.RALs, act.ActionType, config.CgrConfig())); err != nil {
			utils.Logger.Err(
				fmt.Sprintf("Error executing action %s: %v!",
					act.ActionType, err))
			transactionFailed = false
			break
		}
		if act.ActionType == utils.MetaRemoveAccount {
			removeAccountActionFound = true
		}
	}
	if transactionFailed || at.Recurrent {
		at.Executed = false
	}
	if !transactionFailed && acc != nil && !removeAccountActionFound {
		dm.SetAccount(acc)
	}
	return
}

// returns true if the field of the action timing are equeal to the non empty
// fields of the action
func (at *ActionTrigger) Match(a *Action) bool {
	if a == nil || a.Balance == nil {
		return true
	}
	if a.Balance.Type != nil && a.Balance.GetType() != at.Balance.GetType() {
		return false
	}
	thresholdType := true // by default we consider that we don't have ExtraParameters
	if a.ExtraParameters != "" {
		t := struct {
			GroupID       string
			UniqueID      string
			ThresholdType string
		}{}
		json.Unmarshal([]byte(a.ExtraParameters), &t)
		// check Ids first
		if t.GroupID != "" {
			return at.ID == t.GroupID
		}
		if t.UniqueID != "" {
			return at.UniqueID == t.UniqueID
		}
		thresholdType = t.ThresholdType == "" || at.ThresholdType == t.ThresholdType
	}
	return thresholdType && at.Balance.CreateBalance().MatchFilter(a.Balance, "", false, false)
}

func (at *ActionTrigger) CreateBalance() *Balance {
	b := at.Balance.CreateBalance()
	b.ID = at.UniqueID
	return b
}

// makes a shallow copy of the receiver
func (at *ActionTrigger) Clone() *ActionTrigger {
	if at == nil {
		return nil
	}
	return &ActionTrigger{
		ID:                at.ID,
		UniqueID:          at.UniqueID,
		ThresholdType:     at.ThresholdType,
		ThresholdValue:    at.ThresholdValue,
		Recurrent:         at.Recurrent,
		MinSleep:          at.MinSleep,
		ExpirationDate:    at.ExpirationDate,
		ActivationDate:    at.ActivationDate,
		Weight:            at.Weight,
		ActionsID:         at.ActionsID,
		MinQueuedItems:    at.MinQueuedItems,
		Executed:          at.Executed,
		LastExecutionTime: at.LastExecutionTime,
		Balance:           at.Balance.Clone(),
	}
}

func (at *ActionTrigger) Equals(oat *ActionTrigger) bool {
	// ids only
	return at.ID == oat.ID && at.UniqueID == oat.UniqueID
}

func (at *ActionTrigger) IsActive(t time.Time) bool {
	return at.ActivationDate.IsZero() || t.After(at.ActivationDate)
}

func (at *ActionTrigger) IsExpired(t time.Time) bool {
	return !at.ExpirationDate.IsZero() && t.After(at.ExpirationDate)
}

// Structure to store actions according to weight
type ActionTriggers []*ActionTrigger

func (atpl ActionTriggers) Len() int {
	return len(atpl)
}

func (atpl ActionTriggers) Swap(i, j int) {
	atpl[i], atpl[j] = atpl[j], atpl[i]
}

// we need higher weights earlyer in the list
func (atpl ActionTriggers) Less(j, i int) bool {
	return atpl[i].Weight < atpl[j].Weight
}

func (atpl ActionTriggers) Sort() {
	sort.Sort(atpl)
}

func (atpl ActionTriggers) Clone() ActionTriggers {
	if atpl == nil {
		return nil
	}
	clone := make(ActionTriggers, len(atpl))
	for i, at := range atpl {
		clone[i] = at.Clone()
	}
	return clone
}

func (at *ActionTrigger) String() string {
	return utils.ToJSON(at)
}

func (at *ActionTrigger) FieldAsInterface(fldPath []string) (val any, err error) {
	if at == nil || len(fldPath) == 0 {
		return nil, utils.ErrNotFound
	}
	switch fldPath[0] {
	default:
		return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
	case utils.ID:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return at.ID, nil
	case utils.UniqueID:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return at.UniqueID, nil
	case utils.ThresholdType:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return at.ThresholdType, nil
	case utils.ThresholdValue:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return at.ThresholdValue, nil
	case utils.Recurrent:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return at.Recurrent, nil
	case utils.MinSleep:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return at.MinSleep, nil
	case utils.ExpirationDate:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return at.ExpirationDate, nil
	case utils.ActivationDate:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return at.ActivationDate, nil
	case utils.BalanceField:
		if len(fldPath) == 1 {
			return at.Balance, nil
		}
		return at.Balance.FieldAsInterface(fldPath[1:])
	case utils.Weight:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return at.Weight, nil
	case utils.ActionsID:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return at.ActionsID, nil
	case utils.MinQueuedItems:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return at.MinQueuedItems, nil
	case utils.Executed:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return at.Executed, nil
	case utils.LastExecutionTime:
		if len(fldPath) != 1 {
			return nil, utils.ErrNotFound
		}
		return at.LastExecutionTime, nil
	}
}

func (at *ActionTrigger) FieldAsString(fldPath []string) (val string, err error) {
	var iface any
	iface, err = at.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return utils.IfaceAsString(iface), nil
}

type AttrSetActionTrigger struct {
	GroupID       string
	UniqueID      string
	ActionTrigger map[string]any
}

// UpdateActionTrigger updates the ActionTrigger if is matching
func (attr *AttrSetActionTrigger) UpdateActionTrigger(at *ActionTrigger, timezone string) (updated bool, err error) {
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
		at.Balance = &BalanceFilter{}
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
		if mQ, err = utils.IfaceAsTInt64(minQ); err != nil {
			return
		}
		at.MinQueuedItems = int(mQ)
	}
	if accID, has := attr.ActionTrigger[utils.ActionsID]; has {
		at.ActionsID = utils.IfaceAsString(accID)
	}

	if weight, has := attr.ActionTrigger[utils.Weight]; has {
		if at.Weight, err = utils.IfaceAsFloat64(weight); err != nil {
			return
		}
	}
	return
}
