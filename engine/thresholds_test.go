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
	"bytes"
	"log"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	dmTH   *DataManager
	thServ *ThresholdService
	tPrfls = []*ThresholdProfile{
		{
			Tenant:    "cgrates.org",
			ID:        "TH_1",
			FilterIDs: []string{"FLTR_TH_1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     false,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_2",
			FilterIDs: []string{"FLTR_TH_2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			MinSleep:  time.Duration(5 * time.Minute),
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     false,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_3",
			FilterIDs: []string{"FLTR_TH_3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			MinSleep:  time.Duration(5 * time.Minute),
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     false,
		},
	}
	ths = Thresholds{
		&Threshold{
			Tenant: "cgrates.org",
			ID:     "TH_1",
			Hits:   0,
		},
		&Threshold{
			Tenant: "cgrates.org",
			ID:     "TH_2",
			Hits:   0,
		},
		&Threshold{
			Tenant: "cgrates.org",
			ID:     "TH_3",
			Hits:   0,
		},
	}
	argsGetThresholds = []*ArgsProcessEvent{
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "Ev1",
				Event: map[string]any{
					"Threshold": "TH_1",
					"Weight":    "10.0",
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "Ev1",
				Event: map[string]any{
					"Threshold": "TH_2",
					"Weight":    "20.0",
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "Ev1",
				Event: map[string]any{
					"Threshold": "ThresholdPrefix123",
				},
			},
		},
	}
)

func TestThresholdsSort(t *testing.T) {
	ts := Thresholds{
		&Threshold{tPrfl: &ThresholdProfile{ID: "FIRST", Weight: 30.0}},
		&Threshold{tPrfl: &ThresholdProfile{ID: "SECOND", Weight: 40.0}},
		&Threshold{tPrfl: &ThresholdProfile{ID: "THIRD", Weight: 30.0}},
		&Threshold{tPrfl: &ThresholdProfile{ID: "FOURTH", Weight: 35.0}},
	}
	ts.Sort()
	eInst := Thresholds{
		&Threshold{tPrfl: &ThresholdProfile{ID: "SECOND", Weight: 40.0}},
		&Threshold{tPrfl: &ThresholdProfile{ID: "FOURTH", Weight: 35.0}},
		&Threshold{tPrfl: &ThresholdProfile{ID: "FIRST", Weight: 30.0}},
		&Threshold{tPrfl: &ThresholdProfile{ID: "THIRD", Weight: 30.0}},
	}
	if !reflect.DeepEqual(eInst, ts) {
		t.Errorf("expecting: %+v, received: %+v", eInst, ts)
	}
}

func TestThresholdsPopulateThresholdService(t *testing.T) {
	defaultCfg, _ := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, defaultCfg.DataDbCfg().Items)
	dmTH = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.ThresholdSCfg().StoreInterval = 0
	defaultCfg.ThresholdSCfg().StringIndexedFields = nil
	defaultCfg.ThresholdSCfg().PrefixIndexedFields = nil
	thServ, err = NewThresholdService(dmTH, defaultCfg, &FilterS{dm: dmTH, cfg: defaultCfg})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
}

func TestThresholdsAddFilters(t *testing.T) {
	fltrTh1 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_TH_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Threshold",
				Values:  []string{"TH_1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmTH.SetFilter(fltrTh1)
	fltrTh2 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_TH_2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Threshold",
				Values:  []string{"TH_2"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	dmTH.SetFilter(fltrTh2)
	fltrTh3 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_TH_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Threshold",
				Values:  []string{"ThresholdPrefix"},
			},
		},
	}
	dmTH.SetFilter(fltrTh3)
}

