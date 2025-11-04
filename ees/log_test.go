/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package ees

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestNewLogEE(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	em := utils.NewExporterMetrics("", time.Local)

	expected := &LogEE{
		cfg: cfg.EEsCfg().ExporterCfg(utils.MetaDefault),
		em:  em,
	}

	rcv := NewLogEE(cfg.EEsCfg().ExporterCfg(utils.MetaDefault), em)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", expected, rcv)
	}
}

func TestLogEEExportEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	em := utils.NewExporterMetrics("", time.Local)
	logEE := NewLogEE(cfg.EEsCfg().ExporterCfg(utils.MetaDefault), em)
	mp := map[string]any{
		"field1": 2,
		"field2": "value",
	}
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 7)

	logEE.ExportEvent(context.Background(), mp, "")
	exp := `CGRateS <> [INFO] <EEs> <*default> exported: <{"field1":2,"field2":"value"}>`
	if !strings.Contains(buf.String(), exp) {
		t.Errorf("Expected %v to contain %v", exp, buf.String())
	}
}

func TestLogEE_GetMetrics(t *testing.T) {
	mockMetrics := &utils.ExporterMetrics{}

	vEe := &LogEE{
		em: mockMetrics,
	}

	result := vEe.GetMetrics()

	if result != mockMetrics {
		t.Errorf("expected %v, got %v", mockMetrics, result)
	}
}

func TestLogEEPrepareMap(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	em := utils.NewExporterMetrics("", time.Local)
	logEE := NewLogEE(cfg.EEsCfg().ExporterCfg(utils.MetaDefault), em)
	mp := &utils.CGREvent{
		Event: map[string]any{
			"field1": 2,
			"field2": "value",
		},
	}
	rcv, _ := logEE.PrepareMap(mp)
	if !reflect.DeepEqual(rcv, mp.Event) {
		t.Errorf("Expected %v \n but received \n %v", mp, rcv)
	}
}

func TestLogEEPrepareOrderMap(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	em := utils.NewExporterMetrics("", time.Local)
	logEE := NewLogEE(cfg.EEsCfg().ExporterCfg(utils.MetaDefault), em)
	mp := utils.NewOrderedNavigableMap()
	fullPath := &utils.FullPath{
		PathSlice: []string{"*path1"},
		Path:      "*test",
	}
	val := &utils.DataLeaf{
		Data: "payload",
	}
	mp.Append(fullPath, val)
	rcv, _ := logEE.PrepareOrderMap(mp)
	expected := make(map[string]any)
	expected["*path1"] = "payload"
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", mp, rcv)
	}
}
