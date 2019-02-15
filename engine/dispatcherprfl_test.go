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
package engine

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestDispatcherConnsReorderFromIndex(t *testing.T) {
	dConns := DispatcherConns{
		{ID: "DSP_1", Weight: 30},
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_3", Weight: 10},
	}
	eConns := DispatcherConns{
		{ID: "DSP_1", Weight: 30},
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_3", Weight: 10},
	}
	if dConns.ReorderFromIndex(0); !reflect.DeepEqual(eConns, dConns) {
		t.Errorf("expecting: %+v, received: %+v", eConns, dConns)
	}
	dConns = DispatcherConns{
		{ID: "DSP_1", Weight: 30},
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_3", Weight: 10},
	}
	if dConns.ReorderFromIndex(3); !reflect.DeepEqual(eConns, dConns) {
		t.Errorf("expecting: %+v, received: %+v", eConns, dConns)
	}
	dConns = DispatcherConns{
		{ID: "DSP_1", Weight: 30},
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_3", Weight: 10},
	}
	eConns = DispatcherConns{
		{ID: "DSP_3", Weight: 10},
		{ID: "DSP_1", Weight: 30},
		{ID: "DSP_2", Weight: 20},
	}
	if dConns.ReorderFromIndex(2); !reflect.DeepEqual(eConns, dConns) {
		t.Errorf("expecting: %+v, received: %+v", eConns, dConns)
	}
	dConns = DispatcherConns{
		{ID: "DSP_1", Weight: 30},
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_3", Weight: 10},
	}
	eConns = DispatcherConns{
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_3", Weight: 10},
		{ID: "DSP_1", Weight: 30},
	}
	if dConns.ReorderFromIndex(1); !reflect.DeepEqual(eConns, dConns) {
		t.Errorf("expecting: %+v, received: %+v",
			utils.ToJSON(eConns), utils.ToJSON(dConns))
	}
}

func TestDispatcherConnsShuffle(t *testing.T) {
	dConns := DispatcherConns{
		{ID: "DSP_1", Weight: 30},
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_3", Weight: 10},
	}
	oConns := DispatcherConns{
		{ID: "DSP_1", Weight: 30},
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_3", Weight: 10},
	}
	if dConns.Shuffle(); dConns[0] == oConns[0] ||
		dConns[1] == oConns[1] || dConns[2] == oConns[2] {
		t.Errorf("received: %s", utils.ToJSON(dConns))
	}
}
