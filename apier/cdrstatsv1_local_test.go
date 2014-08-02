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

package apier

import (
	"fmt"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os/exec"
	"path"
	"reflect"
	"testing"
	"time"
)

var cdrstCfgPath string
var cdrstCfg *config.CGRConfig
var cdrstRpc *rpc.Client

func init() {
	cdrstCfgPath = path.Join(*dataDir, "conf", "samples", "cdrstatsv1_local_test.cfg")
	cdrstCfg, _ = config.NewCGRConfigFromFile(&cfgPath)
}

func TestCDRStatsLclInitDataDb(t *testing.T) {
	if !*testLocal {
		return
	}
	ratingDb, err := engine.ConfigureRatingStorage(cdrstCfg.RatingDBType, cdrstCfg.RatingDBHost, cdrstCfg.RatingDBPort, cdrstCfg.RatingDBName,
		cdrstCfg.RatingDBUser, cdrstCfg.RatingDBPass, cdrstCfg.DBDataEncoding)
	if err != nil {
		t.Fatal("Cannot connect to dataDb", err)
	}
	accountDb, err := engine.ConfigureAccountingStorage(cdrstCfg.AccountDBType, cdrstCfg.AccountDBHost, cdrstCfg.AccountDBPort, cdrstCfg.AccountDBName,
		cdrstCfg.AccountDBUser, cdrstCfg.AccountDBPass, cdrstCfg.DBDataEncoding)
	if err != nil {
		t.Fatal("Cannot connect to dataDb", err)
	}
	for _, db := range []engine.Storage{ratingDb, accountDb} {
		if err := db.Flush(); err != nil {
			t.Fatal("Cannot reset dataDb", err)
		}
	}
}

func TestCDRStatsLclStartEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	enginePath, err := exec.LookPath("cgr-engine")
	if err != nil {
		t.Fatal("Cannot find cgr-engine executable")
	}
	exec.Command("pkill", "cgr-engine").Run() // Just to make sure another one is not running, bit brutal maybe we can fine tune it
	engine := exec.Command(enginePath, "-config", cdrstCfgPath)
	if err := engine.Start(); err != nil {
		t.Fatal("Cannot start cgr-engine: ", err.Error())
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time to rater to fire up
}

// Connect rpc client to rater
func TestCDRStatsLclRpcConn(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	cdrstRpc, err = jsonrpc.Dial("tcp", cdrstCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func TestCDRStatsLclGetQueueIds(t *testing.T) {
	if !*testLocal {
		return
	}
	var queueIds []string
	eQueueIds := []string{"*default"}
	if err := cdrstRpc.Call("CDRStatsV1.GetQueueIds", "", &queueIds); err != nil {
		t.Error("Calling CDRStatsV1.GetQueueIds, got error: ", err.Error())
	} else if !reflect.DeepEqual(eQueueIds, queueIds) {
		t.Errorf("Expecting: %v, received: %v", eQueueIds, queueIds)
	}
}

func TestCDRStatsLclLoadTariffPlanFromFolder(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	// Simple test that command is executed without errors
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "cdrstats")}
	if err := cdrstRpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.LoadTariffPlanFromFolder: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.LoadTariffPlanFromFolder got reply: ", reply)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func TestCDRStatsLclGetQueueIds2(t *testing.T) {
	if !*testLocal {
		return
	}
	var queueIds []string
	eQueueIds := []string{"*default", "CDRST3", "CDRST4"}
	if err := cdrstRpc.Call("CDRStatsV1.GetQueueIds", "", &queueIds); err != nil {
		t.Error("Calling CDRStatsV1.GetQueueIds, got error: ", err.Error())
	} else if !reflect.DeepEqual(eQueueIds, queueIds) {
		t.Errorf("Expecting: %v, received: %v", eQueueIds, queueIds)
	}
}

