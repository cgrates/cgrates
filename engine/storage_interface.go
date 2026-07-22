// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

type Storage interface {
	Close()
	Flush(string) error
	GetKeysForPrefix(ctx *context.Context, prefix string, search string) ([]string, error)
	GetVersions(itm string) (vrs Versions, err error)
	SetVersions(vrs Versions, overwrite bool) (err error)
	RemoveVersions(vrs Versions) (err error)
	SelectDatabase(dbName string) (err error)
	GetStorageType() string
	IsDBEmpty() (resp bool, err error)
}

// OnlineStorage contains methods to use for administering online data
type DataDB interface {
	Storage
	HasDataDrv(*context.Context, string, string, string) (bool, error)
	GetResourceProfileDrv(*context.Context, string, string) (*utils.ResourceProfile, error)
	SetResourceProfileDrv(*context.Context, *utils.ResourceProfile) error
	RemoveResourceProfileDrv(*context.Context, string, string) error
	GetResourceDrv(*context.Context, string, string) (*utils.Resource, error)
	SetResourceDrv(*context.Context, *utils.Resource) error
	RemoveResourceDrv(*context.Context, string, string) error
	GetIPProfileDrv(*context.Context, string, string) (*utils.IPProfile, error)
	SetIPProfileDrv(*context.Context, *utils.IPProfile) error
	RemoveIPProfileDrv(*context.Context, string, string) error
	GetIPAllocationsDrv(*context.Context, string, string) (*utils.IPAllocations, error)
	SetIPAllocationsDrv(*context.Context, *utils.IPAllocations) error
	RemoveIPAllocationsDrv(*context.Context, string, string) error
	GetIndexesDrv(ctx *context.Context, idxItmType, tntCtx, transactionID string, idxKeys ...string) (indexes map[string]utils.StringSet, err error)
	SetIndexesDrv(ctx *context.Context, idxItmType, tntCtx string,
		indexes map[string]utils.StringSet, commit bool, transactionID string) (err error)
	RemoveIndexesDrv(ctx *context.Context, idxItmType, tntCtx string, idxKeys ...string) (err error)
	GetStatQueueProfileDrv(ctx *context.Context, tenant string, ID string) (sq *utils.StatQueueProfile, err error)
	SetStatQueueProfileDrv(ctx *context.Context, sq *utils.StatQueueProfile) (err error)
	RemStatQueueProfileDrv(ctx *context.Context, tenant, id string) (err error)
	GetStatQueueDrv(ctx *context.Context, tenant, id string) (sq *utils.StatQueue, err error)
	SetStatQueueDrv(ctx *context.Context, ssq *StoredStatQueue, sq *utils.StatQueue) error
	RemStatQueueDrv(ctx *context.Context, tenant, id string) (err error)
	GetThresholdProfileDrv(ctx *context.Context, tenant string, ID string) (tp *utils.ThresholdProfile, err error)
	SetThresholdProfileDrv(ctx *context.Context, tp *utils.ThresholdProfile) (err error)
	RemThresholdProfileDrv(ctx *context.Context, tenant, id string) (err error)
	GetThresholdDrv(*context.Context, string, string) (*utils.Threshold, error)
	SetThresholdDrv(*context.Context, *utils.Threshold) error
	RemoveThresholdDrv(*context.Context, string, string) error
	SetRankingProfileDrv(ctx *context.Context, rp *utils.RankingProfile) (err error)
	GetRankingProfileDrv(ctx *context.Context, tenant string, id string) (sq *utils.RankingProfile, err error)
	RemRankingProfileDrv(ctx *context.Context, tenant string, id string) (err error)
	SetRankingDrv(ctx *context.Context, rn *utils.Ranking) (err error)
	GetRankingDrv(ctx *context.Context, tenant string, id string) (sq *utils.Ranking, err error)
	RemoveRankingDrv(ctx *context.Context, tenant string, id string) (err error)
	SetTrendProfileDrv(ctx *context.Context, sq *utils.TrendProfile) (err error)
	GetTrendProfileDrv(ctx *context.Context, tenant string, id string) (sq *utils.TrendProfile, err error)
	RemTrendProfileDrv(ctx *context.Context, tenant string, id string) (err error)
	GetTrendDrv(ctx *context.Context, tenant string, id string) (*utils.Trend, error)
	SetTrendDrv(ctx *context.Context, tr *utils.Trend) error
	RemoveTrendDrv(ctx *context.Context, tenant string, id string) error
	GetFilterDrv(ctx *context.Context, tnt string, id string) (*Filter, error)
	SetFilterDrv(ctx *context.Context, f *Filter) error
	RemoveFilterDrv(ctx *context.Context, tnt string, id string) error
	GetRouteProfileDrv(*context.Context, string, string) (*utils.RouteProfile, error)
	SetRouteProfileDrv(*context.Context, *utils.RouteProfile) error
	RemoveRouteProfileDrv(*context.Context, string, string) error
	GetAttributeProfileDrv(ctx *context.Context, tnt string, id string) (*utils.AttributeProfile, error)
	SetAttributeProfileDrv(ctx *context.Context, attr *utils.AttributeProfile) error
	RemoveAttributeProfileDrv(*context.Context, string, string) error
	GetChargerProfileDrv(*context.Context, string, string) (*utils.ChargerProfile, error)
	SetChargerProfileDrv(*context.Context, *utils.ChargerProfile) error
	RemoveChargerProfileDrv(*context.Context, string, string) error
	GetItemLoadIDsDrv(ctx *context.Context, itemIDPrefix string) (loadIDs map[string]int64, err error)
	SetLoadIDsDrv(ctx *context.Context, loadIDs map[string]int64) error
	RemoveLoadIDsDrv() error
	GetRateProfileDrv(*context.Context, string, string) (*utils.RateProfile, error)
	GetRateProfileRatesDrv(*context.Context, string, string, string, bool) ([]string, []*utils.Rate, error)
	SetRateProfileDrv(*context.Context, *utils.RateProfile, bool) error
	RemoveRateProfileDrv(*context.Context, string, string, *[]string) error
	GetActionProfileDrv(*context.Context, string, string) (*utils.ActionProfile, error)
	SetActionProfileDrv(*context.Context, *utils.ActionProfile) error
	RemoveActionProfileDrv(*context.Context, string, string) error
	GetAccountDrv(*context.Context, string, string) (*utils.Account, error)
	SetAccountDrv(ctx *context.Context, profile *utils.Account) error
	RemoveAccountDrv(*context.Context, string, string) error
	GetConfigSectionsDrv(*context.Context, string, []string) (map[string][]byte, error)
	SetConfigSectionsDrv(*context.Context, string, map[string][]byte) error
	RemoveConfigSectionsDrv(*context.Context, string, []string) error
	SetCDR(*context.Context, *utils.CGREvent, bool) error
	GetCDRs(*context.Context, []*Filter, map[string]any) ([]*utils.CDR, error)
	RemoveCDRs(*context.Context, []*Filter) error
	DumpDB() error
	RewriteDB() error
	BackupDB(string, bool) error
	RestoreDB(string) error
	SnapshotDB(string, bool) error
}

// DBDriver used as a DataDB but also as a ConfigProvider
type DBDriver interface {
	DataDB
	config.ConfigDB
}

// Decide the value of cacheCommit parameter based on transactionID
func cacheCommit(transactionID string) bool {
	return transactionID == utils.NonTransactional
}
