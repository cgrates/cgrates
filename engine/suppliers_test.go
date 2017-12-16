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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	splserv   SupplierService
	sev       *utils.CGREvent
	dmspl     *DataManager
	sprsmatch SupplierProfiles
)

func TestSuppliersSort(t *testing.T) {
	sprs := SupplierProfiles{
		&SupplierProfile{
			Tenant:    "cgrates.org",
			ID:        "supplierprofile1",
			FilterIDs: []string{},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
			},
			Sorting:       "",
			SortingParams: []string{},
			Suppliers: []*Supplier{
				&Supplier{
					ID:            "supplier1",
					FilterIDs:     []string{},
					AccountIDs:    []string{},
					RatingPlanIDs: []string{},
					ResourceIDs:   []string{},
					StatIDs:       []string{},
					Weight:        10.0,
				},
			},
			Blocker: false,
			Weight:  10,
		},
		&SupplierProfile{
			Tenant:    "cgrates.org",
			ID:        "supplierprofile2",
			FilterIDs: []string{},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
			},
			Sorting:       "",
			SortingParams: []string{},
			Suppliers: []*Supplier{
				&Supplier{
					ID:            "supplier1",
					FilterIDs:     []string{},
					AccountIDs:    []string{},
					RatingPlanIDs: []string{},
					ResourceIDs:   []string{},
					StatIDs:       []string{},
					Weight:        20.0,
				},
			},
			Blocker: false,
			Weight:  20.0,
		},
	}
	eSupplierProfile := SupplierProfiles{
		&SupplierProfile{
			Tenant:    "cgrates.org",
			ID:        "supplierprofile2",
			FilterIDs: []string{},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
			},
			Sorting:       "",
			SortingParams: []string{},
			Suppliers: []*Supplier{
				&Supplier{
					ID:            "supplier1",
					FilterIDs:     []string{},
					AccountIDs:    []string{},
					RatingPlanIDs: []string{},
					ResourceIDs:   []string{},
					StatIDs:       []string{},
					Weight:        20.0,
				},
			},
			Blocker: false,
			Weight:  20.0,
		},
		&SupplierProfile{
			Tenant:    "cgrates.org",
			ID:        "supplierprofile1",
			FilterIDs: []string{},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
			},
			Sorting:       "",
			SortingParams: []string{},
			Suppliers: []*Supplier{
				&Supplier{
					ID:            "supplier1",
					FilterIDs:     []string{},
					AccountIDs:    []string{},
					RatingPlanIDs: []string{},
					ResourceIDs:   []string{},
					StatIDs:       []string{},
					Weight:        10.0,
				},
			},
			Blocker: false,
			Weight:  10.0,
		},
	}
	sprs.Sort()
	if !reflect.DeepEqual(eSupplierProfile, sprs) {
		t.Errorf("Expecting: %+v, received: %+v", eSupplierProfile, sprs)
	}
}