func TestThresholdsCache(t *testing.T) {
	for _, th := range tPrfls {
		if err = dmTH.SetThresholdProfile(th, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each threshold profile from cache
	for _, th := range tPrfls {
		if temptTh, err := dmTH.GetThresholdProfile(th.Tenant,
			th.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(th, temptTh) {
			t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
		}
	}
	for _, th := range ths {
		if err = dmTH.SetThreshold(th); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each threshold profile from cache
	for _, th := range ths {
		if temptTh, err := dmTH.GetThreshold(th.Tenant,
			th.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(th, temptTh) {
			t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
		}
	}

}

func TestThresholdsmatchingThresholdsForEvent(t *testing.T) {
	if thMatched, err := thServ.matchingThresholdsForEvent(argsGetThresholds[0]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(ths[0].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", ths[0].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(ths[0].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", ths[0].ID, thMatched[0].ID)
	} else if !reflect.DeepEqual(ths[0].Hits, thMatched[0].Hits) {
		t.Errorf("Expecting: %+v, received: %+v", ths[0].Hits, thMatched[0].Hits)
	}

	if thMatched, err := thServ.matchingThresholdsForEvent(argsGetThresholds[1]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(ths[1].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", ths[1].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(ths[1].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", ths[1].ID, thMatched[0].ID)
	} else if !reflect.DeepEqual(ths[1].Hits, thMatched[0].Hits) {
		t.Errorf("Expecting: %+v, received: %+v", ths[1].Hits, thMatched[0].Hits)
	}

	if thMatched, err := thServ.matchingThresholdsForEvent(argsGetThresholds[2]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(ths[2].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", ths[2].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(ths[2].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", ths[2].ID, thMatched[0].ID)
	} else if !reflect.DeepEqual(ths[2].Hits, thMatched[0].Hits) {
		t.Errorf("Expecting: %+v, received: %+v", ths[2].Hits, thMatched[0].Hits)
	}
}

func TestThresholdsProcessEvent(t *testing.T) {
	thIDs := []string{"TH_1"}
	if thMatched, err := thServ.processEvent(argsGetThresholds[0]); err != utils.ErrPartiallyExecuted {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(thIDs, thMatched) {
		t.Errorf("Expecting: %+v, received: %+v", thIDs, thMatched)
	}

	thIDs = []string{"TH_2"}
	if thMatched, err := thServ.processEvent(argsGetThresholds[1]); err != utils.ErrPartiallyExecuted {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(thIDs, thMatched) {
		t.Errorf("Expecting: %+v, received: %+v", thIDs, thMatched)
	}

	thIDs = []string{"TH_3"}
	if thMatched, err := thServ.processEvent(argsGetThresholds[2]); err != utils.ErrPartiallyExecuted {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(thIDs, thMatched) {
		t.Errorf("Expecting: %+v, received: %+v", thIDs, thMatched)
	}
}

func TestThresholdsVerifyIfExecuted(t *testing.T) {
	if thMatched, err := thServ.matchingThresholdsForEvent(argsGetThresholds[0]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(ths[0].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", ths[0].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(ths[0].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", ths[0].ID, thMatched[0].ID)
	} else if thMatched[0].Hits != 1 {
		t.Errorf("Expecting: 1, received: %+v", thMatched[0].Hits)
	}

	if thMatched, err := thServ.matchingThresholdsForEvent(argsGetThresholds[1]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(ths[1].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", ths[1].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(ths[1].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", ths[1].ID, thMatched[0].ID)
	} else if thMatched[0].Hits != 1 {
		t.Errorf("Expecting: 1, received: %+v", thMatched[0].Hits)
	}

	if thMatched, err := thServ.matchingThresholdsForEvent(argsGetThresholds[2]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(ths[2].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", ths[2].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(ths[2].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", ths[2].ID, thMatched[0].ID)
	} else if thMatched[0].Hits != 1 {
		t.Errorf("Expecting: 1, received: %+v", thMatched[0].Hits)
	}
}

func TestThresholdsProcessEvent2(t *testing.T) {
	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH_4",
		FilterIDs: []string{"FLTR_TH_1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		MaxHits:   12,
		Blocker:   false,
		Weight:    20.0,
		ActionIDs: []string{"ACT_1", "ACT_2"},
		Async:     false,
	}
	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH_4",
		Hits:   0,
	}
	ev := &ArgsProcessEvent{
		ThresholdIDs: []string{"TH_1", "TH_2", "TH_3", "TH_4"},
		CGREvent:     argsGetThresholds[0].CGREvent,
	}
	if err = dmTH.SetThresholdProfile(thPrf, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if temptTh, err := dmTH.GetThresholdProfile(thPrf.Tenant,
		thPrf.ID, true, false, utils.NonTransactional); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(thPrf, temptTh) {
		t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
	}
	if err = dmTH.SetThreshold(th); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if temptTh, err := dmTH.GetThreshold(th.Tenant,
		th.ID, true, false, utils.NonTransactional); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(th, temptTh) {
		t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
	}
	thIDs := []string{"TH_1", "TH_4"}
	thIDsRev := []string{"TH_4", "TH_1"}
	if thMatched, err := thServ.processEvent(ev); err != utils.ErrPartiallyExecuted {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(thIDs, thMatched) && !reflect.DeepEqual(thIDsRev, thMatched) {
		t.Errorf("Expecting: %+v, received: %+v", thIDs, thMatched)
	}

	if thMatched, err := thServ.matchingThresholdsForEvent(ev); err != nil {
		t.Errorf("Error: %+v", err)
	} else {
		for _, thM := range thMatched {
			if !reflect.DeepEqual(thPrf.Tenant, thM.Tenant) {
				t.Errorf("Expecting: %+v, received: %+v", thPrf.Tenant, thM.Tenant)
			} else if reflect.DeepEqual(thIDs[0], thM.ID) && thM.Hits != 2 {
				t.Errorf("Expecting: 2 for %+v, received: %+v", thM.ID, thM.Hits)
			} else if reflect.DeepEqual(thIDs[1], thM.ID) && thM.Hits != 1 {
				t.Errorf("Expecting: 1 for %+v, received: %+v", thM.ID, thM.Hits)
			}
		}
	}
}

func TestThresholdSProcessEvent22(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	Cache.Clear(nil)
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	tmpDm := dm
	defer func() {
		SetDataStorage(tmpDm)
	}()
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	thS, err := NewThresholdService(dm, cfg, NewFilterS(cfg, nil, dm))
	if err != nil {
		t.Error(err)
	}

	args := &ArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]any{
				utils.EventType:     utils.AccountUpdate,
				utils.Account:       "1002",
				utils.AllowNegative: true,
				utils.Disabled:      false,
				utils.Units:         12.3},
		},
	}
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Account",
				Values:  []string{"1002"},
			},
		},
	}
	SetDataStorage(dm)
	dm.SetFilter(fltr)
	thP := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "THD_Test",
		FilterIDs: []string{"FLTR_1"},
		MaxHits:   -1,
		MinSleep:  time.Duration(time.Second),
		Blocker:   false,
		Weight:    20.0,
		Async:     false,
		ActionIDs: []string{"ACT_LOG"},
	}
	dm.SetThresholdProfile(thP, true)
	dm.SetThreshold(&Threshold{Tenant: thP.Tenant, ID: thP.ID})
	acs := Actions{
		{ActionType: utils.TOPUP,
			Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY),
				Value:          &utils.ValueFormula{Static: 25},
				DestinationIDs: utils.StringMapPointer(utils.NewStringMap("RET")),
				Weight:         utils.Float64Pointer(20)}},
	}
	dm.SetActions("ACT_LOG", acs, utils.NonTransactional)
	var reply []string
	if err := thS.V1ProcessEvent(args, &reply); err != nil {
		t.Error(err)
	}
}

