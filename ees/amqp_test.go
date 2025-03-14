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
