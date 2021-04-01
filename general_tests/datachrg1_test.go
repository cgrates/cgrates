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

func TestSetStorageDtChrg1(t *testing.T) {
	dataDBInt := engine.NewInternalDB(nil, nil, true)
	dataDB = engine.NewDataManager(dataDBInt, config.CgrConfig().CacheCfg(), nil)
	engine.SetDataStorage(dataDB)
	engine.Cache.Clear(nil)
}

func TestLoadCsvTpDtChrg1(t *testing.T) {
	timings := `TM1,*any,*any,*any,*any,00:00:00
TM2,*any,*any,*any,*any,01:00:00`
	rates := `RT_DATA_2c,0,0.002,10s,10s,0
RT_DATA_1c,0,0.001,10,10,0`
	destinationRates := `DR_DATA_1,*any,RT_DATA_2c,*up,4,0,
DR_DATA_2,*any,RT_DATA_1c,*up,4,0,`
	ratingPlans := `RP_DATA1,DR_DATA_1,TM1,10
RP_DATA1,DR_DATA_2,TM2,10`
	ratingProfiles := `cgrates.org,data,*any,2012-01-01T00:00:00Z,RP_DATA1,`
	csvr, err := engine.NewTpReader(dataDB.DataDB(), engine.NewStringCSVStorage(utils.CSVSep,
		utils.EmptyString, timings, rates, destinationRates, ratingPlans, ratingProfiles,
		utils.EmptyString, utils.EmptyString, utils.EmptyString, utils.EmptyString, utils.EmptyString,
		utils.EmptyString, utils.EmptyString, utils.EmptyString, utils.EmptyString, utils.EmptyString,
		utils.EmptyString, utils.EmptyString, utils.EmptyString, utils.EmptyString),
		utils.EmptyString, utils.EmptyString, nil, nil, false)
	if err != nil {
		t.Error(err)
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
	csvr.WriteToDatabase(false, false)
	dataDB.LoadDataDBCache(engine.GetDefaultEmptyArgCachePrefix())

	if cachedRPlans := len(engine.Cache.GetItemIDs(utils.CacheRatingPlans, utils.EmptyString)); cachedRPlans != 1 {
		t.Error("Wrong number of cached rating plans found", cachedRPlans)
	}
	if cachedRProfiles := len(engine.Cache.GetItemIDs(utils.CacheRatingProfiles, utils.EmptyString)); cachedRProfiles != 1 {
		t.Error("Wrong number of cached rating profiles found", cachedRProfiles)
	}
}

func TestGetDataCostDtChrg1(t *testing.T) {
	usedData := 20
	usageDur := time.Duration(usedData) * time.Second
	timeStart := time.Date(2014, 3, 4, 0, 0, 0, 0, time.Local)
	cd := &engine.CallDescriptor{
		Category:      "data",
		Tenant:        "cgrates.org",
		Subject:       "12345",
		Account:       "12345",
		TimeStart:     timeStart,
		TimeEnd:       timeStart.Add(usageDur),
		DurationIndex: usageDur,
		ToR:           utils.MetaData,
	}
	if cc, err := cd.GetCost(); err != nil {
		t.Error(err)
	} else if cc.Cost != 0.004 {
		t.Error("Wrong cost returned: ", cc.Cost)
	}
}

func TestGetDataCostSecondIntDtChrg1(t *testing.T) {
	usedData := 20
	usageDur := time.Duration(usedData)
	timeStart := time.Date(2014, 3, 4, 1, 0, 0, 0, time.Local)
	cd := &engine.CallDescriptor{
		Category:      "data",
		Tenant:        "cgrates.org",
		Subject:       "12345",
		Account:       "12345",
		TimeStart:     timeStart,
		TimeEnd:       timeStart.Add(usageDur),
		DurationIndex: usageDur,
		ToR:           utils.MetaData,
	}
	if cc, err := cd.GetCost(); err != nil {
		t.Error(err)
	} else if cc.Cost != 0.002 {
		t.Error("Wrong cost returned: ", cc.Cost)
	}
}

func TestGetDataBetweenCostDtChrg1(t *testing.T) {
	usedData := 20
	usageDur := time.Duration(usedData) * time.Second
	timeStart := time.Date(2014, 3, 4, 0, 59, 50, 0, time.Local)
	cd := &engine.CallDescriptor{
		Category:      "data",
		Tenant:        "cgrates.org",
		Subject:       "12345",
		Account:       "12345",
		TimeStart:     timeStart,
		TimeEnd:       timeStart.Add(usageDur),
		DurationIndex: usageDur,
		ToR:           utils.MetaData,
	}
	if cc, err := cd.GetCost(); err != nil {
		t.Error(err)
	} else if cc.Cost != 0.004 {
		//t.Logf("%+v", cc.Timespans[1].RateInterval.Timing)
		for _, ts := range cc.Timespans {
			t.Logf("TS: %+v", ts)
		}
		t.Error("Wrong cost returned: ", cc.Cost)
	}
}
