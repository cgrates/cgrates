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

package config

import (
	"errors"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

func TestIPsCfgLoad(t *testing.T) {
	t.Run("successful load", func(t *testing.T) {
		jsnCfg := &IPsJsonCfg{
			Enabled:                utils.BoolPointer(false),
			IndexedSelects:         utils.BoolPointer(true),
			StoreInterval:          utils.StringPointer("72h"),
			PrefixIndexedFields:    &[]string{"*req.prefix1", "*req.prefix2"},
			SuffixIndexedFields:    &[]string{"*req.suffix1"},
			ExistsIndexedFields:    &[]string{"*req.exists1", "*req.exists2"},
			NotExistsIndexedFields: &[]string{"*req.notexists1"},
			NestedFields:           utils.BoolPointer(false),
		}

		expected := &IPsCfg{
			Enabled:                false,
			IndexedSelects:         true,
			StoreInterval:          72 * time.Hour,
			PrefixIndexedFields:    &[]string{"*req.prefix1", "*req.prefix2"},
			SuffixIndexedFields:    &[]string{"*req.suffix1"},
			ExistsIndexedFields:    &[]string{"*req.exists1", "*req.exists2"},
			NotExistsIndexedFields: &[]string{"*req.notexists1"},
			NestedFields:           false,
			Opts:                   &IPsOpts{AllocationID: nil, TTL: nil},
		}

		cfg := &IPsCfg{Opts: &IPsOpts{}}
		ctx := &context.Context{}

		db := &mockDb{
			GetSectionF: func(_ *context.Context, section string, out any) error {
				if section != IPsJSON {
					return errors.New("unexpected section")
				}
				*out.(*IPsJsonCfg) = *jsnCfg
				return nil
			},
		}

		if err := cfg.Load(ctx, db, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		} else if !reflect.DeepEqual(expected, cfg) {
			t.Errorf("expected:\n%v\ngot:\n%v", utils.ToJSON(expected), utils.ToJSON(cfg))
		}
	})

	t.Run("GetSection returns error", func(t *testing.T) {
		cfg := &IPsCfg{Opts: &IPsOpts{}}
		ctx := &context.Context{}

		db := &mockDb{
			GetSectionF: func(_ *context.Context, _ string, _ any) error {
				return errors.New("error")
			},
		}

		err := cfg.Load(ctx, db, nil)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestIPsCfgSName(t *testing.T) {
	var cfg IPsCfg
	expected := IPsJSON
	if got := cfg.SName(); got != expected {
		t.Errorf("SName() = %v; want %v", got, expected)
	}
}

func TestIPsCfgloadFromJSONCfg(t *testing.T) {
	t.Run("full JSON config loads correctly", func(t *testing.T) {
		jsonCfg := &IPsJsonCfg{
			Enabled:                utils.BoolPointer(true),
			IndexedSelects:         utils.BoolPointer(false),
			StoreInterval:          utils.StringPointer("1h30m"),
			StringIndexedFields:    &[]string{"*req.s1"},
			PrefixIndexedFields:    &[]string{"*req.p1"},
			SuffixIndexedFields:    &[]string{"*req.sfx"},
			ExistsIndexedFields:    &[]string{"*req.e1"},
			NotExistsIndexedFields: &[]string{"*req.ne1"},
			NestedFields:           utils.BoolPointer(true),
			Opts:                   &IPsOptsJson{},
		}

		expected := &IPsCfg{
			Enabled:                true,
			IndexedSelects:         false,
			StoreInterval:          time.Hour + 30*time.Minute,
			StringIndexedFields:    &[]string{"*req.s1"},
			PrefixIndexedFields:    &[]string{"*req.p1"},
			SuffixIndexedFields:    &[]string{"*req.sfx"},
			ExistsIndexedFields:    &[]string{"*req.e1"},
			NotExistsIndexedFields: &[]string{"*req.ne1"},
			NestedFields:           true,
			Opts:                   &IPsOpts{},
		}

		cfg := &IPsCfg{Opts: &IPsOpts{}}
		err := cfg.loadFromJSONCfg(jsonCfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !reflect.DeepEqual(cfg, expected) {
			t.Errorf("expected:\n%v\ngot:\n%v", utils.ToJSON(expected), utils.ToJSON(cfg))
		}
	})

	t.Run("nil input returns nil error", func(t *testing.T) {
		cfg := &IPsCfg{Opts: &IPsOpts{}}
		if err := cfg.loadFromJSONCfg(nil); err != nil {
			t.Errorf("expected nil error, got: %v", err)
		}
	})

	t.Run("invalid store interval returns error", func(t *testing.T) {
		jsonCfg := &IPsJsonCfg{
			StoreInterval: utils.StringPointer("invalidDuration"),
		}
		cfg := &IPsCfg{Opts: &IPsOpts{}}
		err := cfg.loadFromJSONCfg(jsonCfg)
		if err == nil {
			t.Error("expected error for invalid duration, got nil")
		}
	})
}

func TestIPsCfgClone(t *testing.T) {
	orig := IPsCfg{
		Enabled:        true,
		IndexedSelects: false,
		StoreInterval:  42,
		NestedFields:   true,
		Opts: &IPsOpts{
			AllocationID: []*DynamicStringOpt{{value: "alloc1"}},
			TTL:          []*DynamicDurationOpt{{value: 123}},
		},
		StringIndexedFields:    &[]string{"c", "g"},
		PrefixIndexedFields:    &[]string{"r"},
		SuffixIndexedFields:    &[]string{"a", "t"},
		ExistsIndexedFields:    &[]string{"e"},
		NotExistsIndexedFields: &[]string{"s", "t"},
	}

	clone := orig.Clone()

	if !reflect.DeepEqual(clone, &orig) {
		t.Errorf("Clone() = %+v; want %+v", clone, orig)
	}

	*clone.StringIndexedFields = append(*clone.StringIndexedFields, "z")
	*clone.PrefixIndexedFields = append(*clone.PrefixIndexedFields, "y")
	clone.Enabled = false
	clone.Opts.AllocationID[0].value = "changed"

	if reflect.DeepEqual(clone, &orig) {
		t.Errorf("Clone() is not a deep copy; modifications to clone affected original")
	}

	if slices.Contains(*clone.StringIndexedFields, "z") && !slices.Contains(*orig.StringIndexedFields, "z") {
	} else {
		t.Errorf("StringIndexedFields slice not cloned properly")
	}

	if slices.Contains(*clone.PrefixIndexedFields, "y") && !slices.Contains(*orig.PrefixIndexedFields, "y") {
	} else {
		t.Errorf("PrefixIndexedFields slice not cloned properly")
	}

	if clone.Opts == orig.Opts {
		t.Errorf("Opts.Clone() did not create a new instance")
	}
}

func TestIPsCfgCloneSection(t *testing.T) {
	orig := IPsCfg{
		Enabled:        true,
		IndexedSelects: false,
		StoreInterval:  42,
		NestedFields:   true,
		Opts:           &IPsOpts{},
	}

	cloneSection := orig.CloneSection()
	clone, ok := cloneSection.(*IPsCfg)
	if !ok {
		t.Fatalf("CloneSection() did not return *IPsCfg, got %T", cloneSection)
	}

	expected := orig.Clone()
	if !reflect.DeepEqual(clone, expected) {
		t.Errorf("CloneSection() = %+v; want %+v", clone, expected)
	}
}
