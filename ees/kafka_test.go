// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ees

import (
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestKafkaEEConnect(t *testing.T) {
	kafkaEE := &KafkaEE{
		cfg: &config.EventExporterCfg{},
		em:  &utils.ExporterMetrics{},
	}
	err := kafkaEE.Connect()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestKafkaEE_Cfg(t *testing.T) {
	expectedCfg := &config.EventExporterCfg{}
	kafkaEE := &KafkaEE{
		cfg: expectedCfg,
	}
	result := kafkaEE.Cfg()
	if result != expectedCfg {
		t.Errorf("expected %v, got %v", expectedCfg, result)
	}
}

func TestKafkaEEGetMetrics(t *testing.T) {
	safeMapStorage := &utils.ExporterMetrics{}
	kafkaEE := &KafkaEE{
		em: safeMapStorage,
	}
	result := kafkaEE.GetMetrics()
	if result != safeMapStorage {
		t.Errorf("expected %v, got %v", safeMapStorage, result)
	}
}
