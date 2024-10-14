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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	accWThdCfgPath string
	accWThdCfg     *config.CGRConfig
	accWThdRpc     *birpc.Client
	accWThdConfDIR string //run tests for specific configuration
	accWThdDelay   int

	sTestsAccWThd = []func(t *testing.T){
		testAccWThdLoadConfig,
		testAccWThdInitDataDb,
		testAccWThdResetStorDb,
		testAccWThdStartEngine,
		testAccWThdRpcConn,
		testAccWThdSetThresholdProfile,
		testAccWThdGetThresholdBeforeDebit,
		testAccWThdSetBalance,
		testAccWThdGetAccountBeforeDebit,
		testAccWThdDebit1,
		testAccWThdGetThresholdBeforeDebit,
		testAccWThdDebit2,
		testAccWThdGetAccountAfterDebit,
		testAccWThdGetThresholdAfterDebit,
		testAccWThdStopEngine,
	}
)

// Test starts here
func TestAccWThdIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		accWThdConfDIR = "tutinternal"
	case utils.MetaMySQL:
		accWThdConfDIR = "tutmysql"
	case utils.MetaMongo:
		accWThdConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsAccWThd {
		t.Run(accWThdConfDIR, stest)
	}
}

func testAccWThdLoadConfig(t *testing.T) {
	var err error
	accWThdCfgPath = path.Join(*utils.DataDir, "conf", "samples", accWThdConfDIR)
	if accWThdCfg, err = config.NewCGRConfigFromPath(accWThdCfgPath); err != nil {
		t.Error(err)
	}
	accWThdDelay = 1000
}

func testAccWThdInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(accWThdCfg); err != nil {
		t.Fatal(err)
	}
}

func testAccWThdResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(accWThdCfg); err != nil {
		t.Fatal(err)
	}
}

func testAccWThdStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(accWThdCfgPath, accWThdDelay); err != nil {
		t.Fatal(err)
	}
}

func testAccWThdRpcConn(t *testing.T) {
	accWThdRpc = engine.NewRPCClient(t, accWThdCfg.ListenCfg())
}

func testAccWThdSetThresholdProfile(t *testing.T) {
	ThdPrf := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*opts.*eventType:AccountUpdate", "*string:~*asm.ID:1002", "*lt:~*asm.BalanceSummaries.testBalanceID.Value:56m", "*gte:~*asm.BalanceSummaries.testBalanceID.Initial:58m"},
			ID:        "THD_ACNT_1002",
			MaxHits:   1,
		},
	}
	var reply string
	if err := accWThdRpc.Call(context.Background(), utils.APIerSv1SetThresholdProfile, ThdPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	args := &utils.TenantID{
		Tenant: "cgrates.org",
		ID:     "THD_ACNT_1002",
	}

	var result1 *engine.ThresholdProfile
	if err := accWThdRpc.Call(context.Background(), utils.APIerSv1GetThresholdProfile, args, &result1); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result1, ThdPrf.ThresholdProfile) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(ThdPrf.ThresholdProfile), utils.ToJSON(result1))
	}
}

func testAccWThdGetThresholdBeforeDebit(t *testing.T) {
	args := &utils.TenantID{
		Tenant: "cgrates.org",
		ID:     "THD_ACNT_1002",
	}

	expThd := &engine.Threshold{
		Tenant: "cgrates.org",
		ID:     "THD_ACNT_1002",
		Hits:   0,
	}

	var result2 *engine.Threshold
	if err := accWThdRpc.Call(context.Background(), utils.ThresholdSv1GetThreshold, &utils.TenantIDWithAPIOpts{TenantID: args}, &result2); err != nil {
		t.Error(err)
	} else if result2.Snooze = expThd.Snooze; !reflect.DeepEqual(result2, expThd) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expThd), utils.ToJSON(result2))
	}
}
func testAccWThdSetBalance(t *testing.T) {
	args := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "1002",
		BalanceType: utils.MetaVoice,
		Value:       float64(time.Hour),
		Balance: map[string]any{
			utils.ID:            "testBalanceID",
			utils.RatingSubject: "*zero1s",
		},
	}
	var reply string
	if err := accWThdRpc.Call(context.Background(), utils.APIerSv2SetBalance, args, &reply); err != nil {
		t.Error("Got error on SetBalance: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling SetBalance received: %s", reply)
	}
}

func testAccWThdGetAccountBeforeDebit(t *testing.T) {
	exp := float64(time.Hour)
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1002",
	}
	if err := accWThdRpc.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MetaVoice].GetTotalValue(); rply != exp {
		t.Errorf("Expecting: %v, received: %v",
			exp, rply)
	}
}

func testAccWThdDebit1(t *testing.T) {
	tStart := time.Date(2021, 5, 5, 12, 0, 0, 0, time.UTC)
	cd := &engine.CallDescriptor{
		Category:      utils.Call,
		Tenant:        "cgrates.org",
		Subject:       "1002",
		Account:       "1002",
		Destination:   "1003",
		TimeStart:     tStart,
		TimeEnd:       tStart.Add(5 * time.Second),
		LoopIndex:     0,
		DurationIndex: 5 * time.Second,
		ToR:           utils.MetaVoice,
		CgrID:         "12345678911",
		RunID:         utils.MetaDefault,
	}
	cc := new(engine.CallCost)
	if err := accWThdRpc.Call(context.Background(), utils.ResponderMaxDebit, &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: cd,
	}, cc); err != nil {
		t.Error(err)
	}
}

func testAccWThdDebit2(t *testing.T) {
	tStart := time.Date(2021, 5, 5, 12, 0, 0, 0, time.UTC)
	cd := &engine.CallDescriptor{
		Category:      utils.Call,
		Tenant:        "cgrates.org",
		Subject:       "1002",
		Account:       "1002",
		Destination:   "1003",
		TimeStart:     tStart,
		TimeEnd:       tStart.Add(5 * time.Minute),
		LoopIndex:     0,
		DurationIndex: 5 * time.Minute,
		ToR:           utils.MetaVoice,
		CgrID:         "12345678910",
		RunID:         utils.MetaDefault,
	}
	cc := new(engine.CallCost)
	if err := accWThdRpc.Call(context.Background(), utils.ResponderMaxDebit, &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: cd,
	}, cc); err != nil {
		t.Error(err)
	}
}

func testAccWThdGetAccountAfterDebit(t *testing.T) {
	exp := float64(time.Hour - 5*time.Minute - 5*time.Second)
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1002",
	}
	if err := accWThdRpc.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MetaVoice].GetTotalValue(); rply != exp {
		t.Errorf("Expecting: %v, received: %v",
			exp, rply)
	}
}

func testAccWThdGetThresholdAfterDebit(t *testing.T) {
	var result2 *engine.Threshold
	if err := accWThdRpc.Call(context.Background(), utils.ThresholdSv1GetThreshold, &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1002"}}, &result2); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testAccWThdStopEngine(t *testing.T) {
	if err := engine.KillEngine(accWThdDelay); err != nil {
		t.Error(err)
	}
}
