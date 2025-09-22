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

func TestActionTarget(t *testing.T) {
	tests := []struct {
		name string
		act  string
		want string
	}{
		{
			name: "ResetStatQueue",
			act:  utils.MetaResetStatQueue,
			want: utils.MetaStats,
		},
		{
			name: "DynamicStats",
			act:  utils.MetaDynamicStats,
			want: utils.MetaStats,
		},
		{
			name: "ResetThreshold",
			act:  utils.MetaResetThreshold,
			want: utils.MetaThresholds,
		},
		{
			name: "DynamicThreshold",
			act:  utils.MetaDynamicThreshold,
			want: utils.MetaThresholds,
		},
		{
			name: "AddBalance",
			act:  utils.MetaAddBalance,
			want: utils.MetaAccounts,
		},
		{
			name: "SetBalance",
			act:  utils.MetaSetBalance,
			want: utils.MetaAccounts,
		},
		{
			name: "RemBalance",
			act:  utils.MetaRemBalance,
			want: utils.MetaAccounts,
		},
		{
			name: "DynamicAttribute",
			act:  utils.MetaDynamicAttribute,
			want: utils.MetaAttributes,
		},
		{
			name: "DynamicResource",
			act:  utils.MetaDynamicResource,
			want: utils.MetaResources,
		},
		{
			name: "DynamicTrend",
			act:  utils.MetaDynamicTrend,
			want: utils.MetaTrends,
		},
		{
			name: "DynamicRanking",
			act:  utils.MetaDynamicRanking,
			want: utils.MetaRankings,
		},
		{
			name: "DynamicFilter",
			act:  utils.MetaDynamicFilter,
			want: utils.MetaFilters,
		},
		{
			name: "DynamicRoute",
			act:  utils.MetaDynamicRoute,
			want: utils.MetaRoutes,
		},
		{
			name: "DynamicRate",
			act:  utils.MetaDynamicRate,
			want: utils.MetaRates,
		},
		{
			name: "DynamicIP",
			act:  utils.MetaDynamicIP,
			want: utils.MetaIPs,
		},
		{
			name: "DynamicAction",
			act:  utils.MetaDynamicAction,
			want: utils.MetaActions,
		},
		{
			name: "UnknownAction",
			act:  "unknown",
			want: utils.MetaNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := actionTarget(tt.act); got != tt.want {
				t.Errorf("actionTarget(%q) = %q, want %q", tt.act, got, tt.want)
			}
		})
	}
}

func TestNewActioner(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	fltr := engine.NewFilterS(cfg, nil, nil)

	ctx := context.Background()
	cgrEv := new(utils.CGREvent)
	connMgr := new(engine.ConnManager)
	tnt := utils.CGRateSorg

	tests := []struct {
		name    string
		aCfg    *utils.APAction
		wantErr string
	}{
		{
			name:    "UnsupportedAction",
			aCfg:    &utils.APAction{Type: "not_a_type"},
			wantErr: "unsupported action type: <not_a_type>",
		},
		{
			name: "MetaLog",
			aCfg: &utils.APAction{Type: utils.MetaLog},
		},
		{
			name: "CDRLog",
			aCfg: &utils.APAction{Type: utils.CDRLog},
		},
		{
			name: "MetaHTTPPost",
			aCfg: &utils.APAction{Type: utils.MetaHTTPPost},
		},
		{
			name: "MetaExport",
			aCfg: &utils.APAction{Type: utils.MetaExport},
		},
		{
			name: "MetaResetStatQueue",
			aCfg: &utils.APAction{Type: utils.MetaResetStatQueue},
		},
		{
			name: "MetaResetThreshold",
			aCfg: &utils.APAction{Type: utils.MetaResetThreshold},
		},
		{
			name: "MetaAddBalance",
			aCfg: &utils.APAction{Type: utils.MetaAddBalance},
		},
		{
			name: "MetaSetBalance",
			aCfg: &utils.APAction{Type: utils.MetaSetBalance},
		},
		{
			name: "MetaRemBalance",
			aCfg: &utils.APAction{Type: utils.MetaRemBalance},
		},
		{
			name: "MetaDynamicThreshold",
			aCfg: &utils.APAction{Type: utils.MetaDynamicThreshold},
		},
		{
			name: "MetaDynamicStats",
			aCfg: &utils.APAction{Type: utils.MetaDynamicStats},
		},
		{
			name: "MetaDynamicAttribute",
			aCfg: &utils.APAction{Type: utils.MetaDynamicAttribute},
		},
		{
			name: "MetaDynamicResource",
			aCfg: &utils.APAction{Type: utils.MetaDynamicResource},
		},
		{
			name: "MetaDynamicTrend",
			aCfg: &utils.APAction{Type: utils.MetaDynamicTrend},
		},
		{
			name: "MetaDynamicRanking",
			aCfg: &utils.APAction{Type: utils.MetaDynamicRanking},
		},
		{
			name: "MetaDynamicFilter",
			aCfg: &utils.APAction{Type: utils.MetaDynamicFilter},
		},
		{
			name: "MetaDynamicRoute",
			aCfg: &utils.APAction{Type: utils.MetaDynamicRoute},
		},
		{
			name: "MetaDynamicRate",
			aCfg: &utils.APAction{Type: utils.MetaDynamicRate},
		},
		{
			name: "MetaDynamicIP",
			aCfg: &utils.APAction{Type: utils.MetaDynamicIP},
		},
		{
			name: "MetaDynamicAction",
			aCfg: &utils.APAction{Type: utils.MetaDynamicAction},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			act, err := newActioner(ctx, cgrEv, cfg, fltr, dm, connMgr, tt.aCfg, tnt)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Errorf("Expected error: %q, got: %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if act == nil {
				t.Errorf("Expected non-nil action, got nil")
			}
		})
	}
}
