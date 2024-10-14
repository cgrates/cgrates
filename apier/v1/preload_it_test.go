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

package v1

import (
	"os"
	"os/exec"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/birpc/jsonrpc"
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

var (
	preloadCfgPath string
	preloadCfgDIR  string
	preloadCfg     *config.CGRConfig
	preloadRPC     *birpc.Client

	preloadTests = []func(t *testing.T){
		testCreateDirs,
		testPreloadITInitConfig,
		testPreloadITStartEngine,
		testPreloadITRpcConn,
		testPreloadITVerifyAttributes,
		testCleanupFiles,
		testPreloadITKillEngine,
	}
)

func TestPreload(t *testing.T) {
	preloadCfgDIR = "tutinternal"
	for _, test := range preloadTests {
		t.Run(preloadCfgDIR, test)
	}
}

func testCreateDirs(t *testing.T) {
	for _, dir := range []string{"/tmp/In", "/tmp/Out", "/tmp/LoaderIn", "/tmp/SubpathWithoutMove",
		"/tmp/SubpathLoaderWithMove", "/tmp/SubpathOut", "/tmp/templateLoaderIn", "/tmp/templateLoaderOut",
		"/tmp/customSepLoaderIn", "/tmp/customSepLoaderOut"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal("Error creating folder: ", dir, err)
		}
	}

	if err := os.WriteFile(path.Join("/tmp/In", utils.AttributesCsv), []byte(`
#Tenant,ID,Contexts,FilterIDs,ActivationInterval,AttributeFilterIDs,Path,Type,Value,Blocker,Weight
cgrates.org,ALS1,con1,*string:~*req.Account:1001,2014-07-29T15:00:00Z,*string:~*req.Field1:Initial,*req.Field1,*variable,Sub1,true,20
cgrates.org,ALS1,con2;con3,,,,*req.Field2,*variable,Sub2,true,20
`), 0644); err != nil {
		t.Fatal(err.Error())
	}
}

func testPreloadITInitConfig(t *testing.T) {
	var err error
	preloadCfgPath = path.Join(*utils.DataDir, "conf", "samples", "loaders", preloadCfgDIR)
	if preloadCfg, err = config.NewCGRConfigFromPath(preloadCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

func testPreloadITStartEngine(t *testing.T) {
	enginePath, err := exec.LookPath("cgr-engine")
	if err != nil {
		t.Error(err)
	}
	engine := exec.Command(enginePath, "-config_path", preloadCfgPath, "-preload", "CustomLoader")
	if err := engine.Start(); err != nil {
		t.Error(err)
	}
	fib := utils.FibDuration(time.Millisecond, 0)
	var connected bool
	for i := 0; i < 25; i++ {
		time.Sleep(fib())
		if _, err := jsonrpc.Dial(utils.TCP, preloadCfg.ListenCfg().RPCJSONListen); err != nil {
			t.Logf("Error <%s> when opening test connection to: <%s>",
				err.Error(), preloadCfg.ListenCfg().RPCJSONListen)
		} else {
			connected = true
			break
		}
	}
	if !connected {
		t.Errorf("engine did not open port <%s>", preloadCfg.ListenCfg().RPCJSONListen)
	}
	time.Sleep(100 * time.Millisecond)
}

func testPreloadITRpcConn(t *testing.T) {
	var err error
	preloadRPC, err = newRPCClient(preloadCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testPreloadITVerifyAttributes(t *testing.T) {
	eAttrPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Contexts:  []string{"con1", "con2", "con3"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC)},
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{"*string:~*req.Field1:Initial"},
				Path:      "*req.Field1",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Sub1", utils.InfieldSep),
			},
			{
				FilterIDs: []string{},
				Path:      "*req.Field2",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Sub2", utils.InfieldSep),
			},
		},
		Blocker: true,
		Weight:  20.0,
	}

	var reply *engine.AttributeProfile
	if err := preloadRPC.Call(context.Background(), utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ALS1"}}, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	sort.Strings(reply.Contexts)
	if !reflect.DeepEqual(eAttrPrf, reply) {
		eAttrPrf.Attributes[1].FilterIDs = nil
		if !reflect.DeepEqual(eAttrPrf, reply) {
			t.Errorf("Expecting : %+v,\n received: %+v", utils.ToJSON(eAttrPrf), utils.ToJSON(reply))
		}
	}
}

func testCleanupFiles(t *testing.T) {
	for _, dir := range []string{"/tmp/In", "/tmp/Out", "/tmp/LoaderIn", "/tmp/SubpathWithoutMove",
		"/tmp/SubpathLoaderWithMove", "/tmp/SubpathOut"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
	}
}

func testPreloadITKillEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
