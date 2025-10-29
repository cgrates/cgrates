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
	"fmt"
	"log"
	"os"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	testThresholdPrfs = []*ThresholdProfile{
		{
			Tenant:    "cgrates.org",
			ID:        "TH_1",
			FilterIDs: []string{"FLTR_TH_1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_2",
			FilterIDs: []string{"FLTR_TH_2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			MinSleep:  5 * time.Minute,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH_3",
			FilterIDs: []string{"FLTR_TH_3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   12,
			MinSleep:  5 * time.Minute,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
		},
	}
	testThresholds = Thresholds{
		&Threshold{
			Tenant: "cgrates.org",
			ID:     "TH_1",
		},
		&Threshold{
			Tenant: "cgrates.org",
			ID:     "TH_2",
		},
		&Threshold{
			Tenant: "cgrates.org",
			ID:     "TH_3",
		},
	}
	testThresholdArgs = []*utils.CGREvent{
		{
			Tenant: "cgrates.org",
			ID:     "Ev1",
			Event: map[string]any{
				"Threshold": "TH_1",
				"Weight":    "10.0",
			},
			APIOpts: map[string]any{},
		},
		{
			Tenant: "cgrates.org",
			ID:     "Ev1",
			Event: map[string]any{
				"Threshold": "TH_2",
				"Weight":    "20.0",
			},
			APIOpts: map[string]any{},
		},
		{
			Tenant: "cgrates.org",
			ID:     "Ev1",
			Event: map[string]any{
				"Threshold": "ThresholdPrefix123",
			},
			APIOpts: map[string]any{},
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

func prepareThresholdData(t *testing.T, dm *DataManager) {
	Cache.Clear(nil)
	if err := dm.SetFilter(&Filter{
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
	}, true); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetFilter(&Filter{
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
	}, true); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetFilter(&Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_TH_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Threshold",
				Values:  []string{"ThresholdPrefix"},
			},
		},
	}, true); err != nil {
		t.Fatal(err)
	}
	for _, th := range testThresholdPrfs {
		if err = dm.SetThresholdProfile(th, true); err != nil {
			t.Fatal(err)
		}
	}
	//Test each threshold profile from cache
	for _, th := range testThresholdPrfs {
		if temptTh, err := dm.GetThresholdProfile(th.Tenant,
			th.ID, true, false, utils.NonTransactional); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(th, temptTh) {
			t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
		}
	}
	for _, th := range testThresholds {
		if err = dm.SetThreshold(&Threshold{
			Tenant: th.Tenant,
			ID:     th.ID,
		}); err != nil {
			t.Fatal(err)
		}
	}
	//Test each threshold profile from cache
	for _, th := range testThresholds {
		if temptTh, err := dm.GetThreshold(th.Tenant,
			th.ID, true, false, utils.NonTransactional); err != nil {
			t.Fatal(err)
		} else {
			th.dirty = temptTh.dirty
			th.tPrfl = temptTh.tPrfl
			if !reflect.DeepEqual(th, temptTh) {
				t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
			}
		}
	}
}

func TestThresholdsCache(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dmTH := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg.ThresholdSCfg().StoreInterval = 0
	cfg.ThresholdSCfg().StringIndexedFields = nil
	cfg.ThresholdSCfg().PrefixIndexedFields = nil
	prepareThresholdData(t, dmTH)
}

func TestThresholdsmatchingThresholdsForEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dmTH := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg.ThresholdSCfg().StoreInterval = 0
	cfg.ThresholdSCfg().StringIndexedFields = nil
	cfg.ThresholdSCfg().PrefixIndexedFields = nil
	thServ := NewThresholdService(dmTH, cfg, &FilterS{dm: dmTH, cfg: cfg}, nil)
	prepareThresholdData(t, dmTH)

	thMatched, err := thServ.matchingThresholdsForEvent(testThresholdArgs[0].Tenant, testThresholdArgs[0])
	if err != nil {
		t.Fatal(err)
	}
	thMatched.unlock()
	if !reflect.DeepEqual(testThresholds[0].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", testThresholds[0].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(testThresholds[0].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", testThresholds[0].ID, thMatched[0].ID)
	} else if !reflect.DeepEqual(testThresholds[0].Hits, thMatched[0].Hits) {
		t.Errorf("Expecting: %+v, received: %+v", testThresholds[0].Hits, thMatched[0].Hits)
	}
	thMatched, err = thServ.matchingThresholdsForEvent(testThresholdArgs[1].Tenant, testThresholdArgs[1])
	if err != nil {
		t.Fatal(err)
	}
	thMatched.unlock()
	if !reflect.DeepEqual(testThresholds[1].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", testThresholds[1].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(testThresholds[1].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", testThresholds[1].ID, thMatched[0].ID)
	} else if !reflect.DeepEqual(testThresholds[1].Hits, thMatched[0].Hits) {
		t.Errorf("Expecting: %+v, received: %+v", testThresholds[1].Hits, thMatched[0].Hits)
	}
	thMatched, err = thServ.matchingThresholdsForEvent(testThresholdArgs[2].Tenant, testThresholdArgs[2])
	if err != nil {
		t.Fatal(err)
	}
	thMatched.unlock()
	if !reflect.DeepEqual(testThresholds[2].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", testThresholds[2].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(testThresholds[2].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", testThresholds[2].ID, thMatched[0].ID)
	} else if !reflect.DeepEqual(testThresholds[2].Hits, thMatched[0].Hits) {
		t.Errorf("Expecting: %+v, received: %+v", testThresholds[2].Hits, thMatched[0].Hits)
	}
}

func TestThresholdsProcessEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dmTH := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg.ThresholdSCfg().StoreInterval = 0
	cfg.ThresholdSCfg().StringIndexedFields = nil
	cfg.ThresholdSCfg().PrefixIndexedFields = nil
	thServ := NewThresholdService(dmTH, cfg, &FilterS{dm: dmTH, cfg: cfg}, nil)
	prepareThresholdData(t, dmTH)

	thIDs := []string{"TH_1"}
	if thMatched, err := thServ.processEvent(testThresholdArgs[0].Tenant, testThresholdArgs[0]); err != utils.ErrPartiallyExecuted {
		t.Fatal(err)
	} else if !reflect.DeepEqual(thIDs, thMatched) {
		t.Errorf("Expecting: %+v, received: %+v", thIDs, thMatched)
	}

	thIDs = []string{"TH_2"}
	if thMatched, err := thServ.processEvent(testThresholdArgs[1].Tenant, testThresholdArgs[1]); err != utils.ErrPartiallyExecuted {
		t.Fatal(err)
	} else if !reflect.DeepEqual(thIDs, thMatched) {
		t.Errorf("Expecting: %+v, received: %+v", thIDs, thMatched)
	}

	thIDs = []string{"TH_3"}
	if thMatched, err := thServ.processEvent(testThresholdArgs[2].Tenant, testThresholdArgs[2]); err != utils.ErrPartiallyExecuted {
		t.Fatal(err)
	} else if !reflect.DeepEqual(thIDs, thMatched) {
		t.Errorf("Expecting: %+v, received: %+v", thIDs, thMatched)
	}
}