func TestThresholdForEvent(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	tmpDm := dm
	defer func() {
		SetDataStorage(tmpDm)
	}()
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	thS, err := NewThresholdService(dm, cfg, NewFilterS(cfg, nil, dm))
	if err != nil {
		t.Error(err)
	}
	args := &ArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSesItProccesCDR",
			Event: map[string]any{
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestTerminate",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1002",
				utils.Subject:     "1001",
				utils.Destination: "1001",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       2 * time.Second,
			},
		},
	}
	fltrs := []*Filter{
		{
			Tenant: "cgrates.org",
			ID:     "FLTR_1",
			Rules: []*FilterRule{
				{
					Type:    utils.MetaString,
					Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.ToR,
					Values:  []string{utils.VOICE},
				},
				{
					Type:    utils.MetaString,
					Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.RequestType,
					Values:  []string{utils.META_PREPAID},
				},
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "FLTR_2",
			Rules: []*FilterRule{
				{
					Type:    utils.MetaString,
					Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.OriginID,
					Values:  []string{"TestTerminate"},
				},
			},
		},
	}
	for _, fltr := range fltrs {
		dm.SetFilter(fltr)
	}
	thp := &ThresholdProfile{Tenant: "cgrates.org",
		ID:        "TH_1",
		FilterIDs: []string{"FLTR_1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		MaxHits:  -1,
		MinSleep: time.Duration(1 * time.Millisecond),
		Weight:   10.0}
	dm.SetThresholdProfile(thp, true)
	thp2 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH_2",
		FilterIDs: []string{"FLTR_2"},
		MaxHits:   -1,
		MinSleep:  time.Duration(1 * time.Millisecond),
		Weight:    90.0,
		Async:     true,
	}
	dm.SetThresholdProfile(thp2, true)
	ths := Thresholds{
		&Threshold{
			Tenant: thp.Tenant,
			ID:     thp.ID,
		},
		&Threshold{
			Tenant: thp2.Tenant,
			ID:     thp2.ID,
		},
	}
	for _, th := range ths {
		dm.SetThreshold(th)
	}
	var reply Thresholds
	if err := thS.V1GetThresholdsForEvent(args, &reply); err != nil {
		t.Error(err)
	}

	sort.Slice(reply, func(i, j int) bool {
		return reply[i].ID < reply[j].ID
	})
	if !reflect.DeepEqual(ths, reply) {
		t.Errorf("Expected %v,Received %v", utils.ToJSON(ths), utils.ToJSON(reply))
	}
	expID := []string{"TH_1", "TH_2"}
	var tIDs []string
	if err := thS.V1GetThresholdIDs("cgrates.org", &tIDs); err != nil {
		t.Error(err)
	}
	sort.Slice(tIDs, func(i, j int) bool {
		return tIDs[i] < tIDs[j]
	})
	if !reflect.DeepEqual(tIDs, expID) {
		t.Errorf("Expected %v,Received %v", expID, tIDs)
	}
}

