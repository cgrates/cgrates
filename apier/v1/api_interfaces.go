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
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

type ThresholdSv1Interface interface {
	GetThresholdIDs(ctx *context.Context, tenant *utils.TenantWithAPIOpts, tIDs *[]string) error
	GetThresholdsForEvent(ctx *context.Context, args *utils.CGREvent, reply *engine.Thresholds) error
	GetThreshold(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, t *engine.Threshold) error
	ProcessEvent(ctx *context.Context, args *utils.CGREvent, tIDs *[]string) error
}

type StatSv1Interface interface {
	GetQueueIDs(ctx *context.Context, tenant *utils.TenantWithAPIOpts, qIDs *[]string) error
	ProcessEvent(ctx *context.Context, args *utils.CGREvent, reply *[]string) error
	GetStatQueuesForEvent(ctx *context.Context, args *utils.CGREvent, reply *[]string) (err error)
	GetQueueStringMetrics(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *map[string]string) (err error)
	GetQueueFloatMetrics(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *map[string]float64) (err error)
}

type ResourceSv1Interface interface {
	GetResourcesForEvent(ctx *context.Context, args *utils.CGREvent, reply *engine.Resources) error
	AuthorizeResources(ctx *context.Context, args *utils.CGREvent, reply *string) error
	AllocateResources(ctx *context.Context, args *utils.CGREvent, reply *string) error
	ReleaseResources(ctx *context.Context, args *utils.CGREvent, reply *string) error
	GetResource(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.Resource) error
	GetResourceWithConfig(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.ResourceWithConfig) error
}

type IPsV1Interface interface {
	GetIPAllocationForEvent(*context.Context, *utils.CGREvent, *engine.IPAllocations) error
	AuthorizeIP(*context.Context, *utils.CGREvent, *engine.AllocatedIP) error
	AllocateIP(*context.Context, *utils.CGREvent, *engine.AllocatedIP) error
	ReleaseIP(*context.Context, *utils.CGREvent, *string) error
	GetIPAllocations(*context.Context, *utils.TenantIDWithAPIOpts, *engine.IPAllocations) error
	ClearIPAllocations(*context.Context, *engine.ClearIPAllocationsArgs, *string) error
}

type RouteSv1Interface interface {
	GetRoutes(ctx *context.Context, args *utils.CGREvent, reply *engine.SortedRoutesList) error
	GetRouteProfilesForEvent(ctx *context.Context, args *utils.CGREvent, reply *[]*engine.RouteProfile) error
	GetRoutesList(ctx *context.Context, args *utils.CGREvent, reply *[]string) error
}

type AttributeSv1Interface interface {
	GetAttributeForEvent(ctx *context.Context, args *utils.CGREvent, reply *engine.AttributeProfile) (err error)
	ProcessEvent(ctx *context.Context, args *utils.CGREvent, reply *engine.AttrSProcessEventReply) error
}

type ChargerSv1Interface interface {
	GetChargersForEvent(ctx *context.Context, cgrEv *utils.CGREvent, reply *engine.ChargerProfiles) error
	ProcessEvent(ctx *context.Context, args *utils.CGREvent, reply *[]*engine.ChrgSProcessEventReply) error
}