func TestThresholdsVerifyIfExecuted(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dmTH := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg.ThresholdSCfg().StoreInterval = 0
	cfg.ThresholdSCfg().StringIndexedFields = nil
	cfg.ThresholdSCfg().PrefixIndexedFields = nil
	thServ := NewThresholdService(dmTH, cfg, &FilterS{dm: dmTH, cfg: cfg}, nil)
	prepareThresholdData(t, dmTH)

	thIDs := []string{"TH_1"}
	if thMatched, err := thServ.processEvent(testThresholdArgs[0].Tenant, testThresholdArgs[0]); err != utils.ErrPartiallyExecuted {
		t.Fatal(err)
	} else if !reflect.DeepEqual(thIDs, thMatched) {
		t.Errorf("Expecting: %+v, received: %+v", thIDs, thMatched)
	}

	thIDs = []string{"TH_2"}
	if thMatched, err := thServ.processEvent(testThresholdArgs[1].Tenant, testThresholdArgs[1]); err != utils.ErrPartiallyExecuted {
		t.Fatal(err)
	} else if !reflect.DeepEqual(thIDs, thMatched) {
		t.Errorf("Expecting: %+v, received: %+v", thIDs, thMatched)
	}

	thIDs = []string{"TH_3"}
	if thMatched, err := thServ.processEvent(testThresholdArgs[2].Tenant, testThresholdArgs[2]); err != utils.ErrPartiallyExecuted {
		t.Fatal(err)
	} else if !reflect.DeepEqual(thIDs, thMatched) {
		t.Errorf("Expecting: %+v, received: %+v", thIDs, thMatched)
	}
	thMatched, err := thServ.matchingThresholdsForEvent(testThresholdArgs[0].Tenant, testThresholdArgs[0])
	if err != nil {
		t.Fatal(err)
	}
	thMatched.unlock()
	if !reflect.DeepEqual(testThresholds[0].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", testThresholds[0].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(testThresholds[0].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", testThresholds[0].ID, thMatched[0].ID)
	} else if thMatched[0].Hits != 1 {
		t.Errorf("Expecting: 1, received: %+v", thMatched[0].Hits)
	}

	thMatched, err = thServ.matchingThresholdsForEvent(testThresholdArgs[1].Tenant, testThresholdArgs[1])
	if err != nil {
		t.Fatal(err)
	}
	thMatched.unlock()
	if !reflect.DeepEqual(testThresholds[1].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", testThresholds[1].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(testThresholds[1].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", testThresholds[1].ID, thMatched[0].ID)
	} else if thMatched[0].Hits != 1 {
		t.Errorf("Expecting: 1, received: %+v", thMatched[0].Hits)
	}

	thMatched, err = thServ.matchingThresholdsForEvent(testThresholdArgs[2].Tenant, testThresholdArgs[2])
	if err != nil {
		t.Fatal(err)
	}
	thMatched.unlock()
	if !reflect.DeepEqual(testThresholds[2].Tenant, thMatched[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", testThresholds[2].Tenant, thMatched[0].Tenant)
	} else if !reflect.DeepEqual(testThresholds[2].ID, thMatched[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", testThresholds[2].ID, thMatched[0].ID)
	} else if thMatched[0].Hits != 1 {
		t.Errorf("Expecting: 1, received: %+v", thMatched[0].Hits)
	}
}

func TestThresholdsProcessEvent2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dmTH := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg.ThresholdSCfg().StoreInterval = 0
	cfg.ThresholdSCfg().StringIndexedFields = nil
	cfg.ThresholdSCfg().PrefixIndexedFields = nil
	thServ := NewThresholdService(dmTH, cfg, &FilterS{dm: dmTH, cfg: cfg}, nil)
	prepareThresholdData(t, dmTH)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH_4",
		FilterIDs: []string{"FLTR_TH_1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		MaxHits:   12,
		Weight:    20.0,
		ActionIDs: []string{"ACT_1", "ACT_2"},
	}
	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH_4",
	}
	testThresholdArgs[0].APIOpts[utils.OptsThresholdsProfileIDs] = []string{"TH_1", "TH_2", "TH_3", "TH_4"}
	ev := testThresholdArgs[0]
	if err = dmTH.SetThresholdProfile(thPrf, true); err != nil {
		t.Fatal(err)
	}
	if temptTh, err := dmTH.GetThresholdProfile(thPrf.Tenant,
		thPrf.ID, true, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(thPrf, temptTh) {
		t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
	}
	if err = dmTH.SetThreshold(th); err != nil {
		t.Fatal(err)
	}
	if temptTh, err := dmTH.GetThreshold(th.Tenant,
		th.ID, true, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(th, temptTh) {
		t.Errorf("Expecting: %+v, received: %+v", th, temptTh)
	}
	thIDs := []string{"TH_1", "TH_4"}
	thIDsRev := []string{"TH_4", "TH_1"}
	if thMatched, err := thServ.processEvent(ev.Tenant, ev); err != utils.ErrPartiallyExecuted {
		t.Fatal(err)
	} else if !reflect.DeepEqual(thIDs, thMatched) && !reflect.DeepEqual(thIDsRev, thMatched) {
		t.Errorf("Expecting: %+v, received: %+v", thIDs, thMatched)
	}
	thMatched, err := thServ.matchingThresholdsForEvent(ev.Tenant, ev)
	if err != nil {
		t.Fatalf("Error: %+v", err)
	}
	thMatched.unlock()
	for _, thM := range thMatched {
		if !reflect.DeepEqual(thPrf.Tenant, thM.Tenant) {
			t.Errorf("Expecting: %+v, received: %+v", thPrf.Tenant, thM.Tenant)
		} else if reflect.DeepEqual(thIDs[0], thM.ID) && thM.Hits != 1 {
			t.Errorf("Expecting: 1 for %+v, received: %+v", thM.ID, thM.Hits)
		} else if reflect.DeepEqual(thIDs[1], thM.ID) && thM.Hits != 1 {
			t.Errorf("Expecting: 1 for %+v, received: %+v", thM.ID, thM.Hits)
		}
	}
}

func TestThresholdsUpdateThreshold(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	idb, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(idb, cfg.CacheCfg(), nil)
	thp := &ThresholdProfile{
		Tenant: "cgrates.org",
		ID:     "THUP1",
	}
	th := &Threshold{
		Tenant: thp.Tenant,
		ID:     thp.ID,
		Hits:   5,
		Snooze: time.Now(),
	}
	expTh := &Threshold{
		Tenant: thp.Tenant,
		ID:     thp.ID,
	}

	if err := dm.SetThresholdProfile(thp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetThreshold(thp.Tenant, thp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.RemoveThreshold(th.Tenant, th.ID); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetThresholdProfile(thp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetThreshold(thp.Tenant, thp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.SetThreshold(th); err != nil {
		t.Fatal(err)
	}
	thp = &ThresholdProfile{
		Tenant:  "cgrates.org",
		ID:      "THUP1",
		MaxHits: 1,
	}

	if err := dm.SetThresholdProfile(thp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetThreshold(thp.Tenant, thp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.SetThreshold(th); err != nil {
		t.Fatal(err)
	}
	thp = &ThresholdProfile{
		Tenant:  "cgrates.org",
		ID:      "THUP1",
		MaxHits: 1,
		MinHits: 1,
	}

	if err := dm.SetThresholdProfile(thp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetThreshold(thp.Tenant, thp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}

	if err := dm.SetThreshold(th); err != nil {
		t.Fatal(err)
	}
	thp = &ThresholdProfile{
		Tenant:   "cgrates.org",
		ID:       "THUP1",
		MaxHits:  1,
		MinHits:  1,
		MinSleep: 1,
	}

	if err := dm.SetThresholdProfile(thp, true); err != nil {
		t.Fatal(err)
	}

	if th, err := dm.GetThreshold(thp.Tenant, thp.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expTh, th) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expTh), utils.ToJSON(th))
	}
	if err := dm.RemoveThresholdProfile(thp.Tenant, thp.ID, true); err != nil {
		t.Fatal(err)
	}
	if _, err := dm.GetThreshold(thp.Tenant, thp.ID, false, false, utils.NonTransactional); err != utils.ErrNotFound {
		t.Fatal(err)
	}
}

func TestThresholdsShutdown(t *testing.T) {
	utils.Logger.SetLogLevel(6)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	cfg := config.NewDefaultCGRConfig()
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	tS := NewThresholdService(dm, cfg, nil, nil)

	expLog1 := `[INFO] <ThresholdS> shutdown initialized`
	expLog2 := `[INFO] <ThresholdS> shutdown complete`
	tS.Shutdown()

	if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog1) ||
		!strings.Contains(rcvLog, expLog2) {
		t.Errorf("expected logs <%+v> and <%+v> \n to be included in <%+v>",
			expLog1, expLog2, rcvLog)
	}
	utils.Logger.SetLogLevel(0)
}

func TestThresholdsStoreThresholdsOK(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	tS := NewThresholdService(dm, cfg, nil, nil)

	exp := &Threshold{
		dirty:  utils.BoolPointer(false),
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
	}
	Cache.SetWithoutReplicate(utils.CacheThresholds, "cgrates.org:TH1", exp, nil, true,
		utils.NonTransactional)
	tS.storedTdIDs.Add("cgrates.org:TH1")
	tS.storeThresholds()

	if rcv, err := tS.dm.GetThreshold("cgrates.org", "TH1", true, false,
		utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(rcv))
	}

	Cache.Remove(utils.CacheThresholds, "cgrates.org:TH1", true, utils.NonTransactional)
}

func TestThresholdsStoreThresholdsStoreThErr(t *testing.T) {

	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	cfg := config.NewDefaultCGRConfig()
	tS := NewThresholdService(nil, cfg, nil, nil)

	value := &Threshold{
		dirty:  utils.BoolPointer(true),
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
	}

	Cache.SetWithoutReplicate(utils.CacheThresholds, "TH1", value, nil, true,
		utils.NonTransactional)
	tS.storedTdIDs.Add("TH1")
	exp := utils.StringSet{
		"TH1": struct{}{},
	}
	expLog := `[WARNING] <ThresholdS> failed saving Threshold with tenant: cgrates.org and ID: TH1, error: NO_DATABASE_CONNECTION`
	tS.storeThresholds()

	if !reflect.DeepEqual(tS.storedTdIDs, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, tS.storedTdIDs)
	}
	if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected log <%+v>\n to be in included in: <%+v>", expLog, rcvLog)
	}

	utils.Logger.SetLogLevel(0)
	Cache.Remove(utils.CacheThresholds, "TH1", true, utils.NonTransactional)
}

func TestThresholdsStoreThresholdsCacheGetErr(t *testing.T) {

	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	cfg := config.NewDefaultCGRConfig()
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	tS := NewThresholdService(dm, cfg, nil, nil)

	value := &Threshold{
		dirty:  utils.BoolPointer(true),
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
	}

	Cache.SetWithoutReplicate(utils.CacheThresholds, "TH2", value, nil, true,
		utils.NonTransactional)
	tS.storedTdIDs.Add("TH1")
	expLog := `[WARNING] <ThresholdS> failed retrieving from cache threshold with ID: TH1`
	tS.storeThresholds()

	if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected <%+v> \nto be included in: <%+v>", expLog, rcvLog)
	}

	utils.Logger.SetLogLevel(0)
	Cache.Remove(utils.CacheThresholds, "TH2", true, utils.NonTransactional)
}

