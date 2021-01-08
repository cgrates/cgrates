// +build race

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

package general_tests

import (
	"fmt"
	"testing"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var (
	sS      *sessions.SessionS
	cfg     *config.CGRConfig
	chrS    *engine.ChargerService
	filterS *engine.FilterS
	connMgr *engine.ConnManager
	dm      *engine.DataManager
	resp    *engine.Responder
)

// this structure will iplement rpcclient.ClientConnector
// and will read forever the Event map
type raceConn struct{}

func (_ raceConn) Call(serviceMethod string, args interface{}, reply interface{}) (err error) {
	cgrev := args.(*engine.ThresholdsArgsProcessEvent)
	for {
		for k := range cgrev.CGREvent.Event {
			if _, has := cgrev.CGREvent.Event[k]; !has {
				fmt.Println(1)
			}
		}
	}
}

// small test to detect races in sessions
func TestSessionSRace(t *testing.T) {
	// config
	var err error
	cfg = config.NewDefaultCGRConfig()
	cfg.SessionSCfg().Enabled = true
	cfg.SessionSCfg().ThreshSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	cfg.SessionSCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	cfg.SessionSCfg().RALsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)}
	cfg.SessionSCfg().DebitInterval = 10
	cfg.ChargerSCfg().Enabled = true

	cfg.CacheCfg().Partitions[utils.CacheChargerProfiles].Limit = -1
	cfg.CacheCfg().Partitions[utils.CacheAccounts].Limit = -1
	// cfg.GeneralCfg().ReplyTimeout = 30 * time.Second

	utils.Logger.SetLogLevel(7)
	// connManager
	raceChan := make(chan rpcclient.ClientConnector, 1)
	chargerSChan := make(chan rpcclient.ClientConnector, 1)
	respChan := make(chan rpcclient.ClientConnector, 1)
	raceChan <- new(raceConn)
	connMgr = engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds): raceChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers):   chargerSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder):  respChan,
	})

	// dataManager
	db := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(db, cfg.CacheCfg(), connMgr)
	engine.SetDataStorage(dm)

	// resp
	resp = &engine.Responder{
		ShdChan:          utils.NewSyncedChan(),
		MaxComputedUsage: cfg.RalsCfg().MaxComputedUsage,
	}
	respChan <- resp

	// filterS
	filterS = engine.NewFilterS(cfg, connMgr, dm)

	// chargerS
	if chrS, err = engine.NewChargerService(dm, filterS, cfg, connMgr); err != nil {
		t.Fatal(err)
	}
	chargerSChan <- v1.NewChargerSv1(chrS)

	// addCharger
	if err = dm.SetChargerProfile(&engine.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "Default",
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{"*none"},
		Weight:       20,
	}, true); err != nil {
		t.Fatal(err)
	}

	// set account
	if err = dm.SetAccount(&engine.Account{
		ID: utils.ConcatenatedKey("cgrates.org", "1001"),
		// AllowNegative: true,
		BalanceMap: map[string]engine.Balances{utils.MetaVoice: {{Value: float64(0 * time.Second), Weight: 10}}}}); err != nil {
		t.Fatal(err)
	}

	// sessionS
	sS = sessions.NewSessionS(cfg, dm, connMgr)

	// the race2
	rply := new(sessions.V1InitSessionReply)
	if err = sS.BiRPCv1InitiateSession(nil, &sessions.V1InitSessionArgs{
		InitSession:       true,
		ProcessThresholds: true,
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testSSv1ItProcessEventInitiateSession",
				Event: map[string]interface{}{
					utils.Tenant:       "cgrates.org",
					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     "testSSv1ItProcessEvent",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1001",
					// utils.RatingSubject: "*zero1ms",
					utils.CGRDebitInterval: 10,
					utils.Destination:      "1002",
					utils.SetupTime:        time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.AnswerTime:       time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
					utils.Usage:            0,
				},
			},
		},
	}, rply); err != utils.ErrPartiallyExecuted {
		t.Fatal(err)
	}
	// the race1
	rply2 := new(sessions.V1ProcessEventReply)
	if err = sS.BiRPCv1ProcessEvent(nil, &sessions.V1ProcessEventArgs{
		Flags: []string{utils.ConcatenatedKey(utils.MetaRALs, utils.MetaInitiate),
			utils.MetaThresholds},
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testSSv1ItProcessEventInitiateSession",
				Event: map[string]interface{}{
					utils.Tenant:       "cgrates.org",
					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     "testSSv1ItProcessEvent",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1001",
					// utils.RatingSubject: "*zero1ms",
					utils.CGRDebitInterval: 10,
					utils.Destination:      "1002",
					utils.SetupTime:        time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.AnswerTime:       time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
					utils.Usage:            0,
				},
			},
		},
	}, rply2); err != utils.ErrPartiallyExecuted {
		t.Fatal(err)
	}
}
