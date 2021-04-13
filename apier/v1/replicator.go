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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewReplicatorSv1 constructs the ReplicatorSv1 object
func NewReplicatorSv1(dm *engine.DataManager, v1 *APIerSv1) *ReplicatorSv1 {
	return &ReplicatorSv1{
		dm: dm,
		v1: v1,
	}
}

// ReplicatorSv1 exports the DataDB methods to RPC
type ReplicatorSv1 struct {
	dm *engine.DataManager
	v1 *APIerSv1 // needed for CallCache only
}

// Call implements birpc.ClientConnector interface for internal RPC
func (rplSv1 *ReplicatorSv1) Call(ctx *context.Context, serviceMethod string, args, reply interface{}) error {
	return utils.APIerRPCCall(rplSv1, serviceMethod, args, reply)
}

// GetDestination is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetDestination(key *utils.StringWithAPIOpts, reply *engine.Destination) error {
	engine.UpdateReplicationFilters(utils.DestinationPrefix, key.Arg, utils.IfaceAsString(key.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetDestinationDrv(key.Arg, utils.NonTransactional)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetReverseDestination is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetReverseDestination(key *utils.StringWithAPIOpts, reply *[]string) error {
	rcv, err := rplSv1.dm.DataDB().GetReverseDestinationDrv(key.Arg, utils.NonTransactional)
	if err != nil {
		return err
	}
	for _, dstID := range rcv {
		engine.UpdateReplicationFilters(utils.DestinationPrefix, dstID, utils.IfaceAsString(key.APIOpts[utils.RemoteHostOpt]))
	}
	*reply = rcv
	return nil
}

// GetStatQueue is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetStatQueue(tntID *utils.TenantIDWithAPIOpts, reply *engine.StatQueue) error {
	engine.UpdateReplicationFilters(utils.StatQueuePrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetStatQueueDrv(tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetFilter is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetFilter(tntID *utils.TenantIDWithAPIOpts, reply *engine.Filter) error {
	engine.UpdateReplicationFilters(utils.FilterPrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetFilterDrv(context.TODO(), tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetThreshold is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetThreshold(tntID *utils.TenantIDWithAPIOpts, reply *engine.Threshold) error {
	engine.UpdateReplicationFilters(utils.ThresholdPrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetThresholdDrv(tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetThresholdProfile is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetThresholdProfile(tntID *utils.TenantIDWithAPIOpts, reply *engine.ThresholdProfile) error {
	engine.UpdateReplicationFilters(utils.ThresholdProfilePrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetThresholdProfileDrv(tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetStatQueueProfile is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetStatQueueProfile(tntID *utils.TenantIDWithAPIOpts, reply *engine.StatQueueProfile) error {
	engine.UpdateReplicationFilters(utils.StatQueueProfilePrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetStatQueueProfileDrv(tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetTiming is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetTiming(id *utils.StringWithAPIOpts, reply *utils.TPTiming) error {
	engine.UpdateReplicationFilters(utils.TimingsPrefix, id.Arg, utils.IfaceAsString(id.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetTimingDrv(id.Arg)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetResource is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetResource(tntID *utils.TenantIDWithAPIOpts, reply *engine.Resource) error {
	engine.UpdateReplicationFilters(utils.ResourcesPrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetResourceDrv(tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetResourceProfile is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetResourceProfile(tntID *utils.TenantIDWithAPIOpts, reply *engine.ResourceProfile) error {
	engine.UpdateReplicationFilters(utils.ResourceProfilesPrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetResourceProfileDrv(tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetRouteProfile is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetRouteProfile(tntID *utils.TenantIDWithAPIOpts, reply *engine.RouteProfile) error {
	engine.UpdateReplicationFilters(utils.RouteProfilePrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetRouteProfileDrv(tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetAttributeProfile is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetAttributeProfile(tntID *utils.TenantIDWithAPIOpts, reply *engine.AttributeProfile) error {
	engine.UpdateReplicationFilters(utils.AttributeProfilePrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetAttributeProfileDrv(context.TODO(), tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetChargerProfile is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetChargerProfile(tntID *utils.TenantIDWithAPIOpts, reply *engine.ChargerProfile) error {
	engine.UpdateReplicationFilters(utils.ChargerProfilePrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetChargerProfileDrv(tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetDispatcherProfile is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetDispatcherProfile(tntID *utils.TenantIDWithAPIOpts, reply *engine.DispatcherProfile) error {
	engine.UpdateReplicationFilters(utils.DispatcherProfilePrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetDispatcherProfileDrv(tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetDispatcherHost is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetDispatcherHost(tntID *utils.TenantIDWithAPIOpts, reply *engine.DispatcherHost) error {
	engine.UpdateReplicationFilters(utils.DispatcherHostPrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetDispatcherHostDrv(tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetRateProfile is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetRateProfile(tntID *utils.TenantIDWithAPIOpts, reply *utils.RateProfile) error {
	engine.UpdateReplicationFilters(utils.RateProfilePrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetRateProfileDrv(tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetActionProfile is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetActionProfile(tntID *utils.TenantIDWithAPIOpts, reply *engine.ActionProfile) error {
	engine.UpdateReplicationFilters(utils.ActionProfilePrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetActionProfileDrv(tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetAccount is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetAccount(tntID *utils.TenantIDWithAPIOpts, reply *utils.Account) error {
	engine.UpdateReplicationFilters(utils.AccountPrefix, tntID.TenantID.TenantID(), utils.IfaceAsString(tntID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetAccountDrv(tntID.Tenant, tntID.ID)
	if err != nil {
		return err
	}
	*reply = *rcv
	return nil
}

// GetItemLoadIDs is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetItemLoadIDs(itemID *utils.StringWithAPIOpts, reply *map[string]int64) error {
	engine.UpdateReplicationFilters(utils.LoadIDPrefix, itemID.Arg, utils.IfaceAsString(itemID.APIOpts[utils.RemoteHostOpt]))
	rcv, err := rplSv1.dm.DataDB().GetItemLoadIDsDrv(itemID.Arg)
	if err != nil {
		return err
	}
	*reply = rcv
	return nil
}

// GetIndexes is the remote method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) GetIndexes(args *utils.GetIndexesArg, reply *map[string]utils.StringSet) error {
	engine.UpdateReplicationFilters(utils.CacheInstanceToPrefix[args.IdxItmType], args.TntCtx, utils.IfaceAsString(args.APIOpts[utils.RemoteHostOpt]))
	indx, err := rplSv1.dm.DataDB().GetIndexesDrv(context.TODO(), args.IdxItmType, args.TntCtx, args.IdxKey)
	if err != nil {
		return err
	}
	*reply = indx
	return nil
}

// SetDestination is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetDestination(dst *engine.DestinationWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetDestinationDrv(dst.Destination, utils.NonTransactional); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(dst.APIOpts[utils.CacheOpt]),
		dst.Tenant, utils.CacheDestinations, dst.ID, nil, nil, dst.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetReverseDestination is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetReverseDestination(dst *engine.DestinationWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetReverseDestinationDrv(dst.Destination.ID, dst.Destination.Prefixes, utils.NonTransactional); err != nil {
		return
	}
	if err = rplSv1.v1.callCacheMultiple(utils.IfaceAsString(dst.APIOpts[utils.CacheOpt]),
		dst.Tenant, utils.CacheReverseDestinations, dst.Prefixes, dst.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetThresholdProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetThresholdProfile(th *engine.ThresholdProfileWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetThresholdProfileDrv(th.ThresholdProfile); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(th.APIOpts[utils.CacheOpt]),
		th.Tenant, utils.CacheThresholdProfiles, th.TenantID(), &th.FilterIDs, nil, th.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetThreshold is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetThreshold(th *engine.ThresholdWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetThresholdDrv(th.Threshold); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(th.APIOpts[utils.CacheOpt]),
		th.Tenant, utils.CacheThresholds, th.TenantID(), nil, nil, th.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetStatQueueProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetStatQueueProfile(sq *engine.StatQueueProfileWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetStatQueueProfileDrv(sq.StatQueueProfile); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(sq.APIOpts[utils.CacheOpt]),
		sq.Tenant, utils.CacheStatQueueProfiles, sq.TenantID(), &sq.FilterIDs, nil, sq.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetStatQueue is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetStatQueue(sq *engine.StatQueueWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetStatQueueDrv(nil, sq.StatQueue); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(sq.APIOpts[utils.CacheOpt]),
		sq.Tenant, utils.CacheStatQueues, sq.TenantID(), nil, nil, sq.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetFilter is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetFilter(fltr *engine.FilterWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetFilterDrv(context.TODO(), fltr.Filter); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(fltr.APIOpts[utils.CacheOpt]),
		fltr.Tenant, utils.CacheFilters, fltr.TenantID(), nil, nil, fltr.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetTiming is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetTiming(tm *utils.TPTimingWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetTimingDrv(tm.TPTiming); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(tm.APIOpts[utils.CacheOpt]),
		tm.Tenant, utils.CacheTimings, tm.ID, nil, nil, tm.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetResourceProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetResourceProfile(rs *engine.ResourceProfileWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetResourceProfileDrv(rs.ResourceProfile); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(rs.APIOpts[utils.CacheOpt]),
		rs.Tenant, utils.CacheResourceProfiles, rs.TenantID(), &rs.FilterIDs, nil, rs.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetResource is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetResource(rs *engine.ResourceWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetResourceDrv(rs.Resource); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(rs.APIOpts[utils.CacheOpt]),
		rs.Tenant, utils.CacheResources, rs.TenantID(), nil, nil, rs.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetRouteProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetRouteProfile(sp *engine.RouteProfileWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetRouteProfileDrv(sp.RouteProfile); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(sp.APIOpts[utils.CacheOpt]),
		sp.Tenant, utils.CacheRouteProfiles, sp.TenantID(), &sp.FilterIDs, nil, sp.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetAttributeProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetAttributeProfile(ap *engine.AttributeProfileWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetAttributeProfileDrv(context.TODO(), ap.AttributeProfile); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(ap.APIOpts[utils.CacheOpt]),
		ap.Tenant, utils.CacheAttributeProfiles, ap.TenantID(), &ap.FilterIDs, ap.Contexts, ap.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetChargerProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetChargerProfile(cp *engine.ChargerProfileWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetChargerProfileDrv(cp.ChargerProfile); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(cp.APIOpts[utils.CacheOpt]),
		cp.Tenant, utils.CacheChargerProfiles, cp.TenantID(), &cp.FilterIDs, nil, cp.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetDispatcherProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetDispatcherProfile(dpp *engine.DispatcherProfileWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetDispatcherProfileDrv(dpp.DispatcherProfile); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(dpp.APIOpts[utils.CacheOpt]),
		dpp.Tenant, utils.CacheDispatcherProfiles, dpp.TenantID(), &dpp.FilterIDs, dpp.Subsystems, dpp.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetDispatcherHost is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetDispatcherHost(dpp *engine.DispatcherHostWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetDispatcherHostDrv(dpp.DispatcherHost); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(dpp.APIOpts[utils.CacheOpt]),
		dpp.Tenant, utils.CacheDispatcherHosts, dpp.TenantID(), nil, nil, dpp.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetRateProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetRateProfile(dpp *utils.RateProfileWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetRateProfileDrv(dpp.RateProfile); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(dpp.APIOpts[utils.CacheOpt]),
		dpp.Tenant, utils.CacheRateProfiles, dpp.TenantID(), &dpp.FilterIDs, nil, dpp.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetActionProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetActionProfile(acp *engine.ActionProfileWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetActionProfileDrv(acp.ActionProfile); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(acp.APIOpts[utils.CacheOpt]),
		acp.Tenant, utils.CacheActionProfiles, acp.TenantID(), &acp.FilterIDs, nil, acp.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetAccount is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetAccount(acp *utils.AccountWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetAccountDrv(acp.Account); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(acp.APIOpts[utils.CacheOpt]),
		acp.Tenant, utils.CacheAccounts, acp.TenantID(), &acp.FilterIDs, nil, acp.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetLoadIDs is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetLoadIDs(args *utils.LoadIDsWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetLoadIDsDrv(context.TODO(), args.LoadIDs); err != nil {
		return
	}
	lIDs := make([]string, 0, len(args.LoadIDs))
	for lID := range args.LoadIDs {
		lIDs = append(lIDs, lID)
	}
	if err = rplSv1.v1.callCacheMultiple(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]),
		args.Tenant, utils.CacheLoadIDs, lIDs, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// SetIndexes is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) SetIndexes(args *utils.SetIndexesArg, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().SetIndexesDrv(context.TODO(), args.IdxItmType, args.TntCtx, args.Indexes, true, utils.NonTransactional); err != nil {
		return
	}
	cIDs := make([]string, 0, len(args.Indexes))
	for idxKey := range args.Indexes {
		cIDs = append(cIDs, utils.ConcatenatedKey(args.TntCtx, idxKey))
	}
	if err = rplSv1.v1.callCacheMultiple(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]),
		args.Tenant, args.IdxItmType, cIDs, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveThreshold is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveThreshold(args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveThresholdDrv(args.Tenant, args.ID); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]),
		args.Tenant, utils.CacheThresholds, args.TenantID.TenantID(), nil, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveDestination is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveDestination(id *utils.StringWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveDestinationDrv(id.Arg, utils.NonTransactional); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(id.APIOpts[utils.CacheOpt]),
		id.Tenant, utils.CacheDestinations, id.Arg, nil, nil, id.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveStatQueue is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveStatQueue(args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemStatQueueDrv(args.Tenant, args.ID); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]),
		args.Tenant, utils.CacheStatQueues, args.TenantID.TenantID(), nil, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveFilter is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveFilter(args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveFilterDrv(args.Tenant, args.ID); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]),
		args.Tenant, utils.CacheFilters, args.TenantID.TenantID(), nil, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveThresholdProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveThresholdProfile(args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemThresholdProfileDrv(args.Tenant, args.ID); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]),
		args.Tenant, utils.CacheThresholdProfiles, args.TenantID.TenantID(), nil, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveStatQueueProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveStatQueueProfile(args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemStatQueueProfileDrv(args.Tenant, args.ID); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]),
		args.Tenant, utils.CacheStatQueueProfiles, args.TenantID.TenantID(), nil, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveTiming is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveTiming(id *utils.StringWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveTimingDrv(id.Arg); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(id.APIOpts[utils.CacheOpt]),
		id.Tenant, utils.CacheTimings, id.Arg, nil, nil, id.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveResource is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveResource(args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveResourceDrv(args.Tenant, args.ID); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]),
		args.Tenant, utils.CacheResources, args.TenantID.TenantID(), nil, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveResourceProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveResourceProfile(args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveResourceProfileDrv(args.Tenant, args.ID); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]),
		args.Tenant, utils.CacheResourceProfiles, args.TenantID.TenantID(), nil, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveRouteProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveRouteProfile(args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveRouteProfileDrv(args.Tenant, args.ID); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]),
		args.Tenant, utils.CacheRouteProfiles, args.TenantID.TenantID(), nil, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveAttributeProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveAttributeProfile(args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveAttributeProfileDrv(context.TODO(), args.Tenant, args.ID); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]),
		args.Tenant, utils.CacheAttributeProfiles, args.TenantID.TenantID(), nil, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveChargerProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveChargerProfile(args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveChargerProfileDrv(args.Tenant, args.ID); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]),
		args.Tenant, utils.CacheChargerProfiles, args.TenantID.TenantID(), nil, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveDispatcherProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveDispatcherProfile(args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveDispatcherProfileDrv(args.Tenant, args.ID); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]),
		args.Tenant, utils.CacheDispatcherProfiles, args.TenantID.TenantID(), nil, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveRateProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveRateProfile(args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveRateProfileDrv(args.Tenant, args.ID); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]),
		args.Tenant, utils.CacheRateProfiles, args.TenantID.TenantID(), nil, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveActionProfile is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveActionProfile(args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveActionProfileDrv(args.Tenant, args.ID); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]),
		args.Tenant, utils.CacheActionProfiles, args.TenantID.TenantID(), nil, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveAccount is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveAccount(args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveAccountDrv(args.Tenant, args.ID); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]),
		args.Tenant, utils.CacheAccounts, args.TenantID.TenantID(), nil, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveDispatcherHost is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveDispatcherHost(args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveDispatcherHostDrv(args.Tenant, args.ID); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]),
		args.Tenant, utils.CacheDispatcherHosts, args.TenantID.TenantID(), nil, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RemoveIndexes  is the replication method coresponding to the dataDb driver method
func (rplSv1 *ReplicatorSv1) RemoveIndexes(args *utils.GetIndexesArg, reply *string) (err error) {
	if err = rplSv1.dm.DataDB().RemoveIndexesDrv(args.IdxItmType, args.TntCtx, args.IdxKey); err != nil {
		return
	}
	if err = rplSv1.v1.CallCache(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]),
		args.Tenant, args.IdxItmType, utils.ConcatenatedKey(args.TntCtx, args.IdxKey), nil, nil, args.APIOpts); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// Ping used to determine if the RPC is active
func (rplSv1 *ReplicatorSv1) Ping(ign *utils.CGREvent, reply *string) error {
	*reply = utils.Pong
	return nil
}
