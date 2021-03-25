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
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestID(t *testing.T) {
	ee := &ElasticEe{
		id: "3",
	}
	if rcv := ee.ID(); !reflect.DeepEqual(rcv, "3") {
		t.Errorf("Expected %+v \n but got %+v", "3", rcv)
	}
}

func TestGetMetrics(t *testing.T) {
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	ee := &ElasticEe{
		dc: dc,
	}

	if rcv := ee.GetMetrics(); !reflect.DeepEqual(rcv, ee.dc) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(rcv), utils.ToJSON(ee.dc))
	}
}

func TestInitClient(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	ee := &ElasticEe{
		cgrCfg: cgrCfg,
	}
	ee.cgrCfg.EEsCfg().Exporters[0].ExportPath = "/\x00"
	errExpect := `cannot create client: parse "/\x00": net/url: invalid control character in URL`
	if err := ee.init(); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %+v \n but got %+v", errExpect, err)
	}
}
func TestInitCase1(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts[utils.ElsIndex] = "test"
	ee := &ElasticEe{
		cgrCfg: cgrCfg,
	}
	if err := ee.init(); err != nil {
		t.Error(err)
	}
	eeExpect := "test"
	if !reflect.DeepEqual(ee.opts.Index, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", eeExpect, ee.opts.Index)
	}
}

func TestInitCase2(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts[utils.ElsIfPrimaryTerm] = 20
	ee := &ElasticEe{
		cgrCfg: cgrCfg,
	}
	if err := ee.init(); err != nil {
		t.Error(err)
	}
	eeExpect := utils.IntPointer(20)
	if !reflect.DeepEqual(ee.opts.IfPrimaryTerm, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(ee.opts.IfPrimaryTerm))
	}
}

func TestInitCase2Err(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts[utils.ElsIfPrimaryTerm] = "test"
	ee := &ElasticEe{
		cgrCfg: cgrCfg,
	}
	errExpect := "strconv.ParseInt: parsing \"test\": invalid syntax"
	if err := ee.init(); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %+v \n but got %+v", errExpect, err)
	}
}

func TestInitCase3(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts[utils.ElsIfSeqNo] = 20
	ee := &ElasticEe{
		cgrCfg: cgrCfg,
	}
	if err := ee.init(); err != nil {
		t.Error(err)
	}
	eeExpect := utils.IntPointer(20)
	if !reflect.DeepEqual(ee.opts.IfSeqNo, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(ee.opts.IfSeqNo))
	}
}

func TestInitCase3Err(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts[utils.ElsIfSeqNo] = "test"
	ee := &ElasticEe{
		cgrCfg: cgrCfg,
	}
	errExpect := "strconv.ParseInt: parsing \"test\": invalid syntax"
	if err := ee.init(); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %+v \n but got %+v", errExpect, err)
	}
}

func TestInitCase4(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts[utils.ElsOpType] = "test"
	ee := &ElasticEe{
		cgrCfg: cgrCfg,
	}
	if err := ee.init(); err != nil {
		t.Error(err)
	}
	eeExpect := "test"
	if !reflect.DeepEqual(ee.opts.OpType, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(ee.opts.OpType))
	}
}

func TestInitCase5(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts[utils.ElsPipeline] = "test"
	ee := &ElasticEe{
		cgrCfg: cgrCfg,
	}
	if err := ee.init(); err != nil {
		t.Error(err)
	}
	eeExpect := "test"
	if !reflect.DeepEqual(ee.opts.Pipeline, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(ee.opts.Pipeline))
	}
}

func TestInitCase6(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts[utils.ElsRouting] = "test"
	ee := &ElasticEe{
		cgrCfg: cgrCfg,
	}
	if err := ee.init(); err != nil {
		t.Error(err)
	}
	eeExpect := "test"
	if !reflect.DeepEqual(ee.opts.Routing, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(ee.opts.Routing))
	}
}

func TestInitCase7(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts[utils.ElsTimeout] = "test"
	ee := &ElasticEe{
		cgrCfg: cgrCfg,
	}
	errExpect := "time: invalid duration \"test\""
	if err := ee.init(); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %+v \n but got %+v", errExpect, err)
	}
}

func TestInitCase8(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts[utils.ElsVersionLow] = 20
	ee := &ElasticEe{
		cgrCfg: cgrCfg,
	}
	if err := ee.init(); err != nil {
		t.Error(err)
	}
	eeExpect := utils.IntPointer(20)
	if !reflect.DeepEqual(ee.opts.Version, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(ee.opts.Version))
	}
}

