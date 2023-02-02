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
	"github.com/cgrates/rpcclient"
)

var (
	expTimeSuppliers = time.Now().Add(time.Duration(20 * time.Minute))
	splService       *SupplierService
	dmSPP            *DataManager
	sppTest          = SupplierProfiles{
		&SupplierProfile{
			Tenant:    "cgrates.org",
			ID:        "SupplierProfile1",
			FilterIDs: []string{"FLTR_SUPP_1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     expTimeSuppliers,
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
				ExpiryTime:     expTimeSuppliers,
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
				ExpiryTime:     expTimeSuppliers,
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
			CGREvent: &utils.CGREvent{
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
			CGREvent: &utils.CGREvent{
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
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "utils.CGREvent1",
				Event: map[string]interface{}{
					"Supplier": "SupplierProfilePrefix",
				},
			},
		},
		{ //matching
			CGREvent: &utils.CGREvent{
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
	defaultCfg, _ := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, defaultCfg.DataDbCfg().Items)
	dmSPP = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.SupplierSCfg().StringIndexedFields = nil
	defaultCfg.SupplierSCfg().PrefixIndexedFields = nil
	splService, err = NewSupplierService(dmSPP, &FilterS{
		dm: dmSPP, cfg: defaultCfg}, defaultCfg, nil)
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
				Type:    utils.MetaString,
				Element: "~*req.Supplier",
				Values:  []string{"SupplierProfile1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.UsageInterval",
				Values:  []string{(1 * time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmSPP.SetFilter(fltrSupp1)
	fltrSupp2 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_SUPP_2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Supplier",
				Values:  []string{"SupplierProfile2"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.PddInterval",
				Values:  []string{(1 * time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	dmSPP.SetFilter(fltrSupp2)
	fltrSupp3 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_SUPP_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Supplier",
				Values:  []string{"SupplierProfilePrefix"},
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
	sprf, err := splService.matchingSupplierProfilesForEvent(argsGetSuppliers[0].CGREvent, true)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[0], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[0], sprf[0])
	}

	sprf, err = splService.matchingSupplierProfilesForEvent(argsGetSuppliers[1].CGREvent, true)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[1], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[1], sprf[0])
	}

	sprf, err = splService.matchingSupplierProfilesForEvent(argsGetSuppliers[2].CGREvent, true)
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
		Count:     1,
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
		Count:     3,
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
		Count:     1,
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
		Count:     2,
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
		Count:     1,
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
		Count:     1,
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
	splService.cgrcfg.SupplierSCfg().IndexedSelects = false
	sprf, err := splService.matchingSupplierProfilesForEvent(argsGetSuppliers[0].CGREvent, true)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[0], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[0], sprf[0])
	}

	sprf, err = splService.matchingSupplierProfilesForEvent(argsGetSuppliers[1].CGREvent, true)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[1], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[1], sprf[0])
	}

	sprf, err = splService.matchingSupplierProfilesForEvent(argsGetSuppliers[2].CGREvent, true)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[2], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[2], sprf[0])
	}
}

type ccMock struct {
	calls map[string]func(args interface{}, reply interface{}) error
}

