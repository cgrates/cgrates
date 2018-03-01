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
	cloneExpTimeSuppliers time.Time
	expTimeSuppliers      = time.Now().Add(time.Duration(20 * time.Minute))
	splserv               SupplierService
	dmSPP                 *DataManager
	sppTest               = SupplierProfiles{
		&SupplierProfile{
			Tenant:    "cgrates.org",
			ID:        "supplierprofile1",
			FilterIDs: []string{"filter3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     cloneExpTimeSuppliers,
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Suppliers: []*Supplier{
				&Supplier{
					ID:                 "supplier1",
					FilterIDs:          []string{"filter3"},
					AccountIDs:         []string{},
					RatingPlanIDs:      []string{},
					ResourceIDs:        []string{},
					StatIDs:            []string{},
					Weight:             10.0,
					Blocker:            false,
					SupplierParameters: "param1",
				},
			},
			Weight: 10,
		},
		&SupplierProfile{
			Tenant:    "cgrates.org",
			ID:        "supplierprofile2",
			FilterIDs: []string{"filter4"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     cloneExpTimeSuppliers,
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
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
					Blocker:            false,
					SupplierParameters: "param1",
				},
			},
			Weight: 20.0,
		},
		&SupplierProfile{
			Tenant:    "cgrates.org",
			ID:        "supplierprofile3",
			FilterIDs: []string{"preffilter2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     cloneExpTimeSuppliers,
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Suppliers: []*Supplier{
				&Supplier{
					ID:                 "supplier1",
					FilterIDs:          []string{"preffilter2"},
					AccountIDs:         []string{},
					RatingPlanIDs:      []string{},
					ResourceIDs:        []string{},
					StatIDs:            []string{},
					Weight:             10.0,
					Blocker:            false,
					SupplierParameters: "param1",
				},
			},
			Weight: 10,
		},
		&SupplierProfile{
			Tenant:    "cgrates.org",
			ID:        "supplierprofile4",
			FilterIDs: []string{"defaultf2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     cloneExpTimeSuppliers,
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
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
					Blocker:            false,
					SupplierParameters: "param1",
				},
			},
			Weight: 20.0,
		},
	}
	argPagEv = &ArgsGetSuppliers{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]interface{}{
				"Supplier":       "SupplierProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				"Weight":         "20.0",
			},
		},
	}
	argPagEv2 = &ArgsGetSuppliers{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]interface{}{
				"Supplier":       "SupplierProfile2",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				"Weight":         "20.0",
			},
		},
	}
	argPagEv3 = &ArgsGetSuppliers{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]interface{}{
				"Supplier": "supplierprofilePrefix",
			},
		},
	}
	argPagEv4 = &ArgsGetSuppliers{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]interface{}{
				"Weight": "200.00",
			},
		},
	}
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
			Sorting:           "",
			SortingParameters: []string{},
			Suppliers: []*Supplier{
				&Supplier{
					ID:                 "supplier1",
					FilterIDs:          []string{},
					AccountIDs:         []string{},
					RatingPlanIDs:      []string{},
					ResourceIDs:        []string{},
					StatIDs:            []string{},
					Weight:             10.0,
					Blocker:            false,
					SupplierParameters: "param1",
				},
			},
			Weight: 10,
		},
		&SupplierProfile{
			Tenant:    "cgrates.org",
			ID:        "supplierprofile2",
			FilterIDs: []string{},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Sorting:           "",
			SortingParameters: []string{},
			Suppliers: []*Supplier{
				&Supplier{
					ID:                 "supplier1",
					FilterIDs:          []string{},
					AccountIDs:         []string{},
					RatingPlanIDs:      []string{},
					ResourceIDs:        []string{},
					StatIDs:            []string{},
					Weight:             20.0,
					Blocker:            false,
					SupplierParameters: "param1",
				},
			},
			Weight: 20.0,
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
			Sorting:           "",
			SortingParameters: []string{},
			Suppliers: []*Supplier{
				&Supplier{
					ID:                 "supplier1",
					FilterIDs:          []string{},
					AccountIDs:         []string{},
					RatingPlanIDs:      []string{},
					ResourceIDs:        []string{},
					StatIDs:            []string{},
					Weight:             20.0,
					Blocker:            false,
					SupplierParameters: "param1",
				},
			},
			Weight: 20.0,
		},
		&SupplierProfile{
			Tenant:    "cgrates.org",
			ID:        "supplierprofile1",
			FilterIDs: []string{},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Sorting:           "",
			SortingParameters: []string{},
			Suppliers: []*Supplier{
				&Supplier{
					ID:                 "supplier1",
					FilterIDs:          []string{},
					AccountIDs:         []string{},
					RatingPlanIDs:      []string{},
					ResourceIDs:        []string{},
					StatIDs:            []string{},
					Weight:             10.0,
					Blocker:            false,
					SupplierParameters: "param1",
				},
			},
			Weight: 10.0,
		},
	}
	sprs.Sort()
	if !reflect.DeepEqual(eSupplierProfile, sprs) {
		t.Errorf("Expecting: %+v, received: %+v", eSupplierProfile, sprs)
	}
}

