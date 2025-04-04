//go:build call
// +build call

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
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/birpc/jsonrpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func TestFsSessionOfflineInternal(t *testing.T) {
	var cfgPath string
	switch *utils.DBType {
	case utils.MetaInternal:
		cfgPath = path.Join(*utils.DataDir, "conf", "samples", "fs_offline_internal")
	case utils.MetaMySQL:
		cfgPath = path.Join(*utils.DataDir, "conf", "samples", "fs_offline_mysql")
	case utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	var err error
	tutorialCallsCfg, err = config.NewCGRConfigFromPath(cfgPath)
	if err != nil {
		t.Error(err)
	}
	if err = os.MkdirAll(tutorialCallsCfg.DataDbCfg().Opts.InternalDBDumpPath, 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(tutorialCallsCfg.StorDbCfg().Opts.InternalDBDumpPath, 0700); err != nil {
		t.Fatal(err)
	}

	tutorialCallsCfg.DataFolderPath = *utils.DataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tutorialCallsCfg)
	if err := engine.PreInitDataDb(tutorialCallsCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.PreInitStorDb(tutorialCallsCfg); err != nil {
		t.Fatal(err)
	}
	engine.KillProcName(utils.Freeswitch, 5000)
	if err := engine.CallScript(path.Join(*fsConfig, "freeswitch", "etc", "init.d", "freeswitch"), "start", 3000); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StopStartEngine(cfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond)
	tutorialCallsRpc, err = jsonrpc.Dial(utils.TCP, tutorialCallsCfg.ListenCfg().RPCJSONListen)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("LoadTPs", func(t *testing.T) {
		var reply string
		attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "tutorial")}
		if err := tutorialCallsRpc.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
			t.Error(err)
		}
		time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond) // Give time for scheduler to execute topups
	})

	t.Run("AccountsBefore", func(t *testing.T) {
		var reply *engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := tutorialCallsRpc.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &reply); err != nil {
			t.Error("Got error on APIerSv2.GetAccount: ", err.Error())
		} else if reply.BalanceMap[utils.MetaMonetary].GetTotalValue() != 10.0 {
			t.Errorf("Calling APIerSv1.GetBalance received: %f", reply.BalanceMap[utils.MetaMonetary].GetTotalValue())
		}
		var reply2 *engine.Account
		attrs2 := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1002"}
		if err := tutorialCallsRpc.Call(context.Background(), utils.APIerSv2GetAccount, attrs2, &reply2); err != nil {
			t.Error("Got error on APIerSv2.GetAccount: ", err.Error())
		} else if reply2.BalanceMap[utils.MetaMonetary].GetTotalValue() != 10.0 {
			t.Errorf("Calling APIerSv1.GetBalance received: %f", reply2.BalanceMap[utils.MetaMonetary].GetTotalValue())
		}
		var reply3 *engine.Account
		attrs3 := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1003"}
		if err := tutorialCallsRpc.Call(context.Background(), utils.APIerSv2GetAccount, attrs3, &reply3); err != nil {
			t.Error("Got error on APIerSv2.GetAccount: ", err.Error())
		} else if reply3.BalanceMap[utils.MetaMonetary].GetTotalValue() != 10.0 {
			t.Errorf("Calling APIerSv1.GetBalance received: %f", reply3.BalanceMap[utils.MetaMonetary].GetTotalValue())
		}
	})

	t.Run("StatMetricsBefore", func(t *testing.T) {
		var metrics map[string]string
		expectedMetrics := map[string]string{
			utils.MetaTCC: utils.NotAvailable,
			utils.MetaTCD: utils.NotAvailable,
		}
		if err := tutorialCallsRpc.Call(context.Background(), utils.StatSv1GetQueueStringMetrics,
			&utils.TenantID{Tenant: "cgrates.org", ID: "Stats2"}, &metrics); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expectedMetrics, metrics) {
			t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
		}
		if err := tutorialCallsRpc.Call(context.Background(), utils.StatSv1GetQueueStringMetrics,
			&utils.TenantID{Tenant: "cgrates.org", ID: "Stats2_1"}, &metrics); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expectedMetrics, metrics) {
			t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
		}
	})

	t.Run("CheckResourceBeforeAllocation", func(t *testing.T) {
		var rs *engine.Resources
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ResourceEvent",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
			},
			APIOpts: map[string]any{
				utils.OptsResourcesUsageID: "OriginID",
			},
		}
		if err := tutorialCallsRpc.Call(context.Background(), utils.ResourceSv1GetResourcesForEvent, args, &rs); err != nil {
			t.Fatal(err)
		} else if len(*rs) != 1 {
			t.Fatalf("Resources: %+v", utils.ToJSON(rs))
		}
		for _, r := range *rs {
			if r.ID == "ResGroup1" &&
				(len(r.Usages) != 0 || len(r.TTLIdx) != 0) {
				t.Errorf("Unexpected resource: %+v", utils.ToJSON(r))
			}
		}
	})

	t.Run("CheckThresholdBefore", func(t *testing.T) {
		var td engine.Threshold
		eTd := engine.Threshold{Tenant: "cgrates.org", ID: "THD_ACNT_1001", Hits: 0}
		if err := tutorialCallsRpc.Call(context.Background(), utils.ThresholdSv1GetThreshold,
			&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1001"}, &td); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(eTd, td) {
			t.Errorf("expecting: %+v, received: %+v", eTd, td)
		}
		eTd = engine.Threshold{Tenant: "cgrates.org", ID: "THD_ACNT_1002", Hits: 0}
		if err := tutorialCallsRpc.Call(context.Background(), utils.ThresholdSv1GetThreshold,
			&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1002"}, &td); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(eTd, td) {
			t.Errorf("expecting: %+v, received: %+v", eTd, td)
		}
	})

	t.Run("StartPjsuaListener", func(t *testing.T) {
		var err error
		acnts := []*engine.PjsuaAccount{
			{Id: "sip:1001@127.0.0.1",
				Username: "1001", Password: "CGRateS.org", Realm: "*", Registrar: "sip:127.0.0.1:5080"},
			{Id: "sip:1002@127.0.0.1",
				Username: "1002", Password: "CGRateS.org", Realm: "*", Registrar: "sip:127.0.0.1:5080"},
			{Id: "sip:1003@127.0.0.1",
				Username: "1003", Password: "CGRateS.org", Realm: "*", Registrar: "sip:127.0.0.1:5080"},
		}
		if tutorialCallsPjSuaListener, err = engine.StartPjsuaListener(
			acnts, 5070, time.Duration(*utils.WaitRater)*time.Millisecond); err != nil {
			t.Fatal(err)
		}
		time.Sleep(3 * time.Second)
	})

	t.Run("Call1001To1002", func(t *testing.T) {
		if err := engine.PjsuaCallUri(
			&engine.PjsuaAccount{Id: "sip:1001@127.0.0.1", Username: "1001", Password: "CGRateS.org", Realm: "*"},
			"sip:1002@127.0.0.1", "sip:127.0.0.1:5080", 67*time.Second, 5071); err != nil {
			t.Fatal(err)
		}
		// give time to session to start so we can check it
		time.Sleep(time.Second)
	})

	t.Run("GetActiveSessions", func(t *testing.T) {
		var reply *[]*sessions.ExternalSession
		expected := &[]*sessions.ExternalSession{
			{
				RequestType: "*prepaid",
				Tenant:      "cgrates.org",
				Category:    "call",
				Account:     "1001",
				Subject:     "1001",
				Destination: "1002",
			},
		}
		if err := tutorialCallsRpc.Call(context.Background(), utils.SessionSv1GetActiveSessions,
			nil, &reply); err != nil {
			t.Error("Got error on SessionSv1.GetActiveSessions: ", err.Error())
		} else {
			if len(*reply) == 2 {
				sort.Slice(*reply, func(i, j int) bool {
					return strings.Compare((*reply)[i].RequestType, (*reply)[j].RequestType) > 0
				})
			}
			// compare some fields (eg. CGRId is generated)
			if !reflect.DeepEqual((*expected)[0].RequestType, (*reply)[0].RequestType) {
				t.Errorf("Expected: %s, received: %s", (*expected)[0].RequestType, (*reply)[0].RequestType)
			} else if !reflect.DeepEqual((*expected)[0].Account, (*reply)[0].Account) {
				t.Errorf("Expected: %s, received: %s", (*expected)[0].Account, (*reply)[0].Account)
			} else if !reflect.DeepEqual((*expected)[0].Destination, (*reply)[0].Destination) {
				t.Errorf("Expected: %s, received: %s", (*expected)[0].Destination, (*reply)[0].Destination)
			}
		}
	})

	t.Run("CutEngineMidCall", func(t *testing.T) {
		if err := engine.KillEngine(1000); err != nil {
			t.Error(err)
		}
		// simulate connection down
		time.Sleep(10 * time.Second)
	})

	t.Run("StartEngineMidCall", func(t *testing.T) {
		if _, err := engine.StopStartEngine(cfgPath, *utils.WaitRater); err != nil {
			t.Fatal(err)
		}
		tutorialCallsRpc, err = jsonrpc.Dial(utils.TCP, tutorialCallsCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(23 * time.Second)
	})

	t.Run("CheckResourceAllocationAndWaitForCallToFinish", func(t *testing.T) {
		var rs engine.Resources
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ResourceAllocation",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
			},
			APIOpts: map[string]any{
				utils.OptsResourcesUsageID: "OriginID1",
			},
		}
		if err := tutorialCallsRpc.Call(context.Background(), utils.ResourceSv1GetResourcesForEvent, args, &rs); err != nil {
			t.Fatal(err)
		} else if len(rs) != 1 {
			t.Fatalf("Resources: %+v", utils.ToJSON(rs))
		}
		if rs[0].ID != "ResGroup1" && len(rs[0].Usages) != 0 {
			t.Errorf("Unexpected resource: %+v", utils.ToJSON(rs[0]))
		}
		// Allow calls to finish before start querying the results
		time.Sleep(50 * time.Second)
	})

	t.Run("GetAccount1001", func(t *testing.T) {
		var reply *engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := tutorialCallsRpc.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &reply); err != nil {
			t.Error(err.Error())
		} else if reply.BalanceMap[utils.MetaMonetary].GetTotalValue() == 10.0 { // Make sure we debitted
			t.Errorf("Expected: 10, received: %+v", reply.BalanceMap[utils.MetaMonetary].GetTotalValue())
		} else if reply.Disabled == true {
			t.Error("Account disabled")
		}
	})

	t.Run("GetCDRs", func(t *testing.T) {
		var reply []*engine.ExternalCDR
		req := utils.RPCCDRsFilter{RunIDs: []string{utils.MetaDefault}, Accounts: []string{"1001"}}
		if err := tutorialCallsRpc.Call(context.Background(), utils.APIerSv2GetCDRs, &req, &reply); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if len(reply) != 1 {
			t.Error("Unexpected number of CDRs returned: ", len(reply))
		} else {
			for _, cdr := range reply {
				if cdr.RequestType != utils.MetaPrepaid {
					t.Errorf("Unexpected RequestType for CDR: %+v", cdr.RequestType)
				}
				if cdr.Destination == "1002" {
					// in case of Asterisk take the integer part from usage
					if optConf == utils.Asterisk {
						cdr.Usage = strings.Split(cdr.Usage, ".")[0] + "s"
					}
					if cdr.Usage != "1m7s" && cdr.Usage != "1m8s" { // Usage as seconds
						t.Errorf("Unexpected Usage for CDR: %+v", cdr.Usage)
					}
					if cdr.CostSource != utils.MetaSessionS {
						t.Errorf("Unexpected CostSource for CDR: %+v", cdr.CostSource)
					}
				} else if cdr.Destination == "1003" {
					// in case of Asterisk take the integer part from usage
					if optConf == utils.Asterisk {
						cdr.Usage = strings.Split(cdr.Usage, ".")[0] + "s"
					}
					if cdr.Usage != "12s" && cdr.Usage != "13s" { // Usage as seconds
						t.Errorf("Unexpected Usage for CDR: %+v", cdr.Usage)
					}
					if cdr.CostSource != utils.MetaSessionS {
						t.Errorf("Unexpected CostSource for CDR: %+v", cdr.CostSource)
					}
				}
			}
		}
	})

	t.Run("StatMetrics", func(t *testing.T) {
		var metrics map[string]string
		firstStatMetrics1 := map[string]string{
			utils.MetaTCC: "0.6117",
			utils.MetaTCD: "1m7s",
		}
		firstStatMetrics2 := map[string]string{
			utils.MetaTCC: "1.35009",
			utils.MetaTCD: "2m25s",
		}
		firstStatMetrics3 := map[string]string{
			utils.MetaTCC: "1.34009",
			utils.MetaTCD: "2m24s",
		}
		firstStatMetrics4 := map[string]string{
			utils.MetaTCC: "1.35346",
			utils.MetaTCD: "2m24s",
		}

		if err := tutorialCallsRpc.Call(context.Background(), utils.StatSv1GetQueueStringMetrics,
			&utils.TenantID{Tenant: "cgrates.org", ID: "Stats2"}, &metrics); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(firstStatMetrics1, metrics) &&
			!reflect.DeepEqual(firstStatMetrics2, metrics) &&
			!reflect.DeepEqual(firstStatMetrics3, metrics) &&
			!reflect.DeepEqual(firstStatMetrics4, metrics) {
			t.Errorf("expecting: %+v, received reply: %s", firstStatMetrics1, metrics)
		}
		if err := tutorialCallsRpc.Call(context.Background(), utils.StatSv1GetQueueStringMetrics,
			&utils.TenantID{Tenant: "cgrates.org", ID: "Stats2_1"}, &metrics); err != nil {
			t.Error(err)
		}
	})

	t.Run("CheckResourceRelease", func(t *testing.T) {
		var rs *engine.Resources
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ResourceRelease",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
			},
			APIOpts: map[string]any{
				utils.OptsResourcesUsageID: "OriginID2",
			},
		}
		if err := tutorialCallsRpc.Call(context.Background(), utils.ResourceSv1GetResourcesForEvent, args, &rs); err != nil {
			t.Fatal(err)
		} else if len(*rs) != 1 {
			t.Fatalf("Resources: %+v", rs)
		}
		for _, r := range *rs {
			if r.ID == "ResGroup1" && len(r.Usages) != 0 {
				t.Errorf("Unexpected resource: %+v", utils.ToJSON(r))
			}
		}
	})

	t.Run("CheckThreshold1001After", func(t *testing.T) {
		var td engine.Threshold
		if err := tutorialCallsRpc.Call(context.Background(), utils.ThresholdSv1GetThreshold,
			&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1001"}, &td); err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			t.Error(err)
		}
	})

	t.Run("CheckThreshold1002After", func(t *testing.T) {
		var td engine.Threshold
		eTd := engine.Threshold{Tenant: "cgrates.org", ID: "THD_ACNT_1002", Hits: 0}
		if err := tutorialCallsRpc.Call(context.Background(), utils.ThresholdSv1GetThreshold,
			&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1002"}, &td); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(eTd.Tenant, td.Tenant) {
			t.Errorf("expecting: %+v, received: %+v", eTd.Tenant, td.Tenant)
		} else if !reflect.DeepEqual(eTd.ID, td.ID) {
			t.Errorf("expecting: %+v, received: %+v", eTd.ID, td.ID)
		} else if !reflect.DeepEqual(eTd.Hits, td.Hits) {
			t.Errorf("expecting: %+v, received: %+v", eTd.Hits, td.Hits)
		}
	})

	t.Run("StopPjsuaListener", func(t *testing.T) {
		tutorialCallsPjSuaListener.Write([]byte("q\n")) // Close pjsua
		time.Sleep(time.Second)                         // Allow pjsua to finish it's tasks, eg un-REGISTER
	})

	t.Run("StopCgrEngine", func(t *testing.T) {
		if err := engine.KillEngine(100); err != nil {
			t.Error(err)
		}
	})

	t.Run("StopFS", func(t *testing.T) {
		engine.ForceKillProcName(utils.Freeswitch, 1000)
	})
	if err := os.RemoveAll("/var/lib/cgrates/internal_db"); err != nil {
		t.Error(err)
	}
}
