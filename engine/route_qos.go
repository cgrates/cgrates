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
	"math"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func NewQOSRouteSorter(cfg *config.CGRConfig, connMgr *ConnManager) *QOSRouteSorter {
	return &QOSRouteSorter{cfg: cfg, connMgr: connMgr}
}

// QOSSorter sorts route based on stats
type QOSRouteSorter struct {
	cfg     *config.CGRConfig
	connMgr *ConnManager
}

func (qos *QOSRouteSorter) SortRoutes(ctx *context.Context, prflID string, routes map[string]*RouteWithWeight,
	ev *utils.CGREvent, extraOpts *optsGetRoutes) (sortedRoutes *SortedRoutes, err error) {
	if len(qos.cfg.RouteSCfg().StatSConns) == 0 {
		return nil, utils.NewErrMandatoryIeMissing("connIDs")
	}
	sortedRoutes = &SortedRoutes{
		ProfileID: prflID,
		Sorting:   utils.MetaQOS,
		Routes:    make([]*SortedRoute, 0, len(routes)),
	}
	for _, route := range routes {
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
		var metricSupp map[string]float64
		if metricSupp, err = populatStatsForQOSRoute(ctx, qos.cfg, qos.connMgr, route.StatIDs, ev.Tenant); err != nil { //create metric map for route
			if extraOpts.ignoreErrors {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> ignoring route with ID: %s, err: %s",
						utils.RouteS, route.ID, err.Error()))
				continue
			}
			return
		}
		// add metrics from statIDs in SortingData
		for key, val := range metricSupp {
			srtRoute.SortingData[key] = val
			srtRoute.sortingDataF64[key] = val
		}
		// check if the route have the metric from sortingParameters
		// in case that the metric don't exist
		// we use 10000000 for *pdd and -1 for others
		for _, metric := range extraOpts.sortingParameters {
			if _, hasMetric := metricSupp[metric]; !hasMetric {
				if metric == utils.MetaPDD {
					srtRoute.SortingData[metric] = math.MaxFloat64
					srtRoute.sortingDataF64[metric] = math.MaxFloat64
				} else {
					srtRoute.SortingData[metric] = -1.0
					srtRoute.sortingDataF64[metric] = -1.0
				}
			}
		}
		var pass bool
		if pass, err = routeLazyPass(ctx, route.lazyCheckRules, ev, srtRoute.SortingData,
			qos.cfg.FilterSCfg().ResourceSConns,
			qos.cfg.FilterSCfg().StatSConns,
			qos.cfg.FilterSCfg().AccountSConns); err != nil {
			return
		} else if pass {
			sortedRoutes.Routes = append(sortedRoutes.Routes, srtRoute)
		}
	}
	sortedRoutes.SortQOS(extraOpts.sortingParameters)
	return
}

// populatStatsForQOSRoute will query a list of statIDs and return composed metric values
// first metric found is always returned
func populatStatsForQOSRoute(ctx *context.Context, cfg *config.CGRConfig,
	connMgr *ConnManager, statIDs []string, tenant string) (stsMetric map[string]float64, err error) {
	type metric struct {
		sum float64
		len int
	}
	stsMetric = make(map[string]float64)
	provStsMetrics := make(map[string]metric)
	for _, statID := range statIDs {
		var metrics map[string]float64
		if err = connMgr.Call(ctx, cfg.RouteSCfg().StatSConns, utils.StatSv1GetQueueFloatMetrics,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: statID}}, &metrics); err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s getting statMetrics for stat : %s", utils.RouteS, err.Error(), statID))
			return
		}
		for key, val := range metrics {
			//add value of metric in a slice in case that we get the same metric from different stat
			provStsMetrics[key] = metric{
				sum: provStsMetrics[key].sum + val,
				len: provStsMetrics[key].len + 1,
			}
		}
	}
	for metric, val := range provStsMetrics {
		stsMetric[metric] = val.sum / float64(val.len)
	}
	return
}
