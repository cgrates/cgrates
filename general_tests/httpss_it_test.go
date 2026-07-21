//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	httpSsCfgPath string
	httpSsCfgDIR  string
	httpSsCfg     *config.CGRConfig
	httpSsRPC     *birpc.Client
	httpSsClnt    *http.Client // so we can cache the connection
	err           error

	httpSsTests = []func(t *testing.T){
		testHttpSsInitCfg,
		testHttpSsHttpClnt,
		// 		testHAitResetDB,
		testHttpSsStartEngine,
		testHttpSsRPC,
		//testHttpSsLoadTPFromFolder,
		testHttpSsProvisionData,
		testHttpSsAuth,
		testHttpSsSessionStart,
		testHttpSsSessionUpdate1,
		testHttpSsSessionUpdate2,
		testHttpSsSessionUpdate3,
		testHttpSsTerminate,
		testHttpSsStopEngine,
	}
)

func TestHttpSessionsIt(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		httpSsCfgDIR = "httpss_internal"
	case utils.MetaRedis:
		t.SkipNow()
	case utils.MetaMySQL:
		t.SkipNow()
	case utils.MetaMongo:
		t.SkipNow()
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range httpSsTests {
		t.Run(httpSsCfgDIR, stest)
	}
}

// // Init config first
func testHttpSsInitCfg(t *testing.T) {
	var err error
	if err := os.RemoveAll("/tmp/TestHttpSessionsIt/loader/in"); err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll("/tmp/TestHttpSessionsIt/loader/in", 0755); err != nil {
		t.Fatal(err)
	}
	httpSsCfgPath = path.Join(*utils.DataDir, "conf", "samples", httpSsCfgDIR)
	httpSsCfg, err = config.NewCGRConfigFromPath(context.Background(), httpSsCfgPath)
	if err != nil {
		t.Error(err)
	}
}

// Initialize the HttpClient so we can cache it's connections
func testHttpSsHttpClnt(t *testing.T) {
	httpSsClnt = new(http.Client)
}

// // Start CGR Engine
func testHttpSsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(httpSsCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// // Connect RPC client to rater
func testHttpSsRPC(t *testing.T) {
	httpSsRPC = engine.NewRPCClient(t, httpSsCfg.ListenCfg(), *utils.Encoding) // We connect over JSON so we can also troubleshoot if needed
}

// Load the data from offline TP
func testHttpSsLoadTPFromFolder(t *testing.T) {
	var reply string
	if err := httpSsRPC.Call(context.Background(), utils.LoaderSv1Run,
		&loaders.ArgsProcessFolder{
			Path: path.Join(*utils.DataDir, "tariffplans", "sessions")}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
	//time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

// testHttpSsProvisionAccount will take care about data provisioning for the tests
func testHttpSsProvisionData(t *testing.T) {
	var reply string
	if err := httpSsRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		&utils.AccountWithAPIOpts{
			Account: &utils.Account{
				Tenant:    "cgrates.org",
				ID:        "2343000000000123",
				FilterIDs: []string{"*string:~*req.IMSI:2343000000000123"},
				Balances: map[string]*utils.Balance{
					"DATA1": {
						ID:      "DATA1",
						Type:    utils.MetaAbstract,
						Weights: utils.DynamicWeights{{Weight: 5}},
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimal(1, 0),
								RecurrentFee: utils.NewDecimal(0, 0)},
						},
						Units: utils.NewDecimalFromFloat64(1000), // 1GB (1 unit = 1kb)
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("SetAccount: %v", err)
	}
	var acnt utils.Account
	if err := httpSsRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "2343000000000123"}},
		&acnt); err != nil {
		t.Fatalf("GetAccount: %v", err)
	} else if acnt.Balances["DATA1"].Units.Compare(utils.NewDecimalFromFloat64(1000)) != 0 {
		t.Error(fmt.Sprintf("Received account: %+v", acnt))
	}
}

func testHttpSsAuth(t *testing.T) {
	reqUrl := fmt.Sprintf("http://localhost:2080%s?requestType=Authorization&imsi=2343000000000123&sessionID=uuidTestHttpSs",
		httpSsCfg.HTTPAgentCfg()[0].URL)
	rply, err := httpSsClnt.Get(reqUrl)
	if err != nil {
		t.Fatal(err)
	}
	eRply := "MaxUsage=1000"
	if rply, err := io.ReadAll(rply.Body); err != nil {
		t.Error(err)
	} else if strings.HasPrefix(string(rply), "Error") {
		t.Errorf("error: <%s>", strings.TrimSuffix(string(rply), "\n"))
	} else if eRply != strings.TrimSuffix(string(rply), "\n") {
		t.Errorf("expecting: %q, received: %q", eRply, rply)
	}
	rply.Body.Close()
	var acnt utils.Account
	if err := httpSsRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "2343000000000123"}},
		&acnt); err != nil {
		t.Fatalf("GetAccount: %v", err)
	} else if acnt.Balances["DATA1"].Units.Compare(utils.NewDecimalFromFloat64(1000)) != 0 {
		t.Error(fmt.Sprintf("Received account: %+v", acnt))
	}
}

