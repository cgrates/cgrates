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
				{"tag": "CSVArrivalTime", "path": "*cgreq.CSVArrivalTime", "type": "*variable", "value": "~*req.{{TIME}}{*timeString}"},
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
					"filters": ["*notexists:~*uch.SetupTimeUCH:"], "value": "~*req.{{TIME}}{*timeString}"},
				{"tag": "SetupTimeUCHUpdated", "path": "*uch.SetupTimeUCH", "type": "*variable",
					"filters": ["*gt:~*uch.SetupTimeUCH:~*req.{{TIME}}{*timeString}"], "value": "~*uch.SetupTimeUCH"},
				{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*variable",
					 "value": "~*uch.SetupTimeUCH"},
				{"tag": "AnswerTimeUCH", "path": "*uch.AnswerTimeUCH", "type": "*variable",
					"filters": ["*notexists:~*uch.AnswerTimeUCH:"], "value": "~*req.{{TIME}}{*timeString}"},
				{"tag": "AnswerTimeUCHUpdated", "path": "*uch.AnswerTimeUCH", "type": "*variable",
					"filters": ["*gt:~*uch.AnswerTimeUCH:~*req.{{TIME}}{*timeString}"], "value": "~*uch.AnswerTimeUCH"},
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
				{"tag": "CSVArrivalTime", "path": "*cgreq.CSVArrivalTime", "type": "*variable", "value": "~*req.{{TIME}}{*timeString}"},
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
					"filters": ["*notexists:~*uch.SetupTimeUCH:"], "value": "~*req.{{TIME}}{*timeString}"},
				{"tag": "SetupTimeUCHUpdated", "path": "*uch.SetupTimeUCH", "type": "*variable",
					"filters": ["*gt:~*uch.SetupTimeUCH:~*req.{{TIME}}{*timeString}"], "value": "~*uch.SetupTimeUCH"},
				{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*variable",
					 "value": "~*uch.SetupTimeUCH"},
				{"tag": "AnswerTimeUCH", "path": "*uch.AnswerTimeUCH", "type": "*variable",
					"filters": ["*notexists:~*uch.AnswerTimeUCH:"], "value": "~*req.{{TIME}}{*timeString}"},
				{"tag": "AnswerTimeUCHUpdated", "path": "*uch.AnswerTimeUCH", "type": "*variable",
					"filters": ["*gt:~*uch.AnswerTimeUCH:~*req.{{TIME}}{*timeString}"], "value": "~*uch.AnswerTimeUCH"},
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
							utils.MetaRates:           "true",
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
				{"tag": "CSVArrivalTime", "path": "*cgreq.CSVArrivalTime", "type": "*variable", "value": "~*req.{{TIME}}{*timeString}"},
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
					"filters": ["*notexists:~*uch.SetupTimeUCH:"], "value": "~*req.{{TIME}}{*timeString}"},
				{"tag": "SetupTimeUCHUpdated", "path": "*uch.SetupTimeUCH", "type": "*variable",
					"filters": ["*gt:~*uch.SetupTimeUCH:~*req.{{TIME}}{*timeString}"], "value": "~*uch.SetupTimeUCH"},
				{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*variable",
					 "value": "~*uch.SetupTimeUCH"},
				{"tag": "AnswerTimeUCH", "path": "*uch.AnswerTimeUCH", "type": "*variable",
					"filters": ["*notexists:~*uch.AnswerTimeUCH:"], "value": "~*req.{{TIME}}{*timeString}"},
				{"tag": "AnswerTimeUCHUpdated", "path": "*uch.AnswerTimeUCH", "type": "*variable",
					"filters": ["*gt:~*uch.AnswerTimeUCH:~*req.{{TIME}}{*timeString}"], "value": "~*uch.AnswerTimeUCH"},
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
