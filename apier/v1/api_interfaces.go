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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

type ThresholdSv1Interface interface {
	GetThresholdIDs(tenant *utils.TenantWithArgDispatcher, tIDs *[]string) error
	GetThresholdsForEvent(args *engine.ArgsProcessEvent, reply *engine.Thresholds) error
	GetThreshold(tntID *utils.TenantIDWithArgDispatcher, t *engine.Threshold) error
	ProcessEvent(args *engine.ArgsProcessEvent, tIDs *[]string) error
	Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error
}

type StatSv1Interface interface {
	GetQueueIDs(tenant *utils.TenantWithArgDispatcher, qIDs *[]string) error
	ProcessEvent(args *engine.StatsArgsProcessEvent, reply *[]string) error
	GetStatQueuesForEvent(args *engine.StatsArgsProcessEvent, reply *[]string) (err error)
	GetQueueStringMetrics(args *utils.TenantIDWithArgDispatcher, reply *map[string]string) (err error)
	GetQueueFloatMetrics(args *utils.TenantIDWithArgDispatcher, reply *map[string]float64) (err error)
	Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error
}

type ResourceSv1Interface interface {
	GetResourcesForEvent(args utils.ArgRSv1ResourceUsage, reply *engine.Resources) error
	AuthorizeResources(args utils.ArgRSv1ResourceUsage, reply *string) error
	AllocateResources(args utils.ArgRSv1ResourceUsage, reply *string) error
	ReleaseResources(args utils.ArgRSv1ResourceUsage, reply *string) error
	GetResource(args *utils.TenantIDWithArgDispatcher, reply *engine.Resource) error
	Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error
}

type SupplierSv1Interface interface {
	GetSuppliers(args *engine.ArgsGetSuppliers, reply *engine.SortedSuppliers) error
	GetSupplierProfilesForEvent(args *utils.CGREventWithArgDispatcher, reply *[]*engine.SupplierProfile) error
	Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error
}

type AttributeSv1Interface interface {
	GetAttributeForEvent(args *engine.AttrArgsProcessEvent, reply *engine.AttributeProfile) (err error)
	ProcessEvent(args *engine.AttrArgsProcessEvent, reply *engine.AttrSProcessEventReply) error
	Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error
}

type ChargerSv1Interface interface {
	Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error
	GetChargersForEvent(cgrEv *utils.CGREventWithArgDispatcher, reply *engine.ChargerProfiles) error
	ProcessEvent(args *utils.CGREventWithArgDispatcher, reply *[]*engine.ChrgSProcessEventReply) error
}

type SessionSv1Interface interface {
	AuthorizeEvent(args *sessions.V1AuthorizeArgs, rply *sessions.V1AuthorizeReply) error
	AuthorizeEventWithDigest(args *sessions.V1AuthorizeArgs, rply *sessions.V1AuthorizeReplyWithDigest) error
	InitiateSession(args *sessions.V1InitSessionArgs, rply *sessions.V1InitSessionReply) error
	InitiateSessionWithDigest(args *sessions.V1InitSessionArgs, rply *sessions.V1InitReplyWithDigest) error
	UpdateSession(args *sessions.V1UpdateSessionArgs, rply *sessions.V1UpdateSessionReply) error
	SyncSessions(args *utils.TenantWithArgDispatcher, rply *string) error
	TerminateSession(args *sessions.V1TerminateSessionArgs, rply *string) error
	ProcessCDR(cgrEv *utils.CGREventWithArgDispatcher, rply *string) error
	ProcessMessage(args *sessions.V1ProcessMessageArgs, rply *sessions.V1ProcessMessageReply) error
	ProcessEvent(args *sessions.V1ProcessEventArgs, rply *sessions.V1ProcessEventReply) error
	GetCost(args *sessions.V1ProcessEventArgs, rply *sessions.V1GetCostReply) error
	GetActiveSessions(args *utils.SessionFilter, rply *[]*sessions.ExternalSession) error
	GetActiveSessionsCount(args *utils.SessionFilter, rply *int) error
	ForceDisconnect(args *utils.SessionFilter, rply *string) error
	GetPassiveSessions(args *utils.SessionFilter, rply *[]*sessions.ExternalSession) error
	GetPassiveSessionsCount(args *utils.SessionFilter, rply *int) error
	Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error
	ReplicateSessions(args dispatchers.ArgsReplicateSessionsWithApiKey, rply *string) error
	SetPassiveSession(args *sessions.Session, reply *string) error
	ActivateSessions(args *utils.SessionIDsWithArgsDispatcher, reply *string) error
	DeactivateSessions(args *utils.SessionIDsWithArgsDispatcher, reply *string) error
}