func TestThresholdsStoreThresholdNilDirtyField(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	tS := NewThresholdService(dm, cfg, nil, nil)

	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
	}

	if err := tS.StoreThreshold(th); err != nil {
		t.Error(err)
	}
}

func TestThresholdsProcessEventOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ExpiryTime: time.Date(2021, 6, 1, 12, 0, 0, 0, time.UTC),
		},
		MinHits: 2,
		MaxHits: 5,
		Weight:  10,
		Blocker: true,
	}
	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
		tPrfl:  thPrf,
	}

	if err := dm.SetThresholdProfile(thPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetThreshold(th); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ThdProcessEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{"TH1"},
		},
	}

	expIDs := []string{"TH1"}
	expStored := utils.StringSet{
		"cgrates.org:TH1": struct{}{},
	}
	if rcvIDs, err := tS.processEvent(args.Tenant, args); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDs, expIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, rcvIDs)
	} else if !reflect.DeepEqual(tS.storedTdIDs, expStored) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expStored, tS.storedTdIDs)
	}

}

func TestThresholdsProcessEventStoreThOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = -1
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH2",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}

	if err := dm.SetThresholdProfile(thPrf, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ThdProcessEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{"TH2"},
		},
	}
	exp := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH2",
		Hits:   1,
	}
	expIDs := []string{"TH2"}
	if rcvIDs, err := tS.processEvent(args.Tenant, args); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvIDs, expIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, rcvIDs)
	} else if rcv, err := tS.dm.GetThreshold("cgrates.org", "TH2", true, false,
		utils.NonTransactional); err != nil {
		t.Error(err)
	} else {
		rcv.tPrfl = nil
		rcv.dirty = nil
		rcv.Snooze = time.Time{}
		if !reflect.DeepEqual(rcv, exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
		}
	}

}

func TestThresholdsProcessEventMaxHitsDMErr(t *testing.T) {
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	tmp := config.CgrConfig()
	tmpCache := Cache
	tmpCMgr := connMgr
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()["test"] = &config.RPCConn{Conns: []*config.RemoteHost{{}}}
	cfg.CacheCfg().ReplicationConns = []string{"test"}
	cfg.CacheCfg().Partitions[utils.CacheThresholds].Replicate = true
	config.SetCgrConfig(cfg)
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	connMgr = NewConnManager(cfg, make(map[string]chan birpc.ClientConnector))
	dm := NewDataManager(data, cfg.CacheCfg(), connMgr)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(nil, cfg, filterS, nil)
	Cache = NewCacheS(cfg, dm, nil)

	defer func() {
		connMgr = tmpCMgr
		config.SetCgrConfig(tmp)
		log.SetOutput(os.Stderr)
		Cache = tmpCache
	}()
	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH3",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ExpiryTime: time.Date(2021, 6, 1, 12, 0, 0, 0, time.UTC),
		},
		MinHits: 2,
		MaxHits: 5,
		Weight:  10,
		Blocker: true,
	}
	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH3",
		Hits:   4,
		tPrfl:  thPrf,
	}

	if err := dm.SetThresholdProfile(thPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetThreshold(th); err != nil {
		t.Error(err)
	}
	Cache.SetWithoutReplicate(utils.CacheThresholdProfiles, thPrf.TenantID(), thPrf, nil, true, utils.NonTransactional)
	Cache.SetWithoutReplicate(utils.CacheThresholds, thPrf.TenantID(), th, nil, true, utils.NonTransactional)

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ThdProcessEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{"TH3"},
		},
	}

	if _, err := tS.processEvent(args.Tenant, args); err != nil {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ErrPartiallyExecuted, err)
	}

	utils.Logger.SetLogLevel(0)
}

func TestThresholdsProcessEventNotFound(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH5",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ExpiryTime: time.Date(2021, 6, 1, 12, 0, 0, 0, time.UTC),
		},
		MinHits: 2,
		MaxHits: 5,
		Weight:  10,
		Blocker: true,
	}
	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH5",
		Hits:   2,
		tPrfl:  thPrf,
	}

	if err := dm.SetThresholdProfile(thPrf, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetThreshold(th); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ThdProcessEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{"TH6"},
		},
	}

	if _, err := tS.processEvent(args.Tenant, args); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

}

func TestThresholdsV1ProcessEventOK(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache.Clear(nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf1 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}
	if err := dm.SetThresholdProfile(thPrf1, true); err != nil {
		t.Error(err)
	}

	thPrf2 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH2",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MaxHits:   7,
		Weight:    20,
	}
	if err := dm.SetThresholdProfile(thPrf2, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		ID: "V1ProcessEventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}
	var reply []string
	exp := []string{"TH1", "TH2"}
	if err := tS.V1ProcessEvent(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(reply)
		if !reflect.DeepEqual(reply, exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, reply)
		}
	}
}

func TestThresholdsV1ProcessEventPartExecErr(t *testing.T) {

	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	cfg := config.NewDefaultCGRConfig()
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache.Clear(nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf1 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH3",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}
	if err := dm.SetThresholdProfile(thPrf1, true); err != nil {
		t.Error(err)
	}

	thPrf2 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH4",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ActionIDs: []string{"ACT1"},
		MaxHits:   7,
		Weight:    20,
	}
	if err := dm.SetThresholdProfile(thPrf2, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		ID: "V1ProcessEventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}
	expLog1 := `[ERROR] Failed to get actions for ACT1: NOT_FOUND`
	expLog2 := `[WARNING] <ThresholdS> failed executing actions: ACT1, error: NOT_FOUND`
	var reply []string
	if err := tS.V1ProcessEvent(context.Background(), args, &reply); err == nil ||
		err != utils.ErrPartiallyExecuted {
		t.Errorf("expected: <%+v>,\nreceived: <%+v>", utils.ErrPartiallyExecuted, err)
	} else {
		if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog1) ||
			!strings.Contains(rcvLog, expLog2) {
			t.Errorf("expected logs <%+v> and <%+v> to be included in: <%+v>",
				expLog1, expLog2, rcvLog)
		}
	}

	utils.Logger.SetLogLevel(0)
}

