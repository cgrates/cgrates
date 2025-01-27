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

	"github.com/cgrates/birpc/context"
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
			Opts:       &config.EventExporterOpts{},
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
				ElsIndex: utils.StringPointer("test"),
			},
		},
	}
	if err := ee.prepareOpts(); err != nil {
		t.Error(err)
	}
	eeExpect := "test"
	if !reflect.DeepEqual(ee.indexReqOpts.Index, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", eeExpect, ee.indexReqOpts.Index)
	}
}

func TestInitCase2(t *testing.T) {
	ee := &ElasticEE{
		cfg: &config.EventExporterCfg{
			Opts: &config.EventExporterOpts{
				ElsIfPrimaryTerm: utils.IntPointer(20),
			},
		},
	}
	if err := ee.prepareOpts(); err != nil {
		t.Error(err)
	}
	eeExpect := utils.IntPointer(20)
	if !reflect.DeepEqual(ee.indexReqOpts.IfPrimaryTerm, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(ee.indexReqOpts.IfPrimaryTerm))
	}
}

func TestInitCase3(t *testing.T) {
	ee := &ElasticEE{
		cfg: &config.EventExporterCfg{
			Opts: &config.EventExporterOpts{
				ElsIfSeqNo: utils.IntPointer(20),
			},
		},
	}
	if err := ee.prepareOpts(); err != nil {
		t.Error(err)
	}
	eeExpect := utils.IntPointer(20)
	if !reflect.DeepEqual(ee.indexReqOpts.IfSeqNo, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(ee.indexReqOpts.IfSeqNo))
	}
}

func TestInitCase4(t *testing.T) {
	ee := &ElasticEE{
		cfg: &config.EventExporterCfg{
			Opts: &config.EventExporterOpts{
				ElsOpType: utils.StringPointer("test"),
			},
		},
	}
	if err := ee.prepareOpts(); err != nil {
		t.Error(err)
	}
	eeExpect := "test"
	if !reflect.DeepEqual(ee.indexReqOpts.OpType, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(ee.indexReqOpts.OpType))
	}
}

func TestInitCase5(t *testing.T) {
	ee := &ElasticEE{
		cfg: &config.EventExporterCfg{
			Opts: &config.EventExporterOpts{
				ElsPipeline: utils.StringPointer("test"),
			},
		},
	}
	if err := ee.prepareOpts(); err != nil {
		t.Error(err)
	}
	eeExpect := "test"
	if !reflect.DeepEqual(ee.indexReqOpts.Pipeline, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(ee.indexReqOpts.Pipeline))
	}
}

func TestInitCase6(t *testing.T) {
	ee := &ElasticEE{
		cfg: &config.EventExporterCfg{
			Opts: &config.EventExporterOpts{
				ElsRouting: utils.StringPointer("test"),
			},
		},
	}
	if err := ee.prepareOpts(); err != nil {
		t.Error(err)
	}
	eeExpect := "test"
	if !reflect.DeepEqual(ee.indexReqOpts.Routing, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(ee.indexReqOpts.Routing))
	}
}

func TestInitCase8(t *testing.T) {
	ee := &ElasticEE{
		cfg: &config.EventExporterCfg{
			Opts: &config.EventExporterOpts{
				ElsVersion: utils.IntPointer(20),
			},
		},
	}
	if err := ee.prepareOpts(); err != nil {
		t.Error(err)
	}
	eeExpect := utils.IntPointer(20)
	if !reflect.DeepEqual(ee.indexReqOpts.Version, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(ee.indexReqOpts.Version))
	}
}

func TestInitCase9(t *testing.T) {
	ee := &ElasticEE{
		cfg: &config.EventExporterCfg{
			Opts: &config.EventExporterOpts{
				ElsVersionType: utils.StringPointer("test"),
			},
		},
	}
	if err := ee.prepareOpts(); err != nil {
		t.Error(err)
	}
	eeExpect := "test"
	if !reflect.DeepEqual(ee.indexReqOpts.VersionType, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(ee.indexReqOpts.VersionType))
	}
}

func TestInitCase10(t *testing.T) {
	ee := &ElasticEE{
		cfg: &config.EventExporterCfg{
			Opts: &config.EventExporterOpts{
				ElsWaitForActiveShards: utils.StringPointer("test"),
			},
		},
	}
	if err := ee.prepareOpts(); err != nil {
		t.Error(err)
	}
	eeExpect := "test"
	if !reflect.DeepEqual(ee.indexReqOpts.WaitForActiveShards, eeExpect) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(eeExpect), utils.ToJSON(ee.indexReqOpts.WaitForActiveShards))
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
	eEe.client.Transport = new(mockClientErr)
	if err := eEe.ExportEvent(context.Background(), []byte{}, ""); err != nil {
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
	eEe.client.Transport = new(mockClientErr2)

	errExpect := io.EOF
	if err := eEe.ExportEvent(context.Background(), []byte{}, ""); err == nil || err != errExpect {
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
	if err = eEe.Connect(); err != nil {
		t.Error(err)
	}
	eEe.client.Transport = new(mockClient)
	errExpect := `the client noticed that the server is not Elasticsearch and we do not support this unknown product`

	cgrCfg.EEsCfg().Exporters[0].ComputeFields()
	if err := eEe.ExportEvent(context.Background(), []byte{}, ""); err == nil || err.Error() != errExpect {
		t.Error(err)
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
	if err := eEe.ExportEvent(context.Background(), []byte{}, ""); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but got %q", errExpect, err)
	}
}
