//go:build integration
// +build integration

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

package engine

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestSetGetRemoveConfigSectionsDrvRedis(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	db, err := NewRedisStorage(cfg.DataDbCfg().Host+":"+cfg.DataDbCfg().Port, 10, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding, cfg.DataDbCfg().Opts.RedisMaxConns, cfg.DataDbCfg().Opts.RedisConnectAttempts,
		utils.EmptyString, false, 0, 0, 0, 0, 0, 150*time.Microsecond, 0, false, utils.EmptyString, utils.EmptyString, utils.EmptyString)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	sectionIDs := []string{"thresholds", "resources"}
	expected := make(map[string][]byte)

	// Try to retrieve the values before setting them (should receive an empty map)
	if rcv, err := db.GetConfigSectionsDrv(context.Background(), "1234", sectionIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	ms, err := utils.NewMarshaler(utils.JSON)
	if err != nil {
		t.Error(err)
	}
	thCfg := &config.ThresholdSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		Store_interval:        utils.StringPointer("2s"),
		String_indexed_fields: &[]string{"req.index11"},
		Prefix_indexed_fields: &[]string{"req.index22"},
		Suffix_indexed_fields: &[]string{"req.index33"},
		Actions_conns:         &[]string{"*internal"},
		Nested_fields:         utils.BoolPointer(true),
		Opts: &config.ThresholdsOptsJson{
			ProfileIDs: []*config.DynamicStringSliceOpt{
				{
					Tenant: "cgrates.org",
					Values: []string{"value1"},
				},
			},
			ProfileIgnoreFilters: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "true",
				},
			},
		},
	}
	thJsnCfg, err := ms.Marshal(thCfg)
	if err != nil {
		t.Error(err)
	}
	rsCfg := &config.ResourceSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		Thresholds_conns:      &[]string{"*birpc"},
		Store_interval:        utils.StringPointer("2s"),
		String_indexed_fields: &[]string{"*req.index11"},
		Prefix_indexed_fields: &[]string{"*req.index22"},
		Suffix_indexed_fields: &[]string{"*req.index33"},
		Nested_fields:         utils.BoolPointer(true),
		Opts: &config.ResourcesOptsJson{
			UsageID: []*config.DynamicInterfaceOpt{
				{
					Value: "usg2",
				},
			},
			UsageTTL: []*config.DynamicInterfaceOpt{
				{
					Value: "1m0s",
				},
			},
			Units: []*config.DynamicInterfaceOpt{
				{
					Value: "2",
				},
			},
		},
	}
	rsJsnCfg, err := ms.Marshal(rsCfg)
	if err != nil {
		t.Error(err)
	}
	sectData := map[string][]byte{
		"thresholds": thJsnCfg,
		"resources":  rsJsnCfg,
	}

	if err := db.SetConfigSectionsDrv(context.Background(), "1234", sectData); err != nil {
		t.Error(err)
	}

	if rcv, err := db.GetConfigSectionsDrv(context.Background(), "1234", sectionIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, sectData) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(sectData), utils.ToJSON(rcv))
	} else {
		rcvThCfg := &config.ThresholdSJsonCfg{}
		ms.Unmarshal(rcv["thresholds"], &rcvThCfg)
		if !reflect.DeepEqual(rcvThCfg, thCfg) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(thCfg), utils.ToJSON(rcvThCfg))
		}
		rcvRsCfg := &config.ResourceSJsonCfg{}
		ms.Unmarshal(rcv["resources"], &rcvRsCfg)
		if !reflect.DeepEqual(rcvRsCfg, rsCfg) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(rsCfg), utils.ToJSON(rcvRsCfg))
		}
	}

	if err := db.RemoveConfigSectionsDrv(context.Background(), "1234", sectionIDs); err != nil {
		t.Error(err)
	}

	if rcv, err := db.GetConfigSectionsDrv(context.Background(), "1234", sectionIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestSetGetRemoveConfigSectionsDrvMongo(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	db, err := NewMongoStorage(cfg.DataDbCfg().Opts.MongoConnScheme, cfg.DataDbCfg().Host, "27017", "10", cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding, utils.DataDB, nil, 10*time.Second)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	sectionIDs := []string{"thresholds", "resources"}

	// Try to retrieve the values before setting them (should receive an empty map)
	if rcv, err := db.GetConfigSectionsDrv(context.Background(), "1234", sectionIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", nil, utils.ToJSON(rcv))
	}

	ms, err := utils.NewMarshaler(utils.JSON)
	if err != nil {
		t.Error(err)
	}
	thCfg := &config.ThresholdSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		Store_interval:        utils.StringPointer("2s"),
		String_indexed_fields: &[]string{"req.index11"},
		Prefix_indexed_fields: &[]string{"req.index22"},
		Suffix_indexed_fields: &[]string{"req.index33"},
		Actions_conns:         &[]string{"*internal"},
		Nested_fields:         utils.BoolPointer(true),
		Opts: &config.ThresholdsOptsJson{
			ProfileIDs: []*config.DynamicStringSliceOpt{
				{
					Tenant: "cgrates.org",
					Values: []string{"value1"},
				},
			},
			ProfileIgnoreFilters: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "true",
				},
			},
		},
	}
	thJsnCfg, err := ms.Marshal(thCfg)
	if err != nil {
		t.Error(err)
	}
	rsCfg := &config.ResourceSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		Thresholds_conns:      &[]string{"*birpc"},
		Store_interval:        utils.StringPointer("2s"),
		String_indexed_fields: &[]string{"*req.index11"},
		Prefix_indexed_fields: &[]string{"*req.index22"},
		Suffix_indexed_fields: &[]string{"*req.index33"},
		Nested_fields:         utils.BoolPointer(true),
		Opts: &config.ResourcesOptsJson{
			UsageID: []*config.DynamicInterfaceOpt{
				{
					Value: "usg2",
				},
			},
			UsageTTL: []*config.DynamicInterfaceOpt{
				{
					Value: "1m0s",
				},
			},
			Units: []*config.DynamicInterfaceOpt{
				{
					Value: "2",
				},
			},
		},
	}
	rsJsnCfg, err := ms.Marshal(rsCfg)
	if err != nil {
		t.Error(err)
	}
	sectData := map[string][]byte{
		"thresholds": thJsnCfg,
		"resources":  rsJsnCfg,
	}

	if err := db.SetConfigSectionsDrv(context.Background(), "1234", sectData); err != nil {
		t.Error(err)
	}

	if rcv, err := db.GetConfigSectionsDrv(context.Background(), "1234", sectionIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, sectData) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(sectData), utils.ToJSON(rcv))
	} else {
		rcvThCfg := &config.ThresholdSJsonCfg{}
		ms.Unmarshal(rcv["thresholds"], &rcvThCfg)
		if !reflect.DeepEqual(rcvThCfg, thCfg) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(thCfg), utils.ToJSON(rcvThCfg))
		}
		rcvRsCfg := &config.ResourceSJsonCfg{}
		ms.Unmarshal(rcv["resources"], &rcvRsCfg)
		if !reflect.DeepEqual(rcvRsCfg, rsCfg) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(rsCfg), utils.ToJSON(rcvRsCfg))
		}
	}

	if err := db.RemoveConfigSectionsDrv(context.Background(), "1234", sectionIDs); err != nil {
		t.Error(err)
	}

	if rcv, err := db.GetConfigSectionsDrv(context.Background(), "1234", sectionIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", nil, utils.ToJSON(rcv))
	}
}