func TestThresholdProcessEvent2(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(4)
	bf := new(bytes.Buffer)
	utils.Logger.SetSyslog(nil)

	log.SetOutput(bf)
	tmpDm := dm
	defer func() {
		SetDataStorage(tmpDm)
		utils.Logger.SetLogLevel(0)
		log.SetOutput(os.Stderr)
	}()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	cdrStor := NewInternalDB(nil, nil, false, cfg.StorDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	tPrfl := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH2",
		FilterIDs: []string{"*string:Account:1001", "*gt:Balance:1000"},
		ActionIDs: []string{"ACT_1"},
		Async:     true,
		MaxHits:   -1,
	}
	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "THD_ACNT_1002",
		Hits:   1,
		tPrfl:  tPrfl,
	}
	args := &ArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testRPCMethodsProcessCDR",
			Event: map[string]any{
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "testRPCMethodsProcessCDR",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Subject:     "ANY2CNT",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       10 * time.Minute,
			},
		},
	}
	SetDataStorage(dm)
	as := Actions{
		&Action{
			ActionType: utils.MetaCDRAccount,
			Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY),
				Value: &utils.ValueFormula{Static: 10}},
		},
	}
	cdr := &CDR{
		Tenant:  "cgrates.org",
		Account: "1001",
		ToR:     utils.VOICE,
		Subject: "ANY2CNT",
		CostDetails: &EventCost{
			Cost: utils.Float64Pointer(10),
			AccountSummary: &AccountSummary{
				Tenant: "cgrates.org",
				ID:     "1001",
				BalanceSummaries: []*BalanceSummary{
					{ID: "voice2", Type: utils.VOICE, Value: 10, Disabled: false},
				},
				AllowNegative: true,
				Disabled:      false,
			},
		},
	}
	if err := cdrStor.SetCDR(cdr, true); err != nil {
		t.Error(err)
	}
	dm.SetActions("ACT_1", as, utils.NonTransactional)
	acc := &Account{
		ID:         utils.ConcatenatedKey("cgrates.org", "1001"),
		BalanceMap: map[string]Balances{utils.MONETARY: {&Balance{Value: 10, Weight: 10}}},
	}
	dm.SetAccount(acc)
	SetCdrStorage(cdrStor)
	if err := th.ProcessEvent(args, dm); err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Millisecond)
	if val := bf.String(); utils.EmptyString != val {
		t.Errorf("Buffer %v", val)
	}
	expAcc := &Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]Balances{
			utils.VOICE: {
				&Balance{
					Uuid:  "1c5b18f6-8afc-4659-b731-acf0f76fa691",
					ID:    "voice2",
					Value: 10,
				},
			},
		},
	}
	if val, err := dm.GetAccount("cgrates.org:1001"); err != nil {
		t.Error(err)
	} else if val.BalanceMap[utils.VOICE][0].ID != expAcc.BalanceMap[utils.VOICE][0].ID {
		t.Errorf("Expected %v,Received %v", utils.ToJSON(val.BalanceMap[utils.VOICE]), utils.ToJSON(expAcc.BalanceMap[utils.VOICE]))
	}
}

