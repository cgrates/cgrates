//go:build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

const ersDryRunCgrUR = "<ERs> DRY_RUN, reader: <cgrur>"

func expectedUREvent() map[string]any {
	ts := timeStart.Format("2006-01-02T15:04:05Z07:00")
	return map[string]any{
		"Account":     "1001",
		"AnswerTime":  ts,
		"Category":    "call",
		"Destination": "1002",
		"ExtraFields": map[string]any{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		"ExtraInfo":   "extraInfo",
		"OrderID":     123,
		"OriginHost":  "192.168.1.1",
		"OriginID":    "oid2",
		"RequestType": "*rated",
		"SetupTime":   ts,
		"Source":      "test",
		"Subject":     "1001",
		"ToR":         "*voice",
		"Usage":       10000000000}

}
func TestERSCgrURFilters(t *testing.T) {
	db := openTestDB(t, "cgrates2", utils.URsTBL, ur1, ur2, ur3)
	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigJSON: `{
"ers": {
"enabled": true,
  "readers": [
    {
      "id": "cgrur",
		"runDelay": "1m",
	   "type": "*cgrur",
	   "sourcePath": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
		"startDelay": "100ms",
		"flags": ["*dryRun"],
		"tenant": "cgrates.org",
	  	"opts": {
				"sqlDBName":"cgrates2",
				"sqlTableName":"urs",
				"sqlBatchSize": 3
		},
		"filters": [
					"*gt:~*req.event.AnswerTime:-168h", 
			],
    }
  ]
}
}`,
		DBCfg:            getDBCfg(t),
		Encoding:         *utils.Encoding,
		LogBuffer:        buf,
		GracefulShutdown: true,
	}
	ng.Run(t)

	waitForLog(t, buf, ersDryRunCgrUR, 2*time.Second)
	if got := strings.Count(buf.String(), ersDryRunCgrUR); got != 1 {
		t.Fatalf("expected 1 DRY_RUN record, got %d", got)
	}
	ev := parseCGREvent(t, buf)
	if got, want := utils.ToJSON(ev.Event), utils.ToJSON(expectedUREvent()); got != want {
		t.Errorf("got event\n%s\nwant\n%s", got, want)
	}
	if got := countRows(t, db, utils.URsTBL); got != 3 {
		t.Fatalf("expected 3 rows, got %d", got)
	}

}

func TestERSCgrURFiltersDelete(t *testing.T) {
	db := openTestDB(t, "cgrates2", utils.URsTBL, ur1, ur2, ur3)

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{

		ConfigJSON: `{
"ers": {
"enabled":true,
  "readers": [
	   {
      "id": "cgrur",
	  "runDelay": "1m",
	  "type": "*cgrur",
	  "sourcePath": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
	  "processedPath": "*delete",
	  "startDelay": "250ms",
	  "flags": ["*dryRun"],
	  "tenant": "cgrates.org",
	  "opts": {
				"sqlDBName":"cgrates2",
				"sqlTableName":"urs",
				"sqlBatchSize": 2
		},
	  "filters": [
					"*gt:~*req.event.AnswerTime:-168h", 
			],
    }
  ]
}
}`,
		DBCfg:            getDBCfg(t),
		Encoding:         *utils.Encoding,
		LogBuffer:        buf,
		GracefulShutdown: true,
	}
	ng.Run(t)

	waitFor(t,
		func() bool { return countRows(t, db, utils.URsTBL) == 2 },
		"expected 2 rows in urs after delete",
		2*time.Second,
	)
	if got := strings.Count(buf.String(), ersDryRunCgrUR); got != 1 {
		t.Fatalf("expected 1 DRY_RUN record, got %d", got)
	}
	ev := parseCGREvent(t, buf)
	if got, want := utils.ToJSON(ev.Event), utils.ToJSON(expectedUREvent()); got != want {
		t.Errorf("got event\n%s\nwant\n%s", got, want)
	}
}
