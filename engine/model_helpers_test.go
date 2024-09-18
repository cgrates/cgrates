/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or56
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
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestTpResourcesAsTpResources(t *testing.T) {
	tps := []*ResourceMdl{
		{
			Tpid:         "TEST_TPID",
			Tenant:       "cgrates.org",
			ID:           "ResGroup1",
			FilterIDs:    "FLTR_RES_GR1;*ai:~*req.AnswerTime:2014-07-29T15:00:00Z",
			Stored:       false,
			Blocker:      false,
			Weights:      ";10",
			Limit:        "45",
			ThresholdIDs: "WARN_RES1;WARN_RES1"},
		{
			Tpid:         "TEST_TPID",
			ID:           "ResGroup1",
			Tenant:       "cgrates.org",
			FilterIDs:    "FLTR_RES_GR1",
			ThresholdIDs: "WARN3"},
		{
			Tpid:      "TEST_TPID",
			Tenant:    "cgrates.org",
			ID:        "ResGroup2",
			FilterIDs: "FLTR_RES_GR2;*ai:~*req.AnswerTime:2014-07-29T15:00:00Z",
			Stored:    false,
			Blocker:   false,
			Weights:   ";10",
			Limit:     "20"},
	}
	eTPs := []*utils.TPResourceProfile{
		{
			TPid:         tps[0].Tpid,
			Tenant:       tps[0].Tenant,
			ID:           tps[0].ID,
			FilterIDs:    []string{"FLTR_RES_GR1", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
			Stored:       tps[0].Stored,
			Blocker:      tps[0].Blocker,
			Weights:      tps[0].Weights,
			Limit:        tps[0].Limit,
			ThresholdIDs: []string{"WARN_RES1", "WARN3"},
		},
		{
			TPid:      tps[2].Tpid,
			Tenant:    tps[2].Tenant,
			ID:        tps[2].ID,
			FilterIDs: []string{"FLTR_RES_GR2", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
			Stored:    tps[2].Stored,
			Blocker:   tps[2].Blocker,
			Weights:   tps[2].Weights,
			Limit:     tps[2].Limit,
		},
	}
	rcvTPs := ResourceMdls(tps).AsTPResources()
	if len(rcvTPs) != len(eTPs) {
		t.Errorf("Expecting: %+v Received: %+v", utils.ToIJSON(eTPs), utils.ToIJSON(rcvTPs))
	}
}

func TestAPItoResource(t *testing.T) {
	tpRL := &utils.TPResourceProfile{
		Tenant:            "cgrates.org",
		TPid:              "tp_test",
		ID:                "ResGroup1",
		FilterIDs:         []string{"FLTR_RES_GR_1", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
		Stored:            false,
		Blocker:           false,
		Weights:           ";10",
		Limit:             "2",
		ThresholdIDs:      []string{"TRes1"},
		AllocationMessage: "asd",
	}
	eRL := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                tpRL.ID,
		Stored:            tpRL.Stored,
		Blocker:           tpRL.Blocker,
		FilterIDs:         []string{"FLTR_RES_GR_1", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
		ThresholdIDs:      []string{"TRes1"},
		AllocationMessage: tpRL.AllocationMessage,
		Limit:             2,
	}

	if tpRL.Weights != utils.EmptyString {
		var err error
		eRL.Weights, err = utils.NewDynamicWeightsFromString(utils.IfaceAsString(tpRL.Weights), utils.InfieldSep, utils.ANDSep)
		if err != nil {
			t.Error(err)
		}
	}
	if rl, err := APItoResource(tpRL, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRL, rl) {
		t.Errorf("Expecting: %+v, received: %+v", eRL, rl)
	}
}

func TestResourceProfileToAPI(t *testing.T) {
	expected := &utils.TPResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "ResGroup1",
		FilterIDs:         []string{"FLTR_RES_GR_1", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
		Weights:           ";10",
		Limit:             "2",
		ThresholdIDs:      []string{"TRes1"},
		AllocationMessage: "asd",
	}
	rp := &ResourceProfile{
		Tenant: "cgrates.org",
		ID:     "ResGroup1",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			}},
		FilterIDs:         []string{"FLTR_RES_GR_1", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
		ThresholdIDs:      []string{"TRes1"},
		AllocationMessage: "asd",
		Limit:             2,
	}

	if rcv := ResourceProfileToAPI(rp); !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestAPItoModelResource(t *testing.T) {
	tpRL := &utils.TPResourceProfile{
		Tenant:            "cgrates.org",
		TPid:              "tp_test",
		ID:                "ResGroup1",
		FilterIDs:         []string{"*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
		Weights:           ";10",
		Limit:             "2",
		ThresholdIDs:      []string{"TRes1"},
		AllocationMessage: "test",
	}
	expModel := &ResourceMdl{
		Tpid:              "tp_test",
		Tenant:            "cgrates.org",
		ID:                "ResGroup1",
		FilterIDs:         "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z",
		Weights:           ";10",
		Limit:             "2",
		ThresholdIDs:      "TRes1",
		AllocationMessage: "test",
	}
	rcv := APItoModelResource(tpRL)
	if len(rcv) != 1 {
		t.Errorf("Expecting: 1, received: %+v", len(rcv))
	} else if !reflect.DeepEqual(rcv[0], expModel) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expModel), utils.ToJSON(rcv[0]))
	}
}

func TestTPStatsAsTPStats(t *testing.T) {
	tps := StatMdls{
		&StatMdl{
			Tpid:           "TEST_TPID",
			Tenant:         "cgrates.org",
			ID:             "Stats1",
			FilterIDs:      "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z;FLTR_1",
			QueueLength:    100,
			TTL:            "1s",
			MinItems:       2,
			MetricIDs:      "*asr;*acc;*tcc;*acd;*tcd;*pdd",
			Stored:         true,
			Blockers:       ";true",
			Weights:        ";20",
			MetricBlockers: ";false",
		},
		&StatMdl{
			Tpid:         "TEST_TPID",
			Tenant:       "cgrates.org",
			ID:           "Stats1",
			FilterIDs:    "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z;FLTR_1",
			QueueLength:  100,
			TTL:          "1s",
			MinItems:     2,
			MetricIDs:    "*sum#BalanceValue;*average#BalanceValue;*tcc",
			ThresholdIDs: "THRESH3",
			Stored:       true,
			Blockers:     ";true",
			Weights:      ";20",
		},
		&StatMdl{
			Tpid:         "TEST_TPID",
			Tenant:       "itsyscom.com",
			ID:           "Stats1",
			FilterIDs:    "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z;FLTR_1",
			QueueLength:  100,
			TTL:          "1s",
			MinItems:     2,
			MetricIDs:    "*sum#BalanceValue;*average#BalanceValue;*tcc",
			ThresholdIDs: "THRESH4",
			Stored:       true,
			Blockers:     ";true",
			Weights:      ";20",
		},
	}
	rcvTPs := tps.AsTPStats()
	if len(rcvTPs) != 2 {
		t.Errorf("Expecting: 2, received: %+v", len(rcvTPs))
	}
	for _, rcvTP := range rcvTPs {
		if rcvTP.Tenant == "cgrates.org" {
			if len(rcvTP.Metrics) != 8 {
				t.Errorf("Expecting: 8, received: %+v", len(rcvTP.Metrics))
			}
		} else {
			if len(rcvTP.Metrics) != 3 {
				t.Errorf("Expecting: 3, received: %+v", len(rcvTP.Metrics))
			}
		}
	}
}

func TestAPItoTPStats(t *testing.T) {
	tps := &utils.TPStatProfile{
		TPid:        "tp_test",
		ID:          "Stats1",
		FilterIDs:   []string{"FLTR_1", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
		QueueLength: 100,
		TTL:         "1s",
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: "*sum#BalanceValue",
			},
			{
				MetricID: "*average#BalanceValue",
			},
			{
				MetricID: "*tcc",
			},
		},
		MinItems:     1,
		ThresholdIDs: []string{"THRESH1", "THRESH2"},
		Stored:       false,
		Blockers:     ";false",
		Weights:      ";20.0",
	}
	eTPs := &StatQueueProfile{ID: tps.ID,
		QueueLength: tps.QueueLength,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*sum#BalanceValue",
			},
			{
				MetricID: "*average#BalanceValue",
			},
			{
				MetricID: "*tcc",
			},
		},
		ThresholdIDs: []string{"THRESH1", "THRESH2"},
		FilterIDs:    []string{"FLTR_1", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
		Stored:       tps.Stored,
		Weights: utils.DynamicWeights{
			{
				Weight: 20.0,
			},
		},
		MinItems: tps.MinItems,
	}
	var err error
	if eTPs.TTL, err = utils.ParseDurationWithNanosecs(tps.TTL); err != nil {
		t.Errorf("Got error: %+v", err)
	}
	if eTPs.Blockers, err = utils.NewDynamicBlockersFromString(tps.Blockers, utils.InfieldSep, utils.ANDSep); err != nil {
		t.Error(err)
	}

	if st, err := APItoStats(tps, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTPs, st) {
		t.Errorf("Expecting: %+v, received: %+v", eTPs, st)
	}
}

func TestStatQueueProfileToAPI(t *testing.T) {
	expected := &utils.TPStatProfile{
		Tenant:      "cgrates.org",
		ID:          "Stats1",
		FilterIDs:   []string{"FLTR_1", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
		QueueLength: 100,
		TTL:         "1s",
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: "*sum#BalanceValue",
			},
			{
				MetricID: "*average#BalanceValue",
			},
			{
				MetricID: "*tcc",
			},
		},
		MinItems:     1,
		ThresholdIDs: []string{"THRESH1", "THRESH2"},
		Weights:      ";20",
	}
	sqPrf := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "Stats1",
		QueueLength: 100,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*sum#BalanceValue",
			},
			{
				MetricID: "*average#BalanceValue",
			},
			{
				MetricID: "*tcc",
			},
		},
		TTL:          time.Second,
		ThresholdIDs: []string{"THRESH1", "THRESH2"},
		FilterIDs:    []string{"FLTR_1", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20.0,
			},
		},
		MinItems: 1,
	}

	if rcv := StatQueueProfileToAPI(sqPrf); !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v,\n received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestAPItoModelStats(t *testing.T) {
	tpS := &utils.TPStatProfile{
		TPid:        "TPS1",
		Tenant:      "cgrates.org",
		ID:          "Stat1",
		FilterIDs:   []string{"*string:Account:1002", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
		QueueLength: 100,
		TTL:         "1s",
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: "*tcc",
			},
			{
				MetricID: "*average#Usage",
			},
		},
		Blockers:     ";true",
		Stored:       true,
		Weights:      ";20",
		MinItems:     2,
		ThresholdIDs: []string{"Th1"},
	}
	rcv := APItoModelStats(tpS)
	eRcv := StatMdls{
		&StatMdl{
			Tpid:         "TPS1",
			Tenant:       "cgrates.org",
			ID:           "Stat1",
			FilterIDs:    "*string:Account:1002;*ai:~*req.AnswerTime:2014-07-29T15:00:00Z",
			QueueLength:  100,
			TTL:          "1s",
			MinItems:     2,
			MetricIDs:    "*tcc",
			Stored:       true,
			Blockers:     ";true",
			Weights:      ";20",
			ThresholdIDs: "Th1",
		},
		&StatMdl{
			Tpid:      "TPS1",
			Tenant:    "cgrates.org",
			ID:        "Stat1",
			MetricIDs: "*average#Usage",
		},
	}
	if len(rcv) != len(eRcv) {
		t.Errorf("Expecting: %+v, received: %+v", len(eRcv), len(rcv))
	} else if !reflect.DeepEqual(eRcv, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v", utils.ToJSON(eRcv), utils.ToJSON(rcv))
	}
}

func TestTPThresholdsAsTPThreshold(t *testing.T) {
	tps := []*ThresholdMdl{
		{
			Tpid:             "TEST_TPID",
			ID:               "Threhold",
			FilterIDs:        "FilterID1;FilterID2;FilterID1;FilterID2;FilterID2;*ai:~*req.AnswerTime:2014-07-29T15:00:00Z",
			MaxHits:          12,
			MinHits:          10,
			MinSleep:         "1s",
			Blocker:          false,
			Weights:          ";20",
			ActionProfileIDs: "WARN3",
		},
	}
	eTPs := []*utils.TPThresholdProfile{
		{
			TPid:             tps[0].Tpid,
			ID:               tps[0].ID,
			FilterIDs:        []string{"*ai:~*req.AnswerTime:2014-07-29T15:00:00Z", "FilterID1", "FilterID2"},
			MinSleep:         tps[0].MinSleep,
			MaxHits:          tps[0].MaxHits,
			MinHits:          tps[0].MinHits,
			Blocker:          tps[0].Blocker,
			Weights:          tps[0].Weights,
			ActionProfileIDs: []string{"WARN3"},
		},
		{
			TPid:             tps[0].Tpid,
			ID:               tps[0].ID,
			FilterIDs:        []string{"*ai:~*req.AnswerTime:2014-07-29T15:00:00Z", "FilterID2", "FilterID1"},
			MinSleep:         tps[0].MinSleep,
			MaxHits:          tps[0].MaxHits,
			MinHits:          tps[0].MinHits,
			Blocker:          tps[0].Blocker,
			Weights:          tps[0].Weights,
			ActionProfileIDs: []string{"WARN3"},
		},
	}
	rcvTPs := ThresholdMdls(tps).AsTPThreshold()
	sort.Strings(rcvTPs[0].FilterIDs)
	if !reflect.DeepEqual(eTPs[0], rcvTPs[0]) && !reflect.DeepEqual(eTPs[1], rcvTPs[0]) {
		t.Errorf("Expecting: %+v , Received: %+v", utils.ToIJSON(eTPs), utils.ToIJSON(rcvTPs))
	}
}

func TestCSVHeader(t *testing.T) {
	var tps ResourceMdls
	eOut := []string{
		"#Tenant", "ID", "FilterIDs", "Weights", "UsageTTL", "Limit", "AllocationMessage", "Blocker", "Stored", "ThresholdIDs",
	}
	if rcv := tps.CSVHeader(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

}

func TestAPItoModelTPThreshold(t *testing.T) {
	th := &utils.TPThresholdProfile{
		TPid:             "TP1",
		Tenant:           "cgrates.org",
		ID:               "TH_1",
		FilterIDs:        []string{"FilterID1", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z"},
		MaxHits:          12,
		MinHits:          10,
		MinSleep:         "1s",
		Blocker:          false,
		Weights:          ";20",
		ActionProfileIDs: []string{"WARN3"},
	}
	models := ThresholdMdls{
		{
			Tpid:             "TP1",
			Tenant:           "cgrates.org",
			ID:               "TH_1",
			FilterIDs:        "FilterID1",
			MaxHits:          12,
			MinHits:          10,
			MinSleep:         "1s",
			Blocker:          false,
			Weights:          ";20",
			ActionProfileIDs: "WARN3",
		},
		{
			Tpid:      "TP1",
			ID:        "TH_1",
			Tenant:    "cgrates.org",
			FilterIDs: "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z",
		},
	}
	rcv := APItoModelTPThreshold(th)
	if !reflect.DeepEqual(models, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(models), utils.ToJSON(rcv))
	}
}

func TestAPItoModelTPThreshold2(t *testing.T) {
	th := &utils.TPThresholdProfile{
		TPid:             "TP1",
		Tenant:           "cgrates.org",
		ID:               "TH_1",
		FilterIDs:        []string{"FLTR_1", "FLTR_2", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z"},
		MaxHits:          12,
		MinHits:          10,
		MinSleep:         "1s",
		Blocker:          false,
		Weights:          ";20",
		ActionProfileIDs: []string{"WARN3"},
	}
	models := ThresholdMdls{
		{
			Tpid:             "TP1",
			Tenant:           "cgrates.org",
			ID:               "TH_1",
			FilterIDs:        "FLTR_1",
			MaxHits:          12,
			MinHits:          10,
			MinSleep:         "1s",
			Blocker:          false,
			Weights:          ";20",
			ActionProfileIDs: "WARN3",
		},
		{
			Tpid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "TH_1",
			FilterIDs: "FLTR_2",
		},
		{
			Tpid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "TH_1",
			FilterIDs: "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z",
		},
	}
	rcv := APItoModelTPThreshold(th)
	if !reflect.DeepEqual(models, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(models), utils.ToJSON(rcv))
	}
}

func TestAPItoModelTPThreshold3(t *testing.T) {
	th := &utils.TPThresholdProfile{
		TPid:             "TP1",
		Tenant:           "cgrates.org",
		ID:               "TH_1",
		FilterIDs:        []string{"FLTR_1", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z"},
		MaxHits:          12,
		MinHits:          10,
		MinSleep:         "1s",
		Blocker:          false,
		Weights:          ";20",
		ActionProfileIDs: []string{"WARN3", "LOG"},
	}
	models := ThresholdMdls{
		{
			Tpid:             "TP1",
			Tenant:           "cgrates.org",
			ID:               "TH_1",
			FilterIDs:        "FLTR_1",
			MaxHits:          12,
			MinHits:          10,
			MinSleep:         "1s",
			Blocker:          false,
			Weights:          ";20",
			ActionProfileIDs: "WARN3",
		},
		{
			Tpid:             "TP1",
			Tenant:           "cgrates.org",
			ID:               "TH_1",
			ActionProfileIDs: "LOG",
			FilterIDs:        "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z",
		},
	}
	rcv := APItoModelTPThreshold(th)
	if !reflect.DeepEqual(models, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(models), utils.ToJSON(rcv))
	}
}

func TestAPItoModelTPThreshold4(t *testing.T) {
	th := &utils.TPThresholdProfile{
		TPid:             "TP1",
		Tenant:           "cgrates.org",
		ID:               "TH_1",
		FilterIDs:        []string{"*ai:~*req.AnswerTime:2014-07-14T14:35:00Z"},
		MaxHits:          12,
		MinHits:          10,
		MinSleep:         "1s",
		Blocker:          false,
		Weights:          ";20",
		ActionProfileIDs: []string{"WARN3", "LOG"},
	}
	models := ThresholdMdls{
		{
			Tpid:             "TP1",
			Tenant:           "cgrates.org",
			ID:               "TH_1",
			FilterIDs:        "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z",
			MaxHits:          12,
			MinHits:          10,
			MinSleep:         "1s",
			Blocker:          false,
			Weights:          ";20",
			ActionProfileIDs: "WARN3",
		},
		{
			Tpid:             "TP1",
			Tenant:           "cgrates.org",
			ID:               "TH_1",
			ActionProfileIDs: "LOG",
		},
	}
	rcv := APItoModelTPThreshold(th)
	if !reflect.DeepEqual(models, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(models), utils.ToJSON(rcv))
	}
}