type SessionSv1Interface interface {
	AuthorizeEvent(ctx *context.Context, args *sessions.V1AuthorizeArgs, rply *sessions.V1AuthorizeReply) error
	AuthorizeEventWithDigest(ctx *context.Context, args *sessions.V1AuthorizeArgs, rply *sessions.V1AuthorizeReplyWithDigest) error
	InitiateSession(ctx *context.Context, args *sessions.V1InitSessionArgs, rply *sessions.V1InitSessionReply) error
	InitiateSessionWithDigest(ctx *context.Context, args *sessions.V1InitSessionArgs, rply *sessions.V1InitReplyWithDigest) error
	UpdateSession(ctx *context.Context, args *sessions.V1UpdateSessionArgs, rply *sessions.V1UpdateSessionReply) error
	SyncSessions(ctx *context.Context, args *utils.TenantWithAPIOpts, rply *string) error
	TerminateSession(ctx *context.Context, args *sessions.V1TerminateSessionArgs, rply *string) error
	ProcessCDR(ctx *context.Context, cgrEv *utils.CGREvent, rply *string) error
	ProcessMessage(ctx *context.Context, args *sessions.V1ProcessMessageArgs, rply *sessions.V1ProcessMessageReply) error
	ProcessEvent(ctx *context.Context, args *sessions.V1ProcessEventArgs, rply *sessions.V1ProcessEventReply) error
	GetCost(ctx *context.Context, args *sessions.V1ProcessEventArgs, rply *sessions.V1GetCostReply) error
	GetActiveSessions(ctx *context.Context, args *utils.SessionFilter, rply *[]*sessions.ExternalSession) error
	GetActiveSessionsCount(ctx *context.Context, args *utils.SessionFilter, rply *int) error
	ForceDisconnect(ctx *context.Context, args utils.SessionFilterWithEvent, rply *string) error
	AlterSessions(ctx *context.Context, args utils.SessionFilterWithEvent, rply *string) error
	GetPassiveSessions(ctx *context.Context, args *utils.SessionFilter, rply *[]*sessions.ExternalSession) error
	GetPassiveSessionsCount(ctx *context.Context, args *utils.SessionFilter, rply *int) error
	ReplicateSessions(ctx *context.Context, args *dispatchers.ArgsReplicateSessionsWithAPIOpts, rply *string) error
	SetPassiveSession(ctx *context.Context, args *sessions.Session, reply *string) error
	ActivateSessions(ctx *context.Context, args *utils.SessionIDsWithArgsDispatcher, reply *string) error
	DeactivateSessions(ctx *context.Context, args *utils.SessionIDsWithArgsDispatcher, reply *string) error

	STIRAuthenticate(ctx *context.Context, args *sessions.V1STIRAuthenticateArgs, reply *string) error
	STIRIdentity(ctx *context.Context, args *sessions.V1STIRIdentityArgs, reply *string) error
}

type ResponderInterface interface {
	GetCost(ctx *context.Context, arg *engine.CallDescriptorWithAPIOpts, reply *engine.CallCost) (err error)
	Debit(ctx *context.Context, arg *engine.CallDescriptorWithAPIOpts, reply *engine.CallCost) (err error)
	MaxDebit(ctx *context.Context, arg *engine.CallDescriptorWithAPIOpts, reply *engine.CallCost) (err error)
	RefundIncrements(ctx *context.Context, arg *engine.CallDescriptorWithAPIOpts, reply *engine.Account) (err error)
	RefundRounding(ctx *context.Context, arg *engine.CallDescriptorWithAPIOpts, reply *engine.Account) (err error)
	GetMaxSessionTime(ctx *context.Context, arg *engine.CallDescriptorWithAPIOpts, reply *time.Duration) (err error)
	GetCostOnRatingPlans(ctx *context.Context, arg *utils.GetCostOnRatingPlansArgs, reply *map[string]any) (err error)
	GetMaxSessionTimeOnAccounts(ctx *context.Context, arg *utils.GetMaxSessionTimeOnAccountsArgs, reply *map[string]any) (err error)
	Shutdown(ctx *context.Context, arg *utils.TenantWithAPIOpts, reply *string) (err error)
}

