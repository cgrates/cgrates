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

// Supplier defines supplier related information used within a SupplierProfile
type Supplier struct {
	ID                 string // SupplierID
	FilterIDs          []string
	AccountIDs         []string
	RatingPlanIDs      []string // used when computing price
	ResourceIDs        []string // queried in some strategies
	StatIDs            []string // queried in some strategies
	Weight             float64
	Blocker            bool // do not process further supplier after this one
	SupplierParameters string

	cacheSupplier map[string]interface{} // cache["*ratio"]=ratio
}

// SupplierProfile represents the configuration of a Supplier profile
type SupplierProfile struct {
	Tenant             string
	ID                 string // LCR Profile ID
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval // Activation interval
	Sorting            string                    // Sorting strategy
	SortingParameters  []string
	Suppliers          []*Supplier
	Weight             float64

	cache map[string]interface{}
}

func (sp *SupplierProfile) compileCacheParameters() error {
	if sp.Sorting == utils.MetaLoad {
		// construct the map for ratio
		ratioMap := make(map[string]int)
		// []string{"supplierID:Ratio"}
		for _, splIDWithRatio := range sp.SortingParameters {
			splitted := strings.Split(splIDWithRatio, utils.CONCATENATED_KEY_SEP)
			ratioVal, err := strconv.Atoi(splitted[1])
			if err != nil {
				return err
			}
			ratioMap[splitted[0]] = ratioVal
		}
		// add the ratio for each supplier
		for _, supplier := range sp.Suppliers {
			supplier.cacheSupplier = make(map[string]interface{})
			if ratioSupplier, has := ratioMap[supplier.ID]; !has { // in case that ratio isn't defined for specific suppliers check for default
				if ratioDefault, has := ratioMap[utils.MetaDefault]; !has { // in case that *default ratio isn't defined take it from config
					supplier.cacheSupplier[utils.MetaRatio] = config.CgrConfig().SupplierSCfg().DefaultRatio
				} else {
					supplier.cacheSupplier[utils.MetaRatio] = ratioDefault
				}
			} else {
				supplier.cacheSupplier[utils.MetaRatio] = ratioSupplier
			}
		}
	}
	return nil
}

// Compile is a wrapper for convenience setting up the SupplierProfile
func (sp *SupplierProfile) Compile() error {
	return sp.compileCacheParameters()
}

// TenantID returns unique identifier of the LCRProfile in a multi-tenant environment
func (rp *SupplierProfile) TenantID() string {
	return utils.ConcatenatedKey(rp.Tenant, rp.ID)
}

// SupplierProfiles is a sortable list of SupplierProfile
type SupplierProfiles []*SupplierProfile

// Sort is part of sort interface, sort based on Weight
func (lps SupplierProfiles) Sort() {
	sort.Slice(lps, func(i, j int) bool { return lps[i].Weight > lps[j].Weight })
}

// NewSupplierService initializes the Supplier Service
func NewSupplierService(dm *DataManager,
	filterS *FilterS, cgrcfg *config.CGRConfig, connMgr *ConnManager) (spS *SupplierService, err error) {
	spS = &SupplierService{
		dm:      dm,
		filterS: filterS,
		cgrcfg:  cgrcfg,
		connMgr: connMgr,
	}
	if spS.sorter, err = NewSupplierSortDispatcher(spS); err != nil {
		return nil, err
	}
	return
}

// SupplierService is the service computing Supplier queries
type SupplierService struct {
	dm      *DataManager
	filterS *FilterS
	cgrcfg  *config.CGRConfig
	sorter  SupplierSortDispatcher
	connMgr *ConnManager
}

// ListenAndServe will initialize the service
func (spS *SupplierService) ListenAndServe(exitChan chan bool) error {
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.SupplierS))
	e := <-exitChan
	exitChan <- e // put back for the others listening for shutdown request
	return nil
}

