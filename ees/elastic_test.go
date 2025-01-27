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
	errExpect := `an error happened during the Index query execution: unsupported protocol scheme ""`
	if err := eEe.ExportEvent(context.Background(), []byte{}, ""); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %q but got %q", errExpect, err)
	}
}
