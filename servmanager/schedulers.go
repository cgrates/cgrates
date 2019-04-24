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

package servmanager

import (
	"errors"

	"github.com/cgrates/cgrates/utils"
)

func NewSchedulerS(srvMngr *ServiceManager) *SchedulerS {
	return &SchedulerS{srvMngr: srvMngr}
}

type SchedulerS struct {
	srvMngr *ServiceManager // access scheduler from servmanager so we can dynamically start/stop
}

// Call gives the ability of SchedulerS to be passed as internal RPC
func (schdS *SchedulerS) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.RPCCall(schdS, serviceMethod, args, reply)
}

// V1ReloadScheduler reloads the scheduler tasks
func (schdS *SchedulerS) V1Reload(_ *utils.CGREventWithArgDispatcher, reply *string) (err error) {
	sched := schdS.srvMngr.GetScheduler()
	if sched == nil {
		return errors.New(utils.SchedulerNotRunningCaps)
	}
	sched.Reload()
	*reply = utils.OK
	return nil
}