func TestAPItoModelTPThreshold5(t *testing.T) {
	th := &utils.TPThresholdProfile{
		TPid:             "TP1",
		Tenant:           "cgrates.org",
		ID:               "TH_1",
		FilterIDs:        []string{"FLTR_1", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z"},
		MaxHits:          12,
		MinHits:          10,
		MinSleep:         "1s",
		Blocker:          false,
		Weights:          ";20",
		ActionProfileIDs: []string{},
	}
	rcv := APItoModelTPThreshold(th)
	if rcv != nil {
		t.Errorf("Expecting : nil, received: %+v", utils.ToJSON(rcv))
	}
}

func TestAPItoTPThreshold(t *testing.T) {
	tps := &utils.TPThresholdProfile{
		TPid:             "tp_test",
		ID:               "TH1",
		FilterIDs:        []string{"FilterID1", "FilterID2", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
		MaxHits:          12,
		MinHits:          10,
		MinSleep:         "1s",
		Blocker:          false,
		Weights:          ";20",
		ActionProfileIDs: []string{"WARN3"},
	}

	eTPs := &ThresholdProfile{
		ID:               tps.ID,
		MaxHits:          tps.MaxHits,
		Blocker:          tps.Blocker,
		MinHits:          tps.MinHits,
		FilterIDs:        tps.FilterIDs,
		ActionProfileIDs: []string{"WARN3"},
	}
	var err error
	eTPs.Weights, err = utils.NewDynamicWeightsFromString(utils.IfaceAsString(tps.Weights), utils.InfieldSep, utils.ANDSep)
	if err != nil {
		t.Errorf("Got error: %+v", err)
	}
	if eTPs.MinSleep, err = utils.ParseDurationWithNanosecs(tps.MinSleep); err != nil {
		t.Errorf("Got error: %+v", err)
	}
	if st, err := APItoThresholdProfile(tps, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTPs, st) {
		t.Errorf("Expecting: %+v, received: %+v", eTPs, st)
	}
}

func TestThresholdProfileToAPI(t *testing.T) {
	expected := &utils.TPThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "TH1",
		FilterIDs:        []string{"FilterID1", "FilterID2", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
		MaxHits:          12,
		MinHits:          10,
		MinSleep:         "1s",
		Weights:          ";20",
		ActionProfileIDs: []string{"WARN3"},
	}

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"FilterID1", "FilterID2", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
		MaxHits:   12,
		MinHits:   10,
		MinSleep:  time.Second,
		Weights: utils.DynamicWeights{
			{
				Weight: 20.0,
			},
		},
		ActionProfileIDs: []string{"WARN3"},
	}

	if rcv := ThresholdProfileToAPI(thPrf); !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v,\n received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestTPFilterAsTPFilter(t *testing.T) {
	tps := []*FilterMdl{
		{
			Tpid:    "TEST_TPID",
			ID:      "Filter1",
			Type:    utils.MetaPrefix,
			Element: "Account",
			Values:  "1001;1002",
		},
	}
	eTPs := []*utils.TPFilterProfile{
		{
			TPid: tps[0].Tpid,
			ID:   tps[0].ID,
			Filters: []*utils.TPFilter{
				{
					Type:    utils.MetaPrefix,
					Element: "Account",
					Values:  []string{"1001", "1002"},
				},
			},
		},
	}

	rcvTPs := FilterMdls(tps).AsTPFilter()
	if !(reflect.DeepEqual(eTPs, rcvTPs) || reflect.DeepEqual(eTPs[0], rcvTPs[0])) {
		t.Errorf("Expecting:\n%+v\nReceived:\n%+v", utils.ToIJSON(eTPs), utils.ToIJSON(rcvTPs))
	}
}

func TestTPFilterAsTPFilterWithDynValues(t *testing.T) {
	tps := []*FilterMdl{
		{
			Tpid:    "TEST_TPID",
			ID:      "Filter1",
			Type:    utils.MetaString,
			Element: "CustomField",
			Values:  "1001;~*uch.<~*opts.*originID;~*rep.RunID;-Cost>;1002;~*uch.<~*opts.*originID;~*rep.RunID>",
		},
	}
	eTPs := []*utils.TPFilterProfile{
		{
			TPid: tps[0].Tpid,
			ID:   tps[0].ID,
			Filters: []*utils.TPFilter{
				{
					Type:    utils.MetaString,
					Element: "CustomField",
					Values:  []string{"1001", "~*uch.<~*opts.*originID;~*rep.RunID;-Cost>", "1002", "~*uch.<~*opts.*originID;~*rep.RunID>"},
				},
			},
		},
	}

	rcvTPs := FilterMdls(tps).AsTPFilter()
	if !(reflect.DeepEqual(eTPs, rcvTPs) || reflect.DeepEqual(eTPs[0], rcvTPs[0])) {
		t.Errorf("Expecting:\n%+v\nReceived:\n%+v", utils.ToIJSON(eTPs), utils.ToIJSON(rcvTPs))
	}
}

func TestTPFilterAsTPFilter2(t *testing.T) {
	tps := []*FilterMdl{
		{
			Tpid:    "TEST_TPID",
			Tenant:  "cgrates.org",
			ID:      "Filter1",
			Type:    utils.MetaPrefix,
			Element: "Account",
			Values:  "1001;1002",
		},
		{
			Tpid:    "TEST_TPID",
			Tenant:  "anotherTenant",
			ID:      "Filter1",
			Type:    utils.MetaPrefix,
			Element: "Account",
			Values:  "1010",
		},
	}
	eTPs := []*utils.TPFilterProfile{
		{
			TPid:   tps[0].Tpid,
			Tenant: "cgrates.org",
			ID:     tps[0].ID,
			Filters: []*utils.TPFilter{
				{
					Type:    utils.MetaPrefix,
					Element: "Account",
					Values:  []string{"1001", "1002"},
				},
			},
		},
		{
			TPid:   tps[1].Tpid,
			Tenant: "anotherTenant",
			ID:     tps[1].ID,
			Filters: []*utils.TPFilter{
				{
					Type:    utils.MetaPrefix,
					Element: "Account",
					Values:  []string{"1010"},
				},
			},
		},
	}

	rcvTPs := FilterMdls(tps).AsTPFilter()
	if len(eTPs) != len(rcvTPs) {
		t.Errorf("Expecting: %+v ,Received: %+v", utils.ToIJSON(eTPs), utils.ToIJSON(rcvTPs))
	}
}

func TestTPFilterAsTPFilter3(t *testing.T) {
	tps := []*FilterMdl{
		{
			Tpid:    "TEST_TPID",
			Tenant:  "cgrates.org",
			ID:      "Filter1",
			Type:    utils.MetaPrefix,
			Element: "Account",
			Values:  "1001",
		},
		{
			Tpid:    "TEST_TPID",
			Tenant:  "cgrates.org",
			ID:      "Filter1",
			Type:    utils.MetaPrefix,
			Element: "Account",
			Values:  "1001",
		},
		{
			Tpid:    "TEST_TPID",
			Tenant:  "anotherTenant",
			ID:      "Filter1",
			Type:    utils.MetaPrefix,
			Element: "Account",
			Values:  "1010",
		},
	}
	eTPs := []*utils.TPFilterProfile{
		{
			TPid:   tps[0].Tpid,
			Tenant: "cgrates.org",
			ID:     tps[0].ID,
			Filters: []*utils.TPFilter{
				{
					Type:    utils.MetaPrefix,
					Element: "Account",
					Values:  []string{"1001", "1001"},
				},
			},
		},
		{
			TPid:   tps[1].Tpid,
			Tenant: "anotherTenant",
			ID:     tps[1].ID,
			Filters: []*utils.TPFilter{
				{
					Type:    utils.MetaPrefix,
					Element: "Account",
					Values:  []string{"1010"},
				},
			},
		},
	}

	rcvTPs := FilterMdls(tps).AsTPFilter()
	sort.Slice(rcvTPs, func(i, j int) bool { return rcvTPs[i].Tenant > rcvTPs[j].Tenant })
	sort.Strings(rcvTPs[0].Filters[0].Values)
	sort.Strings(eTPs[0].Filters[0].Values)
	if !reflect.DeepEqual(eTPs, rcvTPs) {
		t.Errorf("Expecting: %+v \n ,Received: %+v", utils.ToJSON(eTPs), utils.ToJSON(rcvTPs))
	}
}

func TestAPItoModelTPFilter(t *testing.T) {
	var th *utils.TPFilterProfile
	if rcv := APItoModelTPFilter(th); rcv != nil {
		t.Errorf("Expecting: nil ,Received: %+v", utils.ToJSON(rcv))
	}
	th = &utils.TPFilterProfile{
		ID: "someID",
	}
	if rcv := APItoModelTPFilter(th); rcv != nil {
		t.Errorf("Expecting: nil ,Received: %+v", utils.ToJSON(rcv))
	}
	th = &utils.TPFilterProfile{
		ID: "someID",
		Filters: []*utils.TPFilter{
			{
				Type:    utils.MetaPrefix,
				Element: "Account",
				Values:  []string{"1010"},
			},

			{
				Type:    utils.MetaPrefix,
				Element: "Account",
				Values:  []string{"0708"},
			},
		},
	}
	eOut := FilterMdls{
		&FilterMdl{
			ID:      "someID",
			Type:    "*prefix",
			Element: "Account",
			Values:  "1010",
		},
		&FilterMdl{
			ID:      "someID",
			Type:    "*prefix",
			Element: "Account",
			Values:  "0708",
		},
	}
	if rcv := APItoModelTPFilter(th); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	th = &utils.TPFilterProfile{
		TPid:   "TPid",
		Tenant: "cgrates.org",
		ID:     "someID",
		Filters: []*utils.TPFilter{
			{
				Type:    utils.MetaPrefix,
				Element: "Account",
				Values:  []string{"1001", "1002"},
			},
		},
	}
	eOut = FilterMdls{
		{
			Tpid:    "TPid",
			Tenant:  "cgrates.org",
			ID:      "someID",
			Type:    "*prefix",
			Element: "Account",
			Values:  "1001;1002",
		},
	}
	if rcv := APItoModelTPFilter(th); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	th = &utils.TPFilterProfile{
		TPid:   "TPid",
		ID:     "testID",
		Tenant: "cgrates.org",
		Filters: []*utils.TPFilter{
			{
				Type:    utils.MetaString,
				Element: "CustomField",
				Values:  []string{"1001", "~*uch.<~*opts.*originID;~*rep.RunID;-Cost>", "1002", "~*uch.<~*opts.*originID;~*rep.RunID>"},
			},
		},
	}
	eOut = FilterMdls{
		{
			Tpid:    "TPid",
			Tenant:  "cgrates.org",
			ID:      "testID",
			Type:    "*string",
			Element: "CustomField",
			Values:  "1001;~*uch.\u003c~*opts.*originID;~*rep.RunID;-Cost\u003e;1002;~*uch.\u003c~*opts.*originID;~*rep.RunID\u003e",
		},
	}
	if rcv := APItoModelTPFilter(th); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestAPItoTPFilter(t *testing.T) {
	tps := &utils.TPFilterProfile{
		TPid:   "tp_test",
		Tenant: "cgrates.org",
		ID:     "Filter1",
		Filters: []*utils.TPFilter{
			{
				Element: "Account",
				Type:    utils.MetaString,
				Values:  []string{"1001", "1002"},
			},
		},
	}

	eTPs := &Filter{
		Tenant: "cgrates.org",
		ID:     tps.ID,
		Rules: []*FilterRule{
			{
				Element: "Account",
				Type:    utils.MetaString,
				Values:  []string{"1001", "1002"},
			},
		},
	}
	if err := eTPs.Compile(); err != nil {
		t.Fatal(err)
	}
	if st, err := APItoFilter(tps, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTPs, st) {
		t.Errorf("Expecting: %+v, received: %+v", eTPs, st)
	}
}

func TestFilterToTPFilter(t *testing.T) {
	filter := &Filter{
		Tenant: "cgrates.org",
		ID:     "Fltr1",
		Rules: []*FilterRule{
			{
				Element: "Account",
				Type:    utils.MetaString,
				Values:  []string{"1001", "1002"},
			},
		},
	}
	tpfilter := &utils.TPFilterProfile{
		ID:     "Fltr1",
		Tenant: "cgrates.org",
		Filters: []*utils.TPFilter{
			{
				Element: "Account",
				Type:    utils.MetaString,
				Values:  []string{"1001", "1002"},
			},
		},
	}
	eTPFilter := FilterToTPFilter(filter)
	if !reflect.DeepEqual(tpfilter, eTPFilter) {
		t.Errorf("Expecting: %+v, received: %+v", tpfilter, eTPFilter)
	}
}

func TestCsvHeader(t *testing.T) {
	var tps RouteMdls
	eOut := []string{
		"#" + utils.Tenant, utils.ID, utils.FilterIDs, utils.Weights,
		utils.Sorting, utils.SortingParameters, utils.RouteID, utils.RouteFilterIDs,
		utils.RouteAccountIDs, utils.RouteRateProfileIDs, utils.RouteRateProfileIDs, utils.RouteResourceIDs,
		utils.RouteStatIDs, utils.RouteWeights, utils.RouteBlockers,
		utils.RouteParameters,
	}
	if rcv := tps.CSVHeader(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestAPItoAttributeProfile(t *testing.T) {
	tpAlsPrf := &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "*string:~*opts.*context:con1"},
		Attributes: []*utils.TPAttribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FL1",
				Value: "Al1",
			},
		},
		Weights: ";20",
	}
	expected := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "*string:~*opts.*context:con1"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FL1",
				Value: config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
			},
		},
		Weights: make(utils.DynamicWeights, 1),
	}
	expected.Weights[0] = &utils.DynamicWeight{
		Weight: 20,
	}
	if rcv, err := APItoAttributeProfile(tpAlsPrf, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestAttributeProfileToAPI(t *testing.T) {
	exp := &utils.TPAttributeProfile{
		TPid:      utils.EmptyString,
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-15T14:36:00Z", "*string:~*opts.*context:con1"},
		Attributes: []*utils.TPAttribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FL1",
				Value: "Al1",
			},
		},
		Weights: ";20",
	}
	attr := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-15T14:36:00Z", "*string:~*opts.*context:con1"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FL1",
				Value: config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
			},
		},
		Weights: make(utils.DynamicWeights, 1),
	}
	attr.Weights[0] = &utils.DynamicWeight{
		Weight: 20,
	}
	if rcv := AttributeProfileToAPI(attr); !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestAttributeProfileToAPI2(t *testing.T) {
	exp := &utils.TPAttributeProfile{
		TPid:      utils.EmptyString,
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "*string:~*opts.*context:con1"},
		Attributes: []*utils.TPAttribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FL1",
				Value: "Al1",
			},
			{
				Path:  utils.MetaReq + utils.NestingSep + "Test",
				Value: "~*req.Account",
			},
		},
		Weights: ";20",
	}
	attr := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "*string:~*opts.*context:con1"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FL1",
				Value: config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
			},
			{
				Path:  utils.MetaReq + utils.NestingSep + "Test",
				Value: config.NewRSRParsersMustCompile("~*req.Account", utils.InfieldSep),
			},
		},
		Weights: make(utils.DynamicWeights, 1),
	}
	attr.Weights[0] = &utils.DynamicWeight{
		Weight: 20,
	}
	if rcv := AttributeProfileToAPI(attr); !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestAPItoModelTPAttribute(t *testing.T) {
	tpAlsPrf := &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-14T14:36:00Z", "*string:~*opts.*context:con1|con2"},
		Attributes: []*utils.TPAttribute{
			{FilterIDs: []string{"filter_id1", "filter_id2"},
				Path:  utils.MetaReq + utils.NestingSep + "FL1",
				Value: "Al1",
			},
		},
		Weights: ";20",
	}
	expected := AttributeMdls{
		&AttributeMdl{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "ALS1",
			FilterIDs:          "FLTR_ACNT_dan;FLTR_DST_DE;*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-14T14:36:00Z;*string:~*opts.*context:con1|con2",
			AttributeFilterIDs: "filter_id1;filter_id2",
			Path:               utils.MetaReq + utils.NestingSep + "FL1",
			Value:              "Al1",
			Weights:            ";20",
		},
	}
	rcv := APItoModelTPAttribute(tpAlsPrf)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestCsvDumpForAttributeModels(t *testing.T) {
	tpAlsPrf := &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		FilterIDs: []string{"FLTR_ACNT_dan", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "*string:~*opts.*context:con1"},
		Attributes: []*utils.TPAttribute{
			{
				Blockers: ";false",
				Path:     utils.MetaReq + utils.NestingSep + "FL1",
				Value:    "Al1",
			},
			{
				Path:  utils.MetaReq + utils.NestingSep + "FL2",
				Value: "Al2",
			},
		},
		Blockers: ";true",
		Weights:  ";20",
	}
	expected := AttributeMdls{
		&AttributeMdl{
			Tpid:              "TP1",
			Tenant:            "cgrates.org",
			ID:                "ALS1",
			FilterIDs:         "FLTR_ACNT_dan;*ai:~*req.AnswerTime:2014-07-14T14:35:00Z;*string:~*opts.*context:con1",
			Path:              utils.MetaReq + utils.NestingSep + "FL1",
			AttributeBlockers: ";false",
			Value:             "Al1",
			Blockers:          ";true",
			Weights:           ";20",
		},
		&AttributeMdl{
			Tpid:   "TP1",
			Tenant: "cgrates.org",
			ID:     "ALS1",
			Path:   utils.MetaReq + utils.NestingSep + "FL2",
			Value:  "Al2",
		},
	}
	rcv := APItoModelTPAttribute(tpAlsPrf)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v,\n received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
	expRecord := []string{"cgrates.org", "ALS1", "FLTR_ACNT_dan;*ai:~*req.AnswerTime:2014-07-14T14:35:00Z;*string:~*opts.*context:con1", ";20", ";true", "", ";false", "*req.FL1", "", "Al1"}
	for i, model := range rcv {
		if i == 1 {
			expRecord = []string{"cgrates.org", "ALS1", "", "", "", "", "", "*req.FL2", "", "Al2"}
		}
		if csvRecordRcv, _ := CsvDump(model); !reflect.DeepEqual(expRecord, csvRecordRcv) {
			t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expRecord), utils.ToJSON(csvRecordRcv))
		}
	}

}

func TestModelAsTPAttribute2(t *testing.T) {
	models := AttributeMdls{
		&AttributeMdl{
			Tpid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "ALS1",
			FilterIDs: "FLTR_ACNT_dan;FLTR_DST_DE;*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-14T14:36:00Z;*string:~*opts.*context:con1",
			Path:      utils.MetaReq + utils.NestingSep + "FL1",
			Value:     "Al1",
			Weights:   ";20",
		},
	}
	expected := &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		FilterIDs: []string{"*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-14T14:36:00Z", "*string:~*opts.*context:con1", "FLTR_ACNT_dan", "FLTR_DST_DE"},
		Attributes: []*utils.TPAttribute{
			{
				FilterIDs: []string{},
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Value:     "Al1",
			},
		},
		Weights: ";20",
	}
	expected2 := &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		FilterIDs: []string{"*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "*string:~*opts.*context:con1", "FLTR_ACNT_dan", "FLTR_DST_DE"},
		Attributes: []*utils.TPAttribute{
			{
				FilterIDs: []string{},
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Value:     "Al1",
			},
		},
		Weights: ";20",
	}
	rcv := models.AsTPAttributes()
	sort.Strings(rcv[0].FilterIDs)
	if !reflect.DeepEqual(expected, rcv[0]) && !reflect.DeepEqual(expected2, rcv[0]) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv[0]))
	}
}

func TestModelAsTPAttribute(t *testing.T) {
	models := AttributeMdls{
		&AttributeMdl{
			Tpid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "ALS1",
			FilterIDs: "FLTR_ACNT_dan;FLTR_DST_DE;*ai:~*req.AnswerTime:2014-07-14T14:35:00Z;*string:~*opts.*context:con1",
			Weights:   ";20",
			Blockers:  ";true",
			Type:      utils.MetaConstant,
			Path:      utils.MetaReq + utils.NestingSep + "FL1",
			Value:     "Al1",
		},
	}
	expected := &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		FilterIDs: []string{"*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "*string:~*opts.*context:con1", "FLTR_ACNT_dan", "FLTR_DST_DE"},
		Weights:   ";20",
		Blockers:  ";true",
		Attributes: []*utils.TPAttribute{
			{
				FilterIDs: []string{},
				Type:      utils.MetaConstant,
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Value:     "Al1",
			},
		},
	}
	rcv := models.AsTPAttributes()
	sort.Strings(rcv[0].FilterIDs)
	sort.Strings(expected.FilterIDs)
	if !reflect.DeepEqual(expected, rcv[0]) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv[0]))
	}
}

func TestAPItoChargerProfile(t *testing.T) {
	tpCPP := &utils.TPChargerProfile{
		TPid:         "TP1",
		Tenant:       "cgrates.org",
		ID:           "Charger1",
		FilterIDs:    []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z"},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR1", "ATTR2"},
		Weights:      ";20",
	}

	expected := &ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "Charger1",
		FilterIDs:    []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z"},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR1", "ATTR2"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	if rcv := APItoChargerProfile(tpCPP, "UTC"); !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestChargerProfileToAPI(t *testing.T) {
	exp := &utils.TPChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "Charger1",
		FilterIDs:    []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-15T14:36:00Z"},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR1", "ATTR2"},
		Weights:      ";20",
	}

	chargerPrf := &ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "Charger1",
		FilterIDs:    []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-15T14:36:00Z"},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR1", "ATTR2"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	if rcv := ChargerProfileToAPI(chargerPrf); !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expecting : %+v, \n received: %+v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

// Number of FilterIDs and AttributeIDs are equal
func TestAPItoModelTPCharger(t *testing.T) {
	tpCharger := &utils.TPChargerProfile{
		TPid:         "TP1",
		Tenant:       "cgrates.org",
		ID:           "Charger1",
		FilterIDs:    []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-15T14:35:00Z"},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR1", "ATTR2"},
		Weights:      ";20",
	}
	expected := ChargerMdls{
		&ChargerMdl{
			Tpid:         "TP1",
			Tenant:       "cgrates.org",
			ID:           "Charger1",
			FilterIDs:    "FLTR_ACNT_dan",
			RunID:        "*rated",
			AttributeIDs: "ATTR1",
			Weights:      ";20",
		},
		&ChargerMdl{
			Tpid:         "TP1",
			Tenant:       "cgrates.org",
			ID:           "Charger1",
			FilterIDs:    "FLTR_DST_DE",
			AttributeIDs: "ATTR2",
		},
		&ChargerMdl{
			Tpid:         "TP1",
			Tenant:       "cgrates.org",
			ID:           "Charger1",
			FilterIDs:    "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-15T14:35:00Z",
			AttributeIDs: "",
		},
	}
	rcv := APItoModelTPCharger(tpCharger)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

// Number of FilterIDs is smaller than AttributeIDs
func TestAPItoModelTPCharger2(t *testing.T) {
	tpCharger := &utils.TPChargerProfile{
		TPid:         "TP1",
		Tenant:       "cgrates.org",
		ID:           "Charger1",
		FilterIDs:    []string{"FLTR_ACNT_dan", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-15T14:35:00Z"},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR1", "ATTR2"},
		Weights:      ";20",
	}
	expected := ChargerMdls{
		&ChargerMdl{
			Tpid:         "TP1",
			Tenant:       "cgrates.org",
			ID:           "Charger1",
			FilterIDs:    "FLTR_ACNT_dan",
			RunID:        "*rated",
			AttributeIDs: "ATTR1",
			Weights:      ";20",
		},
		&ChargerMdl{
			Tpid:         "TP1",
			Tenant:       "cgrates.org",
			ID:           "Charger1",
			FilterIDs:    "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-15T14:35:00Z",
			AttributeIDs: "ATTR2",
		},
	}
	rcv := APItoModelTPCharger(tpCharger)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

// Number of FilterIDs is greater than AttributeIDs
func TestAPItoModelTPCharger3(t *testing.T) {
	tpCharger := &utils.TPChargerProfile{
		TPid:         "TP1",
		Tenant:       "cgrates.org",
		ID:           "Charger1",
		FilterIDs:    []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z"},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR1"},
		Weights:      ";20",
	}
	expected := ChargerMdls{
		&ChargerMdl{
			Tpid:         "TP1",
			Tenant:       "cgrates.org",
			ID:           "Charger1",
			FilterIDs:    "FLTR_ACNT_dan",
			RunID:        "*rated",
			AttributeIDs: "ATTR1",
			Weights:      ";20",
		},
		&ChargerMdl{
			Tpid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "Charger1",
			FilterIDs: "FLTR_DST_DE",
		},
		&ChargerMdl{
			Tpid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "Charger1",
			FilterIDs: "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z",
		},
	}
	rcv := APItoModelTPCharger(tpCharger)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

// len(AttributeIDs) is 0
func TestAPItoModelTPCharger4(t *testing.T) {
	tpCharger := &utils.TPChargerProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		FilterIDs: []string{"FLTR_ACNT_dan", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-15T14:35:00Z"},
		RunID:     "*rated",
		Weights:   ";20",
	}
	expected := ChargerMdls{
		&ChargerMdl{
			Tpid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "Charger1",
			FilterIDs: "FLTR_ACNT_dan",
			RunID:     "*rated",
			Weights:   ";20",
		},
		&ChargerMdl{
			Tpid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "Charger1",
			FilterIDs: "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-15T14:35:00Z",
			RunID:     "",
			Weights:   "",
		},
	}
	rcv := APItoModelTPCharger(tpCharger)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

// len(FilterIDs) is 0
func TestAPItoModelTPCharger5(t *testing.T) {
	tpCharger := &utils.TPChargerProfile{
		TPid:         "TP1",
		Tenant:       "cgrates.org",
		ID:           "Charger1",
		FilterIDs:    []string{"*ai:~*req.AnswerTime:2014-07-14T14:35:00Z"},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR1"},
		Weights:      ";20",
	}
	expected := ChargerMdls{
		&ChargerMdl{
			Tpid:         "TP1",
			Tenant:       "cgrates.org",
			ID:           "Charger1",
			RunID:        "*rated",
			AttributeIDs: "ATTR1",
			FilterIDs:    "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z",
			Weights:      ";20",
		},
	}
	rcv := APItoModelTPCharger(tpCharger)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

// both len(AttributeIDs) and len(FilterIDs) are 0
func TestAPItoModelTPCharger6(t *testing.T) {
	tpCharger := &utils.TPChargerProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		RunID:     "*rated",
		FilterIDs: []string{"*ai:~*req.AnswerTime:2014-07-14T14:35:00Z"},
		Weights:   ";20",
	}
	expected := ChargerMdls{
		&ChargerMdl{
			Tpid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "Charger1",
			RunID:     "*rated",
			FilterIDs: "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z",
			Weights:   ";20",
		},
	}
	rcv := APItoModelTPCharger(tpCharger)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestModelAsTPChargers(t *testing.T) {
	models := ChargerMdls{
		&ChargerMdl{
			Tpid:         "TP1",
			Tenant:       "cgrates.org",
			ID:           "Charger1",
			FilterIDs:    "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z;FLTR_ACNT_dan;FLTR_DST_DE",
			RunID:        "*rated",
			AttributeIDs: "ATTR1",
			Weights:      ";20",
		},
	}
	expected := &utils.TPChargerProfile{
		TPid:         "TP1",
		Tenant:       "cgrates.org",
		ID:           "Charger1",
		FilterIDs:    []string{"*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "FLTR_ACNT_dan", "FLTR_DST_DE"},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR1"},
		Weights:      ";20",
	}
	expected2 := &utils.TPChargerProfile{
		TPid:         "TP1",
		Tenant:       "cgrates.org",
		ID:           "Charger1",
		FilterIDs:    []string{"*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "FLTR_DST_DE", "FLTR_ACNT_dan"},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR1"},
		Weights:      ";20",
	}
	rcv := models.AsTPChargers()
	sort.Strings(rcv[0].FilterIDs)
	if !reflect.DeepEqual(expected, rcv[0]) && !reflect.DeepEqual(expected2, rcv[0]) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv[0]))
	}
}

