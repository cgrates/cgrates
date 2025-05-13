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
package chargers

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/attributes"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/rpcclient"

	"github.com/cgrates/cgrates/utils"
)

func TestChargerSetChargerProfiles(t *testing.T) {
	var dmCharger *engine.DataManager
	cPPs := []*utils.ChargerProfile{
		{
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
		{
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
		{
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
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dmCharger = engine.NewDataManager(data, cfg, nil)

	fltrCP1 := &engine.Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_1",
		Rules: []*engine.FilterRule{
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
	fltrCP2 := &engine.Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_2",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Charger",
				Values:  []string{"ChargerProfile2"},
			},
		},
	}
	dmCharger.SetFilter(context.Background(), fltrCP2, true)
	fltrCPPrefix := &engine.Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_3",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.harger",
				Values:  []string{"Charger"},
			},
		},
	}
	dmCharger.SetFilter(context.Background(), fltrCPPrefix, true)
	fltrCP4 := &engine.Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_4",
		Rules: []*engine.FilterRule{
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
	var dmCharger *engine.DataManager
	cPPs := []*utils.ChargerProfile{
		{
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
		{
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
		{
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

	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dmCharger = engine.NewDataManager(data, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dmCharger)
	chargerSrv = NewChargerService(dmCharger, fltrs, cfg, nil)

	fltrCP1 := &engine.Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_1",
		Rules: []*engine.FilterRule{
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
	fltrCP2 := &engine.Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_2",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Charger",
				Values:  []string{"ChargerProfile2"},
			},
		},
	}
	dmCharger.SetFilter(context.Background(), fltrCP2, true)
	fltrCPPrefix := &engine.Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_3",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.harger",
				Values:  []string{"Charger"},
			},
		},
	}
	dmCharger.SetFilter(context.Background(), fltrCPPrefix, true)
	fltrCP4 := &engine.Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_4",
		Rules: []*engine.FilterRule{
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
	var dmCharger *engine.DataManager
	cPPs := []*utils.ChargerProfile{
		{
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
		{
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
		{
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

	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dmCharger = engine.NewDataManager(data, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dmCharger)
	chargerSrv = NewChargerService(dmCharger, fltrs, cfg, nil)

	fltrCP1 := &engine.Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_1",
		Rules: []*engine.FilterRule{
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
	fltrCP2 := &engine.Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_2",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Charger",
				Values:  []string{"ChargerProfile2"},
			},
		},
	}
	dmCharger.SetFilter(context.Background(), fltrCP2, true)
	fltrCPPrefix := &engine.Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_3",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.harger",
				Values:  []string{"Charger"},
			},
		},
	}
	dmCharger.SetFilter(context.Background(), fltrCPPrefix, true)
	fltrCP4 := &engine.Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_4",
		Rules: []*engine.FilterRule{
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
			t.Errorf("Expecting: %#v, received: %#v", cp, tempCp)
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
			AlteredFields: []*attributes.FieldsAltered{
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

	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dmCharger := engine.NewDataManager(dataDB, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dmCharger)
	cS := &ChargerS{
		dm:    dmCharger,
		fltrS: fltrs,
		cfg:   cfg,
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

	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dmCharger := engine.NewDataManager(dataDB, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dmCharger)
	cS := &ChargerS{
		dm:    dmCharger,
		fltrS: fltrs,
		cfg:   cfg,
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

	dbm := &engine.DataDBMock{
		GetKeysForPrefixF: func(ctx *context.Context, s string) ([]string, error) {
			return []string{":"}, nil
		},
	}
	dmCharger := engine.NewDataManager(dbm, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dmCharger)
	cS := &ChargerS{
		dm:    dmCharger,
		fltrS: fltrs,
		cfg:   cfg,
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
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	cS := &ChargerS{
		dm:    dm,
		fltrS: fltrs,
		cfg:   cfg,
	}
	cpp := &utils.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "CPP_1",
		RunID:        "TestRunID",
		AttributeIDs: []string{"*none"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
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

	if err := engine.Cache.Set(context.Background(), utils.CacheChargerProfiles, "cgrates.org:CPP_1", nil, []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	_, err := cS.matchingChargerProfilesForEvent(context.Background(), cgrEv.Tenant, cgrEv)

	if err != utils.ErrNotFound {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

}

func TestChargersmatchingChargerProfilesForEventWeightFromDynamicsErr(t *testing.T) {

	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	cS := &ChargerS{
		dm:    dm,
		fltrS: fltrs,
		cfg:   cfg,
	}

	cpp := &utils.ChargerProfile{
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
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	cS := &ChargerS{
		dm:    dm,
		fltrS: fltrs,
		cfg:   cfg,
	}

	cpp := &utils.ChargerProfile{
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
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	cS := &ChargerS{
		dm:    dm,
		fltrS: fltrs,
		cfg:   cfg,
	}

	cpp := &utils.ChargerProfile{
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

	exp := []*utils.ChargerProfile{
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

func TestChargersmatchingChargerProfilesForEventErrPass(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.ChargerSCfg().IndexedSelects = false

	dbm := &engine.DataDBMock{
		GetChargerProfileDrvF: func(ctx *context.Context, s1, s2 string) (*utils.ChargerProfile, error) {
			return &utils.ChargerProfile{
				Tenant:    s1,
				ID:        s2,
				RunID:     utils.MetaDefault,
				FilterIDs: []string{"fltr1"},
			}, nil
		},
		GetKeysForPrefixF: func(ctx *context.Context, s string) ([]string, error) {
			return []string{s + "cgrates.org:chr1"}, nil
		},
		GetFilterDrvF: func(ctx *context.Context, s1, s2 string) (*engine.Filter, error) {
			return nil, utils.ErrNotImplemented
		},
	}
	dmFilter := engine.NewDataManager(dbm, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dmFilter)
	cS := &ChargerS{
		dm:    dmFilter,
		fltrS: fltrs,
		cfg:   cfg,
	}
	cgrEv := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "cgrEvID",
		Event: map[string]any{
			"Charger":        "ChargerProfile1",
			utils.AnswerTime: time.Date(2021, 4, 19, 12, 0, 0, 0, time.UTC),
			"UsageInterval":  "10s",
			utils.Weight:     "10.0",
		},
		APIOpts: map[string]any{
			utils.MetaSubsys: utils.MetaChargers,
		},
	}

	experr := utils.ErrNotImplemented
	rcv, err := cS.matchingChargerProfilesForEvent(context.Background(), cgrEv.Tenant, cgrEv)

	if err == nil || err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
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

func TestChargersprocessEventCallNilErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.ChargerSCfg().IndexedSelects = false
	cfg.ChargerSCfg().AttributeSConns = []string{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}

	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	cP := &utils.ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "1001",
		RunID:     utils.MetaDefault,
		FilterIDs: []string{"*string:~*req.Account:1001"},
	}
	if err := dm.SetChargerProfile(context.Background(), cP, true); err != nil {
		t.Fatal(err)
	}

	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.AttributeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				rply := attributes.AttrSProcessEventReply{
					AlteredFields: []*attributes.FieldsAltered{{
						MatchedProfileID: "attr1",
						Fields:           []string{utils.MetaReq + utils.NestingSep + utils.AccountField},
					}},
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     "cgrEvID",
						Event: map[string]any{
							utils.AccountField: "1002",
						},
					},
				}
				*reply.(*attributes.AttrSProcessEventReply) = rply
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	fltrs := engine.NewFilterS(cfg, nil, dm)
	cS := &ChargerS{
		dm:      dm,
		fltrS:   fltrs,
		cfg:     cfg,
		connMgr: engine.NewConnManager(cfg),
	}
	cS.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), utils.AttributeSv1, rpcInternal)
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "cgrEvID",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}

	exp := []*ChrgSProcessEventReply{
		{
			ChargerSProfile: "1001",
			AlteredFields: []*attributes.FieldsAltered{
				{
					MatchedProfileID: utils.MetaDefault,
					Fields:           []string{utils.MetaOptsRunID, utils.MetaOpts + utils.NestingSep + utils.MetaChargeID, utils.MetaOpts + utils.NestingSep + utils.MetaSubsys},
				},
				{
					MatchedProfileID: "attr1",
					Fields:           []string{utils.MetaReq + utils.NestingSep + utils.AccountField},
				},
			},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "cgrEvID",
				Event: map[string]any{
					utils.AccountField: "1002",
				},
			},
		},
	}
	rcv, err := cS.processEvent(context.Background(), cgrEv.Tenant, cgrEv)
	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(rcv))
	}

	if err := dm.DataDB().Flush(""); err != nil {
		t.Error(err)
	}

}

func TestChargersprocessEventCallErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.ChargerSCfg().IndexedSelects = false
	cfg.ChargerSCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}

	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	cP := &utils.ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "1001",
		RunID:     utils.MetaDefault,
		FilterIDs: []string{"*string:~*req.Account:1001"},
	}
	if err := dm.SetChargerProfile(context.Background(), cP, true); err != nil {
		t.Fatal(err)
	}

	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.AttributeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	fltrs := engine.NewFilterS(cfg, nil, dm)
	cS := &ChargerS{
		dm:      dm,
		fltrS:   fltrs,
		cfg:     cfg,
		connMgr: engine.NewConnManager(cfg),
	}
	cS.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), utils.AttributeSv1, rpcInternal)
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "cgrEvID",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}

	exp := []*ChrgSProcessEventReply{
		{
			ChargerSProfile: "1001",
			AlteredFields: []*attributes.FieldsAltered{
				{
					MatchedProfileID: utils.MetaDefault,
					Fields:           []string{utils.MetaOptsRunID, utils.MetaOpts + utils.NestingSep + utils.MetaChargeID, utils.MetaOpts + utils.NestingSep + utils.MetaSubsys},
				},
			},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "cgrEvID",
				Event: map[string]any{
					utils.AccountField: "1001",
				},
				APIOpts: map[string]any{
					utils.OptsAttributesProfileIDs: []string{},
					utils.OptsContext:              "*chargers",
					utils.MetaRunID:                "*default",
					utils.MetaSubsys:               "*chargers",
				},
			},
		},
	}
	rcv, err := cS.processEvent(context.Background(), cgrEv.Tenant, cgrEv)
	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
	exp[0].CGREvent.APIOpts[utils.MetaChargeID] = rcv[0].CGREvent.APIOpts[utils.MetaChargeID]
	exp[0].CGREvent.APIOpts[utils.OptsAttributesProfileIDs] = rcv[0].CGREvent.APIOpts[utils.OptsAttributesProfileIDs]
	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("\nexpected: <%v>, \nreceived: <%v>",
			utils.ToJSON(exp), utils.ToJSON(rcv))
		t.Errorf("\nexpected: <%T>, \nreceived: <%T>",
			exp[0].CGREvent.APIOpts[utils.OptsAttributesProfileIDs], rcv[0].CGREvent.APIOpts[utils.OptsAttributesProfileIDs])
	}

	if err := dm.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
}

