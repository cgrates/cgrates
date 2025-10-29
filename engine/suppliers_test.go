/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package engine

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
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
				Event: map[string]any{
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
				Event: map[string]any{
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
				Event: map[string]any{
					"Supplier": "SupplierProfilePrefix",
				},
			},
		},
		{ //matching
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "CGR",
				Event: map[string]any{
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
				SortingData: map[string]any{
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
				SortingData: map[string]any{
					"Weight": 30.0,
				},
				SupplierParameters: "param1",
			},
			{
				SupplierID: "supplier2",
				SortingData: map[string]any{
					"Weight": 20.0,
				},
				SupplierParameters: "param2",
			},
			{
				SupplierID: "supplier3",
				SortingData: map[string]any{
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
				SortingData: map[string]any{
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
				SortingData: map[string]any{
					"Weight": 30.0,
				},
				SupplierParameters: "param1",
			},
			{
				SupplierID: "supplier2",
				SortingData: map[string]any{
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
				SortingData: map[string]any{
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
				SortingData: map[string]any{
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
	calls map[string]func(ctx *context.Context, args any, reply any) error
}

func (ccM *ccMock) Call(ctx *context.Context, serviceMethod string, args any, reply any) (err error) {
	if call, has := ccM.calls[serviceMethod]; !has {
		return rpcclient.ErrUnsupporteServiceMethod
	} else {
		return call(ctx, args, reply)
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
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ResponderGetMaxSessionTimeOnAccounts: func(ctx *context.Context, args, reply any) error {
				rpl := map[string]any{
					"Cost": 10,
				}
				*reply.(*map[string]any) = rpl
				return nil
			},
			utils.ResponderGetCostOnRatingPlans: func(ctx *context.Context, args, reply any) error {
				rpl := map[string]any{
					utils.CapMaxUsage: 1000000000,
				}
				*reply.(*map[string]any) = rpl
				return nil

			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
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
			Event: map[string]any{
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
				SortingData: map[string]any{
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
			Event: map[string]any{
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

func TestV1GetSuppliersWithAttributeS(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cfg.SupplierSCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.AttributeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				return utils.ErrNotFound
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): clientConn,
	})
	spS, err := NewSupplierService(dm, nil, cfg, connMgr)
	if err != nil {
		t.Error(err)
	}
	args := &ArgsGetSuppliers{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]any{
				utils.EVENT_NAME:  "Event1",
				utils.Account:     "1002",
				utils.Subject:     "1002",
				utils.Destination: "1001",
				utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:       "1m20s",
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("sup12345"),
		},
	}
	var reply SortedSuppliers

	if err := spS.V1GetSuppliers(args, &reply); err == nil {
		t.Error(err)
	}

}

func TestSupplierServicePopulateSortingData(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_2",
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
	if err := dm.SetFilter(fltr); err != nil {
		t.Error(err)
	}
	spl, err := NewSupplierService(dm, NewFilterS(cfg, nil, dm), cfg, nil)
	if err != nil {
		t.Error(err)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "utils.CGREvent1",
		Event: map[string]any{
			"Supplier":       "SupplierProfile2",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			"Weight":         "20.0",
		},
	}
	sup := &Supplier{
		ID:                 "SPL2",
		FilterIDs:          []string{"FLTR_2"},
		Weight:             20,
		Blocker:            true,
		SupplierParameters: "SortingParameter2",
	}
	extraOpt := &optsGetSuppliers{}

	if _, pass, err := spl.populateSortingData(ev, sup, extraOpt, nil); err != nil || !pass {
		t.Error(err)
	}
}

func TestArgsSuppAsOptsGetSupplier(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	tmpDm := dm
	defer func() {
		dm = tmpDm
	}()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	ev := &ArgsGetSuppliers{
		MaxCost: utils.MetaEventCost,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "CostSuppliers",
			Event: map[string]any{
				utils.Account:     "1003",
				utils.Subject:     "1001",
				utils.Destination: "1002",
				utils.Category:    "call",
				utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:       "1m20s",
			},
		},
	}
	SetDataStorage(dm)
	if _, err := ev.asOptsGetSuppliers(); err == nil {
		t.Error(err)
	}
	//unifinished
}

func TestSupplierscompileCacheParamaters(t *testing.T) {
	sp := SupplierProfile{
		Sorting:           "*load",
		SortingParameters: []string{"1:2", "2:1"},
		Suppliers:         []*Supplier{{}},
	}

	err := sp.compileCacheParameters()

	if err != nil {
		t.Error(err)
	}

	sp2 := SupplierProfile{
		Sorting:           "*load",
		SortingParameters: []string{"test:test"},
		Suppliers:         []*Supplier{{}},
	}

	err = sp2.compileCacheParameters()

	if err.Error() != `strconv.Atoi: parsing "test": invalid syntax` {
		t.Error(err)
	}

	sp3 := SupplierProfile{
		Sorting:           "*load",
		SortingParameters: []string{"*default:1"},
		Suppliers:         []*Supplier{{}},
	}

	err = sp3.compileCacheParameters()

	if err != nil {
		t.Error(err)
	}

	sp4 := SupplierProfile{
		Sorting:           "*load",
		SortingParameters: []string{"test:1"},
		Suppliers:         []*Supplier{{ID: "test"}},
	}

	err = sp4.compileCacheParameters()

	if err != nil {
		t.Error(err)
	}
}

func TestSuppliersCostForEvent(t *testing.T) {
	spS := SupplierService{}
	type args struct {
		ev      *utils.CGREvent
		acntIDs []string
		rpIDs   []string
		argDsp  *utils.ArgDispatcher
	}
	type exp struct {
		costData map[string]any
		err      string
	}
	tests := []struct {
		name string
		args args
		exp  exp
	}{
		{
			name: "CGREvent CheckMandatoryFields",
			args: args{&utils.CGREvent{}, nil, nil, nil},
			exp:  exp{map[string]any{}, "MANDATORY_IE_MISSING: [Account]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv, err := spS.costForEvent(tt.args.ev, tt.args.acntIDs, tt.args.rpIDs, tt.args.argDsp)

			if err != nil {
				if err.Error() != tt.exp.err {
					t.Fatal(err)
				}
			} else {
				t.Error("was expecting an error")
			}

			if !reflect.DeepEqual(rcv, tt.exp.costData) {
				t.Errorf("expected %s, received %s", utils.ToJSON(tt.exp.costData), utils.ToJSON(rcv))
			}
		})
	}
}

func TestSuppliersV1GetSuppliersError(t *testing.T) {
	spS := SupplierService{}
	args := &ArgsGetSuppliers{}

	err := spS.V1GetSuppliers(args, nil)
	if err != nil {
		if err.Error() != "MANDATORY_IE_MISSING: [CGREvent]" {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}
}