func testHttpSsSessionStart(t *testing.T) {
	reqUrl := fmt.Sprintf("http://localhost:2080%s?requestType=SessionStart&imsi=2343000000000123&sessionID=uuidTestHttpSs",
		httpSsCfg.HTTPAgentCfg()[0].URL)
	rply, err := httpSsClnt.Get(reqUrl)
	if err != nil {
		t.Fatal(err)
	}
	eRply := "MaxUsage=50"
	if rply, err := io.ReadAll(rply.Body); err != nil {
		t.Error(err)
	} else if strings.HasPrefix(string(rply), "Error") {
		t.Errorf("error: <%s>", strings.TrimSuffix(string(rply), "\n"))
	} else if eRply != strings.TrimSuffix(string(rply), "\n") {
		t.Errorf("expecting: %q, received: %q", eRply, rply)
	}
	rply.Body.Close()
	var aSessions []*sessions.ExternalSession
	if err := httpSsRPC.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		utils.SessionFilter{}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	}

	var acnt utils.Account
	if err := httpSsRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "2343000000000123"}},
		&acnt); err != nil {
		t.Fatalf("GetAccount: %v", err)
	} else if acnt.Balances["DATA1"].Units.Compare(utils.NewDecimalFromFloat64(950)) != 0 {
		t.Error(fmt.Sprintf("Received account: %+v", acnt))
	}
}

func testHttpSsSessionUpdate1(t *testing.T) {
	reqUrl := fmt.Sprintf("http://localhost:2080%s?requestType=SessionUpdate&imsi=2343000000000123&sessionID=uuidTestHttpSs&usedUnits=40",
		httpSsCfg.HTTPAgentCfg()[0].URL)
	rply, err := httpSsClnt.Get(reqUrl)
	if err != nil {
		t.Fatal(err)
	}
	eRply := "MaxUsage=50"
	if rply, err := io.ReadAll(rply.Body); err != nil {
		t.Error(err)
	} else if strings.HasPrefix(string(rply), "Error") {
		t.Errorf("error: <%s>", strings.TrimSuffix(string(rply), "\n"))
	} else if eRply != strings.TrimSuffix(string(rply), "\n") {
		t.Errorf("expecting: %q, received: %q", eRply, rply)
	}
	rply.Body.Close()
	var aSessions []*sessions.ExternalSession
	if err := httpSsRPC.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		utils.SessionFilter{}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if *aSessions[0].InterimUsage != 50 {
		t.Errorf("Unexpected InterimUsage in session: %+v", aSessions[0])
	} else if *aSessions[0].UsageAdjustment != 0 {
		t.Errorf("Unexpected UsageAdjustment in session: %+v", aSessions[0])
	} else if *aSessions[0].TotalUsage != 90 {
		t.Errorf("Unexpected TotalUsage in session: %+v", aSessions[0])
	}
	var acnt utils.Account
	if err := httpSsRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "2343000000000123"}},
		&acnt); err != nil {
		t.Fatalf("GetAccount: %v", err)
	} else if acnt.Balances["DATA1"].Units.Compare(utils.NewDecimalFromFloat64(910)) != 0 {
		t.Error(fmt.Sprintf("Received account: %+v", acnt))
	}
}

