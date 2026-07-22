// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"fmt"
	"sort"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewSchedulerSv1 retuns the API for SchedulerS
func NewSchedulerSv1(cgrcfg *config.CGRConfig, dm *engine.DataManager, fltrS *engine.FilterS) *SchedulerSv1 {
	return &SchedulerSv1{cgrcfg: cgrcfg, dm: dm, fltrS: fltrS}
}

// SchedulerSv1 is the RPC object implementing scheduler APIs
type SchedulerSv1 struct {
	cgrcfg *config.CGRConfig
	dm     *engine.DataManager
	fltrS  *engine.FilterS
}

// Reload reloads scheduler instructions
func (schdSv1 *SchedulerSv1) Reload(ctx *context.Context, arg *utils.CGREvent, reply *string) error {
	schdSv1.cgrcfg.GetReloadChan(config.SCHEDULER_JSN) <- struct{}{}
	*reply = utils.OK
	return nil
}

// ExecuteActions execute an actionPlan or multiple actionsPlans between a time interval
func (schdSv1 *SchedulerSv1) ExecuteActions(ctx *context.Context, attr *utils.AttrsExecuteActions, reply *string) error {
	if attr.ActionPlanID != utils.EmptyString { // execute by ActionPlanID
		apl, err := schdSv1.dm.GetActionPlan(attr.ActionPlanID, true, true, utils.NonTransactional)
		if err != nil {
			*reply = err.Error()
			return err
		}
		if apl != nil {
			// order by weight
			engine.ActionTimingWeightOnlyPriorityList(apl.ActionTimings).Sort()
			for _, at := range apl.ActionTimings {
				if at.IsASAP() {
					continue
				}

				at.SetAccountIDs(apl.AccountIDs) // copy the accounts
				at.SetActionPlanID(apl.Id)
				err := at.Execute(schdSv1.fltrS, utils.SchedulerS, nil)
				if err != nil {
					*reply = err.Error()
					return err
				}
				utils.Logger.Info(fmt.Sprintf("<Force Scheduler> Executing action %s ", at.ActionsID))
			}
		}
	}
	if !attr.TimeStart.IsZero() && !attr.TimeEnd.IsZero() { // execute between two dates
		actionPlans, err := schdSv1.dm.GetAllActionPlans()
		if err != nil && err != utils.ErrNotFound {
			err := fmt.Errorf("cannot get action plans: %v", err)
			*reply = err.Error()
			return err
		}

		// recreate the queue
		queue := engine.ActionTimingPriorityList{}
		for _, actionPlan := range actionPlans {
			for _, at := range actionPlan.ActionTimings {
				if at.Timing == nil {
					continue
				}
				if at.IsASAP() {
					continue
				}
				if at.GetNextStartTime(attr.TimeStart).Before(attr.TimeStart) {
					// the task is obsolete, do not add it to the queue
					continue
				}
				at.SetAccountIDs(actionPlan.AccountIDs) // copy the accounts
				at.SetActionPlanID(actionPlan.Id)
				at.ResetStartTimeCache()
				queue = append(queue, at)
			}
		}
		sort.Sort(queue)
		// start playback execution loop
		current := attr.TimeStart
		for len(queue) > 0 && current.Before(attr.TimeEnd) {
			a0 := queue[0]
			current = a0.GetNextStartTime(current)
			if current.Before(attr.TimeEnd) || current.Equal(attr.TimeEnd) {
				utils.Logger.Info(fmt.Sprintf("<Replay Scheduler> Executing action %s for time %v", a0.ActionsID, current))
				err := a0.Execute(schdSv1.fltrS, utils.SchedulerS, nil)
				if err != nil {
					*reply = err.Error()
					return err
				}
				// if after execute the next start time is in the past then
				// do not add it to the queue
				a0.ResetStartTimeCache()
				current = current.Add(time.Second)
				start := a0.GetNextStartTime(current)
				if start.Before(current) || start.After(attr.TimeEnd) {
					queue = queue[1:]
				} else {
					queue = append(queue, a0)
					queue = queue[1:]
					sort.Sort(queue)
				}
			}
		}
	}
	*reply = utils.OK
	return nil
}

// ExecuteActionPlans execute multiple actionPlans one by one
func (schdSv1 *SchedulerSv1) ExecuteActionPlans(ctx *context.Context, attr *utils.AttrsExecuteActionPlans, reply *string) (err error) {
	// try get account
	// if not exist set in DM
	accID := utils.ConcatenatedKey(attr.Tenant, attr.AccountID)
	if _, err = schdSv1.dm.GetAccount(accID); err != nil {
		// create account if does not exist
		account := &engine.Account{
			ID: accID,
		}
		if err = schdSv1.dm.SetAccount(account); err != nil {
			return
		}
	}
	for _, apID := range attr.ActionPlanIDs {
		apl, err := schdSv1.dm.GetActionPlan(apID, true, true, utils.NonTransactional)
		if err != nil {
			*reply = err.Error()
			return err
		}
		if apl != nil {
			// order by weight
			engine.ActionTimingWeightOnlyPriorityList(apl.ActionTimings).Sort()
			for _, at := range apl.ActionTimings {
				at.SetAccountIDs(utils.NewStringMap(accID))
				err := at.Execute(schdSv1.fltrS, utils.SchedulerS, nil)
				if err != nil {
					*reply = err.Error()
					return err
				}
				utils.Logger.Info(fmt.Sprintf("<Force Scheduler> Executing action %s ", at.ActionsID))
			}
		}
	}

	*reply = utils.OK
	return nil
}
