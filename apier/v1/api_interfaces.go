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
	GetThresholdIDs(tenant *utils.TenantWithOpts, tIDs *[]string) error
	GetThresholdsForEvent(args *engine.ThresholdsArgsProcessEvent, reply *engine.Thresholds) error
	GetThreshold(tntID *utils.TenantIDWithOpts, t *engine.Threshold) error
	ProcessEvent(args *engine.ThresholdsArgsProcessEvent, tIDs *[]string) error
	Ping(ign *utils.CGREvent, reply *string) error
}

type StatSv1Interface interface {
	GetQueueIDs(tenant *utils.TenantWithOpts, qIDs *[]string) error
	ProcessEvent(args *engine.StatsArgsProcessEvent, reply *[]string) error
	GetStatQueuesForEvent(args *engine.StatsArgsProcessEvent, reply *[]string) (err error)
	GetQueueStringMetrics(args *utils.TenantIDWithOpts, reply *map[string]string) (err error)
	GetQueueFloatMetrics(args *utils.TenantIDWithOpts, reply *map[string]float64) (err error)
	Ping(ign *utils.CGREvent, reply *string) error
}

type ResourceSv1Interface interface {
	GetResourcesForEvent(args *utils.ArgRSv1ResourceUsage, reply *engine.Resources) error
	AuthorizeResources(args *utils.ArgRSv1ResourceUsage, reply *string) error
	AllocateResources(args *utils.ArgRSv1ResourceUsage, reply *string) error
	ReleaseResources(args *utils.ArgRSv1ResourceUsage, reply *string) error
	GetResource(args *utils.TenantIDWithOpts, reply *engine.Resource) error
	GetResourceWithConfig(args *utils.TenantIDWithOpts, reply *engine.ResourceWithConfig) error
	Ping(ign *utils.CGREvent, reply *string) error
}

type RouteSv1Interface interface {
	GetRoutes(args *engine.ArgsGetRoutes, reply *engine.SortedRoutes) error
	GetRouteProfilesForEvent(args *utils.CGREvent, reply *[]*engine.RouteProfile) error
	Ping(ign *utils.CGREvent, reply *string) error
}

type AttributeSv1Interface interface {
	GetAttributeForEvent(args *engine.AttrArgsProcessEvent, reply *engine.AttributeProfile) (err error)
	ProcessEvent(args *engine.AttrArgsProcessEvent, reply *engine.AttrSProcessEventReply) error
	Ping(ign *utils.CGREvent, reply *string) error
}

type ChargerSv1Interface interface {
	Ping(ign *utils.CGREvent, reply *string) error
	GetChargersForEvent(cgrEv *utils.CGREvent, reply *engine.ChargerProfiles) error
	ProcessEvent(args *utils.CGREvent, reply *[]*engine.ChrgSProcessEventReply) error
}

type SessionSv1Interface interface {
	AuthorizeEvent(args *sessions.V1AuthorizeArgs, rply *sessions.V1AuthorizeReply) error
	AuthorizeEventWithDigest(args *sessions.V1AuthorizeArgs, rply *sessions.V1AuthorizeReplyWithDigest) error
	InitiateSession(args *sessions.V1InitSessionArgs, rply *sessions.V1InitSessionReply) error
	InitiateSessionWithDigest(args *sessions.V1InitSessionArgs, rply *sessions.V1InitReplyWithDigest) error
	UpdateSession(args *sessions.V1UpdateSessionArgs, rply *sessions.V1UpdateSessionReply) error
	SyncSessions(args *utils.TenantWithOpts, rply *string) error
	TerminateSession(args *sessions.V1TerminateSessionArgs, rply *string) error
	ProcessCDR(cgrEv *utils.CGREvent, rply *string) error
	ProcessMessage(args *sessions.V1ProcessMessageArgs, rply *sessions.V1ProcessMessageReply) error
	ProcessEvent(args *sessions.V1ProcessEventArgs, rply *sessions.V1ProcessEventReply) error
	GetCost(args *sessions.V1ProcessEventArgs, rply *sessions.V1GetCostReply) error
	GetActiveSessions(args *utils.SessionFilter, rply *[]*sessions.ExternalSession) error
	GetActiveSessionsCount(args *utils.SessionFilter, rply *int) error
	ForceDisconnect(args *utils.SessionFilter, rply *string) error
	GetPassiveSessions(args *utils.SessionFilter, rply *[]*sessions.ExternalSession) error
	GetPassiveSessionsCount(args *utils.SessionFilter, rply *int) error
	Ping(ign *utils.CGREvent, reply *string) error
	ReplicateSessions(args *dispatchers.ArgsReplicateSessionsWithOpts, rply *string) error
	SetPassiveSession(args *sessions.Session, reply *string) error
	ActivateSessions(args *utils.SessionIDsWithArgsDispatcher, reply *string) error
	DeactivateSessions(args *utils.SessionIDsWithArgsDispatcher, reply *string) error

	STIRAuthenticate(args *sessions.V1STIRAuthenticateArgs, reply *string) error
	STIRIdentity(args *sessions.V1STIRIdentityArgs, reply *string) error
}