func TestTHSReload(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = 10 * time.Millisecond
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	ts, err := NewThresholdService(dm, cfg, nil)
	if err != nil {
		t.Error(err)
	}
	go func() {
		time.Sleep(5 * time.Millisecond)
		ts.loopStoped <- struct{}{}
	}()
	ts.Reload()
}

func TestThSStoreThreshold(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	ts, err := NewThresholdService(dm, cfg, nil)
	if err != nil {
		t.Error(err)
	}
	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "THD_ACNT_1002",
		Hits:   1,
		dirty:  utils.BoolPointer(true),
	}
	if err := ts.StoreThreshold(th); err != nil {
		t.Error(err)
	}
	if _, err := ts.dm.GetThreshold("cgrates.org", "THD_ACNT_1002", false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
}

func TestV1GetThreshold(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	thS, err := NewThresholdService(dm, cfg, nil)
	if err != nil {
		t.Error(err)
	}
	dm.SetThreshold(&Threshold{
		Tenant: "cgrates.org",
		ID:     "TH_3",
		Hits:   0,
	})
	var reply Threshold
	tntID := &utils.TenantID{
		Tenant: "cgrates.org",
		ID:     "TH_3",
	}
	if err := thS.V1GetThreshold(tntID, &reply); err != nil {
		t.Error(err)
	}
}

func TestV1ProcessEventErrs(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	Cache.Clear(nil)
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	testCases := []struct {
		name string
		args *ArgsProcessEvent
	}{
		{
			name: "CGREvent missing",
			args: &ArgsProcessEvent{},
		},
		{
			name: "Missing struct fields",
			args: &ArgsProcessEvent{
				CGREvent: &utils.CGREvent{},
			},
		},
		{
			name: "Missing Event",
			args: &ArgsProcessEvent{
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "ThresholdProcessEvent",
				},
			},
		},
		{
			name: "Failed Processing Event",
			args: &ArgsProcessEvent{
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "ProcessEvent",
					Event: map[string]any{
						utils.EventType:     utils.AccountUpdate,
						utils.Account:       "1002",
						utils.AllowNegative: true,
						utils.Disabled:      false,
						utils.Units:         12.3,
					},
				},
			},
		},
	}

	thS, err := NewThresholdService(dm, cfg, NewFilterS(cfg, nil, dm))
	if err != nil {
		t.Error(err)
	}
	var reply []string
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := thS.V1ProcessEvent(tc.args, &reply); err == nil {
				t.Error("Expected error, got nil")
			}
		})
	}
}

