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
	"fmt"
	"strings"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// NewLoadDistributionSorter .
func NewLoadDistributionSorter(cfg *config.CGRConfig, connMgr *ConnManager) *LoadDistributionSorter {
	return &LoadDistributionSorter{cfg: cfg, connMgr: connMgr}
}

// LoadDistributionSorter orders suppliers based on their Resource Usage
type LoadDistributionSorter struct {
	cfg     *config.CGRConfig
	connMgr *ConnManager
}

// SortRoutes .
func (ws *LoadDistributionSorter) SortRoutes(ctx *context.Context, prflID string,
	routes map[string]*RouteWithWeight, ev *utils.CGREvent, extraOpts *optsGetRoutes) (sortedRoutes *SortedRoutes, err error) {
	if len(ws.cfg.RouteSCfg().StatSConns) == 0 {
		return nil, utils.NewErrMandatoryIeMissing("connIDs")
	}
	sortedRoutes = &SortedRoutes{
		ProfileID: prflID,
		Sorting:   utils.MetaLoad,
		Routes:    make([]*SortedRoute, 0, len(routes)),
	}
	for _, route := range routes {
		// we should have at least 1 statID defined for counting CDR (a.k.a *sum:1)
		if len(route.StatIDs) == 0 {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> supplier: <%s> - empty StatIDs",
					utils.RouteS, route.ID))
			return nil, utils.NewErrMandatoryIeMissing("StatIDs")
		}
		srtRoute := &SortedRoute{
			RouteID: route.ID,
			SortingData: map[string]interface{}{
				utils.Weight: route.Weight,
			},
			sortingDataF64: map[string]float64{
				utils.Weight: route.Weight,
			},
			RouteParameters: route.RouteParameters,
		}
		var metricSum float64
		if metricSum, err = populateStatsForLoadRoute(ctx, ws.cfg, ws.connMgr, route.StatIDs, ev.Tenant); err != nil { //create metric map for route
			if extraOpts.ignoreErrors {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> ignoring route with ID: %s, err: %s",
						utils.RouteS, route.ID, err.Error()))
				continue
			}
			return
		}
		srtRoute.SortingData[utils.Load] = metricSum
		srtRoute.sortingDataF64[utils.Load] = metricSum
		var pass bool
		if pass, err = routeLazyPass(ctx, route.lazyCheckRules, ev, srtRoute.SortingData,
			ws.cfg.FilterSCfg().ResourceSConns,
			ws.cfg.FilterSCfg().StatSConns,
			ws.cfg.FilterSCfg().AccountSConns); err != nil {
			return
		} else if pass {
			// Add the ratio in SortingData so we can used it later in SortLoadDistribution
			floatRatio, err := utils.IfaceAsFloat64(route.cacheRoute[utils.MetaRatio])
			if err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> cannot convert ratio <%+v> to float64 supplier: <%s>",
						utils.RouteS, route.cacheRoute[utils.MetaRatio], route.ID))
			}
			srtRoute.SortingData[utils.Ratio] = floatRatio
			srtRoute.sortingDataF64[utils.Ratio] = floatRatio
			sortedRoutes.Routes = append(sortedRoutes.Routes, srtRoute)
		}
	}
	sortedRoutes.SortLoadDistribution()
	return
}

// populateStatsForLoadRoute will query a list of statIDs and return the sum of metrics
// first metric found is always returned
func populateStatsForLoadRoute(ctx *context.Context, cfg *config.CGRConfig,
	connMgr *ConnManager, statIDs []string, tenant string) (result float64, err error) {
	for _, statID := range statIDs {
		// check if we get an ID in the following form (StatID:MetricID)
		statWithMetric := strings.Split(statID, utils.InInFieldSep)
		var metrics map[string]float64
		if err = connMgr.Call(ctx,
			cfg.RouteSCfg().StatSConns,
			utils.StatSv1GetQueueFloatMetrics,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: tenant, ID: statWithMetric[0]}},
			&metrics); err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s getting statMetrics for stat : %s",
					utils.RouteS, err.Error(), statWithMetric[0]))
			continue
		}
		if len(statWithMetric) == 2 { // in case we have MetricID defined with StatID we consider only that metric
			// check if statQueue have metric defined
			metricVal, has := metrics[statWithMetric[1]]
			if !has {
				return 0, fmt.Errorf("<%s> error: %s metric %s for statID: %s",
					utils.RouteS, utils.ErrNotFound, statWithMetric[1], statWithMetric[0])
			}
			result += metricVal
		} else { // otherwise we consider all metrics
			for _, val := range metrics {
				//add value of metric in a slice in case that we get the same metric from different stat
				result += val
			}
		}
	}
	return
}
