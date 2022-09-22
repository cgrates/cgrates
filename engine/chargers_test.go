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
	"bytes"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"

	"github.com/cgrates/cgrates/utils"
)

func TestChargerSetChargerProfiles(t *testing.T) {
	var dmCharger *DataManager
	cPPs := ChargerProfiles{
		&ChargerProfile{
			Tenant:    "cgrates.org",
			ID:        "CPP_1",
			FilterIDs: []string{"FLTR_CP_1", "FLTR_CP_4", "*string:~*opts.*subsys:*chargers"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			RunID:        "TestRunID",
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
		&ChargerProfile{
			Tenant:    "cgrates.org",
			ID:        "CPP_2",
			FilterIDs: []string{"FLTR_CP_2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			RunID:        "*rated",
			AttributeIDs: []string{"ATTR_1"},
			Weight:       20,
		},
		&ChargerProfile{
			Tenant:    "cgrates.org",
			ID:        "CPP_3",
			FilterIDs: []string{"FLTR_CP_3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			RunID:        "*rated",
			AttributeIDs: []string{"ATTR_1"},
			Weight:       20,
		},
	}
	data := NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)
	dmCharger = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)

	fltrCP1 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
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
				Values:  []string{(time.Second).String()},
			},
		},
	}
	if err := fltrCP1.Compile(); err != nil {
		t.Error(err)
	}
	dmCharger.SetFilter(fltrCP1, true)
	fltrCP2 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Charger",
				Values:  []string{"ChargerProfile2"},
			},
		},
	}
	dmCharger.SetFilter(fltrCP2, true)
	fltrCPPrefix := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.harger",
				Values:  []string{"Charger"},
			},
		},
	}
	dmCharger.SetFilter(fltrCPPrefix, true)
	fltrCP4 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_4",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"200.00"},
			},
		},
	}
	dmCharger.SetFilter(fltrCP4, true)
	for _, cp := range cPPs {
		if err = dmCharger.SetChargerProfile(cp, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//verify each charger from cache
	for _, cp := range cPPs {
		if tempCp, err := dmCharger.GetChargerProfile(cp.Tenant, cp.ID,
			true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(cp, tempCp) {
			t.Errorf("Expecting: %+v, received: %+v", cp, tempCp)
		}
	}
}

func TestChargerMatchingChargerProfilesForEvent(t *testing.T) {
	var chargerSrv *ChargerService
	var dmCharger *DataManager
	cPPs := ChargerProfiles{
		&ChargerProfile{
			Tenant:    "cgrates.org",
			ID:        "CPP_1",
			FilterIDs: []string{"FLTR_CP_1", "FLTR_CP_4", "*string:~*opts.*subsys:*chargers"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			RunID:        "TestRunID",
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
		&ChargerProfile{
			Tenant:    "cgrates.org",
			ID:        "CPP_2",
			FilterIDs: []string{"FLTR_CP_2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			RunID:        "*rated",
			AttributeIDs: []string{"ATTR_1"},
			Weight:       20,
		},
		&ChargerProfile{
			Tenant:    "cgrates.org",
			ID:        "CPP_3",
			FilterIDs: []string{"FLTR_CP_3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			RunID:        "*rated",
			AttributeIDs: []string{"ATTR_1"},
			Weight:       20,
		},
	}
	chargerEvents := []*utils.CGREvent{
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"Charger":        "ChargerProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				utils.Weight:     "200.0",
			},
			APIOpts: map[string]interface{}{
				utils.MetaSubsys: utils.MetaChargers,
			},
		},
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"Charger":        "ChargerProfile2",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			},
		},
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"Charger":        "DistinctMatch",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			},
		},
	}

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmCharger = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	chargerSrv = NewChargerService(dmCharger,
		&FilterS{dm: dmCharger, cfg: cfg}, cfg, nil)

	fltrCP1 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
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
				Values:  []string{(time.Second).String()},
			},
		},
	}
	if err := fltrCP1.Compile(); err != nil {
		t.Error(err)
	}
	dmCharger.SetFilter(fltrCP1, true)
	fltrCP2 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Charger",
				Values:  []string{"ChargerProfile2"},
			},
		},
	}
	dmCharger.SetFilter(fltrCP2, true)
	fltrCPPrefix := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.harger",
				Values:  []string{"Charger"},
			},
		},
	}
	dmCharger.SetFilter(fltrCPPrefix, true)
	fltrCP4 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_4",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"200.00"},
			},
		},
	}
	dmCharger.SetFilter(fltrCP4, true)

	for _, cp := range cPPs {
		if err = dmCharger.SetChargerProfile(cp, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//verify each charger from cache
	for _, cp := range cPPs {
		if tempCp, err := dmCharger.GetChargerProfile(cp.Tenant, cp.ID,
			true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(cp, tempCp) {
			t.Errorf("Expecting: %+v, received: %+v", cp, tempCp)
		}
	}

	if _, err = chargerSrv.matchingChargerProfilesForEvent(chargerEvents[2].Tenant, chargerEvents[2]); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Error: %+v", err)
	}

	if rcv, err := chargerSrv.matchingChargerProfilesForEvent(chargerEvents[0].Tenant, chargerEvents[0]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(cPPs[0], rcv[0]) {
		t.Errorf("Expecting: %+v, received: %+v ", cPPs[0], rcv[0])
	}

	if rcv, err := chargerSrv.matchingChargerProfilesForEvent(chargerEvents[1].Tenant, chargerEvents[1]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(cPPs[1], rcv[0]) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(cPPs[1]), utils.ToJSON(rcv))
	}

}

func TestChargerProcessEvent(t *testing.T) {
	var chargerSrv *ChargerService
	var dmCharger *DataManager
	cPPs := ChargerProfiles{
		&ChargerProfile{
			Tenant:    "cgrates.org",
			ID:        "CPP_1",
			FilterIDs: []string{"FLTR_CP_1", "FLTR_CP_4", "*string:~*opts.*subsys:*chargers"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			RunID:        "TestRunID",
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
		&ChargerProfile{
			Tenant:    "cgrates.org",
			ID:        "CPP_2",
			FilterIDs: []string{"FLTR_CP_2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			RunID:        "*rated",
			AttributeIDs: []string{"ATTR_1"},
			Weight:       20,
		},
		&ChargerProfile{
			Tenant:    "cgrates.org",
			ID:        "CPP_3",
			FilterIDs: []string{"FLTR_CP_3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			RunID:        "*rated",
			AttributeIDs: []string{"ATTR_1"},
			Weight:       20,
		},
	}
	chargerEvents := []*utils.CGREvent{
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"Charger":        "ChargerProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				utils.Weight:     "200.0",
			},
			APIOpts: map[string]interface{}{
				utils.MetaSubsys: utils.MetaChargers,
			},
		},
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"Charger":        "ChargerProfile2",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			},
		},
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"Charger":        "DistinctMatch",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			},
		},
	}

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmCharger = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	chargerSrv = NewChargerService(dmCharger,
		&FilterS{dm: dmCharger, cfg: cfg}, cfg, nil)

	fltrCP1 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
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
				Values:  []string{(time.Second).String()},
			},
		},
	}
	if err := fltrCP1.Compile(); err != nil {
		t.Error(err)
	}
	dmCharger.SetFilter(fltrCP1, true)
	fltrCP2 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Charger",
				Values:  []string{"ChargerProfile2"},
			},
		},
	}
	dmCharger.SetFilter(fltrCP2, true)
	fltrCPPrefix := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.harger",
				Values:  []string{"Charger"},
			},
		},
	}
	dmCharger.SetFilter(fltrCPPrefix, true)
	fltrCP4 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_4",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"200.00"},
			},
		},
	}
	dmCharger.SetFilter(fltrCP4, true)

	for _, cp := range cPPs {
		if err = dmCharger.SetChargerProfile(cp, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//verify each charger from cache
	for _, cp := range cPPs {
		if tempCp, err := dmCharger.GetChargerProfile(cp.Tenant, cp.ID,
			true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(cp, tempCp) {
			t.Errorf("Expecting: %+v, received: %+v", cp, tempCp)
		}
	}

	if _, err = chargerSrv.matchingChargerProfilesForEvent(chargerEvents[2].Tenant, chargerEvents[2]); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Error: %+v", err)
	}

	if rcv, err := chargerSrv.matchingChargerProfilesForEvent(chargerEvents[0].Tenant, chargerEvents[0]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(cPPs[0], rcv[0]) {
		t.Errorf("Expecting: %+v, received: %+v ", cPPs[0], rcv[0])
	}

	if rcv, err := chargerSrv.matchingChargerProfilesForEvent(chargerEvents[1].Tenant, chargerEvents[1]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(cPPs[1], rcv[0]) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(cPPs[1]), utils.ToJSON(rcv))
	}
	rpl := []*ChrgSProcessEventReply{
		{
			ChargerSProfile: "CPP_1",
			AlteredFields:   []string{utils.MetaReqRunID},
			CGREvent:        chargerEvents[0],
		},
	}
	rpl[0].CGREvent.Event[utils.RunID] = cPPs[0].RunID
	rcv, err := chargerSrv.processEvent(rpl[0].CGREvent.Tenant, chargerEvents[0])
	if err != nil {
		t.Fatalf("Error: %+v", err)
	}
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

	dataDB := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmCharger := NewDataManager(dataDB, config.CgrConfig().CacheCfg(), nil)
	cS := &ChargerService{
		dm: dmCharger,
		filterS: &FilterS{
			dm:  dmCharger,
			cfg: cfg,
		},
		cfg: cfg,
	}
	cgrEv := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "cgrEvID",
		Event: map[string]interface{}{
			"Charger":        "ChargerProfile1",
			utils.AnswerTime: time.Date(2021, 4, 1, 10, 0, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			utils.Weight:     "10.0",
		},
		APIOpts: map[string]interface{}{
			utils.MetaSubsys: utils.MetaChargers,
		},
	}

	experr := utils.ErrNotFound
	rcv, err := cS.matchingChargerProfilesForEvent("tnt", cgrEv)

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

	dataDB := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmCharger := NewDataManager(dataDB, config.CgrConfig().CacheCfg(), nil)
	cS := &ChargerService{
		dm: dmCharger,
		filterS: &FilterS{
			dm:  dmCharger,
			cfg: cfg,
		},
		cfg: cfg,
	}
	cgrEv := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "cgrEvID",
		Event: map[string]interface{}{
			"Charger":        "ChargerProfile1",
			utils.AnswerTime: time.Date(2021, 4, 1, 10, 0, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			utils.Weight:     "10.0",
		},
		APIOpts: map[string]interface{}{
			utils.MetaSubsys: utils.MetaChargers,
		},
	}

	experr := utils.ErrNotFound
	rcv, err := cS.matchingChargerProfilesForEvent(cgrEv.Tenant, cgrEv)

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
		GetKeysForPrefixF: func(s string) ([]string, error) {
			return []string{":"}, nil
		},
	}
	dmCharger := NewDataManager(dbm, cfg.CacheCfg(), nil)
	cS := &ChargerService{
		dm: dmCharger,
		filterS: &FilterS{
			dm:  dmCharger,
			cfg: cfg,
		},
		cfg: cfg,
	}
	cgrEv := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "cgrEvID",
		Event: map[string]interface{}{
			"Charger":        "ChargerProfile1",
			utils.AnswerTime: time.Date(2021, 4, 1, 10, 0, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			utils.Weight:     "10.0",
		},
		APIOpts: map[string]interface{}{
			utils.MetaSubsys: utils.MetaChargers,
		},
	}

	experr := utils.ErrNotImplemented
	rcv, err := cS.matchingChargerProfilesForEvent(cgrEv.Tenant, cgrEv)

	if err == nil || err != experr {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}

}

func TestChargersprocessEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cS := &ChargerService{
		cfg: cfg,
	}
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProcessRuns: 2,
		},
	}

	experr := "NO_DATABASE_CONNECTION"
	rcv, err := cS.processEvent(cgrEv.Tenant, cgrEv)

	if err == nil || err.Error() != experr {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestChargersV1ProcessEventMissingArgs(t *testing.T) {
	cS := &ChargerService{}
	args := &utils.CGREvent{}
	var reply *[]*ChrgSProcessEventReply

	experr := "MANDATORY_IE_MISSING: [Event]"
	err := cS.V1ProcessEvent(args, reply)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestChargersShutdown(t *testing.T) {
	cS := &ChargerService{}

	utils.Logger.SetLogLevel(6)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	exp := []string{
		"CGRateS <> [INFO] <ChargerS> shutdown initialized",
		"CGRateS <> [INFO] <ChargerS> shutdown complete",
	}
	cS.Shutdown()
	rcv := strings.Split(buf.String(), "\n")

	for i := 0; i < 2; i++ {
		rcv[i] = rcv[i][20:]
		if rcv[i] != exp[i] {
			t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp[i], rcv[i])
		}
	}

	utils.Logger.SetLogLevel(0)
}