func TestThresholdsV1ProcessEventMissingArgs(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache.Clear(nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf1 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}
	if err := dm.SetThresholdProfile(thPrf1, true); err != nil {
		t.Error(err)
	}

	thPrf2 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH2",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MaxHits:   7,
		Weight:    20,
	}
	if err := dm.SetThresholdProfile(thPrf2, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		ID: "V1ProcessEventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}

	args = &utils.CGREvent{
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}
	var reply []string
	experr := `MANDATORY_IE_MISSING: [ID]`
	if err := tS.V1ProcessEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	experr = `MANDATORY_IE_MISSING: [CGREvent]`
	if err := tS.V1ProcessEvent(context.Background(), nil, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = &utils.CGREvent{
		ID:    "V1ProcessEventTest",
		Event: nil,
	}
	experr = `MANDATORY_IE_MISSING: [Event]`
	if err := tS.V1ProcessEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestThresholdsV1GetThresholdOK(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache.Clear(nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}
	if err := dm.SetThresholdProfile(thPrf, true); err != nil {
		t.Error(err)
	}

	expTh := Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
	}
	var rplyTh Threshold
	if err := tS.V1GetThreshold(context.Background(),
		&utils.TenantID{
			ID: "TH1",
		}, &rplyTh); err != nil {
		t.Error(err)
	} else {
		var snooze time.Time
		rplyTh.dirty = nil
		rplyTh.Snooze = snooze
		if !reflect.DeepEqual(rplyTh, expTh) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expTh), utils.ToJSON(rplyTh))
		}
	}
}

func TestThresholdsV1GetThresholdNotFoundErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache.Clear(nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}
	if err := dm.SetThresholdProfile(thPrf, true); err != nil {
		t.Error(err)
	}

	var rplyTh Threshold
	if err := tS.V1GetThreshold(context.Background(),
		&utils.TenantID{
			ID: "TH2",
		}, &rplyTh); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestThresholdMatchingThresholdForEventLocks2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	Cache.Clear(nil)
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(db, config.CgrConfig().CacheCfg(), nil)
	cfg.ThresholdSCfg().StoreInterval = 1
	cfg.ThresholdSCfg().StringIndexedFields = nil
	cfg.ThresholdSCfg().PrefixIndexedFields = nil
	rS := NewThresholdService(dm, cfg,
		&FilterS{dm: dm, cfg: cfg}, nil)

	prfs := make([]*ThresholdProfile, 0)
	ids := utils.StringSet{}
	for i := 0; i < 10; i++ {
		rPrf := &ThresholdProfile{
			Tenant:  "cgrates.org",
			ID:      fmt.Sprintf("TH%d", i),
			Weight:  20.00,
			MaxHits: 5,
		}
		dm.SetThresholdProfile(rPrf, true)
		prfs = append(prfs, rPrf)
		ids.Add(rPrf.ID)
	}
	rPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH20",
		FilterIDs: []string{"FLTR_RES_201"},
		Weight:    20.00,
		MaxHits:   5,
	}
	err = db.SetThresholdProfileDrv(rPrf)
	if err != nil {
		t.Fatal(err)
	}
	prfs = append(prfs, rPrf)
	ids.Add(rPrf.ID)
	_, err := rS.matchingThresholdsForEvent("cgrates.org", &utils.CGREvent{
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: ids.AsSlice(),
		},
	})
	expErr := utils.ErrPrefixNotFound(rPrf.FilterIDs[0])
	if err == nil || err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s ,received: %+v", expErr, err)
	}
	for _, rPrf := range prfs {
		if rPrf.isLocked() {
			t.Fatalf("Expected profile to not be locked %q", rPrf.ID)
		}
		if rPrf.ID == "TH20" {
			continue
		}
		if r, err := dm.GetThreshold(rPrf.Tenant, rPrf.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("error %s for <%s>", err, rPrf.ID)
		} else if r.isLocked() {
			t.Fatalf("Expected Threshold to not be locked %q", rPrf.ID)
		}
	}
}

func TestThresholdMatchingThresholdForEventLocksActivationInterval(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	Cache.Clear(nil)
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(db, config.CgrConfig().CacheCfg(), nil)
	cfg.ThresholdSCfg().StoreInterval = 1
	cfg.ThresholdSCfg().StringIndexedFields = nil
	cfg.ThresholdSCfg().PrefixIndexedFields = nil
	rS := NewThresholdService(dm, cfg,
		&FilterS{dm: dm, cfg: cfg}, nil)

	ids := utils.StringSet{}
	for i := 0; i < 10; i++ {
		rPrf := &ThresholdProfile{
			Tenant:  "cgrates.org",
			ID:      fmt.Sprintf("TH%d", i),
			MaxHits: 5,
			Weight:  20.00,
		}
		dm.SetThresholdProfile(rPrf, true)
		ids.Add(rPrf.ID)
	}
	rPrf := &ThresholdProfile{
		Tenant:  "cgrates.org",
		ID:      "TH21",
		MaxHits: 5,
		Weight:  20.00,
		ActivationInterval: &utils.ActivationInterval{
			ExpiryTime: time.Now().Add(-5 * time.Second),
		},
	}
	dm.SetThresholdProfile(rPrf, true)
	ids.Add(rPrf.ID)
	mres, err := rS.matchingThresholdsForEvent("cgrates.org", &utils.CGREvent{
		Time: utils.TimePointer(time.Now()),
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: ids.AsSlice(),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer mres.unlock()
	if rPrf.isLocked() {
		t.Fatalf("Expected profile to not be locked %q", rPrf.ID)
	}
	if r, err := dm.GetThreshold(rPrf.Tenant, rPrf.ID, true, false, utils.NonTransactional); err != nil {
		t.Errorf("error %s for <%s>", err, rPrf.ID)
	} else if r.isLocked() {
		t.Fatalf("Expected Threshold to not be locked %q", rPrf.ID)
	}
}

func TestThresholdMatchingThresholdForEventLocks3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	prfs := make([]*ThresholdProfile, 0)
	Cache.Clear(nil)
	db := &DataDBMock{
		GetThresholdProfileDrvF: func(tnt, id string) (*ThresholdProfile, error) {
			if id == "TH1" {
				return nil, utils.ErrNotImplemented
			}
			rPrf := &ThresholdProfile{
				Tenant:  "cgrates.org",
				ID:      id,
				MaxHits: 5,
				Weight:  20.00,
			}
			Cache.Set(utils.CacheThresholds, rPrf.TenantID(), &Threshold{
				Tenant: rPrf.Tenant,
				ID:     rPrf.ID,
			}, nil, true, utils.NonTransactional)
			prfs = append(prfs, rPrf)
			return rPrf, nil
		},
	}
	dm := NewDataManager(db, config.CgrConfig().CacheCfg(), nil)
	cfg.ThresholdSCfg().StoreInterval = 1
	cfg.ThresholdSCfg().StringIndexedFields = nil
	cfg.ThresholdSCfg().PrefixIndexedFields = nil
	rS := NewThresholdService(dm, cfg,
		&FilterS{dm: dm, cfg: cfg}, nil)

	ids := utils.StringSet{}
	for i := 0; i < 10; i++ {
		ids.Add(fmt.Sprintf("TH%d", i))
	}
	_, err := rS.matchingThresholdsForEvent("cgrates.org", &utils.CGREvent{
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: ids.AsSlice(),
		},
	})
	if err != utils.ErrNotImplemented {
		t.Fatalf("Error: %+v", err)
	}
	for _, rPrf := range prfs {
		if rPrf.isLocked() {
			t.Fatalf("Expected profile to not be locked %q", rPrf.ID)
		}
	}

}

