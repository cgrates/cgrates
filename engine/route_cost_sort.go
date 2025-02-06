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
	"slices"

	"github.com/cgrates/birpc/context"
	"github.com/ericlagergren/decimal"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func populateCostForRoutes(ctx *context.Context, cfg *config.CGRConfig, connMgr *ConnManager,
	fltrS *FilterS, routes map[string]*RouteWithWeight, ev *utils.CGREvent,
	extraOpts *optsGetRoutes) (sortedRoutes []*SortedRoute, err error) {
	if len(cfg.RouteSCfg().RateSConns) == 0 {
		return nil, utils.NewErrMandatoryIeMissing("connIDs")
	}
	var usage *decimal.Big
	if usage, err = GetDecimalBigOpts(ctx, ev.Tenant, ev.AsDataProvider(), fltrS, cfg.RouteSCfg().Opts.Usage,
		utils.OptsRoutesUsage, utils.MetaUsage); err != nil {
		return
	}
	ev.APIOpts[utils.MetaUsage] = usage
	sortedRoutes = make([]*SortedRoute, 0, len(routes))
	for _, route := range routes {
		if len(route.RateProfileIDs) == 0 && len(route.AccountIDs) == 0 {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> supplier: <%s> - empty RateProfileIDs or AccountIDs",
					utils.RouteS, route.ID))
			return nil, utils.NewErrMandatoryIeMissing("RateProfileIDs or AccountIDs")
		}
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
		var cost *utils.Decimal
		if len(route.AccountIDs) != 0 { // query AccountS for cost

			var acntCost utils.EventCharges
			ev.APIOpts[utils.OptsAccountsProfileIDs] = slices.Clone(route.AccountIDs)
			if err = connMgr.Call(ctx, cfg.RouteSCfg().AccountSConns,
				utils.AccountSv1MaxAbstracts, ev, &acntCost); err != nil {
				if extraOpts.ignoreErrors {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> ignoring route with ID: %s, err: %s",
							utils.RouteS, route.ID, err.Error()))
					err = nil
					continue
				}
				err = utils.NewErrAccountS(err)
				return
			}
			if acntCost.Concretes != nil {
				cost = acntCost.Concretes
				if costFloat64, _ := cost.Float64(); extraOpts.maxCost != 0 && costFloat64 > extraOpts.maxCost {
					continue
				}
			}
			acntIDs := make([]string, 0, len(acntCost.Accounts))
			for acntID := range acntCost.Accounts {
				acntIDs = append(acntIDs, acntID)
			}
			srtRoute.SortingData[utils.AccountIDs] = acntIDs
		} else { // query RateS for cost
			ev.APIOpts[utils.OptsRatesProfileIDs] = slices.Clone(route.RateProfileIDs)
			var rpCost utils.RateProfileCost
			if err = connMgr.Call(ctx, cfg.RouteSCfg().RateSConns,
				utils.RateSv1CostForEvent, ev, &rpCost); err != nil {
				if extraOpts.ignoreErrors {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> ignoring route with ID: %s, err: %s",
							utils.RouteS, route.ID, err.Error()))
					err = nil
					continue
				}
				err = utils.NewErrRateS(err)
				return
			}
			cost = rpCost.Cost
			if costFloat64, _ := cost.Float64(); extraOpts.maxCost != 0 && costFloat64 > extraOpts.maxCost {
				continue
			}
			srtRoute.SortingData[utils.RateProfileID] = rpCost.ID
		}
		if cost != nil {
			srtRoute.sortingDataDecimal[utils.Cost] = cost
		}
		srtRoute.SortingData[utils.Cost] = cost

		var pass bool
		if pass, err = routeLazyPass(ctx, route.lazyCheckRules, ev, srtRoute.SortingData,
			cfg.FilterSCfg().ResourceSConns,
			cfg.FilterSCfg().StatSConns,
			cfg.FilterSCfg().AccountSConns,
			cfg.FilterSCfg().TrendSConns,
			cfg.FilterSCfg().RankingSConns); err != nil {
			return
		} else if pass {
			sortedRoutes = append(sortedRoutes, srtRoute)
		}
	}
	return
}

func NewHighestCostSorter(cfg *config.CGRConfig, connMgr *ConnManager, fltrS *FilterS) *HightCostSorter {
	return &HightCostSorter{cfg: cfg, connMgr: connMgr, fltrS: fltrS}
}

// HightCostSorter sorts routes based on their cost
type HightCostSorter struct {
	cfg     *config.CGRConfig
	connMgr *ConnManager
	fltrS   *FilterS
}

func (hcs *HightCostSorter) SortRoutes(ctx *context.Context, prflID string, routes map[string]*RouteWithWeight,
	ev *utils.CGREvent, extraOpts *optsGetRoutes) (sortedRoutes *SortedRoutes, err error) {
	var sRoutes []*SortedRoute
	if sRoutes, err = populateCostForRoutes(ctx, hcs.cfg, hcs.connMgr, hcs.fltrS, routes, ev, extraOpts); err != nil {
		return
	}

	sortedRoutes = &SortedRoutes{
		ProfileID: prflID,
		Sorting:   utils.MetaHC,
		Routes:    sRoutes,
	}
	sortedRoutes.SortHighestCost()
	return
}

func NewLeastCostSorter(cfg *config.CGRConfig, connMgr *ConnManager, fltrS *FilterS) *LeastCostSorter {
	return &LeastCostSorter{cfg: cfg, connMgr: connMgr, fltrS: fltrS}
}

// LeastCostSorter sorts routes based on their cost
type LeastCostSorter struct {
	cfg     *config.CGRConfig
	connMgr *ConnManager
	fltrS   *FilterS
}

func (lcs *LeastCostSorter) SortRoutes(ctx *context.Context, prflID string, routes map[string]*RouteWithWeight,
	ev *utils.CGREvent, extraOpts *optsGetRoutes) (sortedRoutes *SortedRoutes, err error) {
	var sRoutes []*SortedRoute
	if sRoutes, err = populateCostForRoutes(ctx, lcs.cfg, lcs.connMgr, lcs.fltrS, routes, ev, extraOpts); err != nil {
		return
	}
	sortedRoutes = &SortedRoutes{
		ProfileID: prflID,
		Sorting:   utils.MetaLC,
		Routes:    sRoutes,
	}
	sortedRoutes.SortLeastCost()
	return
}