func TestModelAsTPChargers2(t *testing.T) {
	models := ChargerMdls{
		&ChargerMdl{
			Tpid:         "TP1",
			Tenant:       "cgrates.org",
			ID:           "Charger1",
			FilterIDs:    "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z;FLTR_ACNT_dan;FLTR_DST_DE",
			RunID:        "*rated",
			AttributeIDs: "*constant:*req.RequestType:*rated;*constant:*req.Category:call;ATTR1;*constant:*req.Category:call",
			Weights:      ";20",
		},
	}
	expected := &utils.TPChargerProfile{
		TPid:         "TP1",
		Tenant:       "cgrates.org",
		ID:           "Charger1",
		FilterIDs:    []string{"*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "FLTR_ACNT_dan", "FLTR_DST_DE"},
		RunID:        "*rated",
		AttributeIDs: []string{"*constant:*req.RequestType:*rated;*constant:*req.Category:call", "ATTR1", "*constant:*req.Category:call"},
		Weights:      ";20",
	}
	rcv := models.AsTPChargers()
	sort.Strings(rcv[0].FilterIDs)
	if !reflect.DeepEqual(expected, rcv[0]) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv[0]))
	}
}

func TestModelAsTPChargers3(t *testing.T) {
	models := ChargerMdls{
		&ChargerMdl{
			Tpid:         "TP1",
			Tenant:       "cgrates.org",
			ID:           "Charger1",
			FilterIDs:    "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-15T14:35:00Z;FLTR_ACNT_dan;FLTR_DST_DE",
			RunID:        "*rated",
			AttributeIDs: "*constant:*req.RequestType:*rated;*constant:*req.Category:call;ATTR1;*constant:*req.Category:call&<~*req.OriginID;_suf>",
			Weights:      ";20",
		},
	}
	expected := &utils.TPChargerProfile{
		TPid:         "TP1",
		Tenant:       "cgrates.org",
		ID:           "Charger1",
		FilterIDs:    []string{"*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-15T14:35:00Z", "FLTR_ACNT_dan", "FLTR_DST_DE"},
		RunID:        "*rated",
		AttributeIDs: []string{"*constant:*req.RequestType:*rated;*constant:*req.Category:call", "ATTR1", "*constant:*req.Category:call&<~*req.OriginID;_suf>"},
		Weights:      ";20",
	}
	rcv := models.AsTPChargers()
	sort.Strings(rcv[0].FilterIDs)
	if !reflect.DeepEqual(expected, rcv[0]) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv[0]))
	}
}

func TestAPItoDispatcherProfile(t *testing.T) {
	tpDPP := &utils.TPDispatcherProfile{
		TPid:           "TP1",
		Tenant:         "cgrates.org",
		ID:             "Dsp",
		FilterIDs:      []string{"*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "FLTR_ACNT_dan", "FLTR_DST_DE"},
		Strategy:       utils.MetaFirst,
		StrategyParams: []any{},
		Weight:         20,
		Hosts: []*utils.TPDispatcherHostProfile{
			{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    []any{"192.168.54.203", "*ratio:2"},
				Blocker:   false,
			},
		},
	}

	expected := &DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "Dsp",
		FilterIDs:      []string{"*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "FLTR_ACNT_dan", "FLTR_DST_DE"},
		Strategy:       utils.MetaFirst,
		StrategyParams: map[string]any{},
		Weight:         20,
		Hosts: DispatcherHostProfiles{
			&DispatcherHostProfile{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    map[string]any{"0": "192.168.54.203", utils.MetaRatio: "2"},
				Blocker:   false,
			},
		},
	}
	if rcv := APItoDispatcherProfile(tpDPP, "UTC"); !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestDispatcherProfileToAPI(t *testing.T) {
	exp := &utils.TPDispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "Dsp",
		FilterIDs:      []string{"*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "FLTR_ACNT_dan", "FLTR_DST_DE"},
		Strategy:       utils.MetaFirst,
		StrategyParams: []any{},
		Weight:         20,
		Hosts: []*utils.TPDispatcherHostProfile{
			{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    []any{"192.168.54.203", "*ratio:2"},
				Blocker:   false,
			},
		},
	}
	exp2 := &utils.TPDispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "Dsp",
		FilterIDs:      []string{"*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "FLTR_ACNT_dan", "FLTR_DST_DE"},
		Strategy:       utils.MetaFirst,
		StrategyParams: []any{},
		Weight:         20,
		Hosts: []*utils.TPDispatcherHostProfile{
			{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    []any{"*ratio:2", "192.168.54.203"},
				Blocker:   false,
			},
		},
	}

	dspPrf := &DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "Dsp",
		FilterIDs:      []string{"*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "FLTR_ACNT_dan", "FLTR_DST_DE"},
		Strategy:       utils.MetaFirst,
		StrategyParams: map[string]any{},
		Weight:         20,
		Hosts: DispatcherHostProfiles{
			&DispatcherHostProfile{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    map[string]any{"0": "192.168.54.203", utils.MetaRatio: "2"},
				Blocker:   false,
			},
		},
	}
	if rcv := DispatcherProfileToAPI(dspPrf); !reflect.DeepEqual(exp, rcv) && !reflect.DeepEqual(exp2, rcv) {
		t.Errorf("Expecting : \n %+v \n  or \n %+v \n ,\n received: %+v", utils.ToJSON(exp), utils.ToJSON(exp2), utils.ToJSON(rcv))
	}
}

func TestAPItoModelTPDispatcher(t *testing.T) {
	tpDPP := &utils.TPDispatcherProfile{
		TPid:           "TP1",
		Tenant:         "cgrates.org",
		ID:             "Dsp",
		FilterIDs:      []string{"*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "FLTR_ACNT_dan", "FLTR_DST_DE"},
		Strategy:       utils.MetaFirst,
		StrategyParams: []any{},
		Weight:         20,
		Hosts: []*utils.TPDispatcherHostProfile{
			{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    []any{"192.168.54.203"},
				Blocker:   false,
			},
			{
				ID:        "C2",
				FilterIDs: []string{},
				Weight:    10,
				Params:    []any{"192.168.54.204"},
				Blocker:   false,
			},
		},
	}
	expected := DispatcherProfileMdls{
		&DispatcherProfileMdl{
			Tpid:           "TP1",
			Tenant:         "cgrates.org",
			ID:             "Dsp",
			FilterIDs:      "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z;FLTR_ACNT_dan;FLTR_DST_DE",
			Strategy:       utils.MetaFirst,
			Weight:         20,
			ConnID:         "C1",
			ConnWeight:     10,
			ConnBlocker:    false,
			ConnParameters: "192.168.54.203",
		},
		&DispatcherProfileMdl{
			Tpid:           "TP1",
			Tenant:         "cgrates.org",
			ID:             "Dsp",
			ConnID:         "C2",
			ConnWeight:     10,
			ConnBlocker:    false,
			ConnParameters: "192.168.54.204",
		},
	}
	rcv := APItoModelTPDispatcherProfile(tpDPP)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, \n received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestTPDispatcherHostsCSVHeader(t *testing.T) {
	tps := &DispatcherHostMdls{}
	eOut := []string{"#" + utils.Tenant, utils.ID, utils.Address, utils.Transport, utils.SynchronousCfg, utils.ConnectAttemptsCfg, utils.ReconnectsCfg, utils.MaxReconnectIntervalCfg, utils.ConnectTimeoutCfg, utils.ReplyTimeoutCfg, utils.TLS, utils.ClientKeyCfg, utils.ClientCerificateCfg, utils.CaCertificateCfg}
	if rcv := tps.CSVHeader(); !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestTPDispatcherHostsAsTPDispatcherHosts(t *testing.T) {
	tps := &DispatcherHostMdls{}
	if rcv, err := tps.AsTPDispatcherHosts(); err != nil {
		t.Error(err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil,\nReceived: %+v", utils.ToJSON(rcv))
	}

	tps = &DispatcherHostMdls{
		&DispatcherHostMdl{
			ID:     "ID1",
			Tenant: "Tenant1",
		}}
	if rcv, err := tps.AsTPDispatcherHosts(); err != nil {
		t.Error(err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil,\nReceived: %+v", utils.ToJSON(rcv))
	}

	tps = &DispatcherHostMdls{
		&DispatcherHostMdl{
			ID:                   "ID1",
			Tenant:               "Tenant1",
			Address:              "localhost:6012",
			Transport:            "*json",
			ConnectAttempts:      2,
			Reconnects:           5,
			MaxReconnectInterval: "5m",
			ConnectTimeout:       "2m",
			ReplyTimeout:         "1m",
			TLS:                  true,
			ClientKey:            "client_key",
			ClientCertificate:    "client_certificate",
			CaCertificate:        "ca_certificate",
		}}
	eOut := []*utils.TPDispatcherHost{
		{
			Tenant: "Tenant1",
			ID:     "ID1",
			Conn: &utils.TPDispatcherHostConn{
				Address:              "localhost:6012",
				Transport:            "*json",
				ConnectAttempts:      2,
				Reconnects:           5,
				MaxReconnectInterval: 5 * time.Minute,
				ConnectTimeout:       2 * time.Minute,
				ReplyTimeout:         1 * time.Minute,
				TLS:                  true,
				ClientKey:            "client_key",
				ClientCertificate:    "client_certificate",
				CaCertificate:        "ca_certificate",
			},
		},
	}
	if rcv, err := tps.AsTPDispatcherHosts(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

	tps = &DispatcherHostMdls{
		&DispatcherHostMdl{
			Address:   "Address2",
			ID:        "ID2",
			Tenant:    "Tenant2",
			Transport: "*gob",
		}}
	eOut = []*utils.TPDispatcherHost{
		{
			Tenant: "Tenant2",
			ID:     "ID2",
			Conn: &utils.TPDispatcherHostConn{
				Address:   "Address2",
				Transport: "*gob",
			},
		},
	}
	if rcv, err := tps.AsTPDispatcherHosts(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

	tps = &DispatcherHostMdls{
		&DispatcherHostMdl{
			Address:   "Address3",
			ID:        "ID3",
			Tenant:    "Tenant3",
			Transport: "*gob",
		},
	}
	eOut = []*utils.TPDispatcherHost{
		{
			Tenant: "Tenant3",
			ID:     "ID3",
			Conn: &utils.TPDispatcherHostConn{
				Address:   "Address3",
				Transport: "*gob",
			},
		},
	}
	if rcv, err := tps.AsTPDispatcherHosts(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

	tps = &DispatcherHostMdls{
		&DispatcherHostMdl{
			Address:   "Address4",
			ID:        "ID4",
			Tenant:    "Tenant4",
			Transport: "*gob",
		},
	}
	eOut = []*utils.TPDispatcherHost{
		{
			Tenant: "Tenant4",
			ID:     "ID4",
			Conn: &utils.TPDispatcherHostConn{
				Address:   "Address4",
				Transport: "*gob",
			},
		},
	}
	rcv, err := tps.AsTPDispatcherHosts()
	if err != nil {
		t.Error(err)
	}
	sort.Slice(rcv, func(i, j int) bool { return strings.Compare(rcv[i].ID, rcv[j].ID) < 0 })
	if !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestAPItoModelTPDispatcherHost(t *testing.T) {
	var tpDPH *utils.TPDispatcherHost
	if rcv := APItoModelTPDispatcherHost(tpDPH); rcv != nil {
		t.Errorf("Expecting: nil,\nReceived: %+v", utils.ToJSON(rcv))
	}

	tpDPH = &utils.TPDispatcherHost{
		Tenant: "Tenant",
		ID:     "ID",
		Conn: &utils.TPDispatcherHostConn{
			Address:              "Address1",
			Transport:            "*json",
			ConnectAttempts:      3,
			Reconnects:           5,
			MaxReconnectInterval: 5 * time.Minute,
			ConnectTimeout:       1 * time.Minute,
			ReplyTimeout:         2 * time.Minute,
			TLS:                  true,
			ClientKey:            "client_key",
			ClientCertificate:    "client_certificate",
			CaCertificate:        "ca_certificate",
		},
	}
	eOut := &DispatcherHostMdl{
		Address:              "Address1",
		Transport:            "*json",
		Tenant:               "Tenant",
		ID:                   "ID",
		ConnectAttempts:      3,
		Reconnects:           5,
		MaxReconnectInterval: "5m0s",
		ConnectTimeout:       "1m0s",
		ReplyTimeout:         "2m0s",
		TLS:                  true,
		ClientKey:            "client_key",
		ClientCertificate:    "client_certificate",
		CaCertificate:        "ca_certificate",
	}
	if rcv := APItoModelTPDispatcherHost(tpDPH); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

}

func TestAPItoDispatcherHost(t *testing.T) {
	var tpDPH *utils.TPDispatcherHost
	if rcv := APItoDispatcherHost(tpDPH); rcv != nil {
		t.Errorf("Expecting: nil,\nReceived: %+v", utils.ToJSON(rcv))
	}

	tpDPH = &utils.TPDispatcherHost{
		Tenant: "Tenant1",
		ID:     "ID1",
		Conn: &utils.TPDispatcherHostConn{
			Address:              "localhost:6012",
			Transport:            "*json",
			ConnectAttempts:      3,
			Reconnects:           5,
			MaxReconnectInterval: 5 * time.Minute,
			ConnectTimeout:       1 * time.Minute,
			ReplyTimeout:         2 * time.Minute,
			TLS:                  true,
			ClientKey:            "client_key",
			ClientCertificate:    "client_certificate",
			CaCertificate:        "ca_certificate",
		},
	}

	eOut := &DispatcherHost{
		Tenant: "Tenant1",
		RemoteHost: &config.RemoteHost{
			ID:                   "ID1",
			Address:              "localhost:6012",
			Transport:            "*json",
			Reconnects:           5,
			MaxReconnectInterval: 5 * time.Minute,
			ConnectTimeout:       1 * time.Minute,
			ReplyTimeout:         2 * time.Minute,
			TLS:                  true,
			ClientKey:            "client_key",
			ClientCertificate:    "client_certificate",
			CaCertificate:        "ca_certificate",
			ConnectAttempts:      3,
		},
	}
	if rcv := APItoDispatcherHost(tpDPH); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

	tpDPH = &utils.TPDispatcherHost{
		Tenant: "Tenant2",
		ID:     "ID2",
		Conn: &utils.TPDispatcherHostConn{
			Address:   "Address1",
			Transport: "*json",
			TLS:       true,
		},
	}
	eOut = &DispatcherHost{
		Tenant: "Tenant2",
		RemoteHost: &config.RemoteHost{
			ID:        "ID2",
			Address:   "Address1",
			Transport: "*json",
			TLS:       true,
		},
	}
	if rcv := APItoDispatcherHost(tpDPH); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestDispatcherHostToAPI(t *testing.T) {
	dph := &DispatcherHost{
		Tenant: "Tenant1",
		RemoteHost: &config.RemoteHost{
			Address:              "127.0.0.1:2012",
			Transport:            "*json",
			ConnectAttempts:      0,
			Reconnects:           0,
			MaxReconnectInterval: 5 * time.Minute,
			ConnectTimeout:       1 * time.Minute,
			ReplyTimeout:         1 * time.Minute,
			TLS:                  false,
			ClientKey:            "",
			ClientCertificate:    "",
			CaCertificate:        "",
		},
	}
	eOut := &utils.TPDispatcherHost{
		Tenant: "Tenant1",
		Conn: &utils.TPDispatcherHostConn{
			Address:              "127.0.0.1:2012",
			Transport:            "*json",
			ConnectAttempts:      0,
			Reconnects:           0,
			MaxReconnectInterval: 5 * time.Minute,
			ConnectTimeout:       1 * time.Minute,
			ReplyTimeout:         1 * time.Minute,
			TLS:                  false,
			ClientKey:            "",
			ClientCertificate:    "",
			CaCertificate:        "",
		},
	}
	if rcv := DispatcherHostToAPI(dph); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

}

func TestTPRoutesAsTPRouteProfile(t *testing.T) {
	mdl := RouteMdls{
		&RouteMdl{
			PK:                  1,
			Tpid:                "TP",
			Tenant:              "cgrates.org",
			ID:                  "RoutePrf",
			FilterIDs:           "FltrRoute;*ai:~*req.AnswerTime:2017-11-27T00:00:00Z",
			Sorting:             "*weight",
			SortingParameters:   "srtPrm1",
			RouteID:             "route1",
			RouteFilterIDs:      "",
			RouteAccountIDs:     "",
			RouteRateProfileIDs: "",
			RouteResourceIDs:    "",
			RouteStatIDs:        "",
			RouteWeights:        ";10.0",
			RouteBlockers:       ";false",
			RouteParameters:     "",
			Weights:             ";10",
			CreatedAt:           time.Time{},
		},
		&RouteMdl{
			PK:                  2,
			Tpid:                "TP",
			Tenant:              "cgrates.org",
			ID:                  "RoutePrf",
			FilterIDs:           "",
			Sorting:             "",
			SortingParameters:   "",
			RouteID:             "route2",
			RouteFilterIDs:      "",
			RouteAccountIDs:     "",
			RouteRateProfileIDs: "",
			RouteResourceIDs:    "",
			RouteStatIDs:        "",
			RouteWeights:        ";20.0",
			RouteBlockers:       ";false",
			RouteParameters:     "",
			CreatedAt:           time.Time{},
		},
	}
	expPrf := []*utils.TPRouteProfile{
		{
			TPid:              "TP",
			Tenant:            "cgrates.org",
			ID:                "RoutePrf",
			Sorting:           "*weight",
			SortingParameters: []string{"srtPrm1"},
			FilterIDs:         []string{"*ai:~*req.AnswerTime:2017-11-27T00:00:00Z", "FltrRoute"},
			Routes: []*utils.TPRoute{
				{
					ID:       "route1",
					Weights:  ";10.0",
					Blockers: ";false",
				},
				{
					ID:       "route2",
					Weights:  ";20.0",
					Blockers: ";false",
				},
			},
			Weights: ";10",
		},
	}
	rcv := mdl.AsTPRouteProfile()
	sort.Slice(rcv[0].Routes, func(i, j int) bool {
		return strings.Compare(rcv[0].Routes[i].ID, rcv[0].Routes[j].ID) < 0
	})
	sort.Strings(rcv[0].FilterIDs)
	if !reflect.DeepEqual(rcv, expPrf) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(expPrf), utils.ToJSON(rcv))
	}

	mdlReverse := RouteMdls{
		&RouteMdl{
			PK:                  2,
			Tpid:                "TP",
			Tenant:              "cgrates.org",
			ID:                  "RoutePrf",
			FilterIDs:           "",
			Sorting:             "",
			SortingParameters:   "",
			RouteID:             "route2",
			RouteFilterIDs:      "",
			RouteAccountIDs:     "",
			RouteRateProfileIDs: "",
			RouteResourceIDs:    "",
			RouteStatIDs:        "",
			RouteWeights:        ";20.0",
			RouteBlockers:       ";false",
			RouteParameters:     "",
			Weights:             ";0",
			CreatedAt:           time.Time{},
		},
		&RouteMdl{
			PK:                  1,
			Tpid:                "TP",
			Tenant:              "cgrates.org",
			ID:                  "RoutePrf",
			FilterIDs:           "FltrRoute;*ai:~*req.AnswerTime:2017-11-27T00:00:00Z",
			Sorting:             "*weight",
			SortingParameters:   "srtPrm1",
			RouteID:             "route1",
			RouteFilterIDs:      "",
			RouteAccountIDs:     "",
			RouteRateProfileIDs: "",
			RouteResourceIDs:    "",
			RouteStatIDs:        "",
			RouteWeights:        ";10.0",
			RouteBlockers:       ";false",
			RouteParameters:     "",
			Weights:             ";10",
			CreatedAt:           time.Time{},
		},
	}
	expPrfRev := []*utils.TPRouteProfile{
		{
			TPid:              "TP",
			Tenant:            "cgrates.org",
			ID:                "RoutePrf",
			Sorting:           "*weight",
			SortingParameters: []string{"srtPrm1"},
			FilterIDs:         []string{"*ai:~*req.AnswerTime:2017-11-27T00:00:00Z", "FltrRoute"},
			Routes: []*utils.TPRoute{
				{
					ID:       "route1",
					Weights:  ";10.0",
					Blockers: ";false",
				},
				{
					ID:       "route2",
					Weights:  ";20.0",
					Blockers: ";false",
				},
			},
			Weights: ";10",
		},
	}
	rcvRev := mdlReverse.AsTPRouteProfile()
	sort.Slice(rcvRev[0].Routes, func(i, j int) bool {
		return strings.Compare(rcvRev[0].Routes[i].ID, rcvRev[0].Routes[j].ID) < 0
	})
	sort.Strings(rcvRev[0].SortingParameters)
	sort.Strings(rcvRev[0].FilterIDs)
	if !reflect.DeepEqual(rcvRev, expPrfRev) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(expPrfRev), utils.ToJSON(rcvRev))
	}
}

func TestTPRoutesAsTPRouteProfile2(t *testing.T) {
	mdl := RouteMdls{
		&RouteMdl{
			PK:                  1,
			Tpid:                "TP",
			Tenant:              "cgrates.org",
			ID:                  "RoutePrf",
			FilterIDs:           "FltrRoute;*ai:~*req.AnswerTime:2017-11-27T00:00:00Z|2017-11-28T00:00:00Z",
			Sorting:             "*weight",
			SortingParameters:   "srtPrm1",
			RouteID:             "route1",
			RouteFilterIDs:      "",
			RouteAccountIDs:     "",
			RouteRateProfileIDs: "",
			RouteResourceIDs:    "",
			RouteStatIDs:        "",
			RouteWeights:        ";10.0",
			RouteBlockers:       ";false",
			RouteParameters:     "",
			Weights:             ";10",
			CreatedAt:           time.Time{},
		},
		&RouteMdl{
			PK:                  2,
			Tpid:                "TP",
			Tenant:              "cgrates.org",
			ID:                  "RoutePrf",
			FilterIDs:           "*ai:~*req.AnswerTime:2017-11-27T00:00:00Z|2017-11-28T00:00:00Z",
			Sorting:             "",
			SortingParameters:   "",
			RouteID:             "route2",
			RouteFilterIDs:      "",
			RouteAccountIDs:     "",
			RouteRateProfileIDs: "",
			RouteResourceIDs:    "",
			RouteStatIDs:        "",
			RouteWeights:        ";20.0",
			RouteBlockers:       ";false",
			RouteParameters:     "",
			CreatedAt:           time.Time{},
		},
	}
	expPrf := []*utils.TPRouteProfile{
		{
			TPid:              "TP",
			Tenant:            "cgrates.org",
			ID:                "RoutePrf",
			Sorting:           "*weight",
			SortingParameters: []string{"srtPrm1"},
			FilterIDs:         []string{"*ai:~*req.AnswerTime:2017-11-27T00:00:00Z|2017-11-28T00:00:00Z", "FltrRoute"},
			Routes: []*utils.TPRoute{
				{
					ID:       "route1",
					Weights:  ";10.0",
					Blockers: ";false",
				},
				{
					ID:       "route2",
					Weights:  ";20.0",
					Blockers: ";false",
				},
			},
			Weights: ";10",
		},
	}
	rcv := mdl.AsTPRouteProfile()
	sort.Slice(rcv[0].Routes, func(i, j int) bool {
		return strings.Compare(rcv[0].Routes[i].ID, rcv[0].Routes[j].ID) < 0
	})
	sort.Strings(rcv[0].FilterIDs)
	if !reflect.DeepEqual(rcv, expPrf) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(expPrf), utils.ToJSON(rcv))
	}

	mdlReverse := RouteMdls{
		&RouteMdl{
			PK:                  2,
			Tpid:                "TP",
			Tenant:              "cgrates.org",
			ID:                  "RoutePrf",
			FilterIDs:           "",
			Sorting:             "",
			SortingParameters:   "",
			RouteID:             "route2",
			RouteFilterIDs:      "",
			RouteAccountIDs:     "",
			RouteRateProfileIDs: "",
			RouteResourceIDs:    "",
			RouteStatIDs:        "",
			RouteWeights:        ";20.0",
			RouteBlockers:       ";false",
			RouteParameters:     "",
			Weights:             ";0",
			CreatedAt:           time.Time{},
		},
		&RouteMdl{
			PK:                  1,
			Tpid:                "TP",
			Tenant:              "cgrates.org",
			ID:                  "RoutePrf",
			FilterIDs:           "FltrRoute;*ai:~*req.AnswerTime:2017-11-27T00:00:00Z|2017-11-28T00:00:00Z",
			Sorting:             "*weight",
			SortingParameters:   "srtPrm1",
			RouteID:             "route1",
			RouteFilterIDs:      "",
			RouteAccountIDs:     "",
			RouteRateProfileIDs: "",
			RouteResourceIDs:    "",
			RouteStatIDs:        "",
			RouteWeights:        ";10.0",
			RouteBlockers:       ";false",
			RouteParameters:     "",
			Weights:             ";10",
			CreatedAt:           time.Time{},
		},
	}
	expPrfRev := []*utils.TPRouteProfile{
		{
			TPid:              "TP",
			Tenant:            "cgrates.org",
			ID:                "RoutePrf",
			Sorting:           "*weight",
			SortingParameters: []string{"srtPrm1"},
			FilterIDs:         []string{"*ai:~*req.AnswerTime:2017-11-27T00:00:00Z|2017-11-28T00:00:00Z", "FltrRoute"},
			Routes: []*utils.TPRoute{
				{
					ID:       "route1",
					Weights:  ";10.0",
					Blockers: ";false",
				},
				{
					ID:       "route2",
					Weights:  ";20.0",
					Blockers: ";false",
				},
			},
			Weights: ";10",
		},
	}
	rcvRev := mdlReverse.AsTPRouteProfile()
	sort.Slice(rcvRev[0].Routes, func(i, j int) bool {
		return strings.Compare(rcvRev[0].Routes[i].ID, rcvRev[0].Routes[j].ID) < 0
	})
	sort.Strings(rcvRev[0].SortingParameters)
	sort.Strings(rcvRev[0].FilterIDs)
	if !reflect.DeepEqual(rcvRev, expPrfRev) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(expPrfRev), utils.ToJSON(rcvRev))
	}
}

func TestRateProfileToAPI(t *testing.T) {
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	rPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		MinCost:         utils.NewDecimal(1, 1),
		MaxCost:         utils.NewDecimal(6, 1),
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"RT_WEEK": {
				ID: "RT_WEEK",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(12, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
					{
						IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
						FixedFee:      utils.NewDecimal(234, 5),
						RecurrentFee:  utils.NewDecimal(6, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
				},
			},
			"RT_WEEKEND": {
				ID: "RT_WEEKEND",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(6, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
				},
			},
			"RT_CHRISTMAS": {
				ID: "RT_CHRISTMAS",
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(6, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
				},
			},
		},
	}
	eTPRatePrf := &utils.TPRateProfile{
		Tenant:          "cgrates.org",
		ID:              "RP1",
		FilterIDs:       []string{"*string:~*req.Subject:1001"},
		Weights:         ";0",
		MinCost:         0.1,
		MaxCost:         0.6,
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.TPRate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weights:         ";0",
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*utils.TPIntervalRate{
					{
						IntervalStart: "0",
						RecurrentFee:  0.12,
						Unit:          "60000000000",
						Increment:     "1000000000",
					},
					{
						IntervalStart: "60000000000",
						FixedFee:      0.00234,
						RecurrentFee:  0.06,
						Unit:          "60000000000",
						Increment:     "1000000000",
					},
				},
			},
			"RT_WEEKEND": {
				ID:              "RT_WEEKEND",
				Weights:         ";10",
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*utils.TPIntervalRate{
					{
						IntervalStart: "0",
						RecurrentFee:  0.06,
						Unit:          "60000000000",
						Increment:     "1000000000",
					},
				},
			},
			"RT_CHRISTMAS": {
				ID:              "RT_CHRISTMAS",
				Weights:         ";30",
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*utils.TPIntervalRate{
					{
						IntervalStart: "0",
						RecurrentFee:  0.06,
						Unit:          "60000000000",
						Increment:     "1000000000",
					},
				},
			},
		},
	}
	if rcv := RateProfileToAPI(rPrf); !reflect.DeepEqual(rcv, eTPRatePrf) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eTPRatePrf), utils.ToJSON(rcv))
	}
}

