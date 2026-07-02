//go:build call
// +build call

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
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	voiceblender "github.com/VoiceBlender/voiceblender-go"
	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/birpc/jsonrpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/utils"
)

var (
	tutorialCallsCfg *config.CGRConfig
	tutorialCallsRpc *birpc.Client

	fsConfig           = flag.String("fsConfig", "/usr/share/cgrates/tutorial_tests/fs_evsock", "FreeSwitch tutorial folder")
	ariConf            = flag.String("ariConf", "/usr/share/cgrates/tutorial_tests/asterisk_ari", "Asterisk tutorial folder")
	voiceblenderAPIURL = flag.String("vbAPIURL", "http://127.0.0.1:8080/v1", "VoiceBlender REST API base URL")
	vbClient           *voiceblender.Client
	vbCmd              *exec.Cmd
	optConf            string
)

var sTestsCalls = []func(t *testing.T){
	testCallInitCfg,
	testCallResetDataDb,
	testCallStartFS,
	testCallRestartFS,
	testCallStartEngine,
	testCallRpcConn,
	testCallLoadTariffPlanFromFolder,
	testCallInitVoiceBlender,
	testCallCall1001To1002,
	testCallGetCDRs,
	testCallCheckBalance,
	testCallStopVoiceBlender,
	testCallStopCgrEngine,
	testCallFS,
}

func TestFreeswitchCalls(t *testing.T) {
	optConf = utils.Freeswitch
	for _, stest := range sTestsCalls {
		t.Run("", stest)
	}
}

func TestAsteriskCalls(t *testing.T) {
	optConf = utils.Asterisk
	for _, stest := range sTestsCalls {
		t.Run("", stest)
	}
}

func testCallInitCfg(t *testing.T) {
	var err error
	switch optConf {
	case utils.Freeswitch:
		tutorialCallsCfg, err = config.NewCGRConfigFromPath(context.Background(),
			path.Join(*ariConf, "cgrates", "etc", "cgrates"))
		if err != nil {
			t.Error(err)
		}
	case utils.Asterisk:
		tutorialCallsCfg, err = config.NewCGRConfigFromPath(context.Background(),
			path.Join(*fsConfig, "cgrates", "etc", "cgrates"))
		if err != nil {
			t.Error(err)
		}
	default:
		t.Error("Invalid option")
	}

	tutorialCallsCfg.DataFolderPath = *utils.DataDir // Share DataFolderPath through config towards StoreDb for Flush()
}

func testCallResetDataDb(t *testing.T) {
	if err := engine.InitDB(tutorialCallsCfg); err != nil {
		t.Fatal(err)
	}
}

func testCallStartFS(t *testing.T) {
	switch optConf {
	case utils.Freeswitch:
		engine.KillProcName(utils.Freeswitch, 5000)
		if err := engine.CallScript(path.Join(*fsConfig, "freeswitch", "etc", "init.d", "freeswitch"), "start", 3000); err != nil {
			t.Fatal(err)
		}
	case utils.Asterisk:
		engine.KillProcName(utils.Asterisk, 5000)
		if err := engine.CallScript(path.Join(*ariConf, "asterisk", "etc", "init.d", "asterisk"), "start", 3000); err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatalf("Invalid option")
	}
}

func testCallRestartFS(t *testing.T) {
	switch optConf {
	case utils.Freeswitch:
		engine.KillProcName(utils.Freeswitch, 5000)
		if err := engine.CallScript(path.Join(*fsConfig, "freeswitch", "etc", "init.d", "freeswitch"), "restart", 3000); err != nil {
			t.Fatal(err)
		}
	case utils.Asterisk:
		engine.KillProcName(utils.Asterisk, 5000)
		if err := engine.CallScript(path.Join(*ariConf, "asterisk", "etc", "init.d", "asterisk"), "restart", 3000); err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatalf("Invalid option")
	}
}

func testCallStartEngine(t *testing.T) {
	engine.KillProcName("cgr-engine", *utils.WaitRater)
	switch optConf {
	case utils.Freeswitch:
		if err := engine.CallScript(path.Join(*fsConfig, "cgrates", "etc", "init.d", "cgrates"), "start", 100); err != nil {
			t.Fatal(err)
		}
	case utils.Asterisk:
		if err := engine.CallScript(path.Join(*ariConf, "cgrates", "etc", "init.d", "cgrates"), "start", 100); err != nil {
			t.Fatal(err)
		}
	default:
		t.Error("invalid option")
	}
}

