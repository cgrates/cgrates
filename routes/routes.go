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

package routes

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewRouteService initializes the Route Service
func NewRouteService(dm *engine.DataManager,
	filterS *engine.FilterS, cfg *config.CGRConfig, connMgr *engine.ConnManager) (rS *RouteS) {
	rS = &RouteS{
		dm:      dm,
		fltrS:   filterS,
		cfg:     cfg,
		connMgr: connMgr,
		sorter: RouteSortDispatcher{
			utils.MetaWeight: NewWeightSorter(cfg, connMgr),
			utils.MetaLC:     NewLeastCostSorter(cfg, connMgr, filterS),
			utils.MetaHC:     NewHighestCostSorter(cfg, connMgr, filterS),
			utils.MetaQOS:    NewQOSRouteSorter(cfg, connMgr),
			utils.MetaReas:   NewResourceAscendetSorter(cfg, connMgr),
			utils.MetaReds:   NewResourceDescendentSorter(cfg, connMgr),
			utils.MetaLoad:   NewLoadDistributionSorter(cfg, connMgr),
		},
	}
	return
}

// RouteS is the service computing route queries
type RouteS struct {
	dm      *engine.DataManager
	fltrS   *engine.FilterS
	cfg     *config.CGRConfig
	sorter  RouteSortDispatcher
	connMgr *engine.ConnManager
}

// matchingRouteProfilesForEvent returns ordered list of matching resources which are active by the time of the call
func (rpS *RouteS) matchingRouteProfilesForEvent(ctx *context.Context, tnt string, ev *utils.CGREvent) (matchingRPrf []*utils.RouteProfile, err error) {
	evNm := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
	}
	rPrfIDs, err := engine.MatchingItemIDsForEvent(ctx, evNm,
		rpS.cfg.RouteSCfg().StringIndexedFields,
		rpS.cfg.RouteSCfg().PrefixIndexedFields,
		rpS.cfg.RouteSCfg().SuffixIndexedFields,
		rpS.cfg.RouteSCfg().ExistsIndexedFields,
		rpS.cfg.RouteSCfg().NotExistsIndexedFields,
		rpS.dm, utils.CacheRouteFilterIndexes, tnt,
		rpS.cfg.RouteSCfg().IndexedSelects,
		rpS.cfg.RouteSCfg().NestedFields,
	)
	if err != nil {
		return nil, err
	}
	matchingRPrf = make([]*utils.RouteProfile, 0, len(rPrfIDs))
	weights := make(map[string]float64) // stores sorting weights by profile ID
	for lpID := range rPrfIDs {
		var rPrf *utils.RouteProfile
		if rPrf, err = rpS.dm.GetRouteProfile(ctx, tnt, lpID, true, true, utils.NonTransactional); err != nil {
			if err == utils.ErrNotFound {
				continue
			}
			return
		}
		var pass bool
		if pass, err = rpS.fltrS.Pass(ctx, tnt, rPrf.FilterIDs,
			evNm); err != nil {
			return
		} else if !pass {
			continue
		}
		weight, err := engine.WeightFromDynamics(ctx, rPrf.Weights, rpS.fltrS, ev.Tenant, evNm)
		if err != nil {
			return nil, err
		}
		weights[rPrf.ID] = weight
		matchingRPrf = append(matchingRPrf, rPrf)
	}
	if len(matchingRPrf) == 0 {
		return nil, utils.ErrNotFound
	}

	// Sort by weight (higher values first).
	slices.SortFunc(matchingRPrf, func(a, b *utils.RouteProfile) int {
		return cmp.Compare(weights[b.ID], weights[a.ID])
	})

	for i, rp := range matchingRPrf {
		var blocker bool
		if blocker, err = engine.BlockerFromDynamics(ctx, rp.Blockers, rpS.fltrS, ev.Tenant, evNm); err != nil {
			return
		}
		if blocker {
			matchingRPrf = matchingRPrf[0 : i+1]
			break
		}
	}
	return
}