func TestAPIToRateProfile(t *testing.T) {
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	eRprf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		MinCost:         utils.NewDecimal(1, 1),
		MaxCost:         utils.NewDecimal(6, 1),
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"RT_WEEK": {
				ID: "RT_WEEK",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						FixedFee:      utils.NewDecimal(23451, 4),
						RecurrentFee:  utils.NewDecimal(12, 2),
						Unit:          minDecimal,
						Increment:     minDecimal,
					},
					{
						IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
						FixedFee:      utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(6, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
				},
			},
			"RT_WEEKEND": {
				ID: "RT_WEEKEND",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						FixedFee:      utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(6, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
				},
			},
			"RT_CHRISTMAS": {
				ID: "RT_CHRISTMAS",
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						FixedFee:      utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(6, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
				},
			},
		},
	}
	tpRprf := &utils.TPRateProfile{
		TPid:            "",
		Tenant:          "cgrates.org",
		ID:              "RP1",
		FilterIDs:       []string{"*string:~*req.Subject:1001"},
		Weights:         ";0",
		MinCost:         0.1,
		MaxCost:         0.6,
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.TPRate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weights:         ";0",
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*utils.TPIntervalRate{
					{
						IntervalStart: "0s",
						FixedFee:      2.3451,
						RecurrentFee:  0.12,
						Unit:          "1m0s",
						Increment:     "1m0s",
					},
					{
						IntervalStart: "1m0s",
						RecurrentFee:  0.06,
						Unit:          "1m0s",
						Increment:     "1s",
					},
				},
			},
			"RT_WEEKEND": {
				ID:              "RT_WEEKEND",
				Weights:         ";10",
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*utils.TPIntervalRate{
					{
						IntervalStart: "0s",
						RecurrentFee:  0.06,
						Unit:          "1m0s",
						Increment:     "1s",
					},
				},
			},
			"RT_CHRISTMAS": {
				ID:              "RT_CHRISTMAS",
				Weights:         ";30",
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*utils.TPIntervalRate{
					{
						IntervalStart: "0s",
						RecurrentFee:  0.06,
						Unit:          "1m0s",
						Increment:     "1s",
					},
				},
			},
		},
	}
	if rcv, err := APItoRateProfile(tpRprf, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, eRprf) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eRprf), utils.ToJSON(rcv))
	}
}

func TestAPItoRateProfileError(t *testing.T) {
	tpRprf := &utils.TPRateProfile{
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weights:   "0",
		Rates: map[string]*utils.TPRate{
			"RT_WEEK": {
				ID:      "RT_WEEK",
				Weights: "0",
				IntervalRates: []*utils.TPIntervalRate{
					{
						IntervalStart: "0s",
						RecurrentFee:  0.06,
						Unit:          "1ss",
						Increment:     "1ss",
					},
				},
			},
		},
	}
	expectedErr := "invalid DynamicWeight format for string <0>"
	if _, err := APItoRateProfile(tpRprf, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Error(err)
	}

	tpRprf.Weights = ";0"
	if _, err := APItoRateProfile(tpRprf, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Error(err)
	}

	tpRprf.Rates["RT_WEEK"].Weights = ";0"
	expectedErr = "time: unknown unit \"ss\" in duration \"1ss\""
	if _, err := APItoRateProfile(tpRprf, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Error(err)
	}

	tpRprf.Rates["RT_WEEK"].IntervalRates[0].Unit = "1s"
	if _, err := APItoRateProfile(tpRprf, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Error(err)
	}
}

func TestAPIToRateProfileError(t *testing.T) {
	tpRprf := &utils.TPRateProfile{
		Tenant: "cgrates.org",
		ID:     "RP1",
		Rates: map[string]*utils.TPRate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weights:         ";0",
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*utils.TPIntervalRate{
					{
						IntervalStart: "NOT_A_TIME",
						FixedFee:      2.3451,
						RecurrentFee:  0.12,
					},
				},
			},
		},
	}

	expectedErr := "can't convert <NOT_A_TIME> to decimal"
	if _, err := APItoRateProfile(tpRprf, "UTC"); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+q", expectedErr, err)
	}
}

func TestAPItoModelTPRateProfile(t *testing.T) {
	tpRprf := &utils.TPRateProfile{
		TPid:            "",
		Tenant:          "cgrates.org",
		ID:              "RP1",
		FilterIDs:       []string{"*string:~*req.Subject:1001"},
		Weights:         ";0",
		MinCost:         0.1,
		MaxCost:         0.6,
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.TPRate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weights:         ";0",
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*utils.TPIntervalRate{
					{
						IntervalStart: "0s",
						RecurrentFee:  0.12,
						Unit:          "1m0s",
						Increment:     "1m0s",
					},
					{
						IntervalStart: "1m",
						RecurrentFee:  0.06,
						Unit:          "1m0s",
						Increment:     "1s",
					},
				},
			},
		},
	}

	expModels := RateProfileMdls{
		&RateProfileMdl{
			PK:                  0,
			Tpid:                "",
			Tenant:              "cgrates.org",
			ID:                  "RP1",
			FilterIDs:           "*string:~*req.Subject:1001",
			Weights:             ";0",
			MinCost:             0.1,
			MaxCost:             0.6,
			MaxCostStrategy:     "*free",
			RateID:              "RT_WEEK",
			RateFilterIDs:       "",
			RateActivationTimes: "* * * * 1-5",
			RateWeights:         ";0",
			RateBlocker:         false,
			RateIntervalStart:   "1m",
			RateRecurrentFee:    0.06,
			RateUnit:            "1m0s",
			RateIncrement:       "1s",
			CreatedAt:           time.Time{},
		},
		&RateProfileMdl{
			PK:                  0,
			Tpid:                "",
			Tenant:              "cgrates.org",
			ID:                  "RP1",
			FilterIDs:           "",
			Weights:             ";0",
			MinCost:             0,
			MaxCost:             0,
			MaxCostStrategy:     "",
			RateID:              "RT_WEEK",
			RateFilterIDs:       "",
			RateActivationTimes: "",
			RateWeights:         ";0",
			RateBlocker:         false,
			RateIntervalStart:   "0s",
			RateRecurrentFee:    0.12,
			RateUnit:            "1m0s",
			RateIncrement:       "1m0s",
			CreatedAt:           time.Time{},
		},
	}
	expModelsRev := RateProfileMdls{
		&RateProfileMdl{
			PK:                  0,
			Tpid:                "",
			Tenant:              "cgrates.org",
			ID:                  "RP1",
			FilterIDs:           "*string:~*req.Subject:1001",
			Weights:             ";0",
			MinCost:             0.1,
			MaxCost:             0.6,
			MaxCostStrategy:     "*free",
			RateID:              "RT_WEEK",
			RateFilterIDs:       "",
			RateActivationTimes: "* * * * 1-5",
			RateWeights:         ";0",
			RateBlocker:         false,
			RateIntervalStart:   "0s",
			RateRecurrentFee:    0.12,
			RateUnit:            "1m0s",
			RateIncrement:       "1m0s",
			CreatedAt:           time.Time{},
		},
		&RateProfileMdl{
			PK:                  0,
			Tpid:                "",
			Tenant:              "cgrates.org",
			ID:                  "RP1",
			FilterIDs:           "",
			Weights:             "",
			MinCost:             0,
			MaxCost:             0,
			MaxCostStrategy:     "",
			RateID:              "RT_WEEK",
			RateFilterIDs:       "",
			RateActivationTimes: "",
			RateWeights:         "",
			RateBlocker:         false,
			RateIntervalStart:   "1m",
			RateRecurrentFee:    0.06,
			RateUnit:            "1m0s",
			RateIncrement:       "1s",
			CreatedAt:           time.Time{},
		},
	}
	rcv := APItoModelTPRateProfile(tpRprf)
	if !reflect.DeepEqual(rcv, expModels) && !reflect.DeepEqual(rcv, expModelsRev) {
		t.Errorf("Expecting: %+v or \n%+v,\nReceived: %+v", utils.ToJSON(expModels), utils.ToJSON(expModelsRev), utils.ToJSON(rcv))
	}
}

func TestAsTPRateProfile(t *testing.T) {
	rtMdl := RateProfileMdls{
		&RateProfileMdl{
			PK:                  0,
			Tpid:                "",
			Tenant:              "cgrates.org",
			ID:                  "RP1",
			FilterIDs:           "*string:~*req.Subject:1001",
			Weights:             ";0",
			MinCost:             0.1,
			MaxCost:             0.6,
			MaxCostStrategy:     "*free",
			RateID:              "RT_WEEK",
			RateFilterIDs:       "",
			RateActivationTimes: "* * * * 1-5",
			RateWeights:         ";0",
			RateBlocker:         false,
			RateIntervalStart:   "1m",
			RateRecurrentFee:    0.06,
			RateUnit:            "1m",
			RateIncrement:       "1s",
			CreatedAt:           time.Time{},
		},
		&RateProfileMdl{
			PK:                  0,
			Tpid:                "",
			Tenant:              "cgrates.org",
			ID:                  "RP1",
			FilterIDs:           "",
			Weights:             ";0",
			MinCost:             0,
			MaxCost:             0,
			MaxCostStrategy:     "",
			RateID:              "RT_WEEK",
			RateFilterIDs:       "",
			RateActivationTimes: "",
			RateWeights:         ";0",
			RateBlocker:         false,
			RateIntervalStart:   "0s",
			RateRecurrentFee:    0.12,
			RateUnit:            "1m",
			RateIncrement:       "1m",
			CreatedAt:           time.Time{},
		},
	}

	eRprf := &utils.TPRateProfile{
		TPid:            utils.EmptyString,
		Tenant:          "cgrates.org",
		ID:              "RP1",
		FilterIDs:       []string{"*string:~*req.Subject:1001"},
		Weights:         ";0",
		MinCost:         0.1,
		MaxCost:         0.6,
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.TPRate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weights:         ";0",
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*utils.TPIntervalRate{
					{
						IntervalStart: "1m",
						RecurrentFee:  0.06,
						Unit:          "1m",
						Increment:     "1s",
					},
					{
						IntervalStart: "0s",
						RecurrentFee:  0.12,
						Unit:          "1m",
						Increment:     "1m",
					},
				},
			},
		},
	}
	rcv := rtMdl.AsTPRateProfile()
	if len(rcv) != 1 {
		t.Errorf("Expecting: %+v,\nReceived: %+v", 1, len(rcv))
	} else if !reflect.DeepEqual(rcv[0], eRprf) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eRprf), utils.ToJSON(rcv[0]))
	}
}

func TestModelHelperCsvLoadError(t *testing.T) {
	type testStruct struct {
		ID        int64
		Tpid      string
		Tag       string `index:"cat" re:".*"`
		Prefix    string `index:"1" re:".*"`
		CreatedAt time.Time
	}
	var testStruct1 testStruct
	_, err := csvLoad(testStruct1, []string{"TEST_DEST", "+492"})
	if err == nil || err.Error() != "invalid testStruct.Tag index cat" {
		t.Errorf("Expecting: <invalid testStruct.Tag index cat>,\nReceived: <%+v>", err)
	}
}

func TestModelHelperCsvLoadError2(t *testing.T) {
	type testStruct struct {
		ID        int64
		Tpid      string
		Tag       string `index:"0" re:"cat"`
		Prefix    string `index:"1" re:".*"`
		CreatedAt time.Time
	}
	var testStruct1 testStruct
	_, err := csvLoad(testStruct1, []string{"TEST_DEST", "+492"})

	if err == nil || err.Error() != "invalid testStruct.Tag value TEST_DEST" {
		t.Errorf("Expecting: <invalid testStruct.Tag value TEST_DEST>,\nReceived: <%+v>", err)
	}
}

func TestModelHelpersCsvDumpError(t *testing.T) {
	type testStruct struct {
		ID        int64
		Tpid      string
		Tag       string `index:"cat" re:",*"`
		Prefix    string `index:"1" re:".*"`
		CreatedAt time.Time
	}
	var testStruct1 testStruct
	_, err := CsvDump(testStruct1)
	if err == nil || err.Error() != "invalid testStruct.Tag index cat" {
		t.Errorf("\nExpecting: <invalid testStruct.Tag index cat>,\n  Received: <%+v>", err)
	}
}

func TestAPItoModelTPRoutesCase1(t *testing.T) {
	structTest := &utils.TPRouteProfile{}
	result := APItoModelTPRoutes(structTest)
	if result != nil {
		t.Errorf("Expecting: <%+v>,\n  Received: <%+v>", nil, result)
	}
}

func TestAPItoModelTPRoutesEmptySlice(t *testing.T) {
	tpRoute := []*utils.TPRouteProfile{
		{
			TPid:              "TP1",
			Tenant:            "cgrates.org",
			ID:                "RoutePrf",
			Sorting:           "*lc",
			SortingParameters: []string{},
			Routes: []*utils.TPRoute{
				{
					ID:              "route1",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RateProfileIDs:  []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{"Stat1", "Stat2"},
					Weights:         ";10",
					Blockers:        ";false",
					RouteParameters: "SortingParam1",
				},
			},
			Weights:  ";20",
			Blockers: ";false",
		},
	}
	expMdl := RouteMdls{
		{
			Tpid:                "TP1",
			Tenant:              "cgrates.org",
			ID:                  "RoutePrf",
			Sorting:             "*lc",
			SortingParameters:   "",
			RouteID:             "route1",
			RouteFilterIDs:      "",
			RouteAccountIDs:     "",
			RouteRateProfileIDs: "",
			RouteResourceIDs:    "",
			RouteStatIDs:        "Stat1;Stat2",
			RouteWeights:        ";10",
			RouteBlockers:       ";false",
			RouteParameters:     "SortingParam1",
			Weights:             ";20",
			Blockers:            ";false",
		},
	}
	var mdl RouteMdls
	if mdl = APItoModelTPRoutes(tpRoute[0]); !reflect.DeepEqual(mdl, expMdl) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expMdl), utils.ToJSON(mdl))
	}

	//back to route profile
	//all the empty slices will be nil because of converting back an empty string into a slice
	tpRoute[0].FilterIDs = nil
	tpRoute[0].SortingParameters = nil
	tpRoute[0].Routes[0].FilterIDs = nil
	tpRoute[0].Routes[0].AccountIDs = nil
	tpRoute[0].Routes[0].RateProfileIDs = nil
	tpRoute[0].Routes[0].ResourceIDs = nil
	if newRcv := mdl.AsTPRouteProfile(); !reflect.DeepEqual(newRcv, tpRoute) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(tpRoute), utils.ToJSON(newRcv))
	}
}

func TestAPItoModelTPRoutesCase2(t *testing.T) {
	structTest := &utils.TPRouteProfile{
		TPid:              "TP1",
		Tenant:            "cgrates.org",
		ID:                "RoutePrf",
		FilterIDs:         []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z|2014-08-29T15:00:00Z"},
		Sorting:           "*lc",
		SortingParameters: []string{"PARAM1", "PARAM2"},
		Routes: []*utils.TPRoute{
			{
				ID:              "route1",
				FilterIDs:       []string{"FLTR_1", "FLTR_2"},
				AccountIDs:      []string{"Acc1", "Acc2"},
				RateProfileIDs:  []string{"RPL_1", "RPL_2"},
				ResourceIDs:     []string{"ResGroup1", "ResGroup2"},
				StatIDs:         []string{"Stat1", "Stat2"},
				Weights:         ";10",
				Blockers:        ";false",
				RouteParameters: "SortingParam1",
			},
		},
		Weights: ";20",
	}
	expStructTest := RouteMdls{{
		Tpid:                "TP1",
		Tenant:              "cgrates.org",
		ID:                  "RoutePrf",
		FilterIDs:           "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z|2014-08-29T15:00:00Z;FLTR_ACNT_dan;FLTR_DST_DE",
		Sorting:             "*lc",
		SortingParameters:   "PARAM1;PARAM2",
		RouteID:             "route1",
		RouteFilterIDs:      "FLTR_1;FLTR_2",
		RouteAccountIDs:     "Acc1;Acc2",
		RouteRateProfileIDs: "RPL_1;RPL_2",
		RouteResourceIDs:    "ResGroup1;ResGroup2",
		RouteStatIDs:        "Stat1;Stat2",
		RouteWeights:        ";10",
		RouteBlockers:       ";false",
		RouteParameters:     "SortingParam1",
		Weights:             ";20",
	},
	}
	sort.Strings(structTest.FilterIDs)
	sort.Strings(structTest.SortingParameters)
	result := APItoModelTPRoutes(structTest)
	if !reflect.DeepEqual(result, expStructTest) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStructTest), utils.ToJSON(result))
	}
}

func TestAPItoModelResourceCase1(t *testing.T) {
	var testStruct *utils.TPResourceProfile
	var testStruct2 ResourceMdls
	testStruct = nil
	result := APItoModelResource(testStruct)
	if !reflect.DeepEqual(result, testStruct2) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", testStruct2, result)
	}
}

func TestAPItoModelResourceCase2(t *testing.T) {
	testStruct := &utils.TPResourceProfile{
		Tenant:            "cgrates.org",
		TPid:              "tp_test",
		ID:                "ResGroup1",
		FilterIDs:         []string{"*ai:~*req.AnswerTime:2014-07-29T15:00:00Z|2015-07-29T15:00:00Z"},
		UsageTTL:          "Test_TTL",
		Weights:           ";10",
		Limit:             "2",
		ThresholdIDs:      []string{"TRes1", "TRes2"},
		AllocationMessage: "test",
	}
	expectedStruct := ResourceMdls{
		{
			Tenant:            "cgrates.org",
			Tpid:              "tp_test",
			ID:                "ResGroup1",
			FilterIDs:         "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z|2015-07-29T15:00:00Z",
			UsageTTL:          "Test_TTL",
			Weights:           ";10",
			Limit:             "2",
			ThresholdIDs:      "TRes1;TRes2",
			AllocationMessage: "test",
		},
	}
	result := APItoModelResource(testStruct)
	if !reflect.DeepEqual(result, expectedStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expectedStruct), utils.ToJSON(result))
	}
}

