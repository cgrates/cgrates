//go:build integration
// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package v1

import (
	"context"
	"flag"
	"net/rpc"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	runAMQPv1Test   = flag.Bool("amqpv1_cdre", false, "Run the integration test for the AMQPv1 exporter")
	amqpv1CfgPath   string
	amqpv1Cfg       *config.CGRConfig
	amqpv1RPC       *rpc.Client
	amqpv1ConfigDIR string
	amqpv1DialURL   string

	sTestsCDReAMQPv1 = []func(t *testing.T){
		testAMQPv1InitCfg,
		testAMQPv1InitDataDb,
		testAMQPv1ResetStorDb,
		testAMQPv1StartEngine,
		testAMQPv1RPCConn,
		testAMQPv1AddCDRs,
		testAMQPv1ExportCDRs,
		testAMQPv1VerifyExport,
		testAMQPv1KillEngine,
	}
)

func TestAMQPv1Export(t *testing.T) {
	if !*runAMQPv1Test {
		t.SkipNow()
	}
	amqpv1ConfigDIR = "cdre"
	for _, stest := range sTestsCDReAMQPv1 {
		t.Run(amqpv1ConfigDIR, stest)
	}
}

func testAMQPv1InitCfg(t *testing.T) {
	var err error
	amqpv1CfgPath = path.Join("/usr/share/cgrates", "conf", "samples", amqpv1ConfigDIR)
	amqpv1Cfg, err = config.NewCGRConfigFromPath(amqpv1CfgPath)
	if err != nil {
		t.Fatal(err)
	}
	amqpv1DialURL = amqpv1Cfg.CdreProfiles["amqpv1_exporter"].ExportPath
	amqpv1Cfg.DataFolderPath = "/usr/share/cgrates" // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(amqpv1Cfg)
}

func testAMQPv1InitDataDb(t *testing.T) {
	if err := engine.InitDataDb(amqpv1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testAMQPv1ResetStorDb(t *testing.T) {
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
	amqpv1RPC, err = newRPCClient(amqpv1Cfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testAMQPv1AddCDRs(t *testing.T) {
	storedCdrs := []*engine.CDR{
		{
			CGRID:       "Cdr1",
			OrderID:     101,
			ToR:         utils.VOICE,
			OriginID:    "OriginCDR1",
			OriginHost:  "192.168.1.1",
			Source:      "test",
			RequestType: utils.META_RATED,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "+4986517174963",
			SetupTime:   time.Now(),
			AnswerTime:  time.Now(),
			RunID:       utils.MetaDefault,
			Usage:       time.Duration(10) * time.Second,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
			Cost:        1.01,
		},
		{
			CGRID:       "Cdr2",
			OrderID:     102,
			ToR:         utils.VOICE,
			OriginID:    "OriginCDR2",
			OriginHost:  "192.168.1.1",
			Source:      "test2",
			RequestType: utils.META_RATED,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "+4986517174963",
			SetupTime:   time.Now(),
			AnswerTime:  time.Now(),
			RunID:       utils.MetaDefault,
			Usage:       time.Duration(5) * time.Second,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
			Cost:        1.01,
		},
		{
			CGRID:       "Cdr3",
			OrderID:     103,
			ToR:         utils.VOICE,
			OriginID:    "OriginCDR3",
			OriginHost:  "192.168.1.1",
			Source:      "test2",
			RequestType: utils.META_RATED,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "+4986517174963",
			SetupTime:   time.Now(),
			AnswerTime:  time.Now(),
			RunID:       utils.MetaDefault,
			Usage:       time.Duration(30) * time.Second,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
			Cost:        1.01,
		},
		{
			CGRID:       "Cdr4",
			OrderID:     104,
			ToR:         utils.VOICE,
			OriginID:    "OriginCDR4",
			OriginHost:  "192.168.1.1",
			Source:      "test3",
			RequestType: utils.META_RATED,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "+4986517174963",
			SetupTime:   time.Now(),
			AnswerTime:  time.Time{},
			RunID:       utils.MetaDefault,
			Usage:       time.Duration(0) * time.Second,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
	}
	for _, cdr := range storedCdrs {
		var reply string
		if err := amqpv1RPC.Call(utils.CDRsV1ProcessCDR, &engine.CDRWithArgDispatcher{CDR: cdr}, &reply); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if reply != utils.OK {
			t.Error("Unexpected reply received: ", reply)
		}
	}
	time.Sleep(100 * time.Millisecond)
}

func testAMQPv1ExportCDRs(t *testing.T) {
	attr := ArgExportCDRs{
		ExportArgs: map[string]any{
			utils.ExportTemplate: "amqpv1_exporter",
		},
		Verbose: true,
	}
	var rply RplExportedCDRs
	if err := amqpv1RPC.Call(utils.APIerSv1ExportCDRs, attr, &rply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(rply.ExportedCGRIDs) != 2 {
		t.Errorf("Unexpected number of CDR exported: %s ", utils.ToJSON(rply))
	}
}

func testAMQPv1VerifyExport(t *testing.T) {
	ctx := context.Background()
	// Create client
	client, err := amqp.Dial(ctx, amqpv1DialURL, nil)
	/* an alternative way to create the client
	client, err := amqp.Dial(ctx, "amqps://name-space.servicebus.windows.net", &amqp.ConnOptions{
		SASLType: amqp.SASLTypePlain("access-key-name", "access-key"),
	})
	*/

	if err != nil {
		t.Fatal("Dialing AMQP server:", err)
	}
	defer client.Close()

	// Open a session
	session, err := client.NewSession(ctx, nil)
	if err != nil {
		t.Fatal("Creating AMQP session:", err)
	}

	expCDRs := []string{
		`{"Account":"1001","CGRID":"Cdr2","Category":"call","Cost":"-1.0000","Destination":"+4986517174963","OriginID":"OriginCDR2","RunID":"*default","Source":"test2","Tenant":"cgrates.org","Usage":"5s"}`,
		`{"Account":"1001","CGRID":"Cdr3","Category":"call","Cost":"-1.0000","Destination":"+4986517174963","OriginID":"OriginCDR3","RunID":"*default","Source":"test2","Tenant":"cgrates.org","Usage":"30s"}`,
	}
	rplyCDRs := make([]string, 0)

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

	i := 0
	for i < 2 {
		// Receive next message
		msg, err := receiver.Receive(ctx, nil)
		if err != nil {
			t.Fatal("Reading message from AMQP:", err)
		}

		// Accept message
		receiver.AcceptMessage(ctx, msg)

		rplyCDRs = append(rplyCDRs, string(msg.GetData()))
		i++
	}

	sort.Strings(rplyCDRs)
	if len(rplyCDRs) != 2 {
		t.Fatalf("Expected 2 message received: %d", len(rplyCDRs))
	}
	if !reflect.DeepEqual(rplyCDRs, expCDRs) {
		t.Errorf("expected: %s,\nreceived: %s", expCDRs, rplyCDRs)
	}
}

func testAMQPv1KillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
