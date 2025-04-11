//go:build performance

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

package agents

import (
	"bytes"
	"flag"
	"fmt"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/go-diameter/diam"
	"github.com/cgrates/go-diameter/diam/avp"
	"github.com/cgrates/go-diameter/diam/datatype"
	"github.com/cgrates/go-diameter/diam/dict"
)

var (
	capsLimit    = flag.Int("caps_limit", 0, "caps limit")
	capsStrategy = flag.String("caps_strategy", "*busy", "caps strategy")
	parallelism  = flag.Int("parallelism", 0, "parallelism")
)

func BenchmarkDiameterCaps(b *testing.B) {
	// CoreS config is dynamic for this benchmark.
	jsonCfg := fmt.Sprintf(`{
"cores": {
	"caps": %d,
	"caps_strategy": "%s",

	// use shutdown_timeout option to set the diameter reply timeout.
	"shutdown_timeout": "100ms"
}
}`, *capsLimit, *capsStrategy)

	ng := engine.TestEngine{
		ConfigJSON: jsonCfg,
		ConfigPath: filepath.Join(*utils.DataDir, "conf", "samples", "diambench"),
		TpFiles: map[string]string{
			utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
			utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
			utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_sms,*sms,,,,,*unlimited,,1000000,,,,`,
			utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0`,
		},
		DBCfg: engine.InternalDBCfg,
	}
	client, cfg := ng.Run(b)

	time.Sleep(10 * time.Millisecond) // wait for DiameterAgent service to start
	diamClient, err := NewDiameterClient(cfg.DiameterAgentCfg().Listen, "localhost",
		cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
		cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
		cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().ListenNet)
	if err != nil {
		b.Fatal(err)
	}

	// actual benchmark
	var sent, completed atomic.Int64
	b.SetParallelism(*parallelism)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sent.Add(1)
			if sendDiamCCR(b, diamClient, cfg.CoreSCfg().ShutdownTimeout, "2001") {
				completed.Add(1)
			}
		}
	})

	// check results
	b.Logf("sent %d, completed %d", sent.Load(), completed.Load())
	var acnt *engine.Account
	attrsAcnt := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1001",
	}
	expBalance := float64(1000000 - completed.Load())
	if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
		b.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
	} else if rply := acnt.BalanceMap[utils.MetaSMS].GetTotalValue(); rply != expBalance {
		b.Errorf("APIerSv1.GetAccount: sms_balance: %f, want: %f", rply, expBalance)
	}
}

