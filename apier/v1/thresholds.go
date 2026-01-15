/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package v1

import (
	"fmt"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewThresholdSv1 initializes ThresholdSV1
func NewThresholdSv1(tS *engine.ThresholdService) *ThresholdSv1 {
	return &ThresholdSv1{tS: tS}
}

// ThresholdSv1 exports RPC from RLs
type ThresholdSv1 struct {
	tS *engine.ThresholdService
}

// GetThresholdIDs returns list of threshold IDs registered for a tenant
func (tSv1 *ThresholdSv1) GetThresholdIDs(ctx *context.Context, tenant *utils.TenantWithAPIOpts, tIDs *[]string) error {
	return tSv1.tS.V1GetThresholdIDs(ctx, tenant.Tenant, tIDs)
}

// GetThresholdsForEvent returns a list of thresholds matching an event
func (tSv1 *ThresholdSv1) GetThresholdsForEvent(ctx *context.Context, args *utils.CGREvent, reply *engine.Thresholds) error {
	return tSv1.tS.V1GetThresholdsForEvent(ctx, args, reply)
}

// GetThreshold queries a Threshold
func (tSv1 *ThresholdSv1) GetThreshold(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, t *engine.Threshold) error {
	return tSv1.tS.V1GetThreshold(ctx, tntID.TenantID, t)
}

// ProcessEvent will process an Event
func (tSv1 *ThresholdSv1) ProcessEvent(ctx *context.Context, args *utils.CGREvent, tIDs *[]string) error {
	return tSv1.tS.V1ProcessEvent(ctx, args, tIDs)
}

// ResetThreshold resets the threshold hits
func (tSv1 *ThresholdSv1) ResetThreshold(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *string) error {
	return tSv1.tS.V1ResetThreshold(ctx, tntID.TenantID, reply)
}

func (tSv1 *ThresholdSv1) RegisterInternalBiJSONConn(ctx *context.Context, args string, rply *string) (err error) {
	return tSv1.tS.BiRPCv1RegisterInternalBiJSONConn(ctx, args, rply)
}

// GetThresholdProfile returns a Threshold Profile
func (apierSv1 *APIerSv1) GetThresholdProfile(ctx *context.Context, arg *utils.TenantID, reply *engine.ThresholdProfile) (err error) {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	th, err := apierSv1.DataManager.GetThresholdProfile(tnt, arg.ID, true, true, utils.NonTransactional)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = *th
	return
}

// GetThresholdProfileIDs returns list of thresholdProfile IDs registered for a tenant
func (apierSv1 *APIerSv1) GetThresholdProfileIDs(ctx *context.Context, args *utils.PaginatorWithTenant, thPrfIDs *[]string) error {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	prfx := utils.ThresholdProfilePrefix + tnt + utils.ConcatenatedKeySep
	keys, err := apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx, args.Search)
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

// GetThresholdProfileCount sets in reply var the total number of ThresholdProfileIDs registered for the received tenant
// returns ErrNotFound in case of 0 ThresholdProfileIDs
func (apierSv1 *APIerSv1) GetThresholdProfileCount(ctx *context.Context, args *utils.TenantWithAPIOpts, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	var keys []string
	prfx := utils.ThresholdProfilePrefix + tnt + utils.ConcatenatedKeySep
	if keys, err = apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx, utils.EmptyString); err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(keys)
	return nil
}

// SetThresholdProfile alters/creates a ThresholdProfile
func (apierSv1 *APIerSv1) SetThresholdProfile(ctx *context.Context, args *engine.ThresholdProfileWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(args.ThresholdProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if args.Tenant == utils.EmptyString {
		args.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.DataManager.SetThresholdProfile(args.ThresholdProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheThresholdProfiles and CacheThresholds and store it in database
	//make 1 insert for both ThresholdProfile and Threshold instead of 2
	loadID := time.Now().UnixNano()
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheThresholdProfiles: loadID, utils.CacheThresholds: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if apierSv1.Config.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<SetThresholdProfile> Delaying cache call for %v", apierSv1.Config.GeneralCfg().CachingDelay))
		time.Sleep(apierSv1.Config.GeneralCfg().CachingDelay)
	}
	//handle caching for ThresholdProfile and Threshold
	if err := apierSv1.CallCache(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]), args.Tenant, utils.CacheThresholdProfiles,
		args.TenantID(), utils.EmptyString, &args.FilterIDs, nil, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveThresholdProfile removes a specific Threshold Profile
func (apierSv1 *APIerSv1) RemoveThresholdProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.DataManager.RemoveThresholdProfile(tnt, args.ID, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if apierSv1.Config.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<RemoveThresholdProfile> Delaying cache call for %v", apierSv1.Config.GeneralCfg().CachingDelay))
		time.Sleep(apierSv1.Config.GeneralCfg().CachingDelay)
	}
	//handle caching for ThresholdProfile
	if err := apierSv1.CallCache(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]), tnt, utils.CacheThresholdProfiles,
		utils.ConcatenatedKey(tnt, args.ID), utils.EmptyString, nil, nil, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheThresholdProfiles and CacheThresholds and store it in database
	//make 1 insert for both ThresholdProfile and Threshold instead of 2
	loadID := time.Now().UnixNano()
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheThresholdProfiles: loadID, utils.CacheThresholds: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}
