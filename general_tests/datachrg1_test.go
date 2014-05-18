/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package general_tests

import (
	"testing"
	"time"

	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestDataSetStorageDtChrg1(t *testing.T) {
	ratingDb, _ = engine.NewMapStorageJson()
	engine.SetRatingStorage(ratingDb)
	acntDb, _ = engine.NewMapStorageJson()
	engine.SetAccountingStorage(acntDb)
}

func TestDataLoadCsvTpDtChrg1(t *testing.T) {
	timings := `ALWAYS,*any,*any,*any,*any,00:00:00`
	rates := `RT_DATA_2c,0,0.002,10,10,0`
	destinationRates := `DR_DATA_1,*any,RT_DATA_2c,*up,4`
	ratingPlans := `RP_DATA1,DR_DATA_1,ALWAYS,10`
	ratingProfiles := `*out,cgrates.org,data,*any,2012-01-01T00:00:00Z,RP_DATA1,`
	csvr := engine.NewStringCSVReader(ratingDb, acntDb, ',', "", timings, rates, destinationRates, ratingPlans, ratingProfiles,
		"", "", "", "", "", "", "")
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
	csvr.WriteToDatabase(false, false)
	ratingDb.CacheRating(nil, nil, nil, nil, nil)

	if cachedRPlans := cache2go.CountEntries(engine.RATING_PLAN_PREFIX); cachedRPlans != 1 {
		t.Error("Wrong number of cached rating plans found", cachedRPlans)
	}
	if cachedRProfiles := cache2go.CountEntries(engine.RATING_PROFILE_PREFIX); cachedRProfiles != 1 {
		t.Error("Wrong number of cached rating profiles found", cachedRProfiles)
	}
}

func TestDataGetCostDtChrg1(t *testing.T) {
	usedData := 20
	usageDur := time.Duration(usedData) * time.Second
	timeStart := time.Date(2014, 3, 4, 6, 0, 0, 0, time.UTC)
	cd := &engine.CallDescriptor{
		Direction:     "*out",
		Category:      "data",
		Tenant:        "cgrates.org",
		Subject:       "12345",
		Account:       "12345",
		TimeStart:     timeStart,
		TimeEnd:       timeStart.Add(usageDur),
		DurationIndex: usageDur,
		TOR:           utils.DATA,
	}
	expected := 0.004
	if cc, err := cd.GetCost(); err != nil {
		t.Error(err)
	} else if cc.Cost != expected {
		t.Logf("CC: %+v", cc.Timespans[0].RateInterval.Rating)
		t.Errorf("expected: %v was: %v", expected, cc.Cost)
	}
}
