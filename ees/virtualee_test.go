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
	vEe := &VirtualEE{}
	if err := vEe.ExportEvent([]byte{}, ""); err != nil {
		t.Error(err)
	}
	vEe.Close()
}
