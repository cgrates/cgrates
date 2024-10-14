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
package engine

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"reflect"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	actsLclCfg       *config.CGRConfig
	actsLclRpc       *birpc.Client
	actsLclCfgPath   string
	actionsConfigDIR string

	sTestsActionsit = []func(t *testing.T){
		testActionsitInitCfg,
		testActionsitInitCdrDb,
		testActionsitStartEngine,
		testActionsitRpcConn,
		testActionsitSetCdrlogDebit,
		testActionsitSetCdrlogTopup,
		testActionsitCdrlogEmpty,
		testActionsitCdrlogWithParams,
		testActionsitCdrlogWithParams2,
		testActionsitThresholdCDrLog,
		testActionsitCDRAccount,
		testActionsitThresholdCgrRpcAction,
		testActionsitThresholdPostEvent,
		testActionsitSetSDestinations,
		testActionsitresetAccountCDR,
		testActionsitremoteSetAccount,
		testActionsitStopCgrEngine,
	}
)

func TestActionsit(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		actionsConfigDIR = "actions_internal"
	case utils.MetaMySQL:
		actionsConfigDIR = "actions_mysql"
	case utils.MetaMongo:
		actionsConfigDIR = "actions_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	if *utils.Encoding == utils.MetaGOB {
		actionsConfigDIR += "_gob"
	}

	for _, stest := range sTestsActionsit {
		t.Run(actionsConfigDIR, stest)
	}
}

func testActionsitInitCfg(t *testing.T) {
	actsLclCfgPath = path.Join(*utils.DataDir, "conf", "samples", actionsConfigDIR)
	// Init config first
	var err error
	actsLclCfg, err = config.NewCGRConfigFromPath(actsLclCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testActionsitInitCdrDb(t *testing.T) {
	if err := InitDataDb(actsLclCfg); err != nil { // need it for versions
		t.Fatal(err)
	}
	if err := InitStorDb(actsLclCfg); err != nil {
		t.Fatal(err)
	}
}

// Finds cgr-engine executable and starts it with default configuration
func testActionsitStartEngine(t *testing.T) {
	if _, err := StopStartEngine(actsLclCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testActionsitRpcConn(t *testing.T) {
	actsLclRpc = NewRPCClient(t, actsLclCfg.ListenCfg())
}

func testActionsitSetCdrlogDebit(t *testing.T) {
	var reply string
	attrsSetAccount := &utils.AttrSetAccount{Tenant: "cgrates.org", Account: "dan2904"}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1SetAccount, attrsSetAccount, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetAccount: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.SetAccount received: %s", reply)
	}
	attrsAA := &utils.AttrSetActions{ActionsId: "ACTS_1", Actions: []*utils.TPAction{
		{Identifier: utils.MetaDebit, BalanceType: utils.MetaMonetary, Units: "5", ExpiryTime: utils.MetaUnlimited, Weight: 20.0},
		{Identifier: utils.CDRLog},
	}}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	attrsEA := &utils.AttrExecuteAction{Tenant: attrsSetAccount.Tenant, Account: attrsSetAccount.Account, ActionsId: attrsAA.ActionsId}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.ExecuteAction received: %s", reply)
	}
	var rcvedCdrs []*ExternalCDR
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv2GetCDRs, &utils.RPCCDRsFilter{Sources: []string{utils.CDRLog},
		Accounts: []string{attrsSetAccount.Account}}, &rcvedCdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(rcvedCdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(rcvedCdrs))
	} else if rcvedCdrs[0].ToR != utils.MetaMonetary ||
		rcvedCdrs[0].OriginHost != "127.0.0.1" ||
		rcvedCdrs[0].Source != utils.CDRLog ||
		rcvedCdrs[0].RequestType != utils.MetaNone ||
		rcvedCdrs[0].Tenant != "cgrates.org" ||
		rcvedCdrs[0].Account != "dan2904" ||
		rcvedCdrs[0].Subject != "dan2904" ||
		rcvedCdrs[0].Usage != "1" ||
		rcvedCdrs[0].RunID != utils.MetaDebit ||
		strconv.FormatFloat(rcvedCdrs[0].Cost, 'f', -1, 64) != attrsAA.Actions[0].Units {
		t.Errorf("Received: %+v", rcvedCdrs[0])
	}
}

func testActionsitSetCdrlogTopup(t *testing.T) {
	var reply string
	attrsSetAccount := &utils.AttrSetAccount{Tenant: "cgrates.org", Account: "dan2905"}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1SetAccount, attrsSetAccount, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetAccount: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.SetAccount received: %s", reply)
	}
	attrsAA := &utils.AttrSetActions{ActionsId: "ACTS_2", Actions: []*utils.TPAction{
		{Identifier: utils.MetaTopUp, BalanceType: utils.MetaMonetary, Units: "5", ExpiryTime: utils.MetaUnlimited, Weight: 20.0},
		{Identifier: utils.CDRLog},
	}}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	attrsEA := &utils.AttrExecuteAction{Tenant: attrsSetAccount.Tenant, Account: attrsSetAccount.Account, ActionsId: attrsAA.ActionsId}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.ExecuteAction received: %s", reply)
	}
	var rcvedCdrs []*ExternalCDR
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv2GetCDRs, &utils.RPCCDRsFilter{Sources: []string{utils.CDRLog},
		Accounts: []string{attrsSetAccount.Account}}, &rcvedCdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(rcvedCdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(rcvedCdrs))
	} else if rcvedCdrs[0].ToR != utils.MetaMonetary ||
		rcvedCdrs[0].OriginHost != "127.0.0.1" ||
		rcvedCdrs[0].Source != utils.CDRLog ||
		rcvedCdrs[0].RequestType != utils.MetaNone ||
		rcvedCdrs[0].Tenant != "cgrates.org" ||
		rcvedCdrs[0].Account != "dan2905" ||
		rcvedCdrs[0].Subject != "dan2905" ||
		rcvedCdrs[0].Usage != "1" ||
		rcvedCdrs[0].RunID != utils.MetaTopUp ||
		strconv.FormatFloat(rcvedCdrs[0].Cost, 'f', -1, 64) != attrsAA.Actions[0].Units {
		t.Errorf("Received: %+v", rcvedCdrs[0])
	}
}