func TestAPItoModelResourceCase3(t *testing.T) {
	testStruct := &utils.TPResourceProfile{
		Tenant:            "cgrates.org",
		TPid:              "tp_test",
		ID:                "ResGroup1",
		FilterIDs:         []string{"FilterID1", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z|2015-07-29T15:00:00Z"},
		UsageTTL:          "Test_TTL",
		Weights:           ";10",
		Limit:             "2",
		ThresholdIDs:      []string{"TRes1", "TRes2"},
		AllocationMessage: "test",
	}
	expStruct := ResourceMdls{
		{
			Tenant:            "cgrates.org",
			Tpid:              "tp_test",
			ID:                "ResGroup1",
			FilterIDs:         "FilterID1",
			UsageTTL:          "Test_TTL",
			Weights:           ";10",
			Limit:             "2",
			ThresholdIDs:      "TRes1;TRes2",
			AllocationMessage: "test",
		},
		{
			Tenant:            "cgrates.org",
			Tpid:              "tp_test",
			ID:                "ResGroup1",
			FilterIDs:         "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z|2015-07-29T15:00:00Z",
			UsageTTL:          "",
			Weights:           "",
			Limit:             "",
			ThresholdIDs:      "",
			AllocationMessage: "",
		},
	}
	result := APItoModelResource(testStruct)
	if !reflect.DeepEqual(expStruct, result) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}

}

func TestRouteProfileToAPICase1(t *testing.T) {
	structTest := &RouteProfile{
		FilterIDs:         []string{"FilterID1", "FilterID2", "*ai:~*req.AnswerTime:2020-04-11T21:34:01Z|2020-04-12T21:34:01Z"},
		SortingParameters: []string{"Param1", "Param2"},
		Routes: []*Route{
			{ID: "ResGroup2"},
		},
	}

	expStruct := &utils.TPRouteProfile{
		FilterIDs:         []string{"*ai:~*req.AnswerTime:2020-04-11T21:34:01Z|2020-04-12T21:34:01Z", "FilterID1", "FilterID2"},
		SortingParameters: []string{"Param1", "Param2"},
		Routes: []*utils.TPRoute{{
			ID: "ResGroup2",
		}},
	}

	result := RouteProfileToAPI(structTest)
	sort.Strings(result.FilterIDs)
	sort.Strings(result.SortingParameters)
	if !reflect.DeepEqual(expStruct, result) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}

}

func TestRateProfileToAPIWithActInterval(t *testing.T) {
	testProfile := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001", "*ai:~*req.AnswerTime:2020-04-11T21:34:01Z|2020-04-12T21:34:01Z"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		MinCost:         utils.NewDecimal(1, 1),
		MaxCost:         utils.NewDecimal(6, 1),
		MaxCostStrategy: "*free",
		Rates:           map[string]*utils.Rate{},
	}

	expStruct := &utils.TPRateProfile{
		Tenant:          "cgrates.org",
		ID:              "RP1",
		FilterIDs:       []string{"*string:~*req.Subject:1001", "*ai:~*req.AnswerTime:2020-04-11T21:34:01Z|2020-04-12T21:34:01Z"},
		Weights:         ";0",
		MinCost:         0.1,
		MaxCost:         0.6,
		MaxCostStrategy: "*free",
		Rates:           map[string]*utils.TPRate{},
	}
	if result := RateProfileToAPI(testProfile); !reflect.DeepEqual(expStruct, result) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestAPItoModelTPRateProfileNil(t *testing.T) {
	testStruct := &utils.TPRateProfile{
		Rates: map[string]*utils.TPRate{},
	}
	result := APItoModelTPRateProfile(testStruct)
	if !reflect.DeepEqual(utils.ToJSON(result), "null") {
		t.Errorf("\nExpecting <null>,\n Received <%+v>", utils.ToJSON(result))
	}
}

func TestAPItoModelTPRateProfileCase2(t *testing.T) {
	testStruct := &utils.TPRateProfile{
		FilterIDs: []string{"test_string1", "test_string2", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z|2014-08-29T15:00:00Z"},
		Rates: map[string]*utils.TPRate{"RT_CHRISTMAS": {
			ID:              "RT_CHRISTMAS",
			FilterIDs:       []string{"test_string1", "test_string2"},
			Weights:         ";30",
			ActivationTimes: "* * 24 12 *",
			IntervalRates: []*utils.TPIntervalRate{
				{
					IntervalStart: "0s",
					RecurrentFee:  0.06,
					Unit:          "1m0s",
					Increment:     "1s",
				},
			},
		},
		},
	}
	expStruct := RateProfileMdls{{
		FilterIDs:           "test_string1;test_string2;*ai:~*req.AnswerTime:2014-07-29T15:00:00Z|2014-08-29T15:00:00Z",
		RateID:              "RT_CHRISTMAS",
		RateFilterIDs:       "test_string1;test_string2",
		RateWeights:         ";30",
		RateActivationTimes: "* * 24 12 *",
		RateIntervalStart:   "0s",
		RateRecurrentFee:    0.06,
		RateUnit:            "1m0s",
		RateIncrement:       "1s",
	}}
	result := APItoModelTPRateProfile(testStruct)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestRateProfileMdlsCSVHeader(t *testing.T) {
	testRPMdls := RateProfileMdls{}
	result := testRPMdls.CSVHeader()
	expected := []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs,
		utils.Weights, utils.ConnectFee, utils.MinCost,
		utils.MaxCost, utils.MaxCostStrategy, utils.RateID,
		utils.RateFilterIDs, utils.RateActivationStart, utils.RateWeights, utils.RateBlocker,
		utils.RateIntervalStart, utils.RateFixedFee, utils.RateRecurrentFee, utils.RateUnit, utils.RateIncrement}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, result)
	}
}

func TestDispatcherProfileToAPICase2(t *testing.T) {
	structTest := &DispatcherProfile{
		FilterIDs: []string{"field1", "field2", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-15T14:35:00Z"},
		StrategyParams: map[string]any{
			"Field1": "Params1",
		},
		Hosts: []*DispatcherHostProfile{
			{
				FilterIDs: []string{"fieldA", "fieldB"},
				Params:    map[string]any{},
			},
		},
	}

	expStruct := &utils.TPDispatcherProfile{
		FilterIDs:      []string{"*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-15T14:35:00Z", "field1", "field2"},
		StrategyParams: []any{"Params1"},
		Hosts: []*utils.TPDispatcherHostProfile{
			{
				FilterIDs: []string{"fieldA", "fieldB"},
				Params:    []any{},
			},
		},
	}

	result := DispatcherProfileToAPI(structTest)
	sort.Strings(result.FilterIDs)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestAPItoDispatcherProfileCase2(t *testing.T) {
	structTest := &utils.TPDispatcherProfile{
		FilterIDs:      []string{},
		StrategyParams: []any{"Param1"},
		Hosts: []*utils.TPDispatcherHostProfile{{
			Params: []any{"Param1"},
		}},
	}
	expStruct := &DispatcherProfile{
		FilterIDs: []string{},
		StrategyParams: map[string]any{
			"0": "Param1",
		},
		Hosts: DispatcherHostProfiles{{
			FilterIDs: []string{},
			Params: map[string]any{
				"0": "Param1",
			},
		},
		},
	}
	result := APItoDispatcherProfile(structTest, "")
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestAPItoModelTPDispatcherProfileNil(t *testing.T) {
	var structTest *utils.TPDispatcherProfile = nil
	expected := "null"
	result := APItoModelTPDispatcherProfile(structTest)
	if !reflect.DeepEqual(utils.ToJSON(result), expected) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, utils.ToJSON(result))
	}
}

func TestModelHelpersParamsToString(t *testing.T) {
	testInterface := []any{"Param1", "Param2"}
	result := paramsToString(testInterface)
	if !reflect.DeepEqual(result, "Param1;Param2") {
		t.Errorf("\nExpecting <Param1;Param2>,\n Received <%+v>", result)
	}
}

func TestModelHelpersAsTPDispatcherProfiles(t *testing.T) {
	structTest := DispatcherProfileMdls{
		&DispatcherProfileMdl{
			FilterIDs:          "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z|2014-08-29T15:00:00Z",
			StrategyParameters: "Param1",
		},
	}
	expStruct := []*utils.TPDispatcherProfile{{
		FilterIDs:      []string{"*ai:~*req.AnswerTime:2014-07-29T15:00:00Z|2014-08-29T15:00:00Z"},
		StrategyParams: []any{"Param1"},
	},
	}
	result := structTest.AsTPDispatcherProfiles()
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestTPDispatcherProfilesCSVHeader(t *testing.T) {
	structTest := DispatcherProfileMdls{
		&DispatcherProfileMdl{
			Tpid:           "TP1",
			Tenant:         "cgrates.org",
			ID:             "Dsp",
			FilterIDs:      "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z;FLTR_ACNT_dan;FLTR_DST_DE",
			Strategy:       utils.MetaFirst,
			Weight:         20,
			ConnID:         "C1",
			ConnWeight:     10,
			ConnBlocker:    false,
			ConnParameters: "192.168.54.203",
		},
		&DispatcherProfileMdl{
			Tpid:           "TP1",
			Tenant:         "cgrates.org",
			ID:             "Dsp",
			ConnID:         "C2",
			ConnWeight:     10,
			ConnBlocker:    false,
			ConnParameters: "192.168.54.204",
		},
	}
	expected := []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs, utils.Weight,
		utils.Strategy, utils.StrategyParameters, utils.ConnID, utils.ConnFilterIDs,
		utils.ConnWeight, utils.ConnBlocker, utils.ConnParameters}
	result := structTest.CSVHeader()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, result)
	}
}

func TestChargerProfileToAPILastCase(t *testing.T) {
	testStruct := &ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "CPP_1",
		FilterIDs:    []string{"*string:~*opts.*subsys:*chargers", "FLTR_CP_1", "FLTR_CP_4", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
		RunID:        "TestRunID",
		AttributeIDs: []string{"*none"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	expStruct := &utils.TPChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "CPP_1",
		FilterIDs:    []string{"*ai:~*req.AnswerTime:2014-07-14T14:25:00Z", "*string:~*opts.*subsys:*chargers", "FLTR_CP_1", "FLTR_CP_4"},
		AttributeIDs: []string{"*none"},
		RunID:        "TestRunID",
		Weights:      ";20",
	}

	result := ChargerProfileToAPI(testStruct)
	sort.Strings(result.FilterIDs)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestRateProfileMdlsAsTPRateProfileCase2(t *testing.T) {
	testRPMdls := RateProfileMdls{&RateProfileMdl{
		Tpid:            "",
		Tenant:          "cgrates.org",
		ID:              "RP1",
		FilterIDs:       "*string:~*req.Subject:1001;*ai:~*req.AnswerTime:2014-07-29T15:00:00Z|2014-08-29T15:00:00Z",
		Weights:         ";1.2",
		MinCost:         0.1,
		MaxCost:         0.6,
		MaxCostStrategy: "*free",
		RateID:          "0",
		RateFilterIDs:   "test_filter_id",
		RateWeights:     ";2",
	},
	}
	expStruct := []*utils.TPRateProfile{
		{TPid: "",
			Tenant:          "cgrates.org",
			ID:              "RP1",
			FilterIDs:       []string{"*ai:~*req.AnswerTime:2014-07-29T15:00:00Z|2014-08-29T15:00:00Z", "*string:~*req.Subject:1001"},
			Weights:         ";1.2",
			MinCost:         0.1,
			MaxCost:         0.6,
			MaxCostStrategy: "*free",
			Rates: map[string]*utils.TPRate{
				"0": {
					ID:        "0",
					FilterIDs: []string{"test_filter_id"},
					Weights:   ";2",
					IntervalRates: []*utils.TPIntervalRate{
						{
							IntervalStart: "",
							FixedFee:      0,
							RecurrentFee:  0,
							Unit:          "",
							Increment:     "",
						},
					},
				},
			},
		},
	}
	result := testRPMdls.AsTPRateProfile()
	sort.Strings(result[0].FilterIDs)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}

}

func TestRateProfileMdlsAsTPRateProfileCase3(t *testing.T) {
	testRPMdls := RateProfileMdls{&RateProfileMdl{
		Tpid:            "",
		Tenant:          "cgrates.org",
		ID:              "RP1",
		FilterIDs:       "*string:~*req.Subject:1001;*ai:~*req.AnswerTime:2014-07-29T15:00:00Z",
		Weights:         ";1.2",
		MinCost:         0.1,
		MaxCost:         0.6,
		MaxCostStrategy: "*free",
		RateID:          "0",
		RateFilterIDs:   "test_filter_id",
		RateWeights:     ";2",
	},
	}
	expStruct := []*utils.TPRateProfile{
		{TPid: "",
			Tenant:          "cgrates.org",
			ID:              "RP1",
			FilterIDs:       []string{"*ai:~*req.AnswerTime:2014-07-29T15:00:00Z", "*string:~*req.Subject:1001"},
			Weights:         ";1.2",
			MinCost:         0.1,
			MaxCost:         0.6,
			MaxCostStrategy: "*free",
			Rates: map[string]*utils.TPRate{
				"0": {
					ID:        "0",
					FilterIDs: []string{"test_filter_id"},
					Weights:   ";2",
					IntervalRates: []*utils.TPIntervalRate{
						{
							IntervalStart: "",
							FixedFee:      0,
							RecurrentFee:  0,
							Unit:          "",
							Increment:     "",
						},
					},
				},
			},
		},
	}
	result := testRPMdls.AsTPRateProfile()
	sort.Strings(result[0].FilterIDs)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}

}

func TestAPItoModelTPDispatcherProfileCase2(t *testing.T) {
	structTest := &utils.TPDispatcherProfile{
		FilterIDs: []string{"*ai:~*req.AnswerTime:2014-07-29T15:00:00Z|2014-07-30T15:00:00Z"},
	}
	expStruct := DispatcherProfileMdls{{
		FilterIDs: "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z|2014-07-30T15:00:00Z",
	},
	}
	result := APItoModelTPDispatcherProfile(structTest)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func ModelHelpersTestStatMdlsCSVHeader(t *testing.T) {
	testStruct := ResourceMdls{
		{
			Tpid:         "TEST_TPID",
			Tenant:       "cgrates.org",
			ID:           "ResGroup1",
			FilterIDs:    "FLTR_RES_GR1;*ai:~*req.AnswerTime:2014-07-29T15:00:00Z",
			Stored:       false,
			Blocker:      false,
			Weights:      ";10",
			Limit:        "45",
			ThresholdIDs: "WARN_RES1;WARN_RES1",
		},
	}
	expStruct := []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs, utils.Weights,
		utils.UsageTTL, utils.Limit, utils.AllocationMessage, utils.Blocker, utils.Stored,
		utils.ThresholdIDs}
	result := testStruct.CSVHeader()
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestThresholdMdlsCSVHeader(t *testing.T) {
	testStruct := ThresholdMdls{
		{
			Tpid:   "test_tpid",
			Tenant: "test_tenant",
		},
	}
	expStruct := []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs, utils.Weights,
		utils.MaxHits, utils.MinHits, utils.MinSleep,
		utils.Blocker, utils.ActionProfileIDs, utils.Async}
	result := testStruct.CSVHeader()
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestTrendMdlsCSVHeader(t *testing.T) {
	testStruct := TrendMdls{
		{
			Tpid:   "tpid",
			Tenant: "cgrates.org",
		},
	}

	expStruct := []string{"#" + utils.Tenant, utils.ID, utils.Schedule, utils.StatID,
		utils.Metrics, utils.TTL, utils.QueueLength,
		utils.MinItems, utils.CorrelationType, utils.Tolerance, utils.Stored, utils.ThresholdIDs}
	result := testStruct.CSVHeader()
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}
func TestActionProfileMdlsCSVHeader(t *testing.T) {
	testStruct := ActionProfileMdls{
		{
			Tpid:   "test_tpid",
			Tenant: "test_tenant",
		},
	}
	expStruct := []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs,
		utils.Weights, utils.Blockers, utils.Schedule, utils.TargetType,
		utils.TargetIDs, utils.ActionID, utils.ActionFilterIDs, utils.ActionTTL,
		utils.ActionType, utils.ActionOpts, utils.ActionPath, utils.ActionValue,
	}
	result := testStruct.CSVHeader()
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestActionProfileMdlsAsTPActionProfileTimeLen1(t *testing.T) {
	testStruct := ActionProfileMdls{
		{
			Tpid:            "test_id",
			Tenant:          "cgrates.org",
			ID:              "RP1",
			FilterIDs:       "*string:~*req.Subject:1001;*ai:~*req.AnswerTime:2014-07-29T15:00:00Z",
			Weights:         ";1",
			Schedule:        "test_schedule",
			ActionID:        "test_action_id",
			ActionFilterIDs: "test_action_filter_ids",
		},
	}
	expStruct := []*utils.TPActionProfile{
		{
			TPid:      "test_id",
			Tenant:    "cgrates.org",
			ID:        "RP1",
			FilterIDs: []string{"*ai:~*req.AnswerTime:2014-07-29T15:00:00Z", "*string:~*req.Subject:1001"},
			Weights:   ";1",
			Schedule:  "test_schedule",
			Actions: []*utils.TPAPAction{
				{
					ID:        "test_action_id",
					FilterIDs: []string{"test_action_filter_ids"},
					Diktats:   []*utils.TPAPDiktat{{}},
				},
			},
		},
	}
	result := testStruct.AsTPActionProfile()
	sort.Strings(result[0].FilterIDs)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting %s,\n Received %s", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestActionProfileMdlsAsTPActionProfile(t *testing.T) {
	testStruct := ActionProfileMdls{
		{
			Tpid:            "test_id",
			Tenant:          "cgrates.org",
			ID:              "RP1",
			FilterIDs:       "*string:~*req.Subject:1001;*ai:~*req.AnswerTime:2014-07-29T15:00:00Z|2014-08-29T15:00:00Z",
			Weights:         ";1",
			Schedule:        "test_schedule",
			TargetType:      utils.MetaAccounts,
			TargetIDs:       "test_account_id1;test_account_id2",
			ActionID:        "test_action_id",
			ActionFilterIDs: "test_action_filter_ids",
			Blockers:        ";false",
		},
	}
	expStruct := []*utils.TPActionProfile{
		{
			TPid:      "test_id",
			Tenant:    "cgrates.org",
			ID:        "RP1",
			FilterIDs: []string{"*ai:~*req.AnswerTime:2014-07-29T15:00:00Z|2014-08-29T15:00:00Z", "*string:~*req.Subject:1001"},
			Weights:   ";1",
			Blockers:  ";false",
			Schedule:  "test_schedule",
			Targets: []*utils.TPActionTarget{
				{
					TargetType: utils.MetaAccounts,
					TargetIDs:  []string{"test_account_id1", "test_account_id2"},
				},
			},
			Actions: []*utils.TPAPAction{
				{
					ID:        "test_action_id",
					FilterIDs: []string{"test_action_filter_ids"},
					Diktats:   []*utils.TPAPDiktat{{}},
				},
			},
		},
	}

	result := testStruct.AsTPActionProfile()
	sort.Strings(result[0].FilterIDs)
	sort.Slice(result[0].Targets[0].TargetIDs, func(i, j int) bool {
		return result[0].Targets[0].TargetIDs[i] < result[0].Targets[0].TargetIDs[j]
	})
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestAPItoModelTPActionProfileTPActionProfileNil(t *testing.T) {
	testStruct := &utils.TPActionProfile{}
	var expStruct ActionProfileMdls = nil
	result := APItoModelTPActionProfile(testStruct)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestAPItoModelTPActionProfileTPActionProfile(t *testing.T) {
	testStruct := &utils.TPActionProfile{
		TPid:      "test_id",
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001", "*string:~*req.Subject:1002", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z|2014-08-29T15:00:00Z"},
		Weights:   ";1",
		Schedule:  "test_schedule",
		Targets: []*utils.TPActionTarget{{
			TargetType: utils.MetaAccounts,
			TargetIDs:  []string{"test_account_id1", "test_account_id2"},
		}},
		Actions: []*utils.TPAPAction{
			{
				ID:        "test_action_id",
				FilterIDs: []string{"test_action_filter_id1", "test_action_filter_id2"},
				Diktats:   []*utils.TPAPDiktat{{}},
			},
		},
	}

	expStruct := ActionProfileMdls{{
		Tpid:            "test_id",
		Tenant:          "cgrates.org",
		ID:              "RP1",
		FilterIDs:       "*string:~*req.Subject:1001;*string:~*req.Subject:1002;*ai:~*req.AnswerTime:2014-07-29T15:00:00Z|2014-08-29T15:00:00Z",
		Weights:         ";1",
		Schedule:        "test_schedule",
		TargetType:      utils.MetaAccounts,
		TargetIDs:       "test_account_id1;test_account_id2",
		ActionID:        "test_action_id",
		ActionFilterIDs: "test_action_filter_id1;test_action_filter_id2",
	}}
	result := APItoModelTPActionProfile(testStruct)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting %s,\n Received %s", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestModelHelpersAPItoActionProfile(t *testing.T) {
	testStruct := &utils.TPActionProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001", "*string:~*req.Subject:1002", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-15T14:25:00Z"},
		Weights:   ";1",
		Schedule:  "test_schedule",
		Targets: []*utils.TPActionTarget{
			{
				TargetType: utils.MetaAccounts,
				TargetIDs:  []string{"test_account_id1", "test_account_id2"},
			},
			{
				TargetType: utils.MetaResources,
				TargetIDs:  []string{"test_ID1", "test_ID2"},
			},
		},
		Actions: []*utils.TPAPAction{
			{
				ID:        "test_action_id",
				FilterIDs: []string{"test_action_filter_id1", "test_action_filter_id2"},
				Diktats: []*utils.TPAPDiktat{{
					Path: "test_path",
				}},
				Opts: "key1:val1;key2:val2",
			},
		},
	}

	expStruct := &ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-15T14:25:00Z", "*string:~*req.Subject:1001", "*string:~*req.Subject:1002"},
		Weights: utils.DynamicWeights{
			{
				Weight: 1,
			},
		},
		Schedule: "test_schedule",
		Targets: map[string]utils.StringSet{
			utils.MetaAccounts:  utils.NewStringSet([]string{"test_account_id1", "test_account_id2"}),
			utils.MetaResources: utils.NewStringSet([]string{"test_ID1", "test_ID2"}),
		},
		Actions: []*APAction{
			{
				ID:        "test_action_id",
				FilterIDs: []string{"test_action_filter_id1", "test_action_filter_id2"},
				Diktats: []*APDiktat{{
					Path: "test_path",
				}},
				Opts: map[string]any{
					"key1": "val1",
					"key2": "val2",
				},
			},
		},
	}
	result, _ := APItoActionProfile(testStruct, "")
	sort.Strings(result.FilterIDs)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}

}

