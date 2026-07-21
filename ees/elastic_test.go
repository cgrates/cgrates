// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ees

import (
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestGetMetrics(t *testing.T) {
	em, err := utils.NewExporterMetrics("", "Local")
	if err != nil {
		t.Fatal(err)
	}
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
			Opts:       &config.EventExporterOpts{},
		},
	}
	if err := ee.parseClientOpts(); err != nil {
		t.Error(err)
	}
	errExpect := `parse "/\x00": net/url: invalid control character in URL`
	if err := ee.Connect(); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %+v \n but got %+v", errExpect, err)
	}
}

func TestElasticExportEventErr(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	em, err := utils.NewExporterMetrics("", "Local")
	if err != nil {
		t.Fatal(err)
	}
	eEe, err := NewElasticEE(cgrCfg.EEsCfg().Exporters[0], em)
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
