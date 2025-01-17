//go:build integration
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
	"strings"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	// Globals used
	dataDbCsv       *DataManager // Each dataDb will have it's own sources to collect data
	lCfg            *config.CGRConfig
	loader          *TpReader
	loaderConfigDIR string
	loaderCfgPath   string

	tpCsvScenario = flag.String("tp_scenario", "testtp", "Use this scenario folder to import tp csv data from")

	loaderTests = []func(t *testing.T){
		testLoaderITInitConfig,
		testLoaderITInitDataDB,
		testLoaderITRemoveLoad,
		testLoaderITLoadFromCSV,
		testLoaderITWriteToDatabase,
		testLoaderITInitDataDB,
		testLoaderITInitDataDB,
	}
)

func TestLoaderIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		loaderConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		loaderConfigDIR = "tutmysql"
	case utils.MetaMongo:
		loaderConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		loaderConfigDIR = "tutpostgres"
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range loaderTests {
		t.Run(loaderConfigDIR, stest)
	}
}

func testLoaderITInitConfig(t *testing.T) {
	loaderCfgPath = path.Join(*utils.DataDir, "conf", "samples", loaderConfigDIR)
	var err error
	lCfg, err = config.NewCGRConfigFromPath(context.Background(), loaderCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testLoaderITInitDataDB(t *testing.T) {
	var err error
	dbConn, err := NewDataDBConn(lCfg.DataDbCfg().Type,
		lCfg.DataDbCfg().Host, lCfg.DataDbCfg().Port, lCfg.DataDbCfg().Name,
		lCfg.DataDbCfg().User, lCfg.DataDbCfg().Password, lCfg.GeneralCfg().DBDataEncoding,
		lCfg.DataDbCfg().Opts, lCfg.DataDbCfg().Items)
	if err != nil {
		t.Fatal("Error on dataDb connection: ", err.Error())
	}
	dataDbCsv = NewDataManager(dbConn, lCfg.CacheCfg(), nil)
	if lCfg.DataDbCfg().Type == utils.MetaInternal {
		chIDs := []string{}
		for dbKey := range utils.CacheInstanceToPrefix { // clear only the DataDB
			chIDs = append(chIDs, dbKey)
		}
		Cache.Clear(chIDs)
	} else {
		if err = dbConn.Flush(utils.EmptyString); err != nil {
			t.Fatal("Error when flushing datadb")
		}
	}
	cacheChan := make(chan birpc.ClientConnector, 1)
	srv, _ := birpc.NewServiceWithMethodsRename(NewCacheS(lCfg, dataDbCsv, nil, nil), "", false, func(key string) (newKey string) {
		return strings.TrimPrefix(key, "V1")
	})
	cacheChan <- srv
	connMgr = NewConnManager(lCfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), utils.CacheSv1, cacheChan)
}

// Loads data from csv files in tp scenario to dataDbCsv
func testLoaderITRemoveLoad(t *testing.T) {
	var err error
	/*for fn, v := range FileValidators {
		if err = ValidateCSVData(path.Join(*utils.DataDir, "tariffplans", *tpCsvScenario, fn), v.Rule); err != nil {
			t.Error("Failed validating data: ", err.Error())
		}
	}*/
	csvStorage, err := NewFileCSVStorage(utils.CSVSep, path.Join(*utils.DataDir, "tariffplans", *tpCsvScenario))
	if err != nil {
		t.Fatal(err)
	}
	loader, err = NewTpReader(dataDbCsv.DataDB(), csvStorage, "", "",
		[]string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}, nil, false)
	if err != nil {
		t.Error(err)
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
	if err = loader.LoadRouteProfiles(); err != nil {
		t.Error("Failed loading Route profiles: ", err.Error())
	}
	if err = loader.LoadAttributeProfiles(); err != nil {
		t.Error("Failed loading Attribute profiles: ", err.Error())
	}
	if err = loader.LoadChargerProfiles(); err != nil {
		t.Error("Failed loading Charger profiles: ", err.Error())
	}
	if err := loader.WriteToDatabase(false, false); err != nil {
		t.Error("Could not write data into dataDb: ", err.Error())
	}
	if err := loader.RemoveFromDatabase(false, true); err != nil {
		t.Error("Could not remove data from dataDb: ", err.Error())
	}
}

// Loads data from csv files in tp scenario to dataDbCsv
func testLoaderITLoadFromCSV(t *testing.T) {
	var err error
	/*for fn, v := range FileValidators {
		if err = ValidateCSVData(path.Join(*utils.DataDir, "tariffplans", *tpCsvScenario, fn), v.Rule); err != nil {
			t.Error("Failed validating data: ", err.Error())
		}
	}*/
	csvStorage, err := NewFileCSVStorage(utils.CSVSep, path.Join(*utils.DataDir, "tariffplans", *tpCsvScenario))
	if err != nil {
		t.Fatal(err)
	}
	loader, err = NewTpReader(dataDbCsv.DataDB(), csvStorage, "", "",
		[]string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}, nil, false)
	if err != nil {
		t.Error(err)
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
	if err = loader.LoadRouteProfiles(); err != nil {
		t.Error("Failed loading Route profiles: ", err.Error())
	}
	if err = loader.LoadAttributeProfiles(); err != nil {
		t.Error("Failed loading Attribute profiles: ", err.Error())
	}
	if err = loader.LoadChargerProfiles(); err != nil {
		t.Error("Failed loading Charger profiles: ", err.Error())
	}
	if err := loader.WriteToDatabase(false, false); err != nil {
		t.Error("Could not write data into dataDb: ", err.Error())
	}
}

func testLoaderITWriteToDatabase(t *testing.T) {

	for tenantid, fltr := range loader.filters {
		rcv, err := loader.dm.GetFilter(context.TODO(), tenantid.Tenant, tenantid.ID, false, false, utils.NonTransactional)
		if err != nil {
			t.Error("Failed GetFilter: ", err.Error())
		}
		filter, err := APItoFilter(fltr, "UTC")
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(filter, rcv) {
			t.Errorf("Expecting: %v, received: %v", filter, rcv)
		}
	}

	for tenantid, rl := range loader.resProfiles {
		rcv, err := loader.dm.GetResourceProfile(context.TODO(), tenantid.Tenant, tenantid.ID, false, false, utils.NonTransactional)
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
		rcv, err := loader.dm.GetStatQueueProfile(context.TODO(), tenantid.Tenant, tenantid.ID, false, false, utils.NonTransactional)
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
		rcv, err := loader.dm.GetThresholdProfile(context.TODO(), tenatid.Tenant, tenatid.ID, false, false, utils.NonTransactional)
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

	for tenatid, th := range loader.routeProfiles {
		rcv, err := loader.dm.GetRouteProfile(context.Background(), tenatid.Tenant, tenatid.ID, false, false, utils.NonTransactional)
		if err != nil {
			t.Errorf("Failed GetRouteProfile, tenant: %s, id: %s,  error: %s ", th.Tenant, th.ID, err.Error())
		}
		sts, err := APItoRouteProfile(th, "UTC")
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(sts, rcv) {
			t.Errorf("Expecting: %v, received: %v", sts, rcv)
		}
	}

	for tenatid, attrPrf := range loader.attributeProfiles {
		rcv, err := loader.dm.GetAttributeProfile(context.TODO(), tenatid.Tenant, tenatid.ID, false, false, utils.NonTransactional)
		if err != nil {
			t.Errorf("Failed GetAttributeProfile, tenant: %s, id: %s,  error: %s ", attrPrf.Tenant, attrPrf.ID, err.Error())
		}
		sts, err := APItoAttributeProfile(attrPrf, "UTC")
		if err != nil {
			t.Error(err)
		}
		sts.Compile()
		rcv.Compile()
		if !reflect.DeepEqual(sts, rcv) {
			t.Errorf("Expecting: %v, received: %v", sts, rcv)
		}
	}

	for tenatid, cpp := range loader.chargerProfiles {
		rcv, err := loader.dm.GetChargerProfile(context.Background(), tenatid.Tenant, tenatid.ID, false, false, utils.NonTransactional)
		if err != nil {
			t.Errorf("Failed GetChargerProfile, tenant: %s, id: %s,  error: %s ", cpp.Tenant, cpp.ID, err.Error())
		}
		cp := APItoChargerProfile(cpp, "UTC")
		if !reflect.DeepEqual(cp, rcv) {
			t.Errorf("Expecting: %v, received: %v", cp, rcv)
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
			if key == utils.TasksKey || strings.HasPrefix(key, utils.ActionPlanPrefix) { // action plans are not consistent
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
