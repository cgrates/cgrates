// +build integration

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
package engine

import (
	"flag"
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// Globals used
var dataDbCsv, dataDbStor, dataDbApier *DataManager // Each dataDb will have it's own sources to collect data
var storDb LoadStorage
var lCfg *config.CGRConfig
var loader *TpReader

var tpCsvScenario = flag.String("tp_scenario", "testtp", "Use this scenario folder to import tp csv data from")

// Create connection to dataDb
// Will use 3 different datadbs in order to be able to see differences in data loaded
func TestLoaderITConnDataDbs(t *testing.T) {
	lCfg, _ = config.NewDefaultCGRConfig()
	lCfg.StorDBPass = "CGRateS.org"
	var err error
	if dataDbCsv, err = ConfigureDataStorage(lCfg.DataDbType, lCfg.DataDbHost, lCfg.DataDbPort, "7",
		lCfg.DataDbUser, lCfg.DataDbPass, lCfg.DBDataEncoding, nil, 1); err != nil {
		t.Fatal("Error on dataDb connection: ", err.Error())
	}
	if dataDbStor, err = ConfigureDataStorage(lCfg.DataDbType, lCfg.DataDbHost, lCfg.DataDbPort, "8",
		lCfg.DataDbUser, lCfg.DataDbPass, lCfg.DBDataEncoding, nil, 1); err != nil {
		t.Fatal("Error on dataDb connection: ", err.Error())
	}
	if dataDbApier, err = ConfigureDataStorage(lCfg.DataDbType, lCfg.DataDbHost, lCfg.DataDbPort, "9",
		lCfg.DataDbUser, lCfg.DataDbPass, lCfg.DBDataEncoding, nil, 1); err != nil {
		t.Fatal("Error on dataDb connection: ", err.Error())
	}
	for _, db := range []Storage{dataDbCsv.DataDB(), dataDbStor.DataDB(), dataDbApier.DataDB(), dataDbCsv.DataDB(), dataDbStor.DataDB(), dataDbApier.DataDB()} {
		if err = db.Flush(""); err != nil {
			t.Fatal("Error when flushing datadb")
		}
	}
}

// Create/reset storage tariff plan tables, used as database connectin establishment also
func TestLoaderITCreateStorTpTables(t *testing.T) {
	db, err := NewMySQLStorage(lCfg.StorDBHost, lCfg.StorDBPort, lCfg.StorDBName,
		lCfg.StorDBUser, lCfg.StorDBPass, lCfg.StorDBMaxOpenConns, lCfg.StorDBMaxIdleConns, lCfg.StorDBConnMaxLifetime)
	if err != nil {
		t.Error("Error on opening database connection: ", err)
		return
	} else {
		storDb = db
	}
	// Creating the table serves also as reset since there is a drop prior to create
	if err := db.CreateTablesFromScript(path.Join(*dataDir, "storage", "mysql", utils.CREATE_TARIFFPLAN_TABLES_SQL)); err != nil {
		t.Error("Error on db creation: ", err.Error())
		return // No point in going further
	}
}

// Loads data from csv files in tp scenario to dataDbCsv
func TestLoaderITRemoveLoad(t *testing.T) {
	/*var err error
	for fn, v := range FileValidators {
		if err = ValidateCSVData(path.Join(*dataDir, "tariffplans", *tpCsvScenario, fn), v.Rule); err != nil {
			t.Error("Failed validating data: ", err.Error())
		}
	}*/
	loader = NewTpReader(dataDbCsv.DataDB(), NewFileCSVStorage(utils.CSV_SEP,
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.DESTINATIONS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.TIMINGS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.RATES_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.DESTINATION_RATES_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.RATING_PLANS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.RATING_PROFILES_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.SHARED_GROUPS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.LCRS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.ACTIONS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.ACTION_PLANS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.ACTION_TRIGGERS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.ACCOUNT_ACTIONS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.DERIVED_CHARGERS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.CDR_STATS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.USERS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.ALIASES_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.ResourcesCsv),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.StatsCsv),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.ThresholdsCsv),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.FiltersCsv),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.SuppliersCsv),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.AttributesCsv),
	), "", "")

	if err = loader.LoadDestinations(); err != nil {
		t.Error("Failed loading destinations: ", err.Error())
	}
	if err = loader.LoadTimings(); err != nil {
		t.Error("Failed loading timings: ", err.Error())
	}
	if err = loader.LoadRates(); err != nil {
		t.Error("Failed loading rates: ", err.Error())
	}
	if err = loader.LoadDestinationRates(); err != nil {
		t.Error("Failed loading destination rates: ", err.Error())
	}
	if err = loader.LoadRatingPlans(); err != nil {
		t.Error("Failed loading rating plans: ", err.Error())
	}
	if err = loader.LoadRatingProfiles(); err != nil {
		t.Error("Failed loading rating profiles: ", err.Error())
	}
	if err = loader.LoadActions(); err != nil {
		t.Error("Failed loading actions: ", err.Error())
	}
	if err = loader.LoadActionPlans(); err != nil {
		t.Error("Failed loading action timings: ", err.Error())
	}
	if err = loader.LoadActionTriggers(); err != nil {
		t.Error("Failed loading action triggers: ", err.Error())
	}
	if err = loader.LoadAccountActions(); err != nil {
		t.Error("Failed loading account actions: ", err.Error())
	}
	if err = loader.LoadDerivedChargers(); err != nil {
		t.Error("Failed loading derived chargers: ", err.Error())
	}
	if err = loader.LoadLCRs(); err != nil {
		t.Error("Failed loading lcr rules: ", err.Error())
	}
	if err = loader.LoadUsers(); err != nil {
		t.Error("Failed loading users: ", err.Error())
	}
	if err = loader.LoadAliases(); err != nil {
		t.Error("Failed loading aliases: ", err.Error())
	}
	if err = loader.LoadFilters(); err != nil {
		t.Error("Failed loading filters: ", err.Error())
	}
	if err = loader.LoadResourceProfiles(); err != nil {
		t.Error("Failed loading resource profiles: ", err.Error())
	}
	if err = loader.LoadStats(); err != nil {
		t.Error("Failed loading stats: ", err.Error())
	}
	if err = loader.LoadThresholds(); err != nil {
		t.Error("Failed loading thresholds: ", err.Error())
	}
	if err = loader.LoadSupplierProfiles(); err != nil {
		t.Error("Failed loading Supplier profiles: ", err.Error())
	}
	if err = loader.LoadAttributeProfiles(); err != nil {
		t.Error("Failed loading Alias profiles: ", err.Error())
	}
	if err := loader.WriteToDatabase(true, false, false); err != nil {
		t.Error("Could not write data into dataDb: ", err.Error())
	}
	if err := loader.RemoveFromDatabase(false, true); err != nil {
		t.Error("Could not write data into dataDb: ", err.Error())
	}
}