// Shutdown is called to shutdown the service
func (spS *SupplierService) Shutdown() error {
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown initialized", utils.SupplierS))
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown complete", utils.SupplierS))
	return nil
}

// matchingSupplierProfilesForEvent returns ordered list of matching resources which are active by the time of the call
func (spS *SupplierService) matchingSupplierProfilesForEvent(ev *utils.CGREvent, singleResult bool) (matchingSLP []*SupplierProfile, err error) {
	sPrflIDs, err := MatchingItemIDsForEvent(ev.Event,
		spS.cgrcfg.SupplierSCfg().StringIndexedFields,
		spS.cgrcfg.SupplierSCfg().PrefixIndexedFields,
		spS.dm, utils.CacheSupplierFilterIndexes, ev.Tenant,
		spS.cgrcfg.SupplierSCfg().IndexedSelects,
		spS.cgrcfg.SupplierSCfg().NestedFields,
	)
	if err != nil {
		return nil, err
	}
	if singleResult {
		matchingSLP = make([]*SupplierProfile, 1)
	}
	evNm := utils.MapStorage{utils.MetaReq: ev.Event}
	for lpID := range sPrflIDs {
		splPrfl, err := spS.dm.GetSupplierProfile(ev.Tenant, lpID, true, true, utils.NonTransactional)
		if err != nil {
			if err == utils.ErrNotFound {
				continue
			}
			return nil, err
		}
		if splPrfl.ActivationInterval != nil && ev.Time != nil &&
			!splPrfl.ActivationInterval.IsActiveAtTime(*ev.Time) { // not active
			continue
		}
		if pass, err := spS.filterS.Pass(ev.Tenant, splPrfl.FilterIDs,
			evNm); err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		if singleResult {
			if matchingSLP[0] == nil || matchingSLP[0].Weight < splPrfl.Weight {
				matchingSLP[0] = splPrfl
			}
		} else {
			matchingSLP = append(matchingSLP, splPrfl)
		}
	}
	if singleResult {
		if matchingSLP[0] == nil {
			return nil, utils.ErrNotFound
		}
	} else {
		if len(matchingSLP) == 0 {
			return nil, utils.ErrNotFound
		}
		sort.Slice(matchingSLP, func(i, j int) bool { return matchingSLP[i].Weight > matchingSLP[j].Weight })
	}
	return
}

