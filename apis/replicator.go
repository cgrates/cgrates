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

package apis

import (
	"fmt"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// ReplicatorSv1 exports DataDB methods as RPC endpoints for replication.
type ReplicatorSv1 struct {
	ping
	dm    *engine.DataManager
	admin *AdminSv1
}

// NewReplicatorSv1 creates a new ReplicatorSv1.
func NewReplicatorSv1(dm *engine.DataManager, admin *AdminSv1) *ReplicatorSv1 {
	return &ReplicatorSv1{dm: dm, admin: admin}
}

// GetAccount retrieves an account from the remote database.
func (r *ReplicatorSv1) GetAccount(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *utils.Account) error {
	engine.UpdateReplicationFilters(utils.AccountPrefix, args.TenantID.TenantID(), utils.IfaceAsString(args.APIOpts[utils.RemoteHostOpt]))
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaAccounts)
	if err != nil {
		return err
	}
	rcv, err := dataDB.GetAccountDrv(ctx, args.Tenant, args.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetStatQueue retrieves a stat queue from the remote database.
func (r *ReplicatorSv1) GetStatQueue(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.StatQueue) error {
	engine.UpdateReplicationFilters(utils.StatQueuePrefix, args.TenantID.TenantID(), utils.IfaceAsString(args.APIOpts[utils.RemoteHostOpt]))
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaStatQueues)
	if err != nil {
		return err
	}
	rcv, err := dataDB.GetStatQueueDrv(ctx, args.Tenant, args.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetFilter retrieves a filter from the remote database.
func (r *ReplicatorSv1) GetFilter(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.Filter) error {
	engine.UpdateReplicationFilters(utils.FilterPrefix, args.TenantID.TenantID(), utils.IfaceAsString(args.APIOpts[utils.RemoteHostOpt]))
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaFilters)
	if err != nil {
		return err
	}
	rcv, err := dataDB.GetFilterDrv(ctx, args.Tenant, args.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetThreshold retrieves a threshold from the remote database.
func (r *ReplicatorSv1) GetThreshold(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.Threshold) error {
	engine.UpdateReplicationFilters(utils.ThresholdPrefix, args.TenantID.TenantID(), utils.IfaceAsString(args.APIOpts[utils.RemoteHostOpt]))
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaThresholds)
	if err != nil {
		return err
	}
	rcv, err := dataDB.GetThresholdDrv(ctx, args.Tenant, args.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetThresholdProfile retrieves a threshold profile from the remote database.
func (r *ReplicatorSv1) GetThresholdProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.ThresholdProfile) error {
	engine.UpdateReplicationFilters(utils.ThresholdProfilePrefix, args.TenantID.TenantID(), utils.IfaceAsString(args.APIOpts[utils.RemoteHostOpt]))
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaThresholdProfiles)
	if err != nil {
		return err
	}
	rcv, err := dataDB.GetThresholdProfileDrv(ctx, args.Tenant, args.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetStatQueueProfile retrieves a stat queue profile from the remote database.
func (r *ReplicatorSv1) GetStatQueueProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.StatQueueProfile) error {
	engine.UpdateReplicationFilters(utils.StatQueueProfilePrefix, args.TenantID.TenantID(), utils.IfaceAsString(args.APIOpts[utils.RemoteHostOpt]))
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaStatQueueProfiles)
	if err != nil {
		return err
	}
	rcv, err := dataDB.GetStatQueueProfileDrv(ctx, args.Tenant, args.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetTrendProfile retrieves a trend profile from the remote database.
func (r *ReplicatorSv1) GetTrendProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *utils.TrendProfile) error {
	engine.UpdateReplicationFilters(utils.TrendProfilePrefix, args.TenantID.TenantID(), utils.IfaceAsString(args.APIOpts[utils.RemoteHostOpt]))
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaTrendProfiles)
	if err != nil {
		return err
	}
	rcv, err := dataDB.GetTrendProfileDrv(ctx, args.Tenant, args.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetResource retrieves a resource from the remote database.
func (r *ReplicatorSv1) GetResource(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *utils.Resource) error {
	engine.UpdateReplicationFilters(utils.ResourcesPrefix, args.TenantID.TenantID(), utils.IfaceAsString(args.APIOpts[utils.RemoteHostOpt]))
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaResources)
	if err != nil {
		return err
	}
	rcv, err := dataDB.GetResourceDrv(ctx, args.Tenant, args.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetResourceProfile retrieves a resource profile from the remote database.
func (r *ReplicatorSv1) GetResourceProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *utils.ResourceProfile) error {
	engine.UpdateReplicationFilters(utils.ResourceProfilesPrefix, args.TenantID.TenantID(), utils.IfaceAsString(args.APIOpts[utils.RemoteHostOpt]))
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaResourceProfiles)
	if err != nil {
		return err
	}
	rcv, err := dataDB.GetResourceProfileDrv(ctx, args.Tenant, args.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetIPAllocations retrieves IP allocations from the remote database.
func (r *ReplicatorSv1) GetIPAllocations(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *utils.IPAllocations) error {
	engine.UpdateReplicationFilters(utils.IPAllocationsPrefix, args.TenantID.TenantID(), utils.IfaceAsString(args.APIOpts[utils.RemoteHostOpt]))
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaIPAllocations)
	if err != nil {
		return err
	}
	rcv, err := dataDB.GetIPAllocationsDrv(ctx, args.Tenant, args.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetIPProfile retrieves an IP profile from the remote database.
func (r *ReplicatorSv1) GetIPProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *utils.IPProfile) error {
	engine.UpdateReplicationFilters(utils.IPProfilesPrefix, args.TenantID.TenantID(), utils.IfaceAsString(args.APIOpts[utils.RemoteHostOpt]))
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaIPProfiles)
	if err != nil {
		return err
	}
	rcv, err := dataDB.GetIPProfileDrv(ctx, args.Tenant, args.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetRankingProfile retrieves a ranking profile from the remote database.
func (r *ReplicatorSv1) GetRankingProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *utils.RankingProfile) error {
	engine.UpdateReplicationFilters(utils.RankingProfilePrefix, args.TenantID.TenantID(), utils.IfaceAsString(args.APIOpts[utils.RemoteHostOpt]))
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaRankingProfiles)
	if err != nil {
		return err
	}
	rcv, err := dataDB.GetRankingProfileDrv(ctx, args.Tenant, args.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetRouteProfile retrieves a route profile from the remote database.
func (r *ReplicatorSv1) GetRouteProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *utils.RouteProfile) error {
	engine.UpdateReplicationFilters(utils.RouteProfilePrefix, args.TenantID.TenantID(), utils.IfaceAsString(args.APIOpts[utils.RemoteHostOpt]))
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaRouteProfiles)
	if err != nil {
		return err
	}
	rcv, err := dataDB.GetRouteProfileDrv(ctx, args.Tenant, args.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetAttributeProfile retrieves an attribute profile from the remote database.
func (r *ReplicatorSv1) GetAttributeProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *utils.AttributeProfile) error {
	engine.UpdateReplicationFilters(utils.AttributeProfilePrefix, args.TenantID.TenantID(), utils.IfaceAsString(args.APIOpts[utils.RemoteHostOpt]))
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaAttributeProfiles)
	if err != nil {
		return err
	}
	rcv, err := dataDB.GetAttributeProfileDrv(ctx, args.Tenant, args.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetChargerProfile retrieves a charger profile from the remote database.
func (r *ReplicatorSv1) GetChargerProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *utils.ChargerProfile) error {
	engine.UpdateReplicationFilters(utils.ChargerProfilePrefix, args.TenantID.TenantID(), utils.IfaceAsString(args.APIOpts[utils.RemoteHostOpt]))
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaChargerProfiles)
	if err != nil {
		return err
	}
	rcv, err := dataDB.GetChargerProfileDrv(ctx, args.Tenant, args.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetRateProfile retrieves a rate profile from the remote database.
func (r *ReplicatorSv1) GetRateProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *utils.RateProfile) error {
	engine.UpdateReplicationFilters(utils.RateProfilePrefix, args.TenantID.TenantID(), utils.IfaceAsString(args.APIOpts[utils.RemoteHostOpt]))
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaRateProfiles)
	if err != nil {
		return err
	}
	rcv, err := dataDB.GetRateProfileDrv(ctx, args.Tenant, args.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetActionProfile retrieves an action profile from the remote database.
func (r *ReplicatorSv1) GetActionProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *utils.ActionProfile) error {
	engine.UpdateReplicationFilters(utils.ActionProfilePrefix, args.TenantID.TenantID(), utils.IfaceAsString(args.APIOpts[utils.RemoteHostOpt]))
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaActionProfiles)
	if err != nil {
		return err
	}
	rcv, err := dataDB.GetActionProfileDrv(ctx, args.Tenant, args.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetTrend retrieves a trend from the remote database.
// Unlike the standard Get pattern, Trend contains sync.RWMutex and internal
// caches that cannot be copied with a simple dereference, so each field is
// assigned individually.
func (r *ReplicatorSv1) GetTrend(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *utils.Trend) error {
	engine.UpdateReplicationFilters(utils.TrendPrefix, args.TenantID.TenantID(), utils.IfaceAsString(args.APIOpts[utils.RemoteHostOpt]))
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaTrends)
	if err != nil {
		return err
	}
	rcv, err := dataDB.GetTrendDrv(ctx, args.Tenant, args.ID)
	if err != nil {
		return err
	}
	reply.Tenant = rcv.Tenant
	reply.ID = rcv.ID
	reply.RunTimes = rcv.RunTimes
	reply.CompressedMetrics = rcv.CompressedMetrics
	reply.Metrics = rcv.Metrics
	return nil
}

// GetRanking retrieves a ranking from the remote database.
// Unlike the standard Get pattern, Ranking contains sync.RWMutex so each
// field is assigned individually.
func (r *ReplicatorSv1) GetRanking(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *utils.Ranking) error {
	engine.UpdateReplicationFilters(utils.RankingPrefix, args.TenantID.TenantID(), utils.IfaceAsString(args.APIOpts[utils.RemoteHostOpt]))
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaRankings)
	if err != nil {
		return err
	}
	rcv, err := dataDB.GetRankingDrv(ctx, args.Tenant, args.ID)
	if err != nil {
		return err
	}
	reply.Tenant = rcv.Tenant
	reply.ID = rcv.ID
	reply.LastUpdate = rcv.LastUpdate
	reply.Metrics = rcv.Metrics
	reply.Sorting = rcv.Sorting
	reply.SortingParameters = rcv.SortingParameters
	reply.SortedStatIDs = rcv.SortedStatIDs
	return nil
}

// GetItemLoadIDs retrieves item load IDs from the remote database.
func (r *ReplicatorSv1) GetItemLoadIDs(ctx *context.Context, args *utils.StringWithAPIOpts, reply *map[string]int64) error {
	engine.UpdateReplicationFilters(utils.LoadIDPrefix, args.Arg, utils.IfaceAsString(args.APIOpts[utils.RemoteHostOpt]))
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaLoadIDs)
	if err != nil {
		return err
	}
	rcv, err := dataDB.GetItemLoadIDsDrv(ctx, args.Arg)
	if err != nil {
		return err
	}
	*reply = rcv
	return nil
}

// GetIndexes retrieves indexes from the remote database.
func (r *ReplicatorSv1) GetIndexes(ctx *context.Context, args *utils.GetIndexesArg, reply *map[string]utils.StringSet) error {
	engine.UpdateReplicationFilters(utils.CacheInstanceToPrefix[args.IdxItmType], args.TntCtx, utils.IfaceAsString(args.APIOpts[utils.RemoteHostOpt]))
	dataDB, _, err := r.dm.DBConns().GetConn(args.IdxItmType)
	if err != nil {
		return err
	}
	indx, err := dataDB.GetIndexesDrv(ctx, args.IdxItmType, args.TntCtx, utils.NonTransactional, args.IdxKeys...)
	if err != nil {
		return err
	}
	*reply = indx
	return nil
}

// SetAccount stores an account in the remote database.
func (r *ReplicatorSv1) SetAccount(ctx *context.Context, args *utils.AccountWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaAccounts)
	if err != nil {
		return err
	}
	if err := dataDB.SetAccountDrv(ctx, args.Account); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetThresholdProfile stores a threshold profile in the remote database.
func (r *ReplicatorSv1) SetThresholdProfile(ctx *context.Context, args *engine.ThresholdProfileWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaThresholdProfiles)
	if err != nil {
		return err
	}
	if err := dataDB.SetThresholdProfileDrv(ctx, args.ThresholdProfile); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetThresholdProfile> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheThresholdProfiles, args.TenantID(), "", &args.FilterIDs, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetThreshold stores a threshold in the remote database.
func (r *ReplicatorSv1) SetThreshold(ctx *context.Context, args *engine.ThresholdWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaThresholds)
	if err != nil {
		return err
	}
	if err := dataDB.SetThresholdDrv(ctx, args.Threshold); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetThreshold> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheThresholds, args.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetTrendProfile stores a trend profile in the remote database.
func (r *ReplicatorSv1) SetTrendProfile(ctx *context.Context, args *utils.TrendProfileWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaTrendProfiles)
	if err != nil {
		return err
	}
	if err := dataDB.SetTrendProfileDrv(ctx, args.TrendProfile); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetTrendProfile> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheTrendProfiles, args.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetTrend stores a trend in the remote database.
func (r *ReplicatorSv1) SetTrend(ctx *context.Context, args *utils.TrendWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaTrends)
	if err != nil {
		return err
	}
	if err := dataDB.SetTrendDrv(ctx, args.Trend); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetTrend> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheTrends, args.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetStatQueueProfile stores a stat queue profile in the remote database.
func (r *ReplicatorSv1) SetStatQueueProfile(ctx *context.Context, args *engine.StatQueueProfileWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaStatQueueProfiles)
	if err != nil {
		return err
	}
	if err := dataDB.SetStatQueueProfileDrv(ctx, args.StatQueueProfile); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetStatQueueProfile> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheStatQueueProfiles, args.TenantID(), "", &args.FilterIDs, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetFilter stores a filter in the remote database.
func (r *ReplicatorSv1) SetFilter(ctx *context.Context, args *engine.FilterWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaFilters)
	if err != nil {
		return err
	}
	if err := dataDB.SetFilterDrv(ctx, args.Filter); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetFilter> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheFilters, args.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetResourceProfile stores a resource profile in the remote database.
func (r *ReplicatorSv1) SetResourceProfile(ctx *context.Context, args *utils.ResourceProfileWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaResourceProfiles)
	if err != nil {
		return err
	}
	if err := dataDB.SetResourceProfileDrv(ctx, args.ResourceProfile); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetResourceProfile> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheResourceProfiles, args.TenantID(), "", &args.FilterIDs, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetResource stores a resource in the remote database.
func (r *ReplicatorSv1) SetResource(ctx *context.Context, args *utils.ResourceWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaResources)
	if err != nil {
		return err
	}
	if err := dataDB.SetResourceDrv(ctx, args.Resource); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetResource> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheResources, args.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetIPProfile stores an IP profile in the remote database.
func (r *ReplicatorSv1) SetIPProfile(ctx *context.Context, args *utils.IPProfileWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaIPProfiles)
	if err != nil {
		return err
	}
	if err := dataDB.SetIPProfileDrv(ctx, args.IPProfile); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetIPProfile> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheIPProfiles, args.TenantID(), "", &args.FilterIDs, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetIPAllocations stores IP allocations in the remote database.
func (r *ReplicatorSv1) SetIPAllocations(ctx *context.Context, args *utils.IPAllocationsWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaIPAllocations)
	if err != nil {
		return err
	}
	if err := dataDB.SetIPAllocationsDrv(ctx, args.IPAllocations); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetIPAllocations> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheIPAllocations, args.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetRankingProfile stores a ranking profile in the remote database.
func (r *ReplicatorSv1) SetRankingProfile(ctx *context.Context, args *utils.RankingProfileWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaRankingProfiles)
	if err != nil {
		return err
	}
	if err := dataDB.SetRankingProfileDrv(ctx, args.RankingProfile); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetRankingProfile> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheRankingProfiles, args.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetRanking stores a ranking in the remote database.
func (r *ReplicatorSv1) SetRanking(ctx *context.Context, args *utils.RankingWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaRankings)
	if err != nil {
		return err
	}
	if err := dataDB.SetRankingDrv(ctx, args.Ranking); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetRanking> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheRankings, args.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetRouteProfile stores a route profile in the remote database.
func (r *ReplicatorSv1) SetRouteProfile(ctx *context.Context, args *utils.RouteProfileWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaRouteProfiles)
	if err != nil {
		return err
	}
	if err := dataDB.SetRouteProfileDrv(ctx, args.RouteProfile); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetRouteProfile> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheRouteProfiles, args.TenantID(), "", &args.FilterIDs, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetAttributeProfile stores an attribute profile in the remote database.
func (r *ReplicatorSv1) SetAttributeProfile(ctx *context.Context, args *utils.AttributeProfileWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaAttributeProfiles)
	if err != nil {
		return err
	}
	if err := dataDB.SetAttributeProfileDrv(ctx, args.AttributeProfile); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetAttributeProfile> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheAttributeProfiles, args.TenantID(), "", &args.FilterIDs, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetChargerProfile stores a charger profile in the remote database.
func (r *ReplicatorSv1) SetChargerProfile(ctx *context.Context, args *utils.ChargerProfileWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaChargerProfiles)
	if err != nil {
		return err
	}
	if err := dataDB.SetChargerProfileDrv(ctx, args.ChargerProfile); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetChargerProfile> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheChargerProfiles, args.TenantID(), "", &args.FilterIDs, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetActionProfile stores an action profile in the remote database.
func (r *ReplicatorSv1) SetActionProfile(ctx *context.Context, args *utils.ActionProfileWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaActionProfiles)
	if err != nil {
		return err
	}
	if err := dataDB.SetActionProfileDrv(ctx, args.ActionProfile); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetActionProfile> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheActionProfiles, args.TenantID(), "", &args.FilterIDs, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetStatQueue stores a stat queue in the remote database.
// Unlike the standard Set pattern, StatQueueWithAPIOpts uses a named field
// (not embedded), so tenant/ID access goes through args.StatQueue.
func (r *ReplicatorSv1) SetStatQueue(ctx *context.Context, args *engine.StatQueueWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaStatQueues)
	if err != nil {
		return err
	}
	if err := dataDB.SetStatQueueDrv(ctx, nil, args.StatQueue); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetStatQueue> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.StatQueue.Tenant, utils.CacheStatQueues, args.StatQueue.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetLoadIDs stores load IDs in the remote database.
func (r *ReplicatorSv1) SetLoadIDs(ctx *context.Context, args *utils.LoadIDsWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaLoadIDs)
	if err != nil {
		return err
	}
	if err := dataDB.SetLoadIDsDrv(ctx, args.LoadIDs); err != nil {
		return err
	}
	lIDs := make([]string, 0, len(args.LoadIDs))
	for lID := range args.LoadIDs {
		lIDs = append(lIDs, lID)
	}
	if err := r.admin.callCacheMultiple(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheLoadIDs, lIDs, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetIndexes stores indexes in the remote database.
func (r *ReplicatorSv1) SetIndexes(ctx *context.Context, args *utils.SetIndexesArg, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(args.IdxItmType)
	if err != nil {
		return err
	}
	if err := dataDB.SetIndexesDrv(ctx, args.IdxItmType, args.TntCtx, args.Indexes, true, utils.NonTransactional); err != nil {
		return err
	}
	cIDs := make([]string, 0, len(args.Indexes))
	for idxKey := range args.Indexes {
		cIDs = append(cIDs, utils.ConcatenatedKey(args.TntCtx, idxKey))
	}
	if err := r.admin.callCacheMultiple(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, args.IdxItmType, cIDs, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetRateProfile stores a rate profile in the remote database.
// Unlike the standard Set pattern, SetRateProfileDrv takes an extra overwrite
// argument. Replication always passes false.
func (r *ReplicatorSv1) SetRateProfile(ctx *context.Context, args *utils.RateProfileWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaRateProfiles)
	if err != nil {
		return err
	}
	if err := dataDB.SetRateProfileDrv(ctx, args.RateProfile, false); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.SetRateProfile> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheRateProfiles, args.TenantID(), "", &args.FilterIDs, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveAccount removes an account from the remote database.
func (r *ReplicatorSv1) RemoveAccount(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaAccounts)
	if err != nil {
		return err
	}
	if err := dataDB.RemoveAccountDrv(ctx, args.Tenant, args.ID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveThreshold removes a threshold from the remote database.
func (r *ReplicatorSv1) RemoveThreshold(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaThresholds)
	if err != nil {
		return err
	}
	if err := dataDB.RemoveThresholdDrv(ctx, args.Tenant, args.ID); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveThreshold> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheThresholds, args.TenantID.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveThresholdProfile removes a threshold profile from the remote database.
func (r *ReplicatorSv1) RemoveThresholdProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaThresholdProfiles)
	if err != nil {
		return err
	}
	if err := dataDB.RemThresholdProfileDrv(ctx, args.Tenant, args.ID); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveThresholdProfile> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheThresholdProfiles, args.TenantID.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveTrend removes a trend from the remote database.
func (r *ReplicatorSv1) RemoveTrend(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaTrends)
	if err != nil {
		return err
	}
	if err := dataDB.RemoveTrendDrv(ctx, args.Tenant, args.ID); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveTrend> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheTrends, args.TenantID.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveTrendProfile removes a trend profile from the remote database.
func (r *ReplicatorSv1) RemoveTrendProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaTrendProfiles)
	if err != nil {
		return err
	}
	if err := dataDB.RemTrendProfileDrv(ctx, args.Tenant, args.ID); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveTrendProfile> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheTrendProfiles, args.TenantID.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveStatQueue removes a stat queue from the remote database.
func (r *ReplicatorSv1) RemoveStatQueue(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaStatQueues)
	if err != nil {
		return err
	}
	if err := dataDB.RemStatQueueDrv(ctx, args.Tenant, args.ID); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveStatQueue> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheStatQueues, args.TenantID.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveStatQueueProfile removes a stat queue profile from the remote database.
func (r *ReplicatorSv1) RemoveStatQueueProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaStatQueueProfiles)
	if err != nil {
		return err
	}
	if err := dataDB.RemStatQueueProfileDrv(ctx, args.Tenant, args.ID); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveStatQueueProfile> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheStatQueueProfiles, args.TenantID.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveFilter removes a filter from the remote database.
func (r *ReplicatorSv1) RemoveFilter(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaFilters)
	if err != nil {
		return err
	}
	if err := dataDB.RemoveFilterDrv(ctx, args.Tenant, args.ID); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveFilter> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheFilters, args.TenantID.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveResource removes a resource from the remote database.
func (r *ReplicatorSv1) RemoveResource(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaResources)
	if err != nil {
		return err
	}
	if err := dataDB.RemoveResourceDrv(ctx, args.Tenant, args.ID); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveResource> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheResources, args.TenantID.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveResourceProfile removes a resource profile from the remote database.
func (r *ReplicatorSv1) RemoveResourceProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaResourceProfiles)
	if err != nil {
		return err
	}
	if err := dataDB.RemoveResourceProfileDrv(ctx, args.Tenant, args.ID); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveResourceProfile> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheResourceProfiles, args.TenantID.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveIPAllocations removes IP allocations from the remote database.
func (r *ReplicatorSv1) RemoveIPAllocations(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaIPAllocations)
	if err != nil {
		return err
	}
	if err := dataDB.RemoveIPAllocationsDrv(ctx, args.Tenant, args.ID); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveIPAllocations> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheIPAllocations, args.TenantID.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveIPProfile removes an IP profile from the remote database.
func (r *ReplicatorSv1) RemoveIPProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaIPProfiles)
	if err != nil {
		return err
	}
	if err := dataDB.RemoveIPProfileDrv(ctx, args.Tenant, args.ID); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveIPProfile> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheIPProfiles, args.TenantID.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveRankingProfile removes a ranking profile from the remote database.
func (r *ReplicatorSv1) RemoveRankingProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaRankingProfiles)
	if err != nil {
		return err
	}
	if err := dataDB.RemRankingProfileDrv(ctx, args.Tenant, args.ID); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveRankingProfile> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheRankingProfiles, args.TenantID.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveRanking removes a ranking from the remote database.
func (r *ReplicatorSv1) RemoveRanking(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaRankings)
	if err != nil {
		return err
	}
	if err := dataDB.RemoveRankingDrv(ctx, args.Tenant, args.ID); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveRanking> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheRankings, args.TenantID.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveRouteProfile removes a route profile from the remote database.
func (r *ReplicatorSv1) RemoveRouteProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaRouteProfiles)
	if err != nil {
		return err
	}
	if err := dataDB.RemoveRouteProfileDrv(ctx, args.Tenant, args.ID); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveRouteProfile> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheRouteProfiles, args.TenantID.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveAttributeProfile removes an attribute profile from the remote database.
func (r *ReplicatorSv1) RemoveAttributeProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaAttributeProfiles)
	if err != nil {
		return err
	}
	if err := dataDB.RemoveAttributeProfileDrv(ctx, args.Tenant, args.ID); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveAttributeProfile> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheAttributeProfiles, args.TenantID.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveChargerProfile removes a charger profile from the remote database.
