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

package engine

import (
	"archive/zip"
	"bytes"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestTpExporterNewTPExporter(t *testing.T) {
	str := "test"
	bl := false
	sSCfg, _ := config.NewDefaultCGRConfig()
	mpStr := NewInternalDB(nil, nil, true, sSCfg.DataDbCfg().Items)
	rcv, err := NewTPExporter(mpStr, str, str, str, str, false)

	if err != nil {
		if err.Error() != "Unsupported file format" {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}

	if rcv != nil {
		t.Error(rcv)
	}

	rcv, err = NewTPExporter(mpStr, "", str, utils.CSV, str, false)

	if err != nil {
		if err.Error() != "Missing TPid" {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}

	if rcv != nil {
		t.Error(rcv)
	}

	rcv, err = NewTPExporter(mpStr, str, str, utils.CSV, str, bl)
	if err != nil {
		t.Error(err)
	}

	exp := &TPExporter{
		storDb:     mpStr,
		tpID:       str,
		exportPath: str,
		fileFormat: utils.CSV,
		sep:        't',
		compress:   bl,
		cacheBuff:  new(bytes.Buffer),
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected %v\nreceived %v\n", exp, rcv)
	}

	rcv, err = NewTPExporter(mpStr, str, str, utils.CSV, "", bl)

	if err != nil {
		if err.Error() != "Invalid field separator: " {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}

	if rcv != nil {
		t.Error(rcv)
	}

	rcv, err = NewTPExporter(mpStr, str, str, utils.CSV, str, true)

	if err != nil {
		if err.Error() != "open test/tpexport.zip: no such file or directory" {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}

	if rcv != nil {
		t.Error(rcv)
	}

	rcv, err = NewTPExporter(mpStr, str, "", utils.CSV, str, true)
	if err != nil {
		t.Error(err)
	}

	exp2 := &TPExporter{
		storDb:     mpStr,
		tpID:       str,
		exportPath: "",
		fileFormat: utils.CSV,
		sep:        't',
		compress:   true,
		cacheBuff:  exp.cacheBuff,
		zipWritter: zip.NewWriter(exp.cacheBuff),
	}

	if !reflect.DeepEqual(rcv, exp2) {
		t.Errorf("\nexpected %v\nreceived %v\n", exp2, rcv)
	}
}

func TestTpExporterExportStats(t *testing.T) {
	str := "test"
	bl := false
	sSCfg, _ := config.NewDefaultCGRConfig()
	mpStr := NewInternalDB(nil, nil, true, sSCfg.DataDbCfg().Items)
	self, err := NewTPExporter(mpStr, str, str, utils.CSV, str, bl)
	if err != nil {
		t.Error(err)
	}

	rcv := self.ExportStats()
	exp := &utils.ExportedTPStats{
		ExportPath: str,
		Compressed: bl,
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected %v\nreceived %v\n", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestTpExporterGetCacheBuffer(t *testing.T) {
	str := "test"
	bl := false
	sSCfg, _ := config.NewDefaultCGRConfig()
	mpStr := NewInternalDB(nil, nil, true, sSCfg.DataDbCfg().Items)
	self, err := NewTPExporter(mpStr, str, str, utils.CSV, str, bl)
	if err != nil {
		t.Error(err)
	}

	rcv := self.GetCacheBuffer()
	exp := new(bytes.Buffer)

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected %v\nreceived %v\n", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}
