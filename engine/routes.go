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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// Route defines routes related information used within a RouteProfile
type Route struct {
	ID              string // RouteID
	FilterIDs       []string
	AccountIDs      []string
	RatingPlanIDs   []string // used when computing price
	ResourceIDs     []string // queried in some strategies
	StatIDs         []string // queried in some strategies
	Weight          float64
	Blocker         bool // do not process further route after this one
	RouteParameters string

	cacheRoute     map[string]any // cache["*ratio"]=ratio
	lazyCheckRules []*FilterRule
}

// RouteProfile represents the configuration of a Route profile
type RouteProfile struct {
	Tenant             string
	ID                 string // LCR Profile ID
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval // Activation interval
	Sorting            string                    // Sorting strategy
	SortingParameters  []string
	Routes             []*Route
	Weight             float64
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

// RouteProfiles is a sortable list of RouteProfile
type RouteProfiles []*RouteProfile

// Sort is part of sort interface, sort based on Weight
func (lps RouteProfiles) Sort() {
	sort.Slice(lps, func(i, j int) bool { return lps[i].Weight > lps[j].Weight })
}

// NewRouteService initializes the Route Service
func NewRouteService(dm *DataManager,
	filterS *FilterS, cgrcfg *config.CGRConfig, connMgr *ConnManager) (rS *RouteService) {
	rS = &RouteService{
		dm:      dm,
		filterS: filterS,
		cgrcfg:  cgrcfg,
		connMgr: connMgr,
	}
	rS.sorter = NewRouteSortDispatcher(rS)
	return
}

// RouteService is the service computing route queries
type RouteService struct {
	dm      *DataManager
	filterS *FilterS
	cgrcfg  *config.CGRConfig
	sorter  RouteSortDispatcher
	connMgr *ConnManager
}

// Shutdown is called to shutdown the service
func (rpS *RouteService) Shutdown() {
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown initialized", utils.RouteS))
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown complete", utils.RouteS))
}

// matchingRouteProfilesForEvent returns ordered list of matching resources which are active by the time of the call
func (rpS *RouteService) matchingRouteProfilesForEvent(tnt string, ev *utils.CGREvent) (matchingRPrf []*RouteProfile, err error) {
	evNm := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
	}
	rPrfIDs, err := MatchingItemIDsForEvent(evNm,
		rpS.cgrcfg.RouteSCfg().StringIndexedFields,
		rpS.cgrcfg.RouteSCfg().PrefixIndexedFields,
		rpS.cgrcfg.RouteSCfg().SuffixIndexedFields,
		rpS.dm, utils.CacheRouteFilterIndexes, tnt,
		rpS.cgrcfg.RouteSCfg().IndexedSelects,
		rpS.cgrcfg.RouteSCfg().NestedFields,
	)
	if err != nil {
		return nil, err
	}
	matchingRPrf = make([]*RouteProfile, 0, len(rPrfIDs))
	for lpID := range rPrfIDs {
		rPrf, err := rpS.dm.GetRouteProfile(tnt, lpID, true, true, utils.NonTransactional)
		if err != nil {
			if err == utils.ErrNotFound {
				continue
			}
			return nil, err
		}
		if rPrf.ActivationInterval != nil && ev.Time != nil &&
			!rPrf.ActivationInterval.IsActiveAtTime(*ev.Time) { // not active
			continue
		}
		if pass, err := rpS.filterS.Pass(tnt, rPrf.FilterIDs,
			evNm); err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		matchingRPrf = append(matchingRPrf, rPrf)
	}
	if len(matchingRPrf) == 0 {
		return nil, utils.ErrNotFound
	}
	sort.Slice(matchingRPrf, func(i, j int) bool { return matchingRPrf[i].Weight > matchingRPrf[j].Weight })
	return
}