func testActionsitCdrlogEmpty(t *testing.T) {
	var reply string
	attrsSetAccount := &utils.AttrSetAccount{Tenant: "cgrates.org", Account: "dan2904"}
	attrsAA := &utils.AttrSetActions{ActionsId: "ACTS_3", Actions: []*utils.TPAction{
		{Identifier: utils.MetaDebit, BalanceType: utils.MetaMonetary, DestinationIds: "RET",
			Units: "5", ExpiryTime: utils.MetaUnlimited, Weight: 20.0},
		{Identifier: utils.CDRLog},
	}}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	attrsEA := &utils.AttrExecuteAction{Tenant: attrsSetAccount.Tenant, Account: attrsSetAccount.Account, ActionsId: attrsAA.ActionsId}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.ExecuteAction received: %s", reply)
	}
	var rcvedCdrs []*ExternalCDR
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv2GetCDRs, &utils.RPCCDRsFilter{Sources: []string{utils.CDRLog},
		Accounts: []string{attrsSetAccount.Account}, RunIDs: []string{utils.MetaDebit}}, &rcvedCdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(rcvedCdrs) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(rcvedCdrs))
	} else {
		for _, cdr := range rcvedCdrs {
			if cdr.RunID != utils.MetaDebit {
				t.Errorf("Expecting : MetaDebit, received: %+v", cdr.RunID)
			}
		}
	}
}

