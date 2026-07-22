// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	em, err := utils.NewExporterMetrics("", "Local")
	if err != nil {
		t.Fatal(err)
	}
	fCsv := &FileCSVee{em: em}

	if rcv := fCsv.GetMetrics(); !reflect.DeepEqual(rcv, fCsv.em) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(rcv), utils.ToJSON(fCsv.em))
	}
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

func TestFileCsvComposeHeader(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	newIDb, err := engine.NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
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
		em:        &utils.ExporterMetrics{},
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
	newIDb, err := engine.NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
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
		em:        &utils.ExporterMetrics{},
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
	newIDb, err := engine.NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	newDM := engine.NewDataManager(newIDb, cfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	csvNW := csv.NewWriter(byteBuff)
	em, err := utils.NewExporterMetrics("", "Local")
	if err != nil {
		t.Fatal(err)
	}
	fCsv := &FileCSVee{
		cfg:       cfg.EEsCfg().Exporters[0],
		cgrCfg:    cfg,
		filterS:   filterS,
		file:      nopCloser{byteBuff},
		csvWriter: csvNW,
		em:        em,
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
	newIDb, err := engine.NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
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
		em:        &utils.ExporterMetrics{},
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
	newIDb, err := engine.NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
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
		em:        &utils.ExporterMetrics{},
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
