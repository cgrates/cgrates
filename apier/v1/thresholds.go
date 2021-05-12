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
	"time"

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

// Call implements rpcclient.ClientConnector interface for internal RPC
func (tSv1 *ThresholdSv1) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(tSv1, serviceMethod, args, reply)
}

// GetThresholdIDs returns list of threshold IDs registered for a tenant
func (tSv1 *ThresholdSv1) GetThresholdIDs(tenant *utils.TenantWithArgDispatcher, tIDs *[]string) error {
	return tSv1.tS.V1GetThresholdIDs(tenant.Tenant, tIDs)
}

// GetThresholdsForEvent returns a list of thresholds matching an event
func (tSv1 *ThresholdSv1) GetThresholdsForEvent(args *engine.ArgsProcessEvent, reply *engine.Thresholds) error {
	return tSv1.tS.V1GetThresholdsForEvent(args, reply)
}

// GetThreshold queries a Threshold
func (tSv1 *ThresholdSv1) GetThreshold(tntID *utils.TenantIDWithArgDispatcher, t *engine.Threshold) error {
	return tSv1.tS.V1GetThreshold(tntID.TenantID, t)
}

// ProcessEvent will process an Event
func (tSv1 *ThresholdSv1) ProcessEvent(args *engine.ArgsProcessEvent, tIDs *[]string) error {
	return tSv1.tS.V1ProcessEvent(args, tIDs)
}

// GetThresholdProfile returns a Threshold Profile
func (APIerSv1 *APIerSv1) GetThresholdProfile(arg *utils.TenantID, reply *engine.ThresholdProfile) (err error) {
	if missing := utils.MissingStructFields(arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if th, err := APIerSv1.DataManager.GetThresholdProfile(arg.Tenant, arg.ID, true, true, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	} else {
		*reply = *th
	}
	return
}

// GetThresholdProfileIDs returns list of thresholdProfile IDs registered for a tenant
func (APIerSv1 *APIerSv1) GetThresholdProfileIDs(args utils.TenantArgWithPaginator, thPrfIDs *[]string) error {
	if missing := utils.MissingStructFields(&args, []string{utils.Tenant}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	prfx := utils.ThresholdProfilePrefix + args.Tenant + ":"
	keys, err := APIerSv1.DataManager.DataDB().GetKeysForPrefix(prfx)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	retIDs := make([]string, len(keys))
	for i, key := range keys {
		retIDs[i] = key[len(prfx):]
	}
	*thPrfIDs = args.PaginateStringSlice(retIDs)
	return nil
}

// SetThresholdProfile alters/creates a ThresholdProfile
func (APIerSv1 *APIerSv1) SetThresholdProfile(args *engine.ThresholdWithCache, reply *string) error {
	if missing := utils.MissingStructFields(args.ThresholdProfile, []string{"Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := APIerSv1.DataManager.SetThresholdProfile(args.ThresholdProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheThresholdProfiles and CacheThresholds and store it in database
	//make 1 insert for both ThresholdProfile and Threshold instead of 2
	loadID := time.Now().UnixNano()
	if err := APIerSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheThresholdProfiles: loadID, utils.CacheThresholds: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for ThresholdProfile
	argCache := utils.ArgsGetCacheItem{
		CacheID: utils.CacheThresholdProfiles,
		ItemID:  args.TenantID(),
	}
	if err := APIerSv1.CallCache(args.Tenant, GetCacheOpt(args.Cache), argCache); err != nil {
		return utils.APIErrorHandler(err)
	}

	if has, err := APIerSv1.DataManager.HasData(utils.ThresholdPrefix, args.ID, args.Tenant); err != nil {
		return err
	} else if !has {
		if err := APIerSv1.DataManager.SetThreshold(&engine.Threshold{Tenant: args.Tenant, ID: args.ID}); err != nil {
			return err
		}
		//handle caching for Threshold
		argCache = utils.ArgsGetCacheItem{
			CacheID: utils.CacheThresholds,
			ItemID:  args.TenantID(),
		}
		if err := APIerSv1.CallCache(args.Tenant, GetCacheOpt(args.Cache), argCache); err != nil {
			return utils.APIErrorHandler(err)
		}
	}

	*reply = utils.OK
	return nil
}

// Remove a specific Threshold Profile
func (APIerSv1 *APIerSv1) RemoveThresholdProfile(args *utils.TenantIDWithCache, reply *string) error {
	if missing := utils.MissingStructFields(args, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := APIerSv1.DataManager.RemoveThresholdProfile(args.Tenant, args.ID, utils.NonTransactional, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for ThresholdProfile
	argCache := utils.ArgsGetCacheItem{
		CacheID: utils.CacheThresholdProfiles,
		ItemID:  args.TenantID(),
	}
	if err := APIerSv1.CallCache(args.Tenant, GetCacheOpt(args.Cache), argCache); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := APIerSv1.DataManager.RemoveThreshold(args.Tenant, args.ID, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheThresholdProfiles and CacheThresholds and store it in database
	//make 1 insert for both ThresholdProfile and Threshold instead of 2
	loadID := time.Now().UnixNano()
	if err := APIerSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheThresholdProfiles: loadID, utils.CacheThresholds: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for Threshold
	argCache = utils.ArgsGetCacheItem{
		CacheID: utils.CacheThresholds,
		ItemID:  args.TenantID(),
	}
	if err := APIerSv1.CallCache(args.Tenant, GetCacheOpt(args.Cache), argCache); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

func (tSv1 *ThresholdSv1) Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error {
	*reply = utils.Pong
	return nil
}
