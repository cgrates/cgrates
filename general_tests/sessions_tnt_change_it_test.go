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

package general_tests

import (
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	sesTntChngCfgDir  string
	sesTntChngCfgPath string
	sesTntChngCfg     *config.CGRConfig
	sesTntChngRPC     *birpc.Client

	sesTntChngTests = []func(t *testing.T){
		testSesTntChngLoadConfig,
		testSesTntChngResetDataDB,
		testSesTntChngResetStorDb,
		testSesTntChngStartEngine,
		testSesTntChngRPCConn,
		testSesTntChngSetChargerProfile1,
		testSesTntChngSetChargerProfile2,
		testChargerSAuthProcessEventAuth,
		testSesTntChngStopCgrEngine,
	}
)

func TestSesTntChange(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		sesTntChngCfgDir = "tutinternal"
	case utils.MetaMySQL:
		sesTntChngCfgDir = "tutmysql"
	case utils.MetaMongo:
		sesTntChngCfgDir = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, sestest := range sesTntChngTests {
		t.Run(sesTntChngCfgDir, sestest)
	}
}

func testSesTntChngLoadConfig(t *testing.T) {
	var err error
	sesTntChngCfgPath = path.Join(*utils.DataDir, "conf", "samples", sesTntChngCfgDir)
	if sesTntChngCfg, err = config.NewCGRConfigFromPath(sesTntChngCfgPath); err != nil {
		t.Error(err)
	}
}

func testSesTntChngResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(sesTntChngCfg); err != nil {
		t.Fatal(err)
	}
}

func testSesTntChngResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(sesTntChngCfg); err != nil {
		t.Fatal(err)
	}
}

func testSesTntChngStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sesTntChngCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testSesTntChngRPCConn(t *testing.T) {
	sesTntChngRPC = engine.NewRPCClient(t, sesTntChngCfg.ListenCfg())
}

func testSesTntChngSetChargerProfile1(t *testing.T) {
	var reply *engine.ChargerProfile
	if err := sesTntChngRPC.Call(context.Background(), utils.APIerSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Charger1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatal(err)
	}
	chargerProfile := &v1.ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "Charger1",
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*constant:*tenant:cgrates.ro;*constant:*req.Account:1234"},
		},
	}

	var result string
	if err := sesTntChngRPC.Call(context.Background(), utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	if err := sesTntChngRPC.Call(context.Background(), utils.APIerSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Charger1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(chargerProfile.ChargerProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", chargerProfile.ChargerProfile, reply)
	}
}

func testSesTntChngSetChargerProfile2(t *testing.T) {
	var reply *engine.ChargerProfile
	if err := sesTntChngRPC.Call(context.Background(), utils.APIerSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Charger2"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatal(err)
	}
	chargerProfile := &v1.ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "Charger2",
			RunID:        utils.MetaRaw,
			AttributeIDs: []string{},
		},
	}

	var result string
	if err := sesTntChngRPC.Call(context.Background(), utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	if err := sesTntChngRPC.Call(context.Background(), utils.APIerSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Charger2"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(chargerProfile.ChargerProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", chargerProfile.ChargerProfile, reply)
	}
}

func testChargerSAuthProcessEventAuth(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "1001",
		BalanceType: utils.MetaVoice,
		Value:       float64(2 * time.Minute),
		Balance: map[string]any{
			utils.ID:            "testSes",
			utils.RatingSubject: "*zero1ms",
		},
	}
	var reply string
	if err := sesTntChngRPC.Call(context.Background(), utils.APIerSv2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}

	attrSetBalance2 := utils.AttrSetBalance{
		Tenant:      "cgrates.ro",
		Account:     "1234",
		BalanceType: utils.MetaVoice,
		Value:       float64(2 * time.Minute),
		Balance: map[string]any{
			utils.ID:            "testSes",
			utils.RatingSubject: "*zero1ms",
		},
	}
	var reply2 string
	if err := sesTntChngRPC.Call(context.Background(), utils.APIerSv2SetBalance, attrSetBalance2, &reply2); err != nil {
		t.Error(err)
	} else if reply2 != utils.OK {
		t.Errorf("Received: %s", reply2)
	}

	ev := &sessions.V1AuthorizeArgs{
		GetMaxUsage: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestEv1",
			Event: map[string]any{
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "TestEv1",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.Usage:        time.Minute,
			},
		},
	}
	var rply sessions.V1AuthorizeReply
	if err := sesTntChngRPC.Call(context.Background(), utils.SessionSv1AuthorizeEvent, ev, &rply); err != nil {
		t.Fatal(err)
	}
	expected := &sessions.V1AuthorizeReply{
		MaxUsage: (*time.Duration)(utils.Int64Pointer(60000000000)),
	}
	if !reflect.DeepEqual(utils.ToJSON(&expected), utils.ToJSON(&rply)) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(&expected), utils.ToJSON(&rply))
	}
}

func testSesTntChngStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