// costForEvent will compute cost out of accounts and rating plans for event
// returns map[string]any with cost and relevant matching information inside
func (rpS *RouteService) costForEvent(ev *utils.CGREvent,
	acntIDs, rpIDs []string) (costData map[string]any, err error) {
	costData = make(map[string]any)
	if err = ev.CheckMandatoryFields([]string{utils.AccountField,
		utils.Destination, utils.SetupTime}); err != nil {
		return
	}
	var acnt, subj, dst string
	if acnt, err = ev.FieldAsString(utils.AccountField); err != nil {
		return
	}
	if subj, err = ev.FieldAsString(utils.Subject); err != nil {
		if err != utils.ErrNotFound {
			return
		}
		subj = acnt
	}
	if dst, err = ev.FieldAsString(utils.Destination); err != nil {
		return
	}
	var sTime time.Time
	if sTime, err = ev.FieldAsTime(utils.SetupTime, rpS.cgrcfg.GeneralCfg().DefaultTimezone); err != nil {
		return
	}
	var usage time.Duration
	if usage, err = ev.FieldAsDuration(utils.Usage); err != nil {
		if err != utils.ErrNotFound {
			return
		}
		// in case usage is missing from event we decide to use 1 minute as default
		usage = time.Duration(1 * time.Minute)
		err = nil
	}
	var accountMaxUsage time.Duration
	var acntCost map[string]any
	var initialUsage time.Duration
	if len(acntIDs) != 0 {
		if err := rpS.connMgr.Call(context.TODO(), rpS.cgrcfg.RouteSCfg().RALsConns, utils.ResponderGetMaxSessionTimeOnAccounts,
			&utils.GetMaxSessionTimeOnAccountsArgs{
				Tenant:      ev.Tenant,
				Subject:     subj,
				Destination: dst,
				SetupTime:   sTime,
				Usage:       usage,
				AccountIDs:  acntIDs,
				APIOpts:     ev.APIOpts,
			}, &acntCost); err != nil {
			return nil, err
		}
		if ifaceMaxUsage, has := acntCost[utils.CapMaxUsage]; has {
			if accountMaxUsage, err = utils.IfaceAsDuration(ifaceMaxUsage); err != nil {
				return nil, err
			}
			if usage > accountMaxUsage {
				// remain usage needs to be covered by rating plans
				if len(rpIDs) == 0 {
					return nil, fmt.Errorf("no rating plans defined for remaining usage")
				}
				// update the setup time and the usage
				sTime = sTime.Add(accountMaxUsage)
				initialUsage = usage
				usage = usage - accountMaxUsage
			}
			for k, v := range acntCost { // update the costData with the infos from AccountS
				costData[k] = v
			}
		}
	}

	if accountMaxUsage == 0 || accountMaxUsage < initialUsage {
		var rpCost map[string]any
		if err := rpS.connMgr.Call(context.TODO(), rpS.cgrcfg.RouteSCfg().RALsConns, utils.ResponderGetCostOnRatingPlans,
			&utils.GetCostOnRatingPlansArgs{
				Tenant:        ev.Tenant,
				Account:       acnt,
				Subject:       subj,
				Destination:   dst,
				SetupTime:     sTime,
				Usage:         usage,
				RatingPlanIDs: rpIDs,
				APIOpts:       ev.APIOpts,
			}, &rpCost); err != nil {
			return nil, err
		}
		for k, v := range rpCost { // do not overwrite the return map
			costData[k] = v
		}
	}
	return
}

