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

package apis

import (
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// GetStatQueueProfile returns a StatQueue profile
func (adms *AdminSv1) GetStatQueueProfile(ctx *context.Context, arg *utils.TenantID, reply *engine.StatQueueProfile) (err error) {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	sCfg, err := adms.dm.GetStatQueueProfile(ctx, tnt, arg.ID,
		true, true, utils.NonTransactional)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = *sCfg
	return
}

// GetStatQueueProfileIDs returns list of statQueueProfile IDs registered for a tenant
func (adms *AdminSv1) GetStatQueueProfileIDs(ctx *context.Context, args *utils.PaginatorWithTenant, stsPrfIDs *[]string) error {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.StatQueueProfilePrefix + tnt + utils.ConcatenatedKeySep
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
	*stsPrfIDs = args.PaginateStringSlice(retIDs)
	return nil
}

// SetStatQueueProfile alters/creates a StatQueueProfile
func (adms *AdminSv1) SetStatQueueProfile(ctx *context.Context, arg *engine.StatQueueProfileWithAPIOpts, reply *string) (err error) {
	if missing := utils.MissingStructFields(arg.StatQueueProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = adms.cfg.GeneralCfg().DefaultTenant
	}
	if err = adms.dm.SetStatQueueProfile(ctx, arg.StatQueueProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheStatQueueProfiles and CacheStatQueues and store it in database
	//make 1 insert for both StatQueueProfile and StatQueue instead of 2
	loadID := time.Now().UnixNano()
	if err = adms.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheStatQueueProfiles: loadID, utils.CacheStatQueues: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for StatQueueProfile
	if err = adms.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]), arg.Tenant, utils.CacheStatQueueProfiles,
		arg.TenantID(), &arg.FilterIDs, nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveStatQueueProfile remove a specific stat configuration
func (adms *AdminSv1) RemoveStatQueueProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	if err := adms.dm.RemoveStatQueueProfile(ctx, tnt, args.ID, utils.NonTransactional, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for StatQueueProfile
	if err := adms.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.CacheOpt]), tnt, utils.CacheStatQueueProfiles,
		utils.ConcatenatedKey(tnt, args.ID), nil, nil, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheStatQueueProfiles and CacheStatQueues and store it in database
	//make 1 insert for both StatQueueProfile and StatQueue instead of 2
	loadID := time.Now().UnixNano()
	if err := adms.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheStatQueueProfiles: loadID, utils.CacheStatQueues: loadID}); err != nil {
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
	ping
	sS *engine.StatService
}

// GetQueueIDs returns list of queueIDs registered for a tenant
func (stsv1 *StatSv1) GetQueueIDs(ctx *context.Context, tenant *utils.TenantWithAPIOpts, qIDs *[]string) error {
	return stsv1.sS.V1GetQueueIDs(ctx, tenant.Tenant, qIDs)
}

// ProcessEvent returns processes a new Event
func (stsv1 *StatSv1) ProcessEvent(ctx *context.Context, args *engine.StatsArgsProcessEvent, reply *[]string) error {
	return stsv1.sS.V1ProcessEvent(ctx, args, reply)
}

// GetStatQueuesForEvent returns the list of queues IDs in the system
func (stsv1 *StatSv1) GetStatQueuesForEvent(ctx *context.Context, args *engine.StatsArgsProcessEvent, reply *[]string) (err error) {
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
func (stsv1 *StatSv1) ResetStatQueue(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *string) error {
	return stsv1.sS.V1ResetStatQueue(ctx, tntID.TenantID, reply)
}