func TestSuppliersPopulateSupplierService(t *testing.T) {
	data, _ := NewMapStorage()
	dmspl = NewDataManager(data)
	var filters1 []*RequestFilter
	var filters2 []*RequestFilter
	second := 1 * time.Second
	x, err := NewRequestFilter(MetaString, "supplierprofile1", []string{"Supplier"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters1 = append(filters1, x)
	x, err = NewRequestFilter(MetaGreaterOrEqual, "UsageInterval", []string{second.String()})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters1 = append(filters1, x)
	x, err = NewRequestFilter(MetaGreaterOrEqual, "Weight", []string{"9.0"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters1 = append(filters1, x)
	x, err = NewRequestFilter(MetaString, "supplierprofile2", []string{"Supplier"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters2 = append(filters2, x)
	x, err = NewRequestFilter(MetaGreaterOrEqual, "PddInterval", []string{second.String()})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters2 = append(filters2, x)
	x, err = NewRequestFilter(MetaGreaterOrEqual, "Weight", []string{"15.0"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters2 = append(filters2, x)
	filter1 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "filter1", RequestFilters: filters1}
	filter2 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "filter2", RequestFilters: filters2}
	dmspl.SetFilter(filter1)
	dmspl.SetFilter(filter2)

	ssd := make(map[string]SuppliersSorter)
	ssd[utils.MetaWeight] = NewWeightSorter()
	splserv = SupplierService{
		dm:            dmspl,
		filterS:       &FilterS{dm: dmspl},
		indexedFields: []string{"supplierprofile1", "supplierprofile2"},
		sorter:        ssd,
	}
	ssd[utils.MetaLeastCost] = NewLeastCostSorter(&splserv)
	ev := make(map[string]interface{})
	ev["supplierprofile1"] = "Supplier"
	ev["supplierprofile2"] = "Supplier"
	ev[utils.AnswerTime] = time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC).Local()
	ev["UsageInterval"] = "1s"
	ev["PddInterval"] = "1s"
	ev["Weight"] = "20.0"
	sev = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "utils.CGREvent1",
		Event:  ev,
	}
	sprsmatch = SupplierProfiles{
		&SupplierProfile{
			Tenant:    "cgrates.org",
			ID:        "supplierprofile1",
			FilterIDs: []string{"filter1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
				ExpiryTime:     time.Now().Add(time.Duration(20 * time.Minute)).Local(),
			},
			Sorting:       utils.MetaWeight,
			SortingParams: []string{},
			Suppliers: []*Supplier{
				&Supplier{
					ID:            "supplier1",
					FilterIDs:     []string{"filter1"},
					AccountIDs:    []string{},
					RatingPlanIDs: []string{},
					ResourceIDs:   []string{},
					StatIDs:       []string{},
					Weight:        10.0,
				},
			},
			Blocker: false,
			Weight:  10,
		},
		&SupplierProfile{
			Tenant:    "cgrates.org",
			ID:        "supplierprofile2",
			FilterIDs: []string{"filter2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
				ExpiryTime:     time.Now().Add(time.Duration(20 * time.Minute)).Local(),
			},
			Sorting:       utils.MetaWeight,
			SortingParams: []string{},
			Suppliers: []*Supplier{
				&Supplier{
					ID:            "supplier2",
					FilterIDs:     []string{"filter2"},
					AccountIDs:    []string{},
					RatingPlanIDs: []string{},
					ResourceIDs:   []string{},
					StatIDs:       []string{},
					Weight:        20.0,
				},
				&Supplier{
					ID:            "supplier3",
					FilterIDs:     []string{"filter2"},
					AccountIDs:    []string{},
					RatingPlanIDs: []string{},
					ResourceIDs:   []string{},
					StatIDs:       []string{},
					Weight:        10.0,
				},
				&Supplier{
					ID:            "supplier1",
					FilterIDs:     []string{"filter2"},
					AccountIDs:    []string{},
					RatingPlanIDs: []string{},
					ResourceIDs:   []string{},
					StatIDs:       []string{},
					Weight:        30.0,
				},
			},
			Blocker: false,
			Weight:  20.0,
		},
	}
	for _, spr := range sprsmatch {
		dmspl.DataDB().SetSupplierProfileDrv(spr)
	}
	ref := NewReqFilterIndexer(dmspl, utils.SupplierProfilePrefix, "cgrates.org")
	ref.IndexFilters("supplierprofile1", filters1)
	ref.IndexFilters("supplierprofile2", filters2)
	err = ref.StoreIndexes(false)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	//test here GetReqFilterIndexes with a specific map
	expidx := map[string]map[string]utils.StringMap{
		"supplierprofile1": {
			"Supplier": {
				"supplierprofile1": true,
			},
		},
	}
	splPrf1 := make(map[string]string)
	splPrf1["supplierprofile1"] = "Supplier"
	if rcvidx, err := dmspl.GetReqFilterIndexes(GetDBIndexKey(utils.SupplierProfilePrefix, "cgrates.org", false), splPrf1); err != nil {
		t.Errorf("Error: %+v", err)
	} else {
		if !reflect.DeepEqual(expidx, rcvidx) {
			t.Errorf("Expected: %+v received: %+v", expidx, rcvidx)
		}
	}
}

func TestSuppliersmatchingSupplierProfilesForEvent(t *testing.T) {
	sprf, err := splserv.matchingSupplierProfilesForEvent(sev)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sprsmatch[1], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sprsmatch[1], sprf[0])
	} else if !reflect.DeepEqual(sprsmatch[0], sprf[1]) {
		t.Errorf("Expecting: %+v, received: %+v", sprsmatch[0], sprf[1])
	}
}

func TestSuppliersSortedForEvent(t *testing.T) {
	eFirstSupplierProfile := &SortedSuppliers{
		ProfileID: "supplierprofile2",
		Sorting:   utils.MetaWeight,
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					"Weight": 30.0,
				},
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					"Weight": 20.0,
				},
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					"Weight": 10.0,
				},
			},
		},
	}
	sprf, err := splserv.sortedSuppliersForEvent(sev)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstSupplierProfile, sprf) {
		t.Errorf("Expecting: %+v, received: %+v", eFirstSupplierProfile, sprf)
	}
}