func TestThSProcessEventMaxHits(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	tmpDm := dm
	defer func() {
		dm = tmpDm
	}()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	thS, err := NewThresholdService(dm, cfg, NewFilterS(cfg, nil, dm))
	if err != nil {
		t.Error(err)
	}
	dm.SetFilter(&Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_2",
		Rules: []*FilterRule{
			{
				Type:    "*string",
				Element: "~*req.Category",
				Values:  []string{"call"},
			},
		},
	})
	dm.SetThresholdProfile(&ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH_2",
		FilterIDs: []string{"FLTR_2"},
		MaxHits:   2,
		MinSleep:  time.Duration(1 * time.Millisecond),
		Weight:    90.0,
		Async:     true,
	}, true)
	dm.SetThreshold(&Threshold{
		Tenant: "cgrates.org",
		ID:     "TH_2",
		Hits:   1,
	})
	args := &ArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ProcessEvent",
			Event: map[string]any{
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1002",
				utils.Subject:     "1001",
				utils.Destination: "1001",
			},
		},
	}
	var reply []string
	SetDataStorage(dm)
	if err := thS.V1ProcessEvent(args, &reply); err != nil {
		t.Error(err)
	}
}

func TestThresholdsProcessEvent3(t *testing.T) {
	to := Threshold{
		Snooze: time.Date(
			2030, 11, 17, 20, 34, 58, 651387237, time.UTC),
	}

	err := to.ProcessEvent(nil, nil)

	if err != nil {
		t.Error(err)
	}

	to2 := Threshold{
		Hits: 1,
		tPrfl: &ThresholdProfile{
			MinHits: 2,
		},
	}

	err = to2.ProcessEvent(nil, nil)

	if err != nil {
		t.Error(err)
	}

	to3 := Threshold{
		Hits: 2,
		tPrfl: &ThresholdProfile{
			MinHits: 2,
			MaxHits: 1,
		},
	}

	err = to3.ProcessEvent(nil, nil)

	if err != nil {
		t.Error(err)
	}
}

func TestThresholdsStoreThreshold(t *testing.T) {
	ts := ThresholdService{}

	err := ts.StoreThreshold(&Threshold{})

	if err != nil {
		t.Error(err)
	}
}

func TestThresholdsV1GetThresholdsForEvent(t *testing.T) {
	tS := ThresholdService{}

	err := tS.V1GetThresholdsForEvent(&ArgsProcessEvent{}, &Thresholds{})

	if err != nil {
		if err.Error() != "MANDATORY_IE_MISSING: [CGREvent]" {
			t.Error(err)
		}
	}

	err = tS.V1GetThresholdsForEvent(&ArgsProcessEvent{
		CGREvent: &utils.CGREvent{},
	}, &Thresholds{})

	if err != nil {
		if err.Error() != "MANDATORY_IE_MISSING: [Tenant ID]" {
			t.Error(err)
		}
	}

	err = tS.V1GetThresholdsForEvent(&ArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "test",
			ID:     "test",
		},
	}, &Thresholds{})

	if err != nil {
		if err.Error() != "MANDATORY_IE_MISSING: [Event]" {
			t.Error(err)
		}
	}
}

func TestThresholdSnoozeSleep(t *testing.T) {

	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "th_counter",
		tPrfl: &ThresholdProfile{
			MaxHits:  -1,
			MinHits:  1,
			Blocker:  true,
			Weight:   30,
			MinSleep: 3 * time.Second,
			Async:    true,
		},
	}

	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	var snoozeTime time.Time
	for i, arg := range argsGetThresholds {
		th.ProcessEvent(arg, dm)
		if i > 0 {
			if !th.Snooze.Equal(snoozeTime) {
				t.Error("expecte snooze to not change during sleep time")
			}
		} else {
			snoozeTime = th.Snooze
		}

	}

}
