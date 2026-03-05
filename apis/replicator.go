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

//go:generate go run ../data/scripts/generate_replicator

import (
	"fmt"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

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
	indx, err := dataDB.GetIndexesDrv(ctx, args.IdxItmType, args.TntCtx, args.IdxKey, utils.NonTransactional)
	if err != nil {
		return err
	}
	*reply = indx
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

// RemoveIndexes removes indexes from the remote database.
func (r *ReplicatorSv1) RemoveIndexes(ctx *context.Context, args *utils.GetIndexesArg, reply *string) error {
	dataDB, _, err := r.dm.DBConns().GetConn(args.IdxItmType)
	if err != nil {
		return err
	}
	if err := dataDB.RemoveIndexesDrv(ctx, args.IdxItmType, args.TntCtx, args.IdxKey); err != nil {
		return err
	}
	if r.admin.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<ReplicatorSv1.RemoveIndexes> Delaying cache call for %v", r.admin.cfg.GeneralCfg().CachingDelay))
		time.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)
	}
	if err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),
		args.Tenant, args.IdxItmType, utils.ConcatenatedKey(args.TntCtx, args.IdxKey), "", nil, args.APIOpts); err != nil {
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
