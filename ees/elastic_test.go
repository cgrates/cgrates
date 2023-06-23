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
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestGetMetrics(t *testing.T) {
	dc, err := newEEMetrics("Local")
	if err != nil {
		t.Error(err)
	}
	ee := &ElasticEE{
		dc: dc,
	}

	if rcv := ee.GetMetrics(); !reflect.DeepEqual(rcv, ee.dc) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(rcv), utils.ToJSON(ee.dc))
	}
}

func TestInitClient(t *testing.T) {
	ee := &ElasticEE{
		cfg: &config.EventExporterCfg{
			ExportPath: "/\x00",
			Opts: &config.EventExporterOpts{
				Els: &config.ElsOpts{},
				RPC: &config.RPCOpts{},
			},
		},
	}
	if err := ee.prepareOpts(); err != nil {
		t.Error(err)
	}
	errExpect := `cannot create client: cannot parse url: parse "/\x00": net/url: invalid control character in URL`
	if err := ee.Connect(); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %+v \n but got %+v", errExpect, err)
	}
}
func TestInitCase1(t *testing.T) {
	ee := &ElasticEE{
		cfg: &config.EventExporterCfg{
			Opts: &config.EventExporterOpts{
				Els: &config.ElsOpts{Index: utils.StringPointer("test")},
				RPC: &config.RPCOpts{},
			},
		},
	}
	if err := ee.prepareOpts(); err != nil {
		t.Error(err)
	}
	eeExpect := "test"
	if !reflect.DeepEqual(ee.indxOpts.Index, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", eeExpect, ee.indxOpts.Index)
	}
}

func TestInitCase2(t *testing.T) {
	ee := &ElasticEE{
		cfg: &config.EventExporterCfg{
			Opts: &config.EventExporterOpts{
				Els: &config.ElsOpts{
					IfPrimaryTerm: utils.IntPointer(20)},
				RPC: &config.RPCOpts{},
			},
		},
	}
	if err := ee.prepareOpts(); err != nil {
		t.Error(err)
	}
	eeExpect := utils.IntPointer(20)
	if !reflect.DeepEqual(ee.indxOpts.IfPrimaryTerm, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(ee.indxOpts.IfPrimaryTerm))
	}
}

func TestInitCase3(t *testing.T) {
	ee := &ElasticEE{
		cfg: &config.EventExporterCfg{
			Opts: &config.EventExporterOpts{
				Els: &config.ElsOpts{
					IfSeqNo: utils.IntPointer(20)},
				RPC: &config.RPCOpts{},
			},
		},
	}
	if err := ee.prepareOpts(); err != nil {
		t.Error(err)
	}
	eeExpect := utils.IntPointer(20)
	if !reflect.DeepEqual(ee.indxOpts.IfSeqNo, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(ee.indxOpts.IfSeqNo))
	}
}

func TestInitCase4(t *testing.T) {
	ee := &ElasticEE{
		cfg: &config.EventExporterCfg{
			Opts: &config.EventExporterOpts{
				Els: &config.ElsOpts{
					OpType: utils.StringPointer("test")},
				RPC: &config.RPCOpts{},
			},
		},
	}
	if err := ee.prepareOpts(); err != nil {
		t.Error(err)
	}
	eeExpect := "test"
	if !reflect.DeepEqual(ee.indxOpts.OpType, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(ee.indxOpts.OpType))
	}
}

func TestInitCase5(t *testing.T) {
	ee := &ElasticEE{
		cfg: &config.EventExporterCfg{
			Opts: &config.EventExporterOpts{
				Els: &config.ElsOpts{
					Pipeline: utils.StringPointer("test")},
				RPC: &config.RPCOpts{},
			},
		},
	}
	if err := ee.prepareOpts(); err != nil {
		t.Error(err)
	}
	eeExpect := "test"
	if !reflect.DeepEqual(ee.indxOpts.Pipeline, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(ee.indxOpts.Pipeline))
	}
}

func TestInitCase6(t *testing.T) {
	ee := &ElasticEE{
		cfg: &config.EventExporterCfg{
			Opts: &config.EventExporterOpts{
				Els: &config.ElsOpts{
					Routing: utils.StringPointer("test")},
				RPC: &config.RPCOpts{},
			},
		},
	}
	if err := ee.prepareOpts(); err != nil {
		t.Error(err)
	}
	eeExpect := "test"
	if !reflect.DeepEqual(ee.indxOpts.Routing, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(ee.indxOpts.Routing))
	}
}