type CacheSv1Interface interface {
	GetItemIDs(ctx *context.Context, args *utils.ArgsGetCacheItemIDsWithAPIOpts, reply *[]string) error
	HasItem(ctx *context.Context, args *utils.ArgsGetCacheItemWithAPIOpts, reply *bool) error
	GetItemExpiryTime(ctx *context.Context, args *utils.ArgsGetCacheItemWithAPIOpts, reply *time.Time) error
	RemoveItem(ctx *context.Context, args *utils.ArgsGetCacheItemWithAPIOpts, reply *string) error
	RemoveItems(ctx *context.Context, args *utils.AttrReloadCacheWithAPIOpts, reply *string) error
	Clear(ctx *context.Context, cacheIDs *utils.AttrCacheIDsWithAPIOpts, reply *string) error
	GetCacheStats(ctx *context.Context, cacheIDs *utils.AttrCacheIDsWithAPIOpts, rply *map[string]*ltcache.CacheStats) error
	PrecacheStatus(ctx *context.Context, cacheIDs *utils.AttrCacheIDsWithAPIOpts, rply *map[string]string) error
	HasGroup(ctx *context.Context, args *utils.ArgsGetGroupWithAPIOpts, rply *bool) error
	GetGroupItemIDs(ctx *context.Context, args *utils.ArgsGetGroupWithAPIOpts, rply *[]string) error
	RemoveGroup(ctx *context.Context, args *utils.ArgsGetGroupWithAPIOpts, rply *string) error
	ReloadCache(ctx *context.Context, attrs *utils.AttrReloadCacheWithAPIOpts, reply *string) error
	LoadCache(ctx *context.Context, args *utils.AttrReloadCacheWithAPIOpts, reply *string) error
	ReplicateSet(ctx *context.Context, args *utils.ArgCacheReplicateSet, reply *string) (err error)
	ReplicateRemove(ctx *context.Context, args *utils.ArgCacheReplicateRemove, reply *string) (err error)
}

type GuardianSv1Interface interface {
	RemoteLock(ctx *context.Context, attr *dispatchers.AttrRemoteLockWithAPIOpts, reply *string) (err error)
	RemoteUnlock(ctx *context.Context, refID *dispatchers.AttrRemoteUnlockWithAPIOpts, reply *[]string) (err error)
}

type SchedulerSv1Interface interface {
	Reload(ctx *context.Context, arg *utils.CGREvent, reply *string) error
	ExecuteActions(ctx *context.Context, attr *utils.AttrsExecuteActions, reply *string) error
	ExecuteActionPlans(ctx *context.Context, attr *utils.AttrsExecuteActionPlans, reply *string) error
}

type CDRsV1Interface interface {
	ProcessCDR(ctx *context.Context, cdr *engine.CDRWithAPIOpts, reply *string) error
	ProcessEvent(ctx *context.Context, arg *engine.ArgV1ProcessEvent, reply *string) error
	ProcessExternalCDR(ctx *context.Context, cdr *engine.ExternalCDRWithAPIOpts, reply *string) error
	RateCDRs(ctx *context.Context, arg *engine.ArgRateCDRs, reply *string) error
	StoreSessionCost(ctx *context.Context, attr *engine.AttrCDRSStoreSMCost, reply *string) error
	GetCDRsCount(ctx *context.Context, args *utils.RPCCDRsFilterWithAPIOpts, reply *int64) error
	GetCDRs(ctx *context.Context, args *utils.RPCCDRsFilterWithAPIOpts, reply *[]*engine.CDR) error
}

type ServiceManagerV1Interface interface {
	StartService(ctx *context.Context, args *dispatchers.ArgStartServiceWithAPIOpts, reply *string) error
	StopService(ctx *context.Context, args *dispatchers.ArgStartServiceWithAPIOpts, reply *string) error
	ServiceStatus(ctx *context.Context, args *dispatchers.ArgStartServiceWithAPIOpts, reply *string) error
}

type RALsV1Interface interface {
	GetRatingPlansCost(ctx *context.Context, arg *utils.RatingPlanCostArg, reply *dispatchers.RatingPlanCost) error
}

type ConfigSv1Interface interface {
	GetConfig(ctx *context.Context, section *config.SectionWithAPIOpts, reply *map[string]any) (err error)
	ReloadConfig(ctx *context.Context, section *config.ReloadArgs, reply *string) (err error)
	SetConfig(ctx *context.Context, args *config.SetConfigArgs, reply *string) (err error)
	SetConfigFromJSON(ctx *context.Context, args *config.SetConfigFromJSONArgs, reply *string) (err error)
	GetConfigAsJSON(ctx *context.Context, args *config.SectionWithAPIOpts, reply *string) (err error)
}

