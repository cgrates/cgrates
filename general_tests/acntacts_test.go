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
package general_tests

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var dbAcntActs *engine.DataManager

func TestAcntActsSetStorage(t *testing.T) {
	data, _ := engine.NewMapStorageJson()
	dbAcntActs = engine.NewDataManager(data)
	engine.SetDataStorage(dbAcntActs)
}

func TestAcntActsLoadCsv(t *testing.T) {
	timings := `ASAP,*any,*any,*any,*any,*asap`
	destinations := ``
	rates := ``
	destinationRates := ``
	ratingPlans := ``
	ratingProfiles := ``
	sharedGroups := ``
	actions := `TOPUP10_AC,*topup_reset,,,,*voice,*out,,*any,,,*unlimited,,10s,10,false,false,10
DISABLE_ACNT,*disable_account,,,,,,,,,,,,,,false,false,10
ENABLE_ACNT,*enable_account,,,,,,,,,,,,,,false,false,10`
	actionPlans := `TOPUP10_AT,TOPUP10_AC,ASAP,10`
	actionTriggers := ``
	accountActions := `cgrates.org,1,TOPUP10_AT,,,`
	derivedCharges := ``
	users := ``
	aliases := ``
	resLimits := ``
	stats := ``
	thresholds := ``
	filters := ``
	suppliers := ``
	aliasProfiles := ``
	chargerProfiles := ``
	csvr := engine.NewTpReader(dbAcntActs.DataDB(), engine.NewStringCSVStorage(',', destinations, timings,
		rates, destinationRates, ratingPlans, ratingProfiles, sharedGroups,
		actions, actionPlans, actionTriggers, accountActions, derivedCharges,
		users, aliases, resLimits, stats, thresholds, filters, suppliers, aliasProfiles, chargerProfiles), "", "")
	if err := csvr.LoadAll(); err != nil {
		t.Fatal(err)
	}
	csvr.WriteToDatabase(false, false, false)

	engine.Cache.Clear(nil)
	dbAcntActs.LoadDataDBCache(nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil, nil, nil, nil)

	expectAcnt := &engine.Account{ID: "cgrates.org:1"}
	if acnt, err := dbAcntActs.DataDB().GetAccount("cgrates.org:1"); err != nil {
		t.Error(err)
	} else if acnt == nil {
		t.Error("No account created")
	} else if !reflect.DeepEqual(expectAcnt.ActionTriggers, acnt.ActionTriggers) {
		t.Errorf("Expecting: %+v, received: %+v", expectAcnt, acnt)
	}
}

func TestAcntActsDisableAcnt(t *testing.T) {
	acnt1Tag := "cgrates.org:1"
	at := &engine.ActionTiming{
		ActionsID: "DISABLE_ACNT",
	}
	at.SetAccountIDs(utils.StringMap{acnt1Tag: true})
	if err := at.Execute(nil, nil); err != nil {
		t.Error(err)
	}
	expectAcnt := &engine.Account{ID: "cgrates.org:1", Disabled: true}
	if acnt, err := dbAcntActs.DataDB().GetAccount(acnt1Tag); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectAcnt, acnt) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectAcnt), utils.ToJSON(acnt))
	}
}

func TestAcntActsEnableAcnt(t *testing.T) {
	acnt1Tag := "cgrates.org:1"
	at := &engine.ActionTiming{
		ActionsID: "ENABLE_ACNT",
	}
	at.SetAccountIDs(utils.StringMap{acnt1Tag: true})
	if err := at.Execute(nil, nil); err != nil {
		t.Error(err)
	}
	expectAcnt := &engine.Account{ID: "cgrates.org:1", Disabled: false}
	if acnt, err := dbAcntActs.DataDB().GetAccount(acnt1Tag); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectAcnt, acnt) {
		t.Errorf("Expecting: %+v, received: %+v", expectAcnt, acnt)
	}
}