// Loads data from csv files in tp scenario to dataDbCsv
func TestLoaderITLoadFromCSV(t *testing.T) {
	/*var err error
	for fn, v := range FileValidators {
		if err = ValidateCSVData(path.Join(*dataDir, "tariffplans", *tpCsvScenario, fn), v.Rule); err != nil {
			t.Error("Failed validating data: ", err.Error())
		}
	}*/
	loader = NewTpReader(dataDbCsv.DataDB(), NewFileCSVStorage(utils.CSV_SEP,
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.DESTINATIONS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.TIMINGS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.RATES_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.DESTINATION_RATES_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.RATING_PLANS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.RATING_PROFILES_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.SHARED_GROUPS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.LCRS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.ACTIONS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.ACTION_PLANS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.ACTION_TRIGGERS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.ACCOUNT_ACTIONS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.DERIVED_CHARGERS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.CDR_STATS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.USERS_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.ALIASES_CSV),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.ResourcesCsv),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.StatsCsv),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.ThresholdsCsv),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.FiltersCsv),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.SuppliersCsv),
		path.Join(*dataDir, "tariffplans", *tpCsvScenario, utils.AttributesCsv),
	), "", "")

	if err = loader.LoadDestinations(); err != nil {
		t.Error("Failed loading destinations: ", err.Error())
	}
	if err = loader.LoadTimings(); err != nil {
		t.Error("Failed loading timings: ", err.Error())
	}
	if err = loader.LoadRates(); err != nil {
		t.Error("Failed loading rates: ", err.Error())
	}
	if err = loader.LoadDestinationRates(); err != nil {
		t.Error("Failed loading destination rates: ", err.Error())
	}
	if err = loader.LoadRatingPlans(); err != nil {
		t.Error("Failed loading rating plans: ", err.Error())
	}
	if err = loader.LoadRatingProfiles(); err != nil {
		t.Error("Failed loading rating profiles: ", err.Error())
	}
	if err = loader.LoadActions(); err != nil {
		t.Error("Failed loading actions: ", err.Error())
	}
	if err = loader.LoadActionPlans(); err != nil {
		t.Error("Failed loading action timings: ", err.Error())
	}
	if err = loader.LoadActionTriggers(); err != nil {
		t.Error("Failed loading action triggers: ", err.Error())
	}
	if err = loader.LoadAccountActions(); err != nil {
		t.Error("Failed loading account actions: ", err.Error())
	}
	if err = loader.LoadDerivedChargers(); err != nil {
		t.Error("Failed loading derived chargers: ", err.Error())
	}
	if err = loader.LoadLCRs(); err != nil {
		t.Error("Failed loading lcr rules: ", err.Error())
	}
	if err = loader.LoadUsers(); err != nil {
		t.Error("Failed loading users: ", err.Error())
	}
	if err = loader.LoadAliases(); err != nil {
		t.Error("Failed loading aliases: ", err.Error())
	}
	if err = loader.LoadFilters(); err != nil {
		t.Error("Failed loading filters: ", err.Error())
	}
	if err = loader.LoadResourceProfiles(); err != nil {
		t.Error("Failed loading resource profiles: ", err.Error())
	}
	if err = loader.LoadStats(); err != nil {
		t.Error("Failed loading stats: ", err.Error())
	}
	if err = loader.LoadThresholds(); err != nil {
		t.Error("Failed loading thresholds: ", err.Error())
	}
	if err = loader.LoadSupplierProfiles(); err != nil {
		t.Error("Failed loading Supplier profiles: ", err.Error())
	}
	if err = loader.LoadAttributeProfiles(); err != nil {
		t.Error("Failed loading Alias profiles: ", err.Error())
	}
	if err := loader.WriteToDatabase(true, false, false); err != nil {
		t.Error("Could not write data into dataDb: ", err.Error())
	}
}