func BenchmarkDiameterReplication(b *testing.B) {
	// Setup for storage engine.
	rplNgJSONCfg := `{
"listen": {
	"rpc_json": "127.0.0.1:22012",
	"rpc_gob": "127.0.0.1:22013",
	"http": "127.0.0.1:22080",
},
"data_db": {
	"db_host": "192.168.122.164"
},
"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
},
"schedulers": {
	"enabled": true
}
}`

	rplNg := engine.TestEngine{
		TpFiles: map[string]string{
			utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
			utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
			utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_sms,*sms,,,,,*unlimited,,1000000,,,,`,
			utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,*string:~*req.Account:1001,,*default,*none,10
cgrates.org,Raw,,,*raw,*constant:*req.RequestType:*none,0`,
		},
		ConfigJSON: rplNgJSONCfg,
		DBCfg: engine.DBCfg{
			DataDB: &engine.DBParams{
				Host: utils.StringPointer("192.168.122.164"),
			},
			StorDB: engine.InternalDBCfg.StorDB,
		},
		LogBuffer: new(bytes.Buffer),
	}
	// defer fmt.Println(rplNg.LogBuffer)
	rplClient, _ := rplNg.Run(b)

	// Setup for rater.
	rplInterval := 5 * time.Millisecond
	failedDir := b.TempDir()
	jsonCfg := fmt.Sprintf(`{
"data_db": {
	"db_type": "*internal",
	"remote_conns": ["db_conn"],
	"replication_conns": ["db_conn"],
	"replication_filtered": false,
	"replication_cache": "",
	"replication_failed_dir": "%s",
	"replication_interval": "%s",
	"items":{
		"*accounts": {"limit": -1, "ttl": "", "static_ttl": false, "remote": true, "replicate": true},
		"*reverse_destinations": {"limit": -1, "ttl": "", "static_ttl": false, "remote": true, "replicate": false},
		"*destinations": {"limit": -1, "ttl": "", "static_ttl": false, "remote": true, "replicate": false},
		"*rating_plans": {"limit": -1, "ttl": "", "static_ttl": false, "remote": true, "replicate": false},
		"*rating_profiles": {"limit": -1, "ttl": "", "static_ttl": false, "remote": true, "replicate": false},
		"*actions": {"limit": -1, "ttl": "", "static_ttl": false, "remote": true, "replicate": false},
		"*action_plans": {"limit": -1, "ttl": "", "static_ttl": false, "remote": true, "replicate": false},
		"*account_action_plans": {"limit": -1, "ttl": "", "static_ttl": false, "remote": true, "replicate": false},
		"*action_triggers": {"limit": -1, "ttl": "", "static_ttl": false, "remote": true, "replicate": false},
		"*shared_groups": {"limit": -1, "ttl": "", "static_ttl": false, "remote": true, "replicate": false},
		"*timings": {"limit": -1, "ttl": "", "static_ttl": false, "remote": true, "replicate": false},
		"*filters": {"limit": -1, "ttl": "", "static_ttl": false, "remote": true, "replicate": false},
		"*attribute_profiles": {"limit": -1, "ttl": "", "static_ttl": false, "remote": true, "replicate": false},
		"*charger_profiles": {"limit": -1, "ttl": "", "static_ttl": false, "remote": true, "replicate": false},
		"*load_ids": {"limit": -1, "ttl": "", "static_ttl": false, "remote": true, "replicate": false},
		"*versions": {"limit": -1, "ttl": "", "static_ttl": false, "remote": true, "replicate": false},
		"*attribute_filter_indexes" : {"limit": -1, "ttl": "", "static_ttl": false, "remote": false, "replicate": false},
		"*charger_filter_indexes" : {"limit": -1, "ttl": "", "static_ttl": false, "remote": true, "replicate": false},
		"*reverse_filter_indexes" : {"limit": -1, "ttl": "", "static_ttl": false, "remote": false, "replicate": false},
	}
},
"rpc_conns": {
	"db_conn": {
		"conns": [
			{
				"address": "127.0.0.1:22013",
				"transport":"*gob",
				"connect_attempts": 5,
				"reconnects": 5,
				"max_reconnect_interval": "",
				"connect_timeout":"1s",
				"reply_timeout":"2s"
			}
		]
	}
}
}`, failedDir, rplInterval)

	ng := engine.TestEngine{
		ConfigJSON: jsonCfg,
		ConfigPath: filepath.Join(*utils.DataDir, "conf", "samples", "diambench"),
		LogBuffer:  new(bytes.Buffer),
	}
	// defer fmt.Println(ng.LogBuffer)
	client, cfg := ng.Run(b)

	time.Sleep(50 * time.Millisecond) // wait for DiameterAgent service to start
	diamClient, err := NewDiameterClient(cfg.DiameterAgentCfg().Listen, "localhost",
		cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
		cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
		cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().ListenNet)
	if err != nil {
		b.Fatal(err)
	}

	// actual benchmark
	var sent, completed atomic.Int64
	b.SetParallelism(*parallelism)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sent.Add(1)
			if sendDiamCCR(b, diamClient, 5*time.Second, "2001") {
				completed.Add(1)
			}
		}
	})

	// Wait for last batch of replications to execute.
	time.Sleep(rplInterval)

	// check results
	b.Logf("sent %d, completed %d", sent.Load(), completed.Load())
	var acnt *engine.Account
	attrsAcnt := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1001",
	}
	expBalance := float64(1000000 - completed.Load())
	if err = client.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
		b.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
	} else if rply := acnt.BalanceMap[utils.MetaSMS].GetTotalValue(); rply != expBalance {
		b.Errorf("APIerSv1.GetAccount: sms_balance: %f, want: %f", rply, expBalance)
	}
	if err = rplClient.Call(context.Background(), utils.APIerSv2GetAccount, attrsAcnt, &acnt); err != nil {
		b.Errorf("APIerSv1.GetAccount unexpected err: %v", err)
	} else if rply := acnt.BalanceMap[utils.MetaSMS].GetTotalValue(); rply != expBalance {
		b.Errorf("APIerSv1.GetAccount: sms_balance: %f, want: %f", rply, expBalance)
	}
}

