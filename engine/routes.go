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
	"strconv"
	"strings"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// Route defines routes related information used within a RouteProfile
type Route struct {
	ID              string // RouteID
	FilterIDs       []string
	AccountIDs      []string
	RateProfileIDs  []string // used when computing price
	ResourceIDs     []string // queried in some strategies
	StatIDs         []string // queried in some strategies
	Weights         utils.DynamicWeights
	Blockers        utils.DynamicBlockers // do not process further route after this one
	RouteParameters string

	cacheRoute map[string]any // cache["*ratio"]=ratio
}

// RouteProfile represents the configuration of a Route profile
type RouteProfile struct {
	Tenant            string
	ID                string // LCR Profile ID
	FilterIDs         []string
	Weights           utils.DynamicWeights
	Blockers          utils.DynamicBlockers
	Sorting           string // Sorting strategy
	SortingParameters []string
	Routes            []*Route
}

// RouteProfileWithAPIOpts is used in replicatorV1 for dispatcher
type RouteProfileWithAPIOpts struct {
	*RouteProfile
	APIOpts map[string]any
}

func (rp *RouteProfile) compileCacheParameters() error {
	if rp.Sorting == utils.MetaLoad {
		// construct the map for ratio
		ratioMap := make(map[string]int)
		// []string{"routeID:Ratio"}
		for _, splIDWithRatio := range rp.SortingParameters {
			splitted := strings.Split(splIDWithRatio, utils.ConcatenatedKeySep)
			ratioVal, err := strconv.Atoi(splitted[1])
			if err != nil {
				return err
			}
			ratioMap[splitted[0]] = ratioVal
		}
		// add the ratio for each route
		for _, route := range rp.Routes {
			route.cacheRoute = make(map[string]any)
			if ratioRoute, has := ratioMap[route.ID]; !has { // in case that ratio isn't defined for specific routes check for default
				if ratioDefault, has := ratioMap[utils.MetaDefault]; !has { // in case that *default ratio isn't defined take it from config
					route.cacheRoute[utils.MetaRatio] = config.CgrConfig().RouteSCfg().DefaultRatio
				} else {
					route.cacheRoute[utils.MetaRatio] = ratioDefault
				}
			} else {
				route.cacheRoute[utils.MetaRatio] = ratioRoute
			}
		}
	}
	return nil
}

// Compile is a wrapper for convenience setting up the RouteProfile
func (rp *RouteProfile) Compile() error {
	return rp.compileCacheParameters()
}

// TenantID returns unique identifier of the LCRProfile in a multi-tenant environment
func (rp *RouteProfile) TenantID() string {
	return utils.ConcatenatedKey(rp.Tenant, rp.ID)
}