func TestLoaderITWriteToDatabase(t *testing.T) {
	for k, as := range loader.actions {
		rcv, err := loader.dm.GetActions(k, true, utils.NonTransactional)
		if err != nil {
			t.Error("Failed GetActions: ", err.Error())
		}
		if !reflect.DeepEqual(as[0], rcv[0]) {
			t.Errorf("Expecting: %v, received: %v", as[0], rcv[0])
		}
	}

	for k, ap := range loader.actionPlans {
		rcv, err := loader.dm.DataDB().GetActionPlan(k, true, utils.NonTransactional)
		if err != nil {
			t.Error("Failed GetActionPlan: ", err.Error())
		}
		if !reflect.DeepEqual(ap.Id, rcv.Id) {
			t.Errorf("Expecting: %v, received: %v", ap.Id, rcv.Id)
		}
	}

	for k, atrs := range loader.actionsTriggers {
		rcv, err := loader.dm.GetActionTriggers(k, true, utils.NonTransactional)
		if err != nil {
			t.Error("Failed GetActionTriggers: ", err.Error())
		}
		if !reflect.DeepEqual(atrs[0].ActionsID, rcv[0].ActionsID) {
			t.Errorf("Expecting: %v, received: %v", atrs[0].ActionsID, rcv[0].ActionsID)
		}
	}

	for k, ub := range loader.accountActions {
		rcv, err := loader.dm.DataDB().GetAccount(k)
		if err != nil {
			t.Error("Failed GetAccount: ", err.Error())
		}
		if !reflect.DeepEqual(ub.GetID(), rcv.GetID()) {
			t.Errorf("Expecting: %v, received: %v", ub.GetID(), rcv.GetID())
		}
	}

	for k, d := range loader.destinations {
		rcv, err := loader.dm.DataDB().GetDestination(k, true, utils.NonTransactional)
		if err != nil {
			t.Error("Failed GetDestination: ", err.Error())
		}
		if !reflect.DeepEqual(d, rcv) {
			t.Errorf("Expecting: %v, received: %v", d, rcv)
		}
	}

	for k, tm := range loader.timings {
		rcv, err := loader.dm.GetTiming(k, true, utils.NonTransactional)
		if err != nil {
			t.Error("Failed GetTiming: ", err.Error())
		}
		if !reflect.DeepEqual(tm, rcv) {
			t.Errorf("Expecting: %v, received: %v", tm, rcv)
		}
	}

	for k, rp := range loader.ratingPlans {
		rcv, err := loader.dm.GetRatingPlan(k, true, utils.NonTransactional)
		if err != nil {
			t.Error("Failed GetRatingPlan: ", err.Error())
		}
		if !reflect.DeepEqual(rp.Id, rcv.Id) {
			t.Errorf("Expecting: %v, received: %v", rp.Id, rcv.Id)
		}
	}

	for k, rp := range loader.ratingProfiles {
		rcv, err := loader.dm.GetRatingProfile(k, true, utils.NonTransactional)
		if err != nil {
			t.Error("Failed GetRatingProfile: ", err.Error())
		}
		if !reflect.DeepEqual(rp, rcv) {
			t.Errorf("Expecting: %v, received: %v", rp, rcv)
		}
	}

	for k, sg := range loader.sharedGroups {
		rcv, err := loader.dm.GetSharedGroup(k, true, utils.NonTransactional)
		if err != nil {
			t.Error("Failed GetSharedGroup: ", err.Error())
		}
		if !reflect.DeepEqual(sg, rcv) {
			t.Errorf("Expecting: %v, received: %v", sg, rcv)
		}
	}

	for k, lcr := range loader.lcrs {
		rcv, err := loader.dm.GetLCR(k, true, utils.NonTransactional)
		if err != nil {
			t.Error("Failed GetLCR: ", err.Error())
		}
		if !reflect.DeepEqual(lcr, rcv) {
			t.Errorf("Expecting: %v, received: %v", lcr, rcv)
		}
	}

	for k, dcs := range loader.derivedChargers {
		rcv, err := loader.dm.GetDerivedChargers(k, true, utils.NonTransactional)
		if err != nil {
			t.Error("Failed GetDerivedChargers: ", err.Error())
		}
		if !reflect.DeepEqual(dcs.DestinationIDs, rcv.DestinationIDs) {
			t.Errorf("Expecting: %v, received: %v", dcs.DestinationIDs, rcv.DestinationIDs)
		}
	}

	for k, sq := range loader.cdrStats {
		rcv, err := loader.dm.GetCdrStats(k)
		// t.Log(utils.ToIJSON(sq))
		// t.Log(utils.ToIJSON(rcv))
		t.Log(k)
		if err != nil {
			t.Error("Failed GetCdrStats: ", err.Error())
		}
		if !reflect.DeepEqual(sq, rcv) {
			t.Errorf("Expecting: %v, received: %v", sq, rcv)
		}
	}

	for k, u := range loader.users {
		rcv, err := loader.dm.GetUser(k)
		if err != nil {
			t.Error("Failed GetUser: ", err.Error())
		}
		if !reflect.DeepEqual(u, rcv) {
			t.Errorf("Expecting: %v, received: %v", u, rcv)
		}
	}

	for k, al := range loader.aliases {
		rcv, err := loader.dm.DataDB().GetAlias(k, true, utils.NonTransactional)
		if err != nil {
			t.Error("Failed GetAlias: ", err.Error())
		}
		if !reflect.DeepEqual(al, rcv) {
			t.Errorf("Expecting: %v, received: %v", al, rcv)
		}
	}

	for tenantid, rl := range loader.resProfiles {
		rcv, err := loader.dm.GetResourceProfile(tenantid.Tenant, tenantid.ID, true, utils.NonTransactional)
		if err != nil {
			t.Error("Failed GetResourceProfile: ", err.Error())
		}
		rlT, err := APItoResource(rl, "UTC")
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(rlT, rcv) {
			t.Errorf("Expecting: %v, received: %v", rlT, rcv)
		}
	}
	for tenantid, st := range loader.sqProfiles {
		rcv, err := loader.dm.GetStatQueueProfile(tenantid.Tenant, tenantid.ID, true, utils.NonTransactional)
		if err != nil {
			t.Errorf("Failed GetStatsQueue, tenant: %s, id: %s,  error: %s ", tenantid.Tenant, tenantid.ID, err.Error())
		}
		sts, err := APItoStats(st, "UTC")
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(sts, rcv) {
			t.Errorf("Expecting: %v, received: %v", sts, rcv)
		}
	}

	for tenatid, th := range loader.thProfiles {
		rcv, err := loader.dm.GetThresholdProfile(tenatid.Tenant, tenatid.ID, true, utils.NonTransactional)
		if err != nil {
			t.Errorf("Failed GetThresholdProfile, tenant: %s, id: %s,  error: %s ", th.Tenant, th.ID, err.Error())
		}
		sts, err := APItoThresholdProfile(th, "UTC")
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(sts, rcv) {
			t.Errorf("Expecting: %v, received: %v", sts, rcv)
		}
	}

	for tenatid, th := range loader.sppProfiles {
		rcv, err := loader.dm.GetSupplierProfile(tenatid.Tenant, tenatid.ID, true, utils.NonTransactional)
		if err != nil {
			t.Errorf("Failed GetSupplierProfile, tenant: %s, id: %s,  error: %s ", th.Tenant, th.ID, err.Error())
		}
		sts, err := APItoSupplierProfile(th, "UTC")
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(sts, rcv) {
			t.Errorf("Expecting: %v, received: %v", sts, rcv)
		}
	}

	for tenatid, th := range loader.attributeProfiles {
		rcv, err := loader.dm.GetAttributeProfile(tenatid.Tenant, tenatid.ID, true, utils.NonTransactional)
		if err != nil {
			t.Errorf("Failed GetAttributeProfile, tenant: %s, id: %s,  error: %s ", th.Tenant, th.ID, err.Error())
		}
		sts, err := APItoAttributeProfile(th, "UTC")
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(sts, rcv) {
			t.Errorf("Expecting: %v, received: %v", sts, rcv)
		}
	}

}