// statMetrics will query a list of statIDs and return composed metric values
// first metric found is always returned
func (rpS *RouteService) statMetrics(statIDs []string, tenant string) (stsMetric map[string]float64, err error) {
	stsMetric = make(map[string]float64)
	provStsMetrics := make(map[string][]float64)
	if len(rpS.cgrcfg.RouteSCfg().StatSConns) != 0 {
		for _, statID := range statIDs {
			var metrics map[string]float64
			if err = rpS.connMgr.Call(context.TODO(), rpS.cgrcfg.RouteSCfg().StatSConns, utils.StatSv1GetQueueFloatMetrics,
				&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: statID}}, &metrics); err != nil &&
				err.Error() != utils.ErrNotFound.Error() {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s getting statMetrics for stat : %s", utils.RouteS, err.Error(), statID))
			}
			for key, val := range metrics {
				//add value of metric in a slice in case that we get the same metric from different stat
				provStsMetrics[key] = append(provStsMetrics[key], val)
			}
		}
		for metric, slice := range provStsMetrics {
			sum := 0.0
			for _, val := range slice {
				sum += val
			}
			stsMetric[metric] = sum / float64(len(slice))
		}
	}
	return
}

// statMetricsForLoadDistribution will query a list of statIDs and return the sum of metrics
// first metric found is always returned
func (rpS *RouteService) statMetricsForLoadDistribution(statIDs []string, tenant string) (result float64, err error) {
	provStsMetrics := make(map[string][]float64)
	if len(rpS.cgrcfg.RouteSCfg().StatSConns) != 0 {
		for _, statID := range statIDs {
			// check if we get an ID in the following form (StatID:MetricID)
			statWithMetric := strings.Split(statID, utils.InInFieldSep)
			var metrics map[string]float64
			if err = rpS.connMgr.Call(context.TODO(),
				rpS.cgrcfg.RouteSCfg().StatSConns,
				utils.StatSv1GetQueueFloatMetrics,
				&utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{
						Tenant: tenant, ID: statWithMetric[0]}},
				&metrics); err != nil &&
				err.Error() != utils.ErrNotFound.Error() {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s getting statMetrics for stat : %s",
						utils.RouteS, err.Error(), statWithMetric[0]))
			}
			if len(statWithMetric) == 2 { // in case we have MetricID defined with StatID we consider only that metric
				// check if statQueue have metric defined
				metricVal, has := metrics[statWithMetric[1]]
				if !has {
					return 0, fmt.Errorf("<%s> error: %s metric %s for statID: %s",
						utils.RouteS, utils.ErrNotFound, statWithMetric[1], statWithMetric[0])
				}
				provStsMetrics[statWithMetric[1]] = append(provStsMetrics[statWithMetric[1]], metricVal)
			} else { // otherwise we consider all metrics
				for key, val := range metrics {
					//add value of metric in a slice in case that we get the same metric from different stat
					provStsMetrics[key] = append(provStsMetrics[key], val)
				}
			}
		}
		for _, slice := range provStsMetrics {
			sum := 0.0
			for _, val := range slice {
				sum += val
			}
			result += sum
		}
	}
	return
}

// resourceUsage returns sum of all resource usages out of list
func (rpS *RouteService) resourceUsage(resIDs []string, tenant string) (tUsage float64, err error) {
	if len(rpS.cgrcfg.RouteSCfg().ResourceSConns) != 0 {
		for _, resID := range resIDs {
			var res Resource
			if err = rpS.connMgr.Call(context.TODO(), rpS.cgrcfg.RouteSCfg().ResourceSConns, utils.ResourceSv1GetResource,
				&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: resID}}, &res); err != nil && err.Error() != utils.ErrNotFound.Error() {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s getting resource for ID : %s", utils.RouteS, err.Error(), resID))
				continue
			}
			tUsage += res.TotalUsage()
		}
	}
	return
}

