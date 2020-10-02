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
	"sort"
	"strconv"
	"strings"
	"time"

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

	cacheRoute     map[string]interface{} // cache["*ratio"]=ratio
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

	cache map[string]interface{}
}

// RouteProfileWithOpts is used in replicatorV1 for dispatcher
type RouteProfileWithOpts struct {
	*RouteProfile
	Opts map[string]interface{}
}

func (rp *RouteProfile) compileCacheParameters() error {
	if rp.Sorting == utils.MetaLoad {
		// construct the map for ratio
		ratioMap := make(map[string]int)
		// []string{"routeID:Ratio"}
		for _, splIDWithRatio := range rp.SortingParameters {
			splitted := strings.Split(splIDWithRatio, utils.CONCATENATED_KEY_SEP)
			ratioVal, err := strconv.Atoi(splitted[1])
			if err != nil {
				return err
			}
			ratioMap[splitted[0]] = ratioVal
		}
		// add the ratio for each route
		for _, route := range rp.Routes {
			route.cacheRoute = make(map[string]interface{})
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
	filterS *FilterS, cgrcfg *config.CGRConfig, connMgr *ConnManager) (rS *RouteService, err error) {
	rS = &RouteService{
		dm:      dm,
		filterS: filterS,
		cgrcfg:  cgrcfg,
		connMgr: connMgr,
	}
	if rS.sorter, err = NewRouteSortDispatcher(rS); err != nil {
		return nil, err
	}
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

// ListenAndServe will initialize the service
func (rpS *RouteService) ListenAndServe(exitChan chan bool) error {
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.RouteS))
	e := <-exitChan
	exitChan <- e // put back for the others listening for shutdown request
	return nil
}

// Shutdown is called to shutdown the service
func (rpS *RouteService) Shutdown() error {
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown initialized", utils.RouteS))
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown complete", utils.RouteS))
	return nil
}

// matchingRouteProfilesForEvent returns ordered list of matching resources which are active by the time of the call
func (rpS *RouteService) matchingRouteProfilesForEvent(ev *utils.CGREvent, singleResult bool) (matchingRPrf []*RouteProfile, err error) {
	evNm := utils.MapStorage{utils.MetaReq: ev.Event}
	rPrfIDs, err := MatchingItemIDsForEvent(evNm,
		rpS.cgrcfg.RouteSCfg().StringIndexedFields,
		rpS.cgrcfg.RouteSCfg().PrefixIndexedFields,
		rpS.cgrcfg.RouteSCfg().SuffixIndexedFields,
		rpS.dm, utils.CacheRouteFilterIndexes, ev.Tenant,
		rpS.cgrcfg.RouteSCfg().IndexedSelects,
		rpS.cgrcfg.RouteSCfg().NestedFields,
	)
	if err != nil {
		return nil, err
	}
	if singleResult {
		matchingRPrf = make([]*RouteProfile, 1)
	} else {
		matchingRPrf = make([]*RouteProfile, 0, len(rPrfIDs))
	}
	for lpID := range rPrfIDs {
		rPrf, err := rpS.dm.GetRouteProfile(ev.Tenant, lpID, true, true, utils.NonTransactional)
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
		if pass, err := rpS.filterS.Pass(ev.Tenant, rPrf.FilterIDs,
			evNm); err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		if singleResult {
			if matchingRPrf[0] == nil || matchingRPrf[0].Weight < rPrf.Weight {
				matchingRPrf[0] = rPrf
			}
		} else {
			matchingRPrf = append(matchingRPrf, rPrf)
		}
	}
	if singleResult {
		if matchingRPrf[0] == nil {
			return nil, utils.ErrNotFound
		}
	} else {
		if len(matchingRPrf) == 0 {
			return nil, utils.ErrNotFound
		}
		sort.Slice(matchingRPrf, func(i, j int) bool { return matchingRPrf[i].Weight > matchingRPrf[j].Weight })
	}
	return
}

// costForEvent will compute cost out of accounts and rating plans for event
// returns map[string]interface{} with cost and relevant matching information inside
func (rpS *RouteService) costForEvent(ev *utils.CGREvent,
	acntIDs, rpIDs []string) (costData map[string]interface{}, err error) {
	costData = make(map[string]interface{})
	if err = ev.CheckMandatoryFields([]string{utils.Account,
		utils.Destination, utils.SetupTime}); err != nil {
		return
	}
	var acnt, subj, dst string
	if acnt, err = ev.FieldAsString(utils.Account); err != nil {
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
	var acntCost map[string]interface{}
	var initialUsage time.Duration
	if len(acntIDs) != 0 {
		if err := rpS.connMgr.Call(rpS.cgrcfg.RouteSCfg().RALsConns, nil, utils.ResponderGetMaxSessionTimeOnAccounts,
			&utils.GetMaxSessionTimeOnAccountsArgs{
				Tenant:      ev.Tenant,
				Subject:     subj,
				Destination: dst,
				SetupTime:   sTime,
				Usage:       usage,
				AccountIDs:  acntIDs,
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
		var rpCost map[string]interface{}
		if err := rpS.connMgr.Call(rpS.cgrcfg.RouteSCfg().RALsConns, nil, utils.ResponderGetCostOnRatingPlans,
			&utils.GetCostOnRatingPlansArgs{
				Tenant:        ev.Tenant,
				Account:       acnt,
				Subject:       subj,
				Destination:   dst,
				SetupTime:     sTime,
				Usage:         usage,
				RatingPlanIDs: rpIDs,
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
			if err = rpS.connMgr.Call(rpS.cgrcfg.RouteSCfg().StatSConns, nil, utils.StatSv1GetQueueFloatMetrics,
				&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: statID}}, &metrics); err != nil &&
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
			if err = rpS.connMgr.Call(
				rpS.cgrcfg.RouteSCfg().StatSConns, nil,
				utils.StatSv1GetQueueFloatMetrics,
				&utils.TenantIDWithOpts{
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
			if err = rpS.connMgr.Call(rpS.cgrcfg.RouteSCfg().ResourceSConns, nil, utils.ResourceSv1GetResource,
				&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: resID}}, &res); err != nil && err.Error() != utils.ErrNotFound.Error() {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s getting resource for ID : %s", utils.RouteS, err.Error(), resID))
				continue
			}
			tUsage += res.totalUsage()
		}
	}
	return
}

func (rpS *RouteService) populateSortingData(ev *utils.CGREvent, route *Route,
	extraOpts *optsGetRoutes) (srtRoute *SortedRoute, pass bool, err error) {
	sortedSpl := &SortedRoute{
		RouteID: route.ID,
		SortingData: map[string]interface{}{
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
			}
		}
	}
	//calculate metrics
	//in case we have *load strategy we use statMetricsForLoadDistribution function to calculate the result
	if len(route.StatIDs) != 0 {
		if extraOpts.sortingStragety == utils.MetaLoad {
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
			}
			//check if the route have the metric from sortingParameters
			//in case that the metric don't exist
			//we use 10000000 for *pdd and -1 for others
			for _, metric := range extraOpts.sortingParameters {
				if _, hasMetric := metricSupp[metric]; !hasMetric {
					switch metric {
					default:
						sortedSpl.SortingData[metric] = -1.0
					case utils.MetaPDD:
						sortedSpl.SortingData[metric] = 10000000.0
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
	}
	//filter the route
	if len(route.lazyCheckRules) != 0 {
		//construct the DP and pass it to filterS
		dynDP := newDynamicDP(rpS.cgrcfg, rpS.connMgr, ev.Tenant, utils.MapStorage{
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

// sortedRoutesForEvent will return the list of valid route IDs
// for event based on filters and sorting algorithms
func (rpS *RouteService) sortedRoutesForEvent(args *ArgsGetRoutes) (sortedRoutes *SortedRoutes, err error) {
	if _, has := args.CGREvent.Event[utils.Usage]; !has {
		args.CGREvent.Event[utils.Usage] = time.Duration(time.Minute) // make sure we have default set for Usage
	}
	var rPrfs []*RouteProfile
	if rPrfs, err = rpS.matchingRouteProfilesForEvent(args.CGREvent, true); err != nil {
		return
	}
	rPrfl := rPrfs[0]
	extraOpts, err := args.asOptsGetRoutes() // convert routes arguments into internal options used to limit data
	if err != nil {
		return nil, err
	}
	extraOpts.sortingParameters = rPrfl.SortingParameters // populate sortingParameters in extraOpts
	extraOpts.sortingStragety = rPrfl.Sorting             // populate sortingStrategy in extraOpts

	//construct the DP and pass it to filterS
	nM := utils.MapStorage{utils.MetaReq: args.CGREvent.Event}
	routeNew := make([]*Route, 0)
	// apply filters for event

	for _, route := range rPrfl.Routes {
		pass, lazyCheckRules, err := rpS.filterS.LazyPass(args.CGREvent.Tenant, route.FilterIDs,
			nM, []string{utils.DynamicDataPrefix + utils.MetaReq, utils.DynamicDataPrefix + utils.MetaAccounts,
				utils.DynamicDataPrefix + utils.MetaResources, utils.DynamicDataPrefix + utils.MetaStats})
		if err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		route.lazyCheckRules = lazyCheckRules
		routeNew = append(routeNew, route)
	}

	sortedRoutes, err = rpS.sorter.SortRoutes(rPrfl.ID, rPrfl.Sorting,
		routeNew, args.CGREvent, extraOpts)
	if err != nil {
		return nil, err
	}
	if args.Paginator.Offset != nil {
		if *args.Paginator.Offset <= len(sortedRoutes.SortedRoutes) {
			sortedRoutes.SortedRoutes = sortedRoutes.SortedRoutes[*args.Paginator.Offset:]
		}
	}
	if args.Paginator.Limit != nil {
		if *args.Paginator.Limit <= len(sortedRoutes.SortedRoutes) {
			sortedRoutes.SortedRoutes = sortedRoutes.SortedRoutes[:*args.Paginator.Limit]
		}
	}
	sortedRoutes.Count = len(sortedRoutes.SortedRoutes)
	return
}

// ArgsGetRoutes the argument for GetRoutes API
type ArgsGetRoutes struct {
	IgnoreErrors bool
	MaxCost      string // toDo: try with interface{} here
	*utils.CGREventWithOpts
	utils.Paginator
	clnb bool //rpcclonable
}

// SetCloneable sets if the args should be clonned on internal connections
func (attr *ArgsGetRoutes) SetCloneable(rpcCloneable bool) {
	attr.clnb = rpcCloneable
}

// RPCClone implements rpcclient.RPCCloner interface
func (attr *ArgsGetRoutes) RPCClone() (interface{}, error) {
	if !attr.clnb {
		return attr, nil
	}
	return attr.Clone(), nil
}

// Clone creates a clone of the object
func (attr *ArgsGetRoutes) Clone() *ArgsGetRoutes {
	return &ArgsGetRoutes{
		IgnoreErrors:     attr.IgnoreErrors,
		MaxCost:          attr.MaxCost,
		Paginator:        attr.Paginator.Clone(),
		CGREventWithOpts: attr.CGREventWithOpts.Clone(),
	}
}

func (args *ArgsGetRoutes) asOptsGetRoutes() (opts *optsGetRoutes, err error) {
	opts = &optsGetRoutes{ignoreErrors: args.IgnoreErrors}
	if args.MaxCost == utils.MetaEventCost { // dynamic cost needs to be calculated from event
		if err = args.CGREvent.CheckMandatoryFields([]string{utils.Account,
			utils.Destination, utils.SetupTime, utils.Usage}); err != nil {
			return
		}
		cd, err := NewCallDescriptorFromCGREvent(args.CGREvent,
			config.CgrConfig().GeneralCfg().DefaultTimezone)
		if err != nil {
			return nil, err
		}
		cc, err := cd.GetCost()
		if err != nil {
			return nil, err
		}
		opts.maxCost = cc.Cost
	} else if args.MaxCost != "" {
		if opts.maxCost, err = strconv.ParseFloat(args.MaxCost,
			64); err != nil {
			return nil, err
		}
	}
	return
}

type optsGetRoutes struct {
	ignoreErrors      bool
	maxCost           float64
	sortingParameters []string //used for QOS strategy
	sortingStragety   string
}

// V1GetRoutes returns the list of valid routes
func (rpS *RouteService) V1GetRoutes(args *ArgsGetRoutes, reply *SortedRoutes) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if missing := utils.MissingStructFields(args.CGREvent, []string{utils.Tenant, utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.CGREvent.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if len(rpS.cgrcfg.RouteSCfg().AttributeSConns) != 0 {
		if args.Opts == nil {
			args.Opts = make(map[string]interface{})
		}
		args.Opts[utils.Subsys] = utils.MetaRoutes
		var processRuns *int
		if val, has := args.Opts[utils.OptsAttributesProcessRuns]; has {
			if v, err := utils.IfaceAsTInt64(val); err == nil {
				processRuns = utils.IntPointer(int(v))
			}
		}
		attrArgs := &AttrArgsProcessEvent{
			Context: utils.StringPointer(utils.FirstNonEmpty(
				utils.IfaceAsString(args.CGREvent.Event[utils.OptsContext]),
				utils.MetaRoutes)),
			CGREventWithOpts: args.CGREventWithOpts,
			ProcessRuns:      processRuns,
		}
		var rplyEv AttrSProcessEventReply
		if err := rpS.connMgr.Call(rpS.cgrcfg.RouteSCfg().AttributeSConns, nil,
			utils.AttributeSv1ProcessEvent, attrArgs, &rplyEv); err == nil && len(rplyEv.AlteredFields) != 0 {
			args.CGREvent = rplyEv.CGREvent
			args.Opts = rplyEv.Opts
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrAttributeS(err)
		}
	}
	sSps, err := rpS.sortedRoutesForEvent(args)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *sSps
	return
}

// V1GetRouteProfilesForEvent returns the list of valid route profiles
func (rpS *RouteService) V1GetRouteProfilesForEvent(args *utils.CGREventWithOpts, reply *[]*RouteProfile) (err error) {
	if missing := utils.MissingStructFields(args.CGREvent, []string{utils.Tenant, utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.CGREvent.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	sPs, err := rpS.matchingRouteProfilesForEvent(args.CGREvent, false)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = sPs
	return
}