type CoreSv1Interface interface {
	Status(_ *context.Context, _ *cores.V1StatusParams, _ *map[string]any) error
	Panic(_ *context.Context, _ *utils.PanicMessageArgs, _ *string) error
	Sleep(_ *context.Context, _ *utils.DurationArgs, _ *string) error
	StartCPUProfiling(_ *context.Context, _ *utils.DirectoryArgs, _ *string) error
	StartMemoryProfiling(_ *context.Context, _ cores.MemoryProfilingParams, _ *string) error
	StopCPUProfiling(_ *context.Context, _ *utils.TenantWithAPIOpts, _ *string) error
	StopMemoryProfiling(_ *context.Context, _ utils.TenantWithAPIOpts, _ *string) error
}

type ReplicatorSv1Interface interface {
	GetAccount(ctx *context.Context, args *utils.StringWithAPIOpts, reply *engine.Account) error
	GetDestination(ctx *context.Context, key *utils.StringWithAPIOpts, reply *engine.Destination) error
	GetReverseDestination(ctx *context.Context, key *utils.StringWithAPIOpts, reply *[]string) error
	GetStatQueue(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.StatQueue) error
	GetFilter(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.Filter) error
	GetThreshold(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.Threshold) error
	GetThresholdProfile(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.ThresholdProfile) error
	GetStatQueueProfile(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.StatQueueProfile) error
	GetTiming(ctx *context.Context, id *utils.StringWithAPIOpts, reply *utils.TPTiming) error
	GetResource(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.Resource) error
	GetResourceProfile(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.ResourceProfile) error
	GetIPAllocations(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.IPAllocations) error
	GetIPProfile(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.IPProfile) error
	GetActionTriggers(ctx *context.Context, id *utils.StringWithAPIOpts, reply *engine.ActionTriggers) error
	GetSharedGroup(ctx *context.Context, id *utils.StringWithAPIOpts, reply *engine.SharedGroup) error
	GetActions(ctx *context.Context, id *utils.StringWithAPIOpts, reply *engine.Actions) error
	GetActionPlan(ctx *context.Context, id *utils.StringWithAPIOpts, reply *engine.ActionPlan) error
	GetAllActionPlans(ctx *context.Context, _ *utils.StringWithAPIOpts, reply *map[string]*engine.ActionPlan) error
	GetAccountActionPlans(ctx *context.Context, id *utils.StringWithAPIOpts, reply *[]string) error
	GetRatingPlan(ctx *context.Context, id *utils.StringWithAPIOpts, reply *engine.RatingPlan) error
	GetRatingProfile(ctx *context.Context, id *utils.StringWithAPIOpts, reply *engine.RatingProfile) error
	GetRouteProfile(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.RouteProfile) error
	GetAttributeProfile(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.AttributeProfile) error
	GetChargerProfile(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.ChargerProfile) error
	GetDispatcherProfile(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.DispatcherProfile) error
	GetDispatcherHost(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.DispatcherHost) error
	GetItemLoadIDs(ctx *context.Context, itemID *utils.StringWithAPIOpts, reply *map[string]int64) error
	SetThresholdProfile(ctx *context.Context, th *engine.ThresholdProfileWithAPIOpts, reply *string) error
	SetThreshold(ctx *context.Context, th *engine.ThresholdWithAPIOpts, reply *string) error
	SetAccount(ctx *context.Context, acc *engine.AccountWithAPIOpts, reply *string) error
	SetDestination(ctx *context.Context, dst *engine.DestinationWithAPIOpts, reply *string) error
	SetReverseDestination(ctx *context.Context, dst *engine.DestinationWithAPIOpts, reply *string) error
	SetStatQueue(ctx *context.Context, ssq *engine.StatQueueWithAPIOpts, reply *string) error
	SetFilter(ctx *context.Context, fltr *engine.FilterWithAPIOpts, reply *string) error
	SetStatQueueProfile(ctx *context.Context, sq *engine.StatQueueProfileWithAPIOpts, reply *string) error
	SetTiming(ctx *context.Context, tm *utils.TPTimingWithAPIOpts, reply *string) error
	SetResource(ctx *context.Context, rs *engine.ResourceWithAPIOpts, reply *string) error
	SetResourceProfile(ctx *context.Context, rs *engine.ResourceProfileWithAPIOpts, reply *string) error
	SetIPAllocations(ctx *context.Context, rs *engine.IPAllocationsWithAPIOpts, reply *string) error
	SetIPProfile(ctx *context.Context, rs *engine.IPProfileWithAPIOpts, reply *string) error
	SetActionTriggers(ctx *context.Context, args *engine.SetActionTriggersArgWithAPIOpts, reply *string) error
	SetSharedGroup(ctx *context.Context, shg *engine.SharedGroupWithAPIOpts, reply *string) error
	SetActions(ctx *context.Context, args *engine.SetActionsArgsWithAPIOpts, reply *string) error
	SetRatingPlan(ctx *context.Context, rp *engine.RatingPlanWithAPIOpts, reply *string) error
	SetRatingProfile(ctx *context.Context, rp *engine.RatingProfileWithAPIOpts, reply *string) error
	SetRouteProfile(ctx *context.Context, sp *engine.RouteProfileWithAPIOpts, reply *string) error
	SetAttributeProfile(ctx *context.Context, ap *engine.AttributeProfileWithAPIOpts, reply *string) error
	SetChargerProfile(ctx *context.Context, cp *engine.ChargerProfileWithAPIOpts, reply *string) error
	SetDispatcherProfile(ctx *context.Context, dpp *engine.DispatcherProfileWithAPIOpts, reply *string) error
	SetActionPlan(ctx *context.Context, args *engine.SetActionPlanArgWithAPIOpts, reply *string) error
	SetAccountActionPlans(ctx *context.Context, args *engine.SetAccountActionPlansArgWithAPIOpts, reply *string) error
	SetDispatcherHost(ctx *context.Context, dpp *engine.DispatcherHostWithAPIOpts, reply *string) error
	RemoveThreshold(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error
	SetLoadIDs(ctx *context.Context, args *utils.LoadIDsWithAPIOpts, reply *string) error
	RemoveDestination(ctx *context.Context, id *utils.StringWithAPIOpts, reply *string) error
	RemoveAccount(ctx *context.Context, id *utils.StringWithAPIOpts, reply *string) error
	RemoveStatQueue(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error
	RemoveFilter(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error
	RemoveThresholdProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error
	RemoveStatQueueProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error
	RemoveTiming(ctx *context.Context, id *utils.StringWithAPIOpts, reply *string) error
	RemoveResource(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error
	RemoveResourceProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error
	RemoveIPAllocations(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error
	RemoveIPProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error
	RemoveActionTriggers(ctx *context.Context, id *utils.StringWithAPIOpts, reply *string) error
	RemoveSharedGroup(ctx *context.Context, id *utils.StringWithAPIOpts, reply *string) error
	RemoveActions(ctx *context.Context, id *utils.StringWithAPIOpts, reply *string) error
	RemoveActionPlan(ctx *context.Context, id *utils.StringWithAPIOpts, reply *string) error
	RemAccountActionPlans(ctx *context.Context, args *engine.RemAccountActionPlansArgsWithAPIOpts, reply *string) error
	RemoveRatingPlan(ctx *context.Context, id *utils.StringWithAPIOpts, reply *string) error
	RemoveRatingProfile(ctx *context.Context, id *utils.StringWithAPIOpts, reply *string) error
	RemoveRouteProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error
	RemoveAttributeProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error
	RemoveChargerProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error
	RemoveDispatcherProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error
	RemoveDispatcherHost(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error

	GetIndexes(ctx *context.Context, args *utils.GetIndexesArg, reply *map[string]utils.StringSet) error
	SetIndexes(ctx *context.Context, args *utils.SetIndexesArg, reply *string) error
	RemoveIndexes(ctx *context.Context, args *utils.GetIndexesArg, reply *string) error
}
