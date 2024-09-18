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
	"fmt"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewReplicatorSv1 constructs the ReplicatorSv1 object
func NewReplicatorSv1(dm *engine.DataManager, v1 *AdminSv1) *ReplicatorSv1 {
	return &ReplicatorSv1{
		dm: dm,
		v1: v1,
	}
}

// ReplicatorSv1 exports the DataDB methods to RPC
type ReplicatorSv1 struct {
	ping
	dm *engine.DataManager
	v1 *AdminSv1 // needed for CallCache only
}

// GetAccount is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetAccount(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *utils.Account) error {
	engine.UpdateReplicationFilters(utils.AccountPrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetAccountDrv(ctx, tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetStatQueue is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetStatQueue(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.StatQueue) error {
	engine.UpdateReplicationFilters(utils.StatQueuePrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetStatQueueDrv(ctx, tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetFilter is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetFilter(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.Filter) error {
	engine.UpdateReplicationFilters(utils.FilterPrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetFilterDrv(ctx, tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetThreshold is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetThreshold(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.Threshold) error {
	engine.UpdateReplicationFilters(utils.ThresholdPrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetThresholdDrv(ctx, tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetThresholdProfile is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetThresholdProfile(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.ThresholdProfile) error {
	engine.UpdateReplicationFilters(utils.ThresholdProfilePrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetThresholdProfileDrv(ctx, tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetStatQueueProfile is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetStatQueueProfile(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.StatQueueProfile) error {
	engine.UpdateReplicationFilters(utils.StatQueueProfilePrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetStatQueueProfileDrv(ctx, tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetTrend is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetTrend(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.Trend) error {
	engine.UpdateReplicationFilters(utils.TrendPrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetTrendDrv(tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetTrendProfile is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetTrendProfile(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.TrendProfile) error {
	engine.UpdateReplicationFilters(utils.TrendProfilePrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetTrendProfileDrv(ctx, tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetResource is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetResource(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.Resource) error {
	engine.UpdateReplicationFilters(utils.ResourcesPrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetResourceDrv(ctx, tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetResourceProfile is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetResourceProfile(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.ResourceProfile) error {
	engine.UpdateReplicationFilters(utils.ResourceProfilesPrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetResourceProfileDrv(ctx, tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetRouteProfile is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetRouteProfile(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.RouteProfile) error {
	engine.UpdateReplicationFilters(utils.RouteProfilePrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetRouteProfileDrv(ctx, tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetAttributeProfile is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetAttributeProfile(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.AttributeProfile) error {
	engine.UpdateReplicationFilters(utils.AttributeProfilePrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetAttributeProfileDrv(ctx, tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetChargerProfile is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetChargerProfile(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.ChargerProfile) error {
	engine.UpdateReplicationFilters(utils.ChargerProfilePrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetChargerProfileDrv(ctx, tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetDispatcherProfile is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetDispatcherProfile(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.DispatcherProfile) error {
	engine.UpdateReplicationFilters(utils.DispatcherProfilePrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetDispatcherProfileDrv(ctx, tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetDispatcherHost is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetDispatcherHost(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.DispatcherHost) error {
	engine.UpdateReplicationFilters(utils.DispatcherHostPrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetDispatcherHostDrv(ctx, tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetItemLoadIDs is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetItemLoadIDs(ctx *context.Context, itemID *utils.StringWithAPIOpts, reply *map[string]int64) error {
	engine.UpdateReplicationFilters(utils.LoadIDPrefix, itemID.Arg, utils.IfaceAsString(itemID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetItemLoadIDsDrv(ctx, itemID.Arg)
	if err != nil {
		return err
	}
	*reply = rcv
	return nil
}

// GetIndexes is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetIndexes(ctx *context.Context, args *utils.GetIndexesArg, reply *map[string]utils.StringSet) error {
	engine.UpdateReplicationFilters(utils.CacheInstanceToPrefix[args.IdxItmType], args.TntCtx, utils.IfaceAsString(args.APIOpts[utils.RemoteHostOpt]))
	indx, err := rplSv1.dm.DataDB().GetIndexesDrv(ctx, args.IdxItmType, args.TntCtx, args.IdxKey, utils.NonTransactional)
	if err != nil {
		return err
	}
	*reply = indx
	return nil
}

// SetAccount is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetAccount(ctx *context.Context, acc *utils.AccountWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetAccountDrv(ctx, acc.Account); err != nil {
		return
	}
	// the account doesn't have cache
	*reply = utils.OK
	return
}

// SetThresholdProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetThresholdProfile(ctx *context.Context, th *engine.ThresholdProfileWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetThresholdProfileDrv(ctx, th.ThresholdProfile); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetThresholdProfile> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(th.APIOpts[utils.MetaCache]),
		th.Tenant, utils.CacheThresholdProfiles, th.TenantID(), utils.EmptyString, &th.FilterIDs, th.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetThreshold is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetThreshold(ctx *context.Context, th *engine.ThresholdWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetThresholdDrv(ctx, th.Threshold); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetThreshold> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(th.APIOpts[utils.MetaCache]),
		th.Tenant, utils.CacheThresholds, th.TenantID(), utils.EmptyString, nil, th.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetTrendProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetTrendProfile(ctx *context.Context, trp *engine.TrendProfileWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetTrendProfileDrv(ctx, trp.TrendProfile); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetTrendProfile> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(trp.APIOpts[utils.MetaCache]),
		trp.Tenant, utils.CacheTrendProfiles, trp.TenantID(), utils.EmptyString, nil, trp.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetTrend is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetTrend(ctx *context.Context, tr *engine.TrendWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetTrendDrv(tr.Trend); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetTrend> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(tr.APIOpts[utils.MetaCache]),
		tr.Tenant, utils.CacheTrends, tr.TenantID(), utils.EmptyString, nil, tr.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetStatQueueProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetStatQueueProfile(ctx *context.Context, sq *engine.StatQueueProfileWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetStatQueueProfileDrv(ctx, sq.StatQueueProfile); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetStatQueueProfile> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(sq.APIOpts[utils.MetaCache]),
		sq.Tenant, utils.CacheStatQueueProfiles, sq.TenantID(), utils.EmptyString, &sq.FilterIDs, sq.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetStatQueue is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetStatQueue(ctx *context.Context, sq *engine.StatQueueWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetStatQueueDrv(ctx, nil, sq.StatQueue); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetStatQueue> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(sq.APIOpts[utils.MetaCache]),
		sq.StatQueue.Tenant, utils.CacheStatQueues, sq.StatQueue.TenantID(), utils.EmptyString, nil, sq.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetFilter is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetFilter(ctx *context.Context, fltr *engine.FilterWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetFilterDrv(ctx, fltr.Filter); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetFilter> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(fltr.APIOpts[utils.MetaCache]),
		fltr.Tenant, utils.CacheFilters, fltr.TenantID(), utils.EmptyString, nil, fltr.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetResourceProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetResourceProfile(ctx *context.Context, rs *engine.ResourceProfileWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetResourceProfileDrv(ctx, rs.ResourceProfile); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetResourceProfile> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(rs.APIOpts[utils.MetaCache]),
		rs.Tenant, utils.CacheResourceProfiles, rs.TenantID(), utils.EmptyString, &rs.FilterIDs, rs.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetResource is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetResource(ctx *context.Context, rs *engine.ResourceWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetResourceDrv(ctx, rs.Resource); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetResource> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(rs.APIOpts[utils.MetaCache]),
		rs.Tenant, utils.CacheResources, rs.TenantID(), utils.EmptyString, nil, rs.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetRouteProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetRouteProfile(ctx *context.Context, sp *engine.RouteProfileWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetRouteProfileDrv(ctx, sp.RouteProfile); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetRouteProfile> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(sp.APIOpts[utils.MetaCache]),
		sp.Tenant, utils.CacheRouteProfiles, sp.TenantID(), utils.EmptyString, &sp.FilterIDs, sp.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetAttributeProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetAttributeProfile(ctx *context.Context, ap *engine.AttributeProfileWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetAttributeProfileDrv(ctx, ap.AttributeProfile); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetAttributeProfile> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(ap.APIOpts[utils.MetaCache]),
		ap.Tenant, utils.CacheAttributeProfiles, ap.TenantID(), utils.EmptyString, &ap.FilterIDs, ap.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetChargerProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetChargerProfile(ctx *context.Context, cp *engine.ChargerProfileWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetChargerProfileDrv(ctx, cp.ChargerProfile); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetChargerProfile> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(cp.APIOpts[utils.MetaCache]),
		cp.Tenant, utils.CacheChargerProfiles, cp.TenantID(), utils.EmptyString, &cp.FilterIDs, cp.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetDispatcherProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetDispatcherProfile(ctx *context.Context, dpp *engine.DispatcherProfileWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetDispatcherProfileDrv(ctx, dpp.DispatcherProfile); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetDispatcherProfile> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(dpp.APIOpts[utils.MetaCache]),
		dpp.Tenant, utils.CacheDispatcherProfiles, dpp.TenantID(), utils.EmptyString, &dpp.FilterIDs, dpp.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetDispatcherHost is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetDispatcherHost(ctx *context.Context, dpp *engine.DispatcherHostWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetDispatcherHostDrv(ctx, dpp.DispatcherHost); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetDispatcherHost> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(dpp.APIOpts[utils.MetaCache]),
		dpp.Tenant, utils.CacheDispatcherHosts, dpp.TenantID(), utils.EmptyString, nil, dpp.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetLoadIDs is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetLoadIDs(ctx *context.Context, args *utils.LoadIDsWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetLoadIDsDrv(ctx, args.LoadIDs); err != nil {
		return
	}
	lIDs := make([]string, 0, len(args.LoadIDs))
	for lID := range args.LoadIDs {
		lIDs = append(lIDs, lID)
	}
	if err = rplSv1.v1.callCacheMultiple(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheLoadIDs, lIDs, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetIndexes is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetIndexes(ctx *context.Context, args *utils.SetIndexesArg, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetIndexesDrv(ctx, args.IdxItmType, args.TntCtx, args.Indexes, true, utils.NonTransactional); err != nil {
		return
	}
	cIDs := make([]string, 0, len(args.Indexes))
	for idxKey := range args.Indexes {
		cIDs = append(cIDs, utils.ConcatenatedKey(args.TntCtx, idxKey))
	}
	if err = rplSv1.v1.callCacheMultiple(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, args.IdxItmType, cIDs, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveTrend is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveTrend(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveTrendDrv(args.Tenant, args.ID); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveTrend> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheTrends, args.TenantID.TenantID(), utils.EmptyString, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveTrendProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveTrendProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemTrendProfileDrv(ctx, args.Tenant, args.ID); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveTrendProfile> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheTrendProfiles, args.TenantID.TenantID(), utils.EmptyString, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveThreshold is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveThreshold(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveThresholdDrv(ctx, args.Tenant, args.ID); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveThreshold> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheThresholds, args.TenantID.TenantID(), utils.EmptyString, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveAccount is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveAccount(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveAccountDrv(ctx, args.Tenant, args.ID); err != nil {
		return
	}
	// the account doesn't have cache
	*reply = utils.OK
	return
}

// RemoveStatQueue is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveStatQueue(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemStatQueueDrv(ctx, args.Tenant, args.ID); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveStatQueue> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheStatQueues, args.TenantID.TenantID(), utils.EmptyString, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveFilter is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveFilter(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveFilterDrv(ctx, args.Tenant, args.ID); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveFilter> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheFilters, args.TenantID.TenantID(), utils.EmptyString, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveThresholdProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveThresholdProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemThresholdProfileDrv(ctx, args.Tenant, args.ID); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveThresholdProfile> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheThresholdProfiles, args.TenantID.TenantID(), utils.EmptyString, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveStatQueueProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveStatQueueProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemStatQueueProfileDrv(ctx, args.Tenant, args.ID); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveStatQueueProfile> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheStatQueueProfiles, args.TenantID.TenantID(), utils.EmptyString, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveResource is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveResource(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveResourceDrv(ctx, args.Tenant, args.ID); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveResource> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheResources, args.TenantID.TenantID(), utils.EmptyString, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveResourceProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveResourceProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveResourceProfileDrv(ctx, args.Tenant, args.ID); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveResourceProfile> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheResourceProfiles, args.TenantID.TenantID(), utils.EmptyString, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveRouteProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveRouteProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveRouteProfileDrv(ctx, args.Tenant, args.ID); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveRouteProfile> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheRouteProfiles, args.TenantID.TenantID(), utils.EmptyString, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveAttributeProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveAttributeProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveAttributeProfileDrv(ctx, args.Tenant, args.ID); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveAttributeProfile> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheAttributeProfiles, args.TenantID.TenantID(), utils.EmptyString, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveChargerProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveChargerProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveChargerProfileDrv(ctx, args.Tenant, args.ID); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveChargerProfile> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheChargerProfiles, args.TenantID.TenantID(), utils.EmptyString, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveDispatcherProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveDispatcherProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveDispatcherProfileDrv(ctx, args.Tenant, args.ID); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveDispatcherProfile> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheDispatcherProfiles, args.TenantID.TenantID(), utils.EmptyString, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveDispatcherHost is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveDispatcherHost(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveDispatcherHostDrv(ctx, args.Tenant, args.ID); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveDispatcherHost> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheDispatcherHosts, args.TenantID.TenantID(), utils.EmptyString, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveIndexes  is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveIndexes(ctx *context.Context, args *utils.GetIndexesArg, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveIndexesDrv(ctx, args.IdxItmType, args.TntCtx, args.IdxKey); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveIndexes> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, args.IdxItmType, utils.ConcatenatedKey(args.TntCtx, args.IdxKey), utils.EmptyString, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

func (rplSv1 *ReplicatorSv1) GetRateProfile(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *utils.RateProfile) error {
	engine.UpdateReplicationFilters(utils.RateProfilePrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetRateProfileDrv(ctx, tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}
func (rplSv1 *ReplicatorSv1) GetActionProfile(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.ActionProfile) error {
	engine.UpdateReplicationFilters(utils.ActionProfilePrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetActionProfileDrv(ctx, tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

func (rplSv1 *ReplicatorSv1) SetRateProfile(ctx *context.Context, sp *utils.RateProfileWithAPIOpts, reply *string) (err error) {
	// check if we want to overwrite our profile already existing in database
	var optOverwrite bool
	if _, has := sp.APIOpts[utils.MetaRateSOverwrite]; has {
		optOverwrite, err = utils.IfaceAsBool(sp.APIOpts[utils.MetaRateSOverwrite])
		if err != nil {
			return
		}
	}
	if err = rplSv1.dm.DataDB().SetRateProfileDrv(ctx, sp.RateProfile, optOverwrite); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetRateProfile> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(sp.APIOpts[utils.MetaCache]),
		sp.Tenant, utils.CacheRateProfiles, sp.TenantID(), utils.EmptyString, &sp.FilterIDs, sp.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}
func (rplSv1 *ReplicatorSv1) SetActionProfile(ctx *context.Context, sp *engine.ActionProfileWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetActionProfileDrv(ctx, sp.ActionProfile); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetActionProfile> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(sp.APIOpts[utils.MetaCache]),
		sp.Tenant, utils.CacheActionProfiles, sp.TenantID(), utils.EmptyString, &sp.FilterIDs, sp.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

func (rplSv1 *ReplicatorSv1) RemoveRateProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveRateProfileDrv(ctx, args.Tenant, args.ID, nil); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveRateProfile> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheRateProfiles, args.TenantID.TenantID(), utils.EmptyString, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

func (rplSv1 *ReplicatorSv1) RemoveActionProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveActionProfileDrv(ctx, args.Tenant, args.ID); err != nil {
		return
	}
	// delay if needed before cache call
	if rplSv1.v1.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveActionProfile> Delaying cache call for %v", rplSv1.v1.cfg.GeneralCfg().CachingDelay))
		time.Sleep(rplSv1.v1.cfg.GeneralCfg().CachingDelay)
	}
	if err = rplSv1.v1.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheActionProfiles, args.TenantID.TenantID(), utils.EmptyString, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}
