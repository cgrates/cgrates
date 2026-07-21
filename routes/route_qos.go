// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package routes

import (
	"fmt"
	"math"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewQOSRouteSorter(cfg *config.CGRConfig, connMgr *engine.ConnManager, fltrS *engine.FilterS) *QOSRouteSorter {
	return &QOSRouteSorter{cfg: cfg, connMgr: connMgr, fltrS: fltrS}
}

// QOSSorter sorts route based on stats
type QOSRouteSorter struct {
	cfg     *config.CGRConfig
	connMgr *engine.ConnManager
	fltrS   *engine.FilterS
}

func (qos *QOSRouteSorter) SortRoutes(ctx *context.Context, prflID string, routes map[string]*RouteWithWeight,
	ev *utils.CGREvent, extraOpts *optsGetRoutes) (sortedRoutes *SortedRoutes, err error) {
	statSConns, err := engine.GetConnIDs(ctx, qos.cfg.RouteSCfg().Conns, utils.MetaStats, ev.Tenant, ev.AsDataProvider(), nil, qos.fltrS)
	if err != nil {
		return nil, err
	}
	if len(statSConns) == 0 {
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
			SortingData: map[string]any{
				utils.Weight: route.Weight,
			},
			sortingDataDecimal: map[string]*utils.Decimal{
				utils.Weight: utils.NewDecimalFromFloat64(route.Weight),
			},
			RouteParameters: route.RouteParameters,
		}
		if route.blocker {
			srtRoute.SortingData[utils.Blocker] = true
		}
		var metricSupp map[string]*utils.Decimal
		if metricSupp, err = populatStatsForQOSRoute(ctx, qos.cfg, qos.connMgr, qos.fltrS, route.StatIDs, ev.Tenant); err != nil { //create metric map for route
			if extraOpts.ignoreErrors {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> ignoring route with ID: %s, err: %s",
						utils.RouteS, route.ID, err.Error()))
				err = nil
				continue
			}
			return
		}
		// add metrics from statIDs in SortingData
		for key, val := range metricSupp {
			srtRoute.SortingData[key] = val
			srtRoute.sortingDataDecimal[key] = val
		}
		// check if the route have the metric from sortingParameters
		// in case that the metric don't exist
		// we use 10000000 for *pdd and -1 for others
		for _, metric := range extraOpts.sortingParameters {
			if _, hasMetric := metricSupp[metric]; !hasMetric {
				if metric == utils.MetaPDD {
					srtRoute.SortingData[metric] = math.MaxFloat64
					srtRoute.sortingDataDecimal[metric] = utils.NewDecimalFromFloat64(math.MaxFloat64)
				} else {
					srtRoute.SortingData[metric] = -1.0
					srtRoute.sortingDataDecimal[metric] = utils.NewDecimalFromFloat64(-1.0)
				}
			}
		}
		var pass bool
		if pass, err = routeLazyPass(ctx, route.lazyCheckRules, ev, srtRoute.SortingData,
			qos.cfg, qos.fltrS); err != nil {
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
	connMgr *engine.ConnManager, fltrS *engine.FilterS, statIDs []string, tenant string) (stsMetric map[string]*utils.Decimal, err error) {
	type metric struct {
		sum *utils.Decimal
		len int
	}
	connIDs, err := engine.GetConnIDs(ctx, cfg.RouteSCfg().Conns, utils.MetaStats, tenant, utils.MapStorage{}, nil, fltrS)
	if err != nil {
		return
	}
	stsMetric = make(map[string]*utils.Decimal)
	provStsMetrics := make(map[string]metric)
	for _, statID := range statIDs {
		var metrics map[string]*utils.Decimal
		if err = connMgr.Call(ctx, connIDs, utils.StatSv1GetQueueDecimalMetrics,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: statID}}, &metrics); err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s getting statMetrics for stat : %s", utils.RouteS, err.Error(), statID))
			return
		}
		for key, val := range metrics {
			//add value of metric in a slice in case that we get the same metric from different stat
			provStsMetrics[key] = metric{
				sum: utils.SumDecimal(provStsMetrics[key].sum, val),
				len: provStsMetrics[key].len + 1,
			}
		}
	}
	for metric, val := range provStsMetrics {
		stsMetric[metric] = utils.DivideDecimal(val.sum, utils.NewDecimal(int64(val.len), 0))
	}
	return
}
