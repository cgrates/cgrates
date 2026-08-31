//go:build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/ers"
	"github.com/cgrates/cgrates/utils"
)

type testBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (b *testBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.Write(p)
}

func (b *testBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.String()
}

func TestERSRunReaderFilters(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	sourceDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDir, "events.csv"), []byte(strings.Join([]string{
		"1001,voice,event1",
		"1001,sms,event2",
		"1002,voice,event3",
	}, "\n")), 0644); err != nil {
		t.Fatal(err)
	}
	var buf testBuffer
	testEngine := engine.TestEngine{
		ConfigJSON: fmt.Sprintf(`{
"ers": {
	"enabled": true,
	"readers": [
		{
			"id": "manual_csv",
			"runDelay": "0",
			"type": "*fileCSV",
			"sourcePath": %q,
			"processedPath": "",
			"flags": ["*dryRun"],
			"filters": ["*string:~*req.0:1001"],
			"fields": [
				{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", "value": "~*req.2"}
			]
		}
	]
}
}`, sourceDir),
		DBCfg:            engine.InternalDBCfg,
		Encoding:         *utils.Encoding,
		LogBuffer:        &buf,
		GracefulShutdown: true,
	}
	client, _ := testEngine.Run(t)

	var reply string
	if err := client.Call(context.Background(), utils.ErSv1RunReader,
		&ers.V1RunReaderParams{
			ReaderID: "manual_csv",
			Filters:  []string{"*string:~*req.1:voice"},
		}, &reply); err != nil {
		t.Fatal(err)
	}
	const prefix = "<ERs> DRY_RUN, reader: <manual_csv>"
	waitFor(t, func() bool { return strings.Contains(buf.String(), prefix) },
		"filtered run did not process an event", 2*time.Second)
	if got := strings.Count(buf.String(), prefix); got != 1 {
		t.Fatalf("unexpected number of events after filtered run: %d", got)
	}
	if output := buf.String(); !strings.Contains(output, `"OriginID": "event1"`) ||
		strings.Contains(output, `"OriginID": "event2"`) ||
		strings.Contains(output, `"OriginID": "event3"`) {
		t.Fatalf("unexpected filtered events:\n%s", output)
	}

	if err := client.Call(context.Background(), utils.ErSv1RunReader,
		&ers.V1RunReaderParams{ReaderID: "manual_csv"}, &reply); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool { return strings.Count(buf.String(), prefix) == 3 },
		"second run did not process two events", 2*time.Second)
	if output := buf.String(); !strings.Contains(output, `"OriginID": "event2"`) ||
		strings.Contains(output, `"OriginID": "event3"`) {
		t.Fatalf("request filters changed the next run:\n%s", output)
	}
}

func TestERSRunReaderCgrCDR(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	urID := "record's id"
	cdr1 := &utils.CDR{
		Tenant: "cgrates.org",
		Opts:   map[string]any{utils.MetaURID: utils.Sha1("record1")},
		Event:  map[string]any{utils.OriginID: "event1"},
	}
	cdr2 := &utils.CDR{
		Tenant: "cgrates.org",
		Opts:   map[string]any{utils.MetaURID: urID},
		Event:  map[string]any{utils.OriginID: "event2"},
	}
	db := openTestDB(t, "cgrates2", utils.CDRsTBL, cdr1, cdr2)
	var buf testBuffer
	testEngine := engine.TestEngine{
		ConfigJSON: `{
"ers": {
	"enabled": true,
	"readers": [
		{
			"id": "cgrcdr",
			"runDelay": "0",
			"type": "*cgrcdr",
			"sourcePath": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
			"processedPath": "",
			"flags": ["*dryRun"],
			"opts": {
				"sqlDBName": "cgrates2",
				"sqlTableName": "cdrs",
				"sqlBatchSize": 1
			}
		}
	]
}
}`,
		DBCfg:            engine.InternalDBCfg,
		Encoding:         *utils.Encoding,
		LogBuffer:        &buf,
		GracefulShutdown: true,
	}
	client, _ := testEngine.Run(t)

	var reply string
	if err := client.Call(context.Background(), utils.ErSv1RunReader,
		&ers.V1RunReaderParams{
			ReaderID: "cgrcdr",
			Filters:  []string{"*string:~*req.opts.*urID:" + urID},
		}, &reply); err != nil {
		t.Fatal(err)
	}
	const prefix = "<ERs> DRY_RUN, reader: <cgrcdr>"
	waitFor(t, func() bool { return strings.Contains(buf.String(), prefix) },
		"cgrcdr run did not process an event", 2*time.Second)
	if got := strings.Count(buf.String(), prefix); got != 1 {
		t.Fatalf("unexpected number of cgrcdr events: %d", got)
	}
	if output := buf.String(); !strings.Contains(output, `"OriginID": "event2"`) ||
		strings.Contains(output, `"OriginID": "event1"`) {
		t.Fatalf("unexpected cgrcdr events:\n%s", output)
	}
	if got := countRows(t, db, utils.CDRsTBL); got != 2 {
		t.Fatalf("unexpected row count after manual run: %d", got)
	}
}