func TestChargersV1ProcessEventErrNotFound(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cfg.ChargerSCfg().IndexedSelects = false
	cfg.ChargerSCfg().AttributeSConns = []string{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	dm := engine.NewDataManager(dataDB, cfg, nil)

	cP := &utils.ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "1001",
		RunID:     utils.MetaDefault,
		FilterIDs: []string{"*string:~*req.Account:1001"},
	}
	if err := dm.SetChargerProfile(context.Background(), cP, true); err != nil {
		t.Fatal(err)
	}

	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.AttributeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				rply := attributes.AttrSProcessEventReply{
					AlteredFields: []*attributes.FieldsAltered{{
						Fields: []string{utils.AccountField},
					}},
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     "cgrEvID",
						Event: map[string]any{
							utils.AccountField: "1001",
						},
					},
				}
				*reply.(*attributes.AttrSProcessEventReply) = rply
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	fltrs := engine.NewFilterS(cfg, nil, dm)
	cS := &ChargerS{
		dm:      dm,
		fltrS:   fltrs,
		cfg:     cfg,
		connMgr: engine.NewConnManager(cfg),
	}
	cS.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), utils.AttributeSv1, rpcInternal)
	args := &utils.CGREvent{
		ID: "cgrEvID",
		Event: map[string]any{
			utils.AccountField: "1002",
		},
	}
	reply := &[]*ChrgSProcessEventReply{}

	experr := utils.ErrNotFound
	err := cS.V1ProcessEvent(context.Background(), args, reply)

	if err == nil || err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if err := dm.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
}

