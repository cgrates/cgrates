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
	"flag"
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"

	amqpv1 "github.com/Azure/go-amqp"
	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	runAMQPv1Test  = flag.Bool("amqpv1_ees", false, "Run the integration test for the AMQPv1 exporter")
	amqpv1ConfDir  string
	amqpv1CfgPath  string
	amqpv1Cfg      *config.CGRConfig
	amqpv1RPC      *birpc.Client
	amqpv1DialURL  string
	amqpv1ConnOpts *amqpv1.ConnOptions

	sTestsAMQPv1 = []func(t *testing.T){
		testAMQPv1LoadConfig,
		testAMQPv1ResetDataDB,
		testAMQPv1ResetStorDB,
		testAMQPv1StartEngine,
		testAMQPv1RPCConn,

		testAMQPv1ExportEvent,
		testAMQPv1VerifyExport,

		testStopCgrEngine,
	}
)

func TestAMQPv1Export(t *testing.T) {
	if !*runAMQPv1Test {
		t.SkipNow()
	}
	amqpv1ConfDir = "ees_cloud"
	for _, stest := range sTestsAMQPv1 {
		t.Run(amqpv1ConfDir, stest)
	}
}

func testAMQPv1LoadConfig(t *testing.T) {
	var err error
	amqpv1CfgPath = path.Join(*utils.DataDir, "conf", "samples", amqpv1ConfDir)
	if amqpv1Cfg, err = config.NewCGRConfigFromPath(amqpv1CfgPath); err != nil {
		t.Error(err)
	}
	for _, value := range amqpv1Cfg.EEsCfg().Exporters {
		if value.ID == "amqpv1_test_file" {
			amqpv1DialURL = value.ExportPath
			if value.Opts.AMQP.Username != nil && value.Opts.AMQP.Password != nil {
				amqpv1ConnOpts = &amqpv1.ConnOptions{
					SASLType: amqpv1.SASLTypePlain(*value.Opts.AMQP.Username, *value.Opts.AMQP.Password),
				}
			}
		}
	}
}

func testAMQPv1ResetDataDB(t *testing.T) {
	if err := engine.InitDataDB(amqpv1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testAMQPv1ResetStorDB(t *testing.T) {
	if err := engine.InitStorDb(amqpv1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testAMQPv1StartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(amqpv1CfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testAMQPv1RPCConn(t *testing.T) {
	var err error
	amqpv1RPC, err = newRPCClient(amqpv1Cfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}

func testAMQPv1ExportEvent(t *testing.T) {
	ev := &engine.CGREventWithEeIDs{
		EeIDs: []string{"amqpv1_test_file"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "dataEvent",
			Event: map[string]any{
				utils.ToR:          utils.MetaData,
				utils.OriginID:     "abcdef",
				utils.OriginHost:   "192.168.1.1",
				utils.RequestType:  utils.MetaRated,
				utils.Tenant:       "AnotherTenant",
				utils.Category:     "call", //for data CDR use different Tenant
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Unix(1383813745, 0).UTC(),
				utils.AnswerTime:   time.Unix(1383813746, 0).UTC(),
				utils.Usage:        10 * time.Nanosecond,
				utils.RunID:        utils.MetaDefault,
				utils.Cost:         0.012,
			},
		},
	}

	var reply map[string]utils.MapStorage
	if err := amqpv1RPC.Call(context.Background(), utils.EeSv1ProcessEvent, ev, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(2 * time.Second)
}

func testAMQPv1VerifyExport(t *testing.T) {
	ctx := context.Background()

	// Create client
	client, err := amqpv1.Dial(ctx, amqpv1DialURL, amqpv1ConnOpts)
	if err != nil {
		t.Fatal("Dialing AMQP server:", err)
	}
	defer client.Close()

	// Open a session
	session, err := client.NewSession(ctx, nil)
	if err != nil {
		t.Fatal("Creating AMQP session:", err)
	}

	// Create a receiver
	receiver, err := session.NewReceiver(ctx, "/cgrates_cdrs", nil)
	if err != nil {
		t.Fatal("Creating receiver link:", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
		receiver.Close(ctx)
		cancel()
	}()

	// Receive message
	msg, err := receiver.Receive(ctx, nil)
	if err != nil {
		t.Fatal("Reading message from AMQP:", err)
	}

	// Accept message
	if err = receiver.AcceptMessage(context.Background(), msg); err != nil {
		t.Fatalf("Failure accepting message: %v", err)
	}

	expected := `{"Account":"1001","Category":"call","Destination":"1002","OriginID":"abcdef","RequestType":"*rated","RunID":"*default","Subject":"1001","Tenant":"AnotherTenant","ToR":"*data"}`
	if rply := string(msg.GetData()); rply != expected {
		t.Errorf("expected: %s, \nreceived: %s", expected, rply)
	}
}
