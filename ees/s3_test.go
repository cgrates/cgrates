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

func TestS3GetMetrics(t *testing.T) {
	safeMapStorage := &utils.SafeMapStorage{}
	pstr := &S3EE{
		dc: safeMapStorage,
	}
	result := pstr.GetMetrics()
	if result == nil {
		t.Errorf("GetMetrics() returned nil; expected a non-nil SafeMapStorage")
		return
	}
	if result != safeMapStorage {
		t.Errorf("GetMetrics() returned unexpected result; got %v, want %v", result, safeMapStorage)
	}
}

func TestClose(t *testing.T) {
	pstr := &S3EE{}
	err := pstr.Close()
	if err != nil {
		t.Errorf("Close() returned an error: %v; expected nil", err)
	}
}

func TestS3Cfg(t *testing.T) {
	expectedCfg := &config.EventExporterCfg{}
	pstr := &S3EE{
		cfg: expectedCfg,
	}
	actualCfg := pstr.Cfg()
	if actualCfg != expectedCfg {
		t.Errorf("Cfg() = %v; expected %v", actualCfg, expectedCfg)
	}
}