func TestInitCase8Err(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts[utils.ElsVersionLow] = "test"
	ee := &ElasticEe{
		cgrCfg: cgrCfg,
	}
	errExpect := "strconv.ParseInt: parsing \"test\": invalid syntax"
	if err := ee.init(); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %+v \n but got %+v", errExpect, err)
	}
}

func TestInitCase9(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts[utils.ElsVersionType] = "test"
	ee := &ElasticEe{
		cgrCfg: cgrCfg,
	}
	if err := ee.init(); err != nil {
		t.Error(err)
	}
	eeExpect := "test"
	if !reflect.DeepEqual(ee.opts.VersionType, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(ee.opts.VersionType))
	}
}

func TestInitCase10(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts[utils.ElsWaitForActiveShards] = "test"
	ee := &ElasticEe{
		cgrCfg: cgrCfg,
	}
	if err := ee.init(); err != nil {
		t.Error(err)
	}
	eeExpect := "test"
	if !reflect.DeepEqual(ee.opts.WaitForActiveShards, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(ee.opts.WaitForActiveShards))
	}
}

type mockClientErr struct{}

func (mockClientErr) Perform(req *http.Request) (res *http.Response, err error) {
	res = &http.Response{
		StatusCode: 300,
		Body:       io.NopCloser(bytes.NewBuffer([]byte(`{"test":"test"}`))),
		Header:     http.Header{},
	}
	return res, nil
}

func TestElasticExportEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
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
	eEe, err := NewElasticExporter(cgrCfg, 0, filterS, dc)
	if err != nil {
		t.Error(err)
	}
	eEe.eClnt.Transport = new(mockClientErr)
	cgrEv.Event = map[string]interface{}{
		"test1": "value",
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
	cgrCfg.EEsCfg().Exporters[0].ComputeFields()
	if err := eEe.ExportEvent(cgrEv); err != nil {
		t.Error(err)
	}
}

type mockClientErr2 struct{}

func (mockClientErr2) Perform(req *http.Request) (res *http.Response, err error) {
	res = &http.Response{
		StatusCode: 300,
		Body:       io.NopCloser(bytes.NewBuffer([]byte(""))),
		Header:     http.Header{},
	}
	return res, nil
}

func TestElasticExportEvent2(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
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
	eEe, err := NewElasticExporter(cgrCfg, 0, filterS, dc)
	if err != nil {
		t.Error(err)
	}
	eEe.eClnt.Transport = new(mockClientErr2)
	cgrEv.Event = map[string]interface{}{
		"test1": "value",
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
	cgrCfg.EEsCfg().Exporters[0].ComputeFields()
	errExpect := io.EOF
	if err := eEe.ExportEvent(cgrEv); err == nil || err != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}

type mockClient struct{}

func (mockClient) Perform(req *http.Request) (res *http.Response, err error) {
	res = &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBuffer([]byte(""))),
		Header:     http.Header{},
	}
	return res, nil
}
func TestElasticExportEvent3(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
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
	eEe, err := NewElasticExporter(cgrCfg, 0, filterS, dc)
	if err != nil {
		t.Error(err)
	}
	eEe.eClnt.Transport = new(mockClient)
	cgrEv.Event = map[string]interface{}{
		"test1": "value",
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
	// errExpect := `unsupported protocol scheme ""`
	cgrCfg.EEsCfg().Exporters[0].ComputeFields()
	if err := eEe.ExportEvent(cgrEv); err != nil {
		t.Error(err)
	}
}

func TestElasticExportEvent4(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
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
	eEe, err := NewElasticExporter(cgrCfg, 0, filterS, dc)
	if err != nil {
		t.Error(err)
	}
	// eEe.eClnt.Transport = new(mockClient)
	cgrEv.Event = map[string]interface{}{
		"test1": "value",
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
	errExpect := `unsupported protocol scheme ""`
	if err := eEe.ExportEvent(cgrEv); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but got %q", errExpect, err)
	}
}

func TestElasticExportEvent5(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
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
	eEe, err := NewElasticExporter(cgrCfg, 0, filterS, dc)
	if err != nil {
		t.Error(err)
	}
	// eEe.eClnt.Transport = new(mockClient)
	cgrEv.Event = map[string]interface{}{
		"test1": "value",
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
	errExpect := "inline parse error for string: <*wrong-type>"
	if err := eEe.ExportEvent(cgrEv); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but received %q", errExpect, err)
	}
	eEe.OnEvicted("test", "test")
}