func TestModelHelpersAPItoActionProfileError3(t *testing.T) {
	testStruct := &utils.TPActionProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001", "*string:~*req.Subject:1002", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-15T14:25:00Z"},
		Weights:   ";1",
		Schedule:  "test_schedule",
		Actions: []*utils.TPAPAction{
			{
				ID:        "test_action_id",
				FilterIDs: []string{"test_action_filter_id1", "test_action_filter_id2"},
				Diktats: []*utils.TPAPDiktat{{
					Path: "test_path",
				}},
				TTL: "cat",
			},
		},
	}

	_, err := APItoActionProfile(testStruct, "")
	if err == nil || err.Error() != "time: invalid duration \"cat\"" {
		t.Errorf("\nExpecting <time: invalid duration \"cat\">,\n Received <%+v>", err)
	}
}

func TestModelHelpersAPItoActionProfileError4(t *testing.T) {
	testStruct := &utils.TPActionProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001", "*string:~*req.Subject:1002", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-15T14:25:00Z"},
		Weights:   ";1",
		Schedule:  "test_schedule",
		Actions: []*utils.TPAPAction{
			{
				ID:        "test_action_id",
				FilterIDs: []string{"test_action_filter_id1", "test_action_filter_id2"},
				Diktats: []*utils.TPAPDiktat{{
					Path: "test_path",
				}},
				Opts: "test_opt",
			},
		},
	}

	_, err := APItoActionProfile(testStruct, "")
	if err == nil || err.Error() != "malformed option for ActionProfile <cgrates.org:RP1> for action <test_action_id>" {
		t.Errorf("\nExpecting <malformed option for ActionProfile <cgrates.org:RP1> for action <test_action_id>>,\n Received <%+v>", err)
	}
}

func TestModelHelpersActionProfileToAPI(t *testing.T) {
	testStruct := &ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001", "*string:~*req.Subject:1002", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-15T14:25:00Z"},
		Weights: utils.DynamicWeights{
			{
				Weight: 1,
			},
		},
		Schedule: "test_schedule",
		Actions: []*APAction{
			{
				ID:        "test_action_id",
				FilterIDs: []string{"test_action_filter_id1", "test_action_filter_id2"},
				TTL:       time.Second,
				Diktats: []*APDiktat{{
					Path: "test_path",
				}},
				Opts: map[string]any{
					"key1": "val1",
				},
			},
		},
	}
	expStruct := &utils.TPActionProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001", "*string:~*req.Subject:1002", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-15T14:25:00Z"},
		Weights:   ";1",
		Schedule:  "test_schedule",
		Targets:   []*utils.TPActionTarget{},
		Actions: []*utils.TPAPAction{
			{
				ID:        "test_action_id",
				FilterIDs: []string{"test_action_filter_id1", "test_action_filter_id2"},
				TTL:       "1s",
				Diktats: []*utils.TPAPDiktat{{
					Path: "test_path",
				}},
				Opts: "key1:val1",
			},
		},
	}
	result := ActionProfileToAPI(testStruct)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestChargerMdlsCSVHeader(t *testing.T) {

	testStruct := ChargerMdls{
		{
			Tenant: "cgrates.org",
			ID:     "RP1",
		},
	}
	expStruct := []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs, utils.Weights,
		utils.Blockers, utils.RunID, utils.AttributeIDs}

	result := testStruct.CSVHeader()
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestAPItoAttributeProfileError1(t *testing.T) {
	tpAlsPrf := &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "*string:~*opts.*context:con1"},
		Attributes: []*utils.TPAttribute{
			{
				Path:  "",
				Value: "Al1",
			},
		},
		Weights: ";20",
	}

	_, err := APItoAttributeProfile(tpAlsPrf, "UTC")
	if err == nil || err.Error() != "empty path in AttributeProfile <cgrates.org:ALS1>" {
		t.Errorf("\nExpecting <empty path in AttributeProfile <cgrates.org:ALS1>>,\n Received <%+v>", err)
	}

}

func TestAPItoAttributeProfileError2(t *testing.T) {
	tpAlsPrf := &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "*string:~*opts.*context:con1"},
		Attributes: []*utils.TPAttribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FL1",
				Value: "\"constant;`>;q=0.7;expires=3600constant\"",
			},
		},
		Weights: ";20",
	}

	_, err := APItoAttributeProfile(tpAlsPrf, "UTC")
	expected := "Closed unspilit syntax"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, err)
	}

}

func TestAPItoModelTPAttributeNoAttributes(t *testing.T) {
	testStruct := &utils.TPAttributeProfile{}
	var expStruct AttributeMdls = nil
	result := APItoModelTPAttribute(testStruct)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestAttributeMdlsCSVHeader(t *testing.T) {
	testStruct := AttributeMdls{
		{
			Tenant: "cgrates.org",
			ID:     "ALS1",
		},
	}
	expStruct := []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs, utils.Weights, utils.Blockers, utils.AttributeFilterIDs, utils.AttributeBlockers, utils.Path, utils.Type, utils.Value}
	result := testStruct.CSVHeader()
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestModelHelpersTestAPItoRouteProfile(t *testing.T) {
	testStruct := &utils.TPRouteProfile{
		FilterIDs:         []string{},
		SortingParameters: []string{"param1"},
		Routes:            []*utils.TPRoute{},
	}
	expStruct := &RouteProfile{
		FilterIDs:         []string{},
		SortingParameters: []string{"param1"},
		Routes:            []*Route{},
	}
	result, err := APItoRouteProfile(testStruct, "")
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestModelHelperAPItoFilterError(t *testing.T) {
	testStruct := &utils.TPFilterProfile{
		Filters: []*utils.TPFilter{{
			Type:    "test_type",
			Element: "",
			Values:  []string{"val1"},
		},
		},
	}

	_, err := APItoFilter(testStruct, "")
	if err == nil || err.Error() != "empty RSRParser in rule: <>" {
		t.Errorf("\nExpecting <empty RSRParser in rule: <>>,\n Received <%+v>", err)
	}

}

func TestFilterMdlsCSVHeader(t *testing.T) {
	testStruct := FilterMdls{{
		Tpid:   "test_tpid",
		Tenant: "test_tenant",
	}}
	expStruct := []string{"#" + utils.Tenant, utils.ID, utils.Type, utils.Element,
		utils.Values}
	result := testStruct.CSVHeader()
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}

}

func TestModelHelpersThresholdProfileToAPIExpTime(t *testing.T) {
	testStruct := &ThresholdProfile{
		FilterIDs:        []string{"test_filter_id", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-15T14:25:00Z"},
		ActionProfileIDs: []string{"test_action_id"},
	}
	expStruct := &utils.TPThresholdProfile{
		FilterIDs:        []string{"test_filter_id", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-15T14:25:00Z"},
		ActionProfileIDs: []string{"test_action_id"},
	}
	result := ThresholdProfileToAPI(testStruct)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestModelHelpersAPItoThresholdProfileError1(t *testing.T) {
	testStruct := &utils.TPThresholdProfile{
		TPid:             "",
		Tenant:           "",
		ID:               "",
		FilterIDs:        nil,
		MaxHits:          0,
		MinHits:          0,
		MinSleep:         "cat",
		Blocker:          false,
		Weights:          ";0",
		ActionProfileIDs: nil,
		Async:            false,
	}
	_, err := APItoThresholdProfile(testStruct, "")
	if err == nil || err.Error() != "time: invalid duration \"cat\"" {
		t.Errorf("\nExpecting <time: invalid duration \"cat\">,\n Received <%+v>", err)
	}
}

func TestModelHelpersAPItoModelTPThresholdExpTime1(t *testing.T) {
	testStruct := &utils.TPThresholdProfile{
		TPid:             "TP1",
		Tenant:           "cgrates.org",
		ID:               "TH_1",
		FilterIDs:        []string{"*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-15T14:35:00Z"},
		MaxHits:          12,
		MinHits:          10,
		MinSleep:         "1s",
		Blocker:          false,
		Weights:          ";20",
		ActionProfileIDs: []string{"WARN3", "LOG"},
	}
	expStruct := ThresholdMdls{
		{
			Tpid:             "TP1",
			Tenant:           "cgrates.org",
			ID:               "TH_1",
			FilterIDs:        "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-15T14:35:00Z",
			MaxHits:          12,
			MinHits:          10,
			MinSleep:         "1s",
			Blocker:          false,
			Weights:          ";20",
			ActionProfileIDs: "WARN3",
		},
		{
			Tpid:             "TP1",
			Tenant:           "cgrates.org",
			ID:               "TH_1",
			ActionProfileIDs: "LOG",
		},
	}

	result := APItoModelTPThreshold(testStruct)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestModelHelpersAPItoModelTPThresholdExpTime2(t *testing.T) {
	testStruct := &utils.TPThresholdProfile{
		TPid:             "TP1",
		Tenant:           "cgrates.org",
		ID:               "TH_1",
		FilterIDs:        []string{"FilterID1", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-15T14:35:00Z"},
		MaxHits:          12,
		MinHits:          10,
		MinSleep:         "1s",
		Blocker:          false,
		Weights:          ";20",
		ActionProfileIDs: []string{"WARN3"},
	}
	expStruct := ThresholdMdls{
		{
			Tpid:             "TP1",
			Tenant:           "cgrates.org",
			ID:               "TH_1",
			FilterIDs:        "FilterID1",
			MaxHits:          12,
			MinHits:          10,
			MinSleep:         "1s",
			Blocker:          false,
			Weights:          ";20",
			ActionProfileIDs: "WARN3",
		},
		{
			Tpid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "TH_1",
			FilterIDs: "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-15T14:35:00Z",
		},
	}

	result := APItoModelTPThreshold(testStruct)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestThresholdMdlsAsTPThresholdActivationTime(t *testing.T) {
	testStruct := ThresholdMdls{
		{
			Tpid:             "",
			Tenant:           "",
			ID:               "",
			FilterIDs:        "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-15T14:35:00Z",
			MaxHits:          0,
			MinHits:          0,
			MinSleep:         "",
			Blocker:          false,
			Weights:          "",
			ActionProfileIDs: "",
			Async:            false,
		},
	}
	expStruct := []*utils.TPThresholdProfile{
		{
			TPid:      "",
			Tenant:    "",
			ID:        "",
			FilterIDs: []string{"*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-15T14:35:00Z"},
			MaxHits:   0,
			MinHits:   0,
			MinSleep:  "",
			Blocker:   false,
			Weights:   "",
			Async:     false,
		},
	}
	result := testStruct.AsTPThreshold()
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestModelHelpersStatQueueProfileToAPIFilterIds(t *testing.T) {
	testStruct := &StatQueueProfile{
		Tenant:      "",
		ID:          "",
		FilterIDs:   []string{"test_filter_Id", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-15T14:25:00Z"},
		QueueLength: 0,
		MinItems:    0,
		Metrics: []*MetricWithFilters{{
			FilterIDs: []string{"test_id"},
			Blockers: utils.DynamicBlockers{{
				FilterIDs: []string{"fltr2"},
				Blocker:   false,
			}},
		},
		},
		Stored:   false,
		Blockers: utils.DynamicBlockers{{Blocker: false}},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		ThresholdIDs: []string{"threshold_id"},
	}
	expStruct := &utils.TPStatProfile{
		Tenant:      "",
		ID:          "",
		FilterIDs:   []string{"test_filter_Id", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-15T14:25:00Z"},
		QueueLength: 0,
		MinItems:    0,
		Metrics: []*utils.MetricWithFilters{
			{
				FilterIDs: []string{"test_id"},
				Blockers:  "fltr2;false",
			},
		},
		Blockers:     ";false",
		Stored:       false,
		Weights:      ";0",
		ThresholdIDs: []string{"threshold_id"},
	}
	result := StatQueueProfileToAPI(testStruct)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestModelHelpersAPItoStatsError1(t *testing.T) {
	testStruct := &utils.TPStatProfile{
		TPid:         "",
		Tenant:       "",
		ID:           "",
		QueueLength:  0,
		TTL:          "cat",
		Blockers:     ";false",
		Stored:       false,
		Weights:      ";0",
		MinItems:     0,
		ThresholdIDs: nil,
	}
	_, err := APItoStats(testStruct, "")
	if err == nil || err.Error() != "time: invalid duration \"cat\"" {
		t.Errorf("\nExpecting <time: invalid duration \"cat\">,\n Received <%+v>", err)
	}
}

func TestModelHelpersAPItoModelStatsCase2(t *testing.T) {
	testStruct := &utils.TPStatProfile{
		TPid:        "TPS1",
		Tenant:      "cgrates.org",
		ID:          "Stat1",
		FilterIDs:   []string{"*string:Account:1002", "*string:Account:1003", "*ai:~*req.AnswerTime:2014-07-25T15:00:00Z|2014-07-26T15:00:00Z"},
		QueueLength: 100,
		TTL:         "1s",
		Metrics: []*utils.MetricWithFilters{
			{
				FilterIDs: []string{"test_filter_id1", "test_filter_id2"},
				MetricID:  "*tcc",
				Blockers:  ";false",
			},
		},
		Blockers:     ";true",
		Stored:       true,
		Weights:      ";20",
		MinItems:     2,
		ThresholdIDs: []string{"Th1", "Th2"},
	}
	expStruct := StatMdls{
		&StatMdl{
			Tpid:            "TPS1",
			Tenant:          "cgrates.org",
			ID:              "Stat1",
			FilterIDs:       "*string:Account:1002;*string:Account:1003;*ai:~*req.AnswerTime:2014-07-25T15:00:00Z|2014-07-26T15:00:00Z",
			QueueLength:     100,
			TTL:             "1s",
			MinItems:        2,
			MetricIDs:       "*tcc",
			MetricFilterIDs: "test_filter_id1;test_filter_id2",
			MetricBlockers:  ";false",
			Stored:          true,
			Blockers:        ";true",
			Weights:         ";20",
			ThresholdIDs:    "Th1;Th2",
		},
	}
	result := APItoModelStats(testStruct)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestStatMdlsAsTPStatsCase2(t *testing.T) {
	testStruct := StatMdls{{
		FilterIDs:       "*ai:~*req.AnswerTime:2014-07-25T15:00:00Z|2014-07-26T15:00:00Z",
		MetricIDs:       "test_id",
		MetricFilterIDs: "test_filter_id",
	}}
	expStruct := []*utils.TPStatProfile{{
		FilterIDs: []string{"*ai:~*req.AnswerTime:2014-07-25T15:00:00Z|2014-07-26T15:00:00Z"},
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID:  "test_id",
				FilterIDs: []string{"test_filter_id"},
			},
		},
	}}
	result := testStruct.AsTPStats()
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestStatMdlsCSVHeader(t *testing.T) {
	testStruct := StatMdls{{
		PK:              0,
		Tpid:            "",
		Tenant:          "test_tenant",
		ID:              "test_id",
		FilterIDs:       "test_filter_id",
		QueueLength:     0,
		TTL:             "",
		MinItems:        0,
		MetricIDs:       "",
		MetricFilterIDs: "",
		Stored:          false,
		Blockers:        ";false",
		Weights:         ";0",
		ThresholdIDs:    "",
		CreatedAt:       time.Time{},
	}}
	expStruct := []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs, utils.Weights, utils.Blockers, utils.QueueLength, utils.TTL, utils.MinItems, utils.Stored, utils.ThresholdIDs, utils.MetricIDs, utils.MetricFilterIDs, utils.MetricBlockers}
	result := testStruct.CSVHeader()
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestModelHelpersResourceProfileToAPICase2(t *testing.T) {
	testStruct := &ResourceProfile{
		Tenant:            "",
		ID:                "",
		FilterIDs:         []string{"test_filter_id", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-15T14:25:00Z"},
		UsageTTL:          time.Second,
		Limit:             0,
		AllocationMessage: "",
		Blocker:           false,
		Stored:            false,
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		ThresholdIDs: []string{"test_threshold_id"},
	}
	expStruct := &utils.TPResourceProfile{
		TPid:              "",
		Tenant:            "",
		ID:                "",
		FilterIDs:         []string{"test_filter_id", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-15T14:25:00Z"},
		UsageTTL:          "1s",
		Limit:             "0",
		AllocationMessage: "",
		Blocker:           false,
		Stored:            false,
		Weights:           ";0",
		ThresholdIDs:      []string{"test_threshold_id"},
	}
	result := ResourceProfileToAPI(testStruct)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestModelHelpersAPItoResourceError1(t *testing.T) {
	testStruct := &utils.TPResourceProfile{
		TPid:              "",
		Tenant:            "",
		ID:                "",
		FilterIDs:         nil,
		UsageTTL:          "cat",
		Limit:             "",
		AllocationMessage: "",
		Blocker:           false,
		Stored:            false,
		Weights:           ";0",
		ThresholdIDs:      nil,
	}
	_, err := APItoResource(testStruct, "")
	if err == nil || err.Error() != "time: invalid duration \"cat\"" {
		t.Errorf("\nExpecting <time: invalid duration \"cat\">,\n Received <%+v>", err)
	}
}

func TestModelHelpersAPItoResourceError3(t *testing.T) {
	testStruct := &utils.TPResourceProfile{
		TPid:              "",
		Tenant:            "",
		ID:                "",
		FilterIDs:         nil,
		UsageTTL:          "",
		Limit:             "cat",
		AllocationMessage: "",
		Blocker:           false,
		Stored:            false,
		Weights:           ";0",
		ThresholdIDs:      nil,
	}
	_, err := APItoResource(testStruct, "")
	if err == nil || err.Error() != "strconv.ParseFloat: parsing \"cat\": invalid syntax" {
		t.Errorf("\nExpecting <strconv.ParseFloat: parsing \"cat\": invalid syntax>,\n Received <%+v>", err)
	}
}

func TestTpResourcesAsTpResources2(t *testing.T) {
	testStruct := []*ResourceMdl{
		{
			Tpid:         "TEST_TPID",
			Tenant:       "cgrates.org",
			ID:           "ResGroup1",
			FilterIDs:    "FLTR_RES_GR1;*ai:~*req.AnswerTime:2014-07-27T15:00:00Z|2014-07-28T15:00:00Z",
			ThresholdIDs: "WARN_RES1",
		},
	}
	expStruct := []*utils.TPResourceProfile{
		{
			TPid:         "TEST_TPID",
			Tenant:       "cgrates.org",
			ID:           "ResGroup1",
			FilterIDs:    []string{"*ai:~*req.AnswerTime:2014-07-27T15:00:00Z|2014-07-28T15:00:00Z", "FLTR_RES_GR1"},
			ThresholdIDs: []string{"WARN_RES1"},
		},
	}
	result := ResourceMdls(testStruct).AsTPResources()
	sort.Strings(result[0].FilterIDs)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestModelHelpersCSVLoadErrorInt(t *testing.T) {
	type testStruct struct {
		ID        int64
		Tpid      string
		Tag       int `index:"0" re:".*"`
		CreatedAt time.Time
	}

	_, err := csvLoad(testStruct{}, []string{"TEST_DEST"})
	if err == nil || err.Error() != "invalid value \"TEST_DEST\" for field testStruct.Tag" {
		t.Errorf("\nExpecting <invalid value \"TEST_DEST\" for field testStruct.Tag>,\n Received <%+v>", err)
	}
}

func TestModelHelpersCSVLoadErrorFloat64(t *testing.T) {
	type testStruct struct {
		ID        int64
		Tpid      string
		Tag       float64 `index:"0" re:".*"`
		CreatedAt time.Time
	}

	_, err := csvLoad(testStruct{}, []string{"TEST_DEST"})
	if err == nil || err.Error() != "invalid value \"TEST_DEST\" for field testStruct.Tag" {
		t.Errorf("\nExpecting <invalid value \"TEST_DEST\" for field testStruct.Tag>,\n Received <%+v>", err)
	}
}

func TestModelHelpersCSVLoadErrorBool(t *testing.T) {
	type testStruct struct {
		ID        int64
		Tpid      string
		Tag       bool `index:"0" re:".*"`
		CreatedAt time.Time
	}

	_, err := csvLoad(testStruct{}, []string{"TEST_DEST"})
	if err == nil || err.Error() != "invalid value \"TEST_DEST\" for field testStruct.Tag" {
		t.Errorf("\nExpecting <invalid value \"TEST_DEST\" for field testStruct.Tag>,\n Received <%+v>", err)
	}
}

func TestAccountMdlsCSVHeader(t *testing.T) {
	testStruct := AccountMdls{{
		Tpid:         "TEST_TPID",
		Tenant:       "cgrates.org",
		ID:           "ResGroup1",
		FilterIDs:    "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z;FLTR_RES_GR1",
		Weights:      "10.0",
		ThresholdIDs: "WARN_RES1;WARN_RES1",
	},
	}
	exp := []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs,
		utils.Weights, utils.Blockers, utils.Opts, utils.BalanceID, utils.BalanceFilterIDs, utils.BalanceWeights, utils.BalanceBlockers, utils.BalanceType, utils.BalanceUnits, utils.BalanceUnitFactors, utils.BalanceOpts, utils.BalanceCostIncrements, utils.BalanceAttributeIDs, utils.BalanceRateProfileIDs,
		utils.ThresholdIDs}
	result := testStruct.CSVHeader()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("Expecting: %+v,\nreceived: %+v", utils.ToJSON(exp), utils.ToJSON(result))
	}
}

func TestAccountMdlsAsTPAccount(t *testing.T) {
	testStruct := AccountMdls{{
		PK:                    0,
		Tpid:                  "TEST_TPID",
		Tenant:                "cgrates.org",
		ID:                    "ResGroup1",
		FilterIDs:             "*ai:~*req.AnswerTime:2014-07-24T15:00:00Z|2014-07-25T15:00:00Z;FLTR_RES_GR1",
		Weights:               ";10",
		Blockers:              "*string:~*req.Destination:1003;false",
		BalanceID:             "VoiceBalance",
		BalanceFilterIDs:      "FLTR_RES_GR2",
		BalanceWeights:        ";10",
		BalanceBlockers:       "*string:~*req.Destination:10203;false",
		BalanceRateProfileIDs: "rt1;rt2",
		BalanceType:           utils.MetaVoice,
		BalanceUnits:          "1h",
		ThresholdIDs:          "WARN_RES1",
	},
	}
	exp := []*utils.TPAccount{
		{
			TPid:      "TEST_TPID",
			Tenant:    "cgrates.org",
			ID:        "ResGroup1",
			FilterIDs: []string{"*ai:~*req.AnswerTime:2014-07-24T15:00:00Z|2014-07-25T15:00:00Z", "FLTR_RES_GR1"},
			Weights:   ";10",
			Blockers:  "*string:~*req.Destination:1003;false",
			Balances: map[string]*utils.TPAccountBalance{
				"VoiceBalance": {
					ID:             "VoiceBalance",
					FilterIDs:      []string{"FLTR_RES_GR2"},
					Weights:        ";10",
					Blockers:       "*string:~*req.Destination:10203;false",
					Type:           utils.MetaVoice,
					RateProfileIDs: []string{"rt1", "rt2"},
					Units:          "1h",
				},
			},
			ThresholdIDs: []string{"WARN_RES1"},
		},
	}
	result, err := testStruct.AsTPAccount()
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(result[0].FilterIDs)
	sort.Strings(exp[0].Balances["VoiceBalance"].RateProfileIDs)
	sort.Strings(result[0].Balances["VoiceBalance"].RateProfileIDs)
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("Expecting: %+v,\nreceived: %+v", utils.ToJSON(exp), utils.ToJSON(result))
	}
}

func TestAccountMdlsAsTPAccountCase2(t *testing.T) {
	testStruct := AccountMdls{{
		PK:               0,
		Tpid:             "TEST_TPID",
		Tenant:           "cgrates.org",
		ID:               "ResGroup1",
		FilterIDs:        "*ai:~*req.AnswerTime:2014-07-24T15:00:00Z;FLTR_RES_GR1",
		Weights:          ";10",
		BalanceID:        "VoiceBalance",
		BalanceFilterIDs: "FLTR_RES_GR2",
		BalanceWeights:   ";10",
		BalanceBlockers:  ";false",
		BalanceType:      utils.MetaVoice,
		BalanceUnits:     "1h",
		ThresholdIDs:     "WARN_RES1",
	},
	}
	exp := []*utils.TPAccount{
		{
			TPid:      "TEST_TPID",
			Tenant:    "cgrates.org",
			ID:        "ResGroup1",
			FilterIDs: []string{"*ai:~*req.AnswerTime:2014-07-24T15:00:00Z", "FLTR_RES_GR1"},
			Weights:   ";10",
			Balances: map[string]*utils.TPAccountBalance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"FLTR_RES_GR2"},
					Weights:   ";10",
					Blockers:  ";false",
					Type:      utils.MetaVoice,
					Units:     "1h",
				},
			},
			ThresholdIDs: []string{"WARN_RES1"},
		},
	}
	result, err := testStruct.AsTPAccount()
	sort.Strings(result[0].FilterIDs)
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(result[0].FilterIDs)
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("Expecting: %+v,\nreceived: %+v", utils.ToJSON(exp), utils.ToJSON(result))
	}
}

func TestAccountMdlsAsTPAccountError(t *testing.T) {
	testStruct := AccountMdls{
		{
			PK:                    0,
			Tpid:                  "TEST_TPID",
			Tenant:                "cgrates.org",
			ID:                    "ResGroup1",
			BalanceID:             "VoiceBalance",
			BalanceCostIncrements: "AN;INVALID;COST;INCREMENT;VALUE",
		},
	}
	expectedErr := "invalid key: <AN;INVALID;COST;INCREMENT;VALUE> for BalanceCostIncrements"
	if _, err := testStruct.AsTPAccount(); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

	testStruct[0].BalanceCostIncrements = ";20;not_float;10"
	expectedErr = "strconv.ParseFloat: parsing \"not_float\": invalid syntax"
	if _, err := testStruct.AsTPAccount(); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

	testStruct[0].BalanceCostIncrements = utils.EmptyString
	testStruct[0].BalanceUnitFactors = "NOT;A;VALUE"
	expectedErr = "invalid key: <NOT;A;VALUE> for BalanceUnitFactors"
	if _, err := testStruct.AsTPAccount(); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

	testStruct[0].BalanceUnitFactors = ";float"
	expectedErr = "strconv.ParseFloat: parsing \"float\": invalid syntax"
	if _, err := testStruct.AsTPAccount(); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestAPItoModelTPAccount(t *testing.T) {
	testStruct := &utils.TPAccount{
		TPid:      "TEST_TPID",
		Tenant:    "cgrates.org",
		ID:        "ResGroup1",
		FilterIDs: []string{"*ai:~*req.AnswerTime:2014-07-24T15:00:00Z|2014-07-25T15:00:00Z", "FLTR_RES_GR1"},
		Weights:   ";10",
		Balances: map[string]*utils.TPAccountBalance{
			"VoiceBalance": {
				ID:            "VoiceBalance",
				FilterIDs:     []string{"FLTR_RES_GR2"},
				Weights:       ";10",
				Type:          utils.MetaVoice,
				Units:         "1h",
				CostIncrement: []*utils.TPBalanceCostIncrement{},
			},
		},
		ThresholdIDs: []string{"WARN_RES1"},
	}
	exp := AccountMdls{{
		Tpid:             "TEST_TPID",
		Tenant:           "cgrates.org",
		ID:               "ResGroup1",
		FilterIDs:        "*ai:~*req.AnswerTime:2014-07-24T15:00:00Z|2014-07-25T15:00:00Z;FLTR_RES_GR1",
		Weights:          ";10",
		BalanceID:        "VoiceBalance",
		BalanceFilterIDs: "FLTR_RES_GR2",
		BalanceWeights:   ";10",
		BalanceType:      utils.MetaVoice,
		BalanceUnits:     "1h",
		ThresholdIDs:     "WARN_RES1",
	}}
	result := APItoModelTPAccount(testStruct)
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("Expecting: %+v,\nreceived: %+v", utils.ToJSON(exp), utils.ToJSON(result))
	}
}

func TestAPItoModelTPAccountNoBalance(t *testing.T) {
	testStruct := &utils.TPAccount{
		TPid:         "TEST_TPID",
		Tenant:       "cgrates.org",
		ID:           "ResGroup1",
		FilterIDs:    []string{"FLTR_RES_GR1", "*ai:~*req.AnswerTime:2014-07-24T15:00:00Z|2014-07-25T15:00:00Z"},
		Weights:      "10.0",
		ThresholdIDs: []string{"WARN_RES1"},
	}
	var exp AccountMdls = nil
	result := APItoModelTPAccount(testStruct)
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("Expecting: %+v,\nreceived: %+v", utils.ToJSON(exp), utils.ToJSON(result))
	}
}

func TestAPItoModelTPAccountCase2(t *testing.T) {
	testStruct := &utils.TPAccount{
		TPid:      "TEST_TPID",
		Tenant:    "cgrates.org",
		ID:        "ResGroup1",
		FilterIDs: []string{"FLTR_RES_GR1", "FLTR_RES_GR2", "*ai:~*req.AnswerTime:2014-07-24T15:00:00Z|2014-07-25T15:00:00Z"},
		Weights:   ";10",
		Balances: map[string]*utils.TPAccountBalance{
			"VoiceBalance": {
				ID:        "VoiceBalance",
				FilterIDs: []string{"FLTR_RES_GR1", "FLTR_RES_GR2"},
				Weights:   ";10",
				Type:      utils.MetaVoice,
				Units:     "1h",
				CostIncrement: []*utils.TPBalanceCostIncrement{
					{
						FilterIDs:    []string{"*string:*~req.Account:100"},
						Increment:    "1",
						FixedFee:     utils.Float64Pointer(20),
						RecurrentFee: utils.Float64Pointer(5),
					},
					{
						FilterIDs:    []string{"*string:*~req.Destination:10"},
						Increment:    "2",
						FixedFee:     utils.Float64Pointer(10),
						RecurrentFee: utils.Float64Pointer(7),
					},
				},
				AttributeIDs:   []string{"20", "30"},
				RateProfileIDs: []string{"rt1", "rt2"},
				UnitFactors: []*utils.TPBalanceUnitFactor{
					{
						FilterIDs: []string{"*string:*~req.Account:100"},
						Factor:    21,
					},
					{
						FilterIDs: []string{"*string:*~req.Destination:10"},
						Factor:    27,
					},
				},
			},
		},
		ThresholdIDs: []string{"WARN_RES1", "WARN_RES2"},
	}
	exp := AccountMdls{{
		Tpid:                  "TEST_TPID",
		Tenant:                "cgrates.org",
		ID:                    "ResGroup1",
		FilterIDs:             "*ai:~*req.AnswerTime:2014-07-24T15:00:00Z|2014-07-25T15:00:00Z;FLTR_RES_GR1;FLTR_RES_GR2",
		Weights:               ";10",
		BalanceID:             "VoiceBalance",
		BalanceFilterIDs:      "FLTR_RES_GR1;FLTR_RES_GR2",
		BalanceWeights:        ";10",
		BalanceType:           utils.MetaVoice,
		BalanceCostIncrements: "*string:*~req.Account:100;1;20;5;*string:*~req.Destination:10;2;10;7",
		BalanceAttributeIDs:   "20;30",
		BalanceUnitFactors:    "*string:*~req.Account:100;21;*string:*~req.Destination:10;27",
		BalanceRateProfileIDs: "rt1;rt2",
		BalanceUnits:          "1h",
		ThresholdIDs:          "WARN_RES1;WARN_RES2",
	}}
	sort.Strings(testStruct.FilterIDs)
	sort.Strings(testStruct.ThresholdIDs)
	sort.Strings(testStruct.Balances["VoiceBalance"].FilterIDs)
	result := APItoModelTPAccount(testStruct)
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("Expecting: %+v,\nreceived: %+v", utils.ToJSON(exp), utils.ToJSON(result))
	}
}

func TestApitoAccountCase2(t *testing.T) {
	testStruct := &utils.TPAccount{
		Tenant:    "cgrates.org",
		ID:        "ResGroup1",
		FilterIDs: []string{"FLTR_RES_GR1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-15T14:25:00Z"},
		Weights:   ";10",
		Balances: map[string]*utils.TPAccountBalance{
			"VoiceBalance": {
				ID:             "VoiceBalance",
				FilterIDs:      []string{"FLTR_RES_GR2"},
				Weights:        ";10",
				Blockers:       "*string:~*req.Destination:122;true;;false",
				Type:           utils.MetaVoice,
				RateProfileIDs: []string{"RTPRF1"},
				Units:          "1h",
				Opts:           "key1:val1",
			},
		},
		ThresholdIDs: []string{"WARN_RES1"},
		Blockers:     ";true",
	}
	exp := &utils.Account{
		Tenant:    "cgrates.org",
		ID:        "ResGroup1",
		FilterIDs: []string{"FLTR_RES_GR1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-15T14:25:00Z"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10.0,
			},
		},
		Balances: map[string]*utils.Balance{
			"VoiceBalance": {
				ID:        "VoiceBalance",
				FilterIDs: []string{"FLTR_RES_GR2"},
				Weights: utils.DynamicWeights{
					{
						Weight: 10.0,
					},
				},
				Blockers: utils.DynamicBlockers{
					{
						FilterIDs: []string{"*string:~*req.Destination:122"},
						Blocker:   true,
					},
					{
						Blocker: false,
					},
				},
				Type:           utils.MetaVoice,
				Units:          utils.NewDecimal(3600000000000, 0),
				RateProfileIDs: []string{"RTPRF1"},
				Opts: map[string]any{
					"key1": "val1",
				},
			}},
		ThresholdIDs: []string{"WARN_RES1"},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: true,
			},
		},
	}
	result, err := APItoAccount(testStruct, "")
	if err != nil {
		t.Errorf("Expecting: <nil>,\nreceived: <%+v>", err)
	}
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("Expecting: %+v,\nreceived: %+v", utils.ToJSON(exp), utils.ToJSON(result))
	}
}

func TestApiToAccountWeightsError(t *testing.T) {
	testStruct := &utils.TPAccount{
		Tenant:  "cgrates.org",
		Weights: "10",
		Balances: map[string]*utils.TPAccountBalance{
			"VoiceBalance": {
				Weights: ";10",
				Type:    utils.MetaVoice,
			},
		},
	}
	expectedErr := "invalid DynamicWeight format for string <10>"
	if _, err := APItoAccount(testStruct, ""); err == nil || err.Error() != expectedErr {
		t.Errorf("Expecting: %+v,\nreceived: <%+v>", expectedErr, err)
	}

	testStruct.Weights = ";10"
	testStruct.Balances["VoiceBalance"].Weights = "10"
	if _, err := APItoAccount(testStruct, ""); err == nil || err.Error() != expectedErr {
		t.Errorf("Expecting: %+v,\nreceived: <%+v>", expectedErr, err)
	}
}

func TestApitoAccountCaseTimeError2(t *testing.T) {
	testStruct := &utils.TPAccount{
		Tenant:    "cgrates.org",
		ID:        "ResGroup1",
		FilterIDs: []string{"FLTR_RES_GR1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-15T14:25:00Z"},
		Weights:   ";10.0",
		Balances: map[string]*utils.TPAccountBalance{
			"VoiceBalance": {
				ID:        "VoiceBalance",
				FilterIDs: []string{"FLTR_RES_GR2"},
				Weights:   ";10",
				Type:      utils.MetaVoice,
				Units:     "1h",
				Opts:      "22:22:4fs",
			},
		},
		ThresholdIDs: []string{"WARN_RES1"},
	}
	_, err := APItoAccount(testStruct, "")
	if err == nil || err.Error() != "malformed option for ActionProfile <cgrates.org:ResGroup1> for action <VoiceBalance>" {
		t.Errorf("Expecting: <malformed option for ActionProfile <cgrates.org:ResGroup1> for action <VoiceBalance>>,\nreceived: <%+v>", err)
	}
}

func TestModelHelpersActionProfileToAPICase2(t *testing.T) {
	testStruct := &utils.TPActionProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-15T14:25:00Z"},
		Weights:   ";1",
		Schedule:  "test_schedule",
		Targets: []*utils.TPActionTarget{
			{
				TargetType: utils.MetaAccounts,
				TargetIDs:  []string{"test_account_id1", "test_account_id2"},
			},
		},
		Actions: []*utils.TPAPAction{
			{
				ID:        "test_action_id",
				FilterIDs: []string{"test_action_filter_id1"},
				Diktats: []*utils.TPAPDiktat{{
					Path: "test_path",
				}},
				Opts: "key1:val1",
			},
		},
	}

	expStruct := &utils.TPActionProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-15T14:25:00Z"},
		Weights:   ";1",
		Schedule:  "test_schedule",
		Targets: []*utils.TPActionTarget{
			{
				TargetType: utils.MetaAccounts,
				TargetIDs:  []string{"test_account_id1", "test_account_id2"},
			},
		},
		Actions: []*utils.TPAPAction{
			{
				ID:        "test_action_id",
				FilterIDs: []string{"test_action_filter_id1"},
				Diktats: []*utils.TPAPDiktat{{
					Path: "test_path",
				}},
				Opts: "key1:val1",
				TTL:  "0s",
			},
		},
	}

	result, err := APItoActionProfile(testStruct, "")
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	result2 := ActionProfileToAPI(result)
	sort.Strings(result2.FilterIDs)
	sort.Strings(expStruct.FilterIDs)
	sort.Strings(result2.Targets[0].TargetIDs)
	sort.Strings(expStruct.Targets[0].TargetIDs)
	if !reflect.DeepEqual(result2, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result2))
	}
}

