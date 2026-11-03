//go:build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/ers"
	"github.com/cgrates/cgrates/utils"
)

func TestURRateFromCSV(t *testing.T) {
	ersIn, ersOut, prtPath := t.TempDir(), t.TempDir(), t.TempDir()
	fileIdx := 0
	createFile := func(t *testing.T, dir, ext, content string) {
		fileIdx++
		filePath := filepath.Join(dir, fmt.Sprintf("file%d%s", fileIdx, ext))
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("could not write to file %s: %v", filePath, err)
		}
	}

	// will test for rows containing only updates (no start and no end).
	// after partialCacheTTL, is will post the UR even if it doesnt have a terminate event, and apply cacheDumpFields to the UR posted
	createFile(t, ersIn, utils.CSVSuffix, `1792404000,cgrates.org,update,1001,orgIDHash,,,,,orgHst,csvSrc,,internet,,10,20
1792425600,cgrates.org,update,1001,orgIDHash,,,,,orgHst,csvSrc,,internet,,100,200
1792447200,cgrates.org,update,1001,orgIDHash,,,,,orgHst,csvSrc,,internet,,300,400
1792467000,cgrates.org,update,1001,orgIDHash,,,,,orgHst,csvSrc,,internet,,500,600
`)

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigJSON: fmt.Sprintf(`{
"logger": {
    "level": 7
},
"rates": {
	"enabled": true
},
"sessions": {
	"enabled": true,
	"conns":{
		"*rates": [{"connIDs": ["*localhost"]}],
	},
},
"ers": {
	"enabled": true,
	"conns": {
		"*sessions": [{"connIDs": ["*localhost"]}]
	},
	"partialCacheTTL": "1s",
	"readers": [
		{
			"id": "ReadURCSVs",
			"runDelay": "0",
			"type": "*fileCSV",
			"sourcePath": "%s",
			"processedPath": "%s",
			"startDelay": "100ms",
			"flags": ["*log", "*sessions"],
			"fields":[
				{"tag": "CSVArrivalTime", "path": "*cgreq.CSVArrivalTime", "type": "*variable", "value": "~*req.0{*timeString}"},
				{"tag": "EarliestCSVTime", "path": "*uch.EarliestCSVTime", "type": "*variable", 
				"filters": ["*notexists:~*uch.EarliestCSVTime:"],"value": "~*req.0"},
				{"tag": "EarliestCSVTime", "path": "*uch.EarliestCSVTime", "type": "*variable", 
					"filters": ["*exists:~*uch.EarliestCSVTime:", "*gt:~*uch.EarliestCSVTime:~*req.0"], "value": "~*req.0"},
				{"tag": "Duration", "path": "*cgreq.Duration", "type": "*difference", 
					"value": "~*req.0;~*uch.EarliestCSVTime"},
				{"tag": "Direction", "path": "*cgreq.Direction", "type": "*constant", 
					"value": "*out"},
				{"tag": "Category", "path": "*cgreq.Category", "type": "*constant", 
					"value": "data"},
				{"tag": "Tenant", "path": "*cgreq.Tenant", "type": "*variable", 
					"value": "~*req.1"},
				{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", 
					"value": "~*req.3"},
				{"tag": "Subject", "path": "*cgreq.Subject", "type": "*variable", 
					"value": "~*req.3"},
				{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", 
					"value": "~*req.4"},
				{"tag": "OriginHost", "path": "*cgreq.OriginHost", "type": "*variable", "value": "~*req.9"},
				{"tag": "Source", "path": "*cgreq.Source", "type": "*variable", 
					"value": "~*req.10"},
				{"tag": "Destination", "path": "*cgreq.Destination", "type": 
					"*variable", "value": "~*req.12"},
				{"tag": "SetupTimeUCH", "path": "*uch.SetupTimeUCH", "type": "*variable",
					"filters": ["*notexists:~*uch.SetupTimeUCH:"], "value": "~*req.0{*timeString}"},
				{"tag": "SetupTimeUCHUpdated", "path": "*uch.SetupTimeUCH", "type": "*variable",
					"filters": ["*gt:~*uch.SetupTimeUCH:~*req.0{*timeString}"], "value": "~*uch.SetupTimeUCH"},
				{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*variable",
					 "value": "~*uch.SetupTimeUCH"},
				{"tag": "AnswerTimeUCH", "path": "*uch.AnswerTimeUCH", "type": "*variable",
					"filters": ["*notexists:~*uch.AnswerTimeUCH:"], "value": "~*req.0{*timeString}"},
				{"tag": "AnswerTimeUCHUpdated", "path": "*uch.AnswerTimeUCH", "type": "*variable",
					"filters": ["*gt:~*uch.AnswerTimeUCH:~*req.0{*timeString}"], "value": "~*uch.AnswerTimeUCH"},
				{"tag": "AnswerTime", "path": "*cgreq.AnswerTime", "type": "*variable",
					 "value": "~*uch.AnswerTimeUCH"},


				{"tag": "DataVolumeIncomingOrig", "path": "*cgreq.DataVolumeIncomingOrig", "type": "*variable", "filters": ["*notexists:~*uch.UCHDataVolumeIncomingOrig:"],
					"value": "~*req.14"}, // save first iteration
				{"tag": "DataVolumeIncomingOrig", "path": "*cgreq.DataVolumeIncomingOrig", "type": "*sum", "filters": ["*exists:~*uch.UCHDataVolumeIncomingOrig:"],
					"value": "~*req.14;~*uch.UCHDataVolumeIncomingOrig"}, // on 2nd iteration, save the sum of previous(saved in cache) and current data volume
				{"tag": "UCHDataVolumeIncomingOrig", "path": "*uch.UCHDataVolumeIncomingOrig", "type": "*variable","value": "~*cgreq.DataVolumeIncomingOrig"}, // save in cache the current iteration data volume(overwrite previous)
				
					
				{"tag": "DataVolumeOutgoingOrig", "path": "*cgreq.DataVolumeOutgoingOrig", "type": "*variable", "filters": ["*notexists:~*uch.UCHDataVolumeOutgoingOrig:"],
					"value": "~*req.15"},
				{"tag": "DataVolumeOutgoingOrig", "path": "*cgreq.DataVolumeOutgoingOrig", "type": "*sum", "filters": ["*exists:~*uch.UCHDataVolumeOutgoingOrig:"],
					"value": "~*req.15;~*uch.UCHDataVolumeOutgoingOrig"},
				{"tag": "UCHDataVolumeOutgoingOrig", "path": "*uch.UCHDataVolumeOutgoingOrig", "type": "*variable","value": "~*cgreq.DataVolumeOutgoingOrig"},	

				
				{"tag": "chargeableUnits", "path": "*cgreq.chargeableUnits",
					 "type": "*sum", "value": "~*cgreq.DataVolumeIncomingOrig;~*cgreq.DataVolumeOutgoingOrig"}, 

				{"tag": "InputFile", "path": "*cgreq.InputFile", "type": "*variable", "filters": ["*exists:~*vars.*fileName:", "*exists:~*vars.*fileLineNumber:"],
					"value": "*vars.*fileName;:;*vars.*fileLineNumber"},

				{"tag": "Partial", "path": "*opts.*partial", "type": "*constant", 
					"value": "true"}, // default is partial
				{"tag": "OptSession", "path": "*opts.*session", "type": "*constant",
					"value": "true"},
				{"tag": "OptInterimConsumedOut", "path": "*opts.*interimConsumed", 
					"type": "*sum","value": "~*req.14;~*req.15"},
				{"tag": "MetaOriginID", "path": "*opts.*originID", "type": "*variable", 
					"value": "~*req.4"},
			],
			"partialCommitFields": [
				// UR is completed on terminate but never hit in this case
				{"tag": "Partial", "path": "*opts.*partial", "type": "*constant", 
				"value": "false","filters": ["*exists:~*opts.*terminate:"]}, 
				{"tag": "OptUsageRecord", "path": "*opts.*ur", "type": "*constant",
				"filters": ["*exists:~*opts.*terminate:"], "value": "true"},
				{"tag": "OptRates", "path": "*opts.*rates", "type": "*constant",
				"filters": ["*exists:~*opts.*terminate:"], "value": "true"},
			],
			"cacheDumpFields": [
				{"tag": "CSVArrivalTime", "path": "*cgreq.CSVArrivalTime", "type": "*variable", "value": "~*req.CSVArrivalTime"},
				{"tag": "Duration", "path": "*cgreq.Duration", "type": "*variable", 
					"value": "~*req.Duration"},
				{"tag": "Direction", "path": "*cgreq.Direction", "type": "*constant", 
					"value": "*out"},
				{"tag": "Category", "path": "*cgreq.Category", "type": "*constant", 
					"value": "data"},
				{"tag": "Tenant", "path": "*cgreq.Tenant", "type": "*variable", 
					"value": "~*req.Tenant"},
				{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", 
					"value": "~*req.Account"},
				{"tag": "Subject", "path": "*cgreq.Subject", "type": "*variable", 
					"value": "~*req.Subject"},
				{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", 
					"value": "~*req.OriginID"},
				{"tag": "OriginHost", "path": "*cgreq.OriginHost", "type": "*variable", 
					"value": "~*req.OriginHost"},
				{"tag": "Source", "path": "*cgreq.Source", "type": "*variable", 
					"value": "~*req.Source"},
				{"tag": "Destination", "path": "*cgreq.Destination", "type": 
					"*variable", "value": "~*req.Destination"},
				{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*variable",
					 "value": "~*req.SetupTime"},
				{"tag": "AnswerTime", "path": "*cgreq.AnswerTime", "type": "*variable",
					 "value": "~*req.AnswerTime"},
				
				{"tag": "DataVolumeIncomingOrig", "path": "*cgreq.DataVolumeIncomingOrig", "type": "*variable", "value": "~*req.DataVolumeIncomingOrig"},
				
					
				{"tag": "DataVolumeOutgoingOrig", "path": "*cgreq.DataVolumeOutgoingOrig", "type": "*variable", "value": "~*req.DataVolumeOutgoingOrig"},
				
				{"tag": "OptInterimConsumedOut", "path": "*opts.*interimConsumed", 
					"type": "*constant","value": ""},
				{"tag": "chargeableUnits", "path": "*cgreq.chargeableUnits",
					 "type": "*variable", "value": "~*req.chargeableUnits"}, 
				{"tag": "InputFile", "path": "*cgreq.InputFile", "type": "*variable", 
					"value": "*req.InputFile"},
				{"tag": "Partial", "path": "*opts.*partial", "type": "*constant", 
					"value": "false"}, // not partial when postingUR
				{"tag": "OptSession", "path": "*opts.*session", "type": "*constant",
					"value": "true"},
				{"tag": "OptTerminate", "path": "*opts.*terminate", "type": "*constant",
					"value": "true"},
				{"tag": "MetaOriginID", "path": "*opts.*originID", "type": "*variable", 
					"value": "~*opts.*originID"},
				{"tag": "OptUsageRecord", "path": "*opts.*ur", "type": "*constant",
					"value": "true"},
				{"tag": "OptRates", "path": "*opts.*rates", "type": "*constant",
					"value": "true"},
				{"tag": "OptTotalUsage", "path": "*opts.*totalUsage",
					 "type": "*variable", "value": "~*req.chargeableUnits"},
			],
			"opts": {
				"partialCacheAction": "*postCDR",
				"partialOrderField": "~*req.CSVArrivalTime",
				"partialPath": "%s",
			}
		}
	]
	}
}`, ersIn, ersOut, prtPath),
		TpPath:           filepath.Join(*utils.DataDir, "tariffplans", "tutorial"),
		DBCfg:            getDBCfg(t),
		LogBuffer:        buf,
		GracefulShutdown: true,
	}
	// t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
	client, _ := ng.Run(t)
	time.Sleep(200 * time.Millisecond)
	var rply string
	if err := client.Call(context.Background(), utils.ErSv1RunReader,
		&ers.V1RunReaderParams{
			ReaderID: "ReadURCSVs",
		}, &rply); err != nil {
		t.Fatal(err)
	}
	ersLogCgrUR := "<ERs> LOG, reader: <ReadURCSVs>"
	waitForLog(t, buf, ersLogCgrUR, 3*time.Second)
	if got := strings.Count(buf.String(), ersLogCgrUR); got != 1 {
		t.Fatalf("expected 1 LOG record, got %d", got)
	}
	ev := parseCGREvent(t, buf)
	exp := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     ev.ID,
		Event: map[string]interface{}{
			utils.AccountField:       "1001",
			utils.AnswerTime:         "2026-10-19T10:00:00Z",
			"CSVArrivalTime":         "2026-10-20T03:30:00Z", // the last csv read before posting
			utils.Category:           utils.DataCfg,
			"DataVolumeIncomingOrig": "910",
			"DataVolumeOutgoingOrig": "1220",
			utils.Destination:        "internet",
			"Direction":              "*out",
			"Duration":               "63000",
			"InputFile":              "*req.InputFile",
			utils.OriginHost:         "orgHst",
			utils.OriginID:           "orgIDHash",
			utils.SetupTime:          "2026-10-19T10:00:00Z",
			utils.Source:             "csvSrc",
			utils.Subject:            "1001",
			utils.Tenant:             utils.CGRateSorg,
			"chargeableUnits":        "2130",
		},
		APIOpts: map[string]interface{}{
			utils.MetaInterimConsumed: utils.EmptyString,
			utils.MetaOriginID:        "orgIDHash",
			utils.PartialOpt:          "false",
			utils.MetaRates:           "true",
			utils.MetaSession:         "true",
			utils.MetaTerminate:       "true",
			utils.MetaUR:              "true",
			utils.MetaTotalUsage:      "2130",
		},
	}
	if !reflect.DeepEqual(ev, exp) {
		t.Errorf("expected \n<%v>\nreceived\n<%v>", utils.ToJSON(exp), utils.ToJSON(ev))
	}
}
