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
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strings"
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
	cdrsMasterCfg, cdrsSlaveCfg         *config.CGRConfig
	cdrsMasterRpc                       *rpcclient.RpcClient

	sTestsCDRsOnExp = []func(t *testing.T){
		testCDRsOnExpInitConfig,
		testCDRsOnExpInitCdrDb,
		testCDRsOnExpStartMasterEngine,
		testCDRsOnExpStartSlaveEngine,
		testCDRsOnExpAMQPQueuesCreation,
		testCDRsOnExpInitMasterRPC,
		testCDRsOnExpDisableOnlineExport,
		testCDRsOnExpHttpCdrReplication,
		testCDRsOnExpAMQPReplication,
		testCDRsOnExpHTTPPosterFileFailover,
		testCDRsOnExpAMQPPosterFileFailover,
		testCDRsOnExpAWSAMQPPosterFileFailover,
		testCDRsOnExpKafkaPosterFileFailover,
		testCDRsOnExpSQSPosterFileFailover,
		testCDRsOnExpS3PosterFileFailover,
		testCDRsOnExpStopEngine,
	}
)

func TestCDRsOnExp(t *testing.T) {
	for _, stest := range sTestsCDRsOnExp {
		t.Run("TestCDRsOnExp", stest)
	}
}

func testCDRsOnExpInitConfig(t *testing.T) {
	var err error
	cdrsMasterCfgPath = path.Join(*dataDir, "conf", "samples", "cdrsonexpmaster")
	if cdrsMasterCfg, err = config.NewCGRConfigFromPath(cdrsMasterCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
	cdrsSlaveCfgPath = path.Join(*dataDir, "conf", "samples", "cdrsonexpslave")
	if cdrsSlaveCfg, err = config.NewCGRConfigFromPath(cdrsSlaveCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func testCDRsOnExpInitCdrDb(t *testing.T) {
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
	cdrsMasterRpc, err = rpcclient.NewRpcClient("tcp", cdrsMasterCfg.ListenCfg().RPCJSONListen, false, "", "", "", 1, 1,
		time.Duration(1*time.Second), time.Duration(2*time.Second), "json", nil, false)
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
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
		RunID:       utils.DEFAULT_RUNID,
		Cost:        1.201,
		PreRated:    true,
	}
	var reply string
	if err := cdrsMasterRpc.Call(utils.CDRsV1ProcessEvent,
		&engine.ArgV1ProcessEvent{
			CGREvent: *testCdr.AsCGREvent(),
			Export:   utils.BoolPointer(false),
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
	//add a default charger
	chargerProfile := &engine.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "Default",
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{"*none"},
		Weight:       20,
	}
	var result string
	if err := cdrsMasterRpc.Call(utils.ApierV1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	testCdr1 := &engine.CDR{
		CGRID:       utils.Sha1("httpjsonrpc1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
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
		RunID:       utils.DEFAULT_RUNID,
		Cost:        1.201,
		PreRated:    true,
	}
	var reply string
	if err := cdrsMasterRpc.Call(utils.CDRsV1ProcessEvent,
		&engine.ArgV1ProcessEvent{CGREvent: *testCdr1.AsCGREvent()}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	cdrsSlaveRpc, err := rpcclient.NewRpcClient("tcp", "127.0.0.1:12012", false, "", "", "", 1, 1,
		time.Duration(1*time.Second), time.Duration(2*time.Second), "json", nil, false)
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
	// ToDo: Fix cdr_http to be compatible with rest of processCdr methods
	time.Sleep(200 * time.Millisecond)
	var rcvedCdrs []*engine.ExternalCDR
	if err := cdrsSlaveRpc.Call(utils.ApierV2GetCDRs,
		utils.RPCCDRsFilter{CGRIDs: []string{testCdr1.CGRID}, RunIDs: []string{utils.META_DEFAULT}}, &rcvedCdrs); err != nil {
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
			t.Errorf("Expected: %+v, received: %+v", testCdr1, rcvedCdrs[0])
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
		if rcvCDR[utils.CGRID] != utils.Sha1("httpjsonrpc1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()) {
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
		if rcvCDR[utils.CGRID] != utils.Sha1("httpjsonrpc1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()) {
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
	time.Sleep(time.Duration(5 * time.Second))
	testCdr := &engine.CDR{
		CGRID:       utils.Sha1("amqpreconnect", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
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
		RunID:       utils.DEFAULT_RUNID,
		Cost:        1.201,
		PreRated:    true,
	}
	var reply string
	if err := cdrsMasterRpc.Call(utils.CDRsV1ProcessEvent,
		&engine.ArgV1ProcessEvent{
			CGREvent: *testCdr.AsCGREvent(),
			Export:   utils.BoolPointer(true),
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

func testCDRsOnExpHTTPPosterFileFailover(t *testing.T) {
	time.Sleep(time.Duration(5 * time.Second))
	failoverContent := [][]byte{[]byte(`OriginID=httpjsonrpc1`), []byte(`OriginID=amqpreconnect`)}
	filesInDir, _ := ioutil.ReadDir(cdrsMasterCfg.GeneralCfg().FailedPostsDir)
	if len(filesInDir) == 0 {
		t.Fatalf("No files in directory: %s", cdrsMasterCfg.GeneralCfg().FailedPostsDir)
	}
	var foundFile bool
	var fileName string
	for _, file := range filesInDir { // First file in directory is the one we need, harder to find it's name out of config
		fileName = file.Name()
		if strings.Index(fileName, utils.FormSuffix) != -1 {
			foundFile = true
			filePath := path.Join(cdrsMasterCfg.GeneralCfg().FailedPostsDir, fileName)
			if readBytes, err := ioutil.ReadFile(filePath); err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(failoverContent[0], readBytes) && !reflect.DeepEqual(failoverContent[1], readBytes) { // Checking just the prefix should do since some content is dynamic
				t.Errorf("Expecting: %q or %q, received: %q", string(failoverContent[0]), string(failoverContent[1]), string(readBytes))
			}
			if err := os.Remove(filePath); err != nil {
				t.Error("Failed removing file: ", filePath)
			}
		}
	}
	if !foundFile {
		t.Fatal("Could not find the file in folder")
	}
}

func testCDRsOnExpAMQPPosterFileFailover(t *testing.T) {
	time.Sleep(time.Duration(5 * time.Second))
	failoverContent := [][]byte{[]byte(`{"CGRID":"57548d485d61ebcba55afbe5d939c82a8e9ff670"}`), []byte(`{"CGRID":"88ed9c38005f07576a1e1af293063833b60edcc6"}`)}
	filesInDir, _ := ioutil.ReadDir(cdrsMasterCfg.GeneralCfg().FailedPostsDir)
	if len(filesInDir) == 0 {
		t.Fatalf("No files in directory: %s", cdrsMasterCfg.GeneralCfg().FailedPostsDir)
	}
	var foundFile bool
	var fileName string
	for _, file := range filesInDir { // First file in directory is the one we need, harder to find it's name out of config
		fileName = file.Name()
		if strings.HasPrefix(fileName, "cdr|*amqp_json_map") {
			foundFile = true
			filePath := path.Join(cdrsMasterCfg.GeneralCfg().FailedPostsDir, fileName)
			if readBytes, err := ioutil.ReadFile(filePath); err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(failoverContent[0], readBytes) && !reflect.DeepEqual(failoverContent[1], readBytes) { // Checking just the prefix should do since some content is dynamic
				t.Errorf("Expecting: %v or %v, received: %v", string(failoverContent[0]), string(failoverContent[1]), string(readBytes))
			}
			if err := os.Remove(filePath); err != nil {
				t.Error("Failed removing file: ", filePath)
			}
		}
	}
	if !foundFile {
		t.Fatal("Could not find the file in folder")
	}
}

func testCDRsOnExpAWSAMQPPosterFileFailover(t *testing.T) {
	time.Sleep(time.Duration(10 * time.Second))
	failoverContent := [][]byte{[]byte(`{"CGRID":"57548d485d61ebcba55afbe5d939c82a8e9ff670"}`), []byte(`{"CGRID":"88ed9c38005f07576a1e1af293063833b60edcc6"}`)}
	filesInDir, _ := ioutil.ReadDir(cdrsMasterCfg.GeneralCfg().FailedPostsDir)
	if len(filesInDir) == 0 {
		t.Fatalf("No files in directory: %s", cdrsMasterCfg.GeneralCfg().FailedPostsDir)
	}
	var foundFile bool
	var fileName string
	for _, file := range filesInDir { // First file in directory is the one we need, harder to find it's name out of config
		fileName = file.Name()
		if strings.HasPrefix(fileName, "cdr|*amqpv1_json_map") {
			foundFile = true
			filePath := path.Join(cdrsMasterCfg.GeneralCfg().FailedPostsDir, fileName)
			if readBytes, err := ioutil.ReadFile(filePath); err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(failoverContent[0], readBytes) && !reflect.DeepEqual(failoverContent[1], readBytes) { // Checking just the prefix should do since some content is dynamic
				t.Errorf("Expecting: %v or %v, received: %v", string(failoverContent[0]), string(failoverContent[1]), string(readBytes))
			}
			if err := os.Remove(filePath); err != nil {
				t.Error("Failed removing file: ", filePath)
			}
		}
	}
	if !foundFile {
		t.Fatal("Could not find the file in folder")
	}
}

func testCDRsOnExpKafkaPosterFileFailover(t *testing.T) {
	failoverContent := [][]byte{[]byte(`{"CGRID":"57548d485d61ebcba55afbe5d939c82a8e9ff670"}`), []byte(`{"CGRID":"88ed9c38005f07576a1e1af293063833b60edcc6"}`)}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "cgrates_cdrs",
		GroupID: "tmp",
		MaxWait: time.Millisecond,
	})

	defer reader.Close()

	for i := 0; i < 4; i++ {
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		if m, err := reader.ReadMessage(ctx); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(failoverContent[0], m.Value) && !reflect.DeepEqual(failoverContent[1], m.Value) { // Checking just the prefix should do since some content is dynamic
			t.Errorf("Expecting: %v or %v, received: %v", string(failoverContent[0]), string(failoverContent[1]), string(m.Value))
		}
	}
}

func testCDRsOnExpSQSPosterFileFailover(t *testing.T) {
	time.Sleep(time.Duration(10 * time.Second))
	failoverContent := [][]byte{[]byte(`{"CGRID":"57548d485d61ebcba55afbe5d939c82a8e9ff670"}`), []byte(`{"CGRID":"88ed9c38005f07576a1e1af293063833b60edcc6"}`)}
	filesInDir, _ := ioutil.ReadDir(cdrsMasterCfg.GeneralCfg().FailedPostsDir)
	if len(filesInDir) == 0 {
		t.Fatalf("No files in directory: %s", cdrsMasterCfg.GeneralCfg().FailedPostsDir)
	}
	var foundFile bool
	var fileName string
	for _, file := range filesInDir { // First file in directory is the one we need, harder to find it's name out of config
		fileName = file.Name()
		if strings.HasPrefix(fileName, "cdr|*sqs_json_map") {
			foundFile = true
			filePath := path.Join(cdrsMasterCfg.GeneralCfg().FailedPostsDir, fileName)
			if readBytes, err := ioutil.ReadFile(filePath); err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(failoverContent[0], readBytes) && !reflect.DeepEqual(failoverContent[1], readBytes) { // Checking just the prefix should do since some content is dynamic
				t.Errorf("Expecting: %v or %v, received: %v", string(failoverContent[0]), string(failoverContent[1]), string(readBytes))
			}
			if err := os.Remove(filePath); err != nil {
				t.Error("Failed removing file: ", filePath)
			}
		}
	}
	if !foundFile {
		t.Fatal("Could not find the file in folder")
	}
}

func testCDRsOnExpS3PosterFileFailover(t *testing.T) {
	time.Sleep(time.Duration(10 * time.Second))
	failoverContent := [][]byte{[]byte(`{"CGRID":"57548d485d61ebcba55afbe5d939c82a8e9ff670"}`), []byte(`{"CGRID":"88ed9c38005f07576a1e1af293063833b60edcc6"}`)}
	filesInDir, _ := ioutil.ReadDir(cdrsMasterCfg.GeneralCfg().FailedPostsDir)
	if len(filesInDir) == 0 {
		t.Fatalf("No files in directory: %s", cdrsMasterCfg.GeneralCfg().FailedPostsDir)
	}
	var foundFile bool
	var fileName string
	for _, file := range filesInDir { // First file in directory is the one we need, harder to find it's name out of config
		fileName = file.Name()
		if strings.HasPrefix(fileName, "cdr|*s3_json_map") {
			foundFile = true
			filePath := path.Join(cdrsMasterCfg.GeneralCfg().FailedPostsDir, fileName)
			if readBytes, err := ioutil.ReadFile(filePath); err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(failoverContent[0], readBytes) && !reflect.DeepEqual(failoverContent[1], readBytes) { // Checking just the prefix should do since some content is dynamic
				t.Errorf("Expecting: %v or %v, received: %v", string(failoverContent[0]), string(failoverContent[1]), string(readBytes))
			}
			if err := os.Remove(filePath); err != nil {
				t.Error("Failed removing file: ", filePath)
			}
		}
	}
	if !foundFile {
		t.Fatal("Could not find the file in folder")
	}
}

/*
// Performance test, check `lsof -a -p 8427 | wc -l`

func testCdrsHttpCdrReplication2(t *testing.T) {
	cdrs := make([]*engine.CDR, 0)
	for i := 0; i < 10000; i++ {
		cdr := &engine.CDR{OriginID: fmt.Sprintf("httpjsonrpc_%d", i),
			ToR: utils.VOICE, OriginHost: "192.168.1.1", Source: "UNKNOWN", RequestType: utils.META_PSEUDOPREPAID,
			Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
			SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
			Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}}
		cdrs = append(cdrs, cdr)
	}
	var reply string
	for _, cdr := range cdrs {
		if err := cdrsMasterRpc.Call("CdrsV2.ProcessCdr", cdr, &reply); err != nil {
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