func testHttpSsSessionUpdate2(t *testing.T) {
	reqUrl := fmt.Sprintf("http://localhost:2080%s?requestType=SessionUpdate&imsi=2343000000000123&sessionID=uuidTestHttpSs&usedUnits=60",
		httpSsCfg.HTTPAgentCfg()[0].URL)
	rply, err := httpSsClnt.Get(reqUrl)
	if err != nil {
		t.Fatal(err)
	}
	eRply := "MaxUsage=50"
	if rply, err := io.ReadAll(rply.Body); err != nil {
		t.Error(err)
	} else if strings.HasPrefix(string(rply), "Error") {
		t.Errorf("error: <%s>", strings.TrimSuffix(string(rply), "\n"))
	} else if eRply != strings.TrimSuffix(string(rply), "\n") {
		t.Errorf("expecting: %q, received: %q", eRply, rply)
	}
	rply.Body.Close()
	var aSessions []*sessions.ExternalSession
	if err := httpSsRPC.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		utils.SessionFilter{}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if *aSessions[0].InterimUsage != 50 {
		t.Errorf("Unexpected InterimUsage in session: %+v", aSessions[0])
	} else if *aSessions[0].UsageAdjustment != 0 {
		t.Errorf("Unexpected UsageAdjustment in session: %+v", aSessions[0])
	} else if *aSessions[0].TotalUsage != 150 {
		t.Errorf("Unexpected TotalUsage in session: %+v", aSessions[0])
	}
	var acnt utils.Account
	if err := httpSsRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "2343000000000123"}},
		&acnt); err != nil {
		t.Fatalf("GetAccount: %v", err)
	} else if acnt.Balances["DATA1"].Units.Compare(utils.NewDecimalFromFloat64(850)) != 0 {
		t.Error(fmt.Sprintf("Received account: %+v", acnt))
	}
}

func testHttpSsSessionUpdate3(t *testing.T) {
	reqUrl := fmt.Sprintf("http://localhost:2080%s?requestType=SessionUpdate&imsi=2343000000000123&sessionID=uuidTestHttpSs&usedUnits=30&totalUnits=250",
		httpSsCfg.HTTPAgentCfg()[0].URL)
	rply, err := httpSsClnt.Get(reqUrl)
	if err != nil {
		t.Fatal(err)
	}
	eRply := "MaxUsage=170" // difference between active 130 and  and totalUnits of 250 plus the interimUsage of 50
	if rply, err := io.ReadAll(rply.Body); err != nil {
		t.Error(err)
	} else if strings.HasPrefix(string(rply), "Error") {
		t.Errorf("error: <%s>", strings.TrimSuffix(string(rply), "\n"))
	} else if eRply != strings.TrimSuffix(string(rply), "\n") {
		t.Errorf("expecting: %q, received: %q", eRply, rply)
	}
	rply.Body.Close()
	var aSessions []*sessions.ExternalSession
	if err := httpSsRPC.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		utils.SessionFilter{}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if *aSessions[0].InterimUsage != 170 {
		t.Errorf("Unexpected InterimUsage in session: %+v", aSessions[0])
	} else if *aSessions[0].UsageAdjustment != 0 {
		t.Errorf("Unexpected UsageAdjustment in session: %+v", aSessions[0])
	} else if *aSessions[0].TotalUsage != 300 {
		t.Errorf("Unexpected TotalUsage in session: %+v", aSessions[0])
	}
	var acnt utils.Account
	if err := httpSsRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "2343000000000123"}},
		&acnt); err != nil {
		t.Fatalf("GetAccount: %v", err)
	} else if acnt.Balances["DATA1"].Units.Compare(utils.NewDecimalFromFloat64(700)) != 0 {
		t.Error(fmt.Sprintf("Received account: %+v", acnt))
	}
}

// Should correct the last debit, instead of 150000 have only used 100000, total session should be 200.000
func testHttpSsTerminate(t *testing.T) {
	reqUrl := fmt.Sprintf("http://localhost:2080%s?requestType=Terminate&imsi=2343000000000123&destination=491239440004&sessionID=uuidTestHttpSs&usedUnits=100&totalUnits=200",
		httpSsCfg.HTTPAgentCfg()[0].URL)
	rply, err := httpSsClnt.Get(reqUrl)
	if err != nil {
		t.Fatal(err)
	}
	eRply := "MaxUsage=200" // total session usage
	if rply, err := io.ReadAll(rply.Body); err != nil {
		t.Error(err)
	} else if strings.HasPrefix(string(rply), "Error") {
		t.Errorf("error: <%s>", strings.TrimSuffix(string(rply), "\n"))
	} else if eRply != strings.TrimSuffix(string(rply), "\n") {
		t.Errorf("expecting: %q, received: %q", eRply, rply)
	}
	rply.Body.Close()
	var aSessions []*sessions.ExternalSession
	if err := httpSsRPC.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		utils.SessionFilter{}, &aSessions); err == nil || err.Error() != utils.NotFoundCaps {
		t.Error(err)
	} else if len(aSessions) != 0 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	}
	var acnt utils.Account
	if err := httpSsRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "2343000000000123"}},
		&acnt); err != nil {
		t.Fatalf("GetAccount: %v", err)
	} else if acnt.Balances["DATA1"].Units.Compare(utils.NewDecimalFromFloat64(800)) != 0 {
		t.Error(fmt.Sprintf("Received account: %+v", acnt))
	}
}

func testHttpSsStopEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