func TestThresholdMatchingThresholdForEventLocks5(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpC := config.CgrConfig()
	defer func() {
		config.SetCgrConfig(tmpC)
	}()
	Cache.Clear(nil)
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(db, config.CgrConfig().CacheCfg(), NewConnManager(cfg, make(map[string]chan birpc.ClientConnector)))
	cfg.ThresholdSCfg().StoreInterval = 1
	cfg.ThresholdSCfg().StringIndexedFields = nil
	cfg.ThresholdSCfg().PrefixIndexedFields = nil
	cfg.RPCConns()["test"] = &config.RPCConn{Conns: []*config.RemoteHost{{}}}
	cfg.DataDbCfg().RmtConns = []string{"test"}
	cfg.DataDbCfg().Items[utils.CacheThresholds].Remote = true
	config.SetCgrConfig(cfg)
	rS := NewThresholdService(dm, cfg,
		&FilterS{dm: dm, cfg: cfg}, nil)

	prfs := make([]*ThresholdProfile, 0)
	ids := utils.StringSet{}
	for i := 0; i < 10; i++ {
		rPrf := &ThresholdProfile{
			Tenant:  "cgrates.org",
			ID:      fmt.Sprintf("TH%d", i),
			Weight:  20.00,
			MaxHits: 5,
		}
		dm.SetThresholdProfile(rPrf, true)
		prfs = append(prfs, rPrf)
		ids.Add(rPrf.ID)
	}
	dm.RemoveThreshold("cgrates.org", "TH1")
	_, err := rS.matchingThresholdsForEvent("cgrates.org", &utils.CGREvent{
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: ids.AsSlice(),
		},
	})
	if err != utils.ErrDisconnected {
		t.Fatal(err)
	}
	for _, rPrf := range prfs {
		if rPrf.isLocked() {
			t.Fatalf("Expected profile to not be locked %q", rPrf.ID)
		}
		if rPrf.ID == "TH1" {
			continue
		}
		if r, err := dm.GetThreshold(rPrf.Tenant, rPrf.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("error %s for <%s>", err, rPrf.ID)
		} else if r.isLocked() {
			t.Fatalf("Expected Threshold to not be locked %q", rPrf.ID)
		}
	}
}

func TestThresholdsRunBackupStoreIntervalLessThanZero(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = -1
	tS := &ThresholdService{
		cgrcfg:      cfg,
		loopStopped: make(chan struct{}, 1),
	}

	tS.runBackup()
	select {
	case <-tS.loopStopped:
	case <-time.After(time.Second):
		t.Error("timed out waiting for loop to stop")
	}
}

func TestThresholdsRunBackupStop(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = 5 * time.Millisecond
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	tnt := "cgrates.org"
	thID := "Th1"
	tS := &ThresholdService{
		dm: dm,
		storedTdIDs: utils.StringSet{
			thID: struct{}{},
		},
		cgrcfg:      cfg,
		loopStopped: make(chan struct{}, 1),
		stopBackup:  make(chan struct{}),
	}
	value := &Threshold{
		dirty:  utils.BoolPointer(true),
		Tenant: tnt,
		ID:     thID,
	}
	Cache.SetWithoutReplicate(utils.CacheThresholds, thID, value, nil, true, "")

	// Backup loop checks for the state of the stopBackup
	// channel after storing the threshold. Channel can be
	// safely closed beforehand.
	close(tS.stopBackup)
	tS.runBackup()

	want := &Threshold{
		dirty:  utils.BoolPointer(false),
		Tenant: tnt,
		ID:     thID,
	}
	if got, err := tS.dm.GetThreshold(tnt, thID, true, false, utils.NonTransactional); err != nil {
		t.Errorf("dm.GetThreshold(%q,%q): got unexpected err=%v", tnt, thID, err)
	} else if !reflect.DeepEqual(got, want) {
		t.Errorf("dm.GetThreshold(%q,%q) = %v, want %v", tnt, thID, got, want)
	}

	select {
	case <-tS.loopStopped:
	case <-time.After(time.Second):
		t.Error("timed out waiting for loop to stop")
	}
}

func TestThresholdsReload(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = 5 * time.Millisecond
	tS := &ThresholdService{
		stopBackup:  make(chan struct{}),
		loopStopped: make(chan struct{}, 1),
		cgrcfg:      cfg,
	}
	tS.loopStopped <- struct{}{}
	tS.Reload()
	close(tS.stopBackup)
	select {
	case <-tS.loopStopped:
	case <-time.After(time.Second):
		t.Error("timed out waiting for loop to stop")
	}
}

