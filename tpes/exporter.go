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

package tpes

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type exporter struct {
	fileName  string
	keyPrefix string
	dbItemID  string
	header    []string
	loadRows  func(ctx *context.Context, dm *engine.DataManager, tenant, id string) ([][]string, error)
}

var exporters = map[string]exporter{
	utils.MetaAccounts: {
		fileName:  utils.AccountsCsv,
		keyPrefix: utils.AccountPrefix,
		dbItemID:  utils.MetaAccounts,
		header:    accountHeader,
		loadRows: func(ctx *context.Context, dm *engine.DataManager, tenant, id string) ([][]string, error) {
			account, err := dm.GetAccount(ctx, tenant, id)
			if err != nil {
				return nil, err
			}
			return accountRecords(account), nil
		},
	},
	utils.MetaActions: {
		fileName:  utils.ActionsCsv,
		keyPrefix: utils.ActionProfilePrefix,
		dbItemID:  utils.MetaActionProfiles,
		header:    actionHeader,
		loadRows: func(ctx *context.Context, dm *engine.DataManager, tenant, id string) ([][]string, error) {
			profile, err := dm.GetActionProfile(ctx, tenant, id, true, true, utils.NonTransactional)
			if err != nil {
				return nil, err
			}
			return actionProfileRecords(profile), nil
		},
	},
	utils.MetaAttributes: {
		fileName:  utils.AttributesCsv,
		keyPrefix: utils.AttributeProfilePrefix,
		dbItemID:  utils.MetaAttributeProfiles,
		header:    attributeHeader,
		loadRows: func(ctx *context.Context, dm *engine.DataManager, tenant, id string) ([][]string, error) {
			profile, err := dm.GetAttributeProfile(ctx, tenant, id, true, true, utils.NonTransactional)
			if err != nil {
				return nil, err
			}
			return attributeProfileRecords(profile), nil
		},
	},
	utils.MetaChargers: {
		fileName:  utils.ChargersCsv,
		keyPrefix: utils.ChargerProfilePrefix,
		dbItemID:  utils.MetaChargerProfiles,
		header:    chargerHeader,
		loadRows: func(ctx *context.Context, dm *engine.DataManager, tenant, id string) ([][]string, error) {
			profile, err := dm.GetChargerProfile(ctx, tenant, id, true, true, utils.NonTransactional)
			if err != nil {
				return nil, err
			}
			return chargerProfileRecords(profile), nil
		},
	},
	utils.MetaFilters: {
		fileName:  utils.FiltersCsv,
		keyPrefix: utils.FilterPrefix,
		dbItemID:  utils.MetaFilters,
		header:    filterHeader,
		loadRows: func(ctx *context.Context, dm *engine.DataManager, tenant, id string) ([][]string, error) {
			filter, err := dm.GetFilter(ctx, tenant, id, true, true, utils.NonTransactional)
			if err != nil {
				return nil, err
			}
			return filterRecords(filter), nil
		},
	},
	utils.MetaRates: {
		fileName:  utils.RatesCsv,
		keyPrefix: utils.RateProfilePrefix,
		dbItemID:  utils.MetaRateProfiles,
		header:    rateHeader,
		loadRows: func(ctx *context.Context, dm *engine.DataManager, tenant, id string) ([][]string, error) {
			profile, err := dm.GetRateProfile(ctx, tenant, id, true, true, utils.NonTransactional)
			if err != nil {
				return nil, err
			}
			return rateProfileRecords(profile), nil
		},
	},
	utils.MetaRankings: {
		fileName:  utils.RankingsCsv,
		keyPrefix: utils.RankingProfilePrefix,
		dbItemID:  utils.MetaRankingProfiles,
		header:    rankingHeader,
		loadRows: func(ctx *context.Context, dm *engine.DataManager, tenant, id string) ([][]string, error) {
			profile, err := dm.GetRankingProfile(ctx, tenant, id, true, true, utils.NonTransactional)
			if err != nil {
				return nil, err
			}
			return rankingProfileRecords(profile), nil
		},
	},
	utils.MetaResources: {
		fileName:  utils.ResourcesCsv,
		keyPrefix: utils.ResourceProfilesPrefix,
		dbItemID:  utils.MetaResourceProfiles,
		header:    resourceHeader,
		loadRows: func(ctx *context.Context, dm *engine.DataManager, tenant, id string) ([][]string, error) {
			profile, err := dm.GetResourceProfile(ctx, tenant, id, true, true, utils.NonTransactional)
			if err != nil {
				return nil, err
			}
			return resourceProfileRecords(profile), nil
		},
	},
	utils.MetaRoutes: {
		fileName:  utils.RoutesCsv,
		keyPrefix: utils.RouteProfilePrefix,
		dbItemID:  utils.MetaRouteProfiles,
		header:    routeHeader,
		loadRows: func(ctx *context.Context, dm *engine.DataManager, tenant, id string) ([][]string, error) {
			profile, err := dm.GetRouteProfile(ctx, tenant, id, true, true, utils.NonTransactional)
			if err != nil {
				return nil, err
			}
			return routeProfileRecords(profile), nil
		},
	},
	utils.MetaStats: {
		fileName:  utils.StatsCsv,
		keyPrefix: utils.StatQueueProfilePrefix,
		dbItemID:  utils.MetaStatQueueProfiles,
		header:    statHeader,
		loadRows: func(ctx *context.Context, dm *engine.DataManager, tenant, id string) ([][]string, error) {
			profile, err := dm.GetStatQueueProfile(ctx, tenant, id, true, true, utils.NonTransactional)
			if err != nil {
				return nil, err
			}
			return statQueueProfileRecords(profile), nil
		},
	},
	utils.MetaThresholds: {
		fileName:  utils.ThresholdsCsv,
		keyPrefix: utils.ThresholdProfilePrefix,
		dbItemID:  utils.MetaThresholdProfiles,
		header:    thresholdHeader,
		loadRows: func(ctx *context.Context, dm *engine.DataManager, tenant, id string) ([][]string, error) {
			profile, err := dm.GetThresholdProfile(ctx, tenant, id, true, true, utils.NonTransactional)
			if err != nil {
				return nil, err
			}
			return thresholdProfileRecords(profile), nil
		},
	},
	utils.MetaTrends: {
		fileName:  utils.TrendsCsv,
		keyPrefix: utils.TrendProfilePrefix,
		dbItemID:  utils.MetaTrendProfiles,
		header:    trendHeader,
		loadRows: func(ctx *context.Context, dm *engine.DataManager, tenant, id string) ([][]string, error) {
			profile, err := dm.GetTrendProfile(ctx, tenant, id, true, true, utils.NonTransactional)
			if err != nil {
				return nil, err
			}
			return trendProfileRecords(profile), nil
		},
	},
}

func (e exporter) exportItems(ctx *context.Context, dm *engine.DataManager, exportType string, w io.Writer, tenant string, itemIDs []string) error {
	records := make([][]string, 0, len(itemIDs)+1)
	records = append(records, e.header)
	for _, id := range itemIDs {
		rows, err := e.loadRows(ctx, dm, tenant, id)
		if err != nil {
			if errors.Is(err, utils.ErrNotFound) {
				return fmt.Errorf("cannot find %s with id <%s>: %w", strings.TrimPrefix(exportType, utils.Meta), id, err)
			}
			return err
		}
		records = append(records, rows...)
	}
	return newCSVWriter(w).WriteAll(records)
}
