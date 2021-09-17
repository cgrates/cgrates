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
		cfg: cfg.EEsCfg().GetDefaultExporter(),
		dc:  dc,
	}

	rcv := NewLogEE(cfg.EEsCfg().GetDefaultExporter(), dc)
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
	logEE := NewLogEE(cfg.EEsCfg().GetDefaultExporter(), dc)
	mp := map[string]interface{}{
		"field1": 2,
		"field2": "value",
	}
	logEE.ExportEvent(context.Background(), mp, "")
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
	logEE := NewLogEE(cfg.EEsCfg().GetDefaultExporter(), dc)
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
	logEE := NewLogEE(cfg.EEsCfg().GetDefaultExporter(), dc)
	mp := map[string]interface{}{
		"field1": 2,
		"field2": "value",
	}
	rcv, _ := logEE.PrepareMap(mp)
	if !reflect.DeepEqual(rcv, mp) {
		t.Errorf("Expected %v \n but received \n %v", mp, rcv)
	}
}

func TestLogEEPrepareOrderMap(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dc, err := newEEMetrics("Local")
	if err != nil {
		t.Error(err)
	}
	logEE := NewLogEE(cfg.EEsCfg().GetDefaultExporter(), dc)
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
	expected := make(map[string]interface{})
	expected["*path1"] = "payload"
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", mp, rcv)
	}
}