func TestCDRStatsLclPostCdrs(t *testing.T) {
	if !*testLocal {
		return
	}
	httpClient := new(http.Client)
	storedCdrs := []*utils.StoredCdr{
		&utils.StoredCdr{CgrId: utils.Sha1("dsafdsafa", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: "test",
			ReqType: "rated", Direction: "*out", Tenant: "cgrates.org",
			Category: "call", Account: "1001", Subject: "1001", Destination: "+4986517174963", SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
			AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: utils.DEFAULT_RUNID,
			Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
			Cost: 1.01, RatedAccount: "dan", RatedSubject: "dan",
		},
		&utils.StoredCdr{CgrId: utils.Sha1("dsafdsafb", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: "test",
			ReqType: "rated", Direction: "*out", Tenant: "cgrates.org",
			Category: "call", Account: "1001", Subject: "1001", Destination: "+4986517174963", SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
			AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: utils.DEFAULT_RUNID,
			Usage: time.Duration(5) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01, RatedAccount: "dan", RatedSubject: "dan",
		},
		&utils.StoredCdr{CgrId: utils.Sha1("dsafdsafc", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: "test",
			ReqType: "rated", Direction: "*out", Tenant: "cgrates.org",
			Category: "call", Account: "1001", Subject: "1001", Destination: "+4986517174963", SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
			MediationRunId: utils.DEFAULT_RUNID,
			Usage:          time.Duration(30) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01, RatedAccount: "dan", RatedSubject: "dan",
		},
		&utils.StoredCdr{CgrId: utils.Sha1("dsafdsafd", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: "test",
			ReqType: "rated", Direction: "*out", Tenant: "cgrates.org",
			Category: "call", Account: "1001", Subject: "1001", Destination: "+4986517174963", SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Time{},
			MediationRunId: utils.DEFAULT_RUNID,
			Usage:          time.Duration(0) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01, RatedAccount: "dan", RatedSubject: "dan",
		},
	}
	for _, storedCdr := range storedCdrs {
		if _, err := httpClient.PostForm(fmt.Sprintf("http://%s/cgr", "127.0.0.1:2080"), storedCdr.AsHttpForm()); err != nil {
			t.Error(err.Error())
		}
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)

}

func TestCDRStatsLclGetMetrics1(t *testing.T) {
	if !*testLocal {
		return
	}
	var rcvMetrics1 map[string]float64
	expectedMetrics1 := map[string]float64{"ASR": 75, "ACD": 15, "ACC": 15}
	if err := cdrstRpc.Call("CDRStatsV1.GetMetrics", AttrGetMetrics{StatsQueueId: "*default"}, &rcvMetrics1); err != nil {
		t.Error("Calling CDRStatsV1.GetMetrics, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedMetrics1, rcvMetrics1) {
		t.Errorf("Expecting: %v, received: %v", expectedMetrics1, rcvMetrics1)
	}
	var rcvMetrics2 map[string]float64
	expectedMetrics2 := map[string]float64{"ASR": 75, "ACD": 15}
	if err := cdrstRpc.Call("CDRStatsV1.GetMetrics", AttrGetMetrics{StatsQueueId: "CDRST4"}, &rcvMetrics2); err != nil {
		t.Error("Calling CDRStatsV1.GetMetrics, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedMetrics2, rcvMetrics2) {
		t.Errorf("Expecting: %v, received: %v", expectedMetrics2, rcvMetrics2)
	}
}

func TestCDRStatsLclResetMetrics(t *testing.T) {
	if !*testLocal {
		return
	}
	var reply string
	if err := cdrstRpc.Call("CDRStatsV1.ResetQueues", utils.AttrCDRStatsReloadQueues{StatsQueueIds: []string{"CDRST4"}}, &reply); err != nil {
		t.Error("Calling CDRStatsV1.ResetQueues, got error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	var rcvMetrics1 map[string]float64
	expectedMetrics1 := map[string]float64{"ASR": 75, "ACD": 15, "ACC": 15}
	if err := cdrstRpc.Call("CDRStatsV1.GetMetrics", AttrGetMetrics{StatsQueueId: "*default"}, &rcvMetrics1); err != nil {
		t.Error("Calling CDRStatsV1.GetMetrics, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedMetrics1, rcvMetrics1) {
		t.Errorf("Expecting: %v, received: %v", expectedMetrics1, rcvMetrics1)
	}
	var rcvMetrics2 map[string]float64
	expectedMetrics2 := map[string]float64{"ASR": 0, "ACD": 0}
	if err := cdrstRpc.Call("CDRStatsV1.GetMetrics", AttrGetMetrics{StatsQueueId: "CDRST4"}, &rcvMetrics2); err != nil {
		t.Error("Calling CDRStatsV1.GetMetrics, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedMetrics2, rcvMetrics2) {
		t.Errorf("Expecting: %v, received: %v", expectedMetrics2, rcvMetrics2)
	}
}
