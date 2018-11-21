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
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/scheduler"
	"github.com/cgrates/cgrates/utils"
)

var dataDB *engine.DataManager

func TestDZ1SetStorage(t *testing.T) {
	data, _ := engine.NewMapStorageJson()
	dataDB = engine.NewDataManager(data)
	engine.SetDataStorage(dataDB)
}

func TestDZ1LoadCsvTp(t *testing.T) {
	timings := `ALWAYS,*any,*any,*any,*any,00:00:00
ASAP,*any,*any,*any,*any,*asap`
	destinations := `DST_UK_Mobile_BIG5,447596
DST_UK_Mobile_BIG5,447956`
	rates := `RT_UK_Mobile_BIG5_PKG,0.01,0,20s,20s,0s
RT_UK_Mobile_BIG5,0.01,0.10,1s,1s,0s`
	destinationRates := `DR_UK_Mobile_BIG5_PKG,DST_UK_Mobile_BIG5,RT_UK_Mobile_BIG5_PKG,*up,8,0,
DR_UK_Mobile_BIG5,DST_UK_Mobile_BIG5,RT_UK_Mobile_BIG5,*up,8,0,`
	ratingPlans := `RP_UK_Mobile_BIG5_PKG,DR_UK_Mobile_BIG5_PKG,ALWAYS,10
RP_UK,DR_UK_Mobile_BIG5,ALWAYS,10`
	ratingProfiles := `*out,cgrates.org,call,*any,2013-01-06T00:00:00Z,RP_UK,,
*out,cgrates.org,call,discounted_minutes,2013-01-06T00:00:00Z,RP_UK_Mobile_BIG5_PKG,,`
	sharedGroups := ``
	actions := `TOPUP10_AC,*topup_reset,,,,*monetary,*out,,*any,,,*unlimited,,10,10,false,false,10
TOPUP10_AC1,*topup_reset,,,,*voice,*out,,DST_UK_Mobile_BIG5,discounted_minutes,,*unlimited,,40000000000,10,false,false,10`
	actionPlans := `TOPUP10_AT,TOPUP10_AC,ASAP,10
TOPUP10_AT,TOPUP10_AC1,ASAP,10`
	actionTriggers := ``
	accountActions := `cgrates.org,12344,TOPUP10_AT,,,`
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
	csvr := engine.NewTpReader(dataDB.DataDB(),
		engine.NewStringCSVStorage(',', destinations, timings, rates,
			destinationRates, ratingPlans, ratingProfiles,
			sharedGroups, actions, actionPlans, actionTriggers, accountActions,
			derivedCharges, users, aliases, resLimits, stats,
			thresholds, filters, suppliers, aliasProfiles, chargerProfiles), "", "")
	if err := csvr.LoadDestinations(); err != nil {
		t.Fatal(err)
	}
	if err := csvr.LoadTimings(); err != nil {
		t.Fatal(err)
	}
	if err := csvr.LoadRates(); err != nil {
		t.Fatal(err)
	}
	if err := csvr.LoadDestinationRates(); err != nil {
		t.Fatal(err)
	}
	if err := csvr.LoadRatingPlans(); err != nil {
		t.Fatal(err)
	}
	if err := csvr.LoadRatingProfiles(); err != nil {
		t.Fatal(err)
	}
	if err := csvr.LoadSharedGroups(); err != nil {
		t.Fatal(err)
	}
	if err := csvr.LoadActions(); err != nil {
		t.Fatal(err)
	}
	if err := csvr.LoadActionPlans(); err != nil {
		t.Fatal(err)
	}
	if err := csvr.LoadActionTriggers(); err != nil {
		t.Fatal(err)
	}
	if err := csvr.LoadAccountActions(); err != nil {
		t.Fatal(err)
	}
	if err := csvr.LoadDerivedChargers(); err != nil {
		t.Fatal(err)
	}
	csvr.WriteToDatabase(false, false, false)
	if acnt, err := dataDB.DataDB().GetAccount("cgrates.org:12344"); err != nil {
		t.Error(err)
	} else if acnt == nil {
		t.Error("No account saved")
	}
	engine.Cache.Clear(nil)

	dataDB.LoadDataDBCache(nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	if cachedDests := len(engine.Cache.GetItemIDs(utils.CacheDestinations, "")); cachedDests != 0 {
		t.Error("Wrong number of cached destinations found", cachedDests)
	}
	if cachedRPlans := len(engine.Cache.GetItemIDs(utils.CacheRatingPlans, "")); cachedRPlans != 2 {
		t.Error("Wrong number of cached rating plans found", cachedRPlans)
	}
	if cachedRProfiles := len(engine.Cache.GetItemIDs(utils.CacheRatingProfiles, "")); cachedRProfiles != 0 {
		t.Error("Wrong number of cached rating profiles found", cachedRProfiles)
	}
	if cachedActions := len(engine.Cache.GetItemIDs(utils.CacheActions, "")); cachedActions != 0 {
		t.Error("Wrong number of cached actions found", cachedActions)
	}
}

func TestDZ1ExecuteActions(t *testing.T) {
	scheduler.NewScheduler(dataDB).Reload()
	time.Sleep(10 * time.Millisecond) // Give time to scheduler to topup the account
	if acnt, err := dataDB.DataDB().GetAccount("cgrates.org:12344"); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap) != 2 {
		t.Error("Account does not have enough balances: ", acnt.BalanceMap)
	} else if acnt.BalanceMap[utils.VOICE][0].Value != 40000000000 {
		t.Error("Account does not have enough minutes in balance",
			acnt.BalanceMap[utils.VOICE][0].Value)
	} else if acnt.BalanceMap[utils.MONETARY][0].Value != 10 {
		t.Error("Account does not have enough monetary balance",
			acnt.BalanceMap[utils.MONETARY][0].Value)
	}
}

func TestDZ1Debit(t *testing.T) {
	cd := &engine.CallDescriptor{
		Direction:   "*out",
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "12344",
		Account:     "12344",
		Destination: "447956933443",
		TimeStart:   time.Date(2014, 3, 4, 6, 0, 0, 0, time.UTC),
		TimeEnd:     time.Date(2014, 3, 4, 6, 0, 10, 0, time.UTC),
	}
	if cc, err := cd.Debit(); err != nil {
		t.Error(err)
	} else if cc.Cost != 0.01 {
		t.Error("Wrong cost returned: ", cc.Cost)
	}
	acnt, err := dataDB.DataDB().GetAccount("cgrates.org:12344")
	if err != nil {
		t.Error(err)
	}
	if acnt.BalanceMap[utils.VOICE][0].Value != 20000000000 {
		t.Error("Account does not have expected *voice units in balance",
			acnt.BalanceMap[utils.VOICE][0].Value)
	}
	if acnt.BalanceMap[utils.MONETARY][0].Value != 9.99 {
		t.Error("Account does not have expected *monetary units in balance",
			acnt.BalanceMap[utils.MONETARY][0].Value)
	}
}
