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
	"flag"
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var testLocal = flag.Bool("local", false, "Perform the tests only on local test environment, not by default.") // This flag will be passed here via "go test -local" args
var dataDir = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
var waitRater = flag.Int("wait_rater", 500, "Number of miliseconds to wait for rater to start and cache")

var cdrsCfgPath string
var cdrsCfg *config.CGRConfig
var cdrsRpc *rpc.Client

func TestV2CdrsMysqlInitConfig(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	cdrsCfgPath = path.Join(*dataDir, "conf", "samples", "cdrsv2mysql")
	if cdrsCfg, err = config.NewCGRConfigFromFolder(cdrsCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

func TestV2CdrsMysqlInitDataDb(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.InitDataDb(cdrsCfg); err != nil {
		t.Fatal(err)
	}
}

// InitDb so we can rely on count
func TestV2CdrsMysqlInitCdrDb(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.InitStorDb(cdrsCfg); err != nil {
		t.Fatal(err)
	}
}

func TestV2CdrsMysqlInjectUnratedCdr(t *testing.T) {
	if !*testLocal {
		return
	}
	mysqlDb, err := engine.NewMySQLStorage(cdrsCfg.StorDBHost, cdrsCfg.StorDBPort, cdrsCfg.StorDBName, cdrsCfg.StorDBUser, cdrsCfg.StorDBPass,
		cdrsCfg.StorDBMaxOpenConns, cdrsCfg.StorDBMaxIdleConns)
	if err != nil {
		t.Error("Error on opening database connection: ", err)
		return
	}
	strCdr1 := &engine.CDR{CGRID: utils.Sha1("bbb1", time.Date(2015, 11, 21, 10, 47, 24, 0, time.UTC).String()),
		TOR: utils.VOICE, OriginID: "bbb1", OriginHost: "192.168.1.1", Source: "UNKNOWN", ReqType: utils.META_RATED,
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2015, 11, 21, 10, 47, 24, 0, time.UTC), AnswerTime: time.Date(2015, 11, 21, 10, 47, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		RunID: utils.DEFAULT_RUNID, Cost: 1.201}
	if err := mysqlDb.SetCdr(strCdr1); err != nil {
		t.Error(err.Error())
	}
}

func TestV2CdrsMysqlStartEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	if _, err := engine.StopStartEngine(cdrsCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestV2CdrsMysqlRpcConn(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	cdrsRpc, err = jsonrpc.Dial("tcp", cdrsCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

// Insert some CDRs
func TestV2CdrsMysqlProcessCdr(t *testing.T) {
	if !*testLocal {
		return
	}
	var reply string
	cdrs := []*engine.CDR{
		&engine.CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, TOR: utils.VOICE, OriginID: "dsafdsaf",
			OriginHost: "192.168.1.1", Source: "test", ReqType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
			SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
			Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01, Rated: true,
		},
		&engine.CDR{CGRID: utils.Sha1("abcdeftg", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, TOR: utils.VOICE, OriginID: "dsafdsaf",
			OriginHost: "192.168.1.1", Source: "test", ReqType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1002", Subject: "1002", Destination: "1002",
			SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
			Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
		&engine.CDR{CGRID: utils.Sha1("aererfddf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, TOR: utils.VOICE, OriginID: "dsafdsaf",
			OriginHost: "192.168.1.1", Source: "test", ReqType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1003", Subject: "1003", Destination: "1002",
			SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
			Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
	}
	for _, cdr := range cdrs {
		if err := cdrsRpc.Call("CdrsV2.ProcessCdr", cdr, &reply); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if reply != utils.OK {
			t.Error("Unexpected reply received: ", reply)
		}
	}
}

func TestV2CdrsMysqlGetCdrs(t *testing.T) {
	if !*testLocal {
		return
	}
	var reply []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{}
	if err := cdrsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 4 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
	// CDRs with errors
	req = utils.RPCCDRsFilter{MinCost: utils.Float64Pointer(-1.0), MaxCost: utils.Float64Pointer(0.0)}
	if err := cdrsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", reply)
	}
	// CDRs Rated
	req = utils.RPCCDRsFilter{MinCost: utils.Float64Pointer(-1.0)}
	if err := cdrsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 3 {
		t.Error("Unexpected number of CDRs returned: ", reply)
	}
	// CDRs non rated OR SkipRated
	req = utils.RPCCDRsFilter{MaxCost: utils.Float64Pointer(-1.0)}
	if err := cdrsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", reply)
	}
	// Skip Errors
	req = utils.RPCCDRsFilter{MinCost: utils.Float64Pointer(0.0), MaxCost: utils.Float64Pointer(-1.0)}
	if err := cdrsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", reply)
	}
}

func TestV2CdrsMysqlCountCdrs(t *testing.T) {
	if !*testLocal {
		return
	}
	var reply int64
	req := utils.AttrGetCdrs{}
	if err := cdrsRpc.Call("ApierV2.CountCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != 4 {
		t.Error("Unexpected number of CDRs returned: ", reply)
	}
}

// Test Prepaid CDRs without previous costs being calculated
func TestV2CdrsMysqlProcessPrepaidCdr(t *testing.T) {
	if !*testLocal {
		return
	}
	var reply string
	cdrs := []*engine.CDR{
		&engine.CDR{CGRID: utils.Sha1("dsafdsaf2", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, TOR: utils.VOICE, OriginID: "dsafdsaf",
			OriginHost: "192.168.1.1", Source: "test", ReqType: utils.META_PREPAID, Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
			SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
			Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01, Rated: true,
		},
		&engine.CDR{CGRID: utils.Sha1("abcdeftg2", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, TOR: utils.VOICE, OriginID: "dsafdsaf",
			OriginHost: "192.168.1.1", Source: "test", ReqType: utils.META_PREPAID, Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1002", Subject: "1002", Destination: "1002",
			SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
			Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
		&engine.CDR{CGRID: utils.Sha1("aererfddf2", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, TOR: utils.VOICE, OriginID: "dsafdsaf",
			OriginHost: "192.168.1.1", Source: "test", ReqType: utils.META_PREPAID, Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1003", Subject: "1003", Destination: "1002",
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

func TestV2CdrsMysqlRateWithoutTP(t *testing.T) {
	if !*testLocal {
		return
	}
	rawCdrCGRID := utils.Sha1("bbb1", time.Date(2015, 11, 21, 10, 47, 24, 0, time.UTC).String())
	// Rate the injected CDR, should not rate it since we have no TP loaded
	attrs := utils.AttrRateCdrs{CgrIds: []string{rawCdrCGRID}}
	var reply string
	if err := cdrsRpc.Call("CdrsV2.RateCdrs", attrs, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{CGRIDs: []string{rawCdrCGRID}}
	if err := cdrsRpc.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	} else {
		if cdrs[0].Cost != -1 {
			t.Errorf("Unexpected CDR returned: %+v", cdrs[0])
		}
	}
}

func TestV2CdrsMysqloadTariffPlanFromFolder(t *testing.T) {
	if !*testLocal {
		return
	}
	var loadInst engine.LoadInstance
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := cdrsRpc.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &loadInst); err != nil {
		t.Error(err)
	} else if loadInst.LoadId == "" {
		t.Error("Empty loadId received, loadInstance: ", loadInst)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func TestV2CdrsMysqlRateWithTP(t *testing.T) {
	if !*testLocal {
		return
	}
	rawCdrCGRID := utils.Sha1("bbb1", time.Date(2015, 11, 21, 10, 47, 24, 0, time.UTC).String())
	attrs := utils.AttrRateCdrs{CgrIds: []string{rawCdrCGRID}}
	var reply string
	if err := cdrsRpc.Call("CdrsV2.RateCdrs", attrs, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{CGRIDs: []string{rawCdrCGRID}}
	if err := cdrsRpc.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	} else {
		if cdrs[0].Cost != 0.3 {
			t.Errorf("Unexpected CDR returned: %+v", cdrs[0])
		}
	}
}

/*
// Benchmark speed of processing 1000 CDRs
func TestV2CdrsMysqlProcessRatedExternalCdrBenchmark(t *testing.T) {
	if !*testLocal {
		return
	}
	cdr := &engine.ExternalCDR{TOR: utils.VOICE,
		OriginID: "benchratedcdr", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, ReqType: utils.META_RATED, Direction: utils.OUT,
		Tenant: "cgrates.org", Category: "call", Account: "1003", Subject: "1003", Destination: "1001", Supplier: "SUPPL1",
		SetupTime: "2014-08-04T13:00:00Z", AnswerTime: "2014-08-04T13:00:07Z",
		Usage: "15", Pdd: "7.0", ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	var reply string
	tStart := time.Now()
	nrCdrs := 1000
	for i := 0; i < nrCdrs; i++ {
		cdr.OriginID = "benchratedcdr" + strconv.Itoa(i)
		if err := cdrsRpc.Call("CdrsV2.ProcessExternalCdr", cdr, &reply); err != nil {
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
func TestV2CdrsMysqlReRateWithTPBenchmark(t *testing.T) {
	if !*testLocal {
		return
	}
	var nrCdrs int64
	req := utils.AttrRateCdrs{RerateRated: true, RerateErrors: true}
	if err := cdrsRpc.Call("ApierV2.CountCdrs", req, &nrCdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	}
	tStart := time.Now()
	var reply string
	if err := cdrsRpc.Call("CdrsV2.RateCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	if durExec := time.Now().Sub(tStart); durExec > time.Duration(1)*time.Second {
		t.Errorf("Rerating of %d rated CDRs took: %v", nrCdrs, durExec)
	}
}

// Benchmark speed of processing 1000 postpaid CDRs
func TestV2CdrsMysqlProcessPostpaidExternalCdrBenchmark(t *testing.T) {
	if !*testLocal {
		return
	}
	cdr := &engine.ExternalCDR{TOR: utils.VOICE,
		OriginID: "benchpostpaidcdr", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, ReqType: utils.META_POSTPAID, Direction: utils.OUT,
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002", Supplier: "SUPPL1",
		SetupTime: "2014-08-04T13:00:00Z", AnswerTime: "2014-08-04T13:00:07Z",
		Usage: "15", Pdd: "7.0", ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	var reply string
	tStart := time.Now()
	nrCdrs := 1000
	for i := 0; i < nrCdrs; i++ {
		cdr.OriginID = "benchpostpaidcdr" + strconv.Itoa(i)
		if err := cdrsRpc.Call("CdrsV2.ProcessExternalCdr", cdr, &reply); err != nil {
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

func TestV2CdrsMysqlKillEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
