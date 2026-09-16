// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package routes

import (
	"fmt"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// RoutesSorter is the interface which needs to be implemented by routes sorters
type RoutesSorter interface {
	SortRoutes(*context.Context, string, map[string]*RouteWithWeight, *utils.CGREvent, *optsGetRoutes) (*utils.SortedRoutes, error)
}

// RouteSortDispatcher will initialize strategies
// and dispatch requests to them
type RouteSortDispatcher map[string]RoutesSorter

func (ssd RouteSortDispatcher) SortRoutes(ctx *context.Context, prflID, strategy string,
	suppls map[string]*RouteWithWeight, suplEv *utils.CGREvent, extraOpts *optsGetRoutes) (_ *utils.SortedRoutes, err error) {
	sd, has := ssd[strategy]
	if !has {
		err = fmt.Errorf("unsupported sorting strategy: %s", strategy)
		return
	}
	return sd.SortRoutes(ctx, prflID, suppls, suplEv, extraOpts)
}

// routeLazyPass filters the route based on
func routeLazyPass(ctx *context.Context, filters []*engine.FilterRule, ev *utils.CGREvent, data utils.MapStorage,
	cfg *config.CGRConfig, fS *engine.FilterS) (pass bool, err error) {
	if len(filters) == 0 {
		return true, nil
	}
	dynDP := engine.NewDynamicDP(ctx, cfg, ev.Tenant, utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: data,
	}, fS)
	for _, rule := range filters {
		if pass, err = rule.Pass(ctx, dynDP); err != nil || !pass {
			return
		}
	}
	return true, nil
}

// RouteWithWeight attaches static weight to Route
type RouteWithWeight struct {
	*utils.Route
	Weight         float64
	blocker        bool
	lazyCheckRules []*engine.FilterRule
}
