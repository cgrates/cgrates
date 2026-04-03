//go:build flaky

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

package general_tests

import (
	"bytes"
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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	cdrsMasterCfgPath, cdrsSlaveCfgPath string
	cdrsMasterCfgDIR, cdrsSlaveCfgDIR   string
	cdrsMasterCfg, cdrsSlaveCfg         *config.CGRConfig
	cdrsMasterRpc                       *birpc.Client
	cdrsSlaveRpc                        *birpc.Client
	httpCGRID                           = utils.UUIDSha1Prefix()
	amqpCGRID                           = utils.UUIDSha1Prefix()
	failoverContent                     = [][]byte{[]byte(fmt.Sprintf(`{"CGRID":"%s"}`, httpCGRID)), []byte(fmt.Sprintf(`{"CGRID":"%s"}`, amqpCGRID))}

	sTestsCDRsOnExp = []func(t *testing.T){
		testCDRsOnExpAMQPQueuesCreation,
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
	switch *utils.DBType {
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

	if out, err := exec.Command("pgrep", "-a", "cgr-engine").Output(); err == nil {
		t.Fatalf("stale cgr-engine process from a previous test: %s", bytes.TrimSpace(out))
	}

	cdrsMasterCfgPath = path.Join(*utils.DataDir, "conf", "samples", cdrsMasterCfgDIR)
	cdrsSlaveCfgPath = path.Join(*utils.DataDir, "conf", "samples", cdrsSlaveCfgDIR)

	masterNG := engine.TestEngine{
		ConfigPath: cdrsMasterCfgPath,
		PreStartHook: func(tb testing.TB, cfg *config.CGRConfig) {
			if err := os.RemoveAll(cfg.EEsCfg().FailedPosts.Dir); err != nil {
				tb.Fatal("error removing folder: ", cfg.EEsCfg().FailedPosts.Dir, err)
			}
			if err := os.MkdirAll(cfg.EEsCfg().FailedPosts.Dir, 0700); err != nil {
				tb.Fatal(err)
			}
		},
	}
	cdrsMasterRpc, cdrsMasterCfg = masterNG.Run(t)

	slaveNG := engine.TestEngine{
		ConfigPath:     cdrsSlaveCfgPath,
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	cdrsSlaveRpc, cdrsSlaveCfg = slaveNG.Run(t)

	for _, stest := range sTestsCDRsOnExp {
		t.Run(*utils.DBType, stest)
	}
}

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
	kfkCl, err := kgo.NewClient(kgo.SeedBrokers("localhost:9092"))
	if err != nil {
		t.Fatal(err)
	}
	adm := kadm.NewClient(kfkCl)
	if _, err := adm.CreateTopics(context.Background(), 1, 1, nil, "cgrates_cdrs"); err != nil {
		t.Fatal(err)
	}
	kfkCl.Close()
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
	if err := cdrsMasterRpc.Call(context.Background(), utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
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
		CostDetails: &engine.EventCost{
			Cost: utils.Float64Pointer(10),
		},
	}
	var reply string
	if err := cdrsMasterRpc.Call(context.Background(),
		utils.CDRsV1ProcessEvent,
		&engine.ArgV1ProcessEvent{
			Flags:    []string{"*export:false", "*chargers:false"},
			CGREvent: *testCdr.AsCGREvent(),
		}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond)
	filesInDir, _ := os.ReadDir(cdrsMasterCfg.EEsCfg().FailedPosts.Dir)
	if len(filesInDir) != 0 {
		t.Fatalf("Should be no files in directory: %s", cdrsMasterCfg.EEsCfg().FailedPosts.Dir)
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
		CostDetails: &engine.EventCost{
			Cost: utils.Float64Pointer(10),
		},
	}
	var reply string
	arg := testCdr1.AsCGREvent()
	arg.APIOpts = map[string]any{"ExporterID": "http_localhost"}

	// we expect that the cdr export to fail and go into the failed post directory
	if err := cdrsMasterRpc.Call(context.Background(),
		utils.CDRsV1ProcessEvent,
		&engine.ArgV1ProcessEvent{
			CGREvent: *arg,
		}, &reply); err == nil || err.Error() != utils.ErrPartiallyExecuted.Error() {
		t.Error("Unexpected error: ", err)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond)
	var rcvedCdrs []*engine.ExternalCDR
	if err := cdrsSlaveRpc.Call(context.Background(), utils.APIerSv2GetCDRs,
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
		CostDetails: &engine.EventCost{
			Cost: utils.Float64Pointer(10),
		},
	}
	var reply string
	if err := cdrsMasterRpc.Call(context.Background(),
		utils.CDRsV1ProcessEvent,
		&engine.ArgV1ProcessEvent{
			Flags: []string{"*export:true"},

			CGREvent: *testCdr.AsCGREvent(),
		}, &reply); err == nil || err.Error() != utils.ErrPartiallyExecuted.Error() {
		t.Error("Unexpected error: ", err)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond)
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
	httpContent := []any{&ees.HTTPPosterRequest{Body: v1, Header: http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}}},
		&ees.HTTPPosterRequest{Body: v2, Header: http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}}}}
	filesInDir, _ := os.ReadDir(cdrsMasterCfg.EEsCfg().FailedPosts.Dir)
	if len(filesInDir) == 0 {
		t.Fatalf("No files in directory: %s", cdrsMasterCfg.EEsCfg().FailedPosts.Dir)
	}
	for _, file := range filesInDir { // First file in directory is the one we need, harder to find it's name out of config
		fileName := file.Name()
		filePath := path.Join(cdrsMasterCfg.EEsCfg().FailedPosts.Dir, fileName)

		ev, err := ees.NewExportEventsFromFile(filePath)
		if err != nil {
			t.Errorf("<%s> for file <%s>", err, fileName)
			continue
		} else if len(ev.Events) == 0 {
			t.Error("Expected at least one event")
			continue
		}
		if ev.Type != utils.MetaHTTPPost {
			t.Errorf("Expected %s to be only failed exporter,received <%s>", utils.MetaHTTPPost, ev.Type)
		}
		if err := checkContent(ev, httpContent); err != nil {
			t.Errorf("For file <%s> and event <%s> received %s", filePath, utils.ToJSON(ev), err)
		}
	}
}

func testCDRsOnExpKafkaPosterFileFailover(t *testing.T) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers("localhost:9092"),
		kgo.ConsumeTopics("cgrates_cdrs"),
		kgo.ConsumerGroup("tmp"),
		kgo.FetchMaxWait(10*time.Millisecond),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer cl.Close()

	received := 0
	for received < 2 { // no raw CDR
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		fetches := cl.PollFetches(ctx)
		cancel()
		if errs := fetches.Errors(); len(errs) > 0 {
			t.Fatal(errs[0].Err)
		}
		fetches.EachRecord(func(r *kgo.Record) {
			if !reflect.DeepEqual(failoverContent[0], r.Value) && !reflect.DeepEqual(failoverContent[1], r.Value) {
				t.Errorf("Expecting: %v or %v, received: %v", string(failoverContent[0]), string(failoverContent[1]), string(r.Value))
			}
			received++
		})
	}
}

func testCDRsOnExpStopEngine(t *testing.T) {
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

	kCl, err := kgo.NewClient(kgo.SeedBrokers("localhost:9092"))
	if err != nil {
		t.Fatal(err)
	}
	defer kCl.Close()
	adm := kadm.NewClient(kCl)
	_, _ = adm.DeleteTopics(context.Background(), "cgrates_cdrs")
}