type ResponderInterface interface {
	GetCost(arg *engine.CallDescriptorWithArgDispatcher, reply *engine.CallCost) (err error)
	Debit(arg *engine.CallDescriptorWithArgDispatcher, reply *engine.CallCost) (err error)
	MaxDebit(arg *engine.CallDescriptorWithArgDispatcher, reply *engine.CallCost) (err error)
	RefundIncrements(arg *engine.CallDescriptorWithArgDispatcher, reply *engine.Account) (err error)
	RefundRounding(arg *engine.CallDescriptorWithArgDispatcher, reply *float64) (err error)
	GetMaxSessionTime(arg *engine.CallDescriptorWithArgDispatcher, reply *time.Duration) (err error)
	Shutdown(arg *utils.TenantWithArgDispatcher, reply *string) (err error)
	Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error
}

type CacheSv1Interface interface {
	GetItemIDs(args *utils.ArgsGetCacheItemIDsWithArgDispatcher, reply *[]string) error
	HasItem(args *utils.ArgsGetCacheItemWithArgDispatcher, reply *bool) error
	GetItemExpiryTime(args *utils.ArgsGetCacheItemWithArgDispatcher, reply *time.Time) error
	RemoveItem(args *utils.ArgsGetCacheItemWithArgDispatcher, reply *string) error
	Clear(cacheIDs *utils.AttrCacheIDsWithArgDispatcher, reply *string) error
	FlushCache(args utils.AttrReloadCacheWithArgDispatcher, reply *string) error
	GetCacheStats(cacheIDs *utils.AttrCacheIDsWithArgDispatcher, rply *map[string]*ltcache.CacheStats) error
	PrecacheStatus(cacheIDs *utils.AttrCacheIDsWithArgDispatcher, rply *map[string]string) error
	HasGroup(args *utils.ArgsGetGroupWithArgDispatcher, rply *bool) error
	GetGroupItemIDs(args *utils.ArgsGetGroupWithArgDispatcher, rply *[]string) error
	RemoveGroup(args *utils.ArgsGetGroupWithArgDispatcher, rply *string) error
	ReloadCache(attrs utils.AttrReloadCacheWithArgDispatcher, reply *string) error
	LoadCache(args utils.AttrReloadCacheWithArgDispatcher, reply *string) error
	Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error
}

type GuardianSv1Interface interface {
	RemoteLock(attr dispatchers.AttrRemoteLockWithApiKey, reply *string) (err error)
	RemoteUnlock(refID dispatchers.AttrRemoteUnlockWithApiKey, reply *[]string) (err error)
	Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error
}

type SchedulerSv1Interface interface {
	Reload(arg *utils.CGREventWithArgDispatcher, reply *string) error
	Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error
}

type CDRsV1Interface interface {
	ProcessCDR(cdr *engine.CDRWithArgDispatcher, reply *string) error
	ProcessEvent(arg *engine.ArgV1ProcessEvent, reply *string) error
	ProcessExternalCDR(cdr *engine.ExternalCDRWithArgDispatcher, reply *string) error
	RateCDRs(arg *engine.ArgRateCDRs, reply *string) error
	StoreSessionCost(attr *engine.AttrCDRSStoreSMCost, reply *string) error
	GetCDRsCount(args *utils.RPCCDRsFilterWithArgDispatcher, reply *int64) error
	GetCDRs(args utils.RPCCDRsFilterWithArgDispatcher, reply *[]*engine.CDR) error
	Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error
}

