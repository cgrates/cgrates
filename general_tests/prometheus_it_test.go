//go:build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package general_tests

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestPrometheusAgentIT(t *testing.T) {
	t.Skip("test by looking at the log output")

	cfgNg1 := `{
"listen": {
	"http": ":2080",
	"http_tls": ":2280"
},
"apiers": {
	"enabled": true
},
"stats": {
	"enabled": true,
	"store_interval": "-1"
},
"rpc_conns": {
	"external": {
		"conns": [{
			"address": "127.0.0.1:22012",
			"transport": "*json"
		}]
	}
},
"prometheus_agent": {
	"enabled": true,
	"path": "/metrics",
	"caches_conns": ["*localhost", "external"],
	"cache_ids": [
		"*statqueue_profiles",
		"*statqueues",
		"*stat_filter_indexes",
		"*rpc_connections"
	],
	// "apiers_conns": ["*internal", "external"],
	"stats_conns": ["*internal", "external"],
	"stat_queue_ids": ["cgrates.org:SQ_1","SQ_2"]
}
}`

	cfgNg2 := `{
"listen": {
	"rpc_json": "127.0.0.1:22012",
	"rpc_gob": "",
	"http": "127.0.0.1:22080",
	"rpc_json_tls" : "",
	"rpc_gob_tls": "",
	"http_tls": "127.0.0.1:22280"
},
"apiers": {
	"enabled": true
},
"stats": {
	"enabled": true,
	"store_interval": "-1"
}
}`

	tpFiles := map[string]string{
		// definitions of stat queues common to both engines
		utils.StatsCsv: `
#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],QueueLength[4],TTL[5],MinItems[6],Metrics[7],MetricFilterIDs[8],Stored[9],Blocker[10],Weight[11],ThresholdIDs[12]
cgrates.org,SQ_1,,,100,-1,0,*tcc;*tcd;*acc;*acd;*sum#1,,false,,,*none
cgrates.org,SQ_2,,,100,-1,0,*tcc;*tcd;*acc;*acd;*sum#2,,false,,,*none
cgrates.org,SQ_3,,,100,-1,0,*tcc;*tcd;*acc;*acd;*sum#3,,false,,,*none`,
	}

	ng1 := engine.TestEngine{
		ConfigJSON: cfgNg1,
		TpFiles:    tpFiles,
		DBCfg:      engine.InternalDBCfg,
		LogBuffer:  &bytes.Buffer{},
	}
	defer fmt.Println(ng1.LogBuffer)
	clientNg1, _ := ng1.Run(t)

	ng2 := engine.TestEngine{
		ConfigJSON: cfgNg2,
		TpFiles:    tpFiles,
		DBCfg:      engine.InternalDBCfg,
		LogBuffer:  &bytes.Buffer{},
	}
	defer fmt.Println(ng2.LogBuffer)
	clientNg2, _ := ng2.Run(t)

	scrapePromURL(t)
	for range 3 {
		processStats(t, clientNg1)
		processStats(t, clientNg2)
		scrapePromURL(t)
	}
}

func scrapePromURL(t *testing.T) {
	t.Helper()
	url := "http://localhost:2080/metrics"
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	bodyString := string(body)
	fmt.Println(bodyString)
}

func processStats(t *testing.T, client *birpc.Client) {
	t.Helper()
	var reply []string
	for i := range 3 {
		if err := client.Call(context.Background(), utils.StatSv1ProcessEvent, &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.GenUUID(),
			Event: map[string]any{
				utils.Usage: time.Duration(i) * time.Second,
				utils.Cost:  i * 10,
			},
			APIOpts: map[string]any{
				utils.OptsStatsProfileIDs: []string{fmt.Sprintf("SQ_%d", i+1)},
			},
		}, &reply); err != nil {
			t.Error(err)
		}
	}
}
