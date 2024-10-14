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
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/nats-io/nats.go"
)

// TestEEsExportEventChanges tests if the event that's about to be exported can be changed from one exporter to
// another through AttributeS. Additionally (unrelated to the case presented in the previous sentence), it also
// checks that *opts.exporterID field is being set correctly.
//
// The test steps are as follows:
//  1. Configure inside ees two event exporters. First one has the *attribute flag, while the second doesn't.
//  2. Set an Attribute without filter that changes the request type to *prepaid.
//  3. Send an event with CGRID and RequestType *rated in a request to EeSv1.ProcessEvent.
//  4. Verify that first export has RequestType *prepaid, while the second one has *rated.
func TestEEsExportEventChanges(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	content := `{

"general": {
	"log_level": 7
},

"data_db": {								
	"db_type": "*internal"
},

"stor_db": {
	"db_type": "*internal"
},

"apiers": {
	"enabled": true
},

"attributes": {
	"enabled": true
},

"ees": {
	"enabled": true,
	"attributes_conns":["*localhost"],
	"exporters": [
		{
			"id": "exporter1",
			"type": "*virt",
			"flags": ["*attributes"],
			"attempts": 1,
			"synchronous": true,
			"fields":[
				{"tag": "CGRID", "path": "*uch.CGRID1", "type": "*variable", "value": "~*req.CGRID"},
				{"tag": "RequestType", "path": "*uch.RequestType1", "type": "*variable", "value": "~*req.RequestType"},
				{"tag": "BalanceID", "path": "*uch.BalanceID", "type": "*variable", "value": "~*req.CostDetails.Charges[0].Increments[0].Accounting.Balance.ID"},
				{"tag": "BalanceType", "path": "*uch.BalanceType", "type": "*variable", "value": "~*req.CostDetails.Charges[0].Increments[0].Accounting.Balance.Type"},
				{"tag": "BalanceFound", "path": "*uch.BalanceFound", "type": "*variable", "value": "~*req.BalanceFound"},
				{"tag": "ExporterID", "path": "*uch.ExporterID1", "type": "*variable", "value": "~*opts.*exporterID"},
				{"tag": "ChangedValue", "path": "*uch.ChangedValue", "type": "*variable", "value": "~*req.CostDetails.Charges[0].Increments[0].Accounting.Balance.Value"}
			],
		},
		{
			"id": "exporter2",
			"type": "*virt",
			"flags": [],
			"attempts": 1,
			"synchronous": true,
			"fields":[
				{"tag": "CGRID", "path": "*uch.CGRID2", "type": "*variable", "value": "~*req.CGRID"},
				{"tag": "RequestType", "path": "*uch.RequestType2", "type": "*variable", "value": "~*req.RequestType"},
				{"tag": "ExporterID", "path": "*uch.ExporterID2", "type": "*variable", "value": "~*opts.*exporterID"}
			]
		}
	]
}

}`

	ng := engine.TestEngine{
		ConfigJSON: content,
	}
	client, _ := ng.Run(t)

	t.Run("SetAttributeProfile", func(t *testing.T) {
		attrPrf := &engine.AttributeProfileWithAPIOpts{
			AttributeProfile: &engine.AttributeProfile{
				Tenant: "cgrates.org",
				ID:     "ATTR_TEST",
				Attributes: []*engine.Attribute{
					{
						Path: "*req.RequestType",
						Type: utils.MetaVariable,
						Value: config.RSRParsers{
							&config.RSRParser{
								Rules: utils.MetaPrepaid,
							},
						},
					},

					// This attribute will change BALANCE_TEST Value from 10 to 11
					{
						Path: "*req.CostDetails",
						Type: utils.MetaVariable,
						Value: config.RSRParsers{
							&config.RSRParser{
								Rules: `{"CGRID":"","RunID":"","StartTime":"0001-01-01T00:00:00Z","Usage":null,"Cost":null,"Charges":[{"RatingID":"","Increments":[{"Usage":0,"Cost":0,"AccountingID":"ACCOUNTING_TEST","CompressFactor":0}],"CompressFactor":0}],"AccountSummary":{"Tenant":"","ID":"","BalanceSummaries":[{"UUID":"123456","ID":"BALANCE_TEST","Type":"*voice","Initial":0,"Value":11,"Disabled":false}],"AllowNegative":false,"Disabled":false},"Rating":null,"Accounting":{"ACCOUNTING_TEST":{"AccountID":"","BalanceUUID":"123456","RatingID":"","Units":0,"ExtraChargeID":""}},"RatingFilters":null,"Rates":null,"Timings":null}`,
							},
						},
					},
					{
						FilterIDs: []string{
							"*string:~*req.CostDetails.Charges[0].Increments[0].Accounting.Balance.ID:BALANCE_TEST",
							"*string:~*req.CostDetails.Charges[0].Increments[0].Accounting.Balance.Type:*voice",
							"*string:~*req.AccountSummary.BalanceSummaries.BALANCE_TEST.ID:BALANCE_TEST",
							"*string:~*req.AccountSummary.BalanceSummaries.BALANCE_TEST.Type:*voice",
						},
						Path: "*req.BalanceFound",
						Type: utils.MetaVariable,
						Value: config.RSRParsers{
							&config.RSRParser{
								Rules: utils.TrueStr,
							},
						},
					},
				},
				Blocker: false,
				Weight:  10,
			},
		}
		attrPrf.Compile()
		var result string
		if err := client.Call(context.Background(), utils.APIerSv1SetAttributeProfile, attrPrf, &result); err != nil {
			t.Error(err)
		}
	})

	t.Run("ExportEvent", func(t *testing.T) {
		eventToExport := &engine.CGREventWithEeIDs{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "voiceEvent",
				Time:   utils.TimePointer(time.Now()),
				Event: map[string]any{
					utils.CGRID:       "TEST",
					utils.RequestType: utils.MetaRated,
					utils.CostDetails: &engine.EventCost{
						Charges: []*engine.ChargingInterval{
							{
								Increments: []*engine.ChargingIncrement{
									{
										AccountingID: "ACCOUNTING_TEST",
									},
								},
							},
						},
						Accounting: engine.Accounting{
							"ACCOUNTING_TEST": &engine.BalanceCharge{
								BalanceUUID: "123456",
							},
						},
						AccountSummary: &engine.AccountSummary{
							BalanceSummaries: engine.BalanceSummaries{
								{
									ID:    "BALANCE_TEST",
									Type:  utils.MetaVoice,
									UUID:  "123456",
									Value: 10,
								},
							},
						},
					},
					utils.AccountSummary: &engine.AccountSummary{
						BalanceSummaries: engine.BalanceSummaries{
							{
								ID:    "BALANCE_TEST",
								Type:  utils.MetaVoice,
								UUID:  "123456",
								Value: 10,
							},
						},
					},
				},
			},
		}

		var reply map[string]map[string]any
		if err := client.Call(context.Background(), utils.EeSv1ProcessEvent, eventToExport, &reply); err != nil {
			t.Error(err)
		}

		var err error
		var requestTypeExport1 any
		if err = client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
			Tenant: "cgrates.org",
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheUCH,
				ItemID:  "RequestType1",
			},
		}, &requestTypeExport1); err != nil {
			t.Error(err)
		} else if requestTypeExport1 != utils.MetaPrepaid {
			t.Errorf("expected %v, received %v", utils.MetaPrepaid, requestTypeExport1)
		}

		var requestTypeExport2 any
		if err = client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
			Tenant: "cgrates.org",
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheUCH,
				ItemID:  "RequestType2",
			},
		}, &requestTypeExport2); err != nil {
			t.Error(err)
		} else if requestTypeExport2 != utils.MetaRated {
			t.Errorf("expected %v, received %v", utils.MetaRated, requestTypeExport2)
		}
	})

	t.Run("CheckExporterIDs", func(t *testing.T) {
		var exporterID1 any
		if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
			Tenant: "cgrates.org",
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheUCH,
				ItemID:  "ExporterID1",
			},
		}, &exporterID1); err != nil {
			t.Error(err)
		} else if exporterID1 != "exporter1" {
			t.Errorf("expected %v, received %v", "exporter1", exporterID1)
		}

		var exporterID2 any
		if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
			Tenant: "cgrates.org",
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheUCH,
				ItemID:  "ExporterID2",
			},
		}, &exporterID2); err != nil {
			t.Error(err)
		} else if exporterID2 != "exporter2" {
			t.Errorf("expected %v, received %v", "exporter2", exporterID2)
		}
	})

	t.Run("CheckAttributesAlteredFields", func(t *testing.T) {
		var changedValue any
		if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
			Tenant: "cgrates.org",
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheUCH,
				ItemID:  "ChangedValue",
			},
		}, &changedValue); err != nil {
			t.Error(err)
		} else if changedValue != "11" {
			t.Errorf("expected %v, received %v", "11", changedValue)
		}
		var balanceID any
		if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
			Tenant: "cgrates.org",
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheUCH,
				ItemID:  "BalanceID",
			},
		}, &balanceID); err != nil {
			t.Error(err)
		} else if balanceID != "BALANCE_TEST" {
			t.Errorf("expected %v, received %v", "BALANCE_TEST", balanceID)
		}
		var balanceType any
		if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
			Tenant: "cgrates.org",
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheUCH,
				ItemID:  "BalanceType",
			},
		}, &balanceType); err != nil {
			t.Error(err)
		} else if balanceType != utils.MetaVoice {
			t.Errorf("expected %v, received %v", "BALANCE_TEST", balanceType)
		}
		var balanceFound any
		if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
			Tenant: "cgrates.org",
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheUCH,
				ItemID:  "BalanceFound",
			},
		}, &balanceFound); err != nil {
			t.Error(err)
		} else if balanceFound != "true" {
			t.Errorf("expected %v, received %v", "true", balanceFound)
		}
	})
}

