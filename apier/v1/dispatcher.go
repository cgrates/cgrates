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

	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

// GetDispatcherProfile returns a Dispatcher Profile
func (apierV1 *ApierV1) GetDispatcherProfile(arg *utils.TenantID, reply *engine.DispatcherProfile) error {
	if missing := utils.MissingStructFields(arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if dpp, err := apierV1.DataManager.GetDispatcherProfile(arg.Tenant, arg.ID, true, true, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	} else {
		*reply = *dpp
	}
	return nil
}

// GetDispatcherProfileIDs returns list of dispatcherProfile IDs registered for a tenant
func (apierV1 *ApierV1) GetDispatcherProfileIDs(tenant string, dPrfIDs *[]string) error {
	prfx := utils.DispatcherProfilePrefix + tenant + ":"
	keys, err := apierV1.DataManager.DataDB().GetKeysForPrefix(prfx)
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
	*dPrfIDs = retIDs
	return nil
}

type DispatcherWithCache struct {
	*engine.DispatcherProfile
	Cache *string
}

//SetDispatcherProfile add/update a new Dispatcher Profile
func (apierV1 *ApierV1) SetDispatcherProfile(args *DispatcherWithCache, reply *string) error {
	if missing := utils.MissingStructFields(args.DispatcherProfile, []string{"Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierV1.DataManager.SetDispatcherProfile(args.DispatcherProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheDispatcherProfiles and store it in database
	if err := apierV1.DataManager.SetLoadIDs(map[string]string{utils.CacheDispatcherProfiles: utils.UUIDSha1Prefix()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for DispatcherProfile
	argCache := engine.ArgsGetCacheItem{
		CacheID: utils.CacheDispatcherProfiles,
		ItemID:  args.TenantID(),
	}
	if err := apierV1.CallCache(GetCacheOpt(args.Cache), argCache); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

//RemoveDispatcherProfile remove a specific Dispatcher Profile
func (apierV1 *ApierV1) RemoveDispatcherProfile(arg *utils.TenantIDWrapper, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierV1.DataManager.RemoveDispatcherProfile(arg.Tenant,
		arg.ID, utils.NonTransactional, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheDispatcherProfiles and store it in database
	if err := apierV1.DataManager.SetLoadIDs(map[string]string{utils.CacheDispatcherProfiles: utils.UUIDSha1Prefix()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for DispatcherProfile
	argCache := engine.ArgsGetCacheItem{
		CacheID: utils.CacheDispatcherProfiles,
		ItemID:  arg.TenantID(),
	}
	if err := apierV1.CallCache(GetCacheOpt(arg.Cache), argCache); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// GetDispatcherHost returns a Dispatcher Host
func (apierV1 *ApierV1) GetDispatcherHost(arg *utils.TenantID, reply *engine.DispatcherHost) error {
	if missing := utils.MissingStructFields(arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if dpp, err := apierV1.DataManager.GetDispatcherHost(arg.Tenant, arg.ID, true, true, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	} else {
		*reply = *dpp
	}
	return nil
}

// GetDispatcherHostIDs returns list of dispatcherHost IDs registered for a tenant
func (apierV1 *ApierV1) GetDispatcherHostIDs(tenant string, dPrfIDs *[]string) error {
	prfx := utils.DispatcherHostPrefix + tenant + ":"
	keys, err := apierV1.DataManager.DataDB().GetKeysForPrefix(prfx)
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
	*dPrfIDs = retIDs
	return nil
}

type DispatcherHostWrapper struct {
	*engine.DispatcherHost
	Cache *string
}

//SetDispatcherHost add/update a new Dispatcher Host
func (apierV1 *ApierV1) SetDispatcherHost(args *DispatcherHostWrapper, reply *string) error {
	if missing := utils.MissingStructFields(args.DispatcherHost, []string{"Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierV1.DataManager.SetDispatcherHost(args.DispatcherHost); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheDispatcherHosts and store it in database
	if err := apierV1.DataManager.SetLoadIDs(map[string]string{utils.CacheDispatcherHosts: utils.UUIDSha1Prefix()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for DispatcherProfile
	argCache := engine.ArgsGetCacheItem{
		CacheID: utils.CacheDispatcherHosts,
		ItemID:  args.TenantID(),
	}
	if err := apierV1.CallCache(GetCacheOpt(args.Cache), argCache); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

//RemoveDispatcherHost remove a specific Dispatcher Host
func (apierV1 *ApierV1) RemoveDispatcherHost(arg *utils.TenantIDWrapper, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierV1.DataManager.RemoveDispatcherHost(arg.Tenant,
		arg.ID, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheDispatcherHosts and store it in database
	if err := apierV1.DataManager.SetLoadIDs(map[string]string{utils.CacheDispatcherHosts: utils.UUIDSha1Prefix()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for DispatcherProfile
	argCache := engine.ArgsGetCacheItem{
		CacheID: utils.CacheDispatcherHosts,
		ItemID:  arg.TenantID(),
	}
	if err := apierV1.CallCache(GetCacheOpt(arg.Cache), argCache); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

func NewDispatcherThresholdSv1(dps *dispatchers.DispatcherService) *DispatcherThresholdSv1 {
	return &DispatcherThresholdSv1{dS: dps}
}

// Exports RPC from RLs
type DispatcherThresholdSv1 struct {
	dS *dispatchers.DispatcherService
}

// Ping implements ThresholdSv1Ping
func (dT *DispatcherThresholdSv1) Ping(args *dispatchers.CGREvWithApiKey, reply *string) error {
	return dT.dS.ThresholdSv1Ping(args, reply)
}

// GetThresholdsForEvent implements ThresholdSv1GetThresholdsForEvent
func (dT *DispatcherThresholdSv1) GetThresholdsForEvent(tntID *dispatchers.ArgsProcessEventWithApiKey,
	t *engine.Thresholds) error {
	return dT.dS.ThresholdSv1GetThresholdsForEvent(tntID, t)
}

// ProcessEvent implements ThresholdSv1ProcessEvent
func (dT *DispatcherThresholdSv1) ProcessEvent(args *dispatchers.ArgsProcessEventWithApiKey,
	tIDs *[]string) error {
	return dT.dS.ThresholdSv1ProcessEvent(args, tIDs)
}

func (dT *DispatcherThresholdSv1) GetThresholdIDs(args *dispatchers.TntWithApiKey,
	tIDs *[]string) error {
	return dT.dS.ThresholdSv1GetThresholdIDs(args, tIDs)
}

func (dT *DispatcherThresholdSv1) GetThreshold(args *dispatchers.TntIDWithApiKey,
	th *engine.Threshold) error {
	return dT.dS.ThresholdSv1GetThreshold(args, th)
}

func NewDispatcherStatSv1(dps *dispatchers.DispatcherService) *DispatcherStatSv1 {
	return &DispatcherStatSv1{dS: dps}
}

// Exports RPC from RLs
type DispatcherStatSv1 struct {
	dS *dispatchers.DispatcherService
}

// Ping implements StatSv1Ping
func (dSts *DispatcherStatSv1) Ping(args *dispatchers.CGREvWithApiKey, reply *string) error {
	return dSts.dS.StatSv1Ping(args, reply)
}

// GetStatQueuesForEvent implements StatSv1GetStatQueuesForEvent
func (dSts *DispatcherStatSv1) GetStatQueuesForEvent(args *dispatchers.ArgsStatProcessEventWithApiKey, reply *[]string) error {
	return dSts.dS.StatSv1GetStatQueuesForEvent(args, reply)
}

// GetQueueStringMetrics implements StatSv1GetQueueStringMetrics
func (dSts *DispatcherStatSv1) GetQueueStringMetrics(args *dispatchers.TntIDWithApiKey,
	reply *map[string]string) error {
	return dSts.dS.StatSv1GetQueueStringMetrics(args, reply)
}

func (dSts *DispatcherStatSv1) GetQueueFloatMetrics(args *dispatchers.TntIDWithApiKey,
	reply *map[string]float64) error {
	return dSts.dS.StatSv1GetQueueFloatMetrics(args, reply)
}

func (dSts *DispatcherStatSv1) GetQueueIDs(args *dispatchers.TntWithApiKey,
	reply *[]string) error {
	return dSts.dS.StatSv1GetQueueIDs(args, reply)
}

// GetQueueStringMetrics implements StatSv1ProcessEvent
func (dSts *DispatcherStatSv1) ProcessEvent(args *dispatchers.ArgsStatProcessEventWithApiKey, reply *[]string) error {
	return dSts.dS.StatSv1ProcessEvent(args, reply)
}

func NewDispatcherResourceSv1(dps *dispatchers.DispatcherService) *DispatcherResourceSv1 {
	return &DispatcherResourceSv1{dRs: dps}
}

// Exports RPC from RLs
type DispatcherResourceSv1 struct {
	dRs *dispatchers.DispatcherService
}

// Ping implements ResourceSv1Ping
func (dRs *DispatcherResourceSv1) Ping(args *dispatchers.CGREvWithApiKey, reply *string) error {
	return dRs.dRs.ResourceSv1Ping(args, reply)
}

// GetResourcesForEvent implements ResourceSv1GetResourcesForEvent
func (dRs *DispatcherResourceSv1) GetResourcesForEvent(args *dispatchers.ArgsV1ResUsageWithApiKey,
	reply *engine.Resources) error {
	return dRs.dRs.ResourceSv1GetResourcesForEvent(args, reply)
}

func (dRs *DispatcherResourceSv1) AuthorizeResources(args *dispatchers.ArgsV1ResUsageWithApiKey,
	reply *string) error {
	return dRs.dRs.ResourceSv1AuthorizeResources(args, reply)
}

func (dRs *DispatcherResourceSv1) AllocateResources(args *dispatchers.ArgsV1ResUsageWithApiKey,
	reply *string) error {
	return dRs.dRs.ResourceSv1AllocateResources(args, reply)
}

func (dRs *DispatcherResourceSv1) ReleaseResources(args *dispatchers.ArgsV1ResUsageWithApiKey,
	reply *string) error {
	return dRs.dRs.ResourceSv1ReleaseResources(args, reply)
}

func NewDispatcherSupplierSv1(dps *dispatchers.DispatcherService) *DispatcherSupplierSv1 {
	return &DispatcherSupplierSv1{dSup: dps}
}

// Exports RPC from RLs
type DispatcherSupplierSv1 struct {
	dSup *dispatchers.DispatcherService
}

// Ping implements SupplierSv1Ping
func (dSup *DispatcherSupplierSv1) Ping(args *dispatchers.CGREvWithApiKey, reply *string) error {
	return dSup.dSup.SupplierSv1Ping(args, reply)
}

// GetSuppliers implements SupplierSv1GetSuppliers
func (dSup *DispatcherSupplierSv1) GetSuppliers(args *dispatchers.ArgsGetSuppliersWithApiKey,
	reply *engine.SortedSuppliers) error {
	return dSup.dSup.SupplierSv1GetSuppliers(args, reply)
}

func NewDispatcherAttributeSv1(dps *dispatchers.DispatcherService) *DispatcherAttributeSv1 {
	return &DispatcherAttributeSv1{dA: dps}
}

// Exports RPC from RLs
type DispatcherAttributeSv1 struct {
	dA *dispatchers.DispatcherService
}

// Ping implements SupplierSv1Ping
func (dA *DispatcherAttributeSv1) Ping(args *dispatchers.CGREvWithApiKey, reply *string) error {
	return dA.dA.AttributeSv1Ping(args, reply)
}

// GetAttributeForEvent implements AttributeSv1GetAttributeForEvent
func (dA *DispatcherAttributeSv1) GetAttributeForEvent(args *dispatchers.ArgsAttrProcessEventWithApiKey,
	reply *engine.AttributeProfile) error {
	return dA.dA.AttributeSv1GetAttributeForEvent(args, reply)
}

// ProcessEvent implements AttributeSv1ProcessEvent
func (dA *DispatcherAttributeSv1) ProcessEvent(args *dispatchers.ArgsAttrProcessEventWithApiKey,
	reply *engine.AttrSProcessEventReply) error {
	return dA.dA.AttributeSv1ProcessEvent(args, reply)
}

func NewDispatcherChargerSv1(dps *dispatchers.DispatcherService) *DispatcherChargerSv1 {
	return &DispatcherChargerSv1{dC: dps}
}

// Exports RPC from RLs
type DispatcherChargerSv1 struct {
	dC *dispatchers.DispatcherService
}

// Ping implements ChargerSv1Ping
func (dC *DispatcherChargerSv1) Ping(args *dispatchers.CGREvWithApiKey, reply *string) error {
	return dC.dC.ChargerSv1Ping(args, reply)
}

// GetChargersForEvent implements ChargerSv1GetChargersForEvent
func (dC *DispatcherChargerSv1) GetChargersForEvent(args *dispatchers.CGREvWithApiKey,
	reply *engine.ChargerProfiles) (err error) {
	return dC.dC.ChargerSv1GetChargersForEvent(args, reply)
}

// ProcessEvent implements ChargerSv1ProcessEvent
func (dC *DispatcherChargerSv1) ProcessEvent(args *dispatchers.CGREvWithApiKey,
	reply *[]*engine.ChrgSProcessEventReply) (err error) {
	return dC.dC.ChargerSv1ProcessEvent(args, reply)
}

func NewDispatcherSessionSv1(dps *dispatchers.DispatcherService) *DispatcherSessionSv1 {
	return &DispatcherSessionSv1{dS: dps}
}

// Exports RPC from RLs
type DispatcherSessionSv1 struct {
	dS *dispatchers.DispatcherService
}

// Ping implements SessionSv1Ping
func (dS *DispatcherSessionSv1) Ping(args *dispatchers.CGREvWithApiKey, reply *string) error {
	return dS.dS.SessionSv1Ping(args, reply)
}

// AuthorizeEventWithDigest implements SessionSv1AuthorizeEventWithDigest
func (dS *DispatcherSessionSv1) AuthorizeEventWithDigest(args *dispatchers.AuthorizeArgsWithApiKey,
	reply *sessions.V1AuthorizeReplyWithDigest) error {
	return dS.dS.SessionSv1AuthorizeEventWithDigest(args, reply)
}

func (dS *DispatcherSessionSv1) AuthorizeEvent(args *dispatchers.AuthorizeArgsWithApiKey,
	reply *sessions.V1AuthorizeReply) error {
	return dS.dS.SessionSv1AuthorizeEvent(args, reply)
}

// InitiateSessionWithDigest implements SessionSv1InitiateSessionWithDigest
func (dS *DispatcherSessionSv1) InitiateSessionWithDigest(args *dispatchers.InitArgsWithApiKey,
	reply *sessions.V1InitReplyWithDigest) (err error) {
	return dS.dS.SessionSv1InitiateSessionWithDigest(args, reply)
}

// InitiateSessionWithDigest implements SessionSv1InitiateSessionWithDigest
func (dS *DispatcherSessionSv1) InitiateSession(args *dispatchers.InitArgsWithApiKey,
	reply *sessions.V1InitSessionReply) (err error) {
	return dS.dS.SessionSv1InitiateSession(args, reply)
}

// ProcessCDR implements SessionSv1ProcessCDR
func (dS *DispatcherSessionSv1) ProcessCDR(args *dispatchers.CGREvWithApiKey,
	reply *string) (err error) {
	return dS.dS.SessionSv1ProcessCDR(args, reply)
}

// ProcessEvent implements SessionSv1ProcessEvent
func (dS *DispatcherSessionSv1) ProcessEvent(args *dispatchers.ProcessEventWithApiKey,
	reply *sessions.V1ProcessEventReply) (err error) {
	return dS.dS.SessionSv1ProcessEvent(args, reply)
}

// TerminateSession implements SessionSv1TerminateSession
func (dS *DispatcherSessionSv1) TerminateSession(args *dispatchers.TerminateSessionWithApiKey,
	reply *string) (err error) {
	return dS.dS.SessionSv1TerminateSession(args, reply)
}

// UpdateSession implements SessionSv1UpdateSession
func (dS *DispatcherSessionSv1) UpdateSession(args *dispatchers.UpdateSessionWithApiKey,
	reply *sessions.V1UpdateSessionReply) (err error) {
	return dS.dS.SessionSv1UpdateSession(args, reply)
}

func (dS *DispatcherSessionSv1) GetActiveSessions(args *dispatchers.FilterSessionWithApiKey,
	reply *[]*sessions.ActiveSession) (err error) {
	return dS.dS.SessionSv1GetActiveSessions(args, reply)
}

func (dS *DispatcherSessionSv1) GetActiveSessionsCount(args *dispatchers.FilterSessionWithApiKey,
	reply *int) (err error) {
	return dS.dS.SessionSv1GetActiveSessionsCount(args, reply)
}

func (dS *DispatcherSessionSv1) ForceDisconnect(args *dispatchers.FilterSessionWithApiKey,
	reply *string) (err error) {
	return dS.dS.SessionSv1ForceDisconnect(args, reply)
}

func (dS *DispatcherSessionSv1) GetPassiveSessions(args *dispatchers.FilterSessionWithApiKey,
	reply *[]*sessions.ActiveSession) (err error) {
	return dS.dS.SessionSv1GetPassiveSessions(args, reply)
}

func (dS *DispatcherSessionSv1) GetPassiveSessionsCount(args *dispatchers.FilterSessionWithApiKey,
	reply *int) (err error) {
	return dS.dS.SessionSv1GetPassiveSessionsCount(args, reply)
}

func (dS *DispatcherSessionSv1) ReplicateSessions(args *dispatchers.ArgsReplicateSessionsWithApiKey,
	reply *string) (err error) {
	return dS.dS.SessionSv1ReplicateSessions(args, reply)
}

func (dS *DispatcherSessionSv1) SetPassiveSession(args *dispatchers.SessionWithApiKey,
	reply *string) (err error) {
	return dS.dS.SessionSv1SetPassiveSession(args, reply)
}

func NewDispatcherResponder(dps *dispatchers.DispatcherService) *DispatcherResponder {
	return &DispatcherResponder{dS: dps}
}

// Exports RPC from RLs
type DispatcherResponder struct {
	dS *dispatchers.DispatcherService
}

func (dS *DispatcherResponder) Status(args *dispatchers.TntWithApiKey, reply *map[string]interface{}) error {
	return dS.dS.ResponderStatus(args, reply)
}

func (dS *DispatcherResponder) GetCost(args *dispatchers.CallDescriptorWithApiKey, reply *engine.CallCost) error {
	return dS.dS.ResponderGetCost(args, reply)
}

func (dS *DispatcherResponder) Debit(args *dispatchers.CallDescriptorWithApiKey, reply *engine.CallCost) error {
	return dS.dS.ResponderDebit(args, reply)
}

func (dS *DispatcherResponder) MaxDebit(args *dispatchers.CallDescriptorWithApiKey, reply *engine.CallCost) error {
	return dS.dS.ResponderMaxDebit(args, reply)
}

func (dS *DispatcherResponder) RefundIncrements(args *dispatchers.CallDescriptorWithApiKey, reply *engine.Account) error {
	return dS.dS.ResponderRefundIncrements(args, reply)
}

func (dS *DispatcherResponder) RefundRounding(args *dispatchers.CallDescriptorWithApiKey, reply *float64) error {
	return dS.dS.ResponderRefundRounding(args, reply)
}

func (dS *DispatcherResponder) GetMaxSessionTime(args *dispatchers.CallDescriptorWithApiKey, reply *time.Duration) error {
	return dS.dS.ResponderGetMaxSessionTime(args, reply)
}

func (dS *DispatcherResponder) Shutdown(args *dispatchers.TntWithApiKey, reply *string) error {
	return dS.dS.ResponderShutdown(args, reply)
}

func (dS *DispatcherResponder) GetTimeout(args *dispatchers.TntWithApiKey, reply *time.Duration) error {
	return dS.dS.ResponderGetTimeout(args, reply)
}

// Ping used to detreminate if component is active
func (dS *DispatcherResponder) Ping(args *dispatchers.CGREvWithApiKey, reply *string) error {
	return dS.dS.ResponderPing(args, reply)
}

func NewDispatcherCacheSv1(dps *dispatchers.DispatcherService) *DispatcherCacheSv1 {
	return &DispatcherCacheSv1{dS: dps}
}

// Exports RPC from CacheSv1
type DispatcherCacheSv1 struct {
	dS *dispatchers.DispatcherService
}

// GetItemIDs returns the IDs for cacheID with given prefix
func (dS *DispatcherCacheSv1) GetItemIDs(args *dispatchers.ArgsGetCacheItemIDsWithApiKey,
	reply *[]string) error {
	return dS.dS.CacheSv1GetItemIDs(args, reply)
}

// HasItem verifies the existence of an Item in cache
func (dS *DispatcherCacheSv1) HasItem(args *dispatchers.ArgsGetCacheItemWithApiKey,
	reply *bool) error {
	return dS.dS.CacheSv1HasItem(args, reply)
}

// GetItemExpiryTime returns the expiryTime for an item
func (dS *DispatcherCacheSv1) GetItemExpiryTime(args *dispatchers.ArgsGetCacheItemWithApiKey,
	reply *time.Time) error {
	return dS.dS.CacheSv1GetItemExpiryTime(args, reply)
}

// RemoveItem removes the Item with ID from cache
func (dS *DispatcherCacheSv1) RemoveItem(args *dispatchers.ArgsGetCacheItemWithApiKey,
	reply *string) error {
	return dS.dS.CacheSv1RemoveItem(args, reply)
}

// Clear will clear partitions in the cache (nil fol all, empty slice for none)
func (dS *DispatcherCacheSv1) Clear(args *dispatchers.AttrCacheIDsWithApiKey,
	reply *string) error {
	return dS.dS.CacheSv1Clear(args, reply)
}

// FlushCache wipes out cache for a prefix or completely
func (dS *DispatcherCacheSv1) FlushCache(args dispatchers.AttrReloadCacheWithApiKey, reply *string) (err error) {
	return dS.dS.CacheSv1FlushCache(args, reply)
}

// GetCacheStats returns CacheStats filtered by cacheIDs
func (dS *DispatcherCacheSv1) GetCacheStats(args *dispatchers.AttrCacheIDsWithApiKey,
	reply *map[string]*ltcache.CacheStats) error {
	return dS.dS.CacheSv1GetCacheStats(args, reply)
}

// PrecacheStatus checks status of active precache processes
func (dS *DispatcherCacheSv1) PrecacheStatus(args *dispatchers.AttrCacheIDsWithApiKey, reply *map[string]string) error {
	return dS.dS.CacheSv1PrecacheStatus(args, reply)
}

// HasGroup checks existence of a group in cache
func (dS *DispatcherCacheSv1) HasGroup(args *dispatchers.ArgsGetGroupWithApiKey,
	reply *bool) (err error) {
	return dS.dS.CacheSv1HasGroup(args, reply)
}

// GetGroupItemIDs returns a list of itemIDs in a cache group
func (dS *DispatcherCacheSv1) GetGroupItemIDs(args *dispatchers.ArgsGetGroupWithApiKey,
	reply *[]string) (err error) {
	return dS.dS.CacheSv1GetGroupItemIDs(args, reply)
}

// RemoveGroup will remove a group and all items belonging to it from cache
func (dS *DispatcherCacheSv1) RemoveGroup(args *dispatchers.ArgsGetGroupWithApiKey,
	reply *string) (err error) {
	return dS.dS.CacheSv1RemoveGroup(args, reply)
}

// ReloadCache reloads cache from DB for a prefix or completely
func (dS *DispatcherCacheSv1) ReloadCache(args dispatchers.AttrReloadCacheWithApiKey, reply *string) (err error) {
	return dS.dS.CacheSv1ReloadCache(args, reply)
}

// LoadCache loads cache from DB for a prefix or completely
func (dS *DispatcherCacheSv1) LoadCache(args dispatchers.AttrReloadCacheWithApiKey, reply *string) (err error) {
	return dS.dS.CacheSv1LoadCache(args, reply)
}

// Ping used to detreminate if component is active
func (dS *DispatcherCacheSv1) Ping(args *dispatchers.CGREvWithApiKey, reply *string) error {
	return dS.dS.CacheSv1Ping(args, reply)
}

func NewDispatcherGuardianSv1(dps *dispatchers.DispatcherService) *DispatcherGuardianSv1 {
	return &DispatcherGuardianSv1{dS: dps}
}

// Exports RPC from CacheSv1
type DispatcherGuardianSv1 struct {
	dS *dispatchers.DispatcherService
}

// RemoteLock will lock a key from remote
func (dS *DispatcherGuardianSv1) RemoteLock(attr *dispatchers.AttrRemoteLockWithApiKey, reply *string) (err error) {
	return dS.dS.GuardianSv1RemoteLock(attr, reply)
}

// RemoteUnlock will unlock a key from remote based on reference ID
func (dS *DispatcherGuardianSv1) RemoteUnlock(attr *dispatchers.AttrRemoteUnlockWithApiKey, reply *[]string) (err error) {
	return dS.dS.GuardianSv1RemoteUnlock(attr, reply)
}

// Ping used to detreminate if component is active
func (dS *DispatcherGuardianSv1) Ping(args *dispatchers.CGREvWithApiKey, reply *string) error {
	return dS.dS.GuardianSv1Ping(args, reply)
}
