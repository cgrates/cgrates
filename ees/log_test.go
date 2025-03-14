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
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestLogEECfg(t *testing.T) {
	expectedCfg := &config.EventExporterCfg{}
	vEe := &LogEE{
		cfg: expectedCfg,
	}
	result := vEe.Cfg()
	if result != expectedCfg {
		t.Errorf("expected %v, got %v", expectedCfg, result)
	}
}

func TestLogEEConnect(t *testing.T) {
	vEe := &LogEE{}
	err := vEe.Connect()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestLogEE_ExportEvent(t *testing.T) {

	cfg := &config.EventExporterCfg{ID: "testID"}

	vEe := &LogEE{
		cfg: cfg,
	}

	eventData := map[string]interface{}{
		"key": "value",
	}

	err := vEe.ExportEvent(eventData, "")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

}

func TestLogEE_Close(t *testing.T) {
	vEe := &LogEE{}
	err := vEe.Close()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestLogEE_GetMetrics(t *testing.T) {
	mockMetrics := &utils.ExporterMetrics{}

	vEe := &LogEE{
		dc: mockMetrics,
	}

	result := vEe.GetMetrics()

	if result != mockMetrics {
		t.Errorf("expected %v, got %v", mockMetrics, result)
	}
}

func TestLogEE_PrepareMap(t *testing.T) {

	cgrevent := &utils.CGREvent{}

	vEe := &LogEE{}

	_, err := vEe.PrepareMap(cgrevent)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

}

func TestNewLogEE(t *testing.T) {

	cfg := &config.EventExporterCfg{}
	dc := &utils.ExporterMetrics{}

	logEE := NewLogEE(cfg, dc)

	if logEE == nil {
		t.Fatal("NewLogEE returned nil")
	}

	if logEE.cfg != cfg {
		t.Errorf("Expected cfg to be %v, but got %v", cfg, logEE.cfg)
	}

	if logEE.dc != dc {
		t.Errorf("Expected dc to be %v, but got %v", dc, logEE.dc)
	}
}
