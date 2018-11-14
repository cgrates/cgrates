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
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
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

// NewLCRService initializes a LCRService
func NewSupplierService(dm *DataManager, timezone string,
	filterS *FilterS, stringIndexedFields, prefixIndexedFields *[]string, resourceS,
	statS, attributeS rpcclient.RpcClientConnection) (spS *SupplierService, err error) {
	if attributeS != nil && reflect.ValueOf(attributeS).IsNil() { // fix nil value in interface
		attributeS = nil
	}
	if resourceS != nil && reflect.ValueOf(resourceS).IsNil() { // fix nil value in interface
		resourceS = nil
	}
	if statS != nil && reflect.ValueOf(statS).IsNil() { // fix nil value in interface
		statS = nil
	}
	spS = &SupplierService{
		dm:                  dm,
		timezone:            timezone,
		filterS:             filterS,
		attributeS:          attributeS,
		resourceS:           resourceS,
		statS:               statS,
		stringIndexedFields: stringIndexedFields,
		prefixIndexedFields: prefixIndexedFields}
	if spS.sorter, err = NewSupplierSortDispatcher(spS); err != nil {
		return nil, err
	}
	return
}

// SupplierService is the service computing Supplier queries
type SupplierService struct {
	dm                  *DataManager
	timezone            string
	filterS             *FilterS
	stringIndexedFields *[]string
	prefixIndexedFields *[]string
	attributeS,
	resourceS,
	statS rpcclient.RpcClientConnection
	sorter SupplierSortDispatcher
}

// ListenAndServe will initialize the service
func (spS *SupplierService) ListenAndServe(exitChan chan bool) error {
	utils.Logger.Info("Starting Supplier Service")
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
func (spS *SupplierService) matchingSupplierProfilesForEvent(ev *utils.CGREvent) (sPrfls SupplierProfiles, err error) {
	matchingLPs := make(map[string]*SupplierProfile)
	sPrflIDs, err := matchingItemIDsForEvent(ev.Event, spS.stringIndexedFields, spS.prefixIndexedFields,
		spS.dm, utils.CacheSupplierFilterIndexes, ev.Tenant, spS.filterS.cfg.FilterSCfg().IndexedSelects)
	if err != nil {
		return nil, err
	}
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
			config.NewNavigableMap(ev.Event)); err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		matchingLPs[lpID] = splPrfl
	}
	// All good, convert from Map to Slice so we can sort
	sPrfls = make(SupplierProfiles, len(matchingLPs))
	i := 0
	for _, sPrfl := range matchingLPs {
		sPrfls[i] = sPrfl
		i++
	}
	sPrfls.Sort()
	return
}

// costForEvent will compute cost out of accounts and rating plans for event
// returns map[string]interface{} with cost and relevant matching information inside
func (spS *SupplierService) costForEvent(ev *utils.CGREvent,
	acntIDs, rpIDs []string) (costData map[string]interface{}, err error) {
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
	if sTime, err = ev.FieldAsTime(utils.SetupTime, spS.timezone); err != nil {
		return
	}
	var usage time.Duration
	if usage, err = ev.FieldAsDuration(utils.Usage); err != nil {
		return
	}
	for _, anctID := range acntIDs {
		cd := &CallDescriptor{
			Direction:     utils.OUT,
			Category:      utils.MetaSuppliers,
			Tenant:        ev.Tenant,
			Subject:       subj,
			Account:       anctID,
			Destination:   dst,
			TimeStart:     sTime,
			TimeEnd:       sTime.Add(usage),
			DurationIndex: usage,
		}
		if maxDur, err := cd.GetMaxSessionDuration(); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> ignoring cost for account: %s, err: %s",
					utils.SupplierS, anctID, err.Error()))
		} else if maxDur >= usage {
			return map[string]interface{}{
				utils.Cost:    0.0,
				utils.Account: anctID,
			}, nil
		}
	}
	for _, rp := range rpIDs { // loop through RatingPlans until we find one without errors
		rPrfl := &RatingProfile{
			Id: utils.ConcatenatedKey(utils.OUT,
				ev.Tenant, utils.MetaSuppliers, subj),
			RatingPlanActivations: RatingPlanActivations{
				&RatingPlanActivation{
					ActivationTime: sTime,
					RatingPlanId:   rp,
				},
			},
		}
		// force cache set so it can be picked by calldescriptor for cost calculation
		Cache.Set(utils.CacheRatingProfiles, rPrfl.Id, rPrfl, nil,
			true, utils.NonTransactional)
		cd := &CallDescriptor{
			Direction:     utils.OUT,
			Category:      utils.MetaSuppliers,
			Tenant:        ev.Tenant,
			Subject:       subj,
			Account:       acnt,
			Destination:   dst,
			TimeStart:     sTime,
			TimeEnd:       sTime.Add(usage),
			DurationIndex: usage,
		}
		cc, err := cd.GetCost()
		Cache.Remove(utils.CacheRatingProfiles, rPrfl.Id,
			true, utils.NonTransactional) // Remove here so we don't overload memory
		if err != nil {
			if err != utils.ErrNotFound {
				return nil, err
			}
			continue
		}
		ec := NewEventCostFromCallCost(cc, "", "")
		return map[string]interface{}{
			utils.Cost:         ec.GetCost(),
			utils.RatingPlanID: rp}, nil
	}
	return
}