func newOptsGetRoutes(ctx *context.Context, ev *utils.CGREvent, fS *engine.FilterS, cfgOpts *config.RoutesOpts) (opts *optsGetRoutes, err error) {
	var ignoreErrors bool
	evNM := ev.AsDataProvider()
	if ignoreErrors, err = engine.GetBoolOpts(ctx, ev.Tenant, evNM, nil, fS, cfgOpts.IgnoreErrors,
		utils.OptsRoutesIgnoreErrors); err != nil {
		return
	}
	opts = &optsGetRoutes{
		ignoreErrors: ignoreErrors,
		paginator:    &utils.Paginator{},
	}
	var limit *int
	if limit, err = engine.GetIntPointerOpts(ctx, ev.Tenant, evNM, nil, fS, cfgOpts.Limit,
		utils.OptsRoutesLimit); err != nil {
		return
	} else {
		opts.paginator.Limit = limit
	}
	var offset *int
	if offset, err = engine.GetIntPointerOpts(ctx, ev.Tenant, evNM, nil, fS, cfgOpts.Offset,
		utils.OptsRoutesOffset); err != nil {
		return
	} else {
		opts.paginator.Offset = offset
	}
	var maxItems *int
	if maxItems, err = engine.GetIntPointerOpts(ctx, ev.Tenant, evNM, nil, fS, cfgOpts.MaxItems,
		utils.OptsRoutesMaxItems); err != nil {
		return
	} else {
		opts.paginator.MaxItems = maxItems
	}

	var maxCost any
	if maxCost, err = engine.GetInterfaceOpts(ctx, ev.Tenant, evNM, nil, fS, cfgOpts.MaxCost, config.RoutesMaxCostDftOpt,
		utils.OptsRoutesMaxCost); err != nil {
		return
	}

	switch maxCost {
	case utils.EmptyString, nil:
	case utils.MetaEventCost:
		if err = ev.CheckMandatoryFields([]string{utils.AccountField,
			utils.Destination, utils.SetupTime, utils.Usage}); err != nil {
			return
		}
	// ToDoNext: rates.V1CostForEvent
	// cd, err := NewCallDescriptorFromCGREvent(attr.CGREvent,
	// 	config.CgrConfig().GeneralCfg().DefaultTimezone)
	// if err != nil {
	// 	return nil, err
	// }
	// cc, err := cd.GetCost()
	// if err != nil {
	// 	return nil, err
	// }
	// opts.maxCost = cc.Cost
	default:
		if opts.maxCost, err = utils.IfaceAsFloat64(maxCost); err != nil {
			return nil, err
		}
	}

	return
}

type optsGetRoutes struct {
	ignoreErrors      bool
	maxCost           float64
	paginator         *utils.Paginator
	sortingParameters []string //used for QOS strategy
	sortingStrategy   string
}

var lazyRouteFltrPrfxs = []string{utils.DynamicDataPrefix + utils.MetaReq,
	utils.DynamicDataPrefix + utils.MetaAccounts,
	utils.DynamicDataPrefix + utils.MetaResources,
	utils.DynamicDataPrefix + utils.MetaStats}

// sortedRoutesForEvent will return the list of valid route IDs
// for event based on filters and sorting algorithms
func (rpS *RouteS) sortedRoutesForProfile(ctx *context.Context, tnt string, rPrfl *utils.RouteProfile, ev *utils.CGREvent,
	pag utils.Paginator, extraOpts *optsGetRoutes) (sortedRoutes *SortedRoutes, err error) {
	extraOpts.sortingParameters = rPrfl.SortingParameters // populate sortingParameters in extraOpts
	extraOpts.sortingStrategy = rPrfl.Sorting             // populate sortingStrategy in extraOpts
	//construct the DP and pass it to filterS
	nM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
	}
	passedRoutes := make(map[string]*RouteWithWeight)
	// apply filters for event
	for _, route := range rPrfl.Routes {
		var pass bool
		var lazyCheckRules []*engine.FilterRule
		if pass, lazyCheckRules, err = rpS.fltrS.LazyPass(ctx, tnt,
			route.FilterIDs, nM, lazyRouteFltrPrfxs); err != nil {
			return
		} else if !pass {
			continue
		}
		var weight float64
		if weight, err = engine.WeightFromDynamics(ctx, route.Weights,
			rpS.fltrS, ev.Tenant, nM); err != nil {
			return
		}
		if prev, has := passedRoutes[route.ID]; !has || prev.Weight < weight {
			var blocker bool
			if blocker, err = engine.BlockerFromDynamics(ctx, route.Blockers, rpS.fltrS, tnt, nM); err != nil {
				return
			}
			passedRoutes[route.ID] = &RouteWithWeight{
				Route:          route,
				lazyCheckRules: lazyCheckRules,
				Weight:         weight,
				blocker:        blocker,
			}
		}
	}

	if sortedRoutes, err = rpS.sorter.SortRoutes(ctx, rPrfl.ID, rPrfl.Sorting,
		passedRoutes, ev, extraOpts); err != nil {
		return nil, err
	}
	for i, sortedRoute := range sortedRoutes.Routes {
		if _, has := sortedRoute.SortingData[utils.Blocker]; has {
			sortedRoutes.Routes = sortedRoutes.Routes[:i+1]
			break
		}
	}
	if pag.Offset != nil {
		if *pag.Offset <= len(sortedRoutes.Routes) {
			sortedRoutes.Routes = sortedRoutes.Routes[*pag.Offset:]
		}
	}
	if pag.Limit != nil {
		if *pag.Limit <= len(sortedRoutes.Routes) {
			sortedRoutes.Routes = sortedRoutes.Routes[:*pag.Limit]
		}
	}
	return
}

