/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or56
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
	"sort"
	"strings"
	"testing"
	"time"

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
	eRL := &utils.ResourceProfile{
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
		Metrics: []*utils.TPMetricWithFilters{
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
	eTPs := &utils.StatQueueProfile{ID: tps.ID,
		QueueLength: tps.QueueLength,
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

	eTPs := &utils.ThresholdProfile{
		ID:               tps.ID,
		MaxHits:          tps.MaxHits,
		Blocker:          tps.Blocker,
		MinHits:          tps.MinHits,
		FilterIDs:        tps.FilterIDs,
		ActionProfileIDs: []string{"WARN3"},
		EeIDs:            []string{},
		AttributeIDs:     []string{},
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
	expected := &utils.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z", "*string:~*opts.*context:con1"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FL1",
				Value: utils.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
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

	expected := &utils.ChargerProfile{
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

// Number of FilterIDs and AttributeIDs are equal

// Number of FilterIDs is smaller than AttributeIDs

// Number of FilterIDs is greater than AttributeIDs

// len(AttributeIDs) is 0

// len(FilterIDs) is 0

// both len(AttributeIDs) and len(FilterIDs) are 0

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

func TestModelHelpersParamsToString(t *testing.T) {
	testInterface := []any{"Param1", "Param2"}
	result := paramsToString(testInterface)
	if !reflect.DeepEqual(result, "Param1;Param2") {
		t.Errorf("\nExpecting <Param1;Param2>,\n Received <%+v>", result)
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
					ID:   "actDiktatID",
					Opts: "*balancePath:test_path",
				}},
				Opts: "key1:val1;key2:val2",
			},
		},
	}

	expStruct := &utils.ActionProfile{
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
		Actions: []*utils.APAction{
			{
				ID:        "test_action_id",
				FilterIDs: []string{"test_action_filter_id1", "test_action_filter_id2"},
				Diktats: []*utils.APDiktat{{
					ID: "actDiktatID",
					Opts: map[string]any{
						"*balancePath": "test_path",
					},
				}},
				Opts: map[string]any{
					"key1": "val1",
					"key2": "val2",
				},
			},
		},
	}
	result, err := APItoActionProfile(testStruct, "")
	if err != nil {
		t.Fatal(err)
	}
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
					ID:   "actDiktatID",
					Opts: "*balancePath:test_path",
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
					ID:   "actDiktatID",
					Opts: "*balancePath:test_path",
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

func TestModelHelpersTestAPItoRouteProfile(t *testing.T) {
	testStruct := &utils.TPRouteProfile{
		FilterIDs:         []string{},
		SortingParameters: []string{"param1"},
		Routes:            []*utils.TPRoute{},
	}
	expStruct := &utils.RouteProfile{
		FilterIDs:         []string{},
		SortingParameters: []string{"param1"},
		Routes:            []*utils.Route{},
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

func TestStatMdlsAsTPStatsCase2(t *testing.T) {
	testStruct := StatMdls{{
		FilterIDs:       "*ai:~*req.AnswerTime:2014-07-25T15:00:00Z|2014-07-26T15:00:00Z",
		MetricIDs:       "test_id",
		MetricFilterIDs: "test_filter_id",
	}}
	expStruct := []*utils.TPStatProfile{{
		FilterIDs: []string{"*ai:~*req.AnswerTime:2014-07-25T15:00:00Z|2014-07-26T15:00:00Z"},
		Metrics: []*utils.TPMetricWithFilters{
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
		Metrics: []*utils.TPMetricWithFilters{
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
					Opts: "*balancePath:test_path",
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
					Opts: "*balancePath:test_path",
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

	expected := &utils.ChargerProfile{
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

	expected := &utils.ChargerProfile{
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

func TestAPItoTPStatsNewDynamicBlockersFromStringErr(t *testing.T) {
	tps := &utils.TPStatProfile{
		TPid:        "tp_test",
		ID:          "Stats1",
		FilterIDs:   []string{"FLTR_1", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
		QueueLength: 100,
		TTL:         "1s",
		Metrics: []*utils.TPMetricWithFilters{
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
		Metrics: []*utils.TPMetricWithFilters{
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
