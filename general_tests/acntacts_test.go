// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var dbAcntActs *engine.DataManager

func TestAcntActsSetStorage(t *testing.T) {
	dataDB, err := engine.NewInternalDB(nil, nil, true, nil, config.CgrConfig().DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbAcntActs = engine.NewDataManager(dataDB, config.CgrConfig().CacheCfg(), nil)
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
	actions := `TOPUP10_AC,*topup_reset,,,,*voice,,*any,,,*unlimited,,10s,10,false,false,10
DISABLE_ACNT,*disable_account,,,,,,,,,,,,,false,false,10
ENABLE_ACNT,*enable_account,,,,,,,,,,,,,false,false,10`
	actionPlans := `TOPUP10_AT,TOPUP10_AC,ASAP,10`
	actionTriggers := ``
	accountActions := `cgrates.org,1,TOPUP10_AT,,,`
	resLimits := ``
	ips := ``
	stats := ``
	trends := ``
	rankings := ``
	thresholds := ``
	filters := ``
	suppliers := ``
	attrProfiles := ``
	chargerProfiles := ``
	csvr, err := engine.NewTpReader(dbAcntActs.DataDB(), engine.NewStringCSVStorage(utils.CSVSep, destinations, timings,
		rates, destinationRates, ratingPlans, ratingProfiles, sharedGroups,
		actions, actionPlans, actionTriggers, accountActions,
		resLimits, ips, stats, trends, rankings, thresholds, filters, suppliers, attrProfiles, chargerProfiles, ``, ""), "", "", nil, nil)
	if err != nil {
		t.Error(err)
	}
	if err := csvr.LoadAll(); err != nil {
		t.Fatal(err)
	}
	csvr.WriteToDatabase(false, false)
	engine.LoadAllDataDBToCache(dbAcntActs)

	expectAcnt := &engine.Account{ID: "cgrates.org:1"}
	if acnt, err := dbAcntActs.GetAccount("cgrates.org:1"); err != nil {
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
	if err := at.Execute(nil, "", nil); err != nil {
		t.Error(err)
	}
	expectAcnt := &engine.Account{ID: "cgrates.org:1", Disabled: true}
	if acnt, err := dbAcntActs.GetAccount(acnt1Tag); err != nil {
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
	if err := at.Execute(nil, "", nil); err != nil {
		t.Error(err)
	}
	expectAcnt := &engine.Account{ID: "cgrates.org:1", Disabled: false}
	if acnt, err := dbAcntActs.GetAccount(acnt1Tag); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectAcnt, acnt) {
		t.Errorf("Expecting: %+v, received: %+v", expectAcnt, acnt)
	}
}