func (rpS *RouteService) populateSortingData(ev *utils.CGREvent, route *Route,
	extraOpts *optsGetRoutes) (srtRoute *SortedRoute, pass bool, err error) {
	sortedSpl := &SortedRoute{
		RouteID: route.ID,
		SortingData: map[string]any{
			utils.Weight: route.Weight,
		},
		sortingDataF64: map[string]float64{
			utils.Weight: route.Weight,
		},
		RouteParameters: route.RouteParameters,
	}
	//calculate costData if we have fields
	if len(route.AccountIDs) != 0 || len(route.RatingPlanIDs) != 0 {
		costData, err := rpS.costForEvent(ev, route.AccountIDs, route.RatingPlanIDs)
		if err != nil {
			if extraOpts.ignoreErrors {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> ignoring route with ID: %s, err: %s",
						utils.RouteS, route.ID, err.Error()))
				return nil, false, nil
			}
			return nil, false, err
		} else if len(costData) == 0 {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> ignoring route with ID: %s, missing cost information",
					utils.RouteS, route.ID))
			return nil, false, nil
		} else {
			if extraOpts.maxCost != 0 &&
				costData[utils.Cost].(float64) > extraOpts.maxCost {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> ignoring route with ID: %s, err: %s",
						utils.RouteS, route.ID, utils.ErrMaxCostExceeded.Error()))
				return nil, false, nil
			}
			for k, v := range costData {
				sortedSpl.SortingData[k] = v
				sortedSpl.sortingDataF64[k], _ = v.(float64)
			}
		}
	}
	//calculate metrics
	//in case we have *load strategy we use statMetricsForLoadDistribution function to calculate the result
	if len(route.StatIDs) != 0 {
		if extraOpts.sortingStrategy == utils.MetaLoad {
			metricSum, err := rpS.statMetricsForLoadDistribution(route.StatIDs, ev.Tenant) //create metric map for route
			if err != nil {
				if extraOpts.ignoreErrors {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> ignoring route with ID: %s, err: %s",
							utils.RouteS, route.ID, err.Error()))
					return nil, false, nil
				}
				return nil, false, err
			}
			sortedSpl.SortingData[utils.Load] = metricSum
			sortedSpl.sortingDataF64[utils.Load] = metricSum
		} else {
			metricSupp, err := rpS.statMetrics(route.StatIDs, ev.Tenant) //create metric map for route
			if err != nil {
				if extraOpts.ignoreErrors {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> ignoring route with ID: %s, err: %s",
							utils.RouteS, route.ID, err.Error()))
					return nil, false, nil
				}
				return nil, false, err
			}
			//add metrics from statIDs in SortingData
			for key, val := range metricSupp {
				sortedSpl.SortingData[key] = val
				sortedSpl.sortingDataF64[key] = val
			}
			//check if the route have the metric from sortingParameters
			//in case that the metric don't exist
			//we use 10000000 for *pdd and -1 for others
			for _, metric := range extraOpts.sortingParameters {
				if _, hasMetric := metricSupp[metric]; !hasMetric {
					switch metric {
					default:
						sortedSpl.SortingData[metric] = -1.0
						sortedSpl.sortingDataF64[metric] = -1.0
					case utils.MetaPDD:
						sortedSpl.SortingData[metric] = math.MaxFloat64
						sortedSpl.sortingDataF64[metric] = math.MaxFloat64
					}
				}
			}
		}
	}
	//calculate resourceUsage
	if len(route.ResourceIDs) != 0 {
		resTotalUsage, err := rpS.resourceUsage(route.ResourceIDs, ev.Tenant)
		if err != nil {
			if extraOpts.ignoreErrors {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> ignoring route with ID: %s, err: %s",
						utils.RouteS, route.ID, err.Error()))
				return nil, false, nil
			}
			return nil, false, err
		}
		sortedSpl.SortingData[utils.ResourceUsage] = resTotalUsage
		sortedSpl.sortingDataF64[utils.ResourceUsage] = resTotalUsage
	}
	//filter the route
	if len(route.lazyCheckRules) != 0 {
		//construct the DP and pass it to filterS
		dynDP := newDynamicDP(rpS.cgrcfg.FilterSCfg().ResourceSConns, rpS.cgrcfg.FilterSCfg().StatSConns,
			rpS.cgrcfg.FilterSCfg().ApierSConns, rpS.cgrcfg.FilterSCfg().TrendSConns,
			ev.Tenant, utils.MapStorage{
				utils.MetaReq:  ev.Event,
				utils.MetaVars: sortedSpl.SortingData,
			})

		for _, rule := range route.lazyCheckRules { // verify the rules remaining from PartialPass
			if pass, err = rule.Pass(dynDP); err != nil {
				return nil, false, err
			} else if !pass {
				return nil, false, nil
			}
		}
	}
	return sortedSpl, true, nil
}

