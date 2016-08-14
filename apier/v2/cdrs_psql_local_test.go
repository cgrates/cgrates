/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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

package v2

import (
	"net/rpc"
	"net/rpc/jsonrpc"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var cdrsPsqlCfgPath string
var cdrsPsqlCfg *config.CGRConfig
var cdrsPsqlRpc *rpc.Client
var cmdEngineCdrPsql *exec.Cmd

func TestV2CDRsPSQLInitConfig(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	cdrsPsqlCfgPath = path.Join(*dataDir, "conf", "samples", "cdrsv2psql")
	if cdrsPsqlCfg, err = config.NewCGRConfigFromFolder(cdrsPsqlCfgPath); err != nil {
		t.Fatal(err)
	}
}

func TestV2CDRsPSQLInitDataDb(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.InitDataDb(cdrsPsqlCfg); err != nil {
		t.Fatal(err)
	}
}

// InitDb so we can rely on count
func TestV2CDRsPSQLInitCdrDb(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.InitStorDb(cdrsPsqlCfg); err != nil {
		t.Fatal(err)
	}
}

func TestV2CDRsPSQLInjectUnratedCdr(t *testing.T) {
	if !*testLocal {
		return
	}
	psqlDb, err := engine.NewPostgresStorage(cdrsPsqlCfg.StorDBHost, cdrsPsqlCfg.StorDBPort, cdrsPsqlCfg.StorDBName, cdrsPsqlCfg.StorDBUser, cdrsPsqlCfg.StorDBPass,
		cdrsPsqlCfg.StorDBMaxOpenConns, cdrsPsqlCfg.StorDBMaxIdleConns)
	if err != nil {
		t.Error("Error on opening database connection: ", err)
		return
	}
	strCdr1 := &engine.CDR{CGRID: utils.Sha1("bbb1", time.Date(2015, 11, 21, 10, 47, 24, 0, time.UTC).String()), RunID: utils.MetaRaw,
		ToR: utils.VOICE, OriginID: "bbb1", OriginHost: "192.168.1.1", Source: "TestV2CDRsPSQLInjectUnratedCdr", RequestType: utils.META_RATED,
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2015, 11, 21, 10, 47, 24, 0, time.UTC), AnswerTime: time.Date(2015, 11, 21, 10, 47, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost: -1}
	if err := psqlDb.SetCDR(strCdr1, false); err != nil {
		t.Error(err.Error())
	}
}

func TestV2CDRsPSQLStartEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	if _, err := engine.StopStartEngine(cdrsPsqlCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestV2CDRsPSQLRpcConn(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	cdrsPsqlRpc, err = jsonrpc.Dial("tcp", cdrsPsqlCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func TestV2CDRsPSQLProcessCdrRated(t *testing.T) {
	if !*testLocal {
		return
	}
	cdr := &engine.CDR{
		CGRID: utils.Sha1("dsafdsaf", time.Date(2015, 12, 13, 18, 15, 26, 0, time.UTC).String()), RunID: utils.DEFAULT_RUNID,
		OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf",
		OriginHost: "192.168.1.1", Source: "TestV2CDRsPSQLProcessCdrRated", RequestType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org", Category: "call",
		Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2015, 12, 13, 18, 15, 26, 0, time.UTC), AnswerTime: time.Date(2015, 12, 13, 18, 15, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost: 1.01, CostSource: "TestV2CDRsPSQLProcessCdrRated", Rated: true,
	}
	var reply string
	if err := cdrsPsqlRpc.Call("CdrsV2.ProcessCDR", cdr, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

func TestV2CDRsPSQLProcessCdrRaw(t *testing.T) {
	if !*testLocal {
		return
	}
	cdr := &engine.CDR{
		CGRID: utils.Sha1("abcdeftg", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, RunID: utils.MetaRaw,
		ToR: utils.VOICE, OriginID: "abcdeftg",
		OriginHost: "192.168.1.1", Source: "TestV2CDRsPSQLProcessCdrRaw", RequestType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org", Category: "call",
		Account: "1002", Subject: "1002", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	var reply string
	if err := cdrsPsqlRpc.Call("CdrsV2.ProcessCDR", cdr, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
}

func TestV2CDRsPSQLGetCdrs(t *testing.T) {
	if !*testLocal {
		return
	}
	var reply []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{}
	if err := cdrsPsqlRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 4 { // 1 injected, 1 rated, 1 *raw and it's pair in *default run
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
	// CDRs with rating errors
	req = utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}, MinCost: utils.Float64Pointer(-1.0), MaxCost: utils.Float64Pointer(0.0)}
	if err := cdrsPsqlRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", reply)
	}
	// CDRs Rated
	req = utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}}
	if err := cdrsPsqlRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", reply)
	}
	// Raw CDRs
	req = utils.RPCCDRsFilter{RunIDs: []string{utils.MetaRaw}}
	if err := cdrsPsqlRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", reply)
	}
	// Skip Errors
	req = utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}, MinCost: utils.Float64Pointer(0.0), MaxCost: utils.Float64Pointer(-1.0)}
	if err := cdrsPsqlRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", reply)
	}
}

