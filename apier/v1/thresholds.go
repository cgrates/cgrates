/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNEtS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package v1

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewThresholdSV1 initializes ThresholdSV1
func NewThresholdSV1(tS *engine.ThresholdService) *ThresholdSV1 {
	return &ThresholdSV1{tS: tS}
}

// Exports RPC from RLs
type ThresholdSV1 struct {
	tS *engine.ThresholdService
}

// Call implements rpcclient.RpcClientConnection interface for internal RPC
func (tSv1 *ThresholdSV1) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(tSv1, serviceMethod, args, reply)
}

// GetThresholdIDs returns list of threshold IDs registered for a tenant
func (tSv1 *ThresholdSV1) GetThresholdIDs(tenant string, tIDs *[]string) error {
	return tSv1.tS.V1GetThresholdIDs(tenant, tIDs)
}

// GetThresholdsForEvent returns a list of thresholds matching an event
func (tSv1 *ThresholdSV1) GetThresholdsForEvent(ev *engine.ThresholdEvent, reply *engine.Thresholds) error {
	return tSv1.tS.V1GetThresholdsForEvent(ev, reply)
}

// ProcessEvent will process an Event
func (tSv1 *ThresholdSV1) ProcessEvent(ev *engine.ThresholdEvent, reply *string) error {
	return tSv1.tS.V1ProcessEvent(ev, reply)
}
