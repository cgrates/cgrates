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
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

/*
[
    {
        u'ActionsId': u'BONUS_1',
        u'Uuid': u'5b5ba53b40b1d44380cce52379ec5c0d',
        u'Weight': 10,
        u'Timing': {
            u'Timing': {
                u'MonthDays': [

                ],
                u'Months': [

                ],
                u'WeekDays': [

                ],
                u'Years': [
                    2013
                ],
                u'StartTime': u'11: 00: 00',
                u'EndTime': u''
            },
            u'Rating': None,
            u'Weight': 0
        },
        u'AccountIds': [
            u'*out: cgrates.org: 1001',
            u'*out: cgrates.org: 1002',
            u'*out: cgrates.org: 1003',
            u'*out: cgrates.org: 1004',
            u'*out: cgrates.org: 1005'
        ],
        u'Id': u'PREPAID_10'
    },
    {
        u'ActionsId': u'PREPAID_10',
        u'Uuid': u'b16ab12740e2e6c380ff7660e8b55528',
        u'Weight': 10,
        u'Timing': {
            u'Timing': {
                u'MonthDays': [

                ],
                u'Months': [

                ],
                u'WeekDays': [

                ],
                u'Years': [
                    2013
                ],
                u'StartTime': u'11: 00: 00',
                u'EndTime': u''
            },
            u'Rating': None,
            u'Weight': 0
        },
        u'AccountIds': [
            u'*out: cgrates.org: 1001',
            u'*out: cgrates.org: 1002',
            u'*out: cgrates.org: 1003',
            u'*out: cgrates.org: 1004',
            u'*out: cgrates.org: 1005'
        ],
        u'Id': u'PREPAID_10'
    }
]
*/

type AttrsGetScheduledActions struct {
	Tenant, Account    string
	TimeStart, TimeEnd time.Time // Filter based on next runTime
	utils.Paginator
}

type ScheduledActions struct {
	NextRunTime                               time.Time
	Accounts                                  int
	ActionsId, ActionPlanId, ActionTimingUuid string
}

func (self *ApierV1) GetScheduledActions(attrs AttrsGetScheduledActions, reply *[]*ScheduledActions) error {
	if self.Sched == nil {
		return errors.New("SCHEDULER_NOT_ENABLED")
	}
	schedActions := make([]*ScheduledActions, 0) // needs to be initialized if remains empty
	scheduledActions := self.Sched.GetQueue()
	for _, qActions := range scheduledActions {
		sas := &ScheduledActions{ActionsId: qActions.ActionsID, ActionPlanId: qActions.GetActionPlanID(), ActionTimingUuid: qActions.Uuid, Accounts: len(qActions.GetAccountIDs())}
		if attrs.SearchTerm != "" &&
			!(strings.Contains(sas.ActionPlanId, attrs.SearchTerm) ||
				strings.Contains(sas.ActionsId, attrs.SearchTerm)) {
			continue
		}
		sas.NextRunTime = qActions.GetNextStartTime(time.Now())
		if !attrs.TimeStart.IsZero() && sas.NextRunTime.Before(attrs.TimeStart) {
			continue // Filter here only requests in the filtered interval
		}
		if !attrs.TimeEnd.IsZero() && (sas.NextRunTime.After(attrs.TimeEnd) || sas.NextRunTime.Equal(attrs.TimeEnd)) {
			continue
		}
		// filter on account
		if attrs.Tenant != "" || attrs.Account != "" {
			found := false
			for accID := range qActions.GetAccountIDs() {
				split := strings.Split(accID, utils.CONCATENATED_KEY_SEP)
				if len(split) != 2 {
					continue // malformed account id
				}
				if attrs.Tenant != "" && attrs.Tenant != split[0] {
					continue
				}
				if attrs.Account != "" && attrs.Account != split[1] {
					continue
				}
				found = true
				break
			}
			if !found {
				continue
			}
		}

		// we have a winner

		schedActions = append(schedActions, sas)
	}
	if attrs.Paginator.Offset != nil {
		if *attrs.Paginator.Offset <= len(schedActions) {
			schedActions = schedActions[*attrs.Paginator.Offset:]
		}
	}
	if attrs.Paginator.Limit != nil {
		if *attrs.Paginator.Limit <= len(schedActions) {
			schedActions = schedActions[:*attrs.Paginator.Limit]
		}
	}
	*reply = schedActions
	return nil
}

type AttrsExecuteScheduledActions struct {
	ActionPlanID       string
	TimeStart, TimeEnd time.Time // replay the action timings between the two dates
}

func (self *ApierV1) ExecuteScheduledActions(attr AttrsExecuteScheduledActions, reply *string) error {
	if attr.ActionPlanID != "" { // execute by ActionPlanID
		apl, err := self.RatingDb.GetActionPlan(attr.ActionPlanID, false)
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
				err := at.Execute()
				if err != nil {
					*reply = err.Error()
					return err
				}
				utils.Logger.Info(fmt.Sprintf("<Force Scheduler> Executing action %s ", at.ActionsID))
			}
		}
	}
	if !attr.TimeStart.IsZero() && !attr.TimeEnd.IsZero() { // execute between two dates
		actionPlans, err := self.RatingDb.GetAllActionPlans()
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
				err := a0.Execute()
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
