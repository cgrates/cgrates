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
