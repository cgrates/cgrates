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

func TestFileFwvGetMetrics(t *testing.T) {
	dc, err := newEEMetrics("Local")
	if err != nil {
		t.Error(err)
	}
	fFwv := &FileFWVee{dc: dc}

	if rcv := fFwv.GetMetrics(); !reflect.DeepEqual(rcv, fFwv.dc) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(rcv), utils.ToJSON(fFwv.dc))
	}
}

func TestFileFwvComposeHeader(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	newIDb := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	newDM := engine.NewDataManager(newIDb, cfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	csvNW := csv.NewWriter(byteBuff)
	fFwv := &FileFWVee{
		cfg:     cfg.EEsCfg().Exporters[0],
		cgrCfg:  cfg,
		filterS: filterS,
		file:    nopCloser{byteBuff},
		dc:      &utils.SafeMapStorage{},
	}
	fFwv.Cfg().Fields = []*config.FCTemplate{
		{
			Path: "*hdr.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
		},
		{
			Path: "*hdr.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field2", utils.InfieldSep),
		},
	}
	for _, field := range fFwv.Cfg().Fields {
		field.ComputePath()
	}
	if err := fFwv.composeHeader(); err != nil {
		t.Error(err)
	}
	fFwv.Cfg().ComputeFields()
	if err := fFwv.composeHeader(); err != nil {
		t.Error(err)
	}
	csvNW.Flush()
	expected := "field1field2\n"
	if expected != byteBuff.String() {
		t.Errorf("Expected %q but received %q", expected, byteBuff.String())
	}
	fFwv.Cfg().Fields = []*config.FCTemplate{
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
	for _, field := range fFwv.Cfg().Fields {
		field.ComputePath()
	}
	fFwv.Cfg().ComputeFields()
	byteBuff.Reset()
	errExpect := "inline parse error for string: <*wrong-type>"
	if err := fFwv.composeHeader(); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
}

func TestFileFwvComposeTrailer(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	newIDb := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	newDM := engine.NewDataManager(newIDb, cfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	csvNW := csv.NewWriter(byteBuff)
	fFwv := &FileFWVee{
		cfg:     cfg.EEsCfg().Exporters[0],
		cgrCfg:  cfg,
		filterS: filterS,
		file:    nopCloser{byteBuff},
		dc:      &utils.SafeMapStorage{},
	}
	fFwv.Cfg().Fields = []*config.FCTemplate{
		{
			Path: "*trl.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
		},
		{
			Path: "*trl.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field2", utils.InfieldSep),
		},
	}
	for _, field := range fFwv.Cfg().Fields {
		field.ComputePath()
	}
	if err := fFwv.composeTrailer(); err != nil {
		t.Error(err)
	}
	fFwv.Cfg().ComputeFields()
	if err := fFwv.composeTrailer(); err != nil {
		t.Error(err)
	}
	csvNW.Flush()
	expected := "field1field2\n"
	if expected != byteBuff.String() {
		t.Errorf("Expected %q but received %q", expected, byteBuff.String())
	}
	fFwv.Cfg().Fields = []*config.FCTemplate{
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
	for _, field := range fFwv.Cfg().Fields {
		field.ComputePath()
	}
	fFwv.Cfg().ComputeFields()
	byteBuff.Reset()
	errExpect := "inline parse error for string: <*wrong-type>"
	if err := fFwv.composeTrailer(); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
}

func TestFileFwvExportEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	newIDb := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	newDM := engine.NewDataManager(newIDb, cfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	csvNW := csv.NewWriter(byteBuff)
	dc, err := newEEMetrics("Local")
	if err != nil {
		t.Error(err)
	}
	fFwv := &FileFWVee{
		cfg:     cfg.EEsCfg().Exporters[0],
		cgrCfg:  cfg,
		filterS: filterS,
		file:    nopCloser{byteBuff},
		dc:      dc,
	}
	if err := fFwv.ExportEvent([]string{"value", "3"}, ""); err != nil {
		t.Error(err)
	}
	csvNW.Flush()
	expected := "value3\n"
	if expected != byteBuff.String() {
		t.Errorf("Expected %q but received %q", expected, byteBuff.String())
	}
}

type nopCloserWrite struct {
	io.Writer
}

func (nopCloserWrite) Close() error { return nil }
func (nopCloserWrite) Write(s []byte) (n int, err error) {
	return 0, utils.ErrNotImplemented
}

func TestFileFwvExportEventWriteError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	newIDb := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	newDM := engine.NewDataManager(newIDb, cfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	dc, err := newEEMetrics("Local")
	if err != nil {
		t.Error(err)
	}
	fFwv := &FileFWVee{
		cfg:     cfg.EEsCfg().Exporters[0],
		cgrCfg:  cfg,
		filterS: filterS,
		file:    nopCloserWrite{byteBuff},
		dc:      dc,
	}
	if err := fFwv.ExportEvent([]string{""}, ""); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Expected %q but received %q", utils.ErrNotImplemented, err)
	}
}

func TestFileFwvComposeHeaderWriteError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	newIDb := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	newDM := engine.NewDataManager(newIDb, cfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	fFwv := &FileFWVee{
		cfg:     cfg.EEsCfg().Exporters[0],
		cgrCfg:  cfg,
		filterS: filterS,
		file:    nopCloserWrite{byteBuff},
		dc:      &utils.SafeMapStorage{},
	}
	fFwv.Cfg().Fields = []*config.FCTemplate{
		{
			Path: "*hdr.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
		},
		{
			Path: "*hdr.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field2", utils.InfieldSep),
		},
	}
	for _, field := range fFwv.Cfg().Fields {
		field.ComputePath()
	}
	fFwv.Cfg().ComputeFields()
	if err := fFwv.composeHeader(); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Expected %q but received %q", utils.ErrNotImplemented, err)
	}
}

