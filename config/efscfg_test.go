/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
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

func TestEFsCfgLoad(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	m := &mockDb{}
	efsCfg := &EFsCfg{}
	if err := efsCfg.Load(context.Background(), m, cfg); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, Received errpr <%v>", utils.ErrNotImplemented, err)
	}
}

func TestEFsCfgLoadFromJSONCfg(t *testing.T) {
	efsCfg := &EFsCfg{}
	var jsonEFsCfg *EfsJsonCfg
	efsCfgClone := efsCfg.Clone()
	if err := efsCfg.loadFromJSONCfg(jsonEFsCfg); err != nil {
		t.Errorf("Expected error <nil>, Received error <%v>", err)
	} else if !reflect.DeepEqual(efsCfg, efsCfgClone) {
		t.Errorf("Expected EFsCfg to not change, was <%v>\nNow is <%v>",
			utils.ToJSON(efsCfgClone), utils.ToJSON(efsCfg))
	}
}
func TestEFsCfgLoadFromJSONCfgFailedPostsTTL(t *testing.T) {
	efsCfg := &EFsCfg{}
	jsonEFsCfg := &EfsJsonCfg{
		FailedPostsTTL: utils.StringPointer("failedPost"),
	}
	expErr := `time: invalid duration "failedPost"`
	if err := efsCfg.loadFromJSONCfg(jsonEFsCfg); err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}
}

func TestEFsCfgCloneSection(t *testing.T) {
	efsCfg := EFsCfg{
		Enabled: true,
	}
	if !reflect.DeepEqual(efsCfg.CloneSection(), efsCfg.Clone()) {
		t.Errorf("Expected EFsCfg.CloneSection result <%v>, Received result <%v>", efsCfg.Clone(), efsCfg.CloneSection())
	}

}
func TestDiffEFsJsonCfgEfsJsonCfgNil(t *testing.T) {
	var d *EfsJsonCfg

	v1 := &EFsCfg{}

	v2 := &EFsCfg{}

	expected := &EfsJsonCfg{}

	rcv := diffEFsJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

}

func TestDiffEFsJsonCfg(t *testing.T) {
	d := &EfsJsonCfg{}

	v1 := &EFsCfg{
		Enabled:        false,
		PosterAttempts: 2,
		FailedPostsDir: "2",
		FailedPostsTTL: 2,
	}

	v2 := &EFsCfg{
		Enabled:        true,
		PosterAttempts: 3,
		FailedPostsDir: "3",
		FailedPostsTTL: 3,
	}

	expected := &EfsJsonCfg{
		Enabled:        utils.BoolPointer(true),
		PosterAttempts: utils.IntPointer(3),
		FailedPostsDir: utils.StringPointer("3"),
		FailedPostsTTL: utils.StringPointer("3ns"),
	}

	rcv := diffEFsJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v2_2 := v1
	expected2 := &EfsJsonCfg{
		Enabled:        utils.BoolPointer(true),
		PosterAttempts: utils.IntPointer(3),
		FailedPostsDir: utils.StringPointer("3"),
		FailedPostsTTL: utils.StringPointer("3ns"),
	}

	rcv = diffEFsJsonCfg(d, v1, v2_2)
	if !reflect.DeepEqual(rcv, expected2) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected2), utils.ToJSON(rcv))
	}
}
