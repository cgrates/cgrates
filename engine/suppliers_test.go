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
	argPagEv  *ArgsGetSuppliers
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
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Sorting:       "",
			SortingParams: []string{},
			Suppliers: []*Supplier{
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
			},
			Blocker: false,
			Weight:  10,
		},
		&SupplierProfile{
			Tenant:    "cgrates.org",
			ID:        "supplierprofile2",
			FilterIDs: []string{},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Sorting:       "",
			SortingParams: []string{},
			Suppliers: []*Supplier{
				&Supplier{
					ID:                 "supplier1",
					FilterIDs:          []string{},
					AccountIDs:         []string{},
					RatingPlanIDs:      []string{},
					ResourceIDs:        []string{},
					StatIDs:            []string{},
					Weight:             20.0,
					SupplierParameters: "param1",
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
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Sorting:       "",
			SortingParams: []string{},
			Suppliers: []*Supplier{
				&Supplier{
					ID:                 "supplier1",
					FilterIDs:          []string{},
					AccountIDs:         []string{},
					RatingPlanIDs:      []string{},
					ResourceIDs:        []string{},
					StatIDs:            []string{},
					Weight:             20.0,
					SupplierParameters: "param1",
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
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Sorting:       "",
			SortingParams: []string{},
			Suppliers: []*Supplier{
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
	var preffilter []*RequestFilter
	var defaultf []*RequestFilter
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
	x, err = NewRequestFilter(MetaPrefix, "Supplier", []string{"supplierprofilePrefix"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	preffilter = append(preffilter, x)
	x, err = NewRequestFilter(MetaGreaterOrEqual, "Weight", []string{"200.00"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	defaultf = append(defaultf, x)
	filter3 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "filter3", RequestFilters: filters1}
	filter4 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "filter4", RequestFilters: filters2}
	preffilter2 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "preffilter2", RequestFilters: preffilter}
	defaultf2 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "defaultf2", RequestFilters: defaultf}
	dmspl.SetFilter(filter3)
	dmspl.SetFilter(filter4)
	dmspl.SetFilter(preffilter2)
	dmspl.SetFilter(defaultf2)
	ssd := make(map[string]SuppliersSorter)
	ssd[utils.MetaWeight] = NewWeightSorter()
	splserv = SupplierService{
		dm:      dmspl,
		filterS: &FilterS{dm: dmspl},
		sorter:  ssd,
	}
	ssd[utils.MetaLeastCost] = NewLeastCostSorter(&splserv)
	argPagEv = &ArgsGetSuppliers{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]interface{}{
				"supplierprofile1": "Supplier",
				"supplierprofile2": "Supplier",
				utils.AnswerTime:   time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":    "1s",
				"PddInterval":      "1s",
				"Weight":           "20.0",
			},
		},
	}
	var cloneExpTime time.Time
	expTime := time.Now().Add(time.Duration(20 * time.Minute))
	if err := utils.Clone(expTime, &cloneExpTime); err != nil {
		t.Error(err)
	}
	sprsmatch = SupplierProfiles{
		&SupplierProfile{
			Tenant:    "cgrates.org",
			ID:        "supplierprofile1",
			FilterIDs: []string{"filter3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     cloneExpTime,
			},
			Sorting:       utils.MetaWeight,
			SortingParams: []string{},
			Suppliers: []*Supplier{
				&Supplier{
					ID:                 "supplier1",
					FilterIDs:          []string{"filter3"},
					AccountIDs:         []string{},
					RatingPlanIDs:      []string{},
					ResourceIDs:        []string{},
					StatIDs:            []string{},
					Weight:             10.0,
					SupplierParameters: "param1",
				},
			},
			Blocker: false,
			Weight:  10,
		},
		&SupplierProfile{
			Tenant:    "cgrates.org",
			ID:        "supplierprofile2",
			FilterIDs: []string{"filter4"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     cloneExpTime,
			},
			Sorting:       utils.MetaWeight,
			SortingParams: []string{},
			Suppliers: []*Supplier{
				&Supplier{
					ID:                 "supplier2",
					FilterIDs:          []string{"filter4"},
					AccountIDs:         []string{},
					RatingPlanIDs:      []string{},
					ResourceIDs:        []string{},
					StatIDs:            []string{},
					Weight:             20.0,
					SupplierParameters: "param2",
				},
				&Supplier{
					ID:                 "supplier3",
					FilterIDs:          []string{"filter4"},
					AccountIDs:         []string{},
					RatingPlanIDs:      []string{},
					ResourceIDs:        []string{},
					StatIDs:            []string{},
					Weight:             10.0,
					SupplierParameters: "param3",
				},
				&Supplier{
					ID:                 "supplier1",
					FilterIDs:          []string{"filter4"},
					AccountIDs:         []string{},
					RatingPlanIDs:      []string{},
					ResourceIDs:        []string{},
					StatIDs:            []string{},
					Weight:             30.0,
					SupplierParameters: "param1",
				},
			},
			Blocker: false,
			Weight:  20.0,
		},
		&SupplierProfile{
			Tenant:    "cgrates.org",
			ID:        "supplierprofile3",
			FilterIDs: []string{"preffilter2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     cloneExpTime,
			},
			Sorting:       utils.MetaWeight,
			SortingParams: []string{},
			Suppliers: []*Supplier{
				&Supplier{
					ID:                 "supplier1",
					FilterIDs:          []string{"preffilter2"},
					AccountIDs:         []string{},
					RatingPlanIDs:      []string{},
					ResourceIDs:        []string{},
					StatIDs:            []string{},
					Weight:             10.0,
					SupplierParameters: "param1",
				},
			},
			Blocker: false,
			Weight:  10,
		},
		&SupplierProfile{
			Tenant:    "cgrates.org",
			ID:        "supplierprofile4",
			FilterIDs: []string{"defaultf2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     cloneExpTime,
			},
			Sorting:       utils.MetaWeight,
			SortingParams: []string{},
			Suppliers: []*Supplier{
				&Supplier{
					ID:                 "supplier2",
					FilterIDs:          []string{"defaultf2"},
					AccountIDs:         []string{},
					RatingPlanIDs:      []string{},
					ResourceIDs:        []string{},
					StatIDs:            []string{},
					Weight:             20.0,
					SupplierParameters: "param2",
				},
				&Supplier{
					ID:                 "supplier3",
					FilterIDs:          []string{"defaultf2"},
					AccountIDs:         []string{},
					RatingPlanIDs:      []string{},
					ResourceIDs:        []string{},
					StatIDs:            []string{},
					Weight:             10.0,
					SupplierParameters: "param3",
				},
				&Supplier{
					ID:                 "supplier1",
					FilterIDs:          []string{"defaultf2"},
					AccountIDs:         []string{},
					RatingPlanIDs:      []string{},
					ResourceIDs:        []string{},
					StatIDs:            []string{},
					Weight:             30.0,
					SupplierParameters: "param1",
				},
			},
			Blocker: false,
			Weight:  20.0,
		},
	}

	for _, spr := range sprsmatch {
		dmspl.SetSupplierProfile(spr, true)
	}
	ref := NewReqFilterIndexer(dmspl, utils.SupplierProfilePrefix, "cgrates.org")
	ref.IndexTPFilter(FilterToTPFilter(filter3), "supplierprofile1")
	ref.IndexTPFilter(FilterToTPFilter(filter4), "supplierprofile2")
	ref.IndexTPFilter(FilterToTPFilter(filter3), "supplierprofile3")
	ref.IndexTPFilter(FilterToTPFilter(filter4), "supplierprofile4")
	err = ref.StoreIndexes()
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
}

func TestSuppliersmatchingSupplierProfilesForEvent(t *testing.T) {
	sprf, err := splserv.matchingSupplierProfilesForEvent(&argPagEv.CGREvent)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sprsmatch[1], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sprsmatch[1], sprf[0])
	} else if !reflect.DeepEqual(sprsmatch[0], sprf[1]) {
		t.Errorf("Expecting: %+v, received: %+v", sprsmatch[0], sprf[1])
	}

	stringPref := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "utils.CGREvent1",
		Event: map[string]interface{}{
			"Supplier": "supplierprofilePrefix",
		},
	}
	sprf, err = splserv.matchingSupplierProfilesForEvent(stringPref)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sprsmatch[2], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sprsmatch[2], sprf[0])
	}
	evDefault := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "utils.CGREvent1",
		Event: map[string]interface{}{
			"Weight": "200.00",
		},
	}
	sprf, err = splserv.matchingSupplierProfilesForEvent(evDefault)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sprsmatch[3], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sprsmatch[3], sprf[0])
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
				SupplierParameters: "param1",
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					"Weight": 20.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					"Weight": 10.0,
				},
				SupplierParameters: "param3",
			},
		},
	}
	sprf, err := splserv.sortedSuppliersForEvent(argPagEv)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstSupplierProfile, sprf) {
		t.Errorf("Expecting: %+v, received: %+v", eFirstSupplierProfile, sprf)
	}
}

