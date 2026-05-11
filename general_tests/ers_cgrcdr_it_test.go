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
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

const ersDryRunCgrCDR = "<ERs> DRY_RUN, reader: <cgrcdr>"

func expectedCDREvent() map[string]any {
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
func TestERSCgrCDRFilters(t *testing.T) {
	db := openTestDB(t, cdr1, cdr2, cdr3)

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigJSON: `{
"ers": {
"enabled": true,
  "readers": [
    {
      "id": "cgrcdr",
		"run_delay": "1m",
	   "type": "*cgrcdr",
	   "source_path": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
		"start_delay": "100ms",
		"flags": ["*dryRun"],
		"tenant": "cgrates.org",
	  	"opts": {
				"sqlDBName":"cgrates2",
				"sqlTableName":"cdrs",
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

	waitForERsLog(t, buf, ersDryRunCgrCDR, 2*time.Second)
	if got := strings.Count(buf.String(), ersDryRunCgrCDR); got != 1 {
		t.Fatalf("expected 1 DRY_RUN record, got %d", got)
	}
	ev := parseCGREvent(t, buf)
	if got, want := utils.ToJSON(ev.Event), utils.ToJSON(expectedCDREvent()); got != want {
		t.Errorf("got event\n%s\nwant\n%s", got, want)
	}
	if got := countRows(t, db, utils.CDRsTBL); got != 3 {
		t.Fatalf("expected 3 rows, got %d", got)
	}

}

func TestERSCgrCDRFiltersDelete(t *testing.T) {
	db := openTestDB(t, cdr1, cdr2, cdr3)

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{

		ConfigJSON: `{
"ers": {
"enabled":true,
  "readers": [
	   {
      "id": "cgrcdr",
	  "run_delay": "1m",
	  "type": "*cgrcdr",
	  "source_path": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
	  "processed_path": "*delete",
	  "start_delay": "250ms",
	  "flags": ["*dryRun"],
	  "tenant": "cgrates.org",
	  "opts": {
				"sqlDBName":"cgrates2",
				"sqlTableName":"cdrs",
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
		func() bool { return countRows(t, db, utils.CDRsTBL) == 2 },
		"expected 2 rows in cdrs after delete",
		2*time.Second,
	)
	if got := strings.Count(buf.String(), ersDryRunCgrCDR); got != 1 {
		t.Fatalf("expected 1 DRY_RUN record, got %d", got)
	}
	ev := parseCGREvent(t, buf)
	if got, want := utils.ToJSON(ev.Event), utils.ToJSON(expectedCDREvent()); got != want {
		t.Errorf("got event\n%s\nwant\n%s", got, want)
	}
}
