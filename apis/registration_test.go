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

package apis_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/utils"
)

func TestRegisteredAPIMethodConstants(t *testing.T) {
	tests := []struct {
		name   string
		rcvr   any
		consts map[string]string
	}{
		{
			name: utils.AccountSv1,
			rcvr: apis.NewAccountSv1(nil),
			consts: map[string]string{
				"AccountsForEvent":    utils.AccountSv1AccountsForEvent,
				"ActionRemoveBalance": utils.AccountSv1ActionRemoveBalance,
				"ActionSetBalance":    utils.AccountSv1ActionSetBalance,
				"DebitAbstracts":      utils.AccountSv1DebitAbstracts,
				"DebitConcretes":      utils.AccountSv1DebitConcretes,
				"GetAccount":          utils.AccountSv1GetAccount,
				"MaxAbstracts":        utils.AccountSv1MaxAbstracts,
				"MaxConcretes":        utils.AccountSv1MaxConcretes,
				"Ping":                utils.AccountSv1Ping,
				"RefundCharges":       utils.AccountSv1RefundCharges,
			},
		},
		{
			name: utils.ActionSv1,
			rcvr: apis.NewActionSv1(nil),
			consts: map[string]string{
				"ExecuteActions":  utils.ActionSv1ExecuteActions,
				"Ping":            utils.ActionSv1Ping,
				"ScheduleActions": utils.ActionSv1ScheduleActions,
			},
		},
		{
			name: utils.AdminSv1,
			rcvr: apis.NewAdminSv1(nil, nil, nil, nil),
			consts: map[string]string{
				"BackupDB":                     utils.AdminSv1BackupDB,
				"ComputeFilterIndexIDs":        utils.AdminSv1ComputeFilterIndexIDs,
				"ComputeFilterIndexes":         utils.AdminSv1ComputeFilterIndexes,
				"DumpDB":                       utils.AdminSv1DumpDB,
				"FiltersMatch":                 utils.AdminSv1FiltersMatch,
				"GetAccount":                   utils.AdminSv1GetAccount,
				"GetAccountIDs":                utils.AdminSv1GetAccountIDs,
				"GetAccounts":                  utils.AdminSv1GetAccounts,
				"GetAccountsCount":             utils.AdminSv1GetAccountsCount,
				"GetAccountsIndexesHealth":     utils.AdminSv1GetAccountsIndexesHealth,
				"GetActionProfile":             utils.AdminSv1GetActionProfile,
				"GetActionProfileIDs":          utils.AdminSv1GetActionProfileIDs,
				"GetActionProfiles":            utils.AdminSv1GetActionProfiles,
				"GetActionProfilesCount":       utils.AdminSv1GetActionProfilesCount,
				"GetActionsIndexesHealth":      utils.AdminSv1GetActionsIndexesHealth,
				"GetAttributeProfile":          utils.AdminSv1GetAttributeProfile,
				"GetAttributeProfileIDs":       utils.AdminSv1GetAttributeProfileIDs,
				"GetAttributeProfiles":         utils.AdminSv1GetAttributeProfiles,
				"GetAttributeProfilesCount":    utils.AdminSv1GetAttributeProfilesCount,
				"GetAttributesIndexesHealth":   utils.AdminSv1GetAttributesIndexesHealth,
				"GetCDRs":                      utils.AdminSv1GetCDRs,
				"GetChargerProfile":            utils.AdminSv1GetChargerProfile,
				"GetChargerProfileIDs":         utils.AdminSv1GetChargerProfileIDs,
				"GetChargerProfiles":           utils.AdminSv1GetChargerProfiles,
				"GetChargerProfilesCount":      utils.AdminSv1GetChargerProfilesCount,
				"GetChargersIndexesHealth":     utils.AdminSv1GetChargersIndexesHealth,
				"GetFilter":                    utils.AdminSv1GetFilter,
				"GetFilterIDs":                 utils.AdminSv1GetFilterIDs,
				"GetFilterIndexes":             utils.AdminSv1GetFilterIndexes,
				"GetFilters":                   utils.AdminSv1GetFilters,
				"GetFiltersCount":              utils.AdminSv1GetFiltersCount,
				"GetIPProfile":                 utils.AdminSv1GetIPProfile,
				"GetIPProfileIDs":              utils.AdminSv1GetIPProfileIDs,
				"GetIPProfiles":                utils.AdminSv1GetIPProfiles,
				"GetIPProfilesCount":           utils.AdminSv1GetIPProfilesCount,
				"GetIPsIndexesHealth":          utils.AdminSv1GetIPsIndexesHealth,
				"GetRankingProfile":            utils.AdminSv1GetRankingProfile,
				"GetRankingProfileIDs":         utils.AdminSv1GetRankingProfileIDs,
				"GetRankingProfiles":           utils.AdminSv1GetRankingProfiles,
				"GetRankingProfilesCount":      utils.AdminSv1GetRankingProfilesCount,
				"GetRateProfile":               utils.AdminSv1GetRateProfile,
				"GetRateProfileIDs":            utils.AdminSv1GetRateProfileIDs,
				"GetRateProfileRateIDs":        utils.AdminSv1GetRateProfileRateIDs,
				"GetRateProfileRates":          utils.AdminSv1GetRateProfileRates,
				"GetRateProfileRatesCount":     utils.AdminSv1GetRateProfileRatesCount,
				"GetRateProfiles":              utils.AdminSv1GetRateProfiles,
				"GetRateProfilesCount":         utils.AdminSv1GetRateProfilesCount,
				"GetRateProfilesIndexesHealth": utils.AdminSv1GetRateProfilesIndexesHealth,
				"GetRateRatesIndexesHealth":    utils.AdminSv1GetRateRatesIndexesHealth,
				"GetResourceProfile":           utils.AdminSv1GetResourceProfile,
				"GetResourceProfileIDs":        utils.AdminSv1GetResourceProfileIDs,
				"GetResourceProfiles":          utils.AdminSv1GetResourceProfiles,
				"GetResourceProfilesCount":     utils.AdminSv1GetResourceProfilesCount,
				"GetResourcesIndexesHealth":    utils.AdminSv1GetResourcesIndexesHealth,
				"GetReverseFilterHealth":       utils.AdminSv1GetReverseFilterHealth,
				"GetRouteProfile":              utils.AdminSv1GetRouteProfile,
				"GetRouteProfileIDs":           utils.AdminSv1GetRouteProfileIDs,
				"GetRouteProfiles":             utils.AdminSv1GetRouteProfiles,
				"GetRouteProfilesCount":        utils.AdminSv1GetRouteProfilesCount,
				"GetRoutesIndexesHealth":       utils.AdminSv1GetRoutesIndexesHealth,
				"GetStatQueueProfile":          utils.AdminSv1GetStatQueueProfile,
				"GetStatQueueProfileIDs":       utils.AdminSv1GetStatQueueProfileIDs,
				"GetStatQueueProfiles":         utils.AdminSv1GetStatQueueProfiles,
				"GetStatQueueProfilesCount":    utils.AdminSv1GetStatQueueProfilesCount,
				"GetStatsIndexesHealth":        utils.AdminSv1GetStatsIndexesHealth,
				"GetThresholdProfile":          utils.AdminSv1GetThresholdProfile,
				"GetThresholdProfileIDs":       utils.AdminSv1GetThresholdProfileIDs,
				"GetThresholdProfiles":         utils.AdminSv1GetThresholdProfiles,
				"GetThresholdProfilesCount":    utils.AdminSv1GetThresholdProfilesCount,
				"GetThresholdsIndexesHealth":   utils.AdminSv1GetThresholdsIndexesHealth,
				"GetTrendProfile":              utils.AdminSv1GetTrendProfile,
				"GetTrendProfileIDs":           utils.AdminSv1GetTrendProfileIDs,
				"GetTrendProfiles":             utils.AdminSv1GetTrendProfiles,
				"GetTrendProfilesCount":        utils.AdminSv1GetTrendProfilesCount,
				"Ping":                         utils.AdminSv1Ping,
				"RemoveAccount":                utils.AdminSv1RemoveAccount,
				"RemoveActionProfile":          utils.AdminSv1RemoveActionProfile,
				"RemoveAttributeProfile":       utils.AdminSv1RemoveAttributeProfile,
				"RemoveCDRs":                   utils.AdminSv1RemoveCDRs,
				"RemoveChargerProfile":         utils.AdminSv1RemoveChargerProfile,
				"RemoveFilter":                 utils.AdminSv1RemoveFilter,
				"RemoveFilterIndexes":          utils.AdminSv1RemoveFilterIndexes,
				"RemoveIPProfile":              utils.AdminSv1RemoveIPProfile,
				"RemoveRankingProfile":         utils.AdminSv1RemoveRankingProfile,
				"RemoveRateProfile":            utils.AdminSv1RemoveRateProfile,
				"RemoveRateProfileRates":       utils.AdminSv1RemoveRateProfileRates,
				"RemoveResourceProfile":        utils.AdminSv1RemoveResourceProfile,
				"RemoveRouteProfile":           utils.AdminSv1RemoveRouteProfile,
				"RemoveStatQueueProfile":       utils.AdminSv1RemoveStatQueueProfile,
				"RemoveThresholdProfile":       utils.AdminSv1RemoveThresholdProfile,
				"RemoveTrendProfile":           utils.AdminSv1RemoveTrendProfile,
				"ReplayFailedReplications":     utils.AdminSv1ReplayFailedReplications,
				"RestoreDB":                    utils.AdminSv1RestoreDB,
				"RewriteDB":                    utils.AdminSv1RewriteDB,
				"SetAccount":                   utils.AdminSv1SetAccount,
				"SetActionProfile":             utils.AdminSv1SetActionProfile,
				"SetAttributeProfile":          utils.AdminSv1SetAttributeProfile,
				"SetChargerProfile":            utils.AdminSv1SetChargerProfile,
				"SetFilter":                    utils.AdminSv1SetFilter,
				"SetIPProfile":                 utils.AdminSv1SetIPProfile,
				"SetRankingProfile":            utils.AdminSv1SetRankingProfile,
				"SetRateProfile":               utils.AdminSv1SetRateProfile,
				"SetResourceProfile":           utils.AdminSv1SetResourceProfile,
				"SetRouteProfile":              utils.AdminSv1SetRouteProfile,
				"SetStatQueueProfile":          utils.AdminSv1SetStatQueueProfile,
				"SetThresholdProfile":          utils.AdminSv1SetThresholdProfile,
				"SetTrendProfile":              utils.AdminSv1SetTrendProfile,
				"SnapshotDB":                   utils.AdminSv1SnapshotDB,
			},
		},
		{
			name: utils.AnalyzerSv1,
			rcvr: apis.NewAnalyzerSv1(nil),
			consts: map[string]string{
				"Ping":        utils.AnalyzerSv1Ping,
				"StringQuery": utils.AnalyzerSv1StringQuery,
			},
		},
		{
			name: utils.AttributeSv1,
			rcvr: apis.NewAttributeSv1(nil),
			consts: map[string]string{
				"GetAttributeForEvent": utils.AttributeSv1GetAttributeForEvent,
				"Ping":                 utils.AttributeSv1Ping,
				"ProcessEvent":         utils.AttributeSv1ProcessEvent,
			},
		},
		{
			name: utils.CacheSv1,
			rcvr: apis.NewCacheSv1(nil),
			consts: map[string]string{
				"Clear":             utils.CacheSv1Clear,
				"GetCacheStats":     utils.CacheSv1GetCacheStats,
				"GetGroupItemIDs":   utils.CacheSv1GetGroupItemIDs,
				"GetItem":           utils.CacheSv1GetItem,
				"GetItemExpiryTime": utils.CacheSv1GetItemExpiryTime,
				"GetItemIDs":        utils.CacheSv1GetItemIDs,
				"GetItemWithRemote": utils.CacheSv1GetItemWithRemote,
				"GetStats":          utils.CacheSv1GetStats,
				"HasGroup":          utils.CacheSv1HasGroup,
				"HasItem":           utils.CacheSv1HasItem,
				"LoadCache":         utils.CacheSv1LoadCache,
				"Ping":              utils.CacheSv1Ping,
				"PrecacheStatus":    utils.CacheSv1PrecacheStatus,
				"ReloadCache":       utils.CacheSv1ReloadCache,
				"RemoveGroup":       utils.CacheSv1RemoveGroup,
				"RemoveItem":        utils.CacheSv1RemoveItem,
				"RemoveItems":       utils.CacheSv1RemoveItems,
				"ReplicateRemove":   utils.CacheSv1ReplicateRemove,
				"ReplicateSet":      utils.CacheSv1ReplicateSet,
			},
		},
		{
			name: utils.CDRsV1,
			rcvr: apis.NewCdrSv1(nil),
			consts: map[string]string{
				"Ping":                utils.CDRsV1Ping,
				"ProcessEvent":        utils.CDRsV1ProcessEvent,
				"ProcessEventWithGet": utils.CDRsV1ProcessEventWithGet,
				"ProcessStoredEvents": utils.CDRsV1ProcessStoredEvents,
			},
		},
		{
			name: utils.ChargerSv1,
			rcvr: apis.NewChargerSv1(nil),
			consts: map[string]string{
				"GetChargersForEvent": utils.ChargerSv1GetChargersForEvent,
				"Ping":                utils.ChargerSv1Ping,
				"ProcessEvent":        utils.ChargerSv1ProcessEvent,
			},
		},
		{
			name: utils.ConfigSv1,
			rcvr: apis.NewConfigSv1(nil),
			consts: map[string]string{
				"BackupConfigDB":    utils.ConfigSv1BackupConfigDB,
				"DumpConfigDB":      utils.ConfigSv1DumpConfigDB,
				"GetConfig":         utils.ConfigSv1GetConfig,
				"GetConfigAsJSON":   utils.ConfigSv1GetConfigAsJSON,
				"Ping":              utils.ConfigSv1Ping,
				"ReloadConfig":      utils.ConfigSv1ReloadConfig,
				"RewriteConfigDB":   utils.ConfigSv1RewriteConfigDB,
				"SetConfig":         utils.ConfigSv1SetConfig,
				"SetConfigFromJSON": utils.ConfigSv1SetConfigFromJSON,
				"StoreCfgInDB":      utils.ConfigSv1StoreCfgInDB,
			},
		},
		{
			name: utils.CoreSv1,
			rcvr: apis.NewCoreSv1(nil),
			consts: map[string]string{
				"DescribeMethods":      utils.CoreSv1DescribeMethods,
				"Panic":                utils.CoreSv1Panic,
				"Ping":                 utils.CoreSv1Ping,
				"Shutdown":             utils.CoreSv1Shutdown,
				"Sleep":                utils.CoreSv1Sleep,
				"StartCPUProfiling":    utils.CoreSv1StartCPUProfiling,
				"StartMemoryProfiling": utils.CoreSv1StartMemoryProfiling,
				"Status":               utils.CoreSv1Status,
				"StopCPUProfiling":     utils.CoreSv1StopCPUProfiling,
				"StopMemoryProfiling":  utils.CoreSv1StopMemoryProfiling,
			},
		},
		{
			name: utils.EeSv1,
			rcvr: apis.NewEeSv1(nil),
			consts: map[string]string{
				"ArchiveEventsInReply": utils.EeSv1ArchiveEventsInReply,
				"Ping":                 utils.EeSv1Ping,
				"ProcessEvent":         utils.EeSv1ProcessEvent,
				"ResetExporterMetrics": utils.EeSv1ResetExporterMetrics,
			},
		},
		{
			name: utils.EfSv1,
			rcvr: apis.NewEfSv1(nil),
			consts: map[string]string{
				"Ping":         utils.EfSv1Ping,
				"ProcessEvent": utils.EfSv1ProcessEvent,
				"ReplayEvents": utils.EfSv1ReplayEvents,
			},
		},
		{
			name: utils.ErSv1,
			rcvr: apis.NewErSv1(nil),
			consts: map[string]string{
				"Ping":      utils.ErSv1Ping,
				"RunReader": utils.ErSv1RunReader,
			},
		},
		{
			name: utils.IPsV1,
			rcvr: apis.NewIPSv1(nil),
			consts: map[string]string{
				"AllocateIP":              utils.IPsV1AllocateIP,
				"AuthorizeIP":             utils.IPsV1AuthorizeIP,
				"ClearIPAllocations":      utils.IPsV1ClearIPAllocations,
				"GetIPAllocationForEvent": utils.IPsV1GetIPAllocationForEvent,
				"GetIPAllocations":        utils.IPsV1GetIPAllocations,
				"Ping":                    utils.IPsV1Ping,
				"ReleaseIP":               utils.IPsV1ReleaseIP,
			},
		},
		{
			name: utils.LoaderSv1,
			rcvr: apis.NewLoaderSv1(nil),
			consts: map[string]string{
				"ImportZip": utils.LoaderSv1ImportZip,
				"Ping":      utils.LoaderSv1Ping,
				"Run":       utils.LoaderSv1Run,
			},
		},
		{
			name: utils.RankingSv1,
			rcvr: apis.NewRankingSv1(nil),
			consts: map[string]string{
				"GetRanking":        utils.RankingSv1GetRanking,
				"GetRankingSummary": utils.RankingSv1GetRankingSummary,
				"GetSchedule":       utils.RankingSv1GetSchedule,
				"Ping":              utils.RankingSv1Ping,
				"ScheduleQueries":   utils.RankingSv1ScheduleQueries,
			},
		},
		{
			name: utils.RateSv1,
			rcvr: apis.NewRateSv1(nil),
			consts: map[string]string{
				"CostForEvent":             utils.RateSv1CostForEvent,
				"Ping":                     utils.RateSv1Ping,
				"RateProfileRatesForEvent": utils.RateSv1RateProfileRatesForEvent,
				"RateProfilesForEvent":     utils.RateSv1RateProfilesForEvent,
			},
		},
		{
			name: utils.ReplicatorSv1,
			rcvr: apis.NewReplicatorSv1(nil, nil),
			consts: map[string]string{
				"GetAccount":             utils.ReplicatorSv1GetAccount,
				"GetActionProfile":       utils.ReplicatorSv1GetActionProfile,
				"GetAttributeProfile":    utils.ReplicatorSv1GetAttributeProfile,
				"GetChargerProfile":      utils.ReplicatorSv1GetChargerProfile,
				"GetFilter":              utils.ReplicatorSv1GetFilter,
				"GetIPAllocations":       utils.ReplicatorSv1GetIPAllocations,
				"GetIPProfile":           utils.ReplicatorSv1GetIPProfile,
				"GetIndexes":             utils.ReplicatorSv1GetIndexes,
				"GetItemLoadIDs":         utils.ReplicatorSv1GetItemLoadIDs,
				"GetRanking":             utils.ReplicatorSv1GetRanking,
				"GetRankingProfile":      utils.ReplicatorSv1GetRankingProfile,
				"GetRateProfile":         utils.ReplicatorSv1GetRateProfile,
				"GetResource":            utils.ReplicatorSv1GetResource,
				"GetResourceProfile":     utils.ReplicatorSv1GetResourceProfile,
				"GetRouteProfile":        utils.ReplicatorSv1GetRouteProfile,
				"GetStatQueue":           utils.ReplicatorSv1GetStatQueue,
				"GetStatQueueProfile":    utils.ReplicatorSv1GetStatQueueProfile,
				"GetThreshold":           utils.ReplicatorSv1GetThreshold,
				"GetThresholdProfile":    utils.ReplicatorSv1GetThresholdProfile,
				"GetTrend":               utils.ReplicatorSv1GetTrend,
				"GetTrendProfile":        utils.ReplicatorSv1GetTrendProfile,
				"Ping":                   utils.ReplicatorSv1Ping,
				"RemoveAccount":          utils.ReplicatorSv1RemoveAccount,
				"RemoveActionProfile":    utils.ReplicatorSv1RemoveActionProfile,
				"RemoveAttributeProfile": utils.ReplicatorSv1RemoveAttributeProfile,
				"RemoveChargerProfile":   utils.ReplicatorSv1RemoveChargerProfile,
				"RemoveFilter":           utils.ReplicatorSv1RemoveFilter,
				"RemoveIPAllocations":    utils.ReplicatorSv1RemoveIPAllocations,
				"RemoveIPProfile":        utils.ReplicatorSv1RemoveIPProfile,
				"RemoveIndexes":          utils.ReplicatorSv1RemoveIndexes,
				"RemoveRanking":          utils.ReplicatorSv1RemoveRanking,
				"RemoveRankingProfile":   utils.ReplicatorSv1RemoveRankingProfile,
				"RemoveRateProfile":      utils.ReplicatorSv1RemoveRateProfile,
				"RemoveResource":         utils.ReplicatorSv1RemoveResource,
				"RemoveResourceProfile":  utils.ReplicatorSv1RemoveResourceProfile,
				"RemoveRouteProfile":     utils.ReplicatorSv1RemoveRouteProfile,
				"RemoveStatQueue":        utils.ReplicatorSv1RemoveStatQueue,
				"RemoveStatQueueProfile": utils.ReplicatorSv1RemoveStatQueueProfile,
				"RemoveThreshold":        utils.ReplicatorSv1RemoveThreshold,
				"RemoveThresholdProfile": utils.ReplicatorSv1RemoveThresholdProfile,
				"RemoveTrend":            utils.ReplicatorSv1RemoveTrend,
				"RemoveTrendProfile":     utils.ReplicatorSv1RemoveTrendProfile,
				"SetAccount":             utils.ReplicatorSv1SetAccount,
				"SetActionProfile":       utils.ReplicatorSv1SetActionProfile,
				"SetAttributeProfile":    utils.ReplicatorSv1SetAttributeProfile,
				"SetChargerProfile":      utils.ReplicatorSv1SetChargerProfile,
				"SetFilter":              utils.ReplicatorSv1SetFilter,
				"SetIPAllocations":       utils.ReplicatorSv1SetIPAllocations,
				"SetIPProfile":           utils.ReplicatorSv1SetIPProfile,
				"SetIndexes":             utils.ReplicatorSv1SetIndexes,
				"SetLoadIDs":             utils.ReplicatorSv1SetLoadIDs,
				"SetRanking":             utils.ReplicatorSv1SetRanking,
				"SetRankingProfile":      utils.ReplicatorSv1SetRankingProfile,
				"SetRateProfile":         utils.ReplicatorSv1SetRateProfile,
				"SetResource":            utils.ReplicatorSv1SetResource,
				"SetResourceProfile":     utils.ReplicatorSv1SetResourceProfile,
				"SetRouteProfile":        utils.ReplicatorSv1SetRouteProfile,
				"SetStatQueue":           utils.ReplicatorSv1SetStatQueue,
				"SetStatQueueProfile":    utils.ReplicatorSv1SetStatQueueProfile,
				"SetThreshold":           utils.ReplicatorSv1SetThreshold,
				"SetThresholdProfile":    utils.ReplicatorSv1SetThresholdProfile,
				"SetTrend":               utils.ReplicatorSv1SetTrend,
				"SetTrendProfile":        utils.ReplicatorSv1SetTrendProfile,
			},
		},
		{
			name: utils.ResourceSv1,
			rcvr: apis.NewResourceSv1(nil),
			consts: map[string]string{
				"AllocateResources":     utils.ResourceSv1AllocateResources,
				"AuthorizeResources":    utils.ResourceSv1AuthorizeResources,
				"GetResource":           utils.ResourceSv1GetResource,
				"GetResourceWithConfig": utils.ResourceSv1GetResourceWithConfig,
				"GetResourcesForEvent":  utils.ResourceSv1GetResourcesForEvent,
				"Ping":                  utils.ResourceSv1Ping,
				"ReleaseResources":      utils.ResourceSv1ReleaseResources,
			},
		},
		{
			name: utils.RouteSv1,
			rcvr: apis.NewRouteSv1(nil),
			consts: map[string]string{
				"GetRouteProfilesForEvent": utils.RouteSv1GetRouteProfilesForEvent,
				"GetRoutes":                utils.RouteSv1GetRoutes,
				"GetRoutesList":            utils.RouteSv1GetRoutesList,
				"Ping":                     utils.RouteSv1Ping,
			},
		},
		{
			name: utils.ServiceManagerV1,
			rcvr: apis.NewServiceManagerV1(nil),
			consts: map[string]string{
				"Ping":          utils.ServiceManagerV1Ping,
				"ServiceStatus": utils.ServiceManagerV1ServiceStatus,
				"StartService":  utils.ServiceManagerV1StartService,
				"StopService":   utils.ServiceManagerV1StopService,
			},
		},
		{
			name: utils.SessionSv1,
			rcvr: apis.NewSessionSv1(nil),
			consts: map[string]string{
				"ActivateSessions":           utils.SessionSv1ActivateSessions,
				"AlterSession":               utils.SessionSv1AlterSession,
				"AuthorizeEvent":             utils.SessionSv1AuthorizeEvent,
				"AuthorizeEventWithDigest":   utils.SessionSv1AuthorizeEventWithDigest,
				"DeactivateSessions":         utils.SessionSv1DeactivateSessions,
				"DisconnectPeer":             utils.SessionSv1DisconnectPeer,
				"ForceDisconnect":            utils.SessionSv1ForceDisconnect,
				"GetActiveSessions":          utils.SessionSv1GetActiveSessions,
				"GetActiveSessionsCount":     utils.SessionSv1GetActiveSessionsCount,
				"GetPassiveSessions":         utils.SessionSv1GetPassiveSessions,
				"GetPassiveSessionsCount":    utils.SessionSv1GetPassiveSessionsCount,
				"InitiateSession":            utils.SessionSv1InitiateSession,
				"InitiateSessionWithDigest":  utils.SessionSv1InitiateSessionWithDigest,
				"Ping":                       utils.SessionSv1Ping,
				"ProcessCDR":                 utils.SessionSv1ProcessCDR,
				"ProcessEvent":               utils.SessionSv1ProcessEvent,
				"ProcessMessage":             utils.SessionSv1ProcessMessage,
				"RegisterInternalBiJSONConn": utils.SessionSv1RegisterInternalBiJSONConn,
				"STIRAuthenticate":           utils.SessionSv1STIRAuthenticate,
				"STIRIdentity":               utils.SessionSv1STIRIdentity,
				"SetPassiveSession":          utils.SessionSv1SetPassiveSession,
				"SyncSessions":               utils.SessionSv1SyncSessions,
				"TerminateSession":           utils.SessionSv1TerminateSession,
				"UpdateSession":              utils.SessionSv1UpdateSession,
			},
		},
		{
			name: utils.StatSv1,
			rcvr: apis.NewStatSv1(nil),
			consts: map[string]string{
				"GetQueueDecimalMetrics": utils.StatSv1GetQueueDecimalMetrics,
				"GetQueueFloatMetrics":   utils.StatSv1GetQueueFloatMetrics,
				"GetQueueIDs":            utils.StatSv1GetQueueIDs,
				"GetQueueStringMetrics":  utils.StatSv1GetQueueStringMetrics,
				"GetStatQueue":           utils.StatSv1GetStatQueue,
				"GetStatQueuesForEvent":  utils.StatSv1GetStatQueuesForEvent,
				"Ping":                   utils.StatSv1Ping,
				"ProcessEvent":           utils.StatSv1ProcessEvent,
				"ResetStatQueue":         utils.StatSv1ResetStatQueue,
			},
		},
		{
			name: utils.ThresholdSv1,
			rcvr: apis.NewThresholdSv1(nil),
			consts: map[string]string{
				"GetThreshold":          utils.ThresholdSv1GetThreshold,
				"GetThresholdIDs":       utils.ThresholdSv1GetThresholdIDs,
				"GetThresholdsForEvent": utils.ThresholdSv1GetThresholdsForEvent,
				"Ping":                  utils.ThresholdSv1Ping,
				"ProcessEvent":          utils.ThresholdSv1ProcessEvent,
				"ResetThreshold":        utils.ThresholdSv1ResetThreshold,
			},
		},
		{
			name: utils.TPeSv1,
			rcvr: apis.NewTPeSv1(nil),
			consts: map[string]string{
				"ExportTariffPlan": utils.TPeSv1ExportTariffPlan,
				"Ping":             utils.TPeSv1Ping,
			},
		},
		{
			name: utils.TrendSv1,
			rcvr: apis.NewTrendSv1(nil),
			consts: map[string]string{
				"GetScheduledTrends": utils.TrendSv1GetScheduledTrends,
				"GetTrend":           utils.TrendSv1GetTrend,
				"GetTrendSummary":    utils.TrendSv1GetTrendSummary,
				"Ping":               utils.TrendSv1Ping,
				"ScheduleQueries":    utils.TrendSv1ScheduleQueries,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newRegisteredAPIService(t, tt.rcvr, tt.name)
			if _, has := srv.Methods[utils.Ping]; !has {
				t.Fatalf("%s missing Ping", tt.name)
			}
			for method, val := range tt.consts {
				if _, has := srv.Methods[method]; !has {
					t.Errorf("constant %s references method %s that is not registered on %s", val, method, tt.name)
				}
				if exp := tt.name + utils.NestingSep + method; val != exp {
					t.Errorf("constant for %s = %q, want %q", method, val, exp)
				}
			}
			for method := range srv.Methods {
				if strings.HasPrefix(method, utils.V1Prfx) || strings.HasPrefix(method, "BiRPCv1") {
					t.Errorf("%s exposes prefixed method %s", tt.name, method)
				}
				if _, has := tt.consts[method]; !has {
					t.Errorf("method %s is registered on %s but has no constant in test map", method, tt.name)
				}
			}
			var reply string
			if err := srv.Call(context.Background(), tt.name+utils.NestingSep+utils.Ping, &utils.CGREvent{}, &reply); err != nil {
				t.Error(err)
			} else if reply != utils.Pong {
				t.Errorf("Ping reply = %q, want %q", reply, utils.Pong)
			}
		})
	}
}

func newRegisteredAPIService(t *testing.T, rcvr any, name string) *birpc.Service {
	t.Helper()
	srv, err := birpc.NewService(rcvr, name, true)
	if err != nil {
		t.Fatal(err)
	}
	srv.Methods[utils.Ping] = testPingM
	return srv
}

func testPing(_ any, _ *context.Context, _ *utils.CGREvent, reply *string) error {
	*reply = utils.Pong
	return nil
}

var testPingM = &birpc.MethodType{
	Method: reflect.Method{
		Name: utils.Ping,
		Type: reflect.TypeOf(testPing),
		Func: reflect.ValueOf(testPing),
	},
	ArgType:   reflect.TypeFor[*utils.CGREvent](),
	ReplyType: reflect.TypeFor[*string](),
}
