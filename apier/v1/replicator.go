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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewReplicatorSv1(dm *engine.DataManager) *ReplicatorSv1 {
	return &ReplicatorSv1{dm: dm}
}

// Exports RPC
type ReplicatorSv1 struct {
	dm *engine.DataManager
}

// Call implements rpcclient.RpcClientConnection interface for internal RPC
func (rplSv1 *ReplicatorSv1) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(rplSv1, serviceMethod, args, reply)
}

// SetThresholdProfile alters/creates a ThresholdProfile
func (rplSv1 *ReplicatorSv1) SetThresholdProfile(th *engine.ThresholdProfile, reply *string) error {
	if err := rplSv1.dm.DataDB().SetThresholdProfileDrv(th); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error {
	*reply = utils.Pong
	return nil
}
