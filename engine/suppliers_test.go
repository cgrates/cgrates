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
	splService            *SupplierService
	dmSPP                 *DataManager
	sppTest               = SupplierProfiles{
		&SupplierProfile{
			Tenant:    "cgrates.org",
			ID:        "SupplierProfile1",
			FilterIDs: []string{"FLTR_SUPP_1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     cloneExpTimeSuppliers,
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Suppliers: []*Supplier{
				{
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
			ID:        "SupplierProfile2",
			FilterIDs: []string{"FLTR_SUPP_2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     cloneExpTimeSuppliers,
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Suppliers: []*Supplier{
				{
					ID:                 "supplier2",
					FilterIDs:          []string{},
					AccountIDs:         []string{},
					RatingPlanIDs:      []string{},
					ResourceIDs:        []string{},
					StatIDs:            []string{},
					Weight:             20.0,
					SupplierParameters: "param2",
				},
				{
					ID:                 "supplier3",
					FilterIDs:          []string{},
					AccountIDs:         []string{},
					RatingPlanIDs:      []string{},
					ResourceIDs:        []string{},
					StatIDs:            []string{},
					Weight:             10.0,
					SupplierParameters: "param3",
				},
				{
					ID:                 "supplier1",
					FilterIDs:          []string{},
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
			ID:        "SupplierProfilePrefix",
			FilterIDs: []string{"FLTR_SUPP_3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     cloneExpTimeSuppliers,
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Suppliers: []*Supplier{
				{
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
	}
	argsGetSuppliers = []*ArgsGetSuppliers{
		{ //matching SupplierProfile1
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
		},
		{ //matching SupplierProfile2
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
		},
		{ //matching SupplierProfilePrefix
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "utils.CGREvent1",
				Event: map[string]interface{}{
					"Supplier": "SupplierProfilePrefix",
				},
			},
		},
		{ //matching
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "CGR",
				Event: map[string]interface{}{
					"UsageInterval": "1s",
					"PddInterval":   "1s",
				},
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
				{
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
				{
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
				{
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
				{
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

func TestSuppliersPopulateSupplierService(t *testing.T) {
	//Need clone because time.Now adds extra information that DeepEqual doesn't like
	if err := utils.Clone(expTimeSuppliers, &cloneExpTimeSuppliers); err != nil {
		t.Error(err)
	}
	data, _ := NewMapStorage()
	dmSPP = NewDataManager(data)
	defaultCfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	splService, err = NewSupplierService(dmSPP,
		config.CgrConfig().GeneralCfg().DefaultTimezone, &FilterS{
			dm:  dmSPP,
			cfg: defaultCfg}, nil, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
}

func TestSuppliersAddFilters(t *testing.T) {
	fltrSupp1 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_SUPP_1",
		Rules: []*FilterRule{
			{
				Type:      MetaString,
				FieldName: "Supplier",
				Values:    []string{"SupplierProfile1"},
			},
			{
				Type:      MetaGreaterOrEqual,
				FieldName: "UsageInterval",
				Values:    []string{(1 * time.Second).String()},
			},
			{
				Type:      MetaGreaterOrEqual,
				FieldName: utils.Weight,
				Values:    []string{"9.0"},
			},
		},
	}
	dmSPP.SetFilter(fltrSupp1)
	fltrSupp2 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_SUPP_2",
		Rules: []*FilterRule{
			{
				Type:      MetaString,
				FieldName: "Supplier",
				Values:    []string{"SupplierProfile2"},
			},
			{
				Type:      MetaGreaterOrEqual,
				FieldName: "PddInterval",
				Values:    []string{(1 * time.Second).String()},
			},
			{
				Type:      MetaGreaterOrEqual,
				FieldName: utils.Weight,
				Values:    []string{"15.0"},
			},
		},
	}
	dmSPP.SetFilter(fltrSupp2)
	fltrSupp3 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_SUPP_3",
		Rules: []*FilterRule{
			{
				Type:      MetaPrefix,
				FieldName: "Supplier",
				Values:    []string{"SupplierProfilePrefix"},
			},
		},
	}
	dmSPP.SetFilter(fltrSupp3)
}

func TestSuppliersCache(t *testing.T) {
	for _, spp := range sppTest {
		if err = dmSPP.SetSupplierProfile(spp, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each supplier profile from cache
	for _, spp := range sppTest {
		if tempSpp, err := dmSPP.GetSupplierProfile(spp.Tenant,
			spp.ID, true, true, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(spp, tempSpp) {
			t.Errorf("Expecting: %+v, received: %+v", spp, tempSpp)
		}
	}
}

func TestSuppliersmatchingSupplierProfilesForEvent(t *testing.T) {
	sprf, err := splService.matchingSupplierProfilesForEvent(&argsGetSuppliers[0].CGREvent)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[0], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[0], sprf[0])
	}

	sprf, err = splService.matchingSupplierProfilesForEvent(&argsGetSuppliers[1].CGREvent)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[1], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[1], sprf[0])
	}

	sprf, err = splService.matchingSupplierProfilesForEvent(&argsGetSuppliers[2].CGREvent)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[2], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[2], sprf[0])
	}
}

func TestSuppliersSortedForEvent(t *testing.T) {
	eFirstSupplierProfile := &SortedSuppliers{
		ProfileID: "SupplierProfile1",
		Sorting:   utils.MetaWeight,
		SortedSuppliers: []*SortedSupplier{
			{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					"Weight": 10.0,
				},
				SupplierParameters: "param1",
			},
		},
	}
	sprf, err := splService.sortedSuppliersForEvent(argsGetSuppliers[0])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstSupplierProfile, sprf) {
		t.Errorf("Expecting: %+v, received: %+v", eFirstSupplierProfile, sprf)
	}

	eFirstSupplierProfile = &SortedSuppliers{
		ProfileID: "SupplierProfile2",
		Sorting:   utils.MetaWeight,
		SortedSuppliers: []*SortedSupplier{
			{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					"Weight": 30.0,
				},
				SupplierParameters: "param1",
			},
			{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					"Weight": 20.0,
				},
				SupplierParameters: "param2",
			},
			{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					"Weight": 10.0,
				},
				SupplierParameters: "param3",
			},
		},
	}

	sprf, err = splService.sortedSuppliersForEvent(argsGetSuppliers[1])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstSupplierProfile, sprf) {
		t.Errorf("Expecting: %+v, received: %+v", eFirstSupplierProfile, sprf)
	}

	eFirstSupplierProfile = &SortedSuppliers{
		ProfileID: "SupplierProfilePrefix",
		Sorting:   utils.MetaWeight,
		SortedSuppliers: []*SortedSupplier{
			{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					"Weight": 10.0,
				},
				SupplierParameters: "param1",
			},
		},
	}

	sprf, err = splService.sortedSuppliersForEvent(argsGetSuppliers[2])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstSupplierProfile, sprf) {
		t.Errorf("Expecting: %+v, received: %+v", eFirstSupplierProfile, sprf)
	}
}

func TestSuppliersSortedForEventWithLimit(t *testing.T) {
	eFirstSupplierProfile := &SortedSuppliers{
		ProfileID: "SupplierProfile2",
		Sorting:   utils.MetaWeight,
		SortedSuppliers: []*SortedSupplier{
			{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					"Weight": 30.0,
				},
				SupplierParameters: "param1",
			},
			{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					"Weight": 20.0,
				},
				SupplierParameters: "param2",
			},
		},
	}
	argsGetSuppliers[1].Paginator = utils.Paginator{
		Limit: utils.IntPointer(2),
	}
	sprf, err := splService.sortedSuppliersForEvent(argsGetSuppliers[1])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstSupplierProfile, sprf) {
		t.Errorf("Expecting: %+v, received: %+v", eFirstSupplierProfile, sprf)
	}
}

