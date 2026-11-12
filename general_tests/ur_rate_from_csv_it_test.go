//go:build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/ers"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func TestURRateFromFilesRunReader(t *testing.T) {
	// refs holds how each request value is addressed for the given file type:
	// CSV by column index, XML by element path, FWV by character range, JSON by key.
	type refs struct {
		time, tenant, account, originID, originHost, source, destination, dataIn, dataOut string
	}
	tests := []struct {
		name       string
		readerType string
		suffix     string
		content    string
		refs       refs
	}{
		{
			name:       "CSV",
			readerType: utils.MetaFileCSV,
			suffix:     utils.CSVSuffix,
			content: `1792404000,cgrates.org,update,1001,orgIDHash,,,,,orgHst,csvSrc,,internet,,10,20
1792425600,cgrates.org,update,1001,orgIDHash,,,,,orgHst,csvSrc,,internet,,100,200
1792447200,cgrates.org,update,1001,orgIDHash,,,,,orgHst,csvSrc,,internet,,300,400
1792467000,cgrates.org,update,1001,orgIDHash,,,,,orgHst,csvSrc,,internet,,500,600
`,
			refs: refs{
				time: "0", tenant: "1", account: "3", originID: "4", originHost: "9",
				source: "10", destination: "12", dataIn: "14", dataOut: "15",
			},
		},
		{
			name:       "XML",
			readerType: utils.MetaFileXML,
			suffix:     utils.XMLSuffix,
			content: `<?xml version="1.0" encoding="UTF-8"?>
<Records>
    <Record>
        <Time>1792404000</Time>
        <Tenant>cgrates.org</Tenant>
        <Type>update</Type>
        <Account>1001</Account>
        <OriginID>orgIDHash</OriginID>
        <OriginHost>orgHst</OriginHost>
        <Source>csvSrc</Source>
        <Destination>internet</Destination>
        <DataIn>10</DataIn>
        <DataOut>20</DataOut>
    </Record>
    <Record>
        <Time>1792425600</Time>
        <Tenant>cgrates.org</Tenant>
        <Type>update</Type>
        <Account>1001</Account>
        <OriginID>orgIDHash</OriginID>
        <OriginHost>orgHst</OriginHost>
        <Source>csvSrc</Source>
        <Destination>internet</Destination>
        <DataIn>100</DataIn>
        <DataOut>200</DataOut>
    </Record>
    <Record>
        <Time>1792447200</Time>
        <Tenant>cgrates.org</Tenant>
        <Type>update</Type>
        <Account>1001</Account>
        <OriginID>orgIDHash</OriginID>
        <OriginHost>orgHst</OriginHost>
        <Source>csvSrc</Source>
        <Destination>internet</Destination>
        <DataIn>300</DataIn>
        <DataOut>400</DataOut>
    </Record>
    <Record>
        <Time>1792467000</Time>
        <Tenant>cgrates.org</Tenant>
        <Type>update</Type>
        <Account>1001</Account>
        <OriginID>orgIDHash</OriginID>
        <OriginHost>orgHst</OriginHost>
        <Source>csvSrc</Source>
        <Destination>internet</Destination>
        <DataIn>500</DataIn>
        <DataOut>600</DataOut>
    </Record>
</Records>
`,
			refs: refs{
				time: "Records.Record.Time", tenant: "Records.Record.Tenant", account: "Records.Record.Account",
				originID: "Records.Record.OriginID", originHost: "Records.Record.OriginHost",
				source: "Records.Record.Source", destination: "Records.Record.Destination",
				dataIn: "Records.Record.DataIn", dataOut: "Records.Record.DataOut",
			},
		},
		{
			name:       "FWV",
			readerType: utils.MetaFileFWV,
			suffix:     utils.FWVSuffix,
			// time(0-10) tenant(10-21) account(21-25) originID(25-34) originHost(34-40) source(40-46) destination(46-54) dataIn(54-60) dataOut(60-66)
			content: `1792404000cgrates.org1001orgIDHashorgHstcsvSrcinternet000010000020
1792425600cgrates.org1001orgIDHashorgHstcsvSrcinternet000100000200
1792447200cgrates.org1001orgIDHashorgHstcsvSrcinternet000300000400
1792467000cgrates.org1001orgIDHashorgHstcsvSrcinternet000500000600
`,
			refs: refs{
				time: "0-10", tenant: "10-21", account: "21-25", originID: "25-34", originHost: "34-40",
				source: "40-46", destination: "46-54", dataIn: "54-60", dataOut: "60-66",
			},
		},
		// 		{ // unfinished , fix json reader to read more than 1 object/list of objects
		// 			name:       "JSON",
		// 			readerType: utils.MetaFileJSON,
		// 			suffix:     utils.JSONSuffix,
		// 			content: `[
		// 	{"Time": "1792404000", "Tenant": "cgrates.org", "Type": "update", "Account": "1001", "OriginID": "orgIDHash", "OriginHost": "orgHst", "Source": "csvSrc", "Destination": "internet", "DataIn": "10", "DataOut": "20"},
		// 	{"Time": "1792425600", "Tenant": "cgrates.org", "Type": "update", "Account": "1001", "OriginID": "orgIDHash", "OriginHost": "orgHst", "Source": "csvSrc", "Destination": "internet", "DataIn": "100", "DataOut": "200"},
		// 	{"Time": "1792447200", "Tenant": "cgrates.org", "Type": "update", "Account": "1001", "OriginID": "orgIDHash", "OriginHost": "orgHst", "Source": "csvSrc", "Destination": "internet", "DataIn": "300", "DataOut": "400"},
		// 	{"Time": "1792467000", "Tenant": "cgrates.org", "Type": "update", "Account": "1001", "OriginID": "orgIDHash", "OriginHost": "orgHst", "Source": "csvSrc", "Destination": "internet", "DataIn": "500", "DataOut": "600"}
		// ]
		// `,
		// 			refs: refs{
		// 				time: "Time", tenant: "Tenant", account: "Account", originID: "OriginID", originHost: "OriginHost",
		// 				source: "Source", destination: "Destination", dataIn: "DataIn", dataOut: "DataOut",
		// 			},
		// 		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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
			createFile(t, ersIn, tc.suffix, tc.content)

			// swaps the reader type, paths and the per file type field references
			cfgReplacer := strings.NewReplacer(
				"{{TYPE}}", tc.readerType,
				"{{IN}}", ersIn,
				"{{OUT}}", ersOut,
				"{{PRT}}", prtPath,
				"{{TIME}}", tc.refs.time,
				"{{TENANT}}", tc.refs.tenant,
				"{{ACCOUNT}}", tc.refs.account,
				"{{ORIGINID}}", tc.refs.originID,
				"{{ORIGINHOST}}", tc.refs.originHost,
				"{{SOURCE}}", tc.refs.source,
				"{{DESTINATION}}", tc.refs.destination,
				"{{DATAIN}}", tc.refs.dataIn,
				"{{DATAOUT}}", tc.refs.dataOut,
			)

			buf := &bytes.Buffer{}
			ng := engine.TestEngine{
				ConfigJSON: cfgReplacer.Replace(`{
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
			"type": "{{TYPE}}",
			"sourcePath": "{{IN}}",
			"processedPath": "{{OUT}}",
			"startDelay": "100ms",
			"flags": ["*log", "*sessions"],
			"fields":[
				{"tag": "CSVArrivalTime", "path": "*cgreq.CSVArrivalTime", "type": "*variable", "value": "~*req.{{TIME}}{*time}"},
				{"tag": "EarliestCSVTime", "path": "*uch.EarliestCSVTime", "type": "*variable", 
				"filters": ["*notexists:~*uch.EarliestCSVTime:"],"value": "~*req.{{TIME}}"},
				{"tag": "EarliestCSVTime", "path": "*uch.EarliestCSVTime", "type": "*variable", 
					"filters": ["*exists:~*uch.EarliestCSVTime:", "*gt:~*uch.EarliestCSVTime:~*req.{{TIME}}"], "value": "~*req.{{TIME}}"},
				{"tag": "Duration", "path": "*cgreq.Duration", "type": "*difference", 
					"value": "~*req.{{TIME}};~*uch.EarliestCSVTime"},
				{"tag": "Direction", "path": "*cgreq.Direction", "type": "*constant", 
					"value": "*out"},
				{"tag": "Category", "path": "*cgreq.Category", "type": "*constant", 
					"value": "data"},
				{"tag": "Tenant", "path": "*cgreq.Tenant", "type": "*variable", 
					"value": "~*req.{{TENANT}}"},
				{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", 
					"value": "~*req.{{ACCOUNT}}"},
				{"tag": "Subject", "path": "*cgreq.Subject", "type": "*variable", 
					"value": "~*req.{{ACCOUNT}}"},
				{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", 
					"value": "~*req.{{ORIGINID}}"},
				{"tag": "OriginHost", "path": "*cgreq.OriginHost", "type": "*variable", "value": "~*req.{{ORIGINHOST}}"},
				{"tag": "Source", "path": "*cgreq.Source", "type": "*variable", 
					"value": "~*req.{{SOURCE}}"},
				{"tag": "Destination", "path": "*cgreq.Destination", "type": 
					"*variable", "value": "~*req.{{DESTINATION}}"},
				{"tag": "SetupTimeUCH", "path": "*uch.SetupTimeUCH", "type": "*variable",
					"filters": ["*notexists:~*uch.SetupTimeUCH:"], "value": "~*req.{{TIME}}{*time}"},
				{"tag": "SetupTimeUCHUpdated", "path": "*uch.SetupTimeUCH", "type": "*variable",
					"filters": ["*gt:~*uch.SetupTimeUCH:~*req.{{TIME}}{*time}"], "value": "~*uch.SetupTimeUCH"},
				{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*variable",
					 "value": "~*uch.SetupTimeUCH"},
				{"tag": "AnswerTimeUCH", "path": "*uch.AnswerTimeUCH", "type": "*variable",
					"filters": ["*notexists:~*uch.AnswerTimeUCH:"], "value": "~*req.{{TIME}}{*time}"},
				{"tag": "AnswerTimeUCHUpdated", "path": "*uch.AnswerTimeUCH", "type": "*variable",
					"filters": ["*gt:~*uch.AnswerTimeUCH:~*req.{{TIME}}{*time}"], "value": "~*uch.AnswerTimeUCH"},
				{"tag": "AnswerTime", "path": "*cgreq.AnswerTime", "type": "*variable",
					 "value": "~*uch.AnswerTimeUCH"},


				{"tag": "DataVolumeIncomingOrig", "path": "*cgreq.DataVolumeIncomingOrig", "type": "*variable", "filters": ["*notexists:~*uch.UCHDataVolumeIncomingOrig:"],
					"value": "~*req.{{DATAIN}}"}, // save first iteration
				{"tag": "DataVolumeIncomingOrig", "path": "*cgreq.DataVolumeIncomingOrig", "type": "*sum", "filters": ["*exists:~*uch.UCHDataVolumeIncomingOrig:"],
					"value": "~*req.{{DATAIN}};~*uch.UCHDataVolumeIncomingOrig"}, // on 2nd iteration, save the sum of previous(saved in cache) and current data volume
				{"tag": "UCHDataVolumeIncomingOrig", "path": "*uch.UCHDataVolumeIncomingOrig", "type": "*variable","value": "~*cgreq.DataVolumeIncomingOrig"}, // save in cache the current iteration data volume(overwrite previous)
				
					
				{"tag": "DataVolumeOutgoingOrig", "path": "*cgreq.DataVolumeOutgoingOrig", "type": "*variable", "filters": ["*notexists:~*uch.UCHDataVolumeOutgoingOrig:"],
					"value": "~*req.{{DATAOUT}}"},
				{"tag": "DataVolumeOutgoingOrig", "path": "*cgreq.DataVolumeOutgoingOrig", "type": "*sum", "filters": ["*exists:~*uch.UCHDataVolumeOutgoingOrig:"],
					"value": "~*req.{{DATAOUT}};~*uch.UCHDataVolumeOutgoingOrig"},
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
					"type": "*sum","value": "~*req.{{DATAIN}};~*req.{{DATAOUT}}"},
				{"tag": "MetaOriginID", "path": "*opts.*originID", "type": "*variable", 
					"value": "~*req.{{ORIGINID}}"},
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
				"xmlRootPath": "Records.Record",
				"partialCacheAction": "*processEvent",
				"partialOrderField": "~*req.CSVArrivalTime",
				"partialPath": "{{PRT}}",
			}
		}
	]
	}
}`),
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
		})
	}
}
func TestURRateFromFilesProcessReaderEvents(t *testing.T) {
	// refs holds how each request value is addressed for the given file type:
	// CSV by column index, XML by element path, FWV by character range, JSON by key.
	type refs struct {
		time, tenant, account, originID, originHost, source, destination, dataIn, dataOut string
	}
	tests := []struct {
		name       string
		readerType string
		suffix     string
		content    string
		refs       refs
	}{
		{
			name:       "CSV",
			readerType: utils.MetaFileCSV,
			suffix:     utils.CSVSuffix,
			content: `1792404000,cgrates.org,update,1001,orgIDHash,,,,,orgHst,csvSrc,,internet,,10,20
1792425600,cgrates.org,update,1001,orgIDHash,,,,,orgHst,csvSrc,,internet,,100,200
1792447200,cgrates.org,update,1001,orgIDHash,,,,,orgHst,csvSrc,,internet,,300,400
1792467000,cgrates.org,update,1001,orgIDHash,,,,,orgHst,csvSrc,,internet,,500,600
`,
			refs: refs{
				time: "0", tenant: "1", account: "3", originID: "4", originHost: "9",
				source: "10", destination: "12", dataIn: "14", dataOut: "15",
			},
		},
		{
			name:       "XML",
			readerType: utils.MetaFileXML,
			suffix:     utils.XMLSuffix,
			content: `<?xml version="1.0" encoding="UTF-8"?>
<Records>
    <Record>
        <Time>1792404000</Time>
        <Tenant>cgrates.org</Tenant>
        <Type>update</Type>
        <Account>1001</Account>
        <OriginID>orgIDHash</OriginID>
        <OriginHost>orgHst</OriginHost>
        <Source>csvSrc</Source>
        <Destination>internet</Destination>
        <DataIn>10</DataIn>
        <DataOut>20</DataOut>
    </Record>
    <Record>
        <Time>1792425600</Time>
        <Tenant>cgrates.org</Tenant>
        <Type>update</Type>
        <Account>1001</Account>
        <OriginID>orgIDHash</OriginID>
        <OriginHost>orgHst</OriginHost>
        <Source>csvSrc</Source>
        <Destination>internet</Destination>
        <DataIn>100</DataIn>
        <DataOut>200</DataOut>
    </Record>
    <Record>
        <Time>1792447200</Time>
        <Tenant>cgrates.org</Tenant>
        <Type>update</Type>
        <Account>1001</Account>
        <OriginID>orgIDHash</OriginID>
        <OriginHost>orgHst</OriginHost>
        <Source>csvSrc</Source>
        <Destination>internet</Destination>
        <DataIn>300</DataIn>
        <DataOut>400</DataOut>
    </Record>
    <Record>
        <Time>1792467000</Time>
        <Tenant>cgrates.org</Tenant>
        <Type>update</Type>
        <Account>1001</Account>
        <OriginID>orgIDHash</OriginID>
        <OriginHost>orgHst</OriginHost>
        <Source>csvSrc</Source>
        <Destination>internet</Destination>
        <DataIn>500</DataIn>
        <DataOut>600</DataOut>
    </Record>
</Records>
`,
			refs: refs{
				time: "Records.Record.Time", tenant: "Records.Record.Tenant", account: "Records.Record.Account",
				originID: "Records.Record.OriginID", originHost: "Records.Record.OriginHost",
				source: "Records.Record.Source", destination: "Records.Record.Destination",
				dataIn: "Records.Record.DataIn", dataOut: "Records.Record.DataOut",
			},
		},
		{
			name:       "FWV",
			readerType: utils.MetaFileFWV,
			suffix:     utils.FWVSuffix,
			// time(0-10) tenant(10-21) account(21-25) originID(25-34) originHost(34-40) source(40-46) destination(46-54) dataIn(54-60) dataOut(60-66)
			content: `1792404000cgrates.org1001orgIDHashorgHstcsvSrcinternet000010000020
1792425600cgrates.org1001orgIDHashorgHstcsvSrcinternet000100000200
1792447200cgrates.org1001orgIDHashorgHstcsvSrcinternet000300000400
1792467000cgrates.org1001orgIDHashorgHstcsvSrcinternet000500000600
`,
			refs: refs{
				time: "0-10", tenant: "10-21", account: "21-25", originID: "25-34", originHost: "34-40",
				source: "40-46", destination: "46-54", dataIn: "54-60", dataOut: "60-66",
			},
		},
		// 		{ // unfinished , fix json reader to read more than 1 object/list of objects
		// 			name:       "JSON",
		// 			readerType: utils.MetaFileJSON,
		// 			suffix:     utils.JSONSuffix,
		// 			content: `[
		// 	{"Time": "1792404000", "Tenant": "cgrates.org", "Type": "update", "Account": "1001", "OriginID": "orgIDHash", "OriginHost": "orgHst", "Source": "csvSrc", "Destination": "internet", "DataIn": "10", "DataOut": "20"},
		// 	{"Time": "1792425600", "Tenant": "cgrates.org", "Type": "update", "Account": "1001", "OriginID": "orgIDHash", "OriginHost": "orgHst", "Source": "csvSrc", "Destination": "internet", "DataIn": "100", "DataOut": "200"},
		// 	{"Time": "1792447200", "Tenant": "cgrates.org", "Type": "update", "Account": "1001", "OriginID": "orgIDHash", "OriginHost": "orgHst", "Source": "csvSrc", "Destination": "internet", "DataIn": "300", "DataOut": "400"},
		// 	{"Time": "1792467000", "Tenant": "cgrates.org", "Type": "update", "Account": "1001", "OriginID": "orgIDHash", "OriginHost": "orgHst", "Source": "csvSrc", "Destination": "internet", "DataIn": "500", "DataOut": "600"}
		// ]
		// `,
		// 			refs: refs{
		// 				time: "Time", tenant: "Tenant", account: "Account", originID: "OriginID", originHost: "OriginHost",
		// 				source: "Source", destination: "Destination", dataIn: "DataIn", dataOut: "DataOut",
		// 			},
		// 		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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
			createFile(t, ersIn, tc.suffix, tc.content)

			// swaps the reader type, paths and the per file type field references
			cfgReplacer := strings.NewReplacer(
				"{{TYPE}}", tc.readerType,
				"{{IN}}", ersIn,
				"{{OUT}}", ersOut,
				"{{PRT}}", prtPath,
				"{{TIME}}", tc.refs.time,
				"{{TENANT}}", tc.refs.tenant,
				"{{ACCOUNT}}", tc.refs.account,
				"{{ORIGINID}}", tc.refs.originID,
				"{{ORIGINHOST}}", tc.refs.originHost,
				"{{SOURCE}}", tc.refs.source,
				"{{DESTINATION}}", tc.refs.destination,
				"{{DATAIN}}", tc.refs.dataIn,
				"{{DATAOUT}}", tc.refs.dataOut,
			)

			buf := &bytes.Buffer{}
			ng := engine.TestEngine{
				ConfigJSON: cfgReplacer.Replace(`{
"logger": {
    "level": 7
},
"admins": {
	"enabled": true,
	"conns": {
		"*ers": [{"connIDs": ["*localhost"]}]
	}
},
"rates": {
	"enabled": true
},
"sessions": {
	"enabled": true,
	"conns":{
		"*rates": [{"connIDs": ["*localhost"]}],
		"*ees": [{"connIDs": ["*localhost"]}],
	},
},
"ees": {
	"enabled": true,
	"exporters": [{
		"id": "ur_exporter",
		"type": "*cgrur",
		"exportPath": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
       	"flags": ["*log"],
		"opts": {
			"sqlDBName": "cgrates",
			"sqlTableName": "urs"
		},
	}]
},
"ers": {
	"enabled": true,
	"conns": {
		"*sessions": [{"connIDs": ["*localhost"]}]
	},
	"partialCacheTTL": "1s",
	"readers": [
		{
			"id": "GetURs",
			"runDelay": "0",
			"type": "*cgrur",
			"sourcePath": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
			"processedPath": "",
			"flags": ["*sessions"],
			"fields":[
				{"tag": "NoRate", "path": "*opts.*rates", "type": "*constant", "value": "false"},
			],
			"opts": {
				"sqlDBName": "cgrates",
				"sqlTableName": "urs"
			},
		},
		{
			"id": "ReadURCSVs",
			"runDelay": "0",
			"type": "{{TYPE}}",
			"sourcePath": "{{IN}}",
			"processedPath": "{{OUT}}",
			"startDelay": "100ms",
			"flags": ["*log", "*sessions"],
			"fields":[
				{"tag": "CSVArrivalTime", "path": "*cgreq.CSVArrivalTime", "type": "*variable", "value": "~*req.{{TIME}}{*time}"},
				{"tag": "EarliestCSVTime", "path": "*uch.EarliestCSVTime", "type": "*variable", 
				"filters": ["*notexists:~*uch.EarliestCSVTime:"],"value": "~*req.{{TIME}}"},
				{"tag": "EarliestCSVTime", "path": "*uch.EarliestCSVTime", "type": "*variable", 
					"filters": ["*exists:~*uch.EarliestCSVTime:", "*gt:~*uch.EarliestCSVTime:~*req.{{TIME}}"], "value": "~*req.{{TIME}}"},
				{"tag": "Duration", "path": "*cgreq.Duration", "type": "*difference", 
					"value": "~*req.{{TIME}};~*uch.EarliestCSVTime"},
				{"tag": "Direction", "path": "*cgreq.Direction", "type": "*constant", 
					"value": "*out"},
				{"tag": "Category", "path": "*cgreq.Category", "type": "*constant", 
					"value": "data"},
				{"tag": "Tenant", "path": "*cgreq.Tenant", "type": "*variable", 
					"value": "~*req.{{TENANT}}"},
				{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", 
					"value": "~*req.{{ACCOUNT}}"},
				{"tag": "Subject", "path": "*cgreq.Subject", "type": "*variable", 
					"value": "~*req.{{ACCOUNT}}"},
				{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", 
					"value": "~*req.{{ORIGINID}}"},
				{"tag": "OriginHost", "path": "*cgreq.OriginHost", "type": "*variable", "value": "~*req.{{ORIGINHOST}}"},
				{"tag": "Source", "path": "*cgreq.Source", "type": "*variable", 
					"value": "~*req.{{SOURCE}}"},
				{"tag": "Destination", "path": "*cgreq.Destination", "type": 
					"*variable", "value": "~*req.{{DESTINATION}}"},
				{"tag": "SetupTimeUCH", "path": "*uch.SetupTimeUCH", "type": "*variable",
					"filters": ["*notexists:~*uch.SetupTimeUCH:"], "value": "~*req.{{TIME}}{*time}"},
				{"tag": "SetupTimeUCHUpdated", "path": "*uch.SetupTimeUCH", "type": "*variable",
					"filters": ["*gt:~*uch.SetupTimeUCH:~*req.{{TIME}}{*time}"], "value": "~*uch.SetupTimeUCH"},
				{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*variable",
					 "value": "~*uch.SetupTimeUCH"},
				{"tag": "AnswerTimeUCH", "path": "*uch.AnswerTimeUCH", "type": "*variable",
					"filters": ["*notexists:~*uch.AnswerTimeUCH:"], "value": "~*req.{{TIME}}{*time}"},
				{"tag": "AnswerTimeUCHUpdated", "path": "*uch.AnswerTimeUCH", "type": "*variable",
					"filters": ["*gt:~*uch.AnswerTimeUCH:~*req.{{TIME}}{*time}"], "value": "~*uch.AnswerTimeUCH"},
				{"tag": "AnswerTime", "path": "*cgreq.AnswerTime", "type": "*variable",
					 "value": "~*uch.AnswerTimeUCH"},


				{"tag": "DataVolumeIncomingOrig", "path": "*cgreq.DataVolumeIncomingOrig", "type": "*variable", "filters": ["*notexists:~*uch.UCHDataVolumeIncomingOrig:"],
					"value": "~*req.{{DATAIN}}"}, // save first iteration
				{"tag": "DataVolumeIncomingOrig", "path": "*cgreq.DataVolumeIncomingOrig", "type": "*sum", "filters": ["*exists:~*uch.UCHDataVolumeIncomingOrig:"],
					"value": "~*req.{{DATAIN}};~*uch.UCHDataVolumeIncomingOrig"}, // on 2nd iteration, save the sum of previous(saved in cache) and current data volume
				{"tag": "UCHDataVolumeIncomingOrig", "path": "*uch.UCHDataVolumeIncomingOrig", "type": "*variable","value": "~*cgreq.DataVolumeIncomingOrig"}, // save in cache the current iteration data volume(overwrite previous)
				
					
				{"tag": "DataVolumeOutgoingOrig", "path": "*cgreq.DataVolumeOutgoingOrig", "type": "*variable", "filters": ["*notexists:~*uch.UCHDataVolumeOutgoingOrig:"],
					"value": "~*req.{{DATAOUT}}"},
				{"tag": "DataVolumeOutgoingOrig", "path": "*cgreq.DataVolumeOutgoingOrig", "type": "*sum", "filters": ["*exists:~*uch.UCHDataVolumeOutgoingOrig:"],
					"value": "~*req.{{DATAOUT}};~*uch.UCHDataVolumeOutgoingOrig"},
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
					"type": "*sum","value": "~*req.{{DATAIN}};~*req.{{DATAOUT}}"},
				{"tag": "MetaOriginID", "path": "*opts.*originID", "type": "*variable", 
					"value": "~*req.{{ORIGINID}}"},
			],
			"partialCommitFields": [
				// UR is completed on terminate but never hit in this case
				{"tag": "Partial", "path": "*opts.*partial", "type": "*constant", 
				"value": "false","filters": ["*exists:~*opts.*terminate:"]}, 
				{"tag": "OptUsageRecord", "path": "*opts.*ur", "type": "*constant",
				"filters": ["*exists:~*opts.*terminate:"], "value": "true"},
				{"tag": "OptRates", "path": "*opts.*rates", "type": "*constant",
				"filters": ["*exists:~*opts.*terminate:"], "value": "true"},
				{"tag": "OptEEs", "path": "*opts.*ees", "type": "*constant",
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
				{"tag": "OptEEs", "path": "*opts.*ees", "type": "*constant",
					"filters": ["*exists:~*opts.*terminate:"], "value": "true"},
				{"tag": "OptTotalUsage", "path": "*opts.*totalUsage",
					 "type": "*variable", "value": "~*req.chargeableUnits"},
			],
			"opts": {
				"xmlRootPath": "Records.Record",
				"partialCacheAction": "*processEvent",
				"partialOrderField": "~*req.CSVArrivalTime",
				"partialPath": "{{PRT}}",
			}
		}
	]
	}
}`),
				TpPath:           filepath.Join(*utils.DataDir, "tariffplans", "tutorial"),
				DBCfg:            getDBCfg(t),
				LogBuffer:        buf,
				GracefulShutdown: true,
			}
			// use tutmysql config only for db info to flush after test
			tmpSQLCfg, err := config.NewCGRConfigFromPath(context.TODO(), filepath.Join(*utils.DataDir, "conf", "samples", "tutmysql"))
			if err != nil {
				t.Fatalf("could not init config from path %s: %v", filepath.Join(*utils.DataDir, "conf", "samples", "tutmysql"), err)
			}
			if err := engine.InitDB(tmpSQLCfg); err != nil { // init and flush sql db
				t.Fatalf("failed to flush DataDB err: %v", err)
			}
			// t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
			client, _ := ng.Run(t)
			time.Sleep(200 * time.Millisecond)
			ersReplies := new([]*sessions.V1ProcessEventReply)
			if err := client.Call(context.Background(), utils.ErSv1ProcessReaderEvents,
				&ers.V1RunReaderParams{
					ReaderID: "ReadURCSVs",
				}, &ersReplies); err != nil {
				t.Fatal(err)
			}
			if len(*ersReplies) != 0 {
				t.Errorf("expected no replies populated, received <%v>", ersReplies)
			}
			time.Sleep(3 * time.Second)
			var rows []map[string]*utils.CGREvent
			if err := client.Call(context.Background(), utils.AdminSv1GetURs,
				&apis.URFiltersFromReader{
					ReaderID: "GetURs",
				}, &rows); err != nil {
				t.Fatal(err)
			}
			if len(rows) != 1 {
				t.Fatalf("expected 1 row, received <%v>", rows)
			}
			rateID := rows[0][utils.MetaPrimary].APIOpts[utils.MetaRatesCost].(map[string]any)[utils.CostIntervals].([]any)[0].(map[string]any)[utils.Increments].([]any)[0].(map[string]any)[utils.RateID].(string)
			cgrID := rows[0][utils.MetaPrimary].APIOpts[utils.MetaCGRid]
			expRows := []map[string]*utils.CGREvent{
				{
					utils.MetaPrimary: {
						Tenant: "cgrates.org",
						ID:     rows[0][utils.MetaPrimary].ID,
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
							utils.MetaCGRid:           cgrID,
							utils.MetaRunID:           utils.MetaPrimary,
							utils.MetaInterimConsumed: utils.EmptyString,
							utils.MetaOriginID:        "orgIDHash",
							utils.PartialOpt:          "false",
							utils.MetaRates:           "false",
							utils.MetaRatesCost: map[string]any{
								utils.Altered: nil,
								utils.Cost:    0.05,
								utils.CostIntervals: []any{
									map[string]any{
										utils.CompressFactor: 1.0,
										utils.Increments: []any{
											map[string]any{
												utils.CompressFactor:    60.0,
												utils.RateID:            rateID,
												utils.RateIntervalIndex: 0.0,
												utils.Usage:             60000000000.0,
											},
										},
									},
								},
								utils.ID:              "RP_2",
								utils.MaxCost:         0.0,
								utils.MaxCostStrategy: "",
								utils.MinCost:         0.0,
								utils.Rates: map[string]any{
									rateID: map[string]any{
										utils.FixedFee:      0.0,
										utils.Increment:     1000000000.0,
										utils.IntervalStart: 0.0,
										utils.RecurrentFee:  0.05,
										utils.Unit:          60000000000.0,
									},
								},
							},
							utils.MetaEEs:        "true",
							utils.MetaExporterID: "ur_exporter",
							utils.MetaSession:    "true",
							utils.MetaTerminate:  "true",
							utils.MetaUR:         true,
							utils.MetaTotalUsage: "2130",
							utils.MetaURID:       "orgIDHash",
						},
					},
				},
			}
			if !reflect.DeepEqual(expRows, rows) {
				t.Errorf("expected \n<%v>\nreceived\n<%v>", utils.ToJSON(expRows), utils.ToJSON(rows))
			}
		})
	}
}