// statMetrics will query a list of statIDs and return composed metric values
// first metric found is always returned
func (spS *SupplierService) statMetrics(statIDs []string, tenant string) (stsMetric map[string]float64, err error) {
	stsMetric = make(map[string]float64)
	if spS.statS != nil {
		for _, statID := range statIDs {
			var metrics map[string]float64
			if err := spS.statS.Call(utils.StatSv1GetQueueFloatMetrics,
				&utils.TenantID{Tenant: tenant, ID: statID}, &metrics); err != nil &&
				err.Error() != utils.ErrNotFound.Error() {
				utils.Logger.Warning(
					fmt.Sprintf("<SupplierS> error: %s getting statMetrics for stat : %s", err.Error(), statID))
			}
			for key, val := range metrics {
				stsMetric[utils.ConcatenatedKey(key, statID)] = val
			}

		}
	}
	return
}

// resourceUsage returns sum of all resource usages out of list
func (spS *SupplierService) resourceUsage(resIDs []string) (tUsage float64, err error) {
	return
}

func (spS *SupplierService) populateSortingData(ev *utils.CGREvent, spl *Supplier,
	extraOpts *optsGetSuppliers) (srtSpl *SortedSupplier, pass bool, err error) {
	globalStats := map[string]float64{ //used for QOS strategy
		utils.Weight: spl.Weight,
	}
	sortedSpl := &SortedSupplier{
		SupplierID: spl.ID,
		SortingData: map[string]interface{}{
			utils.Weight: spl.Weight,
		},
		SupplierParameters: spl.SupplierParameters,
	}
	//calculate costData if we have fields
	if len(spl.AccountIDs) != 0 || len(spl.RatingPlanIDs) != 0 {
		costData, err := spS.costForEvent(ev, spl.AccountIDs, spl.RatingPlanIDs)
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
	metricForFilter := map[string]interface{}{
		utils.Weight: spl.Weight,
	}
	//calculate metrics
	if len(spl.StatIDs) != 0 {
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
		for _, metric := range extraOpts.sortingParameters {
			hasMetric := false                         //check if metricSupp have sortingParameter
			for keyWithID, value := range metricSupp { //transfer data from metric into globalStats
				if metric == strings.Split(keyWithID, utils.InInFieldSep)[0] {
					if val, hasKey := globalStats[metric]; !hasKey ||
						(metric == utils.MetaPDD && val < value) || //worst values
						(metric != utils.MetaPDD && val > value) {
						globalStats[metric] = value
						hasMetric = true
					}
				}
			}
			if !hasMetric { //if not have populate with default value
				switch metric {
				default:
					globalStats[metric] = -1
				case utils.MetaPDD:
					globalStats[metric] = 1000000
				}
			}
		}
		for k, v := range metricSupp {
			sortedSpl.SortingData[k] = v
			metricForFilter[strings.Split(k, utils.InInFieldSep)[0]] = v
		}
		sortedSpl.globalStats = globalStats
	}
	//filter the supplier
	if len(spl.FilterIDs) != 0 {
		nM := config.NewNavigableMap(nil)
		nM.Set([]string{"*req"}, ev.Event, false, false)
		nM.Set([]string{"*sd"}, sortedSpl.SortingData, false, false)
		nM.Set([]string{"*gs"}, metricForFilter, false, false)
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
	var suppPrfls SupplierProfiles
	if suppPrfls, err = spS.matchingSupplierProfilesForEvent(&args.CGREvent); err != nil {
		return
	} else if len(suppPrfls) == 0 {
		return nil, utils.ErrNotFound
	}
	splPrfl := suppPrfls[0]                     // pick up the first lcr profile as winner
	extraOpts, err := args.asOptsGetSuppliers() // convert suppliers arguments into internal options used to limit data
	if err != nil {
		return nil, err
	}
	extraOpts.sortingParameters = splPrfl.SortingParameters // populate sortingParameters in extraOpts
	sortedSuppliers, err := spS.sorter.SortSuppliers(splPrfl.ID, splPrfl.Sorting,
		splPrfl.Suppliers, &args.CGREvent, extraOpts)
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
	return sortedSuppliers, nil
}

type ArgsGetSuppliers struct {
	IgnoreErrors bool
	MaxCost      string // toDo: try with interface{} here
	utils.CGREvent
	utils.Paginator
}

func (args *ArgsGetSuppliers) asOptsGetSuppliers() (opts *optsGetSuppliers, err error) {
	opts = &optsGetSuppliers{ignoreErrors: args.IgnoreErrors}
	if args.MaxCost == utils.MetaEventCost { // dynamic cost needs to be calculated from event
		if err = args.CGREvent.CheckMandatoryFields([]string{utils.Account,
			utils.Destination, utils.SetupTime, utils.Usage}); err != nil {
			return
		}
		cd, err := NewCallDescriptorFromCGREvent(&args.CGREvent,
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
}

// V1GetSuppliersForEvent returns the list of valid supplier IDs
func (spS *SupplierService) V1GetSuppliers(args *ArgsGetSuppliers, reply *SortedSuppliers) (err error) {
	if missing := utils.MissingStructFields(&args.CGREvent, []string{"Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.CGREvent.Event == nil {
		return utils.NewErrMandatoryIeMissing("Event")
	}
	if spS.attributeS != nil {
		if args.CGREvent.Context == nil { // populate if not already in
			args.CGREvent.Context = utils.StringPointer(utils.MetaSuppliers)
		}
		attrArgs := &AttrArgsProcessEvent{
			CGREvent: args.CGREvent,
		}
		var rplyEv AttrSProcessEventReply
		if err := spS.attributeS.Call(utils.AttributeSv1ProcessEvent,
			attrArgs, &rplyEv); err == nil && len(rplyEv.AlteredFields) != 0 {
			args.CGREvent = *rplyEv.CGREvent
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