func TestSuppliersSortedForEventWithOffset(t *testing.T) {
	eFirstSupplierProfile := &SortedSuppliers{
		ProfileID: "SupplierProfile2",
		Sorting:   utils.MetaWeight,
		SortedSuppliers: []*SortedSupplier{
			{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					"Weight": 10.0,
				},
				SupplierParameters: "param3",
			},
		},
	}
	argsGetSuppliers[1].Paginator = utils.Paginator{
		Offset: utils.IntPointer(2),
	}
	sprf, err := splService.sortedSuppliersForEvent(argsGetSuppliers[1])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstSupplierProfile, sprf) {
		t.Errorf("Expecting: %+v,received: %+v", utils.ToJSON(eFirstSupplierProfile), utils.ToJSON(sprf))
	}
}

func TestSuppliersSortedForEventWithLimitAndOffset(t *testing.T) {
	eFirstSupplierProfile := &SortedSuppliers{
		ProfileID: "SupplierProfile2",
		Sorting:   utils.MetaWeight,
		SortedSuppliers: []*SortedSupplier{
			{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					"Weight": 20.0,
				},
				SupplierParameters: "param2",
			},
		},
	}
	argsGetSuppliers[1].Paginator = utils.Paginator{
		Limit:  utils.IntPointer(1),
		Offset: utils.IntPointer(1),
	}
	sprf, err := splService.sortedSuppliersForEvent(argsGetSuppliers[1])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstSupplierProfile, sprf) {
		t.Errorf("Expecting: %+v,received: %+v", utils.ToJSON(eFirstSupplierProfile), utils.ToJSON(sprf))
	}
}

func TestSuppliersAsOptsGetSuppliers(t *testing.T) {
	s := &ArgsGetSuppliers{
		IgnoreErrors: true,
		MaxCost:      "10.0",
	}
	spl := &optsGetSuppliers{
		ignoreErrors: true,
		maxCost:      10.0,
	}
	sprf, err := s.asOptsGetSuppliers()
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(spl, sprf) {
		t.Errorf("Expecting: %+v,received: %+v", spl, sprf)
	}
}

func TestSuppliersAsOptsGetSuppliersIgnoreErrors(t *testing.T) {
	s := &ArgsGetSuppliers{
		IgnoreErrors: true,
	}
	spl := &optsGetSuppliers{
		ignoreErrors: true,
	}
	sprf, err := s.asOptsGetSuppliers()
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(spl, sprf) {
		t.Errorf("Expecting: %+v,received: %+v", spl, sprf)
	}
}

func TestSuppliersAsOptsGetSuppliersMaxCost(t *testing.T) {
	s := &ArgsGetSuppliers{
		MaxCost: "10.0",
	}
	spl := &optsGetSuppliers{
		maxCost: 10.0,
	}
	sprf, err := s.asOptsGetSuppliers()
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(spl, sprf) {
		t.Errorf("Expecting: %+v,received: %+v", spl, sprf)
	}
}

func TestSuppliersMatchWithIndexFalse(t *testing.T) {
	splService.filterS.cfg.FilterSCfg().IndexedSelects = false
	sprf, err := splService.matchingSupplierProfilesForEvent(&argsGetSuppliers[0].CGREvent)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[0], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[0], sprf[0])
	}

	sprf, err = splService.matchingSupplierProfilesForEvent(&argsGetSuppliers[1].CGREvent)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[1], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[1], sprf[0])
	}

	sprf, err = splService.matchingSupplierProfilesForEvent(&argsGetSuppliers[2].CGREvent)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[2], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[2], sprf[0])
	}
}
