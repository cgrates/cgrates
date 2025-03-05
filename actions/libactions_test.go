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

package actions

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

func TestACExecuteCDRLog(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	fltr := engine.NewFilterS(cfg, nil, nil)

	actCfg := []*engine.APAction{
		{Type: "not_a_type"},
	}

	expectedErr := "unsupported action type: <not_a_type>"
	if _, err := newActionersFromActions(context.Background(), new(utils.CGREvent), cfg, fltr, dm, nil,
		actCfg, "cgrates.org"); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

	actCfg = []*engine.APAction{
		{Type: utils.CDRLog},
		{Type: utils.MetaHTTPPost},
		{Type: utils.MetaExport},
		{Type: utils.MetaResetStatQueue},
		{Type: utils.MetaResetThreshold},
		{Type: utils.MetaAddBalance},
		{Type: utils.MetaSetBalance},
		{Type: utils.MetaRemBalance},
	}

	actHttp, err := newActHTTPPost(context.Background(), cfg.GeneralCfg().DefaultTenant, new(utils.CGREvent), new(engine.FilterS),
		cfg, &engine.APAction{Type: utils.MetaHTTPPost})
	if err != nil {
		t.Error(err)
	}

	expectedActs := []actioner{
		&actCDRLog{cfg, fltr, nil, &engine.APAction{Type: utils.CDRLog}},
		actHttp,
		&actExport{"cgrates.org", cfg, nil, &engine.APAction{Type: utils.MetaExport}},
		&actResetStat{"cgrates.org", cfg, nil, &engine.APAction{Type: utils.MetaResetStatQueue}},
		&actResetThreshold{"cgrates.org", cfg, nil, &engine.APAction{Type: utils.MetaResetThreshold}},
		&actSetBalance{cfg, nil, &engine.APAction{Type: utils.MetaAddBalance}, "cgrates.org", false},
		&actSetBalance{cfg, nil, &engine.APAction{Type: utils.MetaSetBalance}, "cgrates.org", true},
		&actRemBalance{cfg, nil, &engine.APAction{Type: utils.MetaRemBalance}, "cgrates.org"},
	}

	acts, err := newActionersFromActions(context.Background(), new(utils.CGREvent), cfg, fltr, dm, nil, actCfg, "cgrates.org")
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(acts, expectedActs) {
		t.Errorf("Expected %+v, received %+v", expectedActs, acts)
	}
}

func TestACExecuteScheduledAction(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	acts := []actioner{
		&actCDRLog{cfg, fltr, nil, &engine.APAction{
			ID:   "TEST_ACTION",
			Type: utils.CDRLog,
		}},
	}
	dataStorage := utils.MapStorage{
		utils.LogLevelCfg: 7,
	}

	schedActs := newScheduledActs(nil, "cgrates.org", "FirstAction",
		utils.EmptyString, utils.MetaTopUp, utils.MetaNow,
		dataStorage, acts)

	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 7)

	schedActs.ScheduledExecute()

	expected := "CGRateS <> [WARNING] executing action: <TEST_ACTION>, error: <no connection with CDR Server>"
	if rcv := buf.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}

	//not a data to save
	if err := schedActs.postExec(); err != nil {
		t.Error(err)
	}
}

func TestACActionTarget(t *testing.T) {
	if rcv := actionTarget(utils.MetaResetStatQueue); rcv != utils.MetaStats {
		t.Errorf("Expected %+v, received %+v", utils.MetaStats, rcv)
	}
	if rcv := actionTarget(utils.MetaResetThreshold); rcv != utils.MetaThresholds {
		t.Errorf("Expected %+v, received %+v", utils.MetaThresholds, rcv)
	}
	if rcv := actionTarget(utils.MetaAddBalance); rcv != utils.MetaAccounts {
		t.Errorf("Expected %+v, received %+v", utils.MetaAccounts, rcv)
	}
	if rcv := actionTarget("not_a_target"); rcv != utils.MetaNone {
		t.Errorf("Expected %+v, received %+v", utils.MetaNone, rcv)
	}
}
