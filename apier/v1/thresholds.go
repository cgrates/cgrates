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
func NewThresholdSv1(tS *engine.ThresholdService) *ThresholdSv1 {
	return &ThresholdSv1{tS: tS}
}

// Exports RPC from RLs
type ThresholdSv1 struct {
	tS *engine.ThresholdService
}

// Call implements rpcclient.RpcClientConnection interface for internal RPC
func (tSv1 *ThresholdSv1) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(tSv1, serviceMethod, args, reply)
}

// GetThresholdIDs returns list of threshold IDs registered for a tenant
func (tSv1 *ThresholdSv1) GetThresholdIDs(tenant string, tIDs *[]string) error {
	return tSv1.tS.V1GetThresholdIDs(tenant, tIDs)
}

// GetThresholdsForEvent returns a list of thresholds matching an event
func (tSv1 *ThresholdSv1) GetThresholdsForEvent(args *engine.ArgsProcessEvent, reply *engine.Thresholds) error {
	return tSv1.tS.V1GetThresholdsForEvent(args, reply)
}

// GetThreshold queries a Threshold
func (tSv1 *ThresholdSv1) GetThreshold(tntID *utils.TenantID, t *engine.Threshold) error {
	return tSv1.tS.V1GetThreshold(tntID, t)
}

// ProcessEvent will process an Event
func (tSv1 *ThresholdSv1) ProcessEvent(args *engine.ArgsProcessEvent, hits *int) error {
	return tSv1.tS.V1ProcessEvent(args, hits)
}

// GetThresholdProfile returns a Threshold Profile
func (apierV1 *ApierV1) GetThresholdProfile(arg *utils.TenantID, reply *engine.ThresholdProfile) (err error) {
	if missing := utils.MissingStructFields(arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if th, err := apierV1.DataManager.GetThresholdProfile(arg.Tenant, arg.ID, false, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	} else {
		*reply = *th
	}
	return
}

// SetThresholdProfile alters/creates a ThresholdProfile
func (apierV1 *ApierV1) SetThresholdProfile(thp *engine.ThresholdProfile, reply *string) error {
	if missing := utils.MissingStructFields(thp, []string{"Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierV1.DataManager.SetThresholdProfile(thp, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := apierV1.DataManager.SetThreshold(&engine.Threshold{Tenant: thp.Tenant, ID: thp.ID}); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// Remove a specific Threshold Profile
func (apierV1 *ApierV1) RemoveThresholdProfile(args *utils.TenantID, reply *string) error {
	if missing := utils.MissingStructFields(args, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierV1.DataManager.RemoveThresholdProfile(args.Tenant, args.ID, utils.NonTransactional, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := apierV1.DataManager.RemoveThreshold(args.Tenant, args.ID, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}
