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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	elasticsearch "github.com/elastic/go-elasticsearch/v8"
)

func TestGetMetrics(t *testing.T) {
	em := utils.NewExporterMetrics("", time.Local)
	ee := &ElasticEE{
		em: em,
	}

	if rcv := ee.GetMetrics(); !reflect.DeepEqual(rcv, ee.em) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(rcv), utils.ToJSON(ee.em))
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
	if err := ee.parseClientOpts(); err != nil {
		t.Error(err)
	}
	errExpect := `cannot create client: cannot parse url: parse "/\x00": net/url: invalid control character in URL`
	if err := ee.Connect(); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %+v \n but got %+v", errExpect, err)
	}
}

func TestElasticExportEventErr(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	em := utils.NewExporterMetrics("", time.Local)
	eEe, err := NewElasticEE(cgrCfg.EEsCfg().Exporters[0], em)
	if err != nil {
		t.Error(err)
	}
	if err = eEe.Connect(); err != nil {
		t.Error(err)
	}
	errExpect := `an error happened during the Index query execution: unsupported protocol scheme ""`
	if err := eEe.ExportEvent([]byte{}, ""); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but got %q", errExpect, err)
	}
}

func TestElasticClose(t *testing.T) {
	elasticEE := &ElasticEE{
		client: &elasticsearch.TypedClient{},
	}
	err := elasticEE.Close()
	if elasticEE.client != nil {
		t.Errorf("expected eClnt to be nil, got %v", elasticEE.client)
	}
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestElasticConnect(t *testing.T) {
	t.Run("ClientAlreadyExists", func(t *testing.T) {

		elasticEE := &ElasticEE{
			client: &elasticsearch.TypedClient{},
		}

		err := elasticEE.Connect()

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if elasticEE.client == nil {
			t.Error("expected existing client to remain initialized")
		}
	})

	t.Run("ClientDoesNotExist", func(t *testing.T) {

		elasticEE := &ElasticEE{}

		err := elasticEE.Connect()

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if elasticEE.client == nil {
			t.Error("expected client to be initialized")
		}
	})
}
