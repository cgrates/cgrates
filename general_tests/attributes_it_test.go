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
package general_tests

import (
	"net/rpc"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	attrCfgPath     string
	attrCfg         *config.CGRConfig
	attrRPC         *rpc.Client
	attrDataDir     = "/usr/share/cgrates"
	alsPrf          *v1.AttributeWithCache
	alsPrfConfigDIR string
	sTestsAlsPrf    = []func(t *testing.T){
		testAttributeSInitCfg,
		testAttributeSInitDataDb,
		testAttributeSResetStorDb,
		testAttributeSStartEngine,
		testAttributeSRPCConn,
		testAttributeSLoadFromFolder,
		testAttributeSProcessEvent,
	}
)

func TestAttributeSIT(t *testing.T) {
	attrsTests := sTestsAlsPrf
	switch *dbType {
	case utils.MetaInternal:
		alsPrfConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		alsPrfConfigDIR = "tutmysql"
	case utils.MetaMongo:
		alsPrfConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range attrsTests {
		t.Run(alsPrfConfigDIR, stest)
	}
}

func testAttributeSInitCfg(t *testing.T) {
	var err error
	attrCfgPath = path.Join(attrDataDir, "conf", "samples", alsPrfConfigDIR)
	attrCfg, err = config.NewCGRConfigFromPath(attrCfgPath)
	if err != nil {
		t.Error(err)
	}
	attrCfg.DataFolderPath = attrDataDir // Share DataFolderPath through config towards StoreDb for Flush()
}

func testAttributeSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(attrCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testAttributeSResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(attrCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testAttributeSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(attrCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testAttributeSRPCConn(t *testing.T) {
	var err error
	attrRPC, err = newRPCClient(attrCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testAttributeSLoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testit")}
	if err := attrRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}
func testAttributeSProcessEvent(t *testing.T) {
	ev := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer(utils.MetaSessionS),
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSProcessEvent",
				Event: map[string]interface{}{
					utils.EVENT_NAME: "VariableTest",
					utils.ToR:        utils.VOICE,
				},
			},
		},
	}
	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_VARIABLE"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + utils.Category},
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSProcessEvent",
				Event: map[string]interface{}{
					utils.EVENT_NAME: "VariableTest",
					utils.Category:   utils.VOICE,
					utils.ToR:        utils.VOICE,
				},
			},
		},
	}
	sort.Strings(eRply.AlteredFields)
	var rplyEv engine.AttrSProcessEventReply
	if err := attrRPC.Call(utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Fatal(err)
	}
	sort.Strings(rplyEv.AlteredFields)
	if !reflect.DeepEqual(eRply, rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}
