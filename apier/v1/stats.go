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
	"github.com/cgrates/cgrates/cache"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// GetStatQueueProfile returns a StatQueue profile
func (apierV1 *ApierV1) GetStatQueueProfile(arg *utils.TenantID, reply *engine.StatQueueProfile) (err error) {
	if missing := utils.MissingStructFields(arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if sCfg, err := apierV1.DataManager.GetStatQueueProfile(arg.Tenant, arg.ID,
		false, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	} else {
		*reply = *sCfg
	}
	return
}

// SetStatQueueProfile alters/creates a StatQueueProfile
func (apierV1 *ApierV1) SetStatQueueProfile(sqp *engine.StatQueueProfile, reply *string) error {
	if missing := utils.MissingStructFields(sqp, []string{"Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierV1.DataManager.DataDB().SetStatQueueProfile(sqp); err != nil {
		return utils.APIErrorHandler(err)
	}
	cache.RemKey(utils.StatQueueProfilePrefix+utils.ConcatenatedKey(sqp.Tenant, sqp.ID),
		true, utils.NonTransactional) // Temporary work around util proper cacheDataFromDB will be implemented
	*reply = utils.OK
	return nil
}

// Remove a specific stat configuration
func (apierV1 *ApierV1) RemStatQueueProfile(args *utils.TenantID, reply *string) error {
	if missing := utils.MissingStructFields(args, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierV1.DataManager.DataDB().RemStatQueueProfile(args.Tenant, args.ID, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// NewStatSV1 initializes StatSV1
func NewStatSV1(sS *engine.StatService) *StatSV1 {
	return &StatSV1{sS: sS}
}

// Exports RPC from RLs
type StatSV1 struct {
	sS *engine.StatService
}

// Call implements rpcclient.RpcClientConnection interface for internal RPC
func (stsv1 *StatSV1) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(stsv1, serviceMethod, args, reply)
}

// GetQueueIDs returns list of queueIDs registered for a tenant
func (stsv1 *StatSV1) GetQueueIDs(tenant string, qIDs *[]string) error {
	return stsv1.sS.V1GetQueueIDs(tenant, qIDs)
}

// ProcessEvent returns processes a new Event
func (stsv1 *StatSV1) ProcessEvent(ev *engine.StatEvent, reply *string) error {
	return stsv1.sS.V1ProcessEvent(ev, reply)
}

// GetQueueIDs returns the list of queues IDs in the system
func (stsv1 *StatSV1) GetStatQueuesForEvent(ev *engine.StatEvent, reply *engine.StatQueues) (err error) {
	return stsv1.sS.V1GetStatQueuesForEvent(ev, reply)
}

// GetStringMetrics returns the string metrics for a Queue
func (stsv1 *StatSV1) GetQueueStringMetrics(args *utils.TenantID, reply *map[string]string) (err error) {
	return stsv1.sS.V1GetQueueStringMetrics(args, reply)
}

// GetQueueFloatMetrics returns the float metrics for a Queue
func (stsv1 *StatSV1) GetQueueFloatMetrics(args *utils.TenantID, reply *map[string]float64) (err error) {
	return stsv1.sS.V1GetQueueFloatMetrics(args, reply)
}
