//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	db, err := NewRedisStorage("127.0.0.1:6379", 10, utils.CGRateSLwr,
		cfg.DbCfg().DBConns[utils.MetaDefault].Password, cfg.GeneralCfg().DBDataEncoding, cfg.DbCfg().DBConns[utils.MetaDefault].Opts.RedisMaxConns, cfg.DbCfg().DBConns[utils.MetaDefault].Opts.RedisConnectAttempts,
		utils.EmptyString, false, 0, 0, 0, 0, 0, 150*time.Microsecond, 0, false, utils.EmptyString, utils.EmptyString, utils.EmptyString, 1000, nil, nil)
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
		Conns: map[string][]*config.DynamicConns{
			utils.MetaActions: {
				{
					ConnIDs: []string{"*internal"},
				},
			},
		},
		Nested_fields: utils.BoolPointer(true),
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
		Enabled:         utils.BoolPointer(true),
		Indexed_selects: utils.BoolPointer(true),
		Conns: map[string][]*config.DynamicConns{
			utils.MetaThresholds: {
				{
					ConnIDs: []string{"*birpc"},
				},
			},
		},
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
	db, err := NewMongoStorage(cfg.DbCfg().DBConns[utils.MetaDefault].Opts.MongoConnScheme, cfg.DbCfg().DBConns[utils.MetaDefault].Host, "27017", "10", cfg.DbCfg().DBConns[utils.MetaDefault].User,
		cfg.DbCfg().DBConns[utils.MetaDefault].Password, cfg.GeneralCfg().DBDataEncoding, nil, 10*time.Second)
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
		Conns: map[string][]*config.DynamicConns{
			utils.MetaThresholds: {
				{
					ConnIDs: []string{"*internal"},
				},
			},
		},
		Nested_fields: utils.BoolPointer(true),
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
		Enabled:         utils.BoolPointer(true),
		Indexed_selects: utils.BoolPointer(true),
		Conns: map[string][]*config.DynamicConns{
			utils.MetaThresholds: {
				{
					ConnIDs: []string{"*thresholds"},
				},
			},
		},
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
		Conns: map[string][]*config.DynamicConns{
			utils.Actions: {
				{
					ConnIDs: []string{"*birpc"},
				},
			},
		},
		Nested_fields: utils.BoolPointer(true),
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
		Enabled:         utils.BoolPointer(true),
		Indexed_selects: utils.BoolPointer(true),
		Conns: map[string][]*config.DynamicConns{
			utils.MetaThresholds: {
				{
					ConnIDs: []string{"*birpc"},
				},
			},
		},
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
