// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
		em: mockMetrics,
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
	em := &utils.ExporterMetrics{}

	logEE := NewLogEE(cfg, em)

	if logEE == nil {
		t.Fatal("NewLogEE returned nil")
	}

	if logEE.cfg != cfg {
		t.Errorf("Expected cfg to be %v, but got %v", cfg, logEE.cfg)
	}

	if logEE.em != em {
		t.Errorf("Expected em to be %v, but got %v", em, logEE.em)
	}
}
