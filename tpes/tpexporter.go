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

package tpes

import (
	"io"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var tpExporterTypes = utils.NewStringSet([]string{
	utils.MetaAttributes,
	utils.MetaResources,
	utils.MetaFilters,
	utils.MetaRates,
	utils.MetaTrends,
	utils.MetaChargers,
	utils.MetaRoutes,
	utils.MetaAccounts,
	utils.MetaStats,
	utils.MetaActions,
	utils.MetaThresholds,
})

var exportFileName = map[string]string{
	utils.MetaAttributes: utils.AttributesCsv,
	utils.MetaResources:  utils.ResourcesCsv,
	utils.MetaFilters:    utils.FiltersCsv,
	utils.MetaStats:      utils.StatsCsv,
	utils.MetaThresholds: utils.ThresholdsCsv,
	utils.MetaTrends:     utils.TrendsCsv,
	utils.MetaRoutes:     utils.RoutesCsv,
	utils.MetaChargers:   utils.ChargersCsv,
	utils.MetaRates:      utils.RatesCsv,
	utils.MetaActions:    utils.ActionsCsv,
	utils.MetaAccounts:   utils.AccountsCsv,
}

// tpExporter is the interface implementing exports of tariff plan items
type tpExporter interface {
	exportItems(ctx *context.Context, wrtr io.Writer, tnt string, itmIDs []string) (err error)
}

// newTPExporter constructs tpExporters
func newTPExporter(expType string, dm *engine.DataManager) (tpE tpExporter, err error) {
	switch expType {
	case utils.MetaAttributes:
		return newTPAttributes(dm), nil
	case utils.MetaResources:
		return newTPResources(dm), nil
	case utils.MetaFilters:
		return newTPFilters(dm), nil
	case utils.MetaRates:
		return newTPRates(dm), nil
	case utils.MetaChargers:
		return newTPChargers(dm), nil
	case utils.MetaRoutes:
		return newTPRoutes(dm), nil
	case utils.MetaAccounts:
		return newTPAccounts(dm), nil
	case utils.MetaStats:
		return newTPStats(dm), nil
	case utils.MetaRankings:
		return newTPRankings(dm), nil
	case utils.MetaTrends:
		return newTPTrends(dm), nil
	case utils.MetaActions:
		return newTPActions(dm), nil
	case utils.MetaThresholds:
		return newTPThresholds(dm), nil
	default:
		return nil, utils.ErrPrefix(utils.ErrUnsupportedTPExporterType, expType)
	}
}
