/*
Real-time Charging System for Telecom & ISP environments
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

/*
import (
	"flag"
	"fmt"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"net/url"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)



var cgrCfg *config.CGRConfig
var cgrRpc *rpc.Client
var cdrStor CdrStorage
var httpClient *http.Client

var storDbType = flag.String("stordb_type", utils.MYSQL, "The type of the storDb database <mysql>")
var startDelay = flag.Int("delay_start", 300, "Number of miliseconds to it for rater to start and cache")
var cfgPath = path.Join(*dataDir, "conf", "samples", "mediator1")

func TestMediInitConfig(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	cgrCfg, err = config.NewCGRConfigFromFolder(cfgPath)
	if err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

func TestMediInitDataDb(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := InitDataDb(cgrCfg); err != nil {
		t.Fatal(err)
	}
}

// Empty tables before using them
func TestMediInitStorDb(t *testing.T) {
	if !*testLocal {
		return
	}
	if *storDbType != utils.MYSQL {
		t.Fatal("Unsupported storDbType")
	}
	var mysql *MySQLStorage
	var err error
	if cdrStor, err = ConfigureCdrStorage(cgrCfg.StorDBType, cgrCfg.StorDBHost, cgrCfg.StorDBPort, cgrCfg.StorDBName, cgrCfg.StorDBUser, cgrCfg.StorDBPass,
		cgrCfg.StorDBMaxOpenConns, cgrCfg.StorDBMaxIdleConns); err != nil {
		t.Fatal("Error on opening database connection: ", err)
	} else {
		mysql = cdrStor.(*MySQLStorage)
	}
	if err := mysql.CreateTablesFromScript(path.Join(*dataDir, "storage", *storDbType, CREATE_CDRS_TABLES_SQL)); err != nil {
		t.Fatal("Error on mysql creation: ", err.Error())
		return // No point in going further
	}
	for _, tbl := range []string{utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_EXTRA} {
		if _, err := mysql.Db.Query(fmt.Sprintf("SELECT 1 from %s", tbl)); err != nil {
			t.Fatal(err.Error())
		}
	}
}

// Finds cgr-engine executable and starts it with default configuration
func TestMediStartEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	if _, err := StartEngine(cfgPath, *startDelay); err != nil {
		t.Fatal(err)
	}
	httpClient = new(http.Client)
}

// Connect rpc client
func TestMediRpcConn(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	//cgrRpc, err = rpc.Dial("tcp", cfg.RPCGOBListen) //ToDo: Fix with automatic config
	cgrRpc, err = jsonrpc.Dial("tcp", cgrCfg.RPCJSONListen)
	if err != nil {
		t.Fatal("Could not connect to CGR JSON-RPC Server: ", err.Error())
	}
}

func TestMediPostCdrs(t *testing.T) {
	if !*testLocal {
		return
	}
	cdrForm1 := url.Values{utils.TOR: []string{utils.VOICE}, utils.ACCID: []string{"dsafdsaf"}, utils.CDRHOST: []string{"192.168.1.1"}, utils.REQTYPE: []string{utils.META_RATED}, utils.DIRECTION: []string{"*out"},
		utils.TENANT: []string{"cgrates.org"}, utils.CATEGORY: []string{"call"}, utils.ACCOUNT: []string{"2001"}, utils.SUBJECT: []string{"2001"},
		utils.DESTINATION: []string{"+4986517174963"},
		utils.ANSWER_TIME: []string{"2014-11-07T08:42:26Z"}, utils.USAGE: []string{"10"}, "field_extr1": []string{"val_extr1"}, "fieldextr2": []string{"valextr2"}}
	cdrForm2 := url.Values{utils.TOR: []string{utils.VOICE}, utils.ACCID: []string{"adsafdsaf"}, utils.CDRHOST: []string{"192.168.1.1"}, utils.REQTYPE: []string{utils.META_RATED}, utils.DIRECTION: []string{"*out"},
		utils.TENANT: []string{"itsyscom.com"}, utils.CATEGORY: []string{"call"}, utils.ACCOUNT: []string{"1003"}, utils.SUBJECT: []string{"1003"}, utils.DESTINATION: []string{"+4986517174964"},
		utils.ANSWER_TIME: []string{"2014-11-07T08:42:26Z"}, utils.USAGE: []string{"10"}, "field_extr1": []string{"val_extr1"}, "fieldextr2": []string{"valextr2"}}
	cdrFormData1 := url.Values{utils.TOR: []string{utils.DATA}, utils.ACCID: []string{"616350843"}, utils.CDRHOST: []string{"192.168.1.1"}, utils.REQTYPE: []string{utils.META_RATED},
		utils.DIRECTION: []string{"*out"}, utils.TENANT: []string{"cgrates.org"}, utils.CATEGORY: []string{"data"},
		utils.ACCOUNT: []string{"1010"}, utils.SUBJECT: []string{"1010"}, utils.ANSWER_TIME: []string{"2014-11-07T08:42:26Z"},
		utils.USAGE: []string{"10"}, "field_extr1": []string{"val_extr1"}, "fieldextr2": []string{"valextr2"}}
	for _, cdrForm := range []url.Values{cdrForm1, cdrForm2, cdrFormData1} {
		cdrForm.Set(utils.CDRSOURCE, TEST_SQL)
		if _, err := httpClient.PostForm(fmt.Sprintf("http://%s/cdr_post", cgrCfg.HTTPListen), cdrForm); err != nil {
			t.Error(err.Error())
		}
	}
	time.Sleep(100 * time.Millisecond) // Give time for CDRs to reach database
	if storedCdrs, _, err := cdrStor.GetStoredCdrs(new(utils.CdrsFilter)); err != nil {
		t.Error(err)
	} else if len(storedCdrs) != 3 { // Make sure CDRs made it into StorDb
		t.Error(fmt.Sprintf("Unexpected number of CDRs stored: %d", len(storedCdrs)))
	}
	if nonErrorCdrs, _, err := cdrStor.GetStoredCdrs(&utils.CdrsFilter{CostEnd: utils.Float64Pointer(-1.0)}); err != nil {
		t.Error(err)
	} else if len(nonErrorCdrs) != 0 {
		t.Error(fmt.Sprintf("Unexpected number of CDRs stored: %d", len(nonErrorCdrs)))
	}
}

// Directly inject CDRs into storDb
func TestMediInjectCdrs(t *testing.T) {
	if !*testLocal {
		return
	}
	cgrCdr1 := CgrCdr{utils.TOR: utils.VOICE, utils.ACCID: "aaaaadsafdsaf", "cdrsource": "TEST_INJECT", utils.CDRHOST: "192.168.1.1", utils.REQTYPE: utils.META_RATED, utils.DIRECTION: "*out",
		utils.TENANT: "cgrates.org", utils.CATEGORY: "call", utils.ACCOUNT: "dan", utils.SUBJECT: "dan", utils.DESTINATION: "+4986517174963",
		utils.ANSWER_TIME: "2014-11-07T08:42:26Z", utils.USAGE: "10"}
	cgrCdr2 := CgrCdr{utils.TOR: utils.VOICE, utils.ACCID: "baaaadsafdsaf", "cdrsource": "TEST_INJECT", utils.CDRHOST: "192.168.1.1", utils.REQTYPE: utils.META_RATED, utils.DIRECTION: "*out",
		utils.TENANT: "cgrates.org", utils.CATEGORY: "call", utils.ACCOUNT: "dan", utils.SUBJECT: "dan", utils.DESTINATION: "+4986517173964",
		utils.ANSWER_TIME: "2014-11-07T09:42:26Z", utils.USAGE: "20"}
	for _, cdr := range []CgrCdr{cgrCdr1, cgrCdr2} {
		if err := cdrStor.SetCdr(cdr.AsStoredCdr()); err != nil {
			t.Error(err)
		}
	}
	if storedCdrs, _, err := cdrStor.GetStoredCdrs(new(utils.CdrsFilter)); err != nil {
		t.Error(err)
	} else if len(storedCdrs) != 5 { // Make sure CDRs made it into StorDb
		t.Error(fmt.Sprintf("Unexpected number of CDRs stored: %d", len(storedCdrs)))
	}
	if nonRatedCdrs, _, err := cdrStor.GetStoredCdrs(&utils.CdrsFilter{CostEnd: utils.Float64Pointer(-1.0)}); err != nil {
		t.Error(err)
	} else if len(nonRatedCdrs) != 2 { // Just two of them should be non-rated
		t.Error(fmt.Sprintf("Unexpected number of CDRs non-rated: %d", len(nonRatedCdrs)))
	}
}

// Test here LoadTariffPlanFromFolder
func TestMediLoadTariffPlanFromFolder(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	// Simple test that command is executed without errors
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := cgrRpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.LoadTariffPlanFromFolder: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling ApierV1.LoadTariffPlanFromFolder got reply: ", reply)
	}
}

/*
func TestMediRateCdrs(t *testing.T) {
	if !*testLocal {
		return
	}
	var reply string
	if err := cgrRpc.Call("CdrsV1.RateCdrs", utils.AttrRateCdrs{}, &reply); err != nil {
		t.Error(err.Error())
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply: %s", reply)
	}
	if nonRatedCdrs, _, err := cdrStor.GetStoredCdrs(&utils.CdrsFilter{CostEnd: utils.Float64Pointer(-1.0)}); err != nil {
		t.Error(err)
	} else if len(nonRatedCdrs) != 2 { // All CDRs should be rated
		t.Error(fmt.Sprintf("Unexpected number of CDRs non-rated: %d", len(nonRatedCdrs)))
	}
	if errRatedCdrs, _, err := cdrStor.GetStoredCdrs(&utils.CdrsFilter{CostStart: utils.Float64Pointer(-1.0), CostEnd: utils.Float64Pointer(0)}); err != nil {
		t.Error(err)
	} else if len(errRatedCdrs) != 1 {
		t.Error(fmt.Sprintf("Unexpected number of CDRs with errors: %d", len(errRatedCdrs)))
	}
	if err := cgrRpc.Call("CdrsV1.RateCdrs", utils.AttrRateCdrs{RerateErrors: true}, &reply); err != nil {
		t.Error(err.Error())
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply: %s", reply)
	}
	if errRatedCdrs, _, err := cdrStor.GetStoredCdrs(&utils.CdrsFilter{CostStart: utils.Float64Pointer(-1.0), CostEnd: utils.Float64Pointer(0)}); err != nil {
		t.Error(err)
	} else if len(errRatedCdrs) != 1 {
		t.Error(fmt.Sprintf("Unexpected number of CDRs with errors: %d", len(errRatedCdrs)))
	}
}


// Simply kill the engine after we are done with tests within this file
func TestMediStopEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	exec.Command("pkill", "cgr-engine").Run()
}
*/
