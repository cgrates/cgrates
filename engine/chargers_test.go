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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"

	"github.com/cgrates/cgrates/utils"
)

func TestChargerSetChargerProfiles(t *testing.T) {
	var dmCharger *DataManager
	cPPs := ChargerProfiles{
		&ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "CPP_1",
			FilterIDs:    []string{"FLTR_CP_1", "FLTR_CP_4", "*string:~*opts.*subsys:*chargers", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			RunID:        "TestRunID",
			AttributeIDs: []string{"*none"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			weight: 20,
		},
		&ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "CPP_2",
			FilterIDs:    []string{"FLTR_CP_2", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			RunID:        "*rated",
			AttributeIDs: []string{"ATTR_1"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			weight: 20,
		},
		&ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "CPP_3",
			FilterIDs:    []string{"FLTR_CP_3", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			RunID:        "*rated",
			AttributeIDs: []string{"ATTR_1"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			weight: 20,
		},
	}
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmCharger = NewDataManager(data, cfg, nil)

	fltrCP1 := &Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Charger",
				Values:  []string{"ChargerProfile1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.UsageInterval",
				Values:  []string{time.Second.String()},
			},
		},
	}
	if err := fltrCP1.Compile(); err != nil {
		t.Error(err)
	}
	dmCharger.SetFilter(context.Background(), fltrCP1, true)
	fltrCP2 := &Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Charger",
				Values:  []string{"ChargerProfile2"},
			},
		},
	}
	dmCharger.SetFilter(context.Background(), fltrCP2, true)
	fltrCPPrefix := &Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.harger",
				Values:  []string{"Charger"},
			},
		},
	}
	dmCharger.SetFilter(context.Background(), fltrCPPrefix, true)
	fltrCP4 := &Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_4",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"200.00"},
			},
		},
	}
	dmCharger.SetFilter(context.Background(), fltrCP4, true)
	for _, cp := range cPPs {
		if err := dmCharger.SetChargerProfile(context.Background(), cp, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//verify each charger from cache
	for _, cp := range cPPs {
		if tempCp, err := dmCharger.GetChargerProfile(context.Background(), cp.Tenant, cp.ID,
			true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(cp, tempCp) {
			t.Errorf("Expecting: %+v, received: %+v", cp, tempCp)
		}
	}
}

func TestChargerMatchingChargerProfilesForEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	var chargerSrv *ChargerS
	var dmCharger *DataManager
	cPPs := ChargerProfiles{
		&ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "CPP_1",
			FilterIDs:    []string{"FLTR_CP_1", "FLTR_CP_4", "*string:~*opts.*subsys:*chargers", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			RunID:        "TestRunID",
			AttributeIDs: []string{"*none"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
		&ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "CPP_2",
			FilterIDs:    []string{"FLTR_CP_2", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			RunID:        "*rated",
			AttributeIDs: []string{"ATTR_1"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
		&ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "CPP_3",
			FilterIDs:    []string{"FLTR_CP_3", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			RunID:        "*rated",
			AttributeIDs: []string{"ATTR_1"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
	}
	chargerEvents := []*utils.CGREvent{
		{
			Tenant: cfg.GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Charger":        "ChargerProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				utils.Weight:     "200.0",
			},
			APIOpts: map[string]any{
				utils.MetaSubsys: utils.MetaChargers,
			},
		},
		{
			Tenant: cfg.GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Charger":        "ChargerProfile2",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			},
		},
		{
			Tenant: cfg.GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Charger":        "DistinctMatch",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			},
		},
	}

	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmCharger = NewDataManager(data, cfg, nil)
	chargerSrv = NewChargerService(dmCharger,
		&FilterS{dm: dmCharger, cfg: cfg}, cfg, nil)

	fltrCP1 := &Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Charger",
				Values:  []string{"ChargerProfile1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.UsageInterval",
				Values:  []string{time.Second.String()},
			},
		},
	}
	if err := fltrCP1.Compile(); err != nil {
		t.Error(err)
	}
	dmCharger.SetFilter(context.Background(), fltrCP1, true)
	fltrCP2 := &Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Charger",
				Values:  []string{"ChargerProfile2"},
			},
		},
	}
	dmCharger.SetFilter(context.Background(), fltrCP2, true)
	fltrCPPrefix := &Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.harger",
				Values:  []string{"Charger"},
			},
		},
	}
	dmCharger.SetFilter(context.Background(), fltrCPPrefix, true)
	fltrCP4 := &Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_4",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"200.00"},
			},
		},
	}
	dmCharger.SetFilter(context.Background(), fltrCP4, true)

	for _, cp := range cPPs {
		if err := dmCharger.SetChargerProfile(context.Background(), cp, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//verify each charger from cache
	for _, cp := range cPPs {
		if tempCp, err := dmCharger.GetChargerProfile(context.Background(), cp.Tenant, cp.ID,
			true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(cp, tempCp) {
			t.Errorf("Expecting: %+v, received: %+v", cp, tempCp)
		}
	}

	if _, err := chargerSrv.matchingChargerProfilesForEvent(context.Background(), chargerEvents[2].Tenant, chargerEvents[2]); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Error: %+v", err)
	}

	if rcv, err := chargerSrv.matchingChargerProfilesForEvent(context.Background(), chargerEvents[0].Tenant, chargerEvents[0]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(cPPs[0], rcv[0]) {
		t.Errorf("Expecting: %+v, received: %+v ", cPPs[0], rcv[0])
	}

	if rcv, err := chargerSrv.matchingChargerProfilesForEvent(context.Background(), chargerEvents[1].Tenant, chargerEvents[1]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(cPPs[1], rcv[0]) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(cPPs[1]), utils.ToJSON(rcv))
	}

}

func TestChargerProcessEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	var chargerSrv *ChargerS
	var dmCharger *DataManager
	cPPs := ChargerProfiles{
		&ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "CPP_1",
			FilterIDs:    []string{"FLTR_CP_1", "FLTR_CP_4", "*string:~*opts.*subsys:*chargers", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			RunID:        "TestRunID",
			AttributeIDs: []string{"*none"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			weight: 20,
		},
		&ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "CPP_2",
			FilterIDs:    []string{"FLTR_CP_2", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			RunID:        "*rated",
			AttributeIDs: []string{"ATTR_1"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			weight: 20,
		},
		&ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "CPP_3",
			FilterIDs:    []string{"FLTR_CP_3", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			RunID:        "*rated",
			AttributeIDs: []string{"ATTR_1"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			weight: 20,
		},
	}
	chargerEvents := []*utils.CGREvent{
		{
			Tenant: cfg.GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Charger":        "ChargerProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				utils.Weight:     "200.0",
			},
			APIOpts: map[string]any{
				utils.MetaSubsys: utils.MetaChargers,
			},
		},
		{
			Tenant: cfg.GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Charger":        "ChargerProfile2",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			},
		},
		{
			Tenant: cfg.GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Charger":        "DistinctMatch",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			},
		},
	}

	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmCharger = NewDataManager(data, cfg, nil)
	chargerSrv = NewChargerService(dmCharger,
		&FilterS{dm: dmCharger, cfg: cfg}, cfg, nil)

	fltrCP1 := &Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Charger",
				Values:  []string{"ChargerProfile1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.UsageInterval",
				Values:  []string{time.Second.String()},
			},
		},
	}
	if err := fltrCP1.Compile(); err != nil {
		t.Error(err)
	}
	dmCharger.SetFilter(context.Background(), fltrCP1, true)
	fltrCP2 := &Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Charger",
				Values:  []string{"ChargerProfile2"},
			},
		},
	}
	dmCharger.SetFilter(context.Background(), fltrCP2, true)
	fltrCPPrefix := &Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.harger",
				Values:  []string{"Charger"},
			},
		},
	}
	dmCharger.SetFilter(context.Background(), fltrCPPrefix, true)
	fltrCP4 := &Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_4",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"200.00"},
			},
		},
	}
	dmCharger.SetFilter(context.Background(), fltrCP4, true)

	for _, cp := range cPPs {
		if err := dmCharger.SetChargerProfile(context.Background(), cp, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//verify each charger from cache
	for _, cp := range cPPs {
		if tempCp, err := dmCharger.GetChargerProfile(context.Background(), cp.Tenant, cp.ID,
			true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(cp, tempCp) {
			t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(cp), utils.ToJSON(tempCp))
		}
	}

	if _, err := chargerSrv.matchingChargerProfilesForEvent(context.Background(), chargerEvents[2].Tenant, chargerEvents[2]); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Error: %+v", err)
	}

	if rcv, err := chargerSrv.matchingChargerProfilesForEvent(context.Background(), chargerEvents[0].Tenant, chargerEvents[0]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(cPPs[0], rcv[0]) {
		t.Errorf("Expecting: %+v, received: %+v ", cPPs[0], rcv[0])
	}

	if rcv, err := chargerSrv.matchingChargerProfilesForEvent(context.Background(), chargerEvents[1].Tenant, chargerEvents[1]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(cPPs[1], rcv[0]) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(cPPs[1]), utils.ToJSON(rcv))
	}
	rpl := []*ChrgSProcessEventReply{
		{
			ChargerSProfile: "CPP_1",
			AlteredFields: []*FieldsAltered{
				{
					MatchedProfileID: utils.MetaDefault,
					Fields:           []string{utils.MetaOptsRunID, utils.MetaOpts + utils.NestingSep + utils.MetaChargeID, utils.MetaOpts + utils.NestingSep + utils.MetaSubsys},
				},
			},
			CGREvent: chargerEvents[0],
		},
	}
	rpl[0].CGREvent.APIOpts[utils.MetaRunID] = cPPs[0].RunID
	rcv, err := chargerSrv.processEvent(context.Background(), rpl[0].CGREvent.Tenant, chargerEvents[0])
	if err != nil {
		t.Fatalf("Error: %+v", err)
	}
	rpl[0].CGREvent.APIOpts[utils.MetaChargeID] = rcv[0].CGREvent.APIOpts[utils.MetaChargeID]
	if !reflect.DeepEqual(rpl[0], rcv[0]) {
		t.Errorf("Expecting: %+v, received: %+v ", utils.ToJSON(rpl[0]), utils.ToJSON(rcv[0]))
	}
}

