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
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/google/go-cmp/cmp"
)

func TestVirtualEeGetMetrics(t *testing.T) {
	em := utils.NewExporterMetrics("", time.Local)
	vEe := &VirtualEE{
		em: em,
	}

	if rcv := vEe.GetMetrics(); !reflect.DeepEqual(rcv, vEe.em) {
		t.Errorf("Expected %+v \n but got %+v", utils.ToJSON(rcv), utils.ToJSON(vEe.em))
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
		t.Errorf("Connect() err = %v, want nil", err)
	}
}

func TestVirtualEePrepareMap(t *testing.T) {
	vEe := &VirtualEE{}
	cgrEv := &utils.CGREvent{
		Tenant: "event",
		Event: map[string]any{
			"Key": "value",
		},
	}
	want := map[string]any{
		"Key": "value",
	}
	got, err := vEe.PrepareMap(cgrEv)
	if err != nil {
		t.Errorf("PrepareMap() returned an error: %v, expected nil", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("PrepareMap() returned an unexpected value(-want +got): \n%s", diff)
	}
}
