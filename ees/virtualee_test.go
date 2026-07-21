// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ees

import (
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
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
	vEe := &VirtualEE{}
	if err := vEe.ExportEvent(context.Background(), []byte{}, ""); err != nil {
		t.Error(err)
	}
	vEe.Close()
}
