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
func (APIerSv1 *APIerSv1) GetStatQueueProfile(arg *utils.TenantID, reply *engine.StatQueueProfile) (err error) {
	if missing := utils.MissingStructFields(arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if sCfg, err := APIerSv1.DataManager.GetStatQueueProfile(arg.Tenant, arg.ID,
		true, true, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	} else {
		*reply = *sCfg
	}
	return
}

// GetStatQueueProfileIDs returns list of statQueueProfile IDs registered for a tenant
func (APIerSv1 *APIerSv1) GetStatQueueProfileIDs(args utils.TenantArgWithPaginator, stsPrfIDs *[]string) error {
	if missing := utils.MissingStructFields(&args, []string{utils.Tenant}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	prfx := utils.StatQueueProfilePrefix + args.Tenant + ":"
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
	*stsPrfIDs = args.PaginateStringSlice(retIDs)
	return nil
}

// SetStatQueueProfile alters/creates a StatQueueProfile
func (APIerSv1 *APIerSv1) SetStatQueueProfile(arg *engine.StatQueueWithCache, reply *string) error {
	if missing := utils.MissingStructFields(arg.StatQueueProfile, []string{"Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := APIerSv1.DataManager.SetStatQueueProfile(arg.StatQueueProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheStatQueueProfiles and CacheStatQueues and store it in database
	//make 1 insert for both StatQueueProfile and StatQueue instead of 2
	loadID := time.Now().UnixNano()
	if err := APIerSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheStatQueueProfiles: loadID, utils.CacheStatQueues: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for StatQueueProfile
	argCache := utils.ArgsGetCacheItem{
		CacheID: utils.CacheStatQueueProfiles,
		ItemID:  arg.TenantID(),
	}
	if err := APIerSv1.CallCache(arg.Tenant, GetCacheOpt(arg.Cache), argCache); err != nil {
		return utils.APIErrorHandler(err)
	}
	if has, err := APIerSv1.DataManager.HasData(utils.StatQueuePrefix, arg.ID, arg.Tenant); err != nil {
		return err
	} else if !has {
		//compose metrics for StatQueue
		metrics := make(map[string]engine.StatMetric)
		for _, metric := range arg.Metrics {
			if stsMetric, err := engine.NewStatMetric(metric.MetricID, arg.MinItems, metric.FilterIDs); err != nil {
				return utils.APIErrorHandler(err)
			} else {
				metrics[metric.MetricID] = stsMetric
			}
		}
		if err := APIerSv1.DataManager.SetStatQueue(&engine.StatQueue{Tenant: arg.Tenant, ID: arg.ID, SQMetrics: metrics}); err != nil {
			return utils.APIErrorHandler(err)
		}
		//handle caching for StatQueues
		argCache = utils.ArgsGetCacheItem{
			CacheID: utils.CacheStatQueues,
			ItemID:  arg.TenantID(),
		}
		if err := APIerSv1.CallCache(arg.Tenant, GetCacheOpt(arg.Cache), argCache); err != nil {
			return utils.APIErrorHandler(err)
		}
	}

	*reply = utils.OK
	return nil
}

// Remove a specific stat configuration
func (APIerSv1 *APIerSv1) RemoveStatQueueProfile(args *utils.TenantIDWithCache, reply *string) error {
	if missing := utils.MissingStructFields(args, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := APIerSv1.DataManager.RemoveStatQueueProfile(args.Tenant, args.ID, utils.NonTransactional, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for StatQueueProfile
	argCache := utils.ArgsGetCacheItem{
		CacheID: utils.CacheStatQueueProfiles,
		ItemID:  args.TenantID(),
	}
	if err := APIerSv1.CallCache(args.Tenant, GetCacheOpt(args.Cache), argCache); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := APIerSv1.DataManager.RemoveStatQueue(args.Tenant, args.ID, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheStatQueueProfiles and CacheStatQueues and store it in database
	//make 1 insert for both StatQueueProfile and StatQueue instead of 2
	loadID := time.Now().UnixNano()
	if err := APIerSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheStatQueueProfiles: loadID, utils.CacheStatQueues: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for StatQueues
	argCache = utils.ArgsGetCacheItem{
		CacheID: utils.CacheStatQueues,
		ItemID:  args.TenantID(),
	}
	if err := APIerSv1.CallCache(args.Tenant, GetCacheOpt(args.Cache), argCache); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// NewStatSV1 initializes StatSV1
func NewStatSv1(sS *engine.StatService) *StatSv1 {
	return &StatSv1{sS: sS}
}

// Exports RPC from RLs
type StatSv1 struct {
	sS *engine.StatService
}

// Call implements rpcclient.ClientConnector interface for internal RPC
func (stsv1 *StatSv1) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(stsv1, serviceMethod, args, reply)
}

// GetQueueIDs returns list of queueIDs registered for a tenant
func (stsv1 *StatSv1) GetQueueIDs(tenant *utils.TenantWithArgDispatcher, qIDs *[]string) error {
	return stsv1.sS.V1GetQueueIDs(tenant.Tenant, qIDs)
}

// ProcessEvent returns processes a new Event
func (stsv1 *StatSv1) ProcessEvent(args *engine.StatsArgsProcessEvent, reply *[]string) error {
	return stsv1.sS.V1ProcessEvent(args, reply)
}

// GetQueueIDs returns the list of queues IDs in the system
func (stsv1 *StatSv1) GetStatQueuesForEvent(args *engine.StatsArgsProcessEvent, reply *[]string) (err error) {
	return stsv1.sS.V1GetStatQueuesForEvent(args, reply)
}

// GetStatQueue returns a StatQueue object
func (stsv1 *StatSv1) GetStatQueue(args *utils.TenantIDWithArgDispatcher, reply *engine.StatQueue) (err error) {
	return stsv1.sS.V1GetStatQueue(args, reply)
}

// GetStringMetrics returns the string metrics for a Queue
func (stsv1 *StatSv1) GetQueueStringMetrics(args *utils.TenantIDWithArgDispatcher, reply *map[string]string) (err error) {
	return stsv1.sS.V1GetQueueStringMetrics(args.TenantID, reply)
}

// GetQueueFloatMetrics returns the float metrics for a Queue
func (stsv1 *StatSv1) GetQueueFloatMetrics(args *utils.TenantIDWithArgDispatcher, reply *map[string]float64) (err error) {
	return stsv1.sS.V1GetQueueFloatMetrics(args.TenantID, reply)
}

func (stSv1 *StatSv1) Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error {
	*reply = utils.Pong
	return nil
}
