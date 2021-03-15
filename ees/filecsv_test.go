/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package ees

import (
	"bytes"
	"encoding/csv"
	"io"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestFileCsvID(t *testing.T) {
	fCsv := &FileCSVee{
		id: "3",
	}
	if rcv := fCsv.ID(); !reflect.DeepEqual(rcv, "3") {
		t.Errorf("Expected %+v \n but got %+v", "3", rcv)
	}
}

func TestFileCsvGetMetrics(t *testing.T) {
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	fCsv := &FileCSVee{
		dc: dc,
	}

	if rcv := fCsv.GetMetrics(); !reflect.DeepEqual(rcv, fCsv.dc) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(rcv), utils.ToJSON(fCsv.dc))
	}
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

func TestFileCsvComposeHeader(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	csvNW := csv.NewWriter(byteBuff)
	fCsv := &FileCSVee{
		id:        "string",
		cgrCfg:    cgrCfg,
		cfgIdx:    0,
		filterS:   filterS,
		file:      nopCloser{byteBuff},
		csvWriter: csvNW,
		dc:        utils.MapStorage{},
	}
	cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Fields = []*config.FCTemplate{
		{
			Path: "*hdr.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
		},
		{
			Path: "*hdr.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field2", utils.InfieldSep),
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Fields {
		field.ComputePath()
	}
	if err := fCsv.composeHeader(); err != nil {
		t.Error(err)
	}
	cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].ComputeFields()
	if err := fCsv.composeHeader(); err != nil {
		t.Error(err)
	}
	csvNW.Flush()
	expected := "field1,field2\n"
	if expected != byteBuff.String() {
		t.Errorf("Expected %q but received %q", expected, byteBuff.String())
	}
	cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Fields = []*config.FCTemplate{
		{
			Path: "*hdr.1", Type: utils.MetaVariable,
			Value:   config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
			Filters: []string{"*wrong-type"},
		},
		{
			Path: "*hdr.1", Type: utils.MetaVariable,
			Value:   config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
			Filters: []string{"*wrong-type"},
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Fields {
		field.ComputePath()
	}
	cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].ComputeFields()
	byteBuff.Reset()
	errExpect := "inline parse error for string: <*wrong-type>"
	if err := fCsv.composeHeader(); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
}

func TestFileCsvComposeTrailer(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	csvNW := csv.NewWriter(byteBuff)
	fCsv := &FileCSVee{
		id:        "string",
		cgrCfg:    cgrCfg,
		cfgIdx:    0,
		filterS:   filterS,
		file:      nopCloser{byteBuff},
		csvWriter: csvNW,
		dc:        utils.MapStorage{},
	}
	cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Fields = []*config.FCTemplate{
		{
			Path: "*trl.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
		},
		{
			Path: "*trl.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field2", utils.InfieldSep),
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Fields {
		field.ComputePath()
	}
	if err := fCsv.composeTrailer(); err != nil {
		t.Error(err)
	}
	cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].ComputeFields()
	if err := fCsv.composeTrailer(); err != nil {
		t.Error(err)
	}
	csvNW.Flush()
	expected := "field1,field2\n"
	if expected != byteBuff.String() {
		t.Errorf("Expected %q but received %q", expected, byteBuff.String())
	}
	cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Fields = []*config.FCTemplate{
		{
			Path: "*trl.1", Type: utils.MetaVariable,
			Value:   config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
			Filters: []string{"*wrong-type"},
		},
		{
			Path: "*trl.1", Type: utils.MetaVariable,
			Value:   config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
			Filters: []string{"*wrong-type"},
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Fields {
		field.ComputePath()
	}
	cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].ComputeFields()
	byteBuff.Reset()
	errExpect := "inline parse error for string: <*wrong-type>"
	if err := fCsv.composeTrailer(); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
}

func TestFileCsvExportEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrEv := new(utils.CGREvent)
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	csvNW := csv.NewWriter(byteBuff)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	fCsv := &FileCSVee{
		id:        "string",
		cgrCfg:    cgrCfg,
		cfgIdx:    0,
		filterS:   filterS,
		file:      nopCloser{byteBuff},
		csvWriter: csvNW,
		dc:        dc,
	}
	cgrEv.Event = map[string]interface{}{
		"test1": "value",
	}
	cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Fields = []*config.FCTemplate{
		{
			Path: "*exp.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*req.test1", utils.InfieldSep),
		},
		{
			Path: "*exp.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("3", utils.InfieldSep),
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Fields {
		field.ComputePath()
	}
	if err := fCsv.ExportEvent(cgrEv); err != nil {
		t.Error(err)
	}
	csvNW.Flush()
	expected := "value\n"
	if expected != byteBuff.String() {
		t.Errorf("Expected %q but received %q", expected, byteBuff.String())
	}
	byteBuff.Reset()
	cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].ComputeFields()
	if err := fCsv.ExportEvent(cgrEv); err != nil {
		t.Error(err)
	}
	csvNW.Flush()
	expected = "value,3\n"
	if expected != byteBuff.String() {
		t.Errorf("Expected %q but received %q", expected, byteBuff.String())
	}

	cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Fields = []*config.FCTemplate{
		{
			Path: "*exp.1", Type: utils.MetaVariable,
			Value:   config.NewRSRParsersMustCompile("~*req.field1", utils.InfieldSep),
			Filters: []string{"*wrong-type"},
		},
		{
			Path: "*exp.1", Type: utils.MetaVariable,
			Value:   config.NewRSRParsersMustCompile("~*req.field1", utils.InfieldSep),
			Filters: []string{"*wrong-type"},
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Fields {
		field.ComputePath()
	}
	cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].ComputeFields()
	byteBuff.Reset()
	errExpect := "inline parse error for string: <*wrong-type>"
	if err := fCsv.ExportEvent(cgrEv); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
}

func TestFileCsvOnEvictedTrailer(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	csvNW := csv.NewWriter(byteBuff)
	fCsv := &FileCSVee{
		id:        "string",
		cgrCfg:    cgrCfg,
		cfgIdx:    0,
		filterS:   filterS,
		file:      nopCloserWrite{byteBuff},
		csvWriter: csvNW,
		dc:        utils.MapStorage{},
	}
	cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Fields = []*config.FCTemplate{
		{
			Path: "*trl.1", Type: utils.MetaVariable,
			Value:   config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
			Filters: []string{"*wrong-type"},
		},
		{
			Path: "*trl.2", Type: utils.MetaVariable,
			Value:   config.NewRSRParsersMustCompile("field2", utils.InfieldSep),
			Filters: []string{"*wrong-type"},
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Fields {
		field.ComputePath()
	}
	cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].ComputeFields()
	fCsv.OnEvicted("test", "test")
}

func TestFileCsvOnEvictedClose(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	csvNW := csv.NewWriter(byteBuff)
	fCsv := &FileCSVee{
		id:        "string",
		cgrCfg:    cgrCfg,
		cfgIdx:    0,
		filterS:   filterS,
		file:      nopCloserError{byteBuff},
		csvWriter: csvNW,
		dc:        utils.MapStorage{},
	}
	cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Fields = []*config.FCTemplate{
		{
			Path: "*trl.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
		},
		{
			Path: "*trl.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field2", utils.InfieldSep),
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].Fields {
		field.ComputePath()
	}
	cgrCfg.EEsCfg().Exporters[fCsv.cfgIdx].ComputeFields()
	fCsv.OnEvicted("test", "test")
}
