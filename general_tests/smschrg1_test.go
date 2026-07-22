// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestSMSSetStorageSmsChrg1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	config.SetCgrConfig(cfg)
	config.CgrConfig().CacheCfg().Partitions[utils.CacheRatingPlans].Precache = true // precache rating plan
	data, err := engine.NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dataDB = engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	engine.SetDataStorage(dataDB)
	engine.Cache.Clear(nil)
}

func TestSMSLoadCsvTpSmsChrg1(t *testing.T) {
	timings := `ALWAYS,*any,*any,*any,*any,00:00:00`
	rates := `RT_SMS_5c,0,0.005,1,1,0`
	destinationRates := `DR_SMS_1,*any,RT_SMS_5c,*up,4,0,`
	ratingPlans := `RP_SMS1,DR_SMS_1,ALWAYS,10`
	ratingProfiles := `cgrates.org,sms,*any,2012-01-01T00:00:00Z,RP_SMS1,`
	csvr, err := engine.NewTpReader(dataDB.DataDB(), engine.NewStringCSVStorage(utils.CSVSep,
		utils.EmptyString, timings, rates, destinationRates, ratingPlans, ratingProfiles,
		utils.EmptyString, utils.EmptyString, utils.EmptyString, utils.EmptyString,
		utils.EmptyString, utils.EmptyString, utils.EmptyString, utils.EmptyString,
		utils.EmptyString, utils.EmptyString, utils.EmptyString, utils.EmptyString,
		utils.EmptyString, utils.EmptyString, utils.EmptyString, utils.EmptyString,
		utils.EmptyString), utils.EmptyString, utils.EmptyString, nil, nil)
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
	engine.LoadAllDataDBToCache(dataDB)

	if cachedRPlans := len(engine.Cache.GetItemIDs(utils.CacheRatingPlans, utils.EmptyString)); cachedRPlans != 1 {
		t.Error("Wrong number of cached rating plans found", cachedRPlans)
	}
	if cachedRProfiles := len(engine.Cache.GetItemIDs(utils.CacheRatingProfiles, utils.EmptyString)); cachedRProfiles != 1 {
		t.Error("Wrong number of cached rating profiles found", cachedRProfiles)
	}
}

func TestSMSGetDataCostSmsChrg1(t *testing.T) {
	usageDur := time.Nanosecond
	timeStart := time.Date(2014, 3, 4, 0, 0, 0, 0, time.Local)
	cd := &engine.CallDescriptor{
		Category:      "sms",
		Tenant:        "cgrates.org",
		Subject:       "12345",
		Account:       "12345",
		Destination:   "+4917621621391",
		TimeStart:     timeStart,
		TimeEnd:       timeStart.Add(usageDur),
		DurationIndex: usageDur,
		ToR:           utils.MetaSMS,
	}
	if cc, err := cd.GetCost(); err != nil {
		t.Error(err)
	} else if cc.Cost != 0.005 {
		t.Error("Wrong cost returned: ", cc.Cost)
	}
}
