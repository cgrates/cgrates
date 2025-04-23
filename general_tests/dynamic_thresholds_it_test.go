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
	dynThdCfgPath string
	dynThdCfg     *config.CGRConfig
	dynThdRpc     *birpc.Client
	dymThdConfDIR string
	dynThdDelay   int

	sTestsDynThd = []func(t *testing.T){
		testDynThdLoadConfig,
		testDynThdInitDataDb,
		testDynThdResetStorDb,
		testDynThdStartEngine,
		testDynThdRpcConn,
		testDynThdCheckForThresholdProfile,
		testDynThdSetAction,
		testDynThdSetThresholdProfile,
		testDynThdGetThresholdBeforeDebit,
		testDynThdSetBalance,
		testDynThdGetAccountBeforeDebit,
		testDynThdDebit1,
		testDynThdGetThresholdBeforeDebit,
		testDynThdDebit2,
		testDynThdGetAccountAfterDebit,
		testDynThdGetThresholdAfterDebit,
		testDynThdCheckForDynCreatedThresholdProfile,
		testDynThdStopEngine,
	}
)

// Test starts here
func TestDynThdIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		dymThdConfDIR = "tutinternal"
	case utils.MetaMySQL:
		dymThdConfDIR = "tutmysql"
	case utils.MetaMongo:
		dymThdConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsDynThd {
		t.Run(dymThdConfDIR, stest)
	}
}

func testDynThdLoadConfig(t *testing.T) {
	var err error
	dynThdCfgPath = path.Join(*utils.DataDir, "conf", "samples", dymThdConfDIR)
	if dynThdCfg, err = config.NewCGRConfigFromPath(dynThdCfgPath); err != nil {
		t.Error(err)
	}
	dynThdDelay = 1000
}

func testDynThdInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(dynThdCfg); err != nil {
		t.Fatal(err)
	}
}

func testDynThdResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(dynThdCfg); err != nil {
		t.Fatal(err)
	}
}

func testDynThdStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(dynThdCfgPath, dynThdDelay); err != nil {
		t.Fatal(err)
	}
}

func testDynThdRpcConn(t *testing.T) {
	dynThdRpc = engine.NewRPCClient(t, dynThdCfg.ListenCfg())
}

func testDynThdCheckForThresholdProfile(t *testing.T) {
	var rply *engine.ThresholdProfile
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetThresholdProfile, &utils.TenantID{Tenant: "cgrates.org", ID: "DYNAMICLY_CREATED_THRESHOLD"}, &rply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testDynThdSetAction(t *testing.T) {
	var reply string

	act := &utils.AttrSetActions{
		ActionsId: "DYNAMIC_THRESHOLD_ACTION",
		Actions: []*utils.TPAction{{
			Identifier:      utils.MetaDynamicThreshold,
			ExtraParameters: "cgrates.org;DYNAMICLY_THR_<~*req.ID>;*string:~*opts.*eventType:AccountUpdate;;1;;;true;10;;true;~*opts",
		}}}
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv2SetActions,
		act, &reply); err != nil {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
}

func testDynThdSetThresholdProfile(t *testing.T) {
	ThdPrf := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*opts.*eventType:AccountUpdate", "*string:~*asm.ID:1002", "*lt:~*asm.BalanceSummaries.testBalanceID.Value:56m", "*gte:~*asm.BalanceSummaries.testBalanceID.Initial:58m"},
			ID:        "THD_ACNT_1002",
			MaxHits:   1,
			ActionIDs: []string{"DYNAMIC_THRESHOLD_ACTION"},
			Async:     true,
		},
	}
	var reply string
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1SetThresholdProfile, ThdPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	args := &utils.TenantID{
		Tenant: "cgrates.org",
		ID:     "THD_ACNT_1002",
	}

	var result1 *engine.ThresholdProfile
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetThresholdProfile, args, &result1); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result1, ThdPrf.ThresholdProfile) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(ThdPrf.ThresholdProfile), utils.ToJSON(result1))
	}
}

func testDynThdGetThresholdBeforeDebit(t *testing.T) {
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
	if err := dynThdRpc.Call(context.Background(), utils.ThresholdSv1GetThreshold, &utils.TenantIDWithAPIOpts{TenantID: args}, &result2); err != nil {
		t.Error(err)
	} else if result2.Snooze = expThd.Snooze; !reflect.DeepEqual(result2, expThd) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expThd), utils.ToJSON(result2))
	}
}
func testDynThdSetBalance(t *testing.T) {
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
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv2SetBalance, args, &reply); err != nil {
		t.Error("Got error on SetBalance: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling SetBalance received: %s", reply)
	}
}

func testDynThdGetAccountBeforeDebit(t *testing.T) {
	exp := float64(time.Hour)
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1002",
	}
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MetaVoice].GetTotalValue(); rply != exp {
		t.Errorf("Expecting: %v, received: %v",
			exp, rply)
	}
}

func testDynThdDebit1(t *testing.T) {
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
	if err := dynThdRpc.Call(context.Background(), utils.ResponderMaxDebit, &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: cd,
	}, cc); err != nil {
		t.Error(err)
	}
}

func testDynThdDebit2(t *testing.T) {
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
	if err := dynThdRpc.Call(context.Background(), utils.ResponderMaxDebit, &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: cd,
	}, cc); err != nil {
		t.Error(err)
	}
}

func testDynThdGetAccountAfterDebit(t *testing.T) {
	exp := float64(time.Hour - 5*time.Minute - 5*time.Second)
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1002",
	}
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MetaVoice].GetTotalValue(); rply != exp {
		t.Errorf("Expecting: %v, received: %v",
			exp, rply)
	}
}

func testDynThdGetThresholdAfterDebit(t *testing.T) {
	var result2 *engine.Threshold
	if err := dynThdRpc.Call(context.Background(), utils.ThresholdSv1GetThreshold, &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1002"}}, &result2); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testDynThdCheckForDynCreatedThresholdProfile(t *testing.T) {
	exp := &engine.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "DYNAMICLY_THR_1002",
		FilterIDs: []string{"*string:~*opts.*eventType:AccountUpdate"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Time{},
			ExpiryTime:     time.Time{},
		},
		MaxHits: 1,
		Blocker: true,
		Weight:  10,
		Async:   true,
	}
	args := &utils.TenantID{
		Tenant: "cgrates.org",
		ID:     "DYNAMICLY_THR_1002",
	}
	var result1 *engine.ThresholdProfile
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetThresholdProfile, args, &result1); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result1, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(result1))
	}
}

func testDynThdStopEngine(t *testing.T) {
	if err := engine.KillEngine(dynThdDelay); err != nil {
		t.Error(err)
	}
}
