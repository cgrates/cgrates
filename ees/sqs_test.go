// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ees

import (
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestGetMetricsT(t *testing.T) {
	em := &utils.ExporterMetrics{}
	pstr := &SQSee{em: em}
	result := pstr.GetMetrics()
	if result != em {
		t.Errorf("Expected %v, but got %v", em, result)
	}
}

func TestSqsClose(t *testing.T) {
	pstr := &SQSee{}
	err := pstr.Close()
	if err != nil {
		t.Errorf("Expected nil, but got %v", err)
	}
}

func TestSqsCfg(t *testing.T) {
	testCfg := &config.EventExporterCfg{}
	pstr := &SQSee{cfg: testCfg}
	result := pstr.Cfg()
	if result != testCfg {
		t.Errorf("Expected %v, but got %v", testCfg, result)
	}
}
