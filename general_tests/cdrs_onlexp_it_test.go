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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/ees"
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
	failoverContent                     = [][]byte{[]byte(fmt.Sprintf(`{"CGRID":"%s"}`, httpCGRID)), []byte(fmt.Sprintf(`{"CGRID":"%s"}`, amqpCGRID))}

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
	case utils.MetaInternal, utils.MetaPostgres:
		t.SkipNow()
	case utils.MetaMySQL:
		cdrsMasterCfgDIR = "cdrsonexpmaster_mysql"
		cdrsSlaveCfgDIR = "cdrsonexpslave_mysql"
	case utils.MetaMongo:
		cdrsMasterCfgDIR = "cdrsonexpmaster_mongo"
		cdrsSlaveCfgDIR = "cdrsonexpslave_mongo"
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
	if cdrsMasterCfg, err = config.NewCGRConfigFromPath(context.Background(), cdrsMasterCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
	cdrsSlaveCfgPath = path.Join(*dataDir, "conf", "samples", cdrsSlaveCfgDIR)
	if cdrsSlaveCfg, err = config.NewCGRConfigFromPath(context.Background(), cdrsSlaveCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func testCDRsOnExpInitCdrDb(t *testing.T) {
	if err := engine.InitDataDB(cdrsMasterCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitDataDB(cdrsSlaveCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(cdrsMasterCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(cdrsSlaveCfg); err != nil {
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
	if err = ch.Close(); err != nil {
		t.Error(err)
	}
	if err = conn.Close(); err != nil {
		t.Error(err)
	}
	v, err := kafka.Dial("tcp", "localhost:9092")
	if err != nil {
		t.Fatal(err)
	}
	if err := v.CreateTopics(kafka.TopicConfig{
		Topic:             "cgrates_cdrs",
		NumPartitions:     1,
		ReplicationFactor: 1,
	}); err != nil {
		t.Fatal(err)
	}
	if err = v.Close(); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testCDRsOnExpInitMasterRPC(t *testing.T) {
	var err error
	cdrsMasterRpc, err = rpcclient.NewRPCClient(context.Background(), utils.TCP, cdrsMasterCfg.ListenCfg().RPCJSONListen, false, "", "", "", 1, 1,
		time.Second, 5*time.Second, rpcclient.JSONrpc, nil, false, nil)
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testCDRsOnExpLoadDefaultCharger(t *testing.T) {
	// //add a default charger
	chargerProfile := &engine.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "Default",
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{"*none"},
		Weight:       20,
	}
	var result string
	if err := cdrsMasterRpc.Call(context.Background(), utils.AdminSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

// Disable ExportCDR
func testCDRsOnExpDisableOnlineExport(t *testing.T) {
	testCdr := &engine.CDR{
		CGRID:       utils.Sha1("NoOnlineExport", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
		ToR:         utils.MetaVoice,
		OriginID:    "TestCDRsOnExpDisableOnlineExport",
		OriginHost:  "192.168.1.0",
		Source:      "UNKNOWN",
		RequestType: utils.MetaPseudoPrepaid,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "1002",
		SetupTime:   time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC),
		AnswerTime:  time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage:       10 * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		RunID:       utils.MetaDefault,
		Cost:        1.201,
		PreRated:    true,
		// CostDetails: &engine.EventCost{
		// 	Cost: utils.Float64Pointer(10),
		// },
	}
	testEv := testCdr.AsCGREvent()
	testEv.APIOpts[utils.OptsCDRsExport] = false
	testEv.APIOpts[utils.OptsCDRsChargerS] = false
	var reply string
	if err := cdrsMasterRpc.Call(context.Background(), utils.CDRsV1ProcessEvent,
		testEv, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	filesInDir, _ := os.ReadDir(cdrsMasterCfg.GeneralCfg().FailedPostsDir)
	if len(filesInDir) != 0 {
		t.Fatalf("Should be no files in directory: %s", cdrsMasterCfg.GeneralCfg().FailedPostsDir)
	}
}

func testCDRsOnExpHttpCdrReplication(t *testing.T) {
	testCdr1 := &engine.CDR{
		CGRID:       httpCGRID,
		ToR:         utils.MetaVoice,
		OriginID:    "httpjsonrpc1",
		OriginHost:  "192.168.1.1",
		Source:      "UNKNOWN",
		RequestType: utils.MetaNone,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "1002",
		SetupTime:   time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC),
		AnswerTime:  time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage:       10 * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		RunID:       utils.MetaDefault,
		Cost:        1.201,
		PreRated:    true,
		// CostDetails: &engine.EventCost{
		// 	Cost: utils.Float64Pointer(10),
		// },
	}
	var reply string
	arg := testCdr1.AsCGREvent()
	arg.APIOpts = map[string]interface{}{"ExporterID": "http_localhost"}

	// we expect that the cdr export to fail and go into the failed post directory
	if err := cdrsMasterRpc.Call(context.Background(), utils.CDRsV1ProcessEvent,
		arg, &reply); err == nil || err.Error() != utils.ErrPartiallyExecuted.Error() {
		t.Error("Unexpected error: ", err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	cdrsSlaveRpc, err := rpcclient.NewRPCClient(context.Background(), utils.TCP, "127.0.0.1:12012", false, "", "", "", 1, 1,
		time.Second, 2*time.Second, rpcclient.JSONrpc, nil, false, nil)
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
	// ToDo: Fix cdr_http to be compatible with rest of processCdr methods
	var rcvedCdrs []*engine.ExternalCDR
	if err := cdrsSlaveRpc.Call(utils.APIerSv2GetCDRs,
		&utils.RPCCDRsFilter{CGRIDs: []string{testCdr1.CGRID}, RunIDs: []string{utils.MetaDefault}}, &rcvedCdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(rcvedCdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(rcvedCdrs))
	} else {
		rcvSetupTime, _ := utils.ParseTimeDetectLayout(rcvedCdrs[0].SetupTime, "")
		rcvAnswerTime, _ := utils.ParseTimeDetectLayout(rcvedCdrs[0].AnswerTime, "")
		if rcvedCdrs[0].CGRID != testCdr1.CGRID ||
			rcvedCdrs[0].RunID != testCdr1.RunID ||
			rcvedCdrs[0].ToR != testCdr1.ToR ||
			rcvedCdrs[0].OriginID != testCdr1.OriginID ||
			rcvedCdrs[0].RequestType != testCdr1.RequestType ||
			rcvedCdrs[0].Tenant != testCdr1.Tenant ||
			rcvedCdrs[0].Category != testCdr1.Category ||
			rcvedCdrs[0].Account != testCdr1.Account ||
			rcvedCdrs[0].Subject != testCdr1.Subject ||
			rcvedCdrs[0].Destination != testCdr1.Destination ||
			!rcvSetupTime.Equal(testCdr1.SetupTime) ||
			!rcvAnswerTime.Equal(testCdr1.AnswerTime) ||
			rcvedCdrs[0].Usage != testCdr1.Usage.String() ||
			rcvedCdrs[0].Cost != testCdr1.Cost {
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

	msgs, err := ch.Consume("cgrates_cdrs", "", true, false, false, false, nil)
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
	case <-time.After(100 * time.Millisecond):
		t.Error("No message received from RabbitMQ")
	}
	if msgs, err = ch.Consume("queue1", "consumer", true, false, false, false, nil); err != nil {
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
	case <-time.After(100 * time.Millisecond):
		t.Error("No message received from RabbitMQ")
	}
	conn.Close()
	// restart RabbitMQ server so we can test reconnects
	if err := exec.Command("service", "rabbitmq-server", "restart").Run(); err != nil {
		t.Error(err)
	}
	time.Sleep(2 * time.Second)
	testCdr := &engine.CDR{
		CGRID:       amqpCGRID,
		ToR:         utils.MetaVoice,
		OriginID:    "amqpreconnect",
		OriginHost:  "192.168.1.1",
		Source:      "UNKNOWN",
		RequestType: utils.MetaPseudoPrepaid,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "1002",
		SetupTime:   time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC),
		AnswerTime:  time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage:       10 * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		RunID:       utils.MetaDefault,
		Cost:        1.201,
		PreRated:    true,
		// CostDetails: &engine.EventCost{
		// 	Cost: utils.Float64Pointer(10),
		// },
	}
	testEv := testCdr.AsCGREvent()
	testEv.APIOpts[utils.OptsCDRsExport] = true
	var reply string
	if err := cdrsMasterRpc.Call(context.Background(), utils.CDRsV1ProcessEvent,
		testEv, &reply); err == nil || err.Error() != utils.ErrPartiallyExecuted.Error() {
		t.Error("Unexpected error: ", err)
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

	if msgs, err = ch.Consume("cgrates_cdrs", "", true, false, false, false, nil); err != nil {
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

	if msgs, err = ch.Consume("queue1", "", true, false, false, false, nil); err != nil {
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

func testCDRsOnExpFileFailover(t *testing.T) {
	v1 := url.Values{}
	v2 := url.Values{}
	v1.Set("OriginID", "httpjsonrpc1")
	v2.Set("OriginID", "amqpreconnect")
	httpContent := []interface{}{&ees.HTTPPosterRequest{Body: v1, Header: http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}}},
		&ees.HTTPPosterRequest{Body: v2, Header: http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}}}}
	filesInDir, _ := os.ReadDir(cdrsMasterCfg.GeneralCfg().FailedPostsDir)
	if len(filesInDir) == 0 {
		t.Fatalf("No files in directory: %s", cdrsMasterCfg.GeneralCfg().FailedPostsDir)
	}
	for _, file := range filesInDir { // First file in directory is the one we need, harder to find it's name out of config
		fileName := file.Name()
		filePath := path.Join(cdrsMasterCfg.GeneralCfg().FailedPostsDir, fileName)

		ev, err := ees.NewExportEventsFromFile(filePath)
		if err != nil {
			t.Errorf("<%s> for file <%s>", err, fileName)
			continue
		} else if len(ev.Events) == 0 {
			t.Error("Expected at least one event")
			continue
		}
		if ev.Format != utils.MetaHTTPPost {
			t.Errorf("Expected %s to be only failed exporter,received <%s>", utils.MetaHTTPPost, ev.Format)
		}
		if err := checkContent(ev, httpContent); err != nil {
			t.Errorf("For file <%s> and event <%s> received %s", filePath, utils.ToJSON(ev), err)
		}
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
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		if m, err := reader.ReadMessage(ctx); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(failoverContent[0], m.Value) && !reflect.DeepEqual(failoverContent[1], m.Value) { // Checking just the prefix should do since some content is dynamic
			t.Errorf("Expecting: %v or %v, received: %v", string(failoverContent[0]), string(failoverContent[1]), string(m.Value))
		}
		cancel()
	}
}

func testCDRsOnExpStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		t.Fatal(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer ch.Close()

	if _, err = ch.QueueDelete("cgrates_cdrs", false, false, true); err != nil {
		t.Fatal(err)
	}

	if _, err = ch.QueueDelete("queue1", false, false, true); err != nil {
		t.Fatal(err)
	}
}
