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
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// GetStatQueueProfile returns a StatQueue profile
func (apierSv1 *APIerSv1) GetStatQueueProfile(arg *utils.TenantID, reply *engine.StatQueueProfile) (err error) {
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
func (apierSv1 *APIerSv1) GetStatQueueProfileIDs(args *utils.PaginatorWithTenant, stsPrfIDs *[]string) error {
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
func (apierSv1 *APIerSv1) SetStatQueueProfile(arg *engine.StatQueueWithCache, reply *string) (err error) {
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
	//handle caching for StatQueueProfile
	if err = apierSv1.CallCache(utils.IfaceAsString(arg.Opts[utils.CacheOpt]), arg.Tenant, utils.CacheStatQueueProfiles,
		arg.TenantID(), &arg.FilterIDs, nil, arg.Opts); err != nil {
		return utils.APIErrorHandler(err)
	}
	var ttl *time.Duration
	if arg.TTL > 0 {
		ttl = &arg.TTL
	}
	sq := &engine.StatQueue{
		Tenant: arg.Tenant,
		ID:     arg.ID,
	}
	if !arg.Stored { // for not stored queues create the metrics
		if sq, err = engine.NewStatQueue(arg.Tenant, arg.ID, arg.Metrics,
			arg.MinItems); err != nil {
			return err
		}
	}
	// for non stored we do not save the metrics
	if err = apierSv1.DataManager.SetStatQueue(sq,
		arg.Metrics, arg.MinItems, ttl, arg.QueueLength,
		!arg.Stored); err != nil {
		return err
	}
	//handle caching for StatQueues
	if err := apierSv1.CallCache(utils.IfaceAsString(arg.Opts[utils.CacheOpt]), arg.Tenant, utils.CacheStatQueues,
		arg.TenantID(), nil, nil, arg.Opts); err != nil {
		return utils.APIErrorHandler(err)
	}

	*reply = utils.OK
	return nil
}

// RemoveStatQueueProfile remove a specific stat configuration
func (apierSv1 *APIerSv1) RemoveStatQueueProfile(args *utils.TenantIDWithCache, reply *string) error {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.DataManager.RemoveStatQueueProfile(tnt, args.ID, utils.NonTransactional, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for StatQueueProfile
	if err := apierSv1.CallCache(utils.IfaceAsString(args.Opts[utils.CacheOpt]), tnt, utils.CacheStatQueueProfiles,
		utils.ConcatenatedKey(tnt, args.ID), nil, nil, args.Opts); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := apierSv1.DataManager.RemoveStatQueue(tnt, args.ID, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheStatQueueProfiles and CacheStatQueues and store it in database
	//make 1 insert for both StatQueueProfile and StatQueue instead of 2
	loadID := time.Now().UnixNano()
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheStatQueueProfiles: loadID, utils.CacheStatQueues: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for StatQueues
	if err := apierSv1.CallCache(utils.IfaceAsString(args.Opts[utils.CacheOpt]), tnt, utils.CacheStatQueues,
		utils.ConcatenatedKey(tnt, args.ID), nil, nil, args.Opts); err != nil {
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

// Call implements rpcclient.ClientConnector interface for internal RPC
func (stsv1 *StatSv1) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(stsv1, serviceMethod, args, reply)
}

// GetQueueIDs returns list of queueIDs registered for a tenant
func (stsv1 *StatSv1) GetQueueIDs(tenant *utils.TenantWithOpts, qIDs *[]string) error {
	return stsv1.sS.V1GetQueueIDs(tenant.Tenant, qIDs)
}

// ProcessEvent returns processes a new Event
func (stsv1 *StatSv1) ProcessEvent(args *engine.StatsArgsProcessEvent, reply *[]string) error {
	return stsv1.sS.V1ProcessEvent(args, reply)
}

// GetStatQueuesForEvent returns the list of queues IDs in the system
func (stsv1 *StatSv1) GetStatQueuesForEvent(args *engine.StatsArgsProcessEvent, reply *[]string) (err error) {
	return stsv1.sS.V1GetStatQueuesForEvent(args, reply)
}

// GetStatQueue returns a StatQueue object
func (stsv1 *StatSv1) GetStatQueue(args *utils.TenantIDWithOpts, reply *engine.StatQueue) (err error) {
	return stsv1.sS.V1GetStatQueue(args, reply)
}

// GetQueueStringMetrics returns the string metrics for a Queue
func (stsv1 *StatSv1) GetQueueStringMetrics(args *utils.TenantIDWithOpts, reply *map[string]string) (err error) {
	return stsv1.sS.V1GetQueueStringMetrics(args.TenantID, reply)
}

// GetQueueFloatMetrics returns the float metrics for a Queue
func (stsv1 *StatSv1) GetQueueFloatMetrics(args *utils.TenantIDWithOpts, reply *map[string]float64) (err error) {
	return stsv1.sS.V1GetQueueFloatMetrics(args.TenantID, reply)
}

// ResetStatQueue resets the stat queue
func (stsv1 *StatSv1) ResetStatQueue(tntID *utils.TenantIDWithOpts, reply *string) error {
	return stsv1.sS.V1ResetStatQueue(tntID.TenantID, reply)
}

// Ping .
func (stsv1 *StatSv1) Ping(ign *utils.CGREvent, reply *string) error {
	*reply = utils.Pong
	return nil
}
