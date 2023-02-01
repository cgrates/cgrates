//go:build integration
// +build integration

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

package ees

import (
	"net/rpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/streadway/amqp"
)

var (
	amqpConfDir string
	amqpCfgPath string
	amqpCfg     *config.CGRConfig
	amqpRPC     *rpc.Client

	sTestsAMQP = []func(t *testing.T){
		testAMQPLoadConfig,
		testAMQPResetDataDB,
		testAMQPResetStorDb,
		testAMQPStartEngine,
		testAMQPRPCConn,
		testAMQPExportEvent,
		testAMQPVerifyExport,
		testStopCgrEngine,
	}
)

func TestAMQPExport(t *testing.T) {
	amqpConfDir = "ees_amqp"
	for _, stest := range sTestsAMQP {
		t.Run(amqpConfDir, stest)
	}
}

func testAMQPLoadConfig(t *testing.T) {
	var err error
	amqpCfgPath = path.Join(*dataDir, "conf", "samples", amqpConfDir)
	if amqpCfg, err = config.NewCGRConfigFromPath(amqpCfgPath); err != nil {
		t.Error(err)
	}
}

func testAMQPResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(amqpCfg); err != nil {
		t.Fatal(err)
	}
}

func testAMQPResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(amqpCfg); err != nil {
		t.Fatal(err)
	}
}

func testAMQPStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(amqpCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testAMQPRPCConn(t *testing.T) {
	var err error
	amqpRPC, err = newRPCClient(amqpCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}

func testAMQPExportEvent(t *testing.T) {
	ev := &engine.CGREventWithEeIDs{
		EeIDs: []string{"amqp_test_file"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Time:   utils.TimePointer(time.Now()),
			Event: map[string]interface{}{
				utils.CGRID:        utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "dsafdsaf",
				utils.OriginHost:   "192.168.1.1",
				utils.RequestType:  utils.MetaRated,
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Unix(1383813745, 0).UTC(),
				utils.AnswerTime:   time.Unix(1383813746, 0).UTC(),
				utils.Usage:        10 * time.Second,
				utils.RunID:        utils.MetaDefault,
				utils.Cost:         1.01,
			},
		},
	}

	var reply map[string]utils.MapStorage
	if err := amqpRPC.Call(utils.EeSv1ProcessEvent, ev, &reply); err != nil {
		t.Error(err)
	}

	time.Sleep(2 * time.Second)
}

func testAMQPVerifyExport(t *testing.T) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer ch.Close()
	q, err := ch.QueueDeclare("cgrates_cdrs", true, false, false, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	msgs, err := ch.Consume(q.Name, utils.EmptyString, true, false, false, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	expCDR := `{"Account":"1001","CGRID":"dbafe9c8614c785a65aabd116dd3959c3c56f7f6","Category":"call","Destination":"1002","OriginID":"dsafdsaf","RequestType":"*rated","RunID":"*default","Subject":"1001","Tenant":"cgrates.org","ToR":"*voice"}`
	select {
	case d := <-msgs:
		rcvCDR := string(d.Body)
		if rcvCDR != expCDR {
			t.Errorf("unexpected CDR received: <%s>", rcvCDR)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("No message received from RabbitMQ")
	}
}
