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
	"net/rpc"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	amqpCfgPath   string
	amqpCfg       *config.CGRConfig
	amqpRPC       *rpc.Client
	amqpConfigDIR string

	sTestsCDReAMQP = []func(t *testing.T){

		testAMQPInitCfg,
		testAMQPInitDataDb,
		testAMQPResetStorDb,
		testAMQPStartEngine,
		testAMQPRPCConn,
		testAMQPMapAddCDRs,

		// tests for "*amqp_json_map" exporter
		testAMQPMapExportCDRs,
		testAMQPMapVerifyExport,

		// tests for "*amqp_json_cdr" exporter
		testAMQPCDRExportCDRs,
		testAMQPCDRVerifyExport,

		testAMQPKillEngine,
	}
)

func TestAMQPExport(t *testing.T) {
	amqpConfigDIR = "cdre"
	for _, stest := range sTestsCDReAMQP {
		t.Run(amqpConfigDIR, stest)
	}
}

func testAMQPInitCfg(t *testing.T) {
	var err error
	amqpCfgPath = path.Join("/usr/share/cgrates", "conf", "samples", amqpConfigDIR)
	amqpCfg, err = config.NewCGRConfigFromPath(amqpCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	amqpCfg.DataFolderPath = "/usr/share/cgrates" // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(amqpCfg)
}

func testAMQPInitDataDb(t *testing.T) {
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
	if _, err := engine.StopStartEngine(amqpCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testAMQPRPCConn(t *testing.T) {
	var err error
	amqpRPC, err = newRPCClient(amqpCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testAMQPMapAddCDRs(t *testing.T) {
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
			ExtraFields: map[string]string{"RawCost": "0.17"},
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
			RunID:       utils.MetaDefault,
			Usage:       time.Second,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
	}
	for _, cdr := range storedCdrs {
		var reply string
		if err := amqpRPC.Call(utils.CDRsV1ProcessCDR, &engine.CDRWithArgDispatcher{CDR: cdr}, &reply); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if reply != utils.OK {
			t.Error("Unexpected reply received: ", reply)
		}
	}
	time.Sleep(100 * time.Millisecond)
}

func testAMQPMapExportCDRs(t *testing.T) {
	attr := ArgExportCDRs{
		ExportArgs: map[string]any{
			utils.ExportTemplate: "amqp_exporter_map",
		},
		Verbose: true,
	}
	var rply RplExportedCDRs
	if err := amqpRPC.Call(utils.APIerSv1ExportCDRs, attr, &rply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(rply.ExportedCGRIDs) != 2 {
		t.Errorf("Unexpected number of CDR exported: %s ", utils.ToJSON(rply))
	}
}

func testAMQPMapVerifyExport(t *testing.T) {
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
	expCDRs := []string{
		`{"Account":"1001","CGRID":"Cdr2","Category":"call","Cost":"-1.0000","Destination":"+4986517174963","OriginID":"OriginCDR2","RunID":"*default","Source":"test2","Tenant":"cgrates.org","Usage":"5s"}`,
		`{"Account":"1001","CGRID":"Cdr3","Category":"call","Cost":"0.1700","Destination":"+4986517174963","OriginID":"OriginCDR3","RunID":"*default","Source":"test2","Tenant":"cgrates.org","Usage":"30s"}`,
	}
	rcvCDRs := make([]string, 0)
	waiting := true
	for waiting {
		select {
		case d := <-msgs:
			rcvCDRs = append(rcvCDRs, string(d.Body))
		case <-time.After(100 * time.Millisecond):
			waiting = false
		}
	}
	sort.Strings(rcvCDRs)
	if !reflect.DeepEqual(rcvCDRs, expCDRs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expCDRs, rcvCDRs)
	}
}

func testAMQPKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func testAMQPCDRExportCDRs(t *testing.T) {
	attr := ArgExportCDRs{
		ExportArgs: map[string]any{
			utils.ExportTemplate: "amqp_exporter_cdr",
		},
		Verbose: true,
	}
	var rply RplExportedCDRs
	if err := amqpRPC.Call(utils.APIerSv1ExportCDRs, attr, &rply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(rply.ExportedCGRIDs) != 1 {
		t.Errorf("Unexpected number of CDR exported: %s ", utils.ToJSON(rply))
	}
}

func testAMQPCDRVerifyExport(t *testing.T) {
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
	expCDR := `{"CGRID":"Cdr4","RunID":"*default","OrderID":4,"OriginHost":"192.168.1.1","Source":"test3","OriginID":"OriginCDR4","ToR":"*voice","RequestType":"*rated","Tenant":"cgrates.org","Category":"call","Account":"1001","Subject":"1001","Destination":"+4986517174963","SetupTime":"0001-01-01T00:00:00Z","AnswerTime":"0001-01-01T00:00:00Z","Usage":1000000000,"ExtraFields":{"field_extr1":"val_extr1","fieldextr2":"valextr2"},"ExtraInfo":"NOT_CONNECTED: RALs","Partial":false,"PreRated":false,"CostSource":"","Cost":-1,"CostDetails":null}`
	var rcvCDR string
	select {
	case d := <-msgs:
		rcvCDR = string(d.Body)
	case <-time.After(100 * time.Millisecond):
		t.Error("No message received from RabbitMQ")
	}
	if rcvCDR != expCDR {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expCDR, rcvCDR)
	}
}
