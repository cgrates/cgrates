// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ees

import (
	"bytes"
	"encoding/csv"
	"io"
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"
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
	locker := engine.NewGuardianLocker(cfg)
	newIDb, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: newIDb}, cfg.DbCfg())
	newDM := engine.NewDataManager(dbCM, cfg, nil, locker)
	newDM.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
	filterS := engine.NewFilterS(cfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	csvNW := csv.NewWriter(byteBuff)
	fCsv := &FileCSVee{
		cfg:       cfg.EEsCfg().Exporters[0],
		cgrCfg:    cfg,
		filterS:   filterS,
		wrtr:      nopCloser{byteBuff},
		csvWriter: csvNW,
		em:        &utils.ExporterMetrics{},
	}
	fCsv.Cfg().Fields = []*config.FCTemplate{
		{
			Path: "*hdr.1", Type: utils.MetaVariable,
			Value: utils.NewRSRParsersMustCompile("field1", utils.InfieldSep),
		},
		{
			Path: "*hdr.2", Type: utils.MetaVariable,
			Value: utils.NewRSRParsersMustCompile("field2", utils.InfieldSep),
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
			Value:   utils.NewRSRParsersMustCompile("field1", utils.InfieldSep),
			Filters: []string{"*wrong-type"},
		},
		{
			Path: "*hdr.1", Type: utils.MetaVariable,
			Value:   utils.NewRSRParsersMustCompile("field1", utils.InfieldSep),
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
	locker := engine.NewGuardianLocker(cfg)
	newIDb, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: newIDb}, cfg.DbCfg())
	newDM := engine.NewDataManager(dbCM, cfg, nil, locker)
	newDM.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
	filterS := engine.NewFilterS(cfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	csvNW := csv.NewWriter(byteBuff)
	fCsv := &FileCSVee{
		cfg:       cfg.EEsCfg().Exporters[0],
		cgrCfg:    cfg,
		filterS:   filterS,
		wrtr:      nopCloser{byteBuff},
		csvWriter: csvNW,
		em:        &utils.ExporterMetrics{},
	}
	fCsv.Cfg().Fields = []*config.FCTemplate{
		{
			Path: "*trl.1", Type: utils.MetaVariable,
			Value: utils.NewRSRParsersMustCompile("field1", utils.InfieldSep),
		},
		{
			Path: "*trl.2", Type: utils.MetaVariable,
			Value: utils.NewRSRParsersMustCompile("field2", utils.InfieldSep),
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
			Value:   utils.NewRSRParsersMustCompile("field1", utils.InfieldSep),
			Filters: []string{"*wrong-type"},
		},
		{
			Path: "*trl.1", Type: utils.MetaVariable,
			Value:   utils.NewRSRParsersMustCompile("field1", utils.InfieldSep),
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
	locker := engine.NewGuardianLocker(cfg)
	newIDb, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: newIDb}, cfg.DbCfg())
	newDM := engine.NewDataManager(dbCM, cfg, nil, locker)
	newDM.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
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
		wrtr:      nopCloser{byteBuff},
		csvWriter: csvNW,
		em:        em,
	}

	if err := fCsv.ExportEvent(context.Background(), []string{"value", "3"}, ""); err != nil {
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
	locker := engine.NewGuardianLocker(cfg)
	newIDb, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: newIDb}, cfg.DbCfg())
	newDM := engine.NewDataManager(dbCM, cfg, nil, locker)
	newDM.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
	filterS := engine.NewFilterS(cfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	csvNW := csv.NewWriter(byteBuff)
	fCsv := &FileCSVee{
		cfg:       cfg.EEsCfg().Exporters[0],
		cgrCfg:    cfg,
		filterS:   filterS,
		wrtr:      nopCloserWrite{byteBuff},
		csvWriter: csvNW,
		em:        &utils.ExporterMetrics{},
	}
	fCsv.Cfg().Fields = []*config.FCTemplate{
		{
			Path: "*trl.1", Type: utils.MetaVariable,
			Value:   utils.NewRSRParsersMustCompile("field1", utils.InfieldSep),
			Filters: []string{"*wrong-type"},
		},
		{
			Path: "*trl.2", Type: utils.MetaVariable,
			Value:   utils.NewRSRParsersMustCompile("field2", utils.InfieldSep),
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
	locker := engine.NewGuardianLocker(cfg)
	newIDb, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: newIDb}, cfg.DbCfg())
	newDM := engine.NewDataManager(dbCM, cfg, nil, locker)
	newDM.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
	filterS := engine.NewFilterS(cfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	csvNW := csv.NewWriter(byteBuff)
	fCsv := &FileCSVee{
		cfg:       cfg.EEsCfg().Exporters[0],
		cgrCfg:    cfg,
		filterS:   filterS,
		wrtr:      nopCloserError{byteBuff},
		csvWriter: csvNW,
		em:        &utils.ExporterMetrics{},
	}
	fCsv.Cfg().Fields = []*config.FCTemplate{
		{
			Path: "*trl.1", Type: utils.MetaVariable,
			Value: utils.NewRSRParsersMustCompile("field1", utils.InfieldSep),
		},
		{
			Path: "*trl.2", Type: utils.MetaVariable,
			Value: utils.NewRSRParsersMustCompile("field2", utils.InfieldSep),
		},
	}
	for _, field := range fCsv.Cfg().Fields {
		field.ComputePath()
	}
	fCsv.Cfg().ComputeFields()
	fCsv.Close()
}
