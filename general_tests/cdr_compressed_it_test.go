//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestCDRCompressed(t *testing.T) {
	client := newTestClient(t, true)
	t.Run("ProcessAndSetCDR", func(t *testing.T) {
		var reply string
		err := client.Call(context.Background(), utils.CDRsV1ProcessEvent,
			&engine.ArgV1ProcessEvent{
				Flags: []string{utils.MetaRALs},
				CGREvent: utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "event1",
					Event: map[string]any{
						utils.RunID:        "run_1",
						utils.CGRID:        CGRID,
						utils.Tenant:       "cgrates.org",
						utils.Category:     "call",
						utils.ToR:          utils.MetaVoice,
						utils.OriginID:     "processCDR1",
						utils.OriginHost:   "OriginHost1",
						utils.RequestType:  utils.MetaRated,
						utils.AccountField: "1001",
						utils.Destination:  "1002",
						utils.SetupTime:    time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
						utils.AnswerTime:   time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:        2 * time.Minute,
					},
				},
			}, &reply)
		if err != nil {
			t.Fatal(err)
		}
		var cdrs []*engine.CDR
		err = client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
			RPCCDRsFilter: &utils.RPCCDRsFilter{}}, &cdrs)
		if err != nil {
			t.Fatal(err)
		} else if len(cdrs) != 1 {
			t.Errorf("expected the cdrs length to be 1")
		} else if cdrs[0].CostDetails == nil || cdrs[0].Cost != 1.2 || cdrs[0].RunID != "run_1" {
			t.Error("expected CostDetails to be uncompressed correctly")
		}
	})
}

func newTestClient(t testing.TB, compress bool) *birpc.Client {
	// set up DBCfg exactly as in your TestCDRCompressed
	var dbCfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbCfg = engine.DBCfg{StorDB: &engine.DBParams{Type: utils.StringPointer("*internal")}}
	case utils.MetaMySQL:
		dbCfg = engine.DBCfg{StorDB: &engine.DBParams{
			Type:     utils.StringPointer("*mysql"),
			Password: utils.StringPointer("CGRateS.org"),
		}}
	case utils.MetaMongo:
		dbCfg = engine.DBCfg{StorDB: &engine.DBParams{
			Type: utils.StringPointer("*mongo"),
			Name: utils.StringPointer("cgrates"),
			Port: utils.IntPointer(27017),
		}}
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatalf("unsupported dbtype %v", *utils.DBType)
	}

	content := fmt.Sprintf(`{

	"data_db": {								
		"db_type": "*internal"
	},
	
	"stor_db": {
		"db_type": "*internal"
	},
	
	"attributes":{
		"enabled": true,
		"indexed_selects": false,
	},
	
	"rals": {
		"enabled": true,
	},
	
	"cdrs": {
		"enabled": true,
		"rals_conns": ["*internal"],
		"compress_stored_cost": %t,
	},
	
	"schedulers": {
		"enabled": true
	},
	
	"apiers": {
		"enabled": true,
		"scheduler_conns": ["*internal"]
	}
	
	}`, compress)
	cfgJSON := fmt.Sprintf(content, compress)
	ng := engine.TestEngine{
		ConfigJSON: cfgJSON,
		DBCfg:      dbCfg,
		TpPath:     path.Join(*utils.DataDir, "tariffplans", "reratecdrs"),
	}
	client, _ := ng.Run(t)
	return client
}

func benchmarkProcessCDR(b *testing.B, compress bool) {
	b.ReportAllocs()
	client := newTestClient(b, compress)

	for b.Loop() {
		// pre-build the RPC arg
		arg := &engine.ArgV1ProcessEvent{
			Flags: []string{utils.MetaRALs},
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "eventX",
				Event: map[string]any{
					utils.RunID:        "run",
					utils.CGRID:        utils.GenUUID(),
					utils.Tenant:       "cgrates.org",
					utils.Category:     "call",
					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     "bench",
					utils.OriginHost:   "bench-host",
					utils.RequestType:  utils.MetaRated,
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
					utils.AnswerTime:   time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
					utils.Usage:        2 * time.Minute,
				},
			},
		}
		var reply string
		if err := client.Call(context.Background(), utils.CDRsV1ProcessEvent, arg, &reply); err != nil {
			b.Fatalf("ProcessEvent failed: %v", err)
		}
		var cdrs []*engine.CDR
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
			RPCCDRsFilter: &utils.RPCCDRsFilter{}}, &cdrs); err != nil {
			b.Fatalf("GetCDRs failed: %v", err)
		}
	}
}

func BenchmarkCDRCompressed(b *testing.B) {
	benchmarkProcessCDR(b, true)
}

func BenchmarkCDRUncompressed(b *testing.B) {
	benchmarkProcessCDR(b, false)
}