func TestSuppliersSortedForEventWithLimit(t *testing.T) {
	eFirstSupplierProfile := &SortedSuppliers{
		ProfileID: "supplierprofile2",
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
	argPagEv.Paginator = utils.Paginator{
		Limit: utils.IntPointer(2),
	}
	sprf, err := splserv.sortedSuppliersForEvent(argPagEv)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstSupplierProfile, sprf) {
		t.Errorf("Expecting: %+v, received: %+v", eFirstSupplierProfile, sprf)
	}
}

func TestSuppliersSortedForEventWithOffset(t *testing.T) {
	eFirstSupplierProfile := &SortedSuppliers{
		ProfileID: "supplierprofile2",
		Sorting:   utils.MetaWeight,
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					"Weight": 10.0,
				},
				SupplierParameters: "param3",
			},
		},
	}
	argPagEv.Paginator = utils.Paginator{
		Offset: utils.IntPointer(2),
	}
	sprf, err := splserv.sortedSuppliersForEvent(argPagEv)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstSupplierProfile, sprf) {
		t.Errorf("Expecting: %+v,received: %+v", utils.ToJSON(eFirstSupplierProfile), utils.ToJSON(sprf))
	}
}

func TestSuppliersSortedForEventWithLimitAndOffset(t *testing.T) {
	eFirstSupplierProfile := &SortedSuppliers{
		ProfileID: "supplierprofile2",
		Sorting:   utils.MetaWeight,
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					"Weight": 20.0,
				},
				SupplierParameters: "param2",
			},
		},
	}
	argPagEv.Paginator = utils.Paginator{
		Limit:  utils.IntPointer(1),
		Offset: utils.IntPointer(1),
	}
	sprf, err := splserv.sortedSuppliersForEvent(argPagEv)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstSupplierProfile, sprf) {
		t.Errorf("Expecting: %+v,received: %+v", utils.ToJSON(eFirstSupplierProfile), utils.ToJSON(sprf))
	}
}
