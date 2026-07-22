// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ees

import (
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestAmqpGetMetrics(t *testing.T) {
	expectedMetrics := &utils.ExporterMetrics{}
	pstr := &AMQPee{
		em: expectedMetrics,
	}
	result := pstr.GetMetrics()
	if result != expectedMetrics {
		t.Errorf("expected metrics %v, got %v", expectedMetrics, result)
	}
}

func TestCfg(t *testing.T) {
	expectedCfg := &config.EventExporterCfg{ID: "testCfgID"}
	pstr := &AMQPee{
		cfg: expectedCfg,
	}
	result := pstr.Cfg()
	if result != expectedCfg {
		t.Errorf("expected cfg %v, got %v", expectedCfg, result)
	}
}

func TestAmqpToGetMetrics(t *testing.T) {
	expectedMetrics := &utils.ExporterMetrics{}
	amqp := &AMQPv1EE{
		em: expectedMetrics,
	}
	result := amqp.GetMetrics()
	if result != expectedMetrics {
		t.Errorf("GetMetrics() = %v; want %v", result, expectedMetrics)
	}
}

func TestCfgEvent(t *testing.T) {
	expectedCfg := &config.EventExporterCfg{}
	amqp := &AMQPv1EE{
		cfg: expectedCfg,
	}
	result := amqp.Cfg()
	if result != expectedCfg {
		t.Errorf("Cfg() = %v; want %v", result, expectedCfg)
	}
}