func testActionsitCdrlogWithParams(t *testing.T) {
	var reply string
	attrsSetAccount := &utils.AttrSetAccount{Tenant: "cgrates.org", Account: "dan2904"}
	attrsAA := &utils.AttrSetActions{ActionsId: "ACTS_4",
		Actions: []*utils.TPAction{
			{Identifier: utils.MetaDebit, BalanceType: utils.MetaMonetary,
				DestinationIds: "RET", Units: "25", ExpiryTime: utils.MetaUnlimited, Weight: 20.0},
			{Identifier: utils.CDRLog,
				ExtraParameters: `{"RequestType":"*pseudoprepaid","Subject":"DifferentThanAccount", "ToR":"~ActionType:s/^\\*(.*)$/did_$1/"}`},
			{Identifier: utils.MetaDebitReset, BalanceType: utils.MetaMonetary,
				DestinationIds: "RET", Units: "25", ExpiryTime: utils.MetaUnlimited, Weight: 20.0},
		},
	}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	attrsEA := &utils.AttrExecuteAction{Tenant: attrsSetAccount.Tenant, Account: attrsSetAccount.Account, ActionsId: attrsAA.ActionsId}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.ExecuteAction received: %s", reply)
	}
	var rcvedCdrs []*ExternalCDR
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv2GetCDRs, &utils.RPCCDRsFilter{Sources: []string{utils.CDRLog},
		Accounts: []string{attrsSetAccount.Account}, RunIDs: []string{utils.MetaDebit}, RequestTypes: []string{"*pseudoprepaid"}}, &rcvedCdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(rcvedCdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(rcvedCdrs))
	}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv2GetCDRs, &utils.RPCCDRsFilter{Sources: []string{utils.CDRLog},
		Accounts: []string{attrsSetAccount.Account}, RunIDs: []string{utils.MetaDebitReset}, RequestTypes: []string{"*pseudoprepaid"}}, &rcvedCdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(rcvedCdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(rcvedCdrs))
	}
}

func testActionsitCdrlogWithParams2(t *testing.T) {
	var reply string
	attrsSetAccount := &utils.AttrSetAccount{Tenant: "cgrates.org", Account: "dan2904"}
	attrsAA := &utils.AttrSetActions{
		ActionsId: "CustomAction",
		Actions: []*utils.TPAction{
			{Identifier: utils.MetaDebit, BalanceType: utils.MetaMonetary,
				DestinationIds: "RET", Units: "25", ExpiryTime: utils.MetaUnlimited, Weight: 20.0},
			{Identifier: utils.CDRLog,
				ExtraParameters: `{"RequestType":"*pseudoprepaid", "Usage":"10", "Subject":"testActionsitCdrlogWithParams2", "ToR":"~ActionType:s/^\\*(.*)$/did_$1/"}`},
		},
	}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	attrsEA := &utils.AttrExecuteAction{Tenant: attrsSetAccount.Tenant, Account: attrsSetAccount.Account, ActionsId: attrsAA.ActionsId}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.ExecuteAction received: %s", reply)
	}
	var rcvedCdrs []*ExternalCDR
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv2GetCDRs, &utils.RPCCDRsFilter{Sources: []string{utils.CDRLog},
		Accounts: []string{attrsSetAccount.Account}, Subjects: []string{"testActionsitCdrlogWithParams2"}}, &rcvedCdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(rcvedCdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(rcvedCdrs))
	} else if rcvedCdrs[0].Usage != "10" {
		t.Error("Unexpected usege of CDRs returned: ", rcvedCdrs[0].Usage)
	}

}

