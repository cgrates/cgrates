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

// SupplierEvent is an event processed by Supplier Service
type SupplierEvent struct {
	Tenant string
	ID     string
	Event  map[string]interface{}
}

// AnswerTime returns the AnswerTime of StatEvent
func (le *SupplierEvent) AnswerTime(timezone string) (at time.Time, err error) {
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

// Supplier defines supplier related information used within a SupplierProfile
type Supplier struct {
	ID            string // SupplierID
	FilterIDs     []string
	RatingPlanIDs []string // used when computing price
	ResourceIDs   []string // queried in some strategies
	StatIDs       []string // queried in some strategies
	Weight        float64
}

// Suppliers is a sortable list of Supplier
type Suppliers []*Supplier

// Sort is part of sort interface, sort based on Weight
func (lss Suppliers) Sort() {
	sort.Slice(lss, func(i, j int) bool { return lss[i].Weight > lss[j].Weight })
}

// SupplierProfile represents the configuration of a Supplier profile
type SupplierProfile struct {
	Tenant             string
	ID                 string // LCR Profile ID
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval // Activation interval
	Sorting            string                    // Sorting strategy
	SortingParams      []string
	Suppliers          Suppliers
	Blocker            bool // do not process further profiles after this one
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

// SuppliersReply is returned as part of GetSuppliers call
type SortedSuppliers struct {
	ProfileID       string
	Sorting         string
	SortedSuppliers []*SortedSupplier
}

// SupplierReply represents one supplier in
type SortedSupplier struct {
	SupplierID  string
	SortingData map[string]interface{} // store here extra info like cost or stats
}

// NewLCRService initializes a LCRService
func NewSupplierService(dm *DataManager, timezone string,
	filterS *FilterS, indexedFields []string, resourceS,
	statS rpcclient.RpcClientConnection) (spS *SupplierService, err error) {
	spS = &SupplierService{
		dm:            dm,
		timezone:      timezone,
		filterS:       filterS,
		resourceS:     resourceS,
		statS:         statS,
		indexedFields: indexedFields}
	if spS.sorter, err = NewSupplierSortDispatcher(spS); err != nil {
		return nil, err
	}
	return
}

// SupplierService is the service computing Supplier queries
type SupplierService struct {
	dm            *DataManager
	timezone      string
	filterS       *FilterS
	indexedFields []string
	resourceS,
	statS rpcclient.RpcClientConnection
	sorter SupplierSortDispatcher
}

// ListenAndServe will initialize the service
func (lcrS *SupplierService) ListenAndServe(exitChan chan bool) error {
	utils.Logger.Info(fmt.Sprintf("<%s> start listening for requests", utils.SupplierS))
	e := <-exitChan
	exitChan <- e // put back for the others listening for shutdown request
	return nil
}

// Shutdown is called to shutdown the service
func (lcrS *SupplierService) Shutdown() error {
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown initialized", utils.SupplierS))
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown complete", utils.SupplierS))
	return nil
}

// matchingSupplierProfilesForEvent returns ordered list of matching resources which are active by the time of the call
func (lcrS *SupplierService) matchingSupplierProfilesForEvent(ev *SupplierEvent) (lps SupplierProfiles, err error) {
	matchingLPs := make(map[string]*SupplierProfile)
	lpIDs, err := matchingItemIDsForEvent(ev.Event, lcrS.indexedFields,
		lcrS.dm, utils.SupplierProfilesStringIndex+ev.Tenant)
	if err != nil {
		return nil, err
	}
	lockIDs := utils.PrefixSliceItems(lpIDs.Slice(), utils.SupplierProfilesStringIndex)
	guardian.Guardian.GuardIDs(config.CgrConfig().LockingTimeout, lockIDs...)
	defer guardian.Guardian.UnguardIDs(lockIDs...)
	for lpID := range lpIDs {
		lcrPrfl, err := lcrS.dm.GetSupplierProfile(ev.Tenant, lpID, false, utils.NonTransactional)
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
	lps = make(SupplierProfiles, len(matchingLPs))
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
func (lcrS *SupplierService) costForEvent(ev *SupplierEvent, rpIDs []string) (ec *EventCost, err error) {
	return
}

// statMetrics will query a list of statIDs and return composed metric values
// first metric found is always returned
func (lcrS *SupplierService) statMetrics(statIDs []string, metricIDs []string) (sms map[string]StatMetric, err error) {
	return
}

// resourceUsage returns sum of all resource usages out of list
func (lcrS *SupplierService) resourceUsage(resIDs []string) (tUsage float64, err error) {
	return
}

// supliersForEvent will return the list of valid supplier IDs
// for event based on filters and sorting algorithms
func (spS *SupplierService) supliersForEvent(ev *SupplierEvent) (sortedSuppls *SortedSuppliers, err error) {
	var suppPrfls SupplierProfiles
	if suppPrfls, err = spS.matchingSupplierProfilesForEvent(ev); err != nil {
		return
	} else if len(suppPrfls) == 0 {
		return nil, utils.ErrNotFound
	}
	lcrPrfl := suppPrfls[0] // pick up the first lcr profile as winner
	var lss Suppliers
	for _, s := range lcrPrfl.Suppliers {
		if len(s.FilterIDs) != 0 { // filters should be applied, check them here
			if pass, err := spS.filterS.PassFiltersForEvent(ev.Tenant,
				map[string]interface{}{"SupplierID": s.ID}, s.FilterIDs); err != nil {
				return nil, err
			} else if !pass {
				continue
			}
		}
		lss = append(lss, s)
	}
	return spS.sorter.SortSuppliers(lcrPrfl.ID, lcrPrfl.Sorting, lss)
}
