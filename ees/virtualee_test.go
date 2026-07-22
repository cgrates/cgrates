// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ees

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/google/go-cmp/cmp"
)

func TestVirtualEeGetMetrics(t *testing.T) {
	em, err := utils.NewExporterMetrics("", "Local")
	if err != nil {
		t.Fatal(err)
	}
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