func testActionsitThresholdCDrLog(t *testing.T) {
	var thReply *ThresholdProfile
	var result string
	var reply string

	attrsSetAccount := &utils.AttrSetAccount{Tenant: "cgrates.org", Account: "th_acc"}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1SetAccount, attrsSetAccount, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetAccount: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.SetAccount received: %s", reply)
	}
	attrsAA := &utils.AttrSetActions{ActionsId: "ACT_TH_CDRLOG", Actions: []*utils.TPAction{
		{Identifier: utils.MetaTopUp, BalanceType: utils.MetaMonetary, Units: "5", ExpiryTime: utils.MetaUnlimited, Weight: 20.0},
		{Identifier: utils.CDRLog},
	}}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	//make sure that the threshold don't exit
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test"}, &thReply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	tPrfl := ThresholdProfileWithAPIOpts{
		ThresholdProfile: &ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "THD_Test",
			FilterIDs: []string{"*string:~*req.Account:th_acc"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   -1,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_TH_CDRLOG"},
			Async:     false,
		},
	}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1SetThresholdProfile, tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test"}, &thReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, thReply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, thReply)
	}
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "cdrev1",
		Event: map[string]any{
			utils.EventType:    utils.CDR,
			"field_extr1":      "val_extr1",
			"fieldextr2":       "valextr2",
			utils.CGRID:        utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
			utils.RunID:        utils.MetaRaw,
			utils.OrderID:      123,
			utils.OriginHost:   "192.168.1.1",
			utils.Source:       utils.UnitTest,
			utils.OriginID:     "dsafdsaf",
			utils.ToR:          utils.MetaVoice,
			utils.RequestType:  utils.MetaRated,
			utils.Tenant:       "cgrates.org",
			utils.Category:     "call",
			utils.AccountField: "th_acc",
			utils.Subject:      "th_acc",
			utils.Destination:  "+4986517174963",
			utils.SetupTime:    time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
			utils.PDD:          0 * time.Second,
			utils.AnswerTime:   time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
			utils.Usage:        10 * time.Second,
			utils.Route:        "SUPPL1",
			utils.Cost:         -1.0,
		},
		APIOpts: map[string]any{
			utils.MetaEventType: utils.CDR,
		},
	}
	var ids []string
	eIDs := []string{"THD_Test"}
	if err := actsLclRpc.Call(context.Background(), utils.ThresholdSv1ProcessEvent, ev, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
	var rcvedCdrs []*ExternalCDR
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv2GetCDRs, &utils.RPCCDRsFilter{Sources: []string{utils.CDRLog},
		Accounts: []string{attrsSetAccount.Account}}, &rcvedCdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(rcvedCdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(rcvedCdrs))
	} else if rcvedCdrs[0].ToR != utils.MetaMonetary ||
		rcvedCdrs[0].OriginHost != "127.0.0.1" ||
		rcvedCdrs[0].Source != utils.CDRLog ||
		rcvedCdrs[0].RequestType != utils.MetaNone ||
		rcvedCdrs[0].Tenant != "cgrates.org" ||
		rcvedCdrs[0].Account != "th_acc" ||
		rcvedCdrs[0].Subject != "th_acc" ||
		rcvedCdrs[0].Usage != "1" ||
		rcvedCdrs[0].RunID != utils.MetaTopUp ||
		strconv.FormatFloat(rcvedCdrs[0].Cost, 'f', -1, 64) != attrsAA.Actions[0].Units {
		t.Errorf("Received: %+v", rcvedCdrs[0])
	}
}

func testActionsitCDRAccount(t *testing.T) {
	var reply string
	acnt := "10023456789"

	// redelareted in function with minimum information to avoid cyclic dependencies
	type AttrAddBalance struct {
		Tenant      string
		Account     string
		BalanceType string
		Value       float64
		Balance     map[string]any
		Overwrite   bool
	}
	attrs := &AttrAddBalance{
		Tenant:      "cgrates.org",
		Account:     acnt,
		BalanceType: utils.MetaVoice,
		Value:       float64(30 * time.Second),
		Balance: map[string]any{
			utils.UUID: "testUUID",
			utils.ID:   "TestID",
		},
		Overwrite: true,
	}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1AddBalance, attrs, &reply); err != nil {
		t.Error("Got error on APIerSv1.AddBalance: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.AddBalance received: %s", reply)
	}

	attrsAA := &utils.AttrSetActions{
		ActionsId: "ACTS_RESET1",
		Actions: []*utils.TPAction{
			{Identifier: utils.MetaCDRAccount, ExpiryTime: utils.MetaUnlimited, Weight: 20.0},
		},
	}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}

	var acc Account
	attrs2 := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: acnt}
	var uuid string
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv2GetAccount, attrs2, &acc); err != nil {
		t.Error("Got error on APIerSv1.GetAccount: ", err.Error())
	} else {
		voice := acc.BalanceMap[utils.MetaVoice]
		for _, u := range voice {
			uuid = u.Uuid
			break
		}
	}

	args := &CDRWithAPIOpts{
		CDR: &CDR{
			Tenant:      "cgrates.org",
			OriginID:    "testDspCDRsProcessCDR",
			OriginHost:  "192.168.1.1",
			Source:      "testDspCDRsProcessCDR",
			RequestType: utils.MetaRated,
			RunID:       utils.MetaDefault,
			PreRated:    true,
			Account:     acnt,
			Subject:     acnt,
			Destination: "1002",
			AnswerTime:  time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
			Usage:       2 * time.Minute,
			CostDetails: &EventCost{
				CGRID: utils.UUIDSha1Prefix(),
				RunID: utils.MetaDefault,
				AccountSummary: &AccountSummary{
					Tenant: "cgrates.org",
					ID:     acnt,
					BalanceSummaries: []*BalanceSummary{
						{
							UUID:  uuid,
							ID:    "TestID",
							Type:  utils.MetaVoice,
							Value: float64(10 * time.Second),
						},
					},
				},
			},
		},
	}
	if err := actsLclRpc.Call(context.Background(), utils.CDRsV1ProcessCDR, args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	time.Sleep(100 * time.Millisecond)

	attrsEA := &utils.AttrExecuteAction{Tenant: "cgrates.org", Account: acnt, ActionsId: attrsAA.ActionsId}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.ExecuteAction received: %s", reply)
	}

	if err := actsLclRpc.Call(context.Background(), utils.APIerSv2GetAccount, attrs2, &acc); err != nil {
		t.Error("Got error on APIerSv1.GetAccount: ", err.Error())
	} else if tv := acc.BalanceMap[utils.MetaVoice].GetTotalValue(); tv != float64(10*time.Second) {
		t.Errorf("Calling APIerSv1.GetBalance expected: %f, received: %f", float64(10*time.Second), tv)
	}
}

