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

package main

import (
	"bytes"
	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"os/exec"
	"path"
	"testing"
)

var (
	cgrLdrCfgPath string
	cgrLdrCfgDir string
	cgrLdrCfg  *config.CGRConfig
	cgrLdrBIRPC        *birpc.Client
	cgrLdrTests = []func(t *testing.T) {
		testCgrLdrInitCfg,
		testCgrLdrInitDataDB,
		testCgrLdrInitStorDB,
		testCgrLdrStartEngine,
		testCgrLdrRPCConn,
		testCgrLdrGetSubsystemsBeforeLoad,
		testCgrLdrKillEngine,
		//testCgrLdrLoadData,
	}
)

func TestCGRLoaderRemove(t *testing.T) {
	switch *dbType{
	case utils.MetaInternal:
		cgrLdrCfgDir = "tutinternal"
	case utils.MetaMongo:
		cgrLdrCfgDir = "tutmongo"
	case utils.MetaMySQL:
		cgrLdrCfgDir = "tutmysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, test := range cgrLdrTests {
		t.Run("cgr-loader remove tests", test)
	}
}

func testCgrLdrInitCfg(t *testing.T) {
	var err error
	cgrLdrCfgPath = path.Join(*dataDir, "conf", "samples", cgrLdrCfgDir)
	cgrLdrCfg, err = config.NewCGRConfigFromPath(cgrLdrCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testCgrLdrInitDataDB(t *testing.T) {
	if err := engine.InitDataDB(cgrLdrCfg); err != nil {
		t.Fatal(err)
	}
}

func testCgrLdrInitStorDB(t *testing.T) {
	if err := engine.InitStorDB(cgrLdrCfg); err != nil {
		t.Fatal(err)
	}
}

func testCgrLdrStartEngine(t *testing.T){
	if _, err := engine.StopStartEngine(cgrLdrCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testCgrLdrRPCConn(t *testing.T) {
	var err error
	cgrLdrBIRPC, err = newRPCClient(cgrLdrCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testCgrLdrGetSubsystemsBeforeLoad(t *testing.T) {
	//accountPrf
	var replyAcc *utils.Account
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&replyAcc); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %+q", utils.ErrNotFound.Error(), err.Error())
	}

	//actionsPrf
	var replyAct *engine.ActionProfile
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetActionProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ONE_TIME_ACT"}},
		&replyAct); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %+q", utils.ErrNotFound.Error(), err.Error())
	}

	//attributesPrf
	var replyAttr *engine.APIAttributeProfile
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ALS1"}},
		&replyAttr); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %+q", utils.ErrNotFound.Error(), err.Error())
	}

	//filtersPrf
	var replyFltr *engine.Filter
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetFilter,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "FLTR_1"}},
		&replyFltr); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %+q", utils.ErrNotFound.Error(), err.Error())
	}

	//ratesPrf
	var replyRates *utils.RateProfile
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetRateProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RP1"}},
		&replyRates); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %+q", utils.ErrNotFound.Error(), err.Error())
	}

	// resourcesPrf
	var replyRes *engine.ResourceProfile
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetResourceProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ResGroup21"}},
		&replyRes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %+q", utils.ErrNotFound.Error(), err.Error())
	}

	// routesPrf
	var replyRts *engine.RouteProfile
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetRouteProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RoutePrf1"}},
		&replyRts); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %+q", utils.ErrNotFound.Error(), err.Error())
	}

	// statsPrf
	var replySts *engine.StatQueueProfile
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetStatQueueProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "TestStats"}},
		&replySts); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %+q", utils.ErrNotFound.Error(), err.Error())
	}

	// thresholdPrf
	var replyThdPrf *engine.ThresholdProfile
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetThresholdProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Threshold1"}},
		&replyThdPrf); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %+q", utils.ErrNotFound.Error(), err.Error())
	}

	// threshold
	var rplyThd *engine.Threshold
	if err := cgrLdrBIRPC.Call(context.Background(), utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Threshold1"}},
		&rplyThd); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %+q", utils.ErrNotFound.Error(), err.Error())
	}
}

func testCgrLdrLoadData(t *testing.T) {
	cmd := exec.Command("cgr-loader", "-config_path="+cgrLdrCfgPath, "-path="+path.Join(*dataDir, "tariffplans", "tutorial"))
	output := bytes.NewBuffer(nil)
	outerr := bytes.NewBuffer(nil)
	cmd.Stdout = output
	cmd.Stderr = outerr
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Log(outerr.String())
		t.Fatal(err)
	}
}


//Kill the engine when it is about to be finished
func testCgrLdrKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}