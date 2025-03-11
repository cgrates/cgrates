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
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestCDRCompressed(t *testing.T) {
	var dbCfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbCfg = engine.DBCfg{
			StorDB: &engine.DBParams{
				Type: utils.StringPointer("*internal"),
			},
		}
	case utils.MetaMySQL:
		dbCfg = engine.DBCfg{
			StorDB: &engine.DBParams{
				Type:     utils.StringPointer("*mysql"),
				Password: utils.StringPointer("CGRateS.org"),
			},
		}
	case utils.MetaMongo:
		dbCfg = engine.DBCfg{
			StorDB: &engine.DBParams{
				Type: utils.StringPointer("*mongo"),
				Name: utils.StringPointer("cgrates"),
				Port: utils.IntPointer(27017),
			},
		}
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	content := `{

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
		"compress_stored_cost":true,
	},
	
	"schedulers": {
		"enabled": true
	},
	
	"apiers": {
		"enabled": true,
		"scheduler_conns": ["*internal"]
	}
	
	}`

	ng := engine.TestEngine{
		ConfigJSON: content,
		DBCfg:      dbCfg,
		TpPath:     path.Join(*utils.DataDir, "tariffplans", "reratecdrs"),
	}
	client, _ := ng.Run(t)
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