// costForEvent will compute cost out of accounts and rating plans for event
// returns map[string]interface{} with cost and relevant matching information inside
func (spS *SupplierService) costForEvent(ev *utils.CGREvent,
	acntIDs, rpIDs []string, argDsp *utils.ArgDispatcher) (costData map[string]interface{}, err error) {
	costData = make(map[string]interface{})
	if err = ev.CheckMandatoryFields([]string{utils.Account,
		utils.Destination, utils.SetupTime}); err != nil {
		return
	}
	var acnt, subj, dst string
	if acnt, err = ev.FieldAsString(utils.Account); err != nil {
		return
	}
	if subj, err = ev.FieldAsString(utils.Account); err != nil {
		if err != utils.ErrNotFound {
			return
		}
		subj = acnt
	}
	if dst, err = ev.FieldAsString(utils.Destination); err != nil {
		return
	}
	var sTime time.Time
	if sTime, err = ev.FieldAsTime(utils.SetupTime, spS.cgrcfg.GeneralCfg().DefaultTimezone); err != nil {
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
		if err := spS.connMgr.Call(spS.cgrcfg.SupplierSCfg().RALsConns, nil, utils.ResponderGetMaxSessionTimeOnAccounts,
			&utils.GetMaxSessionTimeOnAccountsArgs{
				Tenant:        ev.Tenant,
				Subject:       subj,
				Destination:   dst,
				SetupTime:     sTime,
				Usage:         usage,
				AccountIDs:    acntIDs,
				ArgDispatcher: argDsp,
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
		if err := spS.connMgr.Call(spS.cgrcfg.SupplierSCfg().RALsConns, nil, utils.ResponderGetCostOnRatingPlans,
			&utils.GetCostOnRatingPlansArgs{
				Tenant:        ev.Tenant,
				Account:       acnt,
				Subject:       subj,
				Destination:   dst,
				SetupTime:     sTime,
				Usage:         usage,
				RatingPlanIDs: rpIDs,
				ArgDispatcher: argDsp,
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
func (spS *SupplierService) statMetrics(statIDs []string, tenant string) (stsMetric map[string]float64, err error) {
	stsMetric = make(map[string]float64)
	provStsMetrics := make(map[string][]float64)
	if len(spS.cgrcfg.SupplierSCfg().StatSConns) != 0 {
		for _, statID := range statIDs {
			var metrics map[string]float64
			if err = spS.connMgr.Call(spS.cgrcfg.SupplierSCfg().StatSConns, nil, utils.StatSv1GetQueueFloatMetrics,
				&utils.TenantIDWithArgDispatcher{TenantID: &utils.TenantID{Tenant: tenant, ID: statID}}, &metrics); err != nil &&
				err.Error() != utils.ErrNotFound.Error() {
				utils.Logger.Warning(
					fmt.Sprintf("<SupplierS> error: %s getting statMetrics for stat : %s", err.Error(), statID))
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
func (spS *SupplierService) statMetricsForLoadDistribution(statIDs []string, tenant string) (result float64, err error) {
	provStsMetrics := make(map[string][]float64)
	if len(spS.cgrcfg.SupplierSCfg().StatSConns) != 0 {
		for _, statID := range statIDs {
			// check if we get an ID in the following form (StatID:MetricID)
			statWithMetric := strings.Split(statID, utils.InInFieldSep)
			var metrics map[string]float64
			if err = spS.connMgr.Call(
				spS.cgrcfg.SupplierSCfg().StatSConns, nil,
				utils.StatSv1GetQueueFloatMetrics,
				&utils.TenantIDWithArgDispatcher{
					TenantID: &utils.TenantID{
						Tenant: tenant, ID: statWithMetric[0]}},
				&metrics); err != nil &&
				err.Error() != utils.ErrNotFound.Error() {
				utils.Logger.Warning(
					fmt.Sprintf("<SupplierS> error: %s getting statMetrics for stat : %s",
						err.Error(), statWithMetric[0]))
			}
			if len(statWithMetric) == 2 { // in case we have MetricID defined with StatID we consider only that metric
				// check if statQueue have metric defined
				if metricVal, has := metrics[statWithMetric[1]]; !has {
					return 0, fmt.Errorf("<%s> error: %s metric %s for statID: %s",
						utils.SupplierS, utils.ErrNotFound, statWithMetric[1], statWithMetric[0])
				} else {
					provStsMetrics[statWithMetric[1]] = append(provStsMetrics[statWithMetric[1]], metricVal)
				}
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
func (spS *SupplierService) resourceUsage(resIDs []string, tenant string) (tUsage float64, err error) {
	if len(spS.cgrcfg.SupplierSCfg().ResourceSConns) != 0 {
		for _, resID := range resIDs {
			var res Resource
			if err = spS.connMgr.Call(spS.cgrcfg.SupplierSCfg().ResourceSConns, nil, utils.ResourceSv1GetResource,
				&utils.TenantID{Tenant: tenant, ID: resID}, &res); err != nil &&
				err.Error() != utils.ErrNotFound.Error() {
				utils.Logger.Warning(
					fmt.Sprintf("<SupplierS> error: %s getting resource for ID : %s", err.Error(), resID))
				continue
			}
			tUsage += res.totalUsage()
		}
	}
	return
}

func (spS *SupplierService) populateSortingData(ev *utils.CGREvent, spl *Supplier,
	extraOpts *optsGetSuppliers, argDsp *utils.ArgDispatcher) (srtSpl *SortedSupplier, pass bool, err error) {
	sortedSpl := &SortedSupplier{
		SupplierID: spl.ID,
		SortingData: map[string]interface{}{
			utils.Weight: spl.Weight,
		},
		SupplierParameters: spl.SupplierParameters,
	}
	//calculate costData if we have fields
	if len(spl.AccountIDs) != 0 || len(spl.RatingPlanIDs) != 0 {
		costData, err := spS.costForEvent(ev, spl.AccountIDs, spl.RatingPlanIDs, argDsp)
		if err != nil {
			if extraOpts.ignoreErrors {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> ignoring supplier with ID: %s, err: %s",
						utils.SupplierS, spl.ID, err.Error()))
				return nil, false, nil
			} else {
				return nil, false, err
			}
		} else if len(costData) == 0 {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> ignoring supplier with ID: %s, missing cost information",
					utils.SupplierS, spl.ID))
		} else {
			if extraOpts.maxCost != 0 &&
				costData[utils.Cost].(float64) > extraOpts.maxCost {
				return nil, false, nil
			}
			for k, v := range costData {
				sortedSpl.SortingData[k] = v
			}
		}
	}
	//calculate metrics
	//in case we have *load strategy we use statMetricsForLoadDistribution function to calculate the result
	if len(spl.StatIDs) != 0 {
		if extraOpts.sortingStragety == utils.MetaLoad {
			metricSum, err := spS.statMetricsForLoadDistribution(spl.StatIDs, ev.Tenant) //create metric map for suppier
			if err != nil {
				if extraOpts.ignoreErrors {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> ignoring supplier with ID: %s, err: %s",
							utils.SupplierS, spl.ID, err.Error()))
					return nil, false, nil
				} else {
					return nil, false, err
				}
			}
			sortedSpl.SortingData[utils.Load] = metricSum
		} else {
			metricSupp, err := spS.statMetrics(spl.StatIDs, ev.Tenant) //create metric map for suppier
			if err != nil {
				if extraOpts.ignoreErrors {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> ignoring supplier with ID: %s, err: %s",
							utils.SupplierS, spl.ID, err.Error()))
					return nil, false, nil
				} else {
					return nil, false, err
				}
			}
			//add metrics from statIDs in SortingData
			for key, val := range metricSupp {
				sortedSpl.SortingData[key] = val
			}
			//check if the supplier have the metric from sortingParameters
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
	if len(spl.ResourceIDs) != 0 {
		resTotalUsage, err := spS.resourceUsage(spl.ResourceIDs, ev.Tenant)
		if err != nil {
			if extraOpts.ignoreErrors {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> ignoring supplier with ID: %s, err: %s",
						utils.SupplierS, spl.ID, err.Error()))
				return nil, false, nil
			} else {
				return nil, false, err
			}
		}
		sortedSpl.SortingData[utils.ResourceUsage] = resTotalUsage
	}
	//filter the supplier
	if len(spl.FilterIDs) != 0 {
		//construct the DP and pass it to filterS
		nM := utils.MapStorage{
			utils.MetaReq:  ev.Event,
			utils.MetaVars: sortedSpl.SortingData,
		}

		if pass, err = spS.filterS.Pass(ev.Tenant, spl.FilterIDs,
			nM); err != nil {
			return nil, false, err
		} else if !pass {
			return nil, false, nil
		}
	}
	return sortedSpl, true, nil
}

// supliersForEvent will return the list of valid supplier IDs
// for event based on filters and sorting algorithms
func (spS *SupplierService) sortedSuppliersForEvent(args *ArgsGetSuppliers) (sortedSuppls *SortedSuppliers, err error) {
	if _, has := args.CGREvent.Event[utils.Usage]; !has {
		args.CGREvent.Event[utils.Usage] = time.Duration(time.Minute) // make sure we have default set for Usage
	}
	var splPrfls []*SupplierProfile
	if splPrfls, err = spS.matchingSupplierProfilesForEvent(args.CGREvent, true); err != nil {
		return
	}
	splPrfl := splPrfls[0]
	extraOpts, err := args.asOptsGetSuppliers() // convert suppliers arguments into internal options used to limit data
	if err != nil {
		return nil, err
	}
	extraOpts.sortingParameters = splPrfl.SortingParameters // populate sortingParameters in extraOpts
	extraOpts.sortingStragety = splPrfl.Sorting             // populate sortinStrategy in extraOpts
	sortedSuppliers, err := spS.sorter.SortSuppliers(splPrfl.ID, splPrfl.Sorting,
		splPrfl.Suppliers, args.CGREvent, extraOpts, args.ArgDispatcher)
	if err != nil {
		return nil, err
	}
	if args.Paginator.Offset != nil {
		if *args.Paginator.Offset <= len(sortedSuppliers.SortedSuppliers) {
			sortedSuppliers.SortedSuppliers = sortedSuppliers.SortedSuppliers[*args.Paginator.Offset:]
		}
	}
	if args.Paginator.Limit != nil {
		if *args.Paginator.Limit <= len(sortedSuppliers.SortedSuppliers) {
			sortedSuppliers.SortedSuppliers = sortedSuppliers.SortedSuppliers[:*args.Paginator.Limit]
		}
	}
	sortedSuppliers.Count = len(sortedSuppliers.SortedSuppliers)
	return sortedSuppliers, nil
}

type ArgsGetSuppliers struct {
	IgnoreErrors bool
	MaxCost      string // toDo: try with interface{} here
	*utils.CGREvent
	utils.Paginator
	*utils.ArgDispatcher
}

func (args *ArgsGetSuppliers) asOptsGetSuppliers() (opts *optsGetSuppliers, err error) {
	opts = &optsGetSuppliers{ignoreErrors: args.IgnoreErrors}
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
		if cc, err := cd.GetCost(); err != nil {
			return nil, err
		} else {
			opts.maxCost = cc.Cost
		}
	} else if args.MaxCost != "" {
		if opts.maxCost, err = strconv.ParseFloat(args.MaxCost,
			64); err != nil {
			return nil, err
		}
	}
	return
}

type optsGetSuppliers struct {
	ignoreErrors      bool
	maxCost           float64
	sortingParameters []string //used for QOS strategy
	sortingStragety   string
}

// V1GetSupplierProfilesForEvent returns the list of valid supplier IDs
func (spS *SupplierService) V1GetSuppliers(args *ArgsGetSuppliers, reply *SortedSuppliers) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if missing := utils.MissingStructFields(args.CGREvent, []string{utils.Tenant, utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.CGREvent.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if len(spS.cgrcfg.SupplierSCfg().AttributeSConns) != 0 {
		attrArgs := &AttrArgsProcessEvent{
			Context: utils.StringPointer(utils.FirstNonEmpty(
				utils.IfaceAsString(args.CGREvent.Event[utils.Context]),
				utils.MetaSuppliers)),
			CGREvent:      args.CGREvent,
			ArgDispatcher: args.ArgDispatcher,
		}
		var rplyEv AttrSProcessEventReply
		if err := spS.connMgr.Call(spS.cgrcfg.SupplierSCfg().AttributeSConns, nil,
			utils.AttributeSv1ProcessEvent, attrArgs, &rplyEv); err == nil && len(rplyEv.AlteredFields) != 0 {
			args.CGREvent = rplyEv.CGREvent
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrAttributeS(err)
		}
	}
	sSps, err := spS.sortedSuppliersForEvent(args)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *sSps
	return
}

// V1GetSupplierProfilesForEvent returns the list of valid supplier profiles
func (spS *SupplierService) V1GetSupplierProfilesForEvent(args *utils.CGREventWithArgDispatcher, reply *[]*SupplierProfile) (err error) {
	if missing := utils.MissingStructFields(args.CGREvent, []string{utils.Tenant, utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.CGREvent.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	sPs, err := spS.matchingSupplierProfilesForEvent(args.CGREvent, false)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = sPs
	return
}
