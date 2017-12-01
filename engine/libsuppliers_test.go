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
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestLibSuppliersSortCost(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					Cost:   0.1,
					Weight: 10.0,
				},
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					Cost:   0.1,
					Weight: 20.0,
				},
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					Cost:   0.05,
					Weight: 10.0,
				},
			},
		},
	}
	sSpls.SortCost()
	eOrderedSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					Cost:   0.05,
					Weight: 10.0,
				},
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					Cost:   0.1,
					Weight: 20.0,
				},
			},
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					Cost:   0.1,
					Weight: 10.0,
				},
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
			ID:            "supplier1",
			FilterIDs:     []string{},
			AccountIDs:    []string{},
			RatingPlanIDs: []string{},
			ResourceIDs:   []string{},
			StatIDs:       []string{},
			Weight:        10.0,
		},
		&Supplier{
			ID:            "supplier2",
			FilterIDs:     []string{},
			AccountIDs:    []string{},
			RatingPlanIDs: []string{},
			ResourceIDs:   []string{},
			StatIDs:       []string{},
			Weight:        20.0,
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
			},
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					"Weight": 10.0,
				},
			},
		},
	}
	se := &SupplierEvent{
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

/*
func TestLibSuppliersAnswerTime(t *testing.T) {
	event := make(map[string]interface{})
	event[utils.ANSWER_TIME] = time.Now().Local()
	se := &SupplierEvent{
		Tenant: "cgrates.org",
		ID:     "supplierevent1",
		Event:  event,
	}
	seErr := &SupplierEvent{
		Tenant: "cgrates.org",
		ID:     "supplierevent1",
		Event:  make(map[string]interface{}),
	}
	answ, err := se.AnswerTime("UTC")
	if err != nil {
		t.Error(err)
	}
	if answ != event[utils.ANSWER_TIME] {
		t.Errorf("Expecting: %+v, received: %+v", event[utils.ANSWER_TIME], answ)
	}
	answ, err = seErr.AnswerTime("CET")
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
*/

func TestLibSuppliersFieldAsString(t *testing.T) {
	event := make(map[string]interface{})
	event["supplierprofile1"] = "Supplier"
	event["UsageInterval"] = time.Duration(1 * time.Second)
	event["PddInterval"] = "1s"
	event["Weight"] = 20.0
	se := &SupplierEvent{
		Tenant: "cgrates.org",
		ID:     "supplierevent1",
		Event:  event,
	}
	answ, err := se.FieldAsString("UsageInterval")
	if err != nil {
		t.Error(err)
	}
	if answ != "1s" {
		t.Errorf("Expecting: %+v, received: %+v", event["UsageInterval"], answ)
	}
	answ, err = se.FieldAsString("PddInterval")
	if err != nil {
		t.Error(err)
	}
	if answ != event["PddInterval"] {
		t.Errorf("Expecting: %+v, received: %+v", event["PddInterval"], answ)
	}
	answ, err = se.FieldAsString("supplierprofile1")
	if err != nil {
		t.Error(err)
	}
	if answ != event["supplierprofile1"] {
		t.Errorf("Expecting: %+v, received: %+v", event["supplierprofile1"], answ)
	}
	answ, err = se.FieldAsString("Weight")
	if err != nil {
		t.Error(err)
	}
	if answ != "20" {
		t.Errorf("Expecting: %+v, received: %+v", event["Weight"], answ)
	}

}

/*
func TestLibSuppliersUsage(t *testing.T) {
	event := make(map[string]interface{})
	event[utils.USAGE] = time.Duration(20 * time.Second)
	se := &SupplierEvent{
		Tenant: "cgrates.org",
		ID:     "supplierevent1",
		Event:  event,
	}
	seErr := &SupplierEvent{
		Tenant: "cgrates.org",
		ID:     "supplierevent1",
		Event:  make(map[string]interface{}),
	}
	answ, err := se.Usage()
	if err != nil {
		t.Error(err)
	}
	if answ != event[utils.USAGE] {
		t.Errorf("Expecting: %+v, received: %+v", event[utils.USAGE], answ)
	}
	answ, err = seErr.Usage()
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
*/