func TestURRateFromFilesProcessReaderEventsUnusualCases(t *testing.T) {
	// refs holds how each request value is addressed for the given file type:
	// CSV by column index, XML by element path, FWV by character range, JSON by key.
	type refs struct {
		time, tenant, account, originID, originHost, source, destination, dataIn, dataOut string
	}
	tests := []struct {
		name       string
		readerType string
		suffix     string
		content    string
		refs       refs
		expLog     string
		expErr     string
	}{
		{
			name:       "CSVEmptyFiles",
			readerType: utils.MetaFileCSV,
			suffix:     utils.CSVSuffix,
			content:    ``,
			refs: refs{
				time: "0", tenant: "1", account: "3", originID: "4", originHost: "9",
				source: "10", destination: "12", dataIn: "14", dataOut: "15",
			},
			expLog: "Total records processed: 0, events posted: 0, run duration:",
		},
		{
			name:       "CSVInvalidContent",
			readerType: utils.MetaFileCSV,
			suffix:     utils.CSVSuffix,
			content:    `1`,
			refs: refs{
				time: "0", tenant: "1", account: "3", originID: "4", originHost: "9",
				source: "10", destination: "12", dataIn: "14", dataOut: "15",
			},
			expLog: "[WARNING] <ERs> processing file file1.csv, error: Unsupported time format",
		},
		{
			name:       "TypeNotSupported",
			readerType: utils.MetaKafkajsonMap,
			suffix:     "",
			content:    ``,
			refs: refs{
				time: "0", tenant: "1", account: "3", originID: "4", originHost: "9",
				source: "10", destination: "12", dataIn: "14", dataOut: "15",
			},
			expErr: "reader type <*kafkaJSONMap> not yet supported",
		},
		{
			name:       "XMLEmptyFiles",
			readerType: utils.MetaFileXML,
			suffix:     utils.XMLSuffix,
			content:    ``,
			refs: refs{
				time: "Records.Record.Time", tenant: "Records.Record.Tenant", account: "Records.Record.Account",
				originID: "Records.Record.OriginID", originHost: "Records.Record.OriginHost",
				source: "Records.Record.Source", destination: "Records.Record.Destination",
				dataIn: "Records.Record.DataIn", dataOut: "Records.Record.DataOut",
			},
			expLog: "[WARNING] <ERs> processing file file1.xml, error: xmlquery: invalid XML document",
		},
		{
			name:       "FWVEmptyFiles",
			readerType: utils.MetaFileFWV,
			suffix:     utils.FWVSuffix,
			// time(0-10) tenant(10-21) account(21-25) originID(25-34) originHost(34-40) source(40-46) destination(46-54) dataIn(54-60) dataOut(60-66)
			content: ``,
			refs: refs{
				time: "0-10", tenant: "10-21", account: "21-25", originID: "25-34", originHost: "34-40",
				source: "40-46", destination: "46-54", dataIn: "54-60", dataOut: "60-66",
			},
			expLog: "[WARNING] <ERs> processing file file1.fwv, error: FinalIndex : 10 is greater than : 0",
		},
		// 		{ // unfinished , fix json reader to read more than 1 object/list of objects
		// 			name:       "JSONEmptyFiles",
		// 			readerType: utils.MetaFileJSON,
		// 			suffix:     utils.JSONSuffix,
		// 			content: ``,
		// 			refs: refs{
		// 				time: "Time", tenant: "Tenant", account: "Account", originID: "OriginID", originHost: "OriginHost",
		// 				source: "Source", destination: "Destination", dataIn: "DataIn", dataOut: "DataOut",
		// 			},
		// 		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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
			createFile(t, ersIn, tc.suffix, tc.content)

			// swaps the reader type, paths and the per file type field references
			cfgReplacer := strings.NewReplacer(
				"{{TYPE}}", tc.readerType,
				"{{IN}}", ersIn,
				"{{OUT}}", ersOut,
				"{{PRT}}", prtPath,
				"{{TIME}}", tc.refs.time,
				"{{TENANT}}", tc.refs.tenant,
				"{{ACCOUNT}}", tc.refs.account,
				"{{ORIGINID}}", tc.refs.originID,
				"{{ORIGINHOST}}", tc.refs.originHost,
				"{{SOURCE}}", tc.refs.source,
				"{{DESTINATION}}", tc.refs.destination,
				"{{DATAIN}}", tc.refs.dataIn,
				"{{DATAOUT}}", tc.refs.dataOut,
			)

			buf := &bytes.Buffer{}
			ng := engine.TestEngine{
				ConfigJSON: cfgReplacer.Replace(`{
"logger": {
    "level": 7
},
"admins": {
	"enabled": true,
	"conns": {
		"*ers": [{"connIDs": ["*localhost"]}]
	}
},
"rates": {
	"enabled": true
},
"sessions": {
	"enabled": true,
	"conns":{
		"*rates": [{"connIDs": ["*localhost"]}],
		"*ees": [{"connIDs": ["*localhost"]}],
	},
},
"ees": {
	"enabled": true,
	"exporters": [{
		"id": "ur_exporter",
		"type": "*cgrur",
		"exportPath": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
       	"flags": ["*log"],
		"opts": {
			"sqlDBName": "cgrates",
			"sqlTableName": "urs"
		},
	}]
},
"ers": {
	"enabled": true,
	"conns": {
		"*sessions": [{"connIDs": ["*localhost"]}]
	},
	"partialCacheTTL": "1s",
	"readers": [
		{
			"id": "GetURs",
			"runDelay": "0",
			"type": "*cgrur",
			"sourcePath": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
			"processedPath": "",
			"flags": ["*sessions"],
			"fields":[
				{"tag": "NoRate", "path": "*opts.*rates", "type": "*constant", "value": "false"},
			],
			"opts": {
				"sqlDBName": "cgrates",
				"sqlTableName": "urs"
			},
		},
		{
			"id": "ReadURCSVs",
			"runDelay": "0",
			"type": "{{TYPE}}",
			"sourcePath": "{{IN}}",
			"processedPath": "{{OUT}}",
			"startDelay": "100ms",
			"flags": ["*log", "*sessions"],
			"fields":[
				{"tag": "CSVArrivalTime", "path": "*cgreq.CSVArrivalTime", "type": "*variable", "value": "~*req.{{TIME}}{*time}"},
				{"tag": "EarliestCSVTime", "path": "*uch.EarliestCSVTime", "type": "*variable", 
				"filters": ["*notexists:~*uch.EarliestCSVTime:"],"value": "~*req.{{TIME}}"},
				{"tag": "EarliestCSVTime", "path": "*uch.EarliestCSVTime", "type": "*variable", 
					"filters": ["*exists:~*uch.EarliestCSVTime:", "*gt:~*uch.EarliestCSVTime:~*req.{{TIME}}"], "value": "~*req.{{TIME}}"},
				{"tag": "Duration", "path": "*cgreq.Duration", "type": "*difference", 
					"value": "~*req.{{TIME}};~*uch.EarliestCSVTime"},
				{"tag": "Direction", "path": "*cgreq.Direction", "type": "*constant", 
					"value": "*out"},
				{"tag": "Category", "path": "*cgreq.Category", "type": "*constant", 
					"value": "data"},
				{"tag": "Tenant", "path": "*cgreq.Tenant", "type": "*variable", 
					"value": "~*req.{{TENANT}}"},
				{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", 
					"value": "~*req.{{ACCOUNT}}"},
				{"tag": "Subject", "path": "*cgreq.Subject", "type": "*variable", 
					"value": "~*req.{{ACCOUNT}}"},
				{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", 
					"value": "~*req.{{ORIGINID}}"},
				{"tag": "OriginHost", "path": "*cgreq.OriginHost", "type": "*variable", "value": "~*req.{{ORIGINHOST}}"},
				{"tag": "Source", "path": "*cgreq.Source", "type": "*variable", 
					"value": "~*req.{{SOURCE}}"},
				{"tag": "Destination", "path": "*cgreq.Destination", "type": 
					"*variable", "value": "~*req.{{DESTINATION}}"},
				{"tag": "SetupTimeUCH", "path": "*uch.SetupTimeUCH", "type": "*variable",
					"filters": ["*notexists:~*uch.SetupTimeUCH:"], "value": "~*req.{{TIME}}{*time}"},
				{"tag": "SetupTimeUCHUpdated", "path": "*uch.SetupTimeUCH", "type": "*variable",
					"filters": ["*gt:~*uch.SetupTimeUCH:~*req.{{TIME}}{*time}"], "value": "~*uch.SetupTimeUCH"},
				{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*variable",
					 "value": "~*uch.SetupTimeUCH"},
				{"tag": "AnswerTimeUCH", "path": "*uch.AnswerTimeUCH", "type": "*variable",
					"filters": ["*notexists:~*uch.AnswerTimeUCH:"], "value": "~*req.{{TIME}}{*time}"},
				{"tag": "AnswerTimeUCHUpdated", "path": "*uch.AnswerTimeUCH", "type": "*variable",
					"filters": ["*gt:~*uch.AnswerTimeUCH:~*req.{{TIME}}{*time}"], "value": "~*uch.AnswerTimeUCH"},
				{"tag": "AnswerTime", "path": "*cgreq.AnswerTime", "type": "*variable",
					 "value": "~*uch.AnswerTimeUCH"},


				{"tag": "DataVolumeIncomingOrig", "path": "*cgreq.DataVolumeIncomingOrig", "type": "*variable", "filters": ["*notexists:~*uch.UCHDataVolumeIncomingOrig:"],
					"value": "~*req.{{DATAIN}}"}, // save first iteration
				{"tag": "DataVolumeIncomingOrig", "path": "*cgreq.DataVolumeIncomingOrig", "type": "*sum", "filters": ["*exists:~*uch.UCHDataVolumeIncomingOrig:"],
					"value": "~*req.{{DATAIN}};~*uch.UCHDataVolumeIncomingOrig"}, // on 2nd iteration, save the sum of previous(saved in cache) and current data volume
				{"tag": "UCHDataVolumeIncomingOrig", "path": "*uch.UCHDataVolumeIncomingOrig", "type": "*variable","value": "~*cgreq.DataVolumeIncomingOrig"}, // save in cache the current iteration data volume(overwrite previous)
				
					
				{"tag": "DataVolumeOutgoingOrig", "path": "*cgreq.DataVolumeOutgoingOrig", "type": "*variable", "filters": ["*notexists:~*uch.UCHDataVolumeOutgoingOrig:"],
					"value": "~*req.{{DATAOUT}}"},
				{"tag": "DataVolumeOutgoingOrig", "path": "*cgreq.DataVolumeOutgoingOrig", "type": "*sum", "filters": ["*exists:~*uch.UCHDataVolumeOutgoingOrig:"],
					"value": "~*req.{{DATAOUT}};~*uch.UCHDataVolumeOutgoingOrig"},
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
					"type": "*sum","value": "~*req.{{DATAIN}};~*req.{{DATAOUT}}"},
				{"tag": "MetaOriginID", "path": "*opts.*originID", "type": "*variable", 
					"value": "~*req.{{ORIGINID}}"},
			],
			"partialCommitFields": [
				// UR is completed on terminate but never hit in this case
				{"tag": "Partial", "path": "*opts.*partial", "type": "*constant", 
				"value": "false","filters": ["*exists:~*opts.*terminate:"]}, 
				{"tag": "OptUsageRecord", "path": "*opts.*ur", "type": "*constant",
				"filters": ["*exists:~*opts.*terminate:"], "value": "true"},
				{"tag": "OptRates", "path": "*opts.*rates", "type": "*constant",
				"filters": ["*exists:~*opts.*terminate:"], "value": "true"},
				{"tag": "OptEEs", "path": "*opts.*ees", "type": "*constant",
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
				{"tag": "OptEEs", "path": "*opts.*ees", "type": "*constant",
					"filters": ["*exists:~*opts.*terminate:"], "value": "true"},
				{"tag": "OptTotalUsage", "path": "*opts.*totalUsage",
					 "type": "*variable", "value": "~*req.chargeableUnits"},
			],
			"opts": {
				"xmlRootPath": "Records.Record",
				"partialCacheAction": "*processEvent",
				"partialOrderField": "~*req.CSVArrivalTime",
				"partialPath": "{{PRT}}",
			}
		}
	]
	}
}`),
				TpPath:           filepath.Join(*utils.DataDir, "tariffplans", "tutorial"),
				DBCfg:            getDBCfg(t),
				LogBuffer:        buf,
				GracefulShutdown: true,
			}
			// use tutmysql config only for db info to flush after test
			tmpSQLCfg, err := config.NewCGRConfigFromPath(context.TODO(), filepath.Join(*utils.DataDir, "conf", "samples", "tutmysql"))
			if err != nil {
				t.Fatalf("could not init config from path %s: %v", filepath.Join(*utils.DataDir, "conf", "samples", "tutmysql"), err)
			}
			if err := engine.InitDB(tmpSQLCfg); err != nil { // init and flush sql db
				t.Fatalf("failed to flush DataDB err: %v", err)
			}
			// t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
			client, _ := ng.Run(t)
			time.Sleep(200 * time.Millisecond)
			ersReplies := new([]*sessions.V1ProcessEventReply)
			if err := client.Call(context.Background(), utils.ErSv1ProcessReaderEvents,
				&ers.V1RunReaderParams{
					ReaderID: "ReadURCSVs",
				}, &ersReplies); err != nil {
				if err.Error() != tc.expErr {
					t.Fatalf("expected error <%v>, received <%v>", tc.expErr, err)
				}
			}
			if len(*ersReplies) != 0 {
				t.Errorf("expected no replies populated, received <%v>", ersReplies)
			}
			time.Sleep(3 * time.Second)
			var rows []map[string]*utils.CGREvent
			if err := client.Call(context.Background(), utils.AdminSv1GetURs,
				&apis.URFiltersFromReader{
					ReaderID: "GetURs",
				}, &rows); err != nil {
				t.Fatal(err)
			}
			if len(rows) != 0 {
				t.Fatalf("expected 0 rows, received <%v>", rows)
			}
			logOutput := buf.String()
			if !strings.Contains(logOutput, tc.expLog) {
				t.Fatalf("log <%v> not found in log output:\n%s", tc.expLog, logOutput)
			}
		})
	}
}

