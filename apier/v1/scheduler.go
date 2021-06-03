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
	"sort"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/scheduler"
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

func (self *APIerSv1) GetScheduledActions(args scheduler.ArgsGetScheduledActions, reply *[]*scheduler.ScheduledAction) error {
	sched := self.SchedulerService.GetScheduler()
	if sched == nil {
		return errors.New(utils.SchedulerNotRunningCaps)
	}
	rpl := sched.GetScheduledActions(args)
	if len(rpl) == 0 {
		return utils.ErrNotFound
	}
	*reply = rpl
	return nil
}

type AttrsExecuteScheduledActions struct {
	ActionPlanID       string
	TimeStart, TimeEnd time.Time // replay the action timings between the two dates
}

func (self *APIerSv1) ExecuteScheduledActions(attr AttrsExecuteScheduledActions, reply *string) error {
	if attr.ActionPlanID != "" { // execute by ActionPlanID
		apl, err := self.DataManager.GetActionPlan(attr.ActionPlanID, true, true, utils.NonTransactional)
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
				err := at.Execute(nil, nil)
				if err != nil {
					*reply = err.Error()
					return err
				}
				utils.Logger.Info(fmt.Sprintf("<Force Scheduler> Executing action %s ", at.ActionsID))
			}
		}
	}
	if !attr.TimeStart.IsZero() && !attr.TimeEnd.IsZero() { // execute between two dates
		actionPlans, err := self.DataManager.GetAllActionPlans()
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
				err := a0.Execute(nil, nil)
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
