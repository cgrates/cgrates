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

func TestFileCsvGetMetrics(t *testing.T) {
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	fCsv := &FileCSVee{dc: dc}

	if rcv := fCsv.GetMetrics(); !reflect.DeepEqual(rcv, fCsv.dc) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(rcv), utils.ToJSON(fCsv.dc))
	}
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

func TestFileCsvComposeHeader(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	newIDb := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	newDM := engine.NewDataManager(newIDb, cfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	csvNW := csv.NewWriter(byteBuff)
	fCsv := &FileCSVee{
		cfg:       cfg.EEsCfg().Exporters[0],
		cgrCfg:    cfg,
		filterS:   filterS,
		file:      nopCloser{byteBuff},
		csvWriter: csvNW,
		dc:        &utils.SafeMapStorage{},
	}
	fCsv.Cfg().Fields = []*config.FCTemplate{
		{
			Path: "*hdr.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
		},
		{
			Path: "*hdr.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field2", utils.InfieldSep),
		},
	}
	for _, field := range fCsv.Cfg().Fields {
		field.ComputePath()
	}
	if err := fCsv.composeHeader(); err != nil {
		t.Error(err)
	}
	fCsv.Cfg().ComputeFields()
	if err := fCsv.composeHeader(); err != nil {
		t.Error(err)
	}
	csvNW.Flush()
	expected := "field1,field2\n"
	if expected != byteBuff.String() {
		t.Errorf("Expected %q but received %q", expected, byteBuff.String())
	}
	fCsv.Cfg().Fields = []*config.FCTemplate{
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
	for _, field := range fCsv.Cfg().Fields {
		field.ComputePath()
	}
	fCsv.Cfg().ComputeFields()
	byteBuff.Reset()
	errExpect := "inline parse error for string: <*wrong-type>"
	if err := fCsv.composeHeader(); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
}

func TestFileCsvComposeTrailer(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	newIDb := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	newDM := engine.NewDataManager(newIDb, cfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	csvNW := csv.NewWriter(byteBuff)
	fCsv := &FileCSVee{
		cfg:       cfg.EEsCfg().Exporters[0],
		cgrCfg:    cfg,
		filterS:   filterS,
		file:      nopCloser{byteBuff},
		csvWriter: csvNW,
		dc:        &utils.SafeMapStorage{},
	}
	fCsv.Cfg().Fields = []*config.FCTemplate{
		{
			Path: "*trl.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
		},
		{
			Path: "*trl.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field2", utils.InfieldSep),
		},
	}
	for _, field := range fCsv.Cfg().Fields {
		field.ComputePath()
	}
	if err := fCsv.composeTrailer(); err != nil {
		t.Error(err)
	}
	fCsv.Cfg().ComputeFields()
	if err := fCsv.composeTrailer(); err != nil {
		t.Error(err)
	}
	csvNW.Flush()
	expected := "field1,field2\n"
	if expected != byteBuff.String() {
		t.Errorf("Expected %q but received %q", expected, byteBuff.String())
	}
	fCsv.Cfg().Fields = []*config.FCTemplate{
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
	for _, field := range fCsv.Cfg().Fields {
		field.ComputePath()
	}
	fCsv.Cfg().ComputeFields()
	byteBuff.Reset()
	errExpect := "inline parse error for string: <*wrong-type>"
	if err := fCsv.composeTrailer(); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
}

func TestFileCsvExportEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	newIDb := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	newDM := engine.NewDataManager(newIDb, cfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
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
		cfg:       cfg.EEsCfg().Exporters[0],
		cgrCfg:    cfg,
		filterS:   filterS,
		file:      nopCloser{byteBuff},
		csvWriter: csvNW,
		dc:        dc,
	}

	if err := fCsv.ExportEvent([]string{"value", "3"}, ""); err != nil {
		t.Error(err)
	}
	csvNW.Flush()
	expected := "value,3\n"
	if expected != byteBuff.String() {
		t.Errorf("Expected %q but received %q", expected, byteBuff.String())
	}
}

func TestFileCsvOnEvictedTrailer(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	newIDb := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	newDM := engine.NewDataManager(newIDb, cfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	csvNW := csv.NewWriter(byteBuff)
	fCsv := &FileCSVee{
		cfg:       cfg.EEsCfg().Exporters[0],
		cgrCfg:    cfg,
		filterS:   filterS,
		file:      nopCloserWrite{byteBuff},
		csvWriter: csvNW,
		dc:        &utils.SafeMapStorage{},
	}
	fCsv.Cfg().Fields = []*config.FCTemplate{
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
	for _, field := range fCsv.Cfg().Fields {
		field.ComputePath()
	}
	fCsv.Cfg().ComputeFields()
	fCsv.Close()
}

func TestFileCsvOnEvictedClose(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	newIDb := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	newDM := engine.NewDataManager(newIDb, cfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	csvNW := csv.NewWriter(byteBuff)
	fCsv := &FileCSVee{
		cfg:       cfg.EEsCfg().Exporters[0],
		cgrCfg:    cfg,
		filterS:   filterS,
		file:      nopCloserError{byteBuff},
		csvWriter: csvNW,
		dc:        &utils.SafeMapStorage{},
	}
	fCsv.Cfg().Fields = []*config.FCTemplate{
		{
			Path: "*trl.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
		},
		{
			Path: "*trl.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field2", utils.InfieldSep),
		},
	}
	for _, field := range fCsv.Cfg().Fields {
		field.ComputePath()
	}
	fCsv.Cfg().ComputeFields()
	fCsv.Close()
}