// Imports data from csv files in tpScenario to storDb
func TestLoaderITImportToStorDb(t *testing.T) {
	csvImporter := TPCSVImporter{
		TPid:     utils.TEST_SQL,
		StorDb:   storDb,
		DirPath:  path.Join(*dataDir, "tariffplans", *tpCsvScenario),
		Sep:      utils.CSV_SEP,
		Verbose:  false,
		ImportId: utils.TEST_SQL}
	if err := csvImporter.Run(); err != nil {
		t.Error("Error when importing tpdata to storDb: ", err)
	}
	if tpids, err := storDb.GetTpIds(""); err != nil {
		t.Error("Error when querying storDb for imported data: ", err)
	} else if len(tpids) != 1 || tpids[0] != utils.TEST_SQL {
		t.Errorf("Data in storDb is different than expected %v", tpids)
	}
}

// Loads data from storDb into dataDb
func TestLoaderITLoadFromStorDb(t *testing.T) {

	loader := NewTpReader(dataDbStor.DataDB(), storDb, utils.TEST_SQL, "")
	if err := loader.LoadDestinations(); err != nil && err.Error() != utils.NotFoundCaps {
		t.Error("Failed loading destinations: ", err.Error())
	}
	if err := loader.LoadTimings(); err != nil && err.Error() != utils.NotFoundCaps {
		t.Error("Failed loading timings: ", err.Error())
	}
	if err := loader.LoadRates(); err != nil && err.Error() != utils.NotFoundCaps {
		t.Error("Failed loading rates: ", err.Error())
	}
	if err := loader.LoadDestinationRates(); err != nil && err.Error() != utils.NotFoundCaps {
		t.Error("Failed loading destination rates: ", err.Error())
	}
	if err := loader.LoadRatingPlans(); err != nil && err.Error() != utils.NotFoundCaps {
		t.Error("Failed loading rating plans: ", err.Error())
	}
	if err := loader.LoadRatingProfiles(); err != nil && err.Error() != utils.NotFoundCaps {
		t.Error("Failed loading rating profiles: ", err.Error())
	}
	if err := loader.LoadActions(); err != nil && err.Error() != utils.NotFoundCaps {
		t.Error("Failed loading actions: ", err.Error())
	}
	if err := loader.LoadActionPlans(); err != nil && err.Error() != utils.NotFoundCaps {
		t.Error("Failed loading action timings: ", err.Error())
	}
	if err := loader.LoadActionTriggers(); err != nil && err.Error() != utils.NotFoundCaps {
		t.Error("Failed loading action triggers: ", err.Error())
	}
	if err := loader.LoadAccountActions(); err != nil && err.Error() != utils.NotFoundCaps {
		t.Error("Failed loading account actions: ", err.Error())
	}
	if err := loader.LoadDerivedChargers(); err != nil && err.Error() != utils.NotFoundCaps {
		t.Error("Failed loading derived chargers: ", err.Error())
	}
	if err := loader.LoadLCRs(); err != nil && err.Error() != utils.NotFoundCaps {
		t.Error("Failed loading lcr rules: ", err.Error())
	}
	if err := loader.LoadUsers(); err != nil && err.Error() != utils.NotFoundCaps {
		t.Error("Failed loading users: ", err.Error())
	}
	if err := loader.LoadAliases(); err != nil && err.Error() != utils.NotFoundCaps {
		t.Error("Failed loading aliases: ", err.Error())
	}
}