func testActionsitThresholdCgrRpcAction(t *testing.T) {
	var thReply *ThresholdProfile
	var result string
	var reply string

	attrsAA := &utils.AttrSetActions{ActionsId: "ACT_TH_CGRRPC", Actions: []*utils.TPAction{
		{Identifier: utils.MetaCgrRpc, ExtraParameters: `{"Address": "127.0.0.1:2012",
"Transport": "*json",
"Method": "RALsV1.Ping",
"Attempts":1,
"Async" :false,
"Params": {}}`}}}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	//make sure that the threshold don't exit
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TH_CGRRPC"}, &thReply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	tPrfl := &ThresholdProfileWithAPIOpts{
		ThresholdProfile: &ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "TH_CGRRPC",
			FilterIDs: []string{"*string:~*req.Method:RALsV1.Ping"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   -1,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_TH_CGRRPC"},
			Async:     false,
		},
	}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1SetThresholdProfile, tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TH_CGRRPC"}, &thReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, thReply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, thReply)
	}
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]any{
			"Method": "RALsV1.Ping",
		},
	}
	var ids []string
	eIDs := []string{"TH_CGRRPC"}
	if err := actsLclRpc.Call(context.Background(), utils.ThresholdSv1ProcessEvent, ev, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}

}

func testActionsitThresholdPostEvent(t *testing.T) {
	var thReply *ThresholdProfile
	var result string
	var reply string

	//if we check syslog we will see that it tries to post
	attrsAA := &utils.AttrSetActions{ActionsId: "ACT_TH_POSTEVENT", Actions: []*utils.TPAction{
		{Identifier: utils.MetaPostEvent, ExtraParameters: "http://127.0.0.1:12080/invalid_json"},
	}}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	//make sure that the threshold don't exit
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_PostEvent"}, &thReply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	tPrfl := &ThresholdProfileWithAPIOpts{
		ThresholdProfile: &ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     "THD_PostEvent",
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   -1,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_TH_POSTEVENT"},
			Async:     false,
		},
	}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1SetThresholdProfile, tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_PostEvent"}, &thReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, thReply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, thReply)
	}
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "cdrev1",
		Event: map[string]any{
			"field_extr1":     "val_extr1",
			"fieldextr2":      "valextr2",
			utils.CGRID:       utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
			utils.RunID:       utils.MetaRaw,
			utils.OrderID:     123,
			utils.OriginHost:  "192.168.1.1",
			utils.Source:      utils.UnitTest,
			utils.OriginID:    "dsafdsaf",
			utils.RequestType: utils.MetaRated,
			utils.Tenant:      "cgrates.org",
			utils.SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
			utils.PDD:         0 * time.Second,
			utils.AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
			utils.Usage:       10 * time.Second,
			utils.Route:       "SUPPL1",
			utils.Cost:        -1.0,
		},
		APIOpts: map[string]any{
			utils.MetaEventType: utils.CDR,
		},
	}
	var ids []string
	eIDs := []string{"THD_PostEvent"}
	if err := actsLclRpc.Call(context.Background(), utils.ThresholdSv1ProcessEvent, ev, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}

}