// sortedRoutesForEvent will return the list of sortedRoutes
// for event based on filters and sorting algorithms
func (rpS *RouteS) sortedRoutesForEvent(ctx *context.Context, tnt string, args *utils.CGREvent) (sortedRoutes SortedRoutesList, err error) {
	rPrfs, err := rpS.matchingRouteProfilesForEvent(ctx, tnt, args)
	if err != nil {
		return
	}
	prfCount := len(rPrfs) // if the option is not present return for all profiles
	var prfCountOpt *int
	if prfCountOpt, err = engine.GetIntPointerOpts(ctx, tnt, args.AsDataProvider(), nil, rpS.fltrS, rpS.cfg.RouteSCfg().Opts.ProfileCount,
		utils.OptsRoutesProfilesCount); err != nil && err != utils.ErrNotFound {
		// if the error is NOT_FOUND, it means that in opts or config, countProfiles field is not defined
		return
	}
	if prfCountOpt != nil && prfCount > *prfCountOpt { // it has the option and is smaller that the current number of profiles
		prfCount = *prfCountOpt
	}
	var extraOpts *optsGetRoutes
	if extraOpts, err = newOptsGetRoutes(ctx, args, rpS.fltrS, rpS.cfg.RouteSCfg().Opts); err != nil { // convert routes arguments into internal options used to limit data
		return
	}
	var startIdx, noSrtRoutes, initialOffset, maxItems int
	if extraOpts.paginator.Offset != nil { // save the offset in a varible to not double check if we have offset and is still not 0
		initialOffset = *extraOpts.paginator.Offset
		startIdx = initialOffset
	}
	if extraOpts.paginator.MaxItems != nil && extraOpts.paginator.Limit != nil {
		maxItems = *extraOpts.paginator.MaxItems
		if maxItems < *extraOpts.paginator.Limit+startIdx {
			return nil, fmt.Errorf("SERVER_ERROR: maximum number of items exceeded")
		}
	}
	sortedRoutes = make(SortedRoutesList, 0, prfCount)
	for _, rPrfl := range rPrfs {
		var prfPag utils.Paginator
		if extraOpts.paginator.Limit != nil { // we have a limit
			if noSrtRoutes >= *extraOpts.paginator.Limit { // the limit was reached
				break
			}
			if noSrtRoutes+len(rPrfl.Routes) > *extraOpts.paginator.Limit { // the limit will be reached in this profile
				limit := *extraOpts.paginator.Limit - noSrtRoutes // make it relative to current profile
				prfPag.Limit = &limit                             // add the limit to the paginator
			}
		}
		if startIdx > 0 { // we have offset
			if idx := startIdx - len(rPrfl.Routes); idx >= 0 { // we still have offset so try the next profile
				startIdx = idx
				continue
			}
			// we have offset but it's in the range of this profile
			offset := startIdx // store in a separate var so when startIdx is updated the prfPag.Offset remains the same
			startIdx = 0       // set it to 0 for the following loop
			prfPag.Offset = &offset
		}
		var sr *SortedRoutes
		if sr, err = rpS.sortedRoutesForProfile(ctx, tnt, rPrfl, args, prfPag, extraOpts); err != nil {
			return
		}
		if len(sr.Routes) != 0 {
			noSrtRoutes += len(sr.Routes)
			sortedRoutes = append(sortedRoutes, sr)
			if len(sortedRoutes) == prfCount { // the profile count was reached
				break
			}
		}
	}
	if maxItems != 0 && maxItems < len(sortedRoutes)+initialOffset {
		return nil, fmt.Errorf("SERVER_ERROR: maximum number of items exceeded")
	}
	if len(sortedRoutes) == 0 {
		err = utils.ErrNotFound
	}
	return
}
