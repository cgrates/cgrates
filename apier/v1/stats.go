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

// GetStatQueueProfile returns a StatQueue profile
func (apierSv1 *APIerSv1) GetStatQueueProfile(ctx *context.Context, arg *utils.TenantID, reply *engine.StatQueueProfile) (err error) {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	sCfg, err := apierSv1.DataManager.GetStatQueueProfile(tnt, arg.ID,
		true, true, utils.NonTransactional)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = *sCfg
	return
}

// GetStatQueueProfileIDs returns list of statQueueProfile IDs registered for a tenant
func (apierSv1 *APIerSv1) GetStatQueueProfileIDs(ctx *context.Context, args *utils.PaginatorWithTenant, stsPrfIDs *[]string) error {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	prfx := utils.StatQueueProfilePrefix + tnt + utils.ConcatenatedKeySep
	keys, err := apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx)
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
	*stsPrfIDs = args.PaginateStringSlice(retIDs)
	return nil
}

// SetStatQueueProfile alters/creates a StatQueueProfile
func (apierSv1 *APIerSv1) SetStatQueueProfile(ctx *context.Context, arg *engine.StatQueueProfileWithAPIOpts, reply *string) (err error) {
	if missing := utils.MissingStructFields(arg.StatQueueProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err = apierSv1.DataManager.SetStatQueueProfile(arg.StatQueueProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheStatQueueProfiles and CacheStatQueues and store it in database
	//make 1 insert for both StatQueueProfile and StatQueue instead of 2
	loadID := time.Now().UnixNano()
	if err = apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheStatQueueProfiles: loadID, utils.CacheStatQueues: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if apierSv1.Config.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<SetStatQueueProfile> Delaying cache call for %v", apierSv1.Config.GeneralCfg().CachingDelay))
		time.Sleep(apierSv1.Config.GeneralCfg().CachingDelay)
	}
	//handle caching for StatQueueProfile
	if err = apierSv1.CallCache(utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]), arg.Tenant, utils.CacheStatQueueProfiles,
		arg.TenantID(), utils.EmptyString, &arg.FilterIDs, nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveStatQueueProfile remove a specific stat configuration
func (apierSv1 *APIerSv1) RemoveStatQueueProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.DataManager.RemoveStatQueueProfile(tnt, args.ID, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if apierSv1.Config.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<RemoveStatQueueProfile> Delaying cache call for %v", apierSv1.Config.GeneralCfg().CachingDelay))
		time.Sleep(apierSv1.Config.GeneralCfg().CachingDelay)
	}
	//handle caching for StatQueueProfile
	if err := apierSv1.CallCache(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]), tnt, utils.CacheStatQueueProfiles,
		utils.ConcatenatedKey(tnt, args.ID), utils.EmptyString, nil, nil, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheStatQueueProfiles and CacheStatQueues and store it in database
	//make 1 insert for both StatQueueProfile and StatQueue instead of 2
	loadID := time.Now().UnixNano()
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheStatQueueProfiles: loadID, utils.CacheStatQueues: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// NewStatSv1 initializes StatSV1
func NewStatSv1(sS *engine.StatService) *StatSv1 {
	return &StatSv1{sS: sS}
}

// StatSv1 exports RPC from RLs
type StatSv1 struct {
	sS *engine.StatService
}

// GetQueueIDs returns list of queueIDs registered for a tenant
func (stsv1 *StatSv1) GetQueueIDs(ctx *context.Context, tenant *utils.TenantWithAPIOpts, qIDs *[]string) error {
	return stsv1.sS.V1GetQueueIDs(ctx, tenant.Tenant, qIDs)
}

// ProcessEvent returns processes a new Event
func (stsv1 *StatSv1) ProcessEvent(ctx *context.Context, args *utils.CGREvent, reply *[]string) error {
	return stsv1.sS.V1ProcessEvent(ctx, args, reply)
}

// GetStatQueuesForEvent returns the list of queues IDs in the system
func (stsv1 *StatSv1) GetStatQueuesForEvent(ctx *context.Context, args *utils.CGREvent, reply *[]string) (err error) {
	return stsv1.sS.V1GetStatQueuesForEvent(ctx, args, reply)
}

// GetStatQueue returns a StatQueue object
func (stsv1 *StatSv1) GetStatQueue(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.StatQueue) (err error) {
	return stsv1.sS.V1GetStatQueue(ctx, args, reply)
}

// GetQueueStringMetrics returns the string metrics for a Queue
func (stsv1 *StatSv1) GetQueueStringMetrics(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *map[string]string) (err error) {
	return stsv1.sS.V1GetQueueStringMetrics(ctx, args.TenantID, reply)
}

// GetQueueFloatMetrics returns the float metrics for a Queue
func (stsv1 *StatSv1) GetQueueFloatMetrics(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *map[string]float64) (err error) {
	return stsv1.sS.V1GetQueueFloatMetrics(ctx, args.TenantID, reply)
}

// ResetStatQueue resets the stat queue
func (stsv1 *StatSv1) ResetStatQueue(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	return stsv1.sS.V1ResetStatQueue(ctx, args.TenantID, reply)
}
