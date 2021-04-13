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

package ers

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
)

func TestNewFlatstoreER(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	expected := &FlatstoreER{
		cgrCfg: cfg,
	}
	cfg.ERsCfg().Readers[0].SourcePath = "/"
	result, err := NewFlatstoreER(cfg, 0, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	result.(*FlatstoreER).cache = nil
	result.(*FlatstoreER).conReqs = nil
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestFlatstoreConfig(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:               "file_reader1",
			Type:             utils.MetaFileCSV,
			RowLength:        5,
			FieldSep:         ",",
			HeaderDefineChar: ":",
			RunDelay:         -1,
			ConcurrentReqs:   1024,
			SourcePath:       "/tmp/ers/in",
			ProcessedPath:    "/tmp/ers/out",
			XMLRootPath:      utils.HierarchyPath{utils.EmptyString},
			Tenant:           nil,
			Timezone:         utils.EmptyString,
			Filters:          []string{},
			Flags:            utils.FlagsWithParams{},
			Opts:             make(map[string]interface{}),
		},
		{
			ID:               "file_reader2",
			Type:             utils.MetaFileCSV,
			RowLength:        5,
			FieldSep:         ",",
			HeaderDefineChar: ":",
			RunDelay:         -1,
			ConcurrentReqs:   1024,
			SourcePath:       "/tmp/ers/in",
			ProcessedPath:    "/tmp/ers/out",
			XMLRootPath:      utils.HierarchyPath{utils.EmptyString},
			Tenant:           nil,
			Timezone:         utils.EmptyString,
			Filters:          []string{},
			Flags:            utils.FlagsWithParams{},
			Opts:             make(map[string]interface{}),
		},
	}
	expected := cfg.ERsCfg().Readers[0]
	rdr, err := NewFlatstoreER(cfg, 0, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	result := rdr.Config()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestFlatstoreServeNil(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	result, err := NewFlatstoreER(cfg, 0, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("\nExpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
	expected := &FlatstoreER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     nil,
		cache:     result.(*FlatstoreER).cache,
		rdrDir:    "/var/spool/cgrates/ers/in",
		rdrEvents: nil,
		rdrError:  nil,
		rdrExit:   nil,
		conReqs:   result.(*FlatstoreER).conReqs,
	}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpected: <%+v>, \nreceived: <%+v>", expected, result)
	}
	result.Config().RunDelay = time.Duration(0)
	err = result.Serve()
	if err != nil {
		t.Errorf("\nExpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
}

func TestNewUnpairedRecordErrTimezone(t *testing.T) {
	record := []string{"TEST1", "TEST2", "TEST3", "TEST4", "TEST5", "TEST6", "TEST7"}
	timezone, _ := time.Time.Zone(time.Now())
	fileName := "testfile.csv"
	errExpect := "unknown time zone EEST"
	_, err := NewUnpairedRecord(record, timezone, fileName)
	if err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}

func TestNewUnpairedRecordErr(t *testing.T) {
	record := []string{"invalid"}
	timezone, _ := time.Time.Zone(time.Now())
	fileName := "testfile.csv"
	errExpect := "MISSING_IE"
	_, err := NewUnpairedRecord(record, timezone, fileName)
	if err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}

func TestPairToRecord(t *testing.T) {
	part1 := &UnpairedRecord{
		Method: "INVITE",
		Values: []string{"value1", "value2", "value3", "value4", "value5", "value6"},
	}
	part2 := &UnpairedRecord{
		Method: "BYE",
		Values: []string{"value1", "value2", "value3", "value4", "value5", "value6"},
	}
	rcv, err := pairToRecord(part1, part2)
	rcvExpect := append(part1.Values, "0")
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, rcvExpect) {
		t.Errorf("Expected %v but received %v", rcvExpect, rcv)
	}

	part1.Values = append(part1.Values, "value7", "value8")
	part2.Values = append(part2.Values, "value7", "value8")
	_, err = pairToRecord(part1, part2)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(part1.Values[7], part2.Values[7]) {
		t.Errorf("Last INVITE value does not match last BYE value")
	}

	cfg := config.NewDefaultCGRConfig()
	fltrs := &engine.FilterS{}
	eR := &FlatstoreER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/flatstoreErs/out",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	eR.conReqs <- struct{}{}
	eR.Config().ProcessedPath = "/tmp"
	part1.FileName = "testfile"
	eR.dumpToFile("ID1", nil)
	eR.dumpToFile("ID1", part1)
	part1.Values = []string{"\n"}
	eR.dumpToFile("ID1", part1)
}

func TestPairToRecordReverse(t *testing.T) {
	part1 := &UnpairedRecord{
		Method: "BYE",
		Values: []string{"value1", "value2", "value3", "value4", "value5"},
	}
	part2 := &UnpairedRecord{
		Method: "INVITE",
		Values: []string{"value1", "value2", "value3", "value4", "value5"},
	}
	rcv, err := pairToRecord(part1, part2)
	rcvExpect := append(part1.Values, "0")
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, rcvExpect) {
		t.Errorf("Expected %v but received %v", rcvExpect, rcv)
	}
}

func TestPairToRecordErrors(t *testing.T) {
	part1 := &UnpairedRecord{
		Method: "INVITE",
		Values: []string{"value1", "value2", "value3", "value4", "value5"},
	}
	part2 := &UnpairedRecord{
		Method: "INVITE",
	}
	errExpect := "MISSING_BYE"
	if _, err := pairToRecord(part1, part2); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}

	errExpect = "INCONSISTENT_VALUES_LENGTH"
	part2.Method = "BYE"
	if _, err := pairToRecord(part1, part2); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}

	part1.Method = "BYE"
	errExpect = "MISSING_INVITE"
	if _, err := pairToRecord(part1, part2); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}