type ResponderInterface interface {
	GetCost(arg *engine.CallDescriptorWithOpts, reply *engine.CallCost) (err error)
	Debit(arg *engine.CallDescriptorWithOpts, reply *engine.CallCost) (err error)
	MaxDebit(arg *engine.CallDescriptorWithOpts, reply *engine.CallCost) (err error)
	RefundIncrements(arg *engine.CallDescriptorWithOpts, reply *engine.Account) (err error)
	RefundRounding(arg *engine.CallDescriptorWithOpts, reply *float64) (err error)
	GetMaxSessionTime(arg *engine.CallDescriptorWithOpts, reply *time.Duration) (err error)
	Shutdown(arg *utils.TenantWithOpts, reply *string) (err error)
	Ping(ign *utils.CGREvent, reply *string) error
}

type CacheSv1Interface interface {
	GetItemIDs(args *utils.ArgsGetCacheItemIDsWithOpts, reply *[]string) error
	HasItem(args *utils.ArgsGetCacheItemWithOpts, reply *bool) error
	GetItemExpiryTime(args *utils.ArgsGetCacheItemWithOpts, reply *time.Time) error
	RemoveItem(args *utils.ArgsGetCacheItemWithOpts, reply *string) error
	RemoveItems(args utils.AttrReloadCacheWithOpts, reply *string) error
	Clear(cacheIDs *utils.AttrCacheIDsWithOpts, reply *string) error
	GetCacheStats(cacheIDs *utils.AttrCacheIDsWithOpts, rply *map[string]*ltcache.CacheStats) error
	PrecacheStatus(cacheIDs *utils.AttrCacheIDsWithOpts, rply *map[string]string) error
	HasGroup(args *utils.ArgsGetGroupWithOpts, rply *bool) error
	GetGroupItemIDs(args *utils.ArgsGetGroupWithOpts, rply *[]string) error
	RemoveGroup(args *utils.ArgsGetGroupWithOpts, rply *string) error
	ReloadCache(attrs *utils.AttrReloadCacheWithOpts, reply *string) error
	LoadCache(args *utils.AttrReloadCacheWithOpts, reply *string) error
	ReplicateSet(args *utils.ArgCacheReplicateSet, reply *string) (err error)
	ReplicateRemove(args *utils.ArgCacheReplicateRemove, reply *string) (err error)
	Ping(ign *utils.CGREvent, reply *string) error
}

type GuardianSv1Interface interface {
	RemoteLock(attr *dispatchers.AttrRemoteLockWithOpts, reply *string) (err error)
	RemoteUnlock(refID *dispatchers.AttrRemoteUnlockWithOpts, reply *[]string) (err error)
	Ping(ign *utils.CGREvent, reply *string) error
}

type SchedulerSv1Interface interface {
	Reload(arg *utils.CGREvent, reply *string) error
	Ping(ign *utils.CGREvent, reply *string) error
	ExecuteActions(attr *utils.AttrsExecuteActions, reply *string) error
	ExecuteActionPlans(attr *utils.AttrsExecuteActionPlans, reply *string) error
}

type CDRsV1Interface interface {
	ProcessCDR(cdr *engine.CDRWithOpts, reply *string) error
	ProcessEvent(arg *engine.ArgV1ProcessEvent, reply *string) error
	ProcessExternalCDR(cdr *engine.ExternalCDRWithOpts, reply *string) error
	RateCDRs(arg *engine.ArgRateCDRs, reply *string) error
	StoreSessionCost(attr *engine.AttrCDRSStoreSMCost, reply *string) error
	GetCDRsCount(args *utils.RPCCDRsFilterWithOpts, reply *int64) error
	GetCDRs(args *utils.RPCCDRsFilterWithOpts, reply *[]*engine.CDR) error
	Ping(ign *utils.CGREvent, reply *string) error
}

type ServiceManagerV1Interface interface {
	StartService(args *dispatchers.ArgStartServiceWithOpts, reply *string) error
	StopService(args *dispatchers.ArgStartServiceWithOpts, reply *string) error
	ServiceStatus(args *dispatchers.ArgStartServiceWithOpts, reply *string) error
	Ping(ign *utils.CGREvent, reply *string) error
}

type RALsV1Interface interface {
	GetRatingPlansCost(arg *utils.RatingPlanCostArg, reply *dispatchers.RatingPlanCost) error
	Ping(ign *utils.CGREvent, reply *string) error
}

