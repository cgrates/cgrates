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

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestCosts1SetStorage(t *testing.T) {
	data, _ := engine.NewMapStorageJson()
	dataDB = engine.NewDataManager(data)
	engine.SetDataStorage(dataDB)
}

func TestCosts1LoadCsvTp(t *testing.T) {
	timings := `ALWAYS,*any,*any,*any,*any,00:00:00
ASAP,*any,*any,*any,*any,*asap`
	dests := `GERMANY,+49
GERMANY_MOBILE,+4915
GERMANY_MOBILE,+4916
GERMANY_MOBILE,+4917`
	rates := `RT_1CENT,0,1,1s,1s,0s
RT_DATA_2c,0,0.002,10,10,0
RT_SMS_5c,0,0.005,1,1,0`
	destinationRates := `DR_RETAIL,GERMANY,RT_1CENT,*up,4,0,
DR_RETAIL,GERMANY_MOBILE,RT_1CENT,*up,4,0,
DR_DATA_1,*any,RT_DATA_2c,*up,4,0,
DR_SMS_1,*any,RT_SMS_5c,*up,4,0,`
	ratingPlans := `RP_RETAIL,DR_RETAIL,ALWAYS,10
RP_DATA1,DR_DATA_1,ALWAYS,10
RP_SMS1,DR_SMS_1,ALWAYS,10`
	ratingProfiles := `*out,cgrates.org,call,*any,2012-01-01T00:00:00Z,RP_RETAIL,,
*out,cgrates.org,data,*any,2012-01-01T00:00:00Z,RP_DATA1,,
*out,cgrates.org,sms,*any,2012-01-01T00:00:00Z,RP_SMS1,,`
	csvr := engine.NewTpReader(dataDB.DataDB(), engine.NewStringCSVStorage(',', dests, timings, rates, destinationRates, ratingPlans, ratingProfiles,
		"", "", "", "", "", "", "", "", "", "", "", "", "", "", ""), "", "")

	if err := csvr.LoadTimings(); err != nil {
		t.Fatal(err)
	}
	if err := csvr.LoadDestinations(); err != nil {
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
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil)

	if cachedRPlans := len(engine.Cache.GetItemIDs(utils.CacheRatingPlans, "")); cachedRPlans != 3 {
		t.Error("Wrong number of cached rating plans found", cachedRPlans)
	}
	if cachedRProfiles := len(engine.Cache.GetItemIDs(utils.CacheRatingProfiles, "")); cachedRProfiles != 0 {
		t.Error("Wrong number of cached rating profiles found", cachedRProfiles)
	}
}

func TestCosts1GetCost1(t *testing.T) {
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:30Z", "")
	cd := &engine.CallDescriptor{
		Direction:   "*out",
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "1001",
		Account:     "1001",
		Destination: "+4986517174963",
		TimeStart:   tStart,
		TimeEnd:     tEnd,
	}
	if cc, err := cd.GetCost(); err != nil {
		t.Error(err)
	} else if cc.Cost != 90 {
		t.Error("Wrong cost returned: ", cc.Cost)
	}
}

func TestCosts1GetCostZeroDuration(t *testing.T) {
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	cd := &engine.CallDescriptor{
		Direction:   "*out",
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "1001",
		Account:     "1001",
		Destination: "+4986517174963",
		TimeStart:   tStart,
		TimeEnd:     tEnd,
	}
	if cc, err := cd.GetCost(); err != nil {
		t.Error(err)
	} else if cc.Cost != 0 {
		t.Error("Wrong cost returned: ", cc.Cost)
	}
}

/* FixMe
func TestCosts1GetCost2(t *testing.T) {
	tStart, _ := utils.ParseTimeDetectLayout("2004-06-04T00:00:01Z")
	tEnd, _ := utils.ParseTimeDetectLayout("2004-06-04T00:01:01Z")
	cd := &engine.CallDescriptor{
		Direction:   "*out",
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "1001",
		Account:     "1001",
		Destination: "+4986517174963",
		TimeStart:   tStart,
		TimeEnd:     tEnd,
	}
	if cc, err := cd.GetCost(); err != nil {
		t.Error(err)
	} else if cc.Cost != 90 {
		t.Error("Wrong cost returned: ", cc.Cost)
	}
}
*/
