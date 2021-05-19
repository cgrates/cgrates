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
	// SyncSessions(args *string, rply *string) error
	TerminateSession(args *sessions.V1TerminateSessionArgs, rply *string) error
	ProcessCDR(cgrEv *utils.CGREventWithArgDispatcher, rply *string) error
	ProcessMessage(args *sessions.V1ProcessMessageArgs, rply *sessions.V1ProcessMessageReply) error
	ProcessEvent(args *sessions.V1ProcessEventArgs, rply *sessions.V1ProcessEventReply) error
	GetActiveSessions(args *utils.SessionFilter, rply *[]*sessions.ExternalSession) error
	GetActiveSessionsCount(args *utils.SessionFilter, rply *int) error
	ForceDisconnect(args *utils.SessionFilter, rply *string) error
	GetPassiveSessions(args *utils.SessionFilter, rply *[]*sessions.ExternalSession) error
	GetPassiveSessionsCount(args *utils.SessionFilter, rply *int) error
	Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error
	ReplicateSessions(args dispatchers.ArgsReplicateSessionsWithApiKey, rply *string) error
	SetPassiveSession(args *sessions.Session, reply *string) error
}

type ResponderInterface interface {
	GetCost(arg *engine.CallDescriptorWithArgDispatcher, reply *engine.CallCost) (err error)
	Debit(arg *engine.CallDescriptorWithArgDispatcher, reply *engine.CallCost) (err error)
	MaxDebit(arg *engine.CallDescriptorWithArgDispatcher, reply *engine.CallCost) (err error)
	RefundIncrements(arg *engine.CallDescriptorWithArgDispatcher, reply *engine.Account) (err error)
	RefundRounding(arg *engine.CallDescriptorWithArgDispatcher, reply *float64) (err error)
	GetMaxSessionTime(arg *engine.CallDescriptorWithArgDispatcher, reply *time.Duration) (err error)
	Shutdown(arg *utils.TenantWithArgDispatcher, reply *string) (err error)
	GetCostOnRatingPlans(arg *utils.GetCostOnRatingPlansArgs, reply *map[string]interface{}) (err error)
	GetMaxSessionTimeOnAccounts(arg *utils.GetMaxSessionTimeOnAccountsArgs, reply *map[string]interface{}) (err error)
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
