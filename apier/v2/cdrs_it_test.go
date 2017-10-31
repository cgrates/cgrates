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
package v2

import (
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var cdrsCfgPath string
var cdrsCfg *config.CGRConfig
var cdrsRpc *rpc.Client
var cdrsConfDIR string // run the tests for specific configuration

// subtests to be executed for each confDIR
var sTestsCDRsIT = []func(t *testing.T){
	testV2CDRsInitConfig,
	testV2CDRsInitDataDb,
	testV2CDRsInitCdrDb,
	testV2CDRsInjectUnratedCdr,
	testV2CDRsStartEngine,
	testV2CDRsRpcConn,
	testV2CDRsProcessCdrRated,
	testV2CDRsProcessCdrRaw,
	testV2CDRsGetCdrs,
	testV2CDRsCountCdrs,
	testV2CDRsProcessPrepaidCdr,
	testV2CDRsRateWithoutTP,
	testV2CDRsLoadTariffPlanFromFolder,
	testV2CDRsRateWithTP,
	// ToDo: test engine shutdown
}

// Tests starting here
func TestCDRsITMySQL(t *testing.T) {
	cdrsConfDIR = "cdrsv2mysql"
	for _, stest := range sTestsCDRsIT {
		t.Run(cdrsConfDIR, stest)
	}
}

func TestCDRsITpg(t *testing.T) {
	cdrsConfDIR = "cdrsv2psql"
	for _, stest := range sTestsCDRsIT {
		t.Run(cdrsConfDIR, stest)
	}
}

func TestCDRsITMongo(t *testing.T) {
	cdrsConfDIR = "cdrsv2mongo"
	for _, stest := range sTestsCDRsIT {
		t.Run(cdrsConfDIR, stest)
	}
}

func testV2CDRsInitConfig(t *testing.T) {
	var err error
	cdrsCfgPath = path.Join(*dataDir, "conf", "samples", cdrsConfDIR)
	if cdrsCfg, err = config.NewCGRConfigFromFolder(cdrsCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

func testV2CDRsInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(cdrsCfg); err != nil {
		t.Fatal(err)
	}
}

// InitDb so we can rely on count
func testV2CDRsInitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(cdrsCfg); err != nil {
		t.Fatal(err)
	}
}

func testV2CDRsInjectUnratedCdr(t *testing.T) {
	var db engine.CdrStorage
	switch cdrsConfDIR {
	case "cdrsv2mysql":
		db, err = engine.NewMySQLStorage(cdrsCfg.StorDBHost, cdrsCfg.StorDBPort, cdrsCfg.StorDBName, cdrsCfg.StorDBUser, cdrsCfg.StorDBPass,
			cdrsCfg.StorDBMaxOpenConns, cdrsCfg.StorDBMaxIdleConns, cdrsCfg.StorDBConnMaxLifetime)
	case "cdrsv2psql":
		db, err = engine.NewPostgresStorage(cdrsCfg.StorDBHost, cdrsCfg.StorDBPort, cdrsCfg.StorDBName, cdrsCfg.StorDBUser, cdrsCfg.StorDBPass,
			cdrsCfg.StorDBMaxOpenConns, cdrsCfg.StorDBMaxIdleConns, cdrsCfg.StorDBConnMaxLifetime)
	case "cdrsv2mongo":
		db, err = engine.NewMongoStorage(cdrsCfg.StorDBHost, cdrsCfg.StorDBPort, cdrsCfg.StorDBName,
			cdrsCfg.StorDBUser, cdrsCfg.StorDBPass, utils.StorDB, cdrsCfg.StorDBCDRSIndexes, nil, 10)
	}
	if err != nil {
		t.Error("Error on opening database connection: ", err)
		return
	}
	strCdr1 := &engine.CDR{CGRID: utils.Sha1("bbb1", time.Date(2015, 11, 21, 10, 47, 24, 0, time.UTC).String()), RunID: utils.MetaRaw,
		ToR: utils.VOICE, OriginID: "bbb1", OriginHost: "192.168.1.1", Source: "TestV2CdrsMongoInjectUnratedCdr", RequestType: utils.META_RATED,
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2015, 11, 21, 10, 47, 24, 0, time.UTC), AnswerTime: time.Date(2015, 11, 21, 10, 47, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost: -1}
	if err := db.SetCDR(strCdr1, false); err != nil {
		t.Error(err.Error())
	}
}

func testV2CDRsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(cdrsCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testV2CDRsRpcConn(t *testing.T) {
	cdrsRpc, err = jsonrpc.Dial("tcp", cdrsCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV2CDRsProcessCdrRated(t *testing.T) {
	cdr := &engine.CDR{
		CGRID: utils.Sha1("dsafdsaf", time.Date(2015, 12, 13, 18, 15, 26, 0, time.UTC).String()), RunID: utils.DEFAULT_RUNID,
		OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf",
		OriginHost: "192.168.1.1", Source: "TestV2CdrsMongoProcessCdrRated", RequestType: utils.META_RATED, Tenant: "cgrates.org", Category: "call",
		Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2015, 12, 13, 18, 15, 26, 0, time.UTC), AnswerTime: time.Date(2015, 12, 13, 18, 15, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost: 1.01, CostSource: "TestV2CdrsMongoProcessCdrRated", Rated: true,
	}
	var reply string
	if err := cdrsRpc.Call("CdrsV2.ProcessCdr", cdr, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

func testV2CDRsProcessCdrRaw(t *testing.T) {
	cdr := &engine.CDR{
		CGRID: utils.Sha1("abcdeftg", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, RunID: utils.MetaRaw,
		ToR: utils.VOICE, OriginID: "abcdeftg",
		OriginHost: "192.168.1.1", Source: "TestV2CdrsMongoProcessCdrRaw", RequestType: utils.META_RATED, Tenant: "cgrates.org", Category: "call",
		Account: "1002", Subject: "1002", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	var reply string
	if err := cdrsRpc.Call("CdrsV2.ProcessCdr", cdr, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
}

func testV2CDRsGetCdrs(t *testing.T) {
	var reply []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{}
	if err := cdrsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 4 { // 1 injected, 1 rated, 1 *raw and it's pair in *default run
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
	// CDRs with rating errors
	req = utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}, MinCost: utils.Float64Pointer(-1.0), MaxCost: utils.Float64Pointer(0.0)}
	if err := cdrsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", reply)
	}
	// CDRs Rated
	req = utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}}
	if err := cdrsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", reply)
	}
	// Raw CDRs
	req = utils.RPCCDRsFilter{RunIDs: []string{utils.MetaRaw}}
	if err := cdrsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", reply)
	}
	// Skip Errors
	req = utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}, MinCost: utils.Float64Pointer(0.0), MaxCost: utils.Float64Pointer(-1.0)}
	if err := cdrsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", reply)
	}
}