func (r *ReplicatorSv1) RemoveChargerProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaChargerProfiles)
	if err != nil {
		return err
	}
	if err := dataDB.RemoveChargerProfileDrv(ctx, args.Tenant, args.ID); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveChargerProfile> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheChargerProfiles, args.TenantID.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveRateProfile removes a rate profile from the remote database.
// Unlike the standard Remove pattern, RemoveRateProfileDrv takes an extra
// rate IDs argument. Replication always passes nil.
func (r *ReplicatorSv1) RemoveRateProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaRateProfiles)
	if err != nil {
		return err
	}
	if err := dataDB.RemoveRateProfileDrv(ctx, args.Tenant, args.ID, nil); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveRateProfile> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheRateProfiles, args.TenantID.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveActionProfile removes an action profile from the remote database.
func (r *ReplicatorSv1) RemoveActionProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(utils.MetaActionProfiles)
	if err != nil {
		return err
	}
	if err := dataDB.RemoveActionProfileDrv(ctx, args.Tenant, args.ID); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveActionProfile> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, utils.CacheActionProfiles, args.TenantID.TenantID(), "", nil, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveIndexes removes indexes from the remote database.
func (r *ReplicatorSv1) RemoveIndexes(ctx *context.Context, args *utils.GetIndexesArg, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(args.IdxItmType)
	if err != nil {
		return err
	}
	if err := dataDB.RemoveIndexesDrv(ctx, args.IdxItmType, args.TntCtx, args.IdxKeys...); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveIndexes> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	for _, idxKey := range args.IdxKeys {
		if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
			args.Tenant, args.IdxItmType, utils.ConcatenatedKey(args.TntCtx, idxKey), "", nil, args.APIOpts); err != nil {
			return err
		}
	}
	*reply = utils.OK
	return nil
}
