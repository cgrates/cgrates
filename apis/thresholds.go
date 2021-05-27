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

package apis

import (
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// GetThresholdProfile returns a Threshold Profile
func (adms *AdminSv1) GetThresholdProfile(ctx *context.Context, arg *utils.TenantID, reply *engine.ThresholdProfile) (err error) {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	th, err := adms.dm.GetThresholdProfile(ctx, tnt, arg.ID, true, true, utils.NonTransactional)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = *th
	return
}

// GetThresholdProfileIDs returns list of thresholdProfile IDs registered for a tenant
func (adms *AdminSv1) GetThresholdProfileIDs(ctx *context.Context, args *utils.PaginatorWithTenant, thPrfIDs *[]string) error {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.ThresholdProfilePrefix + tnt + utils.ConcatenatedKeySep
	keys, err := adms.dm.DataDB().GetKeysForPrefix(ctx, prfx)
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

// GetThresholdProfilesCount sets in reply var the total number of ThresholdProfileIDs registered for the received tenant
// returns ErrNotFound in case of 0 ThresholdProfileIDs
func (adms *AdminSv1) GetThresholdProfilesCount(ctx *context.Context, args *utils.TenantWithAPIOpts, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	var keys []string
	prfx := utils.ThresholdProfilePrefix + tnt + utils.ConcatenatedKeySep
	if keys, err = adms.dm.DataDB().GetKeysForPrefix(ctx, prfx); err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(keys)
	return nil
}

// SetThresholdProfile alters/creates a ThresholdProfile
func (adms *AdminSv1) SetThresholdProfile(ctx *context.Context, args *engine.ThresholdProfileWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(args.ThresholdProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if args.Tenant == utils.EmptyString {
		args.Tenant = adms.cfg.GeneralCfg().DefaultTenant
	}
	if err := adms.dm.SetThresholdProfile(ctx, args.ThresholdProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheThresholdProfiles and CacheThresholds and store it in database
	//make 1 insert for both ThresholdProfile and Threshold instead of 2
	loadID := time.Now().UnixNano()
	if err := adms.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheThresholdProfiles: loadID, utils.CacheThresholds: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for ThresholdProfile and Threshold
	if err := adms.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.CacheOpt]), args.Tenant, utils.CacheThresholdProfiles,
		args.TenantID(), &args.FilterIDs, nil, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveThresholdProfile removes a specific Threshold Profile
func (adms *AdminSv1) RemoveThresholdProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	if err := adms.dm.RemoveThresholdProfile(ctx, tnt, args.ID, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for ThresholdProfile
	if err := adms.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.CacheOpt]), tnt, utils.CacheThresholdProfiles,
		utils.ConcatenatedKey(tnt, args.ID), nil, nil, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheThresholdProfiles and CacheThresholds and store it in database
	//make 1 insert for both ThresholdProfile and Threshold instead of 2
	loadID := time.Now().UnixNano()
	if err := adms.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheThresholdProfiles: loadID, utils.CacheThresholds: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// NewThresholdSv1 initializes ThresholdSV1
func NewThresholdSv1(tS *engine.ThresholdService) *ThresholdSv1 {
	return &ThresholdSv1{tS: tS}
}

// ThresholdSv1 exports RPC from RLs
type ThresholdSv1 struct {
	ping
	tS *engine.ThresholdService
}

// GetThresholdIDs returns list of threshold IDs registered for a tenant
func (tSv1 *ThresholdSv1) GetThresholdIDs(ctx *context.Context, tenant *utils.TenantWithAPIOpts, tIDs *[]string) error {
	return tSv1.tS.V1GetThresholdIDs(ctx, tenant.Tenant, tIDs)
}

// GetThresholdsForEvent returns a list of thresholds matching an event
func (tSv1 *ThresholdSv1) GetThresholdsForEvent(ctx *context.Context, args *engine.ThresholdsArgsProcessEvent, reply *engine.Thresholds) error {
	return tSv1.tS.V1GetThresholdsForEvent(ctx, args, reply)
}

// GetThreshold queries a Threshold
func (tSv1 *ThresholdSv1) GetThreshold(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, t *engine.Threshold) error {
	return tSv1.tS.V1GetThreshold(ctx, tntID.TenantID, t)
}

// ProcessEvent will process an Event
func (tSv1 *ThresholdSv1) ProcessEvent(ctx *context.Context, args *engine.ThresholdsArgsProcessEvent, tIDs *[]string) error {
	return tSv1.tS.V1ProcessEvent(ctx, args, tIDs)
}

// ResetThreshold resets the threshold hits
func (tSv1 *ThresholdSv1) ResetThreshold(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *string) error {
	return tSv1.tS.V1ResetThreshold(ctx, tntID.TenantID, reply)
}