// sendDiamCCR sends a CCR and verifies the expected result code, returning success status
func sendDiamCCR(tb testing.TB, client *DiameterClient, replyTimeout time.Duration, wantResultCode string) bool {
	tb.Helper()
	ccr := diam.NewRequest(diam.CreditControl, 4, nil)
	ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(utils.UUIDSha1Prefix()))
	ccr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	ccr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	ccr.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("message@DiamItCCRSMS"))
	ccr.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(4))
	ccr.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(0))
	ccr.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2018, 10, 5, 11, 43, 10, 0, time.UTC)))
	ccr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("1001")), // Subscription-Id-Data
		}})
	ccr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(1)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("104502200011")), // Subscription-Id-Data
		}})
	ccr.NewAVP(avp.ServiceIdentifier, avp.Mbit, 0, datatype.Unsigned32(0))
	ccr.NewAVP(avp.RequestedAction, avp.Mbit, 0, datatype.Enumerated(0))
	ccr.NewAVP(avp.RequestedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.CCTime, avp.Mbit, 0, datatype.Unsigned32(1))}})
	ccr.NewAVP(873, avp.Mbit, 10415, &diam.GroupedAVP{ //
		AVP: []*diam.AVP{
			diam.NewAVP(20300, avp.Mbit, 2011, &diam.GroupedAVP{ // IN-Information
				AVP: []*diam.AVP{
					diam.NewAVP(20302, avp.Mbit, 2011, datatype.UTF8String("22509")), // Calling-Vlr-Number
					diam.NewAVP(20385, avp.Mbit, 2011, datatype.UTF8String("4002")),  // Called-Party-NP
				},
			}),
			diam.NewAVP(2000, avp.Mbit, 10415, &diam.GroupedAVP{ // SMS-Information
				AVP: []*diam.AVP{
					diam.NewAVP(886, avp.Mbit, 10415, &diam.GroupedAVP{ // Originator-Address
						AVP: []*diam.AVP{
							diam.NewAVP(899, avp.Mbit, 10415, datatype.Enumerated(1)),      // Address-Type
							diam.NewAVP(897, avp.Mbit, 10415, datatype.UTF8String("1001")), // Address-Data
						}}),
					diam.NewAVP(1201, avp.Mbit, 10415, &diam.GroupedAVP{ // Recipient-Address
						AVP: []*diam.AVP{
							diam.NewAVP(899, avp.Mbit, 10415, datatype.Enumerated(1)),      // Address-Type
							diam.NewAVP(897, avp.Mbit, 10415, datatype.UTF8String("1003")), // Address-Data
						}}),
				},
			}),
		}})

	if err := client.SendMessage(ccr); err != nil {
		tb.Errorf("failed to send diameter message: %v", err)
		return false
	}

	reply := client.ReceivedMessage(replyTimeout)
	if reply == nil {
		tb.Error("received empty reply")
		return false
	}

	avps, err := reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
	if err != nil {
		tb.Error(err)
		return false
	}
	if len(avps) == 0 {
		tb.Error("missing AVPs in reply")
		return false
	}

	resultCode, err := diamAVPAsString(avps[0])
	if err != nil {
		tb.Error(err)
		return false
	}
	if resultCode != wantResultCode {
		tb.Errorf("Result-Code=%s, want %s", resultCode, wantResultCode)
		return false
	}
	return true
}
