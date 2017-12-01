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
	"time"

	"github.com/cgrates/cgrates/utils"
)

const (
	Weight = "Weight"
	Cost   = "Cost"
)

// SuppliersReply is returned as part of GetSuppliers call
type SortedSuppliers struct {
	ProfileID       string            // Profile matched
	Sorting         string            // Sorting algorithm
	SortedSuppliers []*SortedSupplier // list of supplier IDs and SortingData data
}

// SupplierReply represents one supplier in
type SortedSupplier struct {
	SupplierID  string
	SortingData map[string]interface{} // store here extra info like cost or stats
}

// SortWeight is part of sort interface, sort based on Weight
func (sSpls *SortedSuppliers) SortWeight() {
	sort.Slice(sSpls.SortedSuppliers, func(i, j int) bool {
		return sSpls.SortedSuppliers[i].SortingData[Weight].(float64) > sSpls.SortedSuppliers[j].SortingData[Weight].(float64)
	})
}

// SortCost is part of sort interface,
// sort based on Cost with fallback on Weight
func (sSpls *SortedSuppliers) SortCost() {
	sort.Slice(sSpls.SortedSuppliers, func(i, j int) bool {
		if sSpls.SortedSuppliers[i].SortingData[Cost].(float64) == sSpls.SortedSuppliers[j].SortingData[Cost].(float64) {
			return sSpls.SortedSuppliers[i].SortingData[Weight].(float64) > sSpls.SortedSuppliers[j].SortingData[Weight].(float64)
		}
		return sSpls.SortedSuppliers[i].SortingData[Cost].(float64) < sSpls.SortedSuppliers[j].SortingData[Cost].(float64)
	})
}

// SupplierEvent is an event processed by Supplier Service
type SupplierEvent struct {
	Tenant string
	ID     string
	Event  map[string]interface{}
}

func (se *SupplierEvent) CheckMandatoryFields(fldNames []string) error {
	for _, fldName := range fldNames {
		if _, has := se.Event[fldName]; !has {
			return utils.NewErrMandatoryIeMissing(fldName)
		}
	}
	return nil
}

// AnswerTime returns the AnswerTime of SupplierEvent
func (le *SupplierEvent) FieldAsString(fldName string) (val string, err error) {
	iface, has := le.Event[fldName]
	if !has {
		return "", utils.ErrNotFound
	}
	val, canCast := utils.CastFieldIfToString(iface)
	if !canCast {
		return "", fmt.Errorf("cannot cast %s to string", fldName)
	}
	return val, nil
}

// FieldAsTime returns the a field as Time instance
func (le *SupplierEvent) FieldAsTime(fldName string, timezone string) (t time.Time, err error) {
	iface, has := le.Event[fldName]
	if !has {
		err = utils.ErrNotFound
		return
	}
	var canCast bool
	if t, canCast = iface.(time.Time); canCast {
		return
	}
	s, canCast := iface.(string)
	if !canCast {
		err = fmt.Errorf("cannot cast %s to string", fldName)
		return
	}
	return utils.ParseTimeDetectLayout(s, timezone)
}

// FieldAsTime returns the a field as Time instance
func (le *SupplierEvent) FieldAsDuration(fldName string) (d time.Duration, err error) {
	iface, has := le.Event[fldName]
	if !has {
		err = utils.ErrNotFound
		return
	}
	var canCast bool
	if d, canCast = iface.(time.Duration); canCast {
		return
	}
	s, canCast := iface.(string)
	if !canCast {
		err = fmt.Errorf("cannot cast %s to string", fldName)
		return
	}
	return utils.ParseDurationWithNanosecs(s)
}

// SuppliersSorter is the interface which needs to be implemented by supplier sorters
type SuppliersSorter interface {
	SortSuppliers(string, []*Supplier, *SupplierEvent) (*SortedSuppliers, error)
}

// NewSupplierSortDispatcher constructs SupplierSortDispatcher
func NewSupplierSortDispatcher(lcrS *SupplierService) (ssd SupplierSortDispatcher, err error) {
	ssd = make(map[string]SuppliersSorter)
	ssd[utils.MetaWeight] = NewWeightSorter()
	ssd[utils.MetaLeastCost] = NewLeastCostSorter(lcrS)
	return
}

// SupplierStrategyHandler will initialize strategies
// and dispatch requests to them
type SupplierSortDispatcher map[string]SuppliersSorter

func (ssd SupplierSortDispatcher) SortSuppliers(prflID, strategy string,
	suppls []*Supplier, suplEv *SupplierEvent) (sortedSuppls *SortedSuppliers, err error) {
	sd, has := ssd[strategy]
	if !has {
		return nil, fmt.Errorf("unsupported sorting strategy: %s", strategy)
	}
	return sd.SortSuppliers(prflID, suppls, suplEv)
}

func NewWeightSorter() *WeightSorter {
	return &WeightSorter{sorting: utils.MetaWeight}
}

// WeightSorter orders suppliers based on their weight, no cost involved
type WeightSorter struct {
	sorting string
}

func (ws *WeightSorter) SortSuppliers(prflID string,
	suppls []*Supplier, suplEv *SupplierEvent) (sortedSuppls *SortedSuppliers, err error) {
	sortedSuppls = &SortedSuppliers{ProfileID: prflID,
		Sorting:         ws.sorting,
		SortedSuppliers: make([]*SortedSupplier, len(suppls))}
	for i, s := range suppls {
		sortedSuppls.SortedSuppliers[i] = &SortedSupplier{
			SupplierID:  s.ID,
			SortingData: map[string]interface{}{Weight: s.Weight}}
	}
	sortedSuppls.SortWeight()
	return
}