type ServiceManagerV1Interface interface {
	StartService(args dispatchers.ArgStartServiceWithApiKey, reply *string) error
	StopService(args dispatchers.ArgStartServiceWithApiKey, reply *string) error
	ServiceStatus(args dispatchers.ArgStartServiceWithApiKey, reply *string) error
	Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error
}

type RALsV1Interface interface {
	GetRatingPlansCost(arg *utils.RatingPlanCostArg, reply *dispatchers.RatingPlanCost) error
	Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error
}

type ConfigSv1Interface interface {
	GetJSONSection(section *config.StringWithArgDispatcher, reply *map[string]interface{}) (err error)
	ReloadConfigFromPath(section *config.ConfigReloadWithArgDispatcher, reply *string) (err error)
	ReloadConfigFromJSON(args *config.JSONReloadWithArgDispatcher, reply *string) (err error)
}

type CoreSv1Interface interface {
	Status(arg *utils.TenantWithArgDispatcher, reply *map[string]interface{}) error
	Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error
}

type ReplicatorSv1Interface interface {
	Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error
	GetAccount(args *utils.StringWithApiKey, reply *engine.Account) error
	GetDestination(key *utils.StringWithApiKey, reply *engine.Destination) error
	GetReverseDestination(key *utils.StringWithApiKey, reply *[]string) error
	GetStatQueue(tntID *utils.TenantIDWithArgDispatcher, reply *engine.StatQueue) error
	GetFilter(tntID *utils.TenantIDWithArgDispatcher, reply *engine.Filter) error
	GetThreshold(tntID *utils.TenantIDWithArgDispatcher, reply *engine.Threshold) error
	GetThresholdProfile(tntID *utils.TenantIDWithArgDispatcher, reply *engine.ThresholdProfile) error
	GetStatQueueProfile(tntID *utils.TenantIDWithArgDispatcher, reply *engine.StatQueueProfile) error
	GetTiming(id *utils.StringWithApiKey, reply *utils.TPTiming) error
	GetResource(tntID *utils.TenantIDWithArgDispatcher, reply *engine.Resource) error
	GetResourceProfile(tntID *utils.TenantIDWithArgDispatcher, reply *engine.ResourceProfile) error
	GetActionTriggers(id *utils.StringWithApiKey, reply *engine.ActionTriggers) error
	GetSharedGroup(id *utils.StringWithApiKey, reply *engine.SharedGroup) error
	GetActions(id *utils.StringWithApiKey, reply *engine.Actions) error
	GetActionPlan(id *utils.StringWithApiKey, reply *engine.ActionPlan) error
	GetAllActionPlans(_ *utils.StringWithApiKey, reply *map[string]*engine.ActionPlan) error
	GetAccountActionPlans(id *utils.StringWithApiKey, reply *[]string) error
	GetRatingPlan(id *utils.StringWithApiKey, reply *engine.RatingPlan) error
	GetRatingProfile(id *utils.StringWithApiKey, reply *engine.RatingProfile) error
	GetSupplierProfile(tntID *utils.TenantIDWithArgDispatcher, reply *engine.SupplierProfile) error
	GetAttributeProfile(tntID *utils.TenantIDWithArgDispatcher, reply *engine.AttributeProfile) error
	GetChargerProfile(tntID *utils.TenantIDWithArgDispatcher, reply *engine.ChargerProfile) error
	GetDispatcherProfile(tntID *utils.TenantIDWithArgDispatcher, reply *engine.DispatcherProfile) error
	GetDispatcherHost(tntID *utils.TenantIDWithArgDispatcher, reply *engine.DispatcherHost) error
	GetItemLoadIDs(itemID *utils.StringWithApiKey, reply *map[string]int64) error
	GetFilterIndexes(args *utils.GetFilterIndexesArgWithArgDispatcher, reply *map[string]utils.StringMap) error
	MatchFilterIndex(args *utils.MatchFilterIndexArgWithArgDispatcher, reply *utils.StringMap) error
	SetThresholdProfile(th *engine.ThresholdProfileWithArgDispatcher, reply *string) error
	SetThreshold(th *engine.ThresholdWithArgDispatcher, reply *string) error
	SetFilterIndexes(args *utils.SetFilterIndexesArgWithArgDispatcher, reply *string) error
	SetAccount(acc *engine.AccountWithArgDispatcher, reply *string) error
	SetDestination(dst *engine.DestinationWithArgDispatcher, reply *string) error
	SetReverseDestination(dst *engine.DestinationWithArgDispatcher, reply *string) error
	SetStatQueue(ssq *engine.StoredStatQueueWithArgDispatcher, reply *string) error
	SetFilter(fltr *engine.FilterWithArgDispatcher, reply *string) error
	SetStatQueueProfile(sq *engine.StatQueueProfileWithArgDispatcher, reply *string) error
	SetTiming(tm *utils.TPTimingWithArgDispatcher, reply *string) error
	SetResource(rs *engine.ResourceWithArgDispatcher, reply *string) error
	SetResourceProfile(rs *engine.ResourceProfileWithArgDispatcher, reply *string) error
	SetActionTriggers(args *engine.SetActionTriggersArgWithArgDispatcher, reply *string) error
	SetSharedGroup(shg *engine.SharedGroupWithArgDispatcher, reply *string) error
	SetActions(args *engine.SetActionsArgsWithArgDispatcher, reply *string) error
	SetRatingPlan(rp *engine.RatingPlanWithArgDispatcher, reply *string) error
	SetRatingProfile(rp *engine.RatingProfileWithArgDispatcher, reply *string) error
	SetSupplierProfile(sp *engine.SupplierProfileWithArgDispatcher, reply *string) error
	SetAttributeProfile(ap *engine.AttributeProfileWithArgDispatcher, reply *string) error
	SetChargerProfile(cp *engine.ChargerProfileWithArgDispatcher, reply *string) error
	SetDispatcherProfile(dpp *engine.DispatcherProfileWithArgDispatcher, reply *string) error
	SetActionPlan(args *engine.SetActionPlanArgWithArgDispatcher, reply *string) error
	SetAccountActionPlans(args *engine.SetAccountActionPlansArgWithArgDispatcher, reply *string) error
	SetDispatcherHost(dpp *engine.DispatcherHostWithArgDispatcher, reply *string) error
	RemoveThreshold(args *utils.TenantIDWithArgDispatcher, reply *string) error
	SetLoadIDs(args *utils.LoadIDsWithArgDispatcher, reply *string) error
	RemoveDestination(id *utils.StringWithApiKey, reply *string) error
	RemoveAccount(id *utils.StringWithApiKey, reply *string) error
	RemoveStatQueue(args *utils.TenantIDWithArgDispatcher, reply *string) error
	RemoveFilter(args *utils.TenantIDWithArgDispatcher, reply *string) error
	RemoveThresholdProfile(args *utils.TenantIDWithArgDispatcher, reply *string) error
	RemoveStatQueueProfile(args *utils.TenantIDWithArgDispatcher, reply *string) error
	RemoveTiming(id *utils.StringWithApiKey, reply *string) error
	RemoveResource(args *utils.TenantIDWithArgDispatcher, reply *string) error
	RemoveResourceProfile(args *utils.TenantIDWithArgDispatcher, reply *string) error
	RemoveActionTriggers(id *utils.StringWithApiKey, reply *string) error
	RemoveSharedGroup(id *utils.StringWithApiKey, reply *string) error
	RemoveActions(id *utils.StringWithApiKey, reply *string) error
	RemoveActionPlan(id *utils.StringWithApiKey, reply *string) error
	RemAccountActionPlans(args *engine.RemAccountActionPlansArgsWithArgDispatcher, reply *string) error
	RemoveRatingPlan(id *utils.StringWithApiKey, reply *string) error
	RemoveRatingProfile(id *utils.StringWithApiKey, reply *string) error
	RemoveSupplierProfile(args *utils.TenantIDWithArgDispatcher, reply *string) error
	RemoveAttributeProfile(args *utils.TenantIDWithArgDispatcher, reply *string) error
	RemoveChargerProfile(args *utils.TenantIDWithArgDispatcher, reply *string) error
	RemoveDispatcherProfile(args *utils.TenantIDWithArgDispatcher, reply *string) error
	RemoveDispatcherHost(args *utils.TenantIDWithArgDispatcher, reply *string) error
}
