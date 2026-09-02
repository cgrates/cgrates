// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package actions

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestACExecuteActURLog(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	cfg.ActionSCfg().Conns[utils.MetaEEs] = []*config.DynamicConns{{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)}}}
	cfg.TemplatesCfg()[utils.MetaUrLog][0].Filters = []string{"invalid_filter_value"}
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm.SetCache(cacheS)
	fltr := engine.NewFilterS(cfg, nil, dm)
	apAction := &utils.APAction{
		ID:   "TEST_ACTION",
		Type: utils.MetaURLog,
	}

	dataStorage := utils.MapStorage{
		utils.MetaReq: map[string]any{
			utils.AccountField: "1001",
		},
		utils.MetaOpts: map[string]any{
			utils.Usage: 10 * time.Minute,
		},
	}

	actUrLG := &actURLog{
		config: cfg,
		fltrS:  fltr,
		aCfg:   apAction,
	}
	expected := "NOT_FOUND:invalid_filter_value"
	if err := actUrLG.execute(nil, dataStorage,
		utils.EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestACActLogger(t *testing.T) {
	actLog := &actLog{
		aCfg: &utils.APAction{
			ID:   "TEST_ACTION",
			Type: utils.MetaURLog,
		},
	}
	if rcv := actLog.id(); rcv != "TEST_ACTION" {
		t.Errorf("Expected %+v, received %+v", "TEST_ACTION", rcv)
	}
	if rcv := actLog.cfg(); !reflect.DeepEqual(rcv, actLog.aCfg) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(actLog.aCfg), utils.ToJSON(rcv))
	}
}

func TestACResetStatsAndThresholds(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	apAction := &utils.APAction{
		ID:   "TEST_ACTION",
		Type: utils.MetaURLog,
	}
	actResStats := &actResetStat{
		tnt:    "cgrates.org",
		config: cfg,
		aCfg:   apAction,
	}

	if rcv := actResStats.id(); rcv != "TEST_ACTION" {
		t.Errorf("Expected %+v, received %+v", "TEST_ACTION", rcv)
	}
	if rcv := actResStats.cfg(); !reflect.DeepEqual(rcv, actResStats.aCfg) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(actResStats.aCfg), utils.ToJSON(rcv))
	}

	actResTh := &actResetThreshold{
		tnt:    "cgrates.org",
		config: cfg,
		aCfg:   apAction,
	}

	if rcv := actResTh.id(); rcv != "TEST_ACTION" {
		t.Errorf("Expected %+v, received %+v", "TEST_ACTION", rcv)
	}
	if rcv := actResTh.cfg(); !reflect.DeepEqual(rcv, actResTh.aCfg) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(actResTh.aCfg), utils.ToJSON(rcv))
	}
}
