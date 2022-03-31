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

package engine

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

type Storage interface {
	Close()
	Flush(string) error
	GetKeysForPrefix(ctx *context.Context, prefix string) ([]string, error)
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
	GetResourceProfileDrv(*context.Context, string, string) (*ResourceProfile, error)
	SetResourceProfileDrv(*context.Context, *ResourceProfile) error
	RemoveResourceProfileDrv(*context.Context, string, string) error
	GetResourceDrv(*context.Context, string, string) (*Resource, error)
	SetResourceDrv(*context.Context, *Resource) error
	RemoveResourceDrv(*context.Context, string, string) error
	GetLoadHistory(int, bool, string) ([]*utils.LoadInstance, error)
	AddLoadHistory(*utils.LoadInstance, int, string) error
	GetIndexesDrv(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error)
	SetIndexesDrv(ctx *context.Context, idxItmType, tntCtx string,
		indexes map[string]utils.StringSet, commit bool, transactionID string) (err error)
	RemoveIndexesDrv(ctx *context.Context, idxItmType, tntCtx, idxKey string) (err error)
	GetStatQueueProfileDrv(ctx *context.Context, tenant string, ID string) (sq *StatQueueProfile, err error)
	SetStatQueueProfileDrv(ctx *context.Context, sq *StatQueueProfile) (err error)
	RemStatQueueProfileDrv(ctx *context.Context, tenant, id string) (err error)
	GetStatQueueDrv(ctx *context.Context, tenant, id string) (sq *StatQueue, err error)
	SetStatQueueDrv(ctx *context.Context, ssq *StoredStatQueue, sq *StatQueue) (err error)
	RemStatQueueDrv(ctx *context.Context, tenant, id string) (err error)
	GetThresholdProfileDrv(ctx *context.Context, tenant string, ID string) (tp *ThresholdProfile, err error)
	SetThresholdProfileDrv(ctx *context.Context, tp *ThresholdProfile) (err error)
	RemThresholdProfileDrv(ctx *context.Context, tenant, id string) (err error)
	GetThresholdDrv(*context.Context, string, string) (*Threshold, error)
	SetThresholdDrv(*context.Context, *Threshold) error
	RemoveThresholdDrv(*context.Context, string, string) error
	GetFilterDrv(ctx *context.Context, tnt string, id string) (*Filter, error)
	SetFilterDrv(ctx *context.Context, f *Filter) error
	RemoveFilterDrv(ctx *context.Context, tnt string, id string) error
	GetRouteProfileDrv(*context.Context, string, string) (*RouteProfile, error)
	SetRouteProfileDrv(*context.Context, *RouteProfile) error
	RemoveRouteProfileDrv(*context.Context, string, string) error
	GetAttributeProfileDrv(ctx *context.Context, tnt string, id string) (*AttributeProfile, error)
	SetAttributeProfileDrv(ctx *context.Context, attr *AttributeProfile) error
	RemoveAttributeProfileDrv(*context.Context, string, string) error
	GetChargerProfileDrv(*context.Context, string, string) (*ChargerProfile, error)
	SetChargerProfileDrv(*context.Context, *ChargerProfile) error
	RemoveChargerProfileDrv(*context.Context, string, string) error
	GetDispatcherProfileDrv(*context.Context, string, string) (*DispatcherProfile, error)
	SetDispatcherProfileDrv(*context.Context, *DispatcherProfile) error
	RemoveDispatcherProfileDrv(*context.Context, string, string) error
	GetItemLoadIDsDrv(ctx *context.Context, itemIDPrefix string) (loadIDs map[string]int64, err error)
	SetLoadIDsDrv(ctx *context.Context, loadIDs map[string]int64) error
	RemoveLoadIDsDrv() error
	GetDispatcherHostDrv(*context.Context, string, string) (*DispatcherHost, error)
	SetDispatcherHostDrv(*context.Context, *DispatcherHost) error
	RemoveDispatcherHostDrv(*context.Context, string, string) error
	GetRateProfileDrv(*context.Context, string, string) (*utils.RateProfile, error)
	GetRateProfileRatesDrv(*context.Context, string, string, string, bool) ([]string, []*utils.Rate, error)
	SetRateProfileDrv(*context.Context, *utils.RateProfile, bool) error
	RemoveRateProfileDrv(*context.Context, string, string, *[]string) error
	GetActionProfileDrv(*context.Context, string, string) (*ActionProfile, error)
	SetActionProfileDrv(*context.Context, *ActionProfile) error
	RemoveActionProfileDrv(*context.Context, string, string) error
	GetAccountDrv(*context.Context, string, string) (*utils.Account, error)
	SetAccountDrv(ctx *context.Context, profile *utils.Account) error
	RemoveAccountDrv(*context.Context, string, string) error
	GetConfigSectionsDrv(*context.Context, string, []string) (map[string][]byte, error)
	SetConfigSectionsDrv(*context.Context, string, map[string][]byte) error
	RemoveConfigSectionsDrv(*context.Context, string, []string) error
}

// DataDBDriver used as a DataDB but also as a ConfigProvider
type DataDBDriver interface {
	DataDB
	config.ConfigDB
}

type LoadStorage interface {
	Storage
	LoadReader
	LoadWriter
}

// LoadReader reads from .csv or TP tables and provides the data ready for the tp_db or data_db.
type LoadReader interface {
	GetTpIds(string) ([]string, error)
	GetTpTableIds(string, string, []string,
		map[string]string, *utils.PaginatorWithSearch) ([]string, error)
	GetTPResources(string, string, string) ([]*utils.TPResourceProfile, error)
	GetTPStats(string, string, string) ([]*utils.TPStatProfile, error)
	GetTPThresholds(string, string, string) ([]*utils.TPThresholdProfile, error)
	GetTPFilters(string, string, string) ([]*utils.TPFilterProfile, error)
	GetTPRoutes(string, string, string) ([]*utils.TPRouteProfile, error)
	GetTPAttributes(string, string, string) ([]*utils.TPAttributeProfile, error)
	GetTPChargers(string, string, string) ([]*utils.TPChargerProfile, error)
	GetTPDispatcherProfiles(string, string, string) ([]*utils.TPDispatcherProfile, error)
	GetTPDispatcherHosts(string, string, string) ([]*utils.TPDispatcherHost, error)
	GetTPRateProfiles(string, string, string) ([]*utils.TPRateProfile, error)
	GetTPActionProfiles(string, string, string) ([]*utils.TPActionProfile, error)
	GetTPAccounts(string, string, string) ([]*utils.TPAccount, error)
}

type LoadWriter interface {
	RemTpData(string, string, map[string]string) error
	SetTPResources([]*utils.TPResourceProfile) error
	SetTPStats([]*utils.TPStatProfile) error
	SetTPThresholds([]*utils.TPThresholdProfile) error
	SetTPFilters([]*utils.TPFilterProfile) error
	SetTPRoutes([]*utils.TPRouteProfile) error
	SetTPAttributes([]*utils.TPAttributeProfile) error
	SetTPChargers([]*utils.TPChargerProfile) error
	SetTPDispatcherProfiles([]*utils.TPDispatcherProfile) error
	SetTPDispatcherHosts([]*utils.TPDispatcherHost) error
	SetTPRateProfiles([]*utils.TPRateProfile) error
	SetTPActionProfiles([]*utils.TPActionProfile) error
	SetTPAccounts([]*utils.TPAccount) error
}

// Decide the value of cacheCommit parameter based on transactionID
func cacheCommit(transactionID string) bool {
	return transactionID == utils.NonTransactional
}