func (ccM *ccMock) Call(serviceMethod string, args interface{}, reply interface{}) (err error) {
	if call, has := ccM.calls[serviceMethod]; !has {
		return rpcclient.ErrUnsupporteServiceMethod
	} else {
		return call(args, reply)
	}
}
func TestSuppliersV1GetSuppliers(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	cfg.SupplierSCfg().IndexedSelects = false
	cfg.SupplierSCfg().RALsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs)}
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResponderGetMaxSessionTimeOnAccounts: func(args, reply interface{}) error {
				rpl := map[string]interface{}{
					"Cost": 10,
				}
				*reply.(*map[string]interface{}) = rpl
				return nil
			},
			utils.ResponderGetCostOnRatingPlans: func(args, reply interface{}) error {
				rpl := map[string]interface{}{
					utils.CapMaxUsage: 1000000000,
				}
				*reply.(*map[string]interface{}) = rpl
				return nil

			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs): clientConn,
	})
	splService, err := NewSupplierService(dm, &FilterS{
		dm: dm, cfg: cfg}, cfg, connMgr)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	arg := &ArgsGetSuppliers{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]interface{}{
				utils.EVENT_NAME:  "Event1",
				utils.Account:     "1001",
				utils.Subject:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:       "1m20s",
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("sup12345"),
		},
	}
	fltr := &Filter{
		ID:     "FLTR_1",
		Tenant: "cgrates.org",
		Rules: []*FilterRule{
			{Type: utils.MetaString,
				Element: "~*req.Account",
				Values:  []string{"1001"},
			},
		},
	}
	if err := dm.SetFilter(fltr); err != nil {
		t.Error(err)
	}
	supplier := &SupplierProfile{
		Tenant:            "cgrates.org",
		ID:                "SUP1",
		FilterIDs:         []string{"FLTR_1"},
		Weight:            10,
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{},
		Suppliers: []*Supplier{
			{
				ID:            "Sup",
				FilterIDs:     []string{},
				AccountIDs:    []string{"1001"},
				RatingPlanIDs: []string{"RT_PLAN1"},
				ResourceIDs:   []string{"RES1"},
				Weight:        10,
			},
		},
	}
	supplier2 := &SupplierProfile{
		Tenant:            "cgrates.org",
		ID:                "SUP2",
		FilterIDs:         []string{"FLTR_1"},
		Weight:            10,
		Sorting:           utils.MetaHC,
		SortingParameters: []string{},
		Suppliers: []*Supplier{
			{
				ID:            "Sup",
				FilterIDs:     []string{},
				AccountIDs:    []string{"1001"},
				RatingPlanIDs: []string{"RT_PLAN1"},
				ResourceIDs:   []string{"RES1"},
				Weight:        10,
			},
		},
	}
	if err := dm.SetSupplierProfile(supplier, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetSupplierProfile(supplier2, true); err != nil {
		t.Error(err)
	}
	exp := &SortedSuppliers{
		ProfileID: "SUP1",
		Sorting:   utils.MetaQOS,
		Count:     1,
		SortedSuppliers: []*SortedSupplier{
			{
				SupplierID: "Sup",
				SortingData: map[string]interface{}{
					"MaxUsage":      1 * time.Second,
					"ResourceUsage": 0,
					"Weight":        10,
				},
			},
		},
	}
	var reply SortedSuppliers
	if err := splService.V1GetSuppliers(arg, &reply); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(exp, reply) {
		t.Errorf("expected %+v,received %+v", utils.ToJSON(exp), utils.ToJSON(reply))
	}
}
func TestSupplierServiceGetSPForEvent(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	cfg.SupplierSCfg().IndexedSelects = false
	splService, err := NewSupplierService(dm, &FilterS{
		dm: dm, cfg: cfg}, cfg, nil)
	if err != nil {
		t.Error(err)
	}
	arg := &utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetHighestCostSuppliers",
			Event: map[string]interface{}{
				utils.Account:     "1001",
				utils.Subject:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:       "1m20s",
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("sup12345"),
		},
	}
	fltr := &Filter{
		ID:     "FLTR_1",
		Tenant: "cgrates.org",
		Rules: []*FilterRule{
			{Type: utils.MetaString,
				Element: "~*req.Account",
				Values:  []string{"1001"},
			},
		},
	}
	if err := dm.SetFilter(fltr); err != nil {
		t.Error(err)
	}
	supplier := &SupplierProfile{
		Tenant:            "cgrates.org",
		ID:                "SUP1",
		FilterIDs:         []string{"FLTR_1"},
		Weight:            10,
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{},
		Suppliers: []*Supplier{
			{
				ID:            "Sup",
				FilterIDs:     []string{},
				AccountIDs:    []string{"1001"},
				RatingPlanIDs: []string{"RT_PLAN1"},
				ResourceIDs:   []string{"RES1"},
				Weight:        10,
			},
		},
	}
	if err := dm.SetSupplierProfile(supplier, true); err != nil {
		t.Error(err)
	}
	var reply []*SupplierProfile
	if err := splService.V1GetSupplierProfilesForEvent(arg, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, []*SupplierProfile{supplier}) {
		t.Errorf("expected %v,received %v", utils.ToJSON([]*SupplierProfile{supplier}), utils.ToJSON(reply))
	}
}
