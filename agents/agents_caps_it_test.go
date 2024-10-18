//go:build integration

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

package agents

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/fiorix/go-diameter/v4/diam"
	"github.com/fiorix/go-diameter/v4/diam/avp"
	"github.com/fiorix/go-diameter/v4/diam/datatype"
	"github.com/fiorix/go-diameter/v4/diam/dict"
)

func TestDiameterAgentCapsIT(t *testing.T) {
	var dbCfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbCfg = engine.DBCfg{
			DataDB: &engine.DBParams{
				Type: utils.StringPointer(utils.MetaInternal),
			},
			StorDB: &engine.DBParams{
				Type: utils.StringPointer(utils.MetaInternal),
			},
		}
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	jsonCfg := `{
"cores": {
	"caps": 2,
	"caps_strategy": "*busy",
	"shutdown_timeout": "5ms"
},
"sessions":{
	"enabled": true
},
"diameter_agent": {
	"enabled": true,
	"synced_conn_requests": true
}
}`

	ng := engine.TestEngine{
		ConfigJSON: jsonCfg,
		DBCfg:      dbCfg,
	}
	client, cfg := ng.Run(t)

	time.Sleep(10 * time.Millisecond) // wait for DiameterAgent service to start
	diamClient, err := NewDiameterClient(cfg.DiameterAgentCfg().Listen, "localhost",
		cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
		cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
		cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().ListenNet)
	if err != nil {
		t.Fatal(err)
	}

	reqIdx := 0
	sendCCR := func(t *testing.T, replyTimeout time.Duration, wg *sync.WaitGroup, wantResultCode string) {
		if wg != nil {
			defer wg.Done()
		}
		reqIdx++
		ccr := diam.NewRequest(diam.CreditControl, 4, nil)
		ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(fmt.Sprintf("session%d", reqIdx)))
		ccr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
		ccr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
		ccr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
		ccr.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(1))
		ccr.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(1))

		if err := diamClient.SendMessage(ccr); err != nil {
			t.Errorf("failed to send diameter message: %v", err)
			return
		}

		reply := diamClient.ReceivedMessage(replyTimeout)
		if reply == nil {
			t.Error("received empty reply")
			return
		}

		avps, err := reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
		if err != nil {
			t.Error(err)
			return
		}
		if len(avps) == 0 {
			t.Error("missing AVPs in reply")
			return
		}

		resultCode, err := diamAVPAsString(avps[0])
		if err != nil {
			t.Error(err)
		}
		if resultCode != wantResultCode {
			t.Errorf("Result-Code=%s, want %s", resultCode, wantResultCode)
		}
	}

	// There is currently no traffic. Expecting Result-Code 5012 (DIAMETER_UNABLE_TO_COMPLY),
	// because there are no request processors enabled.
	diamReplyTimeout := 2 * time.Second
	sendCCR(t, diamReplyTimeout, nil, "5012")

	// Caps limit is 2, therefore expecting the same result as in the scenario above.
	doneCh := simulateCapsTraffic(t, client, 1, *cfg.CoreSCfg())
	time.Sleep(time.Millisecond) // ensure traffic requests have been sent
	sendCCR(t, diamReplyTimeout, nil, "5012")
	<-doneCh

	// With caps limit reached, Result-Code 3004 (DIAMETER_TOO_BUSY) is expected.
	doneCh = simulateCapsTraffic(t, client, 2, *cfg.CoreSCfg())
	time.Sleep(time.Millisecond) // ensure traffic requests have been sent
	sendCCR(t, diamReplyTimeout, nil, "3004")
	<-doneCh

	// TODO: Check caps functionality with async diameter requests.
}

func simulateCapsTraffic(t *testing.T, client *birpc.Client, amount int, coresCfg config.CoreSCfg) <-chan struct{} {
	t.Helper()
	var wg sync.WaitGroup
	var reply string
	for i := range amount {
		wg.Add(1)
		go func() {
			t.Helper()
			defer wg.Done()
			if err := client.Call(context.Background(), utils.CoreSv1Sleep,
				&utils.DurationArgs{
					// Use the ShutdownTimeout CoreS setting
					// instead of having to pass an extra
					// variable to the function.
					Duration: coresCfg.ShutdownTimeout,
				}, &reply); err != nil {
				if coresCfg.CapsStrategy == utils.MetaBusy && i >= coresCfg.Caps {
					return // no need to handle errors for this scenario
				}
				t.Errorf("CoreSv1.Sleep unexpected error: %v", err)
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	return done
}