func Test13CasesURRateFromCSV(t *testing.T) {
	ersIn, ersOut, prtPath := t.TempDir(), t.TempDir(), t.TempDir()
	fileIdx := 0
	createFile := func(t *testing.T, dir, ext, content string) {
		fileIdx++
		filePath := filepath.Join(dir, fmt.Sprintf("file%d%s", fileIdx, ext))
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("could not write to file %s: %v", filePath, err)
		}
	}

	cases := []string{
		// case 1 (one-session-one-day) success
		`1792414800,276011234567890,session_start,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,1500,300
1792416600,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,2000000,150000
1792418400,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,3500000,210000
1792419300,276011234567890,default_bearer_end,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,42000,5100`,

		// case 2 (round-once-at-close) success
		`1792414800,276011234567890,session_start,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,100,100
1792415400,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,100,100
1792416000,276011234567890,default_bearer_end,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,100,100`,

		// case 3 (updates-only-no-start-no-end) success
		`1792404000,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,500000,40000
1792425600,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,750000,61000
1792447200,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,125000,9000
1792467000,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,64000,3000`,

		// case 4 (no-start-has-end) success
		`1792414800,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,800000,70000
1792418400,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,900000,80000
1792422000,276011234567890,default_bearer_end,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,10000,1000`,

		// case 5 (has-start-no-end), success
		`1792432800,276011234567890,session_start,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,2000,1000
1792436400,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,600000,50000
1792440000,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,400000,30000`,

		// case 6 (crosses-local-midnight), success
		`1792452600,276011234567890,session_start,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,4000,1000
1792456200,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,1000000,100000
1792467900,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,2000000,200000
1792469400,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,3000000,300000
1792483200,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,500000,50000
1792501200,276011234567890,default_bearer_end,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,7000,700`,

		// case 7 (three-days-middle-day-updates-only), success
		`1792447200,276011234567890,session_start,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,3000,1000
1792465200,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,1100000,90000
1792490400,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,1200000,95000
1792512000,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,1300000,96000
1792533600,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,1400000,97000
1792558800,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,1500000,98000
1792584000,276011234567890,default_bearer_end,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,6000,500`,

		// case 8 (two-bearers-one-subscriber), success
		`1792414800,276011234567890,session_start,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,5000,1000
1792414805,276011234567890,session_start,1002,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,ims,5,900,900
1792416600,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,2500000,200000
1792416600,276011234567890,default_bearer_update,1002,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,ims,5,40000,38000
1792418400,276011234567890,default_bearer_end,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,10000,2000
1792418700,276011234567890,default_bearer_end,1002,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,ims,5,1200,1100`,

		// case 9 (same-charging-id-different-session), success
		`1792414800,276011234567890,session_start,777,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,100000,10000
1792414800,276011234567891,session_start,777,15550000002,3512057400000202,3606532,1101,1041933,10.1.0.10,100.64.0.2|,8700000000C0A80A2F,internet,9,200000,20000
1792415100,276011234567890,session_start,777,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A3F,internet,9,300000,30000
1792416600,276011234567890,default_bearer_update,0,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,400000,40000
1792416600,276011234567891,default_bearer_update,0,15550000002,3512057400000202,3606532,1101,1041933,10.1.0.10,100.64.0.2|,8700000000C0A80A2F,internet,9,500000,50000
1792418400,276011234567890,default_bearer_end,777,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,1000,100
1792418400,276011234567891,default_bearer_end,777,15550000002,3512057400000202,3606532,1101,1041933,10.1.0.10,100.64.0.2|,8700000000C0A80A2F,internet,9,2000,200
1792418700,276011234567890,default_bearer_end,777,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A3F,internet,9,3000,300`,

		// case 10 (moves-cell-and-tac-within-site), success
		`1792414800,276011234567890,session_start,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,4000,1000
1792416600,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041934,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,1000000,80000
1792418400,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1024,2050001,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,2000000,160000
1792420200,276011234567890,default_bearer_end,1001,15550000001,3512057400000101,3606532,1024,2050001,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,3000,300`,

		// case 11 (moves-to-another-site-and-timezone), success
		`1792461600,276011234567890,session_start,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,4000,1000
1792467000,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,1000000,80000
1792469700,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,2001,3070001,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,2000000,160000
1792471500,276011234567890,default_bearer_end,1001,15550000001,3512057400000101,3606532,2001,3070001,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,3000,300`,

		// case 12 (rows-out-of-order-across-files), success
		`1792422000,276011234567890,default_bearer_update,1001,15550000001,3512057400000199,3606532,1101,1041999,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,700000,60000
1792425600,276011234567890,default_bearer_end,1001,15550000001,3512057400000199,3606532,1101,1041999,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,5000,500`,
		`1792414800,276011234567890,session_start,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,3000,1000
1792418400,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,600000,50000`,

		// case 13 (rows-that-must-be-dropped), success
		`epoch,imsi,event,charging_id,msisdn,ue_imei,plmn,tac,eci,sgw_ip,ue_ip,pgw_ip,apn,qci,octets_in,octets_out
1792414800,276011234567890,session_start,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,1000,1000
1792415400,,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,5000,5000
1792415400,276011234567890,default_bearer_update,2001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,sos,9,5000,5000
1792415400,313380000000009,default_bearer_update,2002,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,5000,5000
1792416000,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,64424509440,5000
1789909200,276011234567890,default_bearer_update,2003,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,5000,5000
1792416600,276011234567890,session_start,2004,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,0,0
1792417200,276011234567890,default_bearer_end,2004,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,0,0
1792417500,276011234567890,default_bearer_update,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,,
1792418400,276011234567890,default_bearer_end,1001,15550000001,3512057400000101,3606532,1101,1041933,10.1.0.10,100.64.0.1|,8700000000C0A80A2F,internet,9,2000,3000`,
	}
	expectedRows := []string{
		// case 1, idx 0 (success)
		`[{"*primary":{"Tenant":"cgrates.org","ID":"da39a3ee5e6b4b0d3255bfef95601890afd80709","Event":{"Account":"PARTNER_276011","AnswerTime":"2026-10-19T13:00:00Z","Category":"data","Destination":"internet","Direction":"*out","OriginHost":"cgrates-sgw","OriginID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","SetupTime":"2026-10-19T13:00:00Z","Source":"cgrates-sgw","Subject":"276011234567890","Tenant":"PARTNER_276011","Usage":"5909504","cellId":"360653210419331101","chargeableUnits":5908900,"chargingId":"1001","dataVolumeIncoming":"5543936","dataVolumeIncomingOrig":5543500,"dataVolumeOutgoing":"365568","dataVolumeOutgoingOrig":365400,"duration":4500,"imei":"3512057400000101","imsi":"276011234567890","inputFile":"file1.csv:1","mSTimeZone":"","msisdn":"15550000001","pGWAddress":"8700000000C0A80A2F","pdpAddress":"100.64.0.1|","qci":"9","sGWAddress":"10.1.0.10","tac":"1101"},"APIOpts":{"*cgrID":"09047612e75fb9e38c66dd73d6158911f8df3bd3","*ees":"false","*exporterID":"ur_exporter","*originID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","*partial":"false","*rates":"false","*runID":"*primary","*session":"true","*terminate":"true","*ur":true,"*urID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","CSVArrivalTime":"2026-10-19T14:15:00Z","NonUpdate":"true"}}}]`,

		// case 2, idx 1 (success)
		`[{"*primary":{"Tenant":"cgrates.org","ID":"da39a3ee5e6b4b0d3255bfef95601890afd80709","Event":{"Account":"PARTNER_276011","AnswerTime":"2026-10-19T13:00:00Z","Category":"data","Destination":"internet","Direction":"*out","OriginHost":"cgrates-sgw","OriginID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","SetupTime":"2026-10-19T13:00:00Z","Source":"cgrates-sgw","Subject":"276011234567890","Tenant":"PARTNER_276011","Usage":"1024","cellId":"360653210419331101","chargeableUnits":600,"chargingId":"1001","dataVolumeIncoming":"1024","dataVolumeIncomingOrig":300,"dataVolumeOutgoing":"1024","dataVolumeOutgoingOrig":300,"duration":1200,"imei":"3512057400000101","imsi":"276011234567890","inputFile":"file2.csv:1","mSTimeZone":"","msisdn":"15550000001","pGWAddress":"8700000000C0A80A2F","pdpAddress":"100.64.0.1|","qci":"9","sGWAddress":"10.1.0.10","tac":"1101"},"APIOpts":{"*cgrID":"09047612e75fb9e38c66dd73d6158911f8df3bd3","*ees":"false","*exporterID":"ur_exporter","*originID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","*partial":"false","*rates":"false","*runID":"*primary","*session":"true","*terminate":"true","*ur":true,"*urID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","CSVArrivalTime":"2026-10-19T13:20:00Z","NonUpdate":"true"}}}]`,

		// case 3, idx 2 (success)
		`[{"*primary":{"Tenant":"cgrates.org","ID":"da39a3ee5e6b4b0d3255bfef95601890afd80709","Event":{"Account":"PARTNER_276011","AnswerTime":"2026-10-19T10:00:00Z","Category":"data","Destination":"internet","Direction":"*out","OriginHost":"cgrates-sgw","OriginID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","SetupTime":"2026-10-19T10:00:00Z","Source":"cgrates-sgw","Subject":"276011234567890","Tenant":"PARTNER_276011","Usage":"1552384","cellId":"360653210419331101","chargeableUnits":"1552000","chargingId":"1001","dataVolumeIncoming":"1439744","dataVolumeIncomingOrig":"1439000","dataVolumeOutgoing":"113664","dataVolumeOutgoingOrig":"113000","duration":"86400","imei":"3512057400000101","imsi":"276011234567890","inputFile":"file3.csv:1","mSTimeZone":"","msisdn":"15550000001","pGWAddress":"8700000000C0A80A2F","pdpAddress":"100.64.0.1|","qci":"9","sGWAddress":"10.1.0.10","tac":"1101"},"APIOpts":{"*cgrID":"09047612e75fb9e38c66dd73d6158911f8df3bd3","*ees":"false","*exporterID":"ur_exporter","*originID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","*partial":"false","*rates":"false","*runID":"*primary","*session":"true","*terminate":"true","*ur":true,"*urID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","CSVArrivalTime":"2026-10-20T03:30:00Z"}}}]`,

		// case 4, idx 3 (success)
		`[{"*primary":{"Tenant":"cgrates.org","ID":"da39a3ee5e6b4b0d3255bfef95601890afd80709","Event":{"Account":"PARTNER_276011","AnswerTime":"2026-10-19T13:00:00Z","Category":"data","Destination":"internet","Direction":"*out","OriginHost":"cgrates-sgw","OriginID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","SetupTime":"2026-10-19T13:00:00Z","Source":"cgrates-sgw","Subject":"276011234567890","Tenant":"PARTNER_276011","Usage":"1861632","cellId":"360653210419331101","chargeableUnits":1861000,"chargingId":"1001","dataVolumeIncoming":"1710080","dataVolumeIncomingOrig":1710000,"dataVolumeOutgoing":"151552","dataVolumeOutgoingOrig":151000,"duration":7200,"imei":"3512057400000101","imsi":"276011234567890","inputFile":"file4.csv:1","mSTimeZone":"","msisdn":"15550000001","pGWAddress":"8700000000C0A80A2F","pdpAddress":"100.64.0.1|","qci":"9","sGWAddress":"10.1.0.10","tac":"1101"},"APIOpts":{"*cgrID":"09047612e75fb9e38c66dd73d6158911f8df3bd3","*ees":"false","*exporterID":"ur_exporter","*originID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","*partial":"false","*rates":"false","*runID":"*primary","*session":"true","*terminate":"true","*ur":true,"*urID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","CSVArrivalTime":"2026-10-19T15:00:00Z","NonUpdate":"true"}}}]`,

		// case 5, idx 4 (success)
		`[{"*primary":{"Tenant":"cgrates.org","ID":"da39a3ee5e6b4b0d3255bfef95601890afd80709","Event":{"Account":"PARTNER_276011","AnswerTime":"2026-10-19T18:00:00Z","Category":"data","Destination":"internet","Direction":"*out","OriginHost":"cgrates-sgw","OriginID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","SetupTime":"2026-10-19T18:00:00Z","Source":"cgrates-sgw","Subject":"276011234567890","Tenant":"PARTNER_276011","Usage":"1083392","cellId":"360653210419331101","chargeableUnits":"1083000","chargingId":"1001","dataVolumeIncoming":"1002496","dataVolumeIncomingOrig":"1002000","dataVolumeOutgoing":"81920","dataVolumeOutgoingOrig":"81000","duration":"7200","imei":"3512057400000101","imsi":"276011234567890","inputFile":"file5.csv:1","mSTimeZone":"","msisdn":"15550000001","pGWAddress":"8700000000C0A80A2F","pdpAddress":"100.64.0.1|","qci":"9","sGWAddress":"10.1.0.10","tac":"1101"},"APIOpts":{"*cgrID":"09047612e75fb9e38c66dd73d6158911f8df3bd3","*ees":"false","*exporterID":"ur_exporter","*originID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","*partial":"false","*rates":"false","*runID":"*primary","*session":"true","*terminate":"true","*ur":true,"*urID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","CSVArrivalTime":"2026-10-19T20:00:00Z","NonUpdate":"true"}}}]`,

		// case 6, idx 5 (success)
		`[{"*primary":{"Tenant":"cgrates.org","ID":"da39a3ee5e6b4b0d3255bfef95601890afd80709","Event":{"Account":"PARTNER_276011","AnswerTime":"2026-10-19T23:30:00Z","Category":"data","Destination":"internet","Direction":"*out","OriginHost":"cgrates-sgw","OriginID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","SetupTime":"2026-10-19T23:30:00Z","Source":"cgrates-sgw","Subject":"276011234567890","Tenant":"PARTNER_276011","Usage":"3305472","cellId":"360653210419331101","chargeableUnits":"3305000","chargingId":"1001","dataVolumeIncoming":"3004416","dataVolumeIncomingOrig":"3004000","dataVolumeOutgoing":"301056","dataVolumeOutgoingOrig":"301000","duration":"15300","imei":"3512057400000101","imsi":"276011234567890","inputFile":"file6.csv:1","mSTimeZone":"","msisdn":"15550000001","pGWAddress":"8700000000C0A80A2F","pdpAddress":"100.64.0.1|","qci":"9","sGWAddress":"10.1.0.10","tac":"1101"},"APIOpts":{"*cgrID":"09047612e75fb9e38c66dd73d6158911f8df3bd3","*ees":"false","*exporterID":"ur_exporter","*originID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","*partial":"false","*rates":"false","*runID":"*primary","*session":"true","*terminate":"true","*ur":true,"*urID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","CSVArrivalTime":"2026-10-20T03:45:00Z","NonUpdate":"true"}}},{"*primary":{"Tenant":"cgrates.org","ID":"da39a3ee5e6b4b0d3255bfef95601890afd80709","Event":{"Account":"PARTNER_276011","AnswerTime":"2026-10-20T04:10:00Z","Category":"data","Destination":"internet","Direction":"*out","OriginHost":"cgrates-sgw","OriginID":"1001_276011234567890_20261020_8700000000C0A80A2F_NewYork_9","SetupTime":"2026-10-20T04:10:00Z","Source":"cgrates-sgw","Subject":"276011234567890","Tenant":"PARTNER_276011","Usage":"3858432","cellId":"360653210419331101","chargeableUnits":3857700,"chargingId":"1001","dataVolumeIncoming":"3507200","dataVolumeIncomingOrig":3507000,"dataVolumeOutgoing":"351232","dataVolumeOutgoingOrig":350700,"duration":31800,"imei":"3512057400000101","imsi":"276011234567890","inputFile":"file6.csv:4","mSTimeZone":"","msisdn":"15550000001","pGWAddress":"8700000000C0A80A2F","pdpAddress":"100.64.0.1|","qci":"9","sGWAddress":"10.1.0.10","tac":"1101"},"APIOpts":{"*cgrID":"f903819ba37a2d34891503559a2c899a6b9296eb","*ees":"false","*exporterID":"ur_exporter","*originID":"1001_276011234567890_20261020_8700000000C0A80A2F_NewYork_9","*partial":"false","*rates":"false","*runID":"*primary","*session":"true","*terminate":"true","*ur":true,"*urID":"1001_276011234567890_20261020_8700000000C0A80A2F_NewYork_9","CSVArrivalTime":"2026-10-20T13:00:00Z","NonUpdate":"true"}}}]`,

		// case 7, idx 6 (success)
		`[{"*primary":{"Tenant":"cgrates.org","ID":"da39a3ee5e6b4b0d3255bfef95601890afd80709","Event":{"Account":"PARTNER_276011","AnswerTime":"2026-10-19T22:00:00Z","Category":"data","Destination":"internet","Direction":"*out","OriginHost":"cgrates-sgw","OriginID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","SetupTime":"2026-10-19T22:00:00Z","Source":"cgrates-sgw","Subject":"276011234567890","Tenant":"PARTNER_276011","Usage":"1195008","cellId":"360653210419331101","chargeableUnits":"1194000","chargingId":"1001","dataVolumeIncoming":"1103872","dataVolumeIncomingOrig":"1103000","dataVolumeOutgoing":"91136","dataVolumeOutgoingOrig":"91000","duration":"18000","imei":"3512057400000101","imsi":"276011234567890","inputFile":"file7.csv:1","mSTimeZone":"","msisdn":"15550000001","pGWAddress":"8700000000C0A80A2F","pdpAddress":"100.64.0.1|","qci":"9","sGWAddress":"10.1.0.10","tac":"1101"},"APIOpts":{"*cgrID":"09047612e75fb9e38c66dd73d6158911f8df3bd3","*ees":"false","*exporterID":"ur_exporter","*originID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","*partial":"false","*rates":"false","*runID":"*primary","*session":"true","*terminate":"true","*ur":true,"*urID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","CSVArrivalTime":"2026-10-20T03:00:00Z","NonUpdate":"true"}}},{"*primary":{"Tenant":"cgrates.org","ID":"da39a3ee5e6b4b0d3255bfef95601890afd80709","Event":{"Account":"PARTNER_276011","AnswerTime":"2026-10-20T10:00:00Z","Category":"data","Destination":"internet","Direction":"*out","OriginHost":"cgrates-sgw","OriginID":"1001_276011234567890_20261020_8700000000C0A80A2F_NewYork_9","SetupTime":"2026-10-20T10:00:00Z","Source":"cgrates-sgw","Subject":"276011234567890","Tenant":"PARTNER_276011","Usage":"4188160","cellId":"360653210419331101","chargeableUnits":"4188000","chargingId":"1001","dataVolumeIncoming":"3900416","dataVolumeIncomingOrig":"3900000","dataVolumeOutgoing":"288768","dataVolumeOutgoingOrig":"288000","duration":"86400","imei":"3512057400000101","imsi":"276011234567890","inputFile":"file7.csv:3","mSTimeZone":"","msisdn":"15550000001","pGWAddress":"8700000000C0A80A2F","pdpAddress":"100.64.0.1|","qci":"9","sGWAddress":"10.1.0.10","tac":"1101"},"APIOpts":{"*cgrID":"f903819ba37a2d34891503559a2c899a6b9296eb","*ees":"false","*exporterID":"ur_exporter","*originID":"1001_276011234567890_20261020_8700000000C0A80A2F_NewYork_9","*partial":"false","*rates":"false","*runID":"*primary","*session":"true","*terminate":"true","*ur":true,"*urID":"1001_276011234567890_20261020_8700000000C0A80A2F_NewYork_9","CSVArrivalTime":"2026-10-20T22:00:00Z"}}},{"*primary":{"Tenant":"cgrates.org","ID":"da39a3ee5e6b4b0d3255bfef95601890afd80709","Event":{"Account":"PARTNER_276011","AnswerTime":"2026-10-21T05:00:00Z","Category":"data","Destination":"internet","Direction":"*out","OriginHost":"cgrates-sgw","OriginID":"1001_276011234567890_20261021_8700000000C0A80A2F_NewYork_9","SetupTime":"2026-10-21T05:00:00Z","Source":"cgrates-sgw","Subject":"276011234567890","Tenant":"PARTNER_276011","Usage":"1604608","cellId":"360653210419331101","chargeableUnits":1604500,"chargingId":"1001","dataVolumeIncoming":"1506304","dataVolumeIncomingOrig":1506000,"dataVolumeOutgoing":"99328","dataVolumeOutgoingOrig":98500,"duration":25200,"imei":"3512057400000101","imsi":"276011234567890","inputFile":"file7.csv:6","mSTimeZone":"","msisdn":"15550000001","pGWAddress":"8700000000C0A80A2F","pdpAddress":"100.64.0.1|","qci":"9","sGWAddress":"10.1.0.10","tac":"1101"},"APIOpts":{"*cgrID":"36b8e1aee637a1b8cd1b5a9f32c2499f1c23132e","*ees":"false","*exporterID":"ur_exporter","*originID":"1001_276011234567890_20261021_8700000000C0A80A2F_NewYork_9","*partial":"false","*rates":"false","*runID":"*primary","*session":"true","*terminate":"true","*ur":true,"*urID":"1001_276011234567890_20261021_8700000000C0A80A2F_NewYork_9","CSVArrivalTime":"2026-10-21T12:00:00Z","NonUpdate":"true"}}}]`,

		// case 8, idx 7 (success)
		`[{"*primary":{"Tenant":"cgrates.org","ID":"da39a3ee5e6b4b0d3255bfef95601890afd80709","Event":{"Account":"PARTNER_276011","AnswerTime":"2026-10-19T13:00:00Z","Category":"data","Destination":"internet","Direction":"*out","OriginHost":"cgrates-sgw","OriginID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","SetupTime":"2026-10-19T13:00:00Z","Source":"cgrates-sgw","Subject":"276011234567890","Tenant":"PARTNER_276011","Usage":"2718720","cellId":"360653210419331101","chargeableUnits":2718000,"chargingId":"1001","dataVolumeIncoming":"2515968","dataVolumeIncomingOrig":2515000,"dataVolumeOutgoing":"203776","dataVolumeOutgoingOrig":203000,"duration":3600,"imei":"3512057400000101","imsi":"276011234567890","inputFile":"file8.csv:1","mSTimeZone":"","msisdn":"15550000001","pGWAddress":"8700000000C0A80A2F","pdpAddress":"100.64.0.1|","qci":"9","sGWAddress":"10.1.0.10","tac":"1101"},"APIOpts":{"*cgrID":"09047612e75fb9e38c66dd73d6158911f8df3bd3","*ees":"false","*exporterID":"ur_exporter","*originID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","*partial":"false","*rates":"false","*runID":"*primary","*session":"true","*terminate":"true","*ur":true,"*urID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","CSVArrivalTime":"2026-10-19T14:00:00Z","NonUpdate":"true"}}},{"*primary":{"Tenant":"cgrates.org","ID":"da39a3ee5e6b4b0d3255bfef95601890afd80709","Event":{"Account":"PARTNER_276011","AnswerTime":"2026-10-19T13:00:05Z","Category":"data","Destination":"ims","Direction":"*out","OriginHost":"cgrates-sgw","OriginID":"1002_276011234567890_20261019_8700000000C0A80A2F_NewYork_5","SetupTime":"2026-10-19T13:00:05Z","Source":"cgrates-sgw","Subject":"276011234567890","Tenant":"PARTNER_276011","Usage":"82944","cellId":"360653210419331101","chargeableUnits":82100,"chargingId":"1002","dataVolumeIncoming":"43008","dataVolumeIncomingOrig":42100,"dataVolumeOutgoing":"40960","dataVolumeOutgoingOrig":40000,"duration":3895,"imei":"3512057400000101","imsi":"276011234567890","inputFile":"file8.csv:2","mSTimeZone":"","msisdn":"15550000001","pGWAddress":"8700000000C0A80A2F","pdpAddress":"100.64.0.1|","qci":"5","sGWAddress":"10.1.0.10","tac":"1101"},"APIOpts":{"*cgrID":"077a78c502924f6341c812df6313f192aa87caad","*ees":"false","*exporterID":"ur_exporter","*originID":"1002_276011234567890_20261019_8700000000C0A80A2F_NewYork_5","*partial":"false","*rates":"false","*runID":"*primary","*session":"true","*terminate":"true","*ur":true,"*urID":"1002_276011234567890_20261019_8700000000C0A80A2F_NewYork_5","CSVArrivalTime":"2026-10-19T14:05:00Z","NonUpdate":"true"}}}]`,

		// case 9, idx 8 (success)
		`[{"*primary":{"Tenant":"cgrates.org","ID":"da39a3ee5e6b4b0d3255bfef95601890afd80709","Event":{"Account":"PARTNER_276011","AnswerTime":"2026-10-19T13:00:00Z","Category":"data","Destination":"internet","Direction":"*out","OriginHost":"cgrates-sgw","OriginID":"777_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","SetupTime":"2026-10-19T13:00:00Z","Source":"cgrates-sgw","Subject":"276011234567890","Tenant":"PARTNER_276011","Usage":"111616","cellId":"360653210419331101","chargeableUnits":111100,"chargingId":"777","dataVolumeIncoming":"101376","dataVolumeIncomingOrig":101000,"dataVolumeOutgoing":"10240","dataVolumeOutgoingOrig":10100,"duration":3600,"imei":"3512057400000101","imsi":"276011234567890","inputFile":"file9.csv:1","mSTimeZone":"","msisdn":"15550000001","pGWAddress":"8700000000C0A80A2F","pdpAddress":"100.64.0.1|","qci":"9","sGWAddress":"10.1.0.10","tac":"1101"},"APIOpts":{"*cgrID":"720cabaa2dbf5716cd154e14a7c5e7cae475482e","*ees":"false","*exporterID":"ur_exporter","*originID":"777_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","*partial":"false","*rates":"false","*runID":"*primary","*session":"true","*terminate":"true","*ur":true,"*urID":"777_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","CSVArrivalTime":"2026-10-19T14:00:00Z","NonUpdate":"true"}}},{"*primary":{"Tenant":"cgrates.org","ID":"da39a3ee5e6b4b0d3255bfef95601890afd80709","Event":{"Account":"PARTNER_276011","AnswerTime":"2026-10-19T13:00:00Z","Category":"data","Destination":"internet","Direction":"*out","OriginHost":"cgrates-sgw","OriginID":"777_276011234567891_20261019_8700000000C0A80A2F_NewYork_9","SetupTime":"2026-10-19T13:00:00Z","Source":"cgrates-sgw","Subject":"276011234567891","Tenant":"PARTNER_276011","Usage":"222208","cellId":"360653210419331101","chargeableUnits":222200,"chargingId":"777","dataVolumeIncoming":"202752","dataVolumeIncomingOrig":202000,"dataVolumeOutgoing":"20480","dataVolumeOutgoingOrig":20200,"duration":3600,"imei":"3512057400000202","imsi":"276011234567891","inputFile":"file9.csv:2","mSTimeZone":"","msisdn":"15550000002","pGWAddress":"8700000000C0A80A2F","pdpAddress":"100.64.0.2|","qci":"9","sGWAddress":"10.1.0.10","tac":"1101"},"APIOpts":{"*cgrID":"cec65f9a34d6c4072d407c0402fe8e1896bbe4df","*ees":"false","*exporterID":"ur_exporter","*originID":"777_276011234567891_20261019_8700000000C0A80A2F_NewYork_9","*partial":"false","*rates":"false","*runID":"*primary","*session":"true","*terminate":"true","*ur":true,"*urID":"777_276011234567891_20261019_8700000000C0A80A2F_NewYork_9","CSVArrivalTime":"2026-10-19T14:00:00Z","NonUpdate":"true"}}},{"*primary":{"Tenant":"cgrates.org","ID":"da39a3ee5e6b4b0d3255bfef95601890afd80709","Event":{"Account":"PARTNER_276011","AnswerTime":"2026-10-19T13:05:00Z","Category":"data","Destination":"internet","Direction":"*out","OriginHost":"cgrates-sgw","OriginID":"777_276011234567890_20261019_8700000000C0A80A3F_NewYork_9","SetupTime":"2026-10-19T13:05:00Z","Source":"cgrates-sgw","Subject":"276011234567890","Tenant":"PARTNER_276011","Usage":"333824","cellId":"360653210419331101","chargeableUnits":333300,"chargingId":"777","dataVolumeIncoming":"303104","dataVolumeIncomingOrig":303000,"dataVolumeOutgoing":"30720","dataVolumeOutgoingOrig":30300,"duration":3600,"imei":"3512057400000101","imsi":"276011234567890","inputFile":"file9.csv:3","mSTimeZone":"","msisdn":"15550000001","pGWAddress":"8700000000C0A80A3F","pdpAddress":"100.64.0.1|","qci":"9","sGWAddress":"10.1.0.10","tac":"1101"},"APIOpts":{"*cgrID":"ccc8d3ae2d082f66927483ce7326adb836f41573","*ees":"false","*exporterID":"ur_exporter","*originID":"777_276011234567890_20261019_8700000000C0A80A3F_NewYork_9","*partial":"false","*rates":"false","*runID":"*primary","*session":"true","*terminate":"true","*ur":true,"*urID":"777_276011234567890_20261019_8700000000C0A80A3F_NewYork_9","CSVArrivalTime":"2026-10-19T14:05:00Z","NonUpdate":"true"}}},{"*primary":{"Tenant":"cgrates.org","ID":"da39a3ee5e6b4b0d3255bfef95601890afd80709","Event":{"Account":"PARTNER_276011","AnswerTime":"2026-10-19T13:30:00Z","Category":"data","Destination":"internet","Direction":"*out","OriginHost":"cgrates-sgw","OriginID":"0_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","SetupTime":"2026-10-19T13:30:00Z","Source":"cgrates-sgw","Subject":"276011234567890","Tenant":"PARTNER_276011","Usage":"440320","cellId":"360653210419331101","chargeableUnits":"440000","chargingId":"0","dataVolumeIncoming":"400384","dataVolumeIncomingOrig":"400000","dataVolumeOutgoing":"40960","dataVolumeOutgoingOrig":"40000","duration":"86400","imei":"3512057400000101","imsi":"276011234567890","inputFile":"file9.csv:4","mSTimeZone":"","msisdn":"15550000001","pGWAddress":"8700000000C0A80A2F","pdpAddress":"100.64.0.1|","qci":"9","sGWAddress":"10.1.0.10","tac":"1101"},"APIOpts":{"*cgrID":"094f22404466194387ab48c09c917ee45698c489","*ees":"false","*exporterID":"ur_exporter","*originID":"0_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","*partial":"false","*rates":"false","*runID":"*primary","*session":"true","*terminate":"true","*ur":true,"*urID":"0_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","CSVArrivalTime":"2026-10-19T13:30:00Z"}}},{"*primary":{"Tenant":"cgrates.org","ID":"da39a3ee5e6b4b0d3255bfef95601890afd80709","Event":{"Account":"PARTNER_276011","AnswerTime":"2026-10-19T13:30:00Z","Category":"data","Destination":"internet","Direction":"*out","OriginHost":"cgrates-sgw","OriginID":"0_276011234567891_20261019_8700000000C0A80A2F_NewYork_9","SetupTime":"2026-10-19T13:30:00Z","Source":"cgrates-sgw","Subject":"276011234567891","Tenant":"PARTNER_276011","Usage":"550912","cellId":"360653210419331101","chargeableUnits":"550000","chargingId":"0","dataVolumeIncoming":"500736","dataVolumeIncomingOrig":"500000","dataVolumeOutgoing":"50176","dataVolumeOutgoingOrig":"50000","duration":"86400","imei":"3512057400000202","imsi":"276011234567891","inputFile":"file9.csv:5","mSTimeZone":"","msisdn":"15550000002","pGWAddress":"8700000000C0A80A2F","pdpAddress":"100.64.0.2|","qci":"9","sGWAddress":"10.1.0.10","tac":"1101"},"APIOpts":{"*cgrID":"71408fd5a03f3c185a8c6d0bd41de9df58cc50be","*ees":"false","*exporterID":"ur_exporter","*originID":"0_276011234567891_20261019_8700000000C0A80A2F_NewYork_9","*partial":"false","*rates":"false","*runID":"*primary","*session":"true","*terminate":"true","*ur":true,"*urID":"0_276011234567891_20261019_8700000000C0A80A2F_NewYork_9","CSVArrivalTime":"2026-10-19T13:30:00Z"}}}]`,

		// case 10, idx 9 (success only because there are no corresponding tac for 1024)
		`[{"*primary":{"Tenant":"cgrates.org","ID":"da39a3ee5e6b4b0d3255bfef95601890afd80709","Event":{"Account":"PARTNER_276011","AnswerTime":"2026-10-19T13:00:00Z","Category":"data","Destination":"internet","Direction":"*out","OriginHost":"cgrates-sgw","OriginID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","SetupTime":"2026-10-19T13:00:00Z","Source":"cgrates-sgw","Subject":"276011234567890","Tenant":"PARTNER_276011","Usage":"3249152","cellId":"360653210419331101","chargeableUnits":3248300,"chargingId":"1001","dataVolumeIncoming":"3007488","dataVolumeIncomingOrig":3007000,"dataVolumeOutgoing":"241664","dataVolumeOutgoingOrig":241300,"duration":5400,"imei":"3512057400000101","imsi":"276011234567890","inputFile":"file10.csv:1","mSTimeZone":"","msisdn":"15550000001","pGWAddress":"8700000000C0A80A2F","pdpAddress":"100.64.0.1|","qci":"9","sGWAddress":"10.1.0.10","tac":"1101"},"APIOpts":{"*cgrID":"09047612e75fb9e38c66dd73d6158911f8df3bd3","*ees":"false","*exporterID":"ur_exporter","*originID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","*partial":"false","*rates":"false","*runID":"*primary","*session":"true","*terminate":"true","*ur":true,"*urID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","CSVArrivalTime":"2026-10-19T14:30:00Z","NonUpdate":"true"}}}]`,

		// case 11, idx 10 (success)
		`[{"*primary":{"Tenant":"cgrates.org","ID":"da39a3ee5e6b4b0d3255bfef95601890afd80709","Event":{"Account":"PARTNER_276011","AnswerTime":"2026-10-20T02:00:00Z","Category":"data","Destination":"internet","Direction":"*out","OriginHost":"cgrates-sgw","OriginID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","SetupTime":"2026-10-20T02:00:00Z","Source":"cgrates-sgw","Subject":"276011234567890","Tenant":"PARTNER_276011","Usage":"1085440","cellId":"360653210419331101","chargeableUnits":"1085000","chargingId":"1001","dataVolumeIncoming":"1004544","dataVolumeIncomingOrig":"1004000","dataVolumeOutgoing":"81920","dataVolumeOutgoingOrig":"81000","duration":"5400","imei":"3512057400000101","imsi":"276011234567890","inputFile":"file11.csv:1","mSTimeZone":"","msisdn":"15550000001","pGWAddress":"8700000000C0A80A2F","pdpAddress":"100.64.0.1|","qci":"9","sGWAddress":"10.1.0.10","tac":"1101"},"APIOpts":{"*cgrID":"09047612e75fb9e38c66dd73d6158911f8df3bd3","*ees":"false","*exporterID":"ur_exporter","*originID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","*partial":"false","*rates":"false","*runID":"*primary","*session":"true","*terminate":"true","*ur":true,"*urID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","CSVArrivalTime":"2026-10-20T03:30:00Z","NonUpdate":"true"}}},{"*primary":{"Tenant":"cgrates.org","ID":"da39a3ee5e6b4b0d3255bfef95601890afd80709","Event":{"Account":"PARTNER_276011","AnswerTime":"2026-10-20T04:15:00Z","Category":"data","Destination":"internet","Direction":"*out","OriginHost":"cgrates-sgw","OriginID":"1001_276011234567890_20261019_8700000000C0A80A2F_Chicago_9","SetupTime":"2026-10-20T04:15:00Z","Source":"cgrates-sgw","Subject":"276011234567890","Tenant":"PARTNER_276011","Usage":"2163712","cellId":"360653230700012001","chargeableUnits":2163300,"chargingId":"1001","dataVolumeIncoming":"2003968","dataVolumeIncomingOrig":2003000,"dataVolumeOutgoing":"160768","dataVolumeOutgoingOrig":160300,"duration":1800,"imei":"3512057400000101","imsi":"276011234567890","inputFile":"file11.csv:3","mSTimeZone":"","msisdn":"15550000001","pGWAddress":"8700000000C0A80A2F","pdpAddress":"100.64.0.1|","qci":"9","sGWAddress":"10.1.0.10","tac":"2001"},"APIOpts":{"*cgrID":"7dd71ab560034cb769bdf022b7d448d211eb2145","*ees":"false","*exporterID":"ur_exporter","*originID":"1001_276011234567890_20261019_8700000000C0A80A2F_Chicago_9","*partial":"false","*rates":"false","*runID":"*primary","*session":"true","*terminate":"true","*ur":true,"*urID":"1001_276011234567890_20261019_8700000000C0A80A2F_Chicago_9","CSVArrivalTime":"2026-10-20T04:45:00Z","NonUpdate":"true"}}}]`,

		// case 12, idx 11
		`[]`, // only 1st csv is read, no result yet

		// case 12, idx 12 (success)
		`[{"*primary":{"Tenant":"cgrates.org","ID":"da39a3ee5e6b4b0d3255bfef95601890afd80709","Event":{"Account":"PARTNER_276011","AnswerTime":"2026-10-19T13:00:00Z","Category":"data","Destination":"internet","Direction":"*out","OriginHost":"cgrates-sgw","OriginID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","SetupTime":"2026-10-19T13:00:00Z","Source":"cgrates-sgw","Subject":"276011234567890","Tenant":"PARTNER_276011","Usage":"1420288","cellId":"360653210419331101","chargeableUnits":"1419500","chargingId":"1001","dataVolumeIncoming":"1308672","dataVolumeIncomingOrig":"1308000","dataVolumeOutgoing":"111616","dataVolumeOutgoingOrig":"111500","duration":"10800","imei":"3512057400000101","imsi":"276011234567890","inputFile":"file13.csv:1","mSTimeZone":"","msisdn":"15550000001","pGWAddress":"8700000000C0A80A2F","pdpAddress":"100.64.0.1|","qci":"9","sGWAddress":"10.1.0.10","tac":"1101"},"APIOpts":{"*cgrID":"09047612e75fb9e38c66dd73d6158911f8df3bd3","*ees":"false","*exporterID":"ur_exporter","*originID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","*partial":"false","*rates":"false","*runID":"*primary","*session":"true","*terminate":"true","*ur":true,"*urID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","CSVArrivalTime":"2026-10-19T14:00:00Z","NonUpdate":"true"}}}]`,

		// case 13, idx 13 (success)
		`[{"*primary":{"Tenant":"cgrates.org","ID":"da39a3ee5e6b4b0d3255bfef95601890afd80709","Event":{"Account":"PARTNER_276011","AnswerTime":"2026-10-19T13:00:00Z","Category":"data","Destination":"internet","Direction":"*out","OriginHost":"cgrates-sgw","OriginID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","SetupTime":"2026-10-19T13:00:00Z","Source":"cgrates-sgw","Subject":"276011234567890","Tenant":"PARTNER_276011","Usage":"7168","cellId":"360653210419331101","chargeableUnits":7000,"chargingId":"1001","dataVolumeIncoming":"3072","dataVolumeIncomingOrig":3000,"dataVolumeOutgoing":"4096","dataVolumeOutgoingOrig":4000,"duration":3600,"imei":"3512057400000101","imsi":"276011234567890","inputFile":"file14.csv:2","mSTimeZone":"","msisdn":"15550000001","pGWAddress":"8700000000C0A80A2F","pdpAddress":"100.64.0.1|","qci":"9","sGWAddress":"10.1.0.10","tac":"1101"},"APIOpts":{"*cgrID":"09047612e75fb9e38c66dd73d6158911f8df3bd3","*ees":"false","*exporterID":"ur_exporter","*originID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","*partial":"false","*rates":"false","*runID":"*primary","*session":"true","*terminate":"true","*ur":true,"*urID":"1001_276011234567890_20261019_8700000000C0A80A2F_NewYork_9","CSVArrivalTime":"2026-10-19T14:00:00Z","NonUpdate":"true"}}}]`,
	}

	ng := engine.TestEngine{
		ConfigJSON: fmt.Sprintf(`{
"logger": {
    "level": 7
},
"admins": {
	"enabled": true,
	"conns": {
		"*ers": [{"connIDs": ["*localhost"]}]
	}
},
"rates": {
	"enabled": true
},
"sessions": {
	"enabled": true,
	"conns":{
		"*ees": [{"connIDs": ["*localhost"]}],
	},
},
"ees": {
	"enabled": true,
	"exporters": [{
		"id": "ur_exporter",
		"type": "*cgrur",
		"exportPath": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
       	"flags": ["*log"],
		"opts": {
			"sqlDBName": "cgrates",
			"sqlTableName": "urs"
		},
	}]
},
"caches":{
	"partitions": {
		// make cache equal to partialCacheTTL
		"*uch": {"limit": -1, "ttl": "1s", "staticTTL": false},
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
			"id": "GetURs",
			"runDelay": "0",
			"type": "*cgrur",
			"sourcePath": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
			"processedPath": "",
			"flags": ["*sessions"],
			"fields":[
				{"tag": "NoRate", "path": "*opts.*rates", "type": "*constant", "value": "false"},
				{"tag": "NoEEs", "path": "*opts.*ees", "type": "*constant", "value": "false"},
			],
			"opts": {
				"sqlDBName": "cgrates",
				"sqlTableName": "urs"
			},
		},
		{
			"id": "ReadURCSVs",
			"concurrentRequests": 1,
			"type": "*fileCSV",
			"sourcePath": "%s",              // templated: input directory
			"processedPath": "%s",           // templated: where finished files are moved
			"startDelay": "100ms",
			"flags": ["*log", "*sessions"],  // *sessions = send events to SessionS
		
			// ------------------------------------------------------------------
			// ROW FILTERS - a row must pass EVERY filter or it is skipped
			// completely (no fields run, nothing is cached or summed).
			// ------------------------------------------------------------------
			"filters": [
				"*regex:~*req.0:^[0-9]+$",        // keeps only rows whose time is all digits -> drops the CSV header
				"*notempty:~*req.1:",             // drops rows with a blank IMSI
				"*notprefix:~*req.1:313380",      // drops home subscribers (IMSI prefix). Add more prefixes after ';'
				"*notprefix:~*req.12:sos",        // drops emergency APNs
				// Volume must be present and under 50 GiB (53687091200 bytes) in both directions.
				"*notempty:~*req.14:", "*lt:~*req.14:53687091200",
				"*notempty:~*req.15:", "*lt:~*req.15:53687091200",
				// Drops rows older than 30 days
				"*gte:~*req.0:-720h{*unixtime}",
			],
		
			"fields":[
				// ============================================================
				// SITE / TIMEZONE LOOKUP -> OriginID (the grouping key)
				// ============================================================
				// TAC -> timezone table. "long" = IANA name used to compute the local
				// date, "short" = label written into the OriginID.
				// EVERY TAC of a site must be listed, otherwise rows with an unmapped
				// TAC get a different OriginID and the session is split in two.
				{"tag": "NYCTime", "path": "*vars.1101.long", "type": "*constant",
					"value": "America/New_York"},
				{"tag": "NYCShort", "path": "*vars.1101.short", "type": "*constant",
					"value": "NewYork"},
				{"tag": "ChicagoTime", "path": "*vars.2001.long", "type": "*constant",
					"value": "America/Chicago"},
				{"tag": "ChicagoShort", "path": "*vars.2001.short", "type": "*constant",
					"value": "Chicago"},
		
				// Looks up the table above using this row's TAC.
				{"tag": "TimezoneShort", "path": "*vars.shorttimezone", "type": "*variable",
					"value": "~*vars.<~*req.7>.short"},
				{"tag": "TimezoneLong", "path": "*vars.longtimezone", "type": "*variable",
					"value": "~*vars.<~*req.7>.long"},
		
				// Local calendar date (YYYYMMDD) of the row's unix time in the site timezone.
				// This is what makes a session "end" at LOCAL midnight instead of UTC midnight.
				{"tag": "LocalDate", "path": "*vars.LocalDate", "type": "*variable",
					"value": "~*req.0{*timezone:<~*vars.longtimezone>&*time:20060102}"},
				// Alternative syntax for the same thing:
				// {"tag": "LocalDate", "path": "*vars.LocalDate", "type": "*variable",
				// 	"value": "*req.0{*time:20060102:<*vars.longtimezone>}"},
		
				// The grouping key. Rows with the same value are merged into one CDR.
				// chargingId _ imsi _ localDate _ pgwIP _ siteName _ qci
				{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable",
					"value": "~*req.3;_;~*req.1;_;~*vars.LocalDate;_;~*req.11;_<~*vars.shorttimezone>_;~*req.13"},
		
				// ============================================================
				// "FIRST ROW BY TIME" VALUES (tac, cellId, inputFile)
				// The CDR must carry these from the EARLIEST row, even if rows arrive
				// out of order. Pattern: write on the first row seen, then overwrite
				// whenever a row with an earlier timestamp shows up.
				// ORDER MATTERS: these must stay ABOVE the EarliestCSVTime fields. 
				// The "...Earlier" fields compare this row against the
				// stored earliest time.
				// ============================================================
				// TAC
				{"tag": "tacUCH", "path": "*uch.<~*cgreq.OriginID>tacUCH", "type": "*variable",
					"filters": ["*notexists:~*uch.<~*cgreq.OriginID>tacUCH:"], "value": "~*req.7"},          // first row seen
				{"tag": "tacUCHEarlier", "path": "*uch.<~*cgreq.OriginID>tacUCH", "type": "*variable",
					"filters": ["*exists:~*uch.<~*cgreq.OriginID>EarliestCSVTime:", "*gt:~*uch.<~*cgreq.OriginID>EarliestCSVTime:~*req.0"],
					"value": "~*req.7"},                                                                     // this row is earlier than the stored earliest
		
				// Cell id = plmn + eci + tac, concatenated
				{"tag": "cellIdUCH", "path": "*uch.<~*cgreq.OriginID>cellIdUCH", "type": "*variable",
					"filters": ["*notexists:~*uch.<~*cgreq.OriginID>cellIdUCH:"], "value": "~*req.6;~*req.8;~*req.7"},
				{"tag": "cellIdUCHEarlier", "path": "*uch.<~*cgreq.OriginID>cellIdUCH", "type": "*variable",
					"filters": ["*exists:~*uch.<~*cgreq.OriginID>EarliestCSVTime:", "*gt:~*uch.<~*cgreq.OriginID>EarliestCSVTime:~*req.0"],
					"value": "~*req.6;~*req.8;~*req.7"},
		
				// Source file name and line number ("file.csv:3") of the earliest row
				{"tag": "InputFileUCH", "path": "*uch.<~*cgreq.OriginID>InputFileUCH", "type": "*variable",
					"filters": ["*notexists:~*uch.<~*cgreq.OriginID>InputFileUCH:"],
					"value": "~*vars.*fileName;:;~*vars.*fileLineNumber"},
				{"tag": "InputFileUCHEarlier", "path": "*uch.<~*cgreq.OriginID>InputFileUCH", "type": "*variable",
					"filters": ["*exists:~*uch.<~*cgreq.OriginID>EarliestCSVTime:", "*gt:~*uch.<~*cgreq.OriginID>EarliestCSVTime:~*req.0"],
					"value": "~*vars.*fileName;:;~*vars.*fileLineNumber"},
		
				// ============================================================
				// EARLIEST / LATEST ROW TIME -> SetupTime, AnswerTime, duration
				// Stored as unix time, independent of the order rows arrive in.
				// ============================================================
				{"tag": "EarliestCSVTime", "path": "*uch.<~*cgreq.OriginID>EarliestCSVTime", "type": "*variable",
					"filters": ["*notexists:~*uch.<~*cgreq.OriginID>EarliestCSVTime:"], "value": "~*req.0"},  // first row seen
				{"tag": "EarliestCSVTimeEarlier", "path": "*uch.<~*cgreq.OriginID>EarliestCSVTime", "type": "*variable",
					"filters": ["*gt:~*uch.<~*cgreq.OriginID>EarliestCSVTime:~*req.0"], "value": "~*req.0"},  // stored value is later than this row
				{"tag": "LatestCSVTime", "path": "*uch.<~*cgreq.OriginID>LatestCSVTime", "type": "*variable",
					"filters": ["*notexists:~*uch.<~*cgreq.OriginID>LatestCSVTime:"], "value": "~*req.0"},    // first row seen
				{"tag": "LatestCSVTimeLater", "path": "*uch.<~*cgreq.OriginID>LatestCSVTime", "type": "*variable",
					"filters": ["*lt:~*uch.<~*cgreq.OriginID>LatestCSVTime:~*req.0"], "value": "~*req.0"},    // stored value is earlier than this row
		
				// Setup and answer time are both the earliest row time
				{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*variable",
					"value": "~*uch.<~*cgreq.OriginID>EarliestCSVTime{*time}"},
				{"tag": "AnswerTime", "path": "*cgreq.AnswerTime", "type": "*variable",
					"value": "~*uch.<~*cgreq.OriginID>EarliestCSVTime{*time}"},
				// duration (seconds) = latest row - earliest row
				{"tag": "duration", "path": "*cgreq.duration", "type": "*difference",
					"value": "~*uch.<~*cgreq.OriginID>LatestCSVTime;~*uch.<~*cgreq.OriginID>EarliestCSVTime"},
		
				// ============================================================
				// CONTROL FLAGS AND FIXED EVENT FIELDS
				// ============================================================
				// Marks the session as containing at least one start/end row. If it
				// never gets set, the session was updates only, and cacheDumpFields
				// forces duration = 86400.
				{"tag": "NonUpdate", "path": "*opts.NonUpdate", "type": "*constant",
					"filters": ["*notstring:~*req.2:default_bearer_update"],
					"value": "true"},
				{"tag": "OptSession", "path": "*opts.*session", "type": "*constant", "value": "true"},
				{"tag": "Direction", "path": "*cgreq.Direction", "type": "*constant", "value": "*out"},
				{"tag": "Category", "path": "*cgreq.Category", "type": "*constant", "value": "data"},
				// Tenant/Account = "PARTNER_" + first 6 digits of the IMSI
				{"tag": "Tenant", "path": "*cgreq.Tenant", "type": "*variable",
					"value": "PARTNER_;~*req.1{*strip:*suffix:9}"},
				{"tag": "Account", "path": "*cgreq.Account", "type": "*variable",
					"value": "PARTNER_;~*req.1{*strip:*suffix:9}"},
				{"tag": "Subject", "path": "*cgreq.Subject", "type": "*variable", "value": "~*req.1"},
				{"tag": "OriginHost", "path": "*cgreq.OriginHost", "type": "*constant", "value": "cgrates-sgw"},
				{"tag": "Source", "path": "*cgreq.Source", "type": "*constant", "value": "cgrates-sgw"},
				{"tag": "Destination", "path": "*cgreq.Destination", "type": "*variable", "value": "~*req.12"},   // APN
				{"tag": "imsi", "path": "*cgreq.imsi", "type": "*variable", "value": "~*req.1"},
				{"tag": "msisdn", "path": "*cgreq.msisdn", "type": "*variable", "value": "~*req.4"},
				{"tag": "imei", "path": "*cgreq.imei", "type": "*variable", "value": "~*req.5"},
				{"tag": "chargingId", "path": "*cgreq.chargingId", "type": "*variable", "value": "~*req.3"},
				// From the earliest row
				{"tag": "tac", "path": "*cgreq.tac", "type": "*variable", "value": "~*uch.<~*cgreq.OriginID>tacUCH"},
				{"tag": "cellId", "path": "*cgreq.cellId", "type": "*variable", "value": "~*uch.<~*cgreq.OriginID>cellIdUCH"},
				{"tag": "inputFile", "path": "*cgreq.inputFile", "type": "*variable", "value": "~*uch.<~*cgreq.OriginID>InputFileUCH"},
				{"tag": "mSTimeZone", "path": "*cgreq.mSTimeZone", "type": "*constant", "value": ""},
				{"tag": "sGWAddress", "path": "*cgreq.sGWAddress", "type": "*variable", "value": "~*req.9"},
				{"tag": "pGWAddress", "path": "*cgreq.pGWAddress", "type": "*variable", "value": "~*req.11"},
				{"tag": "pdpAddress", "path": "*cgreq.pdpAddress", "type": "*variable", "value": "~*req.10"},
				{"tag": "qci", "path": "*cgreq.qci", "type": "*variable", "value": "~*req.13"},
		
				// ============================================================
				// RUNNING VOLUME TOTALS (raw bytes, not rounded)
				// Pattern per direction:
				//   1. first row of the session: take this row's value
				//   2. later rows: this row + stored total
				//   3. write the new total back to *uch
				// ============================================================
				// Incoming
				{"tag": "dataVolumeIncomingOrig", "path": "*cgreq.dataVolumeIncomingOrig", "type": "*variable",
					"filters": ["*notexists:~*uch.<~*cgreq.OriginID>UCHdataVolumeIncomingOrig:"],
					"value": "~*req.14"},
				{"tag": "dataVolumeIncomingOrig", "path": "*cgreq.dataVolumeIncomingOrig", "type": "*sum",
					"filters": ["*exists:~*uch.<~*cgreq.OriginID>UCHdataVolumeIncomingOrig:"],
					"value": "~*req.14;~*uch.<~*cgreq.OriginID>UCHdataVolumeIncomingOrig"},
				{"tag": "UCHdataVolumeIncomingOrig", "path": "*uch.<~*cgreq.OriginID>UCHdataVolumeIncomingOrig", "type": "*variable",
					"value": "~*cgreq.dataVolumeIncomingOrig"},
		
				// Outgoing
				{"tag": "dataVolumeOutgoingOrig", "path": "*cgreq.dataVolumeOutgoingOrig", "type": "*variable",
					"filters": ["*notexists:~*uch.<~*cgreq.OriginID>UCHdataVolumeOutgoingOrig:"],
					"value": "~*req.15"},
				{"tag": "dataVolumeOutgoingOrig", "path": "*cgreq.dataVolumeOutgoingOrig", "type": "*sum",
					"filters": ["*exists:~*uch.<~*cgreq.OriginID>UCHdataVolumeOutgoingOrig:"],
					"value": "~*req.15;~*uch.<~*cgreq.OriginID>UCHdataVolumeOutgoingOrig"},
				{"tag": "UCHdataVolumeOutgoingOrig", "path": "*uch.<~*cgreq.OriginID>UCHdataVolumeOutgoingOrig", "type": "*variable",
					"value": "~*cgreq.dataVolumeOutgoingOrig"},
		
				// Total raw bytes (in + out). chargeableUnits > 0 is also what decides
				// whether a UR is stored (see partialCommitFields / cacheDumpFields).
				{"tag": "chargeableUnits", "path": "*cgreq.chargeableUnits", "type": "*sum",
					"value": "~*cgreq.dataVolumeIncomingOrig;~*cgreq.dataVolumeOutgoingOrig"},
		
				// ============================================================
				// ROUNDING (round UP to the partner increment, 1024 bytes)
				// Always applied to the TOTALS, never to a single row, so the last event
				// of the session carries the correct rounded values. Usage is rounded
				// from the raw sum, not from the sum of the two rounded numbers.
				// ============================================================
				{"tag": "dataVolumeIncoming", "path": "*cgreq.dataVolumeIncoming", "type": "*variable",
					"value": "~*cgreq.dataVolumeIncomingOrig{*divide:1024&*round:0:*up&*multiply:1024}"},
				{"tag": "dataVolumeOutgoing", "path": "*cgreq.dataVolumeOutgoing", "type": "*variable",
					"value": "~*cgreq.dataVolumeOutgoingOrig{*divide:1024&*round:0:*up&*multiply:1024}"},
				{"tag": "Usage", "path": "*cgreq.Usage", "type": "*variable",
					"value": "~*cgreq.chargeableUnits{*divide:1024&*round:0:*up&*multiply:1024}"},
		
				// ============================================================
				// PARTIAL-RECORD CONTROL
				// ============================================================
				// Every row is a partial by default. partialCommitFields flips it to
				// false when the session ends.
				{"tag": "Partial", "path": "*opts.*partial", "type": "*constant", "value": "true"},
				// A default_bearer_end row triggers the commit (CDR is exported).
				{"tag": "OptTerminate", "path": "*opts.*terminate", "type": "*constant",
					"filters": ["*string:~*req.2:default_bearer_end"], "value": "true"},
				// Key under which partial events are cached and merged. Must equal OriginID.
				{"tag": "MetaOriginID", "path": "*opts.*originID", "type": "*variable", "value": "~*cgreq.OriginID"},
				// Row time, used only to order the cached partials before merging
				// (see partialOrderField). Kept in opts so it does not end up in the CDR event.
				{"tag": "CSVArrivalTime", "path": "*opts.CSVArrivalTime", "type": "*variable", "value": "~*req.0{*time}"},
			],
		
			// ------------------------------------------------------------------
			// Runs when the session ENDS (an end row arrived): the cached partials are
			// merged and these fields are applied to the merged event.
			// ------------------------------------------------------------------
			"partialCommitFields": [
				// Merged event is complete, not partial
				{"tag": "Partial", "path": "*opts.*partial", "type": "*constant",
					"value": "false", "filters": ["*exists:~*opts.*terminate:"]},
				// Store the usage record and export it, but only if there was usage.
				// Zero-usage sessions are therefore not exported.
				{"tag": "OptUsageRecord", "path": "*opts.*ur", "type": "*constant",
					"filters": ["*exists:~*opts.*terminate:", "*gt:~*req.chargeableUnits:0"], "value": "true"},
				{"tag": "OptEEs", "path": "*opts.*ees", "type": "*constant",
					"filters": ["*exists:~*opts.*terminate:", "*gt:~*req.chargeableUnits:0"], "value": "true"},
			],
		
			// ------------------------------------------------------------------
			// Runs when a cached partial EXPIRES without an end row (partialCacheTTL).
			// It rebuilds the event from the cached one (~*req.<field>) and posts it
			// as a complete CDR. These are the sessions with no default_bearer_end.
			// ------------------------------------------------------------------
			"cacheDumpFields": [
				// Row-time ordering key
				{"tag": "CSVArrivalTime", "path": "*opts.CSVArrivalTime", "type": "*variable", "value": "~*opts.CSVArrivalTime"},
				{"tag": "duration", "path": "*cgreq.duration", "type": "*variable", "value": "~*req.duration"},
				// Updates-only session (NonUpdate never set): bill a full day
				{"tag": "DurationFullDay", "path": "*cgreq.duration", "type": "*constant",
					"filters": ["*notexists:~*opts.NonUpdate:"], "value": "86400"},
				{"tag": "OptSession", "path": "*opts.*session", "type": "*constant", "value": "true"},
				{"tag": "Direction", "path": "*cgreq.Direction", "type": "*constant", "value": "*out"},
				{"tag": "Category", "path": "*cgreq.Category", "type": "*constant", "value": "data"},
				{"tag": "Tenant", "path": "*cgreq.Tenant", "type": "*variable", "value": "~*req.Tenant"},
				{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", "value": "~*req.Account"},
				{"tag": "Subject", "path": "*cgreq.Subject", "type": "*variable", "value": "~*req.Subject"},
				{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", "value": "~*req.OriginID"},
				{"tag": "OriginHost", "path": "*cgreq.OriginHost", "type": "*constant", "value": "cgrates-sgw"},
				{"tag": "Source", "path": "*cgreq.Source", "type": "*constant", "value": "cgrates-sgw"},
				{"tag": "Destination", "path": "*cgreq.Destination", "type": "*variable", "value": "~*req.Destination"},
				{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*variable", "value": "~*req.SetupTime"},
				{"tag": "AnswerTime", "path": "*cgreq.AnswerTime", "type": "*variable", "value": "~*req.AnswerTime"},
				{"tag": "imsi", "path": "*cgreq.imsi", "type": "*variable", "value": "~*req.imsi"},
				{"tag": "msisdn", "path": "*cgreq.msisdn", "type": "*variable", "value": "~*req.msisdn"},
				{"tag": "imei", "path": "*cgreq.imei", "type": "*variable", "value": "~*req.imei"},
				{"tag": "chargingId", "path": "*cgreq.chargingId", "type": "*variable", "value": "~*req.chargingId"},
				{"tag": "tac", "path": "*cgreq.tac", "type": "*variable", "value": "~*req.tac"},
				{"tag": "cellId", "path": "*cgreq.cellId", "type": "*variable", "value": "~*req.cellId"},
				{"tag": "mSTimeZone", "path": "*cgreq.mSTimeZone", "type": "*constant", "value": ""},
				{"tag": "sGWAddress", "path": "*cgreq.sGWAddress", "type": "*variable", "value": "~*req.sGWAddress"},
				{"tag": "pGWAddress", "path": "*cgreq.pGWAddress", "type": "*variable", "value": "~*req.pGWAddress"},
				{"tag": "pdpAddress", "path": "*cgreq.pdpAddress", "type": "*variable", "value": "~*req.pdpAddress"},
				{"tag": "qci", "path": "*cgreq.qci", "type": "*variable", "value": "~*req.qci"},
				// Rounded volumes and usage, already computed per row
				{"tag": "dataVolumeIncoming", "path": "*cgreq.dataVolumeIncoming", "type": "*variable", "value": "~*req.dataVolumeIncoming"},
				{"tag": "dataVolumeOutgoing", "path": "*cgreq.dataVolumeOutgoing", "type": "*variable", "value": "~*req.dataVolumeOutgoing"},
				{"tag": "Usage", "path": "*cgreq.Usage", "type": "*variable", "value": "~*req.Usage"},
				// Raw volumes
				{"tag": "dataVolumeIncomingOrig", "path": "*cgreq.dataVolumeIncomingOrig", "type": "*variable", "value": "~*req.dataVolumeIncomingOrig"},
				{"tag": "dataVolumeOutgoingOrig", "path": "*cgreq.dataVolumeOutgoingOrig", "type": "*variable", "value": "~*req.dataVolumeOutgoingOrig"},
				{"tag": "chargeableUnits", "path": "*cgreq.chargeableUnits", "type": "*variable", "value": "~*req.chargeableUnits"},
				{"tag": "inputFile", "path": "*cgreq.inputFile", "type": "*variable", "value": "~*req.inputFile"},
				// Dumped record is complete (not partial) and is terminated
				{"tag": "Partial", "path": "*opts.*partial", "type": "*constant", "value": "false"},
				{"tag": "OptTerminate", "path": "*opts.*terminate", "type": "*constant", "value": "true"},
				{"tag": "MetaOriginID", "path": "*opts.*originID", "type": "*variable", "value": "~*opts.*originID"},
				// Store/export only if there was usage
				{"tag": "OptUsageRecord", "path": "*opts.*ur", "type": "*constant",
					"filters": ["*gt:~*req.chargeableUnits:0"], "value": "true"},
				{"tag": "OptEEs", "path": "*opts.*ees", "type": "*constant",
					"filters": ["*gt:~*req.chargeableUnits:0"], "value": "true"},
			],
		
			"opts": {
				// What to do with cached partials that expire without an end row:
				// *processEvent = run cacheDumpFields and post them as CDRs
				"partialCacheAction": "*processEvent",
				// Cached partials are sorted by this before merging, so that the latest
				// row in time wins when fields conflict
				"partialOrderField": "~*opts.CSVArrivalTime",
				"partialPath": "%s",             // templated: directory used for partial dumps
			}
		}
	]
}
}`, ersIn, ersOut, prtPath),
		TpPath:           filepath.Join(*utils.DataDir, "tariffplans", "tutorial"),
		DBCfg:            getDBCfg(t),
		LogBuffer:        &bytes.Buffer{},
		GracefulShutdown: true,
	}
	// use tutmysql config only for db info to flush after test
	tmpSQLCfg, err := config.NewCGRConfigFromPath(context.TODO(), filepath.Join(*utils.DataDir, "conf", "samples", "tutmysql"))
	if err != nil {
		t.Fatalf("could not init config from path %s: %v", filepath.Join(*utils.DataDir, "conf", "samples", "tutmysql"), err)
	}
	if err := engine.InitDB(tmpSQLCfg); err != nil { // init and flush sql db
		t.Fatalf("failed to flush DataDB err: %v", err)
	}
	t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
	client, _ := ng.Run(t)
	time.Sleep(200 * time.Millisecond)
	for i, singleCase := range cases {
		createFile(t, ersIn, utils.CSVSuffix, singleCase)
		var rply string
		if err := client.Call(context.Background(), utils.ErSv1RunReader,
			&ers.V1RunReaderParams{
				ReaderID: "ReadURCSVs",
			}, &rply); err != nil {
			t.Fatal(err)
		}
		if rply != utils.OK {
			t.Errorf("expected <OK>, received <%v>", rply)
		}
		if i == 11 { // for 12th case, multiple files are read, so dont wait for cache to expire
			continue
		}
		time.Sleep(3 * time.Second) // give time for cacheDumpFields ttl
		var rows []map[string]*utils.CGREvent
		if err := client.Call(context.Background(), utils.AdminSv1GetURs,
			&apis.URFiltersFromReader{
				ReaderID: "GetURs",
			}, &rows); err != nil {
			t.Fatal(err)
		}

		// sort for consistancy in testing based on inputFile string
		sort.Slice(rows, func(i, j int) bool {
			eventI := rows[i][utils.MetaPrimary]
			eventJ := rows[j][utils.MetaPrimary]
			return eventI.Event["inputFile"].(string) < eventJ.Event["inputFile"].(string)
		})
		rcv := utils.ToJSON(rows)
		if expectedRows[i] != rcv {
			t.Errorf("idx %v expected <%v>, \nreceived <%v>", i, expectedRows[i], rcv)
			break
		}

		if err := engine.InitDB(tmpSQLCfg); err != nil { // flush sql db since some origin ids are the same in some cases
			t.Fatalf("failed to flush DataDB err: %v", err)
		}

		// t.Log(utils.ToJSON(rows))
	}

}