func newOptsGetRoutes(ev *utils.CGREvent, fS *FilterS, cfgOpts *config.RoutesOpts) (opts *optsGetRoutes, err error) {
	var ignoreErrors bool
	if ignoreErrors, err = utils.GetBoolOpts(ev, cfgOpts.IgnoreErrors, utils.OptsRoutesIgnoreErrors); err != nil {
		return
	}
	opts = &optsGetRoutes{
		ignoreErrors: ignoreErrors,
		paginator:    &utils.Paginator{},
	}
	var limit *int
	if limit, err = utils.GetIntPointerOpts(ev, cfgOpts.Limit, utils.OptsRoutesLimit); err != nil {
		return
	}
	if limit != nil {
		opts.paginator.Limit = limit
	}
	var offset *int
	if offset, err = utils.GetIntPointerOpts(ev, cfgOpts.Offset, utils.OptsRoutesOffset); err != nil {
		return
	}
	if offset != nil {
		opts.paginator.Offset = offset
	}
	maxCost := utils.GetInterfaceOpts(ev, cfgOpts.MaxCost, utils.OptsRoutesMaxCost)

	switch maxCost {
	case utils.EmptyString, nil:
	case utils.MetaEventCost: // dynamic cost needs to be calculated from event
		if err = ev.CheckMandatoryFields([]string{utils.AccountField,
			utils.Destination, utils.SetupTime, utils.Usage}); err != nil {
			return
		}
		cd, err := NewCallDescriptorFromCGREvent(ev, config.CgrConfig().GeneralCfg().DefaultTimezone)
		if err != nil {
			return nil, err
		}
		cc, err := cd.GetCost()
		if err != nil {
			return nil, err
		}
		opts.maxCost = cc.Cost
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
func (rpS *RouteService) V1GetRoutes(ctx *context.Context, args *utils.CGREvent, reply *SortedRoutesList) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = rpS.cgrcfg.GeneralCfg().DefaultTenant
	}
	if len(rpS.cgrcfg.RouteSCfg().AttributeSConns) != 0 {
		if args.APIOpts == nil {
			args.APIOpts = make(map[string]any)
		}
		args.APIOpts[utils.MetaSubsys] = utils.MetaRoutes
		args.APIOpts[utils.OptsContext] = utils.FirstNonEmpty(
			utils.GetStringOpts(args, rpS.cgrcfg.RouteSCfg().Opts.Context, utils.OptsContext),
			utils.MetaRoutes,
		)
		var rplyEv AttrSProcessEventReply
		if err := rpS.connMgr.Call(context.TODO(), rpS.cgrcfg.RouteSCfg().AttributeSConns,
			utils.AttributeSv1ProcessEvent, args, &rplyEv); err == nil && len(rplyEv.AlteredFields) != 0 {
			args = rplyEv.CGREvent
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrRouteS(err)
		}
	}
	var sSps SortedRoutesList
	if sSps, err = rpS.sortedRoutesForEvent(tnt, args); err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return
	}
	*reply = sSps
	return
}

// V1GetRouteProfilesForEvent returns the list of valid route profiles
func (rpS *RouteService) V1GetRouteProfilesForEvent(ctx *context.Context, args *utils.CGREvent, reply *[]*RouteProfile) (err error) {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = rpS.cgrcfg.GeneralCfg().DefaultTenant
	}
	sPs, err := rpS.matchingRouteProfilesForEvent(tnt, args)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = sPs
	return
}

