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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestSMSSetStorageSmsChrg1(t *testing.T) {
	config.CgrConfig().CacheCfg()[utils.CacheRatingPlans].Precache = true // precache rating plan
	data, _ := engine.NewMapStorageJson()
	dataDB = engine.NewDataManager(data)
	engine.SetDataStorage(dataDB)
}

func TestSMSLoadCsvTpSmsChrg1(t *testing.T) {
	timings := `ALWAYS,*any,*any,*any,*any,00:00:00`
	rates := `RT_SMS_5c,0,0.005,1,1,0`
	destinationRates := `DR_SMS_1,*any,RT_SMS_5c,*up,4,0,`
	ratingPlans := `RP_SMS1,DR_SMS_1,ALWAYS,10`
	ratingProfiles := `*out,cgrates.org,sms,*any,2012-01-01T00:00:00Z,RP_SMS1,,`
	csvr := engine.NewTpReader(dataDB.DataDB(), engine.NewStringCSVStorage(',', "", timings, rates, destinationRates, ratingPlans, ratingProfiles,
		"", "", "", "", "", "", "", "", "", "", "", "", "", "", ""), "", "")
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
	csvr.WriteToDatabase(false, false, false)
	engine.Cache.Clear(nil)
	dataDB.LoadDataDBCache(nil, nil, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil, nil, nil)

	if cachedRPlans := len(engine.Cache.GetItemIDs(utils.CacheRatingPlans, "")); cachedRPlans != 1 {
		t.Error("Wrong number of cached rating plans found", cachedRPlans)
	}
	if cachedRProfiles := len(engine.Cache.GetItemIDs(utils.CacheRatingProfiles, "")); cachedRProfiles != 0 {
		t.Error("Wrong number of cached rating profiles found", cachedRProfiles)
	}
}

func TestSMSGetDataCostSmsChrg1(t *testing.T) {
	usageDur := time.Duration(1)
	timeStart := time.Date(2014, 3, 4, 0, 0, 0, 0, time.Local)
	cd := &engine.CallDescriptor{
		Direction:     "*out",
		Category:      "sms",
		Tenant:        "cgrates.org",
		Subject:       "12345",
		Account:       "12345",
		Destination:   "+4917621621391",
		TimeStart:     timeStart,
		TimeEnd:       timeStart.Add(usageDur),
		DurationIndex: usageDur,
		TOR:           utils.SMS,
	}
	if cc, err := cd.GetCost(); err != nil {
		t.Error(err)
	} else if cc.Cost != 0.005 {
		t.Error("Wrong cost returned: ", cc.Cost)
	}
}
