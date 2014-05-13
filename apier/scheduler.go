/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

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

package apier

import (
	"errors"
	"github.com/cgrates/cgrates/utils"
	"time"
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
	Direction, Tenant, Account string
	TimeStart, TimeEnd         time.Time // Filter based on next runTime
}

type ScheduledActions struct {
	NextRunTime                             time.Time
	Accounts                                []*utils.DirectionTenantAccount
	ActionsId, ActionPlanId, ActionPlanUuid string
}

func (self *ApierV1) GetScheduledActions(attrs AttrsGetScheduledActions, reply *[]*ScheduledActions) error {
	schedActions := make([]*ScheduledActions, 0)
	if self.Sched == nil {
		return errors.New("SCHEDULER_NOT_ENABLED")
	}
	for _, qActions := range self.Sched.GetQueue() {
		sas := &ScheduledActions{ActionsId: qActions.ActionsId, ActionPlanId: qActions.Id, ActionPlanUuid: qActions.Uuid}
		sas.NextRunTime = qActions.GetNextStartTime(time.Now())
		if !attrs.TimeStart.IsZero() && sas.NextRunTime.Before(attrs.TimeStart) {
			continue // Filter here only requests in the filtered interval
		}
		if !attrs.TimeEnd.IsZero() && (sas.NextRunTime.After(attrs.TimeEnd) || sas.NextRunTime.Equal(attrs.TimeEnd)) {
			continue
		}
		acntFiltersMatch := false
		for _, acntKey := range qActions.AccountIds {
			directionMatched := len(attrs.Direction) == 0
			tenantMatched := len(attrs.Tenant) == 0
			accountMatched := len(attrs.Account) == 0
			dta, _ := utils.NewDTAFromAccountKey(acntKey)
			sas.Accounts = append(sas.Accounts, dta)
			// One member matching
			if !directionMatched && attrs.Direction == dta.Direction {
				directionMatched = true
			}
			if !tenantMatched && attrs.Tenant == dta.Tenant {
				tenantMatched = true
			}
			if !accountMatched && attrs.Account == dta.Account {
				accountMatched = true
			}
			if directionMatched && tenantMatched && accountMatched {
				acntFiltersMatch = true
			}
		}
		if !acntFiltersMatch {
			continue
		}
		schedActions = append(schedActions, sas)
	}
	*reply = schedActions
	return nil
}
