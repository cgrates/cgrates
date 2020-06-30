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

func (apierSv1 *APIerSv1) GetScheduledActions(args *scheduler.ArgsGetScheduledActions, reply *[]*scheduler.ScheduledAction) error {
	sched := apierSv1.SchedulerService.GetScheduler()
	if sched == nil {
		return errors.New(utils.SchedulerNotRunningCaps)
	}
	rpl := sched.GetScheduledActions(*args)
	if len(rpl) == 0 {
		return utils.ErrNotFound
	}
	*reply = rpl
	return nil
}
