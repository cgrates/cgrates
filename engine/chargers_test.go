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
	chargerSrv *ChargerService
	dmCharger  *DataManager
	cPPs       = ChargerProfiles{
		&ChargerProfile{
			Tenant:    "cgrates.org",
			ID:        "CPP_1",
			FilterIDs: []string{"FLTR_CP_1", "FLTR_CP_4"},
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
	chargerEvents = []*utils.CGREvent{
		{
			Tenant:  config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:      utils.GenUUID(),
			Context: utils.StringPointer(utils.MetaSessionS),
			Event: map[string]interface{}{
				"Charger":        "ChargerProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				utils.Weight:     "200.0",
			},
		},
		{
			Tenant:  config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:      utils.GenUUID(),
			Context: utils.StringPointer(utils.MetaSessionS),
			Event: map[string]interface{}{
				"Charger":        "ChargerProfile2",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			},
		},
		{
			Tenant:  config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:      utils.GenUUID(),
			Context: utils.StringPointer(utils.MetaSessionS),
			Event: map[string]interface{}{
				"Charger":        "DistinctMatch",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			},
		},
	}
)

func TestChargerPopulateChargerService(t *testing.T) {
	data, _ := NewMapStorage()
	dmCharger = NewDataManager(data)
	defaultCfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	chargerSrv, err = NewChargerService(dmCharger,
		&FilterS{dm: dmCharger, cfg: defaultCfg}, nil, defaultCfg)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
}

func TestChargerAddFilter(t *testing.T) {
	fltrCP1 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_1",
		Rules: []*FilterRule{
			{
				Type:      MetaString,
				FieldName: "Charger",
				Values:    []string{"ChargerProfile1"},
			},
			{
				Type:      MetaGreaterOrEqual,
				FieldName: "UsageInterval",
				Values:    []string{(1 * time.Second).String()},
			},
		},
	}
	dmCharger.SetFilter(fltrCP1)
	fltrCP2 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_2",
		Rules: []*FilterRule{
			{
				Type:      MetaString,
				FieldName: "Charger",
				Values:    []string{"ChargerProfile2"},
			},
		},
	}
	dmCharger.SetFilter(fltrCP2)
	fltrCPPrefix := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_3",
		Rules: []*FilterRule{
			{
				Type:      MetaPrefix,
				FieldName: "Charger",
				Values:    []string{"Charger"},
			},
		},
	}
	dmCharger.SetFilter(fltrCPPrefix)
	fltrCP4 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_CP_4",
		Rules: []*FilterRule{
			{
				Type:      MetaGreaterOrEqual,
				FieldName: utils.Weight,
				Values:    []string{"200.00"},
			},
		},
	}
	dmCharger.SetFilter(fltrCP4)
}

func TestChargerSetChargerProfiles(t *testing.T) {
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
	if _, err = chargerSrv.matchingChargerProfilesForEvent(chargerEvents[2]); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Error: %+v", err)
	}

	if rcv, err := chargerSrv.matchingChargerProfilesForEvent(chargerEvents[0]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(cPPs[0], rcv[0]) {
		t.Errorf("Expecting: %+v, received: %+v ", cPPs[0], rcv[0])
	}

	if rcv, err := chargerSrv.matchingChargerProfilesForEvent(chargerEvents[1]); err != nil {
		t.Errorf("Error: %+v", err)
	} else if !reflect.DeepEqual(cPPs[1], rcv[0]) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(cPPs[1]), utils.ToJSON(rcv))
	}

}

func TestChargerProcessEvent(t *testing.T) {
	rpl := []*ChrgSProcessEventReply{
		{
			ChargerSProfile: "CPP_1",
			CGREvent:        chargerEvents[0],
		},
	}
	rpl[0].CGREvent.Event[utils.RunID] = cPPs[0].RunID
	rcv, err := chargerSrv.processEvent(chargerEvents[0])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(rpl[0], rcv[0]) {
		t.Errorf("Expecting: %+v, received: %+v ", utils.ToJSON(rpl[0]), utils.ToJSON(rcv[0]))
	}
}
