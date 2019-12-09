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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// NewSchedulerSv1 retuns the API for SchedulerS
func NewSchedulerSv1(cgrcfg *config.CGRConfig) *SchedulerSv1 {
	return &SchedulerSv1{cgrcfg: cgrcfg}
}

// SchedulerSv1 is the RPC object implementing scheduler APIs
type SchedulerSv1 struct {
	cgrcfg *config.CGRConfig
}

// Reload reloads scheduler instructions
func (schdSv1 *SchedulerSv1) Reload(arg *utils.CGREventWithArgDispatcher, reply *string) error {
	schdSv1.cgrcfg.GetReloadChan(config.SCHEDULER_JSN) <- struct{}{}
	*reply = utils.OK
	return nil
}

// Ping returns Pong
func (schdSv1 *SchedulerSv1) Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error {
	*reply = utils.Pong
	return nil
}

// Call implements rpcclient.ClientConnector interface for internal RPC
func (schdSv1 *SchedulerSv1) Call(serviceMethod string,
	args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(schdSv1, serviceMethod, args, reply)
}
