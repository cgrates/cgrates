//go:build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH
This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.
You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package general_tests

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
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
		fieldRegex := regexp.MustCompile(`"Field":"test(\d+)"`)
		lineNumberRegex := regexp.MustCompile(`"LineNumber":"(\d+)"`)
		records := 0
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.Contains(line, `"Field":"test`) {
				continue
			}

			records++

			fieldMatch := fieldRegex.FindStringSubmatch(line)
			lineNumberMatch := lineNumberRegex.FindStringSubmatch(line)

			if len(fieldMatch) != 2 || len(lineNumberMatch) != 2 {
				t.Fatalf("invalid log line format: %s", line)
			}

			testNumber, err := strconv.Atoi(fieldMatch[1])
			if err != nil {
				t.Fatal(err)
			}
			lineNumber, err := strconv.Atoi(lineNumberMatch[1])
			if err != nil {
				t.Fatal(err)
			}
			if testNumber != lineNumber {
				t.Errorf("mismatch in line: %s, field number: %d, line number: %d", line, testNumber, lineNumber)
			}
		}

		if err := scanner.Err(); err != nil {
			t.Errorf("error reading input: %v", err)
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

	time.Sleep(5 * time.Millisecond) // wait for the files to be processed

	// Check that the suffixes of the 'test' fields match the LineNumber field.
	logData := strings.NewReader(buf.String())
	verifyLogLines(t, logData)
}
