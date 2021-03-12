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
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestHttpJsonMapID(t *testing.T) {
	httpEE := &HTTPjsonMapEE{
		id: "3",
	}
	if rcv := httpEE.ID(); !reflect.DeepEqual(rcv, "3") {
		t.Errorf("Expected %+v but got %+v", "3", rcv)
	}
}

func TestHttpJsonMapGetMetrics(t *testing.T) {
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	httpEE := &HTTPjsonMapEE{
		dc: dc,
	}

	if rcv := httpEE.GetMetrics(); !reflect.DeepEqual(rcv, httpEE.dc) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(rcv), utils.ToJSON(httpEE.dc))
	}
}

func TestHttpJsonMapExportEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaSQSjsonMap
	cgrEv := new(utils.CGREvent)
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}

	httpEE, err := NewHTTPjsonMapEE(cgrCfg, 0, filterS, dc)
	if err != nil {
		t.Error(err)
	}
	cgrEv.Event = map[string]interface{}{
		"test": "string",
	}
	cgrCfg.EEsCfg().Exporters[0].Fields = []*config.FCTemplate{
		{
			Path: "*exp.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*req.field1", utils.InfieldSep),
		},
		{
			Path: "*exp.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("*req.field2", utils.InfieldSep),
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[0].Fields {
		field.ComputePath()
	}
	errExpect := `Post "/var/spool/cgrates/ees": unsupported protocol scheme ""`
	if err := httpEE.ExportEvent(cgrEv); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
	cgrCfg.EEsCfg().Exporters[0].ComputeFields()
	if err := httpEE.ExportEvent(cgrEv); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
	cgrCfg.EEsCfg().Exporters[0].Fields = []*config.FCTemplate{
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
	for _, field := range cgrCfg.EEsCfg().Exporters[0].Fields {
		field.ComputePath()
	}
	cgrCfg.EEsCfg().Exporters[0].ComputeFields()
	errExpect = "inline parse error for string: <*wrong-type>"
	if err := httpEE.ExportEvent(cgrEv); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
	cgrCfg.EEsCfg().Exporters[0].Fields = []*config.FCTemplate{
		{
			Path: "*exp.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*req.field1", utils.InfieldSep),
		},
		{
			Path: "*exp.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*req.field2", utils.InfieldSep),
		},
		{
			Path: "*hdr.1", Type: utils.MetaVariable,
			Value:   config.NewRSRParsersMustCompile("~*req.field2", utils.InfieldSep),
			Filters: []string{"*wrong-type"},
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[0].Fields {
		field.ComputePath()
	}
	cgrCfg.EEsCfg().Exporters[0].ComputeFields()
	errExpect = "inline parse error for string: <*wrong-type>"
	if err := httpEE.ExportEvent(cgrEv); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
	httpEE.OnEvicted("test", "test")
}

func TestHttpJsonMapComposeHeader(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Type = utils.MetaHTTPjson
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	httpEE, err := NewHTTPjsonMapEE(cgrCfg, 0, filterS, dc)
	if err != nil {
		t.Error(err)
	}
	cgrCfg.EEsCfg().Exporters[0].Fields = []*config.FCTemplate{
		{
			Path: "*hdr.1", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field1", utils.InfieldSep),
		},
		{
			Path: "*hdr.2", Type: utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("field2", utils.InfieldSep),
		},
	}
	for _, field := range cgrCfg.EEsCfg().Exporters[0].Fields {
		field.ComputePath()
	}
	if _, err := httpEE.composeHeader(); err != nil {
		t.Error(err)
	}
	cgrCfg.EEsCfg().Exporters[0].ComputeFields()
	if _, err := httpEE.composeHeader(); err != nil {
		t.Error(err)
	}
	cgrCfg.EEsCfg().Exporters[0].Fields = []*config.FCTemplate{
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
	for _, field := range cgrCfg.EEsCfg().Exporters[0].Fields {
		field.ComputePath()
	}
	cgrCfg.EEsCfg().Exporters[0].ComputeFields()
	errExpect := "inline parse error for string: <*wrong-type>"
	if _, err := httpEE.composeHeader(); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
}
