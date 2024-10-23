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

package cdrs

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/utils"
)

func TestCDRsIT(t *testing.T) {
	tpPath := "/tmp/tps/cdrs/TestCDRsProcessEvent"
	csvFiles := map[string]string{

		// How to disable charging for a specific event?
		utils.ChargersCsv: `#Tenant,ID,FilterIDs,Weights,Blockers,RunID,AttributeIDs
#cgrates.org,Raw,,;20,,raw,*constant:*req.RequestType:*none
cgrates.org,CustomerCharges,,;20,,CustomerCharges,*none`,

		utils.RatesCsv: `#Tenant,ID,FilterIDs,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement
cgrates.org,DEFAULT_RATE,,;0,0,0,*free,RT_ALWAYS,,"* * * * *",;0,false,0s,,0.1,1s,1s`,
	}

	err := os.MkdirAll(tpPath, 0755)
	if err != nil {
		t.Fatalf("could not create folder %s: %v", tpPath, err)
	}
	defer os.RemoveAll(tpPath)

	for fileName, content := range csvFiles {
		filePath := path.Join(tpPath, fileName)
		err = os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("could not write to file %s: %v", filePath, err)
		}
	}

	cfgContent := `{

"logger": {
	"level": 7
},

"data_db": {
	"db_type": "*internal"
},

"stor_db": {
	"db_type": "*internal",
	"string_indexed_fields": ["RunID"]
},

"rates": {
	"enabled": true
},

"cdrs": {
	"enabled": true,
	"attributes_conns":["*internal"],
	"chargers_conns":["*localhost"],
	"rates_conns": ["*localhost"]
},

"attributes": {
	"enabled": true
},

"chargers": {
	"enabled": true,
	"attributes_conns": ["*localhost"]
},

"admins": {
	"enabled": true
},

"loaders": [
	{
		"id": "*default",
		"enabled": true,
		"tenant": "cgrates.org",
		"lockfile_path": ".cgr.lck",
		"tp_in_dir": "%s",
		"tp_out_dir": ""
	}
]

}
`
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unknown dbtype")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, cfgPath, clean, err := initCfg(ctx, fmt.Sprintf(cfgContent, tpPath))
	if err != nil {
		t.Fatalf("parsing configuration file failed: %v", err)
	}
	defer clean()

	// Flush DBs.
	if err := engine.InitDataDB(cfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(cfg); err != nil {
		t.Fatal(err)
	}

	if _, err := engine.StopStartEngine(cfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
	defer engine.KillEngine(*utils.WaitRater)

	client := engine.NewRPCClient(t, cfg.ListenCfg(), *utils.Encoding)

	var reply string
	err = client.Call(context.Background(), utils.LoaderSv1Run,
		&loaders.ArgsProcessFolder{
			APIOpts: map[string]any{
				utils.MetaCache:       utils.MetaNone,
				utils.MetaStopOnError: false,
			},
		}, &reply)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("ProcessCDR1", func(t *testing.T) {
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destination:  "1002",
			},
			APIOpts: map[string]any{
				utils.MetaOriginID:   "processCDR1",
				utils.MetaUsage:      20 * time.Second,
				utils.MetaChargers:   true,
				utils.MetaRates:      true,
				utils.OptsCDRsExport: false,
			},
		}

		var reply []*utils.EventsWithOpts
		if err := client.Call(context.Background(), utils.CDRsV1ProcessEventWithGet, args,
			&reply); err != nil {
			t.Error(err)
		}
		if len(reply) != 1 {
			t.Fatal("expecting only 1 event")
		}
		if reply[0].Opts[utils.MetaCost] != 2. {
			t.Errorf("expected %v, received: %v", 2., reply[0].Opts[utils.MetaCost])
		}
	})

	t.Run("ProcessCDR2", func(t *testing.T) {
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destination:  "1002",
			},
			APIOpts: map[string]any{
				utils.MetaOriginID:   "processCDR2",
				utils.MetaUsage:      45 * time.Second,
				utils.MetaChargers:   true,
				utils.MetaRates:      true,
				utils.OptsCDRsExport: false,
			},
		}

		var reply []*utils.EventsWithOpts
		if err := client.Call(context.Background(), utils.CDRsV1ProcessEventWithGet, args,
			&reply); err != nil {
			t.Error(err)
		}
		if len(reply) != 1 {
			t.Fatal("expecting only 1 event")
		}
		if reply[0].Opts[utils.MetaCost] != 4.5 {
			t.Errorf("expected %v, received: %v", 4.5, reply[0].Opts[utils.MetaCost])
		}
	})

	t.Run("ProcessCDR2ErrExists", func(t *testing.T) {
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destination:  "1002",
			},
			APIOpts: map[string]any{
				utils.MetaOriginID:   "processCDR2",
				utils.MetaUsage:      time.Minute + 10*time.Second,
				utils.MetaChargers:   true,
				utils.MetaRates:      true,
				utils.OptsCDRsExport: false,
			},
		}

		var reply string
		if err := client.Call(context.Background(), utils.CDRsV1ProcessEvent, args,
			&reply); err == nil || !strings.Contains(err.Error(), "EXISTS") {
			t.Errorf("expecting an %v error, received %v", utils.ErrExists, err)
		}
	})

	t.Run("ProcessCDR2Update", func(t *testing.T) {
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destination:  "1002",
			},
			APIOpts: map[string]any{
				utils.MetaOriginID:   "processCDR2",
				utils.MetaUsage:      time.Minute + 10*time.Second,
				utils.MetaChargers:   true,
				utils.MetaRates:      true,
				utils.MetaRerate:     true,
				utils.OptsCDRsExport: false,
			},
		}

		var reply []*utils.EventsWithOpts
		if err := client.Call(context.Background(), utils.CDRsV1ProcessEventWithGet, args,
			&reply); err != nil {
			t.Error(err)
		}
		if len(reply) != 1 {
			t.Fatal("expecting only 1 event")
		}
		if reply[0].Opts[utils.MetaCost] != 7. {
			t.Errorf("expected %v, received: %v", 7., reply[0].Opts[utils.MetaCost])
		}
	})

	t.Run("GetCDRs", func(t *testing.T) {
		args := &utils.CDRFilters{
			Tenant: "cgrates.org",
			ID:     "GetCDRs1",
		}

		var cdrs []*utils.CDR
		if err := client.Call(context.Background(), utils.AdminSv1GetCDRs, args,
			&cdrs); err != nil {
			t.Error(err)
		}
		sort.Slice(cdrs, func(i, j int) bool {
			return cdrs[i].Opts[utils.MetaCost].(float64) < cdrs[j].Opts[utils.MetaCost].(float64)
		})
		if cdrs[0].Opts[utils.MetaCost] != 2. ||
			cdrs[0].Opts[utils.MetaOriginID] != "processCDR1" {
			t.Errorf("expected first cdr to have originID %s and cost %.2f, received %s and %.1f",
				"processCDR1", 2.,
				cdrs[0].Opts[utils.MetaOriginID], cdrs[0].Opts[utils.MetaCost])
		}
		if cdrs[1].Opts[utils.MetaCost] != 7. ||
			cdrs[1].Opts[utils.MetaOriginID] != "processCDR2" {
			t.Errorf("expected first cdr to have originID %s and cost %.2f, received %s and %.1f",
				"processCDR1", 7.,
				cdrs[1].Opts[utils.MetaOriginID], cdrs[1].Opts[utils.MetaCost])
		}
	})
}
