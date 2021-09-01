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

package general_tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
	kafka "github.com/segmentio/kafka-go"
	"github.com/streadway/amqp"
)

var (
	cdrsMasterCfgPath, cdrsSlaveCfgPath string
	cdrsMasterCfgDIR, cdrsSlaveCfgDIR   string
	cdrsMasterCfg, cdrsSlaveCfg         *config.CGRConfig
	cdrsMasterRpc                       *rpcclient.RPCClient
	httpCGRID                           = utils.UUIDSha1Prefix()
	amqpCGRID                           = utils.UUIDSha1Prefix()
	failoverContent                     = []interface{}{[]byte(fmt.Sprintf(`{"CGRID":"%s"}`, httpCGRID)), []byte(fmt.Sprintf(`{"CGRID":"%s"}`, amqpCGRID))}

	sTestsCDRsOnExp = []func(t *testing.T){
		testCDRsOnExpInitConfig,
		testCDRsOnExpInitCdrDb,
		testCDRsOnExpStartMasterEngine,
		testCDRsOnExpStartSlaveEngine,
		testCDRsOnExpAMQPQueuesCreation,
		testCDRsOnExpInitMasterRPC,
		testCDRsOnExpLoadDefaultCharger,
		testCDRsOnExpDisableOnlineExport,
		testCDRsOnExpHttpCdrReplication,
		testCDRsOnExpAMQPReplication,
		testCDRsOnExpFileFailover,
		testCDRsOnExpKafkaPosterFileFailover,
		testCDRsOnExpStopEngine,
	}
)

func TestCDRsOnExp(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		cdrsMasterCfgDIR = "cdrsonexpmaster_mysql"
		cdrsSlaveCfgDIR = "cdrsonexpslave_mysql"
	case utils.MetaMongo:
		cdrsMasterCfgDIR = "cdrsonexpmaster_mongo"
		cdrsSlaveCfgDIR = "cdrsonexpslave_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsCDRsOnExp {
		t.Run(*dbType, stest)
	}
}