func testCallRpcConn(t *testing.T) {
	var err error
	for i := 0; i < 20; i++ {
		if tutorialCallsRpc, err = jsonrpc.Dial(utils.TCP, tutorialCallsCfg.ListenCfg().RPCJSONListen); err == nil {
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatal(err)
}

func testCallLoadTariffPlanFromFolder(t *testing.T) {
	var reply string
	args := &loaders.ArgsProcessFolder{
		Path: path.Join(*utils.DataDir, "tariffplans", "tutorial"),
		APIOpts: map[string]any{
			utils.MetaCache:       utils.MetaReload,
			utils.MetaStopOnError: true,
		},
	}
	if err := tutorialCallsRpc.Call(context.Background(), utils.LoaderSv1Run, args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned: %s", reply)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond)
}

func testCallInitVoiceBlender(t *testing.T) {
	engine.KillProcName("voiceblender", 1000)

	vbPath, err := exec.LookPath("voiceblender")
	if err != nil {
		t.Fatal(err)
	}
	vbCmd = exec.Command(vbPath)
	vbCmd.Stdout = os.Stdout
	vbCmd.Stderr = os.Stderr
	if err := vbCmd.Start(); err != nil {
		t.Fatalf("starting voiceblender: %v", err)
	}

	u, err := url.Parse(*voiceblenderAPIURL)
	if err != nil {
		t.Fatal(err)
	}
	addr := u.Host
	if u.Port() == "" {
		addr = net.JoinHostPort(u.Hostname(), "80")
	}
	var ready bool
	for range 10 {
		conn, derr := net.DialTimeout(utils.TCP, addr, 200*time.Millisecond)
		if derr == nil {
			conn.Close()
			ready = true
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	if !ready {
		t.Fatalf("voiceblender REST API %s not reachable", addr)
	}

	vbClient = voiceblender.New(voiceblender.WithBaseURL(*voiceblenderAPIURL))
}

func vbOriginate(t *testing.T, from, dst string, dur time.Duration) {
	t.Helper()
	if _, err := vbClient.CreateLeg(context.Background(), voiceblender.CreateLegRequest{
		Type:        "sip",
		To:          fmt.Sprintf("sip:%s@127.0.0.1:5080", dst),
		From:        from,
		MaxDuration: int(dur.Seconds()),
	}); err != nil {
		t.Fatal(err)
	}
}

func testCallCall1001To1002(t *testing.T) {
	vbOriginate(t, "1001", "1002", 15*time.Second)
	time.Sleep(time.Second)
}

func testCallGetCDRs(t *testing.T) {
	args := &utils.CDRFilters{
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Account:1001"},
	}
	var cdrs []*utils.CDR
	var err error
	for i := 0; i < 30; i++ {
		err = tutorialCallsRpc.Call(context.Background(), utils.AdminSv1GetCDRs, args, &cdrs)
		if err == nil && len(cdrs) != 0 {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil || len(cdrs) == 0 {
		t.Fatalf("expected CDRs for account 1001, received none (err=%v)", err)
	}
	for _, cdr := range cdrs {
		if utils.IfaceAsString(cdr.Event[utils.AccountField]) != "1001" {
			t.Errorf("unexpected CDR account: %s", utils.ToJSON(cdr))
		}
	}
}

func testCallCheckBalance(t *testing.T) {
	const initialUnits = 10.0
	var acc utils.Account
	if err := tutorialCallsRpc.Call(context.Background(), utils.AccountSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&acc); err != nil {
		t.Fatal(err)
	}
	units, ok := acc.Balances["MonetaryBalance"].Units.Float64()
	if !ok {
		t.Fatal("expected MonetaryBalance to exists")
	}
	if units < initialUnits {
		t.Errorf("expected account 1001 balance below initial %v (debited), got %v: %s",
			initialUnits, units, utils.ToJSON(acc))
		return
	}
}

func testCallStopVoiceBlender(t *testing.T) {
	if vbCmd == nil || vbCmd.Process == nil {
		return
	}
	if err := vbCmd.Process.Kill(); err != nil {
		t.Error(err)
	}
	vbCmd.Wait()
}

func testCallStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func testCallFS(t *testing.T) {
	switch optConf {
	case utils.Freeswitch:
		engine.ForceKillProcName(utils.Freeswitch, 1000)
	case utils.Asterisk:
		engine.ForceKillProcName(utils.Asterisk, 1000)
	default:
		t.Errorf("invalid option")
	}
}