func TestThresholdsStartLoop(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = -1
	tS := &ThresholdService{
		loopStopped: make(chan struct{}, 1),
		cgrcfg:      cfg,
	}

	tS.StartLoop()
	select {
	case <-tS.loopStopped:
	case <-time.After(time.Second):
		t.Error("timed out waiting for loop to stop")
	}
}

func TestThresholdsV1GetThresholdsForEventOK(t *testing.T) {
	tmpC := config.CgrConfig()
	defer func() {
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache.Clear(nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}
	if err := dm.SetThresholdProfile(thPrf, true); err != nil {
		t.Error(err)
	}
	args := &utils.CGREvent{
		ID: "TestGetThresholdsForEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{},
		},
	}

	exp := Thresholds{
		{
			Tenant: "cgrates.org",
			ID:     "TH1",
			tPrfl:  thPrf,
			dirty:  utils.BoolPointer(false),
		},
	}
	var reply Thresholds
	if err := tS.V1GetThresholdsForEvent(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, reply)
	}
}

func TestThresholdsV1GetThresholdsForEventMissingArgs(t *testing.T) {
	tmpC := config.CgrConfig()
	defer func() {
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache.Clear(nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}
	if err := dm.SetThresholdProfile(thPrf, true); err != nil {
		t.Error(err)
	}
	experr := `MANDATORY_IE_MISSING: [CGREvent]`
	var reply Thresholds
	if err := tS.V1GetThresholdsForEvent(context.Background(), nil, &reply); err == nil ||
		err.Error() != experr {
		t.Error(err)
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{},
		},
	}

	experr = `MANDATORY_IE_MISSING: [ID]`
	if err := tS.V1GetThresholdsForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Error(err)
	}

	args = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestGetThresholdsForEvent",
		Event:  nil,
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{},
		},
	}

	experr = `MANDATORY_IE_MISSING: [Event]`
	if err := tS.V1GetThresholdsForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Error(err)
	}
}

func TestThresholdsV1GetThresholdIDsOK(t *testing.T) {
	tmpC := config.CgrConfig()
	defer func() {
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache.Clear(nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf1 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}
	if err := dm.SetThresholdProfile(thPrf1, true); err != nil {
		t.Error(err)
	}

	thPrf2 := &ThresholdProfile{
		Tenant:  "cgrates.org",
		ID:      "TH2",
		MinHits: 0,
		MaxHits: 7,
		Weight:  20,
		Blocker: true,
	}
	if err := dm.SetThresholdProfile(thPrf2, true); err != nil {
		t.Error(err)
	}

	expIDs := []string{"TH1", "TH2"}
	var reply []string
	if err := tS.V1GetThresholdIDs(context.Background(), "", &reply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(reply)
		if !reflect.DeepEqual(reply, expIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, reply)
		}
	}
}