func testCDRsOnExpInitConfig(t *testing.T) {
	var err error
	cdrsMasterCfgPath = path.Join(*dataDir, "conf", "samples", cdrsMasterCfgDIR)
	if cdrsMasterCfg, err = config.NewCGRConfigFromPath(cdrsMasterCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
	cdrsSlaveCfgPath = path.Join(*dataDir, "conf", "samples", cdrsSlaveCfgDIR)
	if cdrsSlaveCfg, err = config.NewCGRConfigFromPath(cdrsSlaveCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func testCDRsOnExpInitCdrDb(t *testing.T) {
	if err := engine.InitDataDb(cdrsMasterCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitDataDb(cdrsSlaveCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(cdrsMasterCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(cdrsSlaveCfg); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(cdrsMasterCfg.GeneralCfg().FailedPostsDir); err != nil {
		t.Fatal("Error removing folder: ", cdrsMasterCfg.GeneralCfg().FailedPostsDir, err)
	}

	if err := os.MkdirAll(cdrsMasterCfg.GeneralCfg().FailedPostsDir, 0700); err != nil {
		t.Error(err)
	}

}

func testCDRsOnExpStartMasterEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(cdrsMasterCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testCDRsOnExpStartSlaveEngine(t *testing.T) {
	if _, err := engine.StartEngine(cdrsSlaveCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Create Queues dor amq

func testCDRsOnExpAMQPQueuesCreation(t *testing.T) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		t.Fatal(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer ch.Close()

	if err = ch.ExchangeDeclare("exchangename", "fanout", true, false, false, false, nil); err != nil {
		return
	}
	q1, err := ch.QueueDeclare("queue1", true, false, false, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = ch.QueueBind(q1.Name, "cgr_cdrs", "exchangename", false, nil); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testCDRsOnExpInitMasterRPC(t *testing.T) {
	var err error
	cdrsMasterRpc, err = rpcclient.NewRPCClient(utils.TCP, cdrsMasterCfg.ListenCfg().RPCJSONListen, false, "", "", "", 1, 1,
		time.Duration(1*time.Second), time.Duration(2*time.Second), rpcclient.JSONrpc, nil, false)
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testCDRsOnExpLoadDefaultCharger(t *testing.T) {
	//add a default charger
	chargerProfile := &engine.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "Default",
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{"*none"},
		Weight:       20,
	}
	var result string
	if err := cdrsMasterRpc.Call(utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

// Disable ExportCDR
func testCDRsOnExpDisableOnlineExport(t *testing.T) {
	// stop RabbitMQ server so we can test reconnects
	if err := exec.Command("service", "rabbitmq-server", "stop").Run(); err != nil {
		t.Error(err)
	}
	testCdr := &engine.CDR{
		CGRID:       utils.Sha1("NoOnlineExport", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
		ToR:         utils.VOICE,
		OriginID:    "TestCDRsOnExpDisableOnlineExport",
		OriginHost:  "192.168.1.0",
		Source:      "UNKNOWN",
		RequestType: utils.META_PSEUDOPREPAID,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "1002",
		SetupTime:   time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC),
		AnswerTime:  time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage:       time.Duration(10) * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		RunID:       utils.MetaDefault,
		Cost:        1.201,
		PreRated:    true,
		CostDetails: &engine.EventCost{
			Cost: utils.Float64Pointer(10),
		},
	}
	var reply string
	if err := cdrsMasterRpc.Call(utils.CDRsV1ProcessEvent,
		&engine.ArgV1ProcessEvent{
			Flags:    []string{"*export:false"},
			CGREvent: *testCdr.AsCGREvent(),
		}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	filesInDir, _ := ioutil.ReadDir(cdrsMasterCfg.GeneralCfg().FailedPostsDir)
	if len(filesInDir) != 0 {
		t.Fatalf("Should be no files in directory: %s", cdrsMasterCfg.GeneralCfg().FailedPostsDir)
	}
	// start RabbitMQ server so we can test reconnects
	if err := exec.Command("service", "rabbitmq-server", "start").Run(); err != nil {
		t.Error(err)
	}
	time.Sleep(5 * time.Second)
}

func testCDRsOnExpHttpCdrReplication(t *testing.T) {
	testCdr1 := &engine.CDR{
		CGRID:       httpCGRID,
		ToR:         utils.VOICE,
		OriginID:    "httpjsonrpc1",
		OriginHost:  "192.168.1.1",
		Source:      "UNKNOWN",
		RequestType: utils.META_PSEUDOPREPAID,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "1002",
		SetupTime:   time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC),
		AnswerTime:  time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage:       time.Duration(10) * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		RunID:       utils.MetaDefault,
		Cost:        1.201,
		PreRated:    true,
		CostDetails: &engine.EventCost{
			Cost: utils.Float64Pointer(10),
		},
	}
	var reply string
	if err := cdrsMasterRpc.Call(utils.CDRsV1ProcessEvent,
		&engine.ArgV1ProcessEvent{CGREvent: *testCdr1.AsCGREvent()}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	cdrsSlaveRpc, err := rpcclient.NewRPCClient(utils.TCP, "127.0.0.1:12012", false, "", "", "", 1, 1,
		time.Duration(1*time.Second), time.Duration(2*time.Second), rpcclient.JSONrpc, nil, false)
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
	// ToDo: Fix cdr_http to be compatible with rest of processCdr methods
	time.Sleep(200 * time.Millisecond)
	var rcvedCdrs []*engine.ExternalCDR
	if err := cdrsSlaveRpc.Call(utils.APIerSv2GetCDRs,
		utils.RPCCDRsFilter{CGRIDs: []string{testCdr1.CGRID}, RunIDs: []string{utils.MetaDefault}}, &rcvedCdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(rcvedCdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(rcvedCdrs))
	} else {
		rcvSetupTime, _ := utils.ParseTimeDetectLayout(rcvedCdrs[0].SetupTime, "")
		rcvAnswerTime, _ := utils.ParseTimeDetectLayout(rcvedCdrs[0].AnswerTime, "")
		//rcvUsage, _ := utils.ParseDurationWithSecs(rcvedCdrs[0].Usage)
		if rcvedCdrs[0].CGRID != testCdr1.CGRID ||
			rcvedCdrs[0].ToR != testCdr1.ToR ||
			rcvedCdrs[0].OriginHost != testCdr1.OriginHost ||
			rcvedCdrs[0].RequestType != testCdr1.RequestType ||
			rcvedCdrs[0].Tenant != testCdr1.Tenant ||
			rcvedCdrs[0].Category != testCdr1.Category ||
			rcvedCdrs[0].Account != testCdr1.Account ||
			rcvedCdrs[0].Subject != testCdr1.Subject ||
			rcvedCdrs[0].Destination != testCdr1.Destination ||
			!rcvSetupTime.Equal(testCdr1.SetupTime) ||
			!rcvAnswerTime.Equal(testCdr1.AnswerTime) ||
			//rcvUsage != 10 ||
			rcvedCdrs[0].RunID != testCdr1.RunID {
			//rcvedCdrs[0].Cost != testCdr1.Cost ||
			//!reflect.DeepEqual(rcvedCdrs[0].ExtraFields, testCdr1.ExtraFields) {
			t.Errorf("Expected: %+v, received: %+v", utils.ToJSON(testCdr1), utils.ToJSON(rcvedCdrs[0]))
		}
	}
}

func testCDRsOnExpAMQPReplication(t *testing.T) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		t.Fatal(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("cgrates_cdrs", true, false, false, false, nil)
	if err != nil {
		conn.Close()
		t.Fatal(err)
	}
	q1, err := ch.QueueDeclare("queue1", true, false, false, false, nil)
	if err != nil {
		conn.Close()
		t.Fatal(err)
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		conn.Close()
		t.Fatal(err)
	}
	select {
	case d := <-msgs:
		var rcvCDR map[string]string
		if err := json.Unmarshal(d.Body, &rcvCDR); err != nil {
			t.Error(err)
		}
		if rcvCDR[utils.CGRID] != httpCGRID {
			t.Errorf("Unexpected CDR received: %+v", rcvCDR)
		}
	case <-time.After(time.Duration(100 * time.Millisecond)):
		t.Error("No message received from RabbitMQ")
	}
	if msgs, err = ch.Consume(q1.Name, "consumer", true, false, false, false, nil); err != nil {
		conn.Close()
		t.Fatal(err)
	}
	select {
	case d := <-msgs:
		var rcvCDR map[string]string
		if err := json.Unmarshal(d.Body, &rcvCDR); err != nil {
			t.Error(err)
		}
		if rcvCDR[utils.CGRID] != httpCGRID {
			t.Errorf("Unexpected CDR received: %+v", rcvCDR)
		}
	case <-time.After(time.Duration(100 * time.Millisecond)):
		t.Error("No message received from RabbitMQ")
	}
	conn.Close()
	time.Sleep(500 * time.Millisecond)
	// restart RabbitMQ server so we can test reconnects
	if err := exec.Command("service", "rabbitmq-server", "restart").Run(); err != nil {
		t.Error(err)
	}
	time.Sleep(5 * time.Second)
	testCdr := &engine.CDR{
		CGRID:       amqpCGRID,
		ToR:         utils.VOICE,
		OriginID:    "amqpreconnect",
		OriginHost:  "192.168.1.1",
		Source:      "UNKNOWN",
		RequestType: utils.META_PSEUDOPREPAID,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "1002",
		SetupTime:   time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC),
		AnswerTime:  time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage:       time.Duration(10) * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		RunID:       utils.MetaDefault,
		Cost:        1.201,
		PreRated:    true,
		CostDetails: &engine.EventCost{
			Cost: utils.Float64Pointer(10),
		},
	}
	var reply string
	if err := cdrsMasterRpc.Call(utils.CDRsV1ProcessEvent,
		&engine.ArgV1ProcessEvent{
			Flags:    []string{"*export:true"},
			CGREvent: *testCdr.AsCGREvent(),
		}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	if conn, err = amqp.Dial("amqp://guest:guest@localhost:5672/"); err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if ch, err = conn.Channel(); err != nil {
		t.Fatal(err)
	}
	defer ch.Close()

	if msgs, err = ch.Consume(q.Name, "", true, false, false, false, nil); err != nil {
		t.Fatal(err)
	}
	select {
	case d := <-msgs:
		var rcvCDR map[string]string
		if err := json.Unmarshal(d.Body, &rcvCDR); err != nil {
			t.Error(err)
		}
		if rcvCDR[utils.CGRID] != testCdr.CGRID {
			t.Errorf("Unexpected CDR received: %+v", rcvCDR)
		}
	case <-time.After(150 * time.Millisecond):
		t.Error("No message received from RabbitMQ")
	}

	if msgs, err = ch.Consume(q1.Name, "", true, false, false, false, nil); err != nil {
		t.Fatal(err)
	}
	select {
	case d := <-msgs:
		var rcvCDR map[string]string
		if err := json.Unmarshal(d.Body, &rcvCDR); err != nil {
			t.Error(err)
		}
		if rcvCDR[utils.CGRID] != testCdr.CGRID {
			t.Errorf("Unexpected CDR received: %s expeced: %s", utils.ToJSON(rcvCDR), utils.ToJSON(testCdr))
		}
	case <-time.After(150 * time.Millisecond):
		t.Error("No message received from RabbitMQ")
	}

}

func checkContent(ev *engine.ExportEvents, content []interface{}) error {
	match := false
	for _, bev := range ev.Events {
		for _, con := range content {
			if reflect.DeepEqual(bev, con) {
				match = true
				break
			}
		}
		if match {
			break
		}
	}
	if !match {
		exp := make([]string, len(content))
		for i, con := range content {
			exp[i] = utils.IfaceAsString(con)
		}
		recv := make([]string, len(ev.Events))
		for i, con := range ev.Events {
			recv[i] = utils.IfaceAsString(con)
		}
		return fmt.Errorf("Expecting: one of %q, received: %q", utils.ToJSON(exp), utils.ToJSON(recv))
	}
	return nil
}

func testCDRsOnExpFileFailover(t *testing.T) {
	time.Sleep(5 * time.Second)
	v1 := url.Values{}
	v2 := url.Values{}
	v1.Set("OriginID", "httpjsonrpc1")
	v2.Set("OriginID", "amqpreconnect")
	httpContent := []interface{}{v1, v2}
	filesInDir, _ := ioutil.ReadDir(cdrsMasterCfg.GeneralCfg().FailedPostsDir)
	if len(filesInDir) == 0 {
		t.Fatalf("No files in directory: %s", cdrsMasterCfg.GeneralCfg().FailedPostsDir)
	}
	expectedFormats := utils.NewStringSet([]string{utils.MetaHTTPPost, utils.MetaAMQPjsonMap,
		utils.MetaAMQPV1jsonMap, utils.MetaSQSjsonMap, utils.MetaS3jsonMap})
	rcvFormats := utils.NewStringSet([]string{})
	for _, file := range filesInDir { // First file in directory is the one we need, harder to find it's name out of config
		fileName := file.Name()
		filePath := path.Join(cdrsMasterCfg.GeneralCfg().FailedPostsDir, fileName)

		ev, err := engine.NewExportEventsFromFile(filePath)
		if err != nil {
			t.Errorf("<%s> for file <%s>", err, fileName)
			continue
		} else if len(ev.Events) == 0 {
			t.Error("Expected at least one event")
			continue
		}
		rcvFormats.Add(ev.Format)
		content := failoverContent
		if ev.Format == utils.MetaHTTPPost {
			content = httpContent
		}
		if err := checkContent(ev, content); err != nil {
			t.Errorf("For file <%s> and event <%s> received %s", filePath, utils.ToJSON(ev), err)
		}
	}
	if !reflect.DeepEqual(expectedFormats, rcvFormats) {
		t.Errorf("Missing format expecting: %s received: %s", utils.ToJSON(expectedFormats), utils.ToJSON(rcvFormats))
	}
}

func testCDRsOnExpKafkaPosterFileFailover(t *testing.T) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "cgrates_cdrs",
		GroupID: "tmp",
		MaxWait: time.Millisecond,
	})

	defer reader.Close()

	for i := 0; i < 2; i++ { // no raw CDR
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if m, err := reader.ReadMessage(ctx); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(failoverContent[0].([]byte), m.Value) && !reflect.DeepEqual(failoverContent[1].([]byte), m.Value) { // Checking just the prefix should do since some content is dynamic
			t.Errorf("Expecting: %v or %v, received: %v", utils.IfaceAsString(failoverContent[0]), utils.IfaceAsString(failoverContent[1]), string(m.Value))
		}
		cancel()
	}
}

/*
// Performance test, check `lsof -a -p 8427 | wc -l`

func testCdrsHttpCdrReplication2(t *testing.T) {
	cdrs := make([]*engine.CDR, 0)
	for i := 0; i < 10000; i++ {
		cdr := &engine.CDR{OriginID: fmt.Sprintf("httpjsonrpc_%d", i),
			ToR: utils.VOICE, OriginHost: "192.168.1.1", Source: "UNKNOWN", RequestType: utils.META_PSEUDOPREPAID,
			Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
			SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
			Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}}
		cdrs = append(cdrs, cdr)
	}
	var reply string
	for _, cdr := range cdrs {
		if err := cdrsMasterRpc.Call(utils.CdrsV2ProcessCdr, cdr, &reply); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if reply != utils.OK {
			t.Error("Unexpected reply received: ", reply)
		}
	}
}
*/

func testCDRsOnExpStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