func TestModelHelpersAccountToAPI(t *testing.T) {
	testStruct := &utils.Account{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"test_filterId", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-15T14:25:00Z"},
		Weights: utils.DynamicWeights{
			{
				Weight: 2,
			},
		},
		Balances: map[string]*utils.Balance{
			"VoiceBalance": {
				ID:        "VoiceBalance",
				FilterIDs: []string{"FLTR_RES_GR2"},
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Type:  utils.MetaVoice,
				Units: utils.NewDecimal(int64(time.Hour), 0),
				CostIncrements: []*utils.CostIncrement{
					{
						FilterIDs:    []string{"*string:*~req.Account:100"},
						Increment:    utils.NewDecimal(1, 0),
						FixedFee:     utils.NewDecimal(20, 0),
						RecurrentFee: utils.NewDecimal(5, 0),
					},
				},
				AttributeIDs:   []string{"20"},
				RateProfileIDs: []string{"rtprf1"},
				UnitFactors: []*utils.UnitFactor{
					{
						FilterIDs: []string{"*string:*~req.Account:100"},
						Factor:    utils.NewDecimal(21, 0),
					},
				},
				Opts: map[string]any{
					"key1": "val1",
				},
			}},
		ThresholdIDs: []string{"test_thrs"},
	}
	expStruct := &utils.TPAccount{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"test_filterId", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-15T14:25:00Z"},
		Weights:   ";2",
		Balances: map[string]*utils.TPAccountBalance{
			"VoiceBalance": {
				ID:        "VoiceBalance",
				FilterIDs: []string{"FLTR_RES_GR2"},
				Weights:   ";10",
				Type:      utils.MetaVoice,
				Units:     "3600000000000",
				Opts:      "key1:val1",
				CostIncrement: []*utils.TPBalanceCostIncrement{
					{
						FilterIDs:    []string{"*string:*~req.Account:100"},
						Increment:    "1",
						FixedFee:     utils.Float64Pointer(20),
						RecurrentFee: utils.Float64Pointer(5),
					},
				},
				AttributeIDs:   []string{"20"},
				RateProfileIDs: []string{"rtprf1"},
				UnitFactors: []*utils.TPBalanceUnitFactor{
					{
						FilterIDs: []string{"*string:*~req.Account:100"},
						Factor:    21,
					},
				},
			},
		},
		ThresholdIDs: []string{"test_thrs"},
	}
	if result := AccountToAPI(testStruct); !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestAPItoModelResourceNoFilterIDs(t *testing.T) {
	tpRL := &utils.TPResourceProfile{
		Tenant:            "cgrates.org",
		TPid:              "tp_test",
		ID:                "ResGroup1",
		FilterIDs:         []string{},
		Weights:           ";10",
		Limit:             "2",
		ThresholdIDs:      []string{"TRes1", "TRes2"},
		AllocationMessage: "test",
	}
	expModel := &ResourceMdl{
		Tpid:              "tp_test",
		Tenant:            "cgrates.org",
		ID:                "ResGroup1",
		FilterIDs:         "",
		Weights:           ";10",
		Limit:             "2",
		ThresholdIDs:      "TRes1;TRes2",
		AllocationMessage: "test",
	}
	rcv := APItoModelResource(tpRL)
	if len(rcv) != 1 {
		t.Errorf("Expecting: 1, received: %+v", len(rcv))
	} else if !reflect.DeepEqual(rcv[0], expModel) {
		t.Errorf("Expecting: \n%+v, \nreceived: \n%+v", utils.ToJSON(expModel), utils.ToJSON(rcv[0]))
	}
}

func TestAPItoResourceNewDynamicWeightsFromStringErr(t *testing.T) {
	tpRL := &utils.TPResourceProfile{
		Tenant:            "cgrates.org",
		TPid:              "tp_test",
		ID:                "ResGroup1",
		FilterIDs:         []string{"FLTR_RES_GR_1", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
		Stored:            false,
		Blocker:           false,
		Weights:           ";10",
		Limit:             "2",
		ThresholdIDs:      []string{"TRes1"},
		AllocationMessage: "asd",
	}

	expErr := "invalid Weight <not_a_float64> in string: <fltr1&fltr2;not_a_float64>"
	tpRL.Weights = "fltr1&fltr2;not_a_float64"
	if _, err := APItoResource(tpRL, "UTC"); err == nil || err.Error() != expErr {
		t.Errorf("expecting: %+v, received: %+v", expErr, err)
	}
}

func TestAPItoTPStatsNewDynamicWeightsFromStringErr(t *testing.T) {
	tps := &utils.TPStatProfile{
		TPid:        "tp_test",
		ID:          "Stats1",
		FilterIDs:   []string{"FLTR_1", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
		QueueLength: 100,
		TTL:         "1s",
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: "*sum#BalanceValue",
			},
			{
				MetricID: "*average#BalanceValue",
			},
			{
				MetricID: "*tcc",
			},
		},
		MinItems:     1,
		ThresholdIDs: []string{"THRESH1", "THRESH2"},
		Stored:       false,
		Blockers:     ";false",
		Weights:      "fltr1&fltr2;not_a_float64",
	}

	expErr := "invalid Weight <not_a_float64> in string: <fltr1&fltr2;not_a_float64>"
	if _, err := APItoStats(tps, "UTC"); err == nil || err.Error() != expErr {
		t.Errorf("expecting: %+v, received: %+v", expErr, err)
	}
}

func TestAPItoAccountNewDynamicBlockersFromStringErr(t *testing.T) {
	testStruct := &utils.TPAccount{
		Tenant:    "cgrates.org",
		ID:        "ResGroup1",
		FilterIDs: []string{"FLTR_RES_GR1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-15T14:25:00Z"},
		Weights:   ";10",
		Balances: map[string]*utils.TPAccountBalance{
			"VoiceBalance": {
				ID:             "VoiceBalance",
				FilterIDs:      []string{"FLTR_RES_GR2"},
				Weights:        ";10",
				Blockers:       "*string:~*req.Destination:122;true;;false",
				Type:           utils.MetaVoice,
				RateProfileIDs: []string{"RTPRF1"},
				Units:          "1h",
				Opts:           "key1:val1",
			},
		},
		ThresholdIDs: []string{"WARN_RES1"},
		Blockers:     "wrong input",
	}

	expErr := "invalid DynamicBlocker format for string <wrong input>"
	_, err := APItoAccount(testStruct, "")
	if err == nil || err.Error() != expErr {
		t.Errorf("expecting: %+v, received: %+v", expErr, err)
	}

}

func TestAPItoAccountNewDecimalFromUsageErr(t *testing.T) {
	testStruct := &utils.TPAccount{
		Tenant:    "cgrates.org",
		ID:        "ResGroup1",
		FilterIDs: []string{"FLTR_RES_GR1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-15T14:25:00Z"},
		Weights:   ";10",
		Balances: map[string]*utils.TPAccountBalance{
			"VoiceBalance": {
				ID:             "VoiceBalance",
				FilterIDs:      []string{"FLTR_RES_GR2"},
				Weights:        ";10",
				Blockers:       "*string:~*req.Destination:122;true;;false",
				Type:           utils.MetaVoice,
				RateProfileIDs: []string{"RTPRF1"},
				Units:          "wrong input",
				Opts:           "key1:val1",
			},
		},
		ThresholdIDs: []string{"WARN_RES1"},
		Blockers:     ";true",
	}

	expErr := "can't convert <wrong input> to decimal"
	_, err := APItoAccount(testStruct, "")
	if err == nil || err.Error() != expErr {
		t.Errorf("expecting: %+v, received: %+v", expErr, err)
	}

}

func TestAPItoAccountBalancesNewDynamicBlockersFromStringErr(t *testing.T) {
	testStruct := &utils.TPAccount{
		Tenant:    "cgrates.org",
		ID:        "ResGroup1",
		FilterIDs: []string{"FLTR_RES_GR1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-15T14:25:00Z"},
		Weights:   ";10",
		Balances: map[string]*utils.TPAccountBalance{
			"VoiceBalance": {
				ID:             "VoiceBalance",
				FilterIDs:      []string{"FLTR_RES_GR2"},
				Weights:        ";10",
				Blockers:       "wrong input",
				Type:           utils.MetaVoice,
				RateProfileIDs: []string{"RTPRF1"},
				Units:          "1h",
				Opts:           "key1:val1",
			},
		},
		ThresholdIDs: []string{"WARN_RES1"},
		Blockers:     ";true",
	}

	expErr := "invalid DynamicBlocker format for string <wrong input>"
	_, err := APItoAccount(testStruct, "")
	if err == nil || err.Error() != expErr {
		t.Errorf("expecting: %+v, received: %+v", expErr, err)
	}

}

func TestAPItoActionProfileNewDynamicWeightsFromStringErr(t *testing.T) {
	testStruct := &utils.TPActionProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-15T14:25:00Z"},
		Weights:   "wrong input",
		Schedule:  "test_schedule",
		Targets: []*utils.TPActionTarget{
			{
				TargetType: utils.MetaAccounts,
				TargetIDs:  []string{"test_account_id1", "test_account_id2"},
			},
		},
		Actions: []*utils.TPAPAction{
			{
				ID:        "test_action_id",
				FilterIDs: []string{"test_action_filter_id1"},
				Diktats: []*utils.TPAPDiktat{{
					Path: "test_path",
				}},
				Opts: "key1:val1",
			},
		},
	}

	expErr := "invalid DynamicWeight format for string <wrong input>"
	_, err := APItoActionProfile(testStruct, "")
	if err == nil || err.Error() != expErr {
		t.Errorf("expecting: %+v, received: %+v", expErr, err)
	}

}

