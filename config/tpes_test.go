/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PU:3474RPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

func TestTpeSCfgLoad(t *testing.T) {
	tp := &TpeSCfg{}
	ctx := &context.Context{}
	jsnCfg := new(mockDb)
	cgrcfg := &CGRConfig{}
	if err := tp.Load(ctx, jsnCfg, cgrcfg); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestTpeSCfgCloneSection(t *testing.T) {
	tp := TpeSCfg{}
	tpClone := tp.Clone()
	if rcv := tp.CloneSection(); !reflect.DeepEqual(rcv, tpClone) {
		t.Errorf("Expected <%v>, Received <%v>", tpClone, rcv)
	}
}

func TestDiffTpeSCfgJson(t *testing.T) {
	var d *TpeSCfgJson

	v1 := &TpeSCfg{Enabled: false}

	v2 := &TpeSCfg{Enabled: true}

	expected := &TpeSCfgJson{Enabled: utils.BoolPointer(true)}

	rcv := diffTpeSCfgJson(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &TpeSCfgJson{}
	rcv = diffTpeSCfgJson(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}
