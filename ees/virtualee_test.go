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
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestVirtualEeGetMetrics(t *testing.T) {
	dc, err := newEEMetrics("Local")
	if err != nil {
		t.Error(err)
	}
	vEe := &VirtualEE{
		dc: dc,
	}

	if rcv := vEe.GetMetrics(); !reflect.DeepEqual(rcv, vEe.dc) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(rcv), utils.ToJSON(vEe.dc))
	}
}
func TestVirtualEeExportEvent(t *testing.T) {
	vEe := &VirtualEE{
		cfg: &config.EventExporterCfg{
			ID: "testEE",
		},
	}
	if err := vEe.ExportEvent([]byte{}, ""); err != nil {
		t.Error(err)
	}
	vEe.Close()
}

func TestVirtualEeConnect(t *testing.T) {
	vEe := &VirtualEE{}
	err := vEe.Connect()
	if err != nil {
		t.Errorf("Connect() returned an error: %v, expected nil", err)
	}
}

func TestVirtualEE_PrepareMap(t *testing.T) {
	vEe := &VirtualEE{}
	event := "test event"
	cgrEv := &utils.CGREvent{
		Tenant: "event",
	}

	// Test case when PrepareMap is called
	result, err := vEe.PrepareMap(cgrEv)

	// Check that the returned error is nil
	if err != nil {
		t.Errorf("PrepareMap() returned an error: %v, expected nil", err)
	}

	// Check that the returned result is the expected event
	if result != event {
		t.Errorf("PrepareMap() returned %v, expected %v", result, event)
	}
}
