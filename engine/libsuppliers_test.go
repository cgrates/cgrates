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
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestLibSuppliersSortCost(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				SupplierParameters: "param1",
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 20.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Cost:   0.05,
					utils.Weight: 10.0,
				},
				SupplierParameters: "param3",
			},
		},
	}
	sSpls.SortCost()
	eOrderedSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Cost:   0.05,
					utils.Weight: 10.0,
				},
				SupplierParameters: "param3",
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 20.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				SupplierParameters: "param1",
			},
		},
	}
	if !reflect.DeepEqual(eOrderedSpls, sSpls) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eOrderedSpls), utils.ToJSON(sSpls))
	}
}

func TestLibSuppliersSortWeight(t *testing.T) {
	spl := []*Supplier{
		&Supplier{
			ID:                 "supplier1",
			FilterIDs:          []string{},
			AccountIDs:         []string{},
			RatingPlanIDs:      []string{},
			ResourceIDs:        []string{},
			StatIDs:            []string{},
			Weight:             10.0,
			SupplierParameters: "param1",
		},
		&Supplier{
			ID:                 "supplier2",
			FilterIDs:          []string{},
			AccountIDs:         []string{},
			RatingPlanIDs:      []string{},
			ResourceIDs:        []string{},
			StatIDs:            []string{},
			Weight:             20.0,
			SupplierParameters: "param2",
		},
	}
	eSpls := SortedSuppliers{
		ProfileID: "SPL_WEIGHT_1",
		Sorting:   utils.MetaWeight,
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					"Weight": 20.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					"Weight": 10.0,
				},
				SupplierParameters: "param1",
			},
		},
	}
	se := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierevent1",
		Event:  make(map[string]interface{}),
	}
	ws := NewWeightSorter()
	result, err := ws.SortSuppliers("SPL_WEIGHT_1", spl, se)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eSpls.ProfileID, result.ProfileID) {
		t.Errorf("Expecting: %+v, received: %+v", eSpls.ProfileID, result.ProfileID)
	} else if !reflect.DeepEqual(eSpls.SortedSuppliers, result.SortedSuppliers) {
		t.Errorf("Expecting: %+v, received: %+v", eSpls.SortedSuppliers, result.SortedSuppliers)
	} else if !reflect.DeepEqual(eSpls.Sorting, result.Sorting) {
		t.Errorf("Expecting: %+v, received: %+v", eSpls.Sorting, result.Sorting)
	}
}

func TestSortedSuppliersDigest(t *testing.T) {
	eSpls := SortedSuppliers{
		ProfileID: "SPL_WEIGHT_1",
		Sorting:   utils.MetaWeight,
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					"Weight": 20.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					"Weight": 10.0,
				},
				SupplierParameters: "param1",
			},
		},
	}
	exp := "supplier2:param2,supplier1:param1"
	rcv := eSpls.Digest()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expecting: %+v, received: %+v", exp, rcv)
	}
}

func TestSortedSuppliersDigest2(t *testing.T) {
	eSpls := SortedSuppliers{
		ProfileID: "SPL_WEIGHT_1",
		Sorting:   utils.MetaWeight,
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					"Weight": 30.0,
				},
				SupplierParameters: "param1",
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					"Weight": 20.0,
				},
				SupplierParameters: "param2",
			},
		},
	}
	exp := "supplier1:param1,supplier2:param2"
	rcv := eSpls.Digest()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expecting: %+v, received: %+v", exp, rcv)
	}
}

func TestSortedSuppliersDigest3(t *testing.T) {
	eSpls := SortedSuppliers{
		ProfileID:       "SPL_WEIGHT_1",
		Sorting:         utils.MetaWeight,
		SortedSuppliers: []*SortedSupplier{},
	}
	exp := ""
	rcv := eSpls.Digest()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expecting: %+v, received: %+v", exp, rcv)
	}
}