// NewRouteService initializes the Route Service
func NewRouteService(dm *DataManager,
	filterS *FilterS, cfg *config.CGRConfig, connMgr *ConnManager) (rS *RouteS) {
	rS = &RouteS{
		dm:      dm,
		fltrS:   filterS,
		cfg:     cfg,
		connMgr: connMgr,
		sorter: RouteSortDispatcher{
			utils.MetaWeight: NewWeightSorter(cfg),
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
	dm      *DataManager
	fltrS   *FilterS
	cfg     *config.CGRConfig
	sorter  RouteSortDispatcher
	connMgr *ConnManager
}

// matchingRouteProfilesForEvent returns ordered list of matching resources which are active by the time of the call
func (rpS *RouteS) matchingRouteProfilesForEvent(ctx *context.Context, tnt string, ev *utils.CGREvent) (matchingRPrf RouteProfilesWithWeight, err error) {
	evNm := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
	}
	rPrfIDs, err := MatchingItemIDsForEvent(ctx, evNm,
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
	matchingRPrf = make(RouteProfilesWithWeight, 0, len(rPrfIDs))
	for lpID := range rPrfIDs {
		var rPrf *RouteProfile
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
		var weight float64
		if weight, err = WeightFromDynamics(ctx, rPrf.Weights,
			rpS.fltrS, ev.Tenant, evNm); err != nil {
			return
		}
		matchingRPrf = append(matchingRPrf, &RouteProfileWithWeight{RouteProfile: rPrf, Weight: weight})
	}
	if len(matchingRPrf) == 0 {
		return nil, utils.ErrNotFound
	}
	matchingRPrf.Sort()
	for i, rp := range matchingRPrf {
		var blocker bool
		if blocker, err = BlockerFromDynamics(ctx, rp.Blockers, rpS.fltrS, ev.Tenant, evNm); err != nil {
			return
		}
		if blocker {
			matchingRPrf = matchingRPrf[0 : i+1]
			break
		}
	}
	return
}

func newOptsGetRoutes(ctx *context.Context, ev *utils.CGREvent, fS *FilterS, cfgOpts *config.RoutesOpts) (opts *optsGetRoutes, err error) {
	var ignoreErrors bool
	evNM := ev.AsDataProvider()
	if ignoreErrors, err = GetBoolOpts(ctx, ev.Tenant, evNM, fS, cfgOpts.IgnoreErrors,
		utils.OptsRoutesIgnoreErrors); err != nil {
		return
	}
	opts = &optsGetRoutes{
		ignoreErrors: ignoreErrors,
		paginator:    &utils.Paginator{},
	}
	var limit *int
	if limit, err = GetIntPointerOpts(ctx, ev.Tenant, evNM, fS, cfgOpts.Limit,
		utils.OptsRoutesLimit); err != nil {
		return
	} else {
		opts.paginator.Limit = limit
	}
	var offset *int
	if offset, err = GetIntPointerOpts(ctx, ev.Tenant, evNM, fS, cfgOpts.Offset,
		utils.OptsRoutesOffset); err != nil {
		return
	} else {
		opts.paginator.Offset = offset
	}
	var maxItems *int
	if maxItems, err = GetIntPointerOpts(ctx, ev.Tenant, evNM, fS, cfgOpts.MaxItems,
		utils.OptsRoutesMaxItems); err != nil {
		return
	} else {
		opts.paginator.MaxItems = maxItems
	}

	var maxCost any
	if maxCost, err = GetInterfaceOpts(ctx, ev.Tenant, evNM, fS, cfgOpts.MaxCost, config.RoutesMaxCostDftOpt,
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

// V1GetRoutes returns the list of valid routes
func (rpS *RouteS) V1GetRoutes(ctx *context.Context, args *utils.CGREvent, reply *SortedRoutesList) (err error) {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = rpS.cfg.GeneralCfg().DefaultTenant
	}
	if args.APIOpts == nil {
		args.APIOpts = make(map[string]any)
	}
	if len(rpS.cfg.RouteSCfg().AttributeSConns) != 0 {
		args.APIOpts[utils.MetaSubsys] = utils.MetaRoutes
		var context string
		if context, err = GetStringOpts(ctx, tnt, args.AsDataProvider(), rpS.fltrS, rpS.cfg.RouteSCfg().Opts.Context,
			utils.OptsContext); err != nil {
			return
		}
		args.APIOpts[utils.OptsContext] = context
		var rplyEv AttrSProcessEventReply
		if err := rpS.connMgr.Call(ctx, rpS.cfg.RouteSCfg().AttributeSConns,
			utils.AttributeSv1ProcessEvent, args, &rplyEv); err == nil && len(rplyEv.AlteredFields) != 0 {
			args = rplyEv.CGREvent
			args.APIOpts = rplyEv.APIOpts
		} else if err = utils.CastRPCErr(err); err != utils.ErrNotFound {
			return utils.NewErrRouteS(err)
		}
	}
	var sSps SortedRoutesList
	if sSps, err = rpS.sortedRoutesForEvent(ctx, tnt, args); err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return
	}
	*reply = sSps
	return
}

// V1GetRouteProfilesForEvent returns the list of valid route profiles
func (rpS *RouteS) V1GetRouteProfilesForEvent(ctx *context.Context, args *utils.CGREvent, reply *[]*RouteProfile) (_ error) {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = rpS.cfg.GeneralCfg().DefaultTenant
	}
	sPs, err := rpS.matchingRouteProfilesForEvent(ctx, tnt, args)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = make([]*RouteProfile, len(sPs))
	for i, sP := range sPs {
		(*reply)[i] = sP.RouteProfile
	}
	return
}

var lazyRouteFltrPrfxs = []string{utils.DynamicDataPrefix + utils.MetaReq,
	utils.DynamicDataPrefix + utils.MetaAccounts,
	utils.DynamicDataPrefix + utils.MetaResources,
	utils.DynamicDataPrefix + utils.MetaStats}

// sortedRoutesForEvent will return the list of valid route IDs
// for event based on filters and sorting algorithms
func (rpS *RouteS) sortedRoutesForProfile(ctx *context.Context, tnt string, rPrfl *RouteProfile, ev *utils.CGREvent,
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
		var lazyCheckRules []*FilterRule
		if pass, lazyCheckRules, err = rpS.fltrS.LazyPass(ctx, tnt,
			route.FilterIDs, nM, lazyRouteFltrPrfxs); err != nil {
			return
		} else if !pass {
			continue
		}
		var weight float64
		if weight, err = WeightFromDynamics(ctx, route.Weights,
			rpS.fltrS, ev.Tenant, nM); err != nil {
			return
		}
		if prev, has := passedRoutes[route.ID]; !has || prev.Weight < weight {
			var blocker bool
			if blocker, err = BlockerFromDynamics(ctx, route.Blockers, rpS.fltrS, tnt, nM); err != nil {
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
	var rPrfs RouteProfilesWithWeight
	if rPrfs, err = rpS.matchingRouteProfilesForEvent(ctx, tnt, args); err != nil {
		return
	}
	prfCount := len(rPrfs) // if the option is not present return for all profiles
	var prfCountOpt *int
	if prfCountOpt, err = GetIntPointerOpts(ctx, tnt, args.AsDataProvider(), rpS.fltrS, rpS.cfg.RouteSCfg().Opts.ProfileCount,
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
		if sr, err = rpS.sortedRoutesForProfile(ctx, tnt, rPrfl.RouteProfile, args, prfPag, extraOpts); err != nil {
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

// V1GetRoutesList returns the list of valid routes
func (rpS *RouteS) V1GetRoutesList(ctx *context.Context, args *utils.CGREvent, reply *[]string) (err error) {
	sR := new(SortedRoutesList)
	if err = rpS.V1GetRoutes(ctx, args, sR); err != nil {
		return
	}
	*reply = sR.RoutesWithParams()
	return
}

func (rp *RouteProfile) Set(path []string, val any, newBranch bool) (err error) {
	switch len(path) {
	default:
		return utils.ErrWrongPath
	case 1:
		switch path[0] {
		default:
			return utils.ErrWrongPath
		case utils.Tenant:
			rp.Tenant = utils.IfaceAsString(val)
		case utils.ID:
			rp.ID = utils.IfaceAsString(val)
		case utils.FilterIDs:
			var valA []string
			valA, err = utils.IfaceAsStringSlice(val)
			rp.FilterIDs = append(rp.FilterIDs, valA...)
		case utils.SortingParameters:
			var valA []string
			valA, err = utils.IfaceAsStringSlice(val)
			rp.SortingParameters = append(rp.SortingParameters, valA...)
		case utils.Sorting:
			if valStr := utils.IfaceAsString(val); len(valStr) != 0 {
				rp.Sorting = valStr
			}
		case utils.Weights:
			if val != utils.EmptyString {
				rp.Weights, err = utils.NewDynamicWeightsFromString(utils.IfaceAsString(val), utils.InfieldSep, utils.ANDSep)
			}
		case utils.Blockers:
			if val != utils.EmptyString {
				rp.Blockers, err = utils.NewDynamicBlockersFromString(utils.IfaceAsString(val), utils.InfieldSep, utils.ANDSep)
			}
		}
	case 2:
		if val == utils.EmptyString {
			return
		}
		if path[0] != utils.Routes {
			return utils.ErrWrongPath
		}
		if len(rp.Routes) == 0 || newBranch {
			rp.Routes = append(rp.Routes, new(Route))
		}
		rt := rp.Routes[len(rp.Routes)-1]
		switch path[1] {
		case utils.ID:
			rt.ID = utils.IfaceAsString(val)
		case utils.FilterIDs:
			var valA []string
			valA, err = utils.IfaceAsStringSlice(val)
			rt.FilterIDs = append(rt.FilterIDs, valA...)
		case utils.AccountIDs:
			var valA []string
			valA, err = utils.IfaceAsStringSlice(val)
			rt.AccountIDs = append(rt.AccountIDs, valA...)
		case utils.RateProfileIDs:
			var valA []string
			valA, err = utils.IfaceAsStringSlice(val)
			rt.RateProfileIDs = append(rt.RateProfileIDs, valA...)
		case utils.ResourceIDs:
			var valA []string
			valA, err = utils.IfaceAsStringSlice(val)
			rt.ResourceIDs = append(rt.ResourceIDs, valA...)
		case utils.StatIDs:
			var valA []string
			valA, err = utils.IfaceAsStringSlice(val)
			rt.StatIDs = append(rt.StatIDs, valA...)
		case utils.Weights:
			if val != utils.EmptyString {
				rt.Weights, err = utils.NewDynamicWeightsFromString(utils.IfaceAsString(val), utils.InfieldSep, utils.ANDSep)
			}
		case utils.Blockers:
			if val != utils.EmptyString {
				rt.Blockers, err = utils.NewDynamicBlockersFromString(utils.IfaceAsString(val), utils.InfieldSep, utils.ANDSep)
			}
		case utils.RouteParameters:
			rt.RouteParameters = utils.IfaceAsString(val)
		default:
			return utils.ErrWrongPath
		}
	}
	return
}

func (rp *RouteProfile) Merge(v2 any) {
	vi := v2.(*RouteProfile)
	if len(vi.Tenant) != 0 {
		rp.Tenant = vi.Tenant
	}
	if len(vi.ID) != 0 {
		rp.ID = vi.ID
	}
	rp.FilterIDs = append(rp.FilterIDs, vi.FilterIDs...)
	rp.SortingParameters = append(rp.SortingParameters, vi.SortingParameters...)
	var equal bool
	for _, routeV2 := range vi.Routes {
		for _, route := range rp.Routes {
			if route.ID == routeV2.ID {
				route.Merge(routeV2)
				equal = true
				break
			}
		}
		if !equal {
			rp.Routes = append(rp.Routes, routeV2)
		}
		equal = false
	}
	rp.Weights = append(rp.Weights, vi.Weights...)
	rp.Blockers = append(rp.Blockers, vi.Blockers...)
	if len(vi.Sorting) != 0 {
		rp.Sorting = vi.Sorting
	}
}

func (route *Route) Merge(v2 *Route) {
	if len(v2.ID) != 0 {
		route.ID = v2.ID
	}
	if len(v2.RouteParameters) != 0 {
		route.RouteParameters = v2.RouteParameters
	}
	route.Weights = append(route.Weights, v2.Weights...)
	route.Blockers = append(route.Blockers, v2.Blockers...)
	route.FilterIDs = append(route.FilterIDs, v2.FilterIDs...)
	route.AccountIDs = append(route.AccountIDs, v2.AccountIDs...)
	route.RateProfileIDs = append(route.RateProfileIDs, v2.RateProfileIDs...)
	route.ResourceIDs = append(route.ResourceIDs, v2.ResourceIDs...)
	route.StatIDs = append(route.StatIDs, v2.StatIDs...)
}

func (rp *RouteProfile) String() string { return utils.ToJSON(rp) }
func (rp *RouteProfile) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = rp.FieldAsInterface(fldPath); err != nil {
		return
	}
	return utils.IfaceAsString(val), nil
}
func (rp *RouteProfile) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) == 1 {
		switch fldPath[0] {
		default:
			fld, idx := utils.GetPathIndex(fldPath[0])
			if idx != nil {
				switch fld {
				case utils.SortingParameters:
					if *idx < len(rp.SortingParameters) {
						return rp.SortingParameters[*idx], nil
					}
				case utils.FilterIDs:
					if *idx < len(rp.FilterIDs) {
						return rp.FilterIDs[*idx], nil
					}
				case utils.Routes:
					if *idx < len(rp.Routes) {
						return rp.Routes[*idx], nil
					}
				}
			}
			return nil, utils.ErrNotFound
		case utils.Tenant:
			return rp.Tenant, nil
		case utils.ID:
			return rp.ID, nil
		case utils.FilterIDs:
			return rp.FilterIDs, nil
		case utils.Weights:
			return rp.Weights.String(utils.InfieldSep, utils.ANDSep), nil
		case utils.SortingParameters:
			return rp.SortingParameters, nil
		case utils.Sorting:
			return rp.Sorting, nil
		case utils.Blockers:
			return rp.Blockers.String(utils.InfieldSep, utils.ANDSep), nil
		case utils.Routes:
			return rp.Routes, nil
		}
	}
	if len(fldPath) == 0 {
		return nil, utils.ErrNotFound
	}
	fld, idx := utils.GetPathIndex(fldPath[0])
	if fld != utils.Routes ||
		idx == nil {
		return nil, utils.ErrNotFound
	}
	if *idx >= len(rp.Routes) {
		return nil, utils.ErrNotFound
	}
	return rp.Routes[*idx].FieldAsInterface(fldPath[1:])
}

func (rt *Route) String() string { return utils.ToJSON(rt) }
func (rt *Route) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = rt.FieldAsInterface(fldPath); err != nil {
		return
	}
	return utils.IfaceAsString(val), nil
}
func (rt *Route) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) != 1 {
		return nil, utils.ErrNotFound
	}
	switch fldPath[0] {
	default:
		fld, idx := utils.GetPathIndex(fldPath[0])
		if idx != nil {
			switch fld {
			case utils.AccountIDs:
				if *idx < len(rt.AccountIDs) {
					return rt.AccountIDs[*idx], nil
				}
			case utils.FilterIDs:
				if *idx < len(rt.FilterIDs) {
					return rt.FilterIDs[*idx], nil
				}
			case utils.RateProfileIDs:
				if *idx < len(rt.RateProfileIDs) {
					return rt.RateProfileIDs[*idx], nil
				}
			case utils.ResourceIDs:
				if *idx < len(rt.ResourceIDs) {
					return rt.ResourceIDs[*idx], nil
				}
			case utils.StatIDs:
				if *idx < len(rt.StatIDs) {
					return rt.StatIDs[*idx], nil
				}
			}
		}
		return nil, utils.ErrNotFound
	case utils.ID:
		return rt.ID, nil
	case utils.FilterIDs:
		return rt.FilterIDs, nil
	case utils.AccountIDs:
		return rt.AccountIDs, nil
	case utils.RateProfileIDs:
		return rt.RateProfileIDs, nil
	case utils.ResourceIDs:
		return rt.ResourceIDs, nil
	case utils.StatIDs:
		return rt.StatIDs, nil
	case utils.Weights:
		return rt.Weights.String(utils.InfieldSep, utils.ANDSep), nil
	case utils.Blockers:
		return rt.Blockers.String(utils.InfieldSep, utils.ANDSep), nil
	case utils.RouteParameters:
		return rt.RouteParameters, nil
	}
}