func testActionsitSetSDestinations(t *testing.T) {
	var reply string
	attrsSetAccount := &utils.AttrSetAccount{
		Tenant:  "cgrates.org",
		Account: "testAccSetDDestination",
	}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1SetAccount, attrsSetAccount, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetAccount: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.SetAccount received: %s", reply)
	}
	attrsAA := &utils.AttrSetActions{ActionsId: "ACT_AddBalance", Actions: []*utils.TPAction{
		{Identifier: utils.MetaTopUp, BalanceType: utils.MetaMonetary, DestinationIds: "*ddc_test",
			Units: "5", ExpiryTime: utils.MetaUnlimited, Weight: 20.0},
	}}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}

	attrsEA := &utils.AttrExecuteAction{Tenant: attrsSetAccount.Tenant, Account: attrsSetAccount.Account, ActionsId: attrsAA.ActionsId}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.ExecuteAction received: %s", reply)
	}

	var acc Account
	attrs2 := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "testAccSetDDestination"}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv2GetAccount, attrs2, &acc); err != nil {
		t.Error(err.Error())
	} else if _, has := acc.BalanceMap[utils.MetaMonetary][0].DestinationIDs["*ddc_test"]; !has {
		t.Errorf("Unexpected destinationIDs: %+v", acc.BalanceMap[utils.MetaMonetary][0].DestinationIDs)
	}

	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1SetDestination,
		&utils.AttrSetDestination{Id: "*ddc_test", Prefixes: []string{"111", "222"}}, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.ExecuteAction received: %s", reply)
	}
	//verify destinations
	var dest Destination
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1GetDestination,
		utils.StringPointer("*ddc_test"), &dest); err != nil {
		t.Error(err.Error())
	} else {
		if len(dest.Prefixes) != 2 || !slices.Contains(dest.Prefixes, "111") || !slices.Contains(dest.Prefixes, "222") {
			t.Errorf("Unexpected destination : %+v", dest)
		}
	}

	// set a StatQueueProfile and simulate process event
	statConfig := &StatQueueProfileWithAPIOpts{
		StatQueueProfile: &StatQueueProfile{
			Tenant:      "cgrates.org",
			ID:          "DistinctMetricProfile",
			QueueLength: 10,
			TTL:         10 * time.Second,
			Metrics: []*MetricWithFilters{{
				MetricID: utils.MetaDDC,
			}},
			ThresholdIDs: []string{utils.MetaNone},
			Stored:       true,
			Weight:       20,
		},
	}

	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1SetStatQueueProfile, statConfig, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	var reply2 []string
	expected := []string{"DistinctMetricProfile"}
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]any{
			utils.Destination: "333",
			utils.Usage:       6 * time.Second,
		},
	}
	if err := actsLclRpc.Call(context.Background(), utils.StatSv1ProcessEvent, &args, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply2, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply2)
	}

	args = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event2",
		Event: map[string]any{
			utils.Destination: "777",
			utils.Usage:       6 * time.Second,
		},
	}
	if err := actsLclRpc.Call(context.Background(), utils.StatSv1ProcessEvent, &args, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply2, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply2)
	}

	//Execute setDDestinations
	attrSetDDest := &utils.AttrSetActions{ActionsId: "ACT_setDDestination", Actions: []*utils.TPAction{
		{Identifier: utils.MetaSetDDestinations, ExtraParameters: "DistinctMetricProfile"},
	}}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv2SetActions, attrSetDDest, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}

	attrsetDDest := &utils.AttrExecuteAction{Tenant: attrsSetAccount.Tenant,
		Account: attrsSetAccount.Account, ActionsId: attrSetDDest.ActionsId}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsetDDest, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.ExecuteAction received: %s", reply)
	}

	//verify destinations
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1GetDestination,
		utils.StringPointer("*ddc_test"), &dest); err != nil {
		t.Error(err.Error())
	} else {
		if len(dest.Prefixes) != 2 || !slices.Contains(dest.Prefixes, "333") || !slices.Contains(dest.Prefixes, "777") {
			t.Errorf("Unexpected destination : %+v", dest)
		}
	}

}

