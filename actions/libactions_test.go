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
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	fltr := engine.NewFilterS(cfg, nil, nil)

	actCfg := []*utils.APAction{
		{Type: "not_a_type"},
	}

	expectedErr := "unsupported action type: <not_a_type>"
	if _, err := newActionersFromActions(context.Background(), new(utils.CGREvent), cfg, fltr, dm, nil,
		actCfg, utils.CGRateSorg); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

	actCfg = []*utils.APAction{
		{Type: utils.CDRLog},
		{Type: utils.MetaHTTPPost},
		{Type: utils.MetaExport},
		{Type: utils.MetaResetStatQueue},
		{Type: utils.MetaResetThreshold},
		{Type: utils.MetaAddBalance},
		{Type: utils.MetaSetBalance},
		{Type: utils.MetaRemBalance},
		{Type: utils.MetaDynamicThreshold},
		{Type: utils.MetaDynamicStats},
		{Type: utils.MetaDynamicAttribute},
		{Type: utils.MetaDynamicResource},
	}

	actHttp, err := newActHTTPPost(context.Background(), cfg.GeneralCfg().DefaultTenant, new(utils.CGREvent), new(engine.FilterS),
		cfg, &utils.APAction{Type: utils.MetaHTTPPost})
	if err != nil {
		t.Error(err)
	}

	expectedActs := []actioner{
		&actCDRLog{cfg, fltr, nil, &utils.APAction{Type: utils.CDRLog}},
		actHttp,
		&actExport{utils.CGRateSorg, cfg, nil, &utils.APAction{Type: utils.MetaExport}},
		&actResetStat{utils.CGRateSorg, cfg, nil, &utils.APAction{Type: utils.MetaResetStatQueue}},
		&actResetThreshold{utils.CGRateSorg, cfg, nil, &utils.APAction{Type: utils.MetaResetThreshold}},
		&actSetBalance{cfg, nil, fltr, &utils.APAction{Type: utils.MetaAddBalance}, utils.CGRateSorg, false},
		&actSetBalance{cfg, nil, fltr, &utils.APAction{Type: utils.MetaSetBalance}, utils.CGRateSorg, true},
		&actRemBalance{cfg, nil, fltr, &utils.APAction{Type: utils.MetaRemBalance}, utils.CGRateSorg},
		&actDynamicThreshold{cfg, nil, fltr, &utils.APAction{Type: utils.MetaDynamicThreshold}, utils.CGRateSorg, new(utils.CGREvent)},
		&actDynamicStats{cfg, nil, fltr, &utils.APAction{Type: utils.MetaDynamicStats}, utils.CGRateSorg, new(utils.CGREvent)},
		&actDynamicAttribute{cfg, nil, fltr, &utils.APAction{Type: utils.MetaDynamicAttribute}, utils.CGRateSorg, new(utils.CGREvent)},
		&actDynamicResource{cfg, nil, fltr, &utils.APAction{Type: utils.MetaDynamicResource}, utils.CGRateSorg, new(utils.CGREvent)},
	}

	acts, err := newActionersFromActions(context.Background(), new(utils.CGREvent), cfg, fltr, dm, nil, actCfg, utils.CGRateSorg)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(acts, expectedActs) {
		t.Errorf("Expected %+v, received %+v", expectedActs, acts)
	}
}

func TestACExecuteScheduledAction(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	acts := []actioner{
		&actCDRLog{cfg, fltr, nil, &utils.APAction{
			ID:   "TEST_ACTION",
			Type: utils.CDRLog,
		}},
	}
	dataStorage := utils.MapStorage{
		utils.LogLevelCfg: 7,
	}

	schedActs := newScheduledActs(nil, utils.CGRateSorg, "FirstAction",
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