func TestFileFwvComposeTrailerWriteError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	newIDb := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	newDM := engine.NewDataManager(newIDb, cfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	fFwv := &FileFWVee{
		cfg:     cfg.EEsCfg().Exporters[0],
		cgrCfg:  cfg,
		filterS: filterS,
		file:    nopCloserWrite{byteBuff},
		dc:      &utils.SafeMapStorage{},
	}
	fFwv.Cfg().Fields = []*config.FCTemplate{
		{
			Path: "*trl.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
		},
		{
			Path: "*trl.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field2", utils.InfieldSep),
		},
	}
	for _, field := range fFwv.Cfg().Fields {
		field.ComputePath()
	}
	fFwv.Cfg().ComputeFields()
	if err := fFwv.composeTrailer(); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Expected %q but received %q", utils.ErrNotImplemented, err)
	}
}
func TestFileFwvOnEvictedTrailer(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	newIDb := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	newDM := engine.NewDataManager(newIDb, cfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	fFwv := &FileFWVee{
		cfg:     cfg.EEsCfg().Exporters[0],
		cgrCfg:  cfg,
		filterS: filterS,
		file:    nopCloserWrite{byteBuff},
		dc:      &utils.SafeMapStorage{},
	}
	fFwv.Cfg().Fields = []*config.FCTemplate{
		{
			Path: "*trl.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
		},
		{
			Path: "*trl.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field2", utils.InfieldSep),
		},
	}
	for _, field := range fFwv.Cfg().Fields {
		field.ComputePath()
	}
	fFwv.Cfg().ComputeFields()
	fFwv.Close()
}

type nopCloserError struct {
	io.Writer
}

func (nopCloserError) Close() error { return utils.ErrNotImplemented }
func (nopCloserError) Write(s []byte) (n int, err error) {
	return 0, utils.ErrNotImplemented
}
func TestFileFwvOnEvictedClose(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	newIDb := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	newDM := engine.NewDataManager(newIDb, cfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
	byteBuff := new(bytes.Buffer)
	fFwv := &FileFWVee{
		cfg:     cfg.EEsCfg().Exporters[0],
		cgrCfg:  cfg,
		filterS: filterS,
		file:    nopCloserError{byteBuff},
		dc:      &utils.SafeMapStorage{},
	}
	fFwv.Cfg().Fields = []*config.FCTemplate{
		{
			Path: "*trl.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
		},
		{
			Path: "*trl.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field2", utils.InfieldSep),
		},
	}
	for _, field := range fFwv.Cfg().Fields {
		field.ComputePath()
	}
	fFwv.Cfg().ComputeFields()
	fFwv.Close()
}
