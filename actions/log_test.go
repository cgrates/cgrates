/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package actions

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestACExecuteActCDRLog(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ActionSCfg().CDRsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)}
	cfg.TemplatesCfg()[utils.MetaCdrLog][0].Filters = []string{"invalid_filter_value"}
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	apAction := &engine.APAction{
		ID:   "TEST_ACTION",
		Type: utils.CDRLog,
	}

	dataStorage := utils.MapStorage{
		utils.MetaReq: map[string]interface{}{
			utils.AccountField: "1001",
		},
		utils.MetaOpts: map[string]interface{}{
			utils.Usage: 10 * time.Minute,
		},
	}

	actCdrLG := &actCDRLog{
		config:  cfg,
		filterS: fltr,
		aCfg:    apAction,
	}
	expected := "NOT_FOUND:invalid_filter_value"
	if err := actCdrLG.execute(nil, dataStorage,
		utils.EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestACActLogger(t *testing.T) {
	actLog := &actLog{
		aCfg: &engine.APAction{
			ID:   "TEST_ACTION",
			Type: utils.CDRLog,
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
	apAction := &engine.APAction{
		ID:   "TEST_ACTION",
		Type: utils.CDRLog,
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