func TestSuppliersCache(t *testing.T) {
	//Need clone because time.Now adds extra information that DeepEqual doesn't like
	if err := utils.Clone(expTimeSuppliers, &cloneExpTimeSuppliers); err != nil {
		t.Error(err)
	}
	data, _ := NewMapStorage()
	dmSPP = NewDataManager(data)
	for _, spp := range sppTest {
		if err = dmSPP.SetSupplierProfile(spp, false); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each supplier profile from cache
	for _, spp := range sppTest {
		if tempSpp, err := dmSPP.GetSupplierProfile(spp.Tenant, spp.ID, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(spp, tempSpp) {
			t.Errorf("Expecting: %+v, received: %+v", spp, tempSpp)
		}
	}
}

func TestSuppliersPopulateSupplierService(t *testing.T) {
	data, _ := NewMapStorage()
	dmSPP = NewDataManager(data)
	var filters1 []*FilterRule
	var filters2 []*FilterRule
	var preffilter []*FilterRule
	var defaultf []*FilterRule
	second := 1 * time.Second
	//refresh the DM
	ref := NewFilterIndexer(dmSPP, utils.SupplierProfilePrefix, "cgrates.org")

	//filter1
	x, err := NewFilterRule(MetaString, "Supplier", []string{"SupplierProfile1"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters1 = append(filters1, x)
	x, err = NewFilterRule(MetaGreaterOrEqual, "UsageInterval", []string{second.String()})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters1 = append(filters1, x)
	x, err = NewFilterRule(MetaGreaterOrEqual, "Weight", []string{"9.0"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters1 = append(filters1, x)
	filter3 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "filter3", Rules: filters1}
	dmSPP.SetFilter(filter3)
	ref.IndexTPFilter(FilterToTPFilter(filter3), "supplierprofile1")

	//filter2
	x, err = NewFilterRule(MetaString, "Supplier", []string{"SupplierProfile2"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters2 = append(filters2, x)
	x, err = NewFilterRule(MetaGreaterOrEqual, "PddInterval", []string{second.String()})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters2 = append(filters2, x)
	x, err = NewFilterRule(MetaGreaterOrEqual, "Weight", []string{"15.0"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters2 = append(filters2, x)
	filter4 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "filter4", Rules: filters2}
	dmSPP.SetFilter(filter4)
	ref.IndexTPFilter(FilterToTPFilter(filter4), "supplierprofile2")

	//prefix filter
	x, err = NewFilterRule(MetaPrefix, "Supplier", []string{"supplierprofilePrefix"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	preffilter = append(preffilter, x)
	preffilter2 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "preffilter2", Rules: preffilter}
	dmSPP.SetFilter(preffilter2)
	ref.IndexTPFilter(FilterToTPFilter(preffilter2), "supplierprofile3")

	//default filter
	x, err = NewFilterRule(MetaGreaterOrEqual, "Weight", []string{"200.00"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	defaultf = append(defaultf, x)
	defaultf2 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "defaultf2", Rules: defaultf}
	dmSPP.SetFilter(defaultf2)
	ref.IndexTPFilter(FilterToTPFilter(defaultf2), "supplierprofile4")
	splserv = SupplierService{
		dm:      dmSPP,
		filterS: &FilterS{dm: dmSPP},
		sorter: map[string]SuppliersSorter{
			utils.MetaWeight:    NewWeightSorter(),
			utils.MetaLeastCost: NewLeastCostSorter(&splserv),
		},
	}
	for _, spr := range sppTest {
		dmSPP.SetSupplierProfile(spr, false)
	}
	err = ref.StoreIndexes(true, utils.NonTransactional)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
}

func TestSuppliersmatchingSupplierProfilesForEvent(t *testing.T) {
	sprf, err := splserv.matchingSupplierProfilesForEvent(&argPagEv.CGREvent)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[0], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[0], sprf[0])
	}
	sprf, err = splserv.matchingSupplierProfilesForEvent(&argPagEv2.CGREvent)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[1], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[1], sprf[0])
	}

	sprf, err = splserv.matchingSupplierProfilesForEvent(&argPagEv3.CGREvent)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[2], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[2], sprf[0])
	}
	sprf, err = splserv.matchingSupplierProfilesForEvent(&argPagEv4.CGREvent)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[3], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[3], sprf[0])
	}
}

func TestSuppliersSortedForEvent(t *testing.T) {
	eFirstSupplierProfile := &SortedSuppliers{
		ProfileID: "supplierprofile1",
		Sorting:   utils.MetaWeight,
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					"Weight": 10.0,
				},
				SupplierParameters: "param1",
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
	eFirstSupplierProfile = &SortedSuppliers{
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
	sprf, err = splserv.sortedSuppliersForEvent(argPagEv2)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstSupplierProfile, sprf) {
		t.Errorf("Expecting: %+v, received: %+v", eFirstSupplierProfile, sprf)
	}
	eFirstSupplierProfile = &SortedSuppliers{
		ProfileID: "supplierprofile3",
		Sorting:   utils.MetaWeight,
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					"Weight": 10.0,
				},
				SupplierParameters: "param1",
			},
		},
	}
	sprf, err = splserv.sortedSuppliersForEvent(argPagEv3)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstSupplierProfile, sprf) {
		t.Errorf("Expecting: %+v, received: %+v", eFirstSupplierProfile, sprf)
	}
	eFirstSupplierProfile = &SortedSuppliers{
		ProfileID: "supplierprofile4",
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
	sprf, err = splserv.sortedSuppliersForEvent(argPagEv4)
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
	argPagEv2.Paginator = utils.Paginator{
		Limit: utils.IntPointer(2),
	}
	sprf, err := splserv.sortedSuppliersForEvent(argPagEv2)
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
	argPagEv2.Paginator = utils.Paginator{
		Offset: utils.IntPointer(2),
	}
	sprf, err := splserv.sortedSuppliersForEvent(argPagEv2)
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
	argPagEv2.Paginator = utils.Paginator{
		Limit:  utils.IntPointer(1),
		Offset: utils.IntPointer(1),
	}
	sprf, err := splserv.sortedSuppliersForEvent(argPagEv2)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstSupplierProfile, sprf) {
		t.Errorf("Expecting: %+v,received: %+v", utils.ToJSON(eFirstSupplierProfile), utils.ToJSON(sprf))
	}
}
