//go:build integration
// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

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
		testHttpSsLoadTPFromFolder,
		testHttpSsAuth,
		//testHttpSsSession,
		//testHttpSsStopEngine,
		//testHttpSsTerminate,
	}
)

func TestHttpSessionsIt(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		httpSsCfgDIR = "httpss_internal"
	case utils.MetaRedis:
		t.SkipNow()
	case utils.MetaMySQL:
		httpSsCfgDIR = "httpagent_mysql"
	case utils.MetaMongo:
		httpSsCfgDIR = "httpagent_mongo"
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

func testHttpSsAuth(t *testing.T) {
	reqUrl := fmt.Sprintf("http://localhost:2080%s?requestType=Authorization&imsi=2343000000000123&destination=491239440004&sessionID=uuidTestHttpSs",
		httpSsCfg.HTTPAgentCfg()[0].URL)
	rply, err := httpSsClnt.Get(reqUrl)
	if err != nil {
		t.Fatal(err)
	}
	eRply := "MaxDuration=60"
	if rply, err := io.ReadAll(rply.Body); err != nil {
		t.Error(err)
	} else if strings.HasPrefix(string(rply), "Error") {
		t.Errorf("error: <%s>", strings.TrimSuffix(string(rply), "\n"))
	} else if eRply != strings.TrimSuffix(string(rply), "\n") {
		t.Errorf("expecting: %q, received: %q", eRply, rply)
	}
	rply.Body.Close()
}

func testHttpSsSession(t *testing.T) {
	reqUrl := fmt.Sprintf("http://localhost:2080%s?requestType=Session&imsi=2343000000000123&destination=491239440004&sessionID=uuidTestHttpSs",
		httpSsCfg.HTTPAgentCfg()[0].URL)
	rply, err := httpSsClnt.Get(reqUrl)
	if err != nil {
		t.Fatal(err)
	}
	eRply := "MaxDuration=60"
	if rply, err := io.ReadAll(rply.Body); err != nil {
		t.Error(err)
	} else if strings.HasPrefix(string(rply), "Error") {
		t.Errorf("error: <%s>", strings.TrimSuffix(string(rply), "\n"))
	} else if eRply != strings.TrimSuffix(string(rply), "\n") {
		t.Errorf("expecting: %q, received: %q", eRply, rply)
	}
	rply.Body.Close()
	// 	time.Sleep(time.Millisecond)
}

func testHttpSsTerminate(t *testing.T) {
	reqUrl := fmt.Sprintf("http://localhost:2080%s?requestType=Terminate&imsi=2343000000000123&destination=491239440004&sessionID=uuidTestHttpSs",
		httpSsCfg.HTTPAgentCfg()[0].URL)
	rply, err := httpSsClnt.Get(reqUrl)
	if err != nil {
		t.Fatal(err)
	}
	eRply := "MaxDuration=60"
	if rply, err := io.ReadAll(rply.Body); err != nil {
		t.Error(err)
	} else if strings.HasPrefix(string(rply), "Error") {
		t.Errorf("error: <%s>", strings.TrimSuffix(string(rply), "\n"))
	} else if eRply != strings.TrimSuffix(string(rply), "\n") {
		t.Errorf("expecting: %q, received: %q", eRply, rply)
	}
	rply.Body.Close()
	// 	time.Sleep(time.Millisecond)
}

func testHttpSsStopEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