func TestChargersmatchingChargerProfilesForEventChargerProfileNotFound(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ChargerSCfg().StringIndexedFields = &[]string{
		"string",
	}
	cfg.ChargerSCfg().PrefixIndexedFields = &[]string{"prefix"}
	cfg.ChargerSCfg().SuffixIndexedFields = &[]string{"suffix"}
	cfg.ChargerSCfg().IndexedSelects = false
	cfg.ChargerSCfg().NestedFields = false

	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmCharger := NewDataManager(dataDB, cfg, nil)
	cS := &ChargerS{
		dm: dmCharger,
		fltrS: &FilterS{
			dm:  dmCharger,
			cfg: cfg,
		},
		cfg: cfg,
	}
	cgrEv := &utils.CGREvent{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "cgrEvID",
		Event: map[string]any{
			"Charger":        "ChargerProfile1",
			utils.AnswerTime: time.Date(2021, 4, 1, 10, 0, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			utils.Weight:     "10.0",
		},
		APIOpts: map[string]any{
			utils.MetaSubsys: utils.MetaChargers,
		},
	}

	experr := utils.ErrNotFound
	rcv, err := cS.matchingChargerProfilesForEvent(context.Background(), "tnt", cgrEv)

	if err == nil || err != experr {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestChargersmatchingChargerProfilesForEventDoesNotPass(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ChargerSCfg().StringIndexedFields = &[]string{
		"string",
	}
	cfg.ChargerSCfg().PrefixIndexedFields = &[]string{"prefix"}
	cfg.ChargerSCfg().SuffixIndexedFields = &[]string{"suffix"}
	cfg.ChargerSCfg().IndexedSelects = false
	cfg.ChargerSCfg().NestedFields = false

	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmCharger := NewDataManager(dataDB, cfg, nil)
	cS := &ChargerS{
		dm: dmCharger,
		fltrS: &FilterS{
			dm:  dmCharger,
			cfg: cfg,
		},
		cfg: cfg,
	}
	cgrEv := &utils.CGREvent{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "cgrEvID",
		Event: map[string]any{
			"Charger":        "ChargerProfile1",
			utils.AnswerTime: time.Date(2021, 4, 1, 10, 0, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			utils.Weight:     "10.0",
		},
		APIOpts: map[string]any{
			utils.MetaSubsys: utils.MetaChargers,
		},
	}

	experr := utils.ErrNotFound
	rcv, err := cS.matchingChargerProfilesForEvent(context.Background(), cgrEv.Tenant, cgrEv)

	if err == nil || err != experr {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestChargersmatchingChargerProfilesForEventErrGetChPrf(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ChargerSCfg().StringIndexedFields = &[]string{
		"string",
	}
	cfg.ChargerSCfg().PrefixIndexedFields = &[]string{"prefix"}
	cfg.ChargerSCfg().SuffixIndexedFields = &[]string{"suffix"}
	cfg.ChargerSCfg().IndexedSelects = false
	cfg.ChargerSCfg().NestedFields = false

	dbm := &DataDBMock{
		GetKeysForPrefixF: func(ctx *context.Context, s string) ([]string, error) {
			return []string{":"}, nil
		},
	}
	dmCharger := NewDataManager(dbm, cfg, nil)
	cS := &ChargerS{
		dm: dmCharger,
		fltrS: &FilterS{
			dm:  dmCharger,
			cfg: cfg,
		},
		cfg: cfg,
	}
	cgrEv := &utils.CGREvent{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "cgrEvID",
		Event: map[string]any{
			"Charger":        "ChargerProfile1",
			utils.AnswerTime: time.Date(2021, 4, 1, 10, 0, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			utils.Weight:     "10.0",
		},
		APIOpts: map[string]any{
			utils.MetaSubsys: utils.MetaChargers,
		},
	}

	experr := utils.ErrNotImplemented
	rcv, err := cS.matchingChargerProfilesForEvent(context.Background(), cgrEv.Tenant, cgrEv)

	if err == nil || err != experr {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}

}

func TestChargersprocessEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cS := &ChargerS{
		cfg: cfg,
	}
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 2,
		},
	}

	experr := "NO_DATABASE_CONNECTION"
	rcv, err := cS.processEvent(context.Background(), cgrEv.Tenant, cgrEv)

	if err == nil || err.Error() != experr {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestChargersV1ProcessEventMissingArgs(t *testing.T) {
	cS := &ChargerS{}
	args := &utils.CGREvent{}
	var reply *[]*ChrgSProcessEventReply

	experr := "MANDATORY_IE_MISSING: [Event]"
	err := cS.V1ProcessEvent(context.Background(), args, reply)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestChargersmatchingChargerProfilesForEventCacheReadErr(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg, nil)
	cS := &ChargerS{
		dm: dm,
		fltrS: &FilterS{
			dm:  dm,
			cfg: cfg,
		},
		cfg: cfg,
	}
	cpp := &ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "CPP_1",
		RunID:        "TestRunID",
		AttributeIDs: []string{"*none"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		weight: 20,
	}

	if err := cS.dm.SetChargerProfile(context.Background(), cpp, true); err != nil {
		t.Error(err)
	}

	cgrEv := &utils.CGREvent{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "cgrEvID",
		Event:  map[string]any{},
		APIOpts: map[string]any{
			utils.MetaSubsys: utils.MetaChargers,
		},
	}

	if err := Cache.Set(context.Background(), utils.CacheChargerProfiles, "cgrates.org:CPP_1", nil, []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	_, err := cS.matchingChargerProfilesForEvent(context.Background(), cgrEv.Tenant, cgrEv)

	if err != utils.ErrNotFound {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

}

func TestChargersmatchingChargerProfilesForEventWeightFromDynamicsErr(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg, nil)
	cS := &ChargerS{
		dm: dm,
		fltrS: &FilterS{
			dm:  dm,
			cfg: cfg,
		},
		cfg: cfg,
	}

	cpp := &ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "CPP_2",
		RunID:        "TestRunID2",
		AttributeIDs: []string{"*none"},
		Weights: utils.DynamicWeights{
			{
				FilterIDs: []string{"*stirng:~*req.Account:1001"},
				Weight:    20,
			},
		},
		weight: 20,
	}

	if err := cS.dm.SetChargerProfile(context.Background(), cpp, true); err != nil {
		t.Error(err)
	}

	cgrEv := &utils.CGREvent{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "cgrEvID",
		Event:  map[string]any{},
		APIOpts: map[string]any{
			utils.MetaSubsys: utils.MetaChargers,
		},
	}

	expErr := "NOT_IMPLEMENTED:*stirng"
	_, err := cS.matchingChargerProfilesForEvent(context.Background(), cgrEv.Tenant, cgrEv)
	if err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}

}

func TestChargersmatchingChargerProfilesForEventBlockerFromDynamicsErr(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg, nil)
	cS := &ChargerS{
		dm: dm,
		fltrS: &FilterS{
			dm:  dm,
			cfg: cfg,
		},
		cfg: cfg,
	}

	cpp := &ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "CPP_2",
		RunID:        "TestRunID2",
		AttributeIDs: []string{"*none"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				FilterIDs: []string{"*stirng:~*req.Account:1001"},
				Blocker:   false,
			},
		},
		weight: 20,
	}

	if err := cS.dm.SetChargerProfile(context.Background(), cpp, true); err != nil {
		t.Error(err)
	}

	cgrEv := &utils.CGREvent{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "cgrEvID",
		Event:  map[string]any{},
		APIOpts: map[string]any{
			utils.MetaSubsys: utils.MetaChargers,
		},
	}

	expErr := "NOT_IMPLEMENTED:*stirng"
	_, err := cS.matchingChargerProfilesForEvent(context.Background(), cgrEv.Tenant, cgrEv)
	if err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}

}

func TestChargersmatchingChargerProfilesForEventBlockerTrue(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg, nil)
	cS := &ChargerS{
		dm: dm,
		fltrS: &FilterS{
			dm:  dm,
			cfg: cfg,
		},
		cfg: cfg,
	}

	cpp := &ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "CPP_2",
		RunID:        "TestRunID2",
		AttributeIDs: []string{"*none"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: true,
			},
		},
		weight: 20,
	}

	if err := cS.dm.SetChargerProfile(context.Background(), cpp, true); err != nil {
		t.Error(err)
	}

	cgrEv := &utils.CGREvent{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "cgrEvID",
		Event:  map[string]any{},
		APIOpts: map[string]any{
			utils.MetaSubsys: utils.MetaChargers,
		},
	}

	exp := ChargerProfiles{
		{
			Tenant: "cgrates.org",
			ID:     "CPP_2",
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			Blockers: utils.DynamicBlockers{{
				Blocker: true,
			}},
			RunID:        "TestRunID2",
			AttributeIDs: []string{"*none"},
		},
	}

	rcv, err := cS.matchingChargerProfilesForEvent(context.Background(), cgrEv.Tenant, cgrEv)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(utils.ToJSON(exp), utils.ToJSON(rcv)) {
		t.Fatalf("expected: \n<%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

}