func TestChargersV1ProcessEventErrOther(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cfg.ChargerSCfg().IndexedSelects = false
	cfg.ChargerSCfg().AttributeSConns = []string{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	dm := engine.NewDataManager(dataDB, cfg, nil)

	cP := &utils.ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "1001",
		RunID:     utils.MetaDefault,
		FilterIDs: []string{"*string:~*req.Account:1001"},
	}
	if err := dm.SetChargerProfile(context.Background(), cP, true); err != nil {
		t.Fatal(err)
	}

	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			"invalidMethod": func(ctx *context.Context, args, reply any) error {
				rply := attributes.AttrSProcessEventReply{
					AlteredFields: []*attributes.FieldsAltered{{
						Fields: []string{utils.AccountField},
					}},
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     "cgrEvID",
						Event: map[string]any{
							utils.AccountField: "1001",
						},
					},
				}
				*reply.(*attributes.AttrSProcessEventReply) = rply
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM

	fltrs := engine.NewFilterS(cfg, nil, dm)
	cS := &ChargerS{
		dm:      dm,
		fltrS:   fltrs,
		cfg:     cfg,
		connMgr: engine.NewConnManager(cfg),
	}
	cS.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), utils.AttributeSv1, rpcInternal)
	args := &utils.CGREvent{
		ID: "cgrEvID",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}
	reply := &[]*ChrgSProcessEventReply{}

	exp := &[]*ChrgSProcessEventReply{}
	experr := fmt.Sprintf("SERVER_ERROR: %s", rpcclient.ErrUnsupporteServiceMethod)
	err := cS.V1ProcessEvent(context.Background(), args, reply)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if !reflect.DeepEqual(reply, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(reply))
	}

	if err := dm.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
}