func testActionsitresetAccountCDR(t *testing.T) {
	var reply string
	account := "123456789"

	attrsSetAccount := &utils.AttrSetAccount{
		Tenant:  "cgrates.org",
		Account: "123456789",
	}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1SetAccount, attrsSetAccount, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetAccount: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.SetAccount received: %s", reply)
	}

	attrsAA := &utils.AttrSetActions{
		ActionsId: "resetAccountCDR",
		Actions: []*utils.TPAction{
			{Identifier: utils.MetaCDRAccount, ExpiryTime: utils.MetaUnlimited, Weight: 20.0},
		},
	}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}

	var acc Account
	attrs2 := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: account}
	var uuid string
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv2GetAccount, attrs2, &acc); err != nil {
		t.Error("Got error on APIerSv1.GetAccount: ", err.Error())
	} else {
		voice := acc.BalanceMap[utils.MetaVoice]
		for _, u := range voice {
			uuid = u.Uuid
			break
		}
	}
	args := &CDRWithAPIOpts{
		CDR: &CDR{
			Tenant:      "cgrates.org",
			OriginID:    "testDsp",
			OriginHost:  "192.168.1.1",
			Source:      "testDsp",
			RequestType: utils.MetaRated,
			RunID:       utils.MetaDefault,
			PreRated:    true,
			Account:     account,
			Subject:     account,
			Destination: "1002",
			AnswerTime:  time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
			Usage:       2 * time.Minute,
			CostDetails: &EventCost{
				CGRID: utils.UUIDSha1Prefix(),
				RunID: utils.MetaDefault,
				AccountSummary: &AccountSummary{
					Tenant: "cgrates.org",
					ID:     account,
					BalanceSummaries: []*BalanceSummary{
						{
							UUID:  uuid,
							ID:    "ID",
							Type:  utils.MetaVoice,
							Value: float64(10 * time.Second),
						},
					},
				},
			},
		},
	}
	if err := actsLclRpc.Call(context.Background(), utils.CDRsV1ProcessCDR, args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	time.Sleep(100 * time.Millisecond)

	attrsEA := &utils.AttrExecuteAction{Tenant: "cgrates.org", Account: account, ActionsId: attrsAA.ActionsId}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.ExecuteAction received: %s", reply)
	}

	if err := actsLclRpc.Call(context.Background(), utils.APIerSv2GetAccount, attrs2, &acc); err != nil {
		t.Error("Got error on APIerSv1.GetAccount: ", err.Error())
	} else if tv := acc.BalanceMap[utils.MetaVoice].GetTotalValue(); tv != float64(10*time.Second) {
		t.Errorf("Calling APIerSv1.GetBalance expected: %f, received: %f", float64(10*time.Second), tv)
	}
}

func testActionsitStopCgrEngine(t *testing.T) {
	if err := KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}

func testActionsitremoteSetAccount(t *testing.T) {
	var reply string
	account := "remote1234"
	accID := utils.ConcatenatedKey("cgrates.org", account)
	acc := &Account{
		ID: accID,
	}
	exp := &Account{
		ID: accID,
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: []*Balance{{
				Value:  20,
				Weight: 10,
			}},
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		accStr := utils.ToJSON(acc) + "\n"
		val, err := io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			t.Error(err)
			return
		}
		if string(val) != accStr {
			t.Errorf("Expected %q,received: %q", accStr, string(val))
			return
		}
		rw.Write([]byte(utils.ToJSON(exp)))
	}))

	defer ts.Close()
	attrsAA := &utils.AttrSetActions{
		ActionsId: "remoteSetAccountCDR",
		Actions: []*utils.TPAction{
			{Identifier: utils.MetaRemoteSetAccount, ExtraParameters: ts.URL, Weight: 20.0},
		},
	}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}

	attrsEA := &utils.AttrExecuteAction{Tenant: "cgrates.org", Account: account, ActionsId: attrsAA.ActionsId}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.ExecuteAction received: %s", reply)
	}

	var acc2 Account
	attrs2 := &utils.AttrGetAccount{Account: account}
	if err := actsLclRpc.Call(context.Background(), utils.APIerSv2GetAccount, attrs2, &acc2); err != nil {
		t.Fatal("Got error on APIerSv1.GetAccount: ", err.Error())
	}
	acc2.UpdateTime = exp.UpdateTime
	if utils.ToJSON(exp) != utils.ToJSON(acc2) {
		t.Errorf("Expected: %s,received: %s", utils.ToJSON(exp), utils.ToJSON(acc2))
	}
}
