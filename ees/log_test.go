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
	"reflect"
	"strings"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestNewLogEE(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dc, err := newEEMetrics("Local")
	if err != nil {
		t.Error(err)
	}

	expected := &LogEE{
		cfg: cfg.EEsCfg().ExporterCfg(utils.MetaDefault),
		dc:  dc,
	}

	rcv := NewLogEE(cfg.EEsCfg().ExporterCfg(utils.MetaDefault), dc)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", expected, rcv)
	}
}

func TestLogEEExportEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dc, err := newEEMetrics("Local")
	if err != nil {
		t.Error(err)
	}
	logEE := NewLogEE(cfg.EEsCfg().ExporterCfg(utils.MetaDefault), dc)
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

func TestLogEEGetMetrics(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dc, err := newEEMetrics("Local")
	if err != nil {
		t.Error(err)
	}
	dc.MapStorage = utils.MapStorage{
		"metric1": "value",
	}
	expected := &utils.SafeMapStorage{
		MapStorage: utils.MapStorage{
			"metric1": "value",
		},
	}
	logEE := NewLogEE(cfg.EEsCfg().ExporterCfg(utils.MetaDefault), dc)
	rcv := logEE.GetMetrics()
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %T \n but received \n %T", expected, rcv)
	}
}

func TestLogEEPrepareMap(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dc, err := newEEMetrics("Local")
	if err != nil {
		t.Error(err)
	}
	logEE := NewLogEE(cfg.EEsCfg().ExporterCfg(utils.MetaDefault), dc)
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
	dc, err := newEEMetrics("Local")
	if err != nil {
		t.Error(err)
	}
	logEE := NewLogEE(cfg.EEsCfg().ExporterCfg(utils.MetaDefault), dc)
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