func TestSetGetRemoveConfigSectionsDrvInternal(t *testing.T) {
	db, _ := NewInternalDB(nil, nil, nil, nil)

	defer db.Close()
	sectionIDs := []string{"thresholds", "resources"}
	expected := make(map[string][]byte)

	// Try to retrieve the values before setting them (should receive an empty map)
	if rcv, err := db.GetConfigSectionsDrv(context.Background(), "1234", sectionIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	ms, err := utils.NewMarshaler(utils.JSON)
	if err != nil {
		t.Error(err)
	}
	thCfg := &config.ThresholdSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		Store_interval:        utils.StringPointer("2s"),
		String_indexed_fields: &[]string{"req.index11"},
		Prefix_indexed_fields: &[]string{"req.index22"},
		Suffix_indexed_fields: &[]string{"req.index33"},
		Actions_conns:         &[]string{"*internal"},
		Nested_fields:         utils.BoolPointer(true),
		Opts: &config.ThresholdsOptsJson{
			ProfileIDs: []*config.DynamicStringSliceOpt{
				{
					Tenant: "cgrates.org",
					Values: []string{"value1"},
				},
			},
			ProfileIgnoreFilters: []*config.DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "true",
				},
			},
		},
	}
	thJsnCfg, err := ms.Marshal(thCfg)
	if err != nil {
		t.Error(err)
	}
	rsCfg := &config.ResourceSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		Thresholds_conns:      &[]string{"*birpc"},
		Store_interval:        utils.StringPointer("2s"),
		String_indexed_fields: &[]string{"*req.index11"},
		Prefix_indexed_fields: &[]string{"*req.index22"},
		Suffix_indexed_fields: &[]string{"*req.index33"},
		Nested_fields:         utils.BoolPointer(true),
		Opts: &config.ResourcesOptsJson{
			UsageID: []*config.DynamicInterfaceOpt{
				{
					Value: "usg2",
				},
			},
			UsageTTL: []*config.DynamicInterfaceOpt{
				{
					Value: "1m0s",
				},
			},
			Units: []*config.DynamicInterfaceOpt{
				{
					Value: "2",
				},
			},
		},
	}
	rsJsnCfg, err := ms.Marshal(rsCfg)
	if err != nil {
		t.Error(err)
	}
	sectData := map[string][]byte{
		"thresholds": thJsnCfg,
		"resources":  rsJsnCfg,
	}

	if err := db.SetConfigSectionsDrv(context.Background(), "1234", sectData); err != nil {
		t.Error(err)
	}

	if rcv, err := db.GetConfigSectionsDrv(context.Background(), "1234", sectionIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, sectData) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(sectData), utils.ToJSON(rcv))
	} else {
		rcvThCfg := &config.ThresholdSJsonCfg{}
		ms.Unmarshal(rcv["thresholds"], &rcvThCfg)
		if !reflect.DeepEqual(rcvThCfg, thCfg) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(thCfg), utils.ToJSON(rcvThCfg))
		}
		rcvRsCfg := &config.ResourceSJsonCfg{}
		ms.Unmarshal(rcv["resources"], &rcvRsCfg)
		if !reflect.DeepEqual(rcvRsCfg, rsCfg) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(rsCfg), utils.ToJSON(rcvRsCfg))
		}
	}

	if err := db.RemoveConfigSectionsDrv(context.Background(), "1234", sectionIDs); err != nil {
		t.Error(err)
	}

	if rcv, err := db.GetConfigSectionsDrv(context.Background(), "1234", sectionIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}