// TestEEsReplayFailedPosts tests the implementation of the APIerSv1.ReplayFailedPosts.
func TestEEsReplayFailedPosts(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	failedDir := t.TempDir()

	content := fmt.Sprintf(`{

"general": {
	"log_level": 7,
	"failed_posts_ttl": "1ms",
	"poster_attempts": 1
},

"data_db": {								
	"db_type": "*internal"
},

"stor_db": {
	"db_type": "*internal"
},

"apiers": {
	"enabled": true
},

"ees": {
	"enabled": true,
	"exporters": [
		{
			"id": "nats_exporter",										
			"type": "*nats_json_map",	
			"flags": ["*log"],												
			"export_path": "nats://localhost:4222",		
			"failed_posts_dir": "%s",
			"attempts": 1,
			"synchronous": true,	
			"opts": {
				"natsSubject": "processed_cdrs",
			},
			"fields":[
				{"tag": "CGRID", "path": "*exp.CGRID", "type": "*variable", "value": "~*req.CGRID"},
			]
		}
	]
}

}`, failedDir)

	ng := engine.TestEngine{
		ConfigJSON: content,
	}
	client, _ := ng.Run(t)

	// helper to sort slices
	less := func(a, b string) bool { return a < b }

	// amount of events to export/replay
	count := 5

	t.Run("successful nats export", func(t *testing.T) {
		cmd := exec.Command("nats-server")
		if err := cmd.Start(); err != nil {
			t.Fatalf("failed to start nats-server: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
		defer cmd.Process.Kill()

		nc, err := nats.Connect("nats://localhost:4222", nats.Timeout(time.Second), nats.DrainTimeout(time.Second))
		if err != nil {
			t.Fatalf("failed to connect to nats-server: %v", err)
		}
		defer nc.Drain()

		ch := make(chan *nats.Msg, count)
		sub, err := nc.ChanQueueSubscribe("processed_cdrs", "", ch)
		if err != nil {
			t.Fatalf("failed to subscribe to nats queue: %v", err)
		}

		var reply map[string]map[string]any
		for i := range count {
			if err := client.Call(context.Background(), utils.EeSv1ProcessEvent, &engine.CGREventWithEeIDs{
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "test",
					Event: map[string]any{
						utils.CGRID: i,
					},
				},
			}, &reply); err != nil {
				t.Errorf("EeSv1.ProcessEvent returned unexpected err: %v", err)
			}
		}

		time.Sleep(1 * time.Millisecond) // wait for the channel to receive the replayed exports
		want := make([]string, 0, count)
		for i := range count {
			want = append(want, fmt.Sprintf(`{"CGRID":"%d"}`, i))
		}
		if err := sub.Unsubscribe(); err != nil {
			t.Errorf("failed to unsubscribe from nats subject: %v", err)
		}
		close(ch)
		got := make([]string, 0, count)
		for elem := range ch {
			got = append(got, string(elem.Data))
		}
		if diff := cmp.Diff(want, got, cmpopts.SortSlices(less)); diff != "" {
			t.Errorf("unexpected nats messages received over channel (-want +got): \n%s", diff)
		}
	})

	t.Run("replay failed nats export", func(t *testing.T) {
		var exportReply map[string]map[string]any
		for i := range count {
			err := client.Call(context.Background(), utils.EeSv1ProcessEvent,
				&engine.CGREventWithEeIDs{
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     "test",
						Event: map[string]any{
							utils.CGRID: i,
						},
					},
				}, &exportReply)
			if err == nil || err.Error() != utils.ErrPartiallyExecuted.Error() {
				t.Errorf("EeSv1.ProcessEvent err = %v, want %v", err, utils.ErrPartiallyExecuted)
			}
		}

		time.Sleep(5 * time.Millisecond)
		var replayReply string
		if err := client.Call(context.Background(), utils.APIerSv1ReplayFailedPosts, v1.ReplayFailedPostsParams{
			SourcePath: failedDir,
			Modules:    []string{"test", "EEs"},
		}, &replayReply); err != nil {
			t.Errorf("APIerSv1.ReplayFailedPosts returned unexpected err: %v", err)
		}

		cmd := exec.Command("nats-server")
		if err := cmd.Start(); err != nil {
			t.Fatalf("failed to start nats-server: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
		defer cmd.Process.Kill()

		nc, err := nats.Connect("nats://localhost:4222", nats.Timeout(time.Second), nats.DrainTimeout(time.Second))
		if err != nil {
			t.Fatalf("failed to connect to nats-server: %v", err)
		}
		defer nc.Drain()

		ch := make(chan *nats.Msg, count)
		sub, err := nc.ChanQueueSubscribe("processed_cdrs", "", ch)
		if err != nil {
			t.Fatalf("failed to subscribe to nats queue: %v", err)
		}

		if err := client.Call(context.Background(), utils.APIerSv1ReplayFailedPosts, v1.ReplayFailedPostsParams{
			SourcePath: failedDir,
			Modules:    []string{"test", "EEs"},
		}, &replayReply); err != nil {
			t.Errorf("APIerSv1.ReplayFailedPosts returned unexpected err: %v", err)
		}
		time.Sleep(time.Millisecond) // wait for the channel to receive the replayed exports

		want := make([]string, 0, count)
		for i := range count {
			want = append(want, fmt.Sprintf(`{"CGRID":"%d"}`, i))
		}
		if err := sub.Unsubscribe(); err != nil {
			t.Errorf("failed to unsubscribe from nats subject: %v", err)
		}
		close(ch)
		got := make([]string, 0, count)
		for elem := range ch {
			got = append(got, string(elem.Data))
		}
		if diff := cmp.Diff(want, got, cmpopts.SortSlices(less)); diff != "" {
			t.Errorf("unexpected nats messages received over channel (-want +got): \n%s", diff)
		}
	})
}