func TestThresholdsV1GetThresholdIDsGetKeysForPrefixErr(t *testing.T) {
	tmpC := config.CgrConfig()
	defer func() {
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := &DataDBMock{}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache.Clear(nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	var reply []string
	if err := tS.V1GetThresholdIDs(context.Background(), "", &reply); err == nil ||
		err != utils.ErrNotImplemented {
		t.Error(err)
	}
}

func TestThresholdsV1ResetThresholdOK(t *testing.T) {
	tmpC := config.CgrConfig()
	defer func() {
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache.Clear(nil)
	defer Cache.Clear(nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}

	th := &Threshold{
		tPrfl:  thPrf,
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
		dirty:  utils.BoolPointer(false),
	}
	if err := dm.SetThreshold(th); err != nil {
		t.Error(err)
	}

	expStored := utils.StringSet{
		"cgrates.org:TH1": {},
	}
	var reply string
	if err := tS.V1ResetThreshold(context.Background(),
		&utils.TenantID{
			ID: "TH1",
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned: <%q>", reply)
	}
	if x, ok := Cache.Get(utils.CacheThresholds, "cgrates.org:TH1"); !ok {
		t.Errorf("not ok")
	} else if x.(*Threshold).Hits != 0 {
		t.Errorf("expected nr. of hits to be 0, received: <%+v>", x.(*Threshold).Hits)
	} else if !reflect.DeepEqual(tS.storedTdIDs, expStored) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expStored, tS.storedTdIDs)
	}
}

func TestThresholdsV1ResetThresholdErrNotFound(t *testing.T) {
	tmpC := config.CgrConfig()
	defer func() {
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache.Clear(nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH2",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}

	th := &Threshold{
		tPrfl:  thPrf,
		Tenant: "cgrates.org",
		ID:     "TH2",
		Hits:   2,
		dirty:  utils.BoolPointer(false),
	}
	if err := dm.SetThreshold(th); err != nil {
		t.Error(err)
	}

	var reply string
	if err := tS.V1ResetThreshold(context.Background(),
		&utils.TenantID{
			ID: "TH1",
		}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestThresholdsV1ResetThresholdNegativeStoreIntervalOK(t *testing.T) {
	tmpC := config.CgrConfig()
	defer func() {
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = -1
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache.Clear(nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}

	th := &Threshold{
		tPrfl:  thPrf,
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
		dirty:  utils.BoolPointer(false),
	}
	if err := dm.SetThreshold(th); err != nil {
		t.Error(err)
	}

	var reply string
	if err := tS.V1ResetThreshold(context.Background(),
		&utils.TenantID{
			ID: "TH1",
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned: <%q>", reply)
	}

	if nTh, err := dm.GetThreshold("cgrates.org", "TH1", false, false, ""); err != nil {
		t.Error(err)
	} else if nTh.Hits != 0 {
		t.Errorf("expected nr. of hits to be 0, received: <%+v>", nTh.Hits)
	}
}

func TestThresholdsV1ResetThresholdNegativeStoreIntervalErr(t *testing.T) {
	tmpC := config.CgrConfig()
	defer func() {
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().StoreInterval = -1
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache.Clear(nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(nil, cfg, filterS, nil)

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}

	th := &Threshold{
		tPrfl:  thPrf,
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
		dirty:  utils.BoolPointer(false),
	}
	if err := dm.SetThreshold(th); err != nil {
		t.Error(err)
	}

	var reply string
	if err := tS.V1ResetThreshold(context.Background(),
		&utils.TenantID{
			ID: "TH1",
		}, &reply); err == nil || err != utils.ErrNoDatabaseConn {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNoDatabaseConn, err)
	}
}

func TestThresholdsLockUnlockThresholdProfiles(t *testing.T) {
	thPrf := &ThresholdProfile{
		Tenant:  "cgrates.org",
		ID:      "TH1",
		Weight:  10,
		MaxHits: 5,
		MinHits: 2,
	}

	//lock profile with empty lkID parameter
	thPrf.lock(utils.EmptyString)

	if !thPrf.isLocked() {
		t.Fatal("expected profile to be locked")
	} else if thPrf.lkID == utils.EmptyString {
		t.Fatal("expected struct field \"lkID\" to be non-empty")
	}

	//unlock previously locked profile
	thPrf.unlock()

	if thPrf.isLocked() {
		t.Fatal("expected profile to be unlocked")
	} else if thPrf.lkID != utils.EmptyString {
		t.Fatal("expected struct field \"lkID\" to be empty")
	}

	//unlock an already unlocked profile - nothing happens
	thPrf.unlock()

	if thPrf.isLocked() {
		t.Fatal("expected profile to be unlocked")
	} else if thPrf.lkID != utils.EmptyString {
		t.Fatal("expected struct field \"lkID\" to be empty")
	}
}

func TestThresholdsLockUnlockThresholds(t *testing.T) {
	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
	}

	//lock resource with empty lkID parameter
	th.lock(utils.EmptyString)

	if !th.isLocked() {
		t.Fatal("expected resource to be locked")
	} else if th.lkID == utils.EmptyString {
		t.Fatal("expected struct field \"lkID\" to be non-empty")
	}

	//unlock previously locked resource
	th.unlock()

	if th.isLocked() {
		t.Fatal("expected resource to be unlocked")
	} else if th.lkID != utils.EmptyString {
		t.Fatal("expected struct field \"lkID\" to be empty")
	}

	//unlock an already unlocked resource - nothing happens
	th.unlock()

	if th.isLocked() {
		t.Fatal("expected resource to be unlocked")
	} else if th.lkID != utils.EmptyString {
		t.Fatal("expected struct field \"lkID\" to be empty")
	}
}

func TestThresholdsMatchingThresholdsForEventNotFoundErr(t *testing.T) {
	tmpC := config.CgrConfig()
	defer func() {
		config.SetCgrConfig(tmpC)
	}()

	cfg := config.NewDefaultCGRConfig()
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache.Clear(nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	thPrf1 := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinHits:   2,
		MaxHits:   5,
		Weight:    10,
		Blocker:   true,
	}
	if err := dm.SetThresholdProfile(thPrf1, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event:  map[string]any{},
	}

	if _, err := tS.matchingThresholdsForEvent("cgrates.org", args); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestThresholdsStoreThresholdCacheSetErr(t *testing.T) {
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	tmp := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = tmp
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{"test"}
	cfg.CacheCfg().Partitions[utils.CacheThresholds].Replicate = true
	cfg.RPCConns()["test"] = &config.RPCConn{Conns: []*config.RemoteHost{{}}}
	config.SetCgrConfig(cfg)
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	connMgr = NewConnManager(cfg, make(map[string]chan birpc.ClientConnector))
	Cache = NewCacheS(cfg, dm, nil)
	filterS := NewFilterS(cfg, nil, dm)
	tS := NewThresholdService(dm, cfg, filterS, nil)

	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
		dirty:  utils.BoolPointer(true),
	}
	Cache.SetWithoutReplicate(utils.CacheThresholds, "cgrates.org:TH1", th, nil, true,
		utils.NonTransactional)
	expLog := `[WARNING] <ThresholdService> failed caching Threshold with ID: cgrates.org:TH1, error: DISCONNECTED`
	if err := tS.StoreThreshold(th); err == nil ||
		err.Error() != utils.ErrDisconnected.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrDisconnected, err)
	} else if rcv := buf.String(); !strings.Contains(rcv, expLog) {
		t.Errorf("expected log <%+v> to be included in <%+v>", expLog, rcv)
	}

	utils.Logger.SetLogLevel(0)
}

func TestThresholdProcessEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().IndexedSelects = false
	db, _ := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
	ths := NewThresholdService(dm, cfg, fS, nil)
	thps := []*ThresholdProfile{
		{
			Tenant:    "cgrates.org",
			ID:        "TH1",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			MinHits:   3,
			MaxHits:   2,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH2",
			FilterIDs: []string{"*string:~*req.Account:1002"},
			MinHits:   2,
			MaxHits:   3,
		}, {
			Tenant:    "cgrates.org",
			ID:        "TH3",
			FilterIDs: []string{"*string:~*req.Account:1003"},
			MinHits:   1,
			MaxHits:   -1,
		},
	}
	for _, thP := range thps {
		if err := ths.dm.SetThresholdProfile(thP, false); err != nil {
			t.Error(err)
		}
	}
	tts := []struct {
		name         string
		runs         int
		cgrEvnt      map[string]any
		matchedthIDs []string
	}{
		{
			name: "MinHitsLargerThanMaxHits",
			runs: 3,
			cgrEvnt: map[string]any{
				utils.AccountField: "1001",
			},
		},
		{
			name: "MinHitsLargerThanMaxHits",
			runs: 4,
			cgrEvnt: map[string]any{
				utils.AccountField: "1002",
			},
		},
		{},
	}
	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			var thIDs []string
			for range tt.runs {
				var err error
				thIDs, err = ths.processEvent("cgrates.org", &utils.CGREvent{Event: tt.cgrEvnt})
				if err != nil {
					t.Error(err)
				}
			}
			if !slices.Equal(thIDs, tt.matchedthIDs) {
				t.Errorf("expected: %v, received: %v", tt.matchedthIDs, thIDs)
			}

		})
	}
}