func testV2CDRsCountCdrs(t *testing.T) {
	var reply int64
	req := utils.AttrGetCdrs{}
	if err := cdrsRpc.Call("ApierV2.CountCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != 4 {
		t.Error("Unexpected number of CDRs returned: ", reply)
	}
}

// Make sure *prepaid does not block until finding previous costs
func testV2CDRsProcessPrepaidCdr(t *testing.T) {
	var reply string
	cdrs := []*engine.CDR{
		&engine.CDR{CGRID: utils.Sha1("dsafdsaf2", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf",
			OriginHost: "192.168.1.1", Source: "TestV2CdrsMongoProcessPrepaidCdr1", RequestType: utils.META_PREPAID, Tenant: "cgrates.org",
			Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
			SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
			Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01, Rated: true,
		},
		&engine.CDR{CGRID: utils.Sha1("abcdeftg2", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf",
			OriginHost: "192.168.1.1", Source: "TestV2CdrsMongoProcessPrepaidCdr2", RequestType: utils.META_PREPAID, Tenant: "cgrates.org",
			Category: "call", Account: "1002", Subject: "1002", Destination: "1002",
			SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
			Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
		&engine.CDR{CGRID: utils.Sha1("aererfddf2", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf",
			OriginHost: "192.168.1.1", Source: "TestV2CdrsMongoProcessPrepaidCdr3", RequestType: utils.META_PREPAID, Tenant: "cgrates.org",
			Category: "call", Account: "1003", Subject: "1003", Destination: "1002",
			SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
			Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
	}
	tStart := time.Now()
	for _, cdr := range cdrs {
		if err := cdrsRpc.Call("CdrsV2.ProcessCdr", cdr, &reply); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if reply != utils.OK {
			t.Error("Unexpected reply received: ", reply)
		}
	}
	if processDur := time.Now().Sub(tStart); processDur > 1*time.Second {
		t.Error("Unexpected processing time", processDur)
	}
}

func testV2CDRsRateWithoutTP(t *testing.T) {
	rawCdrCGRID := utils.Sha1("bbb1", time.Date(2015, 11, 21, 10, 47, 24, 0, time.UTC).String())
	// Rate the injected CDR, should not rate it since we have no TP loaded
	attrs := utils.AttrRateCdrs{CgrIds: []string{rawCdrCGRID}}
	var reply string
	if err := cdrsRpc.Call("CdrsV2.RateCdrs", attrs, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{CGRIDs: []string{rawCdrCGRID}, RunIDs: []string{utils.META_DEFAULT}}
	if err := cdrsRpc.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 { // Injected CDR did not have a charging run
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != -1 {
			t.Errorf("Unexpected CDR returned: %+v", cdrs[0])
		}
	}
}

func testV2CDRsLoadTariffPlanFromFolder(t *testing.T) {
	var loadInst utils.LoadInstance
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := cdrsRpc.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func testV2CDRsRateWithTP(t *testing.T) {
	rawCdrCGRID := utils.Sha1("bbb1", time.Date(2015, 11, 21, 10, 47, 24, 0, time.UTC).String())
	attrs := utils.AttrRateCdrs{CgrIds: []string{rawCdrCGRID}}
	var reply string
	if err := cdrsRpc.Call("CdrsV2.RateCdrs", attrs, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{CGRIDs: []string{rawCdrCGRID}, RunIDs: []string{utils.META_DEFAULT}}
	if err := cdrsRpc.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.3 {
			t.Errorf("Unexpected CDR returned: %+v", cdrs[0])
		}
	}
}