func TestLoaderITLoadIndividualProfiles(t *testing.T) {
	loader := NewTpReader(dataDbApier.DataDB(), storDb, utils.TEST_SQL, "")
	// Load ratingPlans. This will also set destination keys
	if rps, err := storDb.GetTPRatingPlans(utils.TEST_SQL, "", nil); err != nil {
		t.Fatal("Could not retrieve rating plans")
	} else {
		for _, r := range rps {
			if loaded, err := loader.LoadRatingPlansFiltered(r.ID); err != nil {
				t.Fatalf("Could not load ratingPlan for id: %s, error: %s", r.ID, err.Error())
			} else if !loaded {
				t.Fatal("Cound not find ratingPLan with id:", r.ID)
			}
		}
	}
	// Load rating profiles
	loadId := utils.CSV_LOAD + "_" + utils.TEST_SQL
	if rprs, err := storDb.GetTPRatingProfiles(&utils.TPRatingProfile{TPid: utils.TEST_SQL, LoadId: loadId}); err != nil {
		t.Fatal("Could not retrieve rating profiles, error: ", err.Error())
	} else if len(rprs) == 0 {
		t.Fatal("Could not retrieve rating profiles")
	} else {
		for _, r := range rprs {
			if err := loader.LoadRatingProfilesFiltered(r); err != nil {
				t.Fatalf("Could not load ratingProfile with id: %s, error: %s", r.KeyId(), err.Error())
			}
		}
	}
	// Load derived chargers
	loadId = utils.CSV_LOAD + "_" + utils.TEST_SQL
	if dcs, err := storDb.GetTPDerivedChargers(&utils.TPDerivedChargers{TPid: utils.TEST_SQL, LoadId: loadId}); err != nil {
		t.Fatal("Could not retrieve derived chargers, error: ", err.Error())
	} else if len(dcs) == 0 {
		t.Fatal("Could not retrieve derived chargers")
	} else {
		for _, d := range dcs {
			if err := loader.LoadDerivedChargersFiltered(d, true); err != nil {
				t.Fatalf("Could not load derived charger with id: %s, error: %s", d.GetDerivedChargesId(), err.Error())
			}
		}
	}
	// Load cdr stats
	//loadId = utils.CSV_LOAD + "_" + utils.TEST_SQL
	if css, err := storDb.GetTPCdrStats(utils.TEST_SQL, ""); err != nil {
		t.Fatal("Could not retrieve cdr stats, error: ", err.Error())
	} else if len(css) == 0 {
		t.Fatal("Could not retrieve cdr stats")
	} else {
		for _, c := range css {
			if err := loader.LoadCdrStatsFiltered(c.ID, true); err != nil {
				t.Fatalf("Could not load cdr stats with id: %s, error: %s", c.ID, err.Error())
			}
		}
	}
	// Load users
	if us, err := storDb.GetTPUsers(&utils.TPUsers{TPid: utils.TEST_SQL}); err != nil {
		t.Fatal("Could not retrieve users, error: ", err.Error())
	} else if len(us) == 0 {
		t.Fatal("Could not retrieve users")
	} else {
		for _, u := range us {
			if found, err := loader.LoadUsersFiltered(u); found && err != nil {
				t.Fatalf("Could not user with id: %s, error: %s", u.GetId(), err.Error())
			}
		}
	}
	// Load aliases
	if aliases, err := storDb.GetTPAliases(&utils.TPAliases{TPid: utils.TEST_SQL}); err != nil {
		t.Fatal("Could not retrieve aliases, error: ", err.Error())
	} else if len(aliases) == 0 {
		t.Fatal("Could not retrieve aliases")
	} else {
		for _, a := range aliases {
			if found, err := loader.LoadAliasesFiltered(a); found && err != nil {
				t.Fatalf("Could not load aliase with id: %s, error: %s", a.GetId(), err.Error())
			}
		}
	}
	// Load account actions
	if aas, err := storDb.GetTPAccountActions(&utils.TPAccountActions{TPid: utils.TEST_SQL, LoadId: loadId}); err != nil {
		t.Fatal("Could not retrieve account action profiles, error: ", err.Error())
	} else if len(aas) == 0 {
		t.Error("No account actions")
	} else {

		for _, a := range aas {
			if err := loader.LoadAccountActionsFiltered(a); err != nil {
				t.Fatalf("Could not load account actions with id: %s, error: %s", a.GetId(), err.Error())
			}
		}
	}
}