func TestV2CDRsPSQLCountCdrs(t *testing.T) {
	if !*testLocal {
		return
	}
	var reply int64
	req := utils.AttrGetCdrs{}
	if err := cdrsPsqlRpc.Call("ApierV2.CountCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != 4 {
		t.Error("Unexpected number of CDRs returned: ", reply)
	}
}

// Make sure *prepaid does not block until finding previous costs
func TestV2CDRsPSQLProcessPrepaidCdr(t *testing.T) {
	if !*testLocal {
		return
	}
	var reply string
	cdrs := []*engine.CDR{
		&engine.CDR{CGRID: utils.Sha1("dsafdsaf2", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf",
			OriginHost: "192.168.1.1", Source: "TestV2CDRsPSQLProcessPrepaidCdr1", RequestType: utils.META_PREPAID, Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
			SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
			Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01, Rated: true,
		},
		&engine.CDR{CGRID: utils.Sha1("abcdeftg2", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf",
			OriginHost: "192.168.1.1", Source: "TestV2CDRsPSQLProcessPrepaidCdr2", RequestType: utils.META_PREPAID, Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1002", Subject: "1002", Destination: "1002",
			SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
			Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
		&engine.CDR{CGRID: utils.Sha1("aererfddf2", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf",
			OriginHost: "192.168.1.1", Source: "TestV2CDRsPSQLProcessPrepaidCdr3", RequestType: utils.META_PREPAID, Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1003", Subject: "1003", Destination: "1002",
			SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
			Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
	}
	tStart := time.Now()
	for _, cdr := range cdrs {
		if err := cdrsPsqlRpc.Call("CdrsV2.ProcessCDR", cdr, &reply); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if reply != utils.OK {
			t.Error("Unexpected reply received: ", reply)
		}
	}
	if processDur := time.Now().Sub(tStart); processDur > 1*time.Second {
		t.Error("Unexpected processing time", processDur)
	}
}

func TestV2CDRsPSQLRateWithoutTP(t *testing.T) {
	if !*testLocal {
		return
	}
	rawCdrCGRID := utils.Sha1("bbb1", time.Date(2015, 11, 21, 10, 47, 24, 0, time.UTC).String())
	// Rate the injected CDR, should not rate it since we have no TP loaded
	attrs := utils.AttrRateCdrs{CgrIds: []string{rawCdrCGRID}}
	var reply string
	if err := cdrsPsqlRpc.Call("CdrsV2.RateCdrs", attrs, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{CGRIDs: []string{rawCdrCGRID}, RunIDs: []string{utils.META_DEFAULT}}
	if err := cdrsPsqlRpc.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 { // Injected CDR did not have a charging run
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != -1 {
			t.Errorf("Unexpected CDR returned: %+v", cdrs[0])
		}
	}
}

func TestV2CDRsPSQLLoadTariffPlanFromFolder(t *testing.T) {
	if !*testLocal {
		return
	}
	var loadInst utils.LoadInstance
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := cdrsPsqlRpc.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &loadInst); err != nil {
		t.Error(err)
	} else if loadInst.RatingLoadID == "" || loadInst.AccountingLoadID == "" {
		// ReThink load instance
		//t.Error("Empty loadId received, loadInstance: ", loadInst)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func TestV2CDRsPSQLRateWithTP(t *testing.T) {
	if !*testLocal {
		return
	}
	rawCdrCGRID := utils.Sha1("bbb1", time.Date(2015, 11, 21, 10, 47, 24, 0, time.UTC).String())
	attrs := utils.AttrRateCdrs{CgrIds: []string{rawCdrCGRID}}
	var reply string
	if err := cdrsPsqlRpc.Call("CdrsV2.RateCdrs", attrs, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{CGRIDs: []string{rawCdrCGRID}, RunIDs: []string{utils.META_DEFAULT}}
	if err := cdrsPsqlRpc.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.3 {
			t.Errorf("Unexpected CDR returned: %+v", cdrs[0])
		}
	}
}

/*
// Benchmark speed of processing 1000 CDRs
func TestV2CDRsPSQLProcessRatedExternalCdrBenchmark(t *testing.T) {
	if !*testLocal {
		return
	}
	cdr := &engine.ExternalCDR{ToR: utils.VOICE,
		OriginID: "benchratedcdr", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED, Direction: utils.OUT,
		Tenant: "cgrates.org", Category: "call", Account: "1003", Subject: "1003", Destination: "1001", Supplier: "SUPPL1",
		SetupTime: "2014-08-04T13:00:00Z", AnswerTime: "2014-08-04T13:00:07Z",
		Usage: "15", PDD: "7.0", ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	var reply string
	tStart := time.Now()
	nrCdrs := 1000
	for i := 0; i < nrCdrs; i++ {
		cdr.OriginID = "benchratedcdr" + strconv.Itoa(i)
		if err := cdrsPsqlRpc.Call("CdrsV2.ProcessExternalCdr", cdr, &reply); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if reply != utils.OK {
			t.Error("Unexpected reply received: ", reply)
		}
	}
	if durExec := time.Now().Sub(tStart); durExec > time.Duration(1)*time.Second {
		t.Errorf("Processing of %d rated CDRs took: %v", nrCdrs, durExec)
	}
}

// Benchmark speed of re-rating 1000 CDRs
func TestV2CDRsPSQLReRateWithTPBenchmark(t *testing.T) {
	if !*testLocal {
		return
	}
	var nrCdrs int64
	req := utils.AttrRateCdrs{RerateRated: true, RerateErrors: true}
	if err := cdrsPsqlRpc.Call("ApierV2.CountCdrs", req, &nrCdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	}
	tStart := time.Now()
	var reply string
	if err := cdrsPsqlRpc.Call("CdrsV2.RateCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	if durExec := time.Now().Sub(tStart); durExec > time.Duration(1)*time.Second {
		t.Errorf("Rerating of %d rated CDRs took: %v", nrCdrs, durExec)
	}
}

// Benchmark speed of processing 1000 postpaid CDRs
func TestV2CDRsPSQLProcessPostpaidExternalCdrBenchmark(t *testing.T) {
	if !*testLocal {
		return
	}
	cdr := &engine.ExternalCDR{ToR: utils.VOICE,
		OriginID: "benchpostpaidcdr", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_POSTPAID, Direction: utils.OUT,
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002", Supplier: "SUPPL1",
		SetupTime: "2014-08-04T13:00:00Z", AnswerTime: "2014-08-04T13:00:07Z",
		Usage: "15", PDD: "7.0", ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	var reply string
	tStart := time.Now()
	nrCdrs := 1000
	for i := 0; i < nrCdrs; i++ {
		cdr.OriginID = "benchpostpaidcdr" + strconv.Itoa(i)
		if err := cdrsPsqlRpc.Call("CdrsV2.ProcessExternalCdr", cdr, &reply); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if reply != utils.OK {
			t.Error("Unexpected reply received: ", reply)
		}
	}
	if durExec := time.Now().Sub(tStart); durExec > time.Duration(1)*time.Second {
		t.Errorf("Processing of %d postpaid CDRs took: %v", nrCdrs, durExec)
	}
}
*/

func TestV2CDRsPSQLKillEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
