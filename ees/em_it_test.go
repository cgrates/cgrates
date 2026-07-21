//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ees

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/engine"
)

func TestExporterMetricsIT(t *testing.T) {
	t.Skip("takes too long - better to analyze manually")
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	exportPath := t.TempDir()
	jsonCfg := fmt.Sprintf(`{
"ees": {
"enabled": true,
"cache": {
  "*fileCSV": { "limit": -1, "ttl": "2s", "staticTTL": false }
},
"exporters": [
  {
	"id": "csv_exporter",
	"type": "*fileCSV",
	"exportPath": "%s",
	"synchronous": true,
	"metricsResetSchedule": "@every 5s",
	"fields": [
	  { "tag": "Number", "path": "*exp.Number", "type": "*variable", "value": "~*em.NumberOfEvents" },
	  { "tag": "OriginID", "path": "*exp.OriginID", "type": "*variable", "value": "~*opts.*originID" },
	  { "tag": "ToR", "path": "*exp.ToR", "type": "*constant", "value": "*sms" },
	  { "tag": "Account", "path": "*exp.Account", "type": "*variable", "value": "~*req.Account" },
	  { "tag": "Destination", "path": "*exp.Destination", "type": "*variable", "value": "~*req.Destination" },
	  { "tag": "AnswerTime", "path": "*exp.AnswerTime", "type": "*variable", "value": "*now", },
	  { "tag": "Usage", "path": "*exp.Usage", "type": "*variable", "value": "~*req.Usage" },
	  { "tag": "Cost", "path": "*exp.Cost", "type": "*variable", "value": "~*req.Cost{*round:4}" },

	  { "tag": "NumberOfEvents", "path": "*trl.NumberOfEvents", "type": "*variable", "value": "~*em.NumberOfEvents" },
	  { "tag": "TotalDuration", "path": "*trl.TotalDuration", "type": "*variable", "value": "~*em.TotalDuration" },
	  { "tag": "TotalSMSUsage", "path": "*trl.TotalSMSUsage", "type": "*variable", "value": "~*em.TotalSMSUsage" },
	  { "tag": "TotalCost", "path": "*trl.TotalCost", "type": "*variable", "value": "~*em.TotalCost{*round:4}" }
	]
  }
]
}
}`, exportPath)

	ng := engine.TestEngine{
		ConfigJSON: jsonCfg,
		DBCfg:      engine.InternalDBCfg,
		Encoding:   *utils.Encoding,
		LogBuffer:  &bytes.Buffer{},
	}
	t.Cleanup(func() {
		fmt.Println(ng.LogBuffer)
	})
	client, _ := ng.Run(t)

	for i := range 7 {
		exportCsvEvent(t, client, i)
		time.Sleep(time.Second)
		if i == 3 {
			resetExporterMetrics(t, client, true)
		}
	}
	time.Sleep(2 * time.Second)
	checkExportedFile(t, exportPath)

	for i := range 7 {
		exportCsvEvent(t, client, i)
		time.Sleep(time.Second)
		if i == 3 {
			resetExporterMetrics(t, client, false)
		}
	}
	time.Sleep(2 * time.Second)
	checkExportedFile(t, exportPath)
}

func exportCsvEvent(t *testing.T, client *birpc.Client, i int) {
	t.Helper()
	var reply map[string]map[string]any
	if err := client.Call(context.Background(), utils.EeSv1ProcessEvent,
		&utils.CGREventWithEeIDs{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.Usage:        time.Duration(i) * time.Second,
					utils.Cost:         i,
				},
				APIOpts: map[string]any{
					utils.MetaOriginID:   fmt.Sprintf("id%d", i),
					utils.OptsEEsVerbose: true,
				},
			},
		}, &reply); err != nil {
		t.Fatal(err)
	}
	fmt.Println(utils.ToJSON(reply))
}

func checkExportedFile(t *testing.T, path string) {
	t.Helper()
	files, err := os.ReadDir(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		fmt.Printf("file %s\n==========\n", file.Name())
		b, err := os.ReadFile(filepath.Join(path, file.Name()))
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(string(b))
	}
}

func resetExporterMetrics(t *testing.T, client *birpc.Client, verbose bool) {
	t.Helper()
	args := V1ResetExporterMetricsParams{
		ExporterID: "csv_exporter",
	}
	if verbose {
		args.APIOpts = map[string]any{
			utils.OptsEEsVerbose: true,
		}
	}
	var reply string
	if err := client.Call(context.Background(), utils.EeSv1ResetExporterMetrics,
		args, &reply); err != nil {
		t.Fatal(err)
	}
}