func TestInitCase8(t *testing.T) {
	ee := &ElasticEE{
		cfg: &config.EventExporterCfg{
			Opts: &config.EventExporterOpts{
				Els: &config.ElsOpts{
					Version: utils.IntPointer(20),
				},
				RPC: &config.RPCOpts{},
			},
		},
	}
	if err := ee.prepareOpts(); err != nil {
		t.Error(err)
	}
	eeExpect := utils.IntPointer(20)
	if !reflect.DeepEqual(ee.indxOpts.Version, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(ee.indxOpts.Version))
	}
}

func TestInitCase9(t *testing.T) {
	ee := &ElasticEE{
		cfg: &config.EventExporterCfg{
			Opts: &config.EventExporterOpts{
				Els: &config.ElsOpts{
					VersionType: utils.StringPointer("test"),
				},
				RPC: &config.RPCOpts{},
			},
		},
	}
	if err := ee.prepareOpts(); err != nil {
		t.Error(err)
	}
	eeExpect := "test"
	if !reflect.DeepEqual(ee.indxOpts.VersionType, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(ee.indxOpts.VersionType))
	}
}

func TestInitCase10(t *testing.T) {
	ee := &ElasticEE{
		cfg: &config.EventExporterCfg{
			Opts: &config.EventExporterOpts{
				Els: &config.ElsOpts{
					WaitForActiveShards: utils.StringPointer("test")},
				RPC: &config.RPCOpts{},
			},
		},
	}
	if err := ee.prepareOpts(); err != nil {
		t.Error(err)
	}
	eeExpect := "test"
	if !reflect.DeepEqual(ee.indxOpts.WaitForActiveShards, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(ee.indxOpts.WaitForActiveShards))
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
	dc, err := newEEMetrics("Local")
	if err != nil {
		t.Error(err)
	}
	eEe, err := NewElasticEE(cgrCfg.EEsCfg().Exporters[0], dc)
	if err != nil {
		t.Error(err)
	}
	if err = eEe.Connect(); err != nil {
		t.Error(err)
	}
	eEe.eClnt.Transport = new(mockClientErr)
	if err := eEe.ExportEvent([]byte{}, ""); err != nil {
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
	dc, err := newEEMetrics("Local")
	if err != nil {
		t.Error(err)
	}
	eEe, err := NewElasticEE(cgrCfg.EEsCfg().Exporters[0], dc)
	if err != nil {
		t.Error(err)
	}
	if err = eEe.Connect(); err != nil {
		t.Error(err)
	}
	eEe.eClnt.Transport = new(mockClientErr2)

	errExpect := io.EOF
	if err := eEe.ExportEvent([]byte{}, ""); err == nil || err != errExpect {
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
	dc, err := newEEMetrics("Local")
	if err != nil {
		t.Error(err)
	}
	eEe, err := NewElasticEE(cgrCfg.EEsCfg().Exporters[0], dc)
	if err != nil {
		t.Error(err)
	}
	if err := eEe.prepareOpts(); err != nil {
		t.Error(err)
	}
	if err = eEe.Connect(); err != nil {
		t.Error(err)
	}
	eEe.eClnt.Transport = new(mockClient)
	errExpect := `the client noticed that the server is not Elasticsearch and we do not support this unknown product`
	cgrCfg.EEsCfg().Exporters[0].ComputeFields()
	if err := eEe.ExportEvent([]byte{}, ""); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but got %q", errExpect, err)
	}
}

func TestElasticExportEvent4(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dc, err := newEEMetrics("Local")
	if err != nil {
		t.Error(err)
	}
	eEe, err := NewElasticEE(cgrCfg.EEsCfg().Exporters[0], dc)
	if err != nil {
		t.Error(err)
	}
	if err = eEe.Connect(); err != nil {
		t.Error(err)
	}
	errExpect := `unsupported protocol scheme ""`
	if err := eEe.ExportEvent([]byte{}, ""); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but got %q", errExpect, err)
	}
}