func TestAPItoActionProfileNewDynamicBlockersFromStringErr(t *testing.T) {
	testStruct := &utils.TPActionProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-15T14:25:00Z"},
		Weights:   ";1",
		Schedule:  "test_schedule",
		Targets: []*utils.TPActionTarget{
			{
				TargetType: utils.MetaAccounts,
				TargetIDs:  []string{"test_account_id1", "test_account_id2"},
			},
		},
		Actions: []*utils.TPAPAction{
			{
				ID:        "test_action_id",
				FilterIDs: []string{"test_action_filter_id1"},
				Diktats: []*utils.TPAPDiktat{{
					Path: "test_path",
				}},
				Opts: "key1:val1",
			},
		},
		Blockers: "wrong input",
	}

	expErr := "invalid DynamicBlocker format for string <wrong input>"
	_, err := APItoActionProfile(testStruct, "")
	if err == nil || err.Error() != expErr {
		t.Errorf("expecting: %+v, received: %+v", expErr, err)
	}

}

func TestAPItoModelTPActionProfileActionProfileMdl(t *testing.T) {
	testStruct := &utils.TPActionProfile{
		TPid:      "test_id",
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001", "*string:~*req.Subject:1002", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z|2014-08-29T15:00:00Z"},
		Weights:   ";1",
		Schedule:  "test_schedule",
		Targets: []*utils.TPActionTarget{{
			TargetType: utils.MetaAccounts,
			TargetIDs:  []string{"test_account_id1", "test_account_id2"},
		}},
		Actions: []*utils.TPAPAction{
			{
				ID:        "test_action_id",
				FilterIDs: []string{"test_action_filter_id1", "test_action_filter_id2"},
				Diktats: []*utils.TPAPDiktat{
					{
						Path:  "*balance.AbstractBalance1.Units",
						Value: "10",
					},
					{
						Path:  "*balance.AbstractBalance1.Units",
						Value: "5",
					}},
			},
		},
	}

	expStruct := ActionProfileMdls{{
		Tpid:        "test_id",
		Tenant:      "cgrates.org",
		ID:          "RP1",
		ActionID:    "test_action_id",
		ActionPath:  "*balance.AbstractBalance1.Units",
		ActionValue: "5",
	}}
	result := APItoModelTPActionProfile(testStruct)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting %s,\n Received %s", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestAsTPDispatcherHostsNilTransport(t *testing.T) {

	tps := &DispatcherHostMdls{
		&DispatcherHostMdl{
			ID:                   "ID1",
			Tenant:               "Tenant1",
			Address:              "localhost:6012",
			ConnectAttempts:      2,
			Reconnects:           5,
			MaxReconnectInterval: "5m",
			ConnectTimeout:       "2m",
			ReplyTimeout:         "1m",
			TLS:                  true,
			ClientKey:            "client_key",
			ClientCertificate:    "client_certificate",
			CaCertificate:        "ca_certificate",
		}}

	eOut := []*utils.TPDispatcherHost{
		{
			Tenant: "Tenant1",
			ID:     "ID1",
			Conn: &utils.TPDispatcherHostConn{
				Address:              "localhost:6012",
				Transport:            "*json",
				ConnectAttempts:      2,
				Reconnects:           5,
				MaxReconnectInterval: 5 * time.Minute,
				ConnectTimeout:       2 * time.Minute,
				ReplyTimeout:         1 * time.Minute,
				TLS:                  true,
				ClientKey:            "client_key",
				ClientCertificate:    "client_certificate",
				CaCertificate:        "ca_certificate",
			},
		},
	}
	if rcv, err := tps.AsTPDispatcherHosts(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

}

func TestAsTPDispatcherHostsMaxReconnectIntervalErr(t *testing.T) {

	tps := &DispatcherHostMdls{
		&DispatcherHostMdl{
			ID:                   "ID1",
			Tenant:               "Tenant1",
			Address:              "localhost:6012",
			Transport:            utils.MetaJSON,
			ConnectAttempts:      2,
			Reconnects:           5,
			MaxReconnectInterval: "wrong input",
			ConnectTimeout:       "2m",
			ReplyTimeout:         "1m",
			TLS:                  true,
			ClientKey:            "client_key",
			ClientCertificate:    "client_certificate",
			CaCertificate:        "ca_certificate",
		}}

	expErr := `time: invalid duration "wrong input"`
	if _, err := tps.AsTPDispatcherHosts(); err == nil || err.Error() != expErr {
		t.Errorf("expecting: \n%+v\n, received: \n%+v", expErr, err)
	}

}

func TestAsTPDispatcherHostsConnectTimeoutErr(t *testing.T) {

	tps := &DispatcherHostMdls{
		&DispatcherHostMdl{
			ID:                   "ID1",
			Tenant:               "Tenant1",
			Address:              "localhost:6012",
			Transport:            utils.MetaJSON,
			ConnectAttempts:      2,
			Reconnects:           5,
			MaxReconnectInterval: "5m",
			ConnectTimeout:       "wrong input",
			ReplyTimeout:         "1m",
			TLS:                  true,
			ClientKey:            "client_key",
			ClientCertificate:    "client_certificate",
			CaCertificate:        "ca_certificate",
		}}

	expErr := `time: invalid duration "wrong input"`
	if _, err := tps.AsTPDispatcherHosts(); err == nil || err.Error() != expErr {
		t.Errorf("expecting: \n%+v\n, received: \n%+v", expErr, err)
	}

}

func TestAsTPDispatcherHostsReplyTimeoutErr(t *testing.T) {

	tps := &DispatcherHostMdls{
		&DispatcherHostMdl{
			ID:                   "ID1",
			Tenant:               "Tenant1",
			Address:              "localhost:6012",
			Transport:            utils.MetaJSON,
			ConnectAttempts:      2,
			Reconnects:           5,
			MaxReconnectInterval: "5m",
			ConnectTimeout:       "2m",
			ReplyTimeout:         "wrong input",
			TLS:                  true,
			ClientKey:            "client_key",
			ClientCertificate:    "client_certificate",
			CaCertificate:        "ca_certificate",
		}}

	expErr := `time: invalid duration "wrong input"`
	if _, err := tps.AsTPDispatcherHosts(); err == nil || err.Error() != expErr {
		t.Errorf("expecting: \n%+v\n, received: \n%+v", expErr, err)
	}

}

func TestAPItoDispatcherProfileNoParams(t *testing.T) {
	tpDPP := &utils.TPDispatcherProfile{
		TPid:           "TP1",
		Tenant:         "cgrates.org",
		ID:             "Dsp",
		FilterIDs:      []string{"*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "FLTR_ACNT_dan", "FLTR_DST_DE"},
		Strategy:       utils.MetaFirst,
		StrategyParams: []any{},
		Weight:         20,
		Hosts: []*utils.TPDispatcherHostProfile{
			{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    []any{""},
				Blocker:   false,
			},
		},
	}

	expected := &DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "Dsp",
		FilterIDs:      []string{"*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "FLTR_ACNT_dan", "FLTR_DST_DE"},
		Strategy:       utils.MetaFirst,
		StrategyParams: map[string]any{},
		Weight:         20,
		Hosts: DispatcherHostProfiles{
			&DispatcherHostProfile{
				ID:        "C1",
				FilterIDs: []string{},
				Params:    make(map[string]any),
				Weight:    10,
				Blocker:   false,
			},
		},
	}
	if rcv := APItoDispatcherProfile(tpDPP, "UTC"); !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : \n%+v\n, received: \n%+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestAPItoChargerProfileNewDynamicWeightsFromStringErr(t *testing.T) {
	tpCPP := &utils.TPChargerProfile{
		TPid:         "TP1",
		Tenant:       "cgrates.org",
		ID:           "Charger1",
		FilterIDs:    []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z"},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR1", "ATTR2"},
		Weights:      "wrong input",
	}

	expected := &ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "Charger1",
		FilterIDs:    []string{"", "", ""},
		Weights:      nil,
		Blockers:     nil,
		RunID:        "*rated",
		AttributeIDs: []string{"", ""},
	}
	if rcv := APItoChargerProfile(tpCPP, "UTC"); !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : \n%+v\n, received: \n%+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestAPItoChargerProfileNewDynamicBlockersFromStringErr(t *testing.T) {
	tpCPP := &utils.TPChargerProfile{
		TPid:         "TP1",
		Tenant:       "cgrates.org",
		ID:           "Charger1",
		FilterIDs:    []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z"},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR1", "ATTR2"},
		Weights:      ";10",
		Blockers:     "wrong input",
	}

	expected := &ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		FilterIDs: []string{"", "", ""},
		Weights: utils.DynamicWeights{
			{
				Weight: float64(10),
			},
		},
		Blockers:     nil,
		RunID:        "*rated",
		AttributeIDs: []string{"", ""},
	}
	if rcv := APItoChargerProfile(tpCPP, "UTC"); !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : \n%+v\n, received: \n%+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

// Number of FilterIDs is 0
func TestAPItoModelTPCharger7(t *testing.T) {
	tpCharger := &utils.TPChargerProfile{
		TPid:         "TP1",
		Tenant:       "cgrates.org",
		ID:           "Charger1",
		FilterIDs:    []string{},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR1", "ATTR2"},
		Weights:      ";20",
	}
	expected := ChargerMdls{
		&ChargerMdl{
			Tpid:         "TP1",
			Tenant:       "cgrates.org",
			ID:           "Charger1",
			FilterIDs:    "",
			RunID:        "*rated",
			AttributeIDs: "ATTR1",
			Weights:      ";20",
		},
		&ChargerMdl{
			Tpid:         "TP1",
			Tenant:       "cgrates.org",
			ID:           "Charger1",
			FilterIDs:    "",
			AttributeIDs: "ATTR2",
		},
	}
	rcv := APItoModelTPCharger(tpCharger)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestModelAsTPChargersWithBlockers(t *testing.T) {
	models := ChargerMdls{
		&ChargerMdl{
			Tpid:         "TP1",
			Tenant:       "cgrates.org",
			ID:           "Charger1",
			FilterIDs:    "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z;FLTR_ACNT_dan;FLTR_DST_DE",
			RunID:        "*rated",
			AttributeIDs: "*constant:*req.RequestType:*rated;*constant:*req.Category:call;ATTR1;*constant:*req.Category:call",
			Weights:      ";20",
			Blockers:     ";true",
		},
	}
	expected := &utils.TPChargerProfile{
		TPid:         "TP1",
		Tenant:       "cgrates.org",
		ID:           "Charger1",
		FilterIDs:    []string{"*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "FLTR_ACNT_dan", "FLTR_DST_DE"},
		RunID:        "*rated",
		AttributeIDs: []string{"*constant:*req.RequestType:*rated;*constant:*req.Category:call", "ATTR1", "*constant:*req.Category:call"},
		Weights:      ";20",
		Blockers:     ";true",
	}
	rcv := models.AsTPChargers()
	sort.Strings(rcv[0].FilterIDs)
	if !reflect.DeepEqual(expected, rcv[0]) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv[0]))
	}
}

func TestAPItoAttributeProfileNewDynamicBlockersFromStringErr(t *testing.T) {
	tpAlsPrf := &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "*string:~*opts.*context:con1"},
		Attributes: []*utils.TPAttribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FL1",
				Value: "Al1",
			},
		},
		Weights:  ";20",
		Blockers: "wrong input",
	}

	expErr := "invalid DynamicBlocker format for string <wrong input>"
	if _, err := APItoAttributeProfile(tpAlsPrf, "UTC"); err == nil || err.Error() != expErr {
		t.Errorf("expecting: %+v, received: %+v", expErr, err)
	}
}

func TestAPItoAttributeProfileNewDynamicWeightsFromStringErr(t *testing.T) {
	tpAlsPrf := &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "*string:~*opts.*context:con1"},
		Attributes: []*utils.TPAttribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FL1",
				Value: "Al1",
			},
		},
		Weights: "wrong input",
	}

	expErr := "invalid DynamicWeight format for string <wrong input>"
	if _, err := APItoAttributeProfile(tpAlsPrf, "UTC"); err == nil || err.Error() != expErr {
		t.Errorf("expecting: %+v, received: %+v", expErr, err)
	}
}

func TestAPItoAttributeProfileAttrNewDynamicBlockersFromStringErr(t *testing.T) {
	tpAlsPrf := &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "*string:~*opts.*context:con1"},
		Attributes: []*utils.TPAttribute{
			{
				Path:     utils.MetaReq + utils.NestingSep + "FL1",
				Value:    "Al1",
				Blockers: "wrong input",
			},
		},
		Weights: ";20",
	}

	expErr := "invalid DynamicBlocker format for string <wrong input>"
	if _, err := APItoAttributeProfile(tpAlsPrf, "UTC"); err == nil || err.Error() != expErr {
		t.Errorf("expecting: %+v, received: %+v", expErr, err)
	}
}

func TestAPItoRouteProfileNewDynamicWeightsFromStringErr(t *testing.T) {
	testStruct := &utils.TPRouteProfile{
		FilterIDs:         []string{},
		SortingParameters: []string{"param1"},
		Routes:            []*utils.TPRoute{},
		Weights:           "wrong input",
	}

	expErr := "invalid DynamicWeight format for string <wrong input>"
	_, err := APItoRouteProfile(testStruct, "")
	if err == nil || err.Error() != expErr {
		t.Errorf("expecting: %+v, received: %+v", expErr, err)
	}
}

func TestAPItoRouteProfileNewDynamicBlockersFromStringErr(t *testing.T) {
	testStruct := &utils.TPRouteProfile{
		FilterIDs:         []string{},
		SortingParameters: []string{"param1"},
		Routes:            []*utils.TPRoute{},
		Weights:           ";10",
		Blockers:          "wrong input",
	}

	expErr := "invalid DynamicBlocker format for string <wrong input>"
	_, err := APItoRouteProfile(testStruct, "")
	if err == nil || err.Error() != expErr {
		t.Errorf("expecting: %+v, received: %+v", expErr, err)
	}
}

func TestAPItoRouteProfileRouteNewDynamicWeightsFromStringErr(t *testing.T) {
	testStruct := &utils.TPRouteProfile{
		FilterIDs:         []string{},
		SortingParameters: []string{"param1"},
		Routes: []*utils.TPRoute{
			{
				ID:      "r1",
				Weights: "wrong input",
			},
		},
		Weights: ";10",
	}

	expErr := "invalid DynamicWeight format for string <wrong input>"
	_, err := APItoRouteProfile(testStruct, "")
	if err == nil || err.Error() != expErr {
		t.Errorf("expecting: %+v, received: %+v", expErr, err)
	}
}

func TestAPItoRouteProfileRouteNewDynamicBlockersFromStringErr(t *testing.T) {
	testStruct := &utils.TPRouteProfile{
		FilterIDs:         []string{},
		SortingParameters: []string{"param1"},
		Routes: []*utils.TPRoute{
			{
				ID:       "r1",
				Weights:  ";10",
				Blockers: "wrong input",
			},
		},
		Weights:  ";10",
		Blockers: ";true",
	}

	expErr := "invalid DynamicBlocker format for string <wrong input>"
	_, err := APItoRouteProfile(testStruct, "")
	if err == nil || err.Error() != expErr {
		t.Errorf("expecting: %+v, received: %+v", expErr, err)
	}
}

func TestAPItoTPThresholdNewDynamicWeightsFromStringErr(t *testing.T) {
	tps := &utils.TPThresholdProfile{
		TPid:             "tp_test",
		ID:               "TH1",
		FilterIDs:        []string{"FilterID1", "FilterID2", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
		MaxHits:          12,
		MinHits:          10,
		MinSleep:         "1s",
		Blocker:          false,
		Weights:          "wrong input",
		ActionProfileIDs: []string{"WARN3"},
	}

	expErr := "invalid DynamicWeight format for string <wrong input>"
	_, err := APItoThresholdProfile(tps, "UTC")
	if err == nil || err.Error() != expErr {
		t.Errorf("expecting: \n%+v\n, received: \n%+v", expErr, err)
	}
}

func TestAPItoModelTPThresholdNoFilters(t *testing.T) {
	testStruct := &utils.TPThresholdProfile{
		TPid:             "TP1",
		Tenant:           "cgrates.org",
		ID:               "TH_1",
		FilterIDs:        []string{},
		MaxHits:          12,
		MinHits:          10,
		MinSleep:         "1s",
		Blocker:          false,
		Weights:          ";20",
		ActionProfileIDs: []string{"WARN3", "LOG"},
	}
	expStruct := ThresholdMdls{
		{
			Tpid:             "TP1",
			Tenant:           "cgrates.org",
			ID:               "TH_1",
			FilterIDs:        "",
			MaxHits:          12,
			MinHits:          10,
			MinSleep:         "1s",
			Blocker:          false,
			Weights:          ";20",
			ActionProfileIDs: "WARN3",
		},
		{
			Tpid:             "TP1",
			Tenant:           "cgrates.org",
			ID:               "TH_1",
			ActionProfileIDs: "LOG",
		},
	}

	result := APItoModelTPThreshold(testStruct)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestAPItoTPStatsNewDynamicBlockersFromStringErr(t *testing.T) {
	tps := &utils.TPStatProfile{
		TPid:        "tp_test",
		ID:          "Stats1",
		FilterIDs:   []string{"FLTR_1", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
		QueueLength: 100,
		TTL:         "1s",
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: "*sum#BalanceValue",
			},
			{
				MetricID: "*average#BalanceValue",
			},
			{
				MetricID: "*tcc",
			},
		},
		MinItems:     1,
		ThresholdIDs: []string{"THRESH1", "THRESH2"},
		Stored:       false,
		Blockers:     "wrong input",
		Weights:      ";20.0",
	}

	expErr := "invalid DynamicBlocker format for string <wrong input>"
	_, err := APItoStats(tps, "UTC")
	if err == nil || err.Error() != expErr {
		t.Errorf("expecting: \n%+v\n, received: \n%+v", expErr, err)
	}
}

func TestAPItoTPStatsMetricNewDynamicBlockersFromStringErr(t *testing.T) {
	tps := &utils.TPStatProfile{
		TPid:        "tp_test",
		ID:          "Stats1",
		FilterIDs:   []string{"FLTR_1", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
		QueueLength: 100,
		TTL:         "1s",
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: "*sum#BalanceValue",
				Blockers: "wrong input",
			},
			{
				MetricID: "*average#BalanceValue",
			},
			{
				MetricID: "*tcc",
			},
		},
		MinItems:     1,
		ThresholdIDs: []string{"THRESH1", "THRESH2"},
		Stored:       false,
		Blockers:     ";false",
		Weights:      ";20.0",
	}

	expErr := "invalid DynamicBlocker format for string <wrong input>"
	_, err := APItoStats(tps, "UTC")
	if err == nil || err.Error() != expErr {
		t.Errorf("expecting: \n%+v\n, received: \n%+v", expErr, err)
	}
}

func TestCsvDumpForThresholdModels(t *testing.T) {
	tpThPrf := &utils.TPThresholdProfile{
		TPid:             "TP1",
		Tenant:           "cgrates.org",
		ID:               "TH_1",
		FilterIDs:        []string{"FilterID1", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z"},
		MaxHits:          12,
		MinHits:          10,
		MinSleep:         "1s",
		Blocker:          false,
		Weights:          ";20",
		ActionProfileIDs: []string{"WARN3"},
	}
	expected := ThresholdMdls{
		{
			Tpid:             "TP1",
			Tenant:           "cgrates.org",
			ID:               "TH_1",
			FilterIDs:        "FilterID1",
			MaxHits:          12,
			MinHits:          10,
			MinSleep:         "1s",
			Blocker:          false,
			Weights:          ";20",
			ActionProfileIDs: "WARN3",
		},
		{
			Tpid:      "TP1",
			ID:        "TH_1",
			Tenant:    "cgrates.org",
			FilterIDs: "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z",
		},
	}
	rcv := APItoModelTPThreshold(tpThPrf)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v,\n received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
	expRecord := []string{"cgrates.org", "TH_1", "FilterID1", ";20", "12", "10", "1s", "false", "WARN3", "false"}
	for i, model := range rcv {
		if i == 1 {
			expRecord = []string{"cgrates.org", "TH_1", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "", "0", "0", "", "false", "", "false"}
		}
		if csvRecordRcv, _ := CsvDump(model); !reflect.DeepEqual(expRecord, csvRecordRcv) {
			t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expRecord), utils.ToJSON(csvRecordRcv))
		}
	}

}

func TestCsvDumpForDispatcherModels(t *testing.T) {
	tpDispPrf := &utils.TPDispatcherProfile{
		TPid:           "TP1",
		Tenant:         "cgrates.org",
		ID:             "Dsp",
		FilterIDs:      []string{"*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "FLTR_ACNT_dan", "FLTR_DST_DE"},
		Strategy:       utils.MetaFirst,
		StrategyParams: []any{},
		Weight:         20,
		Hosts: []*utils.TPDispatcherHostProfile{
			{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    []any{"192.168.54.203"},
				Blocker:   false,
			},
			{
				ID:        "C2",
				FilterIDs: []string{},
				Weight:    10,
				Params:    []any{"192.168.54.204"},
				Blocker:   false,
			},
		},
	}
	expected := DispatcherProfileMdls{
		&DispatcherProfileMdl{
			Tpid:           "TP1",
			Tenant:         "cgrates.org",
			ID:             "Dsp",
			FilterIDs:      "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z;FLTR_ACNT_dan;FLTR_DST_DE",
			Strategy:       utils.MetaFirst,
			Weight:         20,
			ConnID:         "C1",
			ConnWeight:     10,
			ConnBlocker:    false,
			ConnParameters: "192.168.54.203",
		},
		&DispatcherProfileMdl{
			Tpid:           "TP1",
			Tenant:         "cgrates.org",
			ID:             "Dsp",
			ConnID:         "C2",
			ConnWeight:     10,
			ConnBlocker:    false,
			ConnParameters: "192.168.54.204",
		},
	}
	rcv := APItoModelTPDispatcherProfile(tpDispPrf)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v,\n received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
	expRecord := []string{"cgrates.org", "Dsp", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z;FLTR_ACNT_dan;FLTR_DST_DE", "20", "*first", "", "C1", "", "10", "false", "192.168.54.203"}
	for i, model := range rcv {
		if i == 1 {
			expRecord = []string{"cgrates.org", "Dsp", "", "0", "", "", "C2", "", "10", "false", "192.168.54.204"}
		}
		if csvRecordRcv, _ := CsvDump(model); !reflect.DeepEqual(expRecord, csvRecordRcv) {
			t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expRecord), utils.ToJSON(csvRecordRcv))
		}
	}

}
