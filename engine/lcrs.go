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
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// LCREvent is an event processed by LCRService
type LCREvent struct {
	Tenant string
	ID     string
	Event  map[string]interface{}
}

// AnswerTime returns the AnswerTime of StatEvent
func (le *LCREvent) AnswerTime(timezone string) (at time.Time, err error) {
	atIf, has := le.Event[utils.ANSWER_TIME]
	if !has {
		return at, utils.ErrNotFound
	}
	if at, canCast := atIf.(time.Time); canCast {
		return at, nil
	}
	atStr, canCast := atIf.(string)
	if !canCast {
		return at, errors.New("cannot cast to string")
	}
	return utils.ParseTimeDetectLayout(atStr, timezone)
}

// LCRSupplier defines supplier related information used within a LCRProfile
type LCRSupplier struct {
	ID            string // SupplierID
	FilterIDs     []string
	RatingPlanIDs []string // used when computing price
	ResourceIDs   []string // queried in some strategies
	StatIDs       []string // queried in some strategies
	Weight        float64
}

// LCRSuppliers is a sortable list of LCRSupplier
type LCRSuppliers []*LCRSupplier

// Sort is part of sort interface, sort based on Weight
func (lss LCRSuppliers) Sort() {
	sort.Slice(lss, func(i, j int) bool { return lss[i].Weight > lss[j].Weight })
}

// LCRProfile represents the configuration of a LCR profile
type LCRProfile struct {
	Tenant             string
	ID                 string // LCR Profile ID
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval // Activation interval
	Sorting            string                    // Sorting strategy
	SortingParams      []string
	Suppliers          LCRSuppliers
	Blocker            bool // do not process further profiles after this one
	Weight             float64
}

// TenantID returns unique identifier of the LCRProfile in a multi-tenant environment
func (rp *LCRProfile) TenantID() string {
	return utils.ConcatenatedKey(rp.Tenant, rp.ID)
}

// LCRProfiles is a sortable list of LCRProfile
type LCRProfiles []*LCRProfile

// Sort is part of sort interface, sort based on Weight
func (lps LCRProfiles) Sort() {
	sort.Slice(lps, func(i, j int) bool { return lps[i].Weight > lps[j].Weight })
}

// NewLCRService initializes a LCRService
func NewLCRService(dm *DataManager, timezone string,
	filterS *FilterS, indexedFields []string, resourceS,
	statS rpcclient.RpcClientConnection) (lcrS *LCRService, err error) {
	lcrS = &LCRService{
		dm:            dm,
		timezone:      timezone,
		filterS:       filterS,
		resourceS:     resourceS,
		statS:         statS,
		indexedFields: indexedFields}

	if lcrS.sortDispatcher, err = NewSupplierSortDispatcher(lcrS); err != nil {
		return nil, err
	}
	return
}

// LCRService is the service computing LCR queries
type LCRService struct {
	dm            *DataManager
	timezone      string
	filterS       *FilterS
	indexedFields []string
	resourceS,
	statS rpcclient.RpcClientConnection
	sortDispatcher SupplierSortDispatcher
}

// ListenAndServe will initialize the service
func (lcrS *LCRService) ListenAndServe(exitChan chan bool) error {
	utils.Logger.Info(fmt.Sprintf("<%s> start listening for requests", utils.LCRs))
	e := <-exitChan
	exitChan <- e // put back for the others listening for shutdown request
	return nil
}

// Shutdown is called to shutdown the service
func (lcrS *LCRService) Shutdown() error {
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown initialized", utils.LCRs))
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown complete", utils.LCRs))
	return nil
}

// matchingStatQueuesForEvent returns ordered list of matching resources which are active by the time of the call
func (lcrS *LCRService) matchingLCRProfilesForEvent(ev *LCREvent) (lps LCRProfiles, err error) {
	matchingLPs := make(map[string]*LCRProfile)
	lpIDs, err := matchingItemIDsForEvent(ev.Event, lcrS.indexedFields,
		lcrS.dm, utils.LCRProfilesStringIndex+ev.Tenant)
	if err != nil {
		return nil, err
	}
	lockIDs := utils.PrefixSliceItems(lpIDs.Slice(), utils.LCRProfilesStringIndex)
	guardian.Guardian.GuardIDs(config.CgrConfig().LockingTimeout, lockIDs...)
	defer guardian.Guardian.UnguardIDs(lockIDs...)
	for lpID := range lpIDs {
		lcrPrfl, err := lcrS.dm.GetLCRProfile(ev.Tenant, lpID, false, utils.NonTransactional)
		if err != nil {
			if err == utils.ErrNotFound {
				continue
			}
			return nil, err
		}
		aTime, err := ev.AnswerTime(lcrS.timezone)
		if err != nil {
			return nil, err
		}
		if lcrPrfl.ActivationInterval != nil &&
			!lcrPrfl.ActivationInterval.IsActiveAtTime(aTime) { // not active
			continue
		}
		if pass, err := lcrS.filterS.PassFiltersForEvent(ev.Tenant, ev.Event, lcrPrfl.FilterIDs); err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		matchingLPs[lpID] = lcrPrfl
	}
	// All good, convert from Map to Slice so we can sort
	lps = make(LCRProfiles, len(matchingLPs))
	i := 0
	for _, lp := range matchingLPs {
		lps[i] = lp
		i++
	}
	lps.Sort()
	for i, lp := range lps {
		if lp.Blocker { // blocker will stop processing
			lps = lps[:i+1]
			break
		}
	}
	return
}

// costForEvent will compute cost out of ratingPlanIDs for event
func (lcrS *LCRService) costForEvent(ev *LCREvent, rpIDs []string) (ec *EventCost, err error) {
	return
}

// statMetrics will query a list of statIDs and return composed metric values
// first metric found is always returned
func (lcrS *LCRService) statMetrics(statIDs []string, metricIDs []string) (sms map[string]StatMetric, err error) {
	return
}

// resourceUsage returns sum of all resource usages out of list
func (lcrS *LCRService) resourceUsage(resIDs []string) (tUsage float64, err error) {
	return
}

// supliersForEvent will return the list of valid supplier IDs
// for event based on filters and sorting algorithms
func (lcrS *LCRService) supliersForEvent(ev *LCREvent) (lsIDs []string, err error) {
	var lcrPrfls LCRProfiles
	if lcrPrfls, err = lcrS.matchingLCRProfilesForEvent(ev); err != nil {
		return
	} else if len(lcrPrfls) == 0 {
		return nil, utils.ErrNotFound
	}
	lcrPrfl := lcrPrfls[0] // pick up the first lcr profile as winner
	var lss LCRSuppliers
	for _, s := range lcrPrfl.Suppliers {
		if len(s.FilterIDs) != 0 { // filters should be applied, check them here
			if pass, err := lcrS.filterS.PassFiltersForEvent(ev.Tenant,
				map[string]interface{}{"SupplierID": s.ID}, s.FilterIDs); err != nil {
				return nil, err
			} else if !pass {
				continue
			}
		}
		lss = append(lss, s)
	}
	return lcrS.sortDispatcher.SortedSupplierIDs(lcrPrfl.Sorting, lss)
}
