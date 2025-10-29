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
	chargerEvents = []*utils.CGREventWithArgDispatcher{
		{
			CGREvent: &utils.CGREvent{
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				ID:     utils.GenUUID(),
				Event: map[string]any{
					"Charger":        "ChargerProfile1",
					utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
					"UsageInterval":  "1s",
					utils.Weight:     "200.0",
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				ID:     utils.GenUUID(),
				Event: map[string]any{
					"Charger":        "ChargerProfile2",
					utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				ID:     utils.GenUUID(),
				Event: map[string]any{
					"Charger":        "DistinctMatch",
					utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				},
			},
		},
	}
)

func TestChargerPopulateChargerService(t *testing.T) {
	defaultCfg, _ := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, defaultCfg.DataDbCfg().Items)
	dmCharger = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	chargerSrv, err = NewChargerService(dmCharger,
		&FilterS{dm: dmCharger, cfg: defaultCfg}, defaultCfg, nil)
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
				Type:    utils.MetaString,
				Element: "~*req.Charger",
				Values:  []string{"ChargerProfile1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.UsageInterval",
				Values:  []string{(1 * time.Second).String()},
			},
		},
	}
	dmCharger.SetFilter(fltrCP1)
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
	dmCharger.SetFilter(fltrCP2)
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
	dmCharger.SetFilter(fltrCPPrefix)
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
			AlteredFields:   []string{utils.MetaReqRunID},
			CGREvent:        chargerEvents[0].CGREvent,
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

func TestChargerV1GetChargersForEvent(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	chgS := &ChargerService{
		dm:      dm,
		cfg:     cfg,
		filterS: NewFilterS(cfg, nil, dm),
	}
	args := &utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Charger":        "ChargerProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				utils.Weight:     "180.0",
			},
		},
	}
	flt := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_CP_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Charger",
				Values:  []string{"ChargerProfile1"},
			},
		},
	}
	if err := dm.SetFilter(flt); err != nil {
		t.Error(err)
	}
	chP := &ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "CPP_1",
		FilterIDs: []string{"FLTR_CP_1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		RunID:        "TestRunID",
		AttributeIDs: []string{"*none"},
		Weight:       20,
	}
	if err := dm.SetChargerProfile(chP, true); err != nil {
		t.Error(err)
	}
	var reply ChargerProfiles
	exp := ChargerProfiles{
		chP,
	}
	if err := chgS.V1GetChargersForEvent(args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, exp) {
		t.Errorf("Expected %v,Received %v", utils.ToJSON(exp), utils.ToJSON(reply))
	}

}

func TestChargerV1ProcessEvent(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cfg.ChargerSCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(_ *context.Context, _ string, _, _ any) error {
		return nil
	})
	connMngr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): clientConn,
	})
	chgS := &ChargerService{
		dm:      dm,
		cfg:     cfg,
		filterS: NewFilterS(cfg, nil, dm),
		connMgr: connMngr,
	}
	args := &utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Charger":        "ChargerProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				utils.Weight:     "180.0",
			},
		}}
	flt := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_CP_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Charger",
				Values:  []string{"ChargerProfile1"},
			},
		},
	}
	if err := dm.SetFilter(flt); err != nil {
		t.Error(err)
	}
	chP := &ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "CPP_1",
		FilterIDs: []string{"FLTR_CP_1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		RunID:        "TestRunID",
		AttributeIDs: []string{"ATTR_1"},
	}
	if err := dm.SetChargerProfile(chP, true); err != nil {
		t.Error(err)
	}
	var reply []*ChrgSProcessEventReply
	if err := chgS.V1ProcessEvent(args, &reply); err != nil {
		t.Error(err)
	}

}

func TestChargersV1ProcessEvent(t *testing.T) {
	cS := &ChargerService{}
	args := &utils.CGREventWithArgDispatcher{}

	err := cS.V1ProcessEvent(args, nil)

	if err != nil {
		if err.Error() != utils.NewErrMandatoryIeMissing("Event").Error() {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}
}

// func TestChSListenAndServe(t *testing.T) {
// 	cfg, _ := config.NewDefaultCGRConfig()
// 	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
// 	dm := NewDataManager(db, cfg.CacheCfg(), nil)
// 	exitChan := make(chan bool)

// 	go func() {
// 		time.Sleep(3 * time.Millisecond)
// 		exitChan <- true
// 	}()
// 	cS, err := NewChargerService(dm, nil, cfg, nil)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	go func() {
// 		if err := cS.ListenAndServe(exitChan); err != nil {
// 			t.Errorf("ListenAndServe returned an error: %v", err)
// 		}
// 	}()

// 	time.Sleep(5 * time.Millisecond)

// 	exitChan <- true

// 	time.Sleep(5 * time.Millisecond)
// }