func TestChargersV1ProcessEvent(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cfg.ChargerSCfg().IndexedSelects = false
	cfg.ChargerSCfg().AttributeSConns = []string{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	dm := engine.NewDataManager(dataDB, cfg, nil)

	cP := &utils.ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "1001",
		RunID:     utils.MetaDefault,
		FilterIDs: []string{"*string:~*req.Account:1001"},
	}
	if err := dm.SetChargerProfile(context.Background(), cP, true); err != nil {
		t.Fatal(err)
	}

	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.AttributeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				rply := attributes.AttrSProcessEventReply{
					AlteredFields: []*attributes.FieldsAltered{{
						MatchedProfileID: "attr2",
						Fields:           []string{utils.MetaReq + utils.NestingSep + utils.AccountField},
					}},
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     "cgrEvID",
						Event: map[string]any{
							utils.AccountField: "1007",
						},
						APIOpts: map[string]any{
							utils.OptsAttributesProfileIDs: []string{},
							utils.OptsContext:              "*chargers",
							utils.MetaRunID:                "*default",
							utils.MetaSubsys:               "*chargers",
						},
					},
				}
				*reply.(*attributes.AttrSProcessEventReply) = rply
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM

	fltrs := engine.NewFilterS(cfg, nil, dm)
	cS := &ChargerS{
		dm:      dm,
		fltrS:   fltrs,
		cfg:     cfg,
		connMgr: engine.NewConnManager(cfg),
	}
	cS.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), utils.AttributeSv1, rpcInternal)
	args := &utils.CGREvent{
		ID: "cgrEvID",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}
	reply := []*ChrgSProcessEventReply{}

	exp := []*ChrgSProcessEventReply{
		{
			ChargerSProfile: "1001",
			AlteredFields: []*attributes.FieldsAltered{
				{
					MatchedProfileID: utils.MetaDefault,
					Fields:           []string{utils.MetaOptsRunID, utils.MetaOpts + utils.NestingSep + utils.MetaChargeID, utils.MetaOpts + utils.NestingSep + utils.MetaSubsys},
				},
				{
					MatchedProfileID: "attr2",
					Fields:           []string{utils.MetaReq + utils.NestingSep + utils.AccountField},
				},
			},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "cgrEvID",
				Event: map[string]any{
					utils.AccountField: "1007",
				},
				APIOpts: map[string]any{
					utils.OptsAttributesProfileIDs: []string{},
					utils.OptsContext:              "*chargers",
					utils.MetaRunID:                "*default",
					utils.MetaSubsys:               "*chargers",
				},
			},
		},
	}
	//exp[0].CGREvent.APIOpts[utils.MetaChargeID] = reply[0].CGREvent.APIOpts[utils.MetaChargeID]
	err := cS.V1ProcessEvent(context.Background(), args, &reply)

	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if !reflect.DeepEqual(reply, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(reply))
	}

	if err := dm.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
}

