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

// GetStatQueueProfile returns a StatQueue profile
func (apierV1 *ApierV1) GetStatQueueProfile(arg *utils.TenantID, reply *engine.StatQueueProfile) (err error) {
	if missing := utils.MissingStructFields(arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if sCfg, err := apierV1.DataManager.GetStatQueueProfile(arg.Tenant, arg.ID,
		true, true, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	} else {
		*reply = *sCfg
	}
	return
}

// GetStatQueueProfileIDs returns list of statQueueProfile IDs registered for a tenant
func (apierV1 *ApierV1) GetStatQueueProfileIDs(tenant string, stsPrfIDs *[]string) error {
	prfx := utils.StatQueueProfilePrefix + tenant + ":"
	keys, err := apierV1.DataManager.DataDB().GetKeysForPrefix(prfx)
	if err != nil {
		return err
	}
	retIDs := make([]string, len(keys))
	for i, key := range keys {
		retIDs[i] = key[len(prfx):]
	}
	*stsPrfIDs = retIDs
	return nil
}

// SetStatQueueProfile alters/creates a StatQueueProfile
func (apierV1 *ApierV1) SetStatQueueProfile(sqp *engine.StatQueueProfile, reply *string) error {
	if missing := utils.MissingStructFields(sqp, []string{"Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierV1.DataManager.SetStatQueueProfile(sqp, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	metrics := make(map[string]engine.StatMetric)
	for _, metricwithparam := range sqp.Metrics {
		if metric, err := engine.NewStatMetric(metricwithparam.MetricID, sqp.MinItems, metricwithparam.Parameters); err != nil {
			return utils.APIErrorHandler(err)
		} else {
			metrics[metricwithparam.MetricID] = metric
		}
	}
	if err := apierV1.DataManager.SetStatQueue(&engine.StatQueue{Tenant: sqp.Tenant, ID: sqp.ID, SQMetrics: metrics}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// Remove a specific stat configuration
func (apierV1 *ApierV1) RemStatQueueProfile(args *utils.TenantID, reply *string) error {
	if missing := utils.MissingStructFields(args, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierV1.DataManager.RemoveStatQueueProfile(args.Tenant, args.ID, utils.NonTransactional, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := apierV1.DataManager.RemoveStatQueue(args.Tenant, args.ID, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// NewStatSV1 initializes StatSV1
func NewStatSv1(sS *engine.StatService) *StatSv1 {
	return &StatSv1{sS: sS}
}

// Exports RPC from RLs
type StatSv1 struct {
	sS *engine.StatService
}

// Call implements rpcclient.RpcClientConnection interface for internal RPC
func (stsv1 *StatSv1) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(stsv1, serviceMethod, args, reply)
}

// GetQueueIDs returns list of queueIDs registered for a tenant
func (stsv1 *StatSv1) GetQueueIDs(tenant string, qIDs *[]string) error {
	return stsv1.sS.V1GetQueueIDs(tenant, qIDs)
}

// ProcessEvent returns processes a new Event
func (stsv1 *StatSv1) ProcessEvent(args *engine.StatsArgsProcessEvent, reply *[]string) error {
	return stsv1.sS.V1ProcessEvent(args, reply)
}

// GetQueueIDs returns the list of queues IDs in the system
func (stsv1 *StatSv1) GetStatQueuesForEvent(args *engine.StatsArgsProcessEvent, reply *[]string) (err error) {
	return stsv1.sS.V1GetStatQueuesForEvent(args, reply)
}

// GetStringMetrics returns the string metrics for a Queue
func (stsv1 *StatSv1) GetQueueStringMetrics(args *utils.TenantID, reply *map[string]string) (err error) {
	return stsv1.sS.V1GetQueueStringMetrics(args, reply)
}

// GetQueueFloatMetrics returns the float metrics for a Queue
func (stsv1 *StatSv1) GetQueueFloatMetrics(args *utils.TenantID, reply *map[string]float64) (err error) {
	return stsv1.sS.V1GetQueueFloatMetrics(args, reply)
}

func (stSv1 *StatSv1) Ping(ign string, reply *string) error {
	*reply = utils.Pong
	return nil
}
