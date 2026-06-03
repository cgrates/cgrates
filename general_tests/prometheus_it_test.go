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
	"httpTLS": ":2280"
},
"db": {
	"dbConns": {
		"*default": {
			"dbType": "*internal",
				"opts":{
		"internalDBRewriteInterval": "0s",
		"internalDBDumpInterval": "0s"
	}
    	}
	},
},
"stats": {
	"enabled": true,
	"storeInterval": "-1"
},
"rpcConns": {
	"external": {
		"conns": [{
			"address": "127.0.0.1:22012",
			"transport": "*json"
		}]
	}
},
"prometheusAgent": {
	"enabled": true,
	"path": "/metrics",
	"conns": {
		"*caches": [{"connIDs": ["*localhost", "external"]}]
	},
	"cacheIDs": [
		"*statQueueProfiles",
		"*statQueues",
		"*statFilterIndexes",
		"*rpcConnections"
	],
	// "apiers_conns": ["*internal", "external"],
	"conns": {
		"*stats": [{"connIDs": ["*internal", "external"]}]
	},
	"statQueueIDs": ["cgrates.org:SQ_1","SQ_2"]
}
}`

	cfgNg2 := `{
"listen": {
	"rpcJSON": "127.0.0.1:22012",
	"rpcGOB": "",
	"http": "127.0.0.1:22080",
	"rpcJSONtls" : "",
	"rpcGOBtls": "",
	"httpTLS": "127.0.0.1:22280"
},
"db": {
	"dbConns": {
		"*default": {
			"dbType": "*internal",
				"opts":{
		"internalDBRewriteInterval": "0s",
		"internalDBDumpInterval": "0s"
	}
    	}
	},
},
"stats": {
	"enabled": true,
	"storeInterval": "-1"
}
}`

	tpFiles := map[string]string{
		// definitions of stat queues common to both engines
		utils.StatsCsv: `
#Tenant[0],Id[1],FilterIDs[2],Weights[3],Blockers[4],QueueLength[5],TTL[6],MinItems[7],Stored[8],ThresholdIDs[9],MetricIDs[10],MetricFilterIDs[11],MetricBlockers[12]
cgrates.org,SQ_1,,,,100,-1,0,false,*none,*tcc;*tcd;*acc;*acd;*sum#1,,
cgrates.org,SQ_2,,,,100,-1,0,false,*none,*tcc;*tcd;*acc;*acd;*sum#2,,
cgrates.org,SQ_3,,,,100,-1,0,false,*none,*tcc;*tcd;*acc;*acd;*sum#3,,`,
	}

	ng1 := engine.TestEngine{
		ConfigJSON: cfgNg1,
		TpFiles:    tpFiles,
		DBCfg:      engine.InternalDBCfg,
		Encoding:   *utils.Encoding,
		LogBuffer:  &bytes.Buffer{},
	}
	defer fmt.Println(ng1.LogBuffer)
	clientNg1, _ := ng1.Run(t)

	ng2 := engine.TestEngine{
		ConfigJSON: cfgNg2,
		TpFiles:    tpFiles,
		DBCfg:      engine.InternalDBCfg,
		Encoding:   *utils.Encoding,
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
			Event:  map[string]any{},
			APIOpts: map[string]any{
				utils.MetaUsage:           time.Duration(i) * time.Second,
				utils.MetaCost:            i * 10,
				utils.OptsStatsProfileIDs: fmt.Sprintf("SQ_%d", i+1),
			},
		}, &reply); err != nil {
			t.Error(err)
		}
	}
}