/*
// Compares previously loaded data from csv and stor to be identical, redis specific tests
func TestMatchLoadCsvWithStorRating(t *testing.T) {

	rsCsv, redisDb := dataDbCsv.(*RedisStorage)
	if !redisDb {
		return // We only support these tests for redis
	}
	rsStor := dataDbStor.(*RedisStorage)
	rsApier := dataDbApier.(*RedisStorage)
	keysCsv, err := rsCsv.db.Cmd("KEYS", "*").List()
	if err != nil {
		t.Fatal("Failed querying redis keys for csv data")
	}
	for _, key := range keysCsv {
		var refVal []byte
		for idx, rs := range []*RedisStorage{rsCsv, rsStor, rsApier} {
			if key == utils.TASKS_KEY || strings.HasPrefix(key, utils.ACTION_PLAN_PREFIX) { // action plans are not consistent
				continue
			}
			qVal, err := rs.db.Cmd("GET", key).Bytes()
			if err != nil {
				t.Fatalf("Run: %d, could not retrieve key %s, error: %s", idx, key, err.Error())
			}
			if idx == 0 { // Only compare at second iteration, first one is to set reference value
				refVal = qVal
				continue
			}
			if len(refVal) != len(qVal) {
				t.Errorf("Missmatched data for key: %s\n\t reference val: %s \n\t retrieved val: %s\n on iteration: %d", key, refVal, qVal, idx)
			}
		}
	}
}

func TestMatchLoadCsvWithStorAccounting(t *testing.T) {

	rsCsv, redisDb := dataDbCsv.(*RedisStorage)
	if !redisDb {
		return // We only support these tests for redis
	}
	rsStor := dataDbStor.(*RedisStorage)
	rsApier := dataDbApier.(*RedisStorage)
	keysCsv, err := rsCsv.db.Cmd("KEYS", "*").List()
	if err != nil {
		t.Fatal("Failed querying redis keys for csv data")
	}
	for _, key := range keysCsv {
		var refVal []byte
		if key == "load_history" {
			continue
		}
		for idx, rs := range []*RedisStorage{rsCsv, rsStor, rsApier} {
			qVal, err := rs.db.Cmd("GET", key).Bytes()
			if err != nil {
				t.Fatalf("Run: %d, could not retrieve key %s, error: %s", idx, key, err.Error())
			}
			if idx == 0 { // Only compare at second iteration, first one is to set reference value
				refVal = qVal
				continue
			}
			if len(refVal) != len(qVal) {
				t.Errorf("Missmatched data for key: %s\n\t, reference val: %s \n\t retrieved value: %s\n on iteration: %d", key, refVal, qVal, idx)
			}
		}
	}
}
*/
