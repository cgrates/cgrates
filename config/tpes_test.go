// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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

func TestTpeSCfgLoadFromJSONCfg(t *testing.T) {
	tests := []struct {
		name     string
		jsonCfg  *TpeSCfgJson
		expected *TpeSCfg
	}{
		{
			name: "Load with enabled true",
			jsonCfg: &TpeSCfgJson{
				Enabled: utils.BoolPointer(true),
			},
			expected: &TpeSCfg{
				Enabled: true,
			},
		},
		{
			name: "Load with enabled false",
			jsonCfg: &TpeSCfgJson{
				Enabled: utils.BoolPointer(false),
			},
			expected: &TpeSCfg{
				Enabled: false,
			},
		},
		{
			name:    "Load with nil jsonCfg",
			jsonCfg: nil,
			expected: &TpeSCfg{
				Enabled: false,
			},
		},
		{
			name:    "Load with nil Enabled field",
			jsonCfg: &TpeSCfgJson{},
			expected: &TpeSCfg{
				Enabled: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &TpeSCfg{}
			if err := cfg.loadFromJSONCfg(tt.jsonCfg); err != nil {
				t.Fatalf("loadFromJSONCfg() error = %v", err)
			}
			if !reflect.DeepEqual(cfg, tt.expected) {
				t.Errorf("loadFromJSONCfg() got = %+v, want %+v", cfg, tt.expected)
			}
		})
	}
}

func TestTpeSCfgSName(t *testing.T) {
	cfg := TpeSCfg{}
	if name := cfg.SName(); name != TPeSJSON {
		t.Errorf("SName() = %v, want %v", name, TPeSJSON)
	}
}
