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

	"github.com/cgrates/cgrates/cache"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var dbAuth engine.DataDB
var rsponder *engine.Responder

func TestAuthSetStorage(t *testing.T) {
	dbAuth, _ = engine.NewMapStorageJson()
	engine.SetDataStorage(dbAuth)
	cfg, _ := config.NewDefaultCGRConfig()
	config.SetCgrConfig(cfg)
	rsponder = new(engine.Responder)

}

func TestAuthLoadCsv(t *testing.T) {
	timings := ``
	destinations := `DST_GERMANY_LANDLINE,49`
	rates := `RT_1CENTWITHCF,0.02,0.01,60s,60s,0s`
	destinationRates := `DR_GERMANY,DST_GERMANY_LANDLINE,RT_1CENTWITHCF,*up,8,,
DR_ANY_1CNT,*any,RT_1CENTWITHCF,*up,8,,`
	ratingPlans := `RP_1,DR_GERMANY,*any,10
RP_ANY,DR_ANY_1CNT,*any,10`
	ratingProfiles := `*out,cgrates.org,call,testauthpostpaid1,2013-01-06T00:00:00Z,RP_1,,
*out,cgrates.org,call,testauthpostpaid2,2013-01-06T00:00:00Z,RP_1,*any,
*out,cgrates.org,call,*any,2013-01-06T00:00:00Z,RP_ANY,,`
	sharedGroups := ``
	lcrs := ``
	actions := `TOPUP10_AC,*topup_reset,,,,*monetary,*out,,*any,,,*unlimited,,0,10,false,false,10`
	actionPlans := `TOPUP10_AT,TOPUP10_AC,*asap,10`
	actionTriggers := ``
	accountActions := `cgrates.org,testauthpostpaid1,TOPUP10_AT,,,`
	derivedCharges := ``
	cdrStats := ``
	users := ``
	aliases := ``
	resLimits := ``
	csvr := engine.NewTpReader(dbAuth, engine.NewStringCSVStorage(',', destinations, timings, rates, destinationRates, ratingPlans, ratingProfiles,
		sharedGroups, lcrs, actions, actionPlans, actionTriggers, accountActions, derivedCharges, cdrStats, users, aliases, resLimits), "", "")
	if err := csvr.LoadAll(); err != nil {
		t.Fatal(err)
	}
	csvr.WriteToDatabase(false, false, false)
	if acnt, err := dbAuth.GetAccount("cgrates.org:testauthpostpaid1"); err != nil {
		t.Error(err)
	} else if acnt == nil {
		t.Error("No account saved")
	}

	cache.Flush()
	dbAuth.LoadRatingCache(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	dbAuth.LoadAccountingCache(nil, nil, nil)

	if cachedDests := cache.CountEntries(utils.DESTINATION_PREFIX); cachedDests != 0 {
		t.Error("Wrong number of cached destinations found", cachedDests)
	}
	if cachedRPlans := cache.CountEntries(utils.RATING_PLAN_PREFIX); cachedRPlans != 2 {
		t.Error("Wrong number of cached rating plans found", cachedRPlans)
	}
	if cachedRProfiles := cache.CountEntries(utils.RATING_PROFILE_PREFIX); cachedRProfiles != 0 {
		t.Error("Wrong number of cached rating profiles found", cachedRProfiles)
	}
	if cachedActions := cache.CountEntries(utils.ACTION_PREFIX); cachedActions != 0 {
		t.Error("Wrong number of cached actions found", cachedActions)
	}
}

func TestAuthPostpaidNoAcnt(t *testing.T) {
	cdr := &engine.CDR{ToR: utils.VOICE, RequestType: utils.META_POSTPAID, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "nonexistent", Subject: "testauthpostpaid1",
		Destination: "4986517174963", SetupTime: time.Date(2015, 8, 27, 11, 26, 0, 0, time.UTC)}
	var maxSessionTime float64
	if err := rsponder.GetDerivedMaxSessionTime(cdr, &maxSessionTime); err == nil || err != utils.ErrAccountNotFound {
		t.Error(err)
	}
}

func TestAuthPostpaidNoDestination(t *testing.T) {
	// Test subject which does not have destination attached
	cdr := &engine.CDR{ToR: utils.VOICE, RequestType: utils.META_POSTPAID, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "testauthpostpaid1", Subject: "testauthpostpaid1",
		Destination: "441231234", SetupTime: time.Date(2015, 8, 27, 11, 26, 0, 0, time.UTC)}
	var maxSessionTime float64
	if err := rsponder.GetDerivedMaxSessionTime(cdr, &maxSessionTime); err == nil {
		t.Error("Expecting error for destination not allowed to subject")
	}
}

func TestAuthPostpaidFallbackDest(t *testing.T) {
	// Test subject which has fallback for destination
	cdr := &engine.CDR{ToR: utils.VOICE, RequestType: utils.META_POSTPAID, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "testauthpostpaid1", Subject: "testauthpostpaid2",
		Destination: "441231234", SetupTime: time.Date(2015, 8, 27, 11, 26, 0, 0, time.UTC)}
	var maxSessionTime float64
	if err := rsponder.GetDerivedMaxSessionTime(cdr, &maxSessionTime); err != nil {
		t.Error(err)
	} else if maxSessionTime != -1 {
		t.Error("Unexpected maxSessionTime received: ", maxSessionTime)
	}
}

func TestAuthPostpaidWithDestination(t *testing.T) {
	// Test subject which does not have destination attached
	cdr := &engine.CDR{ToR: utils.VOICE, RequestType: utils.META_POSTPAID, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "testauthpostpaid1", Subject: "testauthpostpaid1",
		Destination: "4986517174963", SetupTime: time.Date(2015, 8, 27, 11, 26, 0, 0, time.UTC)}
	var maxSessionTime float64
	if err := rsponder.GetDerivedMaxSessionTime(cdr, &maxSessionTime); err != nil {
		t.Error(err)
	} else if maxSessionTime != -1 {
		t.Error("Unexpected maxSessionTime received: ", maxSessionTime)
	}
}