func TestChargersV1GetChargersForEventNilErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	cfg.ChargerSCfg().IndexedSelects = false
	cfg.ChargerSCfg().AttributeSConns = []string{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	dm := engine.NewDataManager(dataDB, cfg, nil)

	cP := &utils.ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "1001",
		RunID:     utils.MetaDefault,
		FilterIDs: []string{"*string:~*req.Account:1001"},
	}
	if err := dm.SetChargerProfile(context.Background(), cP, true); err != nil {
		t.Fatal(err)
	}

	fltrs := engine.NewFilterS(cfg, nil, dm)
	cS := &ChargerS{
		dm:    dm,
		fltrS: fltrs,
		cfg:   cfg,
	}
	args := &utils.CGREvent{
		ID: "cgrEvID",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}
	var reply []*utils.ChargerProfile

	exp := []*utils.ChargerProfile{
		{
			Tenant:    "cgrates.org",
			ID:        "1001",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			RunID:     "*default",
		},
	}
	err := cS.V1GetChargersForEvent(context.Background(), args, &reply)

	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if !reflect.DeepEqual(reply, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(reply))
	}

	if err := dm.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
}

func TestChargersV1GetChargersForEventErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.ChargerSCfg().IndexedSelects = false

	dbm := &engine.DataDBMock{
		GetKeysForPrefixF: func(ctx *context.Context, s string) ([]string, error) {
			return []string{":"}, nil
		},
	}
	dm := engine.NewDataManager(dbm, cfg, nil)

	fltrs := engine.NewFilterS(cfg, nil, dm)
	cS := &ChargerS{
		dm:    dm,
		fltrS: fltrs,
		cfg:   cfg,
	}
	args := &utils.CGREvent{
		ID: "cgrEvID",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}
	var reply []*utils.ChargerProfile
	var exp []*utils.ChargerProfile
	experr := fmt.Sprintf("SERVER_ERROR: %s", utils.ErrNotImplemented)
	err := cS.V1GetChargersForEvent(context.Background(), args, &reply)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if !reflect.DeepEqual(reply, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(reply))
	}
}