type ConfigSv1Interface interface {
	GetConfig(section *config.SectionWithOpts, reply *map[string]interface{}) (err error)
	ReloadConfig(section *config.ReloadArgs, reply *string) (err error)
	SetConfig(args *config.SetConfigArgs, reply *string) (err error)
	SetConfigFromJSON(args *config.SetConfigFromJSONArgs, reply *string) (err error)
	GetConfigAsJSON(args *config.SectionWithOpts, reply *string) (err error)
}

type CoreSv1Interface interface {
	Status(arg *utils.TenantWithOpts, reply *map[string]interface{}) error
	Ping(ign *utils.CGREvent, reply *string) error
	Sleep(arg *utils.DurationArgs, reply *string) error
}

type RateSv1Interface interface {
	Ping(ign *utils.CGREvent, reply *string) error
	CostForEvent(args *utils.ArgsCostForEvent, rpCost *engine.RateProfileCost) error
}

type RateProfileSv1Interface interface {
	Ping(ign *utils.CGREvent, reply *string) error
}

type ReplicatorSv1Interface interface {
	Ping(ign *utils.CGREvent, reply *string) error
	GetAccount(args *utils.StringWithOpts, reply *engine.Account) error
	GetDestination(key *utils.StringWithOpts, reply *engine.Destination) error
	GetReverseDestination(key *utils.StringWithOpts, reply *[]string) error
	GetStatQueue(tntID *utils.TenantIDWithOpts, reply *engine.StatQueue) error
	GetFilter(tntID *utils.TenantIDWithOpts, reply *engine.Filter) error
	GetThreshold(tntID *utils.TenantIDWithOpts, reply *engine.Threshold) error
	GetThresholdProfile(tntID *utils.TenantIDWithOpts, reply *engine.ThresholdProfile) error
	GetStatQueueProfile(tntID *utils.TenantIDWithOpts, reply *engine.StatQueueProfile) error
	GetTiming(id *utils.StringWithOpts, reply *utils.TPTiming) error
	GetResource(tntID *utils.TenantIDWithOpts, reply *engine.Resource) error
	GetResourceProfile(tntID *utils.TenantIDWithOpts, reply *engine.ResourceProfile) error
	GetActionTriggers(id *utils.StringWithOpts, reply *engine.ActionTriggers) error
	GetSharedGroup(id *utils.StringWithOpts, reply *engine.SharedGroup) error
	GetActions(id *utils.StringWithOpts, reply *engine.Actions) error
	GetActionPlan(id *utils.StringWithOpts, reply *engine.ActionPlan) error
	GetAllActionPlans(_ *utils.StringWithOpts, reply *map[string]*engine.ActionPlan) error
	GetAccountActionPlans(id *utils.StringWithOpts, reply *[]string) error
	GetRatingPlan(id *utils.StringWithOpts, reply *engine.RatingPlan) error
	GetRatingProfile(id *utils.StringWithOpts, reply *engine.RatingProfile) error
	GetRouteProfile(tntID *utils.TenantIDWithOpts, reply *engine.RouteProfile) error
	GetAttributeProfile(tntID *utils.TenantIDWithOpts, reply *engine.AttributeProfile) error
	GetChargerProfile(tntID *utils.TenantIDWithOpts, reply *engine.ChargerProfile) error
	GetDispatcherProfile(tntID *utils.TenantIDWithOpts, reply *engine.DispatcherProfile) error
	GetRateProfile(tntID *utils.TenantIDWithOpts, reply *engine.RateProfile) error
	GetDispatcherHost(tntID *utils.TenantIDWithOpts, reply *engine.DispatcherHost) error
	GetItemLoadIDs(itemID *utils.StringWithOpts, reply *map[string]int64) error
	SetThresholdProfile(th *engine.ThresholdProfileWithOpts, reply *string) error
	SetThreshold(th *engine.ThresholdWithOpts, reply *string) error
	SetAccount(acc *engine.AccountWithOpts, reply *string) error
	SetDestination(dst *engine.DestinationWithOpts, reply *string) error
	SetReverseDestination(dst *engine.DestinationWithOpts, reply *string) error
	SetStatQueue(ssq *engine.StoredStatQueueWithOpts, reply *string) error
	SetFilter(fltr *engine.FilterWithOpts, reply *string) error
	SetStatQueueProfile(sq *engine.StatQueueProfileWithOpts, reply *string) error
	SetTiming(tm *utils.TPTimingWithOpts, reply *string) error
	SetResource(rs *engine.ResourceWithOpts, reply *string) error
	SetResourceProfile(rs *engine.ResourceProfileWithOpts, reply *string) error
	SetActionTriggers(args *engine.SetActionTriggersArgWithOpts, reply *string) error
	SetSharedGroup(shg *engine.SharedGroupWithOpts, reply *string) error
	SetActions(args *engine.SetActionsArgsWithOpts, reply *string) error
	SetRatingPlan(rp *engine.RatingPlanWithOpts, reply *string) error
	SetRatingProfile(rp *engine.RatingProfileWithOpts, reply *string) error
	SetRouteProfile(sp *engine.RouteProfileWithOpts, reply *string) error
	SetAttributeProfile(ap *engine.AttributeProfileWithOpts, reply *string) error
	SetChargerProfile(cp *engine.ChargerProfileWithOpts, reply *string) error
	SetDispatcherProfile(dpp *engine.DispatcherProfileWithOpts, reply *string) error
	SetRateProfile(dpp *engine.RateProfileWithOpts, reply *string) error
	SetActionPlan(args *engine.SetActionPlanArgWithOpts, reply *string) error
	SetAccountActionPlans(args *engine.SetAccountActionPlansArgWithOpts, reply *string) error
	SetDispatcherHost(dpp *engine.DispatcherHostWithOpts, reply *string) error
	RemoveThreshold(args *utils.TenantIDWithOpts, reply *string) error
	SetLoadIDs(args *utils.LoadIDsWithOpts, reply *string) error
	RemoveDestination(id *utils.StringWithOpts, reply *string) error
	RemoveAccount(id *utils.StringWithOpts, reply *string) error
	RemoveStatQueue(args *utils.TenantIDWithOpts, reply *string) error
	RemoveFilter(args *utils.TenantIDWithOpts, reply *string) error
	RemoveThresholdProfile(args *utils.TenantIDWithOpts, reply *string) error
	RemoveStatQueueProfile(args *utils.TenantIDWithOpts, reply *string) error
	RemoveTiming(id *utils.StringWithOpts, reply *string) error
	RemoveResource(args *utils.TenantIDWithOpts, reply *string) error
	RemoveResourceProfile(args *utils.TenantIDWithOpts, reply *string) error
	RemoveActionTriggers(id *utils.StringWithOpts, reply *string) error
	RemoveSharedGroup(id *utils.StringWithOpts, reply *string) error
	RemoveActions(id *utils.StringWithOpts, reply *string) error
	RemoveActionPlan(id *utils.StringWithOpts, reply *string) error
	RemAccountActionPlans(args *engine.RemAccountActionPlansArgsWithOpts, reply *string) error
	RemoveRatingPlan(id *utils.StringWithOpts, reply *string) error
	RemoveRatingProfile(id *utils.StringWithOpts, reply *string) error
	RemoveRouteProfile(args *utils.TenantIDWithOpts, reply *string) error
	RemoveAttributeProfile(args *utils.TenantIDWithOpts, reply *string) error
	RemoveChargerProfile(args *utils.TenantIDWithOpts, reply *string) error
	RemoveDispatcherProfile(args *utils.TenantIDWithOpts, reply *string) error
	RemoveDispatcherHost(args *utils.TenantIDWithOpts, reply *string) error
	RemoveRateProfile(args *utils.TenantIDWithOpts, reply *string) error

	GetIndexes(args *utils.GetIndexesArg, reply *map[string]utils.StringSet) error
	SetIndexes(args *utils.SetIndexesArg, reply *string) error
	RemoveIndexes(args *utils.GetIndexesArg, reply *string) error

	GetAccountProfile(tntID *utils.TenantIDWithOpts, reply *utils.AccountProfile) error
	SetAccountProfile(args *utils.AccountProfileWithOpts, reply *string) error
	RemoveAccountProfile(args *utils.TenantIDWithOpts, reply *string) error

	GetActionProfile(tntID *utils.TenantIDWithOpts, reply *engine.ActionProfile) error
	SetActionProfile(args *engine.ActionProfileWithOpts, reply *string) error
	RemoveActionProfile(args *utils.TenantIDWithOpts, reply *string) error
}

type ActionSv1Interface interface {
	ScheduleActions(args *utils.ArgActionSv1ScheduleActions, rpl *string) error
	ExecuteActions(args *utils.ArgActionSv1ScheduleActions, rpl *string) error
	Ping(ign *utils.CGREvent, reply *string) error
}

type AccountSv1Interface interface {
	Ping(ign *utils.CGREvent, reply *string) error
	AccountProfilesForEvent(args *utils.ArgsAccountsForEvent, aps *[]*utils.AccountProfile) error
	MaxAbstracts(args *utils.ArgsAccountsForEvent, eEc *utils.ExtEventCharges) error
	DebitAbstracts(args *utils.ArgsAccountsForEvent, eEc *utils.ExtEventCharges) error
	ActionSetBalance(args *utils.ArgsActSetBalance, eEc *string) (err error)
	ActionRemoveBalance(args *utils.ArgsActRemoveBalances, eEc *string) (err error)
}