// sortedRoutesForEvent will return the list of valid route IDs
// for event based on filters and sorting algorithms
func (rpS *RouteService) sortedRoutesForProfile(tnt string, rPrfl *RouteProfile, ev *utils.CGREvent,
	pag utils.Paginator, extraOpts *optsGetRoutes) (sortedRoutes *SortedRoutes, err error) {
	extraOpts.sortingParameters = rPrfl.SortingParameters // populate sortingParameters in extraOpts
	extraOpts.sortingStrategy = rPrfl.Sorting             // populate sortingStrategy in extraOpts
	//construct the DP and pass it to filterS
	nM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
	}
	passedRoutes := make(map[string]*Route)
	// apply filters for event
	for _, route := range rPrfl.Routes {
		pass, lazyCheckRules, err := rpS.filterS.LazyPass(tnt,
			route.FilterIDs, nM,
			[]string{utils.DynamicDataPrefix + utils.MetaReq,
				utils.DynamicDataPrefix + utils.MetaAccounts,
				utils.DynamicDataPrefix + utils.MetaResources,
				utils.DynamicDataPrefix + utils.MetaStats})
		if err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		route.lazyCheckRules = lazyCheckRules
		if prev, has := passedRoutes[route.ID]; has && prev.Weight >= route.Weight {
			continue
		}
		passedRoutes[route.ID] = route
	}

	if sortedRoutes, err = rpS.sorter.SortRoutes(rPrfl.ID, rPrfl.Sorting,
		passedRoutes, ev, extraOpts); err != nil {
		return nil, err
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

// sortedRoutesForEvent will return the list of valid route IDs
// for event based on filters and sorting algorithms
func (rpS *RouteService) sortedRoutesForEvent(tnt string, args *utils.CGREvent) (sortedRoutes SortedRoutesList, err error) {
	if _, has := args.Event[utils.Usage]; !has {
		args.Event[utils.Usage] = time.Minute // make sure we have default set for Usage
	}
	var rPrfs []*RouteProfile
	if rPrfs, err = rpS.matchingRouteProfilesForEvent(tnt, args); err != nil {
		return
	}
	prfCount := len(rPrfs) // if the option is not present return for all profiles
	var prfCountOpt *int
	if prfCountOpt, err = utils.GetIntPointerOpts(args, rpS.cgrcfg.RouteSCfg().Opts.ProfileCount,
		utils.OptsRoutesProfileCount); err != nil {
		return
	}
	if prfCountOpt != nil && prfCount > *prfCountOpt { // it has the option and is smaller that the current number of profiles
		prfCount = *prfCountOpt
	}
	var extraOpts *optsGetRoutes
	if extraOpts, err = newOptsGetRoutes(args, rpS.filterS, rpS.cgrcfg.RouteSCfg().Opts); err != nil { // convert routes arguments into internal options used to limit data
		return
	}

	var startIdx, noSrtRoutes int
	if extraOpts.paginator.Offset != nil { // save the offset in a variable to not double check if we have offset and is still not 0
		startIdx = *extraOpts.paginator.Offset
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
			// we have offset but is in the range of this profile
			offset := startIdx // store in a separate var so when startIdx is updated the prfPag.Offset remains the same
			startIdx = 0       // set it to 0 for the following loop
			prfPag.Offset = &offset
		}
		var sr *SortedRoutes
		if sr, err = rpS.sortedRoutesForProfile(tnt, rPrfl, args, prfPag, extraOpts); err != nil {
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
	if len(sortedRoutes) == 0 {
		err = utils.ErrNotFound
	}
	return
}

// V1GetRoutesList returns the list of valid routes
func (rpS *RouteService) V1GetRoutesList(ctx *context.Context, args *utils.CGREvent, reply *[]string) (err error) {
	sR := new(SortedRoutesList)
	if err = rpS.V1GetRoutes(ctx, args, sR); err != nil {
		return
	}
	*reply = sR.RoutesWithParams()
	return
}
