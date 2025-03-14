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
