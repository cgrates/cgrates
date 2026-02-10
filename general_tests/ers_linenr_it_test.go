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
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestERsLineNr(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
	csvFd, fwvFd, xmlFd, procFd := t.TempDir(), t.TempDir(), t.TempDir(), t.TempDir()
	content := fmt.Sprintf(`{
"general": {
	"log_level": 7
},
"data_db": {
	"db_type": "*internal"
},
"stor_db": {
	"db_type": "*internal"
},
"ers": {
	"enabled": true,
	"sessions_conns": [],
	"readers": [
		{
			"id": "file_csv_reader",
			"run_delay":  "-1",
			"type": "*file_csv",
			"source_path": "%s",
			"flags": ["*dryrun"],
			"processed_path": "%s",
			"fields":[
				{"tag": "FileName", "path": "*cgreq.FileName", "type": "*variable", "value": "~*vars.*fileName"},
				{"tag": "LineNumber", "path": "*cgreq.LineNumber", "type": "*variable", "value": "~*vars.*fileLineNumber"},
				{"tag": "Field", "path": "*cgreq.Field", "type": "*variable", "value": "~*req.0"}
			]
		},
		{
			"id": "file_fwv_reader",
			"run_delay":  "-1",
			"type": "*file_fwv",
			"source_path": "%s",
			"flags": ["*dryrun"],
			"processed_path": "%s",
			"fields":[
				{"tag": "FileName2", "path": "*cgreq.FileName", "type": "*variable", "value": "~*vars.*fileName"},
				{"tag": "LineNumber", "path": "*cgreq.LineNumber", "type": "*variable", "value": "~*vars.*fileLineNumber"},
				{"tag": "FileSeqNr", "path": "*cgreq.FileSeqNr", "type": "*variable", "value": "~*hdr.3-6", "padding":"*zeroleft"},
				{"tag": "Field", "path": "*cgreq.Field", "type": "*variable", "value": "~*req.0-5", "padding":"*right"},
				{"tag": "NrOfElements", "type": "*variable", "path":"*cgreq.NrOfElements", "value": "~*trl.3-4"},
			]
		},
		{
			"id": "file_xml_reader",
			"run_delay":  "-1",
			"type": "*file_xml",
			"source_path": "%s",
			"flags": ["*dryrun"],
			"processed_path": "%s",
			"opts": {
				"xmlRootPath": "root.field"
			},
			"fields":[
				{"tag": "FileName", "path": "*cgreq.FileName", "type": "*variable", "value": "~*vars.*fileName"},
				{"tag": "LineNumber", "path": "*cgreq.LineNumber", "type": "*variable", "value": "~*vars.*fileLineNumber"},
				{"tag": "Field", "path": "*cgreq.Field", "type": "*variable", "value": "~*req.root.field"}
			]
		}
	]
}
}`, csvFd, procFd, fwvFd, procFd, xmlFd, procFd)

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigJSON: content,
		LogBuffer:  buf,
	}
	_, _ = ng.Run(t)

	fileIdx := 0
	createFile := func(t *testing.T, dir, ext, content string) {
		fileIdx++
		filePath := filepath.Join(dir, fmt.Sprintf("file%d%s", fileIdx, ext))
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("could not write to file %s: %v", filePath, err)
		}
	}

	verifyLogLines := func(t *testing.T, reader io.Reader) {
		t.Helper()
		data, err := io.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}
		parts := strings.Split(string(data), "CGREvent: ")
		var records int
		for _, part := range parts[1:] {
			var cgrEv utils.CGREvent
			if err := json.NewDecoder(strings.NewReader(part)).Decode(&cgrEv); err != nil {
				t.Fatalf("failed to decode CGREvent: %v", err)
			}
			field, _ := cgrEv.Event["Field"].(string)
			if !strings.HasPrefix(field, "test") {
				continue
			}
			records++
			lineNumber, _ := cgrEv.Event["LineNumber"].(string)
			if want := strings.TrimPrefix(field, "test"); lineNumber != want {
				t.Errorf("got LineNumber=%s, want %s (Field=%s)", lineNumber, want, field)
			}
		}
		if records != 18 {
			t.Errorf("expected ERs to process 18 records, but it processed %d records", records)
		}
	}

	// Create the files inside the source directories of the readers.
	createFile(t, csvFd, utils.CSVSuffix, "test1\ntest2\ntest3\ntest4\ntest5\ntest6")
	createFile(t, fwvFd, utils.FWVSuffix, `HDR002
test1
test2
test3
test4
test5
test6
TRL6
`)
	createFile(t, xmlFd, utils.XMLSuffix, `<?xml version="1.0" encoding="ISO-8859-1"?>
<root>
  <field>test1</field>
  <field>test2</field>
  <field>test3</field>
  <field>test4</field>
  <field>test5</field>
  <field>test6</field>
</root>`)

	time.Sleep(20 * time.Millisecond) // wait for the files to be processed

	// Check that the suffixes of the 'test' fields match the LineNumber field.
	logData := strings.NewReader(buf.String())
	verifyLogLines(t, logData)
}
